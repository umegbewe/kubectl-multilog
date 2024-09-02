package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (t *LogExplorerTUI) createClusterDropdown(clusters []string) *tview.DropDown {
	return tview.NewDropDown().
		SetLabel("Cluster: ").
		SetLabelColor(colors.Accent).
		SetFieldBackgroundColor(colors.Sidebar).
		SetFieldTextColor(colors.Highlight).
		SetOptions(clusters, func(option string, index int) {
			if err := t.k8sClient.SwitchCluster(option); err != nil {
				t.setStatusError(fmt.Sprintf("Error switching cluster: %v", err))
				return
			}
			t.setStatus(fmt.Sprintf("Switched to cluster: %s", option))
			t.refreshHierarchy()
		}).
		SetCurrentOption(0).
		SetFieldWidth(20)
}

func (t *LogExplorerTUI) createTopBar() *tview.Flex {
	startLiveTailBtn := tview.NewButton("Start Live Tail").
		SetSelectedFunc(t.toggleLiveTail).
		SetLabelColor(colors.Accent).
		SetBackgroundColor(colors.Sidebar)

	stopLiveTailBtn := tview.NewButton("Stop Live Tail").
		SetSelectedFunc(t.toggleLiveTail).
		SetLabelColor(colors.Accent).
		SetBackgroundColor(colors.Sidebar)

	return tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox().SetBorder(false).SetBackgroundColor(colors.TopBar), 1, 0, false).
		AddItem(tview.NewBox().SetBorder(true).SetBorderColor(tcell.ColorGray).SetBackgroundColor(colors.TopBar), 1, 0, false).
		AddItem(t.clusterDropdown, 0, 2, false).
		AddItem(t.searchInput.SetFieldBackgroundColor(colors.TopBar).SetFieldTextColor(tcell.ColorWhite), 0, 1, false).
		AddItem(t.filterInput.SetFieldBackgroundColor(colors.TopBar).SetFieldTextColor(tcell.ColorWhite), 0, 1, false).
		AddItem(startLiveTailBtn, 0, 1, false).
		AddItem(stopLiveTailBtn, 0, 1, false)
}
