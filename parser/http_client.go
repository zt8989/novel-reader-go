package parser

import (
	"bytes"
	"io"
	"net/http"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// HttpClient 定义HTTP客户端接口
type HttpClient interface {
	FetchUrl(url string) ([]byte, error)
}

// DefaultHttpClient 默认HTTP客户端实现
type DefaultHttpClient struct{}

// NewDefaultHttpClient 创建默认HTTP客户端
func NewDefaultHttpClient() *DefaultHttpClient {
	return &DefaultHttpClient{}
}

// FetchUrl 实现HttpClient接口
func (p *DefaultHttpClient) FetchUrl(url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 检查是否是 GBK/GB2312 编码
	if bytes.Contains(body, []byte("gbk")) || bytes.Contains(body, []byte("gb2312")) {
		// 尝试自动检测编码
		reader, err := charset.NewReader(bytes.NewReader(body), resp.Header.Get("Content-Type"))
		if err != nil {
			// 如果自动检测失败，尝试使用 GBK 解码
			decoder := simplifiedchinese.GBK.NewDecoder()
			decoded, err := decoder.Bytes(body)
			if err != nil {
				return nil, err
			}
			return decoded, nil
		}
		decoded, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		return decoded, nil
	}

	// 默认使用 UTF-8
	return body, nil
}
