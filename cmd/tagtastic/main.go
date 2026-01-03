// Copyright (c) 2026 InfraVillage
// SPDX-License-Identifier: MIT
//
// This file is part of TAGtastic and is licensed under the MIT License.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/infravillage/tagtastic/internal/cli"
	"github.com/infravillage/tagtastic/internal/config"
	"github.com/infravillage/tagtastic/internal/data"
	"github.com/infravillage/tagtastic/internal/output"
	"golang.org/x/mod/semver"
	"golang.org/x/term"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	repo, err := data.NewEmbeddedThemeRepository()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load themes: %v\n", err)
		os.Exit(1)
	}

	deps := cli.Dependencies{
		Themes:           repo,
		FormatterFactory: output.NewFormatter,
		Out:              os.Stdout,
		VersionInfo: cli.VersionInfo{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
	}

	app := cli.NewCLI(deps)
	parser, err := kong.New(app, kong.Name("tagtastic"), kong.Description("Generate human-readable release codenames."))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize CLI: %v\n", err)
		os.Exit(1)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		if shouldShowBanner(args) {
			fmt.Fprintln(os.Stdout, renderBanner())
			fmt.Fprintln(os.Stdout)
			ctx, err := parser.Parse([]string{"--help"})
			parser.FatalIfErrorf(err)
			parser.FatalIfErrorf(ctx.Run())
			return
		}

		fmt.Fprintln(os.Stderr, "no command provided")
		os.Exit(1)
	}

	if shouldShowBanner(args) && isHelpRequest(args) {
		fmt.Fprintln(os.Stdout, renderBanner())
		fmt.Fprintln(os.Stdout)
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		reportParseError(err, wantsJSONErrors(args))
	}

	if err := ctx.Run(); err != nil {
		reportRunError(err, wantsJSONErrors(args))
	}
}

func shouldShowBanner(args []string) bool {
	if hasFlag(args, "--quiet", "-q") {
		return false
	}
	if os.Getenv("CI") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func isHelpRequest(args []string) bool {
	for _, arg := range args {
		trimmed := strings.TrimSpace(arg)
		switch trimmed {
		case "-h", "--help", "help":
			return true
		}
	}
	return false
}

func wantsJSONErrors(args []string) bool {
	return hasFlag(args, "--json-errors")
}

func hasFlag(args []string, names ...string) bool {
	for _, arg := range args {
		for _, name := range names {
			if arg == name {
				return true
			}
		}
	}
	return false
}

func renderBanner() string {
	info := cli.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
	codename := latestCodename()
	displayVersion := resolveVersion(info.Version)

	lines := []string{
		"╔════════════════════════════════════════════╗",
		"║                  TAGtastic                 ║",
		"║        Release codenames for CI/CD         ║",
		"║          Powered by InfraVillage™          ║",
		"╚════════════════════════════════════════════╝",
		fmt.Sprintf("version: %s", displayVersion),
		fmt.Sprintf("last release codename: %s", defaultValue(codename, "none")),
	}

	return strings.Join(lines, "\n")
}

func latestCodename() string {
	if codename := latestCodenameFromTags(); codename != "" {
		return codename
	}

	cfg, err := config.Load("")
	if err == nil && len(cfg.UsedCodenames) > 0 {
		if codename := latestCodenameFromConfig(cfg.UsedCodenames); codename != "" {
			return codename
		}
	}

	payload, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(payload), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "## [") {
			continue
		}
		if strings.Contains(line, "[Unreleased]") {
			continue
		}
		if idx := strings.Index(line, "– \""); idx != -1 {
			fragment := line[idx+len("– \""):]
			if end := strings.Index(fragment, "\""); end != -1 {
				return fragment[:end]
			}
		}
	}

	return ""
}

func defaultValue(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func latestCodenameFromConfig(values map[string]string) string {
	latest := ""
	latestKey := ""
	for key, value := range values {
		clean := normalizeSemver(key)
		if clean == "" {
			continue
		}
		if latestKey == "" || semver.Compare(clean, latestKey) > 0 {
			latestKey = clean
			latest = value
		}
	}

	return strings.TrimSpace(latest)
}

func latestCodenameFromTags() string {
	if _, err := os.Stat(".git"); err != nil {
		return ""
	}

	output, err := exec.Command("git", "tag", "-l", "v*").Output()
	if err != nil {
		return ""
	}

	latest := ""
	for _, raw := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		tag := strings.TrimSpace(raw)
		if tag == "" {
			continue
		}
		normalized := normalizeSemver(tag)
		if normalized == "" {
			continue
		}
		if latest == "" || semver.Compare(normalized, latest) > 0 {
			latest = normalized
		}
	}

	if latest == "" {
		return ""
	}

	message, err := exec.Command("git", "tag", "-l", latest, "--format=%(contents:subject)").Output()
	if err != nil {
		return ""
	}
	return extractCodename(strings.TrimSpace(string(message)))
}

func extractCodename(message string) string {
	if message == "" {
		return ""
	}
	if idx := strings.Index(message, "– "); idx != -1 {
		return strings.TrimSpace(message[idx+len("– "):])
	}
	if idx := strings.Index(message, "- "); idx != -1 {
		return strings.TrimSpace(message[idx+len("- "):])
	}
	return ""
}

func normalizeSemver(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if !strings.HasPrefix(trimmed, "v") {
		trimmed = "v" + trimmed
	}
	if !semver.IsValid(trimmed) {
		return ""
	}
	return trimmed
}

func resolveVersion(value string) string {
	if strings.TrimSpace(value) == "" {
		return "dev"
	}
	if value != "dev" {
		return value
	}

	payload, err := os.ReadFile("VERSION")
	if err != nil {
		return value
	}

	clean := strings.TrimSpace(string(payload))
	if clean == "" {
		return value
	}

	return fmt.Sprintf("%s (dev)", clean)
}

func reportParseError(err error, jsonErrors bool) {
	if jsonErrors {
		fmt.Fprintf(os.Stderr, "{\"error\":%q,\"type\":\"parse\",\"code\":2}\n", err.Error())
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	fmt.Fprintln(os.Stderr, "Run \"tagtastic --help\" for usage.")
	os.Exit(2)
}

func reportRunError(err error, jsonErrors bool) {
	if jsonErrors {
		fmt.Fprintf(os.Stderr, "{\"error\":%q,\"type\":\"runtime\",\"code\":1}\n", err.Error())
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
