package inidoc

import (
	"fmt"
	"strings"
)

type Entry struct {
	Key   string
	Value string
}

type document struct {
	entries []entry
}

type entry struct {
	section string
	key     string
	value   string
}

func Get(source, path string) (string, error) {
	section, key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	if key == "" {
		return "", fmt.Errorf("INI get requires a key")
	}

	doc, err := parse(source)
	if err != nil {
		return "", err
	}

	var values []string
	for _, entry := range doc.entries {
		if entry.section == section && entry.key == key {
			values = append(values, entry.value)
		}
	}
	if len(values) == 0 {
		return "", fmt.Errorf("%s is not set", formatPath(section, key))
	}
	if len(values) > 1 {
		return "", fmt.Errorf("%s has multiple values", formatPath(section, key))
	}
	return values[0], nil
}

func Dump(source, path string) (any, error) {
	section, key, err := parsePath(path)
	if err != nil {
		return nil, err
	}

	doc, err := parse(source)
	if err != nil {
		return nil, err
	}
	return doc.dump(section, key)
}

func List(source, path string) ([]Entry, error) {
	section, key, err := parsePath(path)
	if err != nil {
		return nil, err
	}

	doc, err := parse(source)
	if err != nil {
		return nil, err
	}
	return doc.list(section, key)
}

func (doc document) list(section, key string) ([]Entry, error) {
	var entries []Entry
	if section == "" && key != "" {
		var keyEntries []Entry
		var sectionEntries []Entry
		for _, item := range doc.entries {
			switch {
			case item.section == "" && item.key == key:
				keyEntries = append(keyEntries, Entry{Key: formatPath(item.section, item.key), Value: item.value})
			case item.section == key:
				sectionEntries = append(sectionEntries, Entry{Key: formatPath(item.section, item.key), Value: item.value})
			}
		}
		if len(keyEntries) != 0 && len(sectionEntries) != 0 {
			return nil, fmt.Errorf("%s matches both a key and section", formatPath("", key))
		}
		if len(sectionEntries) != 0 {
			return sectionEntries, nil
		}
		if len(keyEntries) != 0 {
			return keyEntries, nil
		}
		return nil, fmt.Errorf("%s is not set", formatPath(section, key))
	}

	for _, item := range doc.entries {
		switch {
		case section == "" && key == "":
			entries = append(entries, Entry{Key: formatPath(item.section, item.key), Value: item.value})
		case key == "" && item.section == section:
			entries = append(entries, Entry{Key: formatPath(item.section, item.key), Value: item.value})
		case item.section == section && item.key == key:
			entries = append(entries, Entry{Key: formatPath(item.section, item.key), Value: item.value})
		}
	}
	if len(entries) == 0 {
		if key == "" {
			return nil, fmt.Errorf("%s is not set", formatPath(section, key))
		}
		return nil, fmt.Errorf("%s is not set", formatPath(section, key))
	}
	return entries, nil
}

func (doc document) dump(section, key string) (any, error) {
	if section == "" && key == "" {
		return doc.dumpAll()
	}
	if section == "" {
		var values []string
		sectionMap := make(map[string]any)
		for _, entry := range doc.entries {
			switch {
			case entry.section == "" && entry.key == key:
				values = append(values, entry.value)
			case entry.section == key:
				if _, ok := sectionMap[entry.key]; ok {
					return nil, fmt.Errorf("%s has multiple values", formatPath(entry.section, entry.key))
				}
				sectionMap[entry.key] = entry.value
			}
		}
		if len(values) != 0 && len(sectionMap) != 0 {
			return nil, fmt.Errorf("%s matches both a key and section", formatPath("", key))
		}
		if len(values) > 1 {
			return nil, fmt.Errorf("%s has multiple values", formatPath("", key))
		}
		if len(values) == 1 {
			return values[0], nil
		}
		if len(sectionMap) != 0 {
			return sectionMap, nil
		}
		return nil, fmt.Errorf("%s is not set", formatPath("", key))
	}

	var values []string
	for _, entry := range doc.entries {
		if entry.section == section && entry.key == key {
			values = append(values, entry.value)
		}
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("%s is not set", formatPath(section, key))
	}
	if len(values) > 1 {
		return nil, fmt.Errorf("%s has multiple values", formatPath(section, key))
	}
	return values[0], nil
}

func (doc document) dumpAll() (map[string]any, error) {
	out := make(map[string]any)
	for _, entry := range doc.entries {
		if entry.section == "" {
			if existing, ok := out[entry.key]; ok {
				if _, ok := existing.(map[string]any); ok {
					return nil, fmt.Errorf("%s matches both a key and section", formatPath("", entry.key))
				}
				return nil, fmt.Errorf("%s has multiple values", formatPath("", entry.key))
			}
			out[entry.key] = entry.value
			continue
		}

		existing, ok := out[entry.section]
		if !ok {
			sectionMap := map[string]any{entry.key: entry.value}
			out[entry.section] = sectionMap
			continue
		}
		sectionMap, ok := existing.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s matches both a key and section", formatPath("", entry.section))
		}
		if _, ok := sectionMap[entry.key]; ok {
			return nil, fmt.Errorf("%s has multiple values", formatPath(entry.section, entry.key))
		}
		sectionMap[entry.key] = entry.value
	}
	return out, nil
}

func parse(source string) (document, error) {
	var doc document
	var section string
	lines := strings.Split(source, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}

		if strings.HasPrefix(trimmed, "[") {
			if !strings.HasSuffix(trimmed, "]") {
				return document{}, fmt.Errorf("invalid INI section on line %d", i+1)
			}
			name := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"))
			if name == "" {
				return document{}, fmt.Errorf("empty INI section on line %d", i+1)
			}
			section = name
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			key = line
			value = ""
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return document{}, fmt.Errorf("empty INI key on line %d", i+1)
		}
		doc.entries = append(doc.entries, entry{
			section: section,
			key:     key,
			value:   strings.TrimSpace(value),
		})
	}
	return doc, nil
}

func parsePath(path string) (string, string, error) {
	if path == "" {
		return "", "", nil
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
			return "", "", fmt.Errorf("empty INI path segment in %q", path)
		}
		parts = append(parts, current.String())
		if len(parts) > 1 {
			return "", "", fmt.Errorf("INI paths support only key or section.key")
		}
		current.Reset()
	}
	if current.Len() == 0 {
		return "", "", fmt.Errorf("empty INI path segment in %q", path)
	}
	parts = append(parts, current.String())

	if len(parts) == 1 {
		return "", parts[0], nil
	}
	return parts[0], parts[1], nil
}

func formatPath(section, key string) string {
	if section == "" {
		return escapePathSegment(key)
	}
	if key == "" {
		return escapePathSegment(section)
	}
	return escapePathSegment(section) + "." + escapePathSegment(key)
}

func escapePathSegment(segment string) string {
	return strings.ReplaceAll(segment, ".", "..")
}
