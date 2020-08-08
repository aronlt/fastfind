package main

import (
	"bytes"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"sync"
)
type Input struct {
	widget *widgets.Paragraph
	leftBuffer *bytes.Buffer
	rightBuffer *bytes.Buffer
	inputTextChan *chan string
	historyCommandChan *chan string
	locker sync.Mutex
}

func NewInput(inputTextChan *chan string, historyCommandChan *chan string) *Input {
	widget := widgets.NewParagraph()
	widget.Title = "FastFind"
	widget.Text = "[Enter what you want to search](fg:blue,mod:bold)"
	widget.SetRect(0, 0, 220, 5)
	widget.BorderStyle.Fg = ui.ColorYellow
	input := &Input{
		widget: widget,
		leftBuffer: bytes.NewBuffer([]byte{}),
		rightBuffer: bytes.NewBuffer([]byte{}),
		inputTextChan: inputTextChan,
		historyCommandChan: historyCommandChan,
	}
	go input.handleHistoryCommand()
	return input
}

func (i *Input) Render(elems *[]ui.Drawable) {
	*elems = append(*elems, i.widget)
}

func (i *Input) cursorLeft() {
	i.locker.Lock()
	defer i.locker.Unlock()
	if i.leftBuffer.Len() >= 1 {
		left := ([]rune)(i.leftBuffer.String())
		right := ([]rune)(i.rightBuffer.String())
		newRight := make([]rune, 0)
		newRight = append(newRight, left[len(left) - 1])
		newRight = append(newRight, right[:]...)
		left = left[:len(left) -1]

		i.leftBuffer.Reset()
		i.rightBuffer.Reset()
		for _, r := range left {
			i.leftBuffer.WriteRune(r)
		}
		for _, r := range newRight {
			i.rightBuffer.WriteRune(r)
		}
	}
}

func (i *Input) cursorRight() {
	i.locker.Lock()
	defer i.locker.Unlock()
	if i.rightBuffer.Len() >= 1 {
		left := ([]rune)(i.leftBuffer.String())
		right := ([]rune)(i.rightBuffer.String())

		left = append(left, right[0])
		right = right[1:]

		i.leftBuffer.Reset()
		i.rightBuffer.Reset()
		for _, r := range left {
			i.leftBuffer.WriteRune(r)
		}
		for _, r := range right {
			i.rightBuffer.WriteRune(r)
		}
	}
}

func (i *Input) backspaceCursor() {
	i.locker.Lock()
	defer i.locker.Unlock()
	if i.leftBuffer.Len() >= 1 {
		runes := ([]rune)(i.leftBuffer.String())
		runes = runes[:len(runes) - 1]
		i.leftBuffer.Reset()
		for _, r := range runes {
			i.leftBuffer.WriteRune(r)
		}
	}
}

func (i *Input) handleHistoryCommand() {
	for {
		select {
		case command := <- *i.historyCommandChan:
			i.locker.Lock()
			i.leftBuffer.Reset()
			i.rightBuffer.Reset()
			i.leftBuffer.WriteString(command)
			i.renderDisplayText()
			i.locker.Unlock()
		}
	}
}

func (i *Input) HandleEvent(e ui.Event) ui.Drawable {
	if e.Type == ui.KeyboardEvent {
		switch e.ID {
		case "<Backspace>":
			i.backspaceCursor()
		case "<Space>":
			i.leftBuffer.WriteString(" ")
		case "<Left>":
			i.cursorLeft()
		case "<Right>":
			i.cursorRight()
		case "<Escape>":
		case "<Tab>":
			i.locker.Lock()
			i.leftBuffer.WriteString("\t")
			i.locker.Unlock()
		default:
			i.locker.Lock()
			i.leftBuffer.WriteString(e.ID)
			i.locker.Unlock()
		}

		i.locker.Lock()
		i.renderDisplayText()
		i.locker.Unlock()
		return i.widget
	}
	return i.widget
}

func (i *Input) renderDisplayText() {
	i.widget.Text = i.leftBuffer.String() + "[_](fg:blue,mod:bold)" + i.rightBuffer.String()
	*i.inputTextChan <- i.leftBuffer.String() + i.rightBuffer.String()
}