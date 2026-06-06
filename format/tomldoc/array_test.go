package tomldoc

import (
	"fmt"
	"testing"
)

func TestGetArrayValue(t *testing.T) {
	source := "ports = [3000, 3001]\nnames = [\"api\", \"worker\"]\n"

	got, err := Get(source, "names")

	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got != `[api, worker]` {
		t.Fatalf("Get = %q, want normalized array syntax", got)
	}
}

func TestListArrayValues(t *testing.T) {
	source := "[server]\nports = [3000, 3001]\nnames = [\"api\", \"worker\"]\n"

	got, err := List(source, "server")

	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	want := []Entry{
		{Key: "server.ports.0", Value: "3000"},
		{Key: "server.ports.1", Value: "3001"},
		{Key: "server.names.0", Value: `api`},
		{Key: "server.names.1", Value: `worker`},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}

func TestListArrayValueByPrefix(t *testing.T) {
	source := "[server]\nports = [3000, 3001]\n"

	got, err := List(source, "server.ports")

	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	want := []Entry{
		{Key: "server.ports.0", Value: "3000"},
		{Key: "server.ports.1", Value: "3001"},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}

func TestSetArrayValue(t *testing.T) {
	source := "[server]\nports = [1000]\n"

	got, err := SetArray(source, "server.ports", []string{"3000", "3001"})

	if err != nil {
		t.Fatalf("SetArray returned error: %v", err)
	}
	want := "[server]\nports = [3000, 3001]\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetArrayValueInfersElementTypes(t *testing.T) {
	source := "values = []\n"

	got, err := SetArray(source, "values", []string{"api", "true", "3000"})

	if err != nil {
		t.Fatalf("SetArray returned error: %v", err)
	}
	want := "values = [\"api\", true, 3000]\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetReplacesArrayItemByIndex(t *testing.T) {
	source := "methods = [\"GET\", \"SET\"]\n"

	got, err := Set(source, "methods.1", "POST")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "methods = [\"GET\", \"POST\"]\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}
