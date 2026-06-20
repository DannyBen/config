package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type useOptions struct {
	configFile string
	status     bool
}

func newUseCommand(stdout, stderr io.Writer) *cobra.Command {
	var opts useOptions
	cmd := &cobra.Command{
		Use:   "use [FILE]",
		Short: "Use a config file in a child shell",
		Args: func(cmd *cobra.Command, args []string) error {
			rest, err := parseCommandArgs(args, 0, 1, "usage: config use [FILE]")
			if err != nil {
				return err
			}
			if len(rest) == 0 {
				opts.status = true
				return nil
			}
			opts.configFile = rest[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUse(opts, cmd.InOrStdin(), stdout, stderr)
		},
	}
	cmd.SetHelpFunc(helpPrinter("use"))
	return cmd
}

var runShell = runShellCommand

func runUse(opts useOptions, stdin io.Reader, stdout, stderr io.Writer) error {
	if opts.status {
		return runUseStatus(stdout)
	}

	configFile, err := resolveUseConfigFile(opts.configFile)
	if err != nil {
		return err
	}
	shell, err := resolveShell()
	if err != nil {
		return err
	}

	fmt.Fprintf(stderr, "Using %s.\nExit the shell to stop.\n", opts.configFile)
	return runShell(shell, withConfigFileEnv(os.Environ(), configFile), stdin, stdout, stderr)
}

func runUseStatus(stdout io.Writer) error {
	if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
		fmt.Fprintf(stdout, "Using %s\n", configFile)
	} else {
		fmt.Fprintln(stdout, "No config file is in use")
	}
	fmt.Fprintln(stdout, "usage: config use FILE")
	return nil
}

func resolveUseConfigFile(configFile string) (string, error) {
	absPath, err := filepath.Abs(configFile)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory", absPath)
	}
	file, err := os.Open(absPath)
	if err != nil {
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	return absPath, nil
}

func resolveShell() (string, error) {
	if shell := os.Getenv("SHELL"); shell != "" {
		if resolved, err := exec.LookPath(shell); err == nil {
			return resolved, nil
		}
	}
	if resolved, err := exec.LookPath("/bin/sh"); err == nil {
		return resolved, nil
	}
	return "", errors.New("shell not found")
}

func withConfigFileEnv(env []string, configFile string) []string {
	next := make([]string, 0, len(env)+1)
	replaced := false
	for _, item := range env {
		if strings.HasPrefix(item, "CONFIG_FILE=") {
			next = append(next, "CONFIG_FILE="+configFile)
			replaced = true
			continue
		}
		next = append(next, item)
	}
	if !replaced {
		next = append(next, "CONFIG_FILE="+configFile)
	}
	return next
}

func runShellCommand(shell string, env []string, stdin io.Reader, stdout, stderr io.Writer) error {
	cmd := exec.Command(shell)
	cmd.Env = env
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return silentExitError{code: exitErr.ExitCode()}
		}
		return err
	}
	return nil
}
