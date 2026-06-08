package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestCompletionGeneratesBash(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"completion", "bash"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "# bash completion V2 for config")
	assertContains(t, stdout.String(), "__start_config()")
	assertContains(t, stdout.String(), "completion")
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestCompletionGeneratesZsh(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"completion", "zsh"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "#compdef config")
	assertContains(t, stdout.String(), "# zsh completion for config")
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestCompletionGeneratesFish(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"completion", "fish"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "# fish completion for config")
	assertContains(t, stdout.String(), "complete -c config")
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestCompletionRejectsUnsupportedShell(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"completion", "xonsh"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `unsupported shell "xonsh"`) {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestCompletionRequiresShell(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"completion"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "usage: config completion SHELL" {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestCompleteGetKeyFromFileFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[env.prod.servers]\nport = 3000\nhost = \"api\"\n\n[env.dev.servers]\nport = 3001\n")

	err := Execute([]string{"__complete", "get", "-f", path, "env.prod."}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertCompletionContains(t, stdout.String(), "env.prod.servers.port")
	assertCompletionContains(t, stdout.String(), "env.prod.servers.host")
	assertCompletionOmits(t, stdout.String(), "env.dev.servers.port")
	assertContains(t, stdout.String(), ":4")
}

func TestCompleteListKeyFromConfigFileEnv(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[server]\nport = 3000\nenabled = true\n")
	t.Setenv("CONFIG_FILE", path)

	err := Execute([]string{"__complete", "list", "server."}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertCompletionContains(t, stdout.String(), "server.port")
	assertCompletionContains(t, stdout.String(), "server.enabled")
	assertContains(t, stdout.String(), ":4")
}

func TestCompleteSetDoesNotCompleteValuePosition(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[server]\nport = 3000\n")

	err := Execute([]string{"__complete", "set", "-f", path, "server.port", ""}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertCompletionOmits(t, stdout.String(), "server.port")
	assertContains(t, stdout.String(), ":4")
}

func TestCompleteDeleteKeyUsesContainerKeys(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[server]\nport = 3000\n\n[database]\nhost = \"localhost\"\n")

	err := Execute([]string{"__complete", "delete", "-f", path, "ser"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertCompletionContains(t, stdout.String(), "server")
	assertCompletionOmits(t, stdout.String(), "server.port")
	assertCompletionOmits(t, stdout.String(), "database")
	assertContains(t, stdout.String(), ":4")
}

func TestCompleteUnsetKeyUsesLeafKeys(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[server]\nport = 3000\nhost = \"localhost\"\n")

	err := Execute([]string{"__complete", "unset", "-f", path, "ser"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertCompletionContains(t, stdout.String(), "server.port")
	assertCompletionContains(t, stdout.String(), "server.host")
	assertCompletionOmits(t, stdout.String(), "server")
	assertContains(t, stdout.String(), ":4")
}

func TestCompleteListKeyUsesLeafAndContainerKeys(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[server]\nport = 3000\n")

	err := Execute([]string{"__complete", "list", "-f", path, "ser"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertCompletionContains(t, stdout.String(), "server")
	assertCompletionContains(t, stdout.String(), "server.port")
	assertContains(t, stdout.String(), ":4")
}

func TestCompleteArrayKeyUsesScalarArrayKeys(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempYAML(t, "tags:\n  - api\n  - worker\nservers:\n  - name: api\n    port: 3000\n")

	err := Execute([]string{"__complete", "array", "add", "-f", path, ""}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertCompletionContains(t, stdout.String(), "tags")
	assertCompletionOmits(t, stdout.String(), "tags.0")
	assertCompletionOmits(t, stdout.String(), "servers")
	assertCompletionOmits(t, stdout.String(), "servers.0.name")
	assertContains(t, stdout.String(), ":4")
}

func TestCompleteKeySilentlySkipsMissingConfigFile(t *testing.T) {
	clearConfigFileEnv(t)
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"__complete", "get", "server."}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.String() != ":4\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func assertCompletionContains(t *testing.T, output, value string) {
	t.Helper()
	for _, line := range strings.Split(output, "\n") {
		if line == value {
			return
		}
	}
	t.Fatalf("completion %q not found in:\n%s", value, output)
}

func assertCompletionOmits(t *testing.T, output, value string) {
	t.Helper()
	for _, line := range strings.Split(output, "\n") {
		if line == value {
			t.Fatalf("completion %q found in:\n%s", value, output)
		}
	}
}
