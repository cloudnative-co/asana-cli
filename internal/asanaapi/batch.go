package asanaapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/cloudnative-co/asana-cli/internal/errs"
)

const maxBatchActions = 10

type BatchAction struct {
	RelativePath string
	Method       string
	Data         map[string]any
	Options      map[string]any
}

type BatchResult struct {
	StatusCode int
	Headers    map[string]any
	Body       map[string]any
}

func (c *Client) BatchRequest(ctx context.Context, actions []BatchAction) ([]BatchResult, error) {
	if len(actions) == 0 {
		return nil, nil
	}
	if len(actions) > maxBatchActions {
		return nil, errs.New("invalid_argument", "too many batch actions", "batch requests support up to 10 actions")
	}

	payloadActions := make([]any, 0, len(actions))
	for _, action := range actions {
		if strings.TrimSpace(action.RelativePath) == "" {
			return nil, errs.New("invalid_argument", "batch relative_path is required", "")
		}
		method := strings.ToLower(strings.TrimSpace(action.Method))
		if method == "" {
			method = "get"
		}
		item := map[string]any{
			"relative_path": action.RelativePath,
			"method":        method,
		}
		if len(action.Data) > 0 {
			item["data"] = action.Data
		}
		if len(action.Options) > 0 {
			item["options"] = action.Options
		}
		payloadActions = append(payloadActions, item)
	}

	resp, err := c.requestSingle(ctx, http.MethodPost, "/batch", map[string]string{
		"opt_fields": "body,headers,status_code",
	}, map[string]any{
		"actions": payloadActions,
	})
	if err != nil {
		return nil, err
	}

	rawData, ok := resp["data"].([]any)
	if !ok {
		return nil, errs.New("internal_error", "unexpected batch response shape", "")
	}
	results := make([]BatchResult, 0, len(rawData))
	for _, item := range rawData {
		row, _ := item.(map[string]any)
		if row == nil {
			return nil, errs.New("internal_error", "unexpected batch result entry", "")
		}
		results = append(results, BatchResult{
			StatusCode: int(numberValue(row["status_code"])),
			Headers:    mapValue(row["headers"]),
			Body:       mapValue(row["body"]),
		})
	}
	return results, nil
}

func (c *Client) BatchGetTasks(
	ctx context.Context,
	taskGIDs []string,
	fields []string,
) (map[string]map[string]any, map[string]error, error) {
	unique := uniqueStrings(taskGIDs)
	found := map[string]map[string]any{}
	failures := map[string]error{}

	for start := 0; start < len(unique); start += maxBatchActions {
		end := start + maxBatchActions
		if end > len(unique) {
			end = len(unique)
		}
		chunk := unique[start:end]
		actions := make([]BatchAction, 0, len(chunk))
		for _, gid := range chunk {
			actions = append(actions, BatchAction{
				RelativePath: "/tasks/" + gid,
				Method:       "get",
				Options: map[string]any{
					"fields": fields,
				},
			})
		}
		results, err := c.BatchRequest(ctx, actions)
		if err != nil {
			return nil, nil, err
		}
		for idx, result := range results {
			gid := chunk[idx]
			if result.StatusCode >= 400 {
				failures[gid] = &errs.MachineError{
					Code:    batchErrorCode(result.StatusCode),
					Message: fmt.Sprintf("asana batch action error status=%d", result.StatusCode),
					Hint:    extractErrorMessage(result.Body),
					Status:  result.StatusCode,
				}
				continue
			}
			data := mapValue(result.Body["data"])
			if data == nil {
				failures[gid] = errs.New("internal_error", "unexpected task batch response shape", "")
				continue
			}
			found[gid] = data
		}
	}

	return found, failures, nil
}

func batchErrorCode(status int) string {
	if status == http.StatusNotFound {
		return "api_not_found"
	}
	return "api_error"
}

func mapValue(v any) map[string]any {
	out, _ := v.(map[string]any)
	return out
}

func numberValue(v any) float64 {
	switch typed := v.(type) {
	case float64:
		return typed
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	default:
		return 0
	}
}

func uniqueStrings(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, item := range in {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
