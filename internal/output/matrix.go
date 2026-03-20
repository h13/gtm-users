package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// MatrixEntry represents a cell in the permission matrix.
type MatrixEntry struct {
	Email       string
	Permissions map[string]string // containerID -> permission
}

// PrintMatrix prints a user x container permission matrix.
func PrintMatrix(w io.Writer, entries []MatrixEntry, containerIDs []string) error {
	if len(entries) == 0 {
		_, err := fmt.Fprintln(w, "No users found.")
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Header row.
	header := []string{"USER"}
	header = append(header, containerIDs...)
	if _, err := fmt.Fprintln(tw, strings.Join(header, "\t")); err != nil {
		return err
	}

	// Data rows.
	for _, e := range entries {
		row := []string{e.Email}
		for _, cid := range containerIDs {
			perm, ok := e.Permissions[cid]
			if !ok {
				perm = "-"
			}
			row = append(row, perm)
		}
		if _, err := fmt.Fprintln(tw, strings.Join(row, "\t")); err != nil {
			return err
		}
	}

	return tw.Flush()
}
