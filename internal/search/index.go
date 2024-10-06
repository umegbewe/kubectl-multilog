package search

import (
	"strings"
	"sync"
)

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
