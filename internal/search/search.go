package search

import (
	"regexp"
)

type SearchResult struct {
	Term    string
	Matches []*Match
}

type Match struct {
	LineNumber int
	StartIndex int
	EndIndex   int
	Content    string
	Selected   bool
}

type SearchOptions struct {
	CaseSensitive bool
	WholeWord     bool
	RegexEnabled  bool
}

func PerformSearch(lines []string, term string, options SearchOptions) (*SearchResult, error) {
	re, err := compileSearchRegex(term, options)
	if err != nil {
		return nil, err
	}

	matches := findMatches(lines, re)

	return &SearchResult{
		Term:    term,
		Matches: matches,
	}, nil

}

func compileSearchRegex(term string, options SearchOptions) (*regexp.Regexp, error) {
	var pattern string

	if options.RegexEnabled {
		pattern = term
	} else {
		pattern = regexp.QuoteMeta(term)
	}

	if options.WholeWord {
		pattern = `\b` + pattern + `\b`
	}

	flags := ""
	if !options.CaseSensitive {
		flags = "(?i)"
	}

	re, err := regexp.Compile(flags + pattern)
	if err != nil {
		return nil, err
	}

	return re, nil
}

func findMatches(lines []string, re *regexp.Regexp) []*Match {
	matches := []*Match{}
	for i, line := range lines {
		for _, loc := range re.FindAllStringIndex(line, -1) {
			matches = append(matches, &Match{
				LineNumber: i,
				StartIndex: loc[0],
				EndIndex:   loc[1],
				Content:    line,
			})
		}
	}
	return matches
}
