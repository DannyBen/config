package tomldoc

import "testing"

func TestSetUsesTOMLLiteralWhenValueParses(t *testing.T) {
	source := "port = 3000\n"

	got, err := Set(source, "release.date", "2027-03-24")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "port = 3000\n\n[release]\ndate = 2027-03-24\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetFallsBackToStringWhenValueIsNotTOMLLiteral(t *testing.T) {
	source := "host = \"old\"\n"

	got, err := Set(source, "host", "localhost")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "host = \"localhost\"\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetStoresArrayLookingValueAsString(t *testing.T) {
	source := "ports = []\n"

	got, err := Set(source, "ports", "[3000, 4000]")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "ports = \"[3000, 4000]\"\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetStoresInlineTableLookingValueAsString(t *testing.T) {
	source := "title = \"demo\"\n"

	got, err := Set(source, "pool", "{ min = 1, max = 10 }")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "title = \"demo\"\npool = \"{ min = 1, max = 10 }\"\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetStringForcesString(t *testing.T) {
	source := "version = 1.0\n"

	got, err := SetString(source, "version", "1.0")

	if err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}
	want := "version = \"1.0\"\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetArrayElementsUseTOMLLiteralParsing(t *testing.T) {
	source := "values = []\n"

	got, err := SetArray(source, "values", []string{"2027-03-24", "localhost"})

	if err != nil {
		t.Fatalf("SetArray returned error: %v", err)
	}
	want := "values = [2027-03-24, \"localhost\"]\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}
