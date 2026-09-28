# Iteration 2: JSON Schema & Domain Modeling

## 1. Goal & Rationale
Formalize `music` and `splash_screen` configuration in both the JSON Schema ([`schema/invitation.schema.json`](file:///home/nano/projects/invitation/schema/invitation.schema.json)) and the Go Domain Model ([`pkg/domain/invitation.go`](file:///home/nano/projects/invitation/pkg/domain/invitation.go)).

### Should `splash_screen` be added to the JSON Schema?
**Yes, absolutely.** 
Adding `splash_screen` and `music` directly to `schema/invitation.schema.json` ensures:
1. **Single Source of Truth**: The JSON Schema acts as the strict contract for imports, exports, API payloads, and the Admin Builder Studio.
2. **Deterministic Validation**: Enables strict structural checks (`validate.go`) so malformed properties or invalid URLs are rejected before hitting storage.
3. **Static Generation (SSG) Compatibility**: Seeds and static JSON fixtures remain fully typed and predictable across SSR and SSG pipelines.

---

## 2. JSON Schema Specifications

In [`schema/invitation.schema.json`](file:///home/nano/projects/invitation/schema/invitation.schema.json), add top-level properties `music` and `splash_screen`:

```json
{
  "properties": {
    "music": {
      "$ref": "#/definitions/music_config",
      "description": "Optional background audio track and playback controls"
    },
    "splash_screen": {
      "$ref": "#/definitions/splash_screen_config",
      "description": "Optional tap-to-open welcome screen and envelope gate"
    }
  },
  "definitions": {
    "music_config": {
      "type": "object",
      "required": ["url"],
      "properties": {
        "url": {
          "type": "string",
          "description": "Local path (/uploads/...) or public HTTPS URL to the audio file"
        },
        "title": {
          "type": "string",
          "description": "Song title or artist name displayed on the floating control (e.g. 'A Thousand Years - Christina Perri')"
        },
        "autoplay": {
          "type": "boolean",
          "default": true,
          "description": "Whether playback should start immediately upon the initial user interaction"
        },
        "loop": {
          "type": "boolean",
          "default": true,
          "description": "Whether the audio track should repeat seamlessly when finished"
        }
      }
    },
    "splash_screen_config": {
      "type": "object",
      "properties": {
        "enabled": {
          "type": "boolean",
          "default": true,
          "description": "Whether to display the splash screen before showing the invitation"
        },
        "title": {
          "type": "string",
          "description": "Headline text on the splash card (e.g. 'You are Invited', 'Lucas is turning 30')"
        },
        "message": {
          "type": "string",
          "description": "Warm introductory note or guest instructions"
        },
        "button_text": {
          "type": "string",
          "default": "Open Invitation",
          "description": "Call-to-action text on the envelope button"
        },
        "background_image_url": {
          "type": "string",
          "description": "Optional custom background or envelope texture photo"
        }
      }
    }
  }
}
```

---

## 3. Go Domain Model Extensions (`pkg/domain/invitation.go`)

### 3.1 Type Definitions
```go
// MusicConfig configures optional background audio track and player controls.
type MusicConfig struct {
    URL      string `json:"url"`
    Title    string `json:"title,omitempty"`
    Autoplay bool   `json:"autoplay"`
    Loop     bool   `json:"loop"`
}

func (m *MusicConfig) Validate() error {
    if m == nil {
        return nil
    }
    if strings.TrimSpace(m.URL) == "" {
        return errors.New("music config: url is required")
    }
    return nil
}

func (m *MusicConfig) DisplayTitle(fallback string) string {
    if m != nil && strings.TrimSpace(m.Title) != "" {
        return m.Title
    }
    return fallback
}

// SplashScreenConfig configures the welcome cover and tap-to-open gesture gate.
type SplashScreenConfig struct {
    Enabled            bool   `json:"enabled"`
    Title              string `json:"title,omitempty"`
    Message            string `json:"message,omitempty"`
    ButtonText         string `json:"button_text,omitempty"`
    BackgroundImageURL string `json:"background_image_url,omitempty"`
}

func (s *SplashScreenConfig) Validate() error {
    return nil
}

func (s *SplashScreenConfig) IsActive() bool {
    return s != nil && s.Enabled
}

func (s *SplashScreenConfig) CTAButtonText() string {
    if s != nil && strings.TrimSpace(s.ButtonText) != "" {
        return s.ButtonText
    }
    return "Open Invitation"
}
```

### 3.2 Updating `domain.Invitation`
In `pkg/domain/invitation.go`:
```go
type Invitation struct {
    Version      string              `json:"version"`
    Slug         string              `json:"slug"`
    Title        string              `json:"title"`
    Subtitle     string              `json:"subtitle,omitempty"`
    Hosts        []string            `json:"hosts,omitempty"`
    Description  string              `json:"description,omitempty"`
    DateStart    time.Time           `json:"date_start"`
    DateEnd      *time.Time          `json:"date_end,omitempty"`
    Timezone     string              `json:"timezone"`
    Location     Location            `json:"location"`
    Theme        ThemeConfig         `json:"theme"`
    Music        *MusicConfig        `json:"music,omitempty"`
    SplashScreen *SplashScreenConfig `json:"splash_screen,omitempty"`
    Sections     []Section           `json:"sections"`
}
```

### 3.3 Unmarshaling & Backward Compatibility
Update `UnmarshalJSON` in `pkg/domain/invitation.go` so `rawInvitation` includes:
```go
Music        *MusicConfig        `json:"music,omitempty"`
SplashScreen *SplashScreenConfig `json:"splash_screen,omitempty"`
```
And assign them to `inv.Music` and `inv.SplashScreen`.
Invitations without these keys will have `nil` pointers, guaranteeing 100% backward compatibility for all existing saved events.

---

## 4. Seed Updates & Test Suite

1. **Seed Updates (`seed/wedding.json` & `seed/birthday.json`)**:
   Add sample configurations:
   ```json
   "music": {
     "url": "/uploads/demo-wedding-track.mp3",
     "title": "Acoustic Sunset Vibe",
     "autoplay": true,
     "loop": true
   },
   "splash_screen": {
     "enabled": true,
     "title": "Lucas is turning 30",
     "message": "You're invited to celebrate chapter 30 with spritzes, sourdough pizza, and sunset dancing.",
     "button_text": "Open Invitation"
   }
   ```
2. **Domain Tests (`pkg/domain/invitation_test.go`)**:
   - Test JSON serialization and deserialization of `MusicConfig` and `SplashScreenConfig`.
   - Test validation errors when `Music.URL` is missing.
   - Verify defaults (`CTAButtonText()`, `DisplayTitle()`).
