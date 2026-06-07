package jsondoc

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func Dump(source, path string) (any, error) {
	var data any
	decoder := json.NewDecoder(strings.NewReader(source))
	decoder.UseNumber()
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	data, err := normalizeDumpValue(data)
	if err != nil {
		return nil, err
	}

	if path == "" {
		return data, nil
	}
	key, err := parsePath(path)
	if err != nil {
		return nil, err
	}
	return resolvePath(data, key)
}

func normalizeDumpValue(value any) (any, error) {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			normalized, err := normalizeDumpValue(item)
			if err != nil {
				return nil, err
			}
			out[key] = normalized
		}
		return out, nil
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			normalized, err := normalizeDumpValue(item)
			if err != nil {
				return nil, err
			}
			out[i] = normalized
		}
		return out, nil
	case json.Number:
		if integer, err := strconv.ParseInt(v.String(), 10, 64); err == nil {
			return integer, nil
		}
		floating, err := strconv.ParseFloat(v.String(), 64)
		if err != nil {
			return nil, err
		}
		return floating, nil
	default:
		return value, nil
	}
}
