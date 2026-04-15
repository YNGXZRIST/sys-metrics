// Package stringsparser provides string parsing and normalization helpers.
package stringsparser

import "unicode"

// Capitalize uppercases the first rune and lowercases the rest.
func Capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	for i := 1; i < len(r); i++ {
		r[i] = unicode.ToLower(r[i])
	}
	return string(r)
}
