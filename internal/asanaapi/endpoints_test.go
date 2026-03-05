package asanaapi

import "testing"

func TestEndpointCounts(t *testing.T) {
	if got := len(TaskEndpoints); got != 27 {
		t.Fatalf("tasks endpoints mismatch: got %d want 27", got)
	}
	if got := len(ProjectEndpoints); got != 19 {
		t.Fatalf("projects endpoints mismatch: got %d want 19", got)
	}
	if got := len(UserEndpoints); got != 8 {
		t.Fatalf("users endpoints mismatch: got %d want 8", got)
	}
}
