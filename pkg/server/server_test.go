package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServerEndpoints(t *testing.T) {
	srv, err := New(Config{
		Port:    "8080",
		SeedDir: "../../seed",
	})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	t.Run("GET / renders showcase page with template links", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "Sarah &amp; Alex") && !strings.Contains(body, "Sarah & Alex") {
			t.Errorf("expected body to contain wedding template title, got: %s", body)
		}
		if !strings.Contains(body, "/i/sarah-and-alex-wedding") {
			t.Errorf("expected body to contain link to wedding invitation")
		}
	})

	t.Run("GET /i/{slug} renders dynamic invitation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/i/sarah-and-alex-wedding", nil)
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "Willowbrook Botanical Estate") {
			t.Errorf("expected venue name in body, got: %s", body)
		}
		if !strings.Contains(body, "theme-botanical-elegance") {
			t.Errorf("expected theme class in body, got: %s", body)
		}
	})

	t.Run("GET /i/{slug}/calendar.ics generates valid iCal", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/i/sarah-and-alex-wedding/calendar.ics", nil)
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if !strings.Contains(contentType, "text/calendar") {
			t.Errorf("expected Content-Type text/calendar, got %s", contentType)
		}

		if !strings.Contains(rec.Body.String(), "BEGIN:VCALENDAR") {
			t.Errorf("expected iCal header in body")
		}
	})

	t.Run("POST /i/{slug}/rsvp processes attendance submission", func(t *testing.T) {
		payload := map[string]any{
			"name":          "Eleanor Vance",
			"email":         "eleanor@example.com",
			"attending":     true,
			"guest_count":   2,
			"dietary_needs": "Vegetarian",
			"song_request":  "Dancing Queen",
		}
		jsonBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/i/sarah-and-alex-wedding/rsvp", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["status"] != "confirmed" {
			t.Errorf("expected status confirmed, got %v", resp["status"])
		}
	})

	t.Run("GET /api/invitations returns JSON list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/invitations", nil)
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var list []map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
			t.Fatalf("failed to decode JSON list: %v", err)
		}

		if len(list) < 2 {
			t.Errorf("expected at least 2 starter templates, got %d", len(list))
		}
	})
}
