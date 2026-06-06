package cmd

import (
	"fmt"
	"strings"

	"github.com/dannyben/config/format"
)

const (
	listMaxLineLength = 79
	listMaxKeyLength  = 60
)

func renderList(entries []format.Entry) string {
	var out strings.Builder
	for _, entry := range entries {
		out.WriteString(renderListEntry(entry))
		out.WriteByte('\n')
	}
	return out.String()
}

func renderListEntry(entry format.Entry) string {
	key := truncateRunes(entry.Key, listMaxKeyLength)
	value := strings.Join(strings.Fields(entry.Value), " ")
	availableValue := listMaxLineLength - runeLen(key) - len("=")
	return fmt.Sprintf("%s=%s", key, truncateRunes(value, availableValue))
}

func truncateRunes(value string, maxLength int) string {
	if maxLength <= 0 {
		return ""
	}
	if runeLen(value) <= maxLength {
		return value
	}
	if maxLength == 1 {
		return "…"
	}
	runes := []rune(value)
	return string(runes[:maxLength-1]) + "…"
}

func runeLen(value string) int {
	return len([]rune(value))
}
