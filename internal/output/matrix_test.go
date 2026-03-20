package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/h13/gtm-users/internal/output"
)

func TestPrintMatrix_Empty(t *testing.T) {
	var buf bytes.Buffer
	err := output.PrintMatrix(&buf, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "No users found.\n" {
		t.Errorf("output = %q, want %q", got, "No users found.\n")
	}
}

func TestPrintMatrix_SingleUser(t *testing.T) {
	var buf bytes.Buffer
	entries := []output.MatrixEntry{
		{
			Email: "alice@example.com",
			Permissions: map[string]string{
				"GTM-AAAA1111": "publish",
			},
		},
	}
	containerIDs := []string{"GTM-AAAA1111"}

	err := output.PrintMatrix(&buf, entries, containerIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "USER") {
		t.Error("expected header row with USER")
	}
	if !strings.Contains(out, "alice@example.com") {
		t.Error("expected alice in output")
	}
	if !strings.Contains(out, "publish") {
		t.Error("expected publish permission in output")
	}
}

func TestPrintMatrix_MultipleUsersAndContainers(t *testing.T) {
	var buf bytes.Buffer
	entries := []output.MatrixEntry{
		{
			Email: "alice@example.com",
			Permissions: map[string]string{
				"GTM-AAAA1111": "publish",
				"GTM-BBBB2222": "read",
			},
		},
		{
			Email: "bob@example.com",
			Permissions: map[string]string{
				"GTM-AAAA1111": "edit",
			},
		},
	}
	containerIDs := []string{"GTM-AAAA1111", "GTM-BBBB2222"}

	err := output.PrintMatrix(&buf, entries, containerIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "GTM-AAAA1111") {
		t.Error("expected GTM-AAAA1111 in header")
	}
	if !strings.Contains(out, "GTM-BBBB2222") {
		t.Error("expected GTM-BBBB2222 in header")
	}
	if !strings.Contains(out, "alice@example.com") {
		t.Error("expected alice in output")
	}
	if !strings.Contains(out, "bob@example.com") {
		t.Error("expected bob in output")
	}
}

func TestPrintMatrix_MissingPermission(t *testing.T) {
	var buf bytes.Buffer
	entries := []output.MatrixEntry{
		{
			Email: "alice@example.com",
			Permissions: map[string]string{
				"GTM-AAAA1111": "publish",
			},
		},
	}
	containerIDs := []string{"GTM-AAAA1111", "GTM-BBBB2222"}

	err := output.PrintMatrix(&buf, entries, containerIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}

	// The second line (alice's row) should contain a dash for the missing container.
	if !strings.Contains(lines[1], "-") {
		t.Error("expected dash for missing permission")
	}
}
