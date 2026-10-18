package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"invitation/pkg/domain"

	"golang.org/x/crypto/bcrypt"
)

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(hash)
}

func TestLoadAdminsFile(t *testing.T) {
	tempDir := t.TempDir()
	hash := mustHashPassword(t, "s3cr3t-pass!")

	t.Run("loads object format with admins array", func(t *testing.T) {
		path := filepath.Join(tempDir, "admins_obj.json")
		content := `{"admins": [{"username": "admin", "password_hash": "` + hash + `"}]}`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		users, err := LoadAdminsFile(path)
		if err != nil {
			t.Fatalf("LoadAdminsFile failed: %v", err)
		}
		if len(users) != 1 || users[0].Username != "admin" {
			t.Errorf("unexpected users loaded: %+v", users)
		}
	})

	t.Run("loads top-level array format", func(t *testing.T) {
		path := filepath.Join(tempDir, "admins_arr.json")
		content := `[{"username": "maria", "password_hash": "` + hash + `"}]`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		users, err := LoadAdminsFile(path)
		if err != nil {
			t.Fatalf("LoadAdminsFile failed: %v", err)
		}
		if len(users) != 1 || users[0].Username != "maria" {
			t.Errorf("unexpected users loaded: %+v", users)
		}
	})

	t.Run("rejects invalid bcrypt hash", func(t *testing.T) {
		path := filepath.Join(tempDir, "admins_invalid.json")
		content := `{"admins": [{"username": "admin", "password_hash": "plaintext-not-bcrypt"}]}`
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		if _, err := LoadAdminsFile(path); err == nil {
			t.Errorf("expected error for invalid bcrypt hash, got nil")
		}
	})

	t.Run("rejects empty admins list", func(t *testing.T) {
		path := filepath.Join(tempDir, "admins_empty.json")
		if err := os.WriteFile(path, []byte(`{"admins": []}`), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		if _, err := LoadAdminsFile(path); err == nil {
			t.Errorf("expected error for empty admins list, got nil")
		}
	})
}

func TestAdminAuthFlow(t *testing.T) {
	tempUploads := t.TempDir()
	adminPass := "quince-planner-2026"
	adminHash := mustHashPassword(t, adminPass)

	srv, err := New(Config{
		Port:       "8080",
		SeedDir:    "../../seed",
		UploadsDir: tempUploads,
		Admins: []AdminUser{
			{
				Username:     "planner",
				PasswordHash: adminHash,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Seed an invitation for route testing
	weddingData, err := os.ReadFile("../../seed/wedding.json")
	if err != nil {
		t.Fatalf("failed to read seed file: %v", err)
	}
	var testInv domain.Invitation
	if err := json.Unmarshal(weddingData, &testInv); err != nil {
		t.Fatalf("failed to unmarshal seed file: %v", err)
	}
	if err := srv.store.SaveInvitation(t.Context(), &testInv); err != nil {
		t.Fatalf("failed to save seed invitation: %v", err)
	}

	t.Run("Public routes remain accessible without authentication", func(t *testing.T) {
		for _, path := range []string{"/", "/health", "/i/sarah-and-alex-wedding", "/api/invitations", "/api/invitations/sarah-and-alex-wedding"} {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			srv.Router().ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected GET %s to return 200 OK without auth, got %d", path, rec.Code)
			}
		}
	})

	t.Run("Unauthenticated GET requests to /admin pages redirect to /admin/login", func(t *testing.T) {
		adminPages := []string{
			"/admin",
			"/admin/invitations/new",
			"/admin/invitations/sarah-and-alex-wedding/edit",
			"/admin/invitations/sarah-and-alex-wedding/rsvps",
			"/admin/invitations/sarah-and-alex-wedding/rsvps.csv",
		}
		for _, page := range adminPages {
			req := httptest.NewRequest(http.MethodGet, page, nil)
			rec := httptest.NewRecorder()
			srv.Router().ServeHTTP(rec, req)

			if rec.Code != http.StatusSeeOther {
				t.Errorf("expected GET %s to return 303 See Other, got %d", page, rec.Code)
			}
			if loc := rec.Header().Get("Location"); loc != "/admin/login" {
				t.Errorf("expected redirect Location /admin/login for %s, got %q", page, loc)
			}
		}
	})

	t.Run("Unauthenticated POST /admin/invitations/preview returns 401 Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/admin/invitations/preview", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized, got %d", rec.Code)
		}
	})

	t.Run("Unauthenticated protected /api routes return 401 Unauthorized", func(t *testing.T) {
		cases := []struct {
			method string
			path   string
		}{
			{http.MethodPost, "/api/invitations"},
			{http.MethodPut, "/api/invitations/sarah-and-alex-wedding"},
			{http.MethodDelete, "/api/invitations/sarah-and-alex-wedding"},
			{http.MethodGet, "/api/invitations/sarah-and-alex-wedding/rsvps"},
			{http.MethodPost, "/api/upload"},
		}
		for _, tc := range cases {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			srv.Router().ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected %s %s to return 401 Unauthorized, got %d", tc.method, tc.path, rec.Code)
			}
		}
	})

	t.Run("POST /admin/login rejects invalid credentials with 401", func(t *testing.T) {
		payload := `{"username": "planner", "password": "wrong-password"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized for wrong password, got %d", rec.Code)
		}
		if len(rec.Result().Cookies()) != 0 {
			t.Errorf("expected no session cookie on failed login, got %v", rec.Result().Cookies())
		}
	})

	var sessionCookie *http.Cookie

	t.Run("POST /admin/login with valid form credentials sets session cookie and redirects to /admin", func(t *testing.T) {
		form := url.Values{
			"username": {"planner"},
			"password": {adminPass},
		}
		req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusSeeOther {
			t.Fatalf("expected 303 See Other on form login, got %d: %s", rec.Code, rec.Body.String())
		}
		if loc := rec.Header().Get("Location"); loc != "/admin" {
			t.Errorf("expected redirect to /admin, got %q", loc)
		}

		cookies := rec.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatalf("expected session cookie to be set")
		}
		sessionCookie = cookies[0]
		if !sessionCookie.HttpOnly {
			t.Errorf("expected HttpOnly session cookie")
		}
	})

	t.Run("Authenticated session grants access to /admin and protected /api routes", func(t *testing.T) {
		if sessionCookie == nil {
			t.Skip("no session cookie from login test")
		}

		// 1. Access /admin
		reqAdmin := httptest.NewRequest(http.MethodGet, "/admin", nil)
		reqAdmin.AddCookie(sessionCookie)
		recAdmin := httptest.NewRecorder()
		srv.Router().ServeHTTP(recAdmin, reqAdmin)

		if recAdmin.Code != http.StatusOK {
			t.Errorf("expected 200 OK on /admin with valid session cookie, got %d", recAdmin.Code)
		}

		// 2. Access GET /api/invitations/{slug}/rsvps
		reqAPI := httptest.NewRequest(http.MethodGet, "/api/invitations/sarah-and-alex-wedding/rsvps", nil)
		reqAPI.AddCookie(sessionCookie)
		recAPI := httptest.NewRecorder()
		srv.Router().ServeHTTP(recAPI, reqAPI)

		if recAPI.Code != http.StatusOK {
			t.Errorf("expected 200 OK on protected API route with valid session cookie, got %d", recAPI.Code)
		}
	})

	t.Run("POST /admin/logout destroys session and revokes access", func(t *testing.T) {
		if sessionCookie == nil {
			t.Skip("no session cookie from login test")
		}

		reqLogout := httptest.NewRequest(http.MethodPost, "/admin/logout", nil)
		reqLogout.AddCookie(sessionCookie)
		recLogout := httptest.NewRecorder()
		srv.Router().ServeHTTP(recLogout, reqLogout)

		if recLogout.Code != http.StatusSeeOther {
			t.Fatalf("expected 303 See Other on logout, got %d", recLogout.Code)
		}
		if loc := recLogout.Header().Get("Location"); loc != "/admin/login" {
			t.Errorf("expected redirect to /admin/login, got %q", loc)
		}

		// Verify old session token is deleted from SQLite store and no longer grants access
		reqAfter := httptest.NewRequest(http.MethodGet, "/admin", nil)
		reqAfter.AddCookie(sessionCookie)
		recAfter := httptest.NewRecorder()
		srv.Router().ServeHTTP(recAfter, reqAfter)

		if recAfter.Code != http.StatusSeeOther {
			t.Errorf("expected 303 redirect after logout when reusing old session cookie, got %d", recAfter.Code)
		}
	})

	t.Run("POST /admin/login with JSON returns 200 OK JSON payload", func(t *testing.T) {
		payload, _ := json.Marshal(map[string]string{
			"username": "planner",
			"password": adminPass,
		})
		req := httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on JSON login, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode JSON login response: %v", err)
		}
		if resp["status"] != "authenticated" || resp["username"] != "planner" {
			t.Errorf("unexpected JSON login response: %+v", resp)
		}
	})
}
