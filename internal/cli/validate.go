package cli

import (
	"fmt"
	"strings"

	"github.com/h13/gtm-users/internal/config"
	"github.com/spf13/cobra"
)

func newValidateCmd(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the config file syntax and values",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(opts)
		},
	}
}

func runValidate(opts *rootOptions) error {
	cfg, err := config.Load(opts.configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := checkValidation(cfg); err != nil {
		return err
	}

	fmt.Println("Config is valid.")
	return nil
}

func checkValidation(cfg config.Config) error {
	errs := config.Validate(cfg)
	if len(errs) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString("validation errors:")
	for _, e := range errs {
		fmt.Fprintf(&b, "\n  - %s", e)
	}
	return fmt.Errorf("%s", b.String())
}
