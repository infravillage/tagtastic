// Copyright (c) 2026 InfraVillage
// SPDX-License-Identifier: MIT
//
// This file is part of TAGtastic and is licensed under the MIT License.

package output

import (
	"errors"
	"strings"

	"github.com/infravillage/tagtastic/internal/data"
)

var ErrUnknownFormat = errors.New("unknown format")

type Formatter interface {
	FormatName(item data.CodeName) (string, error)
	FormatList(items []data.CodeName) (string, error)
	FormatThemes(names []string) (string, error)
}

func NewFormatter(format string) (Formatter, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "text":
		return TextFormatter{}, nil
	case "json":
		return JSONFormatter{}, nil
	case "shell":
		return ShellFormatter{}, nil
	default:
		return nil, ErrUnknownFormat
	}
}
