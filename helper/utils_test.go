package helper

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// 上游返回 HTML 时，错误里要带上原文开头 —— 光有 invalid character '<'
// 分不出是风控页还是网关页，处置完全不同。
func TestParseJSONMapKeepsBodySnippet(t *testing.T) {
	_, err := ParseJSONMap([]byte("<!DOCTYPE html><html><head><title>百度安全验证</title></head></html>"))
	if err == nil {
		t.Fatal("HTML 应当解析失败")
	}

	if !strings.Contains(err.Error(), "百度安全验证") {
		t.Errorf("错误里应当带上响应体原文，got = %q", err)
	}
}

// 超长响应体要截断：一整页 HTML 进日志会把上下文冲掉。
func TestParseJSONMapSnippetTruncated(t *testing.T) {
	body := strings.Repeat("啊", 500)

	_, err := ParseJSONMap([]byte(body))
	if err == nil {
		t.Fatal("非 JSON 应当解析失败")
	}

	got := err.Error()
	// 原文那一段截到 200 个 rune、以省略号收尾，整个错误消息因此不会失控
	if !strings.HasSuffix(strings.TrimSuffix(got, ")"), "…") {
		t.Errorf("超长响应体应当被截断并以省略号结尾，got = %q", got)
	}
	if n := utf8.RuneCountInString(got); n > 400 {
		t.Errorf("错误消息不该这么长（%d 个 rune），截断没生效？", n)
	}
	// 截断按 rune 走，不能把多字节字符切成半个
	if !utf8.ValidString(got) {
		t.Errorf("截断把多字节字符切坏了，got = %q", got)
	}
}

// 换行与连续空白先折叠掉，否则日志里会出现一段带缩进的 HTML。
func TestParseJSONMapSnippetCollapsesWhitespace(t *testing.T) {
	_, err := ParseJSONMap([]byte("<html>\n\t  <body>  \n\n  <p>验证</p>\n</html>"))
	if err == nil {
		t.Fatal("HTML 应当解析失败")
	}

	if got := err.Error(); strings.Contains(got, "\n") || strings.Contains(got, "\t") {
		t.Errorf("原文里的空白应当折叠成单个空格，got = %q", got)
	}
}

// 响应体为空（或只有空白）时不加一个空括号。
func TestParseJSONMapSnippetEmpty(t *testing.T) {
	for _, body := range []string{"", "   \n\t "} {
		_, err := ParseJSONMap([]byte(body))
		if err == nil {
			t.Fatalf("空响应体应当解析失败，body = %q", body)
		}
		if strings.Contains(err.Error(), "原文") {
			t.Errorf("响应体为空时不该附原文，body = %q, got = %q", body, err)
		}
	}
}

// 正常的 JSON 不受影响。
func TestParseJSONMapStillWorks(t *testing.T) {
	m, err := ParseJSONMap([]byte(`{"error_code":0,"user_name":"ry","user_id":123456789012345}`))
	if err != nil {
		t.Fatalf("正常 JSON 不该报错: %v", err)
	}
	if got := JSONStr(m, "user_name"); got != "ry" {
		t.Errorf("user_name = %q, 期望 ry", got)
	}
	// 大整数走 json.Number，精度不能丢
	if got := JSONInt(m, "user_id"); got != 123456789012345 {
		t.Errorf("user_id = %d, 期望 123456789012345", got)
	}
}
