package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/h13/gtm-users/internal/diff"
)

// Format represents an output format.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// PrintPlan writes the plan to the writer in the specified format.
func PrintPlan(w io.Writer, plan diff.Plan, format Format) error {
	switch format {
	case FormatJSON:
		return printJSON(w, plan)
	case FormatText:
		return printText(w, plan)
	default:
		return printText(w, plan)
	}
}

func printText(w io.Writer, plan diff.Plan) error {
	if !plan.HasChanges() {
		_, err := fmt.Fprintln(w, "No changes. Infrastructure is up-to-date.")
		return err
	}

	adds, updates, deletes := plan.Summary()
	if _, err := fmt.Fprintf(w, "Plan: %d to add, %d to change, %d to destroy.\n\n", adds, updates, deletes); err != nil {
		return err
	}

	for _, change := range plan.Changes {
		if err := printUserChange(w, change); err != nil {
			return err
		}
	}

	return nil
}

func printUserChange(w io.Writer, change diff.UserChange) error {
	symbol := actionSymbol(change.Action)
	var lines []string

	switch change.Action {
	case diff.ActionAdd:
		lines = append(lines, fmt.Sprintf("%s user %s", symbol, change.Email))
		lines = append(lines, fmt.Sprintf("    account_access: %s", change.NewAccountAccess))
		for _, cc := range change.ContainerChanges {
			lines = append(lines, fmt.Sprintf("    + container %s: %s", cc.ContainerID, cc.NewPermission))
		}

	case diff.ActionUpdate:
		lines = append(lines, fmt.Sprintf("%s user %s", symbol, change.Email))
		if change.OldAccountAccess != change.NewAccountAccess {
			lines = append(lines, fmt.Sprintf("    account_access: %s → %s", change.OldAccountAccess, change.NewAccountAccess))
		}
		for _, cc := range change.ContainerChanges {
			lines = append(lines, formatContainerChange(cc))
		}

	case diff.ActionDelete:
		lines = append(lines, fmt.Sprintf("%s user %s", symbol, change.Email))
		lines = append(lines, fmt.Sprintf("    account_access: %s", change.OldAccountAccess))
	}

	if _, err := fmt.Fprintln(w, strings.Join(lines, "\n")); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w)
	return err
}

func formatContainerChange(cc diff.ContainerChange) string {
	switch cc.Action {
	case diff.ActionAdd:
		return fmt.Sprintf("    + container %s: %s", cc.ContainerID, cc.NewPermission)
	case diff.ActionUpdate:
		return fmt.Sprintf("    ~ container %s: %s → %s", cc.ContainerID, cc.OldPermission, cc.NewPermission)
	case diff.ActionDelete:
		return fmt.Sprintf("    - container %s: %s", cc.ContainerID, cc.OldPermission)
	default:
		return fmt.Sprintf("    ? container %s", cc.ContainerID)
	}
}

func actionSymbol(action diff.ActionType) string {
	switch action {
	case diff.ActionAdd:
		return "+"
	case diff.ActionUpdate:
		return "~"
	case diff.ActionDelete:
		return "-"
	default:
		return "?"
	}
}

func printJSON(w io.Writer, plan diff.Plan) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(plan)
}

// ExportUser represents a user for YAML export.
type ExportUser struct {
	Email           string
	AccountAccess   string
	ContainerAccess []ExportContainerAccess
}

// ExportContainerAccess represents a container permission for export.
type ExportContainerAccess struct {
	ContainerID string
	Permission  string
}

// PrintExport writes the account state as YAML to the writer.
func PrintExport(w io.Writer, accountID string, users []ExportUser) error {
	if _, err := fmt.Fprintf(w, "account_id: %q\nmode: additive\n\nusers:\n", accountID); err != nil {
		return err
	}

	for _, u := range users {
		if err := writeExportUser(w, u); err != nil {
			return err
		}
	}

	return nil
}

func writeExportUser(w io.Writer, u ExportUser) error {
	if _, err := fmt.Fprintf(w, "  - email: %s\n    account_access: %s\n", u.Email, u.AccountAccess); err != nil {
		return err
	}

	if len(u.ContainerAccess) == 0 {
		return nil
	}

	if _, err := fmt.Fprintln(w, "    container_access:"); err != nil {
		return err
	}
	for _, ca := range u.ContainerAccess {
		if _, err := fmt.Fprintf(w, "      - container_id: %q\n        permission: %s\n", ca.ContainerID, ca.Permission); err != nil {
			return err
		}
	}

	return nil
}
