package output_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/h13/gtm-users/internal/config"
	"github.com/h13/gtm-users/internal/diff"
	"github.com/h13/gtm-users/internal/output"
)

type errWriter struct{ err error }

func (w *errWriter) Write([]byte) (int, error) { return 0, w.err }

// callCountWriter fails after n successful Write calls.
type callCountWriter struct {
	remaining int
}

func (w *callCountWriter) Write(p []byte) (int, error) {
	if w.remaining <= 0 {
		return 0, errors.New("write limit exceeded")
	}
	w.remaining--
	return len(p), nil
}

func TestPrintPlan_NoChanges(t *testing.T) {
	plan := diff.Plan{AccountID: "123", Mode: config.ModeAdditive}

	var buf bytes.Buffer
	if err := output.PrintPlan(&buf, plan, output.FormatText); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "No changes") {
		t.Errorf("output = %q, want 'No changes' message", got)
	}
}

func TestPrintPlan_TextFormat(t *testing.T) {
	plan := diff.Plan{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Changes: []diff.UserChange{
			{
				Email:            "alice@example.com",
				Action:           diff.ActionAdd,
				NewAccountAccess: "user",
				ContainerChanges: []diff.ContainerChange{
					{ContainerID: "GTM-AAAA1111", Action: diff.ActionAdd, NewPermission: "publish"},
				},
			},
			{
				Email:            "bob@example.com",
				Action:           diff.ActionUpdate,
				OldAccountAccess: "user",
				NewAccountAccess: "admin",
			},
			{
				Email:            "carol@example.com",
				Action:           diff.ActionDelete,
				OldAccountAccess: "user",
			},
		},
	}

	var buf bytes.Buffer
	if err := output.PrintPlan(&buf, plan, output.FormatText); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()

	if !strings.Contains(got, "1 to add") {
		t.Errorf("missing '1 to add' in output")
	}
	if !strings.Contains(got, "1 to change") {
		t.Errorf("missing '1 to change' in output")
	}
	if !strings.Contains(got, "1 to destroy") {
		t.Errorf("missing '1 to destroy' in output")
	}

	if !strings.Contains(got, "+ user alice@example.com") {
		t.Errorf("missing add for alice")
	}
	if !strings.Contains(got, "~ user bob@example.com") {
		t.Errorf("missing update for bob")
	}
	if !strings.Contains(got, "- user carol@example.com") {
		t.Errorf("missing delete for carol")
	}

	if !strings.Contains(got, "+ container GTM-AAAA1111: publish") {
		t.Errorf("missing container add")
	}

	if !strings.Contains(got, "user → admin") {
		t.Errorf("missing account access change arrow")
	}
}

func TestPrintPlan_JSONFormat(t *testing.T) {
	plan := diff.Plan{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Changes: []diff.UserChange{
			{
				Email:            "alice@example.com",
				Action:           diff.ActionAdd,
				NewAccountAccess: "user",
			},
		},
	}

	var buf bytes.Buffer
	if err := output.PrintPlan(&buf, plan, output.FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if result["account_id"] != "123" {
		t.Errorf("account_id = %v, want 123", result["account_id"])
	}
}

func TestPrintExport(t *testing.T) {
	users := []output.ExportUser{
		{
			Email:         "alice@example.com",
			AccountAccess: "user",
			ContainerAccess: []output.ExportContainerAccess{
				{ContainerID: "GTM-AAAA1111", Permission: "publish"},
			},
		},
		{
			Email:         "bob@example.com",
			AccountAccess: "admin",
		},
	}

	var buf bytes.Buffer
	if err := output.PrintExport(&buf, "123456789", users); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "account_id:") {
		t.Error("missing account_id")
	}
	if !strings.Contains(got, "alice@example.com") {
		t.Error("missing alice")
	}
	if !strings.Contains(got, "bob@example.com") {
		t.Error("missing bob")
	}
	if !strings.Contains(got, "GTM-AAAA1111") {
		t.Error("missing container ID")
	}
	if !strings.Contains(got, "mode: additive") {
		t.Error("missing mode")
	}
}

func TestPrintPlan_DefaultFormat(t *testing.T) {
	plan := diff.Plan{AccountID: "123", Mode: config.ModeAdditive}

	var buf bytes.Buffer
	if err := output.PrintPlan(&buf, plan, "unknown"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "No changes") {
		t.Error("unknown format should fall back to text")
	}
}

func TestPrintExport_NoContainers(t *testing.T) {
	users := []output.ExportUser{
		{Email: "alice@example.com", AccountAccess: "admin"},
	}

	var buf bytes.Buffer
	if err := output.PrintExport(&buf, "123", users); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if strings.Contains(got, "container_access") {
		t.Error("should not have container_access for user without containers")
	}
}

func TestPrintPlan_ContainerUpdateAndDelete(t *testing.T) {
	plan := diff.Plan{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Changes: []diff.UserChange{
			{
				Email:            "alice@example.com",
				Action:           diff.ActionUpdate,
				OldAccountAccess: "user",
				NewAccountAccess: "user",
				ContainerChanges: []diff.ContainerChange{
					{ContainerID: "GTM-AAAA1111", Action: diff.ActionUpdate, OldPermission: "read", NewPermission: "publish"},
					{ContainerID: "GTM-BBBB2222", Action: diff.ActionDelete, OldPermission: "edit"},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := output.PrintPlan(&buf, plan, output.FormatText); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "~ container GTM-AAAA1111: read → publish") {
		t.Errorf("missing container update: %s", got)
	}
	if !strings.Contains(got, "- container GTM-BBBB2222: edit") {
		t.Errorf("missing container delete: %s", got)
	}
}

func TestPrintPlan_UnknownContainerAction(t *testing.T) {
	plan := diff.Plan{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Changes: []diff.UserChange{
			{
				Email:            "alice@example.com",
				Action:           diff.ActionUpdate,
				OldAccountAccess: "user",
				NewAccountAccess: "user",
				ContainerChanges: []diff.ContainerChange{
					{ContainerID: "GTM-AAAA1111", Action: "unknown"},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := output.PrintPlan(&buf, plan, output.FormatText); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "? container GTM-AAAA1111") {
		t.Errorf("missing unknown container action: %s", got)
	}
}

func TestPrintPlan_UnknownUserAction(t *testing.T) {
	plan := diff.Plan{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Changes: []diff.UserChange{
			{
				Email:  "alice@example.com",
				Action: "unknown",
			},
		},
	}

	var buf bytes.Buffer
	if err := output.PrintPlan(&buf, plan, output.FormatText); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Unknown action produces no lines for the user, but summary is still printed
	got := buf.String()
	if !strings.Contains(got, "Plan:") {
		t.Errorf("missing plan summary: %s", got)
	}
}

func TestPrintPlan_UpdateNoAccountAccessChange(t *testing.T) {
	plan := diff.Plan{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Changes: []diff.UserChange{
			{
				Email:            "alice@example.com",
				Action:           diff.ActionUpdate,
				OldAccountAccess: "user",
				NewAccountAccess: "user",
				ContainerChanges: []diff.ContainerChange{
					{ContainerID: "GTM-AAAA1111", Action: diff.ActionAdd, NewPermission: "read"},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := output.PrintPlan(&buf, plan, output.FormatText); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if strings.Contains(got, "→") {
		t.Errorf("should not show arrow when account access unchanged: %s", got)
	}
}

func TestPrintPlan_TextWriteError(t *testing.T) {
	plan := diff.Plan{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Changes: []diff.UserChange{
			{Email: "alice@example.com", Action: diff.ActionAdd, NewAccountAccess: "user"},
		},
	}

	w := &errWriter{err: errors.New("write error")}
	err := output.PrintPlan(w, plan, output.FormatText)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPrintPlan_NoChangesWriteError(t *testing.T) {
	plan := diff.Plan{AccountID: "123", Mode: config.ModeAdditive}

	w := &errWriter{err: errors.New("write error")}
	err := output.PrintPlan(w, plan, output.FormatText)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPrintPlan_JSONWriteError(t *testing.T) {
	plan := diff.Plan{AccountID: "123", Mode: config.ModeAdditive}

	w := &errWriter{err: errors.New("write error")}
	err := output.PrintPlan(w, plan, output.FormatJSON)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPrintExport_WriteError(t *testing.T) {
	users := []output.ExportUser{
		{Email: "alice@example.com", AccountAccess: "admin"},
	}

	w := &errWriter{err: errors.New("write error")}
	err := output.PrintExport(w, "123", users)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPrintExport_EmptyUsers(t *testing.T) {
	var buf bytes.Buffer
	if err := output.PrintExport(&buf, "123", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "account_id:") {
		t.Error("missing account_id header")
	}
	if !strings.Contains(got, "users:") {
		t.Error("missing users header")
	}
}

func TestPrintText_SummaryWriteError(t *testing.T) {
	plan := diff.Plan{
		AccountID: "123",
		Mode:      config.ModeAdditive,
		Changes: []diff.UserChange{
			{Email: "alice@example.com", Action: diff.ActionAdd, NewAccountAccess: "user"},
		},
	}

	// Allow summary line (call 1), fail on user change (call 2)
	w := &callCountWriter{remaining: 1}
	err := output.PrintPlan(w, plan, output.FormatText)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPrintExport_UserWriteError(t *testing.T) {
	users := []output.ExportUser{
		{
			Email:         "alice@example.com",
			AccountAccess: "user",
			ContainerAccess: []output.ExportContainerAccess{
				{ContainerID: "GTM-AAAA1111", Permission: "publish"},
			},
		},
	}

	// Allow header (call 1), fail on user email line (call 2)
	w := &callCountWriter{remaining: 1}
	err := output.PrintExport(w, "123", users)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteExportUser_ContainerHeaderError(t *testing.T) {
	users := []output.ExportUser{
		{
			Email:         "alice@example.com",
			AccountAccess: "user",
			ContainerAccess: []output.ExportContainerAccess{
				{ContainerID: "GTM-AAAA1111", Permission: "publish"},
			},
		},
	}

	// Allow header (call 1) + email line (call 2), fail on container_access header (call 3)
	w := &callCountWriter{remaining: 2}
	err := output.PrintExport(w, "123", users)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteExportUser_ContainerDetailError(t *testing.T) {
	users := []output.ExportUser{
		{
			Email:         "alice@example.com",
			AccountAccess: "user",
			ContainerAccess: []output.ExportContainerAccess{
				{ContainerID: "GTM-AAAA1111", Permission: "publish"},
			},
		},
	}

	// Allow header (1) + email (2) + container_access header (3), fail on container detail (4)
	w := &callCountWriter{remaining: 3}
	err := output.PrintExport(w, "123", users)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
