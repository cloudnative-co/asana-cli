package asanaapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cloudnative-co/asana-cli/internal/errs"
)

const DefaultBaseURL = "https://app.asana.com/api/1.0"

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
	MaxRetries int
	UserAgent  string
}

func NewClient(token string) *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		MaxRetries: 4,
		UserAgent:  "asana-cli/1.0",
	}
}

func (c *Client) Request(
	ctx context.Context,
	method,
	path string,
	query map[string]string,
	body map[string]any,
	autoPaginate bool,
) (map[string]any, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errs.New("invalid_argument", "path is required", "")
	}

	if strings.ToUpper(method) == http.MethodGet && autoPaginate {
		return c.requestPaginated(ctx, method, path, query)
	}
	return c.requestSingle(ctx, method, path, query, body)
}

func (c *Client) requestPaginated(ctx context.Context, method, path string, query map[string]string) (map[string]any, error) {
	accumulated := []any{}
	currentQuery := cloneQuery(query)
	for {
		resp, err := c.requestSingle(ctx, method, path, currentQuery, nil)
		if err != nil {
			return nil, err
		}
		data, ok := resp["data"].([]any)
		if !ok {
			return resp, nil
		}
		accumulated = append(accumulated, data...)

		next, _ := resp["next_page"].(map[string]any)
		if next == nil {
			resp["data"] = accumulated
			resp["schema_version"] = "v1"
			return resp, nil
		}
		offsetValue, _ := next["offset"].(string)
		if strings.TrimSpace(offsetValue) == "" {
			resp["data"] = accumulated
			resp["next_page"] = nil
			resp["schema_version"] = "v1"
			return resp, nil
		}
		currentQuery["offset"] = offsetValue
	}
}

func (c *Client) requestSingle(
	ctx context.Context,
	method,
	path string,
	query map[string]string,
	body map[string]any,
) (map[string]any, error) {
	reqURL, err := c.buildURL(path, query)
	if err != nil {
		return nil, err
	}

	requestBody, err := encodeRequestBody(body)
	if err != nil {
		return nil, err
	}

	attempt := 0
	for {
		attempt++
		respBody, respStatus, headers, reqErr := c.do(ctx, method, reqURL, requestBody)
		if reqErr != nil {
			if attempt <= c.MaxRetries {
				time.Sleep(backoffForAttempt(attempt))
				continue
			}
			return nil, reqErr
		}

		requestID := headers.Get("X-Request-Id")
		if respStatus == http.StatusTooManyRequests {
			retryAfter := parseRetryAfter(headers.Get("Retry-After"))
			if attempt <= c.MaxRetries {
				wait := time.Duration(retryAfter) * time.Second
				if wait <= 0 {
					wait = backoffForAttempt(attempt)
				}
				time.Sleep(wait)
				continue
			}
			return nil, &errs.MachineError{
				Code:       "rate_limited",
				Message:    "asana API rate limited request",
				Status:     respStatus,
				RetryAfter: retryAfter,
				RequestID:  requestID,
			}
		}

		if respStatus >= 500 && attempt <= c.MaxRetries {
			time.Sleep(backoffForAttempt(attempt))
			continue
		}

		parsed := map[string]any{}
		if len(respBody) > 0 {
			if jsonErr := json.Unmarshal(respBody, &parsed); jsonErr != nil {
				return nil, errs.Wrap("api_error", "failed to decode asana response", string(respBody), jsonErr)
			}
		}
		if respStatus >= 400 {
			message := fmt.Sprintf("asana API error status=%d", respStatus)
			hint := extractErrorMessage(parsed)
			code := "api_error"
			if respStatus == 404 {
				code = "api_not_found"
			}
			return nil, &errs.MachineError{
				Code:      code,
				Message:   message,
				Hint:      hint,
				Status:    respStatus,
				RequestID: requestID,
			}
		}
		parsed["schema_version"] = "v1"
		if requestID != "" {
			parsed["request_id"] = requestID
		}
		return parsed, nil
	}
}

func (c *Client) do(ctx context.Context, method, reqURL string, body []byte) ([]byte, int, http.Header, error) {
	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, reader)
	if err != nil {
		return nil, 0, nil, errs.Wrap("internal_error", "failed to create request", "", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, nil, errs.Wrap("api_error", "request failed", "check network connectivity", err)
	}
	defer resp.Body.Close()
	payload, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, 0, nil, errs.Wrap("api_error", "failed to read response body", "", readErr)
	}
	return payload, resp.StatusCode, resp.Header, nil
}

func (c *Client) buildURL(path string, query map[string]string) (string, error) {
	u, err := url.Parse(strings.TrimRight(c.BaseURL, "/") + path)
	if err != nil {
		return "", errs.Wrap("invalid_argument", "invalid endpoint path", path, err)
	}
	values := u.Query()
	for key, value := range query {
		if strings.TrimSpace(key) == "" {
			continue
		}
		values.Set(key, value)
	}
	u.RawQuery = values.Encode()
	return u.String(), nil
}

func encodeRequestBody(body map[string]any) ([]byte, error) {
	if len(body) == 0 {
		return nil, nil
	}
	payload := body
	if _, ok := body["data"]; !ok {
		payload = map[string]any{"data": body}
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, errs.Wrap("invalid_argument", "failed to encode request body", "", err)
	}
	return encoded, nil
}

func parseRetryAfter(raw string) int {
	if strings.TrimSpace(raw) == "" {
		return 0
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return v
}

func backoffForAttempt(attempt int) time.Duration {
	scaled := math.Pow(2, float64(attempt-1))
	if scaled > 16 {
		scaled = 16
	}
	return time.Duration(scaled) * time.Second
}

func extractErrorMessage(payload map[string]any) string {
	errorsValue, ok := payload["errors"].([]any)
	if !ok || len(errorsValue) == 0 {
		return ""
	}
	first, _ := errorsValue[0].(map[string]any)
	if first == nil {
		return ""
	}
	if message, _ := first["message"].(string); message != "" {
		return message
	}
	return ""
}

func cloneQuery(query map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range query {
		out[key] = value
	}
	return out
}
