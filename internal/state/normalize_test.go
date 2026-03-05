package state_test

import (
	"testing"

	"github.com/h13/gtm-users/internal/config"
	"github.com/h13/gtm-users/internal/state"
)

func TestFromConfig(t *testing.T) {
	cfg := config.Config{
		AccountID: "123456789",
		Mode:      config.ModeAdditive,
		Users: []config.User{
			{
				Email:         "Bob@Example.com",
				AccountAccess: config.AccountAccessUser,
				ContainerAccess: []config.ContainerAccess{
					{ContainerID: "GTM-BBBB2222", Permission: config.PermissionEdit},
					{ContainerID: "GTM-AAAA1111", Permission: config.PermissionRead},
				},
			},
			{
				Email:         "Alice@Example.com",
				AccountAccess: config.AccountAccessAdmin,
			},
		},
	}

	s := state.FromConfig(cfg)

	if s.AccountID != "123456789" {
		t.Errorf("account_id = %q, want %q", s.AccountID, "123456789")
	}

	if len(s.Users) != 2 {
		t.Fatalf("len(users) = %d, want 2", len(s.Users))
	}

	// Users should be sorted by email (lowercase)
	if s.Users[0].Email != "alice@example.com" {
		t.Errorf("users[0].email = %q, want %q", s.Users[0].Email, "alice@example.com")
	}
	if s.Users[1].Email != "bob@example.com" {
		t.Errorf("users[1].email = %q, want %q", s.Users[1].Email, "bob@example.com")
	}

	// Containers should be sorted by container ID
	bob := s.Users[1]
	if len(bob.ContainerAccess) != 2 {
		t.Fatalf("len(containers) = %d, want 2", len(bob.ContainerAccess))
	}
	if bob.ContainerAccess[0].ContainerID != "GTM-AAAA1111" {
		t.Errorf("containers[0] = %q, want GTM-AAAA1111", bob.ContainerAccess[0].ContainerID)
	}
	if bob.ContainerAccess[1].ContainerID != "GTM-BBBB2222" {
		t.Errorf("containers[1] = %q, want GTM-BBBB2222", bob.ContainerAccess[1].ContainerID)
	}
}

func TestUserMap(t *testing.T) {
	s := state.AccountState{
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "admin"},
			{Email: "bob@example.com", AccountAccess: "user"},
		},
	}

	m := s.UserMap()
	if len(m) != 2 {
		t.Fatalf("len(map) = %d, want 2", len(m))
	}
	if m["alice@example.com"].AccountAccess != "admin" {
		t.Errorf("alice access = %q, want admin", m["alice@example.com"].AccountAccess)
	}
}

func TestContainerMap(t *testing.T) {
	u := state.UserPermission{
		ContainerAccess: []state.ContainerPermission{
			{ContainerID: "GTM-AAAA1111", Permission: "read"},
			{ContainerID: "GTM-BBBB2222", Permission: "edit"},
		},
	}

	m := u.ContainerMap()
	if len(m) != 2 {
		t.Fatalf("len(map) = %d, want 2", len(m))
	}
	if m["GTM-AAAA1111"] != "read" {
		t.Errorf("GTM-AAAA1111 = %q, want read", m["GTM-AAAA1111"])
	}
}
