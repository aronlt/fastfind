// Copyright 2020 lintong <lintong0825@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT license that can
// be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"log"
	"os"
)

var cursor = 0
var short = false
var fullscreen = false
var h = false

func init() {
	flag.BoolVar(&h, "h", false, "help")
	flag.BoolVar(&fullscreen, "full", false, "进入全屏模式，常驻展示")
	flag.BoolVar(&short, "short", false, "在全屏模式下, 适配小屏幕")
	flag.Usage = usage
}

func usage() {
	fmt.Fprintln(os.Stderr, `fastfind: 快速命令行检索工具，也能够作为一款辅助记忆软件`)
	flag.PrintDefaults()
}

func fastfind() {
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

func main() {
	flag.Parse()
	if h {
		flag.Usage()
		return
	}
	if fullscreen == true {
		fastfind()
	} else {
		fmt.Println("输入exit退出")
		execCommand()
	}
}
