package yamldoc

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func findTarget(root *yaml.Node, path []string) (target, bool) {
	if len(path) == 0 {
		return target{}, false
	}
	if found, ok := findSequenceTarget(root, path); ok {
		return found, true
	}
	parent, ok := resolveParentForSet(root, path[:len(path)-1])
	if !ok || parent.Kind != yaml.MappingNode {
		return target{}, false
	}
	key := path[len(path)-1]
	for i := 0; i+1 < len(parent.Content); i += 2 {
		keyNode := parent.Content[i]
		if keyNode.Kind == yaml.ScalarNode && keyNode.Value == key {
			return target{keyNode: keyNode, valueNode: parent.Content[i+1], parent: parent, key: key, path: path}, true
		}
	}
	return target{}, false
}

func findSequenceTarget(root *yaml.Node, path []string) (target, bool) {
	parent, ok := resolveParentForSet(root, path[:len(path)-1])
	if !ok || parent.Kind != yaml.SequenceNode {
		return target{}, false
	}
	index, err := strconv.Atoi(path[len(path)-1])
	if err != nil || index < 0 || index >= len(parent.Content) {
		return target{}, false
	}
	return target{valueNode: parent.Content[index], parent: parent, path: path}, true
}

func resolveParentForSet(root *yaml.Node, path []string) (*yaml.Node, bool) {
	node := root
	for _, part := range path {
		node, _ = resolveAlias(node, nil)
		switch node.Kind {
		case yaml.MappingNode:
			next, ok := mappingChild(node, part)
			if !ok {
				return nil, false
			}
			node = next
		case yaml.SequenceNode:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(node.Content) {
				return nil, false
			}
			node = node.Content[index]
		default:
			return nil, false
		}
	}
	node, _ = resolveAlias(node, nil)
	return node, true
}

func mappingChild(node *yaml.Node, key string) (*yaml.Node, bool) {
	if err := rejectDuplicateKeys(node); err != nil {
		return nil, false
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Kind == yaml.ScalarNode && node.Content[i].Value == key {
			return node.Content[i+1], true
		}
	}
	return nil, false
}

func valueRange(source string, found target) (int, int, error) {
	if found.valueNode.Style == yaml.LiteralStyle || found.valueNode.Style == yaml.FoldedStyle {
		return blockValueRange(source, found)
	}
	start, end, err := lineRange(source, found.valueNode.Line)
	if err != nil {
		return 0, 0, err
	}
	line := source[start:end]
	col := max(found.valueNode.Column-1, 0)
	if col > len(line) {
		col = len(line)
	}
	if found.valueNode.Anchor != "" {
		col = anchoredValueColumn(line, col, found.valueNode.Anchor)
	}
	valueStart := start + col
	flow := found.parent != nil && found.parent.Style == yaml.FlowStyle
	valueEnd := start + valueEndInLine(line, col, flow)
	return valueStart, valueEnd, nil
}

func blockValueRange(source string, found target) (int, int, error) {
	start, _, err := lineRange(source, found.valueNode.Line)
	if err != nil {
		return 0, 0, err
	}
	lineEnd := strings.IndexByte(source[start:], '\n')
	if lineEnd < 0 {
		lineEnd = len(source) - start
	}
	line := source[start : start+lineEnd]
	col := max(found.valueNode.Column-1, 0)
	valueStart := start + min(col, len(line))
	keyIndent := max(found.keyNode.Column-1, 0)
	end := start + lineEnd
	for end < len(source) {
		if source[end] == '\n' {
			end++
		}
		nextEnd := strings.IndexByte(source[end:], '\n')
		lineStop := len(source)
		if nextEnd >= 0 {
			lineStop = end + nextEnd
		}
		nextLine := source[end:lineStop]
		if strings.TrimSpace(nextLine) != "" && leadingSpaces(nextLine) <= keyIndent {
			break
		}
		end = lineStop
		if nextEnd < 0 {
			break
		}
	}
	return valueStart, end, nil
}

func lineRange(source string, lineNo int) (int, int, error) {
	if lineNo <= 0 {
		return 0, 0, fmt.Errorf("invalid line %d", lineNo)
	}
	line := 1
	start := 0
	for i, ch := range source {
		if line == lineNo {
			start = i
			break
		}
		if ch == '\n' {
			line++
		}
	}
	if line != lineNo {
		return 0, 0, fmt.Errorf("line %d not found", lineNo)
	}
	end := strings.IndexByte(source[start:], '\n')
	if end < 0 {
		return start, len(source), nil
	}
	return start, start + end + 1, nil
}

func valueEndInLine(line string, start int, flow bool) int {
	inQuote := byte(0)
	escaped := false
	for i := start; i < len(line); i++ {
		ch := line[i]
		if inQuote != 0 {
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == inQuote {
				inQuote = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			inQuote = ch
			continue
		}
		if ch == '#' || ch == '\n' {
			return trimRightSpaceIndex(line, i)
		}
		if flow && (ch == ',' || ch == '}' || ch == ']') {
			return trimRightSpaceIndex(line, i)
		}
	}
	return trimRightSpaceIndex(line, len(line))
}

func matchingFlowClose(line string, start int, open, close byte) int {
	inQuote := byte(0)
	escaped := false
	depth := 0
	for i := start; i < len(line); i++ {
		ch := line[i]
		if inQuote != 0 {
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == inQuote {
				inQuote = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			inQuote = ch
			continue
		}
		switch ch {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func anchoredValueColumn(line string, col int, anchor string) int {
	marker := "&" + anchor
	anchorAt := strings.Index(line[col:], marker)
	if anchorAt < 0 {
		return col
	}
	col += anchorAt + len(marker)
	for col < len(line) && (line[col] == ' ' || line[col] == '\t') {
		col++
	}
	return col
}
