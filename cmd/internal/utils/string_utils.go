package utils

import (
	"regexp"
)

func IsSafeString(input string) bool {
	matched, _ := regexp.MatchString(`^[\p{Cyrillic}a-zA-Z0-9@.\-_]{3,50}$`, input)
	return matched
}

func CleanString(input string) string {
	return regexp.MustCompile(`[^\p{Cyrillic}a-zA-Z0-9@.\-_]`).ReplaceAllString(input, "")
}
