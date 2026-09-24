package server

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
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
	UploadsDir   string
	MediaStorage storage.MediaStorage
}

// Server wraps the chi router, template renderer, and storage engine.
type Server struct {
	router       *chi.Mux
	renderer     *renderer.Renderer
	store        storage.Store
	mediaStorage storage.MediaStorage
	config       Config
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

	var mediaStorage storage.MediaStorage
	if cfg.MediaStorage != nil {
		mediaStorage = cfg.MediaStorage
	} else {
		uploadsDir := cfg.UploadsDir
		if uploadsDir == "" {
			uploadsDir = "uploads"
		}
		mediaStorage = storage.NewLocalMediaStorage(uploadsDir, "/uploads")
	}

	s := &Server{
		router:       chi.NewRouter(),
		renderer:     rnd,
		store:        store,
		mediaStorage: mediaStorage,
		config:       cfg,
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
		r.Post("/invitations", s.handleAPICreateInvitation)
		r.Get("/invitations/{slug}", s.handleAPIGetInvitation)
		r.Put("/invitations/{slug}", s.handleAPIUpdateInvitation)
		r.Delete("/invitations/{slug}", s.handleAPIDeleteInvitation)
		r.Get("/invitations/{slug}/rsvps", s.handleAPIListRSVPs)
		r.Post("/upload", s.handleAPIUpload)
	})

	// Uploaded Media Serving Route
	s.router.Get("/uploads/*", s.handleServeUpload)

	// Admin Planner Dashboard Routes
	s.router.Route("/admin", func(r chi.Router) {
		r.Get("/", s.handleAdminIndex)
		r.Get("/invitations/new", s.handleAdminInvitationNew)
		r.Route("/invitations/{slug}", func(r chi.Router) {
			r.Get("/edit", s.handleAdminInvitationEdit)
			r.Get("/rsvps", s.handleAdminRSVPs)
			r.Get("/rsvps.csv", s.handleAdminExportRSVPsCSV)
		})
		r.Post("/invitations/preview", s.handleAdminInvitationPreview)
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

// handleAPICreateInvitation parses, validates, and persists a new invitation document.
func (s *Server) handleAPICreateInvitation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var inv domain.Invitation
	if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Invalid JSON request body: %v", err),
		})
		return
	}

	inv.Slug = strings.TrimSpace(strings.ToLower(inv.Slug))
	if inv.Slug == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "slug is required",
		})
		return
	}

	// TODO: Implement advanced collision resolution (e.g. auto-suffix suggestions)
	if existing, _ := s.store.GetInvitation(inv.Slug); existing != nil {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("An invitation with slug %q already exists.", inv.Slug),
		})
		return
	}

	if err := inv.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := s.store.SaveInvitation(&inv); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Failed to save invitation: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(inv)
}

// handleAPIUpdateInvitation updates an existing invitation document in SQLite.
func (s *Server) handleAPIUpdateInvitation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := chi.URLParam(r, "slug")

	existing, err := s.store.GetInvitation(slug)
	if err != nil || existing == nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Invitation with slug %q not found", slug),
		})
		return
	}

	var inv domain.Invitation
	if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Invalid JSON request body: %v", err),
		})
		return
	}

	// Post-save immutability: enforce slug from URL parameter
	inv.Slug = slug

	if err := inv.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := s.store.SaveInvitation(&inv); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Failed to update invitation: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(inv)
}

// handleAPIDeleteInvitation deletes an invitation and cascades associated RSVPs.
func (s *Server) handleAPIDeleteInvitation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := chi.URLParam(r, "slug")

	if _, err := s.store.GetInvitation(slug); err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Invitation with slug %q not found", slug),
		})
		return
	}

	if err := s.store.DeleteInvitation(slug); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Failed to delete invitation: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "deleted",
		"slug":    slug,
		"message": "Invitation successfully deleted",
	})
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

// handleAPIUpload handles multipart image uploads for invitations.
func (s *Server) handleAPIUpload(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 10 * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+512*1024)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(strings.ToLower(err.Error()), "request body too large") {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "File size exceeds maximum allowed size of 10MB",
			})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Failed to parse multipart form: %v", err),
		})
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Missing 'image' file field in multipart form data",
		})
		return
	}
	defer file.Close()

	slug := r.FormValue("slug")
	if slug == "" {
		slug = r.URL.Query().Get("slug")
	}

	info, err := s.mediaStorage.Save(r.Context(), storage.SaveMediaInput{
		Slug:         slug,
		OriginalName: header.Filename,
		ContentType:  header.Header.Get("Content-Type"),
		Reader:       file,
		Size:         header.Size,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case errors.Is(err, storage.ErrFileTooLarge):
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
		case errors.Is(err, storage.ErrInvalidMediaType):
			w.WriteHeader(http.StatusUnsupportedMediaType)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
		case errors.Is(err, storage.ErrEmptyFile):
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": fmt.Sprintf("Failed to save media: %v", err),
			})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(info)
}

// handleServeUpload safely serves uploaded assets from the media storage directory.
func (s *Server) handleServeUpload(w http.ResponseWriter, r *http.Request) {
	subPath := chi.URLParam(r, "*")
	subPath = filepath.Clean(subPath)
	if subPath == "." || subPath == "/" || strings.HasPrefix(subPath, "..") {
		http.NotFound(w, r)
		return
	}

	localMedia, ok := s.mediaStorage.(*storage.LocalMediaStorage)
	if !ok {
		http.NotFound(w, r)
		return
	}

	fullPath := filepath.Join(localMedia.BaseDir(), subPath)
	stat, err := os.Stat(fullPath)
	if err != nil || stat.IsDir() {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, fullPath)
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

// handleAdminInvitationNew renders the invitation builder pre-populated with starter defaults.
func (s *Server) handleAdminInvitationNew(w http.ResponseWriter, r *http.Request) {
	startDate := time.Now().AddDate(0, 1, 0).Truncate(time.Hour)
	endDate := startDate.Add(6 * time.Hour)

	starterInv := &domain.Invitation{
		Version:     "1.0",
		Slug:        "",
		Title:       "",
		Subtitle:    "",
		Description: "",
		DateStart:   startDate,
		DateEnd:     &endDate,
		Timezone:    "America/New_York",
		Location: domain.Location{
			Name:    "",
			Address: "",
		},
		Theme: domain.ThemeConfig{
			ID: domain.ThemeBotanicalElegance,
		},
		Sections: []domain.Section{
			&domain.HeroSection{
				SectionType:   domain.SectionHero,
				Layout:        "full-bleed",
				Eyebrow:       "Celebration",
				ShowCountdown: true,
			},
			&domain.DetailsSection{
				SectionType:        domain.SectionDetails,
				ShowMapLink:        true,
				ShowCalendarButton: true,
			},
			&domain.QuoteSection{
				SectionType: domain.SectionQuote,
				Text:        "Whatever our souls are made of, his and mine are the same.",
				Author:      "Emily Brontë",
			},
			&domain.TimelineSection{
				SectionType: domain.SectionTimeline,
				Title:       "Schedule of Events",
				Items: []domain.TimelineItem{
					{Time: "4:00 PM", Title: "Welcome & Ceremony", Description: "Main Pavilion"},
					{Time: "6:00 PM", Title: "Dinner & Toasts", Description: "Grand Dining Room"},
					{Time: "8:00 PM", Title: "Music & Dancing", Description: "Courtyard"},
				},
			},
			&domain.DressCodeSection{
				SectionType:  domain.SectionDressCode,
				Title:        "Dress Code",
				Name:         "Garden Formal",
				Description:  "Cocktail attire or suits. Comfortable footwear recommended for outdoors.",
				PaletteHints: []string{"#2A4738", "#D4AF37", "#EFEBE4", "#8A9A86"},
			},
			&domain.RSVPSection{
				SectionType:    domain.SectionRSVP,
				Enabled:        true,
				MaxPartySize:   2,
				AskDietary:     true,
				AskSongRequest: false,
			},
			&domain.ClosingSection{
				SectionType: domain.SectionClosing,
				Message:     "We cannot wait to celebrate our special day with all of you!",
				Signoff:     "With love & gratitude,",
			},
		},
	}

	invJSON, err := json.Marshal(starterInv)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal starter invitation: %v", err), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":          "New Invitation",
		"Invitation":     starterInv,
		"InvitationJSON": string(invJSON),
		"IsNew":          true,
		"CurrentYear":    time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.renderer.RenderAdminEditor(w, data); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render invitation editor: %v", err), http.StatusInternalServerError)
	}
}

// handleAdminInvitationEdit renders the invitation builder populated with an existing invitation.
func (s *Server) handleAdminInvitationEdit(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	inv, err := s.store.GetInvitation(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	invJSON, err := json.Marshal(inv)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal invitation: %v", err), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":          "Edit: " + inv.Title,
		"Invitation":     inv,
		"InvitationJSON": string(invJSON),
		"IsNew":          false,
		"CurrentYear":    time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.renderer.RenderAdminEditor(w, data); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render invitation editor: %v", err), http.StatusInternalServerError)
	}
}

// handleAdminInvitationPreview generates an in-memory preview of an invitation without saving it.
func (s *Server) handleAdminInvitationPreview(w http.ResponseWriter, r *http.Request) {
	var inv domain.Invitation
	if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Sensible defaults for live preview while typing
	if inv.Slug == "" {
		inv.Slug = "preview"
	}
	if inv.Title == "" {
		inv.Title = "Untitled Event"
	}
	if inv.DateStart.IsZero() {
		inv.DateStart = time.Now().AddDate(0, 1, 0)
	}
	if inv.Location.Name == "" {
		inv.Location.Name = "Venue Name"
	}
	if inv.Location.Address == "" {
		inv.Location.Address = "Venue Address"
	}
	if inv.Theme.ID == "" {
		inv.Theme.ID = domain.ThemeBotanicalElegance
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.renderer.RenderInvitation(w, &inv); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render preview: %v", err), http.StatusInternalServerError)
	}
}
