// Copyright (c) 2026 InfraVillage
// SPDX-License-Identifier: MIT
//
// This file is part of TAGtastic and is licensed under the MIT License.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"golang.org/x/term"
)

type colorEntry struct {
	Color string `json:"color"`
}

type colorFile struct {
	Colors []colorEntry `json:"colors"`
}

const unreleasedTemplate = `## [Unreleased]

### Added
- N/A

### Changed
- N/A

### Fixed
- N/A
`

func main() {
	fs := flag.NewFlagSet("release", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	quietArg := hasFlag(os.Args[1:], "--quiet", "-q")
	jsonErrorsArg := hasFlag(os.Args[1:], "--json-errors")

	help := fs.Bool("help", false, "Show help")
	fs.BoolVar(help, "h", false, "Show help (shorthand)")
	codename := fs.String("codename", "", "Optional codename override")
	date := fs.String("date", "", "Release date (YYYY-MM-DD), defaults to today")
	commit := fs.Bool("commit", false, "Commit CHANGELOG.md and VERSION updates")
	quiet := fs.Bool("quiet", false, "Suppress non-essential output")
	fs.BoolVar(quiet, "q", false, "Suppress non-essential output (shorthand)")
	jsonErrors := fs.Bool("json-errors", false, "Emit errors as JSON")
	dryRun := fs.Bool("dry-run", false, "Preview changes without writing files or tagging")

	printUsage := func(showBanner bool) {
		if shouldShowBanner() && showBanner {
			fmt.Fprintln(os.Stdout, renderBanner())
			fmt.Fprintln(os.Stdout)
		}
		fmt.Fprintln(os.Stdout, "TAGtastic release helper (supporting tool)")
		fmt.Fprintln(os.Stdout, "Codename is auto-selected unless --codename is provided.")
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "Usage:")
		fmt.Fprintln(os.Stdout, "  release <version> [--codename NAME] [--date YYYY-MM-DD] [--commit]")
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "Flags:")
		fs.SetOutput(os.Stdout)
		fs.PrintDefaults()
	}

	if err := fs.Parse(os.Args[1:]); err != nil {
		reportError(err, 2, jsonErrorsArg, quietArg, func() { printUsage(!quietArg) })
	}

	quietEnabled := quietArg || *quiet
	jsonEnabled := jsonErrorsArg || *jsonErrors

	if *help {
		printUsage(!quietEnabled)
		os.Exit(0)
	}

	if fs.NArg() == 0 {
		reportError(errors.New("version is required"), 2, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
	}

	version := strings.TrimSpace(fs.Arg(0))
	if version == "" {
		reportError(errors.New("version is required"), 2, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
	}

	resolvedDate := strings.TrimSpace(*date)
	if resolvedDate == "" {
		resolvedDate = time.Now().Format("2006-01-02")
	}

	root, err := os.Getwd()
	if err != nil {
		reportError(err, 1, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
	}

	resolvedCodename := strings.TrimSpace(*codename)
	if resolvedCodename == "" {
		resolvedCodename, err = nextCodename(filepath.Join(root, "data", "crayola.json"), filepath.Join(root, "CHANGELOG.md"))
		if err != nil {
			reportError(err, 1, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
		}
	}

	if *dryRun {
		fmt.Printf("Dry run: would prepare v%s – %s\n", version, resolvedCodename)
		fmt.Printf("Dry run: would update CHANGELOG.md and VERSION (date %s)\n", resolvedDate)
		fmt.Printf("Dry run: would create tag v%s\n", version)
		return
	}

	if err := updateChangelog(filepath.Join(root, "CHANGELOG.md"), version, resolvedCodename, resolvedDate); err != nil {
		reportError(err, 1, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
	}

	if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte(version+"\n"), 0o644); err != nil {
		reportError(err, 1, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
	}

	if *commit {
		if err := commitRelease(version, resolvedCodename); err != nil {
			reportError(err, 1, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
		}
	}

	if err := createTag(version, resolvedCodename); err != nil {
		reportError(err, 1, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
	}

	fmt.Printf("Prepared release v%s – %s\n", version, resolvedCodename)
}

func shouldShowBanner() bool {
	if os.Getenv("CI") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func renderBanner() string {
	lines := []string{
		"╔════════════════════════════════════════════╗",
		"║                  TAGtastic                 ║",
		"║          Release helper (tooling)          ║",
		"║          Powered by InfraVillage™          ║",
		"╚════════════════════════════════════════════╝",
	}
	return strings.Join(lines, "\n")
}

func nextCodename(colorsPath, changelogPath string) (string, error) {
	payload, err := os.ReadFile(colorsPath)
	if err != nil {
		return "", err
	}

	var file colorFile
	if err := json.Unmarshal(payload, &file); err != nil {
		return "", err
	}

	used := make(map[string]struct{})
	if changelog, err := os.ReadFile(changelogPath); err == nil {
		re := regexp.MustCompile(`–\s+"([^"]+)"`)
		for _, match := range re.FindAllStringSubmatch(string(changelog), -1) {
			if len(match) > 1 {
				used[strings.TrimSpace(match[1])] = struct{}{}
			}
		}
	}

	for _, entry := range file.Colors {
		name := strings.TrimSpace(entry.Color)
		if name == "" {
			continue
		}
		if _, ok := used[name]; ok {
			continue
		}
		return name, nil
	}

	return "", errors.New("no available codenames left")
}

func updateChangelog(path, version, codename, date string) error {
	payload, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(payload)
	sections := splitChangelog(content)
	if sections.unreleased == "" {
		return errors.New("missing [Unreleased] section in CHANGELOG.md")
	}

	releaseHeader := fmt.Sprintf("## [%s] – \"%s\" – %s", version, codename, date)
	newRelease := strings.TrimSpace(sections.unreleased)
	if newRelease == "" || newRelease == "### Added\n- N/A\n\n### Changed\n- N/A\n\n### Fixed\n- N/A" {
		newRelease = "### Added\n- Placeholder version entry for this release."
	}

	releaseBlock := fmt.Sprintf("%s\n\n%s\n\n", releaseHeader, newRelease)

	updated := sections.preamble + unreleasedTemplate + "\n" + releaseBlock + sections.remainder
	updated = updateChangelogLinks(updated, version)

	return os.WriteFile(path, []byte(updated), 0o644)
}

type changelogSections struct {
	preamble   string
	unreleased string
	remainder  string
}

func splitChangelog(content string) changelogSections {
	lines := strings.Split(content, "\n")
	state := "preamble"
	var preamble, unreleased, remainder []string

	for _, line := range lines {
		if strings.HasPrefix(line, "## [Unreleased]") {
			state = "unreleased"
			continue
		}
		if strings.HasPrefix(line, "## [") && state == "unreleased" {
			state = "remainder"
		}

		switch state {
		case "preamble":
			preamble = append(preamble, line)
		case "unreleased":
			unreleased = append(unreleased, line)
		case "remainder":
			remainder = append(remainder, line)
		}
	}

	return changelogSections{
		preamble:   strings.Join(preamble, "\n"),
		unreleased: strings.TrimSpace(strings.Join(unreleased, "\n")),
		remainder:  strings.Join(remainder, "\n"),
	}
}

func updateChangelogLinks(content, version string) string {
	lines := strings.Split(content, "\n")
	var output []string
	var references []string
	inRefs := false

	for _, line := range lines {
		if strings.HasPrefix(line, "[") && strings.Contains(line, "]: ") {
			inRefs = true
			references = append(references, line)
			continue
		}
		if inRefs {
			if strings.TrimSpace(line) == "" {
				continue
			}
			if strings.HasPrefix(line, "#") {
				inRefs = false
			}
		}
		output = append(output, line)
	}

	updatedRefs := updateReferenceLines(references, version)
	if len(updatedRefs) > 0 {
		output = append(output, "")
		output = append(output, updatedRefs...)
	}

	return strings.Join(output, "\n")
}

func updateReferenceLines(lines []string, version string) []string {
	var refs []string
	var releases []string

	for _, line := range lines {
		if strings.HasPrefix(line, "[") {
			releases = append(releases, line)
			continue
		}
	}

	prevVersion := ""
	if len(releases) > 0 {
		sort.SliceStable(releases, func(i, j int) bool {
			return releases[i] > releases[j]
		})
		prevVersion = extractVersion(releases[0])
	}

	if prevVersion != "" {
		releases = append([]string{fmt.Sprintf("[%s]: https://github.com/infravillage/tagtastic/compare/v%s...v%s", version, prevVersion, version)}, releases...)
	} else {
		releases = append([]string{fmt.Sprintf("[%s]: https://github.com/infravillage/tagtastic/releases/tag/v%s", version, version)}, releases...)
	}

	unreleased := fmt.Sprintf("[Unreleased]: https://github.com/infravillage/tagtastic/compare/v%s...HEAD", version)

	refs = append(refs, unreleased)
	refs = append(refs, releases...)
	return refs
}

func extractVersion(line string) string {
	start := strings.Index(line, "[")
	end := strings.Index(line, "]")
	if start == -1 || end == -1 || end <= start+1 {
		return ""
	}
	return line[start+1 : end]
}

func createTag(version, codename string) error {
	if _, err := os.Stat(".git"); err != nil {
		return errors.New("git repository not found")
	}
	message := fmt.Sprintf("v%s – %s", version, codename)
	cmd := exec.Command("git", "tag", "-a", "v"+version, "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func commitRelease(version, codename string) error {
	cmd := exec.Command("git", "add", "CHANGELOG.md", "VERSION")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	message := fmt.Sprintf("chore: prepare release v%s (%s)", version, codename)
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr
	return commitCmd.Run()
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

func reportError(err error, code int, jsonErrors, quiet bool, usage func()) {
	if jsonErrors {
		fmt.Fprintf(os.Stderr, "{\"error\":%q,\"code\":%d}\n", err.Error(), code)
		os.Exit(code)
	}
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	if !quiet {
		usage()
	}
	os.Exit(code)
}
