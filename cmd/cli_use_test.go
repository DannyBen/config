package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUseStartsShellWithConfigFileEnv(t *testing.T) {
	var stdout, stderr bytes.Buffer
	configFile := writeTempTOML(t, "app = { name = \"demo\" }\n")
	shell := os.Args[0]
	t.Setenv("SHELL", shell)
	t.Setenv("CONFIG_FILE", "previous.toml")

	var gotShell string
	var gotEnv []string
	oldRunShell := runShell
	runShell = func(shell string, env []string, stdin io.Reader, stdout, stderr io.Writer) error {
		gotShell = shell
		gotEnv = env
		return nil
	}
	t.Cleanup(func() {
		runShell = oldRunShell
	})

	err := Execute([]string{"use", configFile}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if gotShell != shell {
		t.Fatalf("shell = %q, want %q", gotShell, shell)
	}
	assertEnvValue(t, gotEnv, "CONFIG_FILE", configFile)
	if os.Getenv("CONFIG_FILE") != "previous.toml" {
		t.Fatalf("parent CONFIG_FILE changed to %q", os.Getenv("CONFIG_FILE"))
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	assertContains(t, stderr.String(), "Using "+configFile+".\nExit the shell to stop.")
}

func TestUseResolvesRelativeConfigFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	tmpdir := t.TempDir()
	configFile := filepath.Join(tmpdir, "config.toml")
	if err := os.WriteFile(configFile, []byte("app = { name = \"demo\" }\n"), 0644); err != nil {
		t.Fatal(err)
	}
	shell := os.Args[0]
	t.Setenv("SHELL", shell)

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatal(err)
		}
	})

	var gotEnv []string
	oldRunShell := runShell
	runShell = func(shell string, env []string, stdin io.Reader, stdout, stderr io.Writer) error {
		gotEnv = env
		return nil
	}
	t.Cleanup(func() {
		runShell = oldRunShell
	})

	err = Execute([]string{"use", "config.toml"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertEnvValue(t, gotEnv, "CONFIG_FILE", configFile)
}

func TestUseFallsBackToBinShWhenShellIsUnavailable(t *testing.T) {
	var stdout, stderr bytes.Buffer
	configFile := writeTempTOML(t, "app = { name = \"demo\" }\n")
	t.Setenv("SHELL", "/no/such/shell")

	var gotShell string
	oldRunShell := runShell
	runShell = func(shell string, env []string, stdin io.Reader, stdout, stderr io.Writer) error {
		gotShell = shell
		return nil
	}
	t.Cleanup(func() {
		runShell = oldRunShell
	})

	err := Execute([]string{"use", configFile}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if gotShell != "/bin/sh" {
		t.Fatalf("shell = %q, want /bin/sh", gotShell)
	}
}

func TestUseRejectsDirectory(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"use", t.TempDir()}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	assertContains(t, err.Error(), "is a directory")
}

func TestUseReturnsShellExitCodeWithoutPrintingError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	configFile := writeTempTOML(t, "app = { name = \"demo\" }\n")
	t.Setenv("SHELL", os.Args[0])

	oldRunShell := runShell
	runShell = func(shell string, env []string, stdin io.Reader, stdout, stderr io.Writer) error {
		return silentExitError{code: 42}
	}
	t.Cleanup(func() {
		runShell = oldRunShell
	})

	err := Execute([]string{"use", configFile}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if ExitCode(err) != 42 {
		t.Fatalf("ExitCode = %d, want 42", ExitCode(err))
	}
	var printed bytes.Buffer
	PrintError(err, &printed)
	if printed.Len() != 0 {
		t.Fatalf("PrintError output = %q", printed.String())
	}
}

func TestUseReturnsShellStartError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	configFile := writeTempTOML(t, "app = { name = \"demo\" }\n")
	t.Setenv("SHELL", os.Args[0])
	wantErr := errors.New("start failed")

	oldRunShell := runShell
	runShell = func(shell string, env []string, stdin io.Reader, stdout, stderr io.Writer) error {
		return wantErr
	}
	t.Cleanup(func() {
		runShell = oldRunShell
	})

	err := Execute([]string{"use", configFile}, "1.2.3", &stdout, &stderr)

	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

func assertEnvValue(t *testing.T, env []string, key, want string) {
	t.Helper()
	prefix := key + "="
	var matches []string
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			matches = append(matches, strings.TrimPrefix(item, prefix))
		}
	}
	if len(matches) != 1 {
		t.Fatalf("%s entries = %v, want one value %q", key, matches, want)
	}
	if matches[0] != want {
		t.Fatalf("%s = %q, want %q", key, matches[0], want)
	}
}
