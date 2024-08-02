package sliceutil

import (
	"strings"
)

// StringInSlice 判断字符串是否在dst 切片中
func StringInSlice(dst []string, str string) bool {
	for _, v := range dst {
		if strings.TrimSpace(v) == strings.TrimSpace(str) {
			return true
		}
	}
	return false
}
