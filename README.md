# novel-reader-go

一个命令行小说阅读器，支持自动解析小说网站内容并提供舒适的阅读体验。

## 功能特点

- 自动解析小说网站内容
- 支持章节导航（上一章/下一章）
- 简洁的命令行界面
- 可自定义显示行数
- 实时显示阅读进度

## 安装

1. 确保已安装Go (1.16+) 
2. 克隆仓库并安装依赖：

```bash
git clone https://github.com/zt8989/novel-reader-go.git
cd novel-reader-go
go mod download
```

3. 编译并安装：

```bash
# 通用编译
GOOS=linux GOARCH=amd64 go build -o novel-reader-linux
GOOS=darwin GOARCH=amd64 go build -o novel-reader-mac
GOOS=windows GOARCH=amd64 go build -o novel-reader.exe

# 编译为ARM64架构
# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o novel-reader-arm64

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o novel-reader-arm64

# 交叉编译示例
# 编译Windows版本
GOOS=windows GOARCH=amd64 go build -o novel-reader.exe

# 编译为ARM64架构
# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o novel-reader-arm64

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o novel-reader-arm64

# 编译Mac版本
GOOS=darwin GOARCH=amd64 go build -o novel-reader-mac

# 编译Linux版本
GOOS=linux GOARCH=amd64 go build -o novel-reader-linux

# 安装到系统路径(需要管理员权限)
sudo cp novel-reader /usr/local/bin/
```

注意事项：
- 交叉编译需要设置GOOS和GOARCH环境变量
- Windows可执行文件需要添加.exe后缀
- 不同架构需要调整GOARCH参数(如arm64)
- 编译前请确保已安装对应平台的工具链

## 使用说明

基本用法：

```bash
./novel-reader -read <章节地址> [-n 行数]
```

示例：

```bash
./novel-reader -read https://www.example.com/chapter1 -n 10
```

## 快捷键

| 快捷键 | 功能 |
|--------|------|
| j/↓ | 向下滚动 |
| k/↑ | 向上滚动 |
| Ctrl+f/PageDown | 向下翻页 |
| Ctrl+b/PageUp | 向上翻页 |
| g | 跳转到开头 |
| G | 跳转到末尾 |
| q/Ctrl+c | 退出程序 |

## 开发

```bash
# 运行测试
make test

# 构建
make build
```

## 许可证

MIT License