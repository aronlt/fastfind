package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type Entry struct {
	command bytes.Buffer
	explain bytes.Buffer
	detail bytes.Buffer
}


type EntrySlice []*Entry
func (e EntrySlice) Len() int           { return len(e) }
func (e EntrySlice) Less(i, j int) bool {  return strings.Compare(e[i].command.String(), e[j].command.String()) <= 0 }
func (e EntrySlice) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }

func LoadContent() []*Entry {

	var files []string
	entries := make([]*Entry, 0)

	if runtime.GOOS == "darwin" || runtime.GOOS == "linux"{
		homedir, err := os.UserHomeDir()
		if err != nil {
			homedir = "/usr/local/bin/"
		}
		files = Files(homedir + "/.files")
	} else {
		panic("nonsupport system")
	}

	for _, file := range files {
		subEntries := make([]*Entry, 0)
		b, err := ReadContent(file)
		if err != nil {
			log.Printf("can't read content from:%s, error:%s", file, err.Error())
			continue
		}
		content := string(b)
		lines := strings.Split(content, "\n")

		var entry *Entry

		index := -1
		for i, line := range lines {
			if index == 0 || index == 1 {
				line = strings.Trim(line, "\t")
				line = strings.Trim(line, " ")
			}
			if strings.HasPrefix(line, "%%##") {
				if entry != nil {
					subEntries = append(subEntries, entry)
				}
				entry = &Entry{
					command: bytes.Buffer{},
					explain: bytes.Buffer{},
					detail:  bytes.Buffer{},
				}
				index = -1
				continue
			} else if strings.HasPrefix(line, "%%") {
				index += 1
				line = line[2:]
			}
			if entry == nil {
				continue
			}
			if line == "" && index < 2{
				continue
			}
			if index == 0 {
				file = filepath.Base(file)
				file = strings.Trim(file, ".txt")
				line  = file + "-" + line
				if len(line) > maxLineNum {
					entry.command.WriteString(line[0:maxLineNum])
				} else {
					entry.command.Reset()
					entry.command.WriteString(line)
					for entry.command.Len() < maxLineNum {
						entry.command.WriteString(" ")
					}
				}
			} else if index == 1{
				entry.explain.WriteString(line)
			} else {
				entry.detail.WriteString(line)
				entry.detail.WriteString("\n")
			}

			if i == len(lines) - 1 && entry.command.Len() != 0 {
				subEntries = append(subEntries, entry)
			}
		}
		sort.Sort((EntrySlice)(subEntries))
		entries = append(entries, subEntries...)
	}
	if len(entries) == 0 {
		panic("读取空的内容")
	}
	return entries
}


