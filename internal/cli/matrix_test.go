package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/h13/gtm-users/internal/state"
)

func TestRunMatrix_Success(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{
			AccountID: "123",
			Users: []state.UserPermission{
				{
					Email:         "alice@example.com",
					AccountAccess: "user",
					ContainerAccess: []state.ContainerPermission{
						{ContainerID: "GTM-AAAA1111", Permission: "publish"},
						{ContainerID: "GTM-BBBB2222", Permission: "read"},
					},
				},
				{
					Email:         "bob@example.com",
					AccountAccess: "admin",
					ContainerAccess: []state.ContainerPermission{
						{ContainerID: "GTM-AAAA1111", Permission: "edit"},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	opts := &rootOptions{
		credentialsPath: "fake-creds.json",
		stdout:          &buf,
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return mock, nil
		},
	}

	if err := runMatrix(opts, "123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "alice@example.com") {
		t.Error("expected alice in output")
	}
	if !strings.Contains(out, "bob@example.com") {
		t.Error("expected bob in output")
	}
	if !strings.Contains(out, "GTM-AAAA1111") {
		t.Error("expected GTM-AAAA1111 in output")
	}
	if !strings.Contains(out, "GTM-BBBB2222") {
		t.Error("expected GTM-BBBB2222 in output")
	}
}

func TestRunMatrix_EmptyState(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{AccountID: "123"},
	}

	var buf bytes.Buffer
	opts := &rootOptions{
		credentialsPath: "fake-creds.json",
		stdout:          &buf,
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return mock, nil
		},
	}

	if err := runMatrix(opts, "123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "No users found.") {
		t.Error("expected 'No users found.' in output")
	}
}

func TestRunMatrix_MissingCredentials(t *testing.T) {
	opts := &rootOptions{
		credentialsPath: "",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return &mockClient{}, nil
		},
	}

	err := runMatrix(opts, "123")
	if err == nil {
		t.Fatal("expected error for missing credentials, got nil")
	}
	if !strings.Contains(err.Error(), "--credentials") {
		t.Errorf("error = %q, want mention of --credentials", err.Error())
	}
}

func TestRunMatrix_ClientError(t *testing.T) {
	opts := &rootOptions{
		credentialsPath: "fake-creds.json",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return nil, errors.New("auth failed")
		},
	}

	err := runMatrix(opts, "123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunMatrix_FetchError(t *testing.T) {
	mock := &mockClient{
		fetchErr: errors.New("API unavailable"),
	}

	opts := &rootOptions{
		credentialsPath: "fake-creds.json",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return mock, nil
		},
	}

	err := runMatrix(opts, "123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewMatrixCmd_Structure(t *testing.T) {
	opts := &rootOptions{}
	cmd := newMatrixCmd(opts)

	if cmd.Use != "matrix" {
		t.Errorf("use = %q, want %q", cmd.Use, "matrix")
	}

	f := cmd.Flags().Lookup("account-id")
	if f == nil {
		t.Error("missing account-id flag")
	}
}
