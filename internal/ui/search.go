package ui

import (
	"fmt"
	"strings"

	"github.com/umegbewe/kubectl-multilog/internal/search"
)

func (t *App) performSearch(term string) {
	if term == "" {
		t.resetSearch()
		return
	}

	options := t.Model.SearchOptions
	searchResult, err := t.Model.PerformSearch(term, options)
	if err != nil {
		t.setStatusError(fmt.Sprintf("Search error: %v", err))
		return
	}

	t.Model.SearchResult = searchResult
	t.Model.CurrentMatchIndex = 0

	if len(searchResult.Matches) > 0 {
		t.highlightMatches()
		t.updateSearchStatus()
		t.setupSearchNavigation()
	} else {
		t.resetSearch()
	}
}

func (t *App) highlightMatches() {
	t.logView.Clear()
	lines := t.Model.LogBuffer.GetLines()
	matchIndices := make(map[int][]*search.Match)

	for idx, match := range t.Model.SearchResult.Matches {
		lineMatches := matchIndices[match.LineNumber]
		if idx == t.Model.CurrentMatchIndex {
			match.Selected = true
		} else {
			match.Selected = false
		}
		lineMatches = append(lineMatches, match)
		matchIndices[match.LineNumber] = lineMatches
	}

	for lineNumber, line := range lines {
		content := line.Content
		if matches, ok := matchIndices[lineNumber]; ok {
			highlightedContent := highlightMatchesInLineWithSelection(content, matches)
			fmt.Fprintln(t.logView, highlightedContent)
		} else {
			fmt.Fprintln(t.logView, content)
		}
	}
}

func highlightMatchesInLineWithSelection(line string, matches []*search.Match) string {
	var result strings.Builder
	lastIndex := 0
	for _, match := range matches {
		result.WriteString(line[lastIndex:match.StartIndex])
		if match.Selected {
			result.WriteString("[#FF00FF]")
		} else {
			result.WriteString("[#00FF00]")
		}
		result.WriteString(line[match.StartIndex:match.EndIndex])
		result.WriteString("[-]")
		lastIndex = match.EndIndex
	}
	result.WriteString(line[lastIndex:])
	return result.String()
}

func (t *App) updateSearchForNewLogs() {
	if t.Model.SearchResult == nil || t.Model.SearchResult.Term == "" {
		return
	}

	options := t.Model.SearchOptions

	searchResult, err := t.Model.PerformSearch(t.Model.SearchResult.Term, options)
	if err != nil {
		t.setStatusError(fmt.Sprintf("Search error: %v", err))
		return
	}

	t.Model.SearchResult = searchResult

	t.highlightMatches()
	t.updateSearchStatus()
	t.setupSearchNavigation()
}

func (t *App) setupSearchNavigation() {
	matchCount := len(t.Model.SearchResult.Matches)
	t.prevMatchBtn.SetDisabled(matchCount == 0)
	t.nextMatchBtn.SetDisabled(matchCount == 0)
	t.Model.CurrentMatchIndex = 0
	if matchCount > 0 {
		t.navigateToMatch(0)
	}
}

func (t *App) navigateToMatch(direction int) {
	matchCount := len(t.Model.SearchResult.Matches)
	if matchCount == 0 {
		return
	}

	t.Model.CurrentMatchIndex = (t.Model.CurrentMatchIndex + direction + matchCount) % matchCount

	t.highlightMatches()

	currentMatch := t.Model.SearchResult.Matches[t.Model.CurrentMatchIndex]
	t.logView.ScrollTo(currentMatch.LineNumber, 0)

	t.matchCountText.SetText(fmt.Sprintf("Match %d/%d", t.Model.CurrentMatchIndex+1, matchCount))
}

func (t *App) updateSearchStatus() {
	matchCount := len(t.Model.SearchResult.Matches)
	t.matchCountText.SetText(fmt.Sprintf("%d matches", matchCount))
	t.setStatus(fmt.Sprintf("Found %d matches for '%s'", matchCount, t.Model.SearchResult.Term))
}

func (t *App) resetSearch() {
	t.Model.SearchResult = nil
	t.Model.CurrentMatchIndex = 0
	t.logView.Clear()
	t.logView.SetText(strings.Join(t.getVisibleLogLines(), "\n"))
	t.logView.ScrollToEnd()
	t.setStatus("Search cleared")
	t.matchCountText.SetText("")
	t.prevMatchBtn.SetDisabled(true)
	t.nextMatchBtn.SetDisabled(true)
}

func (t *App) getVisibleLogLines() []string {
	lines := t.Model.LogBuffer.GetLines()
	visibleLines := make([]string, len(lines))
	for i, line := range lines {
		visibleLines[i] = line.Content
	}
	return visibleLines
}
