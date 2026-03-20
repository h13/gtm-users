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

func TestRunBackup_Success(t *testing.T) {
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

	outPath := filepath.Join(t.TempDir(), "test-backup.yaml")
	var buf bytes.Buffer
	opts := &rootOptions{
		credentialsPath: "fake-creds.json",
		stdout:          &buf,
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return mock, nil
		},
	}

	if err := runBackup(opts, "123", outPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created with content.
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading backup file: %v", err)
	}
	if len(data) == 0 {
		t.Error("backup file is empty")
	}
	if !strings.Contains(string(data), "alice@example.com") {
		t.Error("backup should contain alice@example.com")
	}
	if !strings.Contains(string(data), "123") {
		t.Error("backup should contain account ID")
	}

	// Verify stdout message.
	if !strings.Contains(buf.String(), "Backup saved to") {
		t.Errorf("stdout = %q, want 'Backup saved to' message", buf.String())
	}
}

func TestRunBackup_DefaultOutputPath(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{
			AccountID: "123",
			Users: []state.UserPermission{
				{Email: "alice@example.com", AccountAccess: "user"},
			},
		},
	}

	// Change to temp dir so the default backup file is written there.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting working dir: %v", err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("changing dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	var buf bytes.Buffer
	opts := &rootOptions{
		credentialsPath: "fake-creds.json",
		stdout:          &buf,
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return mock, nil
		},
	}

	if err := runBackup(opts, "123", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the output message mentions backup-.
	if !strings.Contains(buf.String(), "backup-") {
		t.Errorf("stdout = %q, want timestamp-based filename", buf.String())
	}
}

func TestRunBackup_CustomOutputPath(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{
			AccountID: "123",
			Users:     []state.UserPermission{},
		},
	}

	outPath := filepath.Join(t.TempDir(), "custom-name.yaml")
	var buf bytes.Buffer
	opts := &rootOptions{
		credentialsPath: "fake-creds.json",
		stdout:          &buf,
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return mock, nil
		},
	}

	if err := runBackup(opts, "123", outPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("expected file at %s, got error: %v", outPath, err)
	}
}

func TestRunBackup_MissingCredentials(t *testing.T) {
	opts := &rootOptions{
		credentialsPath: "",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return &mockClient{}, nil
		},
	}

	err := runBackup(opts, "123", "out.yaml")
	if err == nil {
		t.Fatal("expected error for missing credentials, got nil")
	}
	if !strings.Contains(err.Error(), "--credentials") {
		t.Errorf("error = %q, want mention of --credentials", err.Error())
	}
}

func TestRunBackup_ClientError(t *testing.T) {
	opts := &rootOptions{
		credentialsPath: "fake-creds.json",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return nil, errors.New("auth failed")
		},
	}

	err := runBackup(opts, "123", filepath.Join(t.TempDir(), "out.yaml"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunBackup_FetchError(t *testing.T) {
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

	err := runBackup(opts, "123", filepath.Join(t.TempDir(), "out.yaml"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunBackup_CreateFileError(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{AccountID: "123"},
	}
	opts := &rootOptions{
		credentialsPath: "fake-creds.json",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return mock, nil
		},
	}

	err := runBackup(opts, "123", "/nonexistent-dir/backup.yaml")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewBackupCmd_Structure(t *testing.T) {
	opts := &rootOptions{}
	cmd := newBackupCmd(opts)

	if cmd.Use != "backup" {
		t.Errorf("use = %q, want %q", cmd.Use, "backup")
	}

	f := cmd.Flags().Lookup("account-id")
	if f == nil {
		t.Error("missing account-id flag")
	}

	f = cmd.Flags().Lookup("output")
	if f == nil {
		t.Fatal("missing output flag")
	}
	if f.Shorthand != "o" {
		t.Errorf("output shorthand = %q, want 'o'", f.Shorthand)
	}
}
