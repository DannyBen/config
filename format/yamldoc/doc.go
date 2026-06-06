package yamldoc

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Entry struct {
	Key   string
	Value string
}

func validateYAML(source string) error {
	_, err := parseYAML(source)
	return err
}

func parseYAML(source string) (*yaml.Node, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(source), &doc); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}
	if len(doc.Content) == 0 {
		return nil, nil
	}
	return doc.Content[0], nil
}

func unsupported(operation string) error {
	return fmt.Errorf("YAML %s is not supported yet", operation)
}

func parseThenUnsupported(source, operation string) error {
	if err := validateYAML(source); err != nil {
		return err
	}
	return unsupported(operation)
}

func Delete(source, key string, selectors []string) (string, error) {
	return deleteValue(source, key, selectors)
}

func Get(source, key string) (string, error) {
	root, err := parseYAML(source)
	if err != nil {
		return "", err
	}
	return getValue(root, key)
}

func GetIn(source, collection string, selectors []string, key string) (string, error) {
	root, err := parseYAML(source)
	if err != nil {
		return "", err
	}
	return getValueIn(root, collection, selectors, key)
}

func Set(source, key, value string) (string, error) {
	return setValue(source, key, value, setModeInfer)
}

func SetArray(source, key string, values []string) (string, error) {
	return setArrayValue(source, key, values)
}

func SetIn(source, collection, on, key, value string) (string, error) {
	return setValueIn(source, collection, on, key, value, setModeInfer)
}

func SetInArray(source, collection, on, key string, values []string) (string, error) {
	return setArrayValueIn(source, collection, on, key, values)
}

func SetInString(source, collection, on, key, value string) (string, error) {
	return setValueIn(source, collection, on, key, value, setModeString)
}

func SetString(source, key, value string) (string, error) {
	return setValue(source, key, value, setModeString)
}

func Unset(source, key string) (string, error) {
	return unsetValue(source, key)
}

func UnsetIn(source, collection string, selectors []string, key string) (string, error) {
	return unsetValueIn(source, collection, selectors, key)
}

func List(source, key string) ([]Entry, error) {
	root, err := parseYAML(source)
	if err != nil {
		return nil, err
	}
	return listValues(root, key)
}
