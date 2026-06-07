package jsondoc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type object []member

type member struct {
	key   string
	value any
}

func List(source, path string) ([]Entry, error) {
	root, err := parseJSON(source)
	if err != nil {
		return nil, err
	}
	prefix, err := parsePath(path)
	if err != nil {
		return nil, err
	}
	value, err := resolvePath(root, prefix)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	if err := collectEntries(value, prefix, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func parseJSON(source string) (any, error) {
	decoder := json.NewDecoder(strings.NewReader(source))
	decoder.UseNumber()
	value, err := parseJSONValue(decoder)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("invalid JSON: multiple top-level values")
	}
	return value, nil
}

func parseJSONValue(decoder *json.Decoder) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delimiter, ok := token.(json.Delim); ok {
		switch delimiter {
		case '{':
			return parseJSONObject(decoder)
		case '[':
			return parseJSONArray(decoder)
		default:
			return nil, fmt.Errorf("unexpected delimiter %q", delimiter)
		}
	}
	return token, nil
}

func parseJSONObject(decoder *json.Decoder) (object, error) {
	var out object
	for decoder.More() {
		key, err := parseJSONObjectKey(decoder)
		if err != nil {
			return nil, err
		}
		value, err := parseJSONValue(decoder)
		if err != nil {
			return nil, err
		}
		out = append(out, member{key: key, value: value})
	}
	if err := expectJSONDelimiter(decoder, '}'); err != nil {
		return nil, err
	}
	return out, nil
}

func parseJSONObjectKey(decoder *json.Decoder) (string, error) {
	token, err := decoder.Token()
	if err != nil {
		return "", err
	}
	key, ok := token.(string)
	if !ok {
		return "", fmt.Errorf("object key is not a string")
	}
	return key, nil
}

func parseJSONArray(decoder *json.Decoder) ([]any, error) {
	var out []any
	for decoder.More() {
		value, err := parseJSONValue(decoder)
		if err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	if err := expectJSONDelimiter(decoder, ']'); err != nil {
		return nil, err
	}
	return out, nil
}

func expectJSONDelimiter(decoder *json.Decoder, want json.Delim) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token != want {
		return fmt.Errorf("expected delimiter %q", want)
	}
	return nil
}

func resolvePath(value any, path []string) (any, error) {
	current := value
	for i, segment := range path {
		switch node := current.(type) {
		case map[string]any:
			next, ok := node[segment]
			if !ok {
				return nil, fmt.Errorf("%s is not set", formatPath(path[:i+1]))
			}
			current = next
		case object:
			next, ok := objectValue(node, segment)
			if !ok {
				return nil, fmt.Errorf("%s is not set", formatPath(path[:i+1]))
			}
			current = next
		case []any:
			index, err := strconv.Atoi(segment)
			if err != nil || index < 0 || index >= len(node) {
				return nil, fmt.Errorf("%s is not set", formatPath(path[:i+1]))
			}
			current = node[index]
		default:
			return nil, fmt.Errorf("%s is not set", formatPath(path[:i+1]))
		}
	}
	return current, nil
}

func collectEntries(value any, path []string, entries *[]Entry) error {
	switch node := value.(type) {
	case object:
		for _, item := range node {
			if err := collectEntries(item.value, appendPath(path, item.key), entries); err != nil {
				return err
			}
		}
	case []any:
		for i, child := range node {
			if err := collectEntries(child, appendPath(path, strconv.Itoa(i)), entries); err != nil {
				return err
			}
		}
	default:
		if len(path) == 0 {
			return fmt.Errorf("document root is a scalar, not a collection")
		}
		*entries = append(*entries, Entry{Key: formatPath(path), Value: formatScalar(node)})
	}
	return nil
}

func parsePath(path string) ([]string, error) {
	if path == "" {
		return nil, nil
	}
	var parts []string
	var current strings.Builder
	for i := 0; i < len(path); i++ {
		if path[i] != '.' {
			current.WriteByte(path[i])
			continue
		}
		if i+1 < len(path) && path[i+1] == '.' {
			current.WriteByte('.')
			i++
			continue
		}
		if current.Len() == 0 {
			return nil, fmt.Errorf("empty path segment in %q", path)
		}
		parts = append(parts, current.String())
		current.Reset()
	}
	if current.Len() == 0 {
		return nil, fmt.Errorf("empty path segment in %q", path)
	}
	parts = append(parts, current.String())
	return parts, nil
}

func appendPath(path []string, segment string) []string {
	next := make([]string, len(path), len(path)+1)
	copy(next, path)
	next = append(next, segment)
	return next
}

func formatPath(path []string) string {
	parts := make([]string, len(path))
	for i, segment := range path {
		parts[i] = strings.ReplaceAll(segment, ".", "..")
	}
	return strings.Join(parts, ".")
}

func formatScalar(value any) string {
	switch v := value.(type) {
	case nil:
		return "null"
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case json.Number:
		return v.String()
	default:
		var out bytes.Buffer
		encoder := json.NewEncoder(&out)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(v); err != nil {
			return fmt.Sprint(v)
		}
		return strings.TrimSpace(out.String())
	}
}

func objectValue(value object, key string) (any, bool) {
	for _, item := range value {
		if item.key == key {
			return item.value, true
		}
	}
	return nil, false
}
