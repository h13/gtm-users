package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/h13/gtm-users/internal/state"
)

func TestRunExport_Success(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{
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
		},
	}
	opts := &rootOptions{
		credentialsPath: "fake-creds.json",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return mock, nil
		},
	}

	if err := runExport(opts, "123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunExport_NoContainers(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{
			AccountID: "123",
			Users: []state.UserPermission{
				{Email: "bob@example.com", AccountAccess: "admin"},
			},
		},
	}
	opts := &rootOptions{
		credentialsPath: "fake-creds.json",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return mock, nil
		},
	}

	if err := runExport(opts, "123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunExport_MissingCredentials(t *testing.T) {
	opts := &rootOptions{
		credentialsPath: "",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return &mockClient{}, nil
		},
	}

	err := runExport(opts, "123")
	if err == nil {
		t.Fatal("expected error for missing credentials, got nil")
	}
}

func TestRunExport_ClientError(t *testing.T) {
	opts := &rootOptions{
		credentialsPath: "fake-creds.json",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return nil, errors.New("auth failed")
		},
	}

	err := runExport(opts, "123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunExport_FetchStateError(t *testing.T) {
	mock := &mockClient{
		fetchErr: errors.New("API error"),
	}
	opts := &rootOptions{
		credentialsPath: "fake-creds.json",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return mock, nil
		},
	}

	err := runExport(opts, "123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
