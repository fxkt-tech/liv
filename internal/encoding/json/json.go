package json

import (
	"bytes"
	"encoding/json"
)

func ToString(i any) string {
	if i == nil {
		return ""
	}
	b, _ := json.Marshal(i)
	return string(b)
}

func ToBytes(i any) []byte {
	if i == nil {
		return nil
	}
	b, _ := json.Marshal(i)
	return b
}

// json字节流转struct/map等，不会返回错误，所以确保传入的内容一定是能解析的
func ToV[T any](b []byte) T {
	var i T
	json.Unmarshal(b, &i)
	return i
}

func Pretty(b []byte) string {
	var str bytes.Buffer
	_ = json.Indent(&str, b, "", "    ")
	return str.String()
}
