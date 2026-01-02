// Copyright (c) 2026 InfraVillage
// SPDX-License-Identifier: MIT
//
// This file is part of TAGtastic and is licensed under the MIT License.

package data

import (
	"embed"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed themes.yaml
var embeddedThemes embed.FS

var ErrThemeNotFound = errors.New("theme not found")

var normalizePattern = regexp.MustCompile(`[^a-z0-9]+`)

type ThemeRepository interface {
	GetThemeByName(name string) (*Theme, error)
	GetAllThemeNames() []string
}

type EmbeddedThemeRepository struct {
	themes map[string]*Theme
	names  []string
}

type themeFile struct {
	Version string            `yaml:"version"`
	Themes  map[string]*Theme `yaml:"themes"`
}

func NewEmbeddedThemeRepository() (*EmbeddedThemeRepository, error) {
	payload, err := embeddedThemes.ReadFile("themes.yaml")
	if err != nil {
		return nil, fmt.Errorf("read embedded themes: %w", err)
	}

	var file themeFile
	if err := yaml.Unmarshal(payload, &file); err != nil {
		return nil, fmt.Errorf("parse themes: %w", err)
	}

	repo := &EmbeddedThemeRepository{
		themes: make(map[string]*Theme),
	}

	for key, theme := range file.Themes {
		if theme == nil {
			continue
		}
		if theme.ID == "" {
			theme.ID = key
		}
		normalized := normalizeName(key)
		repo.themes[normalized] = theme
		repo.names = append(repo.names, theme.ID)
	}

	sort.Strings(repo.names)
	return repo, nil
}

func (r *EmbeddedThemeRepository) GetThemeByName(name string) (*Theme, error) {
	if name == "" {
		return nil, fmt.Errorf("theme name is required")
	}
	if theme, ok := r.themes[normalizeName(name)]; ok {
		return theme, nil
	}
	return nil, ErrThemeNotFound
}

func (r *EmbeddedThemeRepository) GetAllThemeNames() []string {
	return append([]string(nil), r.names...)
}

func FilterItems(items []CodeName, exclude []string) []CodeName {
	if len(exclude) == 0 {
		return append([]CodeName(nil), items...)
	}

	deny := make(map[string]struct{}, len(exclude))
	for _, raw := range exclude {
		key := normalizeName(raw)
		if key != "" {
			deny[key] = struct{}{}
		}
	}

	filtered := make([]CodeName, 0, len(items))
	for _, item := range items {
		if shouldExclude(item, deny) {
			continue
		}
		filtered = append(filtered, item)
	}

	return filtered
}

func shouldExclude(item CodeName, deny map[string]struct{}) bool {
	candidates := append([]string{item.Name}, item.Aliases...)
	for _, candidate := range candidates {
		if _, ok := deny[normalizeName(candidate)]; ok {
			return true
		}
	}
	return false
}

func normalizeName(input string) string {
	trimmed := strings.TrimSpace(strings.ToLower(input))
	if trimmed == "" {
		return ""
	}
	normalized := normalizePattern.ReplaceAllString(trimmed, "-")
	return strings.Trim(normalized, "-")
}

// NormalizeName exposes the internal normalization for CLI matching.
func NormalizeName(input string) string {
	return normalizeName(input)
}
