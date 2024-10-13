package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/umegbewe/kubectl-multilog/internal/model"
)

type App struct {
	App              *tview.Application
	layout           *tview.Flex
	hierarchy        *tview.TreeView
	logViewContainer *tview.Flex
	logTextView      *ScrollableTextView
	scrollBar        *ScrollBar
	searchInput      *tview.InputField
	statusBar        *tview.TextView
	clusterDropdown  *tview.DropDown
	liveTailBtn      *tview.Button
	caseSensitiveBtn *tview.Button
	wholeWordBtn     *tview.Button
	regexBtn         *tview.Button
	prevMatchBtn     *tview.Button
	nextMatchBtn     *tview.Button
	matchCountText   *tview.TextView
	searchTimer      *time.Timer
	Model            *model.Model
}

func LogExplorerTUI(model *model.Model) *App {
	tui := &App{
		App:         tview.NewApplication(),
		layout:      tview.NewFlex(),
		hierarchy:   tview.NewTreeView().SetGraphics(false),
		searchInput: tview.NewInputField().SetLabel("Search: ").SetLabelColor(colors.Accent),
		statusBar:   tview.NewTextView().SetTextAlign(tview.AlignLeft),
		Model:       model,
	}

	tui.setupUI()
	tui.refreshHierarchy()
	return tui
}

func (t *App) setupUI() error {
	clusters := t.Model.GetClusterNames()
	if len(clusters) == 0 {
		return fmt.Errorf("no clusters found")
	}

	t.clusterDropdown = t.initClusterDropdown(clusters)
	t.hierarchy.SetBackgroundColor(colors.Sidebar)
	t.statusBar.SetBackgroundColor(colors.TopBar)

	root := tview.NewTreeNode("Pods")
	t.hierarchy.SetRoot(root)

	t.logViewContainer = t.initLogView()

	topBar := t.initTopBar()
	mainArea := t.initMainArea()

	t.layout.SetDirection(tview.FlexRow).
		AddItem(topBar, 1, 0, false).
		AddItem(mainArea, 0, 1, true).
		AddItem(t.statusBar, 1, 0, false)

	initialCluster := t.Model.GetCurrentContext()
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
			if t.Model.LiveTailActive {
				t.stopLiveTail()
			}
			t.App.Stop()
			return nil
		}
		return event
	})

	return t.App.SetRoot(t.layout, true).EnableMouse(true).Run()
}
