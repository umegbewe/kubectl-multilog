package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (t *LogExplorerTUI) createClusterDropdown(clusters []string) *tview.DropDown {
	return tview.NewDropDown().
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
	topBar := tview.NewFlex().SetDirection(tview.FlexColumn)

	t.clusterDropdown.SetLabel("Cluster: ")
	t.clusterDropdown.SetLabelColor(colors.Accent)
	t.clusterDropdown.SetFieldTextColor(colors.Text)
	t.clusterDropdown.SetBackgroundColor(colors.TopBar)
	t.clusterDropdown.SetListStyles(tcell.StyleDefault.Background(colors.Sidebar), tcell.StyleDefault.Background(colors.Highlight).Foreground(colors.Text))

	t.searchInput.SetLabel("Search: ")
	t.searchInput.SetLabelColor(colors.Accent)
	t.searchInput.SetFieldBackgroundColor(colors.TopBar)
	t.searchInput.SetFieldTextColor(colors.Text)

	t.filterInput.SetLabel("Filter: ")
    t.filterInput.SetLabelColor(colors.Accent)
    t.filterInput.SetFieldBackgroundColor(colors.TopBar)
    t.filterInput.SetFieldTextColor(colors.Text)

	t.liveTailBtn = tview.NewButton("Start Live Tail")
	t.liveTailBtn.SetLabelColor(colors.Text)
	t.liveTailBtn.SetBackgroundColor(colors.TopBar)
	t.liveTailBtn.SetSelectedFunc(t.toggleLiveTail)

	topBar.AddItem(t.clusterDropdown, 0, 1, false)
	topBar.AddItem(tview.NewBox().SetBackgroundColor(colors.TopBar), 1, 0, false)
	topBar.AddItem(t.searchInput, 0, 1, false)
	topBar.AddItem(tview.NewBox().SetBackgroundColor(colors.TopBar), 1, 0, false)
	topBar.AddItem(t.filterInput, 0, 1, false)
	topBar.AddItem(tview.NewBox().SetBackgroundColor(colors.TopBar), 1, 0, false)
	topBar.AddItem(t.liveTailBtn, 0, 1, false)

	topSection := tview.NewFlex().SetDirection(tview.FlexRow)
	topSection.AddItem(topBar, 1, 0, false)
	topSection.AddItem(tview.NewBox().SetBackgroundColor(colors.TopBar), 1, 0, false)

	return topSection
}
