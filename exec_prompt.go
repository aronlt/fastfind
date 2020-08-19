package main

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"os"
)


type ExePrompt struct {
	entries []*Entry
	entryMap map[string]*Entry
}

func (e *ExePrompt) completer() prompt.Completer {
	f := func(d prompt.Document) []prompt.Suggest {
		s := make([]prompt.Suggest, 0)
		for _, entry := range e.entries {
			suggest := prompt.Suggest{
				Text: entry.command.String(),
				Description: entry.explain.String(),
			}
			s = append(s, suggest)
		}
		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
	}
	return f
}


func execCommand() {
	exePrompt := &ExePrompt{
		entries: LoadContent(),
		entryMap: make(map[string]*Entry, 0),
	}
	for _, entry := range exePrompt.entries {
		exePrompt.entryMap[entry.command.String()] = entry
	}

	for {
		t := prompt.Input(">> ", exePrompt.completer())
		entry, ok := exePrompt.entryMap[t]
		if t == "exit" {
			os.Exit(0)
		}
		if ok {
			fmt.Println(entry.detail.String())
		}

	}
}