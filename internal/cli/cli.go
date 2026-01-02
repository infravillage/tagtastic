package cli

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/infravillage/tagtastic/internal/config"
	"github.com/infravillage/tagtastic/internal/data"
	"github.com/infravillage/tagtastic/internal/output"
)

type Dependencies struct {
	Themes           data.ThemeRepository
	FormatterFactory func(format string) (output.Formatter, error)
	Out              io.Writer
	VersionInfo      VersionInfo
}

type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

type CLI struct {
	Generate GenerateCmd `cmd:"" help:"Generate a codename"`
	List     ListCmd     `cmd:"" help:"List codenames in a theme"`
	Themes   ThemesCmd   `cmd:"" help:"List available themes"`
	Validate ValidateCmd `cmd:"" help:"Validate a codename"`
	Config   ConfigCmd   `cmd:"" help:"Manage local config"`
	Version  VersionCmd  `cmd:"" help:"Show version"`
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

	app.Config.Init.deps = deps
	app.Config.Show.deps = deps
	app.Config.Reset.deps = deps

	return app
}

type GenerateCmd struct {
	Theme   string   `short:"t" long:"theme" help:"Theme to use" default:"crayola_colors"`
	Seed    int64    `short:"s" long:"seed" help:"Random seed (0 uses time)" default:"0"`
	Exclude []string `short:"e" long:"exclude" help:"Comma-separated names to exclude" sep:","`
	Format  string   `short:"f" long:"format" help:"Output format (text, json, shell)" default:"text"`
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

	picker := rand.New(rand.NewSource(seed))
	selected := available[picker.Intn(len(available))]

	outputText, err := formatter.FormatName(selected)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.deps.Out, outputText)
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
				fmt.Fprintf(cmd.deps.Out, "Found in theme '%s'\n", themeName)
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
		fmt.Fprintf(cmd.deps.Out, "Found in theme '%s'\n", cmd.Theme)
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
	Path  string `short:"p" long:"path" help:"Config file path"`
	Force bool   `long:"force" help:"Overwrite existing config"`
	deps  Dependencies
}

func (cmd ConfigInitCmd) Run() error {
	path, err := config.ResolvePath(cmd.Path)
	if err != nil {
		return err
	}

	if !cmd.Force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config already exists at %s", path)
		}
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
	path, err := config.ResolvePath(cmd.Path)
	if err != nil {
		return err
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fmt.Fprint(cmd.deps.Out, string(payload))
	return nil
}

type ConfigResetCmd struct {
	Path string `short:"p" long:"path" help:"Config file path"`
	deps Dependencies
}

func (cmd ConfigResetCmd) Run() error {
	path, err := config.ResolvePath(cmd.Path)
	if err != nil {
		return err
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
