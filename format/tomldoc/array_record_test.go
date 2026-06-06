package tomldoc

import (
	"fmt"
	"strings"
	"testing"
)

func TestListArrayTableValuesUsesIndexedPaths(t *testing.T) {
	source := strings.Join([]string{
		"[[servers]]",
		"name = \"api\"",
		"port = 3000",
		"",
		"[[servers]]",
		"name = \"worker\"",
		"port = 3001",
		"",
	}, "\n")

	got, err := List(source, "servers")

	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	want := []Entry{
		{Key: "servers.0.name", Value: "api"},
		{Key: "servers.0.port", Value: "3000"},
		{Key: "servers.1.name", Value: "worker"},
		{Key: "servers.1.port", Value: "3001"},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}

func TestSetArrayTableValueByIndexedPath(t *testing.T) {
	source := strings.Join([]string{
		"[[servers]]",
		"name = \"api\"",
		"port = 3000 # keep",
		"",
	}, "\n")

	got, err := Set(source, "servers.0.port", "8080")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := strings.Join([]string{
		"[[servers]]",
		"name = \"api\"",
		"port = 8080 # keep",
		"",
	}, "\n")
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetArrayTableValueBySelector(t *testing.T) {
	source := strings.Join([]string{
		"[[servers]]",
		"name = \"api\"",
		"port = 3000",
		"",
		"[[servers]]",
		"name = \"worker\"",
		"port = 3001",
		"",
	}, "\n")

	got, err := SetIn(source, "servers", "name:worker", "port", "8080")

	if err != nil {
		t.Fatalf("SetIn returned error: %v", err)
	}
	want := strings.Join([]string{
		"[[servers]]",
		"name = \"api\"",
		"port = 3000",
		"",
		"[[servers]]",
		"name = \"worker\"",
		"port = 8080",
		"",
	}, "\n")
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetArrayTableValueBySelectorCreatesRecord(t *testing.T) {
	source := "title = \"demo\"\n"

	got, err := SetIn(source, "servers", "name:web", "port", "3000")

	if err != nil {
		t.Fatalf("SetIn returned error: %v", err)
	}
	want := strings.Join([]string{
		"title = \"demo\"",
		"",
		"[[servers]]",
		"name = \"web\"",
		"port = 3000",
		"",
	}, "\n")
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetArrayTableValueBySelectorCreatesCRLFRecord(t *testing.T) {
	source := "title = \"demo\"\r\n"

	got, err := SetIn(source, "servers", "name:web", "port", "3000")

	if err != nil {
		t.Fatalf("SetIn returned error: %v", err)
	}
	want := "title = \"demo\"\r\n\r\n[[servers]]\r\nname = \"web\"\r\nport = 3000\r\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%q\ngot:\n%q", want, got)
	}
}

func TestSetArrayTableValueBySelectorRejectsMultipleMatches(t *testing.T) {
	source := strings.Join([]string{
		"[[servers]]",
		"name = \"web\"",
		"port = 3000",
		"",
		"[[servers]]",
		"name = \"web\"",
		"port = 3001",
		"",
	}, "\n")

	_, err := SetIn(source, "servers", "name:web", "port", "8080")

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "servers has multiple records matching name:web" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetArrayTableValueBySelectorRejectsEqualsSyntax(t *testing.T) {
	source := "[[servers]]\nname = \"api\"\nport = 3000\n"

	_, err := SetIn(source, "servers", "name=api", "port", "8080")

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "--on must use FIELD:VALUE" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetArrayTableValueBySelector(t *testing.T) {
	source := strings.Join([]string{
		"[[servers]]",
		"name = \"api\"",
		"port = 3000",
		"",
		"[[servers]]",
		"name = \"worker\"",
		"port = 3001",
		"",
	}, "\n")

	got, err := GetIn(source, "servers", []string{"name:worker"}, "port")

	if err != nil {
		t.Fatalf("GetIn returned error: %v", err)
	}
	if got != "3001" {
		t.Fatalf("GetIn = %q, want 3001", got)
	}
}

func TestGetArrayTableValueByMultipleSelectors(t *testing.T) {
	source := strings.Join([]string{
		"[[servers]]",
		"name = \"api\"",
		"host = \"api.local\"",
		"port = 3000",
		"",
		"[[servers]]",
		"name = \"api\"",
		"host = \"backup.local\"",
		"port = 3001",
		"",
	}, "\n")

	got, err := GetIn(source, "servers", []string{"name:api", "host:backup.local"}, "port")

	if err != nil {
		t.Fatalf("GetIn returned error: %v", err)
	}
	if got != "3001" {
		t.Fatalf("GetIn = %q, want 3001", got)
	}
}

func TestGetArrayTableValueBySelectorRejectsMissingMatch(t *testing.T) {
	source := "[[servers]]\nname = \"api\"\nport = 3000\n"

	_, err := GetIn(source, "servers", []string{"name:web"}, "port")

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "servers has no records matching name:web" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetArrayTableValueBySelectorRejectsMultipleMatches(t *testing.T) {
	source := strings.Join([]string{
		"[[servers]]",
		"name = \"api\"",
		"port = 3000",
		"",
		"[[servers]]",
		"name = \"api\"",
		"port = 3001",
		"",
	}, "\n")

	_, err := GetIn(source, "servers", []string{"name:api"}, "port")

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "servers has multiple records matching name:api" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetArrayTableCollectionFails(t *testing.T) {
	source := "[[servers]]\nname = \"api\"\nport = 3000\n"

	_, err := Get(source, "servers")

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "servers is an array of records, not a value" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnsetArrayTableValueBySelector(t *testing.T) {
	source := strings.Join([]string{
		"[[servers]]",
		"name = \"api\"",
		"port = 3000",
		"",
		"[[servers]]",
		"name = \"worker\"",
		"port = 3001",
		"",
	}, "\n")

	got, err := UnsetIn(source, "servers", []string{"name:worker"}, "port")

	if err != nil {
		t.Fatalf("UnsetIn returned error: %v", err)
	}
	want := strings.Join([]string{
		"[[servers]]",
		"name = \"api\"",
		"port = 3000",
		"",
		"[[servers]]",
		"name = \"worker\"",
		"",
	}, "\n")
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestUnsetArrayTableValueByMultipleSelectors(t *testing.T) {
	source := strings.Join([]string{
		"[[servers]]",
		"name = \"api\"",
		"host = \"api.local\"",
		"port = 3000",
		"",
		"[[servers]]",
		"name = \"api\"",
		"host = \"backup.local\"",
		"port = 3001",
		"",
	}, "\n")

	got, err := UnsetIn(source, "servers", []string{"name:api", "host:backup.local"}, "port")

	if err != nil {
		t.Fatalf("UnsetIn returned error: %v", err)
	}
	want := strings.Join([]string{
		"[[servers]]",
		"name = \"api\"",
		"host = \"api.local\"",
		"port = 3000",
		"",
		"[[servers]]",
		"name = \"api\"",
		"host = \"backup.local\"",
		"",
	}, "\n")
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestUnsetArrayTableValueBySelectorRejectsMissingField(t *testing.T) {
	source := "[[servers]]\nname = \"api\"\nport = 3000\n"

	_, err := UnsetIn(source, "servers", []string{"name:api"}, "host")

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "servers.0.host is not set" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetArrayTableIndexedPathRejectsMissingIndex(t *testing.T) {
	source := "[[servers]]\nname = \"api\"\n"

	_, err := Set(source, "servers.1.port", "8080")

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "servers.1 is not set" {
		t.Fatalf("unexpected error: %v", err)
	}
}
