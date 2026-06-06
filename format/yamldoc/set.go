package yamldoc

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type setMode int

const (
	setModeInfer setMode = iota
	setModeString
)

type edit struct {
	start int
	end   int
	text  string
}

type target struct {
	keyNode   *yaml.Node
	valueNode *yaml.Node
	parent    *yaml.Node
	key       string
	path      []string
}

func setValue(source, path, raw string, mode setMode) (string, error) {
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
		return "", fmt.Errorf("%s parent is not set", formatPath(key[:len(key)-1]))
	}
	planned, logical, err := planSet(source, root, key, raw, mode)
	if err != nil {
		return "", err
	}
	return applyVerifiedSet(source, planned, semanticScalar(key, logical))
}

func setArrayValue(source, path string, values []string) (string, error) {
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
		return "", fmt.Errorf("%s parent is not set", formatPath(key[:len(key)-1]))
	}
	planned, err := planSetArray(source, root, key, values)
	if err != nil {
		return "", err
	}
	return applyVerifiedSet(source, planned, semanticArray(key, values))
}

func setValueIn(source, collectionPath, rawSelector, path, raw string, mode setMode) (string, error) {
	return setIn(source, collectionPath, rawSelector, path, false, func(fullPath []string, logical string) semanticSet {
		return semanticScalar(fullPath, logical)
	}, func(indent int, existing *yaml.Node) (string, string, error) {
		value, logical, err := formatSetValue(raw, mode, indent, existing)
		return value, logical, err
	})
}

func setArrayValueIn(source, collectionPath, rawSelector, path string, values []string) (string, error) {
	return setIn(source, collectionPath, rawSelector, path, true, func(fullPath []string, _ string) semanticSet {
		return semanticArray(fullPath, values)
	}, func(indent int, existing *yaml.Node) (string, string, error) {
		value, err := formatArray(values, indent, existing)
		if err != nil {
			return "", "", err
		}
		return value, "", nil
	})
}

func setIn(source, collectionPath, rawSelector, path string, allowContainer bool, semantic func([]string, string) semanticSet, format func(int, *yaml.Node) (string, string, error)) (string, error) {
	root, err := parseYAML(source)
	if err != nil {
		return "", err
	}
	collection, selectors, key, err := parseSetInArgs(collectionPath, rawSelector, path)
	if err != nil {
		return "", err
	}
	if root == nil {
		if len(collection) == 0 {
			return "", fmt.Errorf("root collection is not set")
		}
		return "", fmt.Errorf("%s parent is not set", formatPath(collection[:len(collection)-1]))
	}

	collectionNode := root
	if len(collection) > 0 {
		var collectionErr error
		collectionNode, collectionErr = resolvePath(root, collection, nil)
		if collectionErr != nil {
			return insertMissingCollectionRecord(source, root, collection, selectors[0], key, semantic, format)
		}
	}
	if len(collection) > 0 && collectionNode == nil {
		return insertMissingCollectionRecord(source, root, collection, selectors[0], key, semantic, format)
	}
	collectionNode, err = resolveAlias(collectionNode, nil)
	if err != nil {
		return "", err
	}
	if collectionNode.Kind != yaml.SequenceNode {
		return "", fmt.Errorf("%s is not a sequence of records", formatPath(collection))
	}
	if isRecordSequence(collectionNode) {
		index, err := selectedRecordIndex(collectionNode, collection, selectors)
		if err != nil {
			if strings.Contains(err.Error(), " has no records matching ") {
				return appendRecord(source, collectionNode, collection, selectors[0], key, semantic, format)
			}
			return "", err
		}
		record, err := resolveAlias(collectionNode.Content[index], nil)
		if err != nil {
			return "", err
		}
		return setRelative(source, root, record, append(appendPath(collection, strconv.Itoa(index)), key...), key, allowContainer, semantic, format)
	}
	return appendRecord(source, collectionNode, collection, selectors[0], key, semantic, format)
}

func parseSetInArgs(collectionPath, rawSelector, path string) ([]string, []selector, []string, error) {
	collection, err := parsePath(collectionPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("--in: %w", err)
	}
	key, err := parsePath(path)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(key) == 0 {
		return nil, nil, nil, fmt.Errorf("empty path")
	}
	selectors, err := parseSelectors([]string{rawSelector})
	if err != nil {
		return nil, nil, nil, err
	}
	return collection, selectors, key, nil
}

func setRelative(source string, root, base *yaml.Node, fullPath, relative []string, allowContainer bool, semantic func([]string, string) semanticSet, format func(int, *yaml.Node) (string, string, error)) (string, error) {
	if found, ok := findTarget(base, relative); ok {
		indent := max(found.keyNode.Column+1, 2)
		value, logical, err := format(indent, found.valueNode)
		if err != nil {
			return "", err
		}
		planned, err := replaceTarget(source, found, value, logical, allowContainer)
		if err != nil {
			return "", err
		}
		return applyVerifiedSet(source, planned, semantic(fullPath, logical))
	}
	planned, err := planInsertMissing(source, root, base, fullPath, relative, format)
	if err != nil {
		return "", err
	}
	return applyVerifiedInsertedPath(source, planned, fullPath)
}

func planSet(source string, root *yaml.Node, key []string, raw string, mode setMode) (edit, string, error) {
	if found, ok := findTarget(root, key); ok {
		indent := replacementIndent(found)
		value, logical, err := formatSetValue(raw, mode, indent, found.valueNode)
		if err != nil {
			return edit{}, "", err
		}
		planned, err := replaceTarget(source, found, value, logical, false)
		return planned, logical, err
	}
	var logical string
	planned, err := planInsertMissing(source, root, root, key, key, func(indent int, existing *yaml.Node) (string, string, error) {
		value, formattedLogical, err := formatSetValue(raw, mode, indent, existing)
		logical = formattedLogical
		return value, formattedLogical, err
	})
	return planned, logical, err
}

func replacementIndent(found target) int {
	if found.keyNode != nil {
		return max(found.keyNode.Column+1, 2)
	}
	if found.valueNode != nil {
		return max(found.valueNode.Column-1, 0)
	}
	return 2
}

func planSetFormatted(source string, root *yaml.Node, key []string, value, logical string) (edit, error) {
	if found, ok := findTarget(root, key); ok {
		return replaceTarget(source, found, value, logical, true)
	}
	return planInsertMissing(source, root, root, key, key, func(int, *yaml.Node) (string, string, error) {
		return value, logical, nil
	})
}

func planSetArray(source string, root *yaml.Node, key []string, values []string) (edit, error) {
	if found, ok := findTarget(root, key); ok {
		indent := replacementIndent(found)
		value, err := formatArray(values, indent, found.valueNode)
		if err != nil {
			return edit{}, err
		}
		return replaceTarget(source, found, value, "", true)
	}
	return planInsertMissing(source, root, root, key, key, func(indent int, existing *yaml.Node) (string, string, error) {
		value, err := formatArray(values, indent, existing)
		if err != nil {
			return "", "", err
		}
		return value, "", nil
	})
}

func replaceTarget(source string, found target, value, logical string, allowContainer bool) (edit, error) {
	if found.valueNode.Kind == yaml.AliasNode {
		return replaceAliasTarget(found, logical)
	}
	if found.valueNode.Kind == yaml.MappingNode || (found.valueNode.Kind == yaml.SequenceNode && !allowContainer) {
		return edit{}, fmt.Errorf("%s is a container, not a scalar value", formatPath(found.path))
	}
	if found.parent != nil && found.parent.Kind == yaml.SequenceNode {
		start, end, err := valueRange(source, found)
		if err != nil {
			return edit{}, err
		}
		return edit{start: start, end: end, text: value}, nil
	}
	if found.valueNode.Kind == yaml.SequenceNode {
		return replaceSequenceTarget(source, found, value)
	}
	return replaceScalarTarget(source, found, value)
}

func replaceAliasTarget(found target, logical string) (edit, error) {
	got, err := renderGetValue(found.valueNode, found.path)
	if err != nil {
		return edit{}, err
	}
	if logical != "" && got == logical {
		return edit{start: 0, end: 0, text: ""}, nil
	}
	return edit{}, fmt.Errorf("%s is an alias; refusing to change shared YAML state", formatPath(found.path))
}

func replaceSequenceTarget(source string, found target, value string) (edit, error) {
	start, end, err := sequenceValueRange(source, found)
	if err != nil {
		return edit{}, err
	}
	if strings.HasPrefix(value, "\n") && found.valueNode.Line == found.keyNode.Line {
		start = trimLeftInlineValueSpace(source, start)
	}
	text := value
	if found.valueNode.Line != found.keyNode.Line {
		if !strings.HasPrefix(value, "\n") {
			text = " " + value
		}
		if end > start && source[end-1] == '\n' {
			text += "\n"
		}
	}
	return edit{start: start, end: end, text: text}, nil
}

func replaceScalarTarget(source string, found target, value string) (edit, error) {
	start, end, err := valueRange(source, found)
	if err != nil {
		return edit{}, err
	}
	if strings.HasPrefix(value, "\n") {
		start = trimLeftInlineValueSpace(source, start)
	}
	if (found.valueNode.Style == yaml.LiteralStyle || found.valueNode.Style == yaml.FoldedStyle) && end > start && source[end-1] == '\n' {
		value += "\n"
	}
	return edit{start: start, end: end, text: value}, nil
}

func trimLeftInlineValueSpace(source string, start int) int {
	for start > 0 && (source[start-1] == ' ' || source[start-1] == '\t') {
		start--
	}
	return start
}

func sequenceValueRange(source string, found target) (int, int, error) {
	if found.valueNode.Line == found.keyNode.Line {
		return valueRange(source, found)
	}
	keyStart, keyEnd, err := lineRange(source, found.keyNode.Line)
	if err != nil {
		return 0, 0, err
	}
	keyLine := source[keyStart:keyEnd]
	colon := strings.IndexByte(keyLine, ':')
	if colon < 0 {
		return 0, 0, fmt.Errorf("%s key line is not editable", formatPath(found.path))
	}
	_, end, err := nodeRange(source, found.valueNode, max(found.keyNode.Column-1, 0))
	if err != nil {
		return 0, 0, err
	}
	return keyStart + colon + 1, end, nil
}
