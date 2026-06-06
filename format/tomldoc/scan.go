package tomldoc

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/creachadair/tomledit/parser"
	"github.com/creachadair/tomledit/scanner"
)

func scan(source string) ([]assignment, []section, error) {
	var assignments []assignment
	var sections []section
	var current parser.Key
	arrayCounts := map[string]int{}

	for _, stmt := range statements(source) {
		if len(stmt.tokens) == 0 {
			continue
		}
		if key, ok := arrayTableKey(stmt.tokens); ok {
			index := arrayCounts[key.String()]
			arrayCounts[key.String()] = index + 1
			current = arrayRecordKey(key, index)
			sections = append(sections, section{key: current, start: stmt.start, insertAt: stmt.end})
			continue
		}
		if key, ok := tableKey(stmt.tokens); ok {
			current = key
			sections = append(sections, section{key: key, start: stmt.start, insertAt: stmt.end})
			continue
		}
		key, valueSpan, valueTokens, ok := assignmentKey(stmt.tokens)
		if !ok {
			continue
		}
		fullKey := append(parser.Key{}, current...)
		fullKey = append(fullKey, key...)
		if isInlineTableValue(valueTokens) {
			assignments = append(assignments, inlineAssignments(fullKey, valueTokens)...)
			sections = append(sections, inlineSections(fullKey, valueTokens)...)
			sections = append(sections, inlineSection(fullKey, valueTokens))
			updateCurrentSectionInsertAt(sections, current, stmt.end)
			continue
		}
		assignments = append(assignments, arrayItemAssignments(fullKey, valueTokens)...)
		assignments = append(assignments, assignment{
			key:       fullKey,
			scope:     append(parser.Key{}, current...),
			lineSpan:  scanner.Span{Pos: stmt.start, End: stmt.end},
			valueSpan: valueSpan,
		})
		updateCurrentSectionInsertAt(sections, current, stmt.end)
	}

	return assignments, sections, nil
}

func arrayItemAssignments(parent parser.Key, tokens []token) []assignment {
	if len(tokens) < 2 || tokens[0].kind != scanner.LBracket || tokens[len(tokens)-1].kind != scanner.RBracket {
		return nil
	}
	segments := arraySegments(tokens)
	out := make([]assignment, 0, len(segments))
	for i, segment := range segments {
		if len(segment) == 0 {
			continue
		}
		key := append(parser.Key{}, parent...)
		key = append(key, strconv.Itoa(i))
		out = append(out, assignment{
			key:       key,
			lineSpan:  scanner.Span{Pos: tokens[0].span.Pos, End: tokens[len(tokens)-1].span.End},
			valueSpan: scanner.Span{Pos: segment[0].span.Pos, End: segment[len(segment)-1].span.End},
			internal:  true,
		})
	}
	return out
}

func updateCurrentSectionInsertAt(sections []section, current parser.Key, insertAt int) {
	if len(current) == 0 {
		return
	}
	for i := len(sections) - 1; i >= 0; i-- {
		if sections[i].key.Equals(current) && !sections[i].inline {
			sections[i].insertAt = insertAt
			return
		}
	}
}

type statement struct {
	start  int
	end    int
	tokens []token
}

func statements(source string) []statement {
	var out []statement
	var current statement
	s := scanner.New(strings.NewReader(source))

	flush := func(end int) {
		if len(current.tokens) > 0 {
			current.end = end
			out = append(out, current)
			current = statement{}
		}
	}

	for {
		err := s.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			break
		}
		kind := s.Token()
		span := s.Span()
		if kind == scanner.Newline {
			flush(span.End)
			continue
		}
		if kind == scanner.Comment {
			if len(current.tokens) > 0 {
				flush(lineEnd(source, span.Pos))
			}
			continue
		}
		if len(current.tokens) == 0 {
			current.start = span.Pos
		}
		current.tokens = append(current.tokens, token{kind: kind, text: string(s.Text()), span: span})
	}
	flush(len(source))
	return out
}

func lineEnd(source string, pos int) int {
	end := strings.IndexByte(source[pos:], '\n')
	if end < 0 {
		return len(source)
	}
	return pos + end + 1
}

func tableKey(tokens []token) (parser.Key, bool) {
	if len(tokens) < 3 || tokens[0].kind != scanner.LBracket || tokens[len(tokens)-1].kind != scanner.RBracket {
		return nil, false
	}
	key, err := parser.ParseKey(joinTokenText(tokens[1 : len(tokens)-1]))
	return key, err == nil
}

func arrayTableKey(tokens []token) (parser.Key, bool) {
	if len(tokens) < 5 ||
		tokens[0].kind != scanner.LBracket ||
		tokens[1].kind != scanner.LBracket ||
		tokens[len(tokens)-2].kind != scanner.RBracket ||
		tokens[len(tokens)-1].kind != scanner.RBracket {
		return nil, false
	}
	key, err := parser.ParseKey(joinTokenText(tokens[2 : len(tokens)-2]))
	return key, err == nil
}

func assignmentKey(tokens []token) (parser.Key, scanner.Span, []token, bool) {
	eq := -1
	for i, tok := range tokens {
		if tok.kind == scanner.Equal {
			eq = i
			break
		}
	}
	if eq <= 0 || eq == len(tokens)-1 {
		return nil, scanner.Span{}, nil, false
	}
	key, err := parser.ParseKey(joinTokenText(tokens[:eq]))
	if err != nil {
		return nil, scanner.Span{}, nil, false
	}
	valueTokens := tokens[eq+1:]
	return key, scanner.Span{Pos: valueTokens[0].span.Pos, End: valueTokens[len(valueTokens)-1].span.End}, valueTokens, true
}

func isInlineTableValue(tokens []token) bool {
	return len(tokens) >= 2 && tokens[0].kind == scanner.LInline && tokens[len(tokens)-1].kind == scanner.RInline
}

func inlineAssignments(parent parser.Key, tokens []token) []assignment {
	items := inlineItems(parent, tokens)
	var out []assignment
	for _, item := range items {
		if item.inline {
			out = append(out, inlineAssignments(item.key, item.valueTokens)...)
			continue
		}
		out = append(out, assignment{key: item.key, lineSpan: item.deleteSpan, valueSpan: item.valueSpan})
	}
	return out
}

func inlineSections(parent parser.Key, tokens []token) []section {
	items := inlineItems(parent, tokens)
	var out []section
	for _, item := range items {
		if item.inline {
			out = append(out, inlineSection(item.key, item.valueTokens))
			out = append(out, inlineSections(item.key, item.valueTokens)...)
		}
	}
	return out
}

func inlineSection(key parser.Key, tokens []token) section {
	return section{
		key:      key,
		insertAt: tokens[len(tokens)-1].span.Pos,
		inline:   true,
		empty:    len(inlineSegments(tokens)) == 0,
	}
}

type inlineItem struct {
	key         parser.Key
	valueSpan   scanner.Span
	deleteSpan  scanner.Span
	valueTokens []token
	inline      bool
}

func inlineItems(parent parser.Key, tokens []token) []inlineItem {
	segments := inlineSegments(tokens)
	out := make([]inlineItem, 0, len(segments))
	for i, segment := range segments {
		eq := topLevelEqual(segment)
		if eq <= 0 || eq == len(segment)-1 {
			continue
		}
		key, err := parser.ParseKey(joinTokenText(segment[:eq]))
		if err != nil {
			continue
		}
		valueTokens := segment[eq+1:]
		fullKey := append(parser.Key{}, parent...)
		fullKey = append(fullKey, key...)
		valueSpan := scanner.Span{Pos: valueTokens[0].span.Pos, End: valueTokens[len(valueTokens)-1].span.End}
		out = append(out, inlineItem{
			key:         fullKey,
			valueSpan:   valueSpan,
			deleteSpan:  inlineDeleteSpan(segments, i),
			valueTokens: valueTokens,
			inline:      isInlineTableValue(valueTokens),
		})
	}
	return out
}

func inlineSegments(tokens []token) [][]token {
	if !isInlineTableValue(tokens) {
		return nil
	}
	inner := tokens[1 : len(tokens)-1]
	if len(inner) == 0 {
		return nil
	}
	var segments [][]token
	start := 0
	bracketDepth := 0
	inlineDepth := 0
	for i, tok := range inner {
		switch tok.kind {
		case scanner.LBracket:
			bracketDepth++
		case scanner.RBracket:
			bracketDepth--
		case scanner.LInline:
			inlineDepth++
		case scanner.RInline:
			inlineDepth--
		case scanner.Comma:
			if bracketDepth == 0 && inlineDepth == 0 {
				segments = append(segments, inner[start:i])
				start = i + 1
			}
		}
	}
	segments = append(segments, inner[start:])
	return segments
}

func arraySegments(tokens []token) [][]token {
	if len(tokens) < 2 || tokens[0].kind != scanner.LBracket || tokens[len(tokens)-1].kind != scanner.RBracket {
		return nil
	}
	inner := tokens[1 : len(tokens)-1]
	if len(inner) == 0 {
		return nil
	}
	var segments [][]token
	start := 0
	bracketDepth := 0
	inlineDepth := 0
	for i, tok := range inner {
		switch tok.kind {
		case scanner.LBracket:
			bracketDepth++
		case scanner.RBracket:
			bracketDepth--
		case scanner.LInline:
			inlineDepth++
		case scanner.RInline:
			inlineDepth--
		case scanner.Comma:
			if bracketDepth == 0 && inlineDepth == 0 {
				segments = append(segments, inner[start:i])
				start = i + 1
			}
		}
	}
	segments = append(segments, inner[start:])
	return segments
}

func topLevelEqual(tokens []token) int {
	bracketDepth := 0
	inlineDepth := 0
	for i, tok := range tokens {
		switch tok.kind {
		case scanner.LBracket:
			bracketDepth++
		case scanner.RBracket:
			bracketDepth--
		case scanner.LInline:
			inlineDepth++
		case scanner.RInline:
			inlineDepth--
		case scanner.Equal:
			if bracketDepth == 0 && inlineDepth == 0 {
				return i
			}
		}
	}
	return -1
}

func inlineDeleteSpan(segments [][]token, index int) scanner.Span {
	segment := segments[index]
	if len(segments) == 1 {
		return scanner.Span{Pos: segment[0].span.Pos, End: inlineSegmentValueEnd(segment)}
	}
	if index == 0 {
		return scanner.Span{Pos: segment[0].span.Pos, End: segments[index+1][0].span.Pos}
	}
	return scanner.Span{Pos: inlineSegmentValueEnd(segments[index-1]), End: inlineSegmentValueEnd(segment)}
}

func inlineSegmentValueEnd(segment []token) int {
	eq := topLevelEqual(segment)
	if eq < 0 || eq == len(segment)-1 {
		return segment[len(segment)-1].span.End
	}
	return segment[len(segment)-1].span.End
}
