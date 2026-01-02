package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultTheme  string            `yaml:"default_theme"`
	DefaultFormat string            `yaml:"default_format"`
	UsedCodenames map[string]string `yaml:"used_codenames"`
	API           APIConfig         `yaml:"api"`
}

type APIConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
	CacheDir string `yaml:"cache_dir"`
	CacheTTL string `yaml:"cache_ttl"`
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".tagtastic", "config.yaml"), nil
}

func Default() Config {
	return Config{
		DefaultTheme:  "crayola_colors",
		DefaultFormat: "text",
		UsedCodenames: map[string]string{},
		API: APIConfig{
			Enabled:  false,
			Endpoint: "",
			CacheDir: "",
			CacheTTL: "",
		},
	}
}

func Marshal(cfg Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}

func ResolvePath(path string) (string, error) {
	if path == "" {
		resolved, err := DefaultPath()
		if err != nil {
			return "", err
		}
		path = resolved
	}

	resolved, err := expandHome(path)
	if err != nil {
		return "", err
	}

	return resolved, nil
}

func Load(path string) (Config, error) {
	resolved, err := ResolvePath(path)
	if err != nil {
		return Config{}, err
	}

	payload, err := os.ReadFile(resolved)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(payload, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

func expandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	trimmed := strings.TrimPrefix(path, "~")
	trimmed = strings.TrimPrefix(trimmed, string(filepath.Separator))
	return filepath.Join(home, trimmed), nil
}
