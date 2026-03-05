package cli

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/cloudnative-co/asana-cli/internal/errs"
)

var placeholderRe = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)

func parseKeyValueFlags(items []string) (map[string]string, error) {
	parsed := map[string]string{}
	for _, item := range items {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			return nil, errs.New("invalid_argument", fmt.Sprintf("invalid key=value: %s", item), "")
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, errs.New("invalid_argument", fmt.Sprintf("invalid key in: %s", item), "")
		}
		parsed[key] = value
	}
	return parsed, nil
}

func parseBodyFromFlags(jsonData string, fields []string) (map[string]any, error) {
	if strings.TrimSpace(jsonData) != "" {
		var payload map[string]any
		if err := json.Unmarshal([]byte(jsonData), &payload); err != nil {
			return nil, errs.Wrap("invalid_argument", "failed to parse --data JSON", "", err)
		}
		return payload, nil
	}
	if len(fields) == 0 {
		return nil, nil
	}
	kv, err := parseKeyValueFlags(fields)
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	for key, value := range kv {
		if value == "" {
			out[key] = ""
			continue
		}
		var decoded any
		if decodeErr := json.Unmarshal([]byte(value), &decoded); decodeErr == nil {
			out[key] = decoded
			continue
		}
		out[key] = value
	}
	return out, nil
}

func placeholders(path string) []string {
	matches := placeholderRe.FindAllStringSubmatch(path, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		out = append(out, m[1])
	}
	sort.Strings(out)
	return out
}

func fillPath(path string, values map[string]string) (string, error) {
	result := path
	for _, key := range placeholders(path) {
		value, ok := values[key]
		if !ok || strings.TrimSpace(value) == "" {
			return "", errs.New("invalid_argument", fmt.Sprintf("missing required path param: %s", key), "")
		}
		result = strings.ReplaceAll(result, "{"+key+"}", value)
	}
	return result, nil
}

func normalizeFlagName(in string) string {
	return strings.ReplaceAll(in, "_", "-")
}
