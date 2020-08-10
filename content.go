package main

import (
	"bytes"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"os"
	"sort"
	"strings"
)


type Entry struct {
	command bytes.Buffer
	explain bytes.Buffer
	detail bytes.Buffer
}

type Content struct {
	// 展示列表
	listWidget *widgets.List
	// 展示详情
	contentWidget *widgets.Paragraph
	listEntries []*Entry
	entries []*Entry
	// 用于匹配命令
	inputTextChan *chan string
	// 记录命令
	recHistoryCommandChan *chan string

	finder MultiFind
}

func NewContent(inputTextChan *chan string,  recHistoryCommandChan *chan string) *Content {
	listWidget := widgets.NewList()
	listWidget.Title = "Content"
	listWidget.Rows = make([]string, 0)
	listWidget.TextStyle = ui.NewStyle(ui.ColorYellow)
	listWidget.BorderStyle.Fg = ui.ColorRed
	listWidget.WrapText = false
	listWidget.SetRect(0, 5, 170, 40)

	contentWidget := widgets.NewParagraph()
	contentWidget.Title = "Content"
	contentWidget.Text = ""
	contentWidget.SetRect(0, 5, 200, 40)
	contentWidget.BorderStyle.Fg = ui.ColorRed

	content := &Content{
		listWidget:    listWidget,
		contentWidget: contentWidget,
		recHistoryCommandChan: recHistoryCommandChan,
		inputTextChan: inputTextChan,
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
	homedir, err := os.UserHomeDir()
	if err != nil {
		homedir = "/usr/local/bin/"
	}
	files := Files(homedir + "/.files")

	for _, file := range files {
		b, err := ReadContent(file)
		if err != nil {
			return
		}
		content := string(b)
		lines := strings.Split(content, "\n")
		newLines := make([]string, 0)
		for _, line := range lines {
			if len(line) != 0 {
				newLines = append(newLines, line)
			}
		}
		lines = newLines

		var entry *Entry

		index := -1
		for i, line := range lines {
			line = strings.Trim(line, "\t")
			line = strings.Trim(line, " ")
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
			if index == 0 {
				if len(line) > 60 {
					entry.command.WriteString(line[0:40])
				} else {
					entry.command.Reset()
					entry.command.WriteString(line)
					for entry.command.Len() < 60 {
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
				c.entries = append(c.entries, entry)
			}
		}
	}
	sort.Sort((EntrySlice)(c.entries))
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
	//if i + 1 < 10 {
	//	return fmt.Sprintf("[%d.    ](fg:blue,mod:bold) [%s](fg:blue,mod:bold) [%s](fg:red,mod:light)", i+1, entry.command.String(), entry.explain.String())
	//} else if i + 1 < 100 {
	//	return fmt.Sprintf("[%d.   ](fg:blue,mod:bold) [%s](fg:blue,mod:bold) [%s](fg:red,mod:light)", i+1, entry.command.String(), entry.explain.String())
	//} else {
	//	return fmt.Sprintf("[%d.  ](fg:blue,mod:bold) [%s](fg:blue,mod:bold) [%s](fg:red,mod:light)", i+1, entry.command.String(), entry.explain.String())
	//}
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

// 第一个bool表示要不要下一个组件渲染，第二个bool表示是不是调换了输入框
func (c *Content) HandleEvent(e ui.Event) (ui.Drawable, bool, bool) {
	c.setColor()
	if e.Type == ui.KeyboardEvent {
		switch e.ID {
		case "<C-c>":
			os.Exit(-1)
		case "<Enter>":
			if len(c.listWidget.Rows) > 0 {
				c.contentWidget.Text = c.listEntries[c.listWidget.SelectedRow].detail.String()
				// 记录
				*c.recHistoryCommandChan <- c.listEntries[c.listWidget.SelectedRow].command.String()
				return c.contentWidget, false, false
			}
		case "<Down>":
			if len(c.listWidget.Rows) > 0 {
				c.listWidget.ScrollDown()
			}
		case "<Up>":
			if len(c.listWidget.Rows) > 0 {
				c.listWidget.ScrollUp()
			}
		case "<Left>":
		case "<Right>":
			cursor = 1
			c.setColor()
			return c.listWidget, true, true
		case "<C-d>":
			if len(c.listWidget.Rows) > 0 {
				c.listWidget.ScrollHalfPageDown()
			}
		case "<C-u>":
			if len(c.listWidget.Rows) > 0 {
				c.listWidget.ScrollHalfPageUp()
			}
		case "<C-f>":
			if len(c.listWidget.Rows) > 0 {
				c.listWidget.ScrollPageDown()
			}
		case "<C-b>":
			if len(c.listWidget.Rows) > 0 {
				c.listWidget.ScrollPageUp()
			}
		case "<Home>":
			if len(c.listWidget.Rows) > 0 {
				c.listWidget.ScrollTop()
			}
		case "<End>":
			if len(c.listWidget.Rows) > 0 {
				c.listWidget.ScrollBottom()
			}
		default:
			return c.listWidget, true, false
		}
	}
	return c.listWidget, false, false
}