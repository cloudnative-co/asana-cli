package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudnative-co/asana-cli/internal/asanaapi"
	"github.com/cloudnative-co/asana-cli/internal/auth"
	"github.com/cloudnative-co/asana-cli/internal/errs"
	"github.com/cloudnative-co/asana-cli/internal/profile"
)

type GlobalOptions struct {
	Profile        string
	Output         string
	OutputPath     string
	NoColor        bool
	NonInteractive bool
	DryRun         bool
	Yes            bool
}

type Runtime struct {
	Options GlobalOptions

	Profiles *profile.Manager
	Auth     *auth.Service
}

func NewRuntime(opts GlobalOptions) (*Runtime, error) {
	profileManager, err := profile.NewManager()
	if err != nil {
		return nil, err
	}
	return &Runtime{
		Options:  opts,
		Profiles: profileManager,
		Auth:     auth.NewService(profileManager),
	}, nil
}

func (rt *Runtime) ActiveProfileName() (string, error) {
	name, err := rt.Profiles.ResolveProfileName(rt.Options.Profile)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(name) == "" {
		return "", errs.New("profile_not_found", "active profile could not be resolved", "set --profile or run `asana profile set-default`")
	}
	return name, nil
}

func (rt *Runtime) EnsureProfile(name string) (profile.ProfileConfig, error) {
	profileCfg, ok, err := rt.Profiles.GetProfile(name)
	if err != nil {
		return profile.ProfileConfig{}, err
	}
	if ok {
		return profileCfg, nil
	}
	created := profile.ProfileConfig{Output: "table", OAuth: profile.OAuthConfig{RedirectURI: "urn:ietf:wg:oauth:2.0:oob"}}
	if upsertErr := rt.Profiles.UpsertProfile(name, created); upsertErr != nil {
		return profile.ProfileConfig{}, upsertErr
	}
	return created, nil
}

func (rt *Runtime) GetProfile(name string) (profile.ProfileConfig, bool, error) {
	return rt.Profiles.GetProfile(name)
}

func (rt *Runtime) EffectiveOutput(profileName string) (string, error) {
	if strings.TrimSpace(rt.Options.Output) != "" {
		return strings.ToLower(strings.TrimSpace(rt.Options.Output)), nil
	}
	profileCfg, ok, err := rt.Profiles.GetProfile(profileName)
	if err != nil {
		return "", err
	}
	if ok && strings.TrimSpace(profileCfg.Output) != "" {
		return strings.ToLower(profileCfg.Output), nil
	}
	return "table", nil
}

func (rt *Runtime) NewClient(ctx context.Context) (*asanaapi.Client, string, error) {
	profileName, err := rt.ActiveProfileName()
	if err != nil {
		return nil, "", err
	}
	if _, ensureErr := rt.EnsureProfile(profileName); ensureErr != nil {
		return nil, "", ensureErr
	}
	token, tokenSource, tokenErr := rt.Auth.ResolveBearerToken(ctx, profileName)
	if tokenErr != nil {
		return nil, "", tokenErr
	}
	if strings.TrimSpace(token) == "" {
		return nil, "", errs.New("missing_secret", fmt.Sprintf("no token for profile %q", profileName), "run `asana auth login` or `asana auth import-pat`")
	}
	_ = tokenSource
	return asanaapi.NewClient(token), profileName, nil
}
