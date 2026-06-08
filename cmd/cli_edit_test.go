package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func writeEditorScript(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "editor")
	script := "#!/bin/sh\nprintf '%s\\n' \"$1\" > \"$CONFIG_EDIT_LOG\"\nprintf 'edited = true\\n' >> \"$1\"\n"
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestEditOpensExplicitConfigFileInEditor(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "name = \"demo\"\n")
	logPath := filepath.Join(t.TempDir(), "editor.log")
	t.Setenv("EDITOR", writeEditorScript(t))
	t.Setenv("CONFIG_EDIT_LOG", logPath)

	err := Execute([]string{"edit", "-f", path}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	if readFile(t, logPath) != path+"\n" {
		t.Fatalf("editor log = %q", readFile(t, logPath))
	}
	assertContains(t, readFile(t, path), "edited = true\n")
}

func TestEditUsesConfigFileEnv(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "name = \"demo\"\n")
	logPath := filepath.Join(t.TempDir(), "editor.log")
	t.Setenv("CONFIG_FILE", path)
	t.Setenv("EDITOR", writeEditorScript(t))
	t.Setenv("CONFIG_EDIT_LOG", logPath)

	err := Execute([]string{"edit"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if readFile(t, logPath) != path+"\n" {
		t.Fatalf("editor log = %q", readFile(t, logPath))
	}
	assertContains(t, readFile(t, path), "edited = true\n")
}

func TestEditFailsWhenConfigFileIsNotSpecified(t *testing.T) {
	clearConfigFileEnv(t)
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"edit"}, "1.2.3", &stdout, &stderr)

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

func TestEditRejectsExtraArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "name = \"demo\"\n")

	err := Execute([]string{"edit", "-f", path, "extra"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "usage: config edit [options]" {
		t.Fatalf("unexpected error: %v", err)
	}
}
