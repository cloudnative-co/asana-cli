package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cloudnative-co/asana-cli/internal/app"
	"github.com/cloudnative-co/asana-cli/internal/errs"
)

type RuntimeProvider func() (*app.Runtime, error)

func NewRootCommand() *cobra.Command {
	opts := app.GlobalOptions{}
	provider := func() (*app.Runtime, error) {
		return app.NewRuntime(opts)
	}

	root := &cobra.Command{
		Use:   "asana",
		Short: "Asana CLI with profile-based auth and automation-friendly output",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Output == "" {
				opts.Output = ""
			}
			return nil
		},
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVar(&opts.Profile, "profile", "", "profile name (or ASANA_PROFILE)")
	root.PersistentFlags().StringVar(&opts.Output, "output", "", "output format: table|json|csv")
	root.PersistentFlags().StringVar(&opts.OutputPath, "out", "", "write output to file path")
	root.PersistentFlags().BoolVar(&opts.NoColor, "no-color", false, "disable color output")
	root.PersistentFlags().BoolVar(&opts.NonInteractive, "non-interactive", false, "disable prompts and require explicit inputs")
	root.PersistentFlags().BoolVar(&opts.DryRun, "dry-run", false, "show request without mutating action")
	root.PersistentFlags().BoolVar(&opts.Yes, "yes", false, "assume yes for destructive operations")

	root.AddCommand(NewAuthCommand(provider))
	root.AddCommand(NewProfileCommand(provider))
	root.AddCommand(NewTaskCommand(provider))
	root.AddCommand(NewProjectCommand(provider))
	root.AddCommand(NewUserCommand(provider))
	for _, compatCommand := range NewCompatCommands(provider) {
		root.AddCommand(compatCommand)
	}

	return root
}

func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		m := errs.AsMachine(err)
		fmt.Fprintln(os.Stderr, errs.JSON(m))
		os.Exit(errs.ExitCode(m))
	}
}
