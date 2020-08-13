package main

import (
	"bytes"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"os"
	"strings"
	"sync"
)
type History struct {
	widget *widgets.List
	buffer *bytes.Buffer
	historyCommandTextChan *chan string
	recordHistoryCommandTextChan *chan string
	locker sync.Mutex
}

func NewHistory(historyCommandText *chan string, recordHistoryCommandTextChan *chan string) *History {
	widget := widgets.NewList()
	widget.Title = "History"
	widget.Rows = make([]string, 0)
	widget.SetRect(170, 5, 220, 40)
	widget.TextStyle = ui.NewStyle(ui.ColorYellow)
	widget.WrapText = false
	widget.BorderStyle.Fg = ui.ColorYellow
	history := &History{
		widget: widget,
		buffer: bytes.NewBuffer([]byte{}),
		historyCommandTextChan: historyCommandText,
		recordHistoryCommandTextChan: recordHistoryCommandTextChan,
	}
	go history.handleRecordCommand()
	return history
}

func (i *History) Render(elems *[]ui.Drawable) {
	*elems = append(*elems, i.widget)
}

// 记录下历史命令
func (i *History) handleRecordCommand() {
	for {
		select {
		case command := <- *i.recordHistoryCommandTextChan:
			command = strings.Trim(command, " ")
			command = fmt.Sprintf("[%s](fg:blue,mod:bold)", command)
			i.locker.Lock()
			found := false
			for index, row := range i.widget.Rows {
				// 如果有记录，就提升到第一行
				if row == command {
					newRows := i.widget.Rows[:]
					newRows[index] = newRows[0]
					newRows[0] = command
					i.widget.Rows = newRows
					found = true
					break
				}
			}
			if !found {
				// 否则添加到首部
				newRows := make([]string, 0, len(i.widget.Rows) + 1)
				newRows = append(newRows, command)
				newRows = append(newRows, i.widget.Rows...)
				i.widget.Rows = newRows
			}
			i.locker.Unlock()
		}
	}
}

func (i *History) setColor() {
	if cursor == 1 {
		i.widget.BorderStyle.Fg = ui.ColorRed
	} else {
		i.widget.BorderStyle.Fg = ui.ColorYellow
	}
}

func (i *History) HandleEvent(e ui.Event) (ui.Drawable, bool, bool) {

	i.setColor()

	if e.Type == ui.KeyboardEvent {
		i.locker.Lock()
		defer i.locker.Unlock()
		switch e.ID {
		case "<C-c>":
			os.Exit(-1)
		case "<Enter>":
			if len(i.widget.Rows) > 0 {
				command := i.widget.Rows[i.widget.SelectedRow]
				command = command[1:len(command) - len("](fg:blue,mod:bold)")]
				*i.historyCommandTextChan <- command
				return i.widget, false, false
			}
		case "<Down>":
			if len(i.widget.Rows) > 0 {
				i.widget.ScrollDown()
			}
		case "<Up>":
			if len(i.widget.Rows) > 0 {
				i.widget.ScrollUp()
			}
		case "<Left>":
			cursor = 0
			i.setColor()
			return i.widget, true, true
		case "<Right>":
			cursor = 1
			i.setColor()
			return i.widget, true, false
		default:
			return i.widget, true, false
		}
	}
	return i.widget, false, false
}