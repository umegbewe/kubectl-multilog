package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

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
