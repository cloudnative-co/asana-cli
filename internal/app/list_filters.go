package app

import (
	"regexp"
	"strings"
)

type NameFilterOptions struct {
	Contains string
	Regex    *regexp.Regexp
}

func FilterByName(rows []map[string]any, opts NameFilterOptions) []map[string]any {
	contains := strings.ToLower(strings.TrimSpace(opts.Contains))
	filtered := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		name, _ := row["name"].(string)
		nameFolded := strings.ToLower(name)
		if contains != "" && !strings.Contains(nameFolded, contains) {
			continue
		}
		if opts.Regex != nil && !opts.Regex.MatchString(name) {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}
