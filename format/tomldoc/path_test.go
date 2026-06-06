package tomldoc

import (
	"fmt"
	"testing"
)

func TestSetRootKeyInsertsBeforeFirstSection(t *testing.T) {
	source := "[server]\nport = 3000\n"

	got, err := Set(source, "intro", "hello")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "intro = \"hello\"\n\n[server]\nport = 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetRootKeyInsertsAfterExistingRootKeys(t *testing.T) {
	source := "title = \"demo\"\n\n[server]\nport = 3000\n"

	got, err := Set(source, "intro", "hello")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "title = \"demo\"\nintro = \"hello\"\n\n[server]\nport = 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetMissingParentCreatesTable(t *testing.T) {
	source := "[client]\nport = 3000\n"

	got, err := Set(source, "server.host", "localhost")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "[client]\nport = 3000\n\n[server]\nhost = \"localhost\"\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetMissingParentFollowsSiblingTableStyle(t *testing.T) {
	source := "[env.prod]\nport = 80\n\n[env.dev]\nport = 3000\n"

	got, err := Set(source, "env.debug.port", "8080")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "[env.prod]\nport = 80\n\n[env.dev]\nport = 3000\n\n[env.debug]\nport = 8080\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetMissingParentFollowsFirstLevelParentTableStyle(t *testing.T) {
	source := "[env]\nprod.port = 80\ndev.port = 3000\n"

	got, err := Set(source, "env.debug.port", "8080")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "[env]\nprod.port = 80\ndev.port = 3000\ndebug.port = 8080\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetMissingParentKeepsExistingDottedSiblingStyle(t *testing.T) {
	source := "server.port = 3000\n\n[client]\nport = 3000\n"

	got, err := Set(source, "server.host", "localhost")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "server.port = 3000\nserver.host = \"localhost\"\n\n[client]\nport = 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetMissingValueInExistingEmptyTable(t *testing.T) {
	source := "title = \"demo\"\n\n[tui.keymap]\n\n[tui]\ntheme = \"light\"\n"

	got, err := Set(source, "tui.keymap.composer.submit", "tab")
	if err == nil {
		got, err = Set(got, "tui.keymap.composer.queue", "alt-q")
	}
	if err == nil {
		got, err = Set(got, "tui.keymap.editor.insert_newline", "enter")
	}

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "title = \"demo\"\n\n[tui.keymap]\ncomposer.submit = \"tab\"\ncomposer.queue = \"alt-q\"\neditor.insert_newline = \"enter\"\n\n[tui]\ntheme = \"light\"\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestDoubleDotEscapesLiteralDotInPath(t *testing.T) {
	source := "[server]\n\"public.port\" = 3000\n"

	got, err := Get(source, "server.public..port")

	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got != "3000" {
		t.Fatalf("Get = %q, want 3000", got)
	}
}

func TestSetDoubleDotPathPatchesLiteralDotKey(t *testing.T) {
	source := "[server]\n\"public.port\" = 3000\n"

	got, err := Set(source, "server.public..port", "4000")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "[server]\n\"public.port\" = 4000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetDoubleDotPathInExistingTableQuotesLiteralDotKey(t *testing.T) {
	source := "[server]\nhost = \"localhost\"\n"

	got, err := Set(source, "server.public..port", "3000")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "[server]\nhost = \"localhost\"\n\"public.port\" = 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestListDoubleDotPath(t *testing.T) {
	source := "[server]\n\"public.port\" = 3000\n"

	got, err := List(source, "server.public..port")

	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	want := []Entry{{Key: "server.public..port", Value: "3000"}}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}
