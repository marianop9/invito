# Invito &mdash; Event Invitation Platform Plan & Context

## 1. Purpose & Core Philosophy

**Invito** is a lightweight, high-performance web platform built in **Go** to generate, preview, and share structured event invitations with attendance confirmation (RSVP).

### Core Design Principles
- **Schema-Driven**: Invitations are defined via a structured JSON Schema (`schema/invitation.schema.json`) specifying metadata, schedule/timeline, location, dress code, RSVP rules, and theme configuration.
- **Mobile-First Continuous Canvas**: Invitations render as a single, fluid sheet / digital event pass (`.inv-sheet`), eliminating disjointed floating cards and excessive spacing.
- **Concise & Scannable**: Copy is punchy and direct. Key event details (Date, Time, Venue, 1-click Map, Calendar Sync, RSVP) are immediately accessible.
- **Single-Binary Deployment**: HTML templates, CSS themes, client scripts, database migrations, and seed templates compile into a single standalone binary using Go's `embed.FS` with zero external runtime dependencies—ideal for hosting on a low-cost VPS.
- **Hybrid SSR + On-Demand SSG**: Server-Side Rendering (SSR) powers live RSVP submissions, dynamic OpenGraph meta previews, and `.ics` downloads, while an SSG exporter can generate standalone static HTML packages.

---

## 2. Technology Stack & Architecture

- **Language & Runtime**: Go 1.26+
- **HTTP Router & Middleware**: `github.com/go-chi/chi/v5` with `RequestID`, `RealIP`, `Logger`, `Recoverer`, and `Compress` (gzip).
- **Template Engine**: Go standard library `html/template` with embedded files (`embed.FS`).
- **Data Storage**:
  - *Current Prototype*: `pkg/storage/seed_store.go` (in-memory store initialized from `seed/*.json`).
  - *Target Persistent Layer*: Pure-Go SQLite (`modernc.org/sqlite`) with WAL mode for zero-CGO portability.
- **Styling**: Vanilla CSS with CSS Custom Properties, HSL color tokens, and Google Fonts (`Cormorant Garamond`, `Playfair Display`, `Space Grotesk`, `Outfit`, `Plus Jakarta Sans`, `Inter`).
- **Calendar Standard**: RFC 5545 iCalendar (`.ics`) generator (`pkg/calendar/ics.go`) and Google Calendar deep links.

---

## 3. Work Completed So Far

### ✅ Phase 1: Server Foundation & Routing
- Initialized Go module with Chi router in [`main.go`](file:///home/nano/projects/invitation/main.go) and [`pkg/server/server.go`](file:///home/nano/projects/invitation/pkg/server/server.go).
- Configured embedded asset filesystem in [`web/embed.go`](file:///home/nano/projects/invitation/web/embed.go).
- Implemented graceful server shutdown and logging.

### ✅ Phase 2: JSON Schema & Domain Layer
- Created [`schema/invitation.schema.json`](file:///home/nano/projects/invitation/schema/invitation.schema.json) specification.
- Built strongly-typed domain structs and helper methods in [`pkg/domain/invitation.go`](file:///home/nano/projects/invitation/pkg/domain/invitation.go) (date formatters, Google Calendar links, Schema.org JSON-LD microdata).
- Implemented strict validation in [`pkg/domain/validate.go`](file:///home/nano/projects/invitation/pkg/domain/validate.go).

### ✅ Phase 3: 4 Pre-Defined Visual Themes
1. **Botanical Elegance** ([`botanical-elegance.css`](file:///home/nano/projects/invitation/web/static/css/themes/botanical-elegance.css)) &mdash; Ivory, sage green, *Cormorant Garamond* (Weddings & formal dinners).
2. **Midnight Soirée** ([`midnight-soiree.css`](file:///home/nano/projects/invitation/web/static/css/themes/midnight-soiree.css)) &mdash; Luxury slate dark palette, champagne gold, *Playfair Display* (Galas & evening events).
3. **Modern Minimal** ([`modern-minimal.css`](file:///home/nano/projects/invitation/web/static/css/themes/modern-minimal.css)) &mdash; Monochromatic layout, cobalt blue, *Space Grotesk* (Tech summits & exhibitions).
4. **Golden Sunset** ([`golden-sunset.css`](file:///home/nano/projects/invitation/web/static/css/themes/golden-sunset.css)) &mdash; Terracotta & apricot, *Plus Jakarta Sans* (Milestone birthdays & summer parties).

### ✅ Phase 4: Streamlined Mobile-First Redesign
- Consolidated all invitation sections into a single continuous sheet (`.inv-sheet`) in [`web/static/css/invitation.css`](file:///home/nano/projects/invitation/web/static/css/invitation.css).
- Created a compact **Quick Details Strip** for Date, Time, Venue with 1-click Map, and action buttons.
- Replaced verbose seed data with concise, clear copy in [`seed/*.json`](file:///home/nano/projects/invitation/seed/).
- Compact timeline schedule, inline dress code swatches, and streamlined RSVP attendance form (`✓ Attending` / `✕ Decline`).

### ✅ Phase 5: Calendar Integration & Interactivity
- RFC 5545 iCalendar (`.ics`) generator in [`pkg/calendar/ics.go`](file:///home/nano/projects/invitation/pkg/calendar/ics.go) accessible at `GET /i/{slug}/calendar.ics`.
- Interactive client script in [`web/static/js/invitation.js`](file:///home/nano/projects/invitation/web/static/js/invitation.js) providing countdown clocks and AJAX RSVP submissions with real-time feedback.
- Automated unit and integration test coverage across all packages (`go test ./...` passes 100%).

---

## 4. Remaining Steps & Roadmap

### 🔄 Step 1: Persistent SQLite Storage Layer
- Create `pkg/storage/sqlite.go` using pure-Go SQLite (`modernc.org/sqlite`).
- Automatic table migration on startup:
  - `invitations`: `id`, `slug` (UNIQUE), `title`, `data` (JSON document), `created_at`, `updated_at`.
  - `rsvps`: `id`, `invitation_slug`, `name`, `email`, `attending`, `guest_count`, `dietary_needs`, `song_request`, `personal_message`, `created_at`.
- Auto-seed default templates from `seed/*.json` if database is empty on first boot.

### 🔄 Step 2: Interactive Invitation Preview & Sandbox (`/preview`)
- Create a split-screen playground (`web/templates/sandbox.html`, `web/static/js/sandbox.js`, `web/static/css/sandbox.css`).
- Left pane: Live form fields / raw JSON editor + Theme selector + Template loader.
- Right pane: Real-time mobile/desktop simulated viewport preview.
- "Save / Export" action buttons to test saving to SQLite or downloading JSON.

### 🔄 Step 3: Host Admin Panel & RSVP Management (`/admin/invitations/{slug}/rsvps`)
- Host dashboard template (`web/templates/admin_rsvps.html`) displaying confirmed guests, declines, party sizes, and dietary notes.
- Summary attendance cards (Total Attending, Total Declines, Total Party Count).
- CSV export endpoint: `GET /api/invitations/{slug}/rsvps.csv`.

### 🔄 Step 4: Standalone Static Site Export (SSG)
- Complete SSG exporter in `pkg/renderer/ssg.go`.
- Endpoint `GET /api/invitations/{slug}/export` to download a self-contained static HTML/CSS zip bundle for static hosting.

### 🔄 Step 5: Visual Enhancements (Icons & Refinements)
- Integrate crisp inline SVG icons for timeline event types (ceremony, cocktail, dinner, party, speech).
- Add subtle micro-animations for RSVP submission success.

---

## 5. Project File Map

```
/home/nano/projects/invitation/
├── go.mod                       # Go module definitions
├── go.sum                       # Dependency checksums
├── main.go                      # Application entry point & graceful shutdown
├── plan.md                      # Architecture, context, and roadmap document
├── schema/
│   └── invitation.schema.json   # JSON Schema definition
├── seed/                        # Starter templates
│   ├── wedding.json             # Botanical Elegance wedding template
│   ├── gala.json                # Midnight Soirée charity gala template
│   ├── tech_meetup.json         # Modern Minimal tech summit template
│   └── birthday.json            # Golden Sunset birthday template
├── pkg/
│   ├── domain/                  # Domain models & validation logic
│   │   ├── invitation.go
│   │   ├── validate.go
│   │   └── invitation_test.go
│   ├── calendar/                # RFC 5545 iCalendar (.ics) generator
│   │   ├── ics.go
│   │   └── ics_test.go
│   ├── renderer/                # SSR template engine & SSG exporter
│   │   ├── renderer.go
│   │   └── renderer_test.go
│   ├── storage/                 # Storage layer (MemoryStore & SQLite)
│   │   ├── seed_store.go
│   │   └── seed_store_test.go
│   └── server/                  # Chi HTTP server & route handlers
│       ├── server.go
│       └── server_test.go
└── web/
    ├── embed.go                 # go:embed directive for templates and static files
    ├── templates/               # Go HTML templates
    │   ├── base.html            # Base shell layout for platform pages
    │   ├── index.html           # Showcase landing page
    │   ├── invitation.html      # Continuous single-sheet invitation layout
    │   └── partials/            # Modular invitation components
    │       ├── hero.html
    │       ├── details.html
    │       ├── timeline.html
    │       ├── dress_code.html
    │       ├── rsvp_form.html
    │       ├── faqs.html
    │       └── registry.html
    └── static/                  # Static assets
        ├── css/
        │   ├── base.css         # Design tokens & platform styles
        │   ├── invitation.css   # Mobile-first continuous sheet styles
        │   └── themes/          # Pre-defined theme palettes
        │       ├── botanical-elegance.css
        │       ├── midnight-soiree.css
        │       ├── modern-minimal.css
        │       └── golden-sunset.css
        └── js/
            └── invitation.js    # Client countdown & AJAX RSVP handler
```

---

## 6. Developer Commands

```bash
# Run all automated tests
go test ./... -v

# Run the local server (listens on http://localhost:8080)
go run main.go --port 8080

# Build standalone production binary
go build -o bin/invitation main.go
```
