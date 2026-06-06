package format

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Entry struct {
	Key   string
	Value string
}

type Document interface {
	Delete(source, key string, selectors []string) (string, error)
	Get(source, key string) (string, error)
	GetIn(source, collection string, selectors []string, key string) (string, error)
	Set(source, key, value string) (string, error)
	SetArray(source, key string, values []string) (string, error)
	SetIn(source, collection, on, key, value string) (string, error)
	SetInArray(source, collection, on, key string, values []string) (string, error)
	SetInString(source, collection, on, key, value string) (string, error)
	SetString(source, key, value string) (string, error)
	Unset(source, key string) (string, error)
	UnsetIn(source, collection string, selectors []string, key string) (string, error)
	List(source, key string) ([]Entry, error)
}

func Resolve(path string) (Document, string, error) {
	ext := extension(path)
	switch ext {
	case ".toml":
		return tomlDocument{}, "toml", nil
	case ".yaml", ".yml":
		return yamlDocument{}, "yaml", nil
	default:
		return nil, "", fmt.Errorf("unsupported config format for %s", path)
	}
}

func TargetPath(path string) bool {
	switch extension(path) {
	case ".toml", ".yaml", ".yml", ".json", ".ini":
		return true
	default:
		return false
	}
}

func extension(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	return ext
}
