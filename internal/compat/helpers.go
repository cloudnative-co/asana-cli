package compat

import (
	"sort"
	"strings"
	"time"
)

func SortTasksByDue(tasks []map[string]any) []map[string]any {
	withDue := make([]map[string]any, 0, len(tasks))
	withoutDue := make([]map[string]any, 0, len(tasks))
	for _, task := range tasks {
		due, _ := task["due_on"].(string)
		if strings.TrimSpace(due) == "" {
			withoutDue = append(withoutDue, task)
		} else {
			withDue = append(withDue, task)
		}
	}
	sort.SliceStable(withDue, func(i, j int) bool {
		a, _ := withDue[i]["due_on"].(string)
		b, _ := withDue[j]["due_on"].(string)
		return a < b
	})
	return append(withDue, withoutDue...)
}

func BuildIndexedRows(tasks []map[string]any) []any {
	rows := make([]any, 0, len(tasks))
	for idx, task := range tasks {
		row := map[string]any{"index": idx}
		for key, value := range task {
			row[key] = value
		}
		rows = append(rows, row)
	}
	return rows
}

func ParseCompatDue(input string, now time.Time) string {
	normalized := strings.ToLower(strings.TrimSpace(input))
	switch normalized {
	case "today":
		return now.Format("2006-01-02")
	case "tomorrow":
		return now.Add(24 * time.Hour).Format("2006-01-02")
	default:
		return input
	}
}
