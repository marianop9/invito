package storage

import (
	"strings"
	"testing"
	"time"

	"invitation/pkg/domain"
)

func createSampleInvitation(slug, title string) *domain.Invitation {
	start := time.Now().Add(48 * time.Hour).UTC().Truncate(time.Second)
	end := start.Add(4 * time.Hour)

	return &domain.Invitation{
		Version:     "1.0",
		Slug:        slug,
		Title:       title,
		Description: "A celebration to remember",
		DateStart:   start,
		DateEnd:     &end,
		Location: domain.Location{
			Name:    "Grand Ballroom",
			Address: "123 Elegance Blvd",
		},
		Theme: domain.ThemeConfig{
			ID: domain.ThemeBotanicalElegance,
		},
		Sections: []domain.Section{
			&domain.HeroSection{
				SectionType: domain.SectionHero,
				Eyebrow:     "Celebration",
				Vibe:        "Sophisticated elegance",
			},
			&domain.RSVPSection{
				SectionType:  domain.SectionRSVP,
				Enabled:      true,
				MaxPartySize: 4,
			},
		},
	}
}

func TestSQLiteStore_InitializationAndSeeding(t *testing.T) {
	// Initialize with seed directory
	store, err := NewSQLiteStore(":memory:", "../../seed")
	if err != nil {
		t.Fatalf("failed to initialize sqlite store with seeds: %v", err)
	}
	defer store.Close()

	// List seeded invitations
	invs, err := store.ListInvitations()
	if err != nil {
		t.Fatalf("ListInvitations failed: %v", err)
	}
	if len(invs) < 2 {
		t.Fatalf("expected at least 2 seeded invitations, got %d", len(invs))
	}

	// Verify Sarah & Alex exists
	wedding, err := store.GetInvitation("sarah-and-alex-wedding")
	if err != nil {
		t.Fatalf("failed to get wedding invitation: %v", err)
	}
	if wedding.Title != "Sarah & Alex" {
		t.Errorf("expected title 'Sarah & Alex', got %q", wedding.Title)
	}
	if wedding.Location.Name != "Willowbrook Botanical Estate" {
		t.Errorf("expected venue name Willowbrook Botanical Estate, got %q", wedding.Location.Name)
	}

	// Verify nonexistent returns error
	_, err = store.GetInvitation("non-existent-invitation")
	if err == nil {
		t.Errorf("expected error getting non-existent invitation, got nil")
	}
}

func TestSQLiteStore_InvitationCRUD(t *testing.T) {
	store, err := NewSQLiteStore(":memory:", "")
	if err != nil {
		t.Fatalf("failed to initialize sqlite store: %v", err)
	}
	defer store.Close()

	// 1. Create
	inv := createSampleInvitation("gala-night", "Annual Spring Gala")
	if err := store.SaveInvitation(inv); err != nil {
		t.Fatalf("SaveInvitation failed: %v", err)
	}

	// 2. Read
	fetched, err := store.GetInvitation("gala-night")
	if err != nil {
		t.Fatalf("GetInvitation failed: %v", err)
	}
	if fetched.Title != "Annual Spring Gala" {
		t.Errorf("expected title 'Annual Spring Gala', got %q", fetched.Title)
	}
	if fetched.Slug != "gala-night" {
		t.Errorf("expected slug 'gala-night', got %q", fetched.Slug)
	}
	if len(fetched.Sections) != 2 {
		t.Errorf("expected 2 sections, got %d", len(fetched.Sections))
	}

	// 3. Update (Upsert)
	inv.Title = "Updated Annual Spring Gala"
	inv.Description = "Updated gala description"
	if err := store.SaveInvitation(inv); err != nil {
		t.Fatalf("SaveInvitation (update) failed: %v", err)
	}

	updated, err := store.GetInvitation("gala-night")
	if err != nil {
		t.Fatalf("GetInvitation after update failed: %v", err)
	}
	if updated.Title != "Updated Annual Spring Gala" {
		t.Errorf("expected updated title, got %q", updated.Title)
	}
	if updated.Description != "Updated gala description" {
		t.Errorf("expected updated description, got %q", updated.Description)
	}

	// 4. Delete
	if err := store.DeleteInvitation("gala-night"); err != nil {
		t.Fatalf("DeleteInvitation failed: %v", err)
	}

	// 5. Verify deleted
	_, err = store.GetInvitation("gala-night")
	if err == nil {
		t.Errorf("expected error getting deleted invitation, got nil")
	}

	// 6. Delete already deleted should error
	if err := store.DeleteInvitation("gala-night"); err == nil {
		t.Errorf("expected error deleting non-existent invitation, got nil")
	}
}

func TestSQLiteStore_RSVPSubmissionsAndStats(t *testing.T) {
	store, err := NewSQLiteStore(":memory:", "")
	if err != nil {
		t.Fatalf("failed to initialize sqlite store: %v", err)
	}
	defer store.Close()

	inv := createSampleInvitation("birthday-bash", "Leo's 30th Birthday")
	if err := store.SaveInvitation(inv); err != nil {
		t.Fatalf("SaveInvitation failed: %v", err)
	}

	// Submit RSVPs
	rsvps := []domain.RSVPSubmission{
		{
			Name:         "Alice Martin",
			Email:        "alice@example.com",
			Attending:    true,
			GuestCount:   2,
			DietaryNeeds: "Gluten Free",
			SongRequest:  "Dancing Queen",
			SubmittedAt:  time.Now().Add(-2 * time.Hour),
		},
		{
			Name:         "Bob Vance",
			Email:        "bob@example.com",
			Attending:    true,
			GuestCount:   1,
			DietaryNeeds: "Vegetarian",
			SubmittedAt:  time.Now().Add(-1 * time.Hour),
		},
		{
			Name:            "Charlie Brown",
			Email:           "charlie@example.com",
			Attending:       false,
			GuestCount:      1,
			PersonalMessage: "Sorry I will be traveling!",
			SubmittedAt:     time.Now(),
		},
	}

	for _, sub := range rsvps {
		if err := store.SaveRSVP("birthday-bash", sub); err != nil {
			t.Fatalf("SaveRSVP failed for %s: %v", sub.Name, err)
		}
	}

	// List RSVPs (should be ordered newest first)
	list, err := store.ListRSVPs("birthday-bash")
	if err != nil {
		t.Fatalf("ListRSVPs failed: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 RSVPs, got %d", len(list))
	}
	if list[0].Name != "Charlie Brown" {
		t.Errorf("expected Charlie Brown to be first (newest), got %s", list[0].Name)
	}
	if list[1].Name != "Bob Vance" {
		t.Errorf("expected Bob Vance to be second, got %s", list[1].Name)
	}
	if list[2].Name != "Alice Martin" {
		t.Errorf("expected Alice Martin to be third, got %s", list[2].Name)
	}

	// Test GetRSVPStats
	stats, err := store.GetRSVPStats("birthday-bash")
	if err != nil {
		t.Fatalf("GetRSVPStats failed: %v", err)
	}

	if stats.TotalResponses != 3 {
		t.Errorf("expected TotalResponses 3, got %d", stats.TotalResponses)
	}
	if stats.AttendingCount != 2 {
		t.Errorf("expected AttendingCount 2, got %d", stats.AttendingCount)
	}
	if stats.DeclinedCount != 1 {
		t.Errorf("expected DeclinedCount 1, got %d", stats.DeclinedCount)
	}
	if stats.TotalGuests != 3 { // Alice (2) + Bob (1)
		t.Errorf("expected TotalGuests 3, got %d", stats.TotalGuests)
	}
	if stats.DietaryCount != 2 { // Alice (Gluten Free) + Bob (Vegetarian)
		t.Errorf("expected DietaryCount 2, got %d", stats.DietaryCount)
	}
}

func TestSQLiteStore_CascadeDeletion(t *testing.T) {
	store, err := NewSQLiteStore(":memory:", "")
	if err != nil {
		t.Fatalf("failed to initialize sqlite store: %v", err)
	}
	defer store.Close()

	inv := createSampleInvitation("cascade-event", "Cascade Test Event")
	if err := store.SaveInvitation(inv); err != nil {
		t.Fatalf("SaveInvitation failed: %v", err)
	}

	sub := domain.RSVPSubmission{
		Name:      "Guest One",
		Email:     "guest@example.com",
		Attending: true,
	}
	if err := store.SaveRSVP("cascade-event", sub); err != nil {
		t.Fatalf("SaveRSVP failed: %v", err)
	}

	// Verify RSVP is there
	list, err := store.ListRSVPs("cascade-event")
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 RSVP before deletion, got %d (err: %v)", len(list), err)
	}

	// Delete invitation
	if err := store.DeleteInvitation("cascade-event"); err != nil {
		t.Fatalf("DeleteInvitation failed: %v", err)
	}

	// Verify RSVPs were cascade-deleted
	rsvpsAfter, err := store.ListRSVPs("cascade-event")
	if err != nil {
		t.Fatalf("ListRSVPs failed after invitation deletion: %v", err)
	}
	if len(rsvpsAfter) != 0 {
		t.Errorf("expected 0 RSVPs after cascade delete, got %d", len(rsvpsAfter))
	}
}

func TestSQLiteStore_ValidationAndErrors(t *testing.T) {
	store, err := NewSQLiteStore(":memory:", "")
	if err != nil {
		t.Fatalf("failed to initialize sqlite store: %v", err)
	}
	defer store.Close()

	// 1. Nil invitation
	if err := store.SaveInvitation(nil); err == nil {
		t.Errorf("expected error saving nil invitation, got nil")
	}

	// 2. Invalid invitation (missing required fields)
	invalidInv := &domain.Invitation{}
	if err := store.SaveInvitation(invalidInv); err == nil {
		t.Errorf("expected error saving invalid invitation, got nil")
	}

	// 3. RSVP to non-existent invitation
	sub := domain.RSVPSubmission{
		Name:      "Ghost Guest",
		Email:     "ghost@example.com",
		Attending: true,
	}
	err = store.SaveRSVP("non-existent-slug", sub)
	if err == nil {
		t.Errorf("expected error saving RSVP to non-existent invitation, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error message, got %q", err.Error())
	}
}
