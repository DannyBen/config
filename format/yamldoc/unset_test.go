package yamldoc

import "testing"

func TestUnsetRemovesScalarLineAndKeepsPrecedingComment(t *testing.T) {
	source := "server:\n  # keep this comment\n  port: 3000 # remove this inline comment\n  host: localhost\n"

	got, err := Unset(source, "server.port")

	if err != nil {
		t.Fatalf("Unset returned error: %v", err)
	}
	want := "server:\n  # keep this comment\n  host: localhost\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestUnsetDoesNotPruneEmptyParentMapping(t *testing.T) {
	source := "server:\n  port: 3000\n"

	got, err := Unset(source, "server.port")

	if err != nil {
		t.Fatalf("Unset returned error: %v", err)
	}
	want := "server:\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestUnsetRemovesBlockScalar(t *testing.T) {
	source := "help: |-\n  line one\n  line two\nname: demo\n"

	got, err := Unset(source, "help")

	if err != nil {
		t.Fatalf("Unset returned error: %v", err)
	}
	want := "name: demo\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestUnsetRemovesFlowMappingScalar(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "first item",
			path: "server.host",
			want: "server: { port: 3000, enabled: true }\n",
		},
		{
			name: "last item",
			path: "server.enabled",
			want: "server: { host: localhost, port: 3000 }\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := "server: { host: localhost, port: 3000, enabled: true }\n"

			got, err := Unset(source, tt.path)

			if err != nil {
				t.Fatalf("Unset returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", tt.want, got)
			}
		})
	}
}

func TestUnsetSelectedRecordValue(t *testing.T) {
	source := "servers:\n  - name: api\n    port: 3000\n  - name: worker\n    port: 3001\n    host: worker.local\n"

	got, err := UnsetIn(source, "servers", []string{"name:worker"}, "port")

	if err != nil {
		t.Fatalf("UnsetIn returned error: %v", err)
	}
	want := "servers:\n  - name: api\n    port: 3000\n  - name: worker\n    host: worker.local\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestUnsetSelectedRecordWithMultipleSelectors(t *testing.T) {
	source := "servers:\n  - name: api\n    host: primary.local\n    port: 3000\n  - name: api\n    host: backup.local\n    port: 3001\n"

	got, err := UnsetIn(source, "servers", []string{"name:api", "host:backup.local"}, "port")

	if err != nil {
		t.Fatalf("UnsetIn returned error: %v", err)
	}
	want := "servers:\n  - name: api\n    host: primary.local\n    port: 3000\n  - name: api\n    host: backup.local\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestUnsetMissingScalarFails(t *testing.T) {
	source := "server:\n  port: 3000\n"

	_, err := Unset(source, "server.host")

	if err == nil {
		t.Fatal("expected missing value error")
	}
	if err.Error() != "server.host is not set" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnsetRefusesContainers(t *testing.T) {
	tests := []struct {
		name   string
		source string
		path   string
		want   string
	}{
		{
			name:   "mapping",
			source: "server:\n  port: 3000\n",
			path:   "server",
			want:   "server is a container, not a scalar value",
		},
		{
			name:   "sequence",
			source: "tags:\n  - api\n",
			path:   "tags",
			want:   "tags is a container, not a scalar value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Unset(tt.source, tt.path)
			if err == nil {
				t.Fatal("expected container refusal")
			}
			if err.Error() != tt.want {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestUnsetRefusesAnchoredScalarStillReferenced(t *testing.T) {
	source := "host: &host localhost\nreplica: *host\n"

	_, err := Unset(source, "host")

	if err == nil {
		t.Fatal("expected anchor reference refusal")
	}
	if err.Error() != "host defines anchor \"host\" that is still referenced" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnsetAllowsUnreferencedAnchoredScalar(t *testing.T) {
	source := "host: &host localhost\nport: 3000\n"

	got, err := Unset(source, "host")

	if err != nil {
		t.Fatalf("Unset returned error: %v", err)
	}
	want := "port: 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestUnsetAliasUseSite(t *testing.T) {
	source := "host: &host localhost\nreplica: *host\n"

	got, err := Unset(source, "replica")

	if err != nil {
		t.Fatalf("Unset returned error: %v", err)
	}
	want := "host: &host localhost\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}
