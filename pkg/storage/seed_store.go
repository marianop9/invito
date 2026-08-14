package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"invitation/pkg/domain"
)

// MemoryStore provides in-memory invitation and RSVP storage initialized with seed files.
type MemoryStore struct {
	mu          sync.RWMutex
	invitations map[string]*domain.Invitation
	rsvps       map[string][]domain.RSVPSubmission
}

// NewMemoryStore initializes a memory store by scanning a seed directory.
func NewMemoryStore(seedDir string) (*MemoryStore, error) {
	store := &MemoryStore{
		invitations: make(map[string]*domain.Invitation),
		rsvps:       make(map[string][]domain.RSVPSubmission),
	}

	files, err := filepath.Glob(filepath.Join(seedDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read seed directory: %w", err)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read seed file %s: %w", file, err)
		}

		var inv domain.Invitation
		if err := json.Unmarshal(data, &inv); err != nil {
			return nil, fmt.Errorf("failed to unmarshal seed file %s: %w", file, err)
		}

		if err := inv.Validate(); err != nil {
			return nil, fmt.Errorf("seed invitation %s validation failed: %w", file, err)
		}

		store.invitations[inv.Slug] = &inv
	}

	return store, nil
}

// GetInvitation retrieves an invitation by slug.
func (s *MemoryStore) GetInvitation(slug string) (*domain.Invitation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	inv, exists := s.invitations[slug]
	if !exists {
		return nil, fmt.Errorf("invitation with slug %q not found", slug)
	}
	return inv, nil
}

// ListInvitations returns all registered invitations.
func (s *MemoryStore) ListInvitations() []*domain.Invitation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*domain.Invitation, 0, len(s.invitations))
	for _, inv := range s.invitations {
		list = append(list, inv)
	}
	return list
}

// SaveRSVP records an RSVP submission for an event.
func (s *MemoryStore) SaveRSVP(slug string, sub domain.RSVPSubmission) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.invitations[slug]; !exists {
		return fmt.Errorf("invitation with slug %q not found", slug)
	}

	if sub.SubmittedAt.IsZero() {
		sub.SubmittedAt = time.Now().UTC()
	}

	s.rsvps[slug] = append(s.rsvps[slug], sub)
	return nil
}

// ListRSVPs returns all RSVP submissions recorded for an invitation.
func (s *MemoryStore) ListRSVPs(slug string) []domain.RSVPSubmission {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.rsvps[slug]
}
