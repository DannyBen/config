package jsondoc

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func Delete(source, path string, rawSelectors []string) (string, error) {
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
	if len(rawSelectors) > 0 {
		if err := deleteSelectedRecord(data, key, rawSelectors); err != nil {
			return "", err
		}
	} else {
		updated, err := deletePath(data, key, false)
		if err != nil {
			return "", err
		}
		data = updated
	}
	return marshalJSON(data)
}

func DeleteIfEmpty(source, path string) (string, error) {
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
	if _, err := resolvePath(data, key); err != nil {
		return source, nil
	}
	updated, err := deletePath(data, key, true)
	if err != nil {
		if isNotEmptyError(err, key) {
			return source, nil
		}
		return "", err
	}
	data = updated
	return marshalJSON(data)
}

func deleteSelectedRecord(root any, collectionPath []string, rawSelectors []string) error {
	collectionValue, err := resolvePath(root, collectionPath)
	if err != nil {
		return err
	}
	records, ok := collectionValue.([]any)
	if !ok || !isRecordArray(records) {
		return fmt.Errorf("%s is not an array of records", formatPath(collectionPath))
	}
	selectors, err := parseSelectors(rawSelectors)
	if err != nil {
		return err
	}
	index, err := selectedRecordIndex(records, collectionPath, selectors)
	if err != nil {
		return err
	}
	records = append(records[:index], records[index+1:]...)
	_, err = setPathValue(root, collectionPath, records, true)
	return err
}

func deletePath(root any, path []string, onlyIfEmpty bool) (any, error) {
	return deletePathAt(root, path, nil, onlyIfEmpty)
}

func deletePathAt(root any, path, base []string, onlyIfEmpty bool) (any, error) {
	switch node := root.(type) {
	case map[string]any:
		return deleteObjectPath(node, path, base, onlyIfEmpty)
	case []any:
		return deleteArrayPath(node, path, base, onlyIfEmpty)
	default:
		return nil, fmt.Errorf("%s is not set", formatPath(appendPaths(base, path)))
	}
}

func deleteObjectPath(node map[string]any, path, base []string, onlyIfEmpty bool) (any, error) {
	fullPath := appendPath(base, path[0])
	if len(path) == 1 {
		value, ok := node[path[0]]
		if !ok {
			return nil, fmt.Errorf("%s is not set", formatPath(fullPath))
		}
		if !isContainer(value) {
			return nil, fmt.Errorf("%s is a value, use unset to remove fields", formatPath(fullPath))
		}
		if onlyIfEmpty && !emptyContainer(value) {
			return nil, fmt.Errorf("%s is not empty", formatPath(fullPath))
		}
		delete(node, path[0])
		return node, nil
	}
	child, ok := node[path[0]]
	if !ok {
		return nil, fmt.Errorf("%s is not set", formatPath(fullPath))
	}
	updated, err := deletePathAt(child, path[1:], fullPath, onlyIfEmpty)
	if err != nil {
		return nil, err
	}
	node[path[0]] = updated
	return node, nil
}

func deleteArrayPath(node []any, path, base []string, onlyIfEmpty bool) (any, error) {
	index, err := strconv.Atoi(path[0])
	if err != nil || index < 0 || index >= len(node) {
		return nil, fmt.Errorf("%s has no record at index %s", formatPath(base), path[0])
	}
	fullPath := appendPath(base, path[0])
	if len(path) == 1 {
		if !isContainer(node[index]) {
			return nil, fmt.Errorf("%s is a value, use unset to remove fields", formatPath(fullPath))
		}
		if onlyIfEmpty && !emptyContainer(node[index]) {
			return nil, fmt.Errorf("%s is not empty", formatPath(fullPath))
		}
		return append(node[:index], node[index+1:]...), nil
	}
	updated, err := deletePathAt(node[index], path[1:], fullPath, onlyIfEmpty)
	if err != nil {
		return nil, err
	}
	node[index] = updated
	return node, nil
}

func emptyContainer(value any) bool {
	switch node := value.(type) {
	case map[string]any:
		return len(node) == 0
	case []any:
		return len(node) == 0
	default:
		return false
	}
}

func isNotEmptyError(err error, path []string) bool {
	return err.Error() == fmt.Sprintf("%s is not empty", formatPath(path))
}

func appendPaths(base, suffix []string) []string {
	out := make([]string, len(base), len(base)+len(suffix))
	copy(out, base)
	out = append(out, suffix...)
	return out
}
