package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetPrintsTOMLValue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[database]\nport = 5432\n")

	err := Execute([]string{"get", "-f", path, "database.port"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.String() != "5432\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestGetPrintsSelectedTOMLRecordValue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[[servers]]\nname = \"api\"\nport = 3000\n\n[[servers]]\nname = \"worker\"\nport = 3001\n")

	err := Execute([]string{"get", "-f", path, "port", "--in", "servers", "--on", "name:worker"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.String() != "3001\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestGetSelectedRecordRequiresInWithOn(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[[servers]]\nname = \"api\"\nport = 3000\n")

	err := Execute([]string{"get", "-f", path, "port", "--on", "name:api"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "flag --on requires --in" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetSelectedRecordRequiresOnWithIn(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[[servers]]\nname = \"api\"\nport = 3000\n")

	err := Execute([]string{"get", "-f", path, "port", "--in", "servers"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "flag --in requires --on" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetUsesConfigFileEnv(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[database]\nport = 5432\n")
	t.Setenv("CONFIG_FILE", path)

	err := Execute([]string{"get", "database.port"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.String() != "5432\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestGetFailsWhenConfigFileIsNotSpecified(t *testing.T) {
	clearConfigFileEnv(t)
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"get", "database.port"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if err.Error() != "config file not specified" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetUnsupportedExplicitConfigFileReportsUnsupportedFormat(t *testing.T) {
	clearConfigFileEnv(t)
	var stdout, stderr bytes.Buffer
	path := filepath.Join(t.TempDir(), "config.conf")
	if err := os.WriteFile(path, []byte("port = 3000\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := Execute([]string{"get", "-f", path, "database.port"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if err.Error() != "unsupported config format for "+path {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetPrintsJSONValue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempJSON(t, `{"database":{"port":5432}}`)

	err := Execute([]string{"get", "-f", path, "database.port"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.String() != "5432\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestGetPrintsINIValue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempINI(t, "title = demo app\n[server]\nport = 3000\n")

	err := Execute([]string{"get", "-f", path, "server.port"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.String() != "3000\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestFileFlagOverridesConfigFileEnv(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[database]\nport = 5432\n")
	t.Setenv("CONFIG_FILE", path)
	yamlPath := writeTempYAML(t, "database:\n  port: 5432\n")

	err := Execute([]string{"get", "-f", yamlPath, "database.port"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.String() != "5432\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestPositionalTargetPathIsNotAccepted(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[database]\nport = 5432\n")

	err := Execute([]string{"get", path, "database.port"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if !strings.Contains(err.Error(), "usage: config get KEY") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetMissingRequiredArgFailsAfterConfigFileIsResolved(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[database]\nport = 5432\n")

	err := Execute([]string{"get", "-f", path}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if !strings.Contains(err.Error(), "usage: config get KEY") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListPrintsTOMLValues(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "title = \"demo app\"\n\n[server]\nport = 3000\nenabled = true\n")

	err := Execute([]string{"list", "-f", path}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "title=demo app\nserver.port=3000\nserver.enabled=true\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestListAliasPrintsTOMLValues(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[server]\nport = 3000\n")

	err := Execute([]string{"ls", "-f", path}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "server.port=3000\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestListPrintsTOMLValuesWithColor(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[server]\nport = 3000\n")

	err := Execute([]string{"list", "-f", path, "--color"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "\x1b[36mserver.port\x1b[0m\x1b[33m=\x1b[0m3000\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%q\ngot:\n%q", want, stdout.String())
	}
}

func TestListPrintsTOMLValuesUnderTable(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "title = \"demo app\"\n\n[server]\nport = 3000\nenabled = true\n")

	err := Execute([]string{"list", "-f", path, "server"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "server.port=3000\nserver.enabled=true\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestListPrintsINIValuesUnderSection(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempINI(t, "title = demo app\n[server]\nport = 3000\nenabled = true\n")

	err := Execute([]string{"list", "-f", path, "server"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "server.port=3000\nserver.enabled=true\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestListUsesSingleArgumentAsKeyWhenConfigFileEnvIsSet(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[server]\nports = [3000, 3001]\n")
	t.Setenv("CONFIG_FILE", path)

	err := Execute([]string{"list", "server.ports"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "server.ports.0=3000\nserver.ports.1=3001\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestListFailsWhenConfigFileIsNotSpecified(t *testing.T) {
	clearConfigFileEnv(t)
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"list", "server.ports"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "config file not specified" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDumpPrintsTOMLAsYAML(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "title = \"demo app\"\n\n[server]\nport = 3000\nenabled = true\n")

	err := Execute([]string{"dump", "-f", path}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "server:\n  enabled: true\n  port: 3000\ntitle: demo app\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestDumpPrintsTOMLSubtree(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "title = \"demo app\"\n\n[server]\nports = [3000, 3001]\nenabled = true\n")

	err := Execute([]string{"dump", "-f", path, "server"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "enabled: true\nports:\n  - 3000\n  - 3001\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestDumpPrintsTOMLSubtreeAsJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "title = \"demo app\"\n\n[server]\nports = [3000, 3001]\nenabled = true\n")

	err := Execute([]string{"dump", "-f", path, "server", "--json"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "{\n  \"enabled\": true,\n  \"ports\": [\n    3000,\n    3001\n  ]\n}\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestDumpPrintsYAMLSubtree(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempYAML(t, "server:\n  hosts:\n    - api.example.com\n    - worker.example.com\n  enabled: true\n")

	err := Execute([]string{"dump", "-f", path, "server.hosts"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "- api.example.com\n- worker.example.com\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestDumpPrintsYAMLSubtreeAsJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempYAML(t, "server:\n  hosts:\n    - api.example.com\n    - worker.example.com\n  enabled: true\n")

	err := Execute([]string{"dump", "-f", path, "server", "--json"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "{\n  \"enabled\": true,\n  \"hosts\": [\n    \"api.example.com\",\n    \"worker.example.com\"\n  ]\n}\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestDumpUsesSingleArgumentAsKeyWhenConfigFileEnvIsSet(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[server]\nport = 3000\n")
	t.Setenv("CONFIG_FILE", path)

	err := Execute([]string{"dump", "server"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.String() != "port: 3000\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestDumpFailsWhenConfigFileIsNotSpecified(t *testing.T) {
	clearConfigFileEnv(t)
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"dump", "server"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "config file not specified" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"--version"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.String() != "1.2.3\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}
