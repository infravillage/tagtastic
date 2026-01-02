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

func TestGenerateCommand_AllExcluded(t *testing.T) {
	_, err := runCLI(t, "generate", "--theme", "birds", "--exclude", "albatross,blue-heron,crane,dove,eagle", "--seed", "1")
	if err == nil {
		t.Fatalf("expected error when all items are excluded")
	}
}

func TestValidateCommand_MissingName(t *testing.T) {
	_, err := runCLI(t, "validate")
	if err == nil {
		t.Fatalf("expected error for missing name")
	}
}

func TestConfigInit_Existing(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.yaml")

	if _, err := runCLI(t, "config", "init", "--path", configPath); err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	if _, err := runCLI(t, "config", "init", "--path", configPath); err == nil {
		t.Fatalf("expected error when config exists without --force")
	}
}

func TestConfigShow_MissingFile(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "missing.yaml")

	if _, err := runCLI(t, "config", "show", "--path", configPath); err == nil {
		t.Fatalf("expected error for missing config file")
	}
}

func TestValidateCommand_AllThemes(t *testing.T) {
	output, err := runCLI(t, "validate", "Almond")
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if !strings.Contains(output, "crayola_colors") {
		t.Fatalf("expected Almond to be found in crayola_colors")
	}
}

func TestConfigReset_MissingFile(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "missing.yaml")

	if _, err := runCLI(t, "config", "reset", "--path", configPath); err != nil {
		t.Fatalf("config reset should ignore missing file, got: %v", err)
	}
}

func TestGenerateCommand_Formats(t *testing.T) {
	jsonOutput, err := runCLI(t, "generate", "--theme", "birds", "--seed", "2", "--format", "json")
	if err != nil {
		t.Fatalf("json format failed: %v", err)
	}
	if !strings.Contains(jsonOutput, "\"name\"") {
		t.Fatalf("expected json output to include name field")
	}

	shellOutput, err := runCLI(t, "generate", "--theme", "birds", "--seed", "2", "--format", "shell")
	if err != nil {
		t.Fatalf("shell format failed: %v", err)
	}
	if !strings.HasPrefix(shellOutput, "RELEASE_CODENAME=") {
		t.Fatalf("expected shell output to include RELEASE_CODENAME")
	}
}

func TestThemesCommand_JSON_Golden(t *testing.T) {
	output, err := runCLI(t, "themes", "--format", "json")
	if err != nil {
		t.Fatalf("themes json failed: %v", err)
	}

	goldenPath := filepath.Join("testdata", "themes.json")
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}

	if strings.TrimSpace(output) != strings.TrimSpace(string(golden)) {
		t.Fatalf("themes json output mismatch\\nexpected: %s\\nactual:   %s", strings.TrimSpace(string(golden)), strings.TrimSpace(output))
	}
}

func TestListCommand_JSON_Golden(t *testing.T) {
	output, err := runCLI(t, "list", "--theme", "birds", "--format", "json")
	if err != nil {
		t.Fatalf("list json failed: %v", err)
	}

	goldenPath := filepath.Join("testdata", "birds.json")
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}

	if strings.TrimSpace(output) != strings.TrimSpace(string(golden)) {
		t.Fatalf("list json output mismatch\\nexpected: %s\\nactual:   %s", strings.TrimSpace(string(golden)), strings.TrimSpace(output))
	}
}

func TestGenerateCommand_JSON_Golden(t *testing.T) {
	output, err := runCLI(t, "generate", "--theme", "birds", "--seed", "42", "--format", "json")
	if err != nil {
		t.Fatalf("generate json failed: %v", err)
	}

	goldenPath := filepath.Join("testdata", "generate.json")
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}

	if strings.TrimSpace(output) != strings.TrimSpace(string(golden)) {
		t.Fatalf("generate json output mismatch\\nexpected: %s\\nactual:   %s", strings.TrimSpace(string(golden)), strings.TrimSpace(output))
	}
}
