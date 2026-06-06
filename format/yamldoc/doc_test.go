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

func TestApplyVerifiedSetRejectsSemanticMismatch(t *testing.T) {
	source := "server:\n  host: localhost\n"
	planned := edit{start: len(source), end: len(source), text: "  port: 3000\n"}

	_, err := applyVerifiedSet(source, planned, semanticScalar([]string{"server", "mode"}, "3000"))

	if err == nil {
		t.Fatal("expected semantic verification error")
	}
	if err.Error() != "internal YAML patch verification failed" {
		t.Fatalf("error = %q", err.Error())
	}
}
