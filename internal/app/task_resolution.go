package app

import (
	"context"
	"strings"

	"github.com/cloudnative-co/asana-cli/internal/errs"
)

const ResolveProjectsAncestors = "ancestors"

var taskProjectResolutionFields = []string{
	"gid",
	"name",
	"parent.gid",
	"parent.name",
	"projects.gid",
	"projects.name",
	"memberships.project.gid",
	"memberships.project.name",
}

type TaskFetchFunc func(context.Context, string) (map[string]any, error)

type TaskProjectResolution struct {
	Projects       []map[string]any
	SourceTaskGID  string
	SourceTaskName string
	Depth          int
	Status         string
}

func TaskProjectResolutionFields() []string {
	out := make([]string, 0, len(taskProjectResolutionFields))
	out = append(out, taskProjectResolutionFields...)
	return out
}

func MergeOptFields(existing string, required ...string) string {
	seen := map[string]struct{}{}
	ordered := make([]string, 0, len(required)+1)
	for _, raw := range strings.Split(existing, ",") {
		field := strings.TrimSpace(raw)
		if field == "" {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		ordered = append(ordered, field)
	}
	for _, raw := range required {
		field := strings.TrimSpace(raw)
		if field == "" {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		ordered = append(ordered, field)
	}
	return strings.Join(ordered, ",")
}

func ResolveTaskProjects(ctx context.Context, tasks []map[string]any, fetch TaskFetchFunc) error {
	cache := map[string]map[string]any{}
	for _, task := range tasks {
		if gid := stringValue(task["gid"]); gid != "" {
			cache[gid] = task
		}
	}
	for _, task := range tasks {
		resolution, err := resolveTaskProjects(ctx, task, fetch, cache, map[string]struct{}{})
		if err != nil {
			return err
		}
		task["resolved_projects"] = mapsFrom(resolution.Projects)
		task["resolved_from_task_gid"] = resolution.SourceTaskGID
		task["resolved_from_task_name"] = resolution.SourceTaskName
		task["resolved_from_depth"] = resolution.Depth
		task["resolved_projects_status"] = resolution.Status
	}
	return nil
}

func resolveTaskProjects(
	ctx context.Context,
	task map[string]any,
	fetch TaskFetchFunc,
	cache map[string]map[string]any,
	seen map[string]struct{},
) (TaskProjectResolution, error) {
	current, err := ensureTaskProjectContext(ctx, task, fetch, cache)
	if err != nil {
		return TaskProjectResolution{}, err
	}

	currentGID := stringValue(current["gid"])
	if currentGID != "" {
		if _, ok := seen[currentGID]; ok {
			return TaskProjectResolution{
				Projects:       []map[string]any{},
				SourceTaskGID:  currentGID,
				SourceTaskName: stringValue(current["name"]),
				Depth:          len(seen),
				Status:         "cycle",
			}, nil
		}
		seen[currentGID] = struct{}{}
	}

	projects := taskProjects(current)
	if len(projects) > 0 {
		status := "direct"
		depth := len(seen) - 1
		if depth > 0 {
			status = "ancestor"
		}
		return TaskProjectResolution{
			Projects:       projects,
			SourceTaskGID:  currentGID,
			SourceTaskName: stringValue(current["name"]),
			Depth:          depth,
			Status:         status,
		}, nil
	}

	parent, hasParent := current["parent"]
	if !hasParent || parent == nil {
		return TaskProjectResolution{
			Projects:       []map[string]any{},
			SourceTaskGID:  currentGID,
			SourceTaskName: stringValue(current["name"]),
			Depth:          len(seen) - 1,
			Status:         "not_found",
		}, nil
	}
	parentMap, _ := parent.(map[string]any)
	parentGID := stringValue(parentMap["gid"])
	if parentGID == "" {
		return TaskProjectResolution{
			Projects:       []map[string]any{},
			SourceTaskGID:  currentGID,
			SourceTaskName: stringValue(current["name"]),
			Depth:          len(seen) - 1,
			Status:         "not_found",
		}, nil
	}

	parentTask, err := getOrFetchTask(ctx, parentGID, fetch, cache)
	if err != nil {
		m := errs.AsMachine(err)
		if m.Code == "api_not_found" {
			return TaskProjectResolution{
				Projects:       []map[string]any{},
				SourceTaskGID:  parentGID,
				SourceTaskName: stringValue(parentMap["name"]),
				Depth:          len(seen),
				Status:         "missing_parent",
			}, nil
		}
		return TaskProjectResolution{}, err
	}
	return resolveTaskProjects(ctx, parentTask, fetch, cache, seen)
}

func ensureTaskProjectContext(
	ctx context.Context,
	task map[string]any,
	fetch TaskFetchFunc,
	cache map[string]map[string]any,
) (map[string]any, error) {
	if hasTaskProjectContext(task) {
		return task, nil
	}
	gid := stringValue(task["gid"])
	if gid == "" {
		return task, errs.New("invalid_argument", "task gid is required for project resolution", "")
	}
	return getOrFetchTask(ctx, gid, fetch, cache)
}

func getOrFetchTask(
	ctx context.Context,
	gid string,
	fetch TaskFetchFunc,
	cache map[string]map[string]any,
) (map[string]any, error) {
	if cached, ok := cache[gid]; ok && hasTaskProjectContext(cached) {
		return cached, nil
	}
	fetched, err := fetch(ctx, gid)
	if err != nil {
		return nil, err
	}
	cache[gid] = fetched
	return fetched, nil
}

func hasTaskProjectContext(task map[string]any) bool {
	_, hasProjects := task["projects"]
	_, hasMemberships := task["memberships"]
	_, hasParent := task["parent"]
	return hasProjects && hasMemberships && hasParent
}

func taskProjects(task map[string]any) []map[string]any {
	seen := map[string]struct{}{}
	out := []map[string]any{}

	appendProject := func(project map[string]any) {
		gid := stringValue(project["gid"])
		if gid == "" {
			return
		}
		if _, ok := seen[gid]; ok {
			return
		}
		seen[gid] = struct{}{}
		out = append(out, map[string]any{
			"gid":  gid,
			"name": stringValue(project["name"]),
		})
	}

	if rawProjects, ok := task["projects"].([]any); ok {
		for _, item := range rawProjects {
			project, _ := item.(map[string]any)
			if project != nil {
				appendProject(project)
			}
		}
	}

	if rawMemberships, ok := task["memberships"].([]any); ok {
		for _, item := range rawMemberships {
			membership, _ := item.(map[string]any)
			if membership == nil {
				continue
			}
			project, _ := membership["project"].(map[string]any)
			if project != nil {
				appendProject(project)
			}
		}
	}

	return out
}

func stringValue(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func mapsFrom(in []map[string]any) []any {
	out := make([]any, 0, len(in))
	for _, item := range in {
		out = append(out, item)
	}
	return out
}
