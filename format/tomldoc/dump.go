package tomldoc

import (
	"fmt"
	"strconv"

	pelletiertoml "github.com/pelletier/go-toml/v2"
)

func Dump(source, path string) (any, error) {
	var data any
	if err := pelletiertoml.Unmarshal([]byte(source), &data); err != nil {
		return nil, fmt.Errorf("invalid TOML: %w", err)
	}

	selected := data
	if path != "" {
		key, err := parsePath(path)
		if err != nil {
			return nil, err
		}
		selected, err = selectDumpValue(data, key)
		if err != nil {
			return nil, err
		}
	}

	return selected, nil
}

func selectDumpValue(value any, key []string) (any, error) {
	current := value
	for i, segment := range key {
		switch node := current.(type) {
		case map[string]any:
			next, ok := node[segment]
			if !ok {
				return nil, fmt.Errorf("%s is not set", formatListKey(key[:i+1]))
			}
			current = next
		case []any:
			index, err := strconv.Atoi(segment)
			if err != nil || index < 0 || index >= len(node) {
				return nil, fmt.Errorf("%s is not set", formatListKey(key[:i+1]))
			}
			current = node[index]
		default:
			return nil, fmt.Errorf("%s is not set", formatListKey(key[:i+1]))
		}
	}
	return current, nil
}
