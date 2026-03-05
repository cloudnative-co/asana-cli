package compat

import (
	"testing"
	"time"
)

func TestParseCompatDue(t *testing.T) {
	now := time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC)
	if got := ParseCompatDue("today", now); got != "2026-03-05" {
		t.Fatalf("today parse failed: %s", got)
	}
	if got := ParseCompatDue("tomorrow", now); got != "2026-03-06" {
		t.Fatalf("tomorrow parse failed: %s", got)
	}
	if got := ParseCompatDue("2026-12-01", now); got != "2026-12-01" {
		t.Fatalf("literal date parse failed: %s", got)
	}
}
