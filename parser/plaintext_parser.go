package parser

import (
	"bytes"
	"io"
	"os"
	"path"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type PlainTextParser struct {
	url     string
	context string
}

func NewPlainTextParser(url string) *PlainTextParser {
	return &PlainTextParser{
		url:     url,
		context: "",
	}
}

func (p *PlainTextParser) ParseNovel(url string) (NovelResult, error) {
	// 如果提供了新的URL，则更新parser的url
	if url != "" {
		p.url = url
	}

	// 打开并读取文件
	file, err := os.Open(p.url)
	if err != nil {
		return NovelResult{}, err
	}
	defer file.Close()

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		return NovelResult{}, err
	}

	// 检查是否为有效的UTF-8
	contentStr := string(content)
	if utf8.ValidString(contentStr) {
		// 如果是有效的UTF-8，直接使用
		p.context = contentStr
	} else {
		// 如果不是有效的UTF-8，尝试不同的编码
		var decodedContent []byte
		var decodeErr error

		// 尝试常见的编码
		encodings := []encoding.Encoding{
			simplifiedchinese.GBK,
			simplifiedchinese.GB18030,
			traditionalchinese.Big5,
			japanese.EUCJP,
			japanese.ShiftJIS,
			korean.EUCKR,
			unicode.UTF16(unicode.LittleEndian, unicode.UseBOM),
			unicode.UTF16(unicode.BigEndian, unicode.UseBOM),
		}

		for _, enc := range encodings {
			reader := transform.NewReader(bytes.NewReader(content), enc.NewDecoder())
			decodedContent, decodeErr = io.ReadAll(reader)
			if decodeErr == nil {
				// 检查转换后的内容是否为有效的UTF-8
				decodedStr := string(decodedContent)
				if utf8.ValidString(decodedStr) {
					p.context = decodedStr
					break
				}
			}
		}
	}

	// 构造并返回NovelResult
	result := NovelResult{
		Content: p.context,
		Title:   path.Base(p.url), // 使用文件名作为标题
		Index: IndexResult{
			Next: "", // 纯文本文件没有上一章/下一章概念
			Prev: "",
		},
	}

	return result, nil
}
