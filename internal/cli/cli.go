// Copyright (c) 2026 InfraVillage
// SPDX-License-Identifier: MIT
//
// This file is part of TAGtastic and is licensed under the MIT License.

package cli

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/infravillage/tagtastic/internal/config"
	"github.com/infravillage/tagtastic/internal/data"
	"github.com/infravillage/tagtastic/internal/output"
)

type Dependencies struct {
	Themes             data.ThemeRepository
	FormatterFactory   func(format string) (output.Formatter, error)
	Out                io.Writer
	VersionInfo        VersionInfo
	ConfigPathResolver func() string
}

type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

type CLI struct {
	Quiet      bool        `short:"q" long:"quiet" help:"Suppress non-essential output"`
	JSONErrors bool        `long:"json-errors" help:"Emit errors as JSON"`
	ConfigPath string      `long:"config-path" help:"Config file path override"`
	Generate   GenerateCmd `cmd:"" help:"Generate a codename"`
	List       ListCmd     `cmd:"" help:"List codenames in a theme"`
	Themes     ThemesCmd   `cmd:"" help:"List available themes"`
	Validate   ValidateCmd `cmd:"" help:"Validate a codename"`
	Config     ConfigCmd   `cmd:"" help:"Manage local config"`
	Version    VersionCmd  `cmd:"" help:"Show version"`
}

func NewCLI(deps Dependencies) *CLI {
	if deps.Out == nil {
		deps.Out = os.Stdout
	}
	if deps.FormatterFactory == nil {
		deps.FormatterFactory = output.NewFormatter
	}

	app := &CLI{
		Generate: GenerateCmd{deps: deps},
		List:     ListCmd{deps: deps},
		Themes:   ThemesCmd{deps: deps},
		Validate: ValidateCmd{deps: deps},
		Config:   ConfigCmd{deps: deps},
		Version:  VersionCmd{deps: deps},
	}

	app.Generate.deps.ConfigPathResolver = func() string { return app.ConfigPath }
	app.List.deps.ConfigPathResolver = func() string { return app.ConfigPath }
	app.Themes.deps.ConfigPathResolver = func() string { return app.ConfigPath }
	app.Validate.deps.ConfigPathResolver = func() string { return app.ConfigPath }
	app.Config.Init.deps = deps
	app.Config.Show.deps = deps
	app.Config.Reset.deps = deps
	app.Config.Init.deps.ConfigPathResolver = func() string { return app.ConfigPath }
	app.Config.Show.deps.ConfigPathResolver = func() string { return app.ConfigPath }
	app.Config.Reset.deps.ConfigPathResolver = func() string { return app.ConfigPath }
	app.Config.deps.ConfigPathResolver = func() string { return app.ConfigPath }

	return app
}

type GenerateCmd struct {
	Theme   string   `short:"t" long:"theme" help:"Theme to use" default:"crayola_colors"`
	Seed    int64    `short:"s" long:"seed" help:"Random seed (0 uses time)" default:"0"`
	Exclude []string `short:"e" long:"exclude" help:"Comma-separated names to exclude" sep:","`
	Format  string   `short:"f" long:"format" help:"Output format (text, json, shell)" default:"text"`
	Record  bool     `long:"record" help:"Record the selected codename in config"`
	deps    Dependencies
}

func (cmd GenerateCmd) Run() error {
	formatter, err := cmd.deps.FormatterFactory(cmd.Format)
	if err != nil {
		return err
	}

	theme, err := cmd.deps.Themes.GetThemeByName(cmd.Theme)
	if err != nil {
		return err
	}

	available := data.FilterItems(theme.Items, cmd.Exclude)
	if len(available) == 0 {
		return fmt.Errorf("no available codenames after exclusions")
	}

	seed := cmd.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	// #nosec G404 - math/rand is sufficient for non-cryptographic codename selection
	picker := rand.New(rand.NewSource(seed))
	selected := available[picker.Intn(len(available))]

	outputText, err := formatter.FormatName(selected)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.deps.Out, outputText)

	if cmd.Record {
		if err := recordCodename(cmd, selected); err != nil {
			return err
		}
	}

	return nil
}

type ListCmd struct {
	Theme  string `short:"t" long:"theme" help:"Theme to list" default:"crayola_colors"`
	Format string `short:"f" long:"format" help:"Output format (text, json)" default:"text"`
	deps   Dependencies
}

func (cmd ListCmd) Run() error {
	formatter, err := cmd.deps.FormatterFactory(cmd.Format)
	if err != nil {
		return err
	}

	theme, err := cmd.deps.Themes.GetThemeByName(cmd.Theme)
	if err != nil {
		return err
	}

	outputText, err := formatter.FormatList(theme.Items)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.deps.Out, outputText)
	return nil
}

type ThemesCmd struct {
	Format string `short:"f" long:"format" help:"Output format (text, json)" default:"text"`
	deps   Dependencies
}

func (cmd ThemesCmd) Run() error {
	formatter, err := cmd.deps.FormatterFactory(cmd.Format)
	if err != nil {
		return err
	}

	names := cmd.deps.Themes.GetAllThemeNames()
	outputText, err := formatter.FormatThemes(names)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.deps.Out, outputText)
	return nil
}

type ValidateCmd struct {
	Name  string `arg:"" help:"Name to validate"`
	Theme string `short:"t" long:"theme" help:"Theme to search"`
	deps  Dependencies
}

func (cmd ValidateCmd) Run() error {
	if cmd.Name == "" {
		return fmt.Errorf("name is required")
	}

	if cmd.Theme == "" {
		names := cmd.deps.Themes.GetAllThemeNames()
		for _, themeName := range names {
			theme, err := cmd.deps.Themes.GetThemeByName(themeName)
			if err != nil {
				continue
			}
			if containsName(theme.Items, cmd.Name) {
				_, _ = fmt.Fprintf(cmd.deps.Out, "Found in theme '%s'\n", themeName)
				return nil
			}
		}
		return fmt.Errorf("name '%s' not found", cmd.Name)
	}

	theme, err := cmd.deps.Themes.GetThemeByName(cmd.Theme)
	if err != nil {
		return err
	}

	if containsName(theme.Items, cmd.Name) {
		_, _ = fmt.Fprintf(cmd.deps.Out, "Found in theme '%s'\n", cmd.Theme)
		return nil
	}
	return fmt.Errorf("name '%s' not found in theme '%s'", cmd.Name, cmd.Theme)
}

type ConfigCmd struct {
	Init  ConfigInitCmd  `cmd:"" help:"Initialize local config"`
	Show  ConfigShowCmd  `cmd:"" help:"Show local config"`
	Reset ConfigResetCmd `cmd:"" help:"Remove local config"`
	deps  Dependencies
}

type ConfigInitCmd struct {
	Path   string `short:"p" long:"path" help:"Config file path"`
	Force  bool   `long:"force" help:"Overwrite existing config"`
	DryRun bool   `long:"dry-run" help:"Preview changes without writing"`
	deps   Dependencies
}

func (cmd ConfigInitCmd) Run() error {
	path := cmd.Path
	if strings.TrimSpace(path) == "" {
		resolved, err := resolveConfigPath(cmd.deps)
		if err != nil {
			return err
		}
		path = resolved
	}
	path, err := config.ResolvePath(path)
	if err != nil {
		return err
	}

	if !cmd.Force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config already exists at %s", path)
		}
	}

	if cmd.DryRun {
		_, _ = fmt.Fprintf(cmd.deps.Out, "Dry run: would initialize config at %s\n", path)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	payload, err := config.Marshal(config.Default())
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, payload, 0o600); err != nil {
		return err
	}

	fmt.Fprintf(cmd.deps.Out, "Initialized config at %s\n", path)
	return nil
}

type ConfigShowCmd struct {
	Path string `short:"p" long:"path" help:"Config file path"`
	deps Dependencies
}

func (cmd ConfigShowCmd) Run() error {
	path := cmd.Path
	if strings.TrimSpace(path) == "" {
		resolved, err := resolveConfigPath(cmd.deps)
		if err != nil {
			return err
		}
		path = resolved
	}
	path, err := config.ResolvePath(path)
	if err != nil {
		return err
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config not found at %s (run: tagtastic config init)", path)
		}
		return err
	}

	_, _ = fmt.Fprint(cmd.deps.Out, string(payload))
	return nil
}

type ConfigResetCmd struct {
	Path   string `short:"p" long:"path" help:"Config file path"`
	DryRun bool   `long:"dry-run" help:"Preview changes without deleting"`
	deps   Dependencies
}

func (cmd ConfigResetCmd) Run() error {
	path := cmd.Path
	if strings.TrimSpace(path) == "" {
		resolved, err := resolveConfigPath(cmd.deps)
		if err != nil {
			return err
		}
		path = resolved
	}
	path, err := config.ResolvePath(path)
	if err != nil {
		return err
	}

	if cmd.DryRun {
		fmt.Fprintf(cmd.deps.Out, "Dry run: would remove config at %s\n", path)
		return nil
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	fmt.Fprintf(cmd.deps.Out, "Removed config at %s\n", path)
	return nil
}

type VersionCmd struct {
	deps Dependencies
}

func (cmd VersionCmd) Run() error {
	version := cmd.deps.VersionInfo
	if version.Version == "" {
		version.Version = "dev"
	}
	if version.Commit == "" {
		version.Commit = "none"
	}
	if version.Date == "" {
		version.Date = "unknown"
	}

	fmt.Fprintf(cmd.deps.Out, "tagtastic %s (commit %s, built %s)\n", version.Version, version.Commit, version.Date)
	return nil
}

func containsName(items []data.CodeName, name string) bool {
	needle := data.NormalizeName(name)
	if needle == "" {
		return false
	}

	for _, item := range items {
		if data.NormalizeName(item.Name) == needle {
			return true
		}
		for _, alias := range item.Aliases {
			if data.NormalizeName(alias) == needle {
				return true
			}
		}
	}

	return false
}

func recordCodename(cmd GenerateCmd, selected data.CodeName) error {
	path, err := resolveConfigPath(cmd.deps)
	if err != nil {
		return err
	}

	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

	if cfg.UsedCodenames == nil {
		cfg.UsedCodenames = map[string]string{}
	}

	cfg.DefaultTheme = cmd.Theme
	cfg.DefaultFormat = cmd.Format

	cfg.UsedCodenames["unreleased"] = selected.Name

	payload, err := config.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	if err := os.WriteFile(path, payload, 0o600); err != nil {
		return err
	}

	return nil
}

func resolveConfigPath(deps Dependencies) (string, error) {
	if deps.ConfigPathResolver != nil {
		if resolved := strings.TrimSpace(deps.ConfigPathResolver()); resolved != "" {
			return config.ResolvePath(resolved)
		}
	}
	if env := os.Getenv("TAGTASTIC_CONFIG"); strings.TrimSpace(env) != "" {
		return config.ResolvePath(env)
	}
	return config.ResolvePath(".tagtastic.yaml")
}
