package yamldoc

import "testing"

func TestDeleteMappingContainer(t *testing.T) {
	source := "title: demo\n\nserver:\n  port: 3000\n  host: localhost\n\nowner: ops\n"

	got, err := Delete(source, "server", nil)

	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	want := "title: demo\n\nowner: ops\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestDeleteSequenceContainer(t *testing.T) {
	source := "server:\n  port: 3000\nservers:\n  - name: api\n    port: 3000\n  - name: worker\n    port: 3001\nowner: ops\n"

	got, err := Delete(source, "servers", nil)

	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	want := "server:\n  port: 3000\nowner: ops\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestDeleteSelectedRecord(t *testing.T) {
	source := "servers:\n  - name: api\n    port: 3000\n  - name: worker\n    port: 3001\n    host: worker.local\n"

	got, err := Delete(source, "servers", []string{"name:worker"})

	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	want := "servers:\n  - name: api\n    port: 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestDeleteSelectedRecordWithMultipleSelectors(t *testing.T) {
	source := "servers:\n  - name: api\n    host: primary.local\n    port: 3000\n  - name: api\n    host: backup.local\n    port: 3001\n"

	got, err := Delete(source, "servers", []string{"name:api", "host:backup.local"})

	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	want := "servers:\n  - name: api\n    host: primary.local\n    port: 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestDeleteIndexedRecord(t *testing.T) {
	source := "servers:\n  - name: api\n    port: 3000\n  - name: worker\n    port: 3001\n"

	got, err := Delete(source, "servers.0", nil)

	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	want := "servers:\n  - name: worker\n    port: 3001\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestDeleteDashOnlyRecord(t *testing.T) {
	source := "servers:\n  -\n    name: api\n    port: 3000\n  -\n    name: worker\n    port: 3001\n"

	got, err := Delete(source, "servers.1", nil)

	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	want := "servers:\n  -\n    name: api\n    port: 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestDeleteIfEmptyIgnoresMissingPath(t *testing.T) {
	source := "server:\n  port: 3000\n"

	got, err := DeleteIfEmpty(source, "server.missing")

	if err != nil {
		t.Fatalf("DeleteIfEmpty returned error: %v", err)
	}
	if got != source {
		t.Fatalf("source changed\nwant:\n%s\ngot:\n%s", source, got)
	}
}

func TestDeleteIfEmptyIgnoresMissingPathInEmptyDocument(t *testing.T) {
	got, err := DeleteIfEmpty("", "server")

	if err != nil {
		t.Fatalf("DeleteIfEmpty returned error: %v", err)
	}
	if got != "" {
		t.Fatalf("source changed: %q", got)
	}
}

func TestDeleteRefusesScalarValue(t *testing.T) {
	_, err := Delete("server:\n  port: 3000\n", "server.port", nil)

	if err == nil {
		t.Fatal("expected scalar refusal")
	}
	if err.Error() != "server.port is a value, use unset to remove fields" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteRefusesScalarSequenceItem(t *testing.T) {
	_, err := Delete("tags:\n  - api\n  - worker\n", "tags.0", nil)

	if err == nil {
		t.Fatal("expected scalar sequence refusal")
	}
	if err.Error() != "tags is not a sequence of records" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteRefusesUnsafeFlowMappingContainer(t *testing.T) {
	_, err := Delete("server: { database: { port: 5432 }, host: localhost }\n", "server.database", nil)

	if err == nil {
		t.Fatal("expected unsafe flow mapping refusal")
	}
	if err.Error() != "server.database cannot be safely deleted" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteRefusesUnsafeFlowSequenceRecord(t *testing.T) {
	_, err := Delete("servers: [{ name: api, port: 3000 }, { name: worker, port: 3001 }]\n", "servers.0", nil)

	if err == nil {
		t.Fatal("expected unsafe flow sequence refusal")
	}
	if err.Error() != "servers.0 cannot be safely deleted" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteRefusesAnchoredContainerStillReferenced(t *testing.T) {
	source := "server: &server\n  port: 3000\nreplica: *server\n"

	_, err := Delete(source, "server", nil)

	if err == nil {
		t.Fatal("expected anchor reference refusal")
	}
	if err.Error() != "server defines anchor \"server\" that is still referenced" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteRefusesAliasUseSite(t *testing.T) {
	source := "server: &server\n  port: 3000\nreplica: *server\n"

	_, err := Delete(source, "replica", nil)

	if err == nil {
		t.Fatal("expected alias refusal")
	}
	if err.Error() != "replica is an alias; refusing to delete shared YAML state" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteRefusesAliasedRecordUseSite(t *testing.T) {
	source := "template: &template\n  name: api\n  port: 3000\nservers:\n  - *template\n"

	_, err := Delete(source, "servers", []string{"name:api"})

	if err == nil {
		t.Fatal("expected alias record refusal")
	}
	if err.Error() != "servers.0 is an alias; refusing to delete shared YAML state" {
		t.Fatalf("unexpected error: %v", err)
	}
}
