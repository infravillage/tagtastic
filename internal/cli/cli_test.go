package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/infravillage/tagtastic/internal/data"
	"github.com/alecthomas/kong"
)

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()

	repo, err := data.NewEmbeddedThemeRepository()
	if err != nil {
		t.Fatalf("load themes: %v", err)
	}

	var out bytes.Buffer
	cli := NewCLI(Dependencies{Themes: repo, Out: &out})
	parser, err := kong.New(cli, kong.Name("tagtastic"))
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		return out.String(), err
	}

	if err := ctx.Run(); err != nil {
		return out.String(), err
	}

	return strings.TrimSpace(out.String()), nil
}

func TestGenerateCommand(t *testing.T) {
	output, err := runCLI(t, "generate", "--theme", "birds", "--seed", "1")
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	if output == "" {
		t.Fatalf("expected output")
	}
}

func TestListCommand(t *testing.T) {
	output, err := runCLI(t, "list", "--theme", "birds")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if !strings.Contains(output, "Albatross") {
		t.Fatalf("expected list output to contain Albatross")
	}
}

func TestThemesCommand(t *testing.T) {
	output, err := runCLI(t, "themes")
	if err != nil {
		t.Fatalf("themes failed: %v", err)
	}

	if !strings.Contains(output, "crayola_colors") {
		t.Fatalf("expected themes output to include crayola_colors")
	}
}

func TestValidateCommand(t *testing.T) {
	_, err := runCLI(t, "validate", "Albatross", "--theme", "birds")
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
}

func TestConfigCommands(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.yaml")

	if _, err := runCLI(t, "config", "init", "--path", configPath); err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	payload, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(payload), "default_theme") {
		t.Fatalf("expected config file contents")
	}

	output, err := runCLI(t, "config", "show", "--path", configPath)
	if err != nil {
		t.Fatalf("config show failed: %v", err)
	}
	if !strings.Contains(output, "default_theme") {
		t.Fatalf("expected config show output")
	}

	if _, err := runCLI(t, "config", "reset", "--path", configPath); err != nil {
		t.Fatalf("config reset failed: %v", err)
	}
}
