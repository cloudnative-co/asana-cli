package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudnative-co/asana-cli/internal/errs"
)

type TaskIndex struct {
	GeneratedAt time.Time        `json:"generated_at"`
	Entries     []TaskIndexEntry `json:"entries"`
}

type TaskIndexEntry struct {
	Index int    `json:"index"`
	GID   string `json:"gid"`
	Name  string `json:"name"`
	DueOn string `json:"due_on"`
}

func pathFor(profileName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errs.Wrap("internal_error", "failed to resolve home directory", "", err)
	}
	dir := filepath.Join(home, ".cache", "asana-cli", "task-index")
	if mkErr := os.MkdirAll(dir, 0o700); mkErr != nil {
		return "", errs.Wrap("internal_error", "failed to create cache dir", dir, mkErr)
	}
	name := strings.TrimSpace(profileName)
	if name == "" {
		name = "default"
	}
	return filepath.Join(dir, name+".json"), nil
}

func SaveTaskIndex(profileName string, tasks []map[string]any) error {
	entries := make([]TaskIndexEntry, 0, len(tasks))
	for i, task := range tasks {
		entry := TaskIndexEntry{Index: i}
		if gid, _ := task["gid"].(string); gid != "" {
			entry.GID = gid
		}
		if name, _ := task["name"].(string); name != "" {
			entry.Name = name
		}
		if dueOn, _ := task["due_on"].(string); dueOn != "" {
			entry.DueOn = dueOn
		}
		if entry.GID != "" {
			entries = append(entries, entry)
		}
	}
	index := TaskIndex{GeneratedAt: time.Now().UTC(), Entries: entries}
	b, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return errs.Wrap("internal_error", "failed to encode cache", "", err)
	}
	path, pathErr := pathFor(profileName)
	if pathErr != nil {
		return pathErr
	}
	if writeErr := os.WriteFile(path, b, 0o600); writeErr != nil {
		return errs.Wrap("internal_error", "failed to write cache", path, writeErr)
	}
	return nil
}

func LoadTaskIndex(profileName string) (TaskIndex, bool, error) {
	path, pathErr := pathFor(profileName)
	if pathErr != nil {
		return TaskIndex{}, false, pathErr
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return TaskIndex{}, false, nil
		}
		return TaskIndex{}, false, errs.Wrap("internal_error", "failed to read cache", path, err)
	}
	var index TaskIndex
	if unmarshalErr := json.Unmarshal(b, &index); unmarshalErr != nil {
		return TaskIndex{}, false, errs.Wrap("internal_error", "failed to parse cache", path, unmarshalErr)
	}
	return index, true, nil
}

func ResolveTaskRef(profileName, ref string, autoFirst bool) (string, error) {
	trimmed := strings.TrimSpace(ref)
	if trimmed == "" {
		if !autoFirst {
			return "", errs.New("invalid_argument", "task index or gid is required", "")
		}
		trimmed = "0"
	}
	if _, parseErr := strconv.Atoi(trimmed); parseErr != nil {
		return trimmed, nil
	}
	index, ok, err := LoadTaskIndex(profileName)
	if err != nil {
		return "", err
	}
	if !ok || len(index.Entries) == 0 {
		return "", errs.New("invalid_argument", "task index cache is empty", "run `asana tasks` first or pass a task gid")
	}
	position, _ := strconv.Atoi(trimmed)
	if position < 0 || position >= len(index.Entries) {
		return "", errs.New("invalid_argument", fmt.Sprintf("task index out of range: %d", position), "run `asana tasks` to refresh index cache")
	}
	return index.Entries[position].GID, nil
}

func IsOlder(profileName string, duration time.Duration) (bool, error) {
	path, pathErr := pathFor(profileName)
	if pathErr != nil {
		return true, pathErr
	}
	stat, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		return true, errs.Wrap("internal_error", "failed to stat cache", path, err)
	}
	return time.Now().After(stat.ModTime().Add(duration)), nil
}
