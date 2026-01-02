package output

import (
	"encoding/json"

	"github.com/infravillage/tagtastic/internal/data"
)

type JSONFormatter struct{}

func (JSONFormatter) FormatName(item data.CodeName) (string, error) {
	payload := struct {
		Name        string   `json:"name"`
		Aliases     []string `json:"aliases,omitempty"`
		Description string   `json:"description,omitempty"`
	}{
		Name:        item.Name,
		Aliases:     item.Aliases,
		Description: item.Description,
	}

	output, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (JSONFormatter) FormatList(items []data.CodeName) (string, error) {
	output, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (JSONFormatter) FormatThemes(names []string) (string, error) {
	output, err := json.MarshalIndent(names, "", "  ")
	if err != nil {
		return "", err
	}
	return string(output), nil
}
