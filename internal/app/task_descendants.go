package app

import (
	"context"

	"github.com/cloudnative-co/asana-cli/internal/errs"
)

const IncludeSubtasksDescendants = "descendants"

var taskDescendantExpansionFields = []string{
	"gid",
	"name",
	"parent.gid",
	"parent.name",
	"num_subtasks",
}

type TaskSubtaskListFunc func(context.Context, string) ([]map[string]any, error)

func TaskDescendantExpansionFields() []string {
	out := make([]string, 0, len(taskDescendantExpansionFields))
	out = append(out, taskDescendantExpansionFields...)
	return out
}

func ExpandTaskDescendants(ctx context.Context, tasks []map[string]any, listSubtasks TaskSubtaskListFunc) error {
	cache := map[string][]map[string]any{}
	for _, task := range tasks {
		descendants, err := collectTaskDescendants(ctx, task, task, 1, listSubtasks, cache, map[string]struct{}{})
		if err != nil {
			return err
		}
		task["descendant_subtasks"] = mapsFrom(descendants)
		task["descendant_subtasks_count"] = len(descendants)
	}
	return nil
}

func collectTaskDescendants(
	ctx context.Context,
	root map[string]any,
	current map[string]any,
	depth int,
	listSubtasks TaskSubtaskListFunc,
	cache map[string][]map[string]any,
	seen map[string]struct{},
) ([]map[string]any, error) {
	currentGID := stringValue(current["gid"])
	if currentGID == "" {
		return nil, errs.New("invalid_argument", "task gid is required for subtask expansion", "")
	}
	if _, ok := seen[currentGID]; ok {
		return nil, nil
	}
	seen[currentGID] = struct{}{}

	rawSubtasks, err := getOrListSubtasks(ctx, current, listSubtasks, cache)
	if err != nil {
		return nil, err
	}
	if len(rawSubtasks) == 0 {
		return nil, nil
	}

	rootGID := stringValue(root["gid"])
	rootName := stringValue(root["name"])
	out := make([]map[string]any, 0, len(rawSubtasks))
	for _, raw := range rawSubtasks {
		childGID := stringValue(raw["gid"])
		if childGID == "" {
			continue
		}
		if _, ok := seen[childGID]; ok {
			continue
		}

		child := cloneTaskMap(raw)
		child["expanded_from_task_gid"] = rootGID
		child["expanded_from_task_name"] = rootName
		child["subtask_depth"] = depth
		out = append(out, child)

		descendants, err := collectTaskDescendants(ctx, root, child, depth+1, listSubtasks, cache, seen)
		if err != nil {
			return nil, err
		}
		out = append(out, descendants...)
	}
	return out, nil
}

func getOrListSubtasks(
	ctx context.Context,
	task map[string]any,
	listSubtasks TaskSubtaskListFunc,
	cache map[string][]map[string]any,
) ([]map[string]any, error) {
	gid := stringValue(task["gid"])
	if gid == "" {
		return nil, errs.New("invalid_argument", "task gid is required for subtask listing", "")
	}
	if count, ok := taskSubtaskCount(task); ok && count <= 0 {
		return nil, nil
	}
	if cached, ok := cache[gid]; ok {
		return cached, nil
	}

	subtasks, err := listSubtasks(ctx, gid)
	if err != nil {
		return nil, err
	}
	cache[gid] = cloneTaskMaps(subtasks)
	return cache[gid], nil
}

func taskSubtaskCount(task map[string]any) (int, bool) {
	switch typed := task["num_subtasks"].(type) {
	case float64:
		return int(typed), true
	case float32:
		return int(typed), true
	case int:
		return typed, true
	case int64:
		return int(typed), true
	default:
		return 0, false
	}
}

func cloneTaskMaps(in []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(in))
	for _, item := range in {
		out = append(out, cloneTaskMap(item))
	}
	return out
}

func cloneTaskMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
