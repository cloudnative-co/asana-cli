package asanaapi

import "testing"

func TestEndpointCounts(t *testing.T) {
	if got := len(TaskEndpoints); got != 29 {
		t.Fatalf("tasks endpoints mismatch: got %d want 29", got)
	}
	if got := len(ProjectEndpoints); got != 19 {
		t.Fatalf("projects endpoints mismatch: got %d want 19", got)
	}
	if got := len(UserEndpoints); got != 8 {
		t.Fatalf("users endpoints mismatch: got %d want 8", got)
	}
	if got := len(AttachmentEndpoints); got != 4 {
		t.Fatalf("attachment endpoints mismatch: got %d want 4", got)
	}
	if got := len(StoryEndpoints); got != 5 {
		t.Fatalf("story endpoints mismatch: got %d want 5", got)
	}
	if got := len(TagEndpoints); got != 9 {
		t.Fatalf("tag endpoints mismatch: got %d want 9", got)
	}
	if got := len(SectionEndpoints); got != 8 {
		t.Fatalf("section endpoints mismatch: got %d want 8", got)
	}
	if got := len(CustomFieldEndpoints); got != 8 {
		t.Fatalf("custom field endpoints mismatch: got %d want 8", got)
	}
	if got := len(UserTaskListEndpoints); got != 2 {
		t.Fatalf("user task list endpoints mismatch: got %d want 2", got)
	}
	if got := len(TimeTrackingEntryEndpoints); got != 6 {
		t.Fatalf("time tracking entry endpoints mismatch: got %d want 6", got)
	}
}
