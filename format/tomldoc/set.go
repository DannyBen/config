package tomldoc

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/creachadair/tomledit/parser"
	"github.com/creachadair/tomledit/scanner"
	"github.com/neongreen/tomlsawyer"
	pelletiertoml "github.com/pelletier/go-toml/v2"
)

type token struct {
	kind scanner.Token
	text string
	span scanner.Span
}

type assignment struct {
	key       parser.Key
	scope     parser.Key
	lineSpan  scanner.Span
	valueSpan scanner.Span
	internal  bool
}

type section struct {
	key      parser.Key
	insertAt int
	inline   bool
	empty    bool
}

type edit struct {
	start int
	end   int
	text  string
}

type semanticSet struct {
	key   parser.Key
	value string
}

type Entry struct {
	Key   string
	Value string
}

func validateTOML(source string) error {
	var out any
	if err := pelletiertoml.Unmarshal([]byte(source), &out); err != nil {
		return fmt.Errorf("invalid TOML: %w", err)
	}
	return nil
}

func parseTOMLSource(source string) error {
	if err := validateTOML(source); err != nil {
		return err
	}
	_, err := tomlsawyer.Parse([]byte(source))
	return err
}

func parsePath(path string) (parser.Key, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}
	var key parser.Key
	var current strings.Builder
	for i := 0; i < len(path); i++ {
		if path[i] != '.' {
			current.WriteByte(path[i])
			continue
		}
		if i+1 < len(path) && path[i+1] == '.' {
			current.WriteByte('.')
			i++
			continue
		}
		if current.Len() == 0 {
			return nil, fmt.Errorf("empty path segment in %q", path)
		}
		key = append(key, current.String())
		current.Reset()
	}
	if current.Len() == 0 {
		return nil, fmt.Errorf("empty path segment in %q", path)
	}
	key = append(key, current.String())
	return key, nil
}

func Set(source, path, rawValue string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	value, err := formatValue(rawValue)
	if err != nil {
		return "", err
	}
	planned, err := planSet(source, key, value)
	if err != nil {
		return "", err
	}
	return applyVerifiedSet(source, planned, key, value)
}

func SetString(source, path, value string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	formatted, err := formatString(value)
	if err != nil {
		return "", err
	}
	planned, err := planSet(source, key, formatted)
	if err != nil {
		return "", err
	}
	return applyVerifiedSet(source, planned, key, formatted)
}

func SetArray(source, path string, values []string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	formatted, err := formatArray(values)
	if err != nil {
		return "", err
	}
	planned, err := planSet(source, key, formatted)
	if err != nil {
		return "", err
	}
	return applyVerifiedSet(source, planned, key, formatted)
}

func SetIn(source, collectionPath, selector, path, rawValue string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	collection, onKey, onValue, onCompare, key, value, err := parseSetIn(collectionPath, selector, path, rawValue, formatValue)
	if err != nil {
		return "", err
	}
	planned, changes, err := planSetIn(source, collection, onKey, onValue, onCompare, key, value)
	if err != nil {
		return "", err
	}
	return applyVerifiedSets(source, planned, changes)
}

func SetInString(source, collectionPath, selector, path, value string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	collection, onKey, onValue, onCompare, key, formatted, err := parseSetIn(collectionPath, selector, path, value, formatString)
	if err != nil {
		return "", err
	}
	planned, changes, err := planSetIn(source, collection, onKey, onValue, onCompare, key, formatted)
	if err != nil {
		return "", err
	}
	return applyVerifiedSets(source, planned, changes)
}

func SetInArray(source, collectionPath, selector, path string, values []string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	collection, onKey, onValue, onCompare, key, value, err := parseSetIn(collectionPath, selector, path, "", func(string) (string, error) {
		return formatArray(values)
	})
	if err != nil {
		return "", err
	}
	planned, changes, err := planSetIn(source, collection, onKey, onValue, onCompare, key, value)
	if err != nil {
		return "", err
	}
	return applyVerifiedSets(source, planned, changes)
}

func Unset(source, path string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	planned, err := planUnset(source, key)
	if err != nil {
		return "", err
	}
	return applyVerifiedUnset(source, planned, key)
}

func UnsetIn(source, collectionPath string, selectors []string, path string) (string, error) {
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
	planned, err := planUnset(source, fullKey)
	if err != nil {
		return "", err
	}
	return applyVerifiedUnset(source, planned, fullKey)
}

func Delete(source, path string, selectors []string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	planned, err := planDelete(source, key, selectors)
	if err != nil {
		return "", err
	}
	return applyVerifiedDelete(source, planned, key, selectors)
}

func DeleteIfEmpty(source, path string) (string, error) {
	entries, err := List(source, path)
	if err != nil {
		if isNotSetError(err) {
			return source, nil
		}
		return "", err
	}
	if len(entries) > 0 {
		return source, nil
	}
	return Delete(source, path, nil)
}

func isNotSetError(err error) bool {
	return strings.HasSuffix(err.Error(), " is not set")
}

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

func planSet(source string, key parser.Key, value string) (edit, error) {
	assignments, sections, err := scan(source)
	if err != nil {
		return edit{}, err
	}

	for _, item := range assignments {
		if item.key.Equals(key) {
			return edit{start: item.valueSpan.Pos, end: item.valueSpan.End, text: value}, nil
		}
	}
	for _, sec := range sections {
		if sec.key.Equals(key) {
			return edit{}, fmt.Errorf("%s is a table, not a value", key.String())
		}
	}
	if collection, index, _, ok := indexedRecordPath(key); ok && hasArrayCollection(sections, collection) {
		return edit{}, fmt.Errorf("%s.%d is not set", collection.String(), index)
	}

	if len(key) > 1 {
		return planMissingNestedSet(source, assignments, sections, key, value)
	}

	return insertRootKey(source, key.String()+" = "+value), nil
}

func parseSetIn(collectionPath, selector, path, rawValue string, format func(string) (string, error)) (parser.Key, parser.Key, string, string, parser.Key, string, error) {
	collection, err := parsePath(collectionPath)
	if err != nil {
		return nil, nil, "", "", nil, "", fmt.Errorf("--in: %w", err)
	}
	onPath, onRaw, ok := strings.Cut(selector, ":")
	if !ok || onPath == "" {
		return nil, nil, "", "", nil, "", fmt.Errorf("--on must use FIELD:VALUE")
	}
	onKey, err := parsePath(onPath)
	if err != nil {
		return nil, nil, "", "", nil, "", fmt.Errorf("--on: %w", err)
	}
	onValue, err := formatValue(onRaw)
	if err != nil {
		return nil, nil, "", "", nil, "", fmt.Errorf("--on: %w", err)
	}
	onCompare, err := formatGetValue(onValue)
	if err != nil {
		return nil, nil, "", "", nil, "", fmt.Errorf("--on: %w", err)
	}
	key, err := parsePath(path)
	if err != nil {
		return nil, nil, "", "", nil, "", err
	}
	value, err := format(rawValue)
	if err != nil {
		return nil, nil, "", "", nil, "", err
	}
	return collection, onKey, onValue, onCompare, key, value, nil
}

func planSetIn(source string, collection, onKey parser.Key, onValue, onCompare string, key parser.Key, value string) (edit, []semanticSet, error) {
	assignments, sections, err := scan(source)
	if err != nil {
		return edit{}, nil, err
	}

	var matches []int
	for _, sec := range sections {
		recordIndex, ok := arrayRecordIndex(sec.key, collection)
		if !ok {
			continue
		}
		recordKey := arrayRecordKey(collection, recordIndex)
		for _, item := range assignments {
			if !stripKeyPrefix(item.key, recordKey).Equals(onKey) {
				continue
			}
			got, err := formatGetValue(strings.TrimSpace(source[item.valueSpan.Pos:item.valueSpan.End]))
			if err != nil {
				return edit{}, nil, err
			}
			if got == onCompare {
				matches = append(matches, recordIndex)
			}
			break
		}
	}

	switch len(matches) {
	case 0:
		recordKey := arrayRecordKey(collection, nextArrayRecordIndex(sections, collection))
		changes := []semanticSet{{key: append(append(parser.Key{}, recordKey...), onKey...), value: onValue}}
		if !key.Equals(onKey) {
			changes = append(changes, semanticSet{key: append(append(parser.Key{}, recordKey...), key...), value: value})
		}
		return appendArrayRecord(source, collection, onKey, onValue, key, value), changes, nil
	case 1:
		fullKey := arrayRecordKey(collection, matches[0])
		fullKey = append(fullKey, key...)
		planned, err := planSet(source, fullKey, value)
		if err != nil {
			return edit{}, nil, err
		}
		return planned, []semanticSet{{key: fullKey, value: value}}, nil
	default:
		return edit{}, nil, fmt.Errorf("%s has multiple records matching %s", collection.String(), onKey.String()+":"+onCompare)
	}
}

func appendArrayRecord(source string, collection, onKey parser.Key, onValue string, key parser.Key, value string) edit {
	nl := lineEnding(source)
	lines := []string{"[[" + collection.String() + "]]"}
	lines = append(lines, onKey.String()+" = "+onValue)
	if !key.Equals(onKey) {
		lines = append(lines, key.String()+" = "+value)
	}
	return edit{start: len(source), end: len(source), text: blockPrefix(source) + strings.Join(lines, nl) + nl}
}

func arrayRecordKey(collection parser.Key, index int) parser.Key {
	key := append(parser.Key{}, collection...)
	key = append(key, strconv.Itoa(index))
	return key
}

func indexedRecordPath(key parser.Key) (parser.Key, int, parser.Key, bool) {
	for i, segment := range key {
		index, err := strconv.Atoi(segment)
		if err != nil {
			continue
		}
		if i == 0 || i == len(key)-1 || index < 0 {
			return nil, 0, nil, false
		}
		collection := append(parser.Key{}, key[:i]...)
		field := append(parser.Key{}, key[i+1:]...)
		return collection, index, field, true
	}
	return nil, 0, nil, false
}

func hasArrayCollection(sections []section, collection parser.Key) bool {
	for _, sec := range sections {
		if _, ok := arrayRecordIndex(sec.key, collection); ok {
			return true
		}
	}
	return false
}

func nextArrayRecordIndex(sections []section, collection parser.Key) int {
	next := 0
	for _, sec := range sections {
		index, ok := arrayRecordIndex(sec.key, collection)
		if ok && index >= next {
			next = index + 1
		}
	}
	return next
}

func arrayRecordIndex(key, collection parser.Key) (int, bool) {
	if len(key) != len(collection)+1 {
		return 0, false
	}
	for i := range collection {
		if key[i] != collection[i] {
			return 0, false
		}
	}
	index, err := strconv.Atoi(key[len(collection)])
	return index, err == nil && index >= 0
}

func stripKeyPrefix(key, prefix parser.Key) parser.Key {
	if len(key) < len(prefix) {
		return nil
	}
	for i := range prefix {
		if key[i] != prefix[i] {
			return nil
		}
	}
	return parser.Key(key[len(prefix):])
}

func planMissingNestedSet(source string, assignments []assignment, sections []section, key parser.Key, value string) (edit, error) {
	parent := parser.Key(key[:len(key)-1])
	if sec, ok := findSection(sections, parent); ok {
		if sec.inline {
			return insertInlineTableChild(source, sec, key[len(key)-1], value), nil
		}
		text := parser.Key{key[len(key)-1]}.String() + " = " + value + lineEnding(source)
		return edit{start: sec.insertAt, end: sec.insertAt, text: prefixLineEndingAt(source, sec.insertAt) + text}, nil
	}
	if insertAt, scope, ok := dottedSiblingInsertAt(assignments, parent); ok {
		relative := relativeKey(key, scope)
		text := relative.String() + " = " + value + lineEnding(source)
		return edit{start: insertAt, end: insertAt, text: prefixLineEndingAt(source, insertAt) + text}, nil
	}
	if hasSiblingSection(sections, parent) {
		return appendTable(source, parent, key[len(key)-1], value), nil
	}
	if sec, prefix, ok := longestSectionPrefix(sections, key); ok {
		child := key[len(prefix):]
		if sec.inline {
			return insertInlineTableChild(source, sec, child.String(), value), nil
		}
		text := child.String() + " = " + value + lineEnding(source)
		return edit{start: sec.insertAt, end: sec.insertAt, text: prefixLineEndingAt(source, sec.insertAt) + text}, nil
	}
	return appendTable(source, parent, key[len(key)-1], value), nil
}

func insertInlineTableChild(source string, sec section, leaf string, value string) edit {
	start := sec.insertAt
	text := parser.Key{leaf}.String() + " = " + value
	if sec.empty {
		if sec.insertAt > 0 && sec.insertAt <= len(source) && source[sec.insertAt-1] == ' ' {
			start = sec.insertAt - 1
		}
		return edit{start: start, end: sec.insertAt, text: " " + text + " "}
	}
	if sec.insertAt > 0 && sec.insertAt <= len(source) && source[sec.insertAt-1] == ' ' {
		start = sec.insertAt - 1
		return edit{start: start, end: sec.insertAt, text: ", " + text + " "}
	}
	return edit{start: start, end: sec.insertAt, text: ", " + text}
}

func findSection(sections []section, key parser.Key) (section, bool) {
	for _, sec := range sections {
		if sec.key.Equals(key) {
			return sec, true
		}
	}
	return section{}, false
}

func longestSectionPrefix(sections []section, key parser.Key) (section, parser.Key, bool) {
	var best section
	var bestKey parser.Key
	for _, sec := range sections {
		if len(sec.key) == 0 || len(sec.key) >= len(key) {
			continue
		}
		if len(sec.key) <= len(bestKey) || !sec.key.IsPrefixOf(key) {
			continue
		}
		best = sec
		bestKey = sec.key
	}
	return best, bestKey, len(bestKey) > 0
}

func hasSiblingSection(sections []section, parent parser.Key) bool {
	if len(parent) == 0 {
		return false
	}
	grandparent := parser.Key(parent[:len(parent)-1])
	for _, sec := range sections {
		if sec.key.Equals(parent) {
			continue
		}
		if len(sec.key) == len(parent) && grandparent.IsPrefixOf(sec.key) {
			return true
		}
	}
	return false
}

func dottedSiblingInsertAt(assignments []assignment, parent parser.Key) (int, parser.Key, bool) {
	insertAt := -1
	var scope parser.Key
	for _, item := range assignments {
		if item.key.Equals(parent) {
			continue
		}
		if hasKeyPrefix(item.key, parent) && item.lineSpan.End > insertAt {
			insertAt = item.lineSpan.End
			scope = item.scope
		}
	}
	return insertAt, scope, insertAt >= 0
}

func relativeKey(key, scope parser.Key) parser.Key {
	if len(scope) == 0 || !scope.IsPrefixOf(key) {
		return key
	}
	return key[len(scope):]
}

func insertRootKey(source, text string) edit {
	nl := lineEnding(source)
	insertAt := rootInsertAt(source)
	suffix := nl
	if insertAt < len(source) && !hasLineEndingPrefix(source[insertAt:]) {
		suffix = nl + nl
	}
	return edit{start: insertAt, end: insertAt, text: prefixLineEndingAt(source, insertAt) + text + suffix}
}

func rootInsertAt(source string) int {
	insertAt := len(source)
	for _, stmt := range statements(source) {
		if len(stmt.tokens) == 0 {
			continue
		}
		if _, ok := tableKey(stmt.tokens); ok {
			return insertAtBeforeBlankLine(source, stmt.start)
		}
		if _, _, _, ok := assignmentKey(stmt.tokens); ok {
			insertAt = stmt.end
		}
	}
	return insertAt
}

func insertAtBeforeBlankLine(source string, pos int) int {
	insertAt := pos
	for insertAt > 0 {
		lineStart := strings.LastIndexByte(source[:insertAt-1], '\n') + 1
		if strings.TrimSpace(source[lineStart:insertAt]) != "" {
			break
		}
		insertAt = lineStart
	}
	return insertAt
}

func appendTable(source string, parent parser.Key, leaf string, value string) edit {
	nl := lineEnding(source)
	text := "[" + parent.String() + "]" + nl + parser.Key{leaf}.String() + " = " + value + nl
	return edit{start: len(source), end: len(source), text: blockPrefix(source) + text}
}

func blockPrefix(source string) string {
	nl := lineEnding(source)
	if source == "" {
		return ""
	}
	if strings.HasSuffix(source, nl+nl) {
		return ""
	}
	if strings.HasSuffix(source, nl) {
		return nl
	}
	return nl + nl
}

func lineEnding(source string) string {
	i := strings.IndexByte(source, '\n')
	if i > 0 && source[i-1] == '\r' {
		return "\r\n"
	}
	return "\n"
}

func hasLineEndingPrefix(source string) bool {
	return strings.HasPrefix(source, "\n") || strings.HasPrefix(source, "\r\n")
}

func prefixLineEndingAt(source string, pos int) string {
	if pos <= 0 || pos > len(source) || source[pos-1] == '\n' {
		return ""
	}
	return lineEnding(source)
}

func planUnset(source string, key parser.Key) (edit, error) {
	assignments, sections, err := scan(source)
	if err != nil {
		return edit{}, err
	}

	for _, sec := range sections {
		if sec.key.Equals(key) {
			return edit{}, fmt.Errorf("%s is a table, not a value", key.String())
		}
	}
	for _, item := range assignments {
		if item.key.Equals(key) {
			return edit{start: item.lineSpan.Pos, end: item.lineSpan.End, text: ""}, nil
		}
	}
	return edit{}, fmt.Errorf("%s is not set", key.String())
}

func planDelete(source string, key parser.Key, selectors []string) ([]edit, error) {
	assignments, sections, err := scan(source)
	if err != nil {
		return nil, err
	}

	if len(selectors) > 0 {
		return planDeleteSelectedRecord(source, assignments, sections, key, selectors)
	}
	if collection, index, ok := indexedRecordKey(key); ok {
		if !hasArrayCollection(sections, collection) {
			return nil, fmt.Errorf("%s is not set", key.String())
		}
		start, end, ok := arrayRecordDeleteRange(source, collection, index)
		if !ok {
			return nil, fmt.Errorf("%s has no record at index %d", collection.String(), index)
		}
		return []edit{{start: start, end: end, text: ""}}, nil
	}
	for _, item := range assignments {
		if item.key.Equals(key) {
			return nil, fmt.Errorf("%s is a value, use unset to remove fields", key.String())
		}
	}
	for _, sec := range sections {
		if sec.key.Equals(key) {
			if sec.inline {
				return nil, fmt.Errorf("%s is an inline table; delete is not implemented for inline tables yet", key.String())
			}
			start, end, ok := tableDeleteRange(source, key)
			if !ok {
				return nil, fmt.Errorf("%s cannot be safely deleted", key.String())
			}
			return []edit{{start: start, end: end, text: ""}}, nil
		}
	}
	if hasArrayCollection(sections, key) {
		planned, ok := arrayCollectionDeleteRanges(source, key, sections)
		if !ok {
			return nil, fmt.Errorf("%s cannot be safely deleted", key.String())
		}
		return planned, nil
	}
	return nil, fmt.Errorf("%s is not set", key.String())
}

func indexedRecordKey(key parser.Key) (parser.Key, int, bool) {
	if len(key) < 2 {
		return nil, 0, false
	}
	index, err := strconv.Atoi(key[len(key)-1])
	if err != nil || index < 0 {
		return nil, 0, false
	}
	collection := append(parser.Key{}, key[:len(key)-1]...)
	return collection, index, true
}

type deleteSelector struct {
	key     parser.Key
	compare string
	label   string
}

func planDeleteSelectedRecord(source string, assignments []assignment, sections []section, collection parser.Key, rawSelectors []string) ([]edit, error) {
	index, err := selectedArrayRecordIndex(source, assignments, sections, collection, rawSelectors)
	if err != nil {
		return nil, err
	}
	start, end, ok := arrayRecordDeleteRange(source, collection, index)
	if !ok {
		return nil, fmt.Errorf("%s.%d cannot be safely deleted", collection.String(), index)
	}
	return []edit{{start: start, end: end, text: ""}}, nil
}

func selectedArrayRecordIndex(source string, assignments []assignment, sections []section, collection parser.Key, rawSelectors []string) (int, error) {
	if !hasArrayCollection(sections, collection) {
		return 0, fmt.Errorf("%s is not an array of records", collection.String())
	}
	selectors, err := parseDeleteSelectors(rawSelectors)
	if err != nil {
		return 0, err
	}

	var matches []int
	for _, sec := range sections {
		recordIndex, ok := arrayRecordIndex(sec.key, collection)
		if !ok {
			continue
		}
		recordKey := arrayRecordKey(collection, recordIndex)
		if recordMatchesSelectors(source, assignments, recordKey, selectors) {
			matches = append(matches, recordIndex)
		}
	}

	switch len(matches) {
	case 0:
		return 0, fmt.Errorf("%s has no records matching %s", collection.String(), joinSelectorLabels(selectors))
	case 1:
		return matches[0], nil
	default:
		return 0, fmt.Errorf("%s has multiple records matching %s", collection.String(), joinSelectorLabels(selectors))
	}
}

func parseDeleteSelectors(raw []string) ([]deleteSelector, error) {
	selectors := make([]deleteSelector, 0, len(raw))
	for _, selector := range raw {
		onPath, onRaw, ok := strings.Cut(selector, ":")
		if !ok || onPath == "" {
			return nil, fmt.Errorf("--on must use FIELD:VALUE")
		}
		onKey, err := parsePath(onPath)
		if err != nil {
			return nil, fmt.Errorf("--on: %w", err)
		}
		onValue, err := formatValue(onRaw)
		if err != nil {
			return nil, fmt.Errorf("--on: %w", err)
		}
		onCompare, err := formatGetValue(onValue)
		if err != nil {
			return nil, fmt.Errorf("--on: %w", err)
		}
		selectors = append(selectors, deleteSelector{
			key:     onKey,
			compare: onCompare,
			label:   onKey.String() + ":" + onCompare,
		})
	}
	return selectors, nil
}

func recordMatchesSelectors(source string, assignments []assignment, recordKey parser.Key, selectors []deleteSelector) bool {
	for _, selector := range selectors {
		matched := false
		for _, item := range assignments {
			if !stripKeyPrefix(item.key, recordKey).Equals(selector.key) {
				continue
			}
			got, err := formatGetValue(strings.TrimSpace(source[item.valueSpan.Pos:item.valueSpan.End]))
			if err != nil {
				return false
			}
			if got == selector.compare {
				matched = true
			}
			break
		}
		if !matched {
			return false
		}
	}
	return true
}

func joinSelectorLabels(selectors []deleteSelector) string {
	labels := make([]string, 0, len(selectors))
	for _, selector := range selectors {
		labels = append(labels, selector.label)
	}
	return strings.Join(labels, " ")
}

func tableDeleteRange(source string, key parser.Key) (int, int, bool) {
	stmts := statements(source)
	start := -1
	end := -1
	headerStart := -1
	hasNextTable := false
	for _, stmt := range stmts {
		if len(stmt.tokens) == 0 {
			continue
		}
		if table, ok := tableKey(stmt.tokens); ok {
			if start >= 0 {
				hasNextTable = true
				break
			}
			if table.Equals(key) {
				headerStart = stmt.start
				start = stmt.start
				end = stmt.end
			}
			continue
		}
		if _, ok := arrayTableKey(stmt.tokens); ok {
			if start >= 0 {
				hasNextTable = true
				break
			}
			continue
		}
		if start >= 0 {
			end = stmt.end
		}
	}
	if start < 0 {
		return 0, 0, false
	}
	end = trimDeleteTrailingBlankLines(source, end)
	if !hasNextTable {
		start = insertAtBeforeBlankLine(source, headerStart)
	}
	return start, end, true
}

func arrayRecordDeleteRange(source string, collection parser.Key, index int) (int, int, bool) {
	stmts := statements(source)
	recordIndex := -1
	start := -1
	end := -1
	headerStart := -1
	hasNextTable := false
	for _, stmt := range stmts {
		if len(stmt.tokens) == 0 {
			continue
		}
		if key, ok := arrayTableKey(stmt.tokens); ok && key.Equals(collection) {
			recordIndex++
			if start >= 0 {
				hasNextTable = true
				break
			}
			if recordIndex == index {
				headerStart = stmt.start
				start = stmt.start
				end = stmt.end
			}
			continue
		}
		if _, ok := tableKey(stmt.tokens); ok {
			if start >= 0 {
				hasNextTable = true
				break
			}
			continue
		}
		if _, ok := arrayTableKey(stmt.tokens); ok {
			if start >= 0 {
				hasNextTable = true
				break
			}
			continue
		}
		if start >= 0 {
			end = stmt.end
		}
	}
	if start < 0 {
		return 0, 0, false
	}
	end = trimDeleteTrailingBlankLines(source, end)
	if !hasNextTable {
		start = insertAtBeforeBlankLine(source, headerStart)
	}
	return start, end, true
}

func arrayCollectionDeleteRanges(source string, collection parser.Key, sections []section) ([]edit, bool) {
	var planned []edit
	recordCount := 0
	for _, sec := range sections {
		if _, ok := arrayRecordIndex(sec.key, collection); ok {
			recordCount++
		}
	}
	for i := 0; i < recordCount; i++ {
		start, end, ok := arrayRecordDeleteRange(source, collection, i)
		if !ok {
			return nil, false
		}
		planned = append(planned, edit{start: start, end: end, text: ""})
	}
	return planned, len(planned) > 0
}

func trimDeleteTrailingBlankLines(source string, end int) int {
	for end < len(source) {
		next := lineEnd(source, end)
		if strings.TrimSpace(source[end:next]) != "" {
			break
		}
		end = next
	}
	return end
}

func scan(source string) ([]assignment, []section, error) {
	var assignments []assignment
	var sections []section
	var current parser.Key
	arrayCounts := map[string]int{}

	for _, stmt := range statements(source) {
		if len(stmt.tokens) == 0 {
			continue
		}
		if key, ok := arrayTableKey(stmt.tokens); ok {
			index := arrayCounts[key.String()]
			arrayCounts[key.String()] = index + 1
			current = arrayRecordKey(key, index)
			sections = append(sections, section{key: current, insertAt: stmt.end})
			continue
		}
		if key, ok := tableKey(stmt.tokens); ok {
			current = key
			sections = append(sections, section{key: key, insertAt: stmt.end})
			continue
		}
		key, valueSpan, valueTokens, ok := assignmentKey(stmt.tokens)
		if !ok {
			continue
		}
		fullKey := append(parser.Key{}, current...)
		fullKey = append(fullKey, key...)
		if isInlineTableValue(valueTokens) {
			assignments = append(assignments, inlineAssignments(fullKey, valueTokens)...)
			sections = append(sections, inlineSections(fullKey, valueTokens)...)
			sections = append(sections, inlineSection(fullKey, valueTokens))
			updateCurrentSectionInsertAt(sections, current, stmt.end)
			continue
		}
		assignments = append(assignments, arrayItemAssignments(fullKey, valueTokens)...)
		assignments = append(assignments, assignment{
			key:       fullKey,
			scope:     append(parser.Key{}, current...),
			lineSpan:  scanner.Span{Pos: stmt.start, End: stmt.end},
			valueSpan: valueSpan,
		})
		updateCurrentSectionInsertAt(sections, current, stmt.end)
	}

	return assignments, sections, nil
}

func arrayItemAssignments(parent parser.Key, tokens []token) []assignment {
	if len(tokens) < 2 || tokens[0].kind != scanner.LBracket || tokens[len(tokens)-1].kind != scanner.RBracket {
		return nil
	}
	segments := arraySegments(tokens)
	out := make([]assignment, 0, len(segments))
	for i, segment := range segments {
		if len(segment) == 0 {
			continue
		}
		key := append(parser.Key{}, parent...)
		key = append(key, strconv.Itoa(i))
		out = append(out, assignment{
			key:       key,
			lineSpan:  scanner.Span{Pos: tokens[0].span.Pos, End: tokens[len(tokens)-1].span.End},
			valueSpan: scanner.Span{Pos: segment[0].span.Pos, End: segment[len(segment)-1].span.End},
			internal:  true,
		})
	}
	return out
}

func updateCurrentSectionInsertAt(sections []section, current parser.Key, insertAt int) {
	if len(current) == 0 {
		return
	}
	for i := len(sections) - 1; i >= 0; i-- {
		if sections[i].key.Equals(current) && !sections[i].inline {
			sections[i].insertAt = insertAt
			return
		}
	}
}

type statement struct {
	start  int
	end    int
	tokens []token
}

func statements(source string) []statement {
	var out []statement
	var current statement
	s := scanner.New(strings.NewReader(source))

	flush := func(end int) {
		if len(current.tokens) > 0 {
			current.end = end
			out = append(out, current)
			current = statement{}
		}
	}

	for {
		err := s.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			break
		}
		kind := s.Token()
		span := s.Span()
		if kind == scanner.Newline {
			flush(span.End)
			continue
		}
		if kind == scanner.Comment {
			if len(current.tokens) > 0 {
				flush(lineEnd(source, span.Pos))
			}
			continue
		}
		if len(current.tokens) == 0 {
			current.start = span.Pos
		}
		current.tokens = append(current.tokens, token{kind: kind, text: string(s.Text()), span: span})
	}
	flush(len(source))
	return out
}

func lineEnd(source string, pos int) int {
	end := strings.IndexByte(source[pos:], '\n')
	if end < 0 {
		return len(source)
	}
	return pos + end + 1
}

func tableKey(tokens []token) (parser.Key, bool) {
	if len(tokens) < 3 || tokens[0].kind != scanner.LBracket || tokens[len(tokens)-1].kind != scanner.RBracket {
		return nil, false
	}
	key, err := parser.ParseKey(joinTokenText(tokens[1 : len(tokens)-1]))
	return key, err == nil
}

func arrayTableKey(tokens []token) (parser.Key, bool) {
	if len(tokens) < 5 ||
		tokens[0].kind != scanner.LBracket ||
		tokens[1].kind != scanner.LBracket ||
		tokens[len(tokens)-2].kind != scanner.RBracket ||
		tokens[len(tokens)-1].kind != scanner.RBracket {
		return nil, false
	}
	key, err := parser.ParseKey(joinTokenText(tokens[2 : len(tokens)-2]))
	return key, err == nil
}

func assignmentKey(tokens []token) (parser.Key, scanner.Span, []token, bool) {
	eq := -1
	for i, tok := range tokens {
		if tok.kind == scanner.Equal {
			eq = i
			break
		}
	}
	if eq <= 0 || eq == len(tokens)-1 {
		return nil, scanner.Span{}, nil, false
	}
	key, err := parser.ParseKey(joinTokenText(tokens[:eq]))
	if err != nil {
		return nil, scanner.Span{}, nil, false
	}
	valueTokens := tokens[eq+1:]
	return key, scanner.Span{Pos: valueTokens[0].span.Pos, End: valueTokens[len(valueTokens)-1].span.End}, valueTokens, true
}

func isInlineTableValue(tokens []token) bool {
	return len(tokens) >= 2 && tokens[0].kind == scanner.LInline && tokens[len(tokens)-1].kind == scanner.RInline
}

func inlineAssignments(parent parser.Key, tokens []token) []assignment {
	items := inlineItems(parent, tokens)
	var out []assignment
	for _, item := range items {
		if item.inline {
			out = append(out, inlineAssignments(item.key, item.valueTokens)...)
			continue
		}
		out = append(out, assignment{key: item.key, lineSpan: item.deleteSpan, valueSpan: item.valueSpan})
	}
	return out
}

func inlineSections(parent parser.Key, tokens []token) []section {
	items := inlineItems(parent, tokens)
	var out []section
	for _, item := range items {
		if item.inline {
			out = append(out, inlineSection(item.key, item.valueTokens))
			out = append(out, inlineSections(item.key, item.valueTokens)...)
		}
	}
	return out
}

func inlineSection(key parser.Key, tokens []token) section {
	return section{
		key:      key,
		insertAt: tokens[len(tokens)-1].span.Pos,
		inline:   true,
		empty:    len(inlineSegments(tokens)) == 0,
	}
}

type inlineItem struct {
	key         parser.Key
	valueSpan   scanner.Span
	deleteSpan  scanner.Span
	valueTokens []token
	inline      bool
}

func inlineItems(parent parser.Key, tokens []token) []inlineItem {
	segments := inlineSegments(tokens)
	out := make([]inlineItem, 0, len(segments))
	for i, segment := range segments {
		eq := topLevelEqual(segment)
		if eq <= 0 || eq == len(segment)-1 {
			continue
		}
		key, err := parser.ParseKey(joinTokenText(segment[:eq]))
		if err != nil {
			continue
		}
		valueTokens := segment[eq+1:]
		fullKey := append(parser.Key{}, parent...)
		fullKey = append(fullKey, key...)
		valueSpan := scanner.Span{Pos: valueTokens[0].span.Pos, End: valueTokens[len(valueTokens)-1].span.End}
		out = append(out, inlineItem{
			key:         fullKey,
			valueSpan:   valueSpan,
			deleteSpan:  inlineDeleteSpan(segments, i),
			valueTokens: valueTokens,
			inline:      isInlineTableValue(valueTokens),
		})
	}
	return out
}

func inlineSegments(tokens []token) [][]token {
	if !isInlineTableValue(tokens) {
		return nil
	}
	inner := tokens[1 : len(tokens)-1]
	if len(inner) == 0 {
		return nil
	}
	var segments [][]token
	start := 0
	bracketDepth := 0
	inlineDepth := 0
	for i, tok := range inner {
		switch tok.kind {
		case scanner.LBracket:
			bracketDepth++
		case scanner.RBracket:
			bracketDepth--
		case scanner.LInline:
			inlineDepth++
		case scanner.RInline:
			inlineDepth--
		case scanner.Comma:
			if bracketDepth == 0 && inlineDepth == 0 {
				segments = append(segments, inner[start:i])
				start = i + 1
			}
		}
	}
	segments = append(segments, inner[start:])
	return segments
}

func arraySegments(tokens []token) [][]token {
	if len(tokens) < 2 || tokens[0].kind != scanner.LBracket || tokens[len(tokens)-1].kind != scanner.RBracket {
		return nil
	}
	inner := tokens[1 : len(tokens)-1]
	if len(inner) == 0 {
		return nil
	}
	var segments [][]token
	start := 0
	bracketDepth := 0
	inlineDepth := 0
	for i, tok := range inner {
		switch tok.kind {
		case scanner.LBracket:
			bracketDepth++
		case scanner.RBracket:
			bracketDepth--
		case scanner.LInline:
			inlineDepth++
		case scanner.RInline:
			inlineDepth--
		case scanner.Comma:
			if bracketDepth == 0 && inlineDepth == 0 {
				segments = append(segments, inner[start:i])
				start = i + 1
			}
		}
	}
	segments = append(segments, inner[start:])
	return segments
}

func topLevelEqual(tokens []token) int {
	bracketDepth := 0
	inlineDepth := 0
	for i, tok := range tokens {
		switch tok.kind {
		case scanner.LBracket:
			bracketDepth++
		case scanner.RBracket:
			bracketDepth--
		case scanner.LInline:
			inlineDepth++
		case scanner.RInline:
			inlineDepth--
		case scanner.Equal:
			if bracketDepth == 0 && inlineDepth == 0 {
				return i
			}
		}
	}
	return -1
}

func inlineDeleteSpan(segments [][]token, index int) scanner.Span {
	segment := segments[index]
	if len(segments) == 1 {
		return scanner.Span{Pos: segment[0].span.Pos, End: inlineSegmentValueEnd(segment)}
	}
	if index == 0 {
		return scanner.Span{Pos: segment[0].span.Pos, End: segments[index+1][0].span.Pos}
	}
	return scanner.Span{Pos: inlineSegmentValueEnd(segments[index-1]), End: inlineSegmentValueEnd(segment)}
}

func inlineSegmentValueEnd(segment []token) int {
	eq := topLevelEqual(segment)
	if eq < 0 || eq == len(segment)-1 {
		return segment[len(segment)-1].span.End
	}
	return segment[len(segment)-1].span.End
}

func joinTokenText(tokens []token) string {
	var out strings.Builder
	for _, tok := range tokens {
		out.WriteString(tok.text)
	}
	return out.String()
}

func formatValue(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if parsed, err := parser.ParseValue(value); err == nil && isScalarValue(parsed) {
		return value, nil
	}
	return formatString(raw)
}

func isScalarValue(value parser.Value) bool {
	_, ok := value.X.(parser.Token)
	return ok
}

func formatString(value string) (string, error) {
	if strings.ContainsAny(value, "\r\n") {
		formatted := `"""` + string(scanner.EscapeMultiline(value)) + `"""`
		if _, err := parser.ParseValue(formatted); err != nil {
			return "", err
		}
		return formatted, nil
	}
	formatted, err := tomlsawyer.FormatValueToString(value)
	if err != nil {
		return "", err
	}
	if _, err := parser.ParseValue(formatted); err != nil {
		return "", err
	}
	return formatted, nil
}

func formatArray(values []string) (string, error) {
	items := make([]string, 0, len(values))
	for _, value := range values {
		formatted, err := formatValue(value)
		if err != nil {
			return "", err
		}
		items = append(items, formatted)
	}
	formatted := "[" + strings.Join(items, ", ") + "]"
	if _, err := parser.ParseValue(formatted); err != nil {
		return "", err
	}
	return formatted, nil
}

func appendLine(source, text string) edit {
	nl := lineEnding(source)
	prefix := ""
	if source != "" && !strings.HasSuffix(source, nl) {
		prefix = nl
	}
	return edit{start: len(source), end: len(source), text: prefix + text + nl}
}

func applyVerifiedSet(source string, planned edit, key parser.Key, value string) (string, error) {
	return applyVerifiedSets(source, planned, []semanticSet{{key: key, value: value}})
}

func applyVerifiedSets(source string, planned edit, changes []semanticSet) (string, error) {
	updated := applyEdit(source, planned)
	return verifySemanticPatch(source, updated, func(expected map[string]string) error {
		for _, change := range changes {
			if err := semanticSetValue(expected, change.key, change.value); err != nil {
				return err
			}
		}
		return nil
	})
}

func applyVerifiedUnset(source string, planned edit, key parser.Key) (string, error) {
	updated := applyEdit(source, planned)
	updated, err := verifySemanticPatch(source, updated, func(expected map[string]string) error {
		semanticDeletePrefix(expected, key)
		return nil
	})
	if err != nil {
		return "", err
	}
	return pruneEmptyParentTables(updated, key[:len(key)-1])
}

func applyVerifiedDelete(source string, planned []edit, key parser.Key, selectors []string) (string, error) {
	updated := applyEdits(source, planned)
	if len(selectors) > 0 || indexedRecordDelete(key) {
		if err := parseTOMLSource(updated); err != nil {
			return "", err
		}
		return updated, nil
	}
	return verifySemanticPatch(source, updated, func(expected map[string]string) error {
		for item := range expected {
			itemKey, err := parsePath(item)
			if err != nil {
				return err
			}
			if itemKey.Equals(key) || hasKeyPrefix(itemKey, key) {
				delete(expected, item)
			}
		}
		return nil
	})
}

func semanticSetValue(values map[string]string, key parser.Key, value string) error {
	semanticDeletePrefix(values, key)
	parsed, err := parser.ParseValue(value)
	if err != nil {
		return err
	}
	if array, ok := parsed.X.(parser.Array); ok {
		if len(array) == 0 {
			values[formatListKey(key)] = "[]"
			return nil
		}
		for i, item := range array {
			parsedItem, ok := item.(parser.Value)
			if !ok {
				continue
			}
			logical, err := formatParsedGetValue(parsedItem)
			if err != nil {
				return err
			}
			itemKey := append(append(parser.Key{}, key...), strconv.Itoa(i))
			values[formatListKey(itemKey)] = logical
		}
		return nil
	}
	logical, err := formatParsedGetValue(parsed)
	if err != nil {
		return err
	}
	values[formatListKey(key)] = logical
	return nil
}

func semanticDeletePrefix(values map[string]string, key parser.Key) {
	for item := range values {
		itemKey, err := parsePath(item)
		if err != nil {
			continue
		}
		if itemKey.Equals(key) || hasKeyPrefix(itemKey, key) {
			delete(values, item)
		}
	}
}

func pruneEmptyParentTables(source string, parent parser.Key) (string, error) {
	current := source
	for len(parent) > 0 {
		pruned, ok, err := pruneEmptyTable(current, parent)
		if err != nil {
			return "", err
		}
		if !ok {
			parent = parent[:len(parent)-1]
			continue
		}
		next, err := verifySemanticPatch(current, pruned, func(map[string]string) error {
			return nil
		})
		if err != nil {
			return "", err
		}
		current = next
		parent = parent[:len(parent)-1]
	}
	return current, nil
}

func pruneEmptyTable(source string, key parser.Key) (string, bool, error) {
	assignments, sections, err := scan(source)
	if err != nil {
		return "", false, err
	}
	if !explicitEmptyTable(assignments, sections, key) {
		return source, false, nil
	}
	start, end, ok := tableDeleteRange(source, key)
	if !ok {
		return "", false, fmt.Errorf("%s cannot be safely deleted", key.String())
	}
	return applyEdits(source, []edit{{start: start, end: end, text: ""}}), true, nil
}

func explicitEmptyTable(assignments []assignment, sections []section, key parser.Key) bool {
	found := false
	for _, sec := range sections {
		if sec.key.Equals(key) {
			if sec.inline {
				return false
			}
			found = true
			continue
		}
		if hasKeyPrefix(sec.key, key) {
			return false
		}
	}
	if !found {
		return false
	}
	for _, item := range assignments {
		if item.key.Equals(key) || hasKeyPrefix(item.key, key) {
			return false
		}
	}
	return true
}

func indexedRecordDelete(key parser.Key) bool {
	_, _, ok := indexedRecordKey(key)
	return ok
}

func verifySemanticPatch(source, updated string, mutate func(map[string]string) error) (string, error) {
	expected, err := semanticValues(source)
	if err != nil {
		return "", err
	}
	if err := mutate(expected); err != nil {
		return "", err
	}
	got, err := semanticValues(updated)
	if err != nil {
		return "", err
	}
	if !sameSemanticValues(expected, got) {
		return "", fmt.Errorf("internal TOML patch verification failed")
	}
	return updated, nil
}

func semanticValues(source string) (map[string]string, error) {
	if err := parseTOMLSource(source); err != nil {
		return nil, err
	}
	entries, err := listValues(source, nil)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(entries))
	for _, entry := range entries {
		out[entry.Key] = entry.Value
	}
	return out, nil
}

func sameSemanticValues(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		if b[key] != value {
			return false
		}
	}
	return true
}

func applyEdit(source string, e edit) string {
	return source[:e.start] + e.text + source[e.end:]
}

func applyEdits(source string, edits []edit) string {
	sort.SliceStable(edits, func(i, j int) bool {
		return edits[i].start > edits[j].start
	})
	out := source
	for _, e := range edits {
		out = applyEdit(out, e)
	}
	return out
}
