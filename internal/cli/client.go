package cli

import (
	"context"

	"github.com/h13/gtm-users/internal/state"
)

// gtmClient defines the interface for GTM API operations.
type gtmClient interface {
	FetchState(ctx context.Context) (state.AccountState, error)
	CreateUserPermission(ctx context.Context, user state.UserPermission) error
	UpdateUserPermission(ctx context.Context, user state.UserPermission) error
	DeleteUserPermission(ctx context.Context, email string) error
}
