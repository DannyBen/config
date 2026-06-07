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

func SetIn(source, collectionPath, rawSelector, path, rawValue string) (string, error) {
	return setValueIn(source, collectionPath, rawSelector, path, rawValue, false)
}

func SetInString(source, collectionPath, rawSelector, path, rawValue string) (string, error) {
	return setValueIn(source, collectionPath, rawSelector, path, rawValue, true)
}

func setValue(source, path, rawValue string, forceString bool) (string, error) {
	data, err := parseMutableJSON(source)
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

func setValueIn(source, collectionPath, rawSelector, path, rawValue string, forceString bool) (string, error) {
	data, err := parseMutableJSON(source)
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
	selector, err := parseSetSelector(rawSelector)
	if err != nil {
		return "", err
	}
	value, err := parseSetValue(rawValue, forceString)
	if err != nil {
		return "", err
	}
	updated, err := setRecordPath(data, collection, selector, key, value)
	if err != nil {
		return "", err
	}
	return marshalJSON(updated)
}

func parseSetSelector(raw string) (selector, error) {
	selectors, err := parseSelectors([]string{raw})
	if err != nil {
		return selector{}, err
	}
	return selectors[0], nil
}

func setRecordPath(root any, collectionPath []string, match selector, key []string, value any) (any, error) {
	collectionValue, err := resolvePath(root, collectionPath)
	if err != nil {
		return nil, err
	}
	records, ok := collectionValue.([]any)
	if !ok || !isRecordArray(records) {
		return nil, fmt.Errorf("%s is not an array of records", formatPath(collectionPath))
	}
	index, err := selectedRecordIndex(records, collectionPath, []selector{match})
	if err != nil {
		if !recordNotFound(err, collectionPath, match) {
			return nil, err
		}
		record := map[string]any{}
		if err := seedRecordSelector(record, match); err != nil {
			return nil, err
		}
		records = append(records, record)
		index = len(records) - 1
		root, err = setPathValue(root, collectionPath, records, true)
		if err != nil {
			return nil, err
		}
	}
	record, ok := records[index].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s.%d is not a record", formatPath(collectionPath), index)
	}
	updated, err := setPath(record, key, value)
	if err != nil {
		return nil, err
	}
	records[index] = updated
	return root, nil
}

func recordNotFound(err error, collectionPath []string, match selector) bool {
	want := fmt.Sprintf("%s has no records matching %s", formatPath(collectionPath), match.label)
	return err.Error() == want
}

func seedRecordSelector(record map[string]any, match selector) error {
	value, err := parseSetValue(match.compare, false)
	if err != nil {
		return err
	}
	_, err = setPath(record, match.path, value)
	return err
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

func parseMutableJSON(source string) (any, error) {
	var data any
	if err := json.Unmarshal([]byte(source), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return data, nil
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
	return setPathValue(root, path, value, false)
}

func setPathValue(root any, path []string, value any, allowContainer bool) (any, error) {
	if len(path) == 0 {
		return value, nil
	}
	switch node := root.(type) {
	case map[string]any:
		return setObjectPath(node, path, value, allowContainer)
	case []any:
		return setArrayPath(node, path, value, allowContainer)
	default:
		return nil, fmt.Errorf("%s parent is not set", formatPath(path[:len(path)-1]))
	}
}

func setObjectPath(node map[string]any, path []string, value any, allowContainer bool) (any, error) {
	if len(path) == 1 {
		if existing, ok := node[path[0]]; ok && isContainer(existing) && !allowContainer {
			return nil, fmt.Errorf("%s is a container, not a scalar value", formatPath(path))
		}
		node[path[0]] = value
		return node, nil
	}
	child, ok := node[path[0]]
	if !ok {
		child = map[string]any{}
	}
	updated, err := setPathValue(child, path[1:], value, allowContainer)
	if err != nil {
		return nil, err
	}
	node[path[0]] = updated
	return node, nil
}

func setArrayPath(node []any, path []string, value any, allowContainer bool) (any, error) {
	index, err := strconv.Atoi(path[0])
	if err != nil || index < 0 || index >= len(node) {
		return nil, fmt.Errorf("%s is not set", formatPath(path[:1]))
	}
	if len(path) == 1 {
		if isContainer(node[index]) && !allowContainer {
			return nil, fmt.Errorf("%s is a container, not a scalar value", formatPath(path))
		}
		node[index] = value
		return node, nil
	}
	updated, err := setPathValue(node[index], path[1:], value, allowContainer)
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
