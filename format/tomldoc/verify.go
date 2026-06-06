package tomldoc

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/creachadair/tomledit/parser"
	"github.com/creachadair/tomledit/scanner"
	"github.com/neongreen/tomlsawyer"
)

func joinTokenText(tokens []token) string {
	var out strings.Builder
	for _, tok := range tokens {
		out.WriteString(tok.text)
	}
	return out.String()
}

func formatValue(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if parsed, err := parser.ParseValue(value); err == nil && isScalarValue(parsed) {
		return value, nil
	}
	return formatString(raw)
}

func isScalarValue(value parser.Value) bool {
	_, ok := value.X.(parser.Token)
	return ok
}

func formatString(value string) (string, error) {
	if strings.ContainsAny(value, "\r\n") {
		formatted := `"""` + string(scanner.EscapeMultiline(value)) + `"""`
		if _, err := parser.ParseValue(formatted); err != nil {
			return "", err
		}
		return formatted, nil
	}
	formatted, err := tomlsawyer.FormatValueToString(value)
	if err != nil {
		return "", err
	}
	if _, err := parser.ParseValue(formatted); err != nil {
		return "", err
	}
	return formatted, nil
}

func formatArray(values []string) (string, error) {
	items := make([]string, 0, len(values))
	for _, value := range values {
		formatted, err := formatValue(value)
		if err != nil {
			return "", err
		}
		items = append(items, formatted)
	}
	formatted := "[" + strings.Join(items, ", ") + "]"
	if _, err := parser.ParseValue(formatted); err != nil {
		return "", err
	}
	return formatted, nil
}

func appendLine(source, text string) edit {
	nl := lineEnding(source)
	prefix := ""
	if source != "" && !strings.HasSuffix(source, nl) {
		prefix = nl
	}
	return edit{start: len(source), end: len(source), text: prefix + text + nl}
}

func applyVerifiedSet(source string, planned edit, key parser.Key, value string) (string, error) {
	return applyVerifiedSets(source, planned, []semanticSet{{key: key, value: value}})
}

func applyVerifiedSets(source string, planned edit, changes []semanticSet) (string, error) {
	updated := applyEdit(source, planned)
	return verifySemanticPatch(source, updated, func(expected map[string]string) error {
		for _, change := range changes {
			if err := semanticSetValue(expected, change.key, change.value); err != nil {
				return err
			}
		}
		return nil
	})
}

func applyVerifiedUnset(source string, planned edit, key parser.Key) (string, error) {
	updated := applyEdit(source, planned)
	updated, err := verifySemanticPatch(source, updated, func(expected map[string]string) error {
		semanticDeletePrefix(expected, key)
		return nil
	})
	if err != nil {
		return "", err
	}
	return pruneEmptyParentTables(updated, key[:len(key)-1])
}

func applyVerifiedDelete(source string, planned []edit, key parser.Key, selectors []string) (string, error) {
	updated := applyEdits(source, planned)
	if len(selectors) > 0 || indexedRecordDelete(key) {
		if err := parseTOMLSource(updated); err != nil {
			return "", err
		}
		return updated, nil
	}
	return verifySemanticPatch(source, updated, func(expected map[string]string) error {
		for item := range expected {
			itemKey, err := parsePath(item)
			if err != nil {
				return err
			}
			if itemKey.Equals(key) || hasKeyPrefix(itemKey, key) {
				delete(expected, item)
			}
		}
		return nil
	})
}

func semanticSetValue(values map[string]string, key parser.Key, value string) error {
	semanticDeletePrefix(values, key)
	parsed, err := parser.ParseValue(value)
	if err != nil {
		return err
	}
	if array, ok := parsed.X.(parser.Array); ok {
		if len(array) == 0 {
			values[formatListKey(key)] = "[]"
			return nil
		}
		for i, item := range array {
			parsedItem, ok := item.(parser.Value)
			if !ok {
				continue
			}
			logical, err := formatParsedGetValue(parsedItem)
			if err != nil {
				return err
			}
			itemKey := append(append(parser.Key{}, key...), strconv.Itoa(i))
			values[formatListKey(itemKey)] = logical
		}
		return nil
	}
	logical, err := formatParsedGetValue(parsed)
	if err != nil {
		return err
	}
	values[formatListKey(key)] = logical
	return nil
}

func semanticDeletePrefix(values map[string]string, key parser.Key) {
	for item := range values {
		itemKey, err := parsePath(item)
		if err != nil {
			continue
		}
		if itemKey.Equals(key) || hasKeyPrefix(itemKey, key) {
			delete(values, item)
		}
	}
}

func pruneEmptyParentTables(source string, parent parser.Key) (string, error) {
	current := source
	for len(parent) > 0 {
		pruned, ok, err := pruneEmptyTable(current, parent)
		if err != nil {
			return "", err
		}
		if !ok {
			parent = parent[:len(parent)-1]
			continue
		}
		next, err := verifySemanticPatch(current, pruned, func(map[string]string) error {
			return nil
		})
		if err != nil {
			return "", err
		}
		current = next
		parent = parent[:len(parent)-1]
	}
	return current, nil
}

func pruneEmptyTable(source string, key parser.Key) (string, bool, error) {
	assignments, sections, err := scan(source)
	if err != nil {
		return "", false, err
	}
	if !explicitEmptyTable(assignments, sections, key) {
		return source, false, nil
	}
	start, end, ok := tableDeleteRange(source, key)
	if !ok {
		return "", false, fmt.Errorf("%s cannot be safely deleted", key.String())
	}
	return applyEdits(source, []edit{{start: start, end: end, text: ""}}), true, nil
}

func explicitEmptyTable(assignments []assignment, sections []section, key parser.Key) bool {
	found := false
	for _, sec := range sections {
		if sec.key.Equals(key) {
			if sec.inline {
				return false
			}
			found = true
			continue
		}
		if hasKeyPrefix(sec.key, key) {
			return false
		}
	}
	if !found {
		return false
	}
	for _, item := range assignments {
		if item.key.Equals(key) || hasKeyPrefix(item.key, key) {
			return false
		}
	}
	return true
}

func indexedRecordDelete(key parser.Key) bool {
	_, _, ok := indexedRecordKey(key)
	return ok
}

func verifySemanticPatch(source, updated string, mutate func(map[string]string) error) (string, error) {
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
	if !sameSemanticValues(expected, got) {
		return "", fmt.Errorf("internal TOML patch verification failed")
	}
	return updated, nil
}

func semanticValues(source string) (map[string]string, error) {
	if err := parseTOMLSource(source); err != nil {
		return nil, err
	}
	entries, err := listValues(source, nil)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(entries))
	for _, entry := range entries {
		out[entry.Key] = entry.Value
	}
	return out, nil
}

func sameSemanticValues(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		if b[key] != value {
			return false
		}
	}
	return true
}

func applyEdit(source string, e edit) string {
	return source[:e.start] + e.text + source[e.end:]
}

func applyEdits(source string, edits []edit) string {
	sort.SliceStable(edits, func(i, j int) bool {
		return edits[i].start > edits[j].start
	})
	out := source
	for _, e := range edits {
		out = applyEdit(out, e)
	}
	return out
}
