package auth

import (
	"fmt"
	"os"
	"strings"

	"github.com/zalando/go-keyring"

	"github.com/cloudnative-co/asana-cli/internal/errs"
)

const keyringService = "asana-cli"

const (
	KindPAT          = "pat"
	KindAccessToken  = "access_token"
	KindRefreshToken = "refresh_token"
	KindClientSecret = "client_secret"
)

type SecretStore struct{}

func NewSecretStore() *SecretStore {
	return &SecretStore{}
}

func envForKind(kind string) string {
	switch kind {
	case KindPAT:
		return "ASANA_PAT"
	case KindAccessToken:
		return "ASANA_ACCESS_TOKEN"
	case KindRefreshToken:
		return "ASANA_REFRESH_TOKEN"
	case KindClientSecret:
		return "ASANA_CLIENT_SECRET"
	default:
		return ""
	}
}

func refFor(profileName, kind string) string {
	return fmt.Sprintf("profile:%s:%s", profileName, kind)
}

func (s *SecretStore) Set(profileName, kind, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", errs.New("invalid_argument", "secret value is empty", "")
	}
	ref := refFor(profileName, kind)
	if err := keyring.Set(keyringService, ref, trimmed); err != nil {
		return "", errs.Wrap(
			"missing_secret",
			"failed to store secret in keyring",
			"keyring is unavailable. set env vars (ASANA_PAT / ASANA_ACCESS_TOKEN / ASANA_REFRESH_TOKEN / ASANA_CLIENT_SECRET) as fallback",
			err,
		)
	}
	return ref, nil
}

func (s *SecretStore) Get(ref, kind string) (string, bool, error) {
	envName := envForKind(kind)
	if envName != "" {
		if envValue := strings.TrimSpace(os.Getenv(envName)); envValue != "" {
			return envValue, true, nil
		}
	}
	if strings.TrimSpace(ref) == "" {
		return "", false, nil
	}
	value, err := keyring.Get(keyringService, ref)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", false, nil
		}
		return "", false, errs.Wrap("missing_secret", "failed to read secret from keyring", "", err)
	}
	return value, true, nil
}

func (s *SecretStore) Delete(ref string) error {
	if strings.TrimSpace(ref) == "" {
		return nil
	}
	if err := keyring.Delete(keyringService, ref); err != nil && err != keyring.ErrNotFound {
		return errs.Wrap("internal_error", "failed to delete keyring secret", "", err)
	}
	return nil
}
