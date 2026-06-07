package format

import (
	"fmt"

	"github.com/dannyben/config/format/jsondoc"
)

type jsonDocument struct{}

func (jsonDocument) ArrayAdd(source, key string, values []string) (string, error) {
	return "", unsupportedJSON("array add")
}

func (jsonDocument) ArrayDel(source, key string, values []string) (string, error) {
	return "", unsupportedJSON("array del")
}

func (jsonDocument) Delete(source, key string, selectors []string) (string, error) {
	return "", unsupportedJSON("delete")
}

func (jsonDocument) DeleteIfEmpty(source, key string) (string, error) {
	return "", unsupportedJSON("delete --if-empty")
}

func (jsonDocument) Get(source, key string) (string, error) {
	return jsondoc.Get(source, key)
}

func (jsonDocument) GetIn(source, collection string, selectors []string, key string) (string, error) {
	return jsondoc.GetIn(source, collection, selectors, key)
}

func (jsonDocument) Dump(source, key string) (any, error) {
	return jsondoc.Dump(source, key)
}

func (jsonDocument) Set(source, key, value string) (string, error) {
	return jsondoc.Set(source, key, value)
}

func (jsonDocument) SetArray(source, key string, values []string) (string, error) {
	return "", unsupportedJSON("array set")
}

func (jsonDocument) SetIn(source, collection, on, key, value string) (string, error) {
	return jsondoc.SetIn(source, collection, on, key, value)
}

func (jsonDocument) SetInArray(source, collection, on, key string, values []string) (string, error) {
	return "", unsupportedJSON("set --in array")
}

func (jsonDocument) SetInString(source, collection, on, key, value string) (string, error) {
	return jsondoc.SetInString(source, collection, on, key, value)
}

func (jsonDocument) SetString(source, key, value string) (string, error) {
	return jsondoc.SetString(source, key, value)
}

func (jsonDocument) Unset(source, key string) (string, error) {
	return jsondoc.Unset(source, key)
}

func (jsonDocument) UnsetIn(source, collection string, selectors []string, key string) (string, error) {
	return jsondoc.UnsetIn(source, collection, selectors, key)
}

func (jsonDocument) List(source, key string) ([]Entry, error) {
	entries, err := jsondoc.List(source, key)
	if err != nil {
		return nil, err
	}

	out := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, Entry{Key: entry.Key, Value: entry.Value})
	}
	return out, nil
}

func unsupportedJSON(operation string) error {
	return fmt.Errorf("JSON %s is not supported yet", operation)
}
