// Copyright (c) 2026 InfraVillage
// SPDX-License-Identifier: MIT
//
// This file is part of TAGtastic and is licensed under the MIT License.

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	root, err := repoRoot()
	if err != nil {
		fatal(err)
	}

	source := filepath.Join(root, "data", "themes.yaml")
	dest := filepath.Join(root, "internal", "data", "themes.yaml")

	payload, err := os.ReadFile(source)
	if err != nil {
		fatal(err)
	}

	if err := os.WriteFile(dest, payload, 0o644); err != nil {
		fatal(err)
	}

	fmt.Printf("Synced %s -> %s\n", source, dest)
}

func repoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return wd, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
