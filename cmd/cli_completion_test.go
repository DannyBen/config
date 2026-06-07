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
