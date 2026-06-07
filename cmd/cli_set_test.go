package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

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

func TestSetWritesUpdatedJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempJSON(t, `{"title":"demo app"}`)

	err := Execute([]string{"set", path, "server.port", "3000"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	want := "{\n  \"server\": {\n    \"port\": 3000\n  },\n  \"title\": \"demo app\"\n}\n"
	if got := readFile(t, path); got != want {
		t.Fatalf("file mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestArrayAddWritesUpdatedJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempJSON(t, `{"roots":["$HOME/.cache"]}`)

	err := Execute([]string{"array", "add", path, "roots", "/tmp", "$HOME/.cache"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	want := "{\n  \"roots\": [\n    \"$HOME/.cache\",\n    \"/tmp\"\n  ]\n}\n"
	if got := readFile(t, path); got != want {
		t.Fatalf("file mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetRejectsMultipleValues(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := writeTempTOML(t, "[server]\nports = [1000]\n")

	err := Execute([]string{"set", path, "server.ports", "3000", "3001"}, "1.2.3", &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "usage: config set [CONFIG_FILE] KEY VALUE [options]") {
		t.Fatalf("unexpected error: %v", err)
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
