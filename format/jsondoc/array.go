package jsondoc

import "fmt"

func SetArray(source, path string, values []string) (string, error) {
	var items []any
	for _, value := range values {
		parsed, err := parseSetValue(value, false)
		if err != nil {
			return "", err
		}
		items = append(items, parsed)
	}
	return setValueTo(source, path, items)
}

func ArrayAdd(source, path string, values []string) (string, error) {
	items, ok, err := jsonArrayValues(source, path)
	if err != nil {
		return "", err
	}
	if !ok {
		return SetArray(source, path, values)
	}

	seen := make(map[string]bool, len(items)+len(values))
	next := make([]any, 0, len(items)+len(values))
	for _, item := range items {
		seen[formatScalar(item)] = true
		next = append(next, item)
	}
	for _, value := range values {
		parsed, err := parseSetValue(value, false)
		if err != nil {
			return "", err
		}
		logical := formatScalar(parsed)
		if seen[logical] {
			continue
		}
		seen[logical] = true
		next = append(next, parsed)
	}
	return setValueTo(source, path, next)
}

func ArrayDel(source, path string, values []string) (string, error) {
	items, ok, err := jsonArrayValues(source, path)
	if err != nil {
		return "", err
	}
	if !ok {
		return source, nil
	}

	remove := make(map[string]bool, len(values))
	for _, value := range values {
		parsed, err := parseSetValue(value, false)
		if err != nil {
			return "", err
		}
		remove[formatScalar(parsed)] = true
	}
	next := make([]any, 0, len(items))
	for _, item := range items {
		if !remove[formatScalar(item)] {
			next = append(next, item)
		}
	}
	if len(next) == 0 {
		return Delete(source, path, nil)
	}
	return setValueTo(source, path, next)
}

func jsonArrayValues(source, path string) ([]any, bool, error) {
	data, err := parseMutableJSON(source)
	if err != nil {
		return nil, false, err
	}
	key, err := parsePath(path)
	if err != nil {
		return nil, false, err
	}
	if len(key) == 0 {
		return nil, false, fmt.Errorf("empty path")
	}
	value, err := resolvePath(data, key)
	if err != nil {
		if isNotSetJSONError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	items, ok := value.([]any)
	if !ok {
		return nil, true, fmt.Errorf("%s is not an array", formatPath(key))
	}
	for _, item := range items {
		if isContainer(item) {
			return nil, true, fmt.Errorf("%s is not a scalar array", formatPath(key))
		}
	}
	return items, true, nil
}

func setValueTo(source, path string, value any) (string, error) {
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
	updated, err := setPathValue(data, key, value, true)
	if err != nil {
		return "", err
	}
	return marshalJSON(updated)
}

func isNotSetJSONError(err error) bool {
	return err != nil && len(err.Error()) >= len(" is not set") && err.Error()[len(err.Error())-len(" is not set"):] == " is not set"
}
