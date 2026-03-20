package config

// MergeConfigs merges an included config into the base config.
// Base config values take precedence for roles with the same name.
// Policy is NOT merged (only main config's policy applies).
func MergeConfigs(base, included Config) Config {
	result := Config{
		Includes:  base.Includes,
		AccountID: base.AccountID,
		Mode:      base.Mode,
		Roles:     mergeRoles(base.Roles, included.Roles),
		Policy:    base.Policy,
		Users:     append(append(make([]User, 0, len(base.Users)+len(included.Users)), base.Users...), included.Users...),
	}

	return result
}

// mergeRoles combines two role maps, with base taking precedence on conflict.
func mergeRoles(base, included map[string]Role) map[string]Role {
	if len(base) == 0 && len(included) == 0 {
		return nil
	}

	merged := make(map[string]Role, len(base)+len(included))
	for name, role := range included {
		merged[name] = role
	}
	// Base roles override included roles.
	for name, role := range base {
		merged[name] = role
	}

	return merged
}
