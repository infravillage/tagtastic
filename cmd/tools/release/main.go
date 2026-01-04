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
	"strconv"
	"strings"
	"time"

	"github.com/infravillage/tagtastic/internal/config"
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

	parsedArgs := reorderArgs(os.Args[1:])
	quietArg := hasFlag(parsedArgs, "--quiet", "-q")
	jsonErrorsArg := hasFlag(parsedArgs, "--json-errors")

	help := fs.Bool("help", false, "Show help")
	fs.BoolVar(help, "h", false, "Show help (shorthand)")
	codename := fs.String("codename", "", "Optional codename override")
	date := fs.String("date", "", "Release date (YYYY-MM-DD), defaults to today")
	bump := fs.String("bump", "", "Auto-bump version (major, minor, patch)")
	pre := fs.String("pre", "", "Prerelease label (alpha, beta, rc)")
	preNum := fs.Int("pre-num", 0, "Prerelease number (defaults to next available)")
	commit := fs.Bool("commit", false, "Commit CHANGELOG.md and VERSION updates")
	quiet := fs.Bool("quiet", false, "Suppress non-essential output")
	fs.BoolVar(quiet, "q", false, "Suppress non-essential output (shorthand)")
	jsonErrors := fs.Bool("json-errors", false, "Emit errors as JSON")
	dryRun := fs.Bool("dry-run", false, "Preview changes without writing files or tagging")
	configPath := fs.String("config", "", "Config file path override")
	noConfigUpdate := fs.Bool("no-config-update", false, "Skip updating repo config")

	printUsage := func(showBanner bool) {
		if shouldShowBanner() && showBanner {
			fmt.Fprintln(os.Stdout, renderBanner())
			fmt.Fprintln(os.Stdout)
		}
		fmt.Fprintln(os.Stdout, "TAGtastic release helper (supporting tool)")
		fmt.Fprintln(os.Stdout, "Codename is auto-selected unless --codename is provided.")
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "Usage:")
		fmt.Fprintln(os.Stdout, "  release <version> [--pre <alpha|beta|rc>] [--pre-num N] [--codename NAME] [--date YYYY-MM-DD] [--commit]")
		fmt.Fprintln(os.Stdout, "  release --bump <major|minor|patch> [--pre <alpha|beta|rc>] [--pre-num N] [--codename NAME] [--date YYYY-MM-DD] [--commit]")
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "Flags:")
		fs.SetOutput(os.Stdout)
		fs.PrintDefaults()
	}

	if err := fs.Parse(parsedArgs); err != nil {
		reportError(err, 2, jsonErrorsArg, quietArg, func() { printUsage(!quietArg) })
	}

	quietEnabled := quietArg || *quiet
	jsonEnabled := jsonErrorsArg || *jsonErrors

	if *help {
		printUsage(!quietEnabled)
		os.Exit(0)
	}

	if fs.NArg() == 0 && strings.TrimSpace(*bump) == "" {
		reportError(errors.New("version is required (or use --bump)"), 2, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
	}
	if fs.NArg() > 0 && strings.TrimSpace(*bump) != "" {
		reportError(errors.New("use either a version argument or --bump, not both"), 2, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
	}

	var version string
	if fs.NArg() > 0 {
		version = strings.TrimSpace(fs.Arg(0))
		if version == "" {
			reportError(errors.New("version is required"), 2, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
		}
	}

	root, err := os.Getwd()
	if err != nil {
		reportError(err, 1, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
	}

	latestVersion, err := latestVersion(root)
	if err != nil {
		reportError(err, 1, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
	}

	if strings.TrimSpace(*bump) != "" {
		if latestVersion == "" {
			reportError(errors.New("unable to auto-bump version: no existing version tags or VERSION file found"), 2, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
		}
		version, err = bumpVersion(latestVersion, strings.TrimSpace(*bump))
		if err != nil {
			reportError(err, 2, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
		}
	}

	if strings.TrimSpace(*pre) == "" && *preNum > 0 {
		reportError(errors.New("use --pre when providing --pre-num"), 2, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
	}

	if strings.TrimSpace(*pre) != "" {
		if _, err := parseSemVer(version); err != nil {
			reportError(err, 2, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
		}
		version, err = resolvePreReleaseVersion(version, strings.TrimSpace(*pre), *preNum)
		if err != nil {
			reportError(err, 2, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
		}
	}

	if err := ensureSemVerForward(version, latestVersion); err != nil {
		reportError(err, 2, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
	}

	resolvedDate := strings.TrimSpace(*date)
	if resolvedDate == "" {
		resolvedDate = time.Now().Format("2006-01-02")
	}

	resolvedCodename := strings.TrimSpace(*codename)
	if resolvedCodename == "" {
		resolvedCodename, err = nextCodename(filepath.Join(root, "data", "crayola.json"), filepath.Join(root, "CHANGELOG.md"))
		if err != nil {
			reportError(err, 1, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
		}
	}

	configTarget, err := resolveConfigPath(root, *configPath)
	if err != nil {
		reportError(err, 1, jsonEnabled, quietEnabled, func() { printUsage(!quietEnabled) })
	}

	if *dryRun {
		fmt.Printf("Dry run: would prepare v%s – %s\n", version, resolvedCodename)
		fmt.Printf("Dry run: would update CHANGELOG.md and VERSION (date %s)\n", resolvedDate)
		fmt.Printf("Dry run: would create tag v%s\n", version)
		if !*noConfigUpdate {
			fmt.Printf("Dry run: would update config at %s\n", configTarget)
		}
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

	if !*noConfigUpdate {
		if err := updateRepoConfig(configTarget, resolvedCodename, version); err != nil {
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

type semVer struct {
	major    int
	minor    int
	patch    int
	pre      string
	preLabel string
	preNum   int
	hasPre   bool
}

var semVerPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?$`)

func parseSemVer(input string) (semVer, error) {
	match := semVerPattern.FindStringSubmatch(strings.TrimSpace(input))
	if match == nil {
		return semVer{}, fmt.Errorf("invalid SemVer: %s", input)
	}

	major, err := strconv.Atoi(match[1])
	if err != nil {
		return semVer{}, err
	}
	minor, err := strconv.Atoi(match[2])
	if err != nil {
		return semVer{}, err
	}
	patch, err := strconv.Atoi(match[3])
	if err != nil {
		return semVer{}, err
	}

	version := semVer{major: major, minor: minor, patch: patch}
	if match[4] != "" {
		version.hasPre = true
		version.pre = match[4]
		label, num := splitPreRelease(match[4])
		version.preLabel = label
		version.preNum = num
	}

	return version, nil
}

func splitPreRelease(value string) (string, int) {
	parts := strings.Split(value, ".")
	if len(parts) == 1 {
		return value, 0
	}
	last := parts[len(parts)-1]
	num, err := strconv.Atoi(last)
	if err != nil {
		return value, 0
	}
	label := strings.Join(parts[:len(parts)-1], ".")
	if label == "" {
		label = value
	}
	return label, num
}

func compareSemVer(a, b semVer) int {
	if a.major != b.major {
		return compareInt(a.major, b.major)
	}
	if a.minor != b.minor {
		return compareInt(a.minor, b.minor)
	}
	if a.patch != b.patch {
		return compareInt(a.patch, b.patch)
	}
	if !a.hasPre && !b.hasPre {
		return 0
	}
	if !a.hasPre {
		return 1
	}
	if !b.hasPre {
		return -1
	}

	rankA := preReleaseRank(a.preLabel)
	rankB := preReleaseRank(b.preLabel)
	if rankA != rankB {
		return compareInt(rankA, rankB)
	}

	if a.preLabel != b.preLabel {
		if a.preLabel < b.preLabel {
			return -1
		}
		if a.preLabel > b.preLabel {
			return 1
		}
	}

	if a.preNum != b.preNum {
		return compareInt(a.preNum, b.preNum)
	}
	return 0
}

func compareInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func preReleaseRank(label string) int {
	switch strings.ToLower(label) {
	case "alpha":
		return 0
	case "beta":
		return 1
	case "rc":
		return 2
	default:
		return 3
	}
}

func formatSemVer(version semVer) string {
	base := fmt.Sprintf("%d.%d.%d", version.major, version.minor, version.patch)
	if version.hasPre && version.pre != "" {
		return base + "-" + version.pre
	}
	return base
}

func ensureSemVerForward(version, latest string) error {
	if strings.TrimSpace(version) == "" {
		return errors.New("version is required")
	}
	parsed, err := parseSemVer(version)
	if err != nil {
		return err
	}
	if strings.TrimSpace(latest) == "" {
		return nil
	}
	parsedLatest, err := parseSemVer(latest)
	if err != nil {
		return fmt.Errorf("invalid latest version %s: %w", latest, err)
	}
	if compareSemVer(parsed, parsedLatest) <= 0 {
		return fmt.Errorf("version must be greater than %s", formatSemVer(parsedLatest))
	}
	return nil
}

func bumpVersion(base, bump string) (string, error) {
	parsed, err := parseSemVer(base)
	if err != nil {
		return "", err
	}
	switch strings.ToLower(bump) {
	case "major":
		parsed.major++
		parsed.minor = 0
		parsed.patch = 0
	case "minor":
		parsed.minor++
		parsed.patch = 0
	case "patch":
		parsed.patch++
	default:
		return "", fmt.Errorf("invalid bump value: %s (expected major, minor, or patch)", bump)
	}
	parsed.hasPre = false
	parsed.pre = ""
	parsed.preLabel = ""
	parsed.preNum = 0
	return formatSemVer(parsed), nil
}

func latestVersion(root string) (string, error) {
	latestTag, err := latestTagVersion()
	if err != nil {
		return "", err
	}
	if latestTag != "" {
		return latestTag, nil
	}

	payload, err := os.ReadFile(filepath.Join(root, "VERSION"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	version := strings.TrimSpace(string(payload))
	if version == "" {
		return "", nil
	}
	if _, err := parseSemVer(version); err != nil {
		return "", fmt.Errorf("invalid VERSION file: %w", err)
	}
	return version, nil
}

func latestTagVersion() (string, error) {
	if _, err := os.Stat(".git"); err != nil {
		return "", nil
	}
	tags, err := listTags()
	if err != nil {
		return "", err
	}

	var latest semVer
	found := false
	for _, line := range tags {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parsed, err := parseSemVer(strings.TrimSpace(line))
		if err != nil {
			continue
		}
		if !found || compareSemVer(parsed, latest) > 0 {
			latest = parsed
			found = true
		}
	}

	if !found {
		return "", nil
	}
	return formatSemVer(latest), nil
}

func resolvePreReleaseVersion(baseVersion, label string, num int) (string, error) {
	tags, err := listTags()
	if err != nil {
		return "", err
	}
	return resolvePreReleaseVersionWithTags(baseVersion, label, num, tags)
}

func resolvePreReleaseVersionWithTags(baseVersion, label string, num int, tags []string) (string, error) {
	label = strings.ToLower(strings.TrimSpace(label))
	if err := validatePreLabel(label); err != nil {
		return "", err
	}

	base, err := parseSemVer(baseVersion)
	if err != nil {
		return "", err
	}
	if base.hasPre {
		return "", errors.New("version already includes prerelease; omit --pre")
	}

	if num < 0 {
		return "", errors.New("pre-release number must be positive")
	}
	if num == 0 {
		next, err := nextPreNumFromTags(baseVersion, label, tags)
		if err != nil {
			return "", err
		}
		num = next
	}
	if num < 1 {
		return "", errors.New("pre-release number must be at least 1")
	}

	version := fmt.Sprintf("%d.%d.%d-%s.%d", base.major, base.minor, base.patch, label, num)
	return version, nil
}

func validatePreLabel(label string) error {
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "alpha", "beta", "rc":
		return nil
	default:
		return fmt.Errorf("invalid prerelease label: %s", label)
	}
}

func nextPreNum(baseVersion, label string) (int, error) {
	tags, err := listTags()
	if err != nil {
		return 0, err
	}
	return nextPreNumFromTags(baseVersion, label, tags)
}

func nextPreNumFromTags(baseVersion, label string, tags []string) (int, error) {
	base, err := parseSemVer(baseVersion)
	if err != nil {
		return 0, err
	}
	if base.hasPre {
		return 0, errors.New("base version must not include prerelease")
	}

	label = strings.ToLower(strings.TrimSpace(label))
	max := 0
	for _, tag := range tags {
		parsed, err := parseSemVer(strings.TrimSpace(tag))
		if err != nil || !parsed.hasPre {
			continue
		}
		if parsed.major != base.major || parsed.minor != base.minor || parsed.patch != base.patch {
			continue
		}
		if strings.ToLower(parsed.preLabel) != label {
			continue
		}
		if parsed.preNum > max {
			max = parsed.preNum
		}
	}
	return max + 1, nil
}

func listTags() ([]string, error) {
	if _, err := os.Stat(".git"); err != nil {
		return []string{}, nil
	}
	cmd := exec.Command("git", "tag", "-l", "v*")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && strings.TrimSpace(lines[0]) == "" {
		return []string{}, nil
	}
	return lines, nil
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

func reorderArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}

	valueFlags := map[string]struct{}{
		"--bump":     {},
		"--codename": {},
		"--date":     {},
		"--pre":      {},
		"--pre-num":  {},
		"--config":   {},
	}

	var flags []string
	var positionals []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			if _, ok := valueFlags[arg]; ok {
				if i+1 < len(args) {
					flags = append(flags, args[i+1])
					i++
				}
			}
			continue
		}
		positionals = append(positionals, arg)
	}

	return append(flags, positionals...)
}

func resolveConfigPath(root, override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return config.ResolvePath(override)
	}
	if env := strings.TrimSpace(os.Getenv("TAGTASTIC_CONFIG")); env != "" {
		return config.ResolvePath(env)
	}
	repoPath := filepath.Join(root, ".tagtastic.yaml")
	if _, err := os.Stat(repoPath); err == nil {
		return config.ResolvePath(repoPath)
	}
	return config.ResolvePath(repoPath)
}

func updateRepoConfig(path, codename, version string) error {
	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

	if cfg.DefaultTheme == "" {
		cfg.DefaultTheme = "crayola_colors"
	}
	if cfg.DefaultFormat == "" {
		cfg.DefaultFormat = "text"
	}
	if cfg.UsedCodenames == nil {
		cfg.UsedCodenames = map[string]string{}
	}

	cfg.UsedCodenames[version] = codename

	payload, err := config.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	return os.WriteFile(path, payload, 0o600)
}
