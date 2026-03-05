package profile

import "time"

// AppConfig stores non-secret settings.
type AppConfig struct {
	DefaultProfile string                   `toml:"default_profile"`
	Profiles       map[string]ProfileConfig `toml:"profiles"`
}

// ProfileConfig is a named profile configuration.
type ProfileConfig struct {
	Workspace      string      `toml:"workspace"`
	Output         string      `toml:"output"`
	NoColor        bool        `toml:"no_color"`
	NonInteractive bool        `toml:"non_interactive"`
	OAuth          OAuthConfig `toml:"oauth"`
}

// OAuthConfig stores non-secret OAuth settings.
type OAuthConfig struct {
	ClientID    string   `toml:"client_id"`
	RedirectURI string   `toml:"redirect_uri"`
	Scopes      []string `toml:"scopes"`
}

// Credentials stores references to external secret storage.
type Credentials struct {
	Profiles map[string]CredentialRecord `toml:"profiles"`
}

// CredentialRecord includes only references and metadata.
type CredentialRecord struct {
	PATRef            string `toml:"pat_ref"`
	AccessTokenRef    string `toml:"access_token_ref"`
	RefreshTokenRef   string `toml:"refresh_token_ref"`
	ClientSecretRef   string `toml:"client_secret_ref"`
	AccessTokenExpiry string `toml:"access_token_expiry"`
}

func (r CredentialRecord) AccessExpiryTime() (time.Time, bool) {
	if r.AccessTokenExpiry == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, r.AccessTokenExpiry)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func (r CredentialRecord) WithAccessExpiry(t time.Time) CredentialRecord {
	r.AccessTokenExpiry = t.UTC().Format(time.RFC3339)
	return r
}
