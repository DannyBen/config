package tomldoc

import (
	"bytes"
	"fmt"
	"strconv"

	pelletiertoml "github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

func Dump(source, path string) (string, error) {
	var data any
	if err := pelletiertoml.Unmarshal([]byte(source), &data); err != nil {
		return "", fmt.Errorf("invalid TOML: %w", err)
	}

	selected := data
	if path != "" {
		key, err := parsePath(path)
		if err != nil {
			return "", err
		}
		selected, err = selectDumpValue(data, key)
		if err != nil {
			return "", err
		}
	}

	out, err := marshalYAMLDump(selected)
	if err != nil {
		return "", err
	}
	return out, nil
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

func marshalYAMLDump(value any) (string, error) {
	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(2)
	if err := encoder.Encode(value); err != nil {
		encoder.Close()
		return "", err
	}
	if err := encoder.Close(); err != nil {
		return "", err
	}
	return out.String(), nil
}
