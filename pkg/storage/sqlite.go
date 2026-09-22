package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"invitation/pkg/domain"

	_ "modernc.org/sqlite"
)

// SQLiteStore implements the Store interface using SQLite.
type SQLiteStore struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewSQLiteStore initializes a SQLite database at the specified path and runs migrations.
// If dbPath is empty or ":memory:", an in-memory database is opened.
// If seedDir is non-empty and the database contains zero invitations, seed files are automatically imported.
func NewSQLiteStore(dbPath, seedDir string) (*SQLiteStore, error) {
	isMemory := dbPath == "" || dbPath == ":memory:"
	dsn := dbPath
	if isMemory {
		dsn = "file::memory:?cache=shared"
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database %q: %w", dbPath, err)
	}

	if isMemory {
		db.SetMaxOpenConns(1)
	} else {
		// WAL mode and busy timeout for concurrent safety
		if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("failed to set WAL mode: %w", err)
		}
		if _, err := db.Exec("PRAGMA busy_timeout = 5000;"); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("failed to set busy_timeout: %w", err)
		}
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	if seedDir != "" {
		if err := store.seedIfEmpty(seedDir); err != nil {
			// Log seed warning but allow store initialization to continue
			fmt.Printf("⚠️ Warning: Failed to seed invitations from %s: %v\n", seedDir, err)
		}
	}

	return store, nil
}

func (s *SQLiteStore) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS invitations (
		slug TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		data_json TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS rsvps (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		invitation_slug TEXT NOT NULL,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		attending BOOLEAN NOT NULL,
		guest_count INTEGER NOT NULL DEFAULT 1,
		dietary_needs TEXT,
		song_request TEXT,
		personal_message TEXT,
		submitted_at DATETIME NOT NULL,
		FOREIGN KEY (invitation_slug) REFERENCES invitations(slug) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_rsvps_invitation_slug ON rsvps(invitation_slug);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStore) seedIfEmpty(seedDir string) error {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM invitations;").Scan(&count)
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
	s.mu.RLock()
	defer s.mu.RUnlock()

	var dataJSON string
	err := s.db.QueryRow("SELECT data_json FROM invitations WHERE slug = ?;", slug).Scan(&dataJSON)
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
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT data_json FROM invitations ORDER BY created_at ASC;")
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}
	defer rows.Close()

	var list []*domain.Invitation
	for rows.Next() {
		var dataJSON string
		if err := rows.Scan(&dataJSON); err != nil {
			return nil, fmt.Errorf("failed to scan invitation row: %w", err)
		}

		var inv domain.Invitation
		if err := json.Unmarshal([]byte(dataJSON), &inv); err != nil {
			return nil, fmt.Errorf("failed to unmarshal invitation JSON: %w", err)
		}

		list = append(list, &inv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating invitation rows: %w", err)
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

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	query := `
	INSERT INTO invitations (slug, title, description, data_json, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(slug) DO UPDATE SET
		title = excluded.title,
		description = excluded.description,
		data_json = excluded.data_json,
		updated_at = excluded.updated_at;
	`

	_, err = s.db.Exec(query, inv.Slug, inv.Title, inv.Description, string(dataJSON), now, now)
	if err != nil {
		return fmt.Errorf("failed to save invitation %s: %w", inv.Slug, err)
	}

	return nil
}

// DeleteInvitation removes an invitation and its associated RSVPs (via CASCADE).
func (s *SQLiteStore) DeleteInvitation(slug string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	res, err := s.db.Exec("DELETE FROM invitations WHERE slug = ?;", slug)
	if err != nil {
		return fmt.Errorf("failed to delete invitation %s: %w", slug, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("invitation with slug %q not found", slug)
	}

	return nil
}

// SaveRSVP records an RSVP submission for an invitation.
func (s *SQLiteStore) SaveRSVP(slug string, sub domain.RSVPSubmission) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify invitation exists
	var exists bool
	err := s.db.QueryRow("SELECT 1 FROM invitations WHERE slug = ?;", slug).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("invitation with slug %q not found", slug)
		}
		return fmt.Errorf("failed to check invitation existence: %w", err)
	}

	if sub.SubmittedAt.IsZero() {
		sub.SubmittedAt = time.Now().UTC()
	}

	query := `
	INSERT INTO rsvps (
		invitation_slug, name, email, attending, guest_count, dietary_needs, song_request, personal_message, submitted_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err = s.db.Exec(query,
		slug,
		sub.Name,
		sub.Email,
		sub.Attending,
		sub.GuestCount,
		sub.DietaryNeeds,
		sub.SongRequest,
		sub.PersonalMessage,
		sub.SubmittedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert RSVP: %w", err)
	}

	return nil
}

// ListRSVPs returns all RSVPs for an invitation, ordered newest first.
func (s *SQLiteStore) ListRSVPs(slug string) ([]domain.RSVPSubmission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT name, email, attending, guest_count, dietary_needs, song_request, personal_message, submitted_at
	FROM rsvps
	WHERE invitation_slug = ?
	ORDER BY submitted_at DESC;
	`
	rows, err := s.db.Query(query, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to query RSVPs: %w", err)
	}
	defer rows.Close()

	var list []domain.RSVPSubmission
	for rows.Next() {
		var sub domain.RSVPSubmission
		var dietaryNeeds, songRequest, personalMessage sql.NullString

		err := rows.Scan(
			&sub.Name,
			&sub.Email,
			&sub.Attending,
			&sub.GuestCount,
			&dietaryNeeds,
			&songRequest,
			&personalMessage,
			&sub.SubmittedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan RSVP row: %w", err)
		}

		if dietaryNeeds.Valid {
			sub.DietaryNeeds = dietaryNeeds.String
		}
		if songRequest.Valid {
			sub.SongRequest = songRequest.String
		}
		if personalMessage.Valid {
			sub.PersonalMessage = personalMessage.String
		}

		list = append(list, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating RSVP rows: %w", err)
	}

	return list, nil
}

// GetRSVPStats calculates aggregated metrics for an invitation.
func (s *SQLiteStore) GetRSVPStats(slug string) (RSVPStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT 
		COUNT(*),
		COALESCE(SUM(CASE WHEN attending THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN NOT attending THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN attending THEN guest_count ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN dietary_needs IS NOT NULL AND TRIM(dietary_needs) != '' THEN 1 ELSE 0 END), 0)
	FROM rsvps
	WHERE invitation_slug = ?;
	`

	var stats RSVPStats
	err := s.db.QueryRow(query, slug).Scan(
		&stats.TotalResponses,
		&stats.AttendingCount,
		&stats.DeclinedCount,
		&stats.TotalGuests,
		&stats.DietaryCount,
	)
	if err != nil {
		return RSVPStats{}, fmt.Errorf("failed to calculate RSVP stats: %w", err)
	}

	return stats, nil
}

// Close closes the underlying SQLite database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
