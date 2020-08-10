// Copyright 2020 lintong <lintong0825@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT license that can
// be found in the LICENSE file.

package main

import (
	"flag"
	ui "github.com/gizak/termui/v3"
	"log"
)

var cursor = 0
var short = false

func init() {
	flag.BoolVar(&short, "short", false, "display short")
}

func main() {
	flag.Parse()

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	inputChain := make(chan string, 1)
	historyCommandChain := make(chan string, 1)
	recordHistoryCommandChain := make(chan string, 1)

	input := NewInput(&inputChain, &historyCommandChain)
	content := NewContent(&inputChain, &recordHistoryCommandChain)
	history := NewHistory(&historyCommandChain, &recordHistoryCommandChain)
	elems := make([]ui.Drawable, 0)
	input.Render(&elems)
	content.Render(&elems)
	if !short {
		history.Render(&elems)
	}
	ui.Render(elems...)

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		if e.Type != ui.KeyboardEvent {
			continue
		}
		var next bool
		var repeat = true

		for repeat {
			if short {
				elems[1], next, repeat = content.HandleEvent(e)
			} else {
				if cursor == 0 {
					elems[1], next, repeat = content.HandleEvent(e)
				} else {
					elems[2], next, repeat = history.HandleEvent(e)
				}
			}
		}
		if next {
			 elems[0] = input.HandleEvent(e)
		}
		ui.Render(elems...)
	}
}
