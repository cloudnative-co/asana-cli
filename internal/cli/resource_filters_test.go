package cli

import (
	"testing"

	"github.com/cloudnative-co/asana-cli/internal/asanaapi"
)

func TestShouldAutoPaginateByDefault(t *testing.T) {
	if !shouldAutoPaginateByDefault(asanaapi.Endpoint{Name: "list-project", Method: "GET"}) {
		t.Fatalf("expected list-project GET to auto paginate by default")
	}
	if shouldAutoPaginateByDefault(asanaapi.Endpoint{Name: "get", Method: "GET"}) {
		t.Fatalf("expected get GET to not auto paginate by default")
	}
	if shouldAutoPaginateByDefault(asanaapi.Endpoint{Name: "list", Method: "POST"}) {
		t.Fatalf("expected POST to not auto paginate by default")
	}
}

func TestShouldSetDefaultLimit(t *testing.T) {
	endpoint := asanaapi.Endpoint{Name: "list-project", Method: "GET"}
	if !shouldSetDefaultLimit(endpoint, map[string]string{}) {
		t.Fatalf("expected default limit to be set")
	}
	if shouldSetDefaultLimit(endpoint, map[string]string{"limit": "10"}) {
		t.Fatalf("expected default limit not to be set when query has limit")
	}
	if shouldSetDefaultLimit(asanaapi.Endpoint{Name: "get", Method: "GET"}, map[string]string{}) {
		t.Fatalf("expected default limit not to be set for non-list endpoint")
	}
}

func TestShouldUseTaskAssigneeFlag(t *testing.T) {
	if !shouldUseTaskAssigneeFlag("task", asanaapi.Endpoint{Name: "list"}) {
		t.Fatalf("expected task list to support --assignee")
	}
	if shouldUseTaskAssigneeFlag("task", asanaapi.Endpoint{Name: "get"}) {
		t.Fatalf("expected task get not to support --assignee")
	}
}

func TestShouldSupportTaskProjectResolution(t *testing.T) {
	if !shouldSupportTaskProjectResolution("task", asanaapi.Endpoint{Name: "get"}) {
		t.Fatalf("expected task get to support project resolution")
	}
	if !shouldSupportTaskProjectResolution("task", asanaapi.Endpoint{Name: "list-project"}) {
		t.Fatalf("expected task list-project to support project resolution")
	}
	if shouldSupportTaskProjectResolution("project", asanaapi.Endpoint{Name: "get"}) {
		t.Fatalf("expected non-task command not to support project resolution")
	}
	if shouldSupportTaskProjectResolution("task", asanaapi.Endpoint{Name: "duplicate"}) {
		t.Fatalf("expected task duplicate not to support project resolution")
	}
}

func TestShouldSupportTaskSubtaskExpansion(t *testing.T) {
	if !shouldSupportTaskSubtaskExpansion("task", asanaapi.Endpoint{Name: "get"}) {
		t.Fatalf("expected task get to support subtask expansion")
	}
	if !shouldSupportTaskSubtaskExpansion("task", asanaapi.Endpoint{Name: "list-project"}) {
		t.Fatalf("expected task list-project to support subtask expansion")
	}
	if shouldSupportTaskSubtaskExpansion("project", asanaapi.Endpoint{Name: "get"}) {
		t.Fatalf("expected non-task command not to support subtask expansion")
	}
	if shouldSupportTaskSubtaskExpansion("task", asanaapi.Endpoint{Name: "duplicate"}) {
		t.Fatalf("expected task duplicate not to support subtask expansion")
	}
}

func TestApplyNameFilter(t *testing.T) {
	resp := map[string]any{
		"data": []any{
			map[string]any{"gid": "1", "name": "Pocketalk sync"},
			map[string]any{"gid": "2", "name": "Other task"},
		},
	}
	if err := applyNameFilter(resp, "pocketalk", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := resp["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 row after filter, got %d", len(data))
	}
}

func TestApplyNameFilter_InvalidRegex(t *testing.T) {
	resp := map[string]any{"data": []any{map[string]any{"name": "a"}}}
	if err := applyNameFilter(resp, "", "["); err == nil {
		t.Fatalf("expected invalid regex error")
	}
}

func TestApplyNameFilter_NonListResponse(t *testing.T) {
	resp := map[string]any{"data": map[string]any{"name": "x"}}
	if err := applyNameFilter(resp, "x", ""); err == nil {
		t.Fatalf("expected error for non-list response")
	}
}

func TestExtractTaskMaps(t *testing.T) {
	resp := []any{
		map[string]any{"gid": "1", "name": "Task 1"},
		map[string]any{"gid": "2", "name": "Task 2"},
	}
	tasks, ok := extractTaskMaps(resp)
	if !ok {
		t.Fatalf("expected task extraction to succeed")
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestExtractTaskMapsRejectsNonTask(t *testing.T) {
	if _, ok := extractTaskMaps(map[string]any{"name": "missing gid"}); ok {
		t.Fatalf("expected non-task map to be rejected")
	}
}

func TestExtractAllTaskMaps(t *testing.T) {
	resp := map[string]any{
		"gid":  "1",
		"name": "Task 1",
		"descendant_subtasks": []any{
			map[string]any{"gid": "2", "name": "Task 2"},
			map[string]any{"gid": "3", "name": "Task 3"},
		},
	}
	tasks, ok := extractAllTaskMaps(resp)
	if !ok {
		t.Fatalf("expected recursive task extraction to succeed")
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}
}

func TestFlattenTaskListWithDescendants(t *testing.T) {
	root := map[string]any{
		"gid":                       "1",
		"name":                      "root",
		"descendant_subtasks":       []any{map[string]any{"gid": "2", "name": "child", "subtask_depth": 1}},
		"descendant_subtasks_count": 1,
	}
	rows := flattenTaskListWithDescendants([]map[string]any{root})
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows after flattening, got %d", len(rows))
	}
	if _, ok := rows[0]["descendant_subtasks"]; ok {
		t.Fatalf("expected flattened root row to omit descendant_subtasks")
	}
	if rows[1]["gid"] != "2" {
		t.Fatalf("unexpected descendant row: %#v", rows[1])
	}
}
