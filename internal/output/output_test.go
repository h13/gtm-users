package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/h13/gtm-users/internal/config"
	"github.com/h13/gtm-users/internal/diff"
	"github.com/h13/gtm-users/internal/output"
)

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

	// Check summary line
	if !strings.Contains(got, "1 to add") {
		t.Errorf("missing '1 to add' in output")
	}
	if !strings.Contains(got, "1 to change") {
		t.Errorf("missing '1 to change' in output")
	}
	if !strings.Contains(got, "1 to destroy") {
		t.Errorf("missing '1 to destroy' in output")
	}

	// Check user entries
	if !strings.Contains(got, "+ user alice@example.com") {
		t.Errorf("missing add for alice")
	}
	if !strings.Contains(got, "~ user bob@example.com") {
		t.Errorf("missing update for bob")
	}
	if !strings.Contains(got, "- user carol@example.com") {
		t.Errorf("missing delete for carol")
	}

	// Check container change
	if !strings.Contains(got, "+ container GTM-AAAA1111: publish") {
		t.Errorf("missing container add")
	}

	// Check account access change
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

	if result["AccountID"] != "123" {
		t.Errorf("AccountID = %v, want 123", result["AccountID"])
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
