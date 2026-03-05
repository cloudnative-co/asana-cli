package auth

import "testing"

func TestResolveScopePreset(t *testing.T) {
	scopes, err := ResolveScopePreset("task-full")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(scopes) == 0 {
		t.Fatalf("expected non-empty scopes")
	}
	foundCustomFields := false
	for _, scope := range scopes {
		if scope == "custom_fields:read" {
			foundCustomFields = true
			break
		}
	}
	if !foundCustomFields {
		t.Fatalf("expected custom_fields:read in preset scopes")
	}
}

func TestNormalizeScopes(t *testing.T) {
	got := NormalizeScopes([]string{"tasks:read,tasks:write", "tasks:read", " custom_fields:read "})
	if len(got) != 3 {
		t.Fatalf("expected 3 scopes, got %d", len(got))
	}
	if got[0] != "custom_fields:read" || got[1] != "tasks:read" || got[2] != "tasks:write" {
		t.Fatalf("unexpected scope normalization: %#v", got)
	}
}
