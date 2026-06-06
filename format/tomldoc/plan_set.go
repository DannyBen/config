package tomldoc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/creachadair/tomledit/parser"
)

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
	if insertAt, scope, ok := dottedSiblingInsertAt(assignments, key); ok {
		relative := relativeKey(key, scope)
		text := relative.String() + " = " + value + lineEnding(source)
		return edit{start: insertAt, end: insertAt, text: prefixLineEndingAt(source, insertAt) + text}, nil
	}
	if insertAt, ok := dottedParentInsertAt(assignments, parent); ok {
		text := key.String() + " = " + value + lineEnding(source)
		return edit{start: insertAt, end: insertAt, text: prefixLineEndingAt(source, insertAt) + text}, nil
	}
	if hasSiblingSection(sections, parent) {
		return insertTable(source, sections, parent, key[len(key)-1], value), nil
	}
	if sec, prefix, ok := longestSectionPrefix(sections, key); ok {
		child := key[len(prefix):]
		if len(child) > 1 && !sec.inline && !hasAssignmentWithPrefix(assignments, parent) {
			return insertTable(source, sections, parent, key[len(key)-1], value), nil
		}
		if sec.inline {
			return insertInlineTableChild(source, sec, child.String(), value), nil
		}
		text := child.String() + " = " + value + lineEnding(source)
		return edit{start: sec.insertAt, end: sec.insertAt, text: prefixLineEndingAt(source, sec.insertAt) + text}, nil
	}
	return insertTable(source, sections, parent, key[len(key)-1], value), nil
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

func dottedSiblingInsertAt(assignments []assignment, key parser.Key) (int, parser.Key, bool) {
	insertAt := -1
	var scope parser.Key
	for _, item := range assignments {
		if item.key.Equals(key) {
			continue
		}
		if !dottedSiblingStyleApplies(item, key) {
			continue
		}
		if item.lineSpan.End > insertAt {
			insertAt = item.lineSpan.End
			scope = item.scope
		}
	}
	return insertAt, scope, insertAt >= 0
}

func dottedSiblingStyleApplies(item assignment, key parser.Key) bool {
	if item.internal {
		return false
	}
	if len(item.scope) > 0 && !item.scope.IsPrefixOf(key) {
		return false
	}
	relative := relativeKey(key, item.scope)
	if len(relative) != 2 {
		return false
	}
	if len(item.scope) > 0 && !item.scope.IsPrefixOf(item.key) {
		return false
	}
	itemRelative := relativeKey(item.key, item.scope)
	return len(itemRelative) > 1
}

func dottedParentInsertAt(assignments []assignment, parent parser.Key) (int, bool) {
	insertAt := -1
	for _, item := range assignments {
		if hasKeyPrefix(item.key, parent) && item.lineSpan.End > insertAt {
			insertAt = item.lineSpan.End
		}
	}
	return insertAt, insertAt >= 0
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

func insertTable(source string, sections []section, parent parser.Key, leaf string, value string) edit {
	nl := lineEnding(source)
	text := "[" + parent.String() + "]" + nl + parser.Key{leaf}.String() + " = " + value + nl
	insertAt, ok := familyTableInsertAt(source, sections, parent)
	if !ok {
		return edit{start: len(source), end: len(source), text: blockPrefix(source) + text}
	}
	if insertAt < len(source) && !hasLineEndingPrefix(source[insertAt:]) {
		text += nl
	}
	return edit{start: insertAt, end: insertAt, text: tableInsertPrefix(source, insertAt) + text}
}

func familyTableInsertAt(source string, sections []section, key parser.Key) (int, bool) {
	if len(key) == 0 {
		return 0, false
	}
	keyText := key.String()
	var before section
	beforeText := ""
	foundBefore := false
	var after section
	foundAfter := false
	for _, sec := range sections {
		if sec.inline || len(sec.key) == 0 || sec.key[0] != key[0] {
			continue
		}
		secText := sec.key.String()
		if secText <= keyText {
			if !foundBefore || secText > beforeText {
				before = sec
				beforeText = secText
				foundBefore = true
			}
			continue
		}
		if !foundAfter || sec.start < after.start {
			after = sec
			foundAfter = true
		}
	}
	if foundBefore {
		return before.insertAt, true
	}
	if foundAfter {
		return insertAtBeforeBlankLine(source, after.start), true
	}
	return 0, false
}

func tableInsertPrefix(source string, insertAt int) string {
	nl := lineEnding(source)
	if insertAt <= 0 {
		return ""
	}
	if strings.HasSuffix(source[:insertAt], nl+nl) {
		return ""
	}
	if strings.HasSuffix(source[:insertAt], nl) {
		return nl
	}
	return nl + nl
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
