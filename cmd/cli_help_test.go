package cmd

import (
	"bytes"
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
	assertContains(t, stdout.String(), "set         Create or update config values")
	assertContains(t, stdout.String(), "delete      Delete a config container")
	assertContains(t, stdout.String(), "dump        Dump config data")
	assertContains(t, stdout.String(), "completion  Generate shell completion scripts")
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
	assertContains(t, stdout.String(), "Commands:\n  set\n  get\n  unset\n  delete\n  array\n  list\n  dump\n  completion")
	assertContains(t, stdout.String(), "Other topics:\n  environment\n  formats")
	if strings.Contains(stdout.String(), "Shortcut:") {
		t.Fatalf("help index should not include shortcut text:\n%s", stdout.String())
	}
}

func TestHelpCommandShowsCommandHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"help", "set"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "Create or update config values")
	assertContains(t, stdout.String(), "config set [CONFIG_FILE] KEY VALUE [options]")
}

func TestHelpCommandShowsNestedCommandHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"help", "array", "add"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "Add values to a scalar array")
	assertContains(t, stdout.String(), "config array add [CONFIG_FILE] KEY VALUE... [options]")
	assertContains(t, stdout.String(), "Creates the array when KEY is not set.")
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

func TestHelpCommandShowsFormatsTopic(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"help", "formats"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "Topic: formats")
	assertContains(t, stdout.String(), "TOML")
	assertContains(t, stdout.String(), ".json")
	assertContains(t, stdout.String(), "canonical pretty JSON")
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
	assertContains(t, stdout.String(), "Usage:\n  config set [CONFIG_FILE] KEY VALUE [options]")
	assertContains(t, stdout.String(), "CONFIG_FILE\n    Path to the config file")
	assertContains(t, stdout.String(), "--in COLLECTION\n    Edit a record in COLLECTION")
	assertContains(t, stdout.String(), "--on FIELD:VALUE\n    Select or create a record by FIELD:VALUE")
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
	assertContains(t, stdout.String(), "--if-empty\n    Only delete when the container has no values")
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

func TestArrayHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"array", "--help"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	assertContains(t, stdout.String(), "Usage:\n  config array COMMAND [options]\n  config array --help | -h")
	assertContains(t, stdout.String(), "Commands:\n  set   Replace a scalar array\n  add   Add values to a scalar array\n  del   Remove values from a scalar array")
	if strings.Contains(stdout.String(), "Examples:") {
		t.Fatalf("array group help should not include examples:\n%s", stdout.String())
	}
	if strings.Contains(stdout.String(), "config array add [CONFIG_FILE] KEY VALUE") {
		t.Fatalf("array group help should not include subcommand details:\n%s", stdout.String())
	}
}

func TestArraySubcommandHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"array", "add", "--help"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	assertContains(t, stdout.String(), "Usage:\n  config array add [CONFIG_FILE] KEY VALUE... [options]\n  config array add --help | -h")
	assertContains(t, stdout.String(), "Creates the array when KEY is not set.")
	assertContains(t, stdout.String(), "--dry, -n\n    Print the updated config without modifying the file")
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

func TestDumpHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"dump", "--help"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	assertContains(t, stdout.String(), "Dump config data")
	assertContains(t, stdout.String(), "Usage:\n  config dump [CONFIG_FILE] [KEY] [options]")
	assertContains(t, stdout.String(), "KEY\n    Optional key or table path to dump")
	assertContains(t, stdout.String(), "--json\n    Dump as JSON instead of YAML")
}

func TestCompletionHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"completion", "--help"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	assertContains(t, stdout.String(), "Generate shell completion scripts")
	assertContains(t, stdout.String(), "Usage:\n  config completion SHELL")
	assertContains(t, stdout.String(), "bash, zsh, fish")
}
