package cli

import (
	"context"
	"fmt"
	"sort"

	"github.com/h13/gtm-users/internal/output"
	"github.com/spf13/cobra"
)

func newMatrixCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "matrix",
		Short: "Display a user x container permission matrix",
		RunE: func(cmd *cobra.Command, _ []string) error {
			accountID, _ := cmd.Flags().GetString("account-id")
			return runMatrix(opts, accountID)
		},
	}

	cmd.Flags().String("account-id", "", "GTM account ID (required)")
	_ = cmd.MarkFlagRequired("account-id")

	return cmd
}

func runMatrix(opts *rootOptions, accountID string) error {
	ctx := context.Background()
	client, err := opts.newClient(ctx, accountID, opts.credentialsPath)
	if err != nil {
		return fmt.Errorf("creating GTM client: %w", err)
	}

	st, err := client.FetchState(ctx)
	if err != nil {
		return fmt.Errorf("fetching GTM state: %w", err)
	}

	// Collect all unique container IDs.
	containerSet := make(map[string]bool)
	for _, u := range st.Users {
		for _, ca := range u.ContainerAccess {
			containerSet[ca.ContainerID] = true
		}
	}

	containerIDs := make([]string, 0, len(containerSet))
	for cid := range containerSet {
		containerIDs = append(containerIDs, cid)
	}
	sort.Strings(containerIDs)

	// Build matrix entries.
	entries := make([]output.MatrixEntry, 0, len(st.Users))
	for _, u := range st.Users {
		perms := make(map[string]string, len(u.ContainerAccess))
		for _, ca := range u.ContainerAccess {
			perms[ca.ContainerID] = ca.Permission
		}
		entries = append(entries, output.MatrixEntry{
			Email:       u.Email,
			Permissions: perms,
		})
	}

	return output.PrintMatrix(opts.stdout, entries, containerIDs)
}
