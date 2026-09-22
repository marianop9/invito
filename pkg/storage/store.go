package storage

import (
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
	GetInvitation(slug string) (*domain.Invitation, error)
	ListInvitations() ([]*domain.Invitation, error)
	SaveInvitation(inv *domain.Invitation) error
	DeleteInvitation(slug string) error
	SaveRSVP(slug string, sub domain.RSVPSubmission) error
	ListRSVPs(slug string) ([]domain.RSVPSubmission, error)
	GetRSVPStats(slug string) (RSVPStats, error)
	Close() error
}
