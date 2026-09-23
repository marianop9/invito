package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"invitation/pkg/calendar"
	"invitation/pkg/domain"
	"invitation/pkg/renderer"
	"invitation/pkg/storage"
	"invitation/web"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Config holds configuration options for the HTTP server.
type Config struct {
	Port         string
	SeedDir      string
	DatabasePath string
	Store        storage.Store
}

// Server wraps the chi router, template renderer, and storage engine.
type Server struct {
	router   *chi.Mux
	renderer *renderer.Renderer
	store    storage.Store
	config   Config
}

// New creates and configures a new Server instance.
func New(cfg Config) (*Server, error) {
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.SeedDir == "" {
		cfg.SeedDir = "seed"
	}

	rnd, err := renderer.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize renderer: %w", err)
	}

	var store storage.Store
	if cfg.Store != nil {
		store = cfg.Store
	} else {
		s, err := storage.NewSQLiteStore(cfg.DatabasePath, cfg.SeedDir)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize sqlite store: %w", err)
		}
		store = s
	}

	s := &Server{
		router:   chi.NewRouter(),
		renderer: rnd,
		store:    store,
		config:   cfg,
	}

	s.setupMiddlewares()
	s.setupRoutes()

	return s, nil
}

func (s *Server) setupMiddlewares() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Compress(5))
}

func (s *Server) setupRoutes() {
	// Serve embedded static files at /static/*
	staticFS, err := fs.Sub(web.Files, "static")
	if err == nil {
		s.router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	}

	// Platform Showcase Routes
	s.router.Get("/", s.handleIndex)
	s.router.Get("/health", s.handleHealth)

	// Public Invitation Routes
	s.router.Route("/i/{slug}", func(r chi.Router) {
		r.Get("/", s.handleRenderInvitation)
		r.Get("/calendar.ics", s.handleDownloadICS)
		r.Post("/rsvp", s.handleRSVPSubmit)
	})

	// JSON API Routes
	s.router.Route("/api", func(r chi.Router) {
		r.Get("/invitations", s.handleAPIListInvitations)
		r.Get("/invitations/{slug}", s.handleAPIGetInvitation)
		r.Get("/invitations/{slug}/rsvps", s.handleAPIListRSVPs)
	})

	// Admin Planner Dashboard Routes
	s.router.Route("/admin", func(r chi.Router) {
		r.Get("/", s.handleAdminIndex)
		r.Route("/invitations/{slug}", func(r chi.Router) {
			r.Get("/rsvps", s.handleAdminRSVPs)
			r.Get("/rsvps.csv", s.handleAdminExportRSVPsCSV)
		})
	})
}

// handleIndex renders the showcase platform home page.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	invitations, err := s.store.ListInvitations()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list invitations: %v", err), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Invitations": invitations,
		"CurrentYear": time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.renderer.RenderIndex(w, data); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render index page: %v", err), http.StatusInternalServerError)
	}
}

// handleRenderInvitation dynamically renders an invitation page.
func (s *Server) handleRenderInvitation(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	inv, err := s.store.GetInvitation(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.renderer.RenderInvitation(w, inv); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render invitation: %v", err), http.StatusInternalServerError)
	}
}

// handleDownloadICS generates and serves the RFC 5545 iCalendar file.
func (s *Server) handleDownloadICS(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	inv, err := s.store.GetInvitation(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	icsContent := calendar.GenerateICS(inv)

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.ics\"", inv.Slug))
	_, _ = w.Write([]byte(icsContent))
}

// handleRSVPSubmit processes attendance confirmation submissions (JSON or Form POST).
func (s *Server) handleRSVPSubmit(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	inv, err := s.store.GetInvitation(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if !inv.IsRSVPOpen() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "The RSVP deadline for this event has passed.",
		})
		return
	}

	var sub domain.RSVPSubmission
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid JSON payload",
			})
			return
		}
	} else {
		// Standard form submission
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}
		sub.Name = r.FormValue("name")
		sub.Email = r.FormValue("email")
		sub.Attending = r.FormValue("attending") == "true"
		sub.DietaryNeeds = r.FormValue("dietary_needs")
		sub.SongRequest = r.FormValue("song_request")
		sub.PersonalMessage = r.FormValue("personal_message")
		if sub.Attending {
			sub.GuestCount = 1
		}
	}

	maxParty := 2
	if rsvp := inv.RSVPSection(); rsvp != nil && rsvp.MaxPartySize > 0 {
		maxParty = rsvp.MaxPartySize
	}

	if err := sub.Validate(maxParty); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := s.store.SaveRSVP(slug, sub); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to save RSVP response",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":    "confirmed",
		"attending": sub.Attending,
		"message":   "Your response has been successfully recorded!",
	})
}

// handleAPIListInvitations returns all invitations in JSON format.
func (s *Server) handleAPIListInvitations(w http.ResponseWriter, r *http.Request) {
	invitations, err := s.store.ListInvitations()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list invitations: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invitations)
}

// handleAPIGetInvitation returns a single invitation JSON document.
func (s *Server) handleAPIGetInvitation(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	inv, err := s.store.GetInvitation(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(inv)
}

// handleAPIListRSVPs returns all recorded RSVPs for an invitation.
func (s *Server) handleAPIListRSVPs(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	_, err := s.store.GetInvitation(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	rsvps, err := s.store.ListRSVPs(slug)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list RSVPs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rsvps)
}

// handleHealth returns system status.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	templateCount := 0
	if invs, err := s.store.ListInvitations(); err == nil {
		templateCount = len(invs)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":      "ok",
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"templates":   templateCount,
		"environment": "prototype",
	})
}

// Router returns the configured chi.Mux.
func (s *Server) Router() *chi.Mux {
	return s.router
}

// AdminEventItem bundles an invitation with its aggregated RSVP metrics for the admin overview.
type AdminEventItem struct {
	Invitation *domain.Invitation
	Stats      storage.RSVPStats
}

// handleAdminIndex renders the events management list.
func (s *Server) handleAdminIndex(w http.ResponseWriter, r *http.Request) {
	invitations, err := s.store.ListInvitations()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list invitations: %v", err), http.StatusInternalServerError)
		return
	}

	var events []AdminEventItem
	totalConfirmedGuests := 0

	for _, inv := range invitations {
		stats, err := s.store.GetRSVPStats(inv.Slug)
		if err != nil {
			stats = storage.RSVPStats{}
		}
		events = append(events, AdminEventItem{
			Invitation: inv,
			Stats:      stats,
		})
		totalConfirmedGuests += stats.TotalGuests
	}

	data := map[string]any{
		"Title":                "Events Management",
		"Events":               events,
		"TotalEvents":          len(events),
		"TotalConfirmedGuests": totalConfirmedGuests,
		"CurrentYear":          time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.renderer.RenderAdminIndex(w, data); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render admin index: %v", err), http.StatusInternalServerError)
	}
}

// handleAdminRSVPs displays the RSVP tracking dashboard and guest table for a single event.
func (s *Server) handleAdminRSVPs(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	inv, err := s.store.GetInvitation(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	rsvps, err := s.store.ListRSVPs(slug)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list RSVPs: %v", err), http.StatusInternalServerError)
		return
	}

	stats, err := s.store.GetRSVPStats(slug)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get RSVP stats: %v", err), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":       "RSVPs: " + inv.Title,
		"Invitation":  inv,
		"RSVPs":       rsvps,
		"Stats":       stats,
		"CurrentYear": time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.renderer.RenderAdminRSVPs(w, data); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render admin RSVPs: %v", err), http.StatusInternalServerError)
	}
}

// handleAdminExportRSVPsCSV streams an RFC 4180 CSV export of all guest RSVP responses.
func (s *Server) handleAdminExportRSVPsCSV(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	inv, err := s.store.GetInvitation(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	rsvps, err := s.store.ListRSVPs(slug)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list RSVPs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s-rsvps.csv\"", inv.Slug))

	writer := csv.NewWriter(w)

	header := []string{
		"Name",
		"Email",
		"Attending",
		"Guests",
		"Dietary Needs",
		"Song Request",
		"Personal Message",
		"Submitted At",
	}
	if err := writer.Write(header); err != nil {
		return
	}

	for _, sub := range rsvps {
		attendingStr := "No"
		if sub.Attending {
			attendingStr = "Yes"
		}

		submittedAtStr := ""
		if !sub.SubmittedAt.IsZero() {
			submittedAtStr = sub.SubmittedAt.UTC().Format("2006-01-02 15:04:05 UTC")
		}

		row := []string{
			sub.Name,
			sub.Email,
			attendingStr,
			strconv.Itoa(sub.GuestCount),
			sub.DietaryNeeds,
			sub.SongRequest,
			sub.PersonalMessage,
			submittedAtStr,
		}
		if err := writer.Write(row); err != nil {
			return
		}
	}

	writer.Flush()
}
