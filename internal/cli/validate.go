package cli

import (
	"fmt"
	"os"

	"github.com/h13/gtm-users/internal/config"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the config file syntax and values",
		RunE:  runValidate,
	}
}

func runValidate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	errs := config.Validate(cfg)
	if len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "Validation errors:")
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}
		os.Exit(1)
	}

	fmt.Println("Config is valid.")
	return nil
}
