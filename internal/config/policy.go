package config

import "fmt"

// Policy defines organizational constraints on the configuration.
type Policy struct {
	MaxAdmins      *int     `yaml:"max_admins,omitempty"`
	RequireApprove []string `yaml:"require_approve_for_publish,omitempty"`
}

// ValidatePolicy checks policy constraints against the resolved config.
func ValidatePolicy(cfg Config) []ValidationError {
	if cfg.Policy == nil {
		return nil
	}

	var errs []ValidationError
	errs = validateMaxAdmins(errs, cfg)
	errs = validateRequireApprove(errs, cfg)

	return errs
}

func validateMaxAdmins(errs []ValidationError, cfg Config) []ValidationError {
	if cfg.Policy.MaxAdmins == nil {
		return errs
	}

	limit := *cfg.Policy.MaxAdmins
	var count int
	for _, u := range cfg.Users {
		if u.AccountAccess == AccountAccessAdmin {
			count++
		}
	}

	if count > limit {
		errs = append(errs, ValidationError{
			Field:   "policy.max_admins",
			Message: fmt.Sprintf("admin count %d exceeds limit %d", count, limit),
		})
	}

	return errs
}

func validateRequireApprove(errs []ValidationError, cfg Config) []ValidationError {
	if len(cfg.Policy.RequireApprove) == 0 {
		return errs
	}

	required := make(map[string]bool, len(cfg.Policy.RequireApprove))
	for _, cid := range cfg.Policy.RequireApprove {
		required[cid] = true
	}

	for i, u := range cfg.Users {
		for _, ca := range u.ContainerAccess {
			if ca.Permission != PermissionPublish {
				continue
			}
			if !required[ca.ContainerID] {
				continue
			}

			if !userHasApprove(u, ca.ContainerID) {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("users[%d]", i),
					Message: fmt.Sprintf("user %s has publish on %s but lacks approve", u.Email, ca.ContainerID),
				})
			}
		}
	}

	return errs
}

func userHasApprove(u User, containerID string) bool {
	for _, ca := range u.ContainerAccess {
		if ca.ContainerID == containerID && ca.Permission == PermissionApprove {
			return true
		}
	}
	return false
}
