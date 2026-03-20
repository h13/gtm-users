package cli

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/h13/gtm-users/internal/gtm"
	"google.golang.org/api/option"
)

func TestNewRootCmd_Structure(t *testing.T) {
	cmd := NewRootCmd("1.0.0")

	if cmd.Version != "1.0.0" {
		t.Errorf("version = %q, want %q", cmd.Version, "1.0.0")
	}

	if cmd.Use != "gtm-users" {
		t.Errorf("use = %q, want %q", cmd.Use, "gtm-users")
	}

	subcommands := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subcommands[sub.Name()] = true
	}

	for _, name := range []string{"validate", "plan", "apply", "export", "init", "drift"} {
		if !subcommands[name] {
			t.Errorf("missing subcommand %q", name)
		}
	}
}

func TestNewRootCmd_PersistentFlags(t *testing.T) {
	cmd := NewRootCmd("1.0.0")

	tests := []struct {
		name     string
		defValue string
	}{
		{"config", "gtm-users.yaml"},
		{"credentials", ""},
		{"format", "text"},
		{"no-color", "false"},
	}

	for _, tt := range tests {
		f := cmd.PersistentFlags().Lookup(tt.name)
		if f == nil {
			t.Errorf("missing persistent flag %q", tt.name)
			continue
		}
		if f.DefValue != tt.defValue {
			t.Errorf("flag %q default = %q, want %q", tt.name, f.DefValue, tt.defValue)
		}
	}
}

func TestNewRootCmd_WithGTMOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"userPermission":[]}`)) //nolint:errcheck
	}))
	t.Cleanup(srv.Close)

	cmd := NewRootCmd("1.0.0", WithGTMOptions(
		gtm.WithAPIOptions(
			option.WithHTTPClient(srv.Client()),
			option.WithEndpoint(srv.URL),
		),
	))

	path := writeTempConfig(t, validConfig)
	cmd.SetArgs([]string{"plan", "--config", path, "--credentials", "fake.json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewRootCmd_DefaultClientFactory(t *testing.T) {
	cmd := NewRootCmd("1.0.0")

	path := writeTempConfig(t, validConfig)
	cmd.SetArgs([]string{"plan", "--config", path, "--credentials", "/nonexistent/creds.json"})

	// Default factory calls gtm.NewClient which fails with invalid credentials.
	// This exercises the default newClient closure.
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid credentials, got nil")
	}
}
