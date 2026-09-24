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
}
