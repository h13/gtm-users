package cli

import (
	"context"
	"fmt"

	"github.com/h13/gtm-users/internal/output"
	"github.com/spf13/cobra"
)

func newExportCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export current GTM user permissions as YAML",
		RunE: func(cmd *cobra.Command, _ []string) error {
			accountID, _ := cmd.Flags().GetString("account-id")
			return runExport(opts, accountID)
		},
	}

	cmd.Flags().String("account-id", "", "GTM account ID (required)")
	_ = cmd.MarkFlagRequired("account-id")

	return cmd
}

func runExport(opts *rootOptions, accountID string) error {
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

	return output.PrintExport(opts.stdout, accountID, users)
}
