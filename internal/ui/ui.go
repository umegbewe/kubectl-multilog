package ui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	k8s "github.com/umegbewe/kubectl-multilog/internal/k8s"
)

type LogExplorerTUI struct {
	App         *tview.Application
	Layout      *tview.Flex
	hierarchy   *tview.TreeView
	logView     *tview.TextView
	searchInput *tview.InputField
	filterInput *tview.InputField
	statusBar   *tview.TextView

	k8sClient         *k8s.Client
	liveTailButton    *tview.Button
	isLiveTailActive  bool
	liveTailCtx       context.Context
	liveTailCancel    context.CancelFunc
	logChan           chan k8s.LogEntry
	logMutex          sync.Mutex
	liveTailStartTime time.Time
}

var (
	textColor       = tcell.NewHexColor(0x569cdb)
	backgroundColor = tcell.ColorDefault
	sidebarColor    = tcell.NewRGBColor(51, 51, 51)
)

func NewLogExplorerTUI() (*LogExplorerTUI, error) {
	client, err := k8s.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	tui := &LogExplorerTUI{
		App:         tview.NewApplication(),
		Layout:      tview.NewFlex(),
		hierarchy:   tview.NewTreeView().SetGraphics(false),
		logView:     tview.NewTextView(),
		searchInput: tview.NewInputField().SetLabel("Search: "),
		filterInput: tview.NewInputField().SetLabel("Filter: "),
		statusBar:   tview.NewTextView().SetTextAlign(tview.AlignLeft),
		k8sClient:   client,
	}

	tui.setupUI()
	return tui, nil
}

func (t *LogExplorerTUI) setupUI() {
	root := tview.NewTreeNode("Kubernetes").SetColor(textColor)
	t.hierarchy.SetRoot(root).SetCurrentNode(root)
	t.hierarchy.SetBackgroundColor(sidebarColor)
	t.statusBar.SetBackgroundColor(backgroundColor)
	t.loadNamespaces(root)

	t.logView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetScrollable(true).
		SetWordWrap(true)

	t.logView.SetTitle("Logs")
	t.logView.SetBackgroundColor(backgroundColor)
	t.logView.SetBorder(true)

	t.liveTailButton = tview.NewButton("Start Live Tail").
		SetSelectedFunc(t.toggleLiveTail)

	topBar := tview.NewFlex().
		AddItem(t.searchInput, 0, 1, false).
		AddItem(t.filterInput, 0, 1, false).
		AddItem(t.liveTailButton, 0, 1, false)

	mainArea := tview.NewFlex().
		AddItem(t.hierarchy, 0, 1, true).
		AddItem(t.logView, 0, 5, false)

	t.Layout.SetDirection(tview.FlexRow).
		AddItem(topBar, 1, 0, false).
		AddItem(mainArea, 0, 1, true).
		AddItem(t.statusBar, 1, 0, false)

	t.setupHandlers()
}

func (t *LogExplorerTUI) loadNamespaces(root *tview.TreeNode) {
	namespaces, err := t.k8sClient.GetNamespaces()
	if err != nil {
		t.statusBar.SetText(fmt.Sprintf("Error fetching namespaces: %v", err))
		return
	}

	for _, ns := range namespaces {
		nsNode := createTreeNode(ns, false).SetReference(ns)
		setNodeWithToggleIcon(nsNode, ns, func() {
			nsNode.ClearChildren()
			go t.loadPods(nsNode)
		})
		root.AddChild(nsNode)
	}
}

func (t *LogExplorerTUI) loadPods(nsNode *tview.TreeNode) {
	namespace := nsNode.GetReference().(string)
	t.showLoading(fmt.Sprintf("Fetching pods for %s", namespace))
	pods, err := t.k8sClient.GetPods(namespace)
	if err != nil {
		t.App.QueueUpdateDraw(func() {
			t.statusBar.SetText(fmt.Sprintf("Error fetching pods for %s: %v", namespace, err))
		})
		return
	}

	t.App.QueueUpdateDraw(func() {
		nsNode.ClearChildren()
		for _, pod := range pods {
			podNode := createTreeNode(pod, false).SetReference(pod)
			setNodeWithToggleIcon(podNode, pod, func() {
				podNode.ClearChildren()
				// Use a closure to capture the correct pod name
				podName := podNode.GetReference().(string)
				go t.loadContainers(podNode, namespace, podName)
			})
			nsNode.AddChild(podNode)
		}
		t.statusBar.SetText(fmt.Sprintf("Loaded %d pods in namespace %s", len(pods), namespace))
	})
}

func (t *LogExplorerTUI) loadContainers(podNode *tview.TreeNode, namespace, pod string) {
	t.showLoading(fmt.Sprintf("Fetching containers for %s/%s", namespace, pod))
	containers, err := t.k8sClient.GetContainers(namespace, pod)
	if err != nil {
		t.App.QueueUpdateDraw(func() {
			t.statusBar.SetText(fmt.Sprintf("Error fetching containers for %s/%s: %v", namespace, pod, err))
		})
		return
	}

	t.App.QueueUpdateDraw(func() {
		podNode.ClearChildren() // Clear existing children to avoid duplicates
		for _, container := range containers {
			containerNode := tview.NewTreeNode(container).SetColor(tcell.ColorRed).SetReference(container)
			containerNode.SetSelectedFunc(func() {
				go t.loadLogs(namespace, pod, container)
			})
			podNode.AddChild(containerNode)
		}
		t.statusBar.SetText(fmt.Sprintf("Loaded %d containers for %s/%s", len(containers), namespace, pod))
	})
}

func (t *LogExplorerTUI) loadLogs(namespace, pod, container string) {

	t.showLoading(fmt.Sprintf("Loading logs for %s/%s/%s", namespace, pod, container))

	logs, err := t.k8sClient.GetLogs(namespace, pod, container)
	if err != nil {
		t.App.QueueUpdateDraw(func() {
			t.statusBar.SetText(fmt.Sprintf("Error fetching logs: %v", err))
		})
		return
	}

	t.App.QueueUpdateDraw(func() {
		t.logView.Clear()
		formattedLogs := t.FormatLogs(logs)
		t.logView.SetText(formattedLogs)
		t.logView.ScrollToBeginning()
		t.statusBar.SetText(fmt.Sprintf("Logs loaded for %s/%s/%s", namespace, pod, container))
	})
}

func (t *LogExplorerTUI) setupHandlers() {
	t.searchInput.SetDoneFunc(func(key tcell.Key) {
		t.searchLogs(t.searchInput.GetText())
	})

	t.filterInput.SetDoneFunc(func(key tcell.Key) {
		t.filterLogs(t.filterInput.GetText())
	})
}

func (t *LogExplorerTUI) searchLogs(term string) {
	if term == "" {
		return
	}

	content := t.logView.GetText(false)
	lines := strings.Split(content, "\n")
	var results []string

	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(term)) {
			results = append(results, line)
		}
	}

	t.logView.Clear()
	t.logView.SetText(strings.Join(results, "\n"))
	t.statusBar.SetText(fmt.Sprintf("Found %d matches for '%s'", len(results), term))
}

func (t *LogExplorerTUI) filterLogs(filter string) {
	// TODO: Implement filtering
	t.statusBar.SetText(fmt.Sprintf("Filtering logs with: %s", filter))
}

func (t *LogExplorerTUI) showLoading(message string) {
	t.App.QueueUpdateDraw(func() {
		t.statusBar.SetText(message + " Loading...")
	})

	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(500 * time.Millisecond)
			t.App.QueueUpdateDraw(func() {
				currentText := t.statusBar.GetText(false)
				t.statusBar.SetText(currentText + ".")
			})
		}
	}()
}
func createTreeNode(label string, isLeaf bool) *tview.TreeNode {
	node := tview.NewTreeNode("")

	if isLeaf {
		node.SetText(fmt.Sprintf("  %s", label)) // Leaf nodes don't get icons
	} else {
		node.SetText(fmt.Sprintf("▶ %s", label))
		node.SetExpanded(false) // Initially collapsed
	}

	return node
}

func setNodeWithToggleIcon(node *tview.TreeNode, label string, toggleFunc func()) {
	node.SetSelectedFunc(func() {
		if node.IsExpanded() {
			node.CollapseAll()
			node.SetText(fmt.Sprintf("▶ %s", label))
		} else {
			node.ExpandAll()
			node.SetText(fmt.Sprintf("▼ %s", label))
		}
		toggleFunc()
	})
}

func (t *LogExplorerTUI) toggleLiveTail() {
	if t.isLiveTailActive {
		t.stopLiveTail()
	} else {
		t.startLiveTail()
	}
}

func (t *LogExplorerTUI) startLiveTail() {
	t.isLiveTailActive = true
	t.liveTailButton.SetLabel("Stop Live Tail")
	t.logView.Clear()
	t.statusBar.SetText("Live tail active")

	t.liveTailStartTime = time.Now()
	t.liveTailCtx, t.liveTailCancel = context.WithCancel(context.Background())
	t.logChan = make(chan k8s.LogEntry, 100)

	go t.k8sClient.StreamAllLogs(t.liveTailCtx, t.logChan, t.liveTailStartTime)
	go t.processLiveLogs()
}

func (t *LogExplorerTUI) stopLiveTail() {
	t.isLiveTailActive = false
	t.liveTailButton.SetLabel("Start Live Tail")
	t.liveTailCancel()
	close(t.logChan)
	t.statusBar.SetText("Live tail stopped")
}

func (t *LogExplorerTUI) processLiveLogs() {
	for logEntry := range t.logChan {
		t.App.QueueUpdateDraw(func() {
			t.logMutex.Lock()
			defer t.logMutex.Unlock()

			if logEntry.Timestamp.After(t.liveTailStartTime) {
				formattedLog := fmt.Sprintf("[%s] [%s/%s/%s] %s: %s\n",
					logEntry.Timestamp.Format(time.RFC3339),
					logEntry.Namespace,
					logEntry.Pod,
					logEntry.Container,
					logEntry.Level,
					logEntry.Message)

				t.logView.ScrollToEnd()
				fmt.Fprintf(t.logView, "%s", formattedLog)
			}
		})
	}
}

func (t *LogExplorerTUI) Run() error {
	t.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC || event.Rune() == 'q' {
			if t.isLiveTailActive {
				t.stopLiveTail()
			}
			t.App.Stop()
			return nil
		}
		return event
	})

	return t.App.SetRoot(t.Layout, true).EnableMouse(true).Run()
}
