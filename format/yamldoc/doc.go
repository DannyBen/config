package yamldoc

import (
	"fmt"
	"reflect"
	"strconv"

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

type semanticSet struct {
	path  []string
	value map[string]string
}

func semanticScalar(path []string, value string) semanticSet {
	return semanticSet{
		path:  path,
		value: map[string]string{formatPath(path): value},
	}
}

func semanticArray(path []string, values []string) semanticSet {
	value := make(map[string]string, len(values))
	for i, item := range values {
		itemPath := appendPath(path, strconv.Itoa(i))
		value[formatPath(itemPath)] = item
	}
	return semanticSet{path: path, value: value}
}

func applyVerifiedSet(source string, planned edit, changes ...semanticSet) (string, error) {
	updated := applyEdit(source, planned)
	return verifySemanticPatch(source, updated, func(expected map[string]string) error {
		for _, change := range changes {
			semanticDeletePrefix(expected, change.path)
			for key, value := range change.value {
				expected[key] = value
			}
		}
		return nil
	})
}

func applyVerifiedUnset(source string, planned edit, path []string) (string, error) {
	updated := applyEdit(source, planned)
	return verifySemanticPatchWithOptions(source, updated, verifyOptions{allowEmptyParentNullsFor: path}, func(expected map[string]string) error {
		semanticDeletePrefix(expected, path)
		return nil
	})
}

func applyVerifiedInsertedPath(source string, planned edit, path []string) (string, error) {
	updated := applyEdit(source, planned)
	got, err := semanticValues(updated)
	if err != nil {
		return "", err
	}
	found := false
	for key := range got {
		itemPath, err := parsePath(key)
		if err != nil {
			continue
		}
		if samePath(itemPath, path) || hasPathPrefix(itemPath, path) {
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("internal YAML patch verification failed")
	}
	return updated, nil
}

func applyVerifiedDelete(source string, planned edit, path []string, reindexes bool) (string, error) {
	updated := applyEdit(source, planned)
	if reindexes {
		if err := validateYAML(updated); err != nil {
			return "", err
		}
		return updated, nil
	}
	return verifySemanticPatch(source, updated, func(expected map[string]string) error {
		semanticDeletePrefix(expected, path)
		return nil
	})
}

func verifySemanticPatch(source, updated string, mutate func(map[string]string) error) (string, error) {
	return verifySemanticPatchWithOptions(source, updated, verifyOptions{}, mutate)
}

type verifyOptions struct {
	allowEmptyParentNullsFor []string
}

func verifySemanticPatchWithOptions(source, updated string, options verifyOptions, mutate func(map[string]string) error) (string, error) {
	expected, err := semanticValues(source)
	if err != nil {
		return "", err
	}
	if err := mutate(expected); err != nil {
		return "", err
	}
	got, err := semanticValues(updated)
	if err != nil {
		return "", err
	}
	if len(options.allowEmptyParentNullsFor) > 0 {
		removeAllowedEmptyParentNulls(got, options.allowEmptyParentNullsFor)
	}
	if !reflect.DeepEqual(expected, got) {
		return "", fmt.Errorf("internal YAML patch verification failed")
	}
	return updated, nil
}

func removeAllowedEmptyParentNulls(values map[string]string, path []string) {
	for i := len(path) - 1; i > 0; i-- {
		parent := path[:i]
		key := formatPath(parent)
		if values[key] == "null" {
			delete(values, key)
		}
	}
}

func semanticValues(source string) (map[string]string, error) {
	root, err := parseYAML(source)
	if err != nil {
		return nil, err
	}
	entries, err := listValues(root, "")
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(entries))
	for _, entry := range entries {
		out[entry.Key] = entry.Value
	}
	return out, nil
}

func semanticDeletePrefix(values map[string]string, path []string) {
	for item := range values {
		itemPath, err := parsePath(item)
		if err != nil {
			continue
		}
		if samePath(itemPath, path) || hasPathPrefix(itemPath, path) {
			delete(values, item)
		}
	}
}

func hasPathPrefix(path, prefix []string) bool {
	if len(path) <= len(prefix) {
		return false
	}
	for i := range prefix {
		if path[i] != prefix[i] {
			return false
		}
	}
	return true
}

func Delete(source, key string, selectors []string) (string, error) {
	return deleteValue(source, key, selectors)
}

func DeleteIfEmpty(source, key string) (string, error) {
	return deleteEmptyValue(source, key)
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
