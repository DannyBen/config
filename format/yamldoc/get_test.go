package yamldoc

import (
	"strings"
	"testing"
)

func TestGetScalars(t *testing.T) {
	source := "title: demo app\nserver:\n  port: 3000\n  enabled: true\n  missing: null\n"

	tests := map[string]string{
		"title":          "demo app",
		"server.port":    "3000",
		"server.enabled": "true",
		"server.missing": "null",
	}

	for path, want := range tests {
		got, err := Get(source, path)
		if err != nil {
			t.Fatalf("Get(%q) returned error: %v", path, err)
		}
		if got != want {
			t.Fatalf("Get(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestGetUsesDoubleDotEscapingForLiteralDotPath(t *testing.T) {
	source := "server:\n  public.port: 3000\n"

	got, err := Get(source, "server.public..port")

	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got != "3000" {
		t.Fatalf("Get = %q, want 3000", got)
	}
}

func TestGetReadsScalarSequencesAsFlowValues(t *testing.T) {
	source := "tags:\n  - api\n  - worker\nports:\n  - 3000\n  - 3001\n"

	got, err := Get(source, "tags")
	if err != nil {
		t.Fatalf("Get tags returned error: %v", err)
	}
	if got != "[api, worker]" {
		t.Fatalf("Get tags = %q, want [api, worker]", got)
	}

	got, err = Get(source, "ports")
	if err != nil {
		t.Fatalf("Get ports returned error: %v", err)
	}
	if got != "[3000, 3001]" {
		t.Fatalf("Get ports = %q, want [3000, 3001]", got)
	}
}

func TestGetRefusesMappingsAndRecordSequences(t *testing.T) {
	source := "server:\n  port: 3000\nservers:\n  - name: api\n    port: 3000\n"

	_, err := Get(source, "server")
	if err == nil {
		t.Fatal("expected mapping error")
	}
	if err.Error() != "server is a mapping, not a value" {
		t.Fatalf("unexpected mapping error: %v", err)
	}

	_, err = Get(source, "servers")
	if err == nil {
		t.Fatal("expected record sequence error")
	}
	if err.Error() != "servers is a sequence of records, not a value" {
		t.Fatalf("unexpected record sequence error: %v", err)
	}
}

func TestGetResolvesAliases(t *testing.T) {
	source := "host: &host localhost\nreplica: *host\ntags: &tags\n  - api\n  - worker\nactive_tags: *tags\n"

	got, err := Get(source, "replica")
	if err != nil {
		t.Fatalf("Get alias scalar returned error: %v", err)
	}
	if got != "localhost" {
		t.Fatalf("Get alias scalar = %q, want localhost", got)
	}

	got, err = Get(source, "active_tags")
	if err != nil {
		t.Fatalf("Get alias sequence returned error: %v", err)
	}
	if got != "[api, worker]" {
		t.Fatalf("Get alias sequence = %q, want [api, worker]", got)
	}
}

func TestGetInSelectedRecord(t *testing.T) {
	source := "servers:\n  - name: api\n    role: web\n    port: 3000\n  - name: worker\n    role: worker\n    port: 3001\n"

	got, err := GetIn(source, "servers", []string{"name:worker"}, "port")

	if err != nil {
		t.Fatalf("GetIn returned error: %v", err)
	}
	if got != "3001" {
		t.Fatalf("GetIn = %q, want 3001", got)
	}
}

func TestGetInSupportsMultipleSelectors(t *testing.T) {
	source := "servers:\n  - name: api\n    role: web\n    port: 3000\n  - name: api\n    role: worker\n    port: 3001\n"

	got, err := GetIn(source, "servers", []string{"name:api", "role:worker"}, "port")

	if err != nil {
		t.Fatalf("GetIn returned error: %v", err)
	}
	if got != "3001" {
		t.Fatalf("GetIn = %q, want 3001", got)
	}
}

func TestGetInTreatsNonScalarSelectorValuesAsNonMatches(t *testing.T) {
	source := "commands:\n  - name: remove\n    alias: r\n  - name: pull\n    alias: [p, download, update]\n"

	got, err := GetIn(source, "commands", []string{"alias:r"}, "name")

	if err != nil {
		t.Fatalf("GetIn returned error: %v", err)
	}
	if got != "remove" {
		t.Fatalf("GetIn = %q, want remove", got)
	}
}

func TestGetInRefusals(t *testing.T) {
	source := "server:\n  port: 3000\nservers:\n  - name: api\n    meta:\n      role: web\n    port: 3000\n  - name: api\n    port: 3001\n"

	tests := map[string]string{
		"not sequence":       "server is not a sequence of records",
		"missing selector":   "servers has no records matching name:worker",
		"multiple selector":  "servers has multiple records matching name:api",
		"container no match": "servers has no records matching meta:web",
	}

	_, err := GetIn(source, "server", []string{"port:3000"}, "port")
	assertError(t, err, tests["not sequence"])

	_, err = GetIn(source, "servers", []string{"name:worker"}, "port")
	assertError(t, err, tests["missing selector"])

	_, err = GetIn(source, "servers", []string{"name:api"}, "port")
	assertError(t, err, tests["multiple selector"])

	_, err = GetIn(source, "servers", []string{"meta:web"}, "port")
	assertError(t, err, tests["container no match"])
}

func assertError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %q, want to contain %q", err.Error(), want)
	}
}
