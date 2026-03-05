package state

// UserPermission represents a normalized GTM user permission state.
type UserPermission struct {
	Email           string
	AccountAccess   string
	ContainerAccess []ContainerPermission
}

// ContainerPermission represents a permission for a specific container.
type ContainerPermission struct {
	ContainerID string
	Permission  string
}

// AccountState represents the full state of user permissions for a GTM account.
type AccountState struct {
	AccountID string
	Users     []UserPermission
}

// UserMap returns a map of email -> UserPermission for quick lookups.
func (s AccountState) UserMap() map[string]UserPermission {
	m := make(map[string]UserPermission, len(s.Users))
	for _, u := range s.Users {
		m[u.Email] = u
	}
	return m
}

// ContainerMap returns a map of containerID -> permission for a user.
func (u UserPermission) ContainerMap() map[string]string {
	m := make(map[string]string, len(u.ContainerAccess))
	for _, ca := range u.ContainerAccess {
		m[ca.ContainerID] = ca.Permission
	}
	return m
}
