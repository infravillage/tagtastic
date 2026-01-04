// Copyright (c) 2026 InfraVillage
// SPDX-License-Identifier: MIT
//
// This file is part of TAGtastic and is licensed under the MIT License.
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateRepoConfigCreatesFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".tagtastic.yaml")

	if err := updateRepoConfig(path, "Almond", "0.1.0-beta.1"); err != nil {
		t.Fatalf("updateRepoConfig failed: %v", err)
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	content := string(payload)

	if !strings.Contains(content, "default_theme: crayola_colors") {
		t.Fatalf("expected default_theme to be set")
	}
	if !strings.Contains(content, "0.1.0-beta.1: Almond") {
		t.Fatalf("expected codename to be recorded")
	}
}

func TestResolveConfigPathOverride(t *testing.T) {
	tmp := t.TempDir()
	override := filepath.Join(tmp, "custom.yaml")

	path, err := resolveConfigPath(tmp, override)
	if err != nil {
		t.Fatalf("resolveConfigPath failed: %v", err)
	}
	if path == "" || !strings.HasSuffix(path, "custom.yaml") {
		t.Fatalf("expected override path, got %q", path)
	}
}

func TestEnsureSemVerForward(t *testing.T) {
	if err := ensureSemVerForward("0.1.1", "0.1.0"); err != nil {
		t.Fatalf("expected forward version, got error: %v", err)
	}
	if err := ensureSemVerForward("0.1.0", "0.1.0"); err == nil {
		t.Fatalf("expected error for non-forward version")
	}
}

func TestBumpVersion(t *testing.T) {
	version, err := bumpVersion("0.1.0-beta.2", "patch")
	if err != nil {
		t.Fatalf("bumpVersion failed: %v", err)
	}
	if version != "0.1.1" {
		t.Fatalf("expected 0.1.1, got %q", version)
	}
}
