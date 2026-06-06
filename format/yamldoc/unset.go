package yamldoc

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func unsetValue(source, path string) (string, error) {
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
	planned, err := planUnsetFromBase(source, root, root, key, key)
	if err != nil {
		return "", err
	}
	return applyEdit(source, planned), nil
}

func unsetValueIn(source, collectionPath string, rawSelectors []string, path string) (string, error) {
	root, err := parseYAML(source)
	if err != nil {
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
	if len(collection) == 0 {
		return "", fmt.Errorf("--in: empty path")
	}
	if len(key) == 0 {
		return "", fmt.Errorf("empty path")
	}
	if root == nil {
		return "", fmt.Errorf("%s is not set", formatPath(collection))
	}

	collectionNode, err := resolvePath(root, collection, nil)
	if err != nil {
		return "", err
	}
	collectionNode, err = resolveAlias(collectionNode, nil)
	if err != nil {
		return "", err
	}
	if collectionNode.Kind != yaml.SequenceNode || !isRecordSequence(collectionNode) {
		return "", fmt.Errorf("%s is not a sequence of records", formatPath(collection))
	}
	selectors, err := parseSelectors(rawSelectors)
	if err != nil {
		return "", err
	}
	index, err := selectedRecordIndex(collectionNode, collection, selectors)
	if err != nil {
		return "", err
	}
	record, err := resolveAlias(collectionNode.Content[index], nil)
	if err != nil {
		return "", err
	}
	fullPath := appendPath(collection, fmt.Sprintf("%d", index))
	fullPath = append(fullPath, key...)
	planned, err := planUnsetFromBase(source, root, record, fullPath, key)
	if err != nil {
		return "", err
	}
	return applyEdit(source, planned), nil
}

func planUnsetFromBase(source string, doc, base *yaml.Node, fullPath, relative []string) (edit, error) {
	found, ok := findTarget(base, relative)
	if !ok {
		return edit{}, fmt.Errorf("%s is not set", formatPath(fullPath))
	}
	found.path = fullPath
	if err := ensureUnsettable(found, doc); err != nil {
		return edit{}, err
	}
	return targetLineRange(source, found)
}

func ensureUnsettable(found target, doc *yaml.Node) error {
	if found.valueNode.Kind == yaml.MappingNode || found.valueNode.Kind == yaml.SequenceNode {
		return fmt.Errorf("%s is a container, not a scalar value", formatPath(found.path))
	}
	if found.valueNode.Anchor != "" && hasAliasReference(doc, found.valueNode.Anchor) {
		return fmt.Errorf("%s defines anchor %q that is still referenced", formatPath(found.path), found.valueNode.Anchor)
	}
	return nil
}

func targetLineRange(source string, found target) (edit, error) {
	if found.parent != nil && found.parent.Kind == yaml.MappingNode && found.parent.Style == yaml.FlowStyle {
		return flowMappingTargetRange(source, found)
	}
	start, end, err := lineRange(source, found.keyNode.Line)
	if err != nil {
		return edit{}, err
	}
	if found.valueNode.Style == yaml.LiteralStyle || found.valueNode.Style == yaml.FoldedStyle {
		_, blockEnd, err := blockValueRange(source, found)
		if err != nil {
			return edit{}, err
		}
		end = blockEnd
	}
	return edit{start: start, end: end, text: ""}, nil
}

func flowMappingTargetRange(source string, found target) (edit, error) {
	lineStart, lineEnd, err := lineRange(source, found.keyNode.Line)
	if err != nil {
		return edit{}, err
	}
	line := source[lineStart:lineEnd]
	keyStart := max(found.keyNode.Column-1, 0)
	_, valueEnd, err := valueRange(source, found)
	if err != nil {
		return edit{}, err
	}
	valueEnd -= lineStart
	segmentStart := keyStart
	segmentEnd := valueEnd
	if nextComma := strings.IndexByte(line[segmentEnd:], ','); nextComma == 0 {
		segmentEnd++
		for segmentEnd < len(line) && (line[segmentEnd] == ' ' || line[segmentEnd] == '\t') {
			segmentEnd++
		}
	} else if previousComma := strings.LastIndexByte(line[:keyStart], ','); previousComma >= 0 {
		segmentStart = previousComma
		for segmentStart > 0 && (line[segmentStart-1] == ' ' || line[segmentStart-1] == '\t') {
			segmentStart--
		}
	}
	return edit{start: lineStart + segmentStart, end: lineStart + segmentEnd, text: ""}, nil
}

func hasAliasReference(node *yaml.Node, anchor string) bool {
	if node == nil {
		return false
	}
	if node.Kind == yaml.AliasNode {
		if node.Value == anchor {
			return true
		}
		if node.Alias != nil && node.Alias.Anchor == anchor {
			return true
		}
	}
	for _, child := range node.Content {
		if hasAliasReference(child, anchor) {
			return true
		}
	}
	return false
}
