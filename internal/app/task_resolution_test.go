package app

import (
	"context"
	"testing"

	"github.com/cloudnative-co/asana-cli/internal/errs"
)

func TestMergeOptFields(t *testing.T) {
	got := MergeOptFields("gid,name", "name", "parent.gid", "projects.gid")
	if got != "gid,name,parent.gid,projects.gid" {
		t.Fatalf("unexpected merged opt_fields: %q", got)
	}
}

func TestResolveTaskProjects_Direct(t *testing.T) {
	tasks := []map[string]any{
		{
			"gid":         "1",
			"name":        "child",
			"parent":      nil,
			"projects":    []any{map[string]any{"gid": "p1", "name": "Project 1"}},
			"memberships": []any{},
		},
	}
	err := ResolveTaskProjects(context.Background(), tasks, func(context.Context, string) (map[string]any, error) {
		t.Fatalf("fetch should not be called")
		return nil, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := tasks[0]
	if got["resolved_projects_status"] != "direct" {
		t.Fatalf("expected direct status, got %v", got["resolved_projects_status"])
	}
	if got["resolved_from_task_gid"] != "1" {
		t.Fatalf("expected source gid=1, got %v", got["resolved_from_task_gid"])
	}
}

func TestResolveTaskProjects_Ancestor(t *testing.T) {
	tasks := []map[string]any{
		{
			"gid":         "1",
			"name":        "child",
			"parent":      map[string]any{"gid": "2", "name": "parent"},
			"projects":    []any{},
			"memberships": []any{},
		},
	}
	fetchCount := 0
	err := ResolveTaskProjects(context.Background(), tasks, func(_ context.Context, gid string) (map[string]any, error) {
		fetchCount++
		if gid != "2" {
			t.Fatalf("unexpected gid fetch: %s", gid)
		}
		return map[string]any{
			"gid":         "2",
			"name":        "parent",
			"parent":      nil,
			"projects":    []any{map[string]any{"gid": "p1", "name": "Project 1"}},
			"memberships": []any{},
		}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetchCount != 1 {
		t.Fatalf("expected 1 fetch, got %d", fetchCount)
	}
	got := tasks[0]
	if got["resolved_projects_status"] != "ancestor" {
		t.Fatalf("expected ancestor status, got %v", got["resolved_projects_status"])
	}
	if got["resolved_from_task_gid"] != "2" {
		t.Fatalf("expected source gid=2, got %v", got["resolved_from_task_gid"])
	}
	if got["resolved_from_depth"] != 1 {
		t.Fatalf("expected depth=1, got %v", got["resolved_from_depth"])
	}
}

func TestResolveTaskProjects_FetchesPartialRoot(t *testing.T) {
	tasks := []map[string]any{
		{
			"gid":  "1",
			"name": "child",
		},
	}
	err := ResolveTaskProjects(context.Background(), tasks, func(_ context.Context, gid string) (map[string]any, error) {
		if gid != "1" {
			t.Fatalf("unexpected gid fetch: %s", gid)
		}
		return map[string]any{
			"gid":         "1",
			"name":        "child",
			"parent":      nil,
			"projects":    []any{map[string]any{"gid": "p1", "name": "Project 1"}},
			"memberships": []any{},
		}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tasks[0]["resolved_projects_status"] != "direct" {
		t.Fatalf("expected direct status, got %v", tasks[0]["resolved_projects_status"])
	}
}

func TestResolveTaskProjects_NotFound(t *testing.T) {
	tasks := []map[string]any{
		{
			"gid":         "1",
			"name":        "child",
			"parent":      nil,
			"projects":    []any{},
			"memberships": []any{},
		},
	}
	err := ResolveTaskProjects(context.Background(), tasks, func(context.Context, string) (map[string]any, error) {
		t.Fatalf("fetch should not be called")
		return nil, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tasks[0]["resolved_projects_status"] != "not_found" {
		t.Fatalf("expected not_found, got %v", tasks[0]["resolved_projects_status"])
	}
}

func TestResolveTaskProjects_MissingParent(t *testing.T) {
	tasks := []map[string]any{
		{
			"gid":         "1",
			"name":        "child",
			"parent":      map[string]any{"gid": "2", "name": "parent"},
			"projects":    []any{},
			"memberships": []any{},
		},
	}
	err := ResolveTaskProjects(context.Background(), tasks, func(_ context.Context, gid string) (map[string]any, error) {
		return nil, errs.New("api_not_found", "not found", gid)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tasks[0]["resolved_projects_status"] != "missing_parent" {
		t.Fatalf("expected missing_parent, got %v", tasks[0]["resolved_projects_status"])
	}
}
