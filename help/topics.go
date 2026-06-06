package help

import (
	"embed"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

//go:embed topics/*.txt
var topicFS embed.FS

func TopicNames() []string {
	entries, err := topicFS.ReadDir("topics")
	if err != nil {
		return nil
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

func Lookup(topic string) (string, bool) {
	content, err := topicFS.ReadFile(path.Join("topics", topic+".txt"))
	if err != nil {
		return "", false
	}

	return Render(string(content)), true
}

type styles struct {
	bold   lipgloss.Style
	inline lipgloss.Style
	block  lipgloss.Style
}

func Render(content string) string {
	styles := newStyles()
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	rendered := make([]string, 0, len(lines))
	inBlock := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "```" {
			inBlock = !inBlock
			continue
		}

		if inBlock {
			rendered = append(rendered, styles.block.Render(line))
			continue
		}

		rendered = append(rendered, renderInline(line, styles))
	}

	return strings.Join(rendered, "\n")
}

func newStyles() styles {
	renderer := lipgloss.NewRenderer(io.Discard)
	renderer.SetColorProfile(termenv.ANSI)

	if _, noColor := os.LookupEnv("NO_COLOR"); noColor {
		return styles{}
	}

	styles := styles{
		bold: lipgloss.NewStyle().Renderer(renderer).Bold(true),
	}

	styles.inline = lipgloss.NewStyle().Renderer(renderer).Foreground(lipgloss.Color("2"))
	styles.block = lipgloss.NewStyle().Renderer(renderer).Foreground(lipgloss.Color("4"))
	return styles
}

func renderInline(line string, styles styles) string {
	var builder strings.Builder
	for i := 0; i < len(line); {
		switch {
		case strings.HasPrefix(line[i:], "```"):
			builder.WriteString("```")
			i += 3

		case strings.HasPrefix(line[i:], "**"):
			end := strings.Index(line[i+2:], "**")
			if end == -1 {
				builder.WriteString(line[i:])
				i = len(line)
				continue
			}

			text := line[i+2 : i+2+end]
			builder.WriteString(styles.bold.Render(text))
			i += 2 + end + 2

		case line[i] == '`':
			end := strings.IndexByte(line[i+1:], '`')
			if end == -1 {
				builder.WriteString(line[i:])
				i = len(line)
				continue
			}

			text := line[i+1 : i+1+end]
			builder.WriteString(styles.inline.Render(text))
			i += 1 + end + 1

		default:
			builder.WriteByte(line[i])
			i++
		}
	}

	return builder.String()
}
