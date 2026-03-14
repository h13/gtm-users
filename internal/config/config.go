package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Mode determines how unmanaged users are handled.
type Mode string

const (
	ModeAdditive      Mode = "additive"
	ModeAuthoritative Mode = "authoritative"
)

// AccountAccess represents the GTM account-level permission.
type AccountAccess string

const (
	AccountAccessAdmin AccountAccess = "admin"
	AccountAccessUser  AccountAccess = "user"
	AccountAccessNone  AccountAccess = "noAccess"
)

// ContainerPermission represents a permission level for a GTM container.
type ContainerPermission string

const (
	PermissionRead    ContainerPermission = "read"
	PermissionEdit    ContainerPermission = "edit"
	PermissionApprove ContainerPermission = "approve"
	PermissionPublish ContainerPermission = "publish"
)

// ContainerAccess pairs a container ID with a permission level.
type ContainerAccess struct {
	ContainerID string              `yaml:"container_id"`
	Permission  ContainerPermission `yaml:"permission"`
}

// User represents a GTM user's desired permission state.
type User struct {
	Email           string            `yaml:"email"`
	AccountAccess   AccountAccess     `yaml:"account_access"`
	ContainerAccess []ContainerAccess `yaml:"container_access,omitempty"`
}

// Config is the top-level YAML configuration.
type Config struct {
	AccountID string `yaml:"account_id"`
	Mode      Mode   `yaml:"mode"`
	Users     []User `yaml:"users"`
}

// Load reads and parses a YAML config file.
func Load(path string) (Config, error) {
	cleaned := filepath.Clean(path)
	data, err := os.ReadFile(cleaned)
	if err != nil {
		return Config{}, fmt.Errorf("reading config file: %w", err)
	}

	return Parse(data)
}

// Parse parses YAML bytes into a Config.
// Unknown fields in the YAML are rejected to catch typos early.
func Parse(data []byte) (Config, error) {
	var cfg Config

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)

	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("parsing YAML: %w", err)
	}

	if cfg.Mode == "" {
		cfg.Mode = ModeAdditive
	}

	return cfg, nil
}

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	containerIDRe = regexp.MustCompile(`^GTM-[A-Z0-9]+$`)

	validAccountAccess = map[AccountAccess]bool{
		AccountAccessAdmin: true,
		AccountAccessUser:  true,
		AccountAccessNone:  true,
	}

	validContainerPermissions = map[ContainerPermission]bool{
		PermissionRead:    true,
		PermissionEdit:    true,
		PermissionApprove: true,
		PermissionPublish: true,
	}
)

// ValidationError represents a single validation failure.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate checks the config for structural and semantic errors.
func Validate(cfg Config) []ValidationError {
	var errs []ValidationError

	errs = validateTopLevel(errs, cfg)

	seen := make(map[string]bool)
	for i, u := range cfg.Users {
		errs = validateUser(errs, u, i, seen)
	}

	return errs
}

func validateTopLevel(errs []ValidationError, cfg Config) []ValidationError {
	if cfg.AccountID == "" {
		errs = append(errs, ValidationError{Field: "account_id", Message: "required"})
	}

	if cfg.Mode != ModeAdditive && cfg.Mode != ModeAuthoritative {
		errs = append(errs, ValidationError{
			Field:   "mode",
			Message: fmt.Sprintf("must be 'additive' or 'authoritative', got '%s'", cfg.Mode),
		})
	}

	if len(cfg.Users) == 0 {
		errs = append(errs, ValidationError{Field: "users", Message: "at least one user required"})
	}

	return errs
}

func validateUser(errs []ValidationError, u User, idx int, seen map[string]bool) []ValidationError {
	prefix := fmt.Sprintf("users[%d]", idx)

	email := strings.ToLower(u.Email)
	switch {
	case email == "":
		errs = append(errs, ValidationError{Field: prefix + ".email", Message: "required"})
	case !emailRegex.MatchString(email):
		errs = append(errs, ValidationError{Field: prefix + ".email", Message: fmt.Sprintf("invalid email: %s", u.Email)})
	case seen[email]:
		errs = append(errs, ValidationError{Field: prefix + ".email", Message: fmt.Sprintf("duplicate email: %s", u.Email)})
	}
	if email != "" {
		seen[email] = true
	}

	if !validAccountAccess[u.AccountAccess] {
		errs = append(errs, ValidationError{
			Field:   prefix + ".account_access",
			Message: fmt.Sprintf("must be 'admin', 'user', or 'noAccess', got '%s'", u.AccountAccess),
		})
	}

	containerIDs := make(map[string]bool)
	for j, ca := range u.ContainerAccess {
		errs = validateContainerAccess(errs, ca, prefix, j, containerIDs)
	}

	return errs
}

func validateContainerAccess(errs []ValidationError, ca ContainerAccess, prefix string, idx int, seen map[string]bool) []ValidationError {
	caPrefix := fmt.Sprintf("%s.container_access[%d]", prefix, idx)

	switch {
	case ca.ContainerID == "":
		errs = append(errs, ValidationError{Field: caPrefix + ".container_id", Message: "required"})
	case !containerIDRe.MatchString(ca.ContainerID):
		errs = append(errs, ValidationError{
			Field:   caPrefix + ".container_id",
			Message: fmt.Sprintf("invalid container ID format: %s (expected GTM-XXXXXXXX)", ca.ContainerID),
		})
	case seen[ca.ContainerID]:
		errs = append(errs, ValidationError{
			Field:   caPrefix + ".container_id",
			Message: fmt.Sprintf("duplicate container_id: %s", ca.ContainerID),
		})
	}
	if ca.ContainerID != "" {
		seen[ca.ContainerID] = true
	}

	if !validContainerPermissions[ca.Permission] {
		errs = append(errs, ValidationError{
			Field:   caPrefix + ".permission",
			Message: fmt.Sprintf("must be 'read', 'edit', 'approve', or 'publish', got '%s'", ca.Permission),
		})
	}

	return errs
}
