package ui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	currentMatchIndex int
	totalMatches      int
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
		t.resetLogs()
		return
	}

	content := t.logView.GetText(false)
	lines := strings.Split(content, "\n")
	var highlightedLines []string
	var matchIndices []int

	re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(term))
	if err != nil {
		t.setStatusError(fmt.Sprintf("Invalid search term: %v", err))
		return
	}

	for i, line := range lines {
		if re.MatchString(line) {
			highlightedLine := re.ReplaceAllStringFunc(line, func(match string) string {
				return fmt.Sprintf(`["%d"][#00FF00]%s[-:-:-][""]`, i, match)
			})
			highlightedLines = append(highlightedLines, highlightedLine)
			matchIndices = append(matchIndices, i)
		} else {
			highlightedLines = append(highlightedLines, line)
		}
	}

	totalMatches = len(matchIndices)
	currentMatchIndex = 0

	t.logView.Clear()
	t.logView.SetText(strings.Join(highlightedLines, "\n"))
	t.setStatus(fmt.Sprintf("Found %d matches for '%s'", totalMatches, term))
	t.addNavigationButtons()
	if totalMatches > 0 {
		t.navigateMatches(0) // Highlight the first match
	}
}

func (t *LogExplorerTUI) filterLogs(filter string) {
	// TODO: Implement filtering
	t.statusBar.SetText(fmt.Sprintf("Filtering logs with: %s", filter))
}

func (t *LogExplorerTUI) navigateMatches(direction int) {
	if totalMatches == 0 {
		return
	}

	currentMatchIndex = (currentMatchIndex + direction + totalMatches) % totalMatches
	content := t.logView.GetText(false)
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if strings.Contains(line, "[#00FF00]") || strings.Contains(line, "[#FFA500]") {
			lines[i] = strings.ReplaceAll(strings.ReplaceAll(line, "[#00FF00]", "[#00FF00]"), "[#FFA500]", "[#00FF00]")
		}
	}

	currentLine := lines[currentMatchIndex]
	lines[currentMatchIndex] = strings.ReplaceAll(currentLine, "[#00FF00]", "[#FFA500]")

	t.logView.Clear()
	t.logView.SetText(strings.Join(lines, "\n"))
	t.logView.ScrollToHighlight()
	t.logView.Highlight(strconv.Itoa(currentMatchIndex))
	t.updateNavigationButtons()
}

func (t *LogExplorerTUI) addNavigationButtons() {
	if t.searchNavButton != nil {
		t.layout.RemoveItem(t.searchNavButton)
	}

	t.searchNavButton = tview.NewFlex().SetDirection(tview.FlexColumn)

	prevButton := tview.NewButton("< Prev").SetSelectedFunc(func() {
		t.navigateMatches(-1)
	})
	nextButton := tview.NewButton("Next >").SetSelectedFunc(func() {
		t.navigateMatches(1)
	})

	matchCountText := tview.NewTextView().SetTextAlign(tview.AlignCenter)

	t.searchNavButton.AddItem(prevButton, 0, 1, false)
	t.searchNavButton.AddItem(matchCountText, 0, 1, false)
	t.searchNavButton.AddItem(nextButton, 0, 1, false)

	t.layout.AddItem(t.searchNavButton, 1, 0, false)
	t.updateNavigationButtons()
}

func (t *LogExplorerTUI) updateNavigationButtons() {
	if t.searchNavButton == nil {
		return
	}

	matchCountText := t.searchNavButton.GetItem(1).(*tview.TextView)
	if totalMatches > 0 {
		matchCountText.SetText(fmt.Sprintf("Match %d/%d", currentMatchIndex+1, totalMatches))
	} else {
		matchCountText.SetText("No matches")
	}
}

func (t *LogExplorerTUI) resetLogs() {
	t.App.QueueUpdateDraw(func() {
		content := t.logView.GetText(false)
		lines := strings.Split(content, "\n")

		for i, line := range lines {
			lines[i] = strings.ReplaceAll(strings.ReplaceAll(line, "[#00FF00]", ""), "[#FFA500]", "")
		}

		t.logView.Clear()
		t.logView.SetText(strings.Join(lines, "\n"))
		t.logView.Highlight()
		t.setStatus("Search cleared")
		if t.searchNavButton != nil {
			t.layout.RemoveItem(t.searchNavButton)
			t.searchNavButton = nil
		}
		totalMatches = 0
		currentMatchIndex = 0
	})
}
