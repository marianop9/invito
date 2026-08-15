package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

// Validate checks that an Invitation struct contains all required fields and valid configurations.
func (inv *Invitation) Validate() error {
	var errs []string

	if inv.Version == "" {
		inv.Version = "1.0"
	}

	inv.Slug = strings.TrimSpace(strings.ToLower(inv.Slug))
	if inv.Slug == "" {
		errs = append(errs, "slug is required")
	} else if !slugRegex.MatchString(inv.Slug) {
		errs = append(errs, "slug must only contain lowercase letters, numbers, and hyphens")
	}

	if strings.TrimSpace(inv.Title) == "" {
		errs = append(errs, "title is required")
	}

	if inv.DateStart.IsZero() {
		errs = append(errs, "date_start is required and must be a valid ISO 8601 timestamp")
	}

	if inv.DateEnd != nil && !inv.DateEnd.IsZero() {
		if inv.DateEnd.Before(inv.DateStart) {
			errs = append(errs, "date_end cannot be before date_start")
		}
	}

	if strings.TrimSpace(inv.Location.Name) == "" {
		errs = append(errs, "location.name is required")
	}

	if strings.TrimSpace(inv.Location.Address) == "" {
		errs = append(errs, "location.address is required")
	}

	// Validate theme
	switch inv.Theme.ID {
	case ThemeBotanicalElegance, ThemeMidnightSoiree, ThemeModernMinimal, ThemeGoldenSunset:
		// Valid
	case "":
		inv.Theme.ID = ThemeBotanicalElegance // Default fallback
	default:
		errs = append(errs, fmt.Sprintf("unknown theme id: %q", inv.Theme.ID))
	}

	// Validate each section block using its own Section.Validate() contract
	for idx, sec := range inv.Sections {
		if sec == nil {
			errs = append(errs, fmt.Sprintf("section #%d: section block cannot be nil", idx+1))
			continue
		}
		if err := sec.Validate(); err != nil {
			errs = append(errs, fmt.Sprintf("section #%d (%s): %v", idx+1, sec.Type(), err))
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

// ValidateRSVP validates a guest's RSVP submission.
func (sub *RSVPSubmission) Validate(maxPartySize int) error {
	if strings.TrimSpace(sub.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(sub.Email) == "" || !strings.Contains(sub.Email, "@") {
		return errors.New("a valid email is required")
	}
	if sub.Attending {
		if sub.GuestCount < 1 {
			sub.GuestCount = 1
		}
		if maxPartySize > 0 && sub.GuestCount > maxPartySize {
			return fmt.Errorf("guest count (%d) exceeds maximum allowed party size (%d)", sub.GuestCount, maxPartySize)
		}
	} else {
		sub.GuestCount = 0
	}
	return nil
}
