package cmd

import (
	"testing"

	"github.com/dannyben/config/format"
)

func TestRenderListUsesUnifiedKeyValuePresentation(t *testing.T) {
	entries := []format.Entry{
		{Key: "name", Value: "rush"},
		{Key: "server.port", Value: "3000"},
	}

	got := renderList(entries)

	want := "name=rush\nserver.port=3000\n"
	if got != want {
		t.Fatalf("renderList mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderListEntryOwnsMultilinePresentation(t *testing.T) {
	entry := format.Entry{Key: "help", Value: "line one\nline two"}

	got := renderListEntry(entry)

	want := "help=line one line two"
	if got != want {
		t.Fatalf("renderListEntry = %q, want %q", got, want)
	}
}

func TestRenderListEntryTruncatesLongValue(t *testing.T) {
	entry := format.Entry{
		Key:   "commands.2.help",
		Value: "Clone a GitHub package repository. This command clones the repository and registers it in the configuration file.",
	}

	got := renderListEntry(entry)

	want := "commands.2.help=Clone a GitHub package repository. This command clones the rep…"
	if got != want {
		t.Fatalf("renderListEntry = %q, want %q", got, want)
	}
	if runeLen(got) > listMaxLineLength {
		t.Fatalf("rendered line length = %d, want <= %d", runeLen(got), listMaxLineLength)
	}
}

func TestRenderListEntryPreservesKeyBeforeValue(t *testing.T) {
	entry := format.Entry{
		Key:   "commands.10.args.1.deeply.nested.configuration.path.target_package",
		Value: "Target package name. This can either be the package name without the repository name.",
	}

	got := renderListEntry(entry)

	want := "commands.10.args.1.deeply.nested.configuration.path.target_…=Target package na…"
	if got != want {
		t.Fatalf("renderListEntry = %q, want %q", got, want)
	}
	if runeLen(got) > listMaxLineLength {
		t.Fatalf("rendered line length = %d, want <= %d", runeLen(got), listMaxLineLength)
	}
}
