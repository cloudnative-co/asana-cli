package profile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/cloudnative-co/asana-cli/internal/errs"
)

const (
	configFileName      = "config.toml"
	credentialsFileName = "credentials.toml"
)

type Manager struct {
	configPath      string
	credentialsPath string
}

func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, errs.Wrap("internal_error", "failed to resolve home directory", "", err)
	}
	base := filepath.Join(home, ".config", "asana-cli")
	if mkErr := os.MkdirAll(base, 0o700); mkErr != nil {
		return nil, errs.Wrap("internal_error", "failed to create config directory", "", mkErr)
	}
	return &Manager{
		configPath:      filepath.Join(base, configFileName),
		credentialsPath: filepath.Join(base, credentialsFileName),
	}, nil
}

func (m *Manager) ConfigPath() string {
	return m.configPath
}

func (m *Manager) CredentialsPath() string {
	return m.credentialsPath
}

func (m *Manager) LoadConfig() (AppConfig, error) {
	cfg := AppConfig{Profiles: map[string]ProfileConfig{}}
	b, err := os.ReadFile(m.configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return AppConfig{}, errs.Wrap("invalid_config", "failed to read config file", m.configPath, err)
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		return cfg, nil
	}
	if unmarshalErr := toml.Unmarshal(b, &cfg); unmarshalErr != nil {
		return AppConfig{}, errs.Wrap("invalid_config", "failed to parse config file", m.configPath, unmarshalErr)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]ProfileConfig{}
	}
	return cfg, nil
}

func (m *Manager) SaveConfig(cfg AppConfig) error {
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]ProfileConfig{}
	}
	b, err := toml.Marshal(cfg)
	if err != nil {
		return errs.Wrap("internal_error", "failed to encode config", "", err)
	}
	if writeErr := os.WriteFile(m.configPath, b, 0o600); writeErr != nil {
		return errs.Wrap("internal_error", "failed to write config", m.configPath, writeErr)
	}
	return nil
}

func (m *Manager) LoadCredentials() (Credentials, error) {
	creds := Credentials{Profiles: map[string]CredentialRecord{}}
	b, err := os.ReadFile(m.credentialsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return creds, nil
		}
		return Credentials{}, errs.Wrap("invalid_config", "failed to read credentials file", m.credentialsPath, err)
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		return creds, nil
	}
	if unmarshalErr := toml.Unmarshal(b, &creds); unmarshalErr != nil {
		return Credentials{}, errs.Wrap("invalid_config", "failed to parse credentials file", m.credentialsPath, unmarshalErr)
	}
	if creds.Profiles == nil {
		creds.Profiles = map[string]CredentialRecord{}
	}
	return creds, nil
}

func (m *Manager) SaveCredentials(creds Credentials) error {
	if creds.Profiles == nil {
		creds.Profiles = map[string]CredentialRecord{}
	}
	b, err := toml.Marshal(creds)
	if err != nil {
		return errs.Wrap("internal_error", "failed to encode credentials", "", err)
	}
	if writeErr := os.WriteFile(m.credentialsPath, b, 0o600); writeErr != nil {
		return errs.Wrap("internal_error", "failed to write credentials", m.credentialsPath, writeErr)
	}
	return nil
}

func (m *Manager) ResolveProfileName(flagProfile string) (string, error) {
	if strings.TrimSpace(flagProfile) != "" {
		return strings.TrimSpace(flagProfile), nil
	}
	if envProfile := strings.TrimSpace(os.Getenv("ASANA_PROFILE")); envProfile != "" {
		return envProfile, nil
	}
	cfg, err := m.LoadConfig()
	if err != nil {
		return "", err
	}
	if cfg.DefaultProfile != "" {
		return cfg.DefaultProfile, nil
	}
	if len(cfg.Profiles) == 1 {
		for name := range cfg.Profiles {
			return name, nil
		}
	}
	return "default", nil
}

func (m *Manager) UpsertProfile(name string, profileCfg ProfileConfig) error {
	cfg, err := m.LoadConfig()
	if err != nil {
		return err
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]ProfileConfig{}
	}
	cfg.Profiles[name] = profileCfg
	if cfg.DefaultProfile == "" {
		cfg.DefaultProfile = name
	}
	return m.SaveConfig(cfg)
}

func (m *Manager) GetProfile(name string) (ProfileConfig, bool, error) {
	cfg, err := m.LoadConfig()
	if err != nil {
		return ProfileConfig{}, false, err
	}
	profileCfg, ok := cfg.Profiles[name]
	return profileCfg, ok, nil
}

func (m *Manager) SetDefaultProfile(name string) error {
	cfg, err := m.LoadConfig()
	if err != nil {
		return err
	}
	if _, ok := cfg.Profiles[name]; !ok {
		return errs.New("profile_not_found", fmt.Sprintf("profile %q not found", name), "use `asana profile list` to inspect profiles")
	}
	cfg.DefaultProfile = name
	return m.SaveConfig(cfg)
}

func (m *Manager) RemoveProfile(name string) error {
	cfg, err := m.LoadConfig()
	if err != nil {
		return err
	}
	if _, ok := cfg.Profiles[name]; !ok {
		return errs.New("profile_not_found", fmt.Sprintf("profile %q not found", name), "")
	}
	delete(cfg.Profiles, name)
	if cfg.DefaultProfile == name {
		cfg.DefaultProfile = ""
	}
	if saveErr := m.SaveConfig(cfg); saveErr != nil {
		return saveErr
	}

	creds, credErr := m.LoadCredentials()
	if credErr != nil {
		return credErr
	}
	delete(creds.Profiles, name)
	return m.SaveCredentials(creds)
}

func (m *Manager) GetCredential(name string) (CredentialRecord, bool, error) {
	creds, err := m.LoadCredentials()
	if err != nil {
		return CredentialRecord{}, false, err
	}
	rec, ok := creds.Profiles[name]
	return rec, ok, nil
}

func (m *Manager) UpsertCredential(name string, rec CredentialRecord) error {
	creds, err := m.LoadCredentials()
	if err != nil {
		return err
	}
	if creds.Profiles == nil {
		creds.Profiles = map[string]CredentialRecord{}
	}
	creds.Profiles[name] = rec
	return m.SaveCredentials(creds)
}
