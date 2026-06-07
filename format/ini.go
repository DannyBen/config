package format

import (
	"fmt"

	"github.com/dannyben/config/format/inidoc"
)

type iniDocument struct{}

func (iniDocument) ArrayAdd(source, key string, values []string) (string, error) {
	return "", unsupportedINI("array add")
}

func (iniDocument) ArrayDel(source, key string, values []string) (string, error) {
	return "", unsupportedINI("array del")
}

func (iniDocument) Delete(source, key string, selectors []string) (string, error) {
	return "", unsupportedINI("delete")
}

func (iniDocument) DeleteIfEmpty(source, key string) (string, error) {
	return "", unsupportedINI("delete --empty")
}

func (iniDocument) Get(source, key string) (string, error) {
	return inidoc.Get(source, key)
}

func (iniDocument) GetIn(source, collection string, selectors []string, key string) (string, error) {
	return "", unsupportedINI("get --in")
}

func (iniDocument) Dump(source, key string) (any, error) {
	return inidoc.Dump(source, key)
}

func (iniDocument) Set(source, key, value string) (string, error) {
	return "", unsupportedINI("set")
}

func (iniDocument) SetArray(source, key string, values []string) (string, error) {
	return "", unsupportedINI("array set")
}

func (iniDocument) SetIn(source, collection, on, key, value string) (string, error) {
	return "", unsupportedINI("set --in")
}

func (iniDocument) SetInArray(source, collection, on, key string, values []string) (string, error) {
	return "", unsupportedINI("set --in array")
}

func (iniDocument) SetInString(source, collection, on, key, value string) (string, error) {
	return "", unsupportedINI("set --in --string")
}

func (iniDocument) SetString(source, key, value string) (string, error) {
	return "", unsupportedINI("set --string")
}

func (iniDocument) Unset(source, key string) (string, error) {
	return "", unsupportedINI("unset")
}

func (iniDocument) UnsetIn(source, collection string, selectors []string, key string) (string, error) {
	return "", unsupportedINI("unset --in")
}

func (iniDocument) List(source, key string) ([]Entry, error) {
	entries, err := inidoc.List(source, key)
	if err != nil {
		return nil, err
	}

	out := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, Entry{Key: entry.Key, Value: entry.Value})
	}
	return out, nil
}

func unsupportedINI(operation string) error {
	return fmt.Errorf("INI %s is not supported yet", operation)
}
