package cmd

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"config/format"
	"config/help"

	"github.com/spf13/cobra"
)

//go:embed help/*.txt
var helpFiles embed.FS

type setOptions struct {
	configFile string
	key        string
	value      string
	values     []string
	in         string
	on         string
	array      bool
	string     bool
	dry        bool
	diff       bool
	color      bool
}

type getOptions struct {
	configFile string
	key        string
	in         string
	on         []string
}

type unsetOptions struct {
	configFile string
	key        string
	in         string
	on         []string
	dry        bool
	diff       bool
	color      bool
}

type deleteOptions struct {
	configFile string
	key        string
	on         []string
	dry        bool
	diff       bool
	color      bool
}

type listOptions struct {
	configFile string
	key        string
}

type usageError struct {
	message string
}

func (e usageError) Error() string {
	return e.message
}

func Execute(args []string, version string, stdout, stderr io.Writer) error {
	return ExecuteWithIO(args, version, strings.NewReader(""), stdout, stderr)
}

func ExecuteWithIO(args []string, version string, stdin io.Reader, stdout, stderr io.Writer) error {
	root := NewRootCommand(version, stdout, stderr)
	root.SetIn(stdin)
	root.SetArgs(args)
	return root.Execute()
}

func PrintError(err error, stderr io.Writer) {
	if err == nil {
		return
	}
	var usage usageError
	if errors.As(err, &usage) {
		fmt.Fprintln(stderr, usage.Error())
		return
	}
	reporter, reporterErr := newReporter(stderr)
	if reporterErr != nil {
		fmt.Fprintln(stderr, "error:", reporterErr)
		return
	}
	reporter.Error(err.Error())
}

func NewRootCommand(version string, stdout, stderr io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:           "config COMMAND [options]",
		Short:         "Manipulate config files",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetHelpFunc(helpPrinter("root"))
	root.SetVersionTemplate("{{.Version}}\n")

	root.AddCommand(newSetCommand(stdout, stderr), newGetCommand(stdout), newUnsetCommand(stdout, stderr), newDeleteCommand(stdout, stderr), newListCommand(stdout), newHelpCommand(stdout))
	return root
}

func newSetCommand(stdout, stderr io.Writer) *cobra.Command {
	var opts setOptions
	cmd := &cobra.Command{
		Use:   "set [CONFIG_FILE] KEY VALUE... [options]",
		Short: "Create or update config values",
		Args: func(cmd *cobra.Command, args []string) error {
			configFile, rest, err := parseCommandArgs(args, 2, -1, "usage: config set [CONFIG_FILE] KEY VALUE... [options]")
			if err != nil {
				return err
			}
			opts.configFile = configFile
			opts.key = rest[0]
			opts.value = rest[1]
			opts.values = rest[1:]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validatePreviewOptions(opts.dry, opts.diff, opts.color); err != nil {
				return err
			}
			return runSet(opts, cmd.InOrStdin(), stdout, stderr)
		},
	}
	cmd.SetHelpFunc(helpPrinter("set"))
	cmd.Flags().StringVar(&opts.in, "in", "", "Edit a record in COLLECTION")
	cmd.Flags().StringVar(&opts.on, "on", "", "Select or create a record by FIELD:VALUE")
	cmd.Flags().BoolVarP(&opts.array, "array", "a", false, "Store VALUE as an array")
	cmd.Flags().BoolVarP(&opts.string, "string", "s", false, "Store VALUE as a string")
	cmd.Flags().BoolVarP(&opts.dry, "dry", "n", false, "Print the updated config without modifying the file")
	cmd.Flags().BoolVarP(&opts.diff, "diff", "d", false, "Print a unified diff without modifying the file")
	cmd.Flags().BoolVarP(&opts.color, "color", "c", false, "Colorize diff output")
	cmd.MarkFlagsMutuallyExclusive("array", "string")
	cmd.MarkFlagsMutuallyExclusive("dry", "diff")
	return cmd
}

func newGetCommand(stdout io.Writer) *cobra.Command {
	var opts getOptions
	cmd := &cobra.Command{
		Use:   "get [CONFIG_FILE] KEY",
		Short: "Show a config value",
		Args: func(cmd *cobra.Command, args []string) error {
			configFile, rest, err := parseCommandArgs(args, 1, 1, "usage: config get [CONFIG_FILE] KEY")
			if err != nil {
				return err
			}
			opts.configFile = configFile
			opts.key = rest[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(opts, stdout)
		},
	}
	cmd.SetHelpFunc(helpPrinter("get"))
	cmd.Flags().StringVar(&opts.in, "in", "", "Read a field from a record in COLLECTION")
	cmd.Flags().StringArrayVar(&opts.on, "on", nil, "Select a record by FIELD:VALUE")
	return cmd
}

func newUnsetCommand(stdout, stderr io.Writer) *cobra.Command {
	var opts unsetOptions
	cmd := &cobra.Command{
		Use:   "unset [CONFIG_FILE] KEY [options]",
		Short: "Delete a config value",
		Args: func(cmd *cobra.Command, args []string) error {
			configFile, rest, err := parseCommandArgs(args, 1, 1, "usage: config unset [CONFIG_FILE] KEY [options]")
			if err != nil {
				return err
			}
			opts.configFile = configFile
			opts.key = rest[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validatePreviewOptions(opts.dry, opts.diff, opts.color); err != nil {
				return err
			}
			return runUnset(opts, stdout, stderr)
		},
	}
	cmd.SetHelpFunc(helpPrinter("unset"))
	cmd.Flags().StringVar(&opts.in, "in", "", "Remove a field from a record in COLLECTION")
	cmd.Flags().StringArrayVar(&opts.on, "on", nil, "Select a record by FIELD:VALUE")
	cmd.Flags().BoolVarP(&opts.dry, "dry", "n", false, "Print the updated config without modifying the file")
	cmd.Flags().BoolVarP(&opts.diff, "diff", "d", false, "Print a unified diff without modifying the file")
	cmd.Flags().BoolVarP(&opts.color, "color", "c", false, "Colorize diff output")
	cmd.MarkFlagsMutuallyExclusive("dry", "diff")
	return cmd
}

func newDeleteCommand(stdout, stderr io.Writer) *cobra.Command {
	var opts deleteOptions
	cmd := &cobra.Command{
		Use:   "delete [CONFIG_FILE] KEY [options]",
		Short: "Delete a config container",
		Args: func(cmd *cobra.Command, args []string) error {
			configFile, rest, err := parseCommandArgs(args, 1, 1, "usage: config delete [CONFIG_FILE] KEY [options]")
			if err != nil {
				return err
			}
			opts.configFile = configFile
			opts.key = rest[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validatePreviewOptions(opts.dry, opts.diff, opts.color); err != nil {
				return err
			}
			return runDelete(opts, stdout, stderr)
		},
	}
	cmd.SetHelpFunc(helpPrinter("delete"))
	cmd.Flags().StringArrayVar(&opts.on, "on", nil, "Select a record by FIELD:VALUE")
	cmd.Flags().BoolVarP(&opts.dry, "dry", "n", false, "Print the updated config without modifying the file")
	cmd.Flags().BoolVarP(&opts.diff, "diff", "d", false, "Print a unified diff without modifying the file")
	cmd.Flags().BoolVarP(&opts.color, "color", "c", false, "Colorize diff output")
	cmd.MarkFlagsMutuallyExclusive("dry", "diff")
	return cmd
}

func newListCommand(stdout io.Writer) *cobra.Command {
	var opts listOptions
	cmd := &cobra.Command{
		Use:   "list [CONFIG_FILE] [KEY]",
		Short: "Show config values",
		Args: func(cmd *cobra.Command, args []string) error {
			configFile, rest, err := parseCommandArgs(args, 0, 1, "usage: config list [CONFIG_FILE] [KEY]")
			if err != nil {
				return err
			}
			opts.configFile = configFile
			if len(rest) == 1 {
				opts.key = rest[0]
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(opts, stdout)
		},
	}
	cmd.SetHelpFunc(helpPrinter("list"))
	return cmd
}

func newHelpCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "help [COMMAND|TOPIC]",
		Short: "Show command help or topic help",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return usageError{"usage: config help [COMMAND|TOPIC]"}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				fmt.Fprintln(stdout, helpIndex())
				return nil
			}

			name := args[0]
			if commandHelp, ok := commandHelpTopic(name); ok {
				text, err := helpText(commandHelp)
				if err != nil {
					return err
				}
				fmt.Fprint(stdout, text)
				return nil
			}

			if topic, ok := help.Lookup(name); ok {
				fmt.Fprintln(stdout, topic)
				return nil
			}

			return fmt.Errorf("unknown help topic %q", name)
		},
	}
	cmd.SetHelpFunc(helpPrinter("help"))
	return cmd
}

func helpIndex() string {
	lines := []string{
		"Usage:",
		"  config help [TOPIC]",
		"",
		"Commands:",
	}

	for _, command := range []string{"set", "get", "unset", "delete", "list", "help"} {
		lines = append(lines, "  "+command)
	}

	topics := help.TopicNames()
	otherTopics := make([]string, 0, len(topics))
	for _, topic := range topics {
		if _, ok := commandHelpTopic(topic); !ok && topic != "root" {
			otherTopics = append(otherTopics, topic)
		}
	}
	if len(otherTopics) > 0 {
		lines = append(lines, "", "Other topics:")
		for _, topic := range otherTopics {
			lines = append(lines, "  "+topic)
		}
	}

	lines = append(lines, "", "Shortcut:", "  config COMMAND --help|-h")
	return strings.Join(lines, "\n")
}

func commandHelpTopic(name string) (string, bool) {
	switch name {
	case "root", "set", "get", "unset", "delete", "list":
		return name, true
	default:
		return "", false
	}
}

func parseCommandArgs(args []string, minRest, maxRest int, usage string) (string, []string, error) {
	configFile := ""
	rest := args
	if len(args) > 0 && format.TargetPath(args[0]) {
		configFile = args[0]
		rest = args[1:]
	} else if env := os.Getenv("CONFIG_FILE"); env != "" {
		configFile = env
	} else {
		return "", nil, errors.New("config file not specified")
	}

	if len(rest) < minRest || (maxRest >= 0 && len(rest) > maxRest) {
		return "", nil, usageError{usage}
	}
	return configFile, rest, nil
}

func helpPrinter(name string) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		text, err := helpText(name)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "missing help for %s: %v\n", name, err)
			return
		}
		fmt.Fprint(cmd.OutOrStdout(), text)
	}
}

func helpText(name string) (string, error) {
	text, err := helpFiles.ReadFile("help/" + name + ".txt")
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(text), "\n") + "\n", nil
}

func runSet(opts setOptions, stdin io.Reader, stdout, stderr io.Writer) error {
	if opts.in != "" && opts.on == "" {
		return errors.New("flag --in requires --on")
	}

	values := opts.values
	if opts.value == "-" && len(opts.values) == 1 {
		content, err := io.ReadAll(stdin)
		if err != nil {
			return err
		}
		if opts.array {
			values = []string{string(content)}
		} else if opts.string {
			return runEdit("set", opts.configFile, opts.dry, opts.diff, opts.color, stdout, stderr, func(doc format.Document, source string) (string, error) {
				if opts.on != "" {
					return doc.SetInString(source, opts.in, opts.on, opts.key, string(content))
				}
				return doc.SetString(source, opts.key, string(content))
			})
		} else {
			values = []string{string(content)}
		}
	}
	if opts.string && len(values) > 1 {
		return errors.New("flag --string cannot be used with multiple values")
	}
	if opts.array || len(values) > 1 {
		return runEdit("set", opts.configFile, opts.dry, opts.diff, opts.color, stdout, stderr, func(doc format.Document, source string) (string, error) {
			if opts.on != "" {
				return doc.SetInArray(source, opts.in, opts.on, opts.key, values)
			}
			return doc.SetArray(source, opts.key, values)
		})
	}
	if opts.string {
		return runEdit("set", opts.configFile, opts.dry, opts.diff, opts.color, stdout, stderr, func(doc format.Document, source string) (string, error) {
			if opts.on != "" {
				return doc.SetInString(source, opts.in, opts.on, opts.key, values[0])
			}
			return doc.SetString(source, opts.key, values[0])
		})
	}
	return runEdit("set", opts.configFile, opts.dry, opts.diff, opts.color, stdout, stderr, func(doc format.Document, source string) (string, error) {
		if opts.on != "" {
			return doc.SetIn(source, opts.in, opts.on, opts.key, values[0])
		}
		return doc.Set(source, opts.key, values[0])
	})
}

func runUnset(opts unsetOptions, stdout, stderr io.Writer) error {
	if opts.in == "" && len(opts.on) > 0 {
		return errors.New("flag --on requires --in")
	}
	if opts.in != "" && len(opts.on) == 0 {
		return errors.New("flag --in requires --on")
	}

	return runEdit("unset", opts.configFile, opts.dry, opts.diff, opts.color, stdout, stderr, func(doc format.Document, source string) (string, error) {
		if opts.in != "" {
			return doc.UnsetIn(source, opts.in, opts.on, opts.key)
		}
		return doc.Unset(source, opts.key)
	})
}

func runDelete(opts deleteOptions, stdout, stderr io.Writer) error {
	return runEdit("delete", opts.configFile, opts.dry, opts.diff, opts.color, stdout, stderr, func(doc format.Document, source string) (string, error) {
		return doc.Delete(source, opts.key, opts.on)
	})
}

func runGet(opts getOptions, stdout io.Writer) error {
	if opts.in == "" && len(opts.on) > 0 {
		return errors.New("flag --on requires --in")
	}
	if opts.in != "" && len(opts.on) == 0 {
		return errors.New("flag --in requires --on")
	}

	configFile, err := resolveConfigFile(opts.configFile)
	if err != nil {
		return err
	}
	doc, _, err := format.Resolve(configFile)
	if err != nil {
		return err
	}

	source, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	var value string
	if opts.in != "" {
		value, err = doc.GetIn(string(source), opts.in, opts.on, opts.key)
	} else {
		value, err = doc.Get(string(source), opts.key)
	}
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, value)
	return nil
}

func runList(opts listOptions, stdout io.Writer) error {
	configFile, err := resolveConfigFile(opts.configFile)
	if err != nil {
		return err
	}
	doc, _, err := format.Resolve(configFile)
	if err != nil {
		return err
	}

	source, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	entries, err := doc.List(string(source), opts.key)
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, renderList(entries))
	return nil
}

func runEdit(command, configFile string, dry, diff, color bool, stdout, stderr io.Writer, edit func(format.Document, string) (string, error)) error {
	log, err := newReporter(stderr)
	if err != nil {
		return err
	}

	configFile, err = resolveConfigFile(configFile)
	if err != nil {
		return err
	}
	log.Debug("resolved config file", "command", command, "path", configFile)
	doc, formatName, err := format.Resolve(configFile)
	if err != nil {
		return err
	}
	log.Debug("detected format", "format", formatName)

	source, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	log.Debug("read config", "bytes", len(source))
	updated, err := edit(doc, string(source))
	if err != nil {
		return err
	}
	log.Debug("planned edit", "before", len(source), "after", len(updated))
	unchanged := updated == string(source)

	if diff {
		if unchanged {
			log.Debug("no changes")
			return nil
		}
		log.Debug("printing diff")
		diff := unifiedDiff(configFile, string(source), updated)
		if color {
			diff = colorizeDiff(diff)
		}
		fmt.Fprint(stdout, diff)
		return nil
	}
	if dry {
		log.Debug("printing dry output")
		fmt.Fprint(stdout, updated)
		return nil
	}
	if unchanged {
		log.Debug("no changes")
		return nil
	}

	mode := os.FileMode(0644)
	if info, err := os.Stat(configFile); err == nil {
		mode = info.Mode()
	}
	if err := os.WriteFile(configFile, []byte(updated), mode); err != nil {
		return err
	}
	log.Debug("wrote config", "path", configFile, "bytes", len(updated))
	return nil
}

func resolveConfigFile(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	return "", errors.New("config file not specified")
}

func validatePreviewOptions(dry, diff, color bool) error {
	if dry && diff {
		return errors.New("flags --dry and --diff cannot be used together")
	}
	if color && !diff {
		return errors.New("flag --color requires --diff")
	}
	return nil
}

func unifiedDiff(name, before, after string) string {
	const contextLines = 3

	beforeLines := strings.SplitAfter(before, "\n")
	afterLines := strings.SplitAfter(after, "\n")
	if len(beforeLines) > 0 && beforeLines[len(beforeLines)-1] == "" {
		beforeLines = beforeLines[:len(beforeLines)-1]
	}
	if len(afterLines) > 0 && afterLines[len(afterLines)-1] == "" {
		afterLines = afterLines[:len(afterLines)-1]
	}

	var out strings.Builder
	fmt.Fprintf(&out, "--- %s\n+++ %s\n", name, name)
	for _, hunk := range diffHunks(lineDiff(beforeLines, afterLines), contextLines) {
		fmt.Fprintf(&out, "@@ -%s +%s @@\n", diffRange(hunk.beforeStart, hunk.beforeCount), diffRange(hunk.afterStart, hunk.afterCount))
		for _, op := range hunk.ops {
			writeDiffOp(&out, op)
		}
	}
	return out.String()
}

func writeDiffOp(out *strings.Builder, op diffOp) {
	out.WriteByte(op.kind)
	out.WriteString(op.text)
	if !strings.HasSuffix(op.text, "\n") {
		out.WriteByte('\n')
		out.WriteString("\\ No newline at end of file\n")
	}
}

type diffOp struct {
	kind byte
	text string
}

type diffHunk struct {
	beforeStart int
	beforeCount int
	afterStart  int
	afterCount  int
	ops         []diffOp
}

func lineDiff(before, after []string) []diffOp {
	n, m := len(before), len(after)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if before[i] == after[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	var ops []diffOp
	i, j := 0, 0
	for i < n && j < m {
		if before[i] == after[j] {
			ops = append(ops, diffOp{kind: ' ', text: before[i]})
			i++
			j++
		} else if dp[i+1][j] >= dp[i][j+1] {
			ops = append(ops, diffOp{kind: '-', text: before[i]})
			i++
		} else {
			ops = append(ops, diffOp{kind: '+', text: after[j]})
			j++
		}
	}
	for ; i < n; i++ {
		ops = append(ops, diffOp{kind: '-', text: before[i]})
	}
	for ; j < m; j++ {
		ops = append(ops, diffOp{kind: '+', text: after[j]})
	}
	return ops
}

func diffHunks(ops []diffOp, contextLines int) []diffHunk {
	if len(ops) == 0 {
		return nil
	}

	beforeAt := make([]int, len(ops))
	afterAt := make([]int, len(ops))
	beforeLine, afterLine := 1, 1
	for i, op := range ops {
		beforeAt[i] = beforeLine
		afterAt[i] = afterLine
		switch op.kind {
		case ' ':
			beforeLine++
			afterLine++
		case '-':
			beforeLine++
		case '+':
			afterLine++
		}
	}

	var ranges [][2]int
	for i, op := range ops {
		if op.kind == ' ' {
			continue
		}
		start := max(0, i-contextLines)
		end := min(len(ops), i+contextLines+1)
		if len(ranges) > 0 && start <= ranges[len(ranges)-1][1] {
			if end > ranges[len(ranges)-1][1] {
				ranges[len(ranges)-1][1] = end
			}
			continue
		}
		ranges = append(ranges, [2]int{start, end})
	}

	hunks := make([]diffHunk, 0, len(ranges))
	for _, r := range ranges {
		hunkOps := ops[r[0]:r[1]]
		hunk := diffHunk{
			beforeStart: beforeAt[r[0]],
			afterStart:  afterAt[r[0]],
			ops:         hunkOps,
		}
		for _, op := range hunkOps {
			switch op.kind {
			case ' ':
				hunk.beforeCount++
				hunk.afterCount++
			case '-':
				hunk.beforeCount++
			case '+':
				hunk.afterCount++
			}
		}
		hunks = append(hunks, hunk)
	}
	return hunks
}

func diffRange(start, count int) string {
	if count == 1 {
		return strconv.Itoa(start)
	}
	return fmt.Sprintf("%d,%d", start, count)
}

func colorizeDiff(diff string) string {
	lines := strings.SplitAfter(diff, "\n")
	var out strings.Builder
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ") || strings.HasPrefix(line, "@@ "):
			out.WriteString("\x1b[1m")
			out.WriteString(line)
			out.WriteString("\x1b[0m")
		case strings.HasPrefix(line, "-"):
			out.WriteString("\x1b[31m")
			out.WriteString(line)
			out.WriteString("\x1b[0m")
		case strings.HasPrefix(line, "+"):
			out.WriteString("\x1b[32m")
			out.WriteString(line)
			out.WriteString("\x1b[0m")
		default:
			out.WriteString(line)
		}
	}
	return out.String()
}
