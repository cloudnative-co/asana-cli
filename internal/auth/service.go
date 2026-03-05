package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloudnative-co/asana-cli/internal/errs"
	"github.com/cloudnative-co/asana-cli/internal/profile"
)

type Service struct {
	profiles   *profile.Manager
	secrets    *SecretStore
	httpClient *http.Client
}

func NewService(profileManager *profile.Manager) *Service {
	return &Service{
		profiles: profileManager,
		secrets:  NewSecretStore(),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (s *Service) ImportPAT(profileName, pat string) error {
	rec, _, err := s.profiles.GetCredential(profileName)
	if err != nil {
		return err
	}
	patRef, setErr := s.secrets.Set(profileName, KindPAT, pat)
	if setErr != nil {
		return setErr
	}
	rec.PATRef = patRef
	return s.profiles.UpsertCredential(profileName, rec)
}

func (s *Service) LoginWithCode(
	ctx context.Context,
	profileName,
	clientID,
	clientSecret,
	redirectURI string,
	scopes []string,
	code,
	codeVerifier string,
) (TokenResponse, error) {
	if strings.TrimSpace(clientID) == "" {
		return TokenResponse{}, errs.New("invalid_argument", "client_id is required", "")
	}
	if strings.TrimSpace(clientSecret) == "" {
		return TokenResponse{}, errs.New("invalid_argument", "client_secret is required", "")
	}
	if strings.TrimSpace(redirectURI) == "" {
		redirectURI = "urn:ietf:wg:oauth:2.0:oob"
	}
	if strings.TrimSpace(code) == "" {
		return TokenResponse{}, errs.New("invalid_argument", "oauth code is required", "")
	}
	if strings.TrimSpace(codeVerifier) == "" {
		return TokenResponse{}, errs.New("invalid_argument", "code_verifier is required", "")
	}

	token, err := ExchangeAuthorizationCode(ctx, s.httpClient, clientID, clientSecret, redirectURI, code, codeVerifier)
	if err != nil {
		return TokenResponse{}, err
	}

	secretRef, setErr := s.secrets.Set(profileName, KindClientSecret, clientSecret)
	if setErr != nil {
		return TokenResponse{}, setErr
	}
	accessRef, setAccessErr := s.secrets.Set(profileName, KindAccessToken, token.AccessToken)
	if setAccessErr != nil {
		return TokenResponse{}, setAccessErr
	}
	refreshRef := ""
	if token.RefreshToken != "" {
		ref, setRefreshErr := s.secrets.Set(profileName, KindRefreshToken, token.RefreshToken)
		if setRefreshErr != nil {
			return TokenResponse{}, setRefreshErr
		}
		refreshRef = ref
	}

	rec, _, credErr := s.profiles.GetCredential(profileName)
	if credErr != nil {
		return TokenResponse{}, credErr
	}
	rec.AccessTokenRef = accessRef
	rec.RefreshTokenRef = refreshRef
	rec.ClientSecretRef = secretRef
	rec = rec.WithAccessExpiry(time.Now().UTC().Add(time.Duration(token.ExpiresIn) * time.Second))
	if upsertCredErr := s.profiles.UpsertCredential(profileName, rec); upsertCredErr != nil {
		return TokenResponse{}, upsertCredErr
	}

	profileCfg, ok, cfgErr := s.profiles.GetProfile(profileName)
	if cfgErr != nil {
		return TokenResponse{}, cfgErr
	}
	if !ok {
		profileCfg = profile.ProfileConfig{}
	}
	profileCfg.OAuth.ClientID = clientID
	profileCfg.OAuth.RedirectURI = redirectURI
	if len(scopes) > 0 {
		profileCfg.OAuth.Scopes = scopes
	}
	if upsertProfileErr := s.profiles.UpsertProfile(profileName, profileCfg); upsertProfileErr != nil {
		return TokenResponse{}, upsertProfileErr
	}

	return token, nil
}

func (s *Service) Logout(ctx context.Context, profileName string) error {
	rec, _, err := s.profiles.GetCredential(profileName)
	if err != nil {
		return err
	}

	profileCfg, _, cfgErr := s.profiles.GetProfile(profileName)
	if cfgErr != nil {
		return cfgErr
	}
	clientSecret, _, secretErr := s.secrets.Get(rec.ClientSecretRef, KindClientSecret)
	if secretErr != nil {
		return secretErr
	}
	accessToken, _, accessErr := s.secrets.Get(rec.AccessTokenRef, KindAccessToken)
	if accessErr != nil {
		return accessErr
	}
	if accessToken != "" && profileCfg.OAuth.ClientID != "" && clientSecret != "" {
		_ = RevokeToken(ctx, s.httpClient, profileCfg.OAuth.ClientID, clientSecret, accessToken)
	}

	_ = s.secrets.Delete(rec.PATRef)
	_ = s.secrets.Delete(rec.AccessTokenRef)
	_ = s.secrets.Delete(rec.RefreshTokenRef)
	_ = s.secrets.Delete(rec.ClientSecretRef)

	return s.profiles.UpsertCredential(profileName, profile.CredentialRecord{})
}

func (s *Service) ResolveBearerToken(ctx context.Context, profileName string) (string, string, error) {
	rec, _, err := s.profiles.GetCredential(profileName)
	if err != nil {
		return "", "", err
	}

	if pat, ok, getErr := s.secrets.Get(rec.PATRef, KindPAT); getErr != nil {
		return "", "", getErr
	} else if ok && pat != "" {
		return pat, "pat", nil
	}

	accessToken, accessFound, accessErr := s.secrets.Get(rec.AccessTokenRef, KindAccessToken)
	if accessErr != nil {
		return "", "", accessErr
	}
	if accessFound && accessToken != "" {
		if exp, ok := rec.AccessExpiryTime(); ok && exp.After(time.Now().UTC().Add(60*time.Second)) {
			return accessToken, "oauth", nil
		}
	}

	refreshToken, refreshFound, refreshErr := s.secrets.Get(rec.RefreshTokenRef, KindRefreshToken)
	if refreshErr != nil {
		return "", "", refreshErr
	}
	if !refreshFound || refreshToken == "" {
		if accessToken != "" {
			return accessToken, "oauth", nil
		}
		return "", "", errs.New("missing_secret", fmt.Sprintf("no PAT or OAuth token available for profile %q", profileName), "run `asana auth login` or `asana auth import-pat`")
	}

	profileCfg, ok, cfgErr := s.profiles.GetProfile(profileName)
	if cfgErr != nil {
		return "", "", cfgErr
	}
	if !ok {
		return "", "", errs.New("profile_not_found", fmt.Sprintf("profile %q not found", profileName), "")
	}
	if profileCfg.OAuth.ClientID == "" {
		return "", "", errs.New("token_refresh_failed", "oauth client_id is missing", "run `asana auth login` again")
	}
	clientSecret, foundSecret, secretErr := s.secrets.Get(rec.ClientSecretRef, KindClientSecret)
	if secretErr != nil {
		return "", "", secretErr
	}
	if !foundSecret || clientSecret == "" {
		return "", "", errs.New("token_refresh_failed", "oauth client_secret is missing", "set ASANA_CLIENT_SECRET or run `asana auth login`")
	}

	refreshed, refreshTokenErr := RefreshAccessToken(ctx, s.httpClient, profileCfg.OAuth.ClientID, clientSecret, refreshToken)
	if refreshTokenErr != nil {
		return "", "", errs.Wrap("token_refresh_failed", "failed to refresh access token", "run `asana auth login` again if this persists", refreshTokenErr)
	}

	newAccessRef, setAccessErr := s.secrets.Set(profileName, KindAccessToken, refreshed.AccessToken)
	if setAccessErr != nil {
		return "", "", setAccessErr
	}
	rec.AccessTokenRef = newAccessRef
	rec = rec.WithAccessExpiry(time.Now().UTC().Add(time.Duration(refreshed.ExpiresIn) * time.Second))
	if refreshed.RefreshToken != "" {
		newRefreshRef, setRefreshErr := s.secrets.Set(profileName, KindRefreshToken, refreshed.RefreshToken)
		if setRefreshErr != nil {
			return "", "", setRefreshErr
		}
		rec.RefreshTokenRef = newRefreshRef
	}
	if upsertErr := s.profiles.UpsertCredential(profileName, rec); upsertErr != nil {
		return "", "", upsertErr
	}
	return refreshed.AccessToken, "oauth", nil
}
