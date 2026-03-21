package cli

import (
	"context"
	"fmt"

	"github.com/h13/gtm-users/internal/config"
	"github.com/h13/gtm-users/internal/diff"
	"github.com/h13/gtm-users/internal/output"
	"github.com/h13/gtm-users/internal/state"
	"github.com/spf13/cobra"
)

// DriftDetectedError signals that drift was found (exit code 2).
type DriftDetectedError struct{}

func (DriftDetectedError) Error() string {
	return "drift detected"
}

func newDriftCmd(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "drift",
		Short: "Detect configuration drift (exit code 2 if drift found)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDrift(opts)
		},
	}
}

func runDrift(opts *rootOptions) error {
	cfg, err := config.Load(opts.configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := checkValidation(cfg); err != nil {
		return err
	}

	ctx := context.Background()
	client, err := opts.newClient(ctx, cfg.AccountID, opts.credentialsPath)
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
		_, err := fmt.Fprintln(opts.stdout, "No drift detected.")
		return err
	}

	format := output.Format(opts.format)
	if err := output.PrintPlan(opts.stdout, plan, format, !opts.noColor); err != nil {
		return err
	}

	return DriftDetectedError{}
}
