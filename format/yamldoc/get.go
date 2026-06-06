package yamldoc

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type selector struct {
	path    []string
	compare string
	label   string
}

func getValue(root *yaml.Node, path string) (string, error) {
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
	node, err := resolvePath(root, key, nil)
	if err != nil {
		return "", err
	}
	return renderGetValue(node, key)
}

func getValueIn(root *yaml.Node, collectionPath string, rawSelectors []string, path string) (string, error) {
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

	node, err := resolvePath(root, collection, nil)
	if err != nil {
		return "", err
	}
	node, err = resolveAlias(node, nil)
	if err != nil {
		return "", err
	}
	if node.Kind != yaml.SequenceNode || !isRecordSequence(node) {
		return "", fmt.Errorf("%s is not a sequence of records", formatPath(collection))
	}

	selectors, err := parseSelectors(rawSelectors)
	if err != nil {
		return "", err
	}
	index, err := selectedRecordIndex(node, collection, selectors)
	if err != nil {
		return "", err
	}
	record, err := resolveAlias(node.Content[index], nil)
	if err != nil {
		return "", err
	}
	valueNode, err := resolvePath(record, key, nil)
	if err != nil {
		fullPath := appendPath(collection, fmt.Sprintf("%d", index))
		fullPath = append(fullPath, key...)
		return "", fmt.Errorf("%s is not set", formatPath(fullPath))
	}
	fullPath := appendPath(collection, fmt.Sprintf("%d", index))
	fullPath = append(fullPath, key...)
	return renderGetValue(valueNode, fullPath)
}

func renderGetValue(node *yaml.Node, path []string) (string, error) {
	node, err := resolveAlias(node, nil)
	if err != nil {
		return "", err
	}
	switch node.Kind {
	case yaml.ScalarNode:
		return scalarValue(node), nil
	case yaml.SequenceNode:
		if isRecordSequence(node) {
			return "", fmt.Errorf("%s is a sequence of records, not a value", formatPath(path))
		}
		return flowValue(node)
	case yaml.MappingNode:
		return "", fmt.Errorf("%s is a mapping, not a value", formatPath(path))
	default:
		return "", fmt.Errorf("unsupported YAML node at %s", formatPath(path))
	}
}

func flowValue(node *yaml.Node) (string, error) {
	clone, err := cloneFlowNode(node, make(map[*yaml.Node]bool))
	if err != nil {
		return "", err
	}
	content, err := yaml.Marshal(clone)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

func cloneFlowNode(node *yaml.Node, seen map[*yaml.Node]bool) (*yaml.Node, error) {
	node, err := resolveAlias(node, seen)
	if err != nil {
		return nil, err
	}
	if seen[node] {
		return nil, fmt.Errorf("YAML alias cycle")
	}
	seen[node] = true
	defer delete(seen, node)

	clone := *node
	clone.Anchor = ""
	clone.Alias = nil
	if clone.Kind == yaml.SequenceNode || clone.Kind == yaml.MappingNode {
		clone.Style = yaml.FlowStyle
	}
	clone.Content = make([]*yaml.Node, 0, len(node.Content))
	for _, child := range node.Content {
		childClone, err := cloneFlowNode(child, seen)
		if err != nil {
			return nil, err
		}
		clone.Content = append(clone.Content, childClone)
	}
	return &clone, nil
}

func isRecordSequence(node *yaml.Node) bool {
	if node.Kind != yaml.SequenceNode {
		return false
	}
	for _, child := range node.Content {
		resolved, err := resolveAlias(child, nil)
		if err != nil {
			return false
		}
		if resolved.Kind == yaml.MappingNode {
			return true
		}
	}
	return false
}

func parseSelectors(raw []string) ([]selector, error) {
	selectors := make([]selector, 0, len(raw))
	for _, item := range raw {
		onPath, onRaw, ok := strings.Cut(item, ":")
		if !ok || onPath == "" {
			return nil, fmt.Errorf("--on must use FIELD:VALUE")
		}
		path, err := parsePath(onPath)
		if err != nil {
			return nil, fmt.Errorf("--on: %w", err)
		}
		selectors = append(selectors, selector{
			path:    path,
			compare: onRaw,
			label:   formatPath(path) + ":" + onRaw,
		})
	}
	return selectors, nil
}

func selectedRecordIndex(collection *yaml.Node, collectionPath []string, selectors []selector) (int, error) {
	var matches []int
	for i, child := range collection.Content {
		record, err := resolveAlias(child, nil)
		if err != nil {
			return 0, err
		}
		if record.Kind != yaml.MappingNode {
			return 0, fmt.Errorf("%s.%d is not a record", formatPath(collectionPath), i)
		}
		matched, err := recordMatches(record, selectors)
		if err != nil {
			return 0, err
		}
		if matched {
			matches = append(matches, i)
		}
	}

	label := joinSelectorLabels(selectors)
	switch len(matches) {
	case 0:
		return 0, fmt.Errorf("%s has no records matching %s", formatPath(collectionPath), label)
	case 1:
		return matches[0], nil
	default:
		return 0, fmt.Errorf("%s has multiple records matching %s", formatPath(collectionPath), label)
	}
}

func recordMatches(record *yaml.Node, selectors []selector) (bool, error) {
	for _, selector := range selectors {
		node, err := resolvePath(record, selector.path, nil)
		if err != nil {
			return false, nil
		}
		value, err := scalarSelectorValue(node, selector.path)
		if err != nil {
			return false, nil
		}
		if value != selector.compare {
			return false, nil
		}
	}
	return true, nil
}

func scalarSelectorValue(node *yaml.Node, path []string) (string, error) {
	node, err := resolveAlias(node, nil)
	if err != nil {
		return "", err
	}
	if node.Kind != yaml.ScalarNode {
		return "", fmt.Errorf("%s is not a scalar selector value", formatPath(path))
	}
	return scalarValue(node), nil
}

func joinSelectorLabels(selectors []selector) string {
	labels := make([]string, 0, len(selectors))
	for _, selector := range selectors {
		labels = append(labels, selector.label)
	}
	return strings.Join(labels, ",")
}
