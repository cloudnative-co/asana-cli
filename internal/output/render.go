package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/cloudnative-co/asana-cli/internal/errs"
)

func Render(value any, format string, outPath string) error {
	target := io.Writer(os.Stdout)
	var file *os.File
	var err error
	if strings.TrimSpace(outPath) != "" {
		file, err = os.Create(outPath)
		if err != nil {
			return errs.Wrap("internal_error", "failed to create output file", outPath, err)
		}
		defer file.Close()
		target = file
	}

	normalized := strings.ToLower(strings.TrimSpace(format))
	switch normalized {
	case "", "table":
		return renderTable(target, value)
	case "json":
		return renderJSON(target, value)
	case "csv":
		return renderCSV(target, value)
	default:
		return errs.New("invalid_argument", "unsupported output format", "use table, json, or csv")
	}
}

func renderJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func renderTable(w io.Writer, value any) error {
	switch typed := value.(type) {
	case map[string]any:
		if dataSlice, ok := typed["data"].([]any); ok {
			return renderSliceTable(w, dataSlice)
		}
		keys := mapKeys(typed)
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		for _, key := range keys {
			fmt.Fprintf(tw, "%s\t%v\n", key, stringify(typed[key]))
		}
		return tw.Flush()
	case []any:
		return renderSliceTable(w, typed)
	default:
		_, err := fmt.Fprintf(w, "%v\n", typed)
		return err
	}
}

func renderSliceTable(w io.Writer, data []any) error {
	rows := convertRows(data)
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "(no data)")
		return err
	}
	columns := collectColumns(rows)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, strings.Join(columns, "\t"))
	for _, row := range rows {
		parts := make([]string, 0, len(columns))
		for _, column := range columns {
			parts = append(parts, stringify(row[column]))
		}
		fmt.Fprintln(tw, strings.Join(parts, "\t"))
	}
	return tw.Flush()
}

func renderCSV(w io.Writer, value any) error {
	var data []any
	switch typed := value.(type) {
	case map[string]any:
		if d, ok := typed["data"].([]any); ok {
			data = d
		} else {
			data = []any{typed}
		}
	case []any:
		data = typed
	default:
		return errs.New("invalid_argument", "csv output requires list-like data", "use --output json for scalar response")
	}
	rows := convertRows(data)
	if len(rows) == 0 {
		return nil
	}
	columns := collectColumns(rows)
	writer := csv.NewWriter(w)
	if err := writer.Write(columns); err != nil {
		return err
	}
	for _, row := range rows {
		record := make([]string, 0, len(columns))
		for _, column := range columns {
			record = append(record, stringify(row[column]))
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func convertRows(values []any) []map[string]any {
	rows := make([]map[string]any, 0, len(values))
	for _, item := range values {
		switch typed := item.(type) {
		case map[string]any:
			rows = append(rows, typed)
		default:
			rows = append(rows, map[string]any{"value": typed})
		}
	}
	return rows
}

func collectColumns(rows []map[string]any) []string {
	preferred := []string{"index", "gid", "name", "email", "due_on", "completed", "resource_type"}
	columnSet := map[string]struct{}{}
	for _, row := range rows {
		for key := range row {
			columnSet[key] = struct{}{}
		}
	}
	remaining := make([]string, 0, len(columnSet))
	ordered := make([]string, 0, len(columnSet))
	for _, key := range preferred {
		if _, ok := columnSet[key]; ok {
			ordered = append(ordered, key)
			delete(columnSet, key)
		}
	}
	for key := range columnSet {
		remaining = append(remaining, key)
	}
	sort.Strings(remaining)
	return append(ordered, remaining...)
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func stringify(v any) string {
	if v == nil {
		return ""
	}
	switch typed := v.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case bool, float64, float32, int, int64, int32, uint, uint64:
		return fmt.Sprintf("%v", typed)
	default:
		b, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprintf("%v", typed)
		}
		return string(b)
	}
}
