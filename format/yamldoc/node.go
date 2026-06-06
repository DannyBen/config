package yamldoc

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

func mappingInsertAt(source string, node *yaml.Node) (int, error) {
	if len(node.Content) == 0 {
		start, _, err := lineRange(source, node.Line)
		return start, err
	}
	lastValue := node.Content[len(node.Content)-1]
	if lastValue.Kind == yaml.MappingNode || lastValue.Kind == yaml.SequenceNode || lastValue.Style == yaml.LiteralStyle || lastValue.Style == yaml.FoldedStyle {
		_, end, err := nodeRange(source, lastValue, childIndent(node)-2)
		return end, err
	}
	_, end, err := lineRange(source, lastValue.Line)
	return end, err
}

func sequenceInsertAt(source string, node *yaml.Node) (int, error) {
	if len(node.Content) == 0 {
		start, _, err := lineRange(source, node.Line)
		return start, err
	}
	last := node.Content[len(node.Content)-1]
	_, end, err := nodeRange(source, last, max(node.Column-1, 0))
	return end, err
}

func nodeRange(source string, node *yaml.Node, parentIndent int) (int, int, error) {
	start, end, err := lineRange(source, node.Line)
	if err != nil {
		return 0, 0, err
	}
	for end < len(source) {
		lineEnd := strings.IndexByte(source[end:], '\n')
		stop := len(source)
		if lineEnd >= 0 {
			stop = end + lineEnd
		}
		line := source[end:stop]
		if strings.TrimSpace(line) != "" && leadingSpaces(line) <= parentIndent {
			break
		}
		end = stop
		if lineEnd < 0 {
			break
		}
		end++
	}
	return start, end, nil
}

func childIndent(node *yaml.Node) int {
	if len(node.Content) > 0 {
		return max(node.Content[0].Column-1, 0)
	}
	return max(node.Column+1, 2)
}

func nestedMappingText(path []string, value string, indent int) string {
	text := strings.Join(nestedMappingLines(path, value, indent), "\n")
	if strings.HasSuffix(text, "\n") {
		return text
	}
	return text + "\n"
}

func nestedMappingLines(path []string, value string, indent int) []string {
	if len(path) == 1 {
		if strings.HasPrefix(value, "\n") {
			return []string{strings.Repeat(" ", indent) + path[0] + ":" + value}
		}
		return []string{strings.Repeat(" ", indent) + path[0] + ": " + value}
	}
	lines := []string{strings.Repeat(" ", indent) + path[0] + ":"}
	lines = append(lines, nestedMappingLines(path[1:], value, indent+2)...)
	return lines
}

func prefixLineEndingAt(source string, pos int) string {
	if pos == 0 || source[pos-1] == '\n' {
		return ""
	}
	return "\n"
}

func applyEdit(source string, e edit) string {
	if e.start == 0 && e.end == 0 && e.text == "" {
		return source
	}
	return source[:e.start] + e.text + source[e.end:]
}

func trimRightSpaceIndex(s string, end int) int {
	for end > 0 && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return end
}

func leadingSpaces(line string) int {
	count := 0
	for count < len(line) && line[count] == ' ' {
		count++
	}
	return count
}

func samePath(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var (
	integer = regexp.MustCompile(`^[+-]?[0-9]+$`)
	float   = regexp.MustCompile(`^[+-]?([0-9]+\.[0-9]*|[0-9]*\.[0-9]+)([eE][+-]?[0-9]+)?$|^[+-]?[0-9]+[eE][+-]?[0-9]+$`)
)
