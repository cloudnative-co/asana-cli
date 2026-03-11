package cli

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cloudnative-co/asana-cli/internal/app"
	"github.com/cloudnative-co/asana-cli/internal/asanaapi"
	"github.com/cloudnative-co/asana-cli/internal/errs"
	"github.com/cloudnative-co/asana-cli/internal/output"
)

func NewTaskCommand(provider RuntimeProvider) *cobra.Command {
	cmd := newResourceCommand("task", "Task operations (official tasks endpoints)", asanaapi.TaskEndpoints, provider)
	var verbose bool
	var jsonCompat bool
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "compat mode: include task stories")
	cmd.Flags().BoolVarP(&jsonCompat, "json", "j", false, "compat mode: force json output")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errs.New("invalid_argument", "task compat mode accepts at most one argument", "use `asana task get --task-gid ...` for explicit API usage")
		}
		rt, err := provider()
		if err != nil {
			return err
		}
		ref := ""
		if len(args) == 1 {
			ref = args[0]
		}
		return runCompatTaskShow(rt, ref, verbose, jsonCompat)
	}
	return cmd
}

func NewProjectCommand(provider RuntimeProvider) *cobra.Command {
	return newResourceCommand("project", "Project operations (official projects endpoints)", asanaapi.ProjectEndpoints, provider)
}

func NewUserCommand(provider RuntimeProvider) *cobra.Command {
	command := newResourceCommand("user", "User operations (official users endpoints)", asanaapi.UserEndpoints, provider)
	for i, child := range command.Commands() {
		if child.Name() == "list" {
			command.RemoveCommand(child)
			command.AddCommand(newUserListCommand(provider))
			if i == 0 {
				break
			}
		}
	}
	return command
}

func NewAttachmentCommand(provider RuntimeProvider) *cobra.Command {
	return newResourceCommand("attachment", "Attachment operations", asanaapi.AttachmentEndpoints, provider)
}

func NewStoryCommand(provider RuntimeProvider) *cobra.Command {
	return newResourceCommand("story", "Story operations", asanaapi.StoryEndpoints, provider)
}

func NewTagCommand(provider RuntimeProvider) *cobra.Command {
	return newResourceCommand("tag", "Tag operations", asanaapi.TagEndpoints, provider)
}

func NewSectionCommand(provider RuntimeProvider) *cobra.Command {
	return newResourceCommand("section", "Section operations", asanaapi.SectionEndpoints, provider)
}

func NewCustomFieldCommand(provider RuntimeProvider) *cobra.Command {
	return newResourceCommand("custom-field", "Custom field operations", asanaapi.CustomFieldEndpoints, provider)
}

func NewUserTaskListCommand(provider RuntimeProvider) *cobra.Command {
	return newResourceCommand("user-task-list", "User task list operations", asanaapi.UserTaskListEndpoints, provider)
}

func NewTimeEntryCommand(provider RuntimeProvider) *cobra.Command {
	return newResourceCommand("time-entry", "Time tracking entry operations", asanaapi.TimeTrackingEntryEndpoints, provider)
}

func newResourceCommand(name, short string, endpoints []asanaapi.Endpoint, provider RuntimeProvider) *cobra.Command {
	root := &cobra.Command{Use: name, Short: short, SilenceUsage: true}
	for _, endpoint := range endpoints {
		root.AddCommand(newEndpointSubcommand(name, endpoint, provider))
	}
	return root
}

func newEndpointSubcommand(resourceName string, endpoint asanaapi.Endpoint, provider RuntimeProvider) *cobra.Command {
	var queryFlags []string
	var fieldFlags []string
	var jsonData string
	var autoPaginate bool
	var nameContains string
	var nameRegex string
	var assignee string
	var workspace string
	var resolveProjects string
	var includeSubtasks string

	pathParamNames := placeholders(endpoint.Path)
	pathFlagValues := map[string]*string{}

	command := &cobra.Command{
		Use:   endpoint.Name,
		Short: endpoint.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, profileName, clientErr := rt.NewClient(ctx)
			if clientErr != nil {
				return clientErr
			}

			query, queryErr := parseKeyValueFlags(queryFlags)
			if queryErr != nil {
				return queryErr
			}
			if shouldUseTaskAssigneeFlag(resourceName, endpoint) {
				if strings.TrimSpace(assignee) != "" {
					query["assignee"] = strings.TrimSpace(assignee)
				}
				if strings.TrimSpace(workspace) != "" {
					query["workspace"] = strings.TrimSpace(workspace)
				} else if strings.TrimSpace(query["workspace"]) == "" && strings.TrimSpace(assignee) != "" {
					if profileCfg, profileExists, profileErr := rt.GetProfile(profileName); profileErr == nil && profileExists && profileCfg.Workspace != "" {
						query["workspace"] = profileCfg.Workspace
					}
				}
			}
			if shouldSupportTaskSubtaskExpansion(resourceName, endpoint) && strings.TrimSpace(includeSubtasks) != "" {
				if strings.TrimSpace(includeSubtasks) != app.IncludeSubtasksDescendants {
					return errs.New("invalid_argument", "unsupported --include-subtasks mode", "supported: descendants")
				}
				query["opt_fields"] = app.MergeOptFields(query["opt_fields"], app.TaskDescendantExpansionFields()...)
			}
			if shouldSupportTaskProjectResolution(resourceName, endpoint) && strings.TrimSpace(resolveProjects) != "" {
				if strings.TrimSpace(resolveProjects) != app.ResolveProjectsAncestors {
					return errs.New("invalid_argument", "unsupported --resolve-projects mode", "supported: ancestors")
				}
				query["opt_fields"] = app.MergeOptFields(query["opt_fields"], app.TaskProjectResolutionFields()...)
			}
			pathValues := map[string]string{}
			for name, valuePtr := range pathFlagValues {
				pathValues[name] = strings.TrimSpace(*valuePtr)
			}

			if workspaceValue, ok := pathValues["workspace_gid"]; ok && workspaceValue == "" {
				if profileCfg, profileExists, profileErr := rt.GetProfile(profileName); profileErr == nil && profileExists && profileCfg.Workspace != "" {
					pathValues["workspace_gid"] = profileCfg.Workspace
				}
			}

			resolvedPath, fillErr := fillPath(endpoint.Path, pathValues)
			if fillErr != nil {
				return fillErr
			}

			body, bodyErr := parseBodyFromFlags(jsonData, fieldFlags)
			if bodyErr != nil {
				return bodyErr
			}

			method := strings.ToUpper(endpoint.Method)
			if method == "GET" && autoPaginate && shouldSetDefaultLimit(endpoint, query) {
				query["limit"] = "100"
			}
			if rt.Options.DryRun && method != "GET" {
				payload := map[string]any{
					"schema_version": "v1",
					"dry_run":        true,
					"method":         method,
					"path":           resolvedPath,
					"query":          query,
					"body":           body,
				}
				format, formatErr := rt.EffectiveOutput(profileName)
				if formatErr != nil {
					return formatErr
				}
				return output.Render(payload, format, rt.Options.OutputPath)
			}

			response, requestErr := client.Request(ctx, method, resolvedPath, query, body, autoPaginate && method == "GET")
			if requestErr != nil {
				return requestErr
			}
			if shouldSupportTaskSubtaskExpansion(resourceName, endpoint) && strings.TrimSpace(includeSubtasks) != "" {
				if expansionErr := applyTaskSubtaskExpansion(ctx, client, response, query); expansionErr != nil {
					return expansionErr
				}
			}
			if shouldSupportTaskProjectResolution(resourceName, endpoint) && strings.TrimSpace(resolveProjects) != "" {
				if resolutionErr := applyTaskProjectResolution(ctx, client, response); resolutionErr != nil {
					return resolutionErr
				}
			}
			if filterErr := applyNameFilter(response, nameContains, nameRegex); filterErr != nil {
				return filterErr
			}
			format, formatErr := rt.EffectiveOutput(profileName)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(response, format, rt.Options.OutputPath)
		},
		SilenceUsage: true,
	}

	for _, pathName := range pathParamNames {
		var value string
		flagName := normalizeFlagName(pathName)
		command.Flags().StringVar(&value, flagName, "", fmt.Sprintf("path param: %s", pathName))
		pathFlagValues[pathName] = &value
	}
	command.Flags().StringArrayVar(&queryFlags, "query", nil, "query parameter as key=value (repeatable)")
	command.Flags().StringArrayVar(&fieldFlags, "field", nil, "request data field as key=value, value supports JSON literals")
	command.Flags().StringVar(&jsonData, "data", "", "request body JSON object")
	command.Flags().BoolVar(&autoPaginate, "all", shouldAutoPaginateByDefault(endpoint), "auto-follow pagination for list endpoints")
	if shouldUseTaskAssigneeFlag(resourceName, endpoint) {
		command.Flags().StringVar(&assignee, "assignee", "", "task assignee gid or me")
		command.Flags().StringVar(&workspace, "workspace", "", "workspace gid for task list queries")
	}
	if isListLikeEndpoint(endpoint) {
		command.Flags().StringVar(&nameContains, "name-contains", "", "local filter: include rows where name contains this value (case-insensitive)")
		command.Flags().StringVar(&nameRegex, "name-regex", "", "local filter: include rows where name matches this regular expression")
	}
	if shouldSupportTaskProjectResolution(resourceName, endpoint) {
		command.Flags().StringVar(&resolveProjects, "resolve-projects", "", "resolve task projects after fetch (supported: ancestors)")
	}
	if shouldSupportTaskSubtaskExpansion(resourceName, endpoint) {
		command.Flags().StringVar(&includeSubtasks, "include-subtasks", "", "expand descendant subtasks after fetch (supported: descendants)")
	}

	return command
}

func newUserListCommand(provider RuntimeProvider) *cobra.Command {
	var workspaceGID string
	var teamGID string
	var domainFilters []string
	var includeGuests bool
	var includeDeactivated bool
	var queryFlags []string
	var autoPaginate bool

	command := &cobra.Command{
		Use:   "list",
		Short: "List users (supports workspace/team scope and domain filters)",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			client, profileName, clientErr := rt.NewClient(context.Background())
			if clientErr != nil {
				return clientErr
			}
			query, queryErr := parseKeyValueFlags(queryFlags)
			if queryErr != nil {
				return queryErr
			}
			query["opt_fields"] = "gid,name,email,is_active,is_guest"
			path := "/users"

			if strings.TrimSpace(workspaceGID) == "" {
				if profileCfg, ok, profileErr := rt.GetProfile(profileName); profileErr == nil && ok && profileCfg.Workspace != "" {
					workspaceGID = profileCfg.Workspace
				}
			}
			if strings.TrimSpace(teamGID) != "" {
				path = strings.ReplaceAll("/teams/{team_gid}/users", "{team_gid}", strings.TrimSpace(teamGID))
			} else if strings.TrimSpace(workspaceGID) != "" {
				path = strings.ReplaceAll("/workspaces/{workspace_gid}/users", "{workspace_gid}", strings.TrimSpace(workspaceGID))
			}

			resp, reqErr := client.Request(context.Background(), "GET", path, query, nil, autoPaginate)
			if reqErr != nil {
				return reqErr
			}
			rawData, _ := resp["data"].([]any)
			users := make([]map[string]any, 0, len(rawData))
			for _, item := range rawData {
				row, ok := item.(map[string]any)
				if ok {
					users = append(users, row)
				}
			}

			filtered := app.FilterUsers(users, app.UserFilterOptions{
				Domains:            domainFilters,
				IncludeGuests:      includeGuests,
				IncludeDeactivated: includeDeactivated,
			})
			resp["data"] = mapsFrom(filtered)
			resp["filters"] = map[string]any{
				"domains":             domainFilters,
				"include_guests":      includeGuests,
				"include_deactivated": includeDeactivated,
				"workspace_gid":       workspaceGID,
				"team_gid":            teamGID,
			}

			format, formatErr := rt.EffectiveOutput(profileName)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(resp, format, rt.Options.OutputPath)
		},
		SilenceUsage: true,
	}

	command.Flags().StringVar(&workspaceGID, "workspace", "", "workspace gid scope")
	command.Flags().StringVar(&teamGID, "team", "", "team gid scope")
	command.Flags().StringArrayVar(&domainFilters, "domain", nil, "exact email domain filter (repeatable)")
	command.Flags().BoolVar(&includeGuests, "include-guests", false, "include guest users")
	command.Flags().BoolVar(&includeDeactivated, "include-deactivated", false, "include deactivated users")
	command.Flags().StringArrayVar(&queryFlags, "query", nil, "additional query parameter as key=value")
	command.Flags().BoolVar(&autoPaginate, "all", true, "auto-follow pagination")
	return command
}

func mapsFrom(in []map[string]any) []any {
	out := make([]any, 0, len(in))
	for _, item := range in {
		out = append(out, item)
	}
	return out
}

func shouldUseTaskAssigneeFlag(resourceName string, endpoint asanaapi.Endpoint) bool {
	return resourceName == "task" && endpoint.Name == "list"
}

func shouldSupportTaskProjectResolution(resourceName string, endpoint asanaapi.Endpoint) bool {
	return shouldSupportTaskAugmentation(resourceName, endpoint)
}

func shouldSupportTaskSubtaskExpansion(resourceName string, endpoint asanaapi.Endpoint) bool {
	return shouldSupportTaskAugmentation(resourceName, endpoint)
}

func shouldSupportTaskAugmentation(resourceName string, endpoint asanaapi.Endpoint) bool {
	if resourceName != "task" {
		return false
	}
	switch endpoint.Name {
	case "list",
		"create",
		"get",
		"update",
		"list-project",
		"list-section",
		"list-tag",
		"list-user-task-list",
		"list-subtasks",
		"create-subtask",
		"list-dependencies",
		"list-dependents",
		"get-by-custom-id",
		"search-workspace":
		return true
	default:
		return false
	}
}

func shouldAutoPaginateByDefault(endpoint asanaapi.Endpoint) bool {
	if strings.ToUpper(endpoint.Method) != "GET" {
		return false
	}
	return isListLikeEndpoint(endpoint)
}

func shouldSetDefaultLimit(endpoint asanaapi.Endpoint, query map[string]string) bool {
	if !isListLikeEndpoint(endpoint) {
		return false
	}
	if _, ok := query["limit"]; ok {
		return false
	}
	return true
}

func isListLikeEndpoint(endpoint asanaapi.Endpoint) bool {
	name := strings.ToLower(strings.TrimSpace(endpoint.Name))
	if strings.HasPrefix(name, "list") || strings.HasPrefix(name, "search") {
		return true
	}
	return name == "favorites"
}

func applyTaskSubtaskExpansion(
	ctx context.Context,
	client *asanaapi.Client,
	response map[string]any,
	rootQuery map[string]string,
) error {
	tasks, ok := extractTaskMaps(response["data"])
	if !ok {
		return nil
	}
	listSubtasks := func(ctx context.Context, taskGID string) ([]map[string]any, error) {
		resp, err := client.Request(ctx, "GET", "/tasks/"+taskGID+"/subtasks", taskSubtaskQuery(rootQuery), nil, true)
		if err != nil {
			return nil, err
		}
		rows, ok := extractTaskMaps(resp["data"])
		if !ok {
			return nil, errs.New("internal_error", "unexpected subtask expansion response shape", "")
		}
		return rows, nil
	}
	if err := app.ExpandTaskDescendants(ctx, tasks, listSubtasks); err != nil {
		return err
	}

	if _, isList := response["data"].([]any); isList {
		flattened := flattenTaskListWithDescendants(tasks)
		response["data"] = mapsFrom(flattened)
		response["subtasks"] = map[string]any{
			"mode":             app.IncludeSubtasksDescendants,
			"before_count":     len(tasks),
			"after_count":      len(flattened),
			"descendant_count": len(flattened) - len(tasks),
		}
	}
	return nil
}

func applyTaskProjectResolution(ctx context.Context, client *asanaapi.Client, response map[string]any) error {
	tasks, ok := extractAllTaskMaps(response["data"])
	if !ok {
		return nil
	}
	rootTaskGIDs := taskGIDs(tasks)
	batchCache := map[string]map[string]any{}
	batchFailures := map[string]error{}
	rootsPrefetched := false
	fetch := func(ctx context.Context, taskGID string) (map[string]any, error) {
		if task, ok := batchCache[taskGID]; ok {
			return task, nil
		}
		if err, ok := batchFailures[taskGID]; ok {
			return nil, err
		}
		if !rootsPrefetched {
			prefetched, failures, err := client.BatchGetTasks(ctx, rootTaskGIDs, app.TaskProjectResolutionFields())
			if err != nil {
				return nil, err
			}
			for gid, task := range prefetched {
				batchCache[gid] = task
			}
			for gid, fetchErr := range failures {
				batchFailures[gid] = fetchErr
			}
			rootsPrefetched = true
			if task, ok := batchCache[taskGID]; ok {
				return task, nil
			}
			if err, ok := batchFailures[taskGID]; ok {
				return nil, err
			}
		}
		resp, err := client.Request(ctx, "GET", "/tasks/"+taskGID, map[string]string{
			"opt_fields": app.MergeOptFields("", app.TaskProjectResolutionFields()...),
		}, nil, false)
		if err != nil {
			return nil, err
		}
		data, ok := resp["data"].(map[string]any)
		if !ok {
			return nil, errs.New("internal_error", "unexpected task resolution response shape", "")
		}
		return data, nil
	}
	return app.ResolveTaskProjects(ctx, tasks, fetch)
}

func taskSubtaskQuery(rootQuery map[string]string) map[string]string {
	query := map[string]string{
		"opt_fields": app.MergeOptFields(rootQuery["opt_fields"], app.TaskDescendantExpansionFields()...),
	}
	if completedSince := strings.TrimSpace(rootQuery["completed_since"]); completedSince != "" {
		query["completed_since"] = completedSince
	}
	return query
}

func taskGIDs(tasks []map[string]any) []string {
	out := make([]string, 0, len(tasks))
	for _, task := range tasks {
		gid := strings.TrimSpace(stringValue(task["gid"]))
		if gid == "" {
			continue
		}
		out = append(out, gid)
	}
	return out
}

func extractTaskMaps(data any) ([]map[string]any, bool) {
	switch typed := data.(type) {
	case map[string]any:
		if !looksLikeTask(typed) {
			return nil, false
		}
		return []map[string]any{typed}, true
	case []any:
		tasks := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			task, ok := item.(map[string]any)
			if !ok || !looksLikeTask(task) {
				return nil, false
			}
			tasks = append(tasks, task)
		}
		return tasks, true
	default:
		return nil, false
	}
}

func extractAllTaskMaps(data any) ([]map[string]any, bool) {
	roots, ok := extractTaskMaps(data)
	if !ok {
		return nil, false
	}
	out := make([]map[string]any, 0, len(roots))
	seen := map[string]struct{}{}
	for _, root := range roots {
		collectTaskMapsRecursive(root, seen, &out)
	}
	return out, true
}

func collectTaskMapsRecursive(task map[string]any, seen map[string]struct{}, out *[]map[string]any) {
	if task == nil {
		return
	}
	gid := strings.TrimSpace(stringValue(task["gid"]))
	if gid == "" {
		return
	}
	if _, ok := seen[gid]; ok {
		return
	}
	seen[gid] = struct{}{}
	*out = append(*out, task)

	rawDescendants, _ := task["descendant_subtasks"].([]any)
	for _, item := range rawDescendants {
		descendant, _ := item.(map[string]any)
		collectTaskMapsRecursive(descendant, seen, out)
	}
}

func flattenTaskListWithDescendants(tasks []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(tasks))
	seen := map[string]struct{}{}

	appendRow := func(row map[string]any) {
		if row == nil {
			return
		}
		gid := strings.TrimSpace(stringValue(row["gid"]))
		if gid == "" {
			return
		}
		if _, ok := seen[gid]; ok {
			return
		}
		seen[gid] = struct{}{}
		copied := copyTaskRow(row)
		delete(copied, "descendant_subtasks")
		out = append(out, copied)
	}

	for _, task := range tasks {
		appendRow(task)
		rawDescendants, _ := task["descendant_subtasks"].([]any)
		for _, item := range rawDescendants {
			descendant, _ := item.(map[string]any)
			appendRow(descendant)
		}
	}

	return out
}

func copyTaskRow(task map[string]any) map[string]any {
	out := make(map[string]any, len(task))
	for key, value := range task {
		out[key] = value
	}
	return out
}

func looksLikeTask(task map[string]any) bool {
	if task == nil {
		return false
	}
	gid := strings.TrimSpace(stringValue(task["gid"]))
	return gid != ""
}

func stringValue(v any) string {
	s, _ := v.(string)
	return s
}

func applyNameFilter(response map[string]any, nameContains string, nameRegex string) error {
	contains := strings.TrimSpace(nameContains)
	regexValue := strings.TrimSpace(nameRegex)
	if contains == "" && regexValue == "" {
		return nil
	}

	rawData, ok := response["data"].([]any)
	if !ok {
		return errs.New("invalid_argument", "--name-contains/--name-regex are supported only for list responses", "")
	}
	rows := make([]map[string]any, 0, len(rawData))
	for _, item := range rawData {
		row, rowOK := item.(map[string]any)
		if rowOK {
			rows = append(rows, row)
		}
	}

	var compiled *regexp.Regexp
	if regexValue != "" {
		regex, err := regexp.Compile(regexValue)
		if err != nil {
			return errs.Wrap("invalid_argument", "failed to parse --name-regex", "", err)
		}
		compiled = regex
	}

	filtered := app.FilterByName(rows, app.NameFilterOptions{
		Contains: contains,
		Regex:    compiled,
	})
	response["data"] = mapsFrom(filtered)
	response["filters"] = map[string]any{
		"name_contains": contains,
		"name_regex":    regexValue,
		"before_count":  len(rows),
		"after_count":   len(filtered),
	}
	return nil
}
