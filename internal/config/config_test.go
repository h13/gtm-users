package config_test

import (
	"testing"

	"github.com/h13/gtm-users/internal/config"
)

func TestParse(t *testing.T) {
	data := []byte(`
account_id: "123456789"
mode: additive
users:
  - email: alice@example.com
    account_access: user
    container_access:
      - container_id: "GTM-AAAA1111"
        permission: publish
  - email: bob@example.com
    account_access: admin
`)

	cfg, err := config.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AccountID != "123456789" {
		t.Errorf("account_id = %q, want %q", cfg.AccountID, "123456789")
	}
	if cfg.Mode != config.ModeAdditive {
		t.Errorf("mode = %q, want %q", cfg.Mode, config.ModeAdditive)
	}
	if len(cfg.Users) != 2 {
		t.Fatalf("len(users) = %d, want 2", len(cfg.Users))
	}

	alice := cfg.Users[0]
	if alice.Email != "alice@example.com" {
		t.Errorf("users[0].email = %q, want %q", alice.Email, "alice@example.com")
	}
	if alice.AccountAccess != config.AccountAccessUser {
		t.Errorf("users[0].account_access = %q, want %q", alice.AccountAccess, config.AccountAccessUser)
	}
	if len(alice.ContainerAccess) != 1 {
		t.Fatalf("len(users[0].container_access) = %d, want 1", len(alice.ContainerAccess))
	}
	if alice.ContainerAccess[0].ContainerID != "GTM-AAAA1111" {
		t.Errorf("container_id = %q, want %q", alice.ContainerAccess[0].ContainerID, "GTM-AAAA1111")
	}
	if alice.ContainerAccess[0].Permission != config.PermissionPublish {
		t.Errorf("permission = %q, want %q", alice.ContainerAccess[0].Permission, config.PermissionPublish)
	}
}

func TestParseDefaultMode(t *testing.T) {
	data := []byte(`
account_id: "123456789"
users:
  - email: alice@example.com
    account_access: user
`)

	cfg, err := config.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Mode != config.ModeAdditive {
		t.Errorf("default mode = %q, want %q", cfg.Mode, config.ModeAdditive)
	}
}

func TestLoad(t *testing.T) {
	cfg, err := config.Load("../../testdata/gtm-users.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AccountID != "123456789" {
		t.Errorf("account_id = %q, want %q", cfg.AccountID, "123456789")
	}
	if len(cfg.Users) != 3 {
		t.Errorf("len(users) = %d, want 3", len(cfg.Users))
	}
}

func TestParse_UnknownField(t *testing.T) {
	data := []byte(`
account_id: "123"
unknown_field: oops
users:
  - email: a@b.com
    account_access: user
`)

	_, err := config.Parse(data)
	if err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}
}

func TestLoadNotFound(t *testing.T) {
	_, err := config.Load("nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestValidate_Valid(t *testing.T) {
	cfg := config.Config{
		AccountID: "123456789",
		Mode:      config.ModeAdditive,
		Users: []config.User{
			{
				Email:         "alice@example.com",
				AccountAccess: config.AccountAccessUser,
				ContainerAccess: []config.ContainerAccess{
					{ContainerID: "GTM-AAAA1111", Permission: config.PermissionPublish},
				},
			},
		},
	}

	errs := config.Validate(cfg)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidationError_Error(t *testing.T) {
	e := config.ValidationError{Field: "users[0].email", Message: "required"}
	got := e.Error()
	want := "users[0].email: required"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestValidate_Errors(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.Config
		wantErrs int
	}{
		{
			name:     "empty account_id",
			cfg:      config.Config{Mode: config.ModeAdditive, Users: []config.User{{Email: "a@b.com", AccountAccess: config.AccountAccessUser}}},
			wantErrs: 1,
		},
		{
			name:     "invalid mode",
			cfg:      config.Config{AccountID: "123", Mode: "invalid", Users: []config.User{{Email: "a@b.com", AccountAccess: config.AccountAccessUser}}},
			wantErrs: 1,
		},
		{
			name:     "no users",
			cfg:      config.Config{AccountID: "123", Mode: config.ModeAdditive},
			wantErrs: 1,
		},
		{
			name: "invalid email",
			cfg: config.Config{
				AccountID: "123",
				Mode:      config.ModeAdditive,
				Users:     []config.User{{Email: "not-email", AccountAccess: config.AccountAccessUser}},
			},
			wantErrs: 1,
		},
		{
			name: "duplicate email",
			cfg: config.Config{
				AccountID: "123",
				Mode:      config.ModeAdditive,
				Users: []config.User{
					{Email: "a@b.com", AccountAccess: config.AccountAccessUser},
					{Email: "a@b.com", AccountAccess: config.AccountAccessAdmin},
				},
			},
			wantErrs: 1,
		},
		{
			name: "invalid account_access",
			cfg: config.Config{
				AccountID: "123",
				Mode:      config.ModeAdditive,
				Users:     []config.User{{Email: "a@b.com", AccountAccess: "superadmin"}},
			},
			wantErrs: 1,
		},
		{
			name: "invalid container_id format",
			cfg: config.Config{
				AccountID: "123",
				Mode:      config.ModeAdditive,
				Users: []config.User{{
					Email:         "a@b.com",
					AccountAccess: config.AccountAccessUser,
					ContainerAccess: []config.ContainerAccess{
						{ContainerID: "INVALID", Permission: config.PermissionRead},
					},
				}},
			},
			wantErrs: 1,
		},
		{
			name: "invalid permission",
			cfg: config.Config{
				AccountID: "123",
				Mode:      config.ModeAdditive,
				Users: []config.User{{
					Email:         "a@b.com",
					AccountAccess: config.AccountAccessUser,
					ContainerAccess: []config.ContainerAccess{
						{ContainerID: "GTM-AAAA1111", Permission: "superwrite"},
					},
				}},
			},
			wantErrs: 1,
		},
		{
			name: "duplicate container_id",
			cfg: config.Config{
				AccountID: "123",
				Mode:      config.ModeAdditive,
				Users: []config.User{{
					Email:         "a@b.com",
					AccountAccess: config.AccountAccessUser,
					ContainerAccess: []config.ContainerAccess{
						{ContainerID: "GTM-AAAA1111", Permission: config.PermissionRead},
						{ContainerID: "GTM-AAAA1111", Permission: config.PermissionEdit},
					},
				}},
			},
			wantErrs: 1,
		},
		{
			name: "multiple empty emails reported as required not duplicate",
			cfg: config.Config{
				AccountID: "123",
				Mode:      config.ModeAdditive,
				Users: []config.User{
					{Email: "", AccountAccess: config.AccountAccessUser},
					{Email: "", AccountAccess: config.AccountAccessUser},
				},
			},
			wantErrs: 2,
		},
		{
			name: "multiple empty container_ids reported as required not duplicate",
			cfg: config.Config{
				AccountID: "123",
				Mode:      config.ModeAdditive,
				Users: []config.User{{
					Email:         "a@b.com",
					AccountAccess: config.AccountAccessUser,
					ContainerAccess: []config.ContainerAccess{
						{ContainerID: "", Permission: config.PermissionRead},
						{ContainerID: "", Permission: config.PermissionEdit},
					},
				}},
			},
			wantErrs: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := config.Validate(tt.cfg)
			if len(errs) != tt.wantErrs {
				t.Errorf("got %d errors, want %d: %v", len(errs), tt.wantErrs, errs)
			}
		})
	}
}

func TestParse_WithRoles(t *testing.T) {
	data := []byte(`
account_id: "123"
mode: additive
roles:
  viewer:
    account_access: user
    container_access:
      - container_id: "GTM-AAAA1111"
        permission: read
users:
  - email: alice@example.com
    role: viewer
`)

	cfg, err := config.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Users[0].AccountAccess != config.AccountAccessUser {
		t.Errorf("account_access = %q, want %q", cfg.Users[0].AccountAccess, config.AccountAccessUser)
	}
	if len(cfg.Users[0].ContainerAccess) != 1 {
		t.Fatalf("container_access len = %d, want 1", len(cfg.Users[0].ContainerAccess))
	}
	if cfg.Users[0].ContainerAccess[0].Permission != config.PermissionRead {
		t.Errorf("permission = %q, want %q", cfg.Users[0].ContainerAccess[0].Permission, config.PermissionRead)
	}
}

func TestParse_RoleOverride(t *testing.T) {
	data := []byte(`
account_id: "123"
mode: additive
roles:
  viewer:
    account_access: user
    container_access:
      - container_id: "GTM-AAAA1111"
        permission: read
users:
  - email: alice@example.com
    role: viewer
    account_access: admin
    container_access:
      - container_id: "GTM-BBBB2222"
        permission: publish
`)

	cfg, err := config.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Users[0].AccountAccess != config.AccountAccessAdmin {
		t.Errorf("account_access = %q, want %q (override)", cfg.Users[0].AccountAccess, config.AccountAccessAdmin)
	}
	if len(cfg.Users[0].ContainerAccess) != 1 {
		t.Fatalf("container_access len = %d, want 1", len(cfg.Users[0].ContainerAccess))
	}
	if cfg.Users[0].ContainerAccess[0].ContainerID != "GTM-BBBB2222" {
		t.Errorf("container_id = %q, want GTM-BBBB2222 (override)", cfg.Users[0].ContainerAccess[0].ContainerID)
	}
}

func TestParse_UndefinedRole(t *testing.T) {
	data := []byte(`
account_id: "123"
mode: additive
roles:
  viewer:
    account_access: user
users:
  - email: alice@example.com
    role: editor
`)

	_, err := config.Parse(data)
	if err == nil {
		t.Fatal("expected error for undefined role, got nil")
	}
}

func TestResolveRoles_NoRoles(t *testing.T) {
	cfg := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Users: []config.User{
			{Email: "a@b.com", AccountAccess: config.AccountAccessUser},
		},
	}

	resolved, err := config.ResolveRoles(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resolved.Users) != 1 {
		t.Errorf("users len = %d, want 1", len(resolved.Users))
	}
}

func TestResolveRoles_RolePlusInline(t *testing.T) {
	cfg := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Roles: map[string]config.Role{
			"viewer": {
				AccountAccess: config.AccountAccessUser,
				ContainerAccess: []config.ContainerAccess{
					{ContainerID: "GTM-AAAA1111", Permission: config.PermissionRead},
				},
			},
		},
		Users: []config.User{
			{Email: "a@b.com", Role: "viewer"},
			{Email: "b@c.com", AccountAccess: config.AccountAccessAdmin},
		},
	}

	resolved, err := config.ResolveRoles(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Users[0].AccountAccess != config.AccountAccessUser {
		t.Errorf("users[0].account_access = %q, want user", resolved.Users[0].AccountAccess)
	}
	if resolved.Users[1].AccountAccess != config.AccountAccessAdmin {
		t.Errorf("users[1].account_access = %q, want admin", resolved.Users[1].AccountAccess)
	}
}
