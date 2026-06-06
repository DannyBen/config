package yamldoc

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func ArrayAdd(source, path string, values []string) (string, error) {
	root, key, err := parseArrayOperation(source, path)
	if err != nil {
		return "", err
	}
	items, ok, err := yamlArrayValues(root, key)
	if err != nil {
		return "", err
	}
	if !ok {
		return SetArray(source, path, values)
	}

	seen := make(map[string]bool, len(items)+len(values))
	next := make([]string, 0, len(items)+len(values))
	for _, item := range items {
		seen[item] = true
		next = append(next, item)
	}
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		next = append(next, value)
	}
	return SetArray(source, path, next)
}

func ArrayDel(source, path string, values []string) (string, error) {
	root, key, err := parseArrayOperation(source, path)
	if err != nil {
		return "", err
	}
	items, ok, err := yamlArrayValues(root, key)
	if err != nil {
		return "", err
	}
	if !ok {
		return source, nil
	}

	remove := make(map[string]bool, len(values))
	for _, value := range values {
		remove[value] = true
	}
	next := make([]string, 0, len(items))
	for _, item := range items {
		if !remove[item] {
			next = append(next, item)
		}
	}
	if len(next) == 0 {
		return Delete(source, path, nil)
	}
	return SetArray(source, path, next)
}

func parseArrayOperation(source, path string) (*yaml.Node, []string, error) {
	root, err := parseYAML(source)
	if err != nil {
		return nil, nil, err
	}
	key, err := parsePath(path)
	if err != nil {
		return nil, nil, err
	}
	if len(key) == 0 {
		return nil, nil, fmt.Errorf("empty path")
	}
	return root, key, nil
}

func yamlArrayValues(root *yaml.Node, key []string) ([]string, bool, error) {
	if root == nil {
		return nil, false, nil
	}
	node, err := resolvePath(root, key, nil)
	if err != nil {
		if isNotSetYAMLError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	node, err = resolveAlias(node, nil)
	if err != nil {
		return nil, true, err
	}
	if node.Kind != yaml.SequenceNode {
		return nil, true, fmt.Errorf("%s is not an array", formatPath(key))
	}
	if isRecordSequence(node) {
		return nil, true, fmt.Errorf("%s is not a scalar array", formatPath(key))
	}
	items := make([]string, 0, len(node.Content))
	for _, item := range node.Content {
		item, err = resolveAlias(item, nil)
		if err != nil {
			return nil, true, err
		}
		if item.Kind != yaml.ScalarNode {
			return nil, true, fmt.Errorf("%s is not a scalar array", formatPath(key))
		}
		items = append(items, scalarValue(item))
	}
	return items, true, nil
}

func isNotSetYAMLError(err error) bool {
	return err != nil && len(err.Error()) >= len(" is not set") && err.Error()[len(err.Error())-len(" is not set"):] == " is not set"
}
