package jsondoc

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func Unset(source, path string) (string, error) {
	var data any
	if err := json.Unmarshal([]byte(source), &data); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	if len(key) == 0 {
		return "", fmt.Errorf("empty path")
	}
	if err := unsetPath(data, key); err != nil {
		return "", err
	}
	return marshalJSON(data)
}

func UnsetIn(source, collectionPath string, rawSelectors []string, path string) (string, error) {
	var data any
	if err := json.Unmarshal([]byte(source), &data); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
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
	collectionValue, err := resolvePath(data, collection)
	if err != nil {
		return "", err
	}
	records, ok := collectionValue.([]any)
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
	record, ok := records[index].(map[string]any)
	if !ok {
		return "", fmt.Errorf("%s.%d is not a record", formatPath(collection), index)
	}
	if err := unsetPath(record, key); err != nil {
		fullPath := appendPath(collection, fmt.Sprintf("%d", index))
		fullPath = append(fullPath, key...)
		return "", fmt.Errorf("%s is not set", formatPath(fullPath))
	}
	return marshalJSON(data)
}

func unsetPath(root any, path []string) error {
	switch node := root.(type) {
	case map[string]any:
		return unsetObjectPath(node, path)
	case []any:
		return unsetArrayPath(node, path)
	default:
		return fmt.Errorf("%s is not set", formatPath(path))
	}
}

func unsetObjectPath(node map[string]any, path []string) error {
	if len(path) == 1 {
		value, ok := node[path[0]]
		if !ok {
			return fmt.Errorf("%s is not set", formatPath(path))
		}
		if isContainer(value) {
			return fmt.Errorf("%s is a container, not a scalar value", formatPath(path))
		}
		delete(node, path[0])
		return nil
	}
	child, ok := node[path[0]]
	if !ok {
		return fmt.Errorf("%s is not set", formatPath(path[:1]))
	}
	return unsetPath(child, path[1:])
}

func unsetArrayPath(node []any, path []string) error {
	index, err := strconv.Atoi(path[0])
	if err != nil || index < 0 || index >= len(node) {
		return fmt.Errorf("%s is not set", formatPath(path[:1]))
	}
	if len(path) == 1 {
		if isContainer(node[index]) {
			return fmt.Errorf("%s is a container, not a scalar value", formatPath(path))
		}
		node[index] = nil
		return nil
	}
	return unsetPath(node[index], path[1:])
}
