// Copyright (c) 2026 InfraVillage
// SPDX-License-Identifier: MIT
//
// This file is part of TAGtastic and is licensed under the MIT License.

package output

import (
	"fmt"
	"strings"

	"github.com/infravillage/tagtastic/internal/data"
)

type ShellFormatter struct{}

func (ShellFormatter) FormatName(item data.CodeName) (string, error) {
	value := aliasOrSlug(item)
	return fmt.Sprintf("RELEASE_CODENAME=%s", value), nil
}

func (ShellFormatter) FormatList(items []data.CodeName) (string, error) {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, item.Name)
	}
	return strings.Join(lines, "\n"), nil
}

func (ShellFormatter) FormatThemes(names []string) (string, error) {
	return strings.Join(names, "\n"), nil
}

func aliasOrSlug(item data.CodeName) string {
	if len(item.Aliases) > 0 {
		alias := strings.TrimSpace(item.Aliases[0])
		if alias != "" {
			return alias
		}
	}

	slug := data.NormalizeName(item.Name)
	if slug != "" {
		return slug
	}

	return "unknown"
}
