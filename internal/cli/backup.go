package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/h13/gtm-users/internal/output"
	"github.com/spf13/cobra"
)

func newBackupCmd(opts *rootOptions) *cobra.Command {
	var outputPath string
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Save current GTM state as a YAML backup",
		RunE: func(cmd *cobra.Command, _ []string) error {
			accountID, _ := cmd.Flags().GetString("account-id")
			return runBackup(opts, accountID, outputPath)
		},
	}

	cmd.Flags().String("account-id", "", "GTM account ID (required)")
	_ = cmd.MarkFlagRequired("account-id")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "output file path (default: backup-{timestamp}.yaml)")

	return cmd
}

func runBackup(opts *rootOptions, accountID, outputPath string) error {
	if outputPath == "" {
		outputPath = fmt.Sprintf("backup-%s.yaml", time.Now().Format("20060102-150405"))
	}

	outPath := filepath.Clean(outputPath)

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

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating backup file: %w", err)
	}

	if err := output.PrintExport(f, accountID, users); err != nil {
		_ = f.Close()
		return err
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("closing backup file: %w", err)
	}

	_, err = fmt.Fprintf(opts.stdout, "Backup saved to %s\n", outPath)
	return err
}
