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
	lines   []parsedLine
}

type entry struct {
	section string
	key     string
	value   string
	line    int
}

type parsedLine struct {
	kind    lineKind
	section string
	key     string
}

type lineKind int

const (
	lineOther lineKind = iota
	lineSection
	lineEntry
)

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

func Set(source, path, value string) (string, error) {
	section, key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	if key == "" {
		return "", fmt.Errorf("INI set requires a key")
	}

	doc, err := parse(source)
	if err != nil {
		return "", err
	}
	return doc.set(source, section, key, value)
}

func Unset(source, path string) (string, error) {
	section, key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	if key == "" {
		return "", fmt.Errorf("INI unset requires a key")
	}

	doc, err := parse(source)
	if err != nil {
		return "", err
	}
	return doc.unset(source, section, key)
}

func Delete(source, path string, selectors []string) (string, error) {
	if len(selectors) != 0 {
		return "", fmt.Errorf("INI delete --on is not supported")
	}
	section, key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	if section != "" || key == "" {
		return "", fmt.Errorf("%s is a value, use unset to remove fields", formatPath(section, key))
	}

	doc, err := parse(source)
	if err != nil {
		return "", err
	}
	return doc.deleteSection(source, key, false)
}

func DeleteIfEmpty(source, path string) (string, error) {
	section, key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	if section != "" || key == "" {
		return "", fmt.Errorf("%s is a value, use unset to remove fields", formatPath(section, key))
	}

	doc, err := parse(source)
	if err != nil {
		return "", err
	}
	return doc.deleteSection(source, key, true)
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

func (doc document) set(source, section, key, value string) (string, error) {
	if section == "" && doc.hasSection(key) {
		return "", fmt.Errorf("%s is a section, not a value", formatPath("", key))
	}

	matches := doc.findEntries(section, key)
	if len(matches) > 1 {
		return "", fmt.Errorf("%s has multiple values", formatPath(section, key))
	}
	if len(matches) == 1 {
		return replaceLineValue(source, matches[0].line, value)
	}
	if section == "" {
		return doc.insertGlobalValue(source, key, value), nil
	}
	return doc.insertSectionValue(source, section, key, value)
}

func (doc document) unset(source, section, key string) (string, error) {
	if section == "" && doc.hasSection(key) {
		return "", fmt.Errorf("%s is a section, not a value", formatPath("", key))
	}

	matches := doc.findEntries(section, key)
	if len(matches) == 0 {
		return "", fmt.Errorf("%s is not set", formatPath(section, key))
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("%s has multiple values", formatPath(section, key))
	}
	return deleteLine(source, matches[0].line)
}

func (doc document) deleteSection(source, section string, onlyIfEmpty bool) (string, error) {
	if len(doc.findEntries("", section)) != 0 {
		return "", fmt.Errorf("%s is a value, use unset to remove fields", formatPath("", section))
	}

	start, end := doc.sectionBounds(section)
	if start == -1 {
		if onlyIfEmpty {
			return source, nil
		}
		return "", fmt.Errorf("%s is not set", formatPath("", section))
	}
	if onlyIfEmpty && doc.sectionHasEntries(section) {
		return source, nil
	}

	lines, trailing := splitLines(source)
	removeStart := start
	removeEnd := end
	for removeEnd < len(lines) && strings.TrimSpace(lines[removeEnd]) == "" {
		removeEnd++
	}
	lines = append(lines[:removeStart], lines[removeEnd:]...)
	return joinLines(lines, trailing), nil
}

func (doc document) sectionHasEntries(section string) bool {
	for _, entry := range doc.entries {
		if entry.section == section {
			return true
		}
	}
	return false
}

func (doc document) findEntries(section, key string) []entry {
	var matches []entry
	for _, entry := range doc.entries {
		if entry.section == section && entry.key == key {
			matches = append(matches, entry)
		}
	}
	return matches
}

func (doc document) hasSection(section string) bool {
	for _, line := range doc.lines {
		if line.kind == lineSection && line.section == section {
			return true
		}
	}
	return false
}

func (doc document) insertGlobalValue(source, key, value string) string {
	firstSection := -1
	for i, line := range doc.lines {
		if line.kind == lineSection {
			firstSection = i
			break
		}
	}
	if firstSection == -1 {
		return appendGlobal(source, key, value)
	}

	lines, trailing := splitLines(source)
	insertAt := firstSection
	for insertAt > 0 && strings.TrimSpace(lines[insertAt-1]) == "" {
		insertAt--
	}
	newLine := key + " = " + value
	lines = append(lines[:insertAt], append([]string{newLine}, lines[insertAt:]...)...)
	return joinLines(lines, trailing)
}

func (doc document) insertSectionValue(source, section, key, value string) (string, error) {
	start, end := doc.sectionBounds(section)
	if start == -1 {
		return appendSection(source, section, key, value), nil
	}

	lines, trailing := splitLines(source)
	newLine := key + " = " + value
	lines = append(lines[:end], append([]string{newLine}, lines[end:]...)...)
	return joinLines(lines, trailing), nil
}

func (doc document) sectionBounds(section string) (int, int) {
	start := -1
	end := -1
	for i, line := range doc.lines {
		if line.kind != lineSection {
			continue
		}
		if start == -1 {
			if line.section == section {
				start = i
				end = len(doc.lines)
			}
			continue
		}
		end = i
		break
	}
	return start, end
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
	lines, _ := splitLines(source)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			doc.lines = append(doc.lines, parsedLine{kind: lineOther})
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
			doc.lines = append(doc.lines, parsedLine{kind: lineSection, section: section})
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
		doc.lines = append(doc.lines, parsedLine{kind: lineEntry, section: section, key: key})
		doc.entries = append(doc.entries, entry{
			section: section,
			key:     key,
			value:   strings.TrimSpace(value),
			line:    i,
		})
	}
	return doc, nil
}

func Valid(source string) bool {
	doc, err := parse(source)
	if err != nil {
		return false
	}
	return len(doc.entries) > 0 && hasINIShape(source) && !hasTOMLArrayTable(source)
}

func hasINIShape(source string) bool {
	lines, _ := splitLines(source)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}
		if strings.HasPrefix(trimmed, "[") || strings.Contains(line, "=") {
			return true
		}
	}
	return false
}

func hasTOMLArrayTable(source string) bool {
	lines, _ := splitLines(source)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[[") {
			return true
		}
	}
	return false
}

func replaceLineValue(source string, index int, value string) (string, error) {
	lines, trailing := splitLines(source)
	if index < 0 || index >= len(lines) {
		return "", fmt.Errorf("INI key line is not editable")
	}
	before, _, ok := strings.Cut(lines[index], "=")
	if !ok {
		lines[index] = strings.TrimSpace(before) + " = " + value
		return joinLines(lines, trailing), nil
	}
	lines[index] = before + "= " + value
	return joinLines(lines, trailing), nil
}

func deleteLine(source string, index int) (string, error) {
	lines, trailing := splitLines(source)
	if index < 0 || index >= len(lines) {
		return "", fmt.Errorf("INI key line is not editable")
	}
	lines = append(lines[:index], lines[index+1:]...)
	return joinLines(lines, trailing), nil
}

func appendGlobal(source, key, value string) string {
	line := key + " = " + value
	if source == "" {
		return line + "\n"
	}
	if strings.HasSuffix(source, "\n") {
		return source + line + "\n"
	}
	return source + "\n" + line + "\n"
}

func appendSection(source, section, key, value string) string {
	block := "[" + section + "]\n" + key + " = " + value + "\n"
	if source == "" {
		return block
	}
	if strings.HasSuffix(source, "\n\n") {
		return source + block
	}
	if strings.HasSuffix(source, "\n") {
		return source + "\n" + block
	}
	return source + "\n\n" + block
}

func splitLines(source string) ([]string, bool) {
	trailing := strings.HasSuffix(source, "\n")
	if trailing {
		source = strings.TrimSuffix(source, "\n")
	}
	if source == "" {
		return nil, trailing
	}
	return strings.Split(source, "\n"), trailing
}

func joinLines(lines []string, trailing bool) string {
	if len(lines) == 0 {
		return ""
	}
	out := strings.Join(lines, "\n")
	if trailing || out != "" {
		out += "\n"
	}
	return out
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
