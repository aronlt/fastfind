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
	patterns := strings.Split(text, "&")
	if len(patterns) == 0 {
		return make([]*Entry, 0)
	}

	if len(entries) < threshold {
		return f.match(entries, patterns)
	}

	unitLen :=  len(entries) / 4
	result := make([]*Entry, 0, len(entries))
	locker := sync.Mutex{}

	matchFunc := func(startIndex int, endIndex int) {
		matchResult := f.match(entries[startIndex:endIndex], patterns)
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

func (f *ForceFind) match(entries []*Entry, patterns []string) []*Entry {
	result := make([]*Entry, 0, len(entries))

	// 先筛选出第一个pattern的全部数据
	pattern := strings.TrimSpace(patterns[0])
	for _, entry := range entries {
		if len(pattern) == 0 || strings.Contains(entry.command.String(), pattern) || strings.Contains(entry.explain.String(), pattern) {
			result = append(result, entry)
		}
	}

	// 再根据剩余的pattern，进行筛选
	for _, pattern := range patterns[1:] {
		pattern = strings.TrimSpace(pattern)
		tmp := make([]*Entry, 0, len(result))
		for _, entry := range result {
			if len(pattern) == 0 || strings.Contains(entry.command.String(), pattern) || strings.Contains(entry.explain.String(), pattern) {
				tmp = append(tmp, entry)
			}
		}
		result = tmp
	}
	return result
}
