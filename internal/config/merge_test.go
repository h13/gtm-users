package config_test

import (
	"testing"

	"github.com/h13/gtm-users/internal/config"
)

func TestMergeConfigs_RoleMerging(t *testing.T) {
	base := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Roles: map[string]config.Role{
			"viewer": {AccountAccess: config.AccountAccessUser},
		},
		Users: []config.User{
			{Email: "alice@example.com", Role: "viewer"},
		},
	}

	included := config.Config{
		Roles: map[string]config.Role{
			"editor": {AccountAccess: config.AccountAccessUser},
		},
		Users: []config.User{
			{Email: "bob@example.com", Role: "editor"},
		},
	}

	merged := config.MergeConfigs(base, included)

	if len(merged.Roles) != 2 {
		t.Fatalf("roles count = %d, want 2", len(merged.Roles))
	}
	if _, ok := merged.Roles["viewer"]; !ok {
		t.Error("missing viewer role")
	}
	if _, ok := merged.Roles["editor"]; !ok {
		t.Error("missing editor role")
	}
}

func TestMergeConfigs_RoleOverride(t *testing.T) {
	base := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Roles: map[string]config.Role{
			"viewer": {AccountAccess: config.AccountAccessAdmin},
		},
	}

	included := config.Config{
		Roles: map[string]config.Role{
			"viewer": {AccountAccess: config.AccountAccessUser},
		},
	}

	merged := config.MergeConfigs(base, included)

	if merged.Roles["viewer"].AccountAccess != config.AccountAccessAdmin {
		t.Errorf("base role should override, got %q", merged.Roles["viewer"].AccountAccess)
	}
}

func TestMergeConfigs_UserAppending(t *testing.T) {
	base := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Users: []config.User{
			{Email: "alice@example.com", AccountAccess: config.AccountAccessUser},
		},
	}

	included := config.Config{
		Users: []config.User{
			{Email: "bob@example.com", AccountAccess: config.AccountAccessAdmin},
			{Email: "carol@example.com", AccountAccess: config.AccountAccessUser},
		},
	}

	merged := config.MergeConfigs(base, included)

	if len(merged.Users) != 3 {
		t.Fatalf("users count = %d, want 3", len(merged.Users))
	}
	if merged.Users[0].Email != "alice@example.com" {
		t.Errorf("users[0] = %q, want alice", merged.Users[0].Email)
	}
	if merged.Users[1].Email != "bob@example.com" {
		t.Errorf("users[1] = %q, want bob", merged.Users[1].Email)
	}
	if merged.Users[2].Email != "carol@example.com" {
		t.Errorf("users[2] = %q, want carol", merged.Users[2].Email)
	}
}

func TestMergeConfigs_PolicyNotMerged(t *testing.T) {
	maxAdmins := 2
	base := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Policy:    &config.Policy{MaxAdmins: &maxAdmins},
	}

	includedMax := 5
	included := config.Config{
		Policy: &config.Policy{MaxAdmins: &includedMax},
	}

	merged := config.MergeConfigs(base, included)

	if merged.Policy == nil {
		t.Fatal("expected policy from base")
	}
	if *merged.Policy.MaxAdmins != 2 {
		t.Errorf("max_admins = %d, want 2 (base policy should be used)", *merged.Policy.MaxAdmins)
	}
}

func TestMergeConfigs_PreservesBaseFields(t *testing.T) {
	base := config.Config{
		AccountID: "123",
		Mode:      config.ModeAuthoritative,
	}

	included := config.Config{
		AccountID: "456",
		Mode:      config.ModeAdditive,
	}

	merged := config.MergeConfigs(base, included)

	if merged.AccountID != "123" {
		t.Errorf("account_id = %q, want 123", merged.AccountID)
	}
	if merged.Mode != config.ModeAuthoritative {
		t.Errorf("mode = %q, want authoritative", merged.Mode)
	}
}

func TestMergeConfigs_NilRoles(t *testing.T) {
	base := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
	}
	included := config.Config{}

	merged := config.MergeConfigs(base, included)

	if merged.Roles != nil {
		t.Errorf("expected nil roles when both are nil, got %v", merged.Roles)
	}
}

func TestMergeConfigs_BaseNilIncludedHasRoles(t *testing.T) {
	base := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
	}
	included := config.Config{
		Roles: map[string]config.Role{
			"editor": {AccountAccess: config.AccountAccessUser},
		},
	}

	merged := config.MergeConfigs(base, included)

	if len(merged.Roles) != 1 {
		t.Errorf("roles count = %d, want 1", len(merged.Roles))
	}
}

func TestMergeConfigs_DoesNotMutateInputs(t *testing.T) {
	base := config.Config{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Users: []config.User{
			{Email: "alice@example.com", AccountAccess: config.AccountAccessUser},
		},
	}
	included := config.Config{
		Users: []config.User{
			{Email: "bob@example.com", AccountAccess: config.AccountAccessAdmin},
		},
	}

	_ = config.MergeConfigs(base, included)

	if len(base.Users) != 1 {
		t.Errorf("base.Users was mutated, len = %d, want 1", len(base.Users))
	}
	if len(included.Users) != 1 {
		t.Errorf("included.Users was mutated, len = %d, want 1", len(included.Users))
	}
}
