package storage

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"invitation/pkg/domain"
	"invitation/pkg/storage/db"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// SQLiteStore implements the Store interface using SQLite.
type SQLiteStore struct {
	db      *sql.DB
	queries *db.Queries
}

// NewSQLiteStore initializes a SQLite database at the specified path and runs migrations.
// If dbPath is empty or ":memory:", an in-memory database is opened.
// If seedDir is non-empty and the database contains zero invitations, seed files are automatically imported.
func NewSQLiteStore(dbPath, seedDir string) (*SQLiteStore, error) {
	isMemory := dbPath == "" || dbPath == ":memory:"
	var dsn string
	if isMemory {
		dsn = "file::memory:?cache=shared&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	} else {
		sep := "?"
		if strings.Contains(dbPath, "?") {
			sep = "&"
		}
		dsn = fmt.Sprintf("%s%s_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)", dbPath, sep)
	}

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database %q: %w", dbPath, err)
	}

	if isMemory {
		sqlDB.SetMaxOpenConns(1)
	}

	store := &SQLiteStore{
		db:      sqlDB,
		queries: db.New(sqlDB),
	}
	if err := store.migrate(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	if seedDir != "" {
		// if err := store.seedIfEmpty(seedDir); err != nil {
		// 	// Log seed warning but allow store initialization to continue
		// 	fmt.Printf("⚠️ Warning: Failed to seed invitations from %s: %v\n", seedDir, err)
		// }
	}

	return store, nil
}

func (s *SQLiteStore) migrate() error {
	_, err := s.db.Exec(schemaSQL)
	return err
}

func (s *SQLiteStore) seedIfEmpty(seedDir string) error {
	count, err := s.queries.CountInvitations(context.Background())
	if err != nil {
		return fmt.Errorf("failed to query invitation count: %w", err)
	}
	if count > 0 {
		return nil
	}

	files, err := filepath.Glob(filepath.Join(seedDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to scan seed directory: %w", err)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read seed file %s: %w", file, err)
		}

		var inv domain.Invitation
		if err := json.Unmarshal(data, &inv); err != nil {
			return fmt.Errorf("failed to unmarshal seed file %s: %w", file, err)
		}

		// Skip demo templates from polluting SQLite database
		if inv.IsDemo() {
			continue
		}

		if err := inv.Validate(); err != nil {
			return fmt.Errorf("seed invitation %s validation failed: %w", file, err)
		}

		if err := s.SaveInvitation(&inv); err != nil {
			return fmt.Errorf("failed to save seed invitation %s: %w", file, err)
		}
	}

	return nil
}

// GetInvitation retrieves an invitation by slug.
func (s *SQLiteStore) GetInvitation(slug string) (*domain.Invitation, error) {
	dataJSON, err := s.queries.GetInvitationJSON(context.Background(), slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invitation with slug %q not found", slug)
		}
		return nil, fmt.Errorf("failed to query invitation: %w", err)
	}

	var inv domain.Invitation
	if err := json.Unmarshal([]byte(dataJSON), &inv); err != nil {
		return nil, fmt.Errorf("failed to parse stored invitation JSON: %w", err)
	}

	return &inv, nil
}

// ListInvitations returns all invitations ordered by creation date.
func (s *SQLiteStore) ListInvitations() ([]*domain.Invitation, error) {
	rows, err := s.queries.ListInvitationsJSON(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}

	var list []*domain.Invitation
	for _, dataJSON := range rows {
		var inv domain.Invitation
		if err := json.Unmarshal([]byte(dataJSON), &inv); err != nil {
			return nil, fmt.Errorf("failed to unmarshal invitation JSON: %w", err)
		}

		list = append(list, &inv)
	}

	return list, nil
}

// SaveInvitation creates or updates an invitation record.
func (s *SQLiteStore) SaveInvitation(inv *domain.Invitation) error {
	if inv == nil {
		return fmt.Errorf("invitation cannot be nil")
	}

	if err := inv.Validate(); err != nil {
		return fmt.Errorf("invitation validation failed: %w", err)
	}

	dataJSON, err := json.Marshal(inv)
	if err != nil {
		return fmt.Errorf("failed to marshal invitation: %w", err)
	}

	now := time.Now().UTC()
	err = s.queries.SaveInvitation(context.Background(), db.SaveInvitationParams{
		Slug:        inv.Slug,
		Title:       inv.Title,
		Description: inv.Description,
		DataJson:    string(dataJSON),
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return fmt.Errorf("failed to save invitation %s: %w", inv.Slug, err)
	}

	return nil
}

// DeleteInvitation removes an invitation and its associated RSVPs (via CASCADE).
func (s *SQLiteStore) DeleteInvitation(slug string) error {
	affected, err := s.queries.DeleteInvitation(context.Background(), slug)
	if err != nil {
		return fmt.Errorf("failed to delete invitation %s: %w", slug, err)
	}

	if affected == 0 {
		return fmt.Errorf("invitation with slug %q not found", slug)
	}

	return nil
}

// SaveRSVP records an RSVP submission for an invitation.
func (s *SQLiteStore) SaveRSVP(slug string, sub domain.RSVPSubmission) error {
	ctx := context.Background()

	// Verify invitation exists
	if _, err := s.queries.InvitationExists(ctx, slug); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("invitation with slug %q not found", slug)
		}
		return fmt.Errorf("failed to check invitation existence: %w", err)
	}

	if sub.SubmittedAt.IsZero() {
		sub.SubmittedAt = time.Now().UTC()
	}

	err := s.queries.CreateRSVP(ctx, db.CreateRSVPParams{
		InvitationSlug:  slug,
		Name:            sub.Name,
		Email:           sub.Email,
		Attending:       sub.Attending,
		GuestCount:      int64(sub.GuestCount),
		DietaryNeeds:    sql.NullString{String: sub.DietaryNeeds, Valid: sub.DietaryNeeds != ""},
		SongRequest:     sql.NullString{String: sub.SongRequest, Valid: sub.SongRequest != ""},
		PersonalMessage: sql.NullString{String: sub.PersonalMessage, Valid: sub.PersonalMessage != ""},
		SubmittedAt:     sub.SubmittedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to insert RSVP: %w", err)
	}

	return nil
}

// ListRSVPs returns all RSVPs for an invitation, ordered newest first.
func (s *SQLiteStore) ListRSVPs(slug string) ([]domain.RSVPSubmission, error) {
	rows, err := s.queries.ListRSVPs(context.Background(), slug)
	if err != nil {
		return nil, fmt.Errorf("failed to query RSVPs: %w", err)
	}

	var list []domain.RSVPSubmission
	for _, r := range rows {
		sub := domain.RSVPSubmission{
			Name:        r.Name,
			Email:       r.Email,
			Attending:   r.Attending,
			GuestCount:  int(r.GuestCount),
			SubmittedAt: r.SubmittedAt,
		}

		if r.DietaryNeeds.Valid {
			sub.DietaryNeeds = r.DietaryNeeds.String
		}
		if r.SongRequest.Valid {
			sub.SongRequest = r.SongRequest.String
		}
		if r.PersonalMessage.Valid {
			sub.PersonalMessage = r.PersonalMessage.String
		}

		list = append(list, sub)
	}

	return list, nil
}

// GetRSVPStats calculates aggregated metrics for an invitation.
func (s *SQLiteStore) GetRSVPStats(slug string) (RSVPStats, error) {
	row, err := s.queries.GetRSVPStats(context.Background(), slug)
	if err != nil {
		return RSVPStats{}, fmt.Errorf("failed to calculate RSVP stats: %w", err)
	}

	return RSVPStats{
		TotalResponses: int(row.TotalResponses),
		AttendingCount: int(row.AttendingCount),
		DeclinedCount:  int(row.DeclinedCount),
		TotalGuests:    int(row.TotalGuests),
		DietaryCount:   int(row.DietaryCount),
	}, nil
}

// Close closes the underlying SQLite database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
