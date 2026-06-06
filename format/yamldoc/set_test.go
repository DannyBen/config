package yamldoc

import "testing"

func TestSetReplacesExistingScalarWithMinimalPatch(t *testing.T) {
	source := "server:\n  port: 3000 # keep\n  enabled: true\n"

	got, err := Set(source, "server.port", "3001")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "server:\n  port: 3001 # keep\n  enabled: true\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetReplacesFlowMappingScalarWithMinimalPatch(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		value string
		want  string
	}{
		{
			name:  "first item",
			path:  "server.host",
			value: "example",
			want:  "server: { host: example, port: 3000 }\n",
		},
		{
			name:  "last item",
			path:  "server.port",
			value: "3001",
			want:  "server: { host: localhost, port: 3001 }\n",
		},
		{
			name:  "quoted item containing comma",
			path:  "server.host",
			value: "example",
			want:  "server: { host: \"example\", port: 3000 }\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := "server: { host: \"local,host\", port: 3000 }\n"
			if tt.name != "quoted item containing comma" {
				source = "server: { host: localhost, port: 3000 }\n"
			}

			got, err := Set(source, tt.path, tt.value)

			if err != nil {
				t.Fatalf("Set returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", tt.want, got)
			}
		})
	}
}

func TestSetInfersScalarsAndStringFallback(t *testing.T) {
	source := "count: 1\nratio: 1.0\nenabled: false\nmissing: old\nrelease: old\nname: old\n"

	got, err := Set(source, "count", "42")
	if err != nil {
		t.Fatalf("Set integer returned error: %v", err)
	}
	got, err = Set(got, "ratio", "3.14")
	if err != nil {
		t.Fatalf("Set float returned error: %v", err)
	}
	got, err = Set(got, "enabled", "true")
	if err != nil {
		t.Fatalf("Set bool returned error: %v", err)
	}
	got, err = Set(got, "missing", "null")
	if err != nil {
		t.Fatalf("Set null returned error: %v", err)
	}
	got, err = Set(got, "release", "2024-01-13")
	if err != nil {
		t.Fatalf("Set date returned error: %v", err)
	}
	got, err = Set(got, "name", "hello world")
	if err != nil {
		t.Fatalf("Set string fallback returned error: %v", err)
	}

	want := "count: 42\nratio: 3.14\nenabled: true\nmissing: null\nrelease: 2024-01-13\nname: hello world\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetStringForcesString(t *testing.T) {
	source := "version: 1.0\n"

	got, err := SetString(source, "version", "1.0")

	if err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}
	want := "version: \"1.0\"\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetPreservesExistingStringStyleWhenReplacingStrings(t *testing.T) {
	source := "single: 'old value'\ndouble: \"old value\"\nblock: |-\n  old value\nnumber: \"1\"\n"

	got, err := Set(source, "single", "new value")
	if err != nil {
		t.Fatalf("Set single returned error: %v", err)
	}
	got, err = Set(got, "double", "new value")
	if err != nil {
		t.Fatalf("Set double returned error: %v", err)
	}
	got, err = Set(got, "block", "new value")
	if err != nil {
		t.Fatalf("Set block returned error: %v", err)
	}
	got, err = Set(got, "number", "2")
	if err != nil {
		t.Fatalf("Set number returned error: %v", err)
	}

	want := "single: 'new value'\ndouble: \"new value\"\nblock: |-\n  new value\nnumber: 2\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetInsertsMissingScalarUnderExistingMapping(t *testing.T) {
	source := "server:\n  host: localhost\n"

	got, err := Set(source, "server.port", "3000")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "server:\n  host: localhost\n  port: 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetInsertsMissingScalarIntoFlowMapping(t *testing.T) {
	source := "server: { host: localhost, port: 3000 }\n"

	got, err := Set(source, "server.enabled", "true")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "server: { host: localhost, port: 3000, enabled: true }\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetCreatesMissingParentMappings(t *testing.T) {
	source := "title: demo\n"

	got, err := Set(source, "server.http.port", "3000")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "title: demo\nserver:\n  http:\n    port: 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetInsertsMissingScalarThroughRootSequenceIndex(t *testing.T) {
	source := "- name: admin\n- name: user\n"

	got, err := Set(source, "0.pass", "hello")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "- name: admin\n  pass: hello\n- name: user\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetInsertsMissingScalarThroughNestedSequenceIndex(t *testing.T) {
	source := "servers:\n  - name: api\n    port: 3000\n"

	got, err := Set(source, "servers.0.pass", "hello")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "servers:\n  - name: api\n    port: 3000\n    pass: hello\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetCreatesMissingParentMappingsThroughSequenceIndex(t *testing.T) {
	source := "- name: admin\n"

	got, err := Set(source, "0.auth.pass", "hello")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "- name: admin\n  auth:\n    pass: hello\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetArrayWritesSequenceValue(t *testing.T) {
	source := "tags: [old]\n"

	got, err := SetArray(source, "tags", []string{"api", "worker"})

	if err != nil {
		t.Fatalf("SetArray returned error: %v", err)
	}
	want := "tags: [api, worker]\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetArrayReplacesBlockSequence(t *testing.T) {
	source := "tags:\n  - old\n  - stale\n"

	got, err := SetArray(source, "tags", []string{"api", "worker"})

	if err != nil {
		t.Fatalf("SetArray returned error: %v", err)
	}
	want := "tags:\n  - api\n  - worker\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetArrayConvertsFlowSequenceAtThreshold(t *testing.T) {
	source := "tags: [dev, stage]\n"

	got, err := SetArray(source, "tags", []string{"dev", "stage", "api", "sales", "db"})

	if err != nil {
		t.Fatalf("SetArray returned error: %v", err)
	}
	want := "tags:\n  - dev\n  - stage\n  - api\n  - sales\n  - db\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestArrayAddAppendsMissingValuesAndCreatesArray(t *testing.T) {
	source := "roots: [$HOME/.cache]\n"

	got, err := ArrayAdd(source, "roots", []string{"/tmp", "$HOME/.cache"})

	if err != nil {
		t.Fatalf("ArrayAdd returned error: %v", err)
	}
	got, err = ArrayAdd(got, "extra", []string{"/var/tmp"})
	if err != nil {
		t.Fatalf("ArrayAdd returned error: %v", err)
	}
	want := "roots: [$HOME/.cache, /tmp]\nextra:\n  - /var/tmp\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestArrayDelRemovesValuesAndDeletesEmptyArray(t *testing.T) {
	source := "roots: [$HOME/.cache, /tmp]\nextra: [/var/tmp]\n"

	got, err := ArrayDel(source, "roots", []string{"/tmp", "/missing"})

	if err != nil {
		t.Fatalf("ArrayDel returned error: %v", err)
	}
	got, err = ArrayDel(got, "extra", []string{"/var/tmp"})
	if err != nil {
		t.Fatalf("ArrayDel returned error: %v", err)
	}
	got, err = ArrayDel(got, "absent", []string{"/tmp"})
	if err != nil {
		t.Fatalf("ArrayDel returned error: %v", err)
	}
	want := "roots: [$HOME/.cache]\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestArrayAddRefusesNonArrayValue(t *testing.T) {
	_, err := ArrayAdd("roots: /tmp\n", "roots", []string{"/var/tmp"})

	if err == nil {
		t.Fatal("expected non-array refusal")
	}
	if err.Error() != "roots is not an array" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetReplacesSequenceItemByIndex(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "flow",
			source: "methods: [GET, SET]\n",
			want:   "methods: [GET, POST]\n",
		},
		{
			name:   "block",
			source: "methods:\n  - GET\n  - SET\n",
			want:   "methods:\n  - GET\n  - POST\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Set(tt.source, "methods.1", "POST")
			if err != nil {
				t.Fatalf("Set returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", tt.want, got)
			}
		})
	}
}

func TestSetInUpdatesSelectedRecord(t *testing.T) {
	source := "servers:\n  - name: api\n    port: 3000\n  - name: worker\n    port: 3001\n"

	got, err := SetIn(source, "servers", "name:worker", "port", "3002")

	if err != nil {
		t.Fatalf("SetIn returned error: %v", err)
	}
	want := "servers:\n  - name: api\n    port: 3000\n  - name: worker\n    port: 3002\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetInUpdatesSelectedRootRecord(t *testing.T) {
	source := "- name: admin\n- name: user\n"

	got, err := SetIn(source, "", "name:user", "auth.type", "basic")

	if err != nil {
		t.Fatalf("SetIn returned error: %v", err)
	}
	want := "- name: admin\n- name: user\n  auth:\n    type: basic\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetInCreatesSelectedRecord(t *testing.T) {
	source := "servers:\n  - name: api\n    port: 3000\n"

	got, err := SetIn(source, "servers", "name:worker", "port", "3001")

	if err != nil {
		t.Fatalf("SetIn returned error: %v", err)
	}
	want := "servers:\n  - name: api\n    port: 3000\n  - name: worker\n    port: 3001\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetInCreatesMissingCollection(t *testing.T) {
	source := "title: demo\n"

	got, err := SetIn(source, "servers", "name:worker", "port", "3001")

	if err != nil {
		t.Fatalf("SetIn returned error: %v", err)
	}
	want := "title: demo\nservers:\n  - name: worker\n    port: 3001\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetInArrayUpdatesSelectedRecord(t *testing.T) {
	source := "servers:\n  - name: api\n    tags: [old]\n  - name: worker\n    tags: [old]\n"

	got, err := SetInArray(source, "servers", "name:worker", "tags", []string{"api", "worker"})

	if err != nil {
		t.Fatalf("SetInArray returned error: %v", err)
	}
	want := "servers:\n  - name: api\n    tags: [old]\n  - name: worker\n    tags: [api, worker]\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetPreservesAnchorToken(t *testing.T) {
	source := "host: &db_host localhost\n"

	got, err := Set(source, "host", "otherhost")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "host: &db_host otherhost\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetAliasPathNoopOrRefuse(t *testing.T) {
	source := "host: &host localhost\nreplica: *host\n"

	got, err := Set(source, "replica", "localhost")
	if err != nil {
		t.Fatalf("Set same alias value returned error: %v", err)
	}
	if got != source {
		t.Fatalf("same alias value should be no-op\nwant:\n%s\ngot:\n%s", source, got)
	}

	_, err = Set(source, "replica", "otherhost")
	if err == nil {
		t.Fatal("expected alias mutation refusal")
	}
	if err.Error() != "replica is an alias; refusing to change shared YAML state" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetMultilineStringUsesLiteralBlock(t *testing.T) {
	source := "help: short\n"

	got, err := SetString(source, "help", "line one\nline two\n")

	if err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}
	want := "help: |-\n  line one\n  line two\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetInfersTaggedScalar(t *testing.T) {
	source := "secret: old\n"

	got, err := Set(source, "secret", "!vault hello")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "secret: !vault hello\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetStringEscapesLeadingBang(t *testing.T) {
	source := "secret: old\n"

	got, err := SetString(source, "secret", "!a-secret")

	if err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}
	want := "secret: \"!a-secret\"\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetNestedValuePreservesTaggedMapping(t *testing.T) {
	source := "secret: !vault\n  encrypted: value\n  version: 1\n"

	got, err := Set(source, "secret.version", "2")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "secret: !vault\n  encrypted: value\n  version: 2\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}
