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
