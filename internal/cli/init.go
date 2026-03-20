package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/h13/gtm-users/internal/output"
	"github.com/spf13/cobra"
)

func newInitCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap a config file from current GTM state",
		RunE: func(cmd *cobra.Command, _ []string) error {
			accountID, _ := cmd.Flags().GetString("account-id")
			return runInit(opts, accountID)
		},
	}

	cmd.Flags().String("account-id", "", "GTM account ID (required)")
	_ = cmd.MarkFlagRequired("account-id")

	return cmd
}

func runInit(opts *rootOptions, accountID string) error {
	outPath := filepath.Clean(opts.configPath)

	if _, err := os.Stat(outPath); err == nil {
		return fmt.Errorf("config file already exists: %s", outPath)
	}

	if opts.credentialsPath == "" {
		return fmt.Errorf("--credentials flag is required for init")
	}

	ctx := context.Background()
	client, err := opts.newClient(ctx, accountID, opts.credentialsPath)
	if err != nil {
		return fmt.Errorf("creating GTM client: %w", err)
	}

	st, err := client.FetchState(ctx)
	if err != nil {
		return fmt.Errorf("fetching GTM state: %w", err)
	}

	users := make([]output.ExportUser, 0, len(st.Users))
	for _, u := range st.Users {
		containers := make([]output.ExportContainerAccess, 0, len(u.ContainerAccess))
		for _, ca := range u.ContainerAccess {
			containers = append(containers, output.ExportContainerAccess{
				ContainerID: ca.ContainerID,
				Permission:  ca.Permission,
			})
		}
		users = append(users, output.ExportUser{
			Email:           u.Email,
			AccountAccess:   u.AccountAccess,
			ContainerAccess: containers,
		})
	}

	f, err := os.Create(outPath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer f.Close()

	return output.PrintExport(f, accountID, users)
}
