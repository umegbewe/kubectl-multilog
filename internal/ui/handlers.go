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
		t.searchOptions.CaseSensitive = !t.searchOptions.CaseSensitive
		t.performSearch(t.searchInput.GetText())
	})

	t.wholeWordBtn.SetSelectedFunc(func() {
		t.searchOptions.WholeWord = !t.searchOptions.WholeWord
		t.performSearch(t.searchInput.GetText())
	})

	t.regexBtn.SetSelectedFunc(func() {
		t.searchOptions.RegexEnabled = !t.searchOptions.RegexEnabled
		t.performSearch(t.searchInput.GetText())
	})

	t.prevMatchBtn.SetSelectedFunc(func() {
		t.navigateToMatch(-1)
	})

	t.nextMatchBtn.SetSelectedFunc(func() {
		t.navigateToMatch(1)
	})

}
