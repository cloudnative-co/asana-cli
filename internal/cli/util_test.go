package cli

import "testing"

func TestParseKeyValueFlags(t *testing.T) {
	parsed, err := parseKeyValueFlags([]string{"a=1", "b=two"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed["a"] != "1" || parsed["b"] != "two" {
		t.Fatalf("unexpected parsed map: %#v", parsed)
	}

	if _, err := parseKeyValueFlags([]string{"broken"}); err == nil {
		t.Fatalf("expected error for invalid kv")
	}
}

func TestPlaceholders(t *testing.T) {
	vals := placeholders("/tasks/{task_gid}/stories/{story_gid}")
	if len(vals) != 2 {
		t.Fatalf("expected 2 placeholders, got %d", len(vals))
	}
}

func TestFillPathIncludesFlagHint(t *testing.T) {
	_, err := fillPath("/projects/{project_gid}", map[string]string{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := err.Error(); got != "invalid_argument: missing required path param: project_gid" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyPositionalPathArgs(t *testing.T) {
	values := map[string]string{}
	if err := applyPositionalPathArgs("/projects/{project_gid}/tasks/{task_gid}", values, []string{"123", "456"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if values["project_gid"] != "123" || values["task_gid"] != "456" {
		t.Fatalf("unexpected values: %#v", values)
	}
}

func TestApplyPositionalPathArgsPreservesFlags(t *testing.T) {
	values := map[string]string{"project_gid": "flagged"}
	if err := applyPositionalPathArgs("/projects/{project_gid}/tasks/{task_gid}", values, []string{"123", "456"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if values["project_gid"] != "flagged" {
		t.Fatalf("expected existing flag value to win, got %#v", values)
	}
	if values["task_gid"] != "456" {
		t.Fatalf("expected second arg to fill task_gid, got %#v", values)
	}
}

func TestApplyPositionalPathArgsRejectsTooManyArgs(t *testing.T) {
	err := applyPositionalPathArgs("/projects/{project_gid}", map[string]string{}, []string{"123", "456"})
	if err == nil {
		t.Fatalf("expected error")
	}
}
