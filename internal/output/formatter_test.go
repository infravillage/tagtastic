package output

import (
	"encoding/json"
	"testing"

	"github.com/infravillage/tagtastic/internal/data"
)

func TestTextFormatter(t *testing.T) {
	formatter := TextFormatter{}
	item := data.CodeName{Name: "Almond"}

	name, err := formatter.FormatName(item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Almond" {
		t.Fatalf("expected Almond, got %q", name)
	}
}

func TestJSONFormatter(t *testing.T) {
	formatter := JSONFormatter{}
	item := data.CodeName{Name: "Almond", Aliases: []string{"almond"}, Description: "Hex #EFDECD"}

	payload, err := formatter.FormatName(item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded["name"] != "Almond" {
		t.Fatalf("expected name Almond")
	}
}

func TestShellFormatter(t *testing.T) {
	formatter := ShellFormatter{}
	item := data.CodeName{Name: "Almond"}

	payload, err := formatter.FormatName(item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload != "RELEASE_CODENAME=\"Almond\"" {
		t.Fatalf("unexpected shell payload: %q", payload)
	}
}
