package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type colorEntry struct {
	Color string `json:"color"`
}

type colorFile struct {
	Colors []colorEntry `json:"colors"`
}

func main() {
	root, err := repoRoot()
	if err != nil {
		fatal(err)
	}

	colors, err := loadColors(filepath.Join(root, "data", "crayola.json"))
	if err != nil {
		fatal(err)
	}

	used, err := loadUsedCodenames(filepath.Join(root, "CHANGELOG.md"))
	if err != nil {
		fatal(err)
	}

	for _, color := range colors {
		if _, ok := used[color]; ok {
			continue
		}
		fmt.Println(color)
		return
	}

	fatal(fmt.Errorf("no available codenames left in crayola.json"))
}

func repoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return wd, nil
}

func loadColors(path string) ([]string, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var file colorFile
	if err := json.Unmarshal(payload, &file); err != nil {
		return nil, err
	}

	colors := make([]string, 0, len(file.Colors))
	for _, entry := range file.Colors {
		name := strings.TrimSpace(entry.Color)
		if name == "" {
			continue
		}
		colors = append(colors, name)
	}

	return colors, nil
}

func loadUsedCodenames(path string) (map[string]struct{}, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	defer file.Close()

	used := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.Index(line, "– \""); idx != -1 {
			fragment := line[idx+len("– \""):]
			if end := strings.Index(fragment, "\""); end != -1 {
				codename := fragment[:end]
				codename = strings.TrimSpace(codename)
				if codename != "" {
					used[codename] = struct{}{}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return used, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
