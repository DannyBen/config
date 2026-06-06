package tomldoc

import (
	"fmt"
	"testing"
)

func TestGetMultilineBasicString(t *testing.T) {
	source := "message = \"\"\"hello\nworld\"\"\"\n"

	got, err := Get(source, "message")

	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got != "hello\nworld" {
		t.Fatalf("Get = %q, want multiline value", got)
	}
}

func TestGetMultilineLiteralString(t *testing.T) {
	source := "message = '''hello\nworld'''\n"

	got, err := Get(source, "message")

	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got != "hello\nworld" {
		t.Fatalf("Get = %q, want multiline value", got)
	}
}

func TestListMultilineStrings(t *testing.T) {
	source := "message = \"\"\"hello\nworld\"\"\"\nliteral = '''alpha\nbeta'''\n"

	got, err := List(source, "")

	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	want := []Entry{
		{Key: "message", Value: "hello\nworld"},
		{Key: "literal", Value: "alpha\nbeta"},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}

func TestSetReplacesMultilineString(t *testing.T) {
	source := "message = \"\"\"hello\nworld\"\"\"\n"

	got, err := Set(source, "message", "short")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "message = \"short\"\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetStringWithNewlineWritesMultilineString(t *testing.T) {
	source := "message = \"short\"\n"

	got, err := SetString(source, "message", "hello\nworld")

	if err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}
	want := "message = \"\"\"hello\nworld\"\"\"\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetStringDoesNotInferScalarType(t *testing.T) {
	source := "port = 3000\n"

	got, err := SetString(source, "port", "4000")

	if err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}
	want := "port = \"4000\"\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestUnsetMultilineString(t *testing.T) {
	source := "title = \"demo\"\nmessage = \"\"\"hello\nworld\"\"\"\nnext = true\n"

	got, err := Unset(source, "message")

	if err != nil {
		t.Fatalf("Unset returned error: %v", err)
	}
	want := "title = \"demo\"\nnext = true\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}
