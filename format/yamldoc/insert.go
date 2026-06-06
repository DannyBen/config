package yamldoc

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func planInsertMissing(source string, root, base *yaml.Node, fullPath, relative []string, format func(int, *yaml.Node) (string, string, error)) (edit, error) {
	if len(relative) == 0 {
		return edit{}, fmt.Errorf("%s is not set", formatPath(fullPath))
	}

	current, missing, err := insertionBase(base, fullPath, relative)
	if err != nil {
		return edit{}, err
	}

	keyIndent := childIndent(current)
	value, _, err := format(keyIndent+2*len(missing), nil)
	if err != nil {
		return edit{}, err
	}
	if current.Style == yaml.FlowStyle {
		return insertIntoFlowMapping(source, current, missing, value, fullPath)
	}
	text := nestedMappingText(missing, value, keyIndent)
	insertAt, err := mappingInsertAt(source, current)
	if err != nil {
		return edit{}, err
	}
	return edit{start: insertAt, end: insertAt, text: prefixLineEndingAt(source, insertAt) + text}, nil
}

func insertIntoFlowMapping(source string, node *yaml.Node, missing []string, value string, fullPath []string) (edit, error) {
	if len(missing) != 1 {
		return edit{}, fmt.Errorf("%s parent is a flow mapping; refusing to expand nested YAML", formatPath(fullPath[:len(fullPath)-len(missing)+1]))
	}
	if strings.Contains(value, "\n") {
		return edit{}, fmt.Errorf("%s parent is a flow mapping; refusing to insert multiline YAML", formatPath(fullPath[:len(fullPath)-len(missing)+1]))
	}
	lineStart, lineEnd, err := lineRange(source, node.Line)
	if err != nil {
		return edit{}, err
	}
	line := source[lineStart:lineEnd]
	start := max(node.Column-1, 0)
	closeAt := matchingFlowClose(line, start, '{', '}')
	if closeAt < 0 {
		return edit{}, fmt.Errorf("%s flow mapping is not editable", formatPath(fullPath[:len(fullPath)-len(missing)]))
	}
	insertAt := closeAt
	for insertAt > start && (line[insertAt-1] == ' ' || line[insertAt-1] == '\t') {
		insertAt--
	}
	suffix := line[insertAt:closeAt]
	prefix := ""
	if len(node.Content) > 0 {
		prefix = ", "
	}
	return edit{start: lineStart + insertAt, end: lineStart + closeAt, text: prefix + missing[0] + ": " + value + suffix}, nil
}

func insertionBase(base *yaml.Node, fullPath, relative []string) (*yaml.Node, []string, error) {
	current, err := resolveAlias(base, nil)
	if err != nil {
		return nil, nil, err
	}
	basePathLen := len(fullPath) - len(relative)
	existing := 0
	for existing < len(relative)-1 {
		part := relative[existing]
		switch current.Kind {
		case yaml.MappingNode:
			next, ok := mappingChild(current, part)
			if !ok {
				return current, relative[existing:], nil
			}
			next, err := resolveAlias(next, nil)
			if err != nil {
				return nil, nil, err
			}
			if next.Kind != yaml.MappingNode && next.Kind != yaml.SequenceNode {
				parentPath := fullPath[:basePathLen+existing+1]
				return nil, nil, fmt.Errorf("%s parent is not a mapping", formatPath(parentPath))
			}
			current = next
			existing++
		case yaml.SequenceNode:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(current.Content) {
				parentPath := fullPath[:basePathLen+existing+1]
				return nil, nil, fmt.Errorf("%s is not set", formatPath(parentPath))
			}
			next, err := resolveAlias(current.Content[index], nil)
			if err != nil {
				return nil, nil, err
			}
			if next.Kind != yaml.MappingNode && next.Kind != yaml.SequenceNode {
				parentPath := fullPath[:basePathLen+existing+1]
				return nil, nil, fmt.Errorf("%s parent is not a mapping", formatPath(parentPath))
			}
			current = next
			existing++
		default:
			parentPath := fullPath[:basePathLen+existing]
			return nil, nil, fmt.Errorf("%s parent is not a mapping", formatPath(parentPath))
		}
	}

	if current.Kind != yaml.MappingNode {
		parentPath := fullPath[:len(fullPath)-len(relative)+existing]
		return nil, nil, fmt.Errorf("%s parent is not a mapping", formatPath(parentPath))
	}
	return current, relative[existing:], nil
}

func insertMissingCollectionRecord(source string, root *yaml.Node, collection []string, selector selector, key []string, semantic func([]string, string) semanticSet, format func(int, *yaml.Node) (string, string, error)) (string, error) {
	var logical string
	planned, err := planInsertMissing(source, root, root, collection, collection, func(indent int, existing *yaml.Node) (string, string, error) {
		record, formattedLogical, err := recordText(selector, key, format, indent)
		if err != nil {
			return "", "", err
		}
		logical = formattedLogical
		return "\n" + record, "", nil
	})
	if err != nil {
		return "", err
	}
	return applyVerifiedSet(source, planned, recordSemanticChanges(collection, 0, selector, key, logical, semantic)...)
}

func appendRecord(source string, collectionNode *yaml.Node, collection []string, selector selector, key []string, semantic func([]string, string) semanticSet, format func(int, *yaml.Node) (string, string, error)) (string, error) {
	insertAt, err := sequenceInsertAt(source, collectionNode)
	if err != nil {
		return "", err
	}
	itemIndent := max(collectionNode.Column-1, 0)
	text, logical, err := recordText(selector, key, format, itemIndent)
	if err != nil {
		return "", err
	}
	planned := edit{start: insertAt, end: insertAt, text: prefixLineEndingAt(source, insertAt) + text}
	return applyVerifiedSet(source, planned, recordSemanticChanges(collection, len(collectionNode.Content), selector, key, logical, semantic)...)
}

func recordText(selector selector, key []string, format func(int, *yaml.Node) (string, string, error), itemIndent int) (string, string, error) {
	if len(selector.path) != 1 {
		return "", "", fmt.Errorf("--on nested selectors are not supported for creating YAML records")
	}
	if len(key) == 0 {
		return "", "", fmt.Errorf("empty path")
	}
	value, logical, err := format(itemIndent+2, nil)
	if err != nil {
		return "", "", err
	}
	var lines []string
	lines = append(lines, strings.Repeat(" ", itemIndent)+"- "+selector.path[0]+": "+formatPlainString(selector.compare))
	if !samePath(selector.path, key) {
		lines = append(lines, nestedMappingLines(key, value, itemIndent+2)...)
	}
	return strings.Join(lines, "\n") + "\n", logical, nil
}

func recordSemanticChanges(collection []string, index int, selector selector, key []string, logical string, semantic func([]string, string) semanticSet) []semanticSet {
	recordPath := appendPath(collection, strconv.Itoa(index))
	selectorPath := append(append([]string{}, recordPath...), selector.path...)
	changes := []semanticSet{semanticScalar(selectorPath, selector.compare)}
	if !samePath(selector.path, key) {
		keyPath := append(append([]string{}, recordPath...), key...)
		changes = append(changes, semantic(keyPath, logical))
	}
	return changes
}
