package reporter

import (
	"bytes"
	"strings"
	"testing"
)

func TestReporterDebugRespectsLevel(t *testing.T) {
	var buffer bytes.Buffer
	reporter, err := New(&buffer, Options{Level: "info", NoColor: true})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	reporter.Debug("hidden")
	reporter.Info("visible", "symbol", "SPY.US")

	output := buffer.String()
	if strings.Contains(output, "hidden") {
		t.Fatalf("output = %q, want debug message hidden", output)
	}
	if !strings.Contains(output, "INFO visible symbol=SPY.US") {
		t.Fatalf("output = %q, want info message", output)
	}
}

func TestReporterDebugLevel(t *testing.T) {
	var buffer bytes.Buffer
	reporter, err := New(&buffer, Options{Level: "debug", NoColor: true})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	reporter.With("report").Debug("built model", "assets", 3)

	if !strings.Contains(buffer.String(), "DEBUG report: built model assets=3") {
		t.Fatalf("output = %q, want scoped debug message", buffer.String())
	}
}

func TestReporterColor(t *testing.T) {
	var buffer bytes.Buffer
	reporter, err := New(&buffer, Options{Level: "debug"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	reporter.Debug("colored")

	if !strings.Contains(buffer.String(), "\033[35;1mDEBUG\033[0m") {
		t.Fatalf("output = %q, want colored debug level", buffer.String())
	}
}

func TestReporterRejectsUnknownLevel(t *testing.T) {
	_, err := New(&bytes.Buffer{}, Options{Level: "loud"})
	if err == nil {
		t.Fatal("New returned nil error, want unknown level error")
	}
}
