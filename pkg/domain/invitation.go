package domain

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// ThemeID identifies one of the pre-defined visual themes.
type ThemeID string

const (
	ThemeBotanicalElegance ThemeID = "botanical-elegance"
	ThemeMidnightSoiree    ThemeID = "midnight-soiree"
	ThemeModernMinimal     ThemeID = "modern-minimal"
	ThemeGoldenSunset      ThemeID = "golden-sunset"
)

// PaletteOverride allows customizing key theme color tokens.
type PaletteOverride struct {
	Primary    string `json:"primary,omitempty"`
	Background string `json:"background,omitempty"`
	Text       string `json:"text,omitempty"`
	Accent     string `json:"accent,omitempty"`
	CardBg     string `json:"card_bg,omitempty"`
}

// ThemeConfig defines visual design, typography, and palette options.
type ThemeConfig struct {
	ID              ThemeID          `json:"id"`
	PaletteOverride *PaletteOverride `json:"palette_override,omitempty"`
	FontHeading     string           `json:"font_heading,omitempty"`
	FontBody        string           `json:"font_body,omitempty"`
	CustomCSS       string           `json:"custom_css,omitempty"`
}

// Location describes the physical or virtual venue.
type Location struct {
	Name           string `json:"name"`
	Address        string `json:"address"`
	MapURL         string `json:"map_url,omitempty"`
	DirectionsNote string `json:"directions_note,omitempty"`
}

// HeroSection configures the top introductory card.
type HeroSection struct {
	Badge         string `json:"badge,omitempty"`
	CoverImageURL string `json:"cover_image_url,omitempty"`
	ShowCountdown bool   `json:"show_countdown"`
}

// TimelineItem represents a milestone in the event schedule.
type TimelineItem struct {
	Time        string `json:"time"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// DressCodeSection details the dress code guidelines and color palette suggestions.
type DressCodeSection struct {
	Title        string   `json:"title,omitempty"`
	Description  string   `json:"description,omitempty"`
	PaletteHints []string `json:"palette_hints,omitempty"`
}

// RSVPSection configures the attendance confirmation form.
type RSVPSection struct {
	Enabled        bool       `json:"enabled"`
	Deadline       *time.Time `json:"deadline,omitempty"`
	MaxPartySize   int        `json:"max_party_size"`
	AskDietary     bool       `json:"ask_dietary"`
	AskSongRequest bool       `json:"ask_song_request"`
	CustomNote     string     `json:"custom_note,omitempty"`
}

// FAQItem represents a frequently asked question and answer.
type FAQItem struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// RegistryLink represents a gift registry or donation link.
type RegistryLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// GiftRegistrySection describes gift registry instructions and URLs.
type GiftRegistrySection struct {
	Message string         `json:"message,omitempty"`
	Links   []RegistryLink `json:"links,omitempty"`
}

// Sections holds the modular blocks of an invitation.
type Sections struct {
	Hero         *HeroSection         `json:"hero,omitempty"`
	Timeline     []TimelineItem       `json:"timeline,omitempty"`
	DressCode    *DressCodeSection    `json:"dress_code,omitempty"`
	RSVP         *RSVPSection         `json:"rsvp,omitempty"`
	FAQs         []FAQItem            `json:"faqs,omitempty"`
	GiftRegistry *GiftRegistrySection `json:"gift_registry,omitempty"`
}

// Invitation is the top-level structured event invitation entity.
type Invitation struct {
	Version     string      `json:"version"`
	Slug        string      `json:"slug"`
	Title       string      `json:"title"`
	Subtitle    string      `json:"subtitle,omitempty"`
	Hosts       []string    `json:"hosts,omitempty"`
	Description string      `json:"description,omitempty"`
	DateStart   time.Time   `json:"date_start"`
	DateEnd     *time.Time  `json:"date_end,omitempty"`
	Timezone    string      `json:"timezone"`
	Location    Location    `json:"location"`
	Theme       ThemeConfig `json:"theme"`
	Sections    Sections    `json:"sections"`
}

// HostsDisplay returns a formatted string of hosts.
func (inv *Invitation) HostsDisplay() string {
	if len(inv.Hosts) == 0 {
		return ""
	}
	if len(inv.Hosts) == 1 {
		return inv.Hosts[0]
	}
	if len(inv.Hosts) == 2 {
		return inv.Hosts[0] + " & " + inv.Hosts[1]
	}
	return strings.Join(inv.Hosts[:len(inv.Hosts)-1], ", ") + " & " + inv.Hosts[len(inv.Hosts)-1]
}

// FormattedFullDate returns a human-friendly date string like "Saturday, September 19, 2026".
func (inv *Invitation) FormattedFullDate() string {
	return inv.DateStart.Format("Monday, January 2, 2006")
}

// FormattedDateShort returns a compact date string like "Sep 19, 2026".
func (inv *Invitation) FormattedDateShort() string {
	return inv.DateStart.Format("Jan 2, 2006")
}

// FormattedTime returns a time string like "4:00 PM" or "4:00 PM – 11:30 PM".
func (inv *Invitation) FormattedTime() string {
	startStr := inv.DateStart.Format("3:04 PM")
	if inv.DateEnd != nil && !inv.DateEnd.IsZero() {
		return fmt.Sprintf("%s – %s", startStr, inv.DateEnd.Format("3:04 PM"))
	}
	return startStr
}

// GoogleCalendarURL returns a one-click Google Calendar event creation link.
func (inv *Invitation) GoogleCalendarURL() string {
	baseURL := "https://calendar.google.com/calendar/render?action=TEMPLATE"
	title := url.QueryEscape(inv.Title)
	details := url.QueryEscape(inv.Description)
	location := url.QueryEscape(fmt.Sprintf("%s, %s", inv.Location.Name, inv.Location.Address))

	timeFormat := "20060102T150405Z"
	dates := inv.DateStart.UTC().Format(timeFormat) + "/"
	if inv.DateEnd != nil && !inv.DateEnd.IsZero() {
		dates += inv.DateEnd.UTC().Format(timeFormat)
	} else {
		// Default to 2 hours duration
		dates += inv.DateStart.Add(2 * time.Hour).UTC().Format(timeFormat)
	}

	return fmt.Sprintf("%s&text=%s&dates=%s&details=%s&location=%s", baseURL, title, dates, details, location)
}

// IsRSVPOpen checks if RSVP is enabled and if the deadline has not passed.
func (inv *Invitation) IsRSVPOpen() bool {
	if inv.Sections.RSVP == nil || !inv.Sections.RSVP.Enabled {
		return false
	}
	if inv.Sections.RSVP.Deadline != nil && !inv.Sections.RSVP.Deadline.IsZero() {
		return time.Now().Before(*inv.Sections.RSVP.Deadline)
	}
	return true
}

// JSONLD returns the Schema.org Event structured data for SEO rich cards.
func (inv *Invitation) JSONLD() (string, error) {
	data := map[string]any{
		"@context":    "https://schema.org",
		"@type":       "Event",
		"name":        inv.Title,
		"description": inv.Description,
		"startDate":   inv.DateStart.Format(time.RFC3339),
		"eventStatus": "https://schema.org/EventScheduled",
		"location": map[string]any{
			"@type": "Place",
			"name":  inv.Location.Name,
			"address": map[string]any{
				"@type":         "PostalAddress",
				"streetAddress": inv.Location.Address,
			},
		},
	}

	if inv.DateEnd != nil && !inv.DateEnd.IsZero() {
		data["endDate"] = inv.DateEnd.Format(time.RFC3339)
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// RSVPSubmission represents a guest's attendance confirmation payload.
type RSVPSubmission struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Attending       bool   `json:"attending"`
	GuestCount      int    `json:"guest_count"`
	DietaryNeeds    string `json:"dietary_needs,omitempty"`
	SongRequest     string `json:"song_request,omitempty"`
	PersonalMessage string `json:"personal_message,omitempty"`
	SubmittedAt     time.Time `json:"submitted_at"`
}
