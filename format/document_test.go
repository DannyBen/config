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

func TestResolveSourceUsesKnownExtension(t *testing.T) {
	_, name, err := ResolveSource("config.toml", []byte("server: localhost\n"))

	if err != nil {
		t.Fatalf("ResolveSource returned error: %v", err)
	}
	if name != "toml" {
		t.Fatalf("name = %q, want toml", name)
	}
}

func TestResolveSourceDetectsTOML(t *testing.T) {
	doc, name, err := ResolveSource("config.conf", []byte("[[servers]]\nname = \"api\"\nport = 3000\n"))

	if err != nil {
		t.Fatalf("ResolveSource returned error: %v", err)
	}
	if name != "toml" {
		t.Fatalf("name = %q, want toml", name)
	}
	value, err := doc.Get("[[servers]]\nname = \"api\"\nport = 3000\n", "servers.0.port")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if value != "3000" {
		t.Fatalf("value = %q, want 3000", value)
	}
}

func TestResolveSourceDetectsINI(t *testing.T) {
	doc, name, err := ResolveSource("config.conf", []byte("[server]\nhost = localhost\n"))

	if err != nil {
		t.Fatalf("ResolveSource returned error: %v", err)
	}
	if name != "ini" {
		t.Fatalf("name = %q, want ini", name)
	}
	value, err := doc.Get("[server]\nhost = localhost\n", "server.host")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if value != "localhost" {
		t.Fatalf("value = %q, want localhost", value)
	}
}

func TestResolveSourceDetectsJSONBeforeYAML(t *testing.T) {
	_, name, err := ResolveSource("config.conf", []byte(`{"server":{"port":3000}}`))

	if err != nil {
		t.Fatalf("ResolveSource returned error: %v", err)
	}
	if name != "json" {
		t.Fatalf("name = %q, want json", name)
	}
}

func TestResolveSourceDetectsYAML(t *testing.T) {
	doc, name, err := ResolveSource("config.conf", []byte("server:\n  port: 3000\n"))

	if err != nil {
		t.Fatalf("ResolveSource returned error: %v", err)
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

func TestResolveSourceRejectsAmbiguousTOMLAndINI(t *testing.T) {
	_, _, err := ResolveSource("config.conf", []byte("port = 3000\n"))

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "ambiguous config format for config.conf; add # format: toml or # format: ini" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveSourceUsesFormatHint(t *testing.T) {
	_, name, err := ResolveSource("config.conf", []byte("# format: ini\nport = 3000\n"))

	if err != nil {
		t.Fatalf("ResolveSource returned error: %v", err)
	}
	if name != "ini" {
		t.Fatalf("name = %q, want ini", name)
	}
}

func TestResolveSourceRejectsUnknownFormatHint(t *testing.T) {
	_, _, err := ResolveSource("config.conf", []byte("# format: xml\n<config/>\n"))

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != `unsupported format hint "xml" for config.conf` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveSourceRejectsJSONFormatHint(t *testing.T) {
	_, _, err := ResolveSource("config.conf", []byte("# format: json\n{\"port\":3000}\n"))

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != `unsupported format hint "json" for config.conf; JSON files cannot contain comments` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveSourceRejectsUnknownFormat(t *testing.T) {
	_, _, err := ResolveSource("config.conf", []byte("not config\n"))

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "cannot determine config format for config.conf; add # format: toml, # format: ini, or # format: yaml" {
		t.Fatalf("unexpected error: %v", err)
	}
}
