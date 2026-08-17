package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/goodsign/monday"
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

// SectionType identifies the discriminator type of a section block.
type SectionType string

const (
	SectionHero         SectionType = "hero"
	SectionDetails      SectionType = "details"
	SectionQuote        SectionType = "quote"
	SectionText         SectionType = "text"
	SectionImage        SectionType = "image"
	SectionCarousel     SectionType = "carousel"
	SectionTimeline     SectionType = "timeline"
	SectionDressCode    SectionType = "dress_code"
	SectionRSVP         SectionType = "rsvp"
	SectionRSVPExternal SectionType = "rsvp_external"
	SectionFAQs         SectionType = "faqs"
	SectionGiftRegistry SectionType = "gift_registry"
	SectionClosing      SectionType = "closing"
)

// Section defines the interface contract for any self-contained, ordered invitation building block.
type Section interface {
	Type() SectionType
	Validate() error
	TemplateName() string
}

// SectionFactory instantiates a fresh concrete Section implementation.
type SectionFactory func() Section

var sectionRegistry = map[SectionType]SectionFactory{
	SectionHero: func() Section { return &HeroSection{SectionType: SectionHero, ShowCountdown: true} },
	SectionDetails: func() Section {
		return &DetailsSection{SectionType: SectionDetails, ShowMapLink: true, ShowCalendarButton: true}
	},
	SectionQuote:        func() Section { return &QuoteSection{SectionType: SectionQuote} },
	SectionText:         func() Section { return &TextSection{SectionType: SectionText} },
	SectionImage:        func() Section { return &ImageSection{SectionType: SectionImage} },
	SectionCarousel:     func() Section { return &CarouselSection{SectionType: SectionCarousel} },
	SectionTimeline:     func() Section { return &TimelineSection{SectionType: SectionTimeline} },
	SectionDressCode:    func() Section { return &DressCodeSection{SectionType: SectionDressCode} },
	SectionRSVP:         func() Section { return &RSVPSection{SectionType: SectionRSVP, Enabled: true, MaxPartySize: 2} },
	SectionRSVPExternal: func() Section { return &RSVPExternalSection{SectionType: SectionRSVPExternal, Enabled: true} },
	SectionFAQs:         func() Section { return &FAQsSection{SectionType: SectionFAQs} },
	SectionGiftRegistry: func() Section { return &GiftRegistrySection{SectionType: SectionGiftRegistry} },
	SectionClosing:      func() Section { return &ClosingSection{SectionType: SectionClosing} },
}

// HeroSection configures the top introductory card (supports full-bleed overlay and split banner layouts).
type HeroSection struct {
	SectionType    SectionType `json:"type"`
	Layout         string      `json:"layout,omitempty"`
	Eyebrow        string      `json:"eyebrow,omitempty"`
	Vibe           string      `json:"vibe,omitempty"`
	Badge          string      `json:"badge,omitempty"`
	CoverImageURL  string      `json:"cover_image_url,omitempty"`
	BannerImageURL string      `json:"banner_image_url,omitempty"`
	ShowCountdown  bool        `json:"show_countdown"`
}

func (h *HeroSection) Type() SectionType { return SectionHero }

func (h *HeroSection) TemplateName() string {
	if h != nil && h.Layout == "banner" {
		return "partial_hero_banner"
	}
	return "partial_hero"
}

func (h *HeroSection) Validate() error { return nil }
func (h *HeroSection) HasCoverImage() bool {
	return h != nil && strings.TrimSpace(h.CoverImageURL) != ""
}
func (h *HeroSection) HasBannerImage() bool {
	return h != nil && strings.TrimSpace(h.BannerImageURL) != ""
}

// DetailsSection configures the Date, Time, and Venue quick strip.
type DetailsSection struct {
	SectionType        SectionType `json:"type"`
	DateLabel          string      `json:"date_label,omitempty"`
	LocationLabel      string      `json:"location_label,omitempty"`
	ShowMapLink        bool        `json:"show_map_link"`
	ShowCalendarButton bool        `json:"show_calendar_button"`
}

func (d *DetailsSection) Type() SectionType    { return SectionDetails }
func (d *DetailsSection) TemplateName() string { return "partial_details" }
func (d *DetailsSection) Validate() error      { return nil }

// QuoteSection represents an elegant quote or literary excerpt with quotation marks and author attribution.
type QuoteSection struct {
	SectionType SectionType `json:"type"`
	Text        string      `json:"text"`
	Author      string      `json:"author,omitempty"`
}

func (q *QuoteSection) Type() SectionType    { return SectionQuote }
func (q *QuoteSection) TemplateName() string { return "partial_quote" }
func (q *QuoteSection) Validate() error {
	if strings.TrimSpace(q.Text) == "" {
		return errors.New("quote text is required")
	}
	return nil
}

// TextSection represents a clean, plain text block for announcements, instructions, or notes.
type TextSection struct {
	SectionType SectionType `json:"type"`
	Title       string      `json:"title,omitempty"`
	Text        string      `json:"text"`
	Align       string      `json:"align,omitempty"`
}

func (t *TextSection) Type() SectionType    { return SectionText }
func (t *TextSection) TemplateName() string { return "partial_text" }
func (t *TextSection) Validate() error {
	if strings.TrimSpace(t.Text) == "" {
		return errors.New("text content is required")
	}
	return nil
}

// ImageSection represents a standalone static image block.
type ImageSection struct {
	SectionType SectionType `json:"type"`
	URL         string      `json:"url"`
	Caption     string      `json:"caption,omitempty"`
	Alt         string      `json:"alt,omitempty"`
}

func (img *ImageSection) Type() SectionType    { return SectionImage }
func (img *ImageSection) TemplateName() string { return "partial_image" }
func (img *ImageSection) Validate() error {
	if strings.TrimSpace(img.URL) == "" {
		return errors.New("image url is required")
	}
	return nil
}

func (img *ImageSection) AltText(defaultAlt string) string {
	if strings.TrimSpace(img.Alt) != "" {
		return img.Alt
	}
	if strings.TrimSpace(img.Caption) != "" {
		return img.Caption
	}
	return defaultAlt
}

// CarouselImage represents an individual photo in a carousel or gallery.
type CarouselImage struct {
	URL     string `json:"url"`
	Caption string `json:"caption,omitempty"`
	Alt     string `json:"alt,omitempty"`
}

func (img *CarouselImage) AltText(defaultAlt string) string {
	if strings.TrimSpace(img.Alt) != "" {
		return img.Alt
	}
	if strings.TrimSpace(img.Caption) != "" {
		return img.Caption
	}
	return defaultAlt
}

// CarouselSection configures an image gallery or carousel section.
type CarouselSection struct {
	SectionType SectionType     `json:"type"`
	Title       string          `json:"title,omitempty"`
	AspectRatio string          `json:"aspect_ratio,omitempty"`
	Fit         string          `json:"fit,omitempty"`
	Images      []CarouselImage `json:"images,omitempty"`
}

func (c *CarouselSection) Type() SectionType    { return SectionCarousel }
func (c *CarouselSection) TemplateName() string { return "partial_carousel" }
func (c *CarouselSection) Validate() error {
	if len(c.Images) == 0 {
		return errors.New("carousel images list cannot be empty")
	}
	for i, img := range c.Images {
		if strings.TrimSpace(img.URL) == "" {
			return fmt.Errorf("carousel image #%d: url is required", i+1)
		}
	}
	return nil
}
func (c *CarouselSection) HasImages() bool { return c != nil && len(c.Images) > 0 }
func (c *CarouselSection) ImageCount() int {
	if c == nil {
		return 0
	}
	return len(c.Images)
}

func (c *CarouselSection) AspectRatioClass() string {
	if c == nil {
		return "4-3"
	}
	switch strings.TrimSpace(c.AspectRatio) {
	case "1:1", "1/1", "square":
		return "1-1"
	case "16:9", "16/9", "video":
		return "16-9"
	case "4:5", "4/5", "portrait":
		return "4-5"
	case "4:3", "4/3":
		return "4-3"
	default:
		return "4-3"
	}
}

func (c *CarouselSection) IsCoverFit() bool {
	return c != nil && strings.EqualFold(strings.TrimSpace(c.Fit), "cover")
}

// TimelineItem represents a milestone in the event schedule.
type TimelineItem struct {
	Time        string `json:"time"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// TimelineSection holds a schedule of events.
type TimelineSection struct {
	SectionType SectionType    `json:"type"`
	Title       string         `json:"title,omitempty"`
	Items       []TimelineItem `json:"items"`
}

func (t *TimelineSection) Type() SectionType    { return SectionTimeline }
func (t *TimelineSection) TemplateName() string { return "partial_timeline" }
func (t *TimelineSection) Validate() error {
	if len(t.Items) == 0 {
		return errors.New("timeline items list cannot be empty")
	}
	for i, item := range t.Items {
		if strings.TrimSpace(item.Time) == "" {
			return fmt.Errorf("timeline item #%d: time is required", i+1)
		}
		if strings.TrimSpace(item.Title) == "" {
			return fmt.Errorf("timeline item #%d: title is required", i+1)
		}
	}
	return nil
}

// DressCodeSection details the dress code guidelines and color palette suggestions.
type DressCodeSection struct {
	SectionType  SectionType `json:"type"`
	Title        string      `json:"title,omitempty"`
	Name         string      `json:"name,omitempty"`
	Description  string      `json:"description,omitempty"`
	PaletteHints []string    `json:"palette_hints,omitempty"`
}

func (d *DressCodeSection) Type() SectionType    { return SectionDressCode }
func (d *DressCodeSection) TemplateName() string { return "partial_dress_code" }
func (d *DressCodeSection) Validate() error      { return nil }

// RSVPSection configures the attendance confirmation form.
type RSVPSection struct {
	SectionType    SectionType `json:"type"`
	Enabled        bool        `json:"enabled"`
	Deadline       *time.Time  `json:"deadline,omitempty"`
	MaxPartySize   int         `json:"max_party_size"`
	AskDietary     bool        `json:"ask_dietary"`
	AskSongRequest bool        `json:"ask_song_request"`
	CustomNote     string      `json:"custom_note,omitempty"`
}

func (r *RSVPSection) Type() SectionType    { return SectionRSVP }
func (r *RSVPSection) TemplateName() string { return "partial_rsvp_form" }
func (r *RSVPSection) Validate() error {
	if r.MaxPartySize <= 0 {
		r.MaxPartySize = 2
	}
	return nil
}

// RSVPExternalSection configures RSVP delegation to an external platform (Google Forms, Tally, etc.)
type RSVPExternalSection struct {
	SectionType      SectionType `json:"type"`
	Enabled          bool        `json:"enabled"`
	Title            string      `json:"title,omitempty"`
	Prompt           string      `json:"prompt,omitempty"`
	FormURL          string      `json:"form_url"`
	ButtonLabel      string      `json:"button_label,omitempty"`
	Deadline         *time.Time  `json:"deadline,omitempty"`
	ContributionNote string      `json:"contribution_note,omitempty"`
	ReceiptNote      string      `json:"receipt_note,omitempty"`
	CustomNote       string      `json:"custom_note,omitempty"`
}

func (r *RSVPExternalSection) Type() SectionType    { return SectionRSVPExternal }
func (r *RSVPExternalSection) TemplateName() string { return "partial_rsvp_external" }
func (r *RSVPExternalSection) Validate() error {
	if strings.TrimSpace(r.FormURL) == "" {
		return fmt.Errorf("rsvp_external section: form_url is required")
	}
	return nil
}
func (r *RSVPExternalSection) IsOpen() bool {
	if !r.Enabled {
		return false
	}
	if r.Deadline != nil && !r.Deadline.IsZero() {
		return time.Now().Before(*r.Deadline)
	}
	return true
}
func (r *RSVPExternalSection) HasContributionNote() bool {
	return strings.TrimSpace(r.ContributionNote) != ""
}
func (r *RSVPExternalSection) HasReceiptNote() bool {
	return strings.TrimSpace(r.ReceiptNote) != ""
}
func (r *RSVPExternalSection) HasCustomNote() bool {
	return strings.TrimSpace(r.CustomNote) != ""
}

// FAQItem represents a frequently asked question and answer.
type FAQItem struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// FAQsSection holds frequently asked questions.
type FAQsSection struct {
	SectionType SectionType `json:"type"`
	Title       string      `json:"title,omitempty"`
	Items       []FAQItem   `json:"items"`
}

func (f *FAQsSection) Type() SectionType    { return SectionFAQs }
func (f *FAQsSection) TemplateName() string { return "partial_faqs" }
func (f *FAQsSection) Validate() error {
	for i, item := range f.Items {
		if strings.TrimSpace(item.Question) == "" {
			return fmt.Errorf("FAQ #%d: question is required", i+1)
		}
		if strings.TrimSpace(item.Answer) == "" {
			return fmt.Errorf("FAQ #%d: answer is required", i+1)
		}
	}
	return nil
}

// RegistryLink represents a gift registry or donation link.
type RegistryLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// GiftRegistrySection describes gift registry instructions and URLs.
type GiftRegistrySection struct {
	SectionType SectionType    `json:"type"`
	Message     string         `json:"message,omitempty"`
	Links       []RegistryLink `json:"links,omitempty"`
}

func (g *GiftRegistrySection) Type() SectionType    { return SectionGiftRegistry }
func (g *GiftRegistrySection) TemplateName() string { return "partial_registry" }
func (g *GiftRegistrySection) Validate() error {
	for i, link := range g.Links {
		if strings.TrimSpace(link.Label) == "" {
			return fmt.Errorf("registry link #%d: label is required", i+1)
		}
		if strings.TrimSpace(link.URL) == "" {
			return fmt.Errorf("registry link #%d: url is required", i+1)
		}
	}
	return nil
}

// ClosingSection configures the final warm greeting and sign-off at the end of the invitation.
type ClosingSection struct {
	SectionType SectionType `json:"type"`
	Message     string      `json:"message,omitempty"`
	Signoff     string      `json:"signoff,omitempty"`
	Hosts       string      `json:"hosts,omitempty"`
}

func (c *ClosingSection) Type() SectionType    { return SectionClosing }
func (c *ClosingSection) TemplateName() string { return "partial_closing" }
func (c *ClosingSection) Validate() error      { return nil }

func (c *ClosingSection) DisplayHosts(defaultHosts string) string {
	if c != nil && strings.TrimSpace(c.Hosts) != "" {
		return c.Hosts
	}
	return defaultHosts
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
	Sections    []Section   `json:"sections"`
}

// UnmarshalJSON unmarshals an Invitation, supporting both ordered []Section arrays and legacy section maps.
func (inv *Invitation) UnmarshalJSON(data []byte) error {
	type rawInvitation struct {
		Version     string          `json:"version"`
		Slug        string          `json:"slug"`
		Title       string          `json:"title"`
		Subtitle    string          `json:"subtitle,omitempty"`
		Hosts       []string        `json:"hosts,omitempty"`
		Description string          `json:"description,omitempty"`
		DateStart   time.Time       `json:"date_start"`
		DateEnd     *time.Time      `json:"date_end,omitempty"`
		Timezone    string          `json:"timezone"`
		Location    Location        `json:"location"`
		Theme       ThemeConfig     `json:"theme"`
		Sections    json.RawMessage `json:"sections"`
	}

	var raw rawInvitation
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	inv.Version = raw.Version
	inv.Slug = raw.Slug
	inv.Title = raw.Title
	inv.Subtitle = raw.Subtitle
	inv.Hosts = raw.Hosts
	inv.Description = raw.Description
	inv.DateStart = raw.DateStart
	inv.DateEnd = raw.DateEnd
	inv.Timezone = raw.Timezone
	inv.Location = raw.Location
	inv.Theme = raw.Theme
	inv.Sections = nil

	if len(raw.Sections) == 0 {
		return nil
	}

	trimmed := bytes.TrimSpace(raw.Sections)
	if bytes.HasPrefix(trimmed, []byte("[")) {
		// Canonical Array of Polymorphic Section Blocks
		var rawBlocks []json.RawMessage
		if err := json.Unmarshal(trimmed, &rawBlocks); err != nil {
			return fmt.Errorf("failed to unmarshal sections array: %w", err)
		}

		for idx, blockRaw := range rawBlocks {
			var typeHeader struct {
				Type SectionType `json:"type"`
			}
			if err := json.Unmarshal(blockRaw, &typeHeader); err != nil {
				return fmt.Errorf("section #%d: invalid block header: %w", idx+1, err)
			}
			factory, ok := sectionRegistry[typeHeader.Type]
			if !ok {
				return fmt.Errorf("section #%d: unknown section type %q", idx+1, typeHeader.Type)
			}
			sec := factory()
			if err := json.Unmarshal(blockRaw, sec); err != nil {
				return fmt.Errorf("section #%d (%s): unmarshal error: %w", idx+1, typeHeader.Type, err)
			}
			inv.Sections = append(inv.Sections, sec)
		}
	} else if bytes.HasPrefix(trimmed, []byte("{")) {
		// Legacy Object Map
		var legacy struct {
			Hero         *HeroSection         `json:"hero,omitempty"`
			Quote        *QuoteSection        `json:"quote,omitempty"`
			Text         *TextSection         `json:"text,omitempty"`
			Carousel     *CarouselSection     `json:"carousel,omitempty"`
			Timeline     []TimelineItem       `json:"timeline,omitempty"`
			Image        *ImageSection        `json:"image,omitempty"`
			DressCode    *DressCodeSection    `json:"dress_code,omitempty"`
			RSVP         *RSVPSection         `json:"rsvp,omitempty"`
			FAQs         []FAQItem            `json:"faqs,omitempty"`
			GiftRegistry *GiftRegistrySection `json:"gift_registry,omitempty"`
			Closing      *ClosingSection      `json:"closing,omitempty"`
		}
		if err := json.Unmarshal(trimmed, &legacy); err != nil {
			return fmt.Errorf("failed to unmarshal legacy sections map: %w", err)
		}

		if legacy.Hero != nil {
			legacy.Hero.SectionType = SectionHero
			inv.Sections = append(inv.Sections, legacy.Hero)
		}
		if legacy.Quote != nil {
			legacy.Quote.SectionType = SectionQuote
			inv.Sections = append(inv.Sections, legacy.Quote)
		}
		if legacy.Text != nil {
			legacy.Text.SectionType = SectionText
			inv.Sections = append(inv.Sections, legacy.Text)
		}
		if legacy.Carousel != nil {
			legacy.Carousel.SectionType = SectionCarousel
			inv.Sections = append(inv.Sections, legacy.Carousel)
		}
		// Default Details strip
		inv.Sections = append(inv.Sections, &DetailsSection{SectionType: SectionDetails, ShowMapLink: true, ShowCalendarButton: true})

		if len(legacy.Timeline) > 0 {
			inv.Sections = append(inv.Sections, &TimelineSection{SectionType: SectionTimeline, Items: legacy.Timeline})
		}
		if legacy.Image != nil {
			legacy.Image.SectionType = SectionImage
			inv.Sections = append(inv.Sections, legacy.Image)
		}
		if legacy.DressCode != nil {
			legacy.DressCode.SectionType = SectionDressCode
			inv.Sections = append(inv.Sections, legacy.DressCode)
		}
		if legacy.RSVP != nil {
			legacy.RSVP.SectionType = SectionRSVP
			inv.Sections = append(inv.Sections, legacy.RSVP)
		}
		if len(legacy.FAQs) > 0 {
			inv.Sections = append(inv.Sections, &FAQsSection{SectionType: SectionFAQs, Items: legacy.FAQs})
		}
		if legacy.GiftRegistry != nil {
			legacy.GiftRegistry.SectionType = SectionGiftRegistry
			inv.Sections = append(inv.Sections, legacy.GiftRegistry)
		}
		if legacy.Closing != nil {
			legacy.Closing.SectionType = SectionClosing
			inv.Sections = append(inv.Sections, legacy.Closing)
		}
	}

	return nil
}

// HeroSection returns the first HeroSection configured in the invitation, or nil.
func (inv *Invitation) HeroSection() *HeroSection {
	for _, sec := range inv.Sections {
		if h, ok := sec.(*HeroSection); ok {
			return h
		}
	}
	return nil
}

// RSVPSection returns the first RSVPSection configured in the invitation, or nil.
func (inv *Invitation) RSVPSection() *RSVPSection {
	for _, sec := range inv.Sections {
		if r, ok := sec.(*RSVPSection); ok {
			return r
		}
	}
	return nil
}

// RSVPExternalSection returns the first RSVPExternalSection configured in the invitation, or nil.
func (inv *Invitation) RSVPExternalSection() *RSVPExternalSection {
	for _, sec := range inv.Sections {
		if r, ok := sec.(*RSVPExternalSection); ok {
			return r
		}
	}
	return nil
}

// CarouselSection returns the first CarouselSection configured in the invitation, or nil.
func (inv *Invitation) CarouselSection() *CarouselSection {
	for _, sec := range inv.Sections {
		if c, ok := sec.(*CarouselSection); ok {
			return c
		}
	}
	return nil
}

// HasCoverImage returns true if a hero cover image is present.
func (inv *Invitation) HasCoverImage() bool {
	h := inv.HeroSection()
	return h != nil && h.HasCoverImage()
}

// CoverImage returns the hero cover image URL or empty string.
func (inv *Invitation) CoverImage() string {
	h := inv.HeroSection()
	if h != nil {
		return h.CoverImageURL
	}
	return ""
}

// HasCarousel returns true if any carousel section contains images.
func (inv *Invitation) HasCarousel() bool {
	c := inv.CarouselSection()
	return c != nil && c.HasImages()
}

// IsRSVPOpen checks if RSVP is enabled and if the deadline has not passed.
func (inv *Invitation) IsRSVPOpen() bool {
	if r := inv.RSVPSection(); r != nil && r.Enabled {
		if r.Deadline != nil && !r.Deadline.IsZero() {
			return time.Now().Before(*r.Deadline)
		}
		return true
	}
	if re := inv.RSVPExternalSection(); re != nil && re.Enabled {
		return re.IsOpen()
	}
	return false
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
	return monday.Format(inv.DateStart, "Monday, 2 de January de 2006", monday.LocaleEsES)
	// return inv.DateStart.Format("Monday, January 2, 2006")
}

// FormattedDateShort returns a compact date string.
func (inv *Invitation) FormattedDateShort() string {
	return inv.DateStart.Format("02/01/06")
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

	if inv.HasCoverImage() {
		data["image"] = inv.CoverImage()
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
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Attending       bool      `json:"attending"`
	GuestCount      int       `json:"guest_count"`
	DietaryNeeds    string    `json:"dietary_needs,omitempty"`
	SongRequest     string    `json:"song_request,omitempty"`
	PersonalMessage string    `json:"personal_message,omitempty"`
	SubmittedAt     time.Time `json:"submitted_at"`
}
