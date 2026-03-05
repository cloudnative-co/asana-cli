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
