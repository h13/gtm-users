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
	pathCache map[string]string // email -> API path, populated by FetchState
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

// listAll retrieves all user permissions with pagination support.
func (c *Client) listAll(ctx context.Context) ([]*tagmanager.UserPermission, error) {
	var all []*tagmanager.UserPermission
	pageToken := ""

	for {
		call := c.svc.Accounts.UserPermissions.List(c.accountPath())
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("listing user permissions: %w", err)
		}

		all = append(all, resp.UserPermission...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return all, nil
}

// FetchState retrieves the current user permissions from the GTM API.
func (c *Client) FetchState(ctx context.Context) (state.AccountState, error) {
	allPerms, err := c.listAll(ctx)
	if err != nil {
		return state.AccountState{}, err
	}

	c.pathCache = make(map[string]string, len(allPerms))

	users := make([]state.UserPermission, 0, len(allPerms))
	for _, up := range allPerms {
		email := strings.ToLower(up.EmailAddress)
		c.pathCache[email] = up.Path

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
			Email:           email,
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
	created, err := c.svc.Accounts.UserPermissions.Create(c.accountPath(), up).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating permission for %s: %w", user.Email, err)
	}

	if c.pathCache != nil && created.Path != "" {
		c.pathCache[strings.ToLower(created.EmailAddress)] = created.Path
	}
	return nil
}

// UpdateUserPermission updates an existing user permission.
func (c *Client) UpdateUserPermission(ctx context.Context, user state.UserPermission) error {
	path, err := c.resolveUserPath(ctx, user.Email)
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
	path, err := c.resolveUserPath(ctx, email)
	if err != nil {
		return err
	}

	err = c.svc.Accounts.UserPermissions.Delete(path).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting permission for %s: %w", email, err)
	}
	return nil
}

// resolveUserPath returns the API path for a user, using the cache when available.
func (c *Client) resolveUserPath(ctx context.Context, email string) (string, error) {
	email = strings.ToLower(email)

	if c.pathCache != nil {
		if path, ok := c.pathCache[email]; ok {
			return path, nil
		}
	}

	// Cache miss: fall back to API call.
	allPerms, err := c.listAll(ctx)
	if err != nil {
		return "", err
	}

	c.pathCache = make(map[string]string, len(allPerms))
	for _, up := range allPerms {
		c.pathCache[strings.ToLower(up.EmailAddress)] = up.Path
	}

	if path, ok := c.pathCache[email]; ok {
		return path, nil
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
