package search

import (
	"strings"
	"sync"
)

const MaxLogLines = 10000

type LogBuffer struct {
	Lines           []LogLine
	mutex           sync.RWMutex
	maxLines        int
	SearchIdx       *SearchIndex
	LastSearchedPos int
	NewLinesAdded   bool
}

type LogLine struct {
	Content string
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

type SearchIndex struct {
	index map[string][]int
	mutex sync.RWMutex
}

func NewSearchIndex() *SearchIndex {
	return &SearchIndex{
		index: make(map[string][]int),
	}
}

func (si *SearchIndex) Add(word string, lineIdx int) {
	si.mutex.Lock()
	defer si.mutex.Unlock()
	si.index[word] = append(si.index[word], lineIdx)
}

func (si *SearchIndex) Search(term string) []int {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	return si.index[strings.ToLower(term)]
}
