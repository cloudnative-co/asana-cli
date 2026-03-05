package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cloudnative-co/asana-cli/internal/errs"
)

const (
	AuthorizeEndpoint = "https://app.asana.com/-/oauth_authorize"
	TokenEndpoint     = "https://app.asana.com/-/oauth_token"
	RevokeEndpoint    = "https://app.asana.com/-/oauth_revoke"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func BuildAuthorizeURL(clientID, redirectURI, state, codeChallenge string, scopes []string) (string, error) {
	query := url.Values{}
	query.Set("client_id", clientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("response_type", "code")
	query.Set("state", state)
	query.Set("code_challenge_method", "S256")
	query.Set("code_challenge", codeChallenge)
	if len(scopes) > 0 {
		query.Set("scope", strings.Join(scopes, " "))
	}
	u, err := url.Parse(AuthorizeEndpoint)
	if err != nil {
		return "", errs.Wrap("internal_error", "failed to build authorize URL", "", err)
	}
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func ExchangeAuthorizationCode(ctx context.Context, httpClient *http.Client, clientID, clientSecret, redirectURI, code, codeVerifier string) (TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("redirect_uri", redirectURI)
	form.Set("code", code)
	form.Set("code_verifier", codeVerifier)
	return tokenRequest(ctx, httpClient, form)
}

func RefreshAccessToken(ctx context.Context, httpClient *http.Client, clientID, clientSecret, refreshToken string) (TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("refresh_token", refreshToken)
	return tokenRequest(ctx, httpClient, form)
}

func RevokeToken(ctx context.Context, httpClient *http.Client, clientID, clientSecret, token string) error {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("token", token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, RevokeEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return errs.Wrap("internal_error", "failed to create revoke request", "", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return errs.Wrap("auth_failed", "failed to call revoke endpoint", "", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return &errs.MachineError{
			Code:    "auth_failed",
			Message: "failed to revoke token",
			Hint:    string(body),
			Status:  resp.StatusCode,
		}
	}
	return nil
}

func tokenRequest(ctx context.Context, httpClient *http.Client, form url.Values) (TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return TokenResponse{}, errs.Wrap("internal_error", "failed to create token request", "", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return TokenResponse{}, errs.Wrap("auth_failed", "failed to call token endpoint", "", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TokenResponse{}, errs.Wrap("auth_failed", "failed to read token response", "", err)
	}
	if resp.StatusCode >= 300 {
		return TokenResponse{}, &errs.MachineError{
			Code:    "auth_failed",
			Message: "token endpoint returned non-success status",
			Hint:    string(body),
			Status:  resp.StatusCode,
		}
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return TokenResponse{}, errs.Wrap("auth_failed", "failed to parse token response", string(body), err)
	}

	respToken := TokenResponse{}
	if access, _ := payload["access_token"].(string); access != "" {
		respToken.AccessToken = access
	}
	if tokenType, _ := payload["token_type"].(string); tokenType != "" {
		respToken.TokenType = tokenType
	}
	if refresh, _ := payload["refresh_token"].(string); refresh != "" {
		respToken.RefreshToken = refresh
	}
	switch v := payload["expires_in"].(type) {
	case float64:
		respToken.ExpiresIn = int(v)
	case string:
		if parsed, parseErr := strconv.Atoi(v); parseErr == nil {
			respToken.ExpiresIn = parsed
		}
	}

	if respToken.AccessToken == "" {
		return TokenResponse{}, errs.New("auth_failed", "token response missing access_token", string(body))
	}
	if respToken.ExpiresIn == 0 {
		respToken.ExpiresIn = 3600
	}
	return respToken, nil
}
