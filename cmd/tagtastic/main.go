package main

import (
	"fmt"
	"os"

	"github.com/infravillage/tagtastic/internal/cli"
	"github.com/infravillage/tagtastic/internal/data"
	"github.com/infravillage/tagtastic/internal/output"
	"github.com/alecthomas/kong"
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

	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
	parser.FatalIfErrorf(ctx.Run())
}
