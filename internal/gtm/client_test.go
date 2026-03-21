package gtm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/h13/gtm-users/internal/state"
	"google.golang.org/api/option"
	"google.golang.org/api/tagmanager/v2"
)

func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client, err := NewClient(context.Background(), "12345", "",
		WithAPIOptions(
			option.WithHTTPClient(srv.Client()),
			option.WithEndpoint(srv.URL),
			option.WithoutAuthentication(),
		),
	)
	if err != nil {
		t.Fatalf("creating test client: %v", err)
	}

	return client
}

func jsonResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v) //nolint:errcheck,gosec
}

func TestNewClient_ADC(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions", func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, &tagmanager.ListUserPermissionsResponse{})
	})

	client := newTestClient(t, mux)

	// Verify the client is functional — scopes must be set even without credentials file.
	st, err := client.FetchState(context.Background())
	if err != nil {
		t.Fatalf("FetchState with ADC client: %v", err)
	}
	if st.AccountID != "12345" {
		t.Errorf("account ID = %q, want 12345", st.AccountID)
	}
}

func TestNewClient_InvalidCredentials(t *testing.T) {
	_, err := NewClient(context.Background(), "12345", "/nonexistent/credentials.json")
	if err == nil {
		t.Fatal("expected error for invalid credentials, got nil")
	}
}

func TestAccountPath(t *testing.T) {
	c := &Client{accountID: "99999"}
	got := c.accountPath()
	want := "accounts/99999"
	if got != want {
		t.Errorf("accountPath() = %q, want %q", got, want)
	}
}

func TestMapAPIAccountAccess_Nil(t *testing.T) {
	got := mapAPIAccountAccess(nil)
	if got != permNoAccess {
		t.Errorf("mapAPIAccountAccess(nil) = %q, want %q", got, permNoAccess)
	}
}

func TestMapAPIAccountAccess_NonNil(t *testing.T) {
	got := mapAPIAccountAccess(&tagmanager.AccountAccess{Permission: "admin"})
	if got != "admin" {
		t.Errorf("mapAPIAccountAccess = %q, want %q", got, "admin")
	}
}

func TestBuildAPIUserPermission(t *testing.T) {
	user := state.UserPermission{
		Email:         "alice@example.com",
		AccountAccess: "admin",
		ContainerAccess: []state.ContainerPermission{
			{ContainerID: "GTM-AAAA1111", Permission: "publish"},
			{ContainerID: "GTM-BBBB2222", Permission: "read"},
		},
	}

	up := buildAPIUserPermission(user)

	if up.EmailAddress != "alice@example.com" {
		t.Errorf("email = %q, want alice@example.com", up.EmailAddress)
	}
	if up.AccountAccess.Permission != "admin" {
		t.Errorf("account access = %q, want admin", up.AccountAccess.Permission)
	}
	if len(up.ContainerAccess) != 2 {
		t.Fatalf("container access count = %d, want 2", len(up.ContainerAccess))
	}
	if up.ContainerAccess[0].ContainerId != "GTM-AAAA1111" {
		t.Errorf("container[0] = %q, want GTM-AAAA1111", up.ContainerAccess[0].ContainerId)
	}
}

func TestBuildAPIUserPermission_NoContainers(t *testing.T) {
	user := state.UserPermission{
		Email:         "bob@example.com",
		AccountAccess: "user",
	}

	up := buildAPIUserPermission(user)
	if len(up.ContainerAccess) != 0 {
		t.Errorf("container access count = %d, want 0", len(up.ContainerAccess))
	}
}

func TestFetchState_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		jsonResponse(w, &tagmanager.ListUserPermissionsResponse{
			UserPermission: []*tagmanager.UserPermission{
				{
					EmailAddress:  "Bob@Example.com",
					Path:          "accounts/12345/user_permissions/100",
					AccountAccess: &tagmanager.AccountAccess{Permission: "admin"},
					ContainerAccess: []*tagmanager.ContainerAccess{
						{ContainerId: "GTM-BBBB2222", Permission: "publish"},
						{ContainerId: "GTM-AAAA1111", Permission: "read"},
					},
				},
				{
					EmailAddress:  "alice@example.com",
					Path:          "accounts/12345/user_permissions/101",
					AccountAccess: &tagmanager.AccountAccess{Permission: "user"},
				},
			},
		})
	})

	client := newTestClient(t, mux)
	st, err := client.FetchState(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if st.AccountID != "12345" {
		t.Errorf("account ID = %q, want 12345", st.AccountID)
	}

	// Users sorted by email
	if len(st.Users) != 2 {
		t.Fatalf("user count = %d, want 2", len(st.Users))
	}
	if st.Users[0].Email != "alice@example.com" {
		t.Errorf("users[0].email = %q, want alice@example.com", st.Users[0].Email)
	}
	if st.Users[1].Email != "bob@example.com" {
		t.Errorf("users[1].email = %q, want bob@example.com (lowercased)", st.Users[1].Email)
	}
	if st.Users[1].AccountAccess != "admin" {
		t.Errorf("bob access = %q, want admin", st.Users[1].AccountAccess)
	}

	// Containers sorted by ID
	if len(st.Users[1].ContainerAccess) != 2 {
		t.Fatalf("bob container count = %d, want 2", len(st.Users[1].ContainerAccess))
	}
	if st.Users[1].ContainerAccess[0].ContainerID != "GTM-AAAA1111" {
		t.Errorf("bob containers[0] = %q, want GTM-AAAA1111", st.Users[1].ContainerAccess[0].ContainerID)
	}

	// Path cache populated
	if client.pathCache["alice@example.com"] != "accounts/12345/user_permissions/101" {
		t.Errorf("path cache not populated for alice")
	}
	if client.pathCache["bob@example.com"] != "accounts/12345/user_permissions/100" {
		t.Errorf("path cache not populated for bob")
	}
}

func TestFetchState_NilAccountAccess(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions", func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, &tagmanager.ListUserPermissionsResponse{
			UserPermission: []*tagmanager.UserPermission{
				{
					EmailAddress: "alice@example.com",
					Path:         "accounts/12345/user_permissions/100",
					// AccountAccess intentionally nil
				},
			},
		})
	})

	client := newTestClient(t, mux)
	st, err := client.FetchState(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if st.Users[0].AccountAccess != permNoAccess {
		t.Errorf("account access = %q, want %q", st.Users[0].AccountAccess, permNoAccess)
	}
}

func TestFetchState_Pagination(t *testing.T) {
	callCount := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		pageToken := r.URL.Query().Get("pageToken")
		switch pageToken {
		case "":
			jsonResponse(w, &tagmanager.ListUserPermissionsResponse{
				UserPermission: []*tagmanager.UserPermission{
					{EmailAddress: "alice@example.com", Path: "p/1", AccountAccess: &tagmanager.AccountAccess{Permission: "user"}},
				},
				NextPageToken: "page2",
			})
		case "page2":
			jsonResponse(w, &tagmanager.ListUserPermissionsResponse{
				UserPermission: []*tagmanager.UserPermission{
					{EmailAddress: "bob@example.com", Path: "p/2", AccountAccess: &tagmanager.AccountAccess{Permission: "admin"}},
				},
			})
		}
	})

	client := newTestClient(t, mux)
	st, err := client.FetchState(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(st.Users) != 2 {
		t.Errorf("user count = %d, want 2", len(st.Users))
	}
	if callCount != 2 {
		t.Errorf("API calls = %d, want 2", callCount)
	}
}

func TestFetchState_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions", func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, &tagmanager.ListUserPermissionsResponse{})
	})

	client := newTestClient(t, mux)
	st, err := client.FetchState(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(st.Users) != 0 {
		t.Errorf("user count = %d, want 0", len(st.Users))
	}
}

func TestFetchState_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	})

	client := newTestClient(t, mux)
	_, err := client.FetchState(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateUserPermission_Success(t *testing.T) {
	var gotBody tagmanager.UserPermission
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		json.NewDecoder(r.Body).Decode(&gotBody) //nolint:errcheck,gosec
		jsonResponse(w, &tagmanager.UserPermission{
			EmailAddress: "alice@example.com",
			Path:         "accounts/12345/user_permissions/200",
		})
	})

	client := newTestClient(t, mux)
	client.pathCache = make(map[string]string)

	user := state.UserPermission{
		Email:         "alice@example.com",
		AccountAccess: "user",
		ContainerAccess: []state.ContainerPermission{
			{ContainerID: "GTM-AAAA1111", Permission: "read"},
		},
	}

	if err := client.CreateUserPermission(context.Background(), user); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody.EmailAddress != "alice@example.com" {
		t.Errorf("request email = %q, want alice@example.com", gotBody.EmailAddress)
	}

	if client.pathCache["alice@example.com"] != "accounts/12345/user_permissions/200" {
		t.Errorf("path cache not updated")
	}
}

func TestCreateUserPermission_NilPathCache(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions", func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, &tagmanager.UserPermission{
			EmailAddress: "alice@example.com",
			Path:         "accounts/12345/user_permissions/200",
		})
	})

	client := newTestClient(t, mux)
	// pathCache intentionally nil

	user := state.UserPermission{Email: "alice@example.com", AccountAccess: "user"}
	if err := client.CreateUserPermission(context.Background(), user); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateUserPermission_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	})

	client := newTestClient(t, mux)
	user := state.UserPermission{Email: "alice@example.com", AccountAccess: "user"}

	err := client.CreateUserPermission(context.Background(), user)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateUserPermission_CacheHit(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions/100", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		jsonResponse(w, &tagmanager.UserPermission{
			EmailAddress: "alice@example.com",
			Path:         "accounts/12345/user_permissions/100",
		})
	})

	client := newTestClient(t, mux)
	client.pathCache = map[string]string{
		"alice@example.com": "accounts/12345/user_permissions/100",
	}

	user := state.UserPermission{Email: "alice@example.com", AccountAccess: "admin"}
	if err := client.UpdateUserPermission(context.Background(), user); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateUserPermission_CacheMiss(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions", func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, &tagmanager.ListUserPermissionsResponse{
			UserPermission: []*tagmanager.UserPermission{
				{
					EmailAddress:  "alice@example.com",
					Path:          "accounts/12345/user_permissions/100",
					AccountAccess: &tagmanager.AccountAccess{Permission: "user"},
				},
			},
		})
	})
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions/100", func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, &tagmanager.UserPermission{
			EmailAddress: "alice@example.com",
			Path:         "accounts/12345/user_permissions/100",
		})
	})

	client := newTestClient(t, mux)
	client.pathCache = make(map[string]string) // empty cache

	user := state.UserPermission{Email: "alice@example.com", AccountAccess: "admin"}
	if err := client.UpdateUserPermission(context.Background(), user); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateUserPermission_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions/100", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	})

	client := newTestClient(t, mux)
	client.pathCache = map[string]string{
		"alice@example.com": "accounts/12345/user_permissions/100",
	}

	user := state.UserPermission{Email: "alice@example.com", AccountAccess: "admin"}
	err := client.UpdateUserPermission(context.Background(), user)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteUserPermission_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions/100", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, mux)
	client.pathCache = map[string]string{
		"alice@example.com": "accounts/12345/user_permissions/100",
	}

	if err := client.DeleteUserPermission(context.Background(), "alice@example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := client.pathCache["alice@example.com"]; ok {
		t.Error("path cache should be cleaned after delete")
	}
}

func TestDeleteUserPermission_NilPathCache(t *testing.T) {
	mux := http.NewServeMux()
	// listAll fallback
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions", func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, &tagmanager.ListUserPermissionsResponse{
			UserPermission: []*tagmanager.UserPermission{
				{
					EmailAddress:  "alice@example.com",
					Path:          "accounts/12345/user_permissions/100",
					AccountAccess: &tagmanager.AccountAccess{Permission: "user"},
				},
			},
		})
	})
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions/100", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	client := newTestClient(t, mux)
	// pathCache intentionally nil

	if err := client.DeleteUserPermission(context.Background(), "alice@example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteUserPermission_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions/100", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	})

	client := newTestClient(t, mux)
	client.pathCache = map[string]string{
		"alice@example.com": "accounts/12345/user_permissions/100",
	}

	err := client.DeleteUserPermission(context.Background(), "alice@example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResolveUserPath_CacheHit(t *testing.T) {
	client := &Client{
		accountID: "12345",
		pathCache: map[string]string{
			"alice@example.com": "accounts/12345/user_permissions/100",
		},
	}

	path, err := client.resolveUserPath(context.Background(), "Alice@Example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "accounts/12345/user_permissions/100" {
		t.Errorf("path = %q, want accounts/12345/user_permissions/100", path)
	}
}

func TestResolveUserPath_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions", func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, &tagmanager.ListUserPermissionsResponse{
			UserPermission: []*tagmanager.UserPermission{
				{
					EmailAddress:  "bob@example.com",
					Path:          "accounts/12345/user_permissions/101",
					AccountAccess: &tagmanager.AccountAccess{Permission: "user"},
				},
			},
		})
	})

	client := newTestClient(t, mux)
	client.pathCache = make(map[string]string) // empty

	_, err := client.resolveUserPath(context.Background(), "unknown@example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want 'not found'", err.Error())
	}
}

func TestUpdateUserPermission_ResolveError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	})

	client := newTestClient(t, mux)
	client.pathCache = make(map[string]string) // empty - force fallback which will fail

	user := state.UserPermission{Email: "unknown@example.com", AccountAccess: "admin"}
	err := client.UpdateUserPermission(context.Background(), user)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteUserPermission_ResolveError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	})

	client := newTestClient(t, mux)
	client.pathCache = make(map[string]string) // empty - force fallback which will fail

	err := client.DeleteUserPermission(context.Background(), "unknown@example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResolveUserPath_ListAllError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tagmanager/v2/accounts/12345/user_permissions", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	})

	client := newTestClient(t, mux)
	client.pathCache = make(map[string]string) // empty - force fallback

	_, err := client.resolveUserPath(context.Background(), "alice@example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
