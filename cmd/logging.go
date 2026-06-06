package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/dannyben/config/internal/reporter"
)

const envLogLevel = "CONFIG_LOG_LEVEL"

var cliLogWriter io.Writer = os.Stderr

func newReporter(writer io.Writer) (reporter.Reporter, error) {
	level := os.Getenv(envLogLevel)
	source := envLogLevel
	if level == "" {
		level = os.Getenv("LOG_LEVEL")
		source = "LOG_LEVEL"
	}

	reporter, err := reporter.New(writer, reporter.Options{
		Level:   level,
		NoColor: noColor(),
	})
	if err != nil {
		return reporter, fmt.Errorf("%s: %w", source, err)
	}
	return reporter, nil
}

func noColor() bool {
	_, ok := os.LookupEnv("NO_COLOR")
	return ok
}
