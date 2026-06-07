package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestUnsetWritesUpdatedTOML(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[database]\nhost = \"localhost\"\nport = 5432\n")

	err := Execute([]string{"unset", path, "database.port"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if got := readFile(t, path); got != "[database]\nhost = \"localhost\"\n" {
		t.Fatalf("file mismatch: %q", got)
	}
}

func TestUnsetWritesUpdatedJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempJSON(t, `{"database":{"host":"localhost","port":5432}}`)

	err := Execute([]string{"unset", path, "database.port"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	want := "{\n  \"database\": {\n    \"host\": \"localhost\"\n  }\n}\n"
	if got := readFile(t, path); got != want {
		t.Fatalf("file mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestUnsetUsesConfigFileEnv(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[database]\nhost = \"localhost\"\nport = 5432\n")
	t.Setenv("CONFIG_FILE", path)

	err := Execute([]string{"unset", "database.port"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got := readFile(t, path); got != "[database]\nhost = \"localhost\"\n" {
		t.Fatalf("file mismatch: %q", got)
	}
}

func TestUnsetFailsWhenConfigFileIsNotSpecified(t *testing.T) {
	clearConfigFileEnv(t)
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"unset", "database.port"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "config file not specified" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnsetDryPrintsUpdatedTOML(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[database]\nhost = \"localhost\"\nport = 5432\n")

	err := Execute([]string{"unset", "--dry", path, "database.port"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.String() != "[database]\nhost = \"localhost\"\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if got := readFile(t, path); got != "[database]\nhost = \"localhost\"\nport = 5432\n" {
		t.Fatalf("file changed during dry run: %q", got)
	}
}

func TestUnsetDiffPrintsUnifiedDiff(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[database]\nhost = \"localhost\"\nport = 5432\n")

	err := Execute([]string{"unset", "--diff", path, "database.port"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "--- "+path+"\n+++ "+path+"\n")
	assertContains(t, stdout.String(), "-port = 5432\n")
	if got := readFile(t, path); got != "[database]\nhost = \"localhost\"\nport = 5432\n" {
		t.Fatalf("file changed during diff: %q", got)
	}
}

func TestUnsetIfValueOnlyUnsetsMatchingScalar(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "submit = \"tab\"\nqueue = \"alt-w\"\n")

	err := Execute([]string{"unset", path, "submit", "--if", "tab"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	err = Execute([]string{"unset", path, "queue", "--if", "alt-q"}, "1.2.3", &stdout, &stderr)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	err = Execute([]string{"unset", path, "missing", "--if", "tab"}, "1.2.3", &stdout, &stderr)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if got := readFile(t, path); got != "queue = \"alt-w\"\n" {
		t.Fatalf("file mismatch: %q", got)
	}
}

func TestUnsetIfExistsIgnoresMissingKey(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "submit = \"tab\"\nqueue = \"alt-q\"\n")

	err := Execute([]string{"unset", path, "submit", "--if-exists"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	err = Execute([]string{"unset", path, "missing", "--if-exists"}, "1.2.3", &stdout, &stderr)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if got := readFile(t, path); got != "queue = \"alt-q\"\n" {
		t.Fatalf("file mismatch: %q", got)
	}
}

func TestUnsetSelectedTOMLRecordValue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[[servers]]\nname = \"api\"\nport = 3000\n\n[[servers]]\nname = \"worker\"\nport = 3001\n")

	err := Execute([]string{"unset", "--dry", path, "port", "--in", "servers", "--on", "name:worker"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "[[servers]]\nname = \"api\"\nport = 3000\n\n[[servers]]\nname = \"worker\"\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestUnsetSelectedRecordRequiresInWithOn(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[[servers]]\nname = \"api\"\nport = 3000\n")

	err := Execute([]string{"unset", path, "port", "--on", "name:api"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "flag --on requires --in" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnsetSelectedRecordRequiresOnWithIn(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[[servers]]\nname = \"api\"\nport = 3000\n")

	err := Execute([]string{"unset", path, "port", "--in", "servers"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "flag --in requires --on" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteFailsWhenConfigFileIsNotSpecified(t *testing.T) {
	clearConfigFileEnv(t)
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"delete", "database"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "config file not specified" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteWritesUpdatedJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempJSON(t, `{"title":"demo","style":{"color":"blue"},"server":{"port":3000}}`)

	err := Execute([]string{"delete", path, "style"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	want := "{\n  \"server\": {\n    \"port\": 3000\n  },\n  \"title\": \"demo\"\n}\n"
	if got := readFile(t, path); got != want {
		t.Fatalf("file mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestDeleteAliasWritesUpdatedJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempJSON(t, `{"title":"demo","style":{"color":"blue"},"server":{"port":3000}}`)

	err := Execute([]string{"del", path, "style"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	want := "{\n  \"server\": {\n    \"port\": 3000\n  },\n  \"title\": \"demo\"\n}\n"
	if got := readFile(t, path); got != want {
		t.Fatalf("file mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestDeleteDryPrintsUpdatedTOML(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "title = \"demo\"\n\n[style]\ncolor = \"blue\"\nfont = \"arial\"\n")

	err := Execute([]string{"delete", "--dry", path, "style"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.String() != "title = \"demo\"\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if got := readFile(t, path); got != "title = \"demo\"\n\n[style]\ncolor = \"blue\"\nfont = \"arial\"\n" {
		t.Fatalf("file changed during dry run: %q", got)
	}
}

func TestDeleteDiffPrintsUnifiedDiff(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "title = \"demo\"\n\n[style]\ncolor = \"blue\"\nfont = \"arial\"\n")

	err := Execute([]string{"delete", "--diff", path, "style"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "--- "+path+"\n+++ "+path+"\n")
	assertContains(t, stdout.String(), "-[style]\n")
	assertContains(t, stdout.String(), "-color = \"blue\"\n")
	if got := readFile(t, path); got != "title = \"demo\"\n\n[style]\ncolor = \"blue\"\nfont = \"arial\"\n" {
		t.Fatalf("file changed during diff: %q", got)
	}
}

func TestDeleteIfEmptyNoOpsWhenContainerHasValues(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[style]\ncolor = \"blue\"\n")

	err := Execute([]string{"delete", path, "style", "--if-empty"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if got := readFile(t, path); got != "[style]\ncolor = \"blue\"\n" {
		t.Fatalf("file mismatch: %q", got)
	}
}

func TestDeleteRefusesScalarValue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[style]\ncolor = \"blue\"\nfont = \"arial\"\n")

	err := Execute([]string{"delete", path, "style.color"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	assertContains(t, err.Error(), "style.color is a value, use unset to remove fields")
}

func TestDeleteAcceptsRepeatedSelectors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[[servers]]\nname = \"api\"\nport = 3000\n\n[[servers]]\nname = \"api\"\nport = 3001\n")

	err := Execute([]string{"delete", "--dry", path, "servers", "--on", "name:api", "--on", "port:3000"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "[[servers]]\nname = \"api\"\nport = 3001\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestDeleteColorRequiresDiff(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[database]\nport = 5432\n")

	err := Execute([]string{"delete", "--color", path, "database"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	assertContains(t, err.Error(), "flag --color requires --diff")
}

func TestPrintErrorPrefixesOperationalErrors(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[database]\nport = 5432\n")

	err := Execute([]string{"unset", "-dc", path, "database"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	PrintError(err, &stderr)
	if stderr.String() != "ERROR database is a table, not a value\n" {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestPrintErrorReportsInvalidLogLevel(t *testing.T) {
	t.Setenv("CONFIG_LOG_LEVEL", "loud")
	t.Setenv("NO_COLOR", "1")
	var stderr bytes.Buffer

	PrintError(errors.New("failed"), &stderr)

	if stderr.String() != "error: CONFIG_LOG_LEVEL: unknown level \"loud\"\n" {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestSetFailsWhenConfigFileIsNotSpecified(t *testing.T) {
	clearConfigFileEnv(t)
	t.Setenv("NO_COLOR", "1")
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"set", "server.port", "3000"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if err.Error() != "config file not specified" {
		t.Fatalf("unexpected error: %v", err)
	}

	PrintError(err, &stderr)
	if stderr.String() != "ERROR config file not specified\n" {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestSetMissingRequiredArgFailsAfterConfigFileIsResolved(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "server.port = 2000\n")

	err := Execute([]string{"set", path, "server.port"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if !strings.Contains(err.Error(), "usage: config set [CONFIG_FILE] KEY VALUE [options]") {
		t.Fatalf("unexpected error: %v", err)
	}
}
