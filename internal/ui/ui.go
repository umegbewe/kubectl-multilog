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
	Hierarchy   *tview.TreeView
	LogView     *tview.TextView
	SearchInput *tview.InputField
	FilterInput *tview.InputField
	StatusBar   *tview.TextView

	K8sClient         *k8s.Client
	liveTrailButton   *tview.Button
	isLiveTrailActive bool
	liveTrailCtx      context.Context
	liveTrailCancel   context.CancelFunc
	logChan           chan k8s.LogEntry
	logMutex          sync.Mutex
	liveTailStartTime time.Time
}

func NewLogExplorerTUI() (*LogExplorerTUI, error) {
	client, err := k8s.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	tui := &LogExplorerTUI{
		App:         tview.NewApplication(),
		Layout:      tview.NewFlex(),
		Hierarchy:   tview.NewTreeView(),
		LogView:     tview.NewTextView().SetDynamicColors(true),
		SearchInput: tview.NewInputField().SetLabel("Search: "),
		FilterInput: tview.NewInputField().SetLabel("Filter: "),
		StatusBar:   tview.NewTextView().SetTextAlign(tview.AlignLeft),
		K8sClient:   client,
	}

	tui.setupUI()
	return tui, nil
}

func (t *LogExplorerTUI) setupUI() {
	// Setup Hierarchy
	root := tview.NewTreeNode("Kubernetes").SetColor(tcell.ColorYellow)
	t.Hierarchy.SetRoot(root).SetCurrentNode(root)
	t.loadNamespaces(root)

	// Setup LogView
	t.LogView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)
	t.LogView.SetBorder(true).SetTitle("Logs")

	t.liveTrailButton = tview.NewButton("Start Live Trail").
		SetSelectedFunc(t.toggleLiveTrail)

	// Setup top bar
	topBar := tview.NewFlex().
		AddItem(t.SearchInput, 0, 1, false).
		AddItem(t.FilterInput, 0, 1, false).
		AddItem(t.liveTrailButton, 0, 1, false)

	mainArea := tview.NewFlex().
		AddItem(t.Hierarchy, 0, 1, true).
		AddItem(t.LogView, 0, 3, false)

	// Main layout
	t.Layout.SetDirection(tview.FlexRow).
		AddItem(topBar, 1, 0, false).
		AddItem(mainArea, 0, 1, true).
		AddItem(t.StatusBar, 1, 0, false)

	// Setup handlers
	t.setupHandlers()
}

func (t *LogExplorerTUI) loadNamespaces(root *tview.TreeNode) {
	namespaces, err := t.K8sClient.GetNamespaces()
	if err != nil {
		t.StatusBar.SetText(fmt.Sprintf("Error fetching namespaces: %v", err))
		return
	}

	for _, ns := range namespaces {
		nsNode := tview.NewTreeNode(ns).SetColor(tcell.ColorGreen).SetReference(ns)
		nsNode.SetSelectedFunc(func() {
			nsNode.ClearChildren()
			go t.loadPods(nsNode)
		})
		root.AddChild(nsNode)
	}
}

func (t *LogExplorerTUI) loadPods(nsNode *tview.TreeNode) {
	namespace := nsNode.GetReference().(string)
	t.showLoading(fmt.Sprintf("Fetching pods for %s", namespace))
	pods, err := t.K8sClient.GetPods(namespace)
	if err != nil {
		t.App.QueueUpdateDraw(func() {
			t.StatusBar.SetText(fmt.Sprintf("Error fetching pods for %s: %v", namespace, err))
		})
		return
	}

	t.App.QueueUpdateDraw(func() {
		nsNode.ClearChildren()
		for _, pod := range pods {
			podNode := tview.NewTreeNode(pod).SetColor(tcell.ColorBlue).SetReference(pod)
			podNode.SetSelectedFunc(func() {
				// Use a closure to capture the correct pod name
				podName := podNode.GetReference().(string)
				go t.loadContainers(podNode, namespace, podName)
			})
			nsNode.AddChild(podNode)
		}
		t.StatusBar.SetText(fmt.Sprintf("Loaded %d pods in namespace %s", len(pods), namespace))
	})
}

func (t *LogExplorerTUI) loadContainers(podNode *tview.TreeNode, namespace, pod string) {
	t.showLoading(fmt.Sprintf("Fetching containers for %s/%s", namespace, pod))
	containers, err := t.K8sClient.GetContainers(namespace, pod)
	if err != nil {
		t.App.QueueUpdateDraw(func() {
			t.StatusBar.SetText(fmt.Sprintf("Error fetching containers for %s/%s: %v", namespace, pod, err))
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
		t.StatusBar.SetText(fmt.Sprintf("Loaded %d containers for %s/%s", len(containers), namespace, pod))
	})
}

func (t *LogExplorerTUI) loadLogs(namespace, pod, container string) {

	t.showLoading(fmt.Sprintf("Loading logs for %s/%s/%s", namespace, pod, container))

	logs, err := t.K8sClient.GetLogs(namespace, pod, container)
	if err != nil {
		t.App.QueueUpdateDraw(func() {
			t.StatusBar.SetText(fmt.Sprintf("Error fetching logs: %v", err))
		})
		return
	}

	t.App.QueueUpdateDraw(func() {
		t.LogView.Clear()
		formattedLogs := t.FormatLogs(logs)
		t.LogView.SetText(formattedLogs)
		t.LogView.ScrollToBeginning()
		t.StatusBar.SetText(fmt.Sprintf("Logs loaded for %s/%s/%s", namespace, pod, container))
	})
}

func (t *LogExplorerTUI) setupHandlers() {
	t.SearchInput.SetDoneFunc(func(key tcell.Key) {
		t.searchLogs(t.SearchInput.GetText())
	})

	t.FilterInput.SetDoneFunc(func(key tcell.Key) {
		t.filterLogs(t.FilterInput.GetText())
	})
}

func (t *LogExplorerTUI) searchLogs(term string) {
	if term == "" {
		return
	}

	content := t.LogView.GetText(false)
	lines := strings.Split(content, "\n")
	var results []string

	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(term)) {
			results = append(results, line)
		}
	}

	t.LogView.Clear()
	t.LogView.SetText(strings.Join(results, "\n"))
	t.StatusBar.SetText(fmt.Sprintf("Found %d matches for '%s'", len(results), term))
}

func (t *LogExplorerTUI) filterLogs(filter string) {
	// TODO: Implement filtering
	t.StatusBar.SetText(fmt.Sprintf("Filtering logs with: %s", filter))
}

func (t *LogExplorerTUI) showLoading(message string) {
	t.App.QueueUpdateDraw(func() {
		t.StatusBar.SetText(message + " Loading...")
	})

	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(500 * time.Millisecond)
			t.App.QueueUpdateDraw(func() {
				currentText := t.StatusBar.GetText(false)
				t.StatusBar.SetText(currentText + ".")
			})
		}
	}()
}

func (t *LogExplorerTUI) toggleLiveTrail() {
	if t.isLiveTrailActive {
		t.stopLiveTrail()
	} else {
		t.startLiveTrail()
	}
}

func (t *LogExplorerTUI) startLiveTrail() {
	t.isLiveTrailActive = true
	t.liveTrailButton.SetLabel("Stop Live Trail")
	t.LogView.Clear()
	t.StatusBar.SetText("Live trail active")


	t.liveTailStartTime = time.Now()
	t.liveTrailCtx, t.liveTrailCancel = context.WithCancel(context.Background())
	t.logChan = make(chan k8s.LogEntry, 100)

	go t.K8sClient.StreamAllLogs(t.liveTrailCtx, t.logChan, t.liveTailStartTime)
	go t.processLiveLogs()
}

func (t *LogExplorerTUI) stopLiveTrail() {
	t.isLiveTrailActive = false
	t.liveTrailButton.SetLabel("Start Live Trail")
	t.liveTrailCancel()
	close(t.logChan)
	t.StatusBar.SetText("Live trail stopped")
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

			t.LogView.ScrollToEnd()
			fmt.Fprintf(t.LogView, "%s", formattedLog)
			}
		})	
	}
}

func (t *LogExplorerTUI) Run() error {
	t.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC || event.Rune() == 'q' {
			if t.isLiveTrailActive{
				t.stopLiveTrail()
			}
			t.App.Stop()
			return nil
		}
		return event
	})

	return t.App.SetRoot(t.Layout, true).EnableMouse(true).Run()
}
