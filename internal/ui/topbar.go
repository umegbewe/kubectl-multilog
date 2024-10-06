package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (t *App) initClusterDropdown(clusters []string) *tview.DropDown {
	return tview.NewDropDown().
		SetOptions(clusters, func(option string, index int) {
			if err := t.Model.SwitchCluster(option); err != nil {
				t.setStatusError(fmt.Sprintf("Error switching cluster: %v", err))
				return
			}
			t.setStatus(fmt.Sprintf("Switched to cluster: %s", option))
			t.refreshHierarchy()
		}).
		SetCurrentOption(0).
		SetFieldWidth(20)
}

func (t *App) initTopBar() *tview.Flex {
	topBar := tview.NewFlex().SetDirection(tview.FlexColumn)

	t.clusterDropdown.SetLabel("Context: ")
	t.clusterDropdown.SetLabelColor(colors.Accent)
	t.clusterDropdown.SetFieldBackgroundColor(colors.TopBar)
	t.clusterDropdown.SetFieldTextColor(colors.Text)
	t.clusterDropdown.SetFieldWidth(100)
	t.clusterDropdown.SetBackgroundColor(colors.TopBar)
	t.clusterDropdown.SetListStyles(
		tcell.StyleDefault.Background(colors.Sidebar), 
		tcell.StyleDefault.Background(colors.Highlight).Foreground(colors.Text),
	)

	t.searchInput.SetLabel(" Search: ").SetLabelColor(colors.Accent)
	t.searchInput.SetFieldBackgroundColor(colors.TopBar)
	t.searchInput.SetBackgroundColor(colors.TopBar)
	t.searchInput.SetFieldTextColor(colors.Text)

	t.caseSensitiveBtn = tview.NewButton("Aa").SetSelectedFunc(func() {
		t.Model.SearchOptions.CaseSensitive = !t.Model.SearchOptions.CaseSensitive
		t.performSearch(t.searchInput.GetText())
	})
	t.wholeWordBtn = tview.NewButton("W").SetSelectedFunc(func() {
		t.Model.SearchOptions.WholeWord = !t.Model.SearchOptions.WholeWord
		t.performSearch(t.searchInput.GetText())
	})
	t.regexBtn = tview.NewButton(".*").SetSelectedFunc(func() {
		t.Model.SearchOptions.RegexEnabled = !t.Model.SearchOptions.RegexEnabled
		t.performSearch(t.searchInput.GetText())
	})
	t.prevMatchBtn = tview.NewButton("◀").SetSelectedFunc(func() {
		t.navigateToMatch(-1)
	})
	t.nextMatchBtn = tview.NewButton("▶").SetSelectedFunc(func() {
		t.navigateToMatch(1)
	})
	t.matchCountText = tview.NewTextView().SetTextAlign(tview.AlignRight)
	t.matchCountText.SetBackgroundColor(colors.TopBar)
	t.matchCountText.SetTextAlign(tview.AlignCenter)

	t.liveTailBtn = createButton("Live", colors.Button, t.toggleLiveTail)
	t.liveTailBtn.SetDisabled(true)

	searchBar := tview.NewFlex().
		AddItem(t.searchInput, 0, 1, false).
		AddItem(tview.NewBox().SetBackgroundColor(colors.TopBar), 1, 0, false).
		AddItem(createButton("Aa", colors.Button,  t.toggleCaseSensitive), 4, 0, false).
		AddItem(tview.NewBox().SetBackgroundColor(colors.TopBar), 1, 0, false).
		AddItem(createButton("W", colors.Button, t.toggleWholeWord), 3, 0, false).
		AddItem(tview.NewBox().SetBackgroundColor(colors.TopBar), 1, 0, false).
		AddItem(createButton(".*", colors.Button, t.toggleRegex), 4, 0, false).
		AddItem(tview.NewBox().SetBackgroundColor(colors.TopBar), 1, 0, false).
		AddItem(createButton("◀", colors.NavButton, func() { t.navigateToMatch(-1) }), 3, 0, false).
		AddItem(t.matchCountText, 12, 0, false).
		AddItem(createButton("▶", colors.NavButton, func() { t.navigateToMatch(1) }), 3, 0, false)

	t.setupSearchHandler()
	
	topBar.AddItem(t.clusterDropdown, 0, 1, false)
	topBar.AddItem(searchBar, 0, 3, false)
	topBar.AddItem(t.liveTailBtn, 0, 1, false)

	return topBar
}


func createButton(label string, bgColor tcell.Color, selectedFunc func()) *tview.Button {
	return tview.NewButton(label).
		SetLabelColor(colors.Text).
		SetStyle(tcell.StyleDefault.Background(bgColor)).
		SetSelectedFunc(selectedFunc)
}

func (t *App) toggleCaseSensitive() {
	t.Model.SearchOptions.CaseSensitive = !t.Model.SearchOptions.CaseSensitive
	t.performSearch(t.searchInput.GetText())
}

func (t *App) toggleWholeWord() {
	t.Model.SearchOptions.WholeWord = !t.Model.SearchOptions.WholeWord
	t.performSearch(t.searchInput.GetText())
}

func (t *App) toggleRegex() {
	t.Model.SearchOptions.RegexEnabled = !t.Model.SearchOptions.RegexEnabled
	t.performSearch(t.searchInput.GetText())
}