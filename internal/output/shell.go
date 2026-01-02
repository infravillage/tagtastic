package output

import (
	"fmt"
	"strings"

	"github.com/infravillage/tagtastic/internal/data"
)

type ShellFormatter struct{}

func (ShellFormatter) FormatName(item data.CodeName) (string, error) {
	return fmt.Sprintf("RELEASE_CODENAME=%q", item.Name), nil
}

func (ShellFormatter) FormatList(items []data.CodeName) (string, error) {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, item.Name)
	}
	return strings.Join(lines, "\n"), nil
}

func (ShellFormatter) FormatThemes(names []string) (string, error) {
	return strings.Join(names, "\n"), nil
}
