package tomldoc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/creachadair/tomledit/parser"
)

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
		if sec.key.Equals(key) && sec.inline {
			return nil, fmt.Errorf("%s is an inline table; delete is not implemented for inline tables yet", key.String())
		}
	}
	if planned, ok := tablePrefixDeleteRanges(source, key); ok {
		return planned, nil
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

func tablePrefixDeleteRanges(source string, key parser.Key) ([]edit, bool) {
	var planned []edit
	for _, stmt := range statements(source) {
		if len(stmt.tokens) == 0 {
			continue
		}
		table, ok := tableKey(stmt.tokens)
		if !ok || (!table.Equals(key) && !hasKeyPrefix(table, key)) {
			continue
		}
		start, end, ok := tableDeleteRange(source, table)
		if !ok {
			return nil, false
		}
		planned = append(planned, edit{start: start, end: end, text: ""})
	}
	return planned, len(planned) > 0
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
