package yamldoc

import (
	"fmt"
	"strings"
	"testing"
)

func TestListFlattensMappings(t *testing.T) {
	source := "title: demo app\nserver:\n  port: 3000\n  enabled: true\n"

	got, err := List(source, "")

	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	want := []Entry{
		{Key: "title", Value: "demo app"},
		{Key: "server.port", Value: "3000"},
		{Key: "server.enabled", Value: "true"},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}

func TestListPrefixAndScalar(t *testing.T) {
	source := "server:\n  port: 3000\n  enabled: true\n"

	prefix, err := List(source, "server")
	if err != nil {
		t.Fatalf("List prefix returned error: %v", err)
	}
	wantPrefix := []Entry{
		{Key: "server.port", Value: "3000"},
		{Key: "server.enabled", Value: "true"},
	}
	if fmt.Sprint(prefix) != fmt.Sprint(wantPrefix) {
		t.Fatalf("List prefix = %#v, want %#v", prefix, wantPrefix)
	}

	scalar, err := List(source, "server.port")
	if err != nil {
		t.Fatalf("List scalar returned error: %v", err)
	}
	wantScalar := []Entry{{Key: "server.port", Value: "3000"}}
	if fmt.Sprint(scalar) != fmt.Sprint(wantScalar) {
		t.Fatalf("List scalar = %#v, want %#v", scalar, wantScalar)
	}
}

func TestListRejectsEmptyPathSegments(t *testing.T) {
	_, err := List("server:\n  port: 3000\n", "server.")

	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != `empty path segment in "server."` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListDoubleDotEscapesLiteralDotInPath(t *testing.T) {
	source := "server:\n  public.port: 3000\n"

	got, err := List(source, "server.public..port")

	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	want := []Entry{{Key: "server.public..port", Value: "3000"}}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}

func TestListUsesIndexedPathsForSequences(t *testing.T) {
	source := "tags:\n  - api\n  - worker\nservers:\n  - name: api\n    port: 3000\n  - name: worker\n    port: 3001\n"

	got, err := List(source, "")

	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	want := []Entry{
		{Key: "tags.0", Value: "api"},
		{Key: "tags.1", Value: "worker"},
		{Key: "servers.0.name", Value: "api"},
		{Key: "servers.0.port", Value: "3000"},
		{Key: "servers.1.name", Value: "worker"},
		{Key: "servers.1.port", Value: "3001"},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}

func TestListSequencePrefix(t *testing.T) {
	source := "servers:\n  - name: api\n    port: 3000\n  - name: worker\n    port: 3001\n"

	got, err := List(source, "servers.1")

	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	want := []Entry{
		{Key: "servers.1.name", Value: "worker"},
		{Key: "servers.1.port", Value: "3001"},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}

func TestListDistinguishesMissingFromExistingEmptyContainers(t *testing.T) {
	got, err := List("empty_map: {}\nempty_list: []\n", "empty_map")
	if err != nil {
		t.Fatalf("List existing empty map returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("List existing empty map = %#v, want no entries", got)
	}

	got, err = List("empty_map: {}\nempty_list: []\n", "empty_list")
	if err != nil {
		t.Fatalf("List existing empty list returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("List existing empty list = %#v, want no entries", got)
	}

	_, err = List("empty_map: {}\nempty_list: []\n", "missing")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if err.Error() != "missing is not set" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListResolvesAliases(t *testing.T) {
	source := "defaults: &defaults\n  host: localhost\n  port: 5432\nprimary: *defaults\nhost: &host localhost\nreplica: *host\n"

	got, err := List(source, "primary")

	if err != nil {
		t.Fatalf("List alias mapping returned error: %v", err)
	}
	want := []Entry{
		{Key: "primary.host", Value: "localhost"},
		{Key: "primary.port", Value: "5432"},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("List alias mapping = %#v, want %#v", got, want)
	}

	scalar, err := List(source, "replica")
	if err != nil {
		t.Fatalf("List alias scalar returned error: %v", err)
	}
	wantScalar := []Entry{{Key: "replica", Value: "localhost"}}
	if fmt.Sprint(scalar) != fmt.Sprint(wantScalar) {
		t.Fatalf("List alias scalar = %#v, want %#v", scalar, wantScalar)
	}
}

func TestListRejectsDuplicateKeys(t *testing.T) {
	_, err := List("server:\n  port: 3000\n  port: 3001\n", "")

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `duplicate mapping key "port"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
