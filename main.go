package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"unicode/utf8"

	"novel-reader-go/parser"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// wordWrap 将文本按指定宽度分割成数组
func wordWrap(str string, maxWidth int) []string {
	lines := strings.Split(str, "\n")
	var newLines []string
	for _, line := range lines {
		if utf8.RuneCountInString(line) <= maxWidth {
			if strings.TrimSpace(line) != "" {
				newLines = append(newLines, line)
			}
		} else {
			runes := []rune(line)
			for len(runes) > 0 {
				newLines = append(newLines, string(runes[:min(maxWidth, len(runes))]))
				runes = runes[min(maxWidth, len(runes)):]
			}
		}
	}
	return newLines
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var (
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	reader    = parser.NewReaderWithoutUrl()
)

// 定义模型
type model struct {
	state     string
	textInput textinput.Model
	originUrl string
	lines     int
	content   []string
	cursor    int
	done      bool

	history historyEntry
}

// 初始命令
func (m *model) fetchNovelContent(direction string) tea.Msg {
	var (
		novelContent *parser.NovelResult
		err          error
	)

	switch direction {
	case "up":
		novelContent, err = reader.ReadPrev()
	case "down":
		novelContent, err = reader.ReadNext()
	default:
		novelContent, err = reader.Read()
	}

	saveHistory(m.originUrl, reader.GetUrl(), 0)

	if err != nil {
		return errMsg(err)
	}
	return contentMsg(novelContent)
}

func (m model) Init() tea.Cmd {
	if m.state == "prompt" {
		return textinput.Blink
	}

	return func() tea.Msg {
		return m.fetchNovelContent("")
	}
}

// 自定义消息类型
type contentMsg *parser.NovelResult
type errMsg error

// 更新函数
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.state {
	case "prompt":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "ctrl+c":
				m.done = true
				return m, tea.Quit
			case "enter":
				m.state = "reading"
				if m.textInput.Value() == "Y" || m.textInput.Value() == "y" || m.textInput.Value() == "" {
					entry := m.history
					m.originUrl = entry.OriginURL
					m.cursor = entry.Cursor
					return m, func() tea.Msg {
						reader.SetUrl(entry.LastURL)
						return m.fetchNovelContent("")
					}
				} else {
					return m, func() tea.Msg {
						reader.SetUrl(m.originUrl)
						return m.fetchNovelContent("")
					}
				}
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	case "reading":
		switch msg := msg.(type) {
		case contentMsg:
			m.content = wordWrap(msg.Content, 40)
			return m, nil
		case errMsg:
			m.content = []string{fmt.Sprintf("错误: %v", msg)}
			return m, nil
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "ctrl+c":
				m.done = true
				saveHistory(m.originUrl, reader.GetUrl(), m.cursor)
				return m, tea.Quit
			case "j", "down":
				if m.cursor < len(m.content)-m.lines {
					m.cursor += m.lines
					if m.cursor > len(m.content)-m.lines {
						m.cursor = len(m.content) - m.lines
					}
				} else if reader.HasNext() {
					return m, func() tea.Msg {
						return m.fetchNovelContent("down")
					}
				}
			case "k", "up":
				if m.cursor > 0 {
					m.cursor -= m.lines
					if m.cursor < 0 {
						m.cursor = 0
					}
				} else if reader.HasPrev() {
					return m, func() tea.Msg {
						return m.fetchNovelContent("down")
					}
				}
			case "ctrl+f", "pagedown":
				m.cursor += m.lines
				if m.cursor > len(m.content)-m.lines {
					m.cursor = len(m.content) - m.lines
				}
				if m.cursor < 0 {
					m.cursor = 0
				}
			case "ctrl+b", "pageup":
				m.cursor -= m.lines
				if m.cursor < 0 {
					m.cursor = 0
				}
			case "g":
				m.cursor = 0
			case "G":
				m.cursor = len(m.content) - m.lines
				if m.cursor < 0 {
					m.cursor = 0
				}
			}
		}
	}

	return m, nil
}

// 视图函数
type historyEntry struct {
	OriginURL string `json:"originUrl"`
	LastURL   string `json:"lastUrl"`
	Cursor    int    `json:"cursor"`
}

func saveHistory(originUrl, lastUrl string, cursor int) {
	historyDir := path.Join(os.Getenv("HOME"), ".nvrd")
	historyFile := path.Join(historyDir, "history.json")

	// Create directory if not exists
	if _, err := os.Stat(historyDir); os.IsNotExist(err) {
		os.MkdirAll(historyDir, 0755)
	}

	// Read existing history
	var history historyEntry
	if data, err := os.ReadFile(historyFile); err == nil {
		json.Unmarshal(data, &history)
	}

	// Update entry
	history = historyEntry{
		OriginURL: originUrl,
		LastURL:   lastUrl,
		Cursor:    cursor,
	}

	// Write back to file
	if data, err := json.Marshal(history); err == nil {
		os.WriteFile(historyFile, data, 0644)
	}
}

func (m model) View() string {
	if m.done {
		return "再见！\n"
	}

	switch m.state {
	case "prompt":
		return fmt.Sprintf(
			"是否继续上次的阅读？\n%s\n",
			m.textInput.View(),
		) + "\n"
	}

	output := ""

	if reader.GetLoading() || len(m.content) == 0 {
		for i := 0; i < m.lines; i++ {
			output += "\n"
		}
		output += helpStyle.Render(fmt.Sprintf("%.2f%%\t%d/%d\t%s", 100.0, 0, 0, "内容加载中..."))
	} else {
		end := m.cursor + m.lines
		if end > len(m.content) {
			end = len(m.content)
		}
		for i := m.cursor; i < end; i++ {
			output += fmt.Sprintf("%s\n", m.content[i])
		}
		progress := float64(m.cursor+1) / float64(len(m.content)) * 100
		title := reader.GetTitle()
		output += helpStyle.Render(fmt.Sprintf("%.2f%%\t%d/%d\t%s", progress, m.cursor+1, len(m.content), title))
	}

	return output
}

func main() {
	// 解析命令行参数
	var (
		url   string
		lines int
	)
	flag.StringVar(&url, "read", "", "章节地址")
	flag.IntVar(&lines, "n", 1, "显示的行数")
	flag.Parse()

	if url == "" {
		fmt.Println("使用方法: novel-reader-go -read <章节地址> [-n 行数]")
		return
	}

	// 创建初始模型
	initialModel := model{
		originUrl: url,
		lines:     lines,
		state:     "reading",
	}

	// 检查历史记录
	historyDir := path.Join(os.Getenv("HOME"), ".nvrd")
	historyFile := path.Join(historyDir, "history.json")

	if _, err := os.Stat(historyFile); err == nil {
		data, err := os.ReadFile(historyFile)
		if err == nil {
			var entry historyEntry
			if json.Unmarshal(data, &entry) == nil {
				// 创建textinput模型
				ti := textinput.New()
				ti.Placeholder = "Y/n"
				ti.Focus()
				ti.CharLimit = 156
				ti.Width = 20
				initialModel.textInput = ti
				initialModel.state = "prompt"
				initialModel.history = entry
			}
		}
	}

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("出错了: %v", err)
	}
}
