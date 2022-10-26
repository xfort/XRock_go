package xutil

import "strings"

// Reverse returns its argument string reversed rune-wise left to right.
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func PadStart(data string, minLen int, padChar string) string {
	if len(data) >= minLen {
		return data
	}
	resStr := strings.Builder{}
	resStr.Grow(minLen)
	for i := 0; i < len(data); i++ {
		resStr.WriteString(padChar)
	}
	resStr.WriteString(data)
	return resStr.String()
}
