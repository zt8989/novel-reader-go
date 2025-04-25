package main

import (
	"flag"
	"fmt"
	"net/url"
	"path"
	"strings"
	"unicode/utf8"

	"novel-reader-go/parser"

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
	httpClient = parser.NewDefaultHttpClient()
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// 定义模型
type model struct {
	url          string
	lines        int
	novelContent *parser.NovelResult
	content      []string
	cursor       int
	done         bool
	loading      bool
}

// fetchContent 获取网页内容
func fetchContent(url string) (*parser.NovelResult, error) {
	// 使用通用解析器
	p := parser.NewGeneralParser(httpClient)

	// 获取并解析内容
	novelContent, err := p.ParseNovel(url)
	if err != nil {
		return nil, err
	}

	// 返回完整内容
	return &novelContent, nil
}

// 初始命令
func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		m.loading = true
		novelContent, err := fetchContent(m.url)
		if err != nil {
			m.loading = false
			return errMsg(err)
		}
		m.loading = false
		return contentMsg(novelContent)
	}
}

// 自定义消息类型
type contentMsg *parser.NovelResult
type errMsg error

// 更新函数
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case contentMsg:
		m.novelContent = msg
		m.content = wordWrap(m.novelContent.Content, 40)
		m.loading = false
		return m, nil
	case errMsg:
		m.content = []string{fmt.Sprintf("错误: %v", msg)}
		m.loading = false
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.done = true
			return m, tea.Quit
		case "j", "down":
			if m.cursor < len(m.content)-m.lines {
				m.cursor += m.lines
				if m.cursor > len(m.content)-m.lines {
					m.cursor = len(m.content) - m.lines
				}
			} else if m.novelContent != nil && m.novelContent.Index.Next != "" {
				nextURL := m.novelContent.Index.Next
				if !strings.HasPrefix(nextURL, "http") {
					currentURL, err := url.Parse(m.url)
					if err == nil {
						baseURL := fmt.Sprintf("%s://%s", currentURL.Scheme, currentURL.Host)
						if strings.HasPrefix(nextURL, "/") {
							nextURL = baseURL + nextURL
						} else {
							dir := path.Dir(currentURL.Path)
							nextURL = baseURL + path.Join(dir, nextURL)
						}
					}
				}
				m.url = nextURL
				m.cursor = 0
				m.loading = true
				return m, func() tea.Msg {
					novelContent, err := fetchContent(nextURL)
					if err != nil {
						m.loading = false
						return errMsg(err)
					}
					m.loading = false
					return contentMsg(novelContent)
				}
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor -= m.lines
				if m.cursor < 0 {
					m.cursor = 0
				}
			} else if m.novelContent != nil && m.novelContent.Index.Prev != "" {
				prevURL := m.novelContent.Index.Prev
				if !strings.HasPrefix(prevURL, "http") {
					currentURL, err := url.Parse(m.url)
					if err == nil {
						baseURL := fmt.Sprintf("%s://%s", currentURL.Scheme, currentURL.Host)
						if strings.HasPrefix(prevURL, "/") {
							prevURL = baseURL + prevURL
						} else {
							dir := path.Dir(currentURL.Path)
							prevURL = baseURL + path.Join(dir, prevURL)
						}
					}
				}
				m.url = prevURL
				m.cursor = 0
				m.loading = true
				return m, func() tea.Msg {
					novelContent, err := fetchContent(prevURL)
					if err != nil {
						m.loading = false
						return errMsg(err)
					}
					m.loading = false
					return contentMsg(novelContent)
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
	return m, nil
}

// 视图函数
func (m model) View() string {
	if m.done {
		return "再见！\n"
	}

	output := ""

	if m.loading || len(m.content) == 0 {
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
		title := m.url
		if m.novelContent != nil {
			title = m.novelContent.Title
		}
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
		url:   url,
		lines: lines,
	}

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("出错了: %v", err)
	}
}
