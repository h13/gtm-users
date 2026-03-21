package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/h13/gtm-users/internal/state"
)

func TestRunDrift_NoDrift(t *testing.T) {
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

	var buf bytes.Buffer
	opts.stdout = &buf

	if err := runDrift(opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "No drift detected") {
		t.Errorf("output = %q, want 'No drift detected' message", buf.String())
	}
}

func TestRunDrift_DriftDetected(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{
			AccountID: "123456789",
		},
	}
	opts := newTestOpts(t, mock, validConfig)

	err := runDrift(opts)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var driftErr DriftDetectedError
	if !errors.As(err, &driftErr) {
		t.Errorf("error type = %T, want DriftDetectedError", err)
	}
}

func TestRunDrift_ConfigError(t *testing.T) {
	opts := &rootOptions{
		configPath:      "/nonexistent.yaml",
		credentialsPath: "fake.json",
		stdout:          &bytes.Buffer{},
		newClient: func(_ context.Context, _, _ string) (gtmClient, error) {
			return &mockClient{}, nil
		},
	}

	err := runDrift(opts)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunDrift_ValidationError(t *testing.T) {
	mock := &mockClient{}
	opts := newTestOpts(t, mock, invalidConfig)

	err := runDrift(opts)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestRunDrift_ClientError(t *testing.T) {
	opts := newTestOptsWithClientErr(t, validConfig, errors.New("auth failed"))

	err := runDrift(opts)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunDrift_FetchError(t *testing.T) {
	mock := &mockClient{
		fetchErr: errors.New("API unavailable"),
	}
	opts := newTestOpts(t, mock, validConfig)

	err := runDrift(opts)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRunDrift_PrintPlanError(t *testing.T) {
	mock := &mockClient{
		fetchState: state.AccountState{AccountID: "123456789"},
	}
	opts := newTestOpts(t, mock, validConfig)
	opts.stdout = &errWriter{err: errors.New("write error")}

	err := runDrift(opts)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDriftDetectedError_Error(t *testing.T) {
	err := DriftDetectedError{}
	if err.Error() != "drift detected" {
		t.Errorf("Error() = %q, want %q", err.Error(), "drift detected")
	}
}

func TestNewDriftCmd(t *testing.T) {
	opts := &rootOptions{}
	cmd := newDriftCmd(opts)

	if cmd.Use != "drift" {
		t.Errorf("use = %q, want %q", cmd.Use, "drift")
	}
}
