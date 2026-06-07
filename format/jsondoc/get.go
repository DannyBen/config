package jsondoc

import (
	"fmt"
	"strings"
)

type selector struct {
	path    []string
	compare string
	label   string
}

func Get(source, path string) (string, error) {
	root, err := parseJSON(source)
	if err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	if len(key) == 0 {
		return "", fmt.Errorf("empty path")
	}
	value, err := resolvePath(root, key)
	if err != nil {
		return "", err
	}
	return renderGetValue(value, key)
}

func GetIn(source, collectionPath string, rawSelectors []string, path string) (string, error) {
	root, err := parseJSON(source)
	if err != nil {
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
	if len(collection) == 0 {
		return "", fmt.Errorf("--in: empty path")
	}
	if len(key) == 0 {
		return "", fmt.Errorf("empty path")
	}

	value, err := resolvePath(root, collection)
	if err != nil {
		return "", err
	}
	records, ok := value.([]any)
	if !ok || !isRecordArray(records) {
		return "", fmt.Errorf("%s is not an array of records", formatPath(collection))
	}

	selectors, err := parseSelectors(rawSelectors)
	if err != nil {
		return "", err
	}
	index, err := selectedRecordIndex(records, collection, selectors)
	if err != nil {
		return "", err
	}
	record, ok := records[index].(object)
	if !ok {
		return "", fmt.Errorf("%s.%d is not a record", formatPath(collection), index)
	}
	value, err = resolvePath(record, key)
	fullPath := appendPath(collection, fmt.Sprintf("%d", index))
	fullPath = append(fullPath, key...)
	if err != nil {
		return "", fmt.Errorf("%s is not set", formatPath(fullPath))
	}
	return renderGetValue(value, fullPath)
}

func renderGetValue(value any, path []string) (string, error) {
	switch node := value.(type) {
	case object:
		return "", fmt.Errorf("%s is an object, not a value", formatPath(path))
	case []any:
		if isRecordArray(node) {
			return "", fmt.Errorf("%s is an array of records, not a value", formatPath(path))
		}
		return flowArray(node), nil
	default:
		return formatScalar(node), nil
	}
}

func isRecordArray(values []any) bool {
	for _, value := range values {
		if _, ok := value.(object); ok {
			return true
		}
	}
	return false
}

func flowArray(values []any) string {
	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, formatScalar(value))
	}
	return "[" + strings.Join(items, ", ") + "]"
}

func parseSelectors(raw []string) ([]selector, error) {
	selectors := make([]selector, 0, len(raw))
	for _, item := range raw {
		onPath, onRaw, ok := cutSelector(item)
		if !ok || onPath == "" {
			return nil, fmt.Errorf("--on must use FIELD:VALUE")
		}
		path, err := parsePath(onPath)
		if err != nil {
			return nil, fmt.Errorf("--on: %w", err)
		}
		selectors = append(selectors, selector{
			path:    path,
			compare: onRaw,
			label:   formatPath(path) + ":" + onRaw,
		})
	}
	return selectors, nil
}

func cutSelector(value string) (string, string, bool) {
	for i := 0; i < len(value); i++ {
		if value[i] == ':' {
			return value[:i], value[i+1:], true
		}
	}
	return "", "", false
}

func selectedRecordIndex(records []any, collectionPath []string, selectors []selector) (int, error) {
	var matches []int
	for i, value := range records {
		record, ok := value.(object)
		if !ok {
			return 0, fmt.Errorf("%s.%d is not a record", formatPath(collectionPath), i)
		}
		matched, err := recordMatches(record, selectors)
		if err != nil {
			return 0, err
		}
		if matched {
			matches = append(matches, i)
		}
	}

	label := joinSelectorLabels(selectors)
	switch len(matches) {
	case 0:
		return 0, fmt.Errorf("%s has no records matching %s", formatPath(collectionPath), label)
	case 1:
		return matches[0], nil
	default:
		return 0, fmt.Errorf("%s has multiple records matching %s", formatPath(collectionPath), label)
	}
}

func recordMatches(record object, selectors []selector) (bool, error) {
	for _, selector := range selectors {
		value, err := resolvePath(record, selector.path)
		if err != nil {
			return false, nil
		}
		if isContainer(value) {
			return false, nil
		}
		if formatScalar(value) != selector.compare {
			return false, nil
		}
	}
	return true, nil
}

func isContainer(value any) bool {
	switch value.(type) {
	case map[string]any, object, []any:
		return true
	default:
		return false
	}
}

func joinSelectorLabels(selectors []selector) string {
	labels := make([]string, 0, len(selectors))
	for _, selector := range selectors {
		labels = append(labels, selector.label)
	}
	return strings.Join(labels, ", ")
}
