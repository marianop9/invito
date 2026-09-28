# Iteration 4: Admin Invitation Editor Integration (Bulma UI)

## 1. Goal & Rationale
Empower event planners to configure background music and the welcome splash screen directly inside the Admin Builder Studio ([`web/templates/admin/editor.html`](file:///home/nano/projects/invitation/web/templates/admin/editor.html) and [`web/static/js/admin_editor.js`](file:///home/nano/projects/invitation/web/static/js/admin_editor.js)).

### Design Constraint: Bulma CSS Native Components
Per architectural specifications, **Bulma 1.0 components** are preferred over custom CSS for all modifications in the Admin section. This keeps the admin interface consistent with existing cards, controls, and form elements.

---

## 2. Admin Template Markup (`web/templates/admin/editor.html`)

Insert a new dedicated section **"4. Welcome Splash & Background Music"** into the form pane:

```html
<!-- Card 4: Welcome Splash & Background Music -->
<section class="box mb-5">
  <h2 class="title is-5 mb-1">4. Welcome Splash & Background Music</h2>
  <p class="subtitle is-6 has-text-grey mb-4">
    Configure the interactive envelope opening screen and ambient background soundtrack.
  </p>

  <!-- Subsection A: Welcome Splash Screen -->
  <div class="field mb-4">
    <label class="checkbox has-text-weight-semibold">
      <input type="checkbox" x-model="ensureSplashScreen().enabled">
      Enable Tap-to-Open Welcome Splash Screen
    </label>
    <p class="help">Displays an elegant welcome card or envelope flap before the invitation opens.</p>
  </div>

  <div class="p-4 mb-5 has-background-light" style="border-radius: 8px;" x-show="invitation.splash_screen?.enabled" x-transition>
    <div class="columns is-multiline">
      <div class="column is-6">
        <div class="field">
          <label class="label is-small">Card Headline / Title</label>
          <div class="control">
            <input type="text" class="input is-small" placeholder="e.g. Lucas is turning 30" 
                   x-model="invitation.splash_screen.title">
          </div>
        </div>
      </div>

      <div class="column is-6">
        <div class="field">
          <label class="label is-small">CTA Button Text</label>
          <div class="control">
            <input type="text" class="input is-small" placeholder="Open Invitation" 
                   x-model="invitation.splash_screen.button_text">
          </div>
        </div>
      </div>

      <div class="column is-12">
        <div class="field">
          <label class="label is-small">Welcome Note / Message</label>
          <div class="control">
            <textarea class="textarea is-small" rows="2" 
                      placeholder="You are invited to celebrate chapter 30 with spritzes and sunset dancing..." 
                      x-model="invitation.splash_screen.message"></textarea>
          </div>
        </div>
      </div>
    </div>
  </div>

  <hr class="my-4">

  <!-- Subsection B: Background Music Soundtrack -->
  <div class="field mb-4">
    <label class="checkbox has-text-weight-semibold">
      <input type="checkbox" :checked="hasMusic()" @change="toggleMusic($event.target.checked)">
      Enable Ambient Background Music
    </label>
    <p class="help">Plays ambient music when the guest opens the invitation.</p>
  </div>

  <div class="p-4 has-background-light" style="border-radius: 8px;" x-show="hasMusic()" x-transition>
    <div class="columns is-multiline">
      <!-- Audio File Upload -->
      <div class="column is-12">
        <div class="field">
          <label class="label is-small">Audio Track File (MP3, M4A, OGG, WAV up to 10MB) <span class="has-text-danger">*</span></label>
          <div class="field has-addons">
            <div class="control is-expanded">
              <input type="text" class="input is-small" placeholder="/uploads/... or https://..." 
                     x-model="invitation.music.url">
            </div>
            <div class="control">
              <label class="button is-small is-primary is-light">
                <span>🎵 Upload Audio</span>
                <input type="file" accept="audio/mpeg,audio/mp3,audio/mp4,audio/ogg,audio/wav,audio/webm" 
                       style="display: none;" @change="uploadAudioFile($event)">
              </label>
            </div>
          </div>
        </div>
      </div>

      <!-- Track Name / Metadata -->
      <div class="column is-12">
        <div class="field">
          <label class="label is-small">Track Title or Artist (Shown on Floating Player)</label>
          <div class="control">
            <input type="text" class="input is-small" placeholder="e.g. Ludovico Einaudi - Nuvole Bianche" 
                   x-model="invitation.music.title">
          </div>
        </div>
      </div>

      <!-- Playback Options -->
      <div class="column is-6">
        <div class="field">
          <label class="checkbox is-size-7">
            <input type="checkbox" x-model="invitation.music.autoplay">
            Start playing on first tap / open (Autoplay)
          </label>
        </div>
      </div>

      <div class="column is-6">
        <div class="field">
          <label class="checkbox is-size-7">
            <input type="checkbox" x-model="invitation.music.loop">
            Loop playback continuously
          </label>
        </div>
      </div>

      <!-- Audio Preview In Admin -->
      <div class="column is-12" x-show="invitation.music?.url">
        <div class="field">
          <label class="label is-small">Track Preview</label>
          <audio class="mt-1" style="width: 100%; height: 36px;" controls :src="invitation.music.url" preload="none"></audio>
        </div>
      </div>
    </div>
  </div>
</section>
```

---

## 3. Alpine.js Controller Extensions (`web/static/js/admin_editor.js`)

Add state management and upload helpers to the `invitationEditor()` Alpine component:

```javascript
// State Helpers
ensureSplashScreen() {
  if (!this.invitation.splash_screen) {
    this.invitation.splash_screen = {
      enabled: false,
      title: '',
      message: '',
      button_text: 'Open Invitation'
    };
  }
  return this.invitation.splash_screen;
},

hasMusic() {
  return !!(this.invitation.music && this.invitation.music.url !== undefined);
},

toggleMusic(enable) {
  if (enable) {
    if (!this.invitation.music) {
      this.invitation.music = {
        url: '',
        title: '',
        autoplay: true,
        loop: true
      };
    }
  } else {
    this.invitation.music = null;
  }
},

async uploadAudioFile(event) {
  const file = event.target.files?.[0];
  if (!file) return;

  const fd = new FormData();
  fd.append('audio', file);
  if (this.invitation.slug) fd.append('slug', this.invitation.slug);

  this.syncStatus = 'Uploading audio track...';
  try {
    const res = await fetch('/api/upload', { method: 'POST', body: fd });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Audio upload failed');

    if (!this.invitation.music) {
      this.toggleMusic(true);
    }
    this.invitation.music.url = data.url;
    if (!this.invitation.music.title) {
      // Clean base filename without extension for friendly default track title
      const baseName = file.name.replace(/\.[^/.]+$/, '').replace(/[-_]/g, ' ');
      this.invitation.music.title = baseName;
    }

    this.syncStatus = 'Audio track uploaded!';
    event.target.value = '';
    this.refreshLivePreview();
  } catch (err) {
    alert('Audio upload failed: ' + err.message);
    this.syncStatus = 'Audio upload failed';
  }
}
```

---

## 4. Split-Screen Live Preview Synchronization
* When the planner adjusts splash text or track URLs, the editor triggers `refreshLivePreview()` via debounce.
* The preview iframe rerenders with the splash screen in place, letting the planner click "Open Invitation" inside the preview frame to test the exact guest journey.

---

## 5. Verification Checklist

- [ ] **Bulma Compliance**: Confirm all inputs, buttons, checkboxes, and layout columns use native Bulma classes (`.box`, `.field`, `.control`, `.input.is-small`, `.button.is-light`, `.has-background-light`).
- [ ] **Audio Upload**:
  - Click "🎵 Upload Audio" and choose an MP3/M4A file.
  - Verify loading indicator and that URL populates `/uploads/...`.
  - Verify track preview player works directly inside the admin view.
- [ ] **Splash Screen Configuration**:
  - Toggle splash screen on/off.
  - Modify title and CTA button text.
  - Save invitation and reload to verify persistence in SQLite.
