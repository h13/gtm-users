package cli

import (
	"context"

	"github.com/h13/gtm-users/internal/gtm"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	configPath      string
	credentialsPath string
	format          string
	newClient       func(ctx context.Context, accountID, credentialsPath string) (gtmClient, error)
}

func NewRootCmd(version string) *cobra.Command {
	opts := &rootOptions{}
	opts.newClient = func(ctx context.Context, accountID, credentialsPath string) (gtmClient, error) {
		return gtm.NewClient(ctx, accountID, credentialsPath)
	}

	cmd := &cobra.Command{
		Use:           "gtm-users",
		Short:         "Declarative Google Tag Manager user permission management",
		Long:          "Manage GTM user permissions declaratively via YAML config files.",
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.PersistentFlags().StringVar(&opts.configPath, "config", "gtm-users.yaml", "path to config file")
	cmd.PersistentFlags().StringVar(&opts.credentialsPath, "credentials", "", "path to GCP service account credentials JSON")
	cmd.PersistentFlags().StringVar(&opts.format, "format", "text", "output format (text|json)")

	cmd.AddCommand(
		newValidateCmd(opts),
		newExportCmd(opts),
		newPlanCmd(opts),
		newApplyCmd(opts),
	)

	return cmd
}
