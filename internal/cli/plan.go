package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/h13/gtm-users/internal/config"
	"github.com/h13/gtm-users/internal/diff"
	"github.com/h13/gtm-users/internal/gtm"
	"github.com/h13/gtm-users/internal/output"
	"github.com/h13/gtm-users/internal/state"
	"github.com/spf13/cobra"
)

func newPlanCmd(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "Show the execution plan (what would change)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPlan(opts)
		},
	}
}

func runPlan(opts *rootOptions) error {
	cfg, err := config.Load(opts.configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := checkValidation(cfg); err != nil {
		return err
	}

	if opts.credentialsPath == "" {
		return fmt.Errorf("--credentials flag is required for plan")
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

	format := output.Format(opts.format)
	return output.PrintPlan(os.Stdout, plan, format)
}
