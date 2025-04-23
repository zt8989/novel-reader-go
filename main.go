package main

import (
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
)

// 定义模型
type model struct {
    count int
    done  bool
}

// 初始命令
func (m model) Init() tea.Cmd {
    return nil
}

// 更新函数
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            m.done = true
            return m, tea.Quit
        case "up":
            m.count++
        case "down":
            m.count--
        }
    }
    return m, nil
}

// 视图函数
func (m model) View() string {
    if m.done {
        return "再见！\n"
    }
    return fmt.Sprintf(
        "计数器: %d\n\n"+
            "按 ↑ 增加\n"+
            "按 ↓ 减少\n"+
            "按 q 退出\n",
        m.count)
}

func main() {
    p := tea.NewProgram(model{})
    if _, err := p.Run(); err != nil {
        fmt.Printf("出错了: %v", err)
    }
}