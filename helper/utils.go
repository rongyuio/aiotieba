// Package helper 提供各 API 模块共用的内部工具。
//
// 对应 aiotieba.helper。
package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// PackJSON 把 obj 序列化为紧凑 JSON 字符串，对应 pack_json。
func PackJSON(obj any) string {
	b, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(b)
}

// ParseJSON 把 data 解码到 v，对应 parse_json。
func ParseJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// ParseJSONMap 解码一个 JSON 对象。
//
// 数字保留为 json.Number，使较大的标识（user id、fid）保持完整精度；
// Python 客户端在这里依赖 Python 的任意精度整数。
func ParseJSONMap(data []byte) (map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var out map[string]any
	if err := dec.Decode(&out); err != nil {
		// 上游返回 HTML（风控页 / 网关页）时，光看 invalid character '<' 分不出是哪种，
		// 所以把原文开头一段带上。Python 版这里是裸的 jsonlib.loads、不带原文，
		// 这一处是有意的偏离，只为排查
		return nil, fmt.Errorf("helper: decoding the JSON response: %w%s", err, bodySnippet(data))
	}
	if out == nil {
		return nil, fmt.Errorf("helper: the JSON response is not an object")
	}
	return out, nil
}

// bodySnippet 把响应体收成一小段用于报错，返回形如 " (原文: <html>…)" 的后缀。
//
// 折叠空白后截断，免得一条错误把日志冲掉；响应体为空（或只有空白）时返回空串，
// 不往错误里加一个空括号。截断按 rune 走：data[:200] 会把多字节字符切成半个。
func bodySnippet(data []byte) string {
	s := strings.Join(strings.Fields(string(data)), " ")
	if s == "" {
		return ""
	}

	const maxRunes = 200
	if r := []rune(s); len(r) > maxRunes {
		s = string(r[:maxRunes]) + "…"
	}

	return " (原文: " + s + ")"
}

// JSONInt 把 m[key] 作为 int64 返回，缺失的键返回 0，对应 Python 在数字字段上使用的 int(m[key]) 转换。
func JSONInt(m map[string]any, key string) int64 {
	v, ok := m[key]
	if !ok {
		return 0
	}
	return toInt64(v)
}

// JSONStr 把 m[key] 作为字符串返回。
func JSONStr(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

// JSONBool 把 m[key] 作为 bool 返回。
func JSONBool(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t != "" && t != "0" && t != "false"
	default:
		return toInt64(t) != 0
	}
}

// JSONMap 把 m[key] 作为嵌套对象返回；缺失或类型不符时返回 nil。
func JSONMap(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	nested, _ := v.(map[string]any)
	return nested
}

// JSONSlice 把 m[key] 作为列表返回；缺失或类型不符时返回 nil。
func JSONSlice(m map[string]any, key string) []any {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	list, _ := v.([]any)
	return list
}

// HasJSONKey 报告键是否存在，对应 `"key" in data_map`。
func HasJSONKey(m map[string]any, key string) bool {
	_, ok := m[key]
	return ok
}

func toInt64(v any) int64 {
	switch t := v.(type) {
	case nil:
		return 0
	case json.Number:
		n, err := t.Int64()
		if err != nil {
			f, ferr := t.Float64()
			if ferr != nil {
				return 0
			}
			return int64(f)
		}
		return n
	case float64:
		return int64(t)
	case float32:
		return int64(t)
	case int:
		return int64(t)
	case int64:
		return t
	case int32:
		return int64(t)
	case bool:
		if t {
			return 1
		}
		return 0
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		if err != nil {
			return 0
		}
		return n
	default:
		return 0
	}
}

// JSONRaw 返回 m[key] 解码后的原始值。
//
// 当 Python 代码用 `==` 比较 JSON 标量时需要它，因为这种比较对类型敏感：
// 字符串 "123" 永不等于数字 123。
func JSONRaw(m map[string]any, key string) (any, bool) {
	v, ok := m[key]
	return v, ok
}

// JSONScalarEqual 报告 a 与 b 是否为同一个 JSON 标量，对应 Python 对 json.loads
// 解码结果使用 `==`。复合值永不相等。
func JSONScalarEqual(a, b any) bool {
	switch a.(type) {
	case nil, bool, string, json.Number, float64, float32, int, int64, int32:
	default:
		return false
	}
	return a == b
}

// BoolInt 把 b 转换为 1 或 0，对应 Python API 模块使用的 `int(is_x)` 转换。
func BoolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// AnyInt64 把任意已解码的 JSON 值转换为 int64。
func AnyInt64(v any) int64 { return toInt64(v) }

// AnyFloat64 把任意已解码的 JSON 值转换为 float64。
func AnyFloat64(v any) float64 {
	switch t := v.(type) {
	case nil:
		return 0
	case json.Number:
		f, err := t.Float64()
		if err != nil {
			return 0
		}
		return f
	case float64:
		return t
	case float32:
		return float64(t)
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return float64(toInt64(t))
	}
}

// IsPortrait 简单判断输入是否符合 portrait 格式。
func IsPortrait(portrait any) bool {
	s, ok := portrait.(string)
	return ok && strings.HasPrefix(s, "tb.")
}

// IsUserName 简单判断输入是否符合 user_name 格式。
func IsUserName(userName any) bool {
	s, ok := userName.(string)
	return ok && !strings.HasPrefix(s, "tb.")
}

// DefaultDatetime 返回各 API 模块使用的默认时间，对应 default_datetime。
func DefaultDatetime() time.Time {
	return time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
}

// Timeout 派生一个 delay 之后取消的 context，对应 Python 客户端的 timeout 辅助函数。
func Timeout(parent context.Context, delay time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, delay)
}
