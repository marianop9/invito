package server

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestServerEndpoints(t *testing.T) {
	tempUploads, err := os.MkdirTemp("", "invito-server-uploads-*")
	if err != nil {
		t.Fatalf("failed to create temp uploads dir: %v", err)
	}
	defer os.RemoveAll(tempUploads)

	srv, err := New(Config{
		Port:       "8080",
		SeedDir:    "../../seed",
		UploadsDir: tempUploads,
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

	// Sample 1x1 PNG bytes for upload testing
	validPNGBytes := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
		0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}

	var uploadedURL string

	t.Run("POST /api/upload uploads valid image and returns 201 with URL", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("image", "avatar.png")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		if _, err := io.Copy(part, bytes.NewReader(validPNGBytes)); err != nil {
			t.Fatalf("failed to copy file bytes: %v", err)
		}

		if err := writer.WriteField("slug", "test-wedding"); err != nil {
			t.Fatalf("failed to write slug field: %v", err)
		}
		_ = writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode upload response: %v", err)
		}

		uploadedURL = resp["url"].(string)
		if !strings.HasPrefix(uploadedURL, "/uploads/test-wedding-") || !strings.HasSuffix(uploadedURL, ".png") {
			t.Errorf("unexpected uploaded URL: %s", uploadedURL)
		}
		if resp["content_type"] != "image/png" {
			t.Errorf("expected content_type image/png, got %v", resp["content_type"])
		}
	})

	t.Run("GET /uploads/{filename} serves uploaded asset with correct headers", func(t *testing.T) {
		if uploadedURL == "" {
			t.Skip("skipping because upload test did not populate uploadedURL")
		}

		req := httptest.NewRequest(http.MethodGet, uploadedURL, nil)
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		if !strings.Contains(rec.Header().Get("Content-Type"), "image/png") {
			t.Errorf("expected Content-Type image/png, got %s", rec.Header().Get("Content-Type"))
		}
		if rec.Header().Get("Cache-Control") == "" {
			t.Errorf("expected Cache-Control header")
		}
		if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Errorf("expected X-Content-Type-Options: nosniff")
		}
		if !bytes.Equal(rec.Body.Bytes(), validPNGBytes) {
			t.Errorf("served body does not match uploaded bytes")
		}
	})

	t.Run("POST /api/upload rejects non-image file with 415", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("image", "document.txt")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		_, _ = part.Write([]byte("Not an image! Plain text content."))
		_ = writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusUnsupportedMediaType {
			t.Errorf("expected status 415, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("POST /api/upload rejects missing image field with 400", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("slug", "no-image")
		_ = writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("GET /uploads/ directory request returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/uploads/", nil)
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404 for directory listing, got %d", rec.Code)
		}
	})

	t.Run("GET /uploads/nonexistent.png returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/uploads/nonexistent.png", nil)
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("GET /admin renders events management overview", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Willowbrook Botanical Estate") && !strings.Contains(body, "sarah-and-alex-wedding") {
			t.Errorf("expected wedding event in admin overview")
		}
	})

	t.Run("GET /admin/invitations/{slug}/rsvps renders RSVP tracking dashboard", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/invitations/sarah-and-alex-wedding/rsvps", nil)
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Total Responses") || !strings.Contains(body, "Attending Guests") {
			t.Errorf("expected KPI cards in body, got %s", body)
		}
	})

	t.Run("GET /admin/invitations/{slug}/rsvps.csv exports RFC 4180 CSV", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/invitations/sarah-and-alex-wedding/rsvps.csv", nil)
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Header().Get("Content-Type"), "text/csv") {
			t.Errorf("expected Content-Type text/csv, got %s", rec.Header().Get("Content-Type"))
		}
		if !strings.Contains(rec.Body.String(), "Name,Email,Attending,Guests") {
			t.Errorf("expected CSV header row in response")
		}
	})

	t.Run("GET /admin/invitations/new renders invitation creator form", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/invitations/new", nil)
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "New Invitation") || !strings.Contains(body, "field-title") {
			t.Errorf("expected new invitation builder form in body")
		}
		if !strings.Contains(body, "admin_editor.js") {
			t.Errorf("expected admin_editor.js script tag in body")
		}
	})

	t.Run("GET /admin/invitations/{slug}/edit renders editor with existing event", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/invitations/sarah-and-alex-wedding/edit", nil)
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Sarah &amp; Alex") && !strings.Contains(body, "Sarah & Alex") {
			t.Errorf("expected event title in editor body")
		}
		if !strings.Contains(body, "Willowbrook Botanical Estate") {
			t.Errorf("expected venue name in editor body")
		}
	})

	t.Run("GET /admin/invitations/{slug}/edit returns 404 for nonexistent event", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/invitations/nonexistent-event-slug/edit", nil)
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404 for nonexistent event edit, got %d", rec.Code)
		}
	})

	t.Run("POST /admin/invitations/preview renders live in-memory HTML", func(t *testing.T) {
		payload := map[string]any{
			"slug":        "preview-test",
			"title":       "Golden Jubilee Celebration",
			"description": "Celebrating 50 years of excellence.",
			"date_start":  "2026-10-15T18:00:00Z",
			"location": map[string]string{
				"name":    "Grand Ballroom",
				"address": "100 Milestone Way, Boston, MA",
			},
			"theme": map[string]string{
				"id": "golden-sunset",
			},
			"sections": []map[string]any{
				{
					"type":            "hero",
					"layout":          "full-bleed",
					"eyebrow":         "Golden Anniversary",
					"show_countdown":  true,
				},
				{
					"type": "quote",
					"text": "Half a century of cherished moments.",
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/admin/invitations/preview", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Header().Get("Content-Type"), "text/html") {
			t.Errorf("expected Content-Type text/html, got %s", rec.Header().Get("Content-Type"))
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Golden Jubilee Celebration") {
			t.Errorf("expected preview body to contain title")
		}
		if !strings.Contains(body, "Grand Ballroom") {
			t.Errorf("expected preview body to contain venue name")
		}
	})

	t.Run("POST /api/invitations creates new invitation and persists to SQLite", func(t *testing.T) {
		payload := map[string]any{
			"version":     "1.0",
			"slug":        "spring-gala-2026",
			"title":       "Spring Charity Gala",
			"description": "An evening benefiting the arts foundation.",
			"date_start":  "2026-05-12T19:00:00Z",
			"timezone":    "America/New_York",
			"location": map[string]string{
				"name":    "Metropolitan Opera Hall",
				"address": "30 Lincoln Center Plaza, New York, NY",
			},
			"theme": map[string]string{
				"id": "midnight-soiree",
			},
			"sections": []map[string]any{
				{
					"type":           "hero",
					"layout":         "full-bleed",
					"eyebrow":        "Annual Benefit",
					"show_countdown": true,
				},
				{
					"type":    "rsvp",
					"enabled": true,
					"max_party_size": 2,
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/invitations", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201 Created, got %d: %s", rec.Code, rec.Body.String())
		}

		// Verify retrieval
		getReq := httptest.NewRequest(http.MethodGet, "/api/invitations/spring-gala-2026", nil)
		getRec := httptest.NewRecorder()
		srv.Router().ServeHTTP(getRec, getReq)

		if getRec.Code != http.StatusOK {
			t.Errorf("expected status 200 on GET newly created invitation, got %d", getRec.Code)
		}
	})

	t.Run("POST /api/invitations rejects duplicate slug with 409 Conflict", func(t *testing.T) {
		payload := map[string]any{
			"version":    "1.0",
			"slug":       "spring-gala-2026",
			"title":      "Duplicate Spring Gala",
			"date_start": "2026-05-12T19:00:00Z",
			"location": map[string]string{
				"name":    "Another Hall",
				"address": "123 Main St",
			},
			"theme": map[string]string{"id": "botanical-elegance"},
		}
		bodyBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/invitations", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Errorf("expected status 409 Conflict for duplicate slug, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "already exists") {
			t.Errorf("expected conflict error message, got: %s", rec.Body.String())
		}
	})

	t.Run("POST /api/invitations rejects invalid invitation with 400 Bad Request", func(t *testing.T) {
		payload := map[string]any{
			"slug":  "invalid-event",
			"title": "", // Missing title
		}
		bodyBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/invitations", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 Bad Request, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "title is required") {
			t.Errorf("expected title is required in validation message, got: %s", rec.Body.String())
		}
	})

	t.Run("PUT /api/invitations/{slug} updates existing invitation", func(t *testing.T) {
		payload := map[string]any{
			"version":     "1.0",
			"slug":        "spring-gala-2026",
			"title":       "Spring Charity Gala — Updated Edition",
			"description": "An updated evening program.",
			"date_start":  "2026-05-12T19:30:00Z",
			"timezone":    "America/New_York",
			"location": map[string]string{
				"name":    "Metropolitan Opera Grand Ballroom",
				"address": "30 Lincoln Center Plaza, New York, NY",
			},
			"theme": map[string]string{
				"id": "modern-minimal",
			},
			"sections": []map[string]any{
				{
					"type":           "hero",
					"layout":         "banner",
					"eyebrow":        "Updated Benefit",
					"show_countdown": false,
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPut, "/api/invitations/spring-gala-2026", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK on update, got %d: %s", rec.Code, rec.Body.String())
		}

		// Verify update reflected in GET
		getReq := httptest.NewRequest(http.MethodGet, "/api/invitations/spring-gala-2026", nil)
		getRec := httptest.NewRecorder()
		srv.Router().ServeHTTP(getRec, getReq)

		if !strings.Contains(getRec.Body.String(), "Spring Charity Gala — Updated Edition") {
			t.Errorf("expected updated title in body, got: %s", getRec.Body.String())
		}
	})

	t.Run("PUT /api/invitations/{slug} returns 404 for nonexistent event", func(t *testing.T) {
		payload := map[string]any{
			"title": "Nonexistent",
		}
		bodyBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPut, "/api/invitations/does-not-exist-slug", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404 for nonexistent update, got %d", rec.Code)
		}
	})

	t.Run("DELETE /api/invitations/{slug} deletes invitation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/invitations/spring-gala-2026", nil)
		rec := httptest.NewRecorder()

		srv.Router().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK on delete, got %d: %s", rec.Code, rec.Body.String())
		}

		// Verify subsequent GET returns 404
		getReq := httptest.NewRequest(http.MethodGet, "/api/invitations/spring-gala-2026", nil)
		getRec := httptest.NewRecorder()
		srv.Router().ServeHTTP(getRec, getReq)

		if getRec.Code != http.StatusNotFound {
			t.Errorf("expected status 404 after deletion, got %d", getRec.Code)
		}
	})
}
