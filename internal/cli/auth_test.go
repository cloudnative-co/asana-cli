package cli

import (
	"strings"
	"testing"

	"github.com/cloudnative-co/asana-cli/internal/auth"
)

func TestAuthLoginSuccessPayloadIncludesNextWorkspaceStep(t *testing.T) {
	t.Parallel()

	payload := authLoginSuccessPayload("default", auth.TokenResponse{
		TokenType: "bearer",
		ExpiresIn: 3600,
	})

	if got := payload["next_step"]; got != "Configure your default workspace before running task and project commands." {
		t.Fatalf("unexpected next_step: %#v", got)
	}
	if got := payload["next_command"]; got != "asana config --profile default --workspace <workspace_gid>" {
		t.Fatalf("unexpected next_command: %#v", got)
	}
	if got := payload["token_type"]; got != "bearer" {
		t.Fatalf("unexpected token_type: %#v", got)
	}
	expiresAt, _ := payload["expires_at"].(string)
	if !strings.Contains(expiresAt, "T") {
		t.Fatalf("expected RFC3339-like expires_at, got %#v", payload["expires_at"])
	}
}
