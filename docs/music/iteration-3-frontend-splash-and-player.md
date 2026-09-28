# Iteration 3: Front-End Splash Screen & Floating Music Player

## 1. Goal & Rationale
Deliver the guest-facing frontend experience:
1. **Interactive Splash Screen ("Tap to Open")**: An elegant fullscreen greeting overlay that serves as the explicit user gesture required to unlock unmuted audio playback on modern mobile browsers.
2. **Audio Playback Engine**: An unobtrusive native `<audio>` element with soft volume fade-in.
3. **Floating Music Controller (FAB)**: A theme-aware, fixed-position button that indicates playback status (animated vinyl rotation and audio equalizer waves) and provides play/pause toggling.

---

## 2. Template Architecture (`web/templates/invitation.html`)

### 2.1 Splash Screen Markup
Placed at the top of `<body>`, rendered when `.SplashScreen.IsActive`:

```html
{{ if and .SplashScreen .SplashScreen.Enabled }}
<div id="inv-splash-overlay" class="inv-splash-overlay" 
     {{ if .SplashScreen.BackgroundImageURL }}style="background-image: url('{{ .SplashScreen.BackgroundImageURL }}');"{{ end }}>
  <div class="inv-splash-backdrop"></div>
  <div class="inv-splash-card">
    <div class="inv-splash-seal">
      <span>✉️</span>
    </div>
    <h2 class="inv-splash-title">
      {{ if .SplashScreen.Title }}{{ .SplashScreen.Title }}{{ else }}{{ .Title }}{{ end }}
    </h2>
    {{ if .SplashScreen.Message }}
      <p class="inv-splash-message">{{ .SplashScreen.Message }}</p>
    {{ end }}
    <button id="btn-splash-open" class="inv-splash-cta" type="button">
      <span>{{ .SplashScreen.CTAButtonText }}</span>
      <span class="inv-splash-arrow">&rarr;</span>
    </button>
  </div>
</div>
{{ end }}
```

### 2.2 Audio Element & Floating Controller Markup
Placed adjacent to the main invitation canvas:

```html
{{ if .Music }}
  <!-- Hidden Native Audio Engine -->
  <audio id="inv-bg-audio" preload="auto" {{ if .Music.Loop }}loop{{ end }}>
    <source src="{{ .Music.URL }}" type="audio/mpeg">
  </audio>

  <!-- Floating Audio Player FAB -->
  <button id="inv-music-toggle" class="inv-music-fab" type="button" 
          aria-label="Toggle background music" 
          data-autoplay="{{ .Music.Autoplay }}"
          title="{{ .Music.DisplayTitle "Background Music" }}">
    <div class="music-disc-icon">
      <!-- Vinyl record icon -->
      <svg class="icon-disc" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"></circle>
        <circle cx="12" cy="12" r="3"></circle>
      </svg>
      <!-- Sound wave bars -->
      <div class="music-bars" aria-hidden="true">
        <span></span><span></span><span></span>
      </div>
    </div>
    {{ if .Music.Title }}
      <span class="music-tooltip">{{ .Music.Title }}</span>
    {{ end }}
  </button>
{{ end }}
```

---

## 3. Styling & Animations (`web/static/css/invitation.css`)

### 3.1 Splash Screen Styling
```css
/* Splash Screen Fullscreen Overlay */
.inv-splash-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  background-color: var(--inv-bg, #F5F4EF);
  background-size: cover;
  background-position: center;
  transition: opacity 0.65s cubic-bezier(0.16, 1, 0.3, 1), transform 0.65s cubic-bezier(0.16, 1, 0.3, 1);
}

.inv-splash-overlay.is-dismissed {
  opacity: 0;
  transform: scale(1.04);
  pointer-events: none;
}

.inv-splash-card {
  position: relative;
  z-index: 2;
  background: var(--inv-surface, #FFFFFF);
  border: 1px solid var(--inv-border, #E5E1D8);
  border-radius: var(--inv-radius-lg, 20px);
  padding: 2.5rem 2rem;
  max-width: 420px;
  width: 100%;
  text-align: center;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.12);
}

.inv-splash-cta {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  background: var(--inv-accent, #264336);
  color: #FFFFFF;
  border: none;
  border-radius: var(--inv-radius-pill, 9999px);
  padding: 0.85rem 2rem;
  font-family: var(--inv-font-body);
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.15);
  transition: transform 0.2s, background-color 0.2s;
}

.inv-splash-cta:hover {
  transform: translateY(-2px);
  background: var(--inv-accent-hover, #1B3127);
}
```

### 3.2 Floating Music FAB Styling
```css
/* Floating Music Player FAB */
.inv-music-fab {
  position: fixed;
  bottom: calc(1.25rem + env(safe-area-inset-bottom, 0px));
  right: 1.25rem;
  z-index: 99;
  width: 46px;
  height: 46px;
  border-radius: 50%;
  border: 1px solid var(--inv-border, #E5E1D8);
  background: rgba(255, 255, 255, 0.88);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--inv-accent, #264336);
  transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1), background-color 0.2s;
}

.inv-music-fab:hover {
  transform: scale(1.08);
}

.inv-music-fab.is-playing .icon-disc {
  animation: spinVinyl 3.5s linear infinite;
}

@keyframes spinVinyl {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Sound wave bars */
.music-bars {
  position: absolute;
  display: flex;
  align-items: flex-end;
  gap: 2px;
  height: 12px;
  bottom: 8px;
  opacity: 0;
  transition: opacity 0.2s;
}

.inv-music-fab.is-playing .music-bars {
  opacity: 1;
}

.music-bars span {
  width: 2px;
  background-color: var(--inv-accent, #264336);
  border-radius: 1px;
  animation: eqWave 0.8s ease-in-out infinite alternate;
}

.music-bars span:nth-child(1) { height: 4px; animation-delay: 0.1s; }
.music-bars span:nth-child(2) { height: 10px; animation-delay: 0.3s; }
.music-bars span:nth-child(3) { height: 6px; animation-delay: 0.2s; }

@keyframes eqWave {
  0% { height: 3px; }
  100% { height: 12px; }
}

.inv-music-fab.is-paused {
  opacity: 0.75;
}
```

---

## 4. Interactive JavaScript Logic (`web/static/js/invitation.js`)

```javascript
/* ==========================================================================
   Splash Screen & Background Music Engine
   ========================================================================== */
function initMusicAndSplash() {
  const splash = document.getElementById('inv-splash-overlay');
  const splashBtn = document.getElementById('btn-splash-open');
  const audio = document.getElementById('inv-bg-audio');
  const fab = document.getElementById('inv-music-toggle');

  let isPlaying = false;

  function setPlayingState(playing) {
    isPlaying = playing;
    if (!fab) return;
    if (playing) {
      fab.classList.add('is-playing');
      fab.classList.remove('is-paused');
      fab.setAttribute('aria-label', 'Pause background music');
    } else {
      fab.classList.remove('is-playing');
      fab.classList.add('is-paused');
      fab.setAttribute('aria-label', 'Play background music');
    }
  }

  function fadeInAudio(durationMs = 1200) {
    if (!audio) return;
    audio.volume = 0;
    const playPromise = audio.play();
    if (!playPromise) return;

    playPromise
      .then(() => {
        setPlayingState(true);
        const targetVol = 0.75;
        const steps = 20;
        const stepTime = durationMs / steps;
        const stepVol = targetVol / steps;
        let currentStep = 0;

        const interval = setInterval(() => {
          currentStep++;
          if (audio.volume + stepVol >= targetVol || currentStep >= steps) {
            audio.volume = targetVol;
            clearInterval(interval);
          } else {
            audio.volume += stepVol;
          }
        }, stepTime);
      })
      .catch((err) => {
        console.warn('Playback deferred or blocked:', err.message);
        setPlayingState(false);
      });
  }

  // 1. Splash Screen Gesture Unlock
  if (splash && splashBtn) {
    splashBtn.addEventListener('click', () => {
      splash.classList.add('is-dismissed');
      // Unlock audio synchronously inside the click user gesture
      if (audio) {
        fadeInAudio(1000);
      }
      setTimeout(() => {
        splash.style.display = 'none';
      }, 700);
    });
  } else if (audio && fab && fab.getAttribute('data-autoplay') === 'true') {
    // 2. Fallback when splash is disabled: Play on very first user gesture anywhere
    const unlockOnGesture = () => {
      fadeInAudio(1000);
      window.removeEventListener('pointerdown', unlockOnGesture);
      window.removeEventListener('touchstart', unlockOnGesture);
      window.removeEventListener('keydown', unlockOnGesture);
    };
    window.addEventListener('pointerdown', unlockOnGesture, { once: true, passive: true });
    window.addEventListener('touchstart', unlockOnGesture, { once: true, passive: true });
    window.addEventListener('keydown', unlockOnGesture, { once: true, passive: true });
  }

  // 3. Floating FAB Play/Pause Toggle
  if (fab && audio) {
    fab.addEventListener('click', (e) => {
      e.stopPropagation();
      if (isPlaying) {
        audio.pause();
        setPlayingState(false);
      } else {
        audio.play().then(() => setPlayingState(true)).catch(() => {});
      }
    });

    // 4. Tab Visibility Pause/Resume
    document.addEventListener('visibilitychange', () => {
      if (document.hidden && isPlaying) {
        audio.pause();
        setPlayingState(false);
      }
    });
  }
}
```

---

## 5. Verification Checklist

- [ ] **Splash Screen Dismissal**:
  - Clicking "Open Invitation" smoothly animates the splash card offscreen.
  - Page underneath becomes interactive immediately.
- [ ] **Autoplay Verification**:
  - Test on iOS Mobile Safari and Android Chrome to confirm audio begins playing upon tapping "Open Invitation" without browser console warnings.
- [ ] **Floating FAB Usability**:
  - Clicking FAB pauses audio and stops the vinyl animation.
  - Clicking FAB again resumes audio and spins the vinyl icon.
  - Safe-area inset protects the button from iPhone home indicators.
  - Switching tabs or backgrounding the browser pauses playback.
