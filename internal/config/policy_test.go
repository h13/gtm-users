package config_test

import (
	"strings"
	"testing"

	"github.com/h13/gtm-users/internal/config"
)

func intPtr(n int) *int { return &n }

func TestValidatePolicy_NilPolicy(t *testing.T) {
	cfg := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Users: []config.User{
			{Email: "a@b.com", AccountAccess: config.AccountAccessAdmin},
		},
	}

	errs := config.ValidatePolicy(cfg)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidatePolicy_MaxAdminsOK(t *testing.T) {
	cfg := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Policy:    &config.Policy{MaxAdmins: intPtr(2)},
		Users: []config.User{
			{Email: "a@b.com", AccountAccess: config.AccountAccessAdmin},
			{Email: "b@c.com", AccountAccess: config.AccountAccessUser},
		},
	}

	errs := config.ValidatePolicy(cfg)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidatePolicy_MaxAdminsExceeded(t *testing.T) {
	cfg := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Policy:    &config.Policy{MaxAdmins: intPtr(1)},
		Users: []config.User{
			{Email: "a@b.com", AccountAccess: config.AccountAccessAdmin},
			{Email: "b@c.com", AccountAccess: config.AccountAccessAdmin},
		},
	}

	errs := config.ValidatePolicy(cfg)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Message, "exceeds limit") {
		t.Errorf("error = %q, want 'exceeds limit' message", errs[0].Message)
	}
}

func TestValidatePolicy_MaxAdminsNil(t *testing.T) {
	cfg := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Policy:    &config.Policy{},
		Users: []config.User{
			{Email: "a@b.com", AccountAccess: config.AccountAccessAdmin},
		},
	}

	errs := config.ValidatePolicy(cfg)
	if len(errs) != 0 {
		t.Errorf("expected no errors when max_admins is nil, got %v", errs)
	}
}

func TestValidatePolicy_RequireApproveOK(t *testing.T) {
	cfg := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Policy: &config.Policy{
			RequireApprove: []string{"GTM-AAAA1111"},
		},
		Users: []config.User{
			{
				Email:         "a@b.com",
				AccountAccess: config.AccountAccessUser,
				ContainerAccess: []config.ContainerAccess{
					{ContainerID: "GTM-AAAA1111", Permission: config.PermissionPublish},
					{ContainerID: "GTM-AAAA1111", Permission: config.PermissionApprove},
				},
			},
		},
	}

	errs := config.ValidatePolicy(cfg)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidatePolicy_RequireApproveViolation(t *testing.T) {
	cfg := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Policy: &config.Policy{
			RequireApprove: []string{"GTM-AAAA1111"},
		},
		Users: []config.User{
			{
				Email:         "a@b.com",
				AccountAccess: config.AccountAccessUser,
				ContainerAccess: []config.ContainerAccess{
					{ContainerID: "GTM-AAAA1111", Permission: config.PermissionPublish},
				},
			},
		},
	}

	errs := config.ValidatePolicy(cfg)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Message, "lacks approve") {
		t.Errorf("error = %q, want 'lacks approve' message", errs[0].Message)
	}
}

func TestValidatePolicy_RequireApproveUnrelatedContainer(t *testing.T) {
	cfg := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Policy: &config.Policy{
			RequireApprove: []string{"GTM-AAAA1111"},
		},
		Users: []config.User{
			{
				Email:         "a@b.com",
				AccountAccess: config.AccountAccessUser,
				ContainerAccess: []config.ContainerAccess{
					{ContainerID: "GTM-BBBB2222", Permission: config.PermissionPublish},
				},
			},
		},
	}

	errs := config.ValidatePolicy(cfg)
	if len(errs) != 0 {
		t.Errorf("expected no errors for unrelated container, got %v", errs)
	}
}

func TestValidatePolicy_EmptyRequireApprove(t *testing.T) {
	cfg := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Policy: &config.Policy{
			RequireApprove: []string{},
		},
		Users: []config.User{
			{
				Email:         "a@b.com",
				AccountAccess: config.AccountAccessUser,
				ContainerAccess: []config.ContainerAccess{
					{ContainerID: "GTM-AAAA1111", Permission: config.PermissionPublish},
				},
			},
		},
	}

	errs := config.ValidatePolicy(cfg)
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty require_approve, got %v", errs)
	}
}

func TestValidate_IncludesPolicy(t *testing.T) {
	cfg := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Policy:    &config.Policy{MaxAdmins: intPtr(0)},
		Users: []config.User{
			{Email: "a@b.com", AccountAccess: config.AccountAccessAdmin},
		},
	}

	errs := config.Validate(cfg)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Field, "policy") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected policy error in Validate, got %v", errs)
	}
}

func TestParse_WithPolicy(t *testing.T) {
	data := []byte(`
account_id: "123"
mode: additive
policy:
  max_admins: 2
  require_approve_for_publish:
    - "GTM-AAAA1111"
users:
  - email: alice@example.com
    account_access: user
`)

	cfg, err := config.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Policy == nil {
		t.Fatal("expected policy to be set")
	}
	if *cfg.Policy.MaxAdmins != 2 {
		t.Errorf("max_admins = %d, want 2", *cfg.Policy.MaxAdmins)
	}
	if len(cfg.Policy.RequireApprove) != 1 {
		t.Fatalf("require_approve len = %d, want 1", len(cfg.Policy.RequireApprove))
	}
}
