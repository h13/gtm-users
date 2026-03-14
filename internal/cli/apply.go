package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/h13/gtm-users/internal/config"
	"github.com/h13/gtm-users/internal/diff"
	"github.com/h13/gtm-users/internal/gtm"
	"github.com/h13/gtm-users/internal/output"
	"github.com/h13/gtm-users/internal/state"
	"github.com/spf13/cobra"
)

func newApplyCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply the planned changes to GTM",
		RunE: func(cmd *cobra.Command, args []string) error {
			autoApprove, _ := cmd.Flags().GetBool("auto-approve")
			return runApply(opts, autoApprove)
		},
	}

	cmd.Flags().Bool("auto-approve", false, "skip interactive approval prompt")

	return cmd
}

func runApply(opts *rootOptions, autoApprove bool) error {
	cfg, err := config.Load(opts.configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := checkValidation(cfg); err != nil {
		return err
	}

	if opts.credentialsPath == "" {
		return fmt.Errorf("--credentials flag is required for apply")
	}

	ctx := context.Background()
	client, err := gtm.NewClient(ctx, cfg.AccountID, opts.credentialsPath)
	if err != nil {
		return fmt.Errorf("creating GTM client: %w", err)
	}

	actual, err := client.FetchState(ctx)
	if err != nil {
		return fmt.Errorf("fetching current state: %w", err)
	}

	desired := state.FromConfig(cfg)
	plan := diff.Compute(desired, actual, cfg.Mode)

	if !plan.HasChanges() {
		fmt.Println("No changes. Infrastructure is up-to-date.")
		return nil
	}

	format := output.Format(opts.format)
	if err := output.PrintPlan(os.Stdout, plan, format); err != nil {
		return err
	}

	if !autoApprove && !confirmApply() {
		fmt.Println("Apply cancelled.")
		return nil
	}

	return executeChanges(ctx, client, plan, desired)
}

func confirmApply() bool {
	fmt.Print("Do you want to apply these changes? (yes/no): ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(answer)) == "yes"
}

func executeChanges(ctx context.Context, client gtmClient, plan diff.Plan, desired state.AccountState) error {
	var failures int
	for _, change := range plan.Changes {
		if err := applyChange(ctx, client, change, desired); err != nil {
			failures++
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	if failures > 0 {
		return fmt.Errorf("%d operation(s) failed", failures)
	}

	fmt.Println("\nApply complete!")
	return nil
}

func applyChange(ctx context.Context, client gtmClient, change diff.UserChange, desired state.AccountState) error {
	switch change.Action {
	case diff.ActionAdd:
		user := findDesiredUser(desired, change.Email)
		if err := client.CreateUserPermission(ctx, user); err != nil {
			return fmt.Errorf("adding %s: %w", change.Email, err)
		}
		fmt.Printf("Added %s\n", change.Email)

	case diff.ActionUpdate:
		user := findDesiredUser(desired, change.Email)
		if err := client.UpdateUserPermission(ctx, user); err != nil {
			return fmt.Errorf("updating %s: %w", change.Email, err)
		}
		fmt.Printf("Updated %s\n", change.Email)

	case diff.ActionDelete:
		if err := client.DeleteUserPermission(ctx, change.Email); err != nil {
			return fmt.Errorf("deleting %s: %w", change.Email, err)
		}
		fmt.Printf("Deleted %s\n", change.Email)
	}

	return nil
}

func findDesiredUser(s state.AccountState, email string) state.UserPermission {
	for _, u := range s.Users {
		if u.Email == email {
			return u
		}
	}
	return state.UserPermission{Email: email}
}
