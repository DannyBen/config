package yamldoc

import (
	"strings"
	"testing"
)

func TestDeleteReportsInvalidYAMLBeforePlanning(t *testing.T) {
	_, err := Delete("server:\n  - port: 3000\n    bad\n", "server", nil)

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid YAML:") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetReportsInvalidYAMLBeforeUnsupportedError(t *testing.T) {
	_, err := Set("server:\n  - port: 3000\n    bad\n", "server.port", "3001")

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid YAML:") {
		t.Fatalf("unexpected error: %v", err)
	}
}
