package parser

import (
	"os"
	"testing"
)

func TestParseContentWithRealData(t *testing.T) {
	// 读取HTML测试文件
	htmlContent, err := os.ReadFile("testdata/508701.html")
	if err != nil {
		t.Fatal(err)
	}

	// 读取预期结果文件
	txtContent, err := os.ReadFile("testdata/508701.txt")
	if err != nil {
		t.Fatal(err)
	}
	expected := string(txtContent)

	// 获取解析结果
	p := GeneralParser{}
	actual, err := p.parseContent(string(htmlContent))

	// 比较结果
	if actual != expected {
		t.Errorf("解析结果与预期不符\n期望: %s\n实际: %s", expected, actual)
	}
}
