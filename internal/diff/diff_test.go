package diff_test

import (
	"encoding/json"
	"testing"

	"github.com/h13/gtm-users/internal/config"
	"github.com/h13/gtm-users/internal/diff"
	"github.com/h13/gtm-users/internal/state"
)

func TestCompute_NoChanges(t *testing.T) {
	s := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{
				Email:         "alice@example.com",
				AccountAccess: "user",
				ContainerAccess: []state.ContainerPermission{
					{ContainerID: "GTM-AAAA1111", Permission: "publish"},
				},
			},
		},
	}

	plan := diff.Compute(s, s, config.ModeAdditive)
	if plan.HasChanges() {
		t.Errorf("expected no changes, got %d", len(plan.Changes))
	}
}

func TestCompute_AddUser(t *testing.T) {
	desired := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "user"},
			{Email: "bob@example.com", AccountAccess: "admin"},
		},
	}
	actual := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "user"},
		},
	}

	plan := diff.Compute(desired, actual, config.ModeAdditive)

	if !plan.HasChanges() {
		t.Fatal("expected changes")
	}
	adds, updates, deletes := plan.Summary()
	if adds != 1 || updates != 0 || deletes != 0 {
		t.Errorf("summary = %d add, %d update, %d delete; want 1/0/0", adds, updates, deletes)
	}
	if plan.Changes[0].Email != "bob@example.com" {
		t.Errorf("added user = %q, want %q", plan.Changes[0].Email, "bob@example.com")
	}
	if plan.Changes[0].Action != diff.ActionAdd {
		t.Errorf("action = %q, want %q", plan.Changes[0].Action, diff.ActionAdd)
	}
}

func TestCompute_UpdateAccountAccess(t *testing.T) {
	desired := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "admin"},
		},
	}
	actual := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "user"},
		},
	}

	plan := diff.Compute(desired, actual, config.ModeAdditive)

	if !plan.HasChanges() {
		t.Fatal("expected changes")
	}
	if plan.Changes[0].Action != diff.ActionUpdate {
		t.Errorf("action = %q, want %q", plan.Changes[0].Action, diff.ActionUpdate)
	}
	if plan.Changes[0].OldAccountAccess != "user" {
		t.Errorf("old access = %q, want %q", plan.Changes[0].OldAccountAccess, "user")
	}
	if plan.Changes[0].NewAccountAccess != "admin" {
		t.Errorf("new access = %q, want %q", plan.Changes[0].NewAccountAccess, "admin")
	}
}

func TestCompute_UpdateContainerPermission(t *testing.T) {
	desired := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{
				Email:         "alice@example.com",
				AccountAccess: "user",
				ContainerAccess: []state.ContainerPermission{
					{ContainerID: "GTM-AAAA1111", Permission: "publish"},
				},
			},
		},
	}
	actual := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{
				Email:         "alice@example.com",
				AccountAccess: "user",
				ContainerAccess: []state.ContainerPermission{
					{ContainerID: "GTM-AAAA1111", Permission: "read"},
				},
			},
		},
	}

	plan := diff.Compute(desired, actual, config.ModeAdditive)

	if !plan.HasChanges() {
		t.Fatal("expected changes")
	}
	change := plan.Changes[0]
	if change.Action != diff.ActionUpdate {
		t.Errorf("action = %q, want update", change.Action)
	}
	if len(change.ContainerChanges) != 1 {
		t.Fatalf("container changes = %d, want 1", len(change.ContainerChanges))
	}
	cc := change.ContainerChanges[0]
	if cc.Action != diff.ActionUpdate {
		t.Errorf("container action = %q, want update", cc.Action)
	}
	if cc.OldPermission != "read" || cc.NewPermission != "publish" {
		t.Errorf("permissions = %q → %q, want read → publish", cc.OldPermission, cc.NewPermission)
	}
}

func TestCompute_AddContainer(t *testing.T) {
	desired := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{
				Email:         "alice@example.com",
				AccountAccess: "user",
				ContainerAccess: []state.ContainerPermission{
					{ContainerID: "GTM-AAAA1111", Permission: "read"},
					{ContainerID: "GTM-BBBB2222", Permission: "edit"},
				},
			},
		},
	}
	actual := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{
				Email:         "alice@example.com",
				AccountAccess: "user",
				ContainerAccess: []state.ContainerPermission{
					{ContainerID: "GTM-AAAA1111", Permission: "read"},
				},
			},
		},
	}

	plan := diff.Compute(desired, actual, config.ModeAdditive)

	if !plan.HasChanges() {
		t.Fatal("expected changes")
	}
	change := plan.Changes[0]
	if len(change.ContainerChanges) != 1 {
		t.Fatalf("container changes = %d, want 1", len(change.ContainerChanges))
	}
	if change.ContainerChanges[0].Action != diff.ActionAdd {
		t.Errorf("container action = %q, want add", change.ContainerChanges[0].Action)
	}
}

func TestCompute_DeleteContainer(t *testing.T) {
	desired := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{
				Email:         "alice@example.com",
				AccountAccess: "user",
				ContainerAccess: []state.ContainerPermission{
					{ContainerID: "GTM-AAAA1111", Permission: "read"},
				},
			},
		},
	}
	actual := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{
				Email:         "alice@example.com",
				AccountAccess: "user",
				ContainerAccess: []state.ContainerPermission{
					{ContainerID: "GTM-AAAA1111", Permission: "read"},
					{ContainerID: "GTM-BBBB2222", Permission: "edit"},
				},
			},
		},
	}

	plan := diff.Compute(desired, actual, config.ModeAdditive)

	if !plan.HasChanges() {
		t.Fatal("expected changes")
	}
	change := plan.Changes[0]
	if len(change.ContainerChanges) != 1 {
		t.Fatalf("container changes = %d, want 1", len(change.ContainerChanges))
	}
	if change.ContainerChanges[0].Action != diff.ActionDelete {
		t.Errorf("container action = %q, want delete", change.ContainerChanges[0].Action)
	}
}

func TestCompute_Additive_NoDelete(t *testing.T) {
	desired := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "user"},
		},
	}
	actual := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "user"},
			{Email: "bob@example.com", AccountAccess: "admin"},
		},
	}

	plan := diff.Compute(desired, actual, config.ModeAdditive)

	if plan.HasChanges() {
		t.Errorf("additive mode should not delete unmanaged users, got %d changes", len(plan.Changes))
	}
}

func TestCompute_Authoritative_DeleteUser(t *testing.T) {
	desired := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "user"},
		},
	}
	actual := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "user"},
			{Email: "bob@example.com", AccountAccess: "admin"},
		},
	}

	plan := diff.Compute(desired, actual, config.ModeAuthoritative)

	if !plan.HasChanges() {
		t.Fatal("expected changes in authoritative mode")
	}
	adds, updates, deletes := plan.Summary()
	if adds != 0 || updates != 0 || deletes != 1 {
		t.Errorf("summary = %d/%d/%d, want 0/0/1", adds, updates, deletes)
	}
	if plan.Changes[0].Email != "bob@example.com" {
		t.Errorf("deleted user = %q, want %q", plan.Changes[0].Email, "bob@example.com")
	}
}

func TestCompute_ContainerChangesSorted(t *testing.T) {
	desired := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{
				Email:         "alice@example.com",
				AccountAccess: "user",
				ContainerAccess: []state.ContainerPermission{
					{ContainerID: "GTM-CCCC3333", Permission: "read"},
					{ContainerID: "GTM-AAAA1111", Permission: "edit"},
					{ContainerID: "GTM-BBBB2222", Permission: "publish"},
				},
			},
		},
	}
	actual := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "user"},
		},
	}

	plan := diff.Compute(desired, actual, config.ModeAdditive)

	if !plan.HasChanges() {
		t.Fatal("expected changes")
	}
	changes := plan.Changes[0].ContainerChanges
	if len(changes) != 3 {
		t.Fatalf("container changes = %d, want 3", len(changes))
	}
	if changes[0].ContainerID != "GTM-AAAA1111" {
		t.Errorf("changes[0] = %q, want GTM-AAAA1111", changes[0].ContainerID)
	}
	if changes[1].ContainerID != "GTM-BBBB2222" {
		t.Errorf("changes[1] = %q, want GTM-BBBB2222", changes[1].ContainerID)
	}
	if changes[2].ContainerID != "GTM-CCCC3333" {
		t.Errorf("changes[2] = %q, want GTM-CCCC3333", changes[2].ContainerID)
	}
}

func TestCompute_NoChanges_JSONEmptyArray(t *testing.T) {
	s := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "user"},
		},
	}

	plan := diff.Compute(s, s, config.ModeAdditive)

	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	changes := string(result["changes"])
	if changes != "[]" {
		t.Errorf("changes = %s, want []", changes)
	}
}

func TestCompute_AddUserWithContainers(t *testing.T) {
	desired := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{
				Email:         "alice@example.com",
				AccountAccess: "user",
				ContainerAccess: []state.ContainerPermission{
					{ContainerID: "GTM-AAAA1111", Permission: "read"},
				},
			},
		},
	}
	actual := state.AccountState{AccountID: "123"}

	plan := diff.Compute(desired, actual, config.ModeAdditive)

	if !plan.HasChanges() {
		t.Fatal("expected changes")
	}
	if plan.Changes[0].Action != diff.ActionAdd {
		t.Errorf("action = %q, want add", plan.Changes[0].Action)
	}
	if len(plan.Changes[0].ContainerChanges) != 1 {
		t.Fatalf("container changes = %d, want 1", len(plan.Changes[0].ContainerChanges))
	}
	cc := plan.Changes[0].ContainerChanges[0]
	if cc.ContainerID != "GTM-AAAA1111" || cc.NewPermission != "read" || cc.Action != diff.ActionAdd {
		t.Errorf("unexpected container change: %+v", cc)
	}
}

func TestCompute_AddUserNoContainers(t *testing.T) {
	desired := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "user"},
		},
	}
	actual := state.AccountState{AccountID: "123"}

	plan := diff.Compute(desired, actual, config.ModeAdditive)

	if !plan.HasChanges() {
		t.Fatal("expected changes")
	}
	if len(plan.Changes[0].ContainerChanges) != 0 {
		t.Errorf("container changes = %d, want 0", len(plan.Changes[0].ContainerChanges))
	}
}

func TestCompute_MixedExistingAndNew(t *testing.T) {
	desired := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "user"},
			{Email: "bob@example.com", AccountAccess: "admin"},
		},
	}
	actual := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "user"},
		},
	}

	plan := diff.Compute(desired, actual, config.ModeAuthoritative)

	adds, _, _ := plan.Summary()
	if adds != 1 {
		t.Errorf("adds = %d, want 1", adds)
	}
}

func TestCompute_ComplexScenario(t *testing.T) {
	desired := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{
				Email:         "alice@example.com",
				AccountAccess: "admin", // changed from user
				ContainerAccess: []state.ContainerPermission{
					{ContainerID: "GTM-AAAA1111", Permission: "publish"}, // changed from read
				},
			},
			{
				Email:         "carol@example.com", // new user
				AccountAccess: "user",
			},
		},
	}
	actual := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{
				Email:         "alice@example.com",
				AccountAccess: "user",
				ContainerAccess: []state.ContainerPermission{
					{ContainerID: "GTM-AAAA1111", Permission: "read"},
				},
			},
			{
				Email:         "bob@example.com", // to be deleted in authoritative
				AccountAccess: "user",
			},
		},
	}

	plan := diff.Compute(desired, actual, config.ModeAuthoritative)

	adds, updates, deletes := plan.Summary()
	if adds != 1 || updates != 1 || deletes != 1 {
		t.Errorf("summary = %d/%d/%d, want 1/1/1", adds, updates, deletes)
	}
}
