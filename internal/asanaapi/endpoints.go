package asanaapi

// Endpoint defines a CLI-to-REST mapping.
type Endpoint struct {
	Name        string
	Method      string
	Path        string
	Description string
}

var TaskEndpoints = []Endpoint{
	{Name: "list", Method: "GET", Path: "/tasks", Description: "Get multiple tasks"},
	{Name: "create", Method: "POST", Path: "/tasks", Description: "Create a task"},
	{Name: "get", Method: "GET", Path: "/tasks/{task_gid}", Description: "Get a task"},
	{Name: "update", Method: "PUT", Path: "/tasks/{task_gid}", Description: "Update a task"},
	{Name: "delete", Method: "DELETE", Path: "/tasks/{task_gid}", Description: "Delete a task"},
	{Name: "duplicate", Method: "POST", Path: "/tasks/{task_gid}/duplicate", Description: "Duplicate a task"},
	{Name: "list-project", Method: "GET", Path: "/projects/{project_gid}/tasks", Description: "Get tasks from a project"},
	{Name: "list-section", Method: "GET", Path: "/sections/{section_gid}/tasks", Description: "Get tasks from a section"},
	{Name: "list-tag", Method: "GET", Path: "/tags/{tag_gid}/tasks", Description: "Get tasks from a tag"},
	{Name: "list-user-task-list", Method: "GET", Path: "/user_task_lists/{user_task_list_gid}/tasks", Description: "Get tasks from a user task list"},
	{Name: "list-subtasks", Method: "GET", Path: "/tasks/{task_gid}/subtasks", Description: "Get subtasks from a task"},
	{Name: "create-subtask", Method: "POST", Path: "/tasks/{task_gid}/subtasks", Description: "Create a subtask"},
	{Name: "set-parent", Method: "POST", Path: "/tasks/{task_gid}/setParent", Description: "Set parent task"},
	{Name: "list-dependencies", Method: "GET", Path: "/tasks/{task_gid}/dependencies", Description: "Get dependencies"},
	{Name: "add-dependencies", Method: "POST", Path: "/tasks/{task_gid}/addDependencies", Description: "Add dependencies"},
	{Name: "remove-dependencies", Method: "POST", Path: "/tasks/{task_gid}/removeDependencies", Description: "Remove dependencies"},
	{Name: "list-dependents", Method: "GET", Path: "/tasks/{task_gid}/dependents", Description: "Get dependents"},
	{Name: "add-dependents", Method: "POST", Path: "/tasks/{task_gid}/addDependents", Description: "Add dependents"},
	{Name: "remove-dependents", Method: "POST", Path: "/tasks/{task_gid}/removeDependents", Description: "Remove dependents"},
	{Name: "add-project", Method: "POST", Path: "/tasks/{task_gid}/addProject", Description: "Add project to task"},
	{Name: "remove-project", Method: "POST", Path: "/tasks/{task_gid}/removeProject", Description: "Remove project from task"},
	{Name: "add-tag", Method: "POST", Path: "/tasks/{task_gid}/addTag", Description: "Add tag to task"},
	{Name: "remove-tag", Method: "POST", Path: "/tasks/{task_gid}/removeTag", Description: "Remove tag from task"},
	{Name: "add-followers", Method: "POST", Path: "/tasks/{task_gid}/addFollowers", Description: "Add followers to task"},
	{Name: "remove-followers", Method: "POST", Path: "/tasks/{task_gid}/removeFollowers", Description: "Remove followers from task"},
	{Name: "get-by-custom-id", Method: "GET", Path: "/workspaces/{workspace_gid}/tasks/custom_id/{custom_id}", Description: "Get task by custom ID"},
	{Name: "search-workspace", Method: "GET", Path: "/workspaces/{workspace_gid}/tasks/search", Description: "Search tasks in workspace"},
	{Name: "list-time-tracking-entries", Method: "GET", Path: "/tasks/{task_gid}/time_tracking_entries", Description: "Get time tracking entries for a task"},
	{Name: "create-time-tracking-entry", Method: "POST", Path: "/tasks/{task_gid}/time_tracking_entries", Description: "Create a time tracking entry for a task"},
}

var ProjectEndpoints = []Endpoint{
	{Name: "list", Method: "GET", Path: "/projects", Description: "Get multiple projects"},
	{Name: "create", Method: "POST", Path: "/projects", Description: "Create project"},
	{Name: "get", Method: "GET", Path: "/projects/{project_gid}", Description: "Get project"},
	{Name: "update", Method: "PUT", Path: "/projects/{project_gid}", Description: "Update project"},
	{Name: "delete", Method: "DELETE", Path: "/projects/{project_gid}", Description: "Delete project"},
	{Name: "duplicate", Method: "POST", Path: "/projects/{project_gid}/duplicate", Description: "Duplicate project"},
	{Name: "list-for-task", Method: "GET", Path: "/tasks/{task_gid}/projects", Description: "Get projects for task"},
	{Name: "list-for-team", Method: "GET", Path: "/teams/{team_gid}/projects", Description: "Get projects for team"},
	{Name: "create-for-team", Method: "POST", Path: "/teams/{team_gid}/projects", Description: "Create project for team"},
	{Name: "list-for-workspace", Method: "GET", Path: "/workspaces/{workspace_gid}/projects", Description: "Get projects for workspace"},
	{Name: "create-for-workspace", Method: "POST", Path: "/workspaces/{workspace_gid}/projects", Description: "Create project for workspace"},
	{Name: "add-custom-field-setting", Method: "POST", Path: "/projects/{project_gid}/addCustomFieldSetting", Description: "Add custom field setting"},
	{Name: "remove-custom-field-setting", Method: "POST", Path: "/projects/{project_gid}/removeCustomFieldSetting", Description: "Remove custom field setting"},
	{Name: "task-counts", Method: "GET", Path: "/projects/{project_gid}/task_counts", Description: "Get task counts"},
	{Name: "add-members", Method: "POST", Path: "/projects/{project_gid}/addMembers", Description: "Add project members"},
	{Name: "remove-members", Method: "POST", Path: "/projects/{project_gid}/removeMembers", Description: "Remove project members"},
	{Name: "add-followers", Method: "POST", Path: "/projects/{project_gid}/addFollowers", Description: "Add project followers"},
	{Name: "remove-followers", Method: "POST", Path: "/projects/{project_gid}/removeFollowers", Description: "Remove project followers"},
	{Name: "save-as-template", Method: "POST", Path: "/projects/{project_gid}/saveAsTemplate", Description: "Save project as template"},
}

var UserEndpoints = []Endpoint{
	{Name: "list", Method: "GET", Path: "/users", Description: "Get multiple users"},
	{Name: "get", Method: "GET", Path: "/users/{user_gid}", Description: "Get a user"},
	{Name: "favorites", Method: "GET", Path: "/users/{user_gid}/favorites", Description: "Get user favorites"},
	{Name: "list-for-team", Method: "GET", Path: "/teams/{team_gid}/users", Description: "Get users in team"},
	{Name: "list-for-workspace", Method: "GET", Path: "/workspaces/{workspace_gid}/users", Description: "Get users in workspace"},
	{Name: "update", Method: "PUT", Path: "/users/{user_gid}", Description: "Update user"},
	{Name: "get-for-workspace", Method: "GET", Path: "/workspaces/{workspace_gid}/users/{user_gid}", Description: "Get user in workspace"},
	{Name: "update-for-workspace", Method: "PUT", Path: "/workspaces/{workspace_gid}/users/{user_gid}", Description: "Update user in workspace"},
}

var AttachmentEndpoints = []Endpoint{
	{Name: "list-for-object", Method: "GET", Path: "/attachments", Description: "Get attachments from an object"},
	{Name: "create", Method: "POST", Path: "/attachments", Description: "Create an attachment"},
	{Name: "get", Method: "GET", Path: "/attachments/{attachment_gid}", Description: "Get an attachment"},
	{Name: "delete", Method: "DELETE", Path: "/attachments/{attachment_gid}", Description: "Delete an attachment"},
}

var StoryEndpoints = []Endpoint{
	{Name: "get", Method: "GET", Path: "/stories/{story_gid}", Description: "Get a story"},
	{Name: "update", Method: "PUT", Path: "/stories/{story_gid}", Description: "Update a story"},
	{Name: "delete", Method: "DELETE", Path: "/stories/{story_gid}", Description: "Delete a story"},
	{Name: "list-for-task", Method: "GET", Path: "/tasks/{task_gid}/stories", Description: "Get stories from task"},
	{Name: "create-for-task", Method: "POST", Path: "/tasks/{task_gid}/stories", Description: "Create story on task"},
}

var TagEndpoints = []Endpoint{
	{Name: "list", Method: "GET", Path: "/tags", Description: "Get multiple tags"},
	{Name: "create", Method: "POST", Path: "/tags", Description: "Create a tag"},
	{Name: "get", Method: "GET", Path: "/tags/{tag_gid}", Description: "Get a tag"},
	{Name: "update", Method: "PUT", Path: "/tags/{tag_gid}", Description: "Update a tag"},
	{Name: "delete", Method: "DELETE", Path: "/tags/{tag_gid}", Description: "Delete a tag"},
	{Name: "list-for-task", Method: "GET", Path: "/tasks/{task_gid}/tags", Description: "Get tags for task"},
	{Name: "list-for-workspace", Method: "GET", Path: "/workspaces/{workspace_gid}/tags", Description: "Get tags in workspace"},
	{Name: "create-for-workspace", Method: "POST", Path: "/workspaces/{workspace_gid}/tags", Description: "Create a tag in workspace"},
	{Name: "list-tasks", Method: "GET", Path: "/tags/{tag_gid}/tasks", Description: "Get tasks from a tag"},
}

var SectionEndpoints = []Endpoint{
	{Name: "get", Method: "GET", Path: "/sections/{section_gid}", Description: "Get a section"},
	{Name: "update", Method: "PUT", Path: "/sections/{section_gid}", Description: "Update a section"},
	{Name: "delete", Method: "DELETE", Path: "/sections/{section_gid}", Description: "Delete a section"},
	{Name: "list-for-project", Method: "GET", Path: "/projects/{project_gid}/sections", Description: "Get sections in a project"},
	{Name: "create-for-project", Method: "POST", Path: "/projects/{project_gid}/sections", Description: "Create a section in a project"},
	{Name: "insert-for-project", Method: "POST", Path: "/projects/{project_gid}/sections/insert", Description: "Move or insert a section in project"},
	{Name: "add-task", Method: "POST", Path: "/sections/{section_gid}/addTask", Description: "Add task to section"},
	{Name: "list-tasks", Method: "GET", Path: "/sections/{section_gid}/tasks", Description: "Get tasks from a section"},
}

var CustomFieldEndpoints = []Endpoint{
	{Name: "list-for-workspace", Method: "GET", Path: "/workspaces/{workspace_gid}/custom_fields", Description: "Get custom fields in workspace"},
	{Name: "create", Method: "POST", Path: "/custom_fields", Description: "Create a custom field"},
	{Name: "get", Method: "GET", Path: "/custom_fields/{custom_field_gid}", Description: "Get a custom field"},
	{Name: "update", Method: "PUT", Path: "/custom_fields/{custom_field_gid}", Description: "Update a custom field"},
	{Name: "delete", Method: "DELETE", Path: "/custom_fields/{custom_field_gid}", Description: "Delete a custom field"},
	{Name: "create-enum-option", Method: "POST", Path: "/custom_fields/{custom_field_gid}/enum_options", Description: "Create enum option for a custom field"},
	{Name: "insert-enum-option", Method: "POST", Path: "/custom_fields/{custom_field_gid}/enum_options/insert", Description: "Move or insert enum option"},
	{Name: "update-enum-option", Method: "PUT", Path: "/enum_options/{enum_option_gid}", Description: "Update enum option"},
}

var UserTaskListEndpoints = []Endpoint{
	{Name: "get", Method: "GET", Path: "/user_task_lists/{user_task_list_gid}", Description: "Get a user task list"},
	{Name: "get-for-user", Method: "GET", Path: "/users/{user_gid}/user_task_list", Description: "Get user task list by user"},
}

var TimeTrackingEntryEndpoints = []Endpoint{
	{Name: "list", Method: "GET", Path: "/time_tracking_entries", Description: "Get multiple time tracking entries"},
	{Name: "get", Method: "GET", Path: "/time_tracking_entries/{time_tracking_entry_gid}", Description: "Get a time tracking entry"},
	{Name: "update", Method: "PUT", Path: "/time_tracking_entries/{time_tracking_entry_gid}", Description: "Update a time tracking entry"},
	{Name: "delete", Method: "DELETE", Path: "/time_tracking_entries/{time_tracking_entry_gid}", Description: "Delete a time tracking entry"},
	{Name: "list-for-task", Method: "GET", Path: "/tasks/{task_gid}/time_tracking_entries", Description: "Get time tracking entries for task"},
	{Name: "create-for-task", Method: "POST", Path: "/tasks/{task_gid}/time_tracking_entries", Description: "Create time tracking entry for task"},
}

var CompatSupportEndpoints = []Endpoint{
	{Name: "list-workspaces", Method: "GET", Path: "/workspaces", Description: "Get multiple workspaces"},
	{Name: "get-me", Method: "GET", Path: "/users/me", Description: "Get current user"},
	{Name: "get-task-stories", Method: "GET", Path: "/tasks/{task_gid}/stories", Description: "Get stories from task"},
	{Name: "create-task-story", Method: "POST", Path: "/tasks/{task_gid}/stories", Description: "Create story on task"},
	{Name: "list-task-attachments", Method: "GET", Path: "/tasks/{task_gid}/attachments", Description: "Get task attachments"},
	{Name: "get-attachment", Method: "GET", Path: "/attachments/{attachment_gid}", Description: "Get attachment"},
}
