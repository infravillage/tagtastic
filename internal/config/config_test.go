// Copyright (c) 2026 InfraVillage
// SPDX-License-Identifier: MIT
//
// This file is part of TAGtastic and is licensed under the MIT License.

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yaml")

	payload := []byte("default_theme: birds\nused_codenames:\n  v1.0.0: Almond\n")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.DefaultTheme != "birds" {
		t.Fatalf("expected default_theme birds, got %q", cfg.DefaultTheme)
	}
	if cfg.UsedCodenames["v1.0.0"] != "Almond" {
		t.Fatalf("expected used codename Almond")
	}
}

func TestLoadConfig_ExpandsHome(t *testing.T) {
	home := t.TempDir()
	if err := os.Setenv("HOME", home); err != nil {
		t.Fatalf("set HOME: %v", err)
	}

	configDir := filepath.Join(home, ".tagtastic")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	path := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(path, []byte("default_format: json\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load("~/.tagtastic/config.yaml")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.DefaultFormat != "json" {
		t.Fatalf("expected default_format json, got %q", cfg.DefaultFormat)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	cfg, err := Load("/does/not/exist.yaml")
	if err != nil {
		t.Fatalf("expected missing file to be ignored, got %v", err)
	}
	if cfg.DefaultTheme != "" || cfg.DefaultFormat != "" {
		t.Fatalf("expected empty config on missing file")
	}
}
