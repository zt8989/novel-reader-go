package parser

import (
	"math"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

// IParser 接口定义
type IParser interface {
	ParseNovel(url string) (NovelResult, error)
}

// NovelResult 包含解析结果
type NovelResult struct {
	Index   IndexResult
	Content string
	Title   string
}

// IndexResult 包含上一章和下一章链接
type IndexResult struct {
	Next string
	Prev string
}

// GeneralParser 实现 IParser 接口
type GeneralParser struct {
	client HttpClient
}

// NewGeneralParser 创建新的 GeneralParser 实例
func NewGeneralParser(client HttpClient) *GeneralParser {
	return &GeneralParser{client: client}
}

// fetchUrl 获取 URL 内容并处理编码
func (p *GeneralParser) fetchUrl(url string) (string, error) {
	body, err := p.client.FetchUrl(url)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// handleTextNodes 处理文本节点，类似于原 TypeScript 版本的功能
// handleTextNodes 处理文本节点，与 TypeScript 版本功能一致
// handleTextNodes 处理文本节点，与 TypeScript 版本功能一致
func (p *GeneralParser) handleTextNodes(selection *goquery.Selection) string {
	var lines []string
	const indent = "    " // 4个空格的缩进

	selection.Contents().Each(func(i int, s *goquery.Selection) {
		// 处理文本节点
		if goquery.NodeName(s) == "#text" {
			content := strings.TrimSpace(s.Text())
			if content != "" {
				lines = append(lines, indent+content)
			}
		} else if goquery.NodeName(s) == "p" { // 注意：else if 必须与上一个 } 在同一行
			content := strings.TrimSpace(s.Text())
			if content != "" {
				lines = append(lines, indent+content)
			}
		}
	})

	return strings.Join(lines, "\n")
}

// parseContent 解析内容
func (p *GeneralParser) parseContent(doc string) (string, error) {
	// 加载 HTML 文档
	reader := strings.NewReader(doc)
	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return "", err
	}

	// 递归解析子节点
	var parseChildren func(parent *goquery.Selection, size int, slope float64, variance float64) string
	parseChildren = func(parent *goquery.Selection, size int, slope float64, variance float64) string {
		children := parent.Children()
		if children.Length() == 1 {
			return parseChildren(children.First(), size, slope, variance)
		}

		var maxChildren struct {
			element *goquery.Selection
			size    int
			slope   float64
		}

		sizeList := make([]int, 0)
		children.Each(func(i int, s *goquery.Selection) {
			length := utf8.RuneCountInString(s.Text())
			sizeList = append(sizeList, length)
			tempSlope := float64(size-length) / float64(size)

			if maxChildren.element == nil {
				maxChildren.element = s
				maxChildren.size = length
				maxChildren.slope = tempSlope
			} else if tempSlope < maxChildren.slope {
				maxChildren.element = s
				maxChildren.size = length
				maxChildren.slope = tempSlope
			}
		})

		// 计算平均值和方差
		sum := 0
		for _, x := range sizeList {
			sum += x
		}
		avg := float64(sum) / float64(len(sizeList))

		tempVariance := 0.0
		for _, x := range sizeList {
			tempVariance += math.Pow(float64(x)-avg, 2)
		}
		tempVariance = math.Sqrt(tempVariance / float64(len(sizeList)))

		if maxChildren.element != nil {
			// log.Printf("Debug: %s, slope: %f, tempVariance: %f, variance: %f\n",
			// 	maxChildren.element.Text(), maxChildren.slope, tempVariance, variance)

			if tempVariance > 100 {
				return parseChildren(maxChildren.element, maxChildren.size, maxChildren.slope, tempVariance)
			} else {
				return p.handleTextNodes(parent)
			}
		}
		return ""
	}

	body := document.Find("body")
	body.Find("style, script").Remove()
	bodyText := body.Text()
	return parseChildren(body, utf8.RuneCountInString(bodyText), 1.0, float64(utf8.RuneCountInString(bodyText))), nil
}

// parseIndexChapter 解析章节索引
func (p *GeneralParser) parseIndexChapter(doc string) (IndexResult, error) {
	reader := strings.NewReader(doc)
	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return IndexResult{}, err
	}

	nextHref, err := p.parseNextPage(document)
	if err != nil {
		return IndexResult{}, err
	}

	prevHref, err := p.parsePrevPage(document)
	if err != nil {
		return IndexResult{}, err
	}

	return IndexResult{Next: nextHref, Prev: prevHref}, nil
}

// parsePrevPage 解析上一页链接
func (p *GeneralParser) parsePrevPage(document *goquery.Document) (string, error) {
	prevHref := ""
	document.Find("body").Find("a").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text == "上一章" || text == "上一页" {
			href, exists := s.Attr("href")
			if exists {
				prevHref = href
			}
		}
	})
	return prevHref, nil
}

// parseNextPage 解析下一页链接
func (p *GeneralParser) parseNextPage(document *goquery.Document) (string, error) {
	nextHref := ""
	document.Find("body").Find("a").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text == "下一章" || text == "下一页" {
			href, exists := s.Attr("href")
			if exists {
				nextHref = href
			}
		}
	})
	return nextHref, nil
}

// parseTitle 解析标题
func (p *GeneralParser) parseTitle(doc string) (string, error) {
	reader := strings.NewReader(doc)
	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return "", err
	}
	return document.Find("title").Text(), nil
}

// ParseNovel 解析小说内容
func (p *GeneralParser) ParseNovel(url string) (NovelResult, error) {
	doc, err := p.fetchUrl(url)
	if err != nil {
		return NovelResult{}, err
	}

	content, err := p.parseContent(doc)
	if err != nil {
		return NovelResult{}, err
	}

	index, err := p.parseIndexChapter(doc)
	if err != nil {
		return NovelResult{}, err
	}

	title, err := p.parseTitle(doc)
	if err != nil {
		return NovelResult{}, err
	}

	return NovelResult{
		Index:   index,
		Content: content,
		Title:   title,
	}, nil
}
