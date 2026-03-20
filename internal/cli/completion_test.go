package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func newTestCompletionCmd() (*cobra.Command, *bytes.Buffer) {
	root := &cobra.Command{
		Use: "gtm-users",
	}
	root.AddCommand(newCompletionCmd())

	var buf bytes.Buffer
	root.SetOut(&buf)

	return root, &buf
}

func TestCompletionCmd_Bash(t *testing.T) {
	root, buf := newTestCompletionCmd()
	root.SetArgs([]string{"completion", "bash"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty bash completion output")
	}
}

func TestCompletionCmd_Zsh(t *testing.T) {
	root, buf := newTestCompletionCmd()
	root.SetArgs([]string{"completion", "zsh"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty zsh completion output")
	}
}

func TestCompletionCmd_Fish(t *testing.T) {
	root, buf := newTestCompletionCmd()
	root.SetArgs([]string{"completion", "fish"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty fish completion output")
	}
}

func TestCompletionCmd_Powershell(t *testing.T) {
	root, buf := newTestCompletionCmd()
	root.SetArgs([]string{"completion", "powershell"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty powershell completion output")
	}
}

func TestCompletionCmd_InvalidShell(t *testing.T) {
	root, _ := newTestCompletionCmd()
	root.SetArgs([]string{"completion", "invalid"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid shell, got nil")
	}
}

func TestCompletionCmd_NoArgs(t *testing.T) {
	root, _ := newTestCompletionCmd()
	root.SetArgs([]string{"completion"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for no args, got nil")
	}
}

func TestCompletionCmd_Structure(t *testing.T) {
	cmd := newCompletionCmd()

	if cmd.Use != "completion [bash|zsh|fish|powershell]" {
		t.Errorf("use = %q, want %q", cmd.Use, "completion [bash|zsh|fish|powershell]")
	}

	if len(cmd.ValidArgs) != 4 {
		t.Errorf("valid args = %d, want 4", len(cmd.ValidArgs))
	}

	if !cmd.DisableFlagsInUseLine {
		t.Error("expected DisableFlagsInUseLine to be true")
	}
}
