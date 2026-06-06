package yamldoc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type setMode int

const (
	setModeInfer setMode = iota
	setModeString
)

type edit struct {
	start int
	end   int
	text  string
}

type target struct {
	keyNode   *yaml.Node
	valueNode *yaml.Node
	parent    *yaml.Node
	key       string
	path      []string
}

func setValue(source, path, raw string, mode setMode) (string, error) {
	root, err := parseYAML(source)
	if err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	if len(key) == 0 {
		return "", fmt.Errorf("empty path")
	}
	if root == nil {
		return "", fmt.Errorf("%s parent is not set", formatPath(key[:len(key)-1]))
	}
	planned, err := planSet(source, root, key, raw, mode)
	if err != nil {
		return "", err
	}
	return applyEdit(source, planned), nil
}

func setArrayValue(source, path string, values []string) (string, error) {
	root, err := parseYAML(source)
	if err != nil {
		return "", err
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	if len(key) == 0 {
		return "", fmt.Errorf("empty path")
	}
	if root == nil {
		return "", fmt.Errorf("%s parent is not set", formatPath(key[:len(key)-1]))
	}
	planned, err := planSetArray(source, root, key, values)
	if err != nil {
		return "", err
	}
	return applyEdit(source, planned), nil
}

func setValueIn(source, collectionPath, rawSelector, path, raw string, mode setMode) (string, error) {
	return setIn(source, collectionPath, rawSelector, path, false, func(indent int, existing *yaml.Node) (string, string, error) {
		value, logical, err := formatSetValue(raw, mode, indent, existing)
		return value, logical, err
	})
}

func setArrayValueIn(source, collectionPath, rawSelector, path string, values []string) (string, error) {
	return setIn(source, collectionPath, rawSelector, path, true, func(indent int, existing *yaml.Node) (string, string, error) {
		value, err := formatArray(values, indent, existing)
		if err != nil {
			return "", "", err
		}
		return value, "", nil
	})
}

func setIn(source, collectionPath, rawSelector, path string, allowContainer bool, format func(int, *yaml.Node) (string, string, error)) (string, error) {
	root, err := parseYAML(source)
	if err != nil {
		return "", err
	}
	collection, err := parsePath(collectionPath)
	if err != nil {
		return "", fmt.Errorf("--in: %w", err)
	}
	key, err := parsePath(path)
	if err != nil {
		return "", err
	}
	if len(key) == 0 {
		return "", fmt.Errorf("empty path")
	}
	selectors, err := parseSelectors([]string{rawSelector})
	if err != nil {
		return "", err
	}
	if root == nil {
		if len(collection) == 0 {
			return "", fmt.Errorf("root collection is not set")
		}
		return "", fmt.Errorf("%s parent is not set", formatPath(collection[:len(collection)-1]))
	}

	collectionNode := root
	if len(collection) > 0 {
		var collectionErr error
		collectionNode, collectionErr = resolvePath(root, collection, nil)
		if collectionErr != nil {
			return insertMissingCollectionRecord(source, root, collection, selectors[0], key, format)
		}
	}
	if len(collection) > 0 && collectionNode == nil {
		return insertMissingCollectionRecord(source, root, collection, selectors[0], key, format)
	}
	collectionNode, err = resolveAlias(collectionNode, nil)
	if err != nil {
		return "", err
	}
	if collectionNode.Kind != yaml.SequenceNode {
		return "", fmt.Errorf("%s is not a sequence of records", formatPath(collection))
	}
	if isRecordSequence(collectionNode) {
		index, err := selectedRecordIndex(collectionNode, collection, selectors)
		if err != nil {
			if strings.Contains(err.Error(), " has no records matching ") {
				return appendRecord(source, collectionNode, collection, selectors[0], key, format)
			}
			return "", err
		}
		record, err := resolveAlias(collectionNode.Content[index], nil)
		if err != nil {
			return "", err
		}
		return setRelative(source, root, record, append(appendPath(collection, strconv.Itoa(index)), key...), key, allowContainer, format)
	}
	return appendRecord(source, collectionNode, collection, selectors[0], key, format)
}

func setRelative(source string, root, base *yaml.Node, fullPath, relative []string, allowContainer bool, format func(int, *yaml.Node) (string, string, error)) (string, error) {
	if found, ok := findTarget(base, relative); ok {
		indent := max(found.keyNode.Column+1, 2)
		value, logical, err := format(indent, found.valueNode)
		if err != nil {
			return "", err
		}
		planned, err := replaceTarget(source, found, value, logical, allowContainer)
		if err != nil {
			return "", err
		}
		return applyEdit(source, planned), nil
	}
	planned, err := planInsertMissing(source, root, base, fullPath, relative, format)
	if err != nil {
		return "", err
	}
	return applyEdit(source, planned), nil
}

func planSet(source string, root *yaml.Node, key []string, raw string, mode setMode) (edit, error) {
	if found, ok := findTarget(root, key); ok {
		indent := replacementIndent(found)
		value, logical, err := formatSetValue(raw, mode, indent, found.valueNode)
		if err != nil {
			return edit{}, err
		}
		return replaceTarget(source, found, value, logical, false)
	}
	return planInsertMissing(source, root, root, key, key, func(indent int, existing *yaml.Node) (string, string, error) {
		value, logical, err := formatSetValue(raw, mode, indent, existing)
		return value, logical, err
	})
}

func replacementIndent(found target) int {
	if found.keyNode != nil {
		return max(found.keyNode.Column+1, 2)
	}
	if found.valueNode != nil {
		return max(found.valueNode.Column-1, 0)
	}
	return 2
}

func planSetFormatted(source string, root *yaml.Node, key []string, value, logical string) (edit, error) {
	if found, ok := findTarget(root, key); ok {
		return replaceTarget(source, found, value, logical, true)
	}
	return planInsertMissing(source, root, root, key, key, func(int, *yaml.Node) (string, string, error) {
		return value, logical, nil
	})
}

func planSetArray(source string, root *yaml.Node, key []string, values []string) (edit, error) {
	if found, ok := findTarget(root, key); ok {
		indent := replacementIndent(found)
		value, err := formatArray(values, indent, found.valueNode)
		if err != nil {
			return edit{}, err
		}
		return replaceTarget(source, found, value, "", true)
	}
	return planInsertMissing(source, root, root, key, key, func(indent int, existing *yaml.Node) (string, string, error) {
		value, err := formatArray(values, indent, existing)
		if err != nil {
			return "", "", err
		}
		return value, "", nil
	})
}

func replaceTarget(source string, found target, value, logical string, allowContainer bool) (edit, error) {
	if found.valueNode.Kind == yaml.AliasNode {
		got, err := renderGetValue(found.valueNode, found.path)
		if err != nil {
			return edit{}, err
		}
		if logical != "" && got == logical {
			return edit{start: 0, end: 0, text: ""}, nil
		}
		return edit{}, fmt.Errorf("%s is an alias; refusing to change shared YAML state", formatPath(found.path))
	}
	if found.valueNode.Kind == yaml.MappingNode || (found.valueNode.Kind == yaml.SequenceNode && !allowContainer) {
		return edit{}, fmt.Errorf("%s is a container, not a scalar value", formatPath(found.path))
	}
	if found.parent != nil && found.parent.Kind == yaml.SequenceNode {
		start, end, err := valueRange(source, found)
		if err != nil {
			return edit{}, err
		}
		return edit{start: start, end: end, text: value}, nil
	}
	if found.valueNode.Kind == yaml.SequenceNode {
		start, end, err := sequenceValueRange(source, found)
		if err != nil {
			return edit{}, err
		}
		if strings.HasPrefix(value, "\n") && found.valueNode.Line == found.keyNode.Line {
			start = trimLeftInlineValueSpace(source, start)
		}
		text := value
		if found.valueNode.Line != found.keyNode.Line {
			if !strings.HasPrefix(value, "\n") {
				text = " " + value
			}
			if end > start && source[end-1] == '\n' {
				text += "\n"
			}
		}
		return edit{start: start, end: end, text: text}, nil
	}
	start, end, err := valueRange(source, found)
	if err != nil {
		return edit{}, err
	}
	if strings.HasPrefix(value, "\n") {
		start = trimLeftInlineValueSpace(source, start)
	}
	if (found.valueNode.Style == yaml.LiteralStyle || found.valueNode.Style == yaml.FoldedStyle) && end > start && source[end-1] == '\n' {
		value += "\n"
	}
	return edit{start: start, end: end, text: value}, nil
}

func trimLeftInlineValueSpace(source string, start int) int {
	for start > 0 && (source[start-1] == ' ' || source[start-1] == '\t') {
		start--
	}
	return start
}

func sequenceValueRange(source string, found target) (int, int, error) {
	if found.valueNode.Line == found.keyNode.Line {
		return valueRange(source, found)
	}
	keyStart, keyEnd, err := lineRange(source, found.keyNode.Line)
	if err != nil {
		return 0, 0, err
	}
	keyLine := source[keyStart:keyEnd]
	colon := strings.IndexByte(keyLine, ':')
	if colon < 0 {
		return 0, 0, fmt.Errorf("%s key line is not editable", formatPath(found.path))
	}
	_, end, err := nodeRange(source, found.valueNode, max(found.keyNode.Column-1, 0))
	if err != nil {
		return 0, 0, err
	}
	return keyStart + colon + 1, end, nil
}

func planInsertMissing(source string, root, base *yaml.Node, fullPath, relative []string, format func(int, *yaml.Node) (string, string, error)) (edit, error) {
	if len(relative) == 0 {
		return edit{}, fmt.Errorf("%s is not set", formatPath(fullPath))
	}

	current, missing, err := insertionBase(base, fullPath, relative)
	if err != nil {
		return edit{}, err
	}

	keyIndent := childIndent(current)
	value, _, err := format(keyIndent+2*len(missing), nil)
	if err != nil {
		return edit{}, err
	}
	if current.Style == yaml.FlowStyle {
		return insertIntoFlowMapping(source, current, missing, value, fullPath)
	}
	text := nestedMappingText(missing, value, keyIndent)
	insertAt, err := mappingInsertAt(source, current)
	if err != nil {
		return edit{}, err
	}
	return edit{start: insertAt, end: insertAt, text: prefixLineEndingAt(source, insertAt) + text}, nil
}

func insertIntoFlowMapping(source string, node *yaml.Node, missing []string, value string, fullPath []string) (edit, error) {
	if len(missing) != 1 {
		return edit{}, fmt.Errorf("%s parent is a flow mapping; refusing to expand nested YAML", formatPath(fullPath[:len(fullPath)-len(missing)+1]))
	}
	if strings.Contains(value, "\n") {
		return edit{}, fmt.Errorf("%s parent is a flow mapping; refusing to insert multiline YAML", formatPath(fullPath[:len(fullPath)-len(missing)+1]))
	}
	lineStart, lineEnd, err := lineRange(source, node.Line)
	if err != nil {
		return edit{}, err
	}
	line := source[lineStart:lineEnd]
	start := max(node.Column-1, 0)
	closeAt := matchingFlowClose(line, start, '{', '}')
	if closeAt < 0 {
		return edit{}, fmt.Errorf("%s flow mapping is not editable", formatPath(fullPath[:len(fullPath)-len(missing)]))
	}
	insertAt := closeAt
	for insertAt > start && (line[insertAt-1] == ' ' || line[insertAt-1] == '\t') {
		insertAt--
	}
	suffix := line[insertAt:closeAt]
	prefix := ""
	if len(node.Content) > 0 {
		prefix = ", "
	}
	return edit{start: lineStart + insertAt, end: lineStart + closeAt, text: prefix + missing[0] + ": " + value + suffix}, nil
}

func insertionBase(base *yaml.Node, fullPath, relative []string) (*yaml.Node, []string, error) {
	current, err := resolveAlias(base, nil)
	if err != nil {
		return nil, nil, err
	}
	basePathLen := len(fullPath) - len(relative)
	existing := 0
	for existing < len(relative)-1 {
		part := relative[existing]
		switch current.Kind {
		case yaml.MappingNode:
			next, ok := mappingChild(current, part)
			if !ok {
				return current, relative[existing:], nil
			}
			next, err := resolveAlias(next, nil)
			if err != nil {
				return nil, nil, err
			}
			if next.Kind != yaml.MappingNode && next.Kind != yaml.SequenceNode {
				parentPath := fullPath[:basePathLen+existing+1]
				return nil, nil, fmt.Errorf("%s parent is not a mapping", formatPath(parentPath))
			}
			current = next
			existing++
		case yaml.SequenceNode:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(current.Content) {
				parentPath := fullPath[:basePathLen+existing+1]
				return nil, nil, fmt.Errorf("%s is not set", formatPath(parentPath))
			}
			next, err := resolveAlias(current.Content[index], nil)
			if err != nil {
				return nil, nil, err
			}
			if next.Kind != yaml.MappingNode && next.Kind != yaml.SequenceNode {
				parentPath := fullPath[:basePathLen+existing+1]
				return nil, nil, fmt.Errorf("%s parent is not a mapping", formatPath(parentPath))
			}
			current = next
			existing++
		default:
			parentPath := fullPath[:basePathLen+existing]
			return nil, nil, fmt.Errorf("%s parent is not a mapping", formatPath(parentPath))
		}
	}

	if current.Kind != yaml.MappingNode {
		parentPath := fullPath[:len(fullPath)-len(relative)+existing]
		return nil, nil, fmt.Errorf("%s parent is not a mapping", formatPath(parentPath))
	}
	return current, relative[existing:], nil
}

func insertMissingCollectionRecord(source string, root *yaml.Node, collection []string, selector selector, key []string, format func(int, *yaml.Node) (string, string, error)) (string, error) {
	planned, err := planInsertMissing(source, root, root, collection, collection, func(indent int, existing *yaml.Node) (string, string, error) {
		record, err := recordText(selector, key, format, indent)
		if err != nil {
			return "", "", err
		}
		return "\n" + record, "", nil
	})
	if err != nil {
		return "", err
	}
	return applyEdit(source, planned), nil
}

func appendRecord(source string, collectionNode *yaml.Node, collection []string, selector selector, key []string, format func(int, *yaml.Node) (string, string, error)) (string, error) {
	insertAt, err := sequenceInsertAt(source, collectionNode)
	if err != nil {
		return "", err
	}
	itemIndent := max(collectionNode.Column-1, 0)
	text, err := recordText(selector, key, format, itemIndent)
	if err != nil {
		return "", err
	}
	planned := edit{start: insertAt, end: insertAt, text: prefixLineEndingAt(source, insertAt) + text}
	return applyEdit(source, planned), nil
}

func recordText(selector selector, key []string, format func(int, *yaml.Node) (string, string, error), itemIndent int) (string, error) {
	if len(selector.path) != 1 {
		return "", fmt.Errorf("--on nested selectors are not supported for creating YAML records")
	}
	if len(key) == 0 {
		return "", fmt.Errorf("empty path")
	}
	value, _, err := format(itemIndent+2, nil)
	if err != nil {
		return "", err
	}
	var lines []string
	lines = append(lines, strings.Repeat(" ", itemIndent)+"- "+selector.path[0]+": "+formatPlainString(selector.compare))
	if !samePath(selector.path, key) {
		lines = append(lines, nestedMappingLines(key, value, itemIndent+2)...)
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func findTarget(root *yaml.Node, path []string) (target, bool) {
	if len(path) == 0 {
		return target{}, false
	}
	if found, ok := findSequenceTarget(root, path); ok {
		return found, true
	}
	parent, ok := resolveParentForSet(root, path[:len(path)-1])
	if !ok || parent.Kind != yaml.MappingNode {
		return target{}, false
	}
	key := path[len(path)-1]
	for i := 0; i+1 < len(parent.Content); i += 2 {
		keyNode := parent.Content[i]
		if keyNode.Kind == yaml.ScalarNode && keyNode.Value == key {
			return target{keyNode: keyNode, valueNode: parent.Content[i+1], parent: parent, key: key, path: path}, true
		}
	}
	return target{}, false
}

func findSequenceTarget(root *yaml.Node, path []string) (target, bool) {
	parent, ok := resolveParentForSet(root, path[:len(path)-1])
	if !ok || parent.Kind != yaml.SequenceNode {
		return target{}, false
	}
	index, err := strconv.Atoi(path[len(path)-1])
	if err != nil || index < 0 || index >= len(parent.Content) {
		return target{}, false
	}
	return target{valueNode: parent.Content[index], parent: parent, path: path}, true
}

func resolveParentForSet(root *yaml.Node, path []string) (*yaml.Node, bool) {
	node := root
	for _, part := range path {
		node, _ = resolveAlias(node, nil)
		switch node.Kind {
		case yaml.MappingNode:
			next, ok := mappingChild(node, part)
			if !ok {
				return nil, false
			}
			node = next
		case yaml.SequenceNode:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(node.Content) {
				return nil, false
			}
			node = node.Content[index]
		default:
			return nil, false
		}
	}
	node, _ = resolveAlias(node, nil)
	return node, true
}

func mappingChild(node *yaml.Node, key string) (*yaml.Node, bool) {
	if err := rejectDuplicateKeys(node); err != nil {
		return nil, false
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Kind == yaml.ScalarNode && node.Content[i].Value == key {
			return node.Content[i+1], true
		}
	}
	return nil, false
}

func valueRange(source string, found target) (int, int, error) {
	if found.valueNode.Style == yaml.LiteralStyle || found.valueNode.Style == yaml.FoldedStyle {
		return blockValueRange(source, found)
	}
	start, end, err := lineRange(source, found.valueNode.Line)
	if err != nil {
		return 0, 0, err
	}
	line := source[start:end]
	col := max(found.valueNode.Column-1, 0)
	if col > len(line) {
		col = len(line)
	}
	if found.valueNode.Anchor != "" {
		col = anchoredValueColumn(line, col, found.valueNode.Anchor)
	}
	valueStart := start + col
	flow := found.parent != nil && found.parent.Style == yaml.FlowStyle
	valueEnd := start + valueEndInLine(line, col, flow)
	return valueStart, valueEnd, nil
}

func blockValueRange(source string, found target) (int, int, error) {
	start, _, err := lineRange(source, found.valueNode.Line)
	if err != nil {
		return 0, 0, err
	}
	lineEnd := strings.IndexByte(source[start:], '\n')
	if lineEnd < 0 {
		lineEnd = len(source) - start
	}
	line := source[start : start+lineEnd]
	col := max(found.valueNode.Column-1, 0)
	valueStart := start + min(col, len(line))
	keyIndent := max(found.keyNode.Column-1, 0)
	end := start + lineEnd
	for end < len(source) {
		if source[end] == '\n' {
			end++
		}
		nextEnd := strings.IndexByte(source[end:], '\n')
		lineStop := len(source)
		if nextEnd >= 0 {
			lineStop = end + nextEnd
		}
		nextLine := source[end:lineStop]
		if strings.TrimSpace(nextLine) != "" && leadingSpaces(nextLine) <= keyIndent {
			break
		}
		end = lineStop
		if nextEnd < 0 {
			break
		}
	}
	return valueStart, end, nil
}

func lineRange(source string, lineNo int) (int, int, error) {
	if lineNo <= 0 {
		return 0, 0, fmt.Errorf("invalid line %d", lineNo)
	}
	line := 1
	start := 0
	for i, ch := range source {
		if line == lineNo {
			start = i
			break
		}
		if ch == '\n' {
			line++
		}
	}
	if line != lineNo {
		return 0, 0, fmt.Errorf("line %d not found", lineNo)
	}
	end := strings.IndexByte(source[start:], '\n')
	if end < 0 {
		return start, len(source), nil
	}
	return start, start + end + 1, nil
}

func valueEndInLine(line string, start int, flow bool) int {
	inQuote := byte(0)
	escaped := false
	for i := start; i < len(line); i++ {
		ch := line[i]
		if inQuote != 0 {
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == inQuote {
				inQuote = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			inQuote = ch
			continue
		}
		if ch == '#' || ch == '\n' {
			return trimRightSpaceIndex(line, i)
		}
		if flow && (ch == ',' || ch == '}' || ch == ']') {
			return trimRightSpaceIndex(line, i)
		}
	}
	return trimRightSpaceIndex(line, len(line))
}

func matchingFlowClose(line string, start int, open, close byte) int {
	inQuote := byte(0)
	escaped := false
	depth := 0
	for i := start; i < len(line); i++ {
		ch := line[i]
		if inQuote != 0 {
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == inQuote {
				inQuote = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			inQuote = ch
			continue
		}
		switch ch {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func anchoredValueColumn(line string, col int, anchor string) int {
	marker := "&" + anchor
	anchorAt := strings.Index(line[col:], marker)
	if anchorAt < 0 {
		return col
	}
	col += anchorAt + len(marker)
	for col < len(line) && (line[col] == ' ' || line[col] == '\t') {
		col++
	}
	return col
}

func formatSetValue(raw string, mode setMode, indent int, existing *yaml.Node) (string, string, error) {
	if strings.Contains(raw, "\n") {
		return literalBlock(raw, indent), raw, nil
	}
	if mode == setModeString {
		return formatStringValue(raw, existing), raw, nil
	}
	if inferredScalar(raw) {
		return raw, raw, nil
	}
	if strings.HasPrefix(raw, "!") {
		if err := validateTaggedScalar(raw); err != nil {
			return "", "", err
		}
		return raw, raw, nil
	}
	if existing != nil {
		switch existing.Style {
		case yaml.SingleQuotedStyle, yaml.DoubleQuotedStyle:
			return formatStringValue(raw, existing), raw, nil
		case yaml.LiteralStyle, yaml.FoldedStyle:
			return literalBlock(raw, indent), raw, nil
		}
	}
	value := formatStringFallback(raw)
	return value, raw, nil
}

func validateTaggedScalar(raw string) error {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("value: "+raw+"\n"), &doc); err != nil {
		return fmt.Errorf("invalid YAML tag value; use --string for literal text starting with !: %w", err)
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode || len(doc.Content[0].Content) < 2 {
		return fmt.Errorf("invalid YAML tag value; use --string for literal text starting with !")
	}
	value := doc.Content[0].Content[1]
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("YAML tag value must be a scalar; use --string for literal text starting with !")
	}
	if !strings.HasPrefix(value.Tag, "!") {
		return fmt.Errorf("invalid YAML tag value; use --string for literal text starting with !")
	}
	return nil
}

func formatInferredValue(raw string) string {
	if inferredScalar(raw) {
		return raw
	}
	return formatStringFallback(raw)
}

func inferredScalar(raw string) bool {
	switch {
	case raw == "true" || raw == "false" || raw == "null":
		return true
	case integer.MatchString(raw) || float.MatchString(raw):
		return true
	case timestampScalar(raw):
		return true
	}
	return false
}

func timestampScalar(raw string) bool {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("value: "+raw+"\n"), &doc); err != nil {
		return false
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode || len(doc.Content[0].Content) < 2 {
		return false
	}
	value := doc.Content[0].Content[1]
	return value.Kind == yaml.ScalarNode && value.Tag == "!!timestamp" && value.Value == raw
}

func formatStringFallback(raw string) string {
	if plainString(raw) {
		return raw
	}
	return strconv.Quote(raw)
}

func formatStringValue(raw string, existing *yaml.Node) string {
	if existing != nil {
		switch existing.Style {
		case yaml.SingleQuotedStyle:
			return "'" + strings.ReplaceAll(raw, "'", "''") + "'"
		case yaml.DoubleQuotedStyle:
			return strconv.Quote(raw)
		}
	}
	return strconv.Quote(raw)
}

func formatPlainString(raw string) string {
	if plainString(raw) {
		return raw
	}
	return strconv.Quote(raw)
}

func plainString(raw string) bool {
	if strings.TrimSpace(raw) != raw || raw == "" || strings.Contains(raw, "\n") {
		return false
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("value: "+raw+"\n"), &doc); err != nil {
		return false
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode || len(doc.Content[0].Content) < 2 {
		return false
	}
	value := doc.Content[0].Content[1]
	return value.Kind == yaml.ScalarNode && value.Tag == "!!str" && value.Value == raw
}

func formatArray(values []string, indent int, existing *yaml.Node) (string, error) {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, formatInferredValue(value))
	}
	if blockArrayStyle(values, existing) {
		prefix := strings.Repeat(" ", indent)
		lines := make([]string, 0, len(parts))
		for _, part := range parts {
			lines = append(lines, prefix+"- "+part)
		}
		return "\n" + strings.Join(lines, "\n"), nil
	}
	return "[" + strings.Join(parts, ", ") + "]", nil
}

func blockArrayStyle(values []string, existing *yaml.Node) bool {
	if existing == nil {
		return true
	}
	if existing.Kind == yaml.SequenceNode && existing.Style != yaml.FlowStyle {
		return true
	}
	return len(values) >= 5
}

func literalBlock(raw string, indent int) string {
	lines := strings.Split(raw, "\n")
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return strconv.Quote("")
	}
	prefix := strings.Repeat(" ", indent)
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return "|-\n" + strings.Join(lines, "\n")
}

func mappingInsertAt(source string, node *yaml.Node) (int, error) {
	if len(node.Content) == 0 {
		start, _, err := lineRange(source, node.Line)
		return start, err
	}
	lastValue := node.Content[len(node.Content)-1]
	if lastValue.Kind == yaml.MappingNode || lastValue.Kind == yaml.SequenceNode || lastValue.Style == yaml.LiteralStyle || lastValue.Style == yaml.FoldedStyle {
		_, end, err := nodeRange(source, lastValue, childIndent(node)-2)
		return end, err
	}
	_, end, err := lineRange(source, lastValue.Line)
	return end, err
}

func sequenceInsertAt(source string, node *yaml.Node) (int, error) {
	if len(node.Content) == 0 {
		start, _, err := lineRange(source, node.Line)
		return start, err
	}
	last := node.Content[len(node.Content)-1]
	_, end, err := nodeRange(source, last, max(node.Column-1, 0))
	return end, err
}

func nodeRange(source string, node *yaml.Node, parentIndent int) (int, int, error) {
	start, end, err := lineRange(source, node.Line)
	if err != nil {
		return 0, 0, err
	}
	for end < len(source) {
		lineEnd := strings.IndexByte(source[end:], '\n')
		stop := len(source)
		if lineEnd >= 0 {
			stop = end + lineEnd
		}
		line := source[end:stop]
		if strings.TrimSpace(line) != "" && leadingSpaces(line) <= parentIndent {
			break
		}
		end = stop
		if lineEnd < 0 {
			break
		}
		end++
	}
	return start, end, nil
}

func childIndent(node *yaml.Node) int {
	if len(node.Content) > 0 {
		return max(node.Content[0].Column-1, 0)
	}
	return max(node.Column+1, 2)
}

func nestedMappingText(path []string, value string, indent int) string {
	text := strings.Join(nestedMappingLines(path, value, indent), "\n")
	if strings.HasSuffix(text, "\n") {
		return text
	}
	return text + "\n"
}

func nestedMappingLines(path []string, value string, indent int) []string {
	if len(path) == 1 {
		if strings.HasPrefix(value, "\n") {
			return []string{strings.Repeat(" ", indent) + path[0] + ":" + value}
		}
		return []string{strings.Repeat(" ", indent) + path[0] + ": " + value}
	}
	lines := []string{strings.Repeat(" ", indent) + path[0] + ":"}
	lines = append(lines, nestedMappingLines(path[1:], value, indent+2)...)
	return lines
}

func prefixLineEndingAt(source string, pos int) string {
	if pos == 0 || source[pos-1] == '\n' {
		return ""
	}
	return "\n"
}

func applyEdit(source string, e edit) string {
	if e.start == 0 && e.end == 0 && e.text == "" {
		return source
	}
	return source[:e.start] + e.text + source[e.end:]
}

func trimRightSpaceIndex(s string, end int) int {
	for end > 0 && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return end
}

func leadingSpaces(line string) int {
	count := 0
	for count < len(line) && line[count] == ' ' {
		count++
	}
	return count
}

func samePath(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var (
	integer = regexp.MustCompile(`^[+-]?[0-9]+$`)
	float   = regexp.MustCompile(`^[+-]?([0-9]+\.[0-9]*|[0-9]*\.[0-9]+)([eE][+-]?[0-9]+)?$|^[+-]?[0-9]+[eE][+-]?[0-9]+$`)
)
