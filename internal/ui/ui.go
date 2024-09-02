package ui

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	k8s "github.com/umegbewe/kubectl-multilog/internal/k8s"
)

type LogExplorerTUI struct {
	App             *tview.Application
	layout          *tview.Flex
	hierarchy       *tview.TreeView
	logView         *tview.TextView
	searchInput     *tview.InputField
	filterInput     *tview.InputField
	statusBar       *tview.TextView
	clusterDropdown *tview.DropDown
	liveTailBtn     *tview.Button
	k8sClient         *k8s.Client
	isLiveTailActive  bool
	liveTailCtx       context.Context
	liveTailCancel    context.CancelFunc
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
		layout:      tview.NewFlex(),
		hierarchy:   tview.NewTreeView().SetGraphics(false),
		logView:     tview.NewTextView(),
		searchInput: tview.NewInputField().SetLabel("Search: ").SetLabelColor(colors.Accent),
		filterInput: tview.NewInputField().SetLabel("Filter: ").SetLabelColor(colors.Accent),
		statusBar:   tview.NewTextView().SetTextAlign(tview.AlignLeft),
		k8sClient:   client,
	}

	if err := tui.setupUI(); err != nil {
		return nil, fmt.Errorf("failed to setup UI: %v", err)
	}
	return tui, nil
}

func (t *LogExplorerTUI) setupUI() error {
	clusters := t.k8sClient.GetClusterNames()
	if len(clusters) == 0 {
		return fmt.Errorf("no clusters found")
	}

	t.clusterDropdown = t.createClusterDropdown(clusters)
	t.hierarchy.SetBackgroundColor(colors.Sidebar)
	t.statusBar.SetBackgroundColor(colors.TopBar)

	root := tview.NewTreeNode("Clusters")
	t.hierarchy.SetRoot(root)

	t.logView = t.createLogView()
	topBar := t.createTopBar()
	mainArea := t.createMainArea()

	t.layout.SetDirection(tview.FlexRow).
		AddItem(topBar, 3, 0, false).
		AddItem(mainArea, 0, 1, true).
		AddItem(t.statusBar, 1, 0, false)

	t.setupHandlers()
	return nil
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

	return t.App.SetRoot(t.layout, true).EnableMouse(true).Run()
}
