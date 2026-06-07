package yamldoc

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

func Dump(source, path string) (string, error) {
	root, err := parseYAML(source)
	if err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}

	var node *yaml.Node
	if root == nil {
		node = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}
	} else {
		node, err = resolvePath(root, key, nil)
		if err != nil {
			return "", err
		}
	}

	node, err = resolveAlias(node, nil)
	if err != nil {
		return "", err
	}
	var value any
	if err := node.Decode(&value); err != nil {
		return "", err
	}
	out, err := marshalYAMLDump(value)
	if err != nil {
		return "", err
	}
	return out, nil
}

func marshalYAMLDump(value any) (string, error) {
	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(2)
	if err := encoder.Encode(value); err != nil {
		encoder.Close()
		return "", err
	}
	if err := encoder.Close(); err != nil {
		return "", err
	}
	return out.String(), nil
}
