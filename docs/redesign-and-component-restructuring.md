# Invitation UI/UX Redesign & Component Restructuring

This document outlines the diagnosis, architectural principles, and phased implementation plan for restructuring the digital invitation components in `web/templates/partials/` and updating the design system.

---

## 1. Executive Summary & Context

After evaluating our demo invitations (`/seed` folder, `demo-cari-45`, `demo-recepcion-xyz`, `sarah-and-alex-wedding`) against industry benchmarks (Partiful, Paperless Post, Luma) and real-world high-performing event invitations (analyzed from reference samples in `tmp/`), we identified that our current invitations feel like **dense, administrative web documents** rather than **sleek, celebratory invitations**.

Rather than jumping directly into styling or color picking, this initiative adopts a **structure-first approach**: fixing the information architecture and HTML templates before layering on visual themes and color palettes.

---

## 2. Current Diagnosis & Architectural Decisions

### 2.1. Text Overload & Cognitive Friction
* **Fragmented narrative**: Invitations were broken into too many individual text blocks (hero eyebrow, hero title, hero vibe, quote, details notes, text announcements, FAQs, closing words).
* **Reading like a manual**: Non-critical instructions (e.g., *"Recordar traer cubiertos"*, payment warnings, receipt rules) were displayed as prominent alert boxes or separate sections, distracting from the excitement of the event.
* **Uncollapsible FAQs**: Every question and answer was displayed simultaneously, creating an excessively long vertical scroll on mobile screens.

### 2.2. Button Proliferation & Calendar Actions
* **Calendar actions removed from stationery**: Non-technical guests on mobile were confused by technical download prompts (like `.ics` files) and competing links (Google Calendar vs `.ics` vs Google Maps) before even reading the invitation.
* **Architectural Decision**: All calendar button actions have been **completely excluded** from the invitation templates. The When block focuses purely on clean, legible date and time coordinates with **zero button clutter**. The backend endpoint `/i/{slug}/calendar.ics` remains active for calendar integrations.
* **Single Venue Action**: The Where block retains only a single, clear navigation button: `CÓMO LLEGAR`.

### 2.3. The "Color Scheme Void"
* **Flat white documents**: Although themes define CSS variables, almost every section defaults to a pure white background (`#FFFFFF`) with 1px gray borders (`#E5E1D8`). 
* **Underutilized palettes**: In themes like `midnight-soiree` (intended for formal galas and evening celebrations), the background remains white with dark navy text. The theme's accent color is only applied to small badges and buttons.
* **Lack of tonal rhythm**: The page lacks atmospheric depth, background contrast, and textural warmth (to be addressed in Phase 2 & 3).

### 2.4. Boxy, Administrative Components
* **Dashboard-style badges**: Icons in `details.html` were enclosed in rigid rounded-square boxes (`.detail-icon-badge`), resembling a SaaS settings dashboard rather than fine stationery. Replaced with delicate line-art icons floating above centered headings.
* **Hero clutter & hidden countdown**: The hero was crowded with 3 stacked text paragraphs (`Eyebrow`, `Title`, `Vibe`), a redundant date/time pill, and a browser-style "Desliza hacia abajo" scroll cue, while the countdown was confined to a tiny toggle badge.
* **Architectural Decision**: Hero de-cluttered to a pure 2-element hierarchy (`Eyebrow` + `Title`), removing the redundant `Vibe` paragraph, date pill, and scroll cue. The celebratory countdown is elevated to a bold milestone display (`DÍAS : HORAS : MIN : SEG`).

---

## 3. Reference Analysis & Architectural Benchmarks

Analysis of successful production invitations revealed five core structural principles:

```
┌────────────────────────────────────────────────────────┐
│  [Hero Image / Photo Grid]                             │
├────────────────────────────────────────────────────────┤
│  LIGHT BLOCK (White / Warm Linen)                      │
│  Centered Eyebrow + Big Bold Title                     │
│  Full Celebratory Countdown (DÍAS : HORAS : MIN : SEG) │
├────────────────────────────────────────────────────────┤
│  CONTRAST BLOCK (Theme Primary / Contrast Tone)        │
│  ─── Short Centered Emotional Quote ───                │
├────────────────────────────────────────────────────────┤
│  LIGHT BLOCK (White / Warm Ivory)                      │
│  Fine Line Icon (Calendar) + ¿CUÁNDO? + Date/Time      │
│  (Clean, legible coordinates, zero button clutter)     │
├────────────────────────────────────────────────────────┤
│  CONTRAST BLOCK (Theme Primary / Contrast Tone)        │
│  Fine Line Icon (Pin) + ¿DÓNDE? + Venue Name           │
│  [ CÓMO LLEGAR ] (Single clear outline button)         │
├────────────────────────────────────────────────────────┤
│  LIGHT BLOCK (White / Warm Ivory)                      │
│  Fine Line Icon (Diamond) + DRESS CODE + Style Name    │
│  (Tactile rounded color swatches)                      │
└────────────────────────────────────────────────────────┘
```

### Key Principles Extracted:
1. **The Alternating Rhythm (`Light` $\leftrightarrow$ `Contrast`)**:
   By strictly alternating background tones between sections (e.g. Light $\rightarrow$ Contrast $\rightarrow$ Light), 50% of the screen immediately reflects the theme's core color without complex styling.
2. **"One Block = One Focus" Rule**:
   Each section focuses on a single piece of information (*When*, *Where*, *Dress Code*, *Photos*) rather than cramming multiple details together.
3. **At Most One Action per Block**:
   Blocks only present a single relevant action (e.g., only `CÓMO LLEGAR` under Where, only `CONFIRMAR ASISTENCIA` under RSVP).
4. **Delicate Line-Art Iconography**:
   Fine line icons (calendar, pin, diamond, camera, gift) float cleanly above centered headings instead of being trapped inside colored square boxes.
5. **Centered, Minimal Typography**:
   All headings are centered with tracked uppercase letters. Text is concise, avoiding multi-paragraph explanations.
6. **Tone-Agnostic Foundation**:
   This layout structure works equally well for:
   * **Vibrant & Bold Events**: High contrast black/white or neon pink (quinceañeras, parties).
   * **Delicate & Organic Events**: Soft warm linen (`#F4F2EC`), eucalyptus sage (`#2A4738`), and romantic serif typography (`Cormorant Garamond`), preserving the warm, artisanal feel of `demo-cari-45`.

---

## 4. Component-by-Component Refactoring Blueprint

| Partial Template | Previous Implementation | Implemented Architecture (Phase 1) | Status |
| :--- | :--- | :--- | :--- |
| **`hero.html` & `hero_banner.html`** | Stacked 3 text blocks (`Eyebrow`, `Title`, `Vibe`), hidden countdown in small meta badge, date chip, and swipe cue | **De-Cluttered Hero & Full Countdown**: Pure 2-level hierarchy: `Eyebrow` (Milestone tag) + `Title` (Host/Event). Removed `Vibe`, date chip, and scroll cue. Elevated countdown to a prominent milestone display (`DÍAS : HORAS : MIN : SEG`) updating live each second. | **Implemented** |
| **`details.html`** | Single gray strip with Date, Time, Venue, 3 buttons (Google, `.ics`, Maps), and boxy badges | **Separate / Clear Focused Coordinates**: Centered line-art icons. Distinct When (`#section-when`) and Where (`#section-where`) blocks. Pure date/time coordinates with **zero** calendar buttons. Single outline `CÓMO LLEGAR` navigation button under Where. | **Implemented** |
| **`faqs.html`** | Flat static list displaying all questions and answers simultaneously | **Collapsible Accordions**: Semantic HTML5 `<details class="faq-accordion-item">` and `<summary>` elements with animated Tabler `chevron-down.svg` indicator. | **Implemented** |
| **`dress_code.html`** | Tiny 18px color dots next to text | **Mood Badge & Palette Strip**: Centered line-art `diamond.svg` icon, bold attire label (`.dress-code-badge`), and generous, tactile rounded color swatches (`.dress-code-swatch`). | **Implemented** |
| **`rsvp_external.html`** | 4 separate paragraphs and alert-style notice boxes | **Unified Conversion Card**: Single cohesive card (`.rsvp-conversion-card`) with title, deadline badge, primary CTA button (`.btn-rsvp-primary`), and consolidated payment/receipt notes in a tidy info group (`.rsvp-notes-group`). | **Implemented** |
| **`rsvp_form.html`** | Form inputs with heavy borders and boxy radio inputs | **Streamlined Guest Form**: Clean segmented Yes/No pill switch (`.rsvp-segmented-control`), clean inputs, and high-contrast primary submit button. | **Implemented** |
| **`timeline.html`** | Rigid timetable list with fixed-width time badges | **Connected Milestone Track**: Vertical storyline with track line (`.timeline-track-line`), illuminated node markers (`.timeline-marker`), and spaced milestone cards. | **Implemented** |
| **`quote.html` & `text.html`** | Plain left-aligned paragraphs | **Framed Editorial Callouts**: Centered typography framed by delicate horizontal hairline rules (`─── Quote ───`). | **Implemented** |

---

## 5. Phased Implementation Roadmap

### Phase 1: Structural Refactoring of Templates (`HTML First`) — [COMPLETED]
- [x] **Refactor `details.html`**:
  * Separated When and Where into distinct visual blocks (`#section-when`, `#section-where`) with line icons.
  * Completely excluded calendar button actions from stationery templates (kept backend `/i/{slug}/calendar.ics` route).
  * Retained only `CÓMO LLEGAR` as the venue action button.
- [x] **Refactor `hero.html` & `hero_banner.html`**:
  * Extracted countdown into a prominent, multi-unit block (`DÍAS : HORAS : MIN : SEG`) with live ticking in `invitation.js`.
  * De-cluttered hero: removed redundant `Vibe` text paragraph, date/time chip, and "Desliza hacia abajo" scroll prompt.
- [x] **Refactor `faqs.html`**:
  * Converted static FAQ items into semantic `<details>` and `<summary>` accordions with SVG chevron indicator.
- [x] **Refactor `dress_code.html`**:
  * Adopted centered layout with line icon (`diamond.svg`), attire badge, and tactile palette swatches.
- [x] **Refactor `rsvp_external.html`**:
  * Consolidated prompt, CTA, and payment/WhatsApp notes into a single cohesive card structure (`.rsvp-conversion-card`).
- [x] **Refactor `timeline.html`**:
  * Updated list structure to support connected milestone track nodes (`.timeline-track-container`).
- [x] **Refactor `quote.html` & `text.html`**:
  * Centered editorial callouts framed by delicate horizontal rules (`.inv-quote-frame`).
- [x] **Refactor `rsvp_form.html`**:
  * Modernized segmented Yes/No switch and primary submit button.
- [x] **Icon Assets & Scripts**:
  * Added `chevron-down.svg` and `diamond.svg` to `web/static/icons/`.
  * Added structural CSS scaffolding in `web/static/css/invitation.css`.
  * Decoupled SQLite DB from `seed/` JSON files and updated test suites (`go test ./...` 100% passing).

---

### Phase 2: Layout & Alternating Rhythm (`CSS Foundation`) — [NEXT]
1. Introduce alternating section classes (`.inv-block--light`, `.inv-block--contrast`) and automatic `:nth-child` rhythm.
2. Add clean, centered flexbox/grid layout styles for line icons, headings, and single buttons.
3. Add smooth accordion disclosure transitions in `faqs.html`.
4. Style the countdown with large, bold tabular figures and small uppercase tracking labels.

---

### Phase 3: Thematic Depth & Color Palettes
1. Upgrade existing themes in `web/static/css/themes/`:
   * **`botanical-elegance`**: Warm linen canvas (`#F4F2EC`), delicate ivory cards, deep forest contrast sections (`#2A4738`), soft sage borders, and *Cormorant Garamond* serif headings.
   * **`midnight-soiree`**: True nocturnal palette—deep obsidian slate (`#0B121E`), champagne gold accents (`#E2C082`), and luminous text.
   * **`golden-sunset`**: Warm ecru, soft sand, and rich bronze-gold accents.
   * **`modern-minimal`**: Clean blush or monochrome tones with high contrast.
2. Ensure theme overrides dynamically set both light and contrast surface tokens.

---

### Phase 4: Verification & Demo Updating
1. Update seed files (`seed/demo-cari-45.json`, `seed/demo-recepcion-xyz.json`, `seed/wedding.json`) to verify each theme.
2. Verify mobile viewport responsiveness across standard screen widths (360px, 390px, 420px, tablet/desktop).
3. Run test suites (`go test ./...`) to ensure full template and SSG rendering compatibility.
