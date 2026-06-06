package reporter

import (
	"fmt"
	"io"
	"strings"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

type Options struct {
	Level   string
	NoColor bool
}

type Reporter struct {
	writer  io.Writer
	level   Level
	noColor bool
	prefix  string
}

func New(writer io.Writer, options Options) (Reporter, error) {
	level, err := ParseLevel(options.Level)
	if err != nil {
		return Reporter{}, err
	}

	return Reporter{
		writer:  writer,
		level:   level,
		noColor: options.NoColor,
	}, nil
}

func (r Reporter) With(prefix string) Reporter {
	r.prefix = prefix
	return r
}

func (r Reporter) Debug(message string, keyvals ...any) {
	r.log(DebugLevel, "debug", message, keyvals...)
}

func (r Reporter) Info(message string, keyvals ...any) {
	r.log(InfoLevel, "info", message, keyvals...)
}

func (r Reporter) Warn(message string, keyvals ...any) {
	r.log(WarnLevel, "warn", message, keyvals...)
}

func (r Reporter) Error(message string, keyvals ...any) {
	r.log(ErrorLevel, "error", message, keyvals...)
}

func (r Reporter) log(level Level, label, message string, keyvals ...any) {
	if r.writer == nil || level < r.level {
		return
	}

	if r.prefix != "" {
		message = r.prefix + ": " + message
	}

	fmt.Fprintf(r.writer, "%s %s", r.formatLevel(level, label), message)
	for index := 0; index < len(keyvals); index += 2 {
		key := fmt.Sprint(keyvals[index])
		value := ""
		if index+1 < len(keyvals) {
			value = fmt.Sprint(keyvals[index+1])
		}
		fmt.Fprintf(r.writer, " %s=%s", key, value)
	}
	fmt.Fprintln(r.writer)
}

func (r Reporter) formatLevel(level Level, label string) string {
	if r.noColor {
		return strings.ToUpper(label)
	}

	color := "36"
	switch level {
	case DebugLevel:
		color = "35"
	case InfoLevel:
		color = "36"
	case WarnLevel:
		color = "33"
	case ErrorLevel:
		color = "31"
	}

	return "\033[" + color + ";1m" + strings.ToUpper(label) + "\033[0m"
}

func ParseLevel(value string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	default:
		return 0, fmt.Errorf("unknown level %q", value)
	}
}
