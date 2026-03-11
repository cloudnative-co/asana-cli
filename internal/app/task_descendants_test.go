package app

import (
	"context"
	"testing"
)

func TestExpandTaskDescendants(t *testing.T) {
	tasks := []map[string]any{
		{
			"gid":          "1",
			"name":         "root",
			"num_subtasks": 1,
		},
	}

	err := ExpandTaskDescendants(context.Background(), tasks, func(_ context.Context, gid string) ([]map[string]any, error) {
		switch gid {
		case "1":
			return []map[string]any{
				{
					"gid":          "2",
					"name":         "child",
					"parent":       map[string]any{"gid": "1", "name": "root"},
					"num_subtasks": 1,
				},
			}, nil
		case "2":
			return []map[string]any{
				{
					"gid":          "3",
					"name":         "grandchild",
					"parent":       map[string]any{"gid": "2", "name": "child"},
					"num_subtasks": 0,
				},
			}, nil
		default:
			t.Fatalf("unexpected list-subtasks gid: %s", gid)
			return nil, nil
		}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	root := tasks[0]
	if root["descendant_subtasks_count"] != 2 {
		t.Fatalf("expected 2 descendants, got %v", root["descendant_subtasks_count"])
	}
	rawDescendants, _ := root["descendant_subtasks"].([]any)
	if len(rawDescendants) != 2 {
		t.Fatalf("expected 2 descendant rows, got %d", len(rawDescendants))
	}
	first, _ := rawDescendants[0].(map[string]any)
	second, _ := rawDescendants[1].(map[string]any)
	if first["gid"] != "2" || first["subtask_depth"] != 1 {
		t.Fatalf("unexpected first descendant: %#v", first)
	}
	if second["gid"] != "3" || second["subtask_depth"] != 2 {
		t.Fatalf("unexpected second descendant: %#v", second)
	}
	if first["expanded_from_task_gid"] != "1" || second["expanded_from_task_name"] != "root" {
		t.Fatalf("unexpected expansion metadata: %#v %#v", first, second)
	}
}

func TestExpandTaskDescendants_SkipsKnownLeaf(t *testing.T) {
	tasks := []map[string]any{
		{
			"gid":          "1",
			"name":         "root",
			"num_subtasks": 0,
		},
	}

	err := ExpandTaskDescendants(context.Background(), tasks, func(context.Context, string) ([]map[string]any, error) {
		t.Fatalf("list-subtasks should not be called for num_subtasks=0")
		return nil, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tasks[0]["descendant_subtasks_count"] != 0 {
		t.Fatalf("expected no descendants, got %v", tasks[0]["descendant_subtasks_count"])
	}
}

func TestExpandTaskDescendants_ListsWhenCountMissing(t *testing.T) {
	tasks := []map[string]any{
		{
			"gid":  "1",
			"name": "root",
		},
	}

	listCalls := 0
	err := ExpandTaskDescendants(context.Background(), tasks, func(_ context.Context, gid string) ([]map[string]any, error) {
		listCalls++
		if gid != "1" {
			t.Fatalf("unexpected list-subtasks gid: %s", gid)
		}
		return []map[string]any{}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if listCalls != 1 {
		t.Fatalf("expected one list-subtasks call, got %d", listCalls)
	}
}
