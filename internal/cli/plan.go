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

func newPlanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "Show the execution plan (what would change)",
		RunE:  runPlan,
	}
}

func runPlan(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if errs := config.Validate(cfg); len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "Validation errors:")
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}
		os.Exit(1)
	}

	if credentialsPath == "" {
		return fmt.Errorf("--credentials flag is required for plan")
	}

	ctx := context.Background()
	client, err := gtm.NewClient(ctx, cfg.AccountID, credentialsPath)
	if err != nil {
		return fmt.Errorf("creating GTM client: %w", err)
	}

	actual, err := client.FetchState(ctx)
	if err != nil {
		return fmt.Errorf("fetching current state: %w", err)
	}

	desired := state.FromConfig(cfg)
	plan := diff.Compute(desired, actual, cfg.Mode)

	format := output.Format(formatFlag)
	return output.PrintPlan(os.Stdout, plan, format)
}
