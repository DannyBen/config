package tomldoc

import (
	"fmt"
	"testing"
)

func TestListFlattensInlineTable(t *testing.T) {
	source := "pool = { min = 1, max = 10 }\n"

	got, err := List(source, "pool")

	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	want := []Entry{
		{Key: "pool.min", Value: "1"},
		{Key: "pool.max", Value: "10"},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}

func TestGetInlineTableFails(t *testing.T) {
	source := "pool = { min = 1, max = 10 }\n"

	_, err := Get(source, "pool")

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "pool is a table, not a value" {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestSetInlineTableChildPatchesValue(t *testing.T) {
	source := "pool = { min = 1, max = 10 }\n"

	got, err := Set(source, "pool.min", "2")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "pool = { min = 2, max = 10 }\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestUnsetInlineTableChildMaintainsInlineTable(t *testing.T) {
	source := "pool = { min = 1, max = 10 }\n"

	got, err := Unset(source, "pool.min")

	if err != nil {
		t.Fatalf("Unset returned error: %v", err)
	}
	want := "pool = { max = 10 }\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestUnsetLastInlineTableChildMaintainsInlineTable(t *testing.T) {
	source := "pool = { min = 1, max = 10 }\n"

	got, err := Unset(source, "pool.max")

	if err != nil {
		t.Fatalf("Unset returned error: %v", err)
	}
	want := "pool = { min = 1 }\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetMissingInlineTableChildInsertsInPlace(t *testing.T) {
	source := "pool = { min = 1, max = 10 }\n"

	got, err := Set(source, "pool.default", "10")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "pool = { min = 1, max = 10, default = 10 }\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetMissingNestedInlineTableChildInsertsInPlace(t *testing.T) {
	source := "pool = { limits = { min = 1, max = 10 } }\n"

	got, err := Set(source, "pool.limits.default", "10")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "pool = { limits = { min = 1, max = 10, default = 10 } }\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetMissingEmptyInlineTableChildInsertsInPlace(t *testing.T) {
	source := "pool = {}\n"

	got, err := Set(source, "pool.default", "10")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "pool = { default = 10 }\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}
