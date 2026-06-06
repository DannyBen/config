package tomldoc

import (
	"fmt"
	"io"
	"strings"

	"github.com/creachadair/tomledit/parser"
	"github.com/creachadair/tomledit/scanner"
)

type arrayValue struct {
	raw     string
	logical string
}

func ArrayAdd(source, path string, values []string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	items, ok, err := tomlArrayValues(source, key)
	if err != nil {
		return "", err
	}
	if !ok {
		return SetArray(source, path, values)
	}

	seen := make(map[string]bool, len(items)+len(values))
	raws := make([]string, 0, len(items)+len(values))
	for _, item := range items {
		seen[item.logical] = true
		raws = append(raws, item.raw)
	}
	for _, value := range values {
		formatted, logical, err := formattedArrayInput(value)
		if err != nil {
			return "", err
		}
		if seen[logical] {
			continue
		}
		seen[logical] = true
		raws = append(raws, formatted)
	}
	return replaceArray(source, key, raws)
}

func ArrayDel(source, path string, values []string) (string, error) {
	if err := parseTOMLSource(source); err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	items, ok, err := tomlArrayValues(source, key)
	if err != nil {
		return "", err
	}
	if !ok {
		return source, nil
	}

	remove := make(map[string]bool, len(values))
	for _, value := range values {
		_, logical, err := formattedArrayInput(value)
		if err != nil {
			return "", err
		}
		remove[logical] = true
	}
	raws := make([]string, 0, len(items))
	for _, item := range items {
		if !remove[item.logical] {
			raws = append(raws, item.raw)
		}
	}
	if len(raws) == 0 {
		planned, err := planUnset(source, key)
		if err != nil {
			return "", err
		}
		return applyEdit(source, planned), nil
	}
	return replaceArray(source, key, raws)
}

func tomlArrayValues(source string, key parser.Key) ([]arrayValue, bool, error) {
	value, ok, err := tomlValueText(source, key)
	if err != nil || !ok {
		return nil, ok, err
	}
	tokens, err := valueTokens(value)
	if err != nil {
		return nil, true, err
	}
	if len(tokens) < 2 || tokens[0].kind != scanner.LBracket || tokens[len(tokens)-1].kind != scanner.RBracket {
		return nil, true, fmt.Errorf("%s is not an array", key.String())
	}
	segments := arraySegments(tokens)
	items := make([]arrayValue, 0, len(segments))
	for _, segment := range segments {
		if len(segment) == 0 {
			continue
		}
		raw := strings.TrimSpace(value[segment[0].span.Pos:segment[len(segment)-1].span.End])
		parsed, err := parser.ParseValue(raw)
		if err != nil {
			return nil, true, err
		}
		if !isScalarValue(parsed) {
			return nil, true, fmt.Errorf("%s is not a scalar array", key.String())
		}
		logical, err := formatParsedGetValue(parsed)
		if err != nil {
			return nil, true, err
		}
		items = append(items, arrayValue{raw: raw, logical: logical})
	}
	return items, true, nil
}

func tomlValueText(source string, key parser.Key) (string, bool, error) {
	assignments, sections, err := scan(source)
	if err != nil {
		return "", false, err
	}
	for _, item := range assignments {
		if item.internal {
			continue
		}
		if item.key.Equals(key) {
			return strings.TrimSpace(source[item.valueSpan.Pos:item.valueSpan.End]), true, nil
		}
	}
	for _, sec := range sections {
		if sec.key.Equals(key) {
			return "", false, fmt.Errorf("%s is a table, not a value", key.String())
		}
	}
	if hasArrayCollection(sections, key) {
		return "", false, fmt.Errorf("%s is an array of records, not a value", key.String())
	}
	return "", false, nil
}

func valueTokens(value string) ([]token, error) {
	var tokens []token
	s := scanner.New(strings.NewReader(value))
	for {
		err := s.Next()
		if err != nil {
			if err == io.EOF {
				return tokens, nil
			}
			return nil, err
		}
		if s.Token() == scanner.Newline || s.Token() == scanner.Comment {
			continue
		}
		tokens = append(tokens, token{kind: s.Token(), text: string(s.Text()), span: s.Span()})
	}
}

func formattedArrayInput(value string) (string, string, error) {
	formatted, err := formatValue(value)
	if err != nil {
		return "", "", err
	}
	logical, err := formatGetValue(formatted)
	if err != nil {
		return "", "", err
	}
	return formatted, logical, nil
}

func replaceArray(source string, key parser.Key, raws []string) (string, error) {
	formatted := "[" + strings.Join(raws, ", ") + "]"
	if _, err := parser.ParseValue(formatted); err != nil {
		return "", err
	}
	planned, err := planSet(source, key, formatted)
	if err != nil {
		return "", err
	}
	return applyEdit(source, planned), nil
}
