package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cloudnative-co/asana-cli/internal/output"
)

func NewProfileCommand(provider RuntimeProvider) *cobra.Command {
	profileCommand := &cobra.Command{
		Use:   "profile",
		Short: "Profile operations",
	}
	profileCommand.AddCommand(newProfileListCommand(provider))
	profileCommand.AddCommand(newProfileShowCommand(provider))
	profileCommand.AddCommand(newProfileUseCommand(provider))
	profileCommand.AddCommand(newProfileSetDefaultCommand(provider))
	profileCommand.AddCommand(newProfileRemoveCommand(provider))
	profileCommand.AddCommand(newProfileValidateCommand(provider))
	return profileCommand
}

func newProfileListCommand(provider RuntimeProvider) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			cfg, loadErr := rt.Profiles.LoadConfig()
			if loadErr != nil {
				return loadErr
			}
			rows := []any{}
			for name, profileCfg := range cfg.Profiles {
				rows = append(rows, map[string]any{
					"name":            name,
					"default":         name == cfg.DefaultProfile,
					"workspace":       profileCfg.Workspace,
					"output":          profileCfg.Output,
					"non_interactive": profileCfg.NonInteractive,
				})
			}
			profileName, _ := rt.ActiveProfileName()
			format, formatErr := rt.EffectiveOutput(profileName)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(map[string]any{"schema_version": "v1", "data": rows}, format, rt.Options.OutputPath)
		},
	}
}

func newProfileShowCommand(provider RuntimeProvider) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show profile details",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			if strings.TrimSpace(name) == "" {
				name, err = rt.ActiveProfileName()
				if err != nil {
					return err
				}
			}
			profileCfg, ok, getErr := rt.GetProfile(name)
			if getErr != nil {
				return getErr
			}
			if !ok {
				return fmt.Errorf("profile %q not found", name)
			}
			cred, _, credErr := rt.Profiles.GetCredential(name)
			if credErr != nil {
				return credErr
			}
			payload := map[string]any{
				"schema_version": "v1",
				"name":           name,
				"profile":        profileCfg,
				"credentials":    cred,
			}
			format, formatErr := rt.EffectiveOutput(name)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(payload, format, rt.Options.OutputPath)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "profile name")
	return cmd
}

func newProfileUseCommand(provider RuntimeProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <name>",
		Short: "Set default profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			name := args[0]
			if _, ensureErr := rt.EnsureProfile(name); ensureErr != nil {
				return ensureErr
			}
			if setErr := rt.Profiles.SetDefaultProfile(name); setErr != nil {
				return setErr
			}
			format, formatErr := rt.EffectiveOutput(name)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(map[string]any{"schema_version": "v1", "default_profile": name}, format, rt.Options.OutputPath)
		},
	}
	return cmd
}

func newProfileSetDefaultCommand(provider RuntimeProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-default <name>",
		Short: "Set default profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			name := args[0]
			if _, ensureErr := rt.EnsureProfile(name); ensureErr != nil {
				return ensureErr
			}
			if setErr := rt.Profiles.SetDefaultProfile(name); setErr != nil {
				return setErr
			}
			format, formatErr := rt.EffectiveOutput(name)
			if formatErr != nil {
				return formatErr
			}
			return output.Render(map[string]any{"schema_version": "v1", "default_profile": name}, format, rt.Options.OutputPath)
		},
	}
	return cmd
}

func newProfileRemoveCommand(provider RuntimeProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			if removeErr := rt.Profiles.RemoveProfile(args[0]); removeErr != nil {
				return removeErr
			}
			format, formatErr := rt.EffectiveOutput(args[0])
			if formatErr != nil {
				format = "json"
			}
			return output.Render(map[string]any{"schema_version": "v1", "removed": args[0]}, format, rt.Options.OutputPath)
		},
	}
	return cmd
}

func newProfileValidateCommand(provider RuntimeProvider) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate profile by calling /users/me",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := provider()
			if err != nil {
				return err
			}
			if strings.TrimSpace(name) != "" {
				rt.Options.Profile = name
			}
			client, profileName, clientErr := rt.NewClient(context.Background())
			if clientErr != nil {
				return clientErr
			}
			resp, reqErr := client.Request(context.Background(), "GET", "/users/me", nil, nil, false)
			if reqErr != nil {
				return reqErr
			}
			format, formatErr := rt.EffectiveOutput(profileName)
			if formatErr != nil {
				return formatErr
			}
			payload := map[string]any{
				"schema_version": "v1",
				"profile":        profileName,
				"status":         "ok",
				"user":           resp["data"],
			}
			return output.Render(payload, format, rt.Options.OutputPath)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "profile name override")
	return cmd
}
