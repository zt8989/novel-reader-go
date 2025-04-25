package main

import (
	"flag"
	"fmt"
	"strings"
	"unicode/utf8"

	"novel-reader-go/parser"

	tea "github.com/charmbracelet/bubbletea"
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

// 定义模型
type model struct {
	url     string
	lines   int
	content []string
	cursor  int
	done    bool
}

// fetchContent 获取网页内容
func fetchContent(url string) (string, error) {
	// 使用通用解析器
	httpClient := parser.NewDefaultHttpClient()
	p := parser.NewGeneralParser(httpClient)

	// 获取并解析内容
	novelContent, err := p.ParseNovel(url)
	if err != nil {
		return "", err
	}

	// 返回完整内容
	return novelContent.Content, nil
}

// 初始命令
func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		content, err := fetchContent(m.url)
		if err != nil {
			return errMsg(err)
		}
		return contentMsg(content)
	}
}

// 自定义消息类型
type contentMsg string
type errMsg error

// 更新函数
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case contentMsg:
		// 直接使用wordWrap处理完整内容
		m.content = wordWrap(string(msg), 40)
		return m, nil
	case errMsg:
		m.content = []string{fmt.Sprintf("错误: %v", msg)}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.done = true
			return m, tea.Quit
		case "j", "down":
			if m.cursor < len(m.content)-m.lines {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
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

	if len(m.content) == 0 {
		output += "正在加载内容...\n"
	} else {
		end := m.cursor + m.lines
		if end > len(m.content) {
			end = len(m.content)
		}
		for i := m.cursor; i < end; i++ {
			output += fmt.Sprintf("%s\n", m.content[i])
		}
		output += fmt.Sprintf("\n--- 第 %d-%d 行，共 %d 行 ---\n", m.cursor+1, end, len(m.content))
	}

	return output
}

func main() {
	// 解析命令行参数
	lines := flag.Int("n", 1, "显示的行数")
	flag.Parse()

	// 获取URL参数
	args := flag.Args()
	if len(args) < 2 || args[0] != "read" {
		fmt.Println("使用方法: novel-reader-go read <章节地址> [-n 行数]")
		return
	}
	url := args[1]

	// 创建初始模型
	initialModel := model{
		url:   url,
		lines: *lines,
	}

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("出错了: %v", err)
	}
}
