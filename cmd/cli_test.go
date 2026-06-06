package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"-h"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	assertContains(t, stdout.String(), "Manipulate config files\n\nUsage:")
	assertContains(t, stdout.String(), "config COMMAND [options]")
	assertContains(t, stdout.String(), "set     Create or update config values")
	assertContains(t, stdout.String(), "delete  Delete a config container")
	if strings.Contains(stdout.String(), "Topics:") {
		t.Fatalf("root help should not list topics:\n%s", stdout.String())
	}
}

func TestHelpCommandShowsTopicIndex(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"help"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "Usage:\n  config help [TOPIC]")
	assertContains(t, stdout.String(), "Commands:\n  set\n  get\n  unset\n  delete\n  array\n  list")
	assertContains(t, stdout.String(), "Other topics:\n  environment")
	assertContains(t, stdout.String(), "Shortcut:\n  config COMMAND --help|-h")
}

func TestHelpCommandShowsCommandHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"help", "set"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "Create or update config values")
	assertContains(t, stdout.String(), "config set [CONFIG_FILE] KEY VALUE... [options]")
}

func TestHelpCommandShowsTopicHelp(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"help", "environment"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "Topic: environment")
	assertContains(t, stdout.String(), "CONFIG_FILE")
	assertContains(t, stdout.String(), "export CONFIG_FILE=~/.codex/config.toml")
	assertContains(t, stdout.String(), "CONFIG_LOG_LEVEL")
}

func TestHelpCommandRejectsUnknownTopic(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"help", "unknown"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	assertContains(t, err.Error(), `unknown help topic "unknown"`)
}

func TestSetHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"set", "--help"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	assertContains(t, stdout.String(), "Usage:\n  config set [CONFIG_FILE] KEY VALUE... [options]")
	assertContains(t, stdout.String(), "CONFIG_FILE\n    Path to the config file")
	assertContains(t, stdout.String(), "--in COLLECTION\n    Edit a record in COLLECTION")
	assertContains(t, stdout.String(), "--on FIELD:VALUE\n    Select or create a record by FIELD:VALUE")
	assertContains(t, stdout.String(), "--array, -a\n    Store VALUE as an array")
	assertContains(t, stdout.String(), "--string, -s\n    Store VALUE as a string")
	assertContains(t, stdout.String(), "--dry, -n\n    Print the updated config without modifying the file")
	assertContains(t, stdout.String(), "--diff, -d\n    Print a unified diff without modifying the file")
	assertContains(t, stdout.String(), "--color, -c\n    Colorize diff output")
	assertContains(t, stdout.String(), "export CONFIG_FILE=~/.codex/config.toml")
}

func TestDeleteHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"delete", "--help"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	assertContains(t, stdout.String(), "Usage:\n  config delete [CONFIG_FILE] KEY [options]")
	assertContains(t, stdout.String(), "CONFIG_FILE\n    Path to the config file")
	assertContains(t, stdout.String(), "KEY\n    Dot notation string describing the intended config container")
	assertContains(t, stdout.String(), "--on FIELD:VALUE\n    Select a record by FIELD:VALUE. May be repeated.")
	assertContains(t, stdout.String(), "--dry, -n\n    Print the updated config without modifying the file")
	assertContains(t, stdout.String(), "--diff, -d\n    Print a unified diff without modifying the file")
	assertContains(t, stdout.String(), "--color, -c\n    Colorize diff output")
}

func TestUnsetHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"unset", "--help"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	assertContains(t, stdout.String(), "Usage:\n  config unset [CONFIG_FILE] KEY [options]")
	assertContains(t, stdout.String(), "--in COLLECTION\n    Remove a field from a record in COLLECTION")
	assertContains(t, stdout.String(), "--on FIELD:VALUE\n    Select a record by FIELD:VALUE. May be repeated.")
	assertContains(t, stdout.String(), "--if VALUE\n    Only unset when the current value matches VALUE")
	assertContains(t, stdout.String(), "--if-exists\n    Do nothing when KEY is not set")
	assertContains(t, stdout.String(), "--dry, -n\n    Print the updated config without modifying the file")
	assertContains(t, stdout.String(), "--diff, -d\n    Print a unified diff without modifying the file")
	assertContains(t, stdout.String(), "--color, -c\n    Colorize diff output")
}

func TestGetHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"get", "--help"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	assertContains(t, stdout.String(), "Usage:\n  config get [CONFIG_FILE] KEY [options]")
	assertContains(t, stdout.String(), "Scalar values and arrays of scalar values are returned.")
	assertContains(t, stdout.String(), "Values are printed in a format-neutral display form")
	assertContains(t, stdout.String(), "--in COLLECTION\n    Read a field from a record in COLLECTION")
	assertContains(t, stdout.String(), "--on FIELD:VALUE\n    Select a record by FIELD:VALUE. May be repeated.")
}

func TestSetEchoesParsedArgsWithFlagAfterPositionals(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "server.port = 2000\n")

	err := Execute([]string{"set", path, "server.port", "3000", "--dry"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	if stdout.String() != "server.port = 3000\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if got := readFile(t, path); got != "server.port = 2000\n" {
		t.Fatalf("file changed during dry run: %q", got)
	}
}

func TestSetEchoesParsedArgsWithFlagBeforePositionals(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "server.port = 2000\n")

	err := Execute([]string{"set", "--dry", path, "server.port", "3000"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.String() != "server.port = 3000\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestSetEchoesParsedArgsWithShortFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "server.port = 2000\n")

	err := Execute([]string{"set", "-n", path, "server.port", "3000"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.String() != "server.port = 3000\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestSetEchoesParsedArgsWithDiffFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "server.port = 2000\n")

	err := Execute([]string{"set", "-d", path, "server.port", "3000"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "--- "+path+"\n+++ "+path+"\n")
	assertContains(t, stdout.String(), "-server.port = 2000\n")
	assertContains(t, stdout.String(), "+server.port = 3000\n")
	if got := readFile(t, path); got != "server.port = 2000\n" {
		t.Fatalf("file changed during diff: %q", got)
	}
}

func TestSetColorizesDiff(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "server.port = 2000\n")

	err := Execute([]string{"set", "--diff", "--color", path, "server.port", "3000"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "\x1b[31m-server.port = 2000\n\x1b[0m")
	assertContains(t, stdout.String(), "\x1b[32m+server.port = 3000\n\x1b[0m")
}

func TestSetCompactsDiffAndColorShortFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "server.port = 2000\n")

	err := Execute([]string{"set", "-dc", path, "server.port", "3000"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "\x1b[31m-server.port = 2000\n\x1b[0m")
	assertContains(t, stdout.String(), "\x1b[32m+server.port = 3000\n\x1b[0m")
}

func TestUnifiedDiffShowsContextHunks(t *testing.T) {
	before := strings.Join([]string{
		"line 1",
		"line 2",
		"line 3",
		"line 4",
		"line 5",
		"line 6",
		"line 7",
		"line 8",
		"line 9",
		"line 10",
		"",
	}, "\n")
	after := strings.Replace(before, "line 6\n", "line six\n", 1)

	got := unifiedDiff("config.toml", before, after)

	assertContains(t, got, "--- config.toml\n+++ config.toml\n@@ -3,7 +3,7 @@\n")
	assertContains(t, got, " line 3\n")
	assertContains(t, got, "-line 6\n")
	assertContains(t, got, "+line six\n")
	assertContains(t, got, " line 9\n")
	if strings.Contains(got, " line 1\n") || strings.Contains(got, " line 10\n") {
		t.Fatalf("diff includes context outside the hunk:\n%s", got)
	}
}

func TestUnifiedDiffSeparatesLinesWithoutTrailingNewline(t *testing.T) {
	before := "server: { host: localhost, port: 3000 }"
	after := "server: { host: example, port: 3000 }"

	got := unifiedDiff("config.yaml", before, after)

	assertContains(t, got, "--- config.yaml\n+++ config.yaml\n@@ -1 +1 @@\n")
	assertContains(t, got, "-server: { host: localhost, port: 3000 }\n")
	assertContains(t, got, "+server: { host: example, port: 3000 }\n")
	if strings.Contains(got, "}+server") {
		t.Fatalf("diff joined removed and added lines:\n%s", got)
	}
}

func TestUnifiedDiffKeepsLargeInputsCompact(t *testing.T) {
	var beforeLines []string
	var afterLines []string
	for i := 1; i <= 1500; i++ {
		beforeLines = append(beforeLines, fmt.Sprintf("line %d", i))
		if i == 750 {
			afterLines = append(afterLines, "line seven fifty")
		} else {
			afterLines = append(afterLines, fmt.Sprintf("line %d", i))
		}
	}
	before := strings.Join(beforeLines, "\n") + "\n"
	after := strings.Join(afterLines, "\n") + "\n"

	got := unifiedDiff("config.yaml", before, after)

	assertContains(t, got, "--- config.yaml\n+++ config.yaml\n@@ -747,7 +747,7 @@\n")
	assertContains(t, got, "-line 750\n")
	assertContains(t, got, "+line seven fifty\n")
	if strings.Contains(got, " line 1\n") || strings.Contains(got, " line 1500\n") {
		t.Fatalf("large diff includes full-file context:\n%s", got)
	}
}

func TestSetWritesUpdatedTOML(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "server.port = 2000\n")

	err := Execute([]string{"set", path, "server.port", "3000"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if got := readFile(t, path); got != "server.port = 3000\n" {
		t.Fatalf("file mismatch: %q", got)
	}
}

func TestSetWritesArrayFromMultipleValues(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[server]\nports = [1000]\n")

	err := Execute([]string{"set", path, "server.ports", "3000", "3001"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "[server]\nports = [3000, 3001]\n"
	if got := readFile(t, path); got != want {
		t.Fatalf("file mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetWritesSingleValueArrayWithArrayFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[server]\nports = [1000]\n")

	err := Execute([]string{"set", "--array", path, "server.ports", "3000"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "[server]\nports = [3000]\n"
	if got := readFile(t, path); got != want {
		t.Fatalf("file mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetWritesTOMLLiteralValue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "release.date = \"old\"\n")

	err := Execute([]string{"set", path, "release.date", "2027-03-24"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "release.date = 2027-03-24\n"
	if got := readFile(t, path); got != want {
		t.Fatalf("file mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetIdenticalValueDoesNotWriteFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "server.port = 3000\n")
	if err := os.Chmod(path, 0444); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(path, 0644)

	err := Execute([]string{"set", path, "server.port", "3000"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if got := readFile(t, path); got != "server.port = 3000\n" {
		t.Fatalf("file mismatch: %q", got)
	}
}

func TestSetIdenticalValueDiffPrintsNothing(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "server.port = 3000\n")

	err := Execute([]string{"set", "--diff", path, "server.port", "3000"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if got := readFile(t, path); got != "server.port = 3000\n" {
		t.Fatalf("file mismatch: %q", got)
	}
}

func TestSetIdenticalValueDryPrintsUnchangedConfig(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "server.port = 3000\n")

	err := Execute([]string{"set", "--dry", path, "server.port", "3000"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.String() != "server.port = 3000\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if got := readFile(t, path); got != "server.port = 3000\n" {
		t.Fatalf("file mismatch: %q", got)
	}
}

func TestSetStringFlagForcesStringValue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "version = 1.0\n")

	err := Execute([]string{"set", "--string", path, "version", "1.0"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "version = \"1.0\"\n"
	if got := readFile(t, path); got != want {
		t.Fatalf("file mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetStringFlagRejectsMultipleValues(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "values = []\n")

	err := Execute([]string{"set", "--string", path, "values", "one", "two"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "flag --string cannot be used with multiple values" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetReadsValueFromStdin(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "message = \"short\"\n")

	err := ExecuteWithIO([]string{"set", path, "message", "-"}, "1.2.3", strings.NewReader("hello\nworld"), &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	want := "message = \"\"\"hello\nworld\"\"\"\n"
	if got := readFile(t, path); got != want {
		t.Fatalf("file mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetReadsSingleArrayValueFromStdin(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "ports = [1000]\n")

	err := ExecuteWithIO([]string{"set", "--array", path, "ports", "-"}, "1.2.3", strings.NewReader("3000"), &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "ports = [3000]\n"
	if got := readFile(t, path); got != want {
		t.Fatalf("file mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetDryAndDiffAreMutuallyExclusive(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "server.port = 2000\n")

	err := Execute([]string{"set", "--dry", "--diff", path, "server.port", "3000"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	assertContains(t, err.Error(), `if any flags in the group [dry diff] are set none of the others can be`)
}

func TestSetUsesConfigFileEnv(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "server.port = 2000\n")
	t.Setenv("CONFIG_FILE", path)

	err := Execute([]string{"set", "server.port", "3000"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got := readFile(t, path); got != "server.port = 3000\n" {
		t.Fatalf("file mismatch: %q", got)
	}
}

func TestSetInOnWritesArrayRecord(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[[servers]]\nname = \"api\"\nport = 3000\n")

	err := Execute([]string{"set", path, "port", "8080", "--in", "servers", "--on", "name:api"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "[[servers]]\nname = \"api\"\nport = 8080\n"
	if got := readFile(t, path); got != want {
		t.Fatalf("file mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetInRequiresOnFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[[servers]]\nname = \"api\"\nport = 3000\n")

	err := Execute([]string{"set", path, "port", "8080", "--in", "servers"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "flag --in requires --on" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetColorRequiresDiff(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "server.port = 2000\n")

	err := Execute([]string{"set", "--color", path, "server.port", "3000"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	assertContains(t, err.Error(), "flag --color requires --diff")
}

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
	if !strings.Contains(err.Error(), "usage: config set [CONFIG_FILE] KEY VALUE... [options]") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetPrintsTOMLValue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[database]\nport = 5432\n")

	err := Execute([]string{"get", path, "database.port"}, "1.2.3", &stdout, &stderr)

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

	err := Execute([]string{"get", path, "port", "--in", "servers", "--on", "name:worker"}, "1.2.3", &stdout, &stderr)

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

	err := Execute([]string{"get", path, "port", "--on", "name:api"}, "1.2.3", &stdout, &stderr)

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

	err := Execute([]string{"get", path, "port", "--in", "servers"}, "1.2.3", &stdout, &stderr)

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
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"get", "config.json", "database.port"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if err.Error() != "unsupported config format for config.json" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExplicitTargetPathOverridesConfigFileEnv(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[database]\nport = 5432\n")
	t.Setenv("CONFIG_FILE", path)
	yamlPath := writeTempYAML(t, "database:\n  port: 5432\n")

	err := Execute([]string{"get", yamlPath, "database.port"}, "1.2.3", &stdout, &stderr)

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

func TestGetMissingRequiredArgFailsAfterConfigFileIsResolved(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[database]\nport = 5432\n")

	err := Execute([]string{"get", path}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if !strings.Contains(err.Error(), "usage: config get [CONFIG_FILE] KEY") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListPrintsTOMLValues(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "title = \"demo app\"\n\n[server]\nport = 3000\nenabled = true\n")

	err := Execute([]string{"list", path}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	want := "title=demo app\nserver.port=3000\nserver.enabled=true\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestListPrintsTOMLValuesUnderTable(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "title = \"demo app\"\n\n[server]\nport = 3000\nenabled = true\n")

	err := Execute([]string{"list", path, "server"}, "1.2.3", &stdout, &stderr)

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

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected output to contain %q, got:\n%s", needle, haystack)
	}
}

func clearConfigFileEnv(t *testing.T) {
	t.Helper()
	t.Setenv("CONFIG_FILE", "")
}

func writeTempTOML(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
