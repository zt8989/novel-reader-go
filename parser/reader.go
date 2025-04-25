package parser

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

type Reader struct {
	url     string
	parser  IParser
	content *NovelResult
	loading bool
}

func NewReaderWithParser(parser IParser) *Reader {
	return &Reader{
		parser:  parser,
		loading: false,
	}
}

func NewReaderWithoutUrl() *Reader {
	return &Reader{
		parser: NewGeneralParser(&DefaultHttpClient{}),
	}
}

func (r *Reader) Read() (*NovelResult, error) {
	r.loading = true
	result, err := r.parser.ParseNovel(r.url)
	r.loading = false
	if err == nil {
		r.content = &result
	}
	return &result, err
}

func (r *Reader) ReadNext() (*NovelResult, error) {
	if r.content.Index.Next != "" {
		r.url = r.handlePageNavigation(r.content.Index.Next)
		return r.Read()
	}
	return &NovelResult{}, nil
}

func (r *Reader) ReadPrev() (*NovelResult, error) {
	if r.content.Index.Prev != "" {
		r.url = r.handlePageNavigation(r.content.Index.Prev)
		return r.Read()
	}
	return &NovelResult{}, nil
}

func (r *Reader) GetUrl() string {
	return r.url
}

func (r *Reader) SetUrl(url string) {
	r.url = url
}

func (r *Reader) handlePageNavigation(navURL string) string {
	if !strings.HasPrefix(navURL, "http") {
		currentURL, err := url.Parse(r.url)
		if err == nil {
			baseURL := fmt.Sprintf("%s://%s", currentURL.Scheme, currentURL.Host)
			if strings.HasPrefix(navURL, "/") {
				navURL = baseURL + navURL
			} else {
				dir := path.Dir(currentURL.Path)
				navURL = baseURL + path.Join(dir, navURL)
			}
		}
	}
	return navURL
}

func (r *Reader) HasNext() bool {
	if r.content != nil && r.content.Index.Prev != "" {
		return true
	}
	return false
}

func (r *Reader) HasPrev() bool {
	if r.content != nil && r.content.Index.Prev != "" {
		return true
	}
	return false
}

func (r *Reader) GetTitle() string {
	if r.content != nil {
		return r.content.Title
	}
	return r.url
}

func (r *Reader) GetLoading() bool {
	return r.loading
}
