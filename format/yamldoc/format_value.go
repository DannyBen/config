package yamldoc

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func formatSetValue(raw string, mode setMode, indent int, existing *yaml.Node) (string, string, error) {
	if strings.Contains(raw, "\n") {
		return literalBlock(raw, indent), literalBlockLogical(raw), nil
	}
	if mode == setModeString {
		return formatStringValue(raw, existing), raw, nil
	}
	if inferredScalar(raw) {
		return raw, raw, nil
	}
	if strings.HasPrefix(raw, "!") {
		logical, err := validateTaggedScalar(raw)
		if err != nil {
			return "", "", err
		}
		return raw, logical, nil
	}
	if existing != nil {
		switch existing.Style {
		case yaml.SingleQuotedStyle, yaml.DoubleQuotedStyle:
			return formatStringValue(raw, existing), raw, nil
		case yaml.LiteralStyle, yaml.FoldedStyle:
			return literalBlock(raw, indent), raw, nil
		}
	}
	value := formatStringFallback(raw)
	return value, raw, nil
}

func validateTaggedScalar(raw string) (string, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("value: "+raw+"\n"), &doc); err != nil {
		return "", fmt.Errorf("invalid YAML tag value; use --string for literal text starting with !: %w", err)
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode || len(doc.Content[0].Content) < 2 {
		return "", fmt.Errorf("invalid YAML tag value; use --string for literal text starting with !")
	}
	value := doc.Content[0].Content[1]
	if value.Kind != yaml.ScalarNode {
		return "", fmt.Errorf("YAML tag value must be a scalar; use --string for literal text starting with !")
	}
	if !strings.HasPrefix(value.Tag, "!") {
		return "", fmt.Errorf("invalid YAML tag value; use --string for literal text starting with !")
	}
	return scalarValue(value), nil
}

func formatInferredValue(raw string) string {
	if inferredScalar(raw) {
		return raw
	}
	return formatStringFallback(raw)
}

func inferredScalar(raw string) bool {
	switch {
	case raw == "true" || raw == "false" || raw == "null":
		return true
	case integer.MatchString(raw) || float.MatchString(raw):
		return true
	case timestampScalar(raw):
		return true
	}
	return false
}

func timestampScalar(raw string) bool {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("value: "+raw+"\n"), &doc); err != nil {
		return false
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode || len(doc.Content[0].Content) < 2 {
		return false
	}
	value := doc.Content[0].Content[1]
	return value.Kind == yaml.ScalarNode && value.Tag == "!!timestamp" && value.Value == raw
}

func formatStringFallback(raw string) string {
	if plainString(raw) {
		return raw
	}
	return strconv.Quote(raw)
}

func formatStringValue(raw string, existing *yaml.Node) string {
	if existing != nil {
		switch existing.Style {
		case yaml.SingleQuotedStyle:
			return "'" + strings.ReplaceAll(raw, "'", "''") + "'"
		case yaml.DoubleQuotedStyle:
			return strconv.Quote(raw)
		}
	}
	return strconv.Quote(raw)
}

func formatPlainString(raw string) string {
	if plainString(raw) {
		return raw
	}
	return strconv.Quote(raw)
}

func plainString(raw string) bool {
	if strings.TrimSpace(raw) != raw || raw == "" || strings.Contains(raw, "\n") {
		return false
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("value: "+raw+"\n"), &doc); err != nil {
		return false
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode || len(doc.Content[0].Content) < 2 {
		return false
	}
	value := doc.Content[0].Content[1]
	return value.Kind == yaml.ScalarNode && value.Tag == "!!str" && value.Value == raw
}

func formatArray(values []string, indent int, existing *yaml.Node) (string, error) {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, formatInferredValue(value))
	}
	if blockArrayStyle(values, existing) {
		prefix := strings.Repeat(" ", indent)
		lines := make([]string, 0, len(parts))
		for _, part := range parts {
			lines = append(lines, prefix+"- "+part)
		}
		return "\n" + strings.Join(lines, "\n"), nil
	}
	return "[" + strings.Join(parts, ", ") + "]", nil
}

func blockArrayStyle(values []string, existing *yaml.Node) bool {
	if existing == nil {
		return true
	}
	if existing.Kind == yaml.SequenceNode && existing.Style != yaml.FlowStyle {
		return true
	}
	return len(values) >= 5
}

func literalBlock(raw string, indent int) string {
	lines := strings.Split(raw, "\n")
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return strconv.Quote("")
	}
	prefix := strings.Repeat(" ", indent)
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return "|-\n" + strings.Join(lines, "\n")
}

func literalBlockLogical(raw string) string {
	lines := strings.Split(raw, "\n")
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}
