package helper

import (
	"regexp"
	"strings"
	"time"
)

func GetStringBetween(str, start, end string) string {
	regex := regexp.MustCompile(start + "(.*)" + end)
	result := regex.FindString(str)
	return result
}

func RepairJson(input string) string {
	// Remove trailing comma before the closing brace
	input = strings.TrimRight(input, ", ")

	// Add quotes around keys
	keyRe := regexp.MustCompile(`(\s*{|\s*,\s*)([a-zA-Z0-9_]+)(\s*:)`)
	quotedKeys := keyRe.ReplaceAllString(input, `$1"$2"$3`)

	// Replace single quotes with double quotes
	quoteReplaced := strings.ReplaceAll(quotedKeys, "'", "\"")

	// Remove trailing comma
	trailingCommaRe := regexp.MustCompile(`,(\s*})`)
	trailingCommaRemoved := trailingCommaRe.ReplaceAllString(quoteReplaced, "$1")

	// Escape single quotes inside div class attribute
	divClassRe := regexp.MustCompile(`(class=\\")([^"]*)(\\")`)
	divClassEscaped := divClassRe.ReplaceAllStringFunc(trailingCommaRemoved, func(match string) string {
		return strings.ReplaceAll(match, "\\'", "\\\\'")
	})

	return divClassEscaped
}

func ParseDate(dateStr string) (time.Time, error) {
	const layout = "02.01.2006"
	return time.Parse(layout, dateStr)
}

func ExtractBetweenTags(input string) string {
	start := strings.Index(input, ">") + 1
	end := strings.LastIndex(input, "<")
	if start >= 0 && end >= 0 && end > start {
		return input[start:end]
	}
	return ""
}
