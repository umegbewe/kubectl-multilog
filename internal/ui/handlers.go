package ui

import (
	"time"
)


func (t *App) setupSearchHandler() {
	t.searchInput.SetChangedFunc(func(text string) {
		if t.searchTimer != nil {
			t.searchTimer.Stop()
		}
		t.searchTimer = time.AfterFunc(200*time.Millisecond, func() {
			t.App.QueueUpdateDraw(func() {
				t.performSearch(text)
			})
		})
	})

	t.caseSensitiveBtn.SetSelectedFunc(func() {
		t.Model.SearchOptions.CaseSensitive = !t.Model.SearchOptions.CaseSensitive
		t.performSearch(t.searchInput.GetText())
	})

	t.wholeWordBtn.SetSelectedFunc(func() {
		t.Model.SearchOptions.WholeWord = !t.Model.SearchOptions.WholeWord
		t.performSearch(t.searchInput.GetText())
	})

	t.regexBtn.SetSelectedFunc(func() {
		t.Model.SearchOptions.RegexEnabled = !t.Model.SearchOptions.RegexEnabled
		t.performSearch(t.searchInput.GetText())
	})

	t.prevMatchBtn.SetSelectedFunc(func() {
		t.navigateToMatch(-1)
	})

	t.nextMatchBtn.SetSelectedFunc(func() {
		t.navigateToMatch(1)
	})

}
