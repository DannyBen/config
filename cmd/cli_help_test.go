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
	assertContains(t, stdout.String(), "use         Use a config file in a sub-shell")
	assertContains(t, stdout.String(), "edit        Open the config file in an editor")
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
	assertContains(t, stdout.String(), "Usage:\n  config help [COMMAND|TOPIC]")
	assertContains(t, stdout.String(), "Commands:\n  set\n  get\n  unset\n  delete\n  array\n  array set\n  array add\n  array delete\n  list\n  dump\n  use\n  edit\n  completion")
	assertContains(t, stdout.String(), "Other topics:\n  environment\n  formats")
	if strings.Contains(stdout.String(), "Shortcut:") {
		t.Fatalf("help index should not include shortcut text:\n%s", stdout.String())
	}
}

func TestHelpCommandHelpShowsTopicIndex(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"help", "--help"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "Usage:\n  config help [COMMAND|TOPIC]")
	assertContains(t, stdout.String(), "Commands:\n  set\n  get\n  unset\n  delete\n  array\n  array set\n  array add\n  array delete\n  list\n  dump\n  use\n  edit\n  completion")
	assertContains(t, stdout.String(), "Other topics:\n  environment\n  formats")
}

func TestHelpCommandShowsCommandHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"help", "set"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "Create or update config values")
	assertContains(t, stdout.String(), "config set KEY VALUE [options]")
}

func TestHelpCommandShowsNestedCommandHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"help", "array", "add"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	assertContains(t, stdout.String(), "Add values to a scalar array")
	assertContains(t, stdout.String(), "config array add KEY VALUE... [options]")
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
	assertContains(t, stdout.String(), "config use ~/.codex/config.toml")
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
	assertContains(t, stdout.String(), "Unknown extensions")
	assertContains(t, stdout.String(), "# format: ini")
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
	assertContains(t, stdout.String(), "Usage:\n  config set KEY VALUE [options]")
	assertContains(t, stdout.String(), "--file, -f PATH\n    Path to the config file")
	assertContains(t, stdout.String(), "--in COLLECTION\n    Edit a record in COLLECTION")
	assertContains(t, stdout.String(), "--on FIELD:VALUE\n    Select or create a record by FIELD:VALUE")
	assertContains(t, stdout.String(), "--string, -s\n    Store VALUE as a string")
	assertContains(t, stdout.String(), "--dry, -n\n    Print the updated config without modifying the file")
	assertContains(t, stdout.String(), "--diff, -d\n    Print a unified diff without modifying the file")
	assertContains(t, stdout.String(), "--color, -c\n    Colorize diff output")
	assertContains(t, stdout.String(), "export CONFIG_FILE=config.toml")
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
	assertContains(t, stdout.String(), "Usage:\n  config delete KEY [options]")
	assertContains(t, stdout.String(), "Aliases:\n  del")
	assertContains(t, stdout.String(), "config del servers.1")
	if strings.Contains(stdout.String(), "config del KEY [options]") {
		t.Fatalf("delete usage should not include aliases:\n%s", stdout.String())
	}
	assertContains(t, stdout.String(), "--file, -f PATH\n    Path to the config file")
	assertContains(t, stdout.String(), "KEY\n    Dot notation string describing the intended config container")
	assertContains(t, stdout.String(), "--on FIELD:VALUE\n    Select a record by FIELD:VALUE. May be repeated.")
	assertContains(t, stdout.String(), "--if-empty\n    Only delete when the container has no values")
	assertContains(t, stdout.String(), "--if-exists\n    Do nothing when KEY is not set")
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
	assertContains(t, stdout.String(), "Usage:\n  config unset KEY [options]")
	assertContains(t, stdout.String(), "--file, -f PATH\n    Path to the config file")
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
	assertContains(t, stdout.String(), "Commands:\n  set      Replace a scalar array\n  add      Add values to a scalar array\n  delete   Remove values from a scalar array")
	if strings.Contains(stdout.String(), "Examples:") {
		t.Fatalf("array group help should not include examples:\n%s", stdout.String())
	}
	if strings.Contains(stdout.String(), "config array add KEY VALUE") {
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
	assertContains(t, stdout.String(), "Usage:\n  config array add KEY VALUE... [options]\n  config array add --help | -h")
	assertContains(t, stdout.String(), "Creates the array when KEY is not set.")
	assertContains(t, stdout.String(), "--file, -f PATH\n    Path to the config file")
	assertContains(t, stdout.String(), "--dry, -n\n    Print the updated config without modifying the file")
}

func TestArrayDeleteSubcommandHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"array", "delete", "--help"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	assertContains(t, stdout.String(), "Usage:\n  config array delete KEY VALUE... [options]")
	assertContains(t, stdout.String(), "Aliases:\n  del")
	assertContains(t, stdout.String(), "Deletes KEY when no values remain.")
	assertContains(t, stdout.String(), "config array del roots /tmp /var/tmp")
	if strings.Contains(stdout.String(), "config array del KEY VALUE... [options]") {
		t.Fatalf("array delete usage should not include aliases:\n%s", stdout.String())
	}
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
	assertContains(t, stdout.String(), "Usage:\n  config get KEY [options]")
	assertContains(t, stdout.String(), "Scalar values and arrays of scalar values are returned.")
	assertContains(t, stdout.String(), "Values are printed in a format-neutral display form")
	assertContains(t, stdout.String(), "--file, -f PATH\n    Path to the config file")
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
	assertContains(t, stdout.String(), "Usage:\n  config dump [KEY] [options]")
	assertContains(t, stdout.String(), "KEY\n    Optional key or table path to dump")
	assertContains(t, stdout.String(), "--file, -f PATH\n    Path to the config file")
	assertContains(t, stdout.String(), "--json\n    Dump as JSON instead of YAML")
}

func TestListHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"list", "--help"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	assertContains(t, stdout.String(), "Show config values")
	assertContains(t, stdout.String(), "Usage:\n  config list [KEY]")
	assertContains(t, stdout.String(), "Aliases:\n  ls")
	assertContains(t, stdout.String(), "--file, -f PATH\n    Path to the config file")
	assertContains(t, stdout.String(), "--color, -c\n    Colorize keys and separators")
	assertContains(t, stdout.String(), "config list database.port --color")
	if strings.Contains(stdout.String(), "config ls [KEY]") {
		t.Fatalf("list usage should not include aliases:\n%s", stdout.String())
	}
}

func TestEditHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"edit", "--help"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	assertContains(t, stdout.String(), "Open the config file in an editor")
	assertContains(t, stdout.String(), "Usage:\n  config edit [options]")
	assertContains(t, stdout.String(), "--file, -f PATH\n    Path to the config file")
	assertContains(t, stdout.String(), "EDITOR\n    Editor command to run. Defaults to vi.")
}

func TestUseHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Execute([]string{"use", "--help"}, "1.2.3", &stdout, &stderr)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
	assertContains(t, stdout.String(), "Use a config file in a child shell")
	assertContains(t, stdout.String(), "Usage:\n  config use FILE")
	assertContains(t, stdout.String(), "Starts a child shell with CONFIG_FILE set")
	assertContains(t, stdout.String(), "SHELL\n    Shell executable to start")
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
