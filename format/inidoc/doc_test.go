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
