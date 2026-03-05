package app

import (
	"regexp"
	"testing"
)

func TestFilterByName_ContainsCaseInsensitive(t *testing.T) {
	rows := []map[string]any{
		{"gid": "1", "name": "Pocketalk sync"},
		{"gid": "2", "name": "POCKETALK test"},
		{"gid": "3", "name": "Other task"},
	}
	filtered := FilterByName(rows, NameFilterOptions{Contains: "pocketalk"})
	if len(filtered) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(filtered))
	}
}

func TestFilterByName_Regex(t *testing.T) {
	rows := []map[string]any{
		{"gid": "1", "name": "pocketalk-001"},
		{"gid": "2", "name": "pocketalk-xyz"},
		{"gid": "3", "name": "other"},
	}
	filtered := FilterByName(rows, NameFilterOptions{Regex: regexp.MustCompile(`pocketalk-[0-9]+`)})
	if len(filtered) != 1 {
		t.Fatalf("expected 1 row, got %d", len(filtered))
	}
	if filtered[0]["gid"] != "1" {
		t.Fatalf("expected gid=1, got %v", filtered[0]["gid"])
	}
}

func TestFilterByName_ContainsAndRegex(t *testing.T) {
	rows := []map[string]any{
		{"gid": "1", "name": "pocketalk-001 done"},
		{"gid": "2", "name": "pocketalk-final"},
		{"gid": "3", "name": "other"},
	}
	filtered := FilterByName(rows, NameFilterOptions{
		Contains: "pocketalk",
		Regex:    regexp.MustCompile(`\d+`),
	})
	if len(filtered) != 1 {
		t.Fatalf("expected 1 row, got %d", len(filtered))
	}
}
