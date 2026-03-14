package diff

import (
	"cmp"
	"slices"

	"github.com/h13/gtm-users/internal/config"
	"github.com/h13/gtm-users/internal/state"
)

// ActionType represents the type of change needed.
type ActionType string

const (
	ActionAdd    ActionType = "add"
	ActionUpdate ActionType = "update"
	ActionDelete ActionType = "delete"
)

// ContainerChange represents a change to a container permission.
type ContainerChange struct {
	ContainerID   string     `json:"container_id"`
	Action        ActionType `json:"action"`
	OldPermission string     `json:"old_permission,omitempty"`
	NewPermission string     `json:"new_permission,omitempty"`
}

// UserChange represents a change to a user's permissions.
type UserChange struct {
	Email            string            `json:"email"`
	Action           ActionType        `json:"action"`
	OldAccountAccess string            `json:"old_account_access,omitempty"`
	NewAccountAccess string            `json:"new_account_access,omitempty"`
	ContainerChanges []ContainerChange `json:"container_changes,omitempty"`
}

// Plan represents the full set of changes to apply.
type Plan struct {
	AccountID string       `json:"account_id"`
	Changes   []UserChange `json:"changes"`
	Mode      config.Mode  `json:"mode"`
}

// HasChanges returns true if the plan contains any changes.
func (p Plan) HasChanges() bool {
	return len(p.Changes) > 0
}

// Summary returns counts of add/update/delete operations.
func (p Plan) Summary() (adds, updates, deletes int) {
	for _, c := range p.Changes {
		switch c.Action {
		case ActionAdd:
			adds++
		case ActionUpdate:
			updates++
		case ActionDelete:
			deletes++
		}
	}
	return
}

// Compute calculates the diff between desired and actual states.
func Compute(desired, actual state.AccountState, mode config.Mode) Plan {
	plan := Plan{
		AccountID: desired.AccountID,
		Changes:   []UserChange{},
		Mode:      mode,
	}

	desiredMap := desired.UserMap()
	actualMap := actual.UserMap()

	// Check desired users against actual
	for _, du := range desired.Users {
		au, exists := actualMap[du.Email]
		if !exists {
			// User needs to be added
			containerChanges := make([]ContainerChange, 0, len(du.ContainerAccess))
			for _, ca := range du.ContainerAccess {
				containerChanges = append(containerChanges, ContainerChange{
					ContainerID:   ca.ContainerID,
					Action:        ActionAdd,
					NewPermission: ca.Permission,
				})
			}
			plan.Changes = append(plan.Changes, UserChange{
				Email:            du.Email,
				Action:           ActionAdd,
				NewAccountAccess: du.AccountAccess,
				ContainerChanges: containerChanges,
			})
			continue
		}

		// User exists - check for changes
		change := computeUserChange(du, au)
		if change != nil {
			plan.Changes = append(plan.Changes, *change)
		}
	}

	// In authoritative mode, check for users to delete
	if mode == config.ModeAuthoritative {
		for _, au := range actual.Users {
			if _, exists := desiredMap[au.Email]; !exists {
				plan.Changes = append(plan.Changes, UserChange{
					Email:            au.Email,
					Action:           ActionDelete,
					OldAccountAccess: au.AccountAccess,
				})
			}
		}
	}

	return plan
}

func computeUserChange(desired, actual state.UserPermission) *UserChange {
	accountChanged := desired.AccountAccess != actual.AccountAccess
	containerChanges := computeContainerChanges(desired, actual)

	if !accountChanged && len(containerChanges) == 0 {
		return nil
	}

	return &UserChange{
		Email:            desired.Email,
		Action:           ActionUpdate,
		OldAccountAccess: actual.AccountAccess,
		NewAccountAccess: desired.AccountAccess,
		ContainerChanges: containerChanges,
	}
}

func computeContainerChanges(desired, actual state.UserPermission) []ContainerChange {
	desiredMap := desired.ContainerMap()
	actualMap := actual.ContainerMap()
	var changes []ContainerChange

	// Check desired containers
	for cID, dPerm := range desiredMap {
		aPerm, exists := actualMap[cID]
		if !exists {
			changes = append(changes, ContainerChange{
				ContainerID:   cID,
				Action:        ActionAdd,
				NewPermission: dPerm,
			})
		} else if dPerm != aPerm {
			changes = append(changes, ContainerChange{
				ContainerID:   cID,
				Action:        ActionUpdate,
				OldPermission: aPerm,
				NewPermission: dPerm,
			})
		}
	}

	// Check for containers to remove (present in actual but not in desired)
	for cID, aPerm := range actualMap {
		if _, exists := desiredMap[cID]; !exists {
			changes = append(changes, ContainerChange{
				ContainerID:   cID,
				Action:        ActionDelete,
				OldPermission: aPerm,
			})
		}
	}

	slices.SortFunc(changes, func(a, b ContainerChange) int {
		return cmp.Compare(a.ContainerID, b.ContainerID)
	})

	return changes
}
