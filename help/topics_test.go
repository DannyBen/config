package help

import (
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestTopicNamesIncludesEnvironment(t *testing.T) {
	names := TopicNames()

	if !slices.Contains(names, "environment") {
		t.Fatalf("TopicNames() = %v, want environment", names)
	}
}

func TestLookupEnvironment(t *testing.T) {
	body, ok := Lookup("environment")
	if !ok {
		t.Fatal("Lookup(environment) returned false")
	}

	if !strings.Contains(body, "CONFIG_FILE") {
		t.Fatalf("body = %q, want CONFIG_FILE", body)
	}
}

func TestRenderStylesSupportedMarkers(t *testing.T) {
	unsetEnv(t, "NO_COLOR")

	body := Render("Use **bold** and `inline`.\n```\nblock\nline 2\n```\nDone.")

	if strings.Contains(body, "**") || strings.Contains(body, "`") {
		t.Fatalf("body = %q, want formatting markers removed", body)
	}
	if !regexp.MustCompile(`\x1b\[[0-9;]*1m`).MatchString(body) {
		t.Fatalf("body = %q, want bold ANSI sequence", body)
	}
	if !regexp.MustCompile(`\x1b\[(32|38;5;2)m`).MatchString(body) {
		t.Fatalf("body = %q, want green inline ANSI sequence", body)
	}
	if !regexp.MustCompile(`\x1b\[(34|38;5;4)m`).MatchString(body) {
		t.Fatalf("body = %q, want blue code block ANSI sequence", body)
	}
}

func TestRenderNoColorDisablesStyling(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	body := Render("Use **bold** and `inline`.\n```\nblock\n```\nDone.")

	if strings.Contains(body, "**") || strings.Contains(body, "`") {
		t.Fatalf("body = %q, want formatting markers removed", body)
	}
	if regexp.MustCompile(`\x1b\[[0-9;]*m`).MatchString(body) {
		t.Fatalf("body = %q, want no ANSI styling", body)
	}
}

func TestRenderLeavesUnmatchedMarkersLiteral(t *testing.T) {
	unsetEnv(t, "NO_COLOR")

	body := Render("Use **bold and `inline.\n``` yaml\nnot a fence")

	if !strings.Contains(body, "**bold") {
		t.Fatalf("body = %q, want unmatched bold marker preserved", body)
	}
	if !strings.Contains(body, "`inline") {
		t.Fatalf("body = %q, want unmatched inline marker preserved", body)
	}
	if !strings.Contains(body, "``` yaml") {
		t.Fatalf("body = %q, want non-strict fence preserved", body)
	}
}

func TestLookupUnknownTopic(t *testing.T) {
	if _, ok := Lookup("unknown"); ok {
		t.Fatal("Lookup(unknown) returned true, want false")
	}
}

func unsetEnv(t *testing.T, name string) {
	t.Helper()

	original, ok := os.LookupEnv(name)
	if err := os.Unsetenv(name); err != nil {
		t.Fatalf("unset %s: %v", name, err)
	}
	t.Cleanup(func() {
		if ok {
			_ = os.Setenv(name, original)
			return
		}
		_ = os.Unsetenv(name)
	})
}
