package yamldoc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var errNotEmpty = errors.New("not empty")

func deleteValue(source, path string, rawSelectors []string) (string, error) {
	root, err := parseYAML(source)
	if err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	if len(key) == 0 {
		return "", fmt.Errorf("empty path")
	}
	if root == nil {
		return "", fmt.Errorf("%s is not set", formatPath(key))
	}

	var planned edit
	if len(rawSelectors) > 0 {
		planned, err = planDeleteSelectedRecord(source, root, key, rawSelectors)
	} else {
		planned, err = planDeletePath(source, root, key)
	}
	if err != nil {
		return "", err
	}
	return applyEdit(source, planned), nil
}

func deleteEmptyValue(source, path string) (string, error) {
	root, err := parseYAML(source)
	if err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	if len(key) == 0 {
		return "", fmt.Errorf("empty path")
	}
	if root == nil {
		return source, nil
	}

	found, ok := findTarget(root, key)
	if !ok {
		return source, nil
	}
	found.path = key
	if err := ensureEmptyDeletableContainer(source, found, root); err != nil {
		if errors.Is(err, errNotEmpty) {
			return source, nil
		}
		return "", err
	}
	planned, err := containerTargetRange(source, found)
	if err != nil {
		return "", err
	}
	return applyEdit(source, planned), nil
}

func planDeletePath(source string, root *yaml.Node, path []string) (edit, error) {
	if collection, index, ok := indexedPath(path); ok {
		return planDeleteIndexedRecord(source, root, collection, index)
	}

	found, ok := findTarget(root, path)
	if !ok {
		return edit{}, fmt.Errorf("%s is not set", formatPath(path))
	}
	found.path = path
	if err := ensureDeletableContainer(found.valueNode, root, path); err != nil {
		return edit{}, err
	}
	return containerTargetRange(source, found)
}

func indexedPath(path []string) ([]string, int, bool) {
	if len(path) < 2 {
		return nil, 0, false
	}
	index, err := strconv.Atoi(path[len(path)-1])
	if err != nil || index < 0 {
		return nil, 0, false
	}
	collection := make([]string, len(path)-1)
	copy(collection, path[:len(path)-1])
	return collection, index, true
}

func planDeleteIndexedRecord(source string, root *yaml.Node, collection []string, index int) (edit, error) {
	collectionNode, err := resolvePath(root, collection, nil)
	if err != nil {
		return edit{}, fmt.Errorf("%s is not set", formatPath(appendPath(collection, strconv.Itoa(index))))
	}
	collectionNode, err = resolveAlias(collectionNode, nil)
	if err != nil {
		return edit{}, err
	}
	if collectionNode.Kind != yaml.SequenceNode || !isRecordSequence(collectionNode) {
		return edit{}, fmt.Errorf("%s is not a sequence of records", formatPath(collection))
	}
	if index >= len(collectionNode.Content) {
		return edit{}, fmt.Errorf("%s has no record at index %d", formatPath(collection), index)
	}
	if collectionNode.Content[index].Kind == yaml.AliasNode {
		fullPath := appendPath(collection, strconv.Itoa(index))
		return edit{}, fmt.Errorf("%s is an alias; refusing to delete shared YAML state", formatPath(fullPath))
	}
	record, err := resolveAlias(collectionNode.Content[index], nil)
	if err != nil {
		return edit{}, err
	}
	fullPath := appendPath(collection, strconv.Itoa(index))
	if record.Kind != yaml.MappingNode {
		return edit{}, fmt.Errorf("%s is not a record", formatPath(fullPath))
	}
	if err := ensureDeletableContainer(record, root, fullPath); err != nil {
		return edit{}, err
	}
	return sequenceItemRange(source, collectionNode, collection, index)
}

func planDeleteSelectedRecord(source string, root *yaml.Node, collection []string, rawSelectors []string) (edit, error) {
	collectionNode, err := resolvePath(root, collection, nil)
	if err != nil {
		return edit{}, err
	}
	collectionNode, err = resolveAlias(collectionNode, nil)
	if err != nil {
		return edit{}, err
	}
	if collectionNode.Kind != yaml.SequenceNode || !isRecordSequence(collectionNode) {
		return edit{}, fmt.Errorf("%s is not a sequence of records", formatPath(collection))
	}
	selectors, err := parseSelectors(rawSelectors)
	if err != nil {
		return edit{}, err
	}
	index, err := selectedRecordIndex(collectionNode, collection, selectors)
	if err != nil {
		return edit{}, err
	}
	if collectionNode.Content[index].Kind == yaml.AliasNode {
		fullPath := appendPath(collection, strconv.Itoa(index))
		return edit{}, fmt.Errorf("%s is an alias; refusing to delete shared YAML state", formatPath(fullPath))
	}
	record, err := resolveAlias(collectionNode.Content[index], nil)
	if err != nil {
		return edit{}, err
	}
	fullPath := appendPath(collection, strconv.Itoa(index))
	if err := ensureDeletableContainer(record, root, fullPath); err != nil {
		return edit{}, err
	}
	return sequenceItemRange(source, collectionNode, collection, index)
}

func ensureDeletableContainer(node, doc *yaml.Node, path []string) error {
	if node.Kind == yaml.AliasNode {
		return fmt.Errorf("%s is an alias; refusing to delete shared YAML state", formatPath(path))
	}
	node, err := resolveAlias(node, nil)
	if err != nil {
		return err
	}
	switch node.Kind {
	case yaml.ScalarNode:
		return fmt.Errorf("%s is a value, use unset to remove fields", formatPath(path))
	case yaml.MappingNode, yaml.SequenceNode:
		if node.Anchor != "" && hasAliasReference(doc, node.Anchor) {
			return fmt.Errorf("%s defines anchor %q that is still referenced", formatPath(path), node.Anchor)
		}
		return nil
	default:
		return fmt.Errorf("unsupported YAML node at %s", formatPath(path))
	}
}

func ensureEmptyDeletableContainer(source string, found target, doc *yaml.Node) error {
	if found.valueNode.Kind == yaml.AliasNode {
		return fmt.Errorf("%s is an alias; refusing to delete shared YAML state", formatPath(found.path))
	}
	node, err := resolveAlias(found.valueNode, nil)
	if err != nil {
		return err
	}
	switch node.Kind {
	case yaml.ScalarNode:
		if isImplicitNullValue(source, found) {
			return nil
		}
		return errNotEmpty
	case yaml.MappingNode, yaml.SequenceNode:
		if node.Anchor != "" && hasAliasReference(doc, node.Anchor) {
			return fmt.Errorf("%s defines anchor %q that is still referenced", formatPath(found.path), node.Anchor)
		}
		if len(node.Content) > 0 {
			return errNotEmpty
		}
		return nil
	default:
		return fmt.Errorf("unsupported YAML node at %s", formatPath(found.path))
	}
}

func isImplicitNullValue(source string, found target) bool {
	node := found.valueNode
	if node.Tag != "!!null" || node.Value != "" || node.Line != found.keyNode.Line {
		return false
	}
	start, end, err := lineRange(source, found.keyNode.Line)
	if err != nil {
		return false
	}
	line := source[start:end]
	column := found.keyNode.Column - 1
	if column < 0 || column+len(found.keyNode.Value) > len(line) {
		return false
	}
	afterKey := line[column+len(found.keyNode.Value):]
	colon := strings.IndexByte(afterKey, ':')
	if colon < 0 {
		return false
	}
	afterColon := strings.TrimSpace(afterKey[colon+1:])
	return afterColon == "" || strings.HasPrefix(afterColon, "#")
}

func containerTargetRange(source string, found target) (edit, error) {
	start, _, err := lineRange(source, found.keyNode.Line)
	if err != nil {
		return edit{}, err
	}
	if found.valueNode.Line == found.keyNode.Line {
		lineStart, end, err := lineRange(source, found.keyNode.Line)
		if err != nil {
			return edit{}, err
		}
		line := source[lineStart:end]
		if found.keyNode.Column-1 != leadingSpaces(line) {
			return edit{}, fmt.Errorf("%s cannot be safely deleted", formatPath(found.path))
		}
		return edit{start: start, end: end, text: ""}, nil
	}

	parentIndent := max(found.keyNode.Column-1, 0)
	_, end, err := nodeRange(source, found.valueNode, parentIndent)
	if err != nil {
		return edit{}, err
	}
	return edit{start: start, end: end, text: ""}, nil
}

func sequenceItemRange(source string, sequence *yaml.Node, collection []string, index int) (edit, error) {
	if sequence.Style == yaml.FlowStyle {
		return edit{}, fmt.Errorf("%s.%d cannot be safely deleted", formatPath(collection), index)
	}
	child := sequence.Content[index]
	start, _, err := lineRange(source, child.Line)
	if err != nil {
		return edit{}, err
	}
	itemIndent := max(sequence.Column-1, 0)
	start = sequenceItemStart(source, start, itemIndent)
	_, end, err := nodeRange(source, child, itemIndent)
	if err != nil {
		return edit{}, err
	}
	return edit{start: start, end: end, text: ""}, nil
}

func sequenceItemStart(source string, start, itemIndent int) int {
	current := start
	for current > 0 {
		previousEnd := current - 1
		previousStart := strings.LastIndexByte(source[:previousEnd], '\n') + 1
		line := source[previousStart:previousEnd]
		if strings.TrimSpace(line) == "" {
			current = previousStart
			continue
		}
		if leadingSpaces(line) == itemIndent && strings.HasPrefix(strings.TrimLeft(line, " "), "-") {
			return previousStart
		}
		break
	}
	return start
}
