package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/h13/gtm-users/internal/config"
	"github.com/h13/gtm-users/internal/diff"
	"github.com/h13/gtm-users/internal/state"
)

type mockClient struct {
	createCalls []state.UserPermission
	updateCalls []state.UserPermission
	deleteCalls []string
	createErr   error
	updateErr   error
	deleteErr   error
	fetchState  state.AccountState
	fetchErr    error
}

func (m *mockClient) FetchState(_ context.Context) (state.AccountState, error) {
	return m.fetchState, m.fetchErr
}

func (m *mockClient) CreateUserPermission(_ context.Context, user state.UserPermission) error {
	m.createCalls = append(m.createCalls, user)
	return m.createErr
}

func (m *mockClient) UpdateUserPermission(_ context.Context, user state.UserPermission) error {
	m.updateCalls = append(m.updateCalls, user)
	return m.updateErr
}

func (m *mockClient) DeleteUserPermission(_ context.Context, email string) error {
	m.deleteCalls = append(m.deleteCalls, email)
	return m.deleteErr
}

func TestExecuteChanges_AllActions(t *testing.T) {
	mock := &mockClient{}
	desired := state.AccountState{
		AccountID: "123",
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "admin"},
			{Email: "bob@example.com", AccountAccess: "user"},
		},
	}
	plan := diff.Plan{
		AccountID: "123",
		Mode:      config.ModeAuthoritative,
		Changes: []diff.UserChange{
			{Email: "alice@example.com", Action: diff.ActionAdd, NewAccountAccess: "admin"},
			{Email: "bob@example.com", Action: diff.ActionUpdate, OldAccountAccess: "user", NewAccountAccess: "admin"},
			{Email: "carol@example.com", Action: diff.ActionDelete, OldAccountAccess: "user"},
		},
	}

	err := executeChanges(context.Background(), mock, plan, desired)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.createCalls) != 1 {
		t.Errorf("create calls = %d, want 1", len(mock.createCalls))
	}
	if mock.createCalls[0].Email != "alice@example.com" {
		t.Errorf("created user = %q, want alice@example.com", mock.createCalls[0].Email)
	}

	if len(mock.updateCalls) != 1 {
		t.Errorf("update calls = %d, want 1", len(mock.updateCalls))
	}
	if mock.updateCalls[0].Email != "bob@example.com" {
		t.Errorf("updated user = %q, want bob@example.com", mock.updateCalls[0].Email)
	}

	if len(mock.deleteCalls) != 1 {
		t.Errorf("delete calls = %d, want 1", len(mock.deleteCalls))
	}
	if mock.deleteCalls[0] != "carol@example.com" {
		t.Errorf("deleted user = %q, want carol@example.com", mock.deleteCalls[0])
	}
}

func TestExecuteChanges_CreateError(t *testing.T) {
	mock := &mockClient{
		createErr: errors.New("API error"),
	}
	plan := diff.Plan{
		AccountID: "123",
		Changes: []diff.UserChange{
			{Email: "alice@example.com", Action: diff.ActionAdd},
		},
	}

	err := executeChanges(context.Background(), mock, plan, state.AccountState{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "1 operation(s) failed" {
		t.Errorf("error = %q, want '1 operation(s) failed'", err.Error())
	}
}

func TestExecuteChanges_ContinuesOnError(t *testing.T) {
	mock := &mockClient{
		createErr: errors.New("API error"),
	}
	plan := diff.Plan{
		AccountID: "123",
		Changes: []diff.UserChange{
			{Email: "alice@example.com", Action: diff.ActionAdd},
			{Email: "bob@example.com", Action: diff.ActionDelete},
		},
	}

	_ = executeChanges(context.Background(), mock, plan, state.AccountState{})

	if len(mock.deleteCalls) != 1 {
		t.Errorf("delete calls = %d, want 1 (should continue after create error)", len(mock.deleteCalls))
	}
}

func TestFindDesiredUser_Found(t *testing.T) {
	s := state.AccountState{
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "admin"},
			{Email: "bob@example.com", AccountAccess: "user"},
		},
	}

	user := findDesiredUser(s, "bob@example.com")
	if user.Email != "bob@example.com" {
		t.Errorf("email = %q, want bob@example.com", user.Email)
	}
	if user.AccountAccess != "user" {
		t.Errorf("access = %q, want user", user.AccountAccess)
	}
}

func TestFindDesiredUser_NotFound(t *testing.T) {
	s := state.AccountState{
		Users: []state.UserPermission{
			{Email: "alice@example.com", AccountAccess: "admin"},
		},
	}

	user := findDesiredUser(s, "unknown@example.com")
	if user.Email != "unknown@example.com" {
		t.Errorf("email = %q, want unknown@example.com", user.Email)
	}
	if user.AccountAccess != "" {
		t.Errorf("access = %q, want empty", user.AccountAccess)
	}
}

func TestApplyChange_UpdateError(t *testing.T) {
	mock := &mockClient{
		updateErr: errors.New("update failed"),
	}
	change := diff.UserChange{
		Email:  "alice@example.com",
		Action: diff.ActionUpdate,
	}

	err := applyChange(context.Background(), mock, change, state.AccountState{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(mock.updateCalls) != 1 {
		t.Errorf("update calls = %d, want 1", len(mock.updateCalls))
	}
}

func TestApplyChange_DeleteError(t *testing.T) {
	mock := &mockClient{
		deleteErr: errors.New("delete failed"),
	}
	change := diff.UserChange{
		Email:  "alice@example.com",
		Action: diff.ActionDelete,
	}

	err := applyChange(context.Background(), mock, change, state.AccountState{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(mock.deleteCalls) != 1 {
		t.Errorf("delete calls = %d, want 1", len(mock.deleteCalls))
	}
}
