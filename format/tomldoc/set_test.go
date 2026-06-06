package tomldoc

import (
	"fmt"
	"testing"
)

func TestSetExistingScalarPreservesInlineComment(t *testing.T) {
	source := "[database]\nport = 5432 # default postgres port\n"

	got, err := Set(source, "database.port", "3000")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "[database]\nport = 3000 # default postgres port\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetMissingScalarInExistingTable(t *testing.T) {
	source := "[server]\nbind = \"0.0.0.0\"\n"

	got, err := Set(source, "server.port", "3000")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "[server]\nbind = \"0.0.0.0\"\nport = 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetMissingScalarInExistingCRLFTable(t *testing.T) {
	source := "[server]\r\nbind = \"0.0.0.0\"\r\n"

	got, err := Set(source, "server.port", "3000")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "[server]\r\nbind = \"0.0.0.0\"\r\nport = 3000\r\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%q\ngot:\n%q", want, got)
	}
}

func TestSetMissingDottedKey(t *testing.T) {
	source := "title = \"demo app\"\n"

	got, err := Set(source, "server.port", "3000")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "title = \"demo app\"\n\n[server]\nport = 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSetExistingScalarWithoutFinalNewline(t *testing.T) {
	source := "port = 3000"

	got, err := Set(source, "port", "3001")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "port = 3001"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%q\ngot:\n%q", want, got)
	}
}

func TestSetMissingScalarWithoutFinalNewlineMayAddNewline(t *testing.T) {
	source := "[server]\nbind = \"0.0.0.0\""

	got, err := Set(source, "server.port", "3000")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "[server]\nbind = \"0.0.0.0\"\nport = 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%q\ngot:\n%q", want, got)
	}
}

func TestSetMissingRootScalarWithoutFinalNewlineMayAddNewline(t *testing.T) {
	source := "title = \"demo app\""

	got, err := Set(source, "port", "3000")

	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	want := "title = \"demo app\"\nport = 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%q\ngot:\n%q", want, got)
	}
}

func TestSetTableAsScalarFails(t *testing.T) {
	source := "[database]\nport = 5432\n"

	_, err := Set(source, "database", "3000")

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnsetExistingScalar(t *testing.T) {
	source := "[database]\nhost = \"localhost\"\nport = 5432 # default postgres port\n"

	got, err := Unset(source, "database.port")

	if err != nil {
		t.Fatalf("Unset returned error: %v", err)
	}
	want := "[database]\nhost = \"localhost\"\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestUnsetAfterSetInAnotherTable(t *testing.T) {
	source := "title = \"demo app\"\n\n[server]\nbind = \"0.0.0.0\"\n\n[database]\nhost = \"localhost\"\nport = 5432 # default postgres port\npassword = \"secret\"\n"

	updated, err := Set(source, "server.port", "3000")
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	got, err := Unset(updated, "database.password")
	if err != nil {
		t.Fatalf("Unset returned error: %v\nsource:\n%s", err, updated)
	}
	want := "title = \"demo app\"\n\n[server]\nbind = \"0.0.0.0\"\nport = 3000\n\n[database]\nhost = \"localhost\"\nport = 5432 # default postgres port\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestUnsetMissingScalarFails(t *testing.T) {
	source := "[database]\nport = 5432\n"

	_, err := Unset(source, "database.host")

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnsetTableFails(t *testing.T) {
	source := "[database]\nport = 5432\n"

	_, err := Unset(source, "database")

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnsetFinalScalarWithoutFinalNewline(t *testing.T) {
	source := "[database]\nhost = \"localhost\"\nport = 5432"

	got, err := Unset(source, "database.port")

	if err != nil {
		t.Fatalf("Unset returned error: %v", err)
	}
	want := "[database]\nhost = \"localhost\"\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%q\ngot:\n%q", want, got)
	}
}

func TestDeleteTable(t *testing.T) {
	source := "title = \"demo app\"\n\n[server]\nport = 3000\n\n[style]\ncolor = \"blue\"\nfont = \"arial\"\n"

	got, err := Delete(source, "style", nil)

	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	want := "title = \"demo app\"\n\n[server]\nport = 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestDeleteFinalTableWithoutFinalNewline(t *testing.T) {
	source := "title = \"demo app\"\n\n[style]\ncolor = \"blue\""

	got, err := Delete(source, "style", nil)

	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	want := "title = \"demo app\"\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%q\ngot:\n%q", want, got)
	}
}

func TestDeleteTableBeforeAnotherTable(t *testing.T) {
	source := "title = \"demo app\"\n\n[style]\ncolor = \"blue\"\nfont = \"arial\"\n\n[server]\nport = 3000\n"

	got, err := Delete(source, "style", nil)

	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	want := "title = \"demo app\"\n\n[server]\nport = 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestDeleteArrayRecordByIndex(t *testing.T) {
	source := "title = \"demo app\"\n\n[[servers]]\nname = \"api\"\nport = 3000\n\n[[servers]]\nname = \"worker\"\nport = 3001\n\n[style]\ncolor = \"blue\"\n"

	got, err := Delete(source, "servers.0", nil)

	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	want := "title = \"demo app\"\n\n[[servers]]\nname = \"worker\"\nport = 3001\n\n[style]\ncolor = \"blue\"\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestDeleteArrayRecordBySelector(t *testing.T) {
	source := "title = \"demo app\"\n\n[[servers]]\nname = \"api\"\nhost = \"api.local\"\nport = 3000\n\n[[servers]]\nname = \"worker\"\nhost = \"worker.local\"\nport = 3001\n"

	got, err := Delete(source, "servers", []string{"name:api"})

	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	want := "title = \"demo app\"\n\n[[servers]]\nname = \"worker\"\nhost = \"worker.local\"\nport = 3001\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestDeleteArrayRecordByMultipleSelectors(t *testing.T) {
	source := "[[servers]]\nname = \"api\"\nhost = \"api.local\"\nport = 3000\n\n[[servers]]\nname = \"api\"\nhost = \"backup.local\"\nport = 3001\n"

	got, err := Delete(source, "servers", []string{"name:api", "host:backup.local"})

	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	want := "[[servers]]\nname = \"api\"\nhost = \"api.local\"\nport = 3000\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestDeleteArrayRecordRefusesMissingSelectorMatch(t *testing.T) {
	source := "[[servers]]\nname = \"api\"\nport = 3000\n"

	_, err := Delete(source, "servers", []string{"name:web"})

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "servers has no records matching name:web" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteArrayCollection(t *testing.T) {
	source := "title = \"demo\"\n\n[[servers]]\nname = \"api\"\nport = 3000\n\n[[servers]]\nname = \"worker\"\nport = 3001\n\n[style]\ncolor = \"blue\"\n"

	got, err := Delete(source, "servers", nil)

	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	want := "title = \"demo\"\n\n[style]\ncolor = \"blue\"\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestDeleteArrayCollectionWithInterleavedTable(t *testing.T) {
	source := "[[servers]]\nname = \"api\"\n\n[style]\ncolor = \"blue\"\n\n[[servers]]\nname = \"worker\"\n\n[database]\nport = 5432\n"

	got, err := Delete(source, "servers", nil)

	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	want := "[style]\ncolor = \"blue\"\n\n[database]\nport = 5432\n"
	if got != want {
		t.Fatalf("updated source mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestDeleteArrayRecordIndexOutOfRange(t *testing.T) {
	source := "[[servers]]\nname = \"api\"\nport = 3000\n"

	_, err := Delete(source, "servers.2", nil)

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "servers has no record at index 2" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteArrayRecordRefusesMultipleSelectorMatches(t *testing.T) {
	source := "[[servers]]\nname = \"api\"\nport = 3000\n\n[[servers]]\nname = \"api\"\nport = 3001\n"

	_, err := Delete(source, "servers", []string{"name:api"})

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "servers has multiple records matching name:api" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteRefusesScalarValue(t *testing.T) {
	source := "[style]\ncolor = \"blue\"\nfont = \"arial\"\n"

	_, err := Delete(source, "style.color", nil)

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "style.color is a value, use unset to remove fields" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteSelectorRequiresColon(t *testing.T) {
	source := "[[servers]]\nname = \"api\"\nport = 3000\n"

	_, err := Delete(source, "servers", []string{"name=api"})

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "--on must use FIELD:VALUE" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetExistingScalar(t *testing.T) {
	source := "[database]\nport = 5432 # default postgres port\n"

	got, err := Get(source, "database.port")

	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got != "5432" {
		t.Fatalf("Get = %q, want 5432", got)
	}
}

func TestGetStringUnquotesTOMLRepresentation(t *testing.T) {
	source := "[database]\nhost = \"localhost\"\n"

	got, err := Get(source, "database.host")

	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got != "localhost" {
		t.Fatalf("Get = %q, want unquoted string", got)
	}
}

func TestGetMissingScalarFails(t *testing.T) {
	source := "[database]\nport = 5432\n"

	_, err := Get(source, "database.host")

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetTableFails(t *testing.T) {
	source := "[database]\nport = 5432\n"

	_, err := Get(source, "database")

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListValues(t *testing.T) {
	source := "title = \"demo app\"\n\n[server]\nport = 3000\nenabled = true\n\n[database]\nhost = \"localhost\"\n"

	got, err := List(source, "")

	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	want := []Entry{
		{Key: "title", Value: "demo app"},
		{Key: "server.port", Value: "3000"},
		{Key: "server.enabled", Value: "true"},
		{Key: "database.host", Value: "localhost"},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}

func TestListValuesUnderTable(t *testing.T) {
	source := "title = \"demo app\"\n\n[server]\nport = 3000\nenabled = true\n\n[database]\nhost = \"localhost\"\n"

	got, err := List(source, "server")

	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	want := []Entry{
		{Key: "server.port", Value: "3000"},
		{Key: "server.enabled", Value: "true"},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}

func TestListScalarReturnsSingleEntry(t *testing.T) {
	source := "[server]\nport = 3000\n"

	got, err := List(source, "server.port")

	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	want := []Entry{{Key: "server.port", Value: "3000"}}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}

func TestListMissingPathFails(t *testing.T) {
	source := "[server]\nport = 3000\n"

	_, err := List(source, "database")

	if err == nil {
		t.Fatal("expected error")
	}
}
