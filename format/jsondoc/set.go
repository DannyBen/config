package jsondoc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func Set(source, path, rawValue string) (string, error) {
	return setValue(source, path, rawValue, false)
}

func SetString(source, path, rawValue string) (string, error) {
	return setValue(source, path, rawValue, true)
}

func setValue(source, path, rawValue string, forceString bool) (string, error) {
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
	value, err := parseSetValue(rawValue, forceString)
	if err != nil {
		return "", err
	}
	updated, err := setPath(data, key, value)
	if err != nil {
		return "", err
	}
	return marshalJSON(updated)
}

func parseSetValue(raw string, forceString bool) (any, error) {
	if forceString {
		return raw, nil
	}
	switch raw {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "null":
		return nil, nil
	}
	if number, ok := parseJSONNumber(raw); ok {
		return number, nil
	}
	return raw, nil
}

func parseJSONNumber(raw string) (any, bool) {
	if raw == "" {
		return nil, false
	}
	if integer, err := strconv.ParseInt(raw, 10, 64); err == nil && strconv.FormatInt(integer, 10) == raw {
		return integer, true
	}
	var parsed any
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&parsed); err != nil {
		return nil, false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, false
	}
	number, ok := parsed.(json.Number)
	if !ok {
		return nil, false
	}
	if _, err := strconv.ParseFloat(number.String(), 64); err != nil {
		return nil, false
	}
	return number, true
}

func setPath(root any, path []string, value any) (any, error) {
	if len(path) == 0 {
		return value, nil
	}
	switch node := root.(type) {
	case map[string]any:
		return setObjectPath(node, path, value)
	case []any:
		return setArrayPath(node, path, value)
	default:
		return nil, fmt.Errorf("%s parent is not set", formatPath(path[:len(path)-1]))
	}
}

func setObjectPath(node map[string]any, path []string, value any) (any, error) {
	if len(path) == 1 {
		if existing, ok := node[path[0]]; ok && isContainer(existing) {
			return nil, fmt.Errorf("%s is a container, not a scalar value", formatPath(path))
		}
		node[path[0]] = value
		return node, nil
	}
	child, ok := node[path[0]]
	if !ok {
		child = map[string]any{}
	}
	updated, err := setPath(child, path[1:], value)
	if err != nil {
		return nil, err
	}
	node[path[0]] = updated
	return node, nil
}

func setArrayPath(node []any, path []string, value any) (any, error) {
	index, err := strconv.Atoi(path[0])
	if err != nil || index < 0 || index >= len(node) {
		return nil, fmt.Errorf("%s is not set", formatPath(path[:1]))
	}
	if len(path) == 1 {
		if isContainer(node[index]) {
			return nil, fmt.Errorf("%s is a container, not a scalar value", formatPath(path))
		}
		node[index] = value
		return node, nil
	}
	updated, err := setPath(node[index], path[1:], value)
	if err != nil {
		return nil, err
	}
	node[index] = updated
	return node, nil
}

func marshalJSON(value any) (string, error) {
	out, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out) + "\n", nil
}
