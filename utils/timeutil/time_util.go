package timeutil

import (
	"errors"
	"time"

	"github.com/araddon/dateparse"
)

var specialFormats = []string{
	"2006年01月02日15时04分05秒",
}

const DefaultTimeLayout = "2006-01-02 15:04:05"

// FormatDate 格式化日期
func FormatDate(dateStr, timeLayout string) (string, error) {
	// 先使用第三方库解析
	parsedDate, err := dateparse.ParseAny(dateStr)
	if err == nil {
		return parsedDate.Format(timeLayout), nil
	}
	// 如果第三方库解析失败，使用自定义解析
	return FormatOtherDate(dateStr, timeLayout)
}

// FormatOtherDate 格式化特殊日期
func FormatOtherDate(dateStr, timeLayout string) (string, error) {
	for _, format := range specialFormats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t.Format(timeLayout), nil
		}
	}
	return "", errors.New("constant.FormatDateErr")
}
