package cli

import (
	"github.com/spf13/cobra"
)

var (
	configPath      string
	credentialsPath string
	formatFlag      string
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gtm-users",
		Short: "Declarative Google Tag Manager user permission management",
		Long:  "Manage GTM user permissions declaratively via YAML config files.",
	}

	cmd.PersistentFlags().StringVar(&configPath, "config", "gtm-users.yaml", "path to config file")
	cmd.PersistentFlags().StringVar(&credentialsPath, "credentials", "", "path to GCP service account credentials JSON")
	cmd.PersistentFlags().StringVar(&formatFlag, "format", "text", "output format (text|json)")

	cmd.AddCommand(
		newValidateCmd(),
		newExportCmd(),
		newPlanCmd(),
		newApplyCmd(),
	)

	return cmd
}
