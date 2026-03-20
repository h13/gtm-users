package output

import "github.com/h13/gtm-users/internal/diff"

// ANSI color codes for terminal output.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// ColorForAction returns the ANSI color code for a given action type.
func ColorForAction(action diff.ActionType) string {
	switch action {
	case diff.ActionAdd:
		return colorGreen
	case diff.ActionUpdate:
		return colorYellow
	case diff.ActionDelete:
		return colorRed
	default:
		return ""
	}
}

// Colorize wraps a string with the given color code and reset.
func Colorize(color, s string) string {
	if color == "" {
		return s
	}
	return color + s + colorReset
}
