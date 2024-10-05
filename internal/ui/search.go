package ui

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/umegbewe/kubectl-multilog/internal/search"
)

type SearchOptions struct {
	CaseSensitive bool
	WholeWord     bool
	RegexEnabled  bool
}

type Match struct {
	LineNumber int
	StartIndex int
	EndIndex   int
	Selected   bool
}

type SearchResult struct {
	Term    string
	Matches []Match
	Options SearchOptions
}

func (t *App) performSearch(term string) {
	if term == "" {
		t.resetSearch()
		return
	}

	t.logBuffer.LastSearchedPos = 0
	t.logBuffer.NewLinesAdded = false

	re, err := t.compileSearchRegex(term)
	if err != nil {
		t.setStatusError(fmt.Sprintf("Invalid regex: %v", err))
		return
	}

	lines := t.logBuffer.GetLines()
	matches := t.findMatches(lines, re)

	t.searchResult = &SearchResult{
		Term:    term,
		Matches: matches,
		Options: t.searchOptions,
	}

	t.highlightMatches()
	t.updateSearchStatus()
	t.setupSearchNavigation()

	if len(matches) > 0 {
		t.scrollToMatch(matches[0])
	}
}

func (t *App) compileSearchRegex(term string) (*regexp.Regexp, error) {
	if t.searchOptions.RegexEnabled {
		return regexp.Compile(term)
	}

	if t.searchOptions.WholeWord {
		term = fmt.Sprintf("\\b%s\\b", regexp.QuoteMeta(term))
	}
	if !t.searchOptions.CaseSensitive {
		term = "(?i)" + term
	}
	return regexp.Compile(term)
}

func (t *App) findMatches(lines []search.LogLine, re *regexp.Regexp) []Match {
	if t.searchOptions.RegexEnabled || t.searchOptions.WholeWord || !t.searchOptions.CaseSensitive {
		return t.fullTextSearch(lines, re)
	}
	return t.indexedSearch(t.searchResult.Term)
}

func (t *App) fullTextSearch(lines []search.LogLine, re *regexp.Regexp) []Match {
	var matches []Match
	for i, line := range lines {
		for _, idx := range re.FindAllStringIndex(line.Content, -1) {
			matches = append(matches, Match{
				LineNumber: i,
				StartIndex: idx[0],
				EndIndex:   idx[1],
			})
		}
	}
	return matches
}

func (t *App) indexedSearch(term string) []Match {
	var matches []Match
	lineIndices := t.logBuffer.SearchIdx.Search(term)
	lines := t.logBuffer.GetLines()
	for _, lineIdx := range lineIndices {
		line := lines[lineIdx].Content
		if startIdx := strings.Index(line, term); startIdx != -1 {
			matches = append(matches, Match{
				LineNumber: lineIdx,
				StartIndex: startIdx,
				EndIndex:   startIdx + len(term),
			})
		}
	}
	return matches
}

func (t *App) scrollToMatch(match Match) {
	t.logView.ScrollTo(match.LineNumber, 0)
}

func (t *App) highlightLine(line string, lineNumber int) string {
	if t.searchResult == nil {
		return line
	}

	var highlightedParts []string
	lastIndex := 0

	lineMatches := []Match{}
	for _, match := range t.searchResult.Matches {
		if match.LineNumber == lineNumber {
			lineMatches = append(lineMatches, match)
		}
	}
	sort.Slice(lineMatches, func(i, j int) bool {
		return lineMatches[i].StartIndex < lineMatches[j].StartIndex
	})

	for _, match := range lineMatches {
		if match.StartIndex > lastIndex {
			highlightedParts = append(highlightedParts, line[lastIndex:match.StartIndex])
		}
		if match.Selected {
			highlightedParts = append(highlightedParts, "[#FF00FF]"+line[match.StartIndex:match.EndIndex]+"[-:-:-]")
		} else {
			highlightedParts = append(highlightedParts, "[#00FF00]"+line[match.StartIndex:match.EndIndex]+"[-:-:-]")
		}
		lastIndex = match.EndIndex
	}

	if lastIndex < len(line) {
		highlightedParts = append(highlightedParts, line[lastIndex:])
	}

	return strings.Join(highlightedParts, "")
}

func (t *App) highlightMatches() {
	if t.searchResult == nil || len(t.searchResult.Matches) == 0 {
		return
	}

	lines := t.logBuffer.GetLines()
	highlightedLines := make([]string, len(lines))

	for i, line := range lines {
		highlightedLines[i] = t.highlightLine(line.Content, i)
	}

	t.logView.Clear()
	for _, line := range highlightedLines {
		fmt.Fprintf(t.logView, "%s\n", line)
	}
}

func (t *App) updateSearchStatus() {
	matchCount := len(t.searchResult.Matches)
	t.matchCountText.SetText(fmt.Sprintf("%d matches", matchCount))
	t.setStatus(fmt.Sprintf("Found %d matches for '%s'", matchCount, t.searchResult.Term))
}

func (t *App) updateSearchForNewLogs() {
	if t.searchResult == nil {
		return
	}

	re, err := t.compileSearchRegex(t.searchResult.Term)
	if err != nil {
		t.setStatusError(fmt.Sprintf("Invalid regex: %v", err))
	}

	lines := t.logBuffer.GetLines()
	matches := t.findMatches(lines, re)
	t.searchResult.Matches = matches

	if len(matches) > 0 {
		t.highlightMatches()
		t.updateSearchStatus()
		t.setupSearchNavigation()
	} else {
		t.resetSearch()
	}

	t.logBuffer.NewLinesAdded = false
	t.logBuffer.LastSearchedPos = len(t.logBuffer.GetLines())
}
func (t *App) setupSearchNavigation() {
	matchCount := len(t.searchResult.Matches)
	t.prevMatchBtn.SetDisabled(matchCount == 0)
	t.nextMatchBtn.SetDisabled(matchCount == 0)
	t.currentMatchIdx = 0
}

func (t *App) navigateToMatch(direction int) {
	matchCount := len(t.searchResult.Matches)
	if matchCount == 0 {
		return
	}

	if t.currentMatchIdx >= 0 && t.currentMatchIdx < matchCount {
		t.searchResult.Matches[t.currentMatchIdx].Selected = false
	}

	t.currentMatchIdx = (t.currentMatchIdx + direction + matchCount) % matchCount
	match := &t.searchResult.Matches[t.currentMatchIdx]
	match.Selected = true

	t.logView.ScrollTo(match.LineNumber, 0)
	t.logView.Highlight(fmt.Sprintf("%d", match.LineNumber))
	t.matchCountText.SetText(fmt.Sprintf("Match %d/%d", t.currentMatchIdx+1, matchCount))

	t.highlightMatches()
}

func (t *App) resetSearch() {
	t.searchResult = nil
	t.currentMatchIdx = 0
	t.logView.Clear()
	t.logView.SetText(strings.Join(t.getVisibleLogLines(), "\n"))
	t.logView.ScrollToEnd()
	t.setStatus("Search cleared")
	t.matchCountText.SetText("")
	t.prevMatchBtn.SetDisabled(true)
	t.nextMatchBtn.SetDisabled(true)
}

func (t *App) getVisibleLogLines() []string {
	lines := t.logBuffer.GetLines()
	visibleLines := make([]string, len(lines))
	for i, line := range lines {
		visibleLines[i] = line.Content
	}
	return visibleLines
}
