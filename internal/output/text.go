// Copyright (c) 2026 InfraVillage
// SPDX-License-Identifier: MIT
//
// This file is part of TAGtastic and is licensed under the MIT License.

package output

import (
	"strings"

	"github.com/infravillage/tagtastic/internal/data"
)

type TextFormatter struct{}

func (TextFormatter) FormatName(item data.CodeName) (string, error) {
	return item.Name, nil
}

func (TextFormatter) FormatList(items []data.CodeName) (string, error) {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, item.Name)
	}
	return strings.Join(lines, "\n"), nil
}

func (TextFormatter) FormatThemes(names []string) (string, error) {
	return strings.Join(names, "\n"), nil
}
