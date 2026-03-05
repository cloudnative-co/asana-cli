package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cloudnative-co/asana-cli/internal/auth"
	"github.com/cloudnative-co/asana-cli/internal/errs"
	"github.com/cloudnative-co/asana-cli/internal/output"
)

func NewAuthCommand(provider RuntimeProvider) *cobra.Command {
	authCommand := &cobra.Command{
		Use:   "auth",
		Short: "Authentication management (OAuth + PAT)",
	}

	authCommand.AddCommand(newAuthLoginCommand(provider))
	authCommand.AddCommand(newAuthImportPATCommand(provider))
	authCommand.AddCommand(newAuthLogoutCommand(provider))

	return authCommand
}

func newAuthLoginCommand(provider RuntimeProvider) *cobra.Command {
	var profileName string
	var clientID string
	var clientSecret string
	var redirectURI string
	var scopes []string
	var scopePreset string
	var code string
	var codeVerifier string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate using OAuth authorization code + PKCE",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}

			resolvedProfile := strings.TrimSpace(profileName)
			if resolvedProfile == "" {
				resolvedProfile, err = rt.ActiveProfileName()
				if err != nil {
					resolvedProfile = "default"
				}
			}

			profileCfg, ensureErr := rt.EnsureProfile(resolvedProfile)
			if ensureErr != nil {
				return ensureErr
			}

			if strings.TrimSpace(clientID) == "" {
				clientID = strings.TrimSpace(profileCfg.OAuth.ClientID)
			}
			if strings.TrimSpace(redirectURI) == "" {
				redirectURI = strings.TrimSpace(profileCfg.OAuth.RedirectURI)
				if redirectURI == "" {
					redirectURI = strings.TrimSpace(os.Getenv("ASANA_REDIRECT_URI"))
				}
				if redirectURI == "" {
					redirectURI = "urn:ietf:wg:oauth:2.0:oob"
				}
			}
			resolvedScopes := []string{}
			if strings.TrimSpace(scopePreset) != "" {
				presetScopes, presetErr := auth.ResolveScopePreset(scopePreset)
				if presetErr != nil {
					return presetErr
				}
				resolvedScopes = append(resolvedScopes, presetScopes...)
			}
			resolvedScopes = append(resolvedScopes, scopes...)
			resolvedScopes = auth.NormalizeScopes(resolvedScopes)
			if len(resolvedScopes) == 0 {
				resolvedScopes = auth.NormalizeScopes(profileCfg.OAuth.Scopes)
			}
			scopes = resolvedScopes

			if strings.TrimSpace(clientID) == "" {
				return errs.New("invalid_argument", "client_id is required", "set --client-id or configure profile oauth.client_id")
			}

			if strings.TrimSpace(clientSecret) == "" {
				clientSecret = strings.TrimSpace(os.Getenv("ASANA_CLIENT_SECRET"))
			}
			if strings.TrimSpace(clientSecret) == "" {
				if rt.Options.NonInteractive {
					return errs.New("invalid_argument", "client_secret is required in non-interactive mode", "set --client-secret or ASANA_CLIENT_SECRET")
				}
				input, inputErr := prompt("Client Secret: ")
				if inputErr != nil {
					return inputErr
				}
				clientSecret = input
			}

			if strings.TrimSpace(codeVerifier) == "" {
				generatedVerifier, verifierErr := auth.NewCodeVerifier()
				if verifierErr != nil {
					return errs.Wrap("internal_error", "failed to generate PKCE verifier", "", verifierErr)
				}
				codeVerifier = generatedVerifier
			}
			state, stateErr := auth.NewState()
			if stateErr != nil {
				return errs.Wrap("internal_error", "failed to generate oauth state", "", stateErr)
			}
			challenge := auth.NewCodeChallenge(codeVerifier)
			authURL, authErr := auth.BuildAuthorizeURL(clientID, redirectURI, state, challenge, scopes)
			if authErr != nil {
				return authErr
			}

			if strings.TrimSpace(code) == "" {
				fmt.Fprintf(os.Stderr, "OAuth redirect_uri: %s\n", redirectURI)
				fmt.Fprintln(os.Stderr, "This value must exactly match one of your app's OAuth Redirect URLs in Asana Developer Console.")
				if len(scopes) == 0 {
					fmt.Fprintln(os.Stderr, "OAuth scopes: (not specified; Asana app defaults will be used)")
				} else {
					fmt.Fprintf(os.Stderr, "OAuth scopes: %s\n", strings.Join(scopes, " "))
				}
				fmt.Fprintf(os.Stderr, "Open the following URL and authorize the app:\n%s\n\n", authURL)
				if rt.Options.NonInteractive {
					return errs.New("invalid_argument", "oauth authorization code is required in non-interactive mode", "pass --code <code>")
				}
				input, inputErr := prompt("Paste authorization code: ")
				if inputErr != nil {
					return inputErr
				}
				code = input
			}

			token, loginErr := rt.Auth.LoginWithCode(context.Background(), resolvedProfile, clientID, clientSecret, redirectURI, scopes, code, codeVerifier)
			if loginErr != nil {
				return loginErr
			}

			payload := map[string]any{
				"schema_version": "v1",
				"profile":        resolvedProfile,
				"token_type":     token.TokenType,
				"expires_in":     token.ExpiresIn,
				"expires_at":     time.Now().UTC().Add(time.Duration(token.ExpiresIn) * time.Second).Format(time.RFC3339),
			}
			format, formatErr := rt.EffectiveOutput(resolvedProfile)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(payload, format, rt.Options.OutputPath)
		},
	}

	cmd.Flags().StringVar(&profileName, "profile", "", "profile name override")
	cmd.Flags().StringVar(&clientID, "client-id", "", "asana oauth client id")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "asana oauth client secret")
	cmd.Flags().StringVar(&redirectURI, "redirect-uri", "", "oauth redirect uri (must exactly match app OAuth redirect URL; default: urn:ietf:wg:oauth:2.0:oob)")
	cmd.Flags().StringSliceVar(&scopes, "scopes", nil, "oauth scopes (comma-separated or repeatable, e.g. --scopes tasks:read,users:read)")
	cmd.Flags().StringVar(&scopePreset, "scope-preset", "", "oauth scope preset (supported: task-full)")
	cmd.Flags().StringVar(&code, "code", "", "oauth authorization code")
	cmd.Flags().StringVar(&codeVerifier, "code-verifier", "", "oauth code verifier (advanced)")

	return cmd
}

func newAuthImportPATCommand(provider RuntimeProvider) *cobra.Command {
	var profileName string
	var pat string
	cmd := &cobra.Command{
		Use:   "import-pat",
		Short: "Store PAT in keyring and bind to profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}

			resolvedProfile := strings.TrimSpace(profileName)
			if resolvedProfile == "" {
				resolvedProfile, err = rt.ActiveProfileName()
				if err != nil {
					resolvedProfile = "default"
				}
			}
			if _, ensureErr := rt.EnsureProfile(resolvedProfile); ensureErr != nil {
				return ensureErr
			}

			value := strings.TrimSpace(pat)
			if value == "" {
				value = strings.TrimSpace(os.Getenv("ASANA_PAT"))
			}
			if value == "" {
				if rt.Options.NonInteractive {
					return errs.New("invalid_argument", "pat is required in non-interactive mode", "pass --pat or ASANA_PAT")
				}
				input, inputErr := prompt("Personal Access Token: ")
				if inputErr != nil {
					return inputErr
				}
				value = input
			}
			if importErr := rt.Auth.ImportPAT(resolvedProfile, value); importErr != nil {
				return importErr
			}

			format, formatErr := rt.EffectiveOutput(resolvedProfile)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(map[string]any{"schema_version": "v1", "profile": resolvedProfile, "status": "pat_imported"}, format, rt.Options.OutputPath)
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "profile name override")
	cmd.Flags().StringVar(&pat, "pat", "", "personal access token")
	return cmd
}

func newAuthLogoutCommand(provider RuntimeProvider) *cobra.Command {
	var profileName string
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove PAT/OAuth secrets from active profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			resolvedProfile := strings.TrimSpace(profileName)
			if resolvedProfile == "" {
				resolvedProfile, err = rt.ActiveProfileName()
				if err != nil {
					return err
				}
			}
			if logoutErr := rt.Auth.Logout(context.Background(), resolvedProfile); logoutErr != nil {
				return logoutErr
			}
			format, formatErr := rt.EffectiveOutput(resolvedProfile)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(map[string]any{"schema_version": "v1", "profile": resolvedProfile, "status": "logged_out"}, format, rt.Options.OutputPath)
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "profile name override")
	return cmd
}

func prompt(label string) (string, error) {
	fmt.Fprint(os.Stderr, label)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", errs.Wrap("internal_error", "failed to read user input", "", err)
	}
	return strings.TrimSpace(line), nil
}
