# Milestone Plan & Record: Iteration 1 — Unified SQLite Storage

**Status**: ✅ Completed & Verified

---

## 1. Goal & Rationale
Replace the prototype in-memory storage (`MemoryStore`) with a persistent, pure-Go SQLite storage engine (`modernc.org/sqlite`). This ensures that event templates and guest RSVP submissions persist across server restarts, without introducing CGO toolchains or external database requirements.

---

## 2. Components Implemented

### `pkg/storage/store.go`
- Declared the central `Store` interface:
  ```go
  type Store interface {
      GetInvitation(slug string) (*domain.Invitation, error)
      ListInvitations() ([]*domain.Invitation, error)
      SaveInvitation(inv *domain.Invitation) error
      DeleteInvitation(slug string) error
      SaveRSVP(slug string, sub domain.RSVPSubmission) error
      ListRSVPs(slug string) ([]domain.RSVPSubmission, error)
      GetRSVPStats(slug string) (RSVPStats, error)
      Close() error
  }
  ```
- Declared `RSVPStats` struct for real-time aggregation metrics:
  ```go
  type RSVPStats struct {
      TotalResponses int `json:"total_responses"`
      AttendingCount int `json:"attending_count"`
      DeclinedCount  int `json:"declined_count"`
      TotalGuests    int `json:"total_guests"`
      DietaryCount   int `json:"dietary_count"`
  }
  ```

### `pkg/storage/sqlite.go`
- Driver: `modernc.org/sqlite` (pure Go).
- Connection management: WAL mode (`PRAGMA journal_mode = WAL`), busy timeout (`5000ms`), and foreign key enforcement (`PRAGMA foreign_keys = ON`).
- Automatic migrations:
  - Table `invitations`: `slug` (PK), `title`, `description`, `data_json`, `created_at`, `updated_at`.
  - Table `rsvps`: `id` (PK AUTOINCREMENT), `invitation_slug` (FK to invitations ON DELETE CASCADE), `name`, `email`, `attending`, `guest_count`, `dietary_needs`, `song_request`, `personal_message`, `submitted_at`.
  - Index on `rsvps(invitation_slug)`.
- Auto-seeding: If `invitations` table has 0 records on boot, scans `seed/*.json` and imports demo templates.
- Native in-memory support: `NewSQLiteStore(":memory:", ...)` sets `db.SetMaxOpenConns(1)` for unit testing without disk I/O.

### Deprecations
- Removed `pkg/storage/seed_store.go` and `pkg/storage/seed_store_test.go`.

### Server & CLI Integration
- `pkg/server/server.go`: Updated `Server` to use `storage.Store`. Handlers updated to handle query errors.
- `main.go`: Added `-db` CLI flag (default `invito.db`). Initializes `SQLiteStore` and manages graceful closing on shutdown.
- `pkg/ssg/ssg.go`: Updated to use `storage.Store` in-memory.

---

## 3. Automated Test Suite
- Implemented in `pkg/storage/sqlite_test.go`:
  - `TestSQLiteStore_InitializationAndSeeding`: Verifies initial table migration and seed import.
  - `TestSQLiteStore_InvitationCRUD`: Tests insert, read, upsert, and delete operations.
  - `TestSQLiteStore_RSVPSubmissionsAndStats`: Tests RSVP insertion, ordering (newest first), and `GetRSVPStats` aggregation calculations.
  - `TestSQLiteStore_CascadeDeletion`: Verifies that deleting an invitation cascades and deletes all related RSVP rows.
  - `TestSQLiteStore_ValidationAndErrors`: Tests error handling for nil, invalid models, and non-existent event slugs.

---

## 4. Verification Results
```bash
$ go test ./... -v
ok  	invitation/pkg/calendar	0.003s
ok  	invitation/pkg/domain	0.004s
ok  	invitation/pkg/renderer	0.007s
ok  	invitation/pkg/server	0.017s
ok  	invitation/pkg/ssg	0.108s
ok  	invitation/pkg/storage	0.009s
```
Manual persistence verified across server restart with disk file `test_invito.db`.
