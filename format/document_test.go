package format

import "testing"

func TestResolveTOML(t *testing.T) {
	doc, name, err := Resolve("config.toml")

	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if name != "toml" {
		t.Fatalf("name = %q, want toml", name)
	}
	value, err := doc.Get("title = \"demo app\"\n", "title")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if value != "demo app" {
		t.Fatalf("value = %q, want demo app", value)
	}
}

func TestResolveTOMLCaseInsensitiveExtension(t *testing.T) {
	_, name, err := Resolve("CONFIG.TOML")

	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if name != "toml" {
		t.Fatalf("name = %q, want toml", name)
	}
}

func TestResolveYAML(t *testing.T) {
	for _, path := range []string{"config.yaml", "config.yml", "CONFIG.YAML"} {
		doc, name, err := Resolve(path)

		if err != nil {
			t.Fatalf("Resolve(%q) returned error: %v", path, err)
		}
		if name != "yaml" {
			t.Fatalf("name = %q, want yaml", name)
		}
		value, err := doc.Get("server:\n  port: 3000\n", "server.port")
		if err != nil {
			t.Fatalf("Get returned error: %v", err)
		}
		if value != "3000" {
			t.Fatalf("value = %q, want 3000", value)
		}
	}
}

func TestResolveJSON(t *testing.T) {
	doc, name, err := Resolve("config.json")

	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if name != "json" {
		t.Fatalf("name = %q, want json", name)
	}
	entries, err := doc.List(`{"server":{"port":3000}}`, "server")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(entries) != 1 || entries[0].Key != "server.port" || entries[0].Value != "3000" {
		t.Fatalf("entries = %#v", entries)
	}
}

func TestResolveINI(t *testing.T) {
	doc, name, err := Resolve("config.ini")

	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if name != "ini" {
		t.Fatalf("name = %q, want ini", name)
	}
	entries, err := doc.List("title = config\n[server]\nport = 3000\n", "server")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(entries) != 1 || entries[0].Key != "server.port" || entries[0].Value != "3000" {
		t.Fatalf("entries = %#v", entries)
	}
}

func TestResolveUnsupportedFormat(t *testing.T) {
	_, _, err := Resolve("config.conf")

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTargetPath(t *testing.T) {
	tests := map[string]bool{
		"config.toml":  true,
		"CONFIG.TOML":  true,
		"config.yaml":  true,
		"config.yml":   true,
		"config.json":  true,
		"config.ini":   true,
		"config.conf":  false,
		"server.ports": false,
	}

	for path, want := range tests {
		if got := TargetPath(path); got != want {
			t.Fatalf("TargetPath(%q) = %v, want %v", path, got, want)
		}
	}
}
