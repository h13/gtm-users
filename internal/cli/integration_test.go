package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/h13/gtm-users/internal/state"
	"github.com/spf13/cobra"
)

func newTestRootCmd(t *testing.T, mock *mockClient, configYAML string) (*bytes.Buffer, func(args ...string) error) {
	t.Helper()
	path := writeTempConfig(t, configYAML)

	opts := &rootOptions{
		configPath:      path,
		credentialsPath: "fake-creds.json",
		format:          "text",
		stdout:          &bytes.Buffer{},
		stdin:           strings.NewReader("yes\n"),
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return mock, nil
		},
	}

	cmd := &cobra.Command{
		Use:           "gtm-users",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.PersistentFlags().StringVar(&opts.configPath, "config", path, "path to config file")
	cmd.PersistentFlags().StringVar(&opts.credentialsPath, "credentials", "fake-creds.json", "credentials")
	cmd.PersistentFlags().StringVar(&opts.format, "format", "text", "output format")

	cmd.AddCommand(
		newValidateCmd(opts),
		newExportCmd(opts),
		newPlanCmd(opts),
		newApplyCmd(opts),
	)

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	return &stdout, func(args ...string) error {
		cmd.SetArgs(args)
		return cmd.Execute()
	}
}

func TestIntegration_ValidateCommand(t *testing.T) {
	_, run := newTestRootCmd(t, &mockClient{}, validConfig)

	if err := run("validate"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIntegration_ValidateCommand_InvalidConfig(t *testing.T) {
	_, run := newTestRootCmd(t, &mockClient{}, invalidConfig)

	err := run("validate")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestIntegration_PlanCommand(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{AccountID: "123456789"},
	}
	_, run := newTestRootCmd(t, mock, validConfig)

	if err := run("plan"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIntegration_PlanCommand_JSONFormat(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{AccountID: "123456789"},
	}
	_, run := newTestRootCmd(t, mock, validConfig)

	if err := run("plan", "--format", "json"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIntegration_ApplyCommand_AutoApprove(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{AccountID: "123456789"},
	}
	_, run := newTestRootCmd(t, mock, validConfig)

	if err := run("apply", "--auto-approve"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIntegration_ExportCommand(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{
			AccountID: "123",
			Users: []state.UserPermission{
				{Email: "alice@example.com", AccountAccess: "user"},
			},
		},
	}
	_, run := newTestRootCmd(t, mock, validConfig)

	if err := run("export", "--account-id", "123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIntegration_VersionFlag(t *testing.T) {
	cmd := NewRootCmd("1.2.3")
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
