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
		root.AddCommand(newEndpointSubcommand(endpoint, provider))
	}
	return root
}

func newEndpointSubcommand(endpoint asanaapi.Endpoint, provider RuntimeProvider) *cobra.Command {
	var queryFlags []string
	var fieldFlags []string
	var jsonData string
	var autoPaginate bool
	var nameContains string
	var nameRegex string

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
	if isListLikeEndpoint(endpoint) {
		command.Flags().StringVar(&nameContains, "name-contains", "", "local filter: include rows where name contains this value (case-insensitive)")
		command.Flags().StringVar(&nameRegex, "name-regex", "", "local filter: include rows where name matches this regular expression")
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
