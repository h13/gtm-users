package cli

import (
	"testing"
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

	for _, name := range []string{"validate", "plan", "apply", "export"} {
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
