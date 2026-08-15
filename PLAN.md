# Invito &mdash; Event Invitation Platform Plan & Roadmap

## 1. Purpose & Core Philosophy

**Invito** is a lightweight, high-performance web platform built in **Go** to generate, preview, and share structured event invitations with attendance confirmation (RSVP).

### Core Principles
- **Schema-Driven**: Invitations are defined via a structured JSON Schema (`schema/invitation.schema.json`) specifying metadata, schedule/timeline, location, dress code, RSVP rules, images, and theme configuration.
- **Mobile-First Continuous Canvas**: Invitations render as a single, fluid sheet / digital event pass (`.inv-sheet`), eliminating disjointed floating cards and excessive spacing.
- **Concise & Scannable**: Punchy copy, immediate access to Date, Time, Venue, 1-click Map, Calendar Sync, and RSVP.
- **Softer, Elegant Aesthetics**: Focus on warm, inviting palettes (Botanical Elegance & Golden Sunset) over aggressive dark or high-contrast neon tones.
- **Single-Binary Deployment**: HTML templates, CSS themes, client scripts, and images compile into a single standalone binary using Go's `embed.FS`.

---

## 2. Interactive Roadmap & Priority Checklist

### ✅ Completed Milestones
- [x] **Go Server Foundation**: Chi router, production middlewares (Logger, Recoverer, Compress), and graceful shutdown.
- [x] **Embedded Template System**: `embed.FS` configuration in `web/embed.go`.
- [x] **JSON Schema & Domain Layer**: `schema/invitation.schema.json` and `pkg/domain` models with strict validation.
- [x] **Calendar Sync**: Dynamic RFC 5545 iCalendar (`.ics`) generator and Google Calendar URL builder.
- [x] **Streamlined Mobile Canvas**: Continuous single-sheet layout (`.inv-sheet`), compact quick details strip, inline time badges, and AJAX RSVP confirmation.
- [x] **Modular Content Sections**: Text/quote statement section (`sections.message`), static image block (`sections.image`), and closing greeting with host sign-off (`sections.closing`).

---

### 🎯 Immediate Priority: Working Demo Excellence (Focus on Wedding & Birthday)

- [x] **Image Support in Invitation Schema & Domain**:
  - [x] Add `cover_image_url` and section banner support to `sections.hero` and domain models.
  - [x] Add `carousel` section to JSON Schema (`title`, array of `{ url, caption, alt }`).
  - [x] Update `pkg/domain/invitation.go` with Carousel structs and helper methods.
- [x] **Static Asset Serving for Seed Images**:
  - [x] Static images for demo (`demo-hero.webp`, `demo-carousel-1.webp`, `demo-carousel-2.webp`, `demo-carousel-3.webp`) are accessible under `web/static/img/`.
- [x] **Full-Width Hero Image & Section Images**:
  - [x] Render full-bleed/rounded hero cover image in `web/templates/partials/hero.html` with graceful fallback if omitted.
- [x] **Auto-Scrolling Viewport-Aware Image Carousel**:
  - [x] Build `web/templates/partials/carousel.html` partial (clean slide cards with indicator dots, no arrows or bottom captions).
  - [x] Implement smooth viewport-triggered auto-scrolling with infinite wrap-around and dot indicators in `web/static/css/invitation.css` and `web/static/js/invitation.js`.
- [x] **Refine & Showcase the 2 Core Softer Demos**:
  - [x] Update `seed/wedding.json` (Botanical Elegance) to integrate `demo-hero.webp` and the 3 carousel images.
  - [x] Polish `seed/birthday.json` (Golden Sunset) with warm aesthetic accents.
  - [x] Update showcase landing page (`web/templates/index.html`) to focus exclusively on the Wedding and Birthday demos.

---

### Demo Implementation

- [ ] **Standalone Static Site Exporter (SSG)**:
  - [ ] Export invitation to standalone static HTML/CSS zip bundle. Used to upload to a static site hosting platform for easy access to the demos.
- [ ] **Upload Demos to Static Site Hosting**
  - [ ] Host the generated static sites. Potential targets include GitHub Pages, which should already be setup for this project, configuration in `.github/workflows/static.yml`.
- [ ] **Use Demo Feedback and Implement MVP**
  - [ ] Once the initial demos and invitation sections are developed, a custom invitation will be developed based on the received feedback.

---

### 📦 Future Backlog (Post-Demo Features)

- [ ] **Persistent SQLite Storage Layer (`pkg/storage/sqlite.go`)**:
  - [ ] Pure-Go SQLite driver (`modernc.org/sqlite`).
  - [ ] Auto-migrations for `invitations` and `rsvps` tables.
  - [ ] Auto-seed default templates on first boot.
- [ ] **Interactive Split-Screen Preview / Sandbox (`/preview`)**:
  - [ ] Live visual editor + raw JSON editor with simulated mobile viewport.
- [ ] **Host Admin Dashboard (`/admin/invitations/{slug}/rsvps`)**:
  - [ ] Guest list table with confirmed attendees, declines, and dietary requirements.
  - [ ] CSV export endpoint (`GET /api/invitations/{slug}/rsvps.csv`).
- [ ] **Visual Iconography Enhancements**:
  - [ ] Inline SVG icons for timeline event types (rings, toast, dinner, music).

---

## 3. Active Demo Profiles

| Demo | Theme | Palette & Style | Key Features |
| :--- | :--- | :--- | :--- |
| **Sarah & Alex's Wedding** (`/i/sarah-and-alex-wedding`) | `botanical-elegance` | Warm ivory, sage green, champagne gold (*Cormorant Garamond*) | Hero cover image, 3-image story carousel, ceremony & reception schedule, garden formal dress code, RSVP |
| **Lucas is 30** (`/i/lucas-30th-birthday-sunset-fiesta`) | `golden-sunset` | Warm terracotta, apricot, soft cream (*Plus Jakarta Sans*) | Sunset party timeline, pizza & cocktail bar details, summer chic dress code, RSVP |

---

## 4. Key Files Map

```
/home/nano/projects/invitation/
├── plan.md                      # Roadmap, priority checklist & context
├── schema/
│   └── invitation.schema.json   # JSON Schema definition (including images/carousel)
├── seed/                        # Active demo templates & assets
│   ├── wedding.json             # Botanical Elegance wedding demo
│   ├── birthday.json            # Golden Sunset birthday demo
│   └── img/                     # Demo images (hero.webp, carousel-1..3.webp)
├── pkg/
│   ├── domain/                  # Go domain models & validation
│   ├── calendar/                # RFC 5545 iCal generator
│   ├── renderer/                # SSR template engine
│   ├── storage/                 # MemoryStore seed loader
│   └── server/                  # Chi router & HTTP handlers
└── web/
    ├── embed.go                 # Embedded asset bundle
    ├── templates/
    │   ├── base.html
    │   ├── index.html           # Showcase landing page (Wedding & Birthday)
    │   ├── invitation.html      # Single-sheet canvas
    │   └── partials/            # Hero, Details, Timeline, Carousel, RSVP, etc.
    └── static/
        ├── css/                 # Base, invitation, and theme stylesheets
        └── js/                  # Carousel scroll, countdown, and AJAX RSVP
```

---

## 5. Developer Commands

```bash
# Run all automated tests
go test ./... -v

# Run the local server (listens on http://localhost:8080)
go run main.go --port 8080

# Build standalone binary
go build -o bin/invitation main.go
```

NOTE: It is not necessary to run the build or tests. After implementing the code for the requested features and updating the test suite, I will manually verify the results and run the tests.  