package cmd

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type featureCommand struct {
	text           string
	exitCode       int
	stdout         string
	stderr         string
	stdoutByFormat map[string]string
	stderrByFormat map[string]string
}

type featureSpec struct {
	name          string
	pending       bool
	pendingReason string
	sources       map[string]string
	files         map[string]string
	results       map[string]string
	commands      []featureCommand
}

func TestFeatures(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("CONFIG_LOG_LEVEL", "")
	t.Setenv("LOG_LEVEL", "")

	root := filepath.Join("..", "features")
	paths := filterFeatureSpecPaths(t, root, featureSpecPaths(t, root), flag.Args())

	for _, path := range paths {
		spec := parseFeatureSpec(t, path)
		name := featureTestName(t, root, path)
		t.Run(name, func(t *testing.T) {
			runFeatureSpec(t, spec)
		})
	}
}

func featureSpecPaths(t *testing.T, root string) []string {
	t.Helper()
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() == "README.md" || filepath.Ext(entry.Name()) != ".md" {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatalf("%s: no feature specs found", root)
	}
	return paths
}

func filterFeatureSpecPaths(t *testing.T, root string, paths, filters []string) []string {
	t.Helper()
	if len(filters) == 0 {
		return paths
	}
	var out []string
	for _, path := range paths {
		name := featureTestName(t, root, path)
		for _, filter := range filters {
			filter = strings.TrimSuffix(filepath.ToSlash(filter), ".md")
			if strings.Contains(name, filter) {
				out = append(out, path)
				break
			}
		}
	}
	if len(out) == 0 {
		t.Fatalf("no feature specs matched %q", strings.Join(filters, ", "))
	}
	return out
}

func featureTestName(t *testing.T, root, path string) string {
	t.Helper()
	rel, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSuffix(filepath.ToSlash(rel), filepath.Ext(rel))
}

func runFeatureSpec(t *testing.T, spec featureSpec) {
	t.Helper()
	if spec.pending {
		reason := spec.pendingReason
		if reason == "" {
			reason = "pending feature"
		}
		t.Skip(reason)
	}
	for formatName, source := range spec.sources {
		t.Run(formatName, func(t *testing.T) {
			runFeatureFormat(t, spec, formatName, source)
		})
	}
}

func runFeatureFormat(t *testing.T, spec featureSpec, formatName, source string) {
	t.Helper()
	temp, target, targetName := writeFeatureFiles(t, spec, formatName, source)

	var allStdout, allStderr bytes.Buffer
	for _, command := range spec.commands {
		gotStdout, gotStderr, collect := runFeatureCommand(t, command, temp, target, targetName, formatName)
		if collect {
			allStdout.WriteString(gotStdout)
			allStderr.WriteString(gotStderr)
		}
	}

	verifyFeatureResult(t, spec, formatName, source, target, targetName)
	if allStdout.Len() != 0 {
		t.Fatalf("unexpected stdout\n%s", unifiedDiff("stdout", "", allStdout.String()))
	}
	if allStderr.Len() != 0 {
		t.Fatalf("unexpected stderr\n%s", unifiedDiff("stderr", "", allStderr.String()))
	}
}

func writeFeatureFiles(t *testing.T, spec featureSpec, formatName, source string) (string, string, string) {
	t.Helper()
	targetName := "config." + formatName
	temp := t.TempDir()
	target := filepath.Join(temp, targetName)
	if err := os.WriteFile(target, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	for name, content := range spec.files {
		if err := os.WriteFile(filepath.Join(temp, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return temp, target, targetName
}

func runFeatureCommand(t *testing.T, command featureCommand, temp, target, targetName, formatName string) (string, string, bool) {
	t.Helper()
	commandText, stdin := featureCommandInput(t, command.text, temp)
	args, splitErr := splitCommand(commandText)
	if splitErr != nil {
		t.Fatalf("invalid command %q: %v", commandText, splitErr)
	}
	if len(args) == 0 {
		return "", "", false
	}
	if args[0] != "config" {
		t.Fatalf("command must start with config: %q", commandText)
	}

	var stdout, stderr bytes.Buffer
	t.Setenv("CONFIG_FILE", target)
	err := ExecuteWithIO(args[1:], "1.2.3", strings.NewReader(stdin), &stdout, &stderr)
	verifyFeatureCommandExit(t, command, commandText, err, &stderr)

	gotStdout := normalizeExampleOutput(stdout.String(), target, targetName)
	gotStderr := normalizeExampleOutput(stderr.String(), target, targetName)
	if command.hasOutputExpectations() {
		verifyFeatureCommandOutput(t, command, commandText, formatName, gotStdout, gotStderr)
		return "", "", false
	}
	return gotStdout, gotStderr, true
}

func featureCommandInput(t *testing.T, commandText, temp string) (string, string) {
	t.Helper()
	before, after, ok := strings.Cut(commandText, " < ")
	if !ok {
		return commandText, ""
	}
	return strings.TrimSpace(before), readFile(t, filepath.Join(temp, strings.TrimSpace(after)))
}

func verifyFeatureCommandExit(t *testing.T, command featureCommand, commandText string, err error, stderr *bytes.Buffer) {
	t.Helper()
	if command.exitCode != 0 {
		if err == nil {
			t.Fatalf("Execute(%q) returned nil error, want error", commandText)
		}
		PrintError(err, stderr)
		return
	}
	if err != nil {
		t.Fatalf("Execute(%q) returned error: %v\nstderr:\n%s", commandText, err, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr for %q:\n%s", commandText, stderr.String())
	}
}

func verifyFeatureCommandOutput(t *testing.T, command featureCommand, commandText, formatName, gotStdout, gotStderr string) {
	t.Helper()
	wantStdout := expectedFeatureOutput(command.stdout, command.stdoutByFormat, formatName)
	if gotStdout != wantStdout {
		t.Fatalf("stdout mismatch for %q\n%s", commandText, unifiedDiff("stdout", wantStdout, gotStdout))
	}
	wantStderr := expectedFeatureOutput(command.stderr, command.stderrByFormat, formatName)
	if gotStderr != wantStderr {
		t.Fatalf("stderr mismatch for %q\n%s", commandText, unifiedDiff("stderr", wantStderr, gotStderr))
	}
}

func verifyFeatureResult(t *testing.T, spec featureSpec, formatName, source, target, targetName string) {
	t.Helper()
	want := source
	if result, ok := spec.results[formatName]; ok {
		want = result
	}
	got := readFile(t, target)
	if normalizeFeatureFileResult(got) != normalizeFeatureFileResult(want) {
		t.Fatalf("%s result mismatch\n%s", formatName, unifiedDiff(targetName, want, got))
	}
}

func normalizeFeatureFileResult(value string) string {
	return strings.TrimRight(value, " \t\r\n")
}

func normalizeExampleOutput(output, target, beforeName string) string {
	return strings.ReplaceAll(output, target, beforeName)
}

func (c featureCommand) hasOutputExpectations() bool {
	return c.stdout != "" || c.stderr != "" || len(c.stdoutByFormat) != 0 || len(c.stderrByFormat) != 0
}

func expectedFeatureOutput(common string, byFormat map[string]string, formatName string) string {
	if value, ok := byFormat[formatName]; ok {
		return value
	}
	return common
}

func parseFeatureSpec(t *testing.T, path string) featureSpec {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	spec := newFeatureSpec(path)
	section := ""
	lines := strings.Split(string(content), "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "# ") && spec.name == "" {
			spec.name = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			continue
		}
		if section == "" && strings.HasPrefix(line, "> PENDING") {
			spec.pending = true
			spec.pendingReason = strings.TrimSpace(strings.TrimPrefix(line, "> PENDING"))
			continue
		}
		if strings.HasPrefix(line, "## ") {
			section = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			validateFeatureSection(t, path, section)
			continue
		}
		if strings.HasPrefix(line, "```") {
			language, name := parseFenceInfo(strings.TrimSpace(strings.TrimPrefix(line, "```")))
			if language == "" {
				t.Fatalf("%s: fenced code block missing format at line %d", path, i+1)
			}
			block, next := readFeatureFence(t, path, lines, i+1)
			parseFeatureFenceBlock(t, path, &spec, section, language, name, block)
			i = next
			continue
		}
		if section == "Commands" {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				t.Fatalf("%s: commands must be written in a fenced shell block", path)
			}
			continue
		}
	}
	validateFeatureSpec(t, path, spec)
	return spec
}

func newFeatureSpec(path string) featureSpec {
	spec := featureSpec{
		name:    strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		sources: make(map[string]string),
		files:   make(map[string]string),
		results: make(map[string]string),
	}
	if strings.HasPrefix(filepath.Base(path), "PENDING") {
		spec.pending = true
	}
	return spec
}

func validateFeatureSection(t *testing.T, path, section string) {
	t.Helper()
	if section == "STDOUT" || section == "STDERR" || strings.HasPrefix(section, "STDOUT ") || strings.HasPrefix(section, "STDERR ") {
		t.Fatalf("%s: %s sections are not supported; use command-local arrow directives", path, section)
	}
}

func parseFeatureFenceBlock(t *testing.T, path string, spec *featureSpec, section, language, name, block string) {
	t.Helper()
	switch section {
	case "Source Files":
		if name == "" {
			spec.sources[language] = block
		} else {
			spec.files[name] = block
		}
	case "Result Files":
		if name != "" {
			t.Fatalf("%s: result file block cannot have filename %q", path, name)
		}
		spec.results[language] = block
	case "Commands":
		if language != "shell" && language != "sh" {
			t.Fatalf("%s: command block must use shell, got %q", path, language)
		}
		parseFeatureCommandBlock(t, path, spec, block)
	default:
		t.Fatalf("%s: unexpected fenced block in section %q", path, section)
	}
}

func validateFeatureSpec(t *testing.T, path string, spec featureSpec) {
	t.Helper()
	if len(spec.sources) == 0 {
		t.Fatalf("%s: missing source files", path)
	}
	if len(spec.commands) == 0 {
		t.Fatalf("%s: missing commands", path)
	}
}

func parseFeatureCommandBlock(t *testing.T, path string, spec *featureSpec, block string) {
	t.Helper()
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "# ") {
			continue
		}
		if ok := parseFeatureTranscriptDirective(t, path, &spec.commands, trimmed); ok {
			continue
		}
		if strings.HasPrefix(trimmed, "$ ") {
			t.Fatalf("%s: command lines should not use '$': %q", path, line)
		}
		spec.commands = append(spec.commands, parseFeatureCommandLine(trimmed))
	}
}

func parseFeatureCommandLine(command string) featureCommand {
	return featureCommand{
		text:           command,
		stdoutByFormat: make(map[string]string),
		stderrByFormat: make(map[string]string),
	}
}

func parseFeatureTranscriptDirective(t *testing.T, path string, commands *[]featureCommand, line string) bool {
	t.Helper()
	kind, formatName, text, ok := parseFeatureArrowDirective(line)
	if !ok {
		return false
	}
	if len(*commands) == 0 {
		t.Fatalf("%s: output directive before command: %q", path, line)
	}
	command := &(*commands)[len(*commands)-1]
	switch kind {
	case "stdout":
		if formatName == "" {
			command.stdout += text + "\n"
		} else {
			command.stdoutByFormat[formatName] += text + "\n"
		}
	case "stderr":
		if formatName == "" {
			command.stderr += text + "\n"
		} else {
			command.stderrByFormat[formatName] += text + "\n"
		}
	case "exit":
		if formatName != "" {
			t.Fatalf("%s: exit directive cannot use a format prefix: %q", path, line)
		}
		switch text {
		case "0":
			command.exitCode = 0
		case "1":
			command.exitCode = 1
		default:
			t.Fatalf("%s: unsupported exit directive %q", path, line)
		}
	default:
		t.Fatalf("%s: unsupported directive %q", path, line)
	}
	return true
}

func parseFeatureArrowDirective(line string) (string, string, string, bool) {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "->") {
		return "stdout", "", strings.TrimSpace(strings.TrimPrefix(line, "->")), true
	}
	if strings.HasPrefix(line, "!->") {
		return "stderr", "", strings.TrimSpace(strings.TrimPrefix(line, "!->")), true
	}
	if before, after, ok := strings.Cut(line, " !->"); ok {
		formatName := strings.TrimSpace(before)
		if isFeatureFormat(formatName) {
			return "stderr", formatName, strings.TrimSpace(after), true
		}
	}
	if before, after, ok := strings.Cut(line, " ->"); ok {
		prefix := strings.TrimSpace(before)
		if prefix == "exit" {
			return "exit", "", strings.TrimSpace(after), true
		}
		if isFeatureFormat(prefix) {
			return "stdout", prefix, strings.TrimSpace(after), true
		}
	}
	return "", "", "", false
}

func isFeatureFormat(name string) bool {
	switch name {
	case "yaml", "toml", "json", "ini":
		return true
	default:
		return false
	}
}

func parseFenceInfo(info string) (string, string) {
	parts := strings.Fields(info)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func readFeatureFence(t *testing.T, path string, lines []string, start int) (string, int) {
	t.Helper()
	var out []string
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "```") {
			return strings.Join(out, "\n") + "\n", i
		}
		out = append(out, lines[i])
	}
	t.Fatalf("%s: unterminated fenced code block at line %d", path, start)
	return "", 0
}

func TestParseFeatureSpec(t *testing.T) {
	path := filepath.Join("..", "features", "set", "basic.md")
	spec := parseFeatureSpec(t, path)

	if len(spec.sources) != 2 || spec.sources["yaml"] == "" || spec.sources["toml"] == "" {
		t.Fatalf("sources not parsed: %#v", spec.sources)
	}
	if len(spec.commands) != 3 {
		t.Fatalf("commands = %#v", spec.commands)
	}
	if _, ok := spec.results["yaml"]; !ok {
		t.Fatal("missing yaml result")
	}
	if _, ok := spec.results["toml"]; !ok {
		t.Fatal("missing toml result")
	}

	getPath := filepath.Join("..", "features", "get", "basic.md")
	getSpec := parseFeatureSpec(t, getPath)
	if len(getSpec.results) != 0 {
		t.Fatalf("get results = %#v", getSpec.results)
	}

	refusalsPath := filepath.Join("..", "features", "get", "refusals.md")
	refusalsSpec := parseFeatureSpec(t, refusalsPath)
	if len(refusalsSpec.commands) != 3 || refusalsSpec.commands[2].exitCode != 1 || refusalsSpec.commands[2].text != "config get servers" {
		t.Fatalf("refusal commands = %#v", refusalsSpec.commands)
	}
	if refusalsSpec.commands[2].stderrByFormat["yaml"] == "" || refusalsSpec.commands[2].stderrByFormat["toml"] == "" {
		t.Fatalf("refusal stderr by format = %#v", refusalsSpec.commands[2].stderrByFormat)
	}

	pendingPath := filepath.Join(t.TempDir(), "PENDING-example.md")
	if err := os.WriteFile(pendingPath, []byte("# pending/example\n\n> PENDING Example pending reason.\n\n## Source Files\n\n```yaml\nvalue: old\n```\n\n## Commands\n\n```shell\nconfig set value new\n```\n"), 0644); err != nil {
		t.Fatal(err)
	}
	pending := parseFeatureSpec(t, pendingPath)
	if !pending.pending {
		t.Fatal("pending marker was not parsed")
	}
	if !strings.Contains(pending.pendingReason, "Example pending reason") {
		t.Fatalf("pending reason = %q", pending.pendingReason)
	}

	if language, name := parseFenceInfo("text value.txt"); language != "text" || name != "value.txt" {
		t.Fatalf("parseFenceInfo = %q %q", language, name)
	}
}

func splitCommand(command string) ([]string, error) {
	var args []string
	var current strings.Builder
	var quote rune
	escaped := false

	flush := func() {
		if current.Len() > 0 {
			args = append(args, current.String())
			current.Reset()
		}
	}

	for _, ch := range command {
		if escaped {
			current.WriteRune(ch)
			escaped = false
			continue
		}
		if ch == '\\' && quote != '\'' {
			escaped = true
			continue
		}
		if quote != 0 {
			if ch == quote {
				quote = 0
			} else {
				current.WriteRune(ch)
			}
			continue
		}
		switch ch {
		case '\'', '"':
			quote = ch
		case ' ', '\t':
			flush()
		default:
			current.WriteRune(ch)
		}
	}
	if escaped {
		current.WriteRune('\\')
	}
	if quote != 0 {
		return nil, errors.New("unterminated quote")
	}
	flush()
	return args, nil
}
