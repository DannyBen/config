package yamldoc

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func listValues(root *yaml.Node, path string) ([]Entry, error) {
	prefix, err := parsePath(path)
	if err != nil {
		return nil, err
	}
	if root == nil {
		if len(prefix) == 0 {
			return nil, nil
		}
		return nil, fmt.Errorf("%s is not set", formatPath(prefix))
	}

	node, err := resolvePath(root, prefix, nil)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := collectEntries(node, prefix, nil, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func parsePath(path string) ([]string, error) {
	if path == "" {
		return nil, nil
	}
	var parts []string
	var current strings.Builder
	for i := 0; i < len(path); i++ {
		if path[i] != '.' {
			current.WriteByte(path[i])
			continue
		}
		if i+1 < len(path) && path[i+1] == '.' {
			current.WriteByte('.')
			i++
			continue
		}
		if current.Len() == 0 {
			return nil, fmt.Errorf("empty path segment in %q", path)
		}
		parts = append(parts, current.String())
		current.Reset()
	}
	if current.Len() == 0 {
		return nil, fmt.Errorf("empty path segment in %q", path)
	}
	parts = append(parts, current.String())
	return parts, nil
}

func resolvePath(node *yaml.Node, path []string, aliases map[*yaml.Node]bool) (*yaml.Node, error) {
	node, err := resolveAlias(node, aliases)
	if err != nil {
		return nil, err
	}
	if len(path) == 0 {
		return node, nil
	}

	switch node.Kind {
	case yaml.MappingNode:
		if err := rejectDuplicateKeys(node); err != nil {
			return nil, err
		}
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			if key.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("non-scalar mapping key at %s", formatPath(path))
			}
			if key.Value == path[0] {
				return resolvePath(node.Content[i+1], path[1:], aliases)
			}
		}
		return nil, fmt.Errorf("%s is not set", formatPath(path))
	case yaml.SequenceNode:
		index, err := strconv.Atoi(path[0])
		if err != nil || index < 0 || index >= len(node.Content) {
			return nil, fmt.Errorf("%s is not set", formatPath(path))
		}
		return resolvePath(node.Content[index], path[1:], aliases)
	default:
		return nil, fmt.Errorf("%s is not set", formatPath(path))
	}
}

func collectEntries(node *yaml.Node, path []string, aliases map[*yaml.Node]bool, entries *[]Entry) error {
	if aliases == nil {
		aliases = make(map[*yaml.Node]bool)
	}
	node, err := resolveAlias(node, aliases)
	if err != nil {
		return err
	}

	switch node.Kind {
	case yaml.MappingNode:
		if err := enterCollection(node, path, aliases); err != nil {
			return err
		}
		defer leaveCollection(node, aliases)
		if err := rejectDuplicateKeys(node); err != nil {
			return err
		}
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			if key.Kind != yaml.ScalarNode {
				return fmt.Errorf("non-scalar mapping key at %s", formatPath(path))
			}
			childPath := appendPath(path, key.Value)
			if err := collectEntries(node.Content[i+1], childPath, aliases, entries); err != nil {
				return err
			}
		}
	case yaml.SequenceNode:
		if err := enterCollection(node, path, aliases); err != nil {
			return err
		}
		defer leaveCollection(node, aliases)
		for i, child := range node.Content {
			childPath := appendPath(path, strconv.Itoa(i))
			if err := collectEntries(child, childPath, aliases, entries); err != nil {
				return err
			}
		}
	case yaml.ScalarNode:
		if len(path) == 0 {
			return fmt.Errorf("document root is a scalar, not a collection")
		}
		*entries = append(*entries, Entry{Key: formatPath(path), Value: scalarValue(node)})
	default:
		return fmt.Errorf("unsupported YAML node at %s", formatPath(path))
	}
	return nil
}

func resolveAlias(node *yaml.Node, aliases map[*yaml.Node]bool) (*yaml.Node, error) {
	seen := make(map[*yaml.Node]bool)
	for node != nil && node.Kind == yaml.AliasNode {
		if node.Alias == nil {
			return nil, fmt.Errorf("unresolved YAML alias")
		}
		if seen[node.Alias] {
			return nil, fmt.Errorf("YAML alias cycle")
		}
		seen[node.Alias] = true
		node = node.Alias
	}
	return node, nil
}

func enterCollection(node *yaml.Node, path []string, aliases map[*yaml.Node]bool) error {
	if aliases == nil {
		return nil
	}
	if aliases[node] {
		return fmt.Errorf("YAML alias cycle at %s", formatPath(path))
	}
	aliases[node] = true
	return nil
}

func leaveCollection(node *yaml.Node, aliases map[*yaml.Node]bool) {
	if aliases != nil {
		delete(aliases, node)
	}
}

func rejectDuplicateKeys(node *yaml.Node) error {
	seen := make(map[string]bool)
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		if key.Kind != yaml.ScalarNode {
			continue
		}
		if seen[key.Value] {
			return fmt.Errorf("duplicate mapping key %q", key.Value)
		}
		seen[key.Value] = true
	}
	return nil
}

func scalarValue(node *yaml.Node) string {
	if node.Tag == "!!null" {
		return "null"
	}
	return node.Value
}

func appendPath(path []string, segment string) []string {
	next := make([]string, len(path), len(path)+1)
	copy(next, path)
	next = append(next, segment)
	return next
}

func formatPath(path []string) string {
	parts := make([]string, len(path))
	for i, segment := range path {
		parts[i] = strings.ReplaceAll(segment, ".", "..")
	}
	return strings.Join(parts, ".")
}
