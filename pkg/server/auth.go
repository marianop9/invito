package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const sessionKeyAdminUser = "admin_user"

// dummyBcryptHash is used when a username does not exist to prevent timing-based user enumeration.
var dummyBcryptHash = []byte("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy")

// AdminUser represents a single administrator entry in the config file.
type AdminUser struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

// AdminsConfig represents the top-level JSON structure for the admins configuration file.
type AdminsConfig struct {
	Admins []AdminUser `json:"admins"`
}

// LoadAdminsFile reads and validates admin credentials from a JSON file.
// It supports either {"admins": [...]} or a top-level [...] array.
func LoadAdminsFile(path string) ([]AdminUser, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read admins file %q: %w", path, err)
	}

	trimmed := bytes.TrimSpace(data)
	var users []AdminUser

	if len(trimmed) > 0 && trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &users); err != nil {
			return nil, fmt.Errorf("failed to parse admins array from %q: %w", path, err)
		}
	} else {
		var cfg AdminsConfig
		if err := json.Unmarshal(trimmed, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse admins config from %q: %w", path, err)
		}
		users = cfg.Admins
	}

	if err := validateAdminUsers(users); err != nil {
		return nil, err
	}

	return users, nil
}

func validateAdminUsers(users []AdminUser) error {
	if len(users) == 0 {
		return fmt.Errorf("at least one admin user must be configured")
	}

	seen := make(map[string]bool, len(users))
	for i, u := range users {
		username := strings.TrimSpace(u.Username)
		if username == "" {
			return fmt.Errorf("admin user at index %d has an empty username", i)
		}
		if seen[username] {
			return fmt.Errorf("duplicate admin username %q", username)
		}
		seen[username] = true

		if strings.TrimSpace(u.PasswordHash) == "" {
			return fmt.Errorf("admin user %q has an empty password_hash", username)
		}
		if _, err := bcrypt.Cost([]byte(u.PasswordHash)); err != nil {
			return fmt.Errorf("admin user %q has an invalid bcrypt password_hash: %w", username, err)
		}
	}

	return nil
}

type adminAuthenticator struct {
	users map[string][]byte
}

func newAdminAuthenticator(users []AdminUser) (*adminAuthenticator, error) {
	if err := validateAdminUsers(users); err != nil {
		return nil, err
	}

	m := make(map[string][]byte, len(users))
	for _, u := range users {
		username := strings.TrimSpace(u.Username)
		m[username] = []byte(u.PasswordHash)
	}

	return &adminAuthenticator{users: m}, nil
}

func (a *adminAuthenticator) Authenticate(username, password string) bool {
	if a == nil {
		return false
	}
	username = strings.TrimSpace(username)
	hash, ok := a.users[username]
	if !ok || username == "" || password == "" {
		_ = bcrypt.CompareHashAndPassword(dummyBcryptHash, []byte(password))
		return false
	}

	return bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil
}

func (s *Server) isAuthEnabled() bool {
	return s.auth != nil
}

func (s *Server) authenticatedAdmin(r *http.Request) string {
	if s.sessionManager == nil {
		return ""
	}
	return s.sessionManager.GetString(r.Context(), sessionKeyAdminUser)
}

func wantsJSON(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	accept := r.Header.Get("Accept")
	return strings.Contains(ct, "application/json") || strings.Contains(accept, "application/json")
}

// requireAdminPage protects HTML and preview routes under /admin.
func (s *Server) requireAdminPage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.isAuthEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		if s.authenticatedAdmin(r) == "" {
			if r.Method == http.MethodGet && !wantsJSON(r) {
				http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Authentication required",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// requireAdminAPI protects mutating and admin-only JSON API routes under /api.
func (s *Server) requireAdminAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.isAuthEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		if s.authenticatedAdmin(r) == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Authentication required",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleAdminLoginPage renders the HTML login form for /admin/login.
func (s *Server) handleAdminLoginPage(w http.ResponseWriter, r *http.Request) {
	if !s.isAuthEnabled() || s.authenticatedAdmin(r) != "" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	data := map[string]any{
		"Title": "Iniciar Sesión",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.renderer.RenderAdminLogin(w, data); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render login page: %v", err), http.StatusInternalServerError)
	}
}

// handleAdminLogin authenticates an admin user and creates a persistent session.
func (s *Server) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	if !s.isAuthEnabled() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Admin authentication is not configured",
		})
		return
	}

	var username, password string
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid JSON payload",
			})
			return
		}
		username = creds.Username
		password = creds.Password
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}
		username = r.FormValue("username")
		password = r.FormValue("password")
	}

	username = strings.TrimSpace(username)
	if !s.auth.Authenticate(username, password) {
		if wantsJSON(r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid username or password",
			})
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		_ = s.renderer.RenderAdminLogin(w, map[string]any{
			"Title":    "Iniciar Sesión",
			"Username": username,
			"Error":    "Usuario o contraseña incorrectos.",
		})
		return
	}

	if err := s.sessionManager.RenewToken(r.Context()); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to renew session token",
		})
		return
	}

	s.sessionManager.Put(r.Context(), sessionKeyAdminUser, username)

	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "authenticated",
			"username": username,
		})
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// handleAdminLogout destroys the current admin session.
func (s *Server) handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	if s.sessionManager != nil {
		if err := s.sessionManager.Destroy(r.Context()); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to destroy session",
			})
			return
		}
	}

	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "logged_out",
		})
		return
	}

	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}
