package yamldoc

import "gopkg.in/yaml.v3"

func Dump(source, path string) (any, error) {
	root, err := parseYAML(source)
	if err != nil {
		return nil, err
	}
	key, err := parsePath(path)
	if err != nil {
		return nil, err
	}

	var node *yaml.Node
	if root == nil {
		node = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}
	} else {
		node, err = resolvePath(root, key, nil)
		if err != nil {
			return nil, err
		}
	}

	node, err = resolveAlias(node, nil)
	if err != nil {
		return nil, err
	}
	var value any
	if err := node.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}
