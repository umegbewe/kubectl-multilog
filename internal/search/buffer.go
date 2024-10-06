package search

import (
	"strings"
	"sync"
)
const MaxLogLines = 10000

type LogLine struct {
	Content string
}

type LogBuffer struct {
	Lines           []LogLine
	mutex           sync.RWMutex
	maxLines        int
	SearchIdx       *SearchIndex
	LastSearchedPos int
	NewLinesAdded   bool
}

func NewLogBuffer() *LogBuffer {
	return &LogBuffer{
		Lines:     make([]LogLine, 0, MaxLogLines),
		maxLines:  MaxLogLines,
		SearchIdx: NewSearchIndex(),
	}
}

func (lb *LogBuffer) AddLine(content string) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	if len(lb.Lines) >= lb.maxLines {
		lb.Lines = lb.Lines[1:]
		lb.LastSearchedPos--
		if lb.LastSearchedPos < 0 {
			lb.LastSearchedPos = 0
		}
	}
	lb.Lines = append(lb.Lines, LogLine{Content: content})

	words := strings.Fields(strings.ToLower(content))
	lineIdx := len(lb.Lines) - 1
	for _, word := range words {
		lb.SearchIdx.Add(word, lineIdx)
	}

	lb.NewLinesAdded = true
}

func (lb *LogBuffer) GetLines() []LogLine {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	return lb.Lines
}

func (lb *LogBuffer) Clear() {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	lb.Lines = make([]LogLine, 0, lb.maxLines)
	lb.SearchIdx = NewSearchIndex()
}

func (lb *LogBuffer) GetLinesContent() []string {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	contents := make([]string, len(lb.Lines))
	for i, line := range lb.Lines {
		contents[i] = line.Content
	}
	return contents
}
