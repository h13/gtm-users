package cli

import (
	"context"
	"io"
	"os"

	"github.com/h13/gtm-users/internal/gtm"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	configPath      string
	credentialsPath string
	format          string
	noColor         bool
	newClient       func(ctx context.Context, accountID, credentialsPath string) (gtmClient, error)
	stdout          io.Writer
	stdin           io.Reader
}

// CmdOption configures the root command.
type CmdOption func(*rootOptions)

// WithGTMOptions passes options to the GTM client constructor.
func WithGTMOptions(opts ...gtm.Option) CmdOption {
	return func(o *rootOptions) {
		o.newClient = func(ctx context.Context, accountID, credentialsPath string) (gtmClient, error) {
			return gtm.NewClient(ctx, accountID, credentialsPath, opts...)
		}
	}
}

func NewRootCmd(version string, opts ...CmdOption) *cobra.Command {
	o := &rootOptions{
		stdout: os.Stdout,
		stdin:  os.Stdin,
		newClient: func(ctx context.Context, accountID, credentialsPath string) (gtmClient, error) {
			return gtm.NewClient(ctx, accountID, credentialsPath)
		},
	}
	for _, opt := range opts {
		opt(o)
	}

	cmd := &cobra.Command{
		Use:           "gtm-users",
		Short:         "Declarative Google Tag Manager user permission management",
		Long:          "Manage GTM user permissions declaratively via YAML config files.",
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.PersistentFlags().StringVar(&o.configPath, "config", "gtm-users.yaml", "path to config file")
	cmd.PersistentFlags().StringVar(&o.credentialsPath, "credentials", "", "path to GCP credentials JSON (omit to use Application Default Credentials)")
	cmd.PersistentFlags().StringVar(&o.format, "format", "text", "output format (text|json)")
	cmd.PersistentFlags().BoolVar(&o.noColor, "no-color", false, "disable colored output")

	cmd.AddCommand(
		newValidateCmd(o),
		newExportCmd(o),
		newPlanCmd(o),
		newApplyCmd(o),
		newInitCmd(o),
		newDriftCmd(o),
		newCompletionCmd(),
		newMatrixCmd(o),
		newBackupCmd(o),
	)

	return cmd
}
