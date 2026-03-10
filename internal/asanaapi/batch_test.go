package asanaapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBatchRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/batch" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		data, _ := payload["data"].(map[string]any)
		actions, _ := data["actions"].([]any)
		if len(actions) != 1 {
			t.Fatalf("expected 1 batch action, got %d", len(actions))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"status_code":200,"body":{"data":{"gid":"1","name":"Task 1"}}}]}`))
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL

	results, err := client.BatchRequest(context.Background(), []BatchAction{{
		RelativePath: "/tasks/1",
		Method:       "get",
		Options:      map[string]any{"fields": []string{"gid", "name"}},
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", results[0].StatusCode)
	}
}

func TestBatchGetTasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data":[
				{"status_code":200,"body":{"data":{"gid":"1","name":"Task 1","parent":null,"projects":[],"memberships":[]}}},
				{"status_code":404,"body":{"errors":[{"message":"Not Found"}]}}
			]
		}`))
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL

	found, failures, err := client.BatchGetTasks(context.Background(), []string{"1", "2"}, []string{"gid", "name"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found["1"]["name"] != "Task 1" {
		t.Fatalf("expected task 1 to be returned, got %#v", found["1"])
	}
	if _, ok := failures["2"]; !ok {
		t.Fatalf("expected task 2 failure to be recorded")
	}
}
