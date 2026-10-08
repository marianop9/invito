package domain

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestInvitationValidation(t *testing.T) {
	now := time.Now().Add(24 * time.Hour)
	end := now.Add(4 * time.Hour)

	validInv := Invitation{
		Version:   "1.0",
		Slug:      "test-party",
		Title:     "Test Party Celebration",
		DateStart: now,
		DateEnd:   &end,
		Location: Location{
			Name:    "Grand Ballroom",
			Address: "123 Main St, City, Country",
		},
		Theme: ThemeConfig{
			ID: ThemeBotanicalElegance,
		},
		Sections: []Section{
			&HeroSection{SectionType: SectionHero, Badge: "Save The Date"},
			&RSVPSection{SectionType: SectionRSVP, Enabled: true, MaxPartySize: 2},
		},
	}

	if err := validInv.Validate(); err != nil {
		t.Fatalf("expected valid invitation, got error: %v", err)
	}

	// Test invalid slug with spaces
	invalidSlug := validInv
	invalidSlug.Slug = "invalid slug with spaces"
	if err := invalidSlug.Validate(); err == nil {
		t.Errorf("expected error for invalid slug, got nil")
	}

	// Test invalid date_end before date_start
	invalidDates := validInv
	badEnd := now.Add(-1 * time.Hour)
	invalidDates.DateEnd = &badEnd
	if err := invalidDates.Validate(); err == nil {
		t.Errorf("expected error when date_end is before date_start, got nil")
	}

	// Test invalid carousel with empty image URL
	invalidCarouselInv := validInv
	invalidCarouselInv.Sections = []Section{
		&CarouselSection{
			SectionType: SectionCarousel,
			Title:       "Our Moments",
			Images: []CarouselImage{
				{URL: "/static/img/hero.webp", Caption: "Valid photo"},
				{URL: "", Caption: "Missing URL photo"},
			},
		},
	}
	if err := invalidCarouselInv.Validate(); err == nil {
		t.Errorf("expected error when carousel image url is empty, got nil")
	}

	// Test invalid quote with empty text
	invalidQuoteInv := validInv
	invalidQuoteInv.Sections = []Section{
		&QuoteSection{SectionType: SectionQuote, Text: "   "},
	}
	if err := invalidQuoteInv.Validate(); err == nil {
		t.Errorf("expected error when quote text is empty, got nil")
	}

	// Test invalid text with empty text
	invalidTextInv := validInv
	invalidTextInv.Sections = []Section{
		&TextSection{SectionType: SectionText, Text: "   "},
	}
	if err := invalidTextInv.Validate(); err == nil {
		t.Errorf("expected error when text content is empty, got nil")
	}

	// Test invalid image with empty url
	invalidImgInv := validInv
	invalidImgInv.Sections = []Section{
		&ImageSection{SectionType: SectionImage, URL: "   "},
	}
	if err := invalidImgInv.Validate(); err == nil {
		t.Errorf("expected error when image url is empty, got nil")
	}
}

func TestInvitationJSONUnmarshaling(t *testing.T) {
	t.Run("Polymorphic Array Unmarshaling with Quote, Text, and Images", func(t *testing.T) {
		jsonBlob := `{
			"version": "1.0",
			"slug": "custom-blocks-event",
			"title": "Custom Blocks Wedding",
			"date_start": "2026-09-19T16:00:00Z",
			"location": {
				"name": "Botanical Estate",
				"address": "450 Magnolia Lane"
			},
			"theme": { "id": "botanical-elegance" },
			"sections": [
				{ "type": "hero", "badge": "Special Announcement", "cover_image_url": "/static/img/demo-hero.webp" },
				{ "type": "quote", "text": "Two lives, one path.", "author": "Poet" },
				{ "type": "carousel", "title": "Photo Moments", "images": [{ "url": "/static/img/c1.webp" }, { "url": "/static/img/c2.webp" }] },
				{ "type": "details", "show_map_link": true },
				{ "type": "timeline", "title": "Schedule", "items": [{ "time": "4:00 PM", "title": "Ceremony" }] },
				{ "type": "image", "url": "/static/img/venue.webp", "caption": "The Glasshouse Conservatory" },
				{ "type": "text", "title": "Guest Logistics", "text": "Shuttle departs at 3:15 PM sharp." },
				{ "type": "dress_code", "title": "Garden Formal", "palette_hints": ["#2A4738", "#D4AF37"] },
				{ "type": "image", "url": "/static/img/swatches.webp", "caption": "Color Swatches" },
				{ "type": "rsvp", "enabled": true, "max_party_size": 2 },
				{ "type": "faqs", "title": "Questions", "items": [{ "question": "Parking?", "answer": "Valet at gate" }] },
				{ "type": "gift_registry", "message": "Gifts welcome", "links": [{ "label": "Store", "url": "https://store.com" }] },
				{ "type": "closing", "message": "See you there!", "signoff": "Love,", "hosts": "Sarah & Alex" }
			]
		}`

		var inv Invitation
		if err := json.Unmarshal([]byte(jsonBlob), &inv); err != nil {
			t.Fatalf("failed to unmarshal polymorphic blocks: %v", err)
		}

		if len(inv.Sections) != 13 {
			t.Fatalf("expected 13 sections, got %d", len(inv.Sections))
		}

		// Verify ordering and types
		expectedTypes := []SectionType{
			SectionHero,
			SectionQuote,
			SectionCarousel,
			SectionDetails,
			SectionTimeline,
			SectionImage,
			SectionText,
			SectionDressCode,
			SectionImage,
			SectionRSVP,
			SectionFAQs,
			SectionGiftRegistry,
			SectionClosing,
		}

		for i, expected := range expectedTypes {
			if inv.Sections[i].Type() != expected {
				t.Errorf("section #%d: expected type %s, got %s", i+1, expected, inv.Sections[i].Type())
			}
		}

		// Check quote contents
		quote, ok := inv.Sections[1].(*QuoteSection)
		if !ok || !strings.Contains(quote.Text, "Two lives") || quote.Author != "Poet" {
			t.Errorf("unexpected quote: %+v", inv.Sections[1])
		}

		// Check text announcement contents
		txt, ok := inv.Sections[6].(*TextSection)
		if !ok || !strings.Contains(txt.Text, "Shuttle departs") || txt.Title != "Guest Logistics" {
			t.Errorf("unexpected txt: %+v", inv.Sections[6])
		}

		// Check multiple image contents
		img1, ok := inv.Sections[5].(*ImageSection)
		if !ok || img1.URL != "/static/img/venue.webp" {
			t.Errorf("unexpected img1: %+v", inv.Sections[5])
		}
		img2, ok := inv.Sections[8].(*ImageSection)
		if !ok || img2.URL != "/static/img/swatches.webp" {
			t.Errorf("unexpected img2: %+v", inv.Sections[8])
		}

		if err := inv.Validate(); err != nil {
			t.Fatalf("unmarshaled invitation failed validation: %v", err)
		}
	})

	t.Run("Legacy Object Map Fallback Unmarshaling", func(t *testing.T) {
		legacyJSON := `{
			"version": "1.0",
			"slug": "legacy-party",
			"title": "Legacy Party",
			"date_start": "2026-09-19T16:00:00Z",
			"location": { "name": "Hall", "address": "123 Main" },
			"theme": { "id": "golden-sunset" },
			"sections": {
				"hero": { "badge": "Legacy Eyebrow" },
				"quote": { "text": "Legacy statement", "author": "Host" },
				"text": { "text": "Legacy announcement" },
				"rsvp": { "enabled": true, "max_party_size": 2 }
			}
		}`

		var inv Invitation
		if err := json.Unmarshal([]byte(legacyJSON), &inv); err != nil {
			t.Fatalf("failed to unmarshal legacy map: %v", err)
		}

		if len(inv.Sections) == 0 {
			t.Fatalf("expected legacy sections to convert to slice, got 0")
		}

		if inv.HeroSection() == nil || inv.HeroSection().Badge != "Legacy Eyebrow" {
			t.Errorf("expected hero section to be parsed from legacy map")
		}
		if inv.RSVPSection() == nil || !inv.RSVPSection().Enabled {
			t.Errorf("expected rsvp section to be parsed from legacy map")
		}
	})
}

func TestInvitationHelpers(t *testing.T) {
	date := time.Date(2026, time.September, 19, 16, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.September, 19, 23, 30, 0, 0, time.UTC)

	inv := Invitation{
		Slug:      "sarah-and-alex",
		Title:     "Sarah & Alex Wedding",
		Hosts:     []string{"Sarah", "Alex"},
		DateStart: date,
		DateEnd:   &end,
		Location: Location{
			Name:    "Botanical Gardens",
			Address: "123 Flower Ave",
		},
		Sections: []Section{
			&HeroSection{
				SectionType:    SectionHero,
				Badge:          "Save The Date",
				CoverImageURL:  "/static/img/hero.webp",
				BannerImageURL: "/static/img/banner.webp",
			},
			&QuoteSection{
				SectionType: SectionQuote,
				Text:        "Two lives, one shared journey.",
				Author:      "Poet",
			},
			&CarouselSection{
				SectionType: SectionCarousel,
				Title:       "Our Journey",
				Images: []CarouselImage{
					{URL: "/static/img/carousel-1.webp", Caption: "Engagement Day", Alt: "Sarah and Alex engagement"},
					{URL: "/static/img/carousel-2.webp", Caption: "Summer in Italy"},
				},
			},
			&ImageSection{
				SectionType: SectionImage,
				URL:         "/static/img/venue.webp",
				Caption:     "The conservatory",
				Alt:         "Glasshouse venue",
			},
			&ClosingSection{
				SectionType: SectionClosing,
				Message:     "See you soon!",
				Signoff:     "With love,",
				Hosts:       "Sarah & Alex",
			},
		},
	}

	if inv.HostsDisplay() != "Sarah & Alex" {
		t.Errorf("expected 'Sarah & Alex', got %q", inv.HostsDisplay())
	}

	if inv.FormattedDateShort() != "19/09/26" {
		t.Errorf("expected '19/09/26', got %q", inv.FormattedDateShort())
	}

	if !inv.HasCoverImage() {
		t.Errorf("expected HasCoverImage() to be true")
	}
	if inv.CoverImage() != "/static/img/hero.webp" {
		t.Errorf("expected CoverImage() to return '/static/img/hero.webp', got %q", inv.CoverImage())
	}
	hero := inv.HeroSection()
	if !hero.HasCoverImage() || !hero.HasBannerImage() {
		t.Errorf("expected HeroSection HasCoverImage and HasBannerImage to be true")
	}
	if hero.TemplateName() != "partial_hero" {
		t.Errorf("expected default hero TemplateName to be 'partial_hero', got %q", hero.TemplateName())
	}

	bannerHero := HeroSection{Layout: "banner"}
	if bannerHero.TemplateName() != "partial_hero_banner" {
		t.Errorf("expected banner hero TemplateName to be 'partial_hero_banner', got %q", bannerHero.TemplateName())
	}

	imgSec := inv.Sections[3].(*ImageSection)
	if imgSec.AltText("def") != "Glasshouse venue" {
		t.Errorf("expected image alt text 'Glasshouse venue', got %q", imgSec.AltText("def"))
	}

	closingSec := inv.Sections[4].(*ClosingSection)
	if closingSec.DisplayHosts("fallback") != "Sarah & Alex" {
		t.Errorf("expected 'Sarah & Alex', got %q", closingSec.DisplayHosts("fallback"))
	}

	if !inv.HasCarousel() {
		t.Errorf("expected HasCarousel() to be true")
	}
	carousel := inv.CarouselSection()
	if carousel.ImageCount() != 2 {
		t.Errorf("expected ImageCount() == 2, got %d", carousel.ImageCount())
	}
	if !carousel.HasImages() {
		t.Errorf("expected HasImages() to be true")
	}

	img1 := carousel.Images[0]
	if img1.AltText("default") != "Sarah and Alex engagement" {
		t.Errorf("expected alt text 'Sarah and Alex engagement', got %q", img1.AltText("default"))
	}
	img2 := carousel.Images[1]
	if img2.AltText("default") != "Summer in Italy" {
		t.Errorf("expected caption fallback for alt text 'Summer in Italy', got %q", img2.AltText("default"))
	}
	emptyImg := CarouselImage{}
	if emptyImg.AltText("fallback") != "fallback" {
		t.Errorf("expected fallback 'fallback', got %q", emptyImg.AltText("fallback"))
	}

	calURL := inv.GoogleCalendarURL()
	if !strings.Contains(calURL, "calendar.google.com") {
		t.Errorf("expected google calendar link, got %q", calURL)
	}

	jsonld, err := inv.JSONLD()
	if err != nil {
		t.Fatalf("failed to generate JSON-LD: %v", err)
	}
	if !strings.Contains(jsonld, "EventScheduled") {
		t.Errorf("expected JSON-LD to contain EventScheduled, got %s", jsonld)
	}
	if !strings.Contains(jsonld, "/static/img/hero.webp") {
		t.Errorf("expected JSON-LD to contain cover image, got %s", jsonld)
	}
}

func TestRSVPValidation(t *testing.T) {
	sub := RSVPSubmission{
		Name:       "Jane Doe",
		Email:      "jane@example.com",
		Attending:  true,
		GuestCount: 2,
	}

	if err := sub.Validate(2); err != nil {
		t.Fatalf("expected valid RSVP, got %v", err)
	}

	// Exceed party size
	if err := sub.Validate(1); err == nil {
		t.Errorf("expected error when guest count exceeds max party size, got nil")
	}

	// Invalid email
	sub.Email = "invalid-email"
	if err := sub.Validate(2); err == nil {
		t.Errorf("expected error for invalid email, got nil")
	}
}

func TestCarouselSectionHelperMethods(t *testing.T) {
	c := &CarouselSection{
		SectionType: SectionCarousel,
		Title:       "Test Gallery",
		Images: []CarouselImage{
			{URL: "/static/img/photo1.webp"},
		},
	}

	// Default aspect ratio
	if c.AspectRatioClass() != "4-3" {
		t.Errorf("expected default aspect ratio class '4-3', got %q", c.AspectRatioClass())
	}
	if c.IsCoverFit() {
		t.Errorf("expected default IsCoverFit to be false, got true")
	}

	// 16:9
	c.AspectRatio = "16:9"
	if c.AspectRatioClass() != "16-9" {
		t.Errorf("expected aspect ratio class '16-9', got %q", c.AspectRatioClass())
	}

	// Square 1:1
	c.AspectRatio = "1:1"
	if c.AspectRatioClass() != "1-1" {
		t.Errorf("expected aspect ratio class '1-1', got %q", c.AspectRatioClass())
	}

	// 4:5 Portrait
	c.AspectRatio = "4:5"
	if c.AspectRatioClass() != "4-5" {
		t.Errorf("expected aspect ratio class '4-5', got %q", c.AspectRatioClass())
	}

	// 3:4 Tall Portrait
	c.AspectRatio = "3:4"
	if c.AspectRatioClass() != "3-4" {
		t.Errorf("expected aspect ratio class '3-4', got %q", c.AspectRatioClass())
	}

	// 9:16 Cinematic Story
	c.AspectRatio = "9:16"
	if c.AspectRatioClass() != "9-16" {
		t.Errorf("expected aspect ratio class '9-16', got %q", c.AspectRatioClass())
	}

	// Fit cover
	c.Fit = "cover"
	if !c.IsCoverFit() {
		t.Errorf("expected IsCoverFit to be true when Fit is 'cover', got false")
	}

	// Nil safety
	var nilCarousel *CarouselSection
	if nilCarousel.AspectRatioClass() != "4-3" {
		t.Errorf("expected nil carousel AspectRatioClass to return '4-3', got %q", nilCarousel.AspectRatioClass())
	}
	if nilCarousel.IsCoverFit() {
		t.Errorf("expected nil carousel IsCoverFit to return false, got true")
	}
}

func TestMultipleCarousels(t *testing.T) {
	c1 := &CarouselSection{
		SectionType: SectionCarousel,
		Title:       "Gallery 1",
		Images: []CarouselImage{
			{URL: "/img/1.webp"},
			{URL: "/img/2.webp"},
		},
	}
	c2 := &CarouselSection{
		SectionType: SectionCarousel,
		Title:       "Gallery 2",
		Images: []CarouselImage{
			{URL: "/img/3.webp"},
			{URL: "/img/4.webp"},
			{URL: "/img/5.webp"},
		},
	}

	inv := &Invitation{
		Slug:  "multi-carousel-test",
		Title: "Multi Carousel Test",
		Sections: []Section{
			c1,
			&TextSection{SectionType: SectionText, Title: "Interlude", Text: "Some text"},
			c2,
		},
	}

	if !inv.HasCarousel() {
		t.Errorf("expected HasCarousel to be true")
	}

	sections := inv.CarouselSections()
	if len(sections) != 2 {
		t.Fatalf("expected 2 carousel sections, got %d", len(sections))
	}

	if inv.TotalCarouselImages() != 5 {
		t.Errorf("expected 5 total carousel images, got %d", inv.TotalCarouselImages())
	}
}

func TestRSVPExternalSection(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	ext := &RSVPExternalSection{
		SectionType:      SectionRSVPExternal,
		Enabled:          true,
		Title:            "Confirmación de Asistencia",
		Prompt:           "Completá el formulario para confirmar tu lugar:",
		FormURL:          "https://forms.gle/demo-link",
		ButtonLabel:      "Completar Formulario",
		ContributionNote: "Seña de $15.000 (Alias: cari.cumple45)",
		ReceiptNote:      "Enviá el comprobante por WhatsApp al anfitrión",
		CustomNote:       "Hasta el 20 de Agosto",
		Deadline:         &future,
	}

	if ext.Type() != SectionRSVPExternal {
		t.Errorf("expected type 'rsvp_external', got %s", ext.Type())
	}

	if ext.TemplateName() != "partial_rsvp_external" {
		t.Errorf("expected TemplateName 'partial_rsvp_external', got %s", ext.TemplateName())
	}

	if err := ext.Validate(); err != nil {
		t.Fatalf("expected valid rsvp_external section, got error: %v", err)
	}

	if !ext.IsOpen() {
		t.Errorf("expected rsvp_external to be open with future deadline")
	}

	if !ext.HasContributionNote() {
		t.Errorf("expected HasContributionNote to be true")
	}

	if !ext.HasReceiptNote() {
		t.Errorf("expected HasReceiptNote to be true")
	}

	if !ext.HasCustomNote() {
		t.Errorf("expected HasCustomNote to be true")
	}

	// Missing form_url validation error
	invalidExt := &RSVPExternalSection{
		SectionType: SectionRSVPExternal,
		Enabled:     true,
	}
	if err := invalidExt.Validate(); err == nil {
		t.Errorf("expected error when form_url is empty, got nil")
	}

	// Past deadline
	ext.Deadline = &past
	if ext.IsOpen() {
		t.Errorf("expected IsOpen to be false with past deadline")
	}

	// Disabled
	ext.Enabled = false
	ext.Deadline = nil
	if ext.IsOpen() {
		t.Errorf("expected IsOpen to be false when disabled")
	}

	// Invitation helper methods
	inv := &Invitation{
		Slug: "test-rsvp-ext",
		Sections: []Section{
			ext,
		},
	}
	if inv.RSVPExternalSection() != ext {
		t.Errorf("expected inv.RSVPExternalSection() to return ext")
	}
}

func TestMusicConfig(t *testing.T) {
	t.Run("Validation", func(t *testing.T) {
		var nilMusic *MusicConfig
		if err := nilMusic.Validate(); err != nil {
			t.Errorf("expected nil music config to be valid, got: %v", err)
		}

		emptyURL := &MusicConfig{URL: ""}
		if err := emptyURL.Validate(); err == nil || !strings.Contains(err.Error(), "url is required") {
			t.Errorf("expected url required error, got: %v", err)
		}

		whitespaceURL := &MusicConfig{URL: "   "}
		if err := whitespaceURL.Validate(); err == nil || !strings.Contains(err.Error(), "url is required") {
			t.Errorf("expected url required error for whitespace, got: %v", err)
		}

		validMusic := &MusicConfig{URL: "/uploads/song.mp3", Title: "My Song", Autoplay: true, Loop: true}
		if err := validMusic.Validate(); err != nil {
			t.Errorf("expected valid music config, got: %v", err)
		}
	})

	t.Run("DisplayTitle", func(t *testing.T) {
		var nilMusic *MusicConfig
		if nilMusic.DisplayTitle("Fallback") != "Fallback" {
			t.Errorf("expected 'Fallback' on nil music, got %q", nilMusic.DisplayTitle("Fallback"))
		}

		m := &MusicConfig{Title: "A Thousand Years"}
		if m.DisplayTitle("Fallback") != "A Thousand Years" {
			t.Errorf("expected 'A Thousand Years', got %q", m.DisplayTitle("Fallback"))
		}

		emptyTitle := &MusicConfig{Title: "   "}
		if emptyTitle.DisplayTitle("Fallback") != "Fallback" {
			t.Errorf("expected 'Fallback' for empty title, got %q", emptyTitle.DisplayTitle("Fallback"))
		}
	})

	t.Run("JSON Serialization and Deserialization", func(t *testing.T) {
		orig := MusicConfig{
			URL:      "/uploads/sunset.mp3",
			Title:    "Sunset Melody",
			Autoplay: true,
			Loop:     true,
		}

		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("failed to marshal MusicConfig: %v", err)
		}

		var parsed MusicConfig
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("failed to unmarshal MusicConfig: %v", err)
		}

		if parsed != orig {
			t.Errorf("expected %+v, got %+v", orig, parsed)
		}
	})
}

func TestSplashScreenConfig(t *testing.T) {
	t.Run("Validation", func(t *testing.T) {
		var nilSplash *SplashScreenConfig
		if err := nilSplash.Validate(); err != nil {
			t.Errorf("expected nil splash screen to be valid, got: %v", err)
		}

		s := &SplashScreenConfig{Enabled: true}
		if err := s.Validate(); err != nil {
			t.Errorf("expected splash screen to be valid, got: %v", err)
		}
	})

	t.Run("IsActive", func(t *testing.T) {
		var nilSplash *SplashScreenConfig
		if nilSplash.IsActive() {
			t.Errorf("expected nil splash screen to not be active")
		}

		disabled := &SplashScreenConfig{Enabled: false}
		if disabled.IsActive() {
			t.Errorf("expected disabled splash screen to not be active")
		}

		enabled := &SplashScreenConfig{Enabled: true}
		if !enabled.IsActive() {
			t.Errorf("expected enabled splash screen to be active")
		}
	})

	t.Run("CTAButtonText", func(t *testing.T) {
		var nilSplash *SplashScreenConfig
		if nilSplash.CTAButtonText() != "Open Invitation" {
			t.Errorf("expected default 'Open Invitation' on nil, got %q", nilSplash.CTAButtonText())
		}

		emptyBtn := &SplashScreenConfig{ButtonText: "   "}
		if emptyBtn.CTAButtonText() != "Open Invitation" {
			t.Errorf("expected default 'Open Invitation' on empty button text, got %q", emptyBtn.CTAButtonText())
		}

		customBtn := &SplashScreenConfig{ButtonText: "Abrir Invitación"}
		if customBtn.CTAButtonText() != "Abrir Invitación" {
			t.Errorf("expected custom button text, got %q", customBtn.CTAButtonText())
		}
	})

	t.Run("JSON Serialization and Deserialization", func(t *testing.T) {
		orig := SplashScreenConfig{
			Enabled:            true,
			Title:              "Welcome Guests",
			Message:            "Please tap to enter",
			ButtonText:         "Enter Event",
			BackgroundImageURL: "/uploads/cover.jpg",
		}

		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("failed to marshal SplashScreenConfig: %v", err)
		}

		var parsed SplashScreenConfig
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("failed to unmarshal SplashScreenConfig: %v", err)
		}

		if parsed != orig {
			t.Errorf("expected %+v, got %+v", orig, parsed)
		}
	})
}

func TestInvitationWithMusicAndSplashScreen(t *testing.T) {
	jsonBlob := `{
		"version": "1.0",
		"slug": "audio-splash-party",
		"title": "Audio & Splash Fiesta",
		"date_start": "2026-09-19T18:00:00Z",
		"location": {
			"name": "Sunset Terrace",
			"address": "123 Ocean Blvd"
		},
		"theme": { "id": "golden-sunset" },
		"music": {
			"url": "/uploads/party-track.mp3",
			"title": "Fiesta Vibes",
			"autoplay": true,
			"loop": true
		},
		"splash_screen": {
			"enabled": true,
			"title": "Welcome to the Fiesta",
			"message": "Put on your dancing shoes!",
			"button_text": "Join Party",
			"background_image_url": "/uploads/splash-bg.jpg"
		},
		"sections": [
			{ "type": "hero", "eyebrow": "Fiesta 2026" }
		]
	}`

	var inv Invitation
	if err := json.Unmarshal([]byte(jsonBlob), &inv); err != nil {
		t.Fatalf("failed to unmarshal invitation with music & splash: %v", err)
	}

	if inv.Music == nil {
		t.Fatalf("expected non-nil Music")
	}
	if inv.Music.URL != "/uploads/party-track.mp3" || inv.Music.Title != "Fiesta Vibes" || !inv.Music.Autoplay || !inv.Music.Loop {
		t.Errorf("unexpected MusicConfig values: %+v", inv.Music)
	}
	if !inv.HasMusic() {
		t.Errorf("expected HasMusic() to be true")
	}

	if inv.SplashScreen == nil {
		t.Fatalf("expected non-nil SplashScreen")
	}
	if !inv.SplashScreen.Enabled || inv.SplashScreen.Title != "Welcome to the Fiesta" || inv.SplashScreen.ButtonText != "Join Party" {
		t.Errorf("unexpected SplashScreenConfig values: %+v", inv.SplashScreen)
	}
	if !inv.HasSplashScreen() {
		t.Errorf("expected HasSplashScreen() to be true")
	}

	if err := inv.Validate(); err != nil {
		t.Fatalf("expected valid invitation, got: %v", err)
	}

	// Validation failure when music URL is missing
	inv.Music.URL = ""
	if err := inv.Validate(); err == nil || !strings.Contains(err.Error(), "music config: url is required") {
		t.Errorf("expected validation error for empty music url, got: %v", err)
	}
}

func TestInvitationBackwardCompatibility(t *testing.T) {
	legacyJSON := `{
		"version": "1.0",
		"slug": "legacy-no-audio",
		"title": "Legacy Invitation Without Audio",
		"date_start": "2026-09-19T18:00:00Z",
		"location": {
			"name": "Classic Ballroom",
			"address": "789 Main St"
		},
		"theme": { "id": "botanical-elegance" },
		"sections": [
			{ "type": "hero", "badge": "Legacy Event" }
		]
	}`

	var inv Invitation
	if err := json.Unmarshal([]byte(legacyJSON), &inv); err != nil {
		t.Fatalf("failed to unmarshal legacy invitation without music/splash: %v", err)
	}

	if inv.Music != nil {
		t.Errorf("expected nil Music for legacy invitation, got %+v", inv.Music)
	}
	if inv.SplashScreen != nil {
		t.Errorf("expected nil SplashScreen for legacy invitation, got %+v", inv.SplashScreen)
	}
	if inv.HasMusic() {
		t.Errorf("expected HasMusic() to be false for legacy invitation")
	}
	if inv.HasSplashScreen() {
		t.Errorf("expected HasSplashScreen() to be false for legacy invitation")
	}

	if err := inv.Validate(); err != nil {
		t.Fatalf("expected legacy invitation to validate cleanly, got: %v", err)
	}

	// Re-marshal to verify music and splash_screen are omitted
	marshaled, err := json.Marshal(&inv)
	if err != nil {
		t.Fatalf("failed to re-marshal legacy invitation: %v", err)
	}
	if strings.Contains(string(marshaled), `"music"`) {
		t.Errorf("expected marshaled JSON to omit 'music', got: %s", string(marshaled))
	}
	if strings.Contains(string(marshaled), `"splash_screen"`) {
		t.Errorf("expected marshaled JSON to omit 'splash_screen', got: %s", string(marshaled))
	}
}

func TestComputeSectionSurfaces(t *testing.T) {
	t.Run("Default Auto Alternation after Hero and Details", func(t *testing.T) {
		hero := &HeroSection{SectionType: SectionHero}
		details := &DetailsSection{SectionType: SectionDetails}
		quote := &QuoteSection{SectionType: SectionQuote, Text: "Love is patient"}
		carousel := &CarouselSection{SectionType: SectionCarousel, Images: []CarouselImage{{URL: "/img/1.jpg"}}}
		timeline := &TimelineSection{SectionType: SectionTimeline, Items: []TimelineItem{{Time: "18:00", Title: "Arrival"}}}
		dressCode := &DressCodeSection{SectionType: SectionDressCode, Name: "Formal"}

		inv := &Invitation{
			Sections: []Section{hero, details, quote, carousel, timeline, dressCode},
		}

		surfaces := inv.ComputeSectionSurfaces()

		if surfaces[hero] != SurfaceContrast {
			t.Errorf("expected hero to have contrast surface, got %s", surfaces[hero])
		}
		if surfaces[details] != SurfaceLight {
			t.Errorf("expected details to have light surface, got %s", surfaces[details])
		}
		// Details finishes on contrast (where-strip), so next auto block should be light
		if surfaces[quote] != SurfaceLight {
			t.Errorf("expected quote following details to alternate to light, got %s", surfaces[quote])
		}
		// Carousel follows quote (light), so it should alternate to contrast
		if surfaces[carousel] != SurfaceContrast {
			t.Errorf("expected carousel following quote to alternate to contrast, got %s", surfaces[carousel])
		}
		// Timeline follows carousel (contrast), so it should alternate to light
		if surfaces[timeline] != SurfaceLight {
			t.Errorf("expected timeline to alternate to light, got %s", surfaces[timeline])
		}
		// DressCode follows timeline (light), so it should alternate to contrast
		if surfaces[dressCode] != SurfaceContrast {
			t.Errorf("expected dressCode to alternate to contrast, got %s", surfaces[dressCode])
		}
	})

	t.Run("Explicit User Overrides are Respected", func(t *testing.T) {
		quote := &QuoteSection{SectionType: SectionQuote, Surface: SurfaceContrast, Text: "Love is patient"}
		carousel := &CarouselSection{SectionType: SectionCarousel, Surface: SurfaceContrast, Images: []CarouselImage{{URL: "/img/1.jpg"}}}
		timeline := &TimelineSection{SectionType: SectionTimeline, Surface: SurfaceAuto, Items: []TimelineItem{{Time: "18:00", Title: "Arrival"}}}

		inv := &Invitation{
			Sections: []Section{quote, carousel, timeline},
		}

		surfaces := inv.ComputeSectionSurfaces()

		// Both quote and carousel explicitly set contrast
		if surfaces[quote] != SurfaceContrast {
			t.Errorf("expected explicit quote contrast surface, got %s", surfaces[quote])
		}
		if surfaces[carousel] != SurfaceContrast {
			t.Errorf("expected explicit carousel contrast surface, got %s", surfaces[carousel])
		}
		// Timeline is auto, following carousel (contrast), so it alternates to light
		if surfaces[timeline] != SurfaceLight {
			t.Errorf("expected timeline to alternate to light following explicit contrast, got %s", surfaces[timeline])
		}
	})

	t.Run("Surface JSON Unmarshaling and Marshaling", func(t *testing.T) {
		raw := `{
			"version": "1.0",
			"slug": "surface-test",
			"title": "Surface Test",
			"date_start": "2026-09-19T18:00:00Z",
			"location": {"name": "Venue", "address": "123 St"},
			"theme": {"id": "botanical-elegance"},
			"sections": [
				{ "type": "quote", "surface": "contrast", "text": "A great quote" },
				{ "type": "text", "surface": "light", "text": "Some text" },
				{ "type": "image", "url": "/test.webp" }
			]
		}`

		var inv Invitation
		if err := json.Unmarshal([]byte(raw), &inv); err != nil {
			t.Fatalf("failed to unmarshal JSON with surfaces: %v", err)
		}

		if len(inv.Sections) != 3 {
			t.Fatalf("expected 3 sections, got %d", len(inv.Sections))
		}

		if inv.Sections[0].GetSurface() != "contrast" {
			t.Errorf("expected quote surface 'contrast', got %q", inv.Sections[0].GetSurface())
		}
		if inv.Sections[1].GetSurface() != "light" {
			t.Errorf("expected text surface 'light', got %q", inv.Sections[1].GetSurface())
		}
		if inv.Sections[2].GetSurface() != "" {
			t.Errorf("expected image surface empty, got %q", inv.Sections[2].GetSurface())
		}

		// Verify re-marshaling respects omitempty when empty
		marshaled, err := json.Marshal(&inv)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}
		marshaledStr := string(marshaled)
		if !strings.Contains(marshaledStr, `"surface":"contrast"`) {
			t.Errorf("expected marshaled JSON to retain quote surface contrast")
		}
	})
}

func TestSaveTheDateConfig(t *testing.T) {
	t.Run("JSON Unmarshaling and Defaults", func(t *testing.T) {
		raw := `{
			"version": "1.0",
			"slug": "julieta-mis-15",
			"title": "Mis 15 Julieta",
			"date_start": "2026-11-14T21:00:00Z",
			"location": {"name": "Salón La Reserva", "address": "Pilar, Buenos Aires"},
			"theme": {"id": "botanical-elegance"},
			"save_the_date": {
				"enabled": true,
				"layout": "full-bleed",
				"placement": "bottom-left",
				"phrase": "Reserva la fecha",
				"event_type": "Mis 15",
				"name": "Julieta",
				"location_hint": "Buenos Aires",
				"footer_note": "Invitación formal próximamente",
				"cover_image_url": "/static/img/hero-cumple.webp",
				"show_countdown": true
			}
		}`

		var inv Invitation
		if err := json.Unmarshal([]byte(raw), &inv); err != nil {
			t.Fatalf("failed to unmarshal JSON with save_the_date: %v", err)
		}

		if !inv.HasSaveTheDate() {
			t.Errorf("expected HasSaveTheDate to be true")
		}

		std := inv.SaveTheDate
		if std.DisplayPhrase() != "Reserva la fecha" {
			t.Errorf("expected phrase 'Reserva la fecha', got %q", std.DisplayPhrase())
		}
		if std.DisplayName(&inv) != "Julieta" {
			t.Errorf("expected name 'Julieta', got %q", std.DisplayName(&inv))
		}
		if std.DisplayLocation(&inv) != "Buenos Aires" {
			t.Errorf("expected location 'Buenos Aires', got %q", std.DisplayLocation(&inv))
		}
		if std.DisplayFooterNote() != "Invitación formal próximamente" {
			t.Errorf("expected footer note 'Invitación formal próximamente', got %q", std.DisplayFooterNote())
		}
		if std.PlacementClass() != "std-place--bottom-left" {
			t.Errorf("expected placement class 'std-place--bottom-left', got %q", std.PlacementClass())
		}
		if std.LayoutClass() != "std-layout--full-bleed" {
			t.Errorf("expected layout class 'std-layout--full-bleed', got %q", std.LayoutClass())
		}
		if !std.ShowCountdown {
			t.Errorf("expected ShowCountdown to be true")
		}
	})

	t.Run("Fallbacks when fields are omitted", func(t *testing.T) {
		inv := Invitation{
			Title:     "Event Title Fallback",
			DateStart: time.Date(2026, time.November, 14, 20, 0, 0, 0, time.UTC),
			Location: Location{
				Name:    "Venue Name",
				Address: "City, Country",
			},
			SaveTheDate: &SaveTheDateConfig{
				Enabled: true,
			},
		}

		std := inv.SaveTheDate
		if std.DisplayPhrase() != "Save the Date" {
			t.Errorf("expected default phrase 'Save the Date', got %q", std.DisplayPhrase())
		}
		if std.DisplayName(&inv) != "Event Title Fallback" {
			t.Errorf("expected fallback name from title, got %q", std.DisplayName(&inv))
		}
		if std.DisplayLocation(&inv) != "City, Country" {
			t.Errorf("expected fallback location from address, got %q", std.DisplayLocation(&inv))
		}
		if std.PlacementClass() != "std-place--bottom-center" {
			t.Errorf("expected default placement class 'std-place--bottom-center', got %q", std.PlacementClass())
		}
		if std.LayoutClass() != "std-layout--full-bleed" {
			t.Errorf("expected default layout class 'std-layout--full-bleed', got %q", std.LayoutClass())
		}
		// Formatted date should contain Saturday and Noviembre in Spanish
		dateStr := std.DisplayDate(&inv)
		if !strings.Contains(dateStr, "14") || !strings.Contains(dateStr, "2026") {
			t.Errorf("expected formatted date with 14 and 2026, got %q", dateStr)
		}
	})

	t.Run("Disabled SaveTheDate", func(t *testing.T) {
		inv := Invitation{
			SaveTheDate: &SaveTheDateConfig{
				Enabled: false,
			},
		}
		if inv.HasSaveTheDate() {
			t.Errorf("expected HasSaveTheDate to be false when enabled=false")
		}
	})
}

