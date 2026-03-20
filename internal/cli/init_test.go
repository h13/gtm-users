package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/h13/gtm-users/internal/state"
)

func TestRunInit_HappyPath(t *testing.T) {
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

	dir := t.TempDir()
	configPath := filepath.Join(dir, "gtm-users.yaml")

	opts := &rootOptions{
		configPath:      configPath,
		credentialsPath: "fake-creds.json",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return mock, nil
		},
	}

	if err := runInit(opts, "123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading created file: %v", err)
	}

	got := string(data)
	if !strings.Contains(got, "account_id:") {
		t.Error("missing account_id in output")
	}
	if !strings.Contains(got, "alice@example.com") {
		t.Error("missing alice in output")
	}
	if !strings.Contains(got, "GTM-AAAA1111") {
		t.Error("missing container ID in output")
	}
}

func TestRunInit_FileAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "gtm-users.yaml")
	if err := os.WriteFile(configPath, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}

	opts := &rootOptions{
		configPath:      configPath,
		credentialsPath: "fake-creds.json",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return &mockClient{}, nil
		},
	}

	err := runInit(opts, "123")
	if err == nil {
		t.Fatal("expected error for existing file, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want 'already exists' message", err.Error())
	}
}

func TestRunInit_MissingCredentials(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "gtm-users.yaml")

	opts := &rootOptions{
		configPath:      configPath,
		credentialsPath: "",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return &mockClient{}, nil
		},
	}

	err := runInit(opts, "123")
	if err == nil {
		t.Fatal("expected error for missing credentials, got nil")
	}
	if !strings.Contains(err.Error(), "credentials") {
		t.Errorf("error = %q, want credentials message", err.Error())
	}
}

func TestRunInit_FetchStateError(t *testing.T) {
	mock := &mockClient{
		fetchErr: errors.New("API error"),
	}

	dir := t.TempDir()
	configPath := filepath.Join(dir, "gtm-users.yaml")

	opts := &rootOptions{
		configPath:      configPath,
		credentialsPath: "fake-creds.json",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return mock, nil
		},
	}

	err := runInit(opts, "123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunInit_ClientError(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "gtm-users.yaml")

	opts := &rootOptions{
		configPath:      configPath,
		credentialsPath: "fake-creds.json",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return nil, errors.New("auth failed")
		},
	}

	err := runInit(opts, "123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunInit_CreateFileError(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{AccountID: "123"},
	}

	opts := &rootOptions{
		configPath:      "/nonexistent-dir/subdir/gtm-users.yaml",
		credentialsPath: "fake-creds.json",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return mock, nil
		},
	}

	err := runInit(opts, "123")
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}

func TestNewInitCmd(t *testing.T) {
	opts := &rootOptions{}
	cmd := newInitCmd(opts)

	if cmd.Use != "init" {
		t.Errorf("use = %q, want %q", cmd.Use, "init")
	}

	f := cmd.Flags().Lookup("account-id")
	if f == nil {
		t.Error("missing account-id flag")
	}
}
