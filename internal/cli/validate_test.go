package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/h13/gtm-users/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "gtm-users.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	return path
}

const validConfig = `
account_id: "123456789"
mode: additive
users:
  - email: alice@example.com
    account_access: user
    container_access:
      - container_id: "GTM-AAAA1111"
        permission: publish
`

const invalidConfig = `
account_id: ""
mode: invalid
users: []
`

func TestRunValidate_Valid(t *testing.T) {
	path := writeTempConfig(t, validConfig)
	opts := &rootOptions{configPath: path}

	if err := runValidate(opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunValidate_FileNotFound(t *testing.T) {
	opts := &rootOptions{configPath: "/nonexistent/path.yaml"}

	err := runValidate(opts)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestRunValidate_InvalidConfig(t *testing.T) {
	path := writeTempConfig(t, invalidConfig)
	opts := &rootOptions{configPath: path}

	err := runValidate(opts)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestCheckValidation_NoErrors(t *testing.T) {
	cfg := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Users: []config.User{
			{Email: "a@b.com", AccountAccess: config.AccountAccessUser},
		},
	}

	if err := checkValidation(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckValidation_MultipleErrors(t *testing.T) {
	cfg := config.Config{
		AccountID: "",
		Mode:      "bad",
	}

	err := checkValidation(cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "validation errors:") {
		t.Errorf("error should contain 'validation errors:', got %q", err.Error())
	}
}
