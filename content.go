package main

import (
	"bytes"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

var maxLineNum = 80

var maxContentLineSize = 1
var displayLineSize = 30

type Entry struct {
	command bytes.Buffer
	explain bytes.Buffer
	detail bytes.Buffer
}

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

	widgetType WidgetType

	listEntries []*Entry
	entries []*Entry
	// 用于匹配命令
	inputTextChan *chan string
	// 记录命令
	recHistoryCommandChan *chan string
	// 搜索接口
	finder MultiFind
}

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
	}

	content.loadContent()
	go content.handleInputText()
	return content
}

type EntrySlice []*Entry
func (e EntrySlice) Len() int           { return len(e) }
func (e EntrySlice) Less(i, j int) bool {  return strings.Compare(e[i].command.String(), e[j].command.String()) <= 0 }
func (e EntrySlice) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }



// 加载并解析文件
func (c *Content) loadContent() {

	var files []string
	c.entries = make([]*Entry, 0)
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
		entries := make([]*Entry, 0)
		b, err := ReadContent(file)
		if err != nil {
			return
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
					c.entries = append(c.entries, entry)
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
				entries = append(entries, entry)
			}
		}
		sort.Sort((EntrySlice)(entries))
		c.entries = append(c.entries, entries...)
	}
	if len(c.entries) == 0 {
		panic("读取空的内容")
	}
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
			os.Exit(-1)
		case "<C-r>":
			c.reset()
			c.loadContent()
			return c.listWidget, false, false
		case "<Enter>":
			if len(c.listWidget.Rows) > 0 {
				if c.widgetType == ListType {
					c.contentInitPage()
					c.widgetType = ContentType
					// 记录
					*c.recHistoryCommandChan <- c.listEntries[c.listWidget.SelectedRow].command.String()
				}
			}
			return c.contentWidget, false, false
		case "<Down>":
			if len(c.listWidget.Rows) > 0 {
				if c.widgetType == ListType {
					c.listWidget.ScrollDown()
					return c.listWidget, false, false
				} else {
					c.contentDownPage()
					return c.contentWidget, false, false
				}
			}
		case "<Up>":
			if len(c.listWidget.Rows) > 0 {
				if c.widgetType == ListType {
					c.listWidget.ScrollUp()
					return c.listWidget, false, false
				} else {
					c.contentUpPage()
					return c.contentWidget, false, false
				}
			}
		case "<Left>":
			c.setColor()
			if c.widgetType == ListType {
				return c.listWidget, false, false
			} else {
				return c.contentWidget, false, false
			}
		case "<Right>":
			cursor = 1
			c.setColor()
			if c.widgetType == ListType {
				return c.listWidget, true, true
			} else {
				return c.contentWidget, true, true
			}
		default:
			c.reset()
			return c.listWidget, true, false
		}
	}
	return c.listWidget, false, false
}