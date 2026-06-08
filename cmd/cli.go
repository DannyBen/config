package cmd

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/dannyben/config/format"
	"github.com/dannyben/config/help"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

//go:embed help/*.txt
var helpFiles embed.FS

type setOptions struct {
	configFile string
	key        string
	value      string
	in         string
	on         string
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
	ifValue    string
	ifSet      bool
	ifExists   bool
	dry        bool
	diff       bool
	color      bool
}

type deleteOptions struct {
	configFile string
	key        string
	on         []string
	ifEmpty    bool
	dry        bool
	diff       bool
	color      bool
}

type arrayOptions struct {
	configFile string
	key        string
	values     []string
	dry        bool
	diff       bool
	color      bool
}

type listOptions struct {
	configFile string
	key        string
	color      bool
}

type dumpOptions struct {
	configFile string
	key        string
	json       bool
}

type editOptions struct {
	configFile string
}

type usageError struct {
	message string
}

func (e usageError) Error() string {
	return e.message
}

type silentExitError struct {
	code int
}

func (e silentExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.code)
}

func (e silentExitError) ExitCode() int {
	return e.code
}

type exitCoder interface {
	ExitCode() int
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
	var silent silentExitError
	if errors.As(err, &silent) {
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

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var coder exitCoder
	if errors.As(err, &coder) {
		return coder.ExitCode()
	}
	return 1
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
	root.CompletionOptions.DisableDefaultCmd = true

	root.AddCommand(newSetCommand(stdout, stderr), newGetCommand(stdout), newUnsetCommand(stdout, stderr), newDeleteCommand(stdout, stderr), newArrayCommand(stdout, stderr), newListCommand(stdout), newDumpCommand(stdout), newUseCommand(stdout, stderr), newEditCommand(stdout, stderr), newCompletionCommand(stdout), newHelpCommand(stdout))
	return root
}

func newSetCommand(stdout, stderr io.Writer) *cobra.Command {
	var opts setOptions
	cmd := &cobra.Command{
		Use:   "set KEY VALUE [options]",
		Short: "Create or update config values",
		Args: func(cmd *cobra.Command, args []string) error {
			rest, err := parseCommandArgs(args, 2, 2, "usage: config set KEY VALUE [options]")
			if err != nil {
				return err
			}
			opts.key = rest[0]
			opts.value = rest[1]
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
	addFileFlag(cmd, &opts.configFile)
	cmd.Flags().StringVar(&opts.in, "in", "", "Edit a record in COLLECTION")
	cmd.Flags().StringVar(&opts.on, "on", "", "Select or create a record by FIELD:VALUE")
	cmd.Flags().BoolVarP(&opts.string, "string", "s", false, "Store VALUE as a string")
	cmd.Flags().BoolVarP(&opts.dry, "dry", "n", false, "Print the updated config without modifying the file")
	cmd.Flags().BoolVarP(&opts.diff, "diff", "d", false, "Print a unified diff without modifying the file")
	cmd.Flags().BoolVarP(&opts.color, "color", "c", false, "Colorize diff output")
	cmd.MarkFlagsMutuallyExclusive("dry", "diff")
	return cmd
}

func newGetCommand(stdout io.Writer) *cobra.Command {
	var opts getOptions
	cmd := &cobra.Command{
		Use:   "get KEY",
		Short: "Show a config value",
		Args: func(cmd *cobra.Command, args []string) error {
			rest, err := parseCommandArgs(args, 1, 1, "usage: config get KEY")
			if err != nil {
				return err
			}
			opts.key = rest[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(opts, stdout)
		},
	}
	cmd.SetHelpFunc(helpPrinter("get"))
	addFileFlag(cmd, &opts.configFile)
	cmd.Flags().StringVar(&opts.in, "in", "", "Read a field from a record in COLLECTION")
	cmd.Flags().StringArrayVar(&opts.on, "on", nil, "Select a record by FIELD:VALUE")
	return cmd
}

func newUnsetCommand(stdout, stderr io.Writer) *cobra.Command {
	var opts unsetOptions
	cmd := &cobra.Command{
		Use:   "unset KEY [options]",
		Short: "Delete a config value",
		Args: func(cmd *cobra.Command, args []string) error {
			rest, err := parseCommandArgs(args, 1, 1, "usage: config unset KEY [options]")
			if err != nil {
				return err
			}
			opts.key = rest[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validatePreviewOptions(opts.dry, opts.diff, opts.color); err != nil {
				return err
			}
			opts.ifSet = cmd.Flags().Changed("if")
			return runUnset(opts, stdout, stderr)
		},
	}
	cmd.SetHelpFunc(helpPrinter("unset"))
	addFileFlag(cmd, &opts.configFile)
	cmd.Flags().StringVar(&opts.in, "in", "", "Remove a field from a record in COLLECTION")
	cmd.Flags().StringArrayVar(&opts.on, "on", nil, "Select a record by FIELD:VALUE")
	cmd.Flags().StringVar(&opts.ifValue, "if", "", "Only unset when the current value matches VALUE")
	cmd.Flags().BoolVar(&opts.ifExists, "if-exists", false, "Do nothing when KEY is not set")
	cmd.Flags().BoolVarP(&opts.dry, "dry", "n", false, "Print the updated config without modifying the file")
	cmd.Flags().BoolVarP(&opts.diff, "diff", "d", false, "Print a unified diff without modifying the file")
	cmd.Flags().BoolVarP(&opts.color, "color", "c", false, "Colorize diff output")
	cmd.MarkFlagsMutuallyExclusive("dry", "diff")
	return cmd
}

func newDeleteCommand(stdout, stderr io.Writer) *cobra.Command {
	var opts deleteOptions
	cmd := &cobra.Command{
		Use:     "delete KEY [options]",
		Aliases: []string{"del"},
		Short:   "Delete a config container",
		Args: func(cmd *cobra.Command, args []string) error {
			rest, err := parseCommandArgs(args, 1, 1, "usage: config delete KEY [options]")
			if err != nil {
				return err
			}
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
	addFileFlag(cmd, &opts.configFile)
	cmd.Flags().StringArrayVar(&opts.on, "on", nil, "Select a record by FIELD:VALUE")
	cmd.Flags().BoolVar(&opts.ifEmpty, "if-empty", false, "Only delete when the container has no values")
	cmd.Flags().BoolVarP(&opts.dry, "dry", "n", false, "Print the updated config without modifying the file")
	cmd.Flags().BoolVarP(&opts.diff, "diff", "d", false, "Print a unified diff without modifying the file")
	cmd.Flags().BoolVarP(&opts.color, "color", "c", false, "Colorize diff output")
	cmd.MarkFlagsMutuallyExclusive("dry", "diff")
	return cmd
}

func newListCommand(stdout io.Writer) *cobra.Command {
	var opts listOptions
	cmd := &cobra.Command{
		Use:     "list [KEY]",
		Aliases: []string{"ls"},
		Short:   "Show config values",
		Args: func(cmd *cobra.Command, args []string) error {
			rest, err := parseCommandArgs(args, 0, 1, "usage: config list [KEY]")
			if err != nil {
				return err
			}
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
	addFileFlag(cmd, &opts.configFile)
	cmd.Flags().BoolVarP(&opts.color, "color", "c", false, "Colorize keys and separators")
	return cmd
}

func newDumpCommand(stdout io.Writer) *cobra.Command {
	var opts dumpOptions
	cmd := &cobra.Command{
		Use:   "dump [KEY]",
		Short: "Dump config data",
		Args: func(cmd *cobra.Command, args []string) error {
			rest, err := parseCommandArgs(args, 0, 1, "usage: config dump [KEY]")
			if err != nil {
				return err
			}
			if len(rest) == 1 {
				opts.key = rest[0]
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDump(opts, stdout)
		},
	}
	cmd.SetHelpFunc(helpPrinter("dump"))
	addFileFlag(cmd, &opts.configFile)
	cmd.Flags().BoolVar(&opts.json, "json", false, "Dump as JSON")
	return cmd
}

func newEditCommand(stdout, stderr io.Writer) *cobra.Command {
	var opts editOptions
	cmd := &cobra.Command{
		Use:   "edit [options]",
		Short: "Open the config file in an editor",
		Args: func(cmd *cobra.Command, args []string) error {
			_, err := parseCommandArgs(args, 0, 0, "usage: config edit [options]")
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOpenEditor(opts, cmd.InOrStdin(), stdout, stderr)
		},
	}
	cmd.SetHelpFunc(helpPrinter("edit"))
	addFileFlag(cmd, &opts.configFile)
	return cmd
}

func newCompletionCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion SHELL",
		Short: "Generate shell completion scripts",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return usageError{"usage: config completion SHELL"}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompletion(cmd.Root(), args[0], stdout)
		},
	}
	cmd.SetHelpFunc(helpPrinter("completion"))
	return cmd
}

func newArrayCommand(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "array COMMAND [options]",
		Short: "Manipulate scalar arrays",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return usageError{"usage: config array COMMAND [options]"}
			}
			return nil
		},
	}
	cmd.SetHelpFunc(helpPrinter("array"))
	cmd.AddCommand(
		newArrayEditCommand("set", "Replace a scalar array", stdout, stderr, func(doc format.Document, source string, opts arrayOptions) (string, error) {
			return doc.SetArray(source, opts.key, opts.values)
		}),
		newArrayEditCommand("add", "Add values to a scalar array", stdout, stderr, func(doc format.Document, source string, opts arrayOptions) (string, error) {
			return doc.ArrayAdd(source, opts.key, opts.values)
		}),
		newArrayEditCommand("delete", "Remove values from a scalar array", stdout, stderr, func(doc format.Document, source string, opts arrayOptions) (string, error) {
			return doc.ArrayDel(source, opts.key, opts.values)
		}, "del"),
	)
	return cmd
}

func newArrayEditCommand(name, short string, stdout, stderr io.Writer, edit func(format.Document, string, arrayOptions) (string, error), aliases ...string) *cobra.Command {
	var opts arrayOptions
	cmd := &cobra.Command{
		Use:     name + " KEY VALUE... [options]",
		Aliases: aliases,
		Short:   short,
		Args: func(cmd *cobra.Command, args []string) error {
			rest, err := parseCommandArgs(args, 2, -1, "usage: config array "+name+" KEY VALUE... [options]")
			if err != nil {
				return err
			}
			opts.key = rest[0]
			opts.values = rest[1:]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validatePreviewOptions(opts.dry, opts.diff, opts.color); err != nil {
				return err
			}
			return runArrayEdit(name, opts, stdout, stderr, edit)
		},
	}
	cmd.SetHelpFunc(helpPrinter("array-" + name))
	addFileFlag(cmd, &opts.configFile)
	cmd.Flags().BoolVarP(&opts.dry, "dry", "n", false, "Print the updated config without modifying the file")
	cmd.Flags().BoolVarP(&opts.diff, "diff", "d", false, "Print a unified diff without modifying the file")
	cmd.Flags().BoolVarP(&opts.color, "color", "c", false, "Colorize diff output")
	cmd.MarkFlagsMutuallyExclusive("dry", "diff")
	return cmd
}

func newHelpCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "help [COMMAND|TOPIC]",
		Short: "Show command help or topic help",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 2 {
				return usageError{"usage: " + helpUsage}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				fmt.Fprintln(stdout, helpIndex())
				return nil
			}

			name := strings.Join(args, " ")
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
	cmd.SetHelpFunc(helpIndexPrinter(stdout))
	return cmd
}

func helpIndexPrinter(stdout io.Writer) func(*cobra.Command, []string) {
	return func(*cobra.Command, []string) {
		fmt.Fprintln(stdout, helpIndex())
	}
}

const helpUsage = "config help [COMMAND|TOPIC]"

func helpIndex() string {
	lines := []string{
		"Usage:",
		"  " + helpUsage,
		"",
		"Commands:",
	}

	for _, command := range []string{"set", "get", "unset", "delete", "array", "array set", "array add", "array delete", "list", "dump", "use", "edit", "completion"} {
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

	return strings.Join(lines, "\n")
}

func commandHelpTopic(name string) (string, bool) {
	switch name {
	case "root", "set", "get", "unset", "delete", "array", "list", "dump", "use", "edit", "completion":
		return name, true
	case "del":
		return "delete", true
	case "ls":
		return "list", true
	case "array set":
		return "array-set", true
	case "array add":
		return "array-add", true
	case "array delete", "array del":
		return "array-delete", true
	default:
		return "", false
	}
}

func addFileFlag(cmd *cobra.Command, target *string) {
	cmd.Flags().StringVarP(target, "file", "f", "", "Path to the config file")
}

func parseCommandArgs(args []string, minRest, maxRest int, usage string) ([]string, error) {
	if len(args) < minRest || (maxRest >= 0 && len(args) > maxRest) {
		return nil, usageError{usage}
	}
	return args, nil
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

	value := opts.value
	if opts.value == "-" {
		content, err := io.ReadAll(stdin)
		if err != nil {
			return err
		}
		value = string(content)
	}
	if opts.string {
		return runEdit("set", opts.configFile, opts.dry, opts.diff, opts.color, stdout, stderr, func(doc format.Document, source string) (string, error) {
			if opts.on != "" {
				return doc.SetInString(source, opts.in, opts.on, opts.key, value)
			}
			return doc.SetString(source, opts.key, value)
		})
	}
	return runEdit("set", opts.configFile, opts.dry, opts.diff, opts.color, stdout, stderr, func(doc format.Document, source string) (string, error) {
		if opts.on != "" {
			return doc.SetIn(source, opts.in, opts.on, opts.key, value)
		}
		return doc.Set(source, opts.key, value)
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
		if opts.ifSet || opts.ifExists {
			var value string
			var err error
			if opts.in != "" {
				value, err = doc.GetIn(source, opts.in, opts.on, opts.key)
			} else {
				value, err = doc.Get(source, opts.key)
			}
			if err != nil {
				if isNotSetError(err) {
					return source, nil
				}
				return "", err
			}
			if opts.ifSet && value != opts.ifValue {
				return source, nil
			}
		}
		if opts.in != "" {
			return doc.UnsetIn(source, opts.in, opts.on, opts.key)
		}
		return doc.Unset(source, opts.key)
	})
}

func runArrayEdit(command string, opts arrayOptions, stdout, stderr io.Writer, edit func(format.Document, string, arrayOptions) (string, error)) error {
	return runEdit("array "+command, opts.configFile, opts.dry, opts.diff, opts.color, stdout, stderr, func(doc format.Document, source string) (string, error) {
		return edit(doc, source, opts)
	})
}

func isNotSetError(err error) bool {
	return strings.HasSuffix(err.Error(), " is not set")
}

func runDelete(opts deleteOptions, stdout, stderr io.Writer) error {
	if opts.ifEmpty && len(opts.on) > 0 {
		return errors.New("flag --if-empty cannot be used with --on")
	}
	return runEdit("delete", opts.configFile, opts.dry, opts.diff, opts.color, stdout, stderr, func(doc format.Document, source string) (string, error) {
		if opts.ifEmpty {
			return doc.DeleteIfEmpty(source, opts.key)
		}
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
	source, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	doc, _, err := format.ResolveSource(configFile, source)
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
	source, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	doc, _, err := format.ResolveSource(configFile, source)
	if err != nil {
		return err
	}

	entries, err := doc.List(string(source), opts.key)
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, renderList(entries, opts.color))
	return nil
}

func runDump(opts dumpOptions, stdout io.Writer) error {
	configFile, err := resolveConfigFile(opts.configFile)
	if err != nil {
		return err
	}
	source, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	doc, _, err := format.ResolveSource(configFile, source)
	if err != nil {
		return err
	}

	value, err := doc.Dump(string(source), opts.key)
	if err != nil {
		return err
	}
	output, err := renderDump(value, opts.json)
	if err != nil {
		return err
	}
	fmt.Fprint(stdout, output)
	return nil
}

func runOpenEditor(opts editOptions, stdin io.Reader, stdout, stderr io.Writer) error {
	configFile, err := resolveConfigFile(opts.configFile)
	if err != nil {
		return err
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command("sh", "-c", editor+" \"$1\"", "config-edit", configFile)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func runCompletion(root *cobra.Command, shell string, stdout io.Writer) error {
	switch shell {
	case "bash":
		return root.GenBashCompletionV2(stdout, true)
	case "zsh":
		return root.GenZshCompletion(stdout)
	case "fish":
		return root.GenFishCompletion(stdout, true)
	default:
		return fmt.Errorf("unsupported shell %q", shell)
	}
}

func renderDump(value any, asJSON bool) (string, error) {
	if asJSON {
		out, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return "", err
		}
		return string(out) + "\n", nil
	}

	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(2)
	if err := encoder.Encode(value); err != nil {
		encoder.Close()
		return "", err
	}
	if err := encoder.Close(); err != nil {
		return "", err
	}
	return out.String(), nil
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

	source, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	doc, formatName, err := format.ResolveSource(configFile, source)
	if err != nil {
		return err
	}
	log.Debug("detected format", "format", formatName)
	log.Debug("read config", "bytes", len(source))
	updated, err := edit(doc, string(source))
	if err != nil {
		return err
	}
	log.Debug("planned edit", "before", len(source), "after", len(updated))
	unchanged := updated == string(source)

	handled := printEditPreview(configFile, string(source), updated, dry, diff, color, stdout, log)
	if handled {
		return nil
	}
	if unchanged {
		log.Debug("no changes")
		return nil
	}

	if err := writeConfigFile(configFile, updated); err != nil {
		return err
	}
	log.Debug("wrote config", "path", configFile, "bytes", len(updated))
	return nil
}

type editLogger interface {
	Debug(message string, keyvals ...any)
}

func printEditPreview(configFile, source, updated string, dry, diff, color bool, stdout io.Writer, log editLogger) bool {
	if diff {
		if updated == source {
			log.Debug("no changes")
			return true
		}
		log.Debug("printing diff")
		output := unifiedDiff(configFile, source, updated)
		if color {
			output = colorizeDiff(output)
		}
		fmt.Fprint(stdout, output)
		return true
	}
	if dry {
		log.Debug("printing dry output")
		fmt.Fprint(stdout, updated)
		return true
	}
	return false
}

func writeConfigFile(configFile, updated string) error {
	mode := os.FileMode(0644)
	if info, err := os.Stat(configFile); err == nil {
		mode = info.Mode()
	}
	if err := os.WriteFile(configFile, []byte(updated), mode); err != nil {
		return err
	}
	return nil
}

func resolveConfigFile(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	if env := os.Getenv("CONFIG_FILE"); env != "" {
		return env, nil
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
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(before),
		B:        difflib.SplitLines(after),
		FromFile: name,
		ToFile:   name,
		Context:  3,
	}
	text, _ := difflib.GetUnifiedDiffString(diff)
	return text
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
