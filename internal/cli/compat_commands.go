package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cloudnative-co/asana-cli/internal/app"
	"github.com/cloudnative-co/asana-cli/internal/cache"
	"github.com/cloudnative-co/asana-cli/internal/compat"
	"github.com/cloudnative-co/asana-cli/internal/errs"
	"github.com/cloudnative-co/asana-cli/internal/output"
)

func NewCompatCommands(provider RuntimeProvider) []*cobra.Command {
	return []*cobra.Command{
		newCompatConfigCommand(provider),
		newCompatWorkspacesCommand(provider),
		newCompatTasksCommand(provider),
		newCompatCommentCommand(provider),
		newCompatDoneCommand(provider),
		newCompatDueCommand(provider),
		newCompatBrowseCommand(provider),
		newCompatDownloadCommand(provider),
	}
}

func newCompatConfigCommand(provider RuntimeProvider) *cobra.Command {
	var profileName string
	var pat string
	var workspace string
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Compatibility config command (legacy style)",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			if strings.TrimSpace(profileName) != "" {
				rt.Options.Profile = profileName
			}
			active, activeErr := rt.ActiveProfileName()
			if activeErr != nil {
				active = "default"
			}
			profileCfg, ensureErr := rt.EnsureProfile(active)
			if ensureErr != nil {
				return ensureErr
			}
			if strings.TrimSpace(pat) != "" {
				if importErr := rt.Auth.ImportPAT(active, pat); importErr != nil {
					return importErr
				}
			}

			if strings.TrimSpace(workspace) == "" {
				workspace = profileCfg.Workspace
			}

			if workspace == "" {
				client, _, clientErr := rt.NewClient(context.Background())
				if clientErr != nil {
					return errs.Wrap("missing_secret", "workspace is not set and user lookup failed", "run with --workspace after importing PAT/OAuth", clientErr)
				}
				resp, reqErr := client.Request(context.Background(), http.MethodGet, "/users/me", map[string]string{"opt_fields": "workspaces.gid,workspaces.name"}, nil, false)
				if reqErr != nil {
					return reqErr
				}
				data, _ := resp["data"].(map[string]any)
				workspacesRaw, _ := data["workspaces"].([]any)
				if len(workspacesRaw) == 1 {
					if only, ok := workspacesRaw[0].(map[string]any); ok {
						workspace, _ = only["gid"].(string)
					}
				} else if len(workspacesRaw) > 1 {
					if rt.Options.NonInteractive {
						return errs.New("invalid_argument", "multiple workspaces found", "use --workspace <gid>")
					}
					fmt.Fprintln(os.Stderr, "Workspaces:")
					for idx, item := range workspacesRaw {
						ws, _ := item.(map[string]any)
						fmt.Fprintf(os.Stderr, "[%d] %v %v\n", idx, ws["gid"], ws["name"])
					}
					choice, choiceErr := prompt("Choose workspace index: ")
					if choiceErr != nil {
						return choiceErr
					}
					selected, parseErr := strconv.Atoi(choice)
					if parseErr != nil || selected < 0 || selected >= len(workspacesRaw) {
						return errs.New("invalid_argument", "invalid workspace index", "")
					}
					ws, _ := workspacesRaw[selected].(map[string]any)
					workspace, _ = ws["gid"].(string)
				}
			}

			profileCfg.Workspace = strings.TrimSpace(workspace)
			if profileCfg.Output == "" {
				profileCfg.Output = "table"
			}
			if upsertErr := rt.Profiles.UpsertProfile(active, profileCfg); upsertErr != nil {
				return upsertErr
			}
			format, formatErr := rt.EffectiveOutput(active)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(map[string]any{"schema_version": "v1", "profile": active, "workspace": profileCfg.Workspace}, format, rt.Options.OutputPath)
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "profile name")
	cmd.Flags().StringVar(&pat, "pat", "", "personal access token")
	cmd.Flags().StringVar(&workspace, "workspace", "", "workspace gid")
	return cmd
}

func newCompatWorkspacesCommand(provider RuntimeProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workspaces",
		Aliases: []string{"w"},
		Short:   "Compatibility command: list workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			client, profileName, clientErr := rt.NewClient(context.Background())
			if clientErr != nil {
				return clientErr
			}
			resp, reqErr := client.Request(context.Background(), http.MethodGet, "/workspaces", nil, nil, true)
			if reqErr != nil {
				return reqErr
			}
			format, formatErr := rt.EffectiveOutput(profileName)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(resp, format, rt.Options.OutputPath)
		},
	}
	return cmd
}

func newCompatTasksCommand(provider RuntimeProvider) *cobra.Command {
	var noCache bool
	var refresh bool
	cmd := &cobra.Command{
		Use:     "tasks",
		Aliases: []string{"ts"},
		Short:   "Compatibility command: list tasks assigned to me",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			client, profileName, clientErr := rt.NewClient(context.Background())
			if clientErr != nil {
				return clientErr
			}

			shouldFetch := noCache || refresh
			if !shouldFetch {
				older, olderErr := cache.IsOlder(profileName, 5*time.Minute)
				if olderErr != nil {
					return olderErr
				}
				shouldFetch = older
			}

			if !shouldFetch {
				index, found, loadErr := cache.LoadTaskIndex(profileName)
				if loadErr != nil {
					return loadErr
				}
				if found && len(index.Entries) > 0 {
					rows := make([]any, 0, len(index.Entries))
					for _, entry := range index.Entries {
						rows = append(rows, map[string]any{
							"index":  entry.Index,
							"gid":    entry.GID,
							"due_on": entry.DueOn,
							"name":   entry.Name,
						})
					}
					format, formatErr := rt.EffectiveOutput(profileName)
					if formatErr != nil {
						return formatErr
					}
					return output.Render(map[string]any{"schema_version": "v1", "data": rows, "cached": true}, format, rt.Options.OutputPath)
				}
			}

			query := map[string]string{"assignee": "me", "opt_fields": "gid,name,due_on,completed"}
			if profileCfg, ok, profileErr := rt.GetProfile(profileName); profileErr == nil && ok && profileCfg.Workspace != "" {
				query["workspace"] = profileCfg.Workspace
			}
			resp, reqErr := client.Request(context.Background(), http.MethodGet, "/tasks", query, nil, true)
			if reqErr != nil {
				return reqErr
			}
			rawData, _ := resp["data"].([]any)
			tasks := make([]map[string]any, 0, len(rawData))
			for _, item := range rawData {
				task, _ := item.(map[string]any)
				if task == nil {
					continue
				}
				if completed, ok := task["completed"].(bool); ok && completed {
					continue
				}
				tasks = append(tasks, task)
			}
			sorted := compat.SortTasksByDue(tasks)
			if saveErr := cache.SaveTaskIndex(profileName, sorted); saveErr != nil {
				return saveErr
			}
			resp["data"] = compat.BuildIndexedRows(sorted)
			format, formatErr := rt.EffectiveOutput(profileName)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(resp, format, rt.Options.OutputPath)
		},
	}
	cmd.Flags().BoolVarP(&noCache, "no-cache", "n", false, "skip cache")
	cmd.Flags().BoolVarP(&refresh, "refresh", "r", false, "refresh cache")
	return cmd
}

func runCompatTaskShow(rt *app.Runtime, ref string, verbose bool, forceJSON bool) error {
	client, profileName, err := rt.NewClient(context.Background())
	if err != nil {
		return err
	}
	taskGID, resolveErr := cache.ResolveTaskRef(profileName, ref, true)
	if resolveErr != nil {
		if strings.TrimSpace(ref) == "" {
			return resolveErr
		}
		taskGID = ref
	}

	query := map[string]string{
		"opt_fields": "gid,name,notes,due_on,completed,tags.name,custom_fields.name,custom_fields.display_value,projects.name,assignee.name",
	}
	taskResp, taskErr := client.Request(context.Background(), http.MethodGet, "/tasks/"+taskGID, query, nil, false)
	if taskErr != nil {
		return taskErr
	}
	attachmentsResp, attachErr := client.Request(context.Background(), http.MethodGet, "/tasks/"+taskGID+"/attachments", nil, nil, true)
	if attachErr == nil {
		taskResp["attachments"] = attachmentsResp["data"]
	}
	if verbose {
		storiesResp, storyErr := client.Request(context.Background(), http.MethodGet, "/tasks/"+taskGID+"/stories", nil, nil, true)
		if storyErr == nil {
			taskResp["stories"] = storiesResp["data"]
		}
	}
	format, formatErr := rt.EffectiveOutput(profileName)
	if formatErr != nil {
		return formatErr
	}
	if forceJSON {
		format = "json"
	}
	payload := map[string]any{
		"schema_version": "v1",
		"task":           taskResp["data"],
		"attachments":    taskResp["attachments"],
		"stories":        taskResp["stories"],
	}
	return output.Render(payload, format, rt.Options.OutputPath)
}

func newCompatCommentCommand(provider RuntimeProvider) *cobra.Command {
	var text string
	cmd := &cobra.Command{
		Use:     "comment <task_index_or_gid>",
		Aliases: []string{"cm"},
		Short:   "Compatibility command: post task comment",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			client, profileName, clientErr := rt.NewClient(context.Background())
			if clientErr != nil {
				return clientErr
			}
			taskGID, resolveErr := cache.ResolveTaskRef(profileName, args[0], false)
			if resolveErr != nil {
				taskGID = args[0]
			}
			commentText := strings.TrimSpace(text)
			if commentText == "" {
				if rt.Options.NonInteractive {
					return errs.New("invalid_argument", "comment text is required in non-interactive mode", "pass --text")
				}
				editor, hasEditor := os.LookupEnv("EDITOR")
				if hasEditor && strings.TrimSpace(editor) != "" {
					tmpFile, createErr := os.CreateTemp("", "asana-comment-*.txt")
					if createErr != nil {
						return errs.Wrap("internal_error", "failed to create temp file", "", createErr)
					}
					defer os.Remove(tmpFile.Name())
					defer tmpFile.Close()
					runErr := exec.Command(editor, tmpFile.Name()).Run()
					if runErr != nil {
						return errs.Wrap("internal_error", "failed to run editor", "", runErr)
					}
					body, readErr := os.ReadFile(tmpFile.Name())
					if readErr != nil {
						return errs.Wrap("internal_error", "failed to read editor output", "", readErr)
					}
					commentText = strings.TrimSpace(string(body))
				} else {
					fmt.Fprintln(os.Stderr, "Enter comment. End with Ctrl-D:")
					body, readErr := io.ReadAll(bufio.NewReader(os.Stdin))
					if readErr != nil {
						return errs.Wrap("internal_error", "failed to read stdin", "", readErr)
					}
					commentText = strings.TrimSpace(string(body))
				}
			}
			if commentText == "" {
				return errs.New("invalid_argument", "comment text is empty", "")
			}
			resp, reqErr := client.Request(context.Background(), http.MethodPost, "/tasks/"+taskGID+"/stories", nil, map[string]any{"text": commentText}, false)
			if reqErr != nil {
				return reqErr
			}
			format, formatErr := rt.EffectiveOutput(profileName)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(resp, format, rt.Options.OutputPath)
		},
	}
	cmd.Flags().StringVar(&text, "text", "", "comment text")
	return cmd
}

func newCompatDoneCommand(provider RuntimeProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done <task_index_or_gid>",
		Short: "Compatibility command: mark task completed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			client, profileName, clientErr := rt.NewClient(context.Background())
			if clientErr != nil {
				return clientErr
			}
			taskGID, resolveErr := cache.ResolveTaskRef(profileName, args[0], false)
			if resolveErr != nil {
				taskGID = args[0]
			}
			resp, reqErr := client.Request(context.Background(), http.MethodPut, "/tasks/"+taskGID, nil, map[string]any{"completed": true}, false)
			if reqErr != nil {
				return reqErr
			}
			format, formatErr := rt.EffectiveOutput(profileName)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(resp, format, rt.Options.OutputPath)
		},
	}
	return cmd
}

func newCompatDueCommand(provider RuntimeProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "due <task_index_or_gid> <due_date|today|tomorrow>",
		Short: "Compatibility command: set task due date",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			client, profileName, clientErr := rt.NewClient(context.Background())
			if clientErr != nil {
				return clientErr
			}
			taskGID, resolveErr := cache.ResolveTaskRef(profileName, args[0], false)
			if resolveErr != nil {
				taskGID = args[0]
			}
			dueOn := compat.ParseCompatDue(args[1], time.Now())
			resp, reqErr := client.Request(context.Background(), http.MethodPut, "/tasks/"+taskGID, nil, map[string]any{"due_on": dueOn}, false)
			if reqErr != nil {
				return reqErr
			}
			format, formatErr := rt.EffectiveOutput(profileName)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(resp, format, rt.Options.OutputPath)
		},
	}
	return cmd
}

func newCompatBrowseCommand(provider RuntimeProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "browse <task_index_or_gid>",
		Aliases: []string{"b"},
		Short:   "Compatibility command: open task in browser",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			_, profileName, clientErr := rt.NewClient(context.Background())
			if clientErr != nil {
				return clientErr
			}
			taskGID, resolveErr := cache.ResolveTaskRef(profileName, args[0], false)
			if resolveErr != nil {
				taskGID = args[0]
			}
			profileCfg, _, profileErr := rt.GetProfile(profileName)
			if profileErr != nil {
				return profileErr
			}
			if strings.TrimSpace(profileCfg.Workspace) == "" {
				return errs.New("invalid_argument", "workspace is not configured", "run `asana config --workspace <gid>`")
			}
			target := fmt.Sprintf("https://app.asana.com/0/%s/%s", profileCfg.Workspace, taskGID)
			launcher, launchArgs, launchErr := browserLauncher(target)
			if launchErr != nil {
				return launchErr
			}
			runErr := exec.Command(launcher, launchArgs...).Start()
			if runErr != nil {
				return errs.Wrap("internal_error", "failed to start browser", "", runErr)
			}
			format, formatErr := rt.EffectiveOutput(profileName)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(map[string]any{"schema_version": "v1", "opened": target}, format, rt.Options.OutputPath)
		},
	}
	return cmd
}

func newCompatDownloadCommand(provider RuntimeProvider) *cobra.Command {
	var outputPath string
	cmd := &cobra.Command{
		Use:     "download <task_index_or_gid> <attachment_index> | <attachment_gid>",
		Aliases: []string{"dl"},
		Short:   "Compatibility command: download attachment",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			client, profileName, clientErr := rt.NewClient(context.Background())
			if clientErr != nil {
				return clientErr
			}

			attachmentGID := ""
			if len(args) == 1 && len(args[0]) > 10 {
				attachmentGID = args[0]
			} else if len(args) == 2 {
				taskGID, resolveErr := cache.ResolveTaskRef(profileName, args[0], false)
				if resolveErr != nil {
					taskGID = args[0]
				}
				attachmentsResp, attachmentsErr := client.Request(context.Background(), http.MethodGet, "/tasks/"+taskGID+"/attachments", nil, nil, true)
				if attachmentsErr != nil {
					return attachmentsErr
				}
				attachments, _ := attachmentsResp["data"].([]any)
				idx, parseErr := strconv.Atoi(args[1])
				if parseErr != nil || idx < 0 || idx >= len(attachments) {
					return errs.New("invalid_argument", "invalid attachment index", "")
				}
				attachmentMap, _ := attachments[idx].(map[string]any)
				attachmentGID, _ = attachmentMap["gid"].(string)
			}
			if strings.TrimSpace(attachmentGID) == "" {
				return errs.New("invalid_argument", "attachment gid not resolved", "")
			}
			attachmentResp, attachmentErr := client.Request(context.Background(), http.MethodGet, "/attachments/"+attachmentGID, nil, nil, false)
			if attachmentErr != nil {
				return attachmentErr
			}
			attachment, _ := attachmentResp["data"].(map[string]any)
			downloadURL, _ := attachment["download_url"].(string)
			if strings.TrimSpace(downloadURL) == "" {
				return errs.New("api_error", "attachment does not expose download_url", "")
			}
			name, _ := attachment["name"].(string)
			if strings.TrimSpace(outputPath) == "" {
				outputPath = name
			}
			if strings.TrimSpace(outputPath) == "" {
				outputPath = attachmentGID
			}
			if mkErr := os.MkdirAll(filepath.Dir(outputPath), 0o755); mkErr != nil {
				return errs.Wrap("internal_error", "failed to create output directory", outputPath, mkErr)
			}
			if downloadErr := downloadFile(downloadURL, outputPath); downloadErr != nil {
				return downloadErr
			}
			format, formatErr := rt.EffectiveOutput(profileName)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(map[string]any{"schema_version": "v1", "attachment_gid": attachmentGID, "path": outputPath}, format, rt.Options.OutputPath)
		},
	}
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "output file path")
	return cmd
}

func browserLauncher(target string) (string, []string, error) {
	if browser := strings.TrimSpace(os.Getenv("BROWSER")); browser != "" {
		parts := strings.Fields(browser)
		return parts[0], append(parts[1:], target), nil
	}
	switch runtime.GOOS {
	case "darwin":
		return "open", []string{target}, nil
	case "windows":
		return "cmd", []string{"/c", "start", target}, nil
	default:
		candidates := []string{"xdg-open", "x-www-browser", "firefox", "chromium", "google-chrome"}
		for _, c := range candidates {
			if _, err := exec.LookPath(c); err == nil {
				return c, []string{target}, nil
			}
		}
	}
	return "", nil, errs.New("invalid_argument", "browser launcher not found", "set BROWSER environment variable")
}

func downloadFile(sourceURL, destinationPath string) error {
	resp, err := http.Get(sourceURL)
	if err != nil {
		return errs.Wrap("api_error", "failed to download attachment", sourceURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return errs.New("api_error", fmt.Sprintf("download request failed with status %d", resp.StatusCode), sourceURL)
	}
	out, err := os.Create(destinationPath)
	if err != nil {
		return errs.Wrap("internal_error", "failed to create output file", destinationPath, err)
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return errs.Wrap("internal_error", "failed to write downloaded file", destinationPath, err)
	}
	return nil
}
