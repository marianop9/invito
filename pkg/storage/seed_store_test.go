package storage

import (
	"testing"
	"time"

	"invitation/pkg/domain"
)

func TestMemoryStore(t *testing.T) {
	store, err := NewMemoryStore("../../seed")
	if err != nil {
		t.Fatalf("failed to create memory store: %v", err)
	}

	invs := store.ListInvitations()
	if len(invs) < 2 {
		t.Fatalf("expected at least 2 invitations, got %d", len(invs))
	}

	inv, err := store.GetInvitation("sarah-and-alex-wedding")
	if err != nil {
		t.Fatalf("expected to find wedding invitation, got error: %v", err)
	}
	if inv.Title == "" {
		t.Errorf("expected non-empty title")
	}

	// Test non-existent slug
	if _, err := store.GetInvitation("non-existent"); err == nil {
		t.Errorf("expected error for non-existent slug, got nil")
	}

	// Test RSVP saving and retrieval
	rsvp := domain.RSVPSubmission{
		Name:        "Test Guest",
		Email:       "guest@example.com",
		Attending:   true,
		GuestCount:  2,
		SubmittedAt: time.Now(),
	}

	if err := store.SaveRSVP("sarah-and-alex-wedding", rsvp); err != nil {
		t.Fatalf("failed to save RSVP: %v", err)
	}

	list := store.ListRSVPs("sarah-and-alex-wedding")
	if len(list) != 1 {
		t.Fatalf("expected 1 RSVP, got %d", len(list))
	}
	if list[0].Name != "Test Guest" {
		t.Errorf("expected guest name 'Test Guest', got %q", list[0].Name)
	}
}
