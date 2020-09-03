package main

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"math/rand"
	"os"
	"strings"
)

var maxLineNum = 80

var maxContentLineSize = 1
var displayLineSize = 30

type WidgetType int32

const (
	ListType      WidgetType = 0
	ContentType      WidgetType = 1
)

type Content struct {
	// 展示列表
	listWidget *widgets.List
	// 展示详情
	contentWidget *widgets.Paragraph
	contentStartIdx int
	contentEndIdx int

	// 当前所处的模式
	widgetType WidgetType

	listEntries []*Entry
	entries []*Entry
	// 用于匹配命令
	inputTextChan *chan string
	// 记录命令
	recHistoryCommandChan *chan string
	// 搜索接口
	finder MultiFind
	// 处理函数
	handlerMap map[string]InputHandler
}

type InputHandler func(c *Content) (ui.Drawable, bool, bool)

func NewContent(inputTextChan *chan string,  recHistoryCommandChan *chan string) *Content {
	listWidget := widgets.NewList()
	listWidget.Title = "Content"
	listWidget.Rows = make([]string, 0)
	listWidget.TextStyle = ui.NewStyle(ui.ColorYellow)
	listWidget.BorderStyle.Fg = ui.ColorRed
	listWidget.WrapText = false
	if short {
		listWidget.SetRect(0, 5, 60, 40)
		maxLineNum -= 40
	} else {
		listWidget.SetRect(0, 5, 170, 40)
	}

	contentWidget := widgets.NewParagraph()
	contentWidget.Title = "Content"
	contentWidget.Text = ""
	if short {
		contentWidget.SetRect(0, 5, 60, 40)
	} else {
		contentWidget.SetRect(0, 5, 170, 40)
	}
	contentWidget.BorderStyle.Fg = ui.ColorRed

	content := &Content{
		listWidget:    listWidget,
		contentWidget: contentWidget,
		recHistoryCommandChan: recHistoryCommandChan,
		inputTextChan: inputTextChan,
		contentStartIdx: 0,
		contentEndIdx: 0,
		widgetType: ListType,
		finder: &ForceFind{},
		handlerMap: make(map[string]InputHandler, 0),
	}

	content.register()
	content.loadContent()
	go content.handleInputText()
	return content
}

func (c *Content) register() {

	c.handlerMap["<C-r>"] = func(c *Content) (ui.Drawable, bool, bool) {
		c.reset()
		c.loadContent()
		return c.listWidget, false, false
	}

	c.handlerMap["<PageDown>"] = func(c *Content) (ui.Drawable, bool, bool) {
		c.reset()
		c.widgetType = ContentType
		c.randomPick()
		c.contentInitPage()
		return c.contentWidget, false, false
	}

	c.handlerMap["<PageUp>"] = c.handlerMap["<PageDown>"]

	c.handlerMap["<Enter>"] = func(c *Content) (ui.Drawable, bool, bool) {
		if len(c.listWidget.Rows) > 0 {
			if c.widgetType == ListType {
				c.contentInitPage()
				c.widgetType = ContentType
				// 记录
				*c.recHistoryCommandChan <- c.listEntries[c.listWidget.SelectedRow].command.String()
			}
		}
		return c.contentWidget, false, false
	}

	c.handlerMap["<Down>"] = func(c *Content) (ui.Drawable, bool, bool) {
		if len(c.listWidget.Rows) > 0 {
			if c.widgetType == ListType {
				c.listWidget.ScrollDown()
				return c.listWidget, false, false
			} else {
				c.contentDownPage()
				return c.contentWidget, false, false
			}
		} else {
			return c.listWidget, false, false
		}
	}

	c.handlerMap["<Up>"] = func(c *Content) (ui.Drawable, bool, bool) {
		if len(c.listWidget.Rows) > 0 {
			if c.widgetType == ListType {
				c.listWidget.ScrollUp()
				return c.listWidget, false, false
			} else {
				c.contentUpPage()
				return c.contentWidget, false, false
			}
		} else {
			return c.listWidget, false, false
		}
	}

	c.handlerMap["<Left>"] = func(c *Content) (ui.Drawable, bool, bool) {
		c.setColor()
		if c.widgetType == ListType {
			return c.listWidget, false, false
		} else {
			return c.contentWidget, false, false
		}
	}

	c.handlerMap["<Right>"] = func(c *Content) (ui.Drawable, bool, bool) {
		cursor = 1
		c.setColor()
		if c.widgetType == ListType {
			return c.listWidget, true, true
		} else {
			return c.contentWidget, true, true
		}
	}
}

func (c *Content) randomPick() {
	var idx int
	if len(c.listEntries) == 0 {
		idx = rand.Intn(len(c.listEntries))
	} else {
		idx = rand.Intn(len(c.listWidget.Rows))
	}
	c.listWidget.SelectedRow = idx
}

// 加载并解析文件
func (c *Content) loadContent() {
	c.entries = LoadContent()
	c.listWidget.Rows = make([]string, 0, len(c.entries))
	c.listEntries = make([]*Entry, 0, len(c.entries))
	for _, entry := range c.entries {
		text := c.getRowText(entry)
		c.listWidget.Rows = append(c.listWidget.Rows, text)
		c.listEntries = append(c.listEntries, entry)
	}
}

func (c *Content) Render(elems *[]ui.Drawable)  {
	*elems = append(*elems, c.listWidget)
}

// 实时匹配
func (c *Content) handleInputText() {
	for {
		select {
		case text := <- *c.inputTextChan:
			c.reRenderRows(text)
		}
	}
}

// 展示
func (c *Content) getRowText(entry *Entry) string {
	if short {
		return fmt.Sprintf("[%s](fg:blue,mod:bold)",  entry.command.String())
	}
	return fmt.Sprintf("[%s](fg:blue,mod:bold) [%s](fg:red,mod:light)",  entry.command.String(), entry.explain.String())
}

// 匹配列表
func (c *Content) reRenderRows(text string) {
	newRow := make([]string, 0)
	newEntries := make([]*Entry, 0)

	matchEntries := c.finder.Match(c.entries, text)

	for _, entry := range matchEntries{
		text := c.getRowText(entry)
		newRow = append(newRow, text)
		newEntries = append(newEntries, entry)
	}
	c.listWidget.Rows = newRow
	c.listEntries = newEntries
	c.listWidget.SelectedRow = 0
}

func (c *Content) setColor() {
	if cursor == 0 {
		c.contentWidget.BorderStyle.Fg = ui.ColorRed
		c.listWidget.BorderStyle.Fg = ui.ColorRed
	} else {
		c.contentWidget.BorderStyle.Fg = ui.ColorYellow
		c.listWidget.BorderStyle.Fg = ui.ColorYellow
	}
}

func (c *Content)contentDownPage() {
	idx := c.listWidget.SelectedRow
	lines := strings.Split(c.listEntries[idx].detail.String(), "\n")
	if c.contentEndIdx >= len(lines) {
		return
	}

	oldStart := c.contentStartIdx
	oldEnd := c.contentEndIdx

	// 满足往下走的条件
	if c.contentStartIdx < len(lines) && c.contentStartIdx + maxContentLineSize < len(lines) &&
		c.contentEndIdx < len(lines) && c.contentEndIdx + maxContentLineSize < len(lines) {
		c.contentStartIdx += maxContentLineSize
		c.contentEndIdx += maxContentLineSize
	}

	if c.contentStartIdx >= c.contentEndIdx {
		c.contentStartIdx = oldStart
		c.contentEndIdx = oldEnd
	}

	c.contentWidget.Text = c.decorateText(idx, c.contentStartIdx, c.contentEndIdx)
}

func (c *Content)contentUpPage() {
	if c.contentStartIdx == 0 {
		return
	}
	idx := c.listWidget.SelectedRow
	oldStart := c.contentStartIdx
	oldEnd := c.contentEndIdx

	if c.contentEndIdx > 0 &&  c.contentEndIdx - maxContentLineSize > 0  && c.contentStartIdx > 0 && c.contentStartIdx - maxContentLineSize > 0 {
		c.contentEndIdx -= maxContentLineSize
		c.contentStartIdx -= maxContentLineSize
	}

	if c.contentStartIdx >= c.contentEndIdx {
		c.contentStartIdx = oldStart
		c.contentEndIdx = oldEnd
	}
	c.contentWidget.Text = c.decorateText(idx, c.contentStartIdx, c.contentEndIdx)
}

func (c *Content)contentInitPage() {
	idx := c.listWidget.SelectedRow
	lines := strings.Split(c.listEntries[idx].detail.String(), "\n")
	c.contentStartIdx = 0
	c.contentEndIdx = displayLineSize
	if c.contentEndIdx > len(lines) {
		c.contentEndIdx = len(lines)
	}
	c.contentWidget.Text = c.decorateText(idx, c.contentStartIdx, c.contentEndIdx)
}

func (c *Content)decorateText(idx int, start int, end int) string {
	lines := strings.Split(c.listEntries[idx].detail.String(), "\n")[start: end]
	return "[" + c.listEntries[idx].command.String() + "](fg:blue,mod:bold)" + "\n" +
		"[" + c.listEntries[idx].explain.String() + "](fg:blue,mod:bold)" + "\n" +
		strings.Join(lines, "\n")
}

func (c *Content)reset() {
	c.contentStartIdx = 0
	c.contentEndIdx = 0
	c.widgetType = ListType
}


// 第一个bool表示要不要下一个组件渲染，第二个bool表示是不是调换了输入框
func (c *Content) HandleEvent(e ui.Event) (ui.Drawable, bool, bool) {
	c.setColor()
	if e.Type == ui.KeyboardEvent {
		switch e.ID {
		case "<C-c>":
			ui.Close()
			os.Exit(-1)
		default:
			if handler, ok := c.handlerMap[e.ID]; ok {
				return handler(c)
			} else {
				c.reset()
				return c.listWidget, true, false
			}
		}
	}
	return c.listWidget, false, false
}