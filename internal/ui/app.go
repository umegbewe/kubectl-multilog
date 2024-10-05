package ui

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	k8s "github.com/umegbewe/kubectl-multilog/internal/k8sclient"
	"github.com/umegbewe/kubectl-multilog/internal/search"
)

type App struct {
	App               *tview.Application
	layout            *tview.Flex
	hierarchy         *tview.TreeView
	logView           *tview.TextView
	searchInput       *tview.InputField
	statusBar         *tview.TextView
	clusterDropdown   *tview.DropDown
	liveTailBtn       *tview.Button
	k8sClient         *k8s.Client
	LiveTailActive    bool
	liveTailCtx       context.Context
	liveTailCancel    context.CancelFunc
	logChan           chan k8s.LogEntry
	logMutex          sync.Mutex
	liveTailStartTime time.Time
	logStreamCancel   context.CancelFunc
	logBuffer         *search.LogBuffer
	searchResult      *SearchResult
	searchOptions     SearchOptions
	currentMatchIdx   int
	caseSensitiveBtn  *tview.Button
	wholeWordBtn      *tview.Button
	regexBtn          *tview.Button
	prevMatchBtn      *tview.Button
	nextMatchBtn      *tview.Button
	matchCountText    *tview.TextView
	searchTimer       *time.Timer
}

func LogExplorerTUI() (*App, error) {
	client, err := k8s.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	tui := &App{
		App:         tview.NewApplication(),
		layout:      tview.NewFlex(),
		hierarchy:   tview.NewTreeView().SetGraphics(false),
		logView:     tview.NewTextView(),
		searchInput: tview.NewInputField().SetLabel("Search: ").SetLabelColor(colors.Accent),
		statusBar:   tview.NewTextView().SetTextAlign(tview.AlignLeft),
		k8sClient:   client,
		logBuffer:   search.NewLogBuffer(),
	}

	if err := tui.setupUI(); err != nil {
		return nil, fmt.Errorf("failed to setup UI: %v", err)
	}

	tui.refreshHierarchy()
	return tui, nil
}

func (t *App) setupUI() error {
	clusters := t.k8sClient.GetClusterNames()
	if len(clusters) == 0 {
		return fmt.Errorf("no clusters found")
	}

	t.clusterDropdown = t.initClusterDropdown(clusters)
	t.hierarchy.SetBackgroundColor(colors.Sidebar)
	t.statusBar.SetBackgroundColor(colors.TopBar)

	root := tview.NewTreeNode("Pods")
	t.hierarchy.SetRoot(root)

	t.logView = t.initLogView()
	topBar := t.initTopBar()
	mainArea := t.initMainArea()

	t.layout.SetDirection(tview.FlexRow).
		AddItem(topBar, 1, 0, false).
		AddItem(mainArea, 0, 1, true).
		AddItem(t.statusBar, 1, 0, false)

	t.setupSearchHandler()

	initialCluster := t.k8sClient.GetCurrentContext()
	for i, cluster := range clusters {
		if cluster == initialCluster {
			t.clusterDropdown.SetCurrentOption(i)
			break
		}
	}

	return nil
}

func (t *App) Run() error {
	t.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			if t.LiveTailActive {
				t.stopLiveTail()
			}
			t.App.Stop()
			return nil
		}
		return event
	})

	return t.App.SetRoot(t.layout, true).EnableMouse(true).Run()
}
