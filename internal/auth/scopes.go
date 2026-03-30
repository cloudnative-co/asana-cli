package auth

import (
	"slices"
	"strings"

	"github.com/cloudnative-co/asana-cli/internal/errs"
)

var cliDefaultScopes = []string{
	"attachments:delete",
	"attachments:read",
	"attachments:write",
	"custom_fields:read",
	"custom_fields:write",
	"projects:delete",
	"projects:read",
	"projects:write",
	"stories:read",
	"stories:write",
	"tags:read",
	"tags:write",
	"task_custom_types:read",
	"task_templates:read",
	"tasks:delete",
	"tasks:read",
	"tasks:write",
	"teams:read",
	"time_tracking_entries:read",
	"users:read",
	"workspaces:read",
}

var taskRelatedScopes = append([]string(nil), cliDefaultScopes...)

func ResolveScopePreset(name string) ([]string, error) {
	normalized := strings.ToLower(strings.TrimSpace(name))
	switch normalized {
	case "cli-default", "default", "recommended":
		return append([]string(nil), cliDefaultScopes...), nil
	case "task-full", "tasks-full", "task-related", "task-related-full":
		return append([]string(nil), taskRelatedScopes...), nil
	default:
		return nil, errs.New("invalid_argument", "unknown --scope-preset value", "supported: cli-default, task-full")
	}
}

func NormalizeScopes(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, raw := range in {
		for _, part := range strings.Split(raw, ",") {
			scope := strings.TrimSpace(part)
			if scope == "" {
				continue
			}
			if _, exists := seen[scope]; exists {
				continue
			}
			seen[scope] = struct{}{}
			out = append(out, scope)
		}
	}
	slices.Sort(out)
	return out
}
