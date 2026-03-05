package state

import (
	"sort"
	"strings"

	"github.com/h13/gtm-users/internal/config"
)

// FromConfig converts a Config into an AccountState (the "desired" state).
func FromConfig(cfg config.Config) AccountState {
	users := make([]UserPermission, 0, len(cfg.Users))
	for _, u := range cfg.Users {
		containers := make([]ContainerPermission, 0, len(u.ContainerAccess))
		for _, ca := range u.ContainerAccess {
			containers = append(containers, ContainerPermission{
				ContainerID: ca.ContainerID,
				Permission:  string(ca.Permission),
			})
		}
		sort.Slice(containers, func(i, j int) bool {
			return containers[i].ContainerID < containers[j].ContainerID
		})

		users = append(users, UserPermission{
			Email:           strings.ToLower(u.Email),
			AccountAccess:   string(u.AccountAccess),
			ContainerAccess: containers,
		})
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].Email < users[j].Email
	})

	return AccountState{
		AccountID: cfg.AccountID,
		Users:     users,
	}
}
