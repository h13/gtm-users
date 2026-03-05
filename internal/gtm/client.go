package gtm

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/h13/gtm-users/internal/state"
	"google.golang.org/api/option"
	"google.golang.org/api/tagmanager/v2"
)

const permNoAccess = "noAccess"

// Client wraps the GTM API for user permission operations.
type Client struct {
	svc       *tagmanager.Service
	accountID string
}

// NewClient creates a GTM API client with the given credentials file.
func NewClient(ctx context.Context, accountID string, credentialsFile string) (*Client, error) {
	svc, err := tagmanager.NewService(ctx,
		option.WithCredentialsFile(credentialsFile), //nolint:staticcheck // no replacement available yet
		option.WithScopes(tagmanager.TagmanagerManageUsersScope),
	)
	if err != nil {
		return nil, fmt.Errorf("creating tagmanager service: %w", err)
	}

	return &Client{svc: svc, accountID: accountID}, nil
}

// accountPath returns the GTM API path for the account.
func (c *Client) accountPath() string {
	return fmt.Sprintf("accounts/%s", c.accountID)
}

// FetchState retrieves the current user permissions from the GTM API.
func (c *Client) FetchState(ctx context.Context) (state.AccountState, error) {
	resp, err := c.svc.Accounts.UserPermissions.List(c.accountPath()).Context(ctx).Do()
	if err != nil {
		return state.AccountState{}, fmt.Errorf("listing user permissions: %w", err)
	}

	users := make([]state.UserPermission, 0, len(resp.UserPermission))
	for _, up := range resp.UserPermission {
		containers := make([]state.ContainerPermission, 0, len(up.ContainerAccess))
		for _, ca := range up.ContainerAccess {
			containers = append(containers, state.ContainerPermission{
				ContainerID: ca.ContainerId,
				Permission:  ca.Permission,
			})
		}
		slices.SortFunc(containers, func(a, b state.ContainerPermission) int {
			return cmp.Compare(a.ContainerID, b.ContainerID)
		})

		users = append(users, state.UserPermission{
			Email:           strings.ToLower(up.EmailAddress),
			AccountAccess:   mapAPIAccountAccess(up.AccountAccess),
			ContainerAccess: containers,
		})
	}

	slices.SortFunc(users, func(a, b state.UserPermission) int {
		return cmp.Compare(a.Email, b.Email)
	})

	return state.AccountState{
		AccountID: c.accountID,
		Users:     users,
	}, nil
}

// CreateUserPermission creates a new user permission in the GTM account.
func (c *Client) CreateUserPermission(ctx context.Context, user state.UserPermission) error {
	up := buildAPIUserPermission(user)
	_, err := c.svc.Accounts.UserPermissions.Create(c.accountPath(), up).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating permission for %s: %w", user.Email, err)
	}
	return nil
}

// UpdateUserPermission updates an existing user permission.
func (c *Client) UpdateUserPermission(ctx context.Context, user state.UserPermission) error {
	path, err := c.findUserPermissionPath(ctx, user.Email)
	if err != nil {
		return err
	}

	up := buildAPIUserPermission(user)
	_, err = c.svc.Accounts.UserPermissions.Update(path, up).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating permission for %s: %w", user.Email, err)
	}
	return nil
}

// DeleteUserPermission removes a user's permissions from the account.
func (c *Client) DeleteUserPermission(ctx context.Context, email string) error {
	path, err := c.findUserPermissionPath(ctx, email)
	if err != nil {
		return err
	}

	err = c.svc.Accounts.UserPermissions.Delete(path).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting permission for %s: %w", email, err)
	}
	return nil
}

func (c *Client) findUserPermissionPath(ctx context.Context, email string) (string, error) {
	resp, err := c.svc.Accounts.UserPermissions.List(c.accountPath()).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("listing user permissions: %w", err)
	}

	email = strings.ToLower(email)
	for _, up := range resp.UserPermission {
		if strings.ToLower(up.EmailAddress) == email {
			return up.Path, nil
		}
	}
	return "", fmt.Errorf("user permission not found for %s", email)
}

func buildAPIUserPermission(user state.UserPermission) *tagmanager.UserPermission {
	containers := make([]*tagmanager.ContainerAccess, 0, len(user.ContainerAccess))
	for _, ca := range user.ContainerAccess {
		containers = append(containers, &tagmanager.ContainerAccess{
			ContainerId: ca.ContainerID,
			Permission:  ca.Permission,
		})
	}

	return &tagmanager.UserPermission{
		EmailAddress: user.Email,
		AccountAccess: &tagmanager.AccountAccess{
			Permission: user.AccountAccess,
		},
		ContainerAccess: containers,
	}
}

// mapAPIAccountAccess converts GTM API account access to our internal representation.
func mapAPIAccountAccess(access *tagmanager.AccountAccess) string {
	if access == nil {
		return permNoAccess
	}
	return access.Permission
}
