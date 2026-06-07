package inidoc

import (
	"reflect"
	"testing"
)

func TestListIncludesGlobalAndSectionValues(t *testing.T) {
	source := `
# full-line comment
title = config

[style]
color = red
size = 14
`

	got, err := List(source, "")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	want := []Entry{
		{Key: "title", Value: "config"},
		{Key: "style.color", Value: "red"},
		{Key: "style.size", Value: "14"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}

func TestListSectionAndKey(t *testing.T) {
	source := "title = config\n[style]\ncolor = red\nsize = 14\n"

	got, err := List(source, "style")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	want := []Entry{
		{Key: "style.color", Value: "red"},
		{Key: "style.size", Value: "14"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}

	got, err = List(source, "style.color")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	want = []Entry{{Key: "style.color", Value: "red"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}

func TestGetValue(t *testing.T) {
	got, err := Get("title = config\n[style]\ncolor = red\n", "style.color")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got != "red" {
		t.Fatalf("Get = %q, want red", got)
	}
}

func TestDump(t *testing.T) {
	source := "title = config\nserver.port = 3000\n[style]\ncolor = red\nsize = 14\n"

	got, err := Dump(source, "")
	if err != nil {
		t.Fatalf("Dump returned error: %v", err)
	}

	want := map[string]any{
		"title":       "config",
		"server.port": "3000",
		"style": map[string]any{
			"color": "red",
			"size":  "14",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Dump = %#v, want %#v", got, want)
	}
}

func TestDumpSection(t *testing.T) {
	got, err := Dump("title = config\n[style]\ncolor = red\nsize = 14\n", "style")
	if err != nil {
		t.Fatalf("Dump returned error: %v", err)
	}

	want := map[string]any{"color": "red", "size": "14"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Dump = %#v, want %#v", got, want)
	}
}

func TestDumpValue(t *testing.T) {
	got, err := Dump("title = config\n[style]\ncolor = red\n", "style.color")
	if err != nil {
		t.Fatalf("Dump returned error: %v", err)
	}
	if got != "red" {
		t.Fatalf("Dump = %#v, want red", got)
	}
}

func TestDumpRefusesDuplicateKeys(t *testing.T) {
	_, err := Dump("port = 3000\nport = 3001\n", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDumpRefusesKeySectionConflict(t *testing.T) {
	_, err := Dump("server = localhost\n[server]\nport = 3000\n", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSetUpdatesGlobalKey(t *testing.T) {
	got, err := Set("title = config\n", "title", "demo")
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "title = demo\n"
	if got != want {
		t.Fatalf("Set output mismatch\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestSetUpdatesSectionKey(t *testing.T) {
	source := "[server]\nport=3000\nhost = localhost\n"

	got, err := Set(source, "server.port", "3001")
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	want := "[server]\nport= 3001\nhost = localhost\n"
	if got != want {
		t.Fatalf("Set output mismatch\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestSetAppendsGlobalKey(t *testing.T) {
	got, err := Set("title = config\n", "enabled", "true")
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "title = config\nenabled = true\n"
	if got != want {
		t.Fatalf("Set output mismatch\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestSetInsertsGlobalKeyBeforeFirstSection(t *testing.T) {
	got, err := Set("title = config\n\n[server]\nhost = localhost\n", "enabled", "true")
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "title = config\nenabled = true\n\n[server]\nhost = localhost\n"
	if got != want {
		t.Fatalf("Set output mismatch\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestSetInsertsIntoExistingSection(t *testing.T) {
	source := "title = config\n\n[server]\nhost = localhost\n\n[database]\nhost = db\n"

	got, err := Set(source, "server.port", "3000")
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	want := "title = config\n\n[server]\nhost = localhost\n\nport = 3000\n[database]\nhost = db\n"
	if got != want {
		t.Fatalf("Set output mismatch\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestSetAppendsMissingSection(t *testing.T) {
	got, err := Set("title = config\n", "server.port", "3000")
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "title = config\n\n[server]\nport = 3000\n"
	if got != want {
		t.Fatalf("Set output mismatch\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestSetRefusesDuplicateKeys(t *testing.T) {
	_, err := Set("port = 3000\nport = 3001\n", "port", "3002")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSetRefusesKeySectionConflict(t *testing.T) {
	_, err := Set("[server]\nport = 3000\n", "server", "localhost")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "server is a section, not a value" {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestFullLineCommentsOnly(t *testing.T) {
	source := `
; comment
# comment
port = 3000 # dev
tag = #production
ports = 3000;4000
`

	got, err := List(source, "")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	want := []Entry{
		{Key: "port", Value: "3000 # dev"},
		{Key: "tag", Value: "#production"},
		{Key: "ports", Value: "3000;4000"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}

func TestDoubleDotEscapesLiteralDots(t *testing.T) {
	source := "server.port = 3000\n[env.prod]\nserver = web\n[env]\nprod.server = app\n"

	tests := map[string]string{
		"server..port":     "3000",
		"env..prod.server": "web",
		"env.prod..server": "app",
		"..env.server":     "",
		"env...server":     "",
	}

	for path, want := range tests {
		got, err := Get(source, path)
		if want == "" {
			if err == nil {
				t.Fatalf("Get(%q) expected error", path)
			}
			continue
		}
		if err != nil {
			t.Fatalf("Get(%q) returned error: %v", path, err)
		}
		if got != want {
			t.Fatalf("Get(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestRejectsDeepPaths(t *testing.T) {
	_, err := Get("[env]\nserver = web\n", "env.prod.server")
	if err == nil {
		t.Fatal("expected error")
	}
}
