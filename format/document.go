package format

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dannyben/config/format/inidoc"
	"github.com/dannyben/config/format/jsondoc"
	"github.com/dannyben/config/format/tomldoc"
	"github.com/dannyben/config/format/yamldoc"
)

type Entry struct {
	Key   string
	Value string
}

type Document interface {
	ArrayAdd(source, key string, values []string) (string, error)
	ArrayDel(source, key string, values []string) (string, error)
	Delete(source, key string, selectors []string) (string, error)
	DeleteIfEmpty(source, key string) (string, error)
	Get(source, key string) (string, error)
	GetIn(source, collection string, selectors []string, key string) (string, error)
	Dump(source, key string) (any, error)
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
		return documentForFormat("toml")
	case ".yaml", ".yml":
		return documentForFormat("yaml")
	case ".json":
		return documentForFormat("json")
	case ".ini":
		return documentForFormat("ini")
	default:
		return nil, "", fmt.Errorf("unsupported config format for %s", path)
	}
}

func ResolveSource(path string, source []byte) (Document, string, error) {
	if doc, name, err := Resolve(path); err == nil {
		return doc, name, nil
	}

	text := string(source)
	if hinted, ok := formatHint(text); ok {
		if hinted == "json" {
			return nil, "", fmt.Errorf("unsupported format hint %q for %s; JSON files cannot contain comments", hinted, path)
		}
		doc, name, err := documentForFormat(hinted)
		if err != nil {
			return nil, "", fmt.Errorf("unsupported format hint %q for %s", hinted, path)
		}
		return doc, name, nil
	}

	possibleTOML := tomldoc.Valid(text)
	possibleINI := inidoc.Valid(text)
	switch {
	case possibleTOML && possibleINI:
		return nil, "", fmt.Errorf("ambiguous config format for %s; add # format: toml or # format: ini", path)
	case possibleTOML:
		return documentForFormat("toml")
	case possibleINI:
		return documentForFormat("ini")
	}

	if jsondoc.Valid(text) {
		return documentForFormat("json")
	}
	if yamldoc.Valid(text) {
		return documentForFormat("yaml")
	}

	return nil, "", fmt.Errorf("cannot determine config format for %s; add # format: toml, # format: ini, or # format: yaml", path)
}

func documentForFormat(format string) (Document, string, error) {
	switch format {
	case "toml":
		return tomlDocument{}, "toml", nil
	case "yaml", "yml":
		return yamlDocument{}, "yaml", nil
	case "json":
		return jsonDocument{}, "json", nil
	case "ini":
		return iniDocument{}, "ini", nil
	default:
		return nil, "", fmt.Errorf("unsupported config format %q", format)
	}
}

func formatHint(source string) (string, bool) {
	for _, line := range strings.Split(source, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "#") {
			return "", false
		}
		hint := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
		name, ok := strings.CutPrefix(hint, "format:")
		if !ok {
			return "", false
		}
		return strings.ToLower(strings.TrimSpace(name)), true
	}
	return "", false
}

func extension(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	return ext
}
