// Copyright (c) 2026 InfraVillage
// SPDX-License-Identifier: MIT
//
// This file is part of TAGtastic and is licensed under the MIT License.

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
	if payload != "RELEASE_CODENAME=almond" {
		t.Fatalf("unexpected shell payload: %q", payload)
	}
}

func TestNewFormatter_Unknown(t *testing.T) {
	if _, err := NewFormatter("nope"); err == nil {
		t.Fatalf("expected error for unknown format")
	}
}

func TestShellFormatter_FallbackToSlug(t *testing.T) {
	formatter := ShellFormatter{}
	item := data.CodeName{Name: "Blue Heron"}

	payload, err := formatter.FormatName(item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload != "RELEASE_CODENAME=blue-heron" {
		t.Fatalf("unexpected shell payload: %q", payload)
	}
}

func TestJSONFormatter_ListAndThemes(t *testing.T) {
	formatter := JSONFormatter{}
	items := []data.CodeName{
		{Name: "Almond"},
		{Name: "Antique Brass"},
	}

	listPayload, err := formatter.FormatList(items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var listDecoded []map[string]any
	if err := json.Unmarshal([]byte(listPayload), &listDecoded); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if len(listDecoded) != 2 {
		t.Fatalf("expected 2 list items")
	}

	themesPayload, err := formatter.FormatThemes([]string{"birds", "crayola_colors"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var themesDecoded []string
	if err := json.Unmarshal([]byte(themesPayload), &themesDecoded); err != nil {
		t.Fatalf("unmarshal themes: %v", err)
	}
	if len(themesDecoded) != 2 || themesDecoded[0] != "birds" {
		t.Fatalf("unexpected themes output")
	}
}
