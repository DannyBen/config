package tomldoc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/creachadair/tomledit/parser"
	"github.com/creachadair/tomledit/scanner"
)

func Get(source, path string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	return getValue(source, key)
}

func GetIn(source, collectionPath string, selectors []string, path string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	collection, err := parsePath(collectionPath)
	if err != nil {
		return "", fmt.Errorf("--in: %w", err)
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	assignments, sections, err := scan(source)
	if err != nil {
		return "", err
	}
	index, err := selectedArrayRecordIndex(source, assignments, sections, collection, selectors)
	if err != nil {
		return "", err
	}
	fullKey := arrayRecordKey(collection, index)
	fullKey = append(fullKey, key...)
	return getValue(source, fullKey)
}

func List(source, path string) ([]Entry, error) {
	if err := parseTOMLSource(source); err != nil {
		return nil, err
	}

	var prefix parser.Key
	if path != "" {
		key, err := parsePath(path)
		if err != nil {
			return nil, err
		}
		prefix = key
	}
	return listValues(source, prefix)
}

func listValues(source string, prefix parser.Key) ([]Entry, error) {
	assignments, sections, err := scan(source)
	if err != nil {
		return nil, err
	}

	if len(prefix) > 0 {
		for _, item := range assignments {
			if item.key.Equals(prefix) {
				if hasInternalAssignmentWithPrefix(assignments, prefix) {
					break
				}
				value, err := formatGetValue(strings.TrimSpace(source[item.valueSpan.Pos:item.valueSpan.End]))
				if err != nil {
					return nil, err
				}
				return []Entry{{Key: formatListKey(item.key), Value: value}}, nil
			}
		}
		if !hasSection(sections, prefix) && !hasAssignmentWithPrefix(assignments, prefix) {
			return nil, fmt.Errorf("%s is not set", prefix.String())
		}
	}

	var entries []Entry
	for _, item := range assignments {
		if !hasKeyPrefix(item.key, prefix) {
			continue
		}
		if !item.internal && hasInternalAssignmentWithPrefix(assignments, item.key) {
			continue
		}
		value, err := formatGetValue(strings.TrimSpace(source[item.valueSpan.Pos:item.valueSpan.End]))
		if err != nil {
			return nil, err
		}
		entries = append(entries, Entry{Key: formatListKey(item.key), Value: value})
	}
	return entries, nil
}

func formatListKey(key parser.Key) string {
	parts := make([]string, len(key))
	for i, segment := range key {
		parts[i] = strings.ReplaceAll(segment, ".", "..")
	}
	return strings.Join(parts, ".")
}

func hasInternalAssignmentWithPrefix(assignments []assignment, prefix parser.Key) bool {
	for _, item := range assignments {
		if item.internal && hasKeyPrefix(item.key, prefix) {
			return true
		}
	}
	return false
}

func hasSection(sections []section, key parser.Key) bool {
	for _, sec := range sections {
		if sec.key.Equals(key) {
			return true
		}
	}
	return false
}

func hasAssignmentWithPrefix(assignments []assignment, prefix parser.Key) bool {
	for _, item := range assignments {
		if hasKeyPrefix(item.key, prefix) {
			return true
		}
	}
	return false
}

func hasKeyPrefix(key, prefix parser.Key) bool {
	if len(prefix) == 0 {
		return true
	}
	if len(key) <= len(prefix) {
		return false
	}
	for i := range prefix {
		if key[i] != prefix[i] {
			return false
		}
	}
	return true
}

func getValue(source string, key parser.Key) (string, error) {
	assignments, sections, err := scan(source)
	if err != nil {
		return "", err
	}

	for _, sec := range sections {
		if sec.key.Equals(key) {
			return "", fmt.Errorf("%s is a table, not a value", key.String())
		}
	}
	if hasArrayCollection(sections, key) {
		return "", fmt.Errorf("%s is an array of records, not a value", key.String())
	}
	for _, item := range assignments {
		if item.key.Equals(key) {
			return formatGetValue(strings.TrimSpace(source[item.valueSpan.Pos:item.valueSpan.End]))
		}
	}
	return "", fmt.Errorf("%s is not set", key.String())
}

func formatGetValue(value string) (string, error) {
	parsed, err := parser.ParseValue(value)
	if err != nil {
		return "", err
	}
	return formatParsedGetValue(parsed)
}

func formatParsedGetValue(parsed parser.Value) (string, error) {
	if token, ok := parsed.X.(parser.Token); ok {
		switch token.Type {
		case scanner.String:
			unquoted, err := strconv.Unquote(parsed.String())
			if err != nil {
				return "", err
			}
			return unquoted, nil
		case scanner.LString:
			value := parsed.String()
			return value[1 : len(value)-1], nil
		case scanner.MString:
			return unquoteMultilineBasicString(parsed.String())
		case scanner.MLString:
			value := parsed.String()
			return value[3 : len(value)-3], nil
		}
	}
	if array, ok := parsed.X.(parser.Array); ok {
		return formatGetArray(array)
	}
	return parsed.String(), nil
}

func formatGetArray(array parser.Array) (string, error) {
	var values []string
	for _, item := range array {
		value, ok := item.(parser.Value)
		if !ok {
			continue
		}
		formatted, err := formatParsedGetValue(value)
		if err != nil {
			return "", err
		}
		values = append(values, formatted)
	}
	return "[" + strings.Join(values, ", ") + "]", nil
}

func unquoteMultilineBasicString(value string) (string, error) {
	unquoted, err := scanner.Unescape([]byte(value[3 : len(value)-3]))
	if err != nil {
		return "", err
	}
	return string(unquoted), nil
}
