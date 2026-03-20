package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/h13/gtm-users/internal/state"
)

func newTestOpts(t *testing.T, mock *mockClient, configYAML string) *rootOptions {
	t.Helper()
	path := writeTempConfig(t, configYAML)
	return &rootOptions{
		configPath:      path,
		credentialsPath: "fake-creds.json",
		format:          "text",
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return mock, nil
		},
	}
}

func newTestOptsWithClientErr(t *testing.T, configYAML string, clientErr error) *rootOptions {
	t.Helper()
	path := writeTempConfig(t, configYAML)
	return &rootOptions{
		configPath:      path,
		credentialsPath: "fake-creds.json",
		format:          "text",
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return nil, clientErr
		},
	}
}

func TestRunPlan_Success(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{
			AccountID: "123456789",
			Users: []state.UserPermission{
				{Email: "alice@example.com", AccountAccess: "user"},
			},
		},
	}
	opts := newTestOpts(t, mock, validConfig)

	if err := runPlan(opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPlan_JSONFormat(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{AccountID: "123456789"},
	}
	opts := newTestOpts(t, mock, validConfig)
	opts.format = "json"

	if err := runPlan(opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPlan_ConfigLoadError(t *testing.T) {
	opts := &rootOptions{
		configPath:      "/nonexistent.yaml",
		credentialsPath: "fake.json",
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return &mockClient{}, nil
		},
	}

	err := runPlan(opts)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunPlan_ValidationError(t *testing.T) {
	mock := &mockClient{}
	opts := newTestOpts(t, mock, invalidConfig)

	err := runPlan(opts)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestRunPlan_MissingCredentials(t *testing.T) {
	path := writeTempConfig(t, validConfig)
	opts := &rootOptions{
		configPath:      path,
		credentialsPath: "",
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return &mockClient{}, nil
		},
	}

	err := runPlan(opts)
	if err == nil {
		t.Fatal("expected error for missing credentials, got nil")
	}
}

func TestRunPlan_ClientError(t *testing.T) {
	opts := newTestOptsWithClientErr(t, validConfig, errors.New("auth failed"))

	err := runPlan(opts)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunPlan_FetchStateError(t *testing.T) {
	mock := &mockClient{
		fetchErr: errors.New("API unavailable"),
	}
	opts := newTestOpts(t, mock, validConfig)

	err := runPlan(opts)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunPlan_NoChanges(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{
			AccountID: "123456789",
			Users: []state.UserPermission{
				{
					Email:         "alice@example.com",
					AccountAccess: "user",
					ContainerAccess: []state.ContainerPermission{
						{ContainerID: "GTM-AAAA1111", Permission: "publish"},
					},
				},
			},
		},
	}
	opts := newTestOpts(t, mock, validConfig)

	if err := runPlan(opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
