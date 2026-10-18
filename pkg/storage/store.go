package storage

import (
	"context"

	"invitation/pkg/domain"
)

// RSVPStats provides aggregated attendance and catering metrics for an invitation.
type RSVPStats struct {
	TotalResponses int `json:"total_responses"`
	AttendingCount int `json:"attending_count"`
	DeclinedCount  int `json:"declined_count"`
	TotalGuests    int `json:"total_guests"`
	DietaryCount   int `json:"dietary_count"`
}

// Store defines the persistent storage interface for invitations and RSVP records.
type Store interface {
	GetInvitation(ctx context.Context, slug string) (*domain.Invitation, error)
	ListInvitations(ctx context.Context) ([]*domain.Invitation, error)
	SaveInvitation(ctx context.Context, inv *domain.Invitation) error
	DeleteInvitation(ctx context.Context, slug string) error
	SaveRSVP(ctx context.Context, slug string, sub domain.RSVPSubmission) error
	ListRSVPs(ctx context.Context, slug string) ([]domain.RSVPSubmission, error)
	GetRSVPStats(ctx context.Context, slug string) (RSVPStats, error)
	Close() error
}
