package main

import (
	"strings"
	"sync"
)

const threshold = 1000
type ForceFind struct {
	wg sync.WaitGroup
}


func (f *ForceFind) Match(entries []*Entry, text string) []*Entry {
	if len(entries) < threshold {
		return f.match(entries, text)
	}

	unitLen :=  len(entries) / 4
	result := make([]*Entry, 0, len(entries))
	locker := sync.Mutex{}

	matchFunc := func(startIndex int, endIndex int) {
		matchResult := f.match(entries[startIndex:endIndex], text)
		locker.Lock()
		result = append(result, matchResult...)
		locker.Unlock()
		f.wg.Done()
	}
	f.wg.Add(4)
	go matchFunc(0, unitLen * 1)
	go matchFunc(unitLen * 1, unitLen * 2)
	go matchFunc(unitLen * 2, unitLen * 3)
	go matchFunc(unitLen * 3, len(entries))
	f.wg.Wait()
	return result
}

func (f *ForceFind) match(entries []*Entry, text string) []*Entry {
	result := make([]*Entry, 0, len(entries))
	for _, entry := range entries {
		if len(text) == 0 || strings.Contains(entry.command.String(), text) || strings.Contains(entry.explain.String(), text) {
			result = append(result, entry)
		}
	}
	return result
}
