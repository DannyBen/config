package tomldoc

import (
	"fmt"
	"strings"

	"github.com/creachadair/tomledit/parser"
	"github.com/creachadair/tomledit/scanner"
	"github.com/neongreen/tomlsawyer"
	pelletiertoml "github.com/pelletier/go-toml/v2"
)

type token struct {
	kind scanner.Token
	text string
	span scanner.Span
}

type assignment struct {
	key       parser.Key
	scope     parser.Key
	lineSpan  scanner.Span
	valueSpan scanner.Span
	internal  bool
}

type section struct {
	key      parser.Key
	start    int
	insertAt int
	inline   bool
	empty    bool
}

type edit struct {
	start int
	end   int
	text  string
}

type semanticSet struct {
	key   parser.Key
	value string
}

type Entry struct {
	Key   string
	Value string
}

func validateTOML(source string) error {
	var out any
	if err := pelletiertoml.Unmarshal([]byte(source), &out); err != nil {
		return fmt.Errorf("invalid TOML: %w", err)
	}
	return nil
}

func Valid(source string) bool {
	return parseTOMLSource(source) == nil
}

func parseTOMLSource(source string) error {
	if err := validateTOML(source); err != nil {
		return err
	}
	_, err := tomlsawyer.Parse([]byte(source))
	return err
}

func parsePath(path string) (parser.Key, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}
	var key parser.Key
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
		key = append(key, current.String())
		current.Reset()
	}
	if current.Len() == 0 {
		return nil, fmt.Errorf("empty path segment in %q", path)
	}
	key = append(key, current.String())
	return key, nil
}

func Set(source, path, rawValue string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	value, err := formatValue(rawValue)
	if err != nil {
		return "", err
	}
	planned, err := planSet(source, key, value)
	if err != nil {
		return "", err
	}
	return applyVerifiedSet(source, planned, key, value)
}

func SetString(source, path, value string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	formatted, err := formatString(value)
	if err != nil {
		return "", err
	}
	planned, err := planSet(source, key, formatted)
	if err != nil {
		return "", err
	}
	return applyVerifiedSet(source, planned, key, formatted)
}

func SetArray(source, path string, values []string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	formatted, err := formatArray(values)
	if err != nil {
		return "", err
	}
	planned, err := planSet(source, key, formatted)
	if err != nil {
		return "", err
	}
	return applyVerifiedSet(source, planned, key, formatted)
}

func SetIn(source, collectionPath, selector, path, rawValue string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	collection, onKey, onValue, onCompare, key, value, err := parseSetIn(collectionPath, selector, path, rawValue, formatValue)
	if err != nil {
		return "", err
	}
	planned, changes, err := planSetIn(source, collection, onKey, onValue, onCompare, key, value)
	if err != nil {
		return "", err
	}
	return applyVerifiedSets(source, planned, changes)
}

func SetInString(source, collectionPath, selector, path, value string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	collection, onKey, onValue, onCompare, key, formatted, err := parseSetIn(collectionPath, selector, path, value, formatString)
	if err != nil {
		return "", err
	}
	planned, changes, err := planSetIn(source, collection, onKey, onValue, onCompare, key, formatted)
	if err != nil {
		return "", err
	}
	return applyVerifiedSets(source, planned, changes)
}

func SetInArray(source, collectionPath, selector, path string, values []string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	collection, onKey, onValue, onCompare, key, value, err := parseSetIn(collectionPath, selector, path, "", func(string) (string, error) {
		return formatArray(values)
	})
	if err != nil {
		return "", err
	}
	planned, changes, err := planSetIn(source, collection, onKey, onValue, onCompare, key, value)
	if err != nil {
		return "", err
	}
	return applyVerifiedSets(source, planned, changes)
}

func Unset(source, path string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	planned, err := planUnset(source, key)
	if err != nil {
		return "", err
	}
	return applyVerifiedUnset(source, planned, key)
}

func UnsetIn(source, collectionPath string, selectors []string, path string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
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
	assignments, sections, err := scan(source)
	if err != nil {
		return "", err
	}
	index, err := selectedArrayRecordIndex(source, assignments, sections, collection, selectors)
	if err != nil {
		return "", err
	}
	fullKey := arrayRecordKey(collection, index)
	fullKey = append(fullKey, key...)
	planned, err := planUnset(source, fullKey)
	if err != nil {
		return "", err
	}
	return applyVerifiedUnset(source, planned, fullKey)
}

func Delete(source, path string, selectors []string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	planned, err := planDelete(source, key, selectors)
	if err != nil {
		return "", err
	}
	return applyVerifiedDelete(source, planned, key, selectors)
}

func DeleteIfEmpty(source, path string) (string, error) {
	entries, err := List(source, path)
	if err != nil {
		if isNotSetError(err) {
			return source, nil
		}
		return "", err
	}
	if len(entries) > 0 {
		return source, nil
	}
	return Delete(source, path, nil)
}

func isNotSetError(err error) bool {
	return strings.HasSuffix(err.Error(), " is not set")
}
