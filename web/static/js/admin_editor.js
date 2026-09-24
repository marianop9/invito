/**
 * Invito Admin Invitation Builder & Live Preview Studio
 * Clean Vanilla ES6 - Zero external runtime dependencies
 */

(function () {
  'use strict';

  // --- State Initialization ---
  const state = {
    invitation: null,
    isNew: true,
    manualSlugEdited: false,
    previewDebounceTimer: null,
    isSyncingPreview: false,
  };

  // DOM Elements
  const dom = {
    form: document.getElementById('invitation-form'),
    pageTitle: document.getElementById('page-editor-title'),
    statusBadge: document.getElementById('editor-status-badge'),
    previewSyncStatus: document.getElementById('preview-sync-status'),
    validationAlert: document.getElementById('validation-alert'),
    validationList: document.getElementById('validation-alert-list'),
    successToast: document.getElementById('success-toast'),
    successToastMsg: document.getElementById('success-toast-message'),
    btnSave: document.getElementById('btn-save-invitation'),
    saveSpinner: document.getElementById('save-spinner'),
    saveLabel: document.getElementById('save-label'),
    btnDelete: document.getElementById('btn-delete-invitation'),
    previewIframe: document.getElementById('preview-iframe'),
    previewWrapper: document.getElementById('preview-wrapper'),
    btnDeviceMobile: document.getElementById('btn-device-mobile'),
    btnDeviceDesktop: document.getElementById('btn-device-desktop'),
    btnRefreshPreview: document.getElementById('btn-refresh-preview'),
    btnOpenPreviewTab: document.getElementById('btn-open-preview-tab'),
    sectionsContainer: document.getElementById('sections-container'),
    selectNewSectionType: document.getElementById('select-new-section-type'),
    btnAddSection: document.getElementById('btn-add-section'),
    themeSwatches: document.querySelectorAll('.theme-swatch-card'),
    fieldThemeId: document.getElementById('field-theme-id'),

    // Event basics
    fieldTitle: document.getElementById('field-title'),
    fieldSlug: document.getElementById('field-slug'),
    slugHintText: document.getElementById('slug-hint-text'),
    fieldSubtitle: document.getElementById('field-subtitle'),
    fieldHosts: document.getElementById('field-hosts'),
    fieldDescription: document.getElementById('field-description'),
    fieldDateStart: document.getElementById('field-date-start'),
    fieldDateEnd: document.getElementById('field-date-end'),
    fieldTimezone: document.getElementById('field-timezone'),

    // Venue
    fieldLocName: document.getElementById('field-loc-name'),
    fieldLocAddress: document.getElementById('field-loc-address'),
    fieldLocMapUrl: document.getElementById('field-loc-map-url'),
    fieldLocDirections: document.getElementById('field-loc-directions'),

    // Palette overrides
    fieldPalPrimary: document.getElementById('field-palette-primary'),
    fieldPalBg: document.getElementById('field-palette-bg'),
    fieldPalAccent: document.getElementById('field-palette-accent'),
    fieldPalText: document.getElementById('field-palette-text'),
    fieldPalCardBg: document.getElementById('field-palette-card-bg'),
    fieldCustomCSS: document.getElementById('field-custom-css'),
  };

  // --- Helper Functions ---

  function slugify(text) {
    return text
      .toString()
      .toLowerCase()
      .trim()
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '') // remove accents
      .replace(/\s+/g, '-') // replace spaces with -
      .replace(/[^\w\-]+/g, '') // remove non-word chars
      .replace(/\-\-+/g, '-') // replace multiple - with single -
      .replace(/^-+/, '') // trim - from start
      .replace(/-+$/, ''); // trim - from end
  }

  function formatDateTimeLocal(isoStr) {
    if (!isoStr) return '';
    try {
      const d = new Date(isoStr);
      if (isNaN(d.getTime())) return '';
      // Return YYYY-MM-DDTHH:MM format
      const year = d.getFullYear();
      const month = String(d.getMonth() + 1).padStart(2, '0');
      const day = String(d.getDate()).padStart(2, '0');
      const hours = String(d.getHours()).padStart(2, '0');
      const minutes = String(d.getMinutes()).padStart(2, '0');
      return `${year}-${month}-${day}T${hours}:${minutes}`;
    } catch {
      return '';
    }
  }

  function parseDateTimeLocalToISO(valStr) {
    if (!valStr) return null;
    try {
      const d = new Date(valStr);
      if (isNaN(d.getTime())) return null;
      return d.toISOString();
    } catch {
      return null;
    }
  }

  function escapeHTML(str) {
    if (!str) return '';
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }

  // --- Section Factory & Default Definitions ---

  function createDefaultSection(type) {
    switch (type) {
      case 'hero':
        return {
          type: 'hero',
          layout: 'full-bleed',
          eyebrow: 'Celebration',
          vibe: '',
          cover_image_url: '',
          banner_image_url: '',
          show_countdown: true,
        };
      case 'details':
        return {
          type: 'details',
          show_map_link: true,
          show_calendar_button: true,
        };
      case 'quote':
        return {
          type: 'quote',
          text: 'Love begins in a moment, grows over time, and lasts forever.',
          author: '',
        };
      case 'text':
        return {
          type: 'text',
          title: 'Special Note',
          text: 'We are thrilled to celebrate with you.',
          align: 'center',
        };
      case 'image':
        return {
          type: 'image',
          url: '',
          caption: '',
          alt: '',
        };
      case 'carousel':
        return {
          type: 'carousel',
          title: 'Moments & Memories',
          aspect_ratio: '4-3',
          fit: 'cover',
          images: [],
        };
      case 'timeline':
        return {
          type: 'timeline',
          title: 'Schedule of Events',
          items: [
            { time: '4:00 PM', title: 'Arrival & Welcome', description: 'Main Lawn' },
            { time: '5:00 PM', title: 'Ceremony', description: 'Garden Pavilion' },
            { time: '7:00 PM', title: 'Dinner & Celebration', description: 'Grand Hall' },
          ],
        };
      case 'dress_code':
        return {
          type: 'dress_code',
          title: 'Dress Code',
          name: 'Cocktail Attire',
          description: 'Suits or cocktail dresses. Outdoor footwear recommended.',
          palette_hints: ['#2A4738', '#D4AF37', '#EFEBE4'],
        };
      case 'rsvp':
        return {
          type: 'rsvp',
          enabled: true,
          max_party_size: 2,
          ask_dietary: true,
          ask_song_request: false,
          custom_note: 'Kindly reply at your earliest convenience.',
        };
      case 'rsvp_external':
        return {
          type: 'rsvp_external',
          enabled: true,
          title: 'RSVP Confirmation',
          prompt: 'Please confirm your attendance via our external form',
          form_url: 'https://forms.google.com',
          button_label: 'Open RSVP Form &nearr;',
          custom_note: '',
        };
      case 'faqs':
        return {
          type: 'faqs',
          title: 'Frequently Asked Questions',
          items: [
            { question: 'Is parking available?', answer: 'Yes, valet parking is available on-site.' },
            { question: 'Are children welcome?', answer: 'We kindly request an adults-only celebration.' },
          ],
        };
      case 'gift_registry':
        return {
          type: 'gift_registry',
          message: 'Your presence is our present. If you would like to honor us with a gift:',
          links: [
            { label: 'Honeymoon Registry', url: 'https://www.zola.com' },
          ],
        };
      case 'closing':
        return {
          type: 'closing',
          message: 'We cannot wait to celebrate together!',
          signoff: 'With love & excitement,',
          hosts: '',
        };
      default:
        return { type: type };
    }
  }

  // --- Form Population & Serialization ---

  function populateForm(inv) {
    if (!inv) return;

    dom.fieldTitle.value = inv.title || '';
    dom.fieldSlug.value = inv.slug || '';
    dom.slugHintText.textContent = inv.slug || 'slug';
    dom.fieldSubtitle.value = inv.subtitle || '';
    dom.fieldHosts.value = (inv.hosts && Array.isArray(inv.hosts)) ? inv.hosts.join(', ') : '';
    dom.fieldDescription.value = inv.description || '';

    if (inv.date_start) {
      dom.fieldDateStart.value = formatDateTimeLocal(inv.date_start);
    }
    if (inv.date_end) {
      dom.fieldDateEnd.value = formatDateTimeLocal(inv.date_end);
    }
    dom.fieldTimezone.value = inv.timezone || 'America/New_York';

    if (inv.location) {
      dom.fieldLocName.value = inv.location.name || '';
      dom.fieldLocAddress.value = inv.location.address || '';
      dom.fieldLocMapUrl.value = inv.location.map_url || '';
      dom.fieldLocDirections.value = inv.location.directions_note || '';
    }

    const themeId = (inv.theme && inv.theme.id) ? inv.theme.id : 'botanical-elegance';
    dom.fieldThemeId.value = themeId;
    setActiveThemeSwatch(themeId);

    if (inv.theme && inv.theme.palette_override) {
      const pal = inv.theme.palette_override;
      dom.fieldPalPrimary.value = pal.primary || '';
      dom.fieldPalBg.value = pal.background || '';
      dom.fieldPalAccent.value = pal.accent || '';
      dom.fieldPalText.value = pal.text || '';
      dom.fieldPalCardBg.value = pal.card_bg || '';
    }
    if (inv.theme && inv.theme.custom_css) {
      dom.fieldCustomCSS.value = inv.theme.custom_css;
    }

    renderSections();
  }

  function serializeForm() {
    const title = dom.fieldTitle.value.trim();
    const slug = dom.fieldSlug.value.trim();
    const subtitle = dom.fieldSubtitle.value.trim();
    const hostsRaw = dom.fieldHosts.value.trim();
    const hosts = hostsRaw ? hostsRaw.split(',').map(s => s.trim()).filter(Boolean) : [];
    const description = dom.fieldDescription.value.trim();
    const dateStart = parseDateTimeLocalToISO(dom.fieldDateStart.value);
    const dateEnd = parseDateTimeLocalToISO(dom.fieldDateEnd.value);
    const timezone = dom.fieldTimezone.value.trim() || 'America/New_York';

    const location = {
      name: dom.fieldLocName.value.trim(),
      address: dom.fieldLocAddress.value.trim(),
      map_url: dom.fieldLocMapUrl.value.trim(),
      directions_note: dom.fieldLocDirections.value.trim(),
    };

    const theme = {
      id: dom.fieldThemeId.value || 'botanical-elegance',
    };

    const palPrimary = dom.fieldPalPrimary.value.trim();
    const palBg = dom.fieldPalBg.value.trim();
    const palAccent = dom.fieldPalAccent.value.trim();
    const palText = dom.fieldPalText.value.trim();
    const palCardBg = dom.fieldPalCardBg.value.trim();
    if (palPrimary || palBg || palAccent || palText || palCardBg) {
      theme.palette_override = {
        primary: palPrimary || undefined,
        background: palBg || undefined,
        accent: palAccent || undefined,
        text: palText || undefined,
        card_bg: palCardBg || undefined,
      };
    }
    const customCSS = dom.fieldCustomCSS.value.trim();
    if (customCSS) {
      theme.custom_css = customCSS;
    }

    // Build complete invitation document
    const doc = {
      version: state.invitation.version || '1.0',
      slug: slug,
      title: title,
      subtitle: subtitle || undefined,
      hosts: hosts.length > 0 ? hosts : undefined,
      description: description || undefined,
      date_start: dateStart || undefined,
      date_end: dateEnd || undefined,
      timezone: timezone,
      location: location,
      theme: theme,
      sections: state.invitation.sections || [],
    };

    return doc;
  }

  // --- Theme Swatch Handling ---

  function setActiveThemeSwatch(themeId) {
    dom.themeSwatches.forEach(card => {
      if (card.dataset.theme === themeId) {
        card.classList.add('active');
      } else {
        card.classList.remove('active');
      }
    });
  }

  dom.themeSwatches.forEach(card => {
    card.addEventListener('click', () => {
      const themeId = card.dataset.theme;
      dom.fieldThemeId.value = themeId;
      setActiveThemeSwatch(themeId);
      triggerPreviewSync();
    });
  });

  // --- Dynamic Sections Rendering ---

  function renderSections() {
    dom.sectionsContainer.innerHTML = '';
    const sections = state.invitation.sections || [];

    if (sections.length === 0) {
      dom.sectionsContainer.innerHTML = `
        <div style="text-align: center; padding: 2rem 1rem; color: var(--text-muted); background: var(--bg-subtle); border-radius: var(--radius-sm);">
          No sections configured yet. Use the selector below to add blocks.
        </div>
      `;
      return;
    }

    sections.forEach((sec, idx) => {
      const card = createSectionCardElement(sec, idx, sections.length);
      dom.sectionsContainer.appendChild(card);
    });
  }

  function createSectionCardElement(sec, idx, total) {
    const card = document.createElement('div');
    card.className = 'section-card';
    card.dataset.index = idx;

    const typeLabels = {
      hero: 'Hero Cover Card',
      details: 'Details Strip',
      quote: 'Quote / Excerpt',
      text: 'Text Announcement',
      image: 'Single Image',
      carousel: 'Photo Carousel',
      timeline: 'Schedule / Timeline',
      dress_code: 'Dress Code',
      rsvp: 'Interactive RSVP Form',
      rsvp_external: 'External RSVP Form Link',
      faqs: 'Guest FAQs',
      gift_registry: 'Gift Registry',
      closing: 'Closing Sign-off',
    };

    const typeLabel = typeLabels[sec.type] || sec.type;

    // Header HTML
    const header = document.createElement('div');
    header.className = 'section-card-header';
    header.innerHTML = `
      <div class="section-card-title-group">
        <span class="section-badge">${sec.type}</span>
        <span class="section-label">${typeLabel}</span>
      </div>
      <div class="section-card-actions">
        <button type="button" class="section-action-btn btn-move-up" title="Move Up" ${idx === 0 ? 'disabled style="opacity:0.4;cursor:default;"' : ''}>▲</button>
        <button type="button" class="section-action-btn btn-move-down" title="Move Down" ${idx === total - 1 ? 'disabled style="opacity:0.4;cursor:default;"' : ''}>▼</button>
        <button type="button" class="section-action-btn delete-btn btn-delete-section" title="Remove Section">🗑</button>
      </div>
    `;

    // Action button listeners
    const btnUp = header.querySelector('.btn-move-up');
    if (idx > 0) {
      btnUp.addEventListener('click', (e) => {
        e.stopPropagation();
        moveSection(idx, -1);
      });
    }

    const btnDown = header.querySelector('.btn-move-down');
    if (idx < total - 1) {
      btnDown.addEventListener('click', (e) => {
        e.stopPropagation();
        moveSection(idx, 1);
      });
    }

    const btnDel = header.querySelector('.btn-delete-section');
    btnDel.addEventListener('click', (e) => {
      e.stopPropagation();
      removeSection(idx);
    });

    // Body HTML based on section type
    const body = document.createElement('div');
    body.className = 'section-card-body';
    body.appendChild(buildSectionFields(sec, idx));

    card.appendChild(header);
    card.appendChild(body);
    return card;
  }

  function buildSectionFields(sec, idx) {
    const container = document.createElement('div');
    container.className = 'form-grid';

    switch (sec.type) {
      case 'hero': {
        container.innerHTML = `
          <div class="form-group col-6">
            <label class="form-label">Layout Style</label>
            <select class="form-input sec-field" data-key="layout">
              <option value="full-bleed" ${sec.layout === 'full-bleed' ? 'selected' : ''}>Full-Bleed Cover</option>
              <option value="banner" ${sec.layout === 'banner' ? 'selected' : ''}>Split Banner</option>
            </select>
          </div>
          <div class="form-group col-6">
            <label class="form-label">Eyebrow Tag</label>
            <input type="text" class="form-input sec-field" data-key="eyebrow" value="${escapeHTML(sec.eyebrow || '')}" placeholder="e.g. Wedding Celebration">
          </div>
          <div class="form-group col-12">
            <label class="form-label">Vibe / Tagline</label>
            <input type="text" class="form-input sec-field" data-key="vibe" value="${escapeHTML(sec.vibe || '')}" placeholder="e.g. An intimate evening of vows and dinner">
          </div>
          <div class="form-group col-12">
            <label class="form-label">Hero Cover Image URL</label>
            <div class="image-upload-wrapper">
              <div class="image-upload-row">
                <input type="text" class="form-input sec-field img-url-input" data-key="cover_image_url" value="${escapeHTML(sec.cover_image_url || '')}" placeholder="/uploads/... or https://...">
                <label class="upload-btn-label">
                  <span>📷 Upload</span>
                  <input type="file" accept="image/png,image/jpeg,image/webp,image/gif" class="uploader-file-input" style="display:none;">
                </label>
              </div>
              ${sec.cover_image_url ? `<img src="${escapeHTML(sec.cover_image_url)}" class="image-upload-thumb">` : ''}
            </div>
          </div>
          <div class="form-group col-12" style="flex-direction:row; align-items:center; gap:0.5rem; margin-top:0.25rem;">
            <input type="checkbox" id="hero-countdown-${idx}" class="sec-field" data-key="show_countdown" ${sec.show_countdown ? 'checked' : ''}>
            <label for="hero-countdown-${idx}" class="form-label" style="margin-bottom:0;cursor:pointer;">Show Countdown Clock</label>
          </div>
        `;
        break;
      }

      case 'details': {
        container.innerHTML = `
          <div class="form-group col-6" style="flex-direction:row; align-items:center; gap:0.5rem;">
            <input type="checkbox" id="details-map-${idx}" class="sec-field" data-key="show_map_link" ${sec.show_map_link ? 'checked' : ''}>
            <label for="details-map-${idx}" class="form-label" style="margin-bottom:0;cursor:pointer;">Show Google Maps Link</label>
          </div>
          <div class="form-group col-6" style="flex-direction:row; align-items:center; gap:0.5rem;">
            <input type="checkbox" id="details-cal-${idx}" class="sec-field" data-key="show_calendar_button" ${sec.show_calendar_button ? 'checked' : ''}>
            <label for="details-cal-${idx}" class="form-label" style="margin-bottom:0;cursor:pointer;">Show Add to Calendar Button</label>
          </div>
        `;
        break;
      }

      case 'quote': {
        container.innerHTML = `
          <div class="form-group col-12">
            <label class="form-label required">Quote Text</label>
            <textarea class="form-textarea sec-field" data-key="text" rows="2" placeholder="e.g. Whatever our souls are made of, his and mine are the same.">${escapeHTML(sec.text || '')}</textarea>
          </div>
          <div class="form-group col-12">
            <label class="form-label">Author / Attribution</label>
            <input type="text" class="form-input sec-field" data-key="author" value="${escapeHTML(sec.author || '')}" placeholder="e.g. Emily Brontë">
          </div>
        `;
        break;
      }

      case 'text': {
        container.innerHTML = `
          <div class="form-group col-8">
            <label class="form-label">Announcement Title</label>
            <input type="text" class="form-input sec-field" data-key="title" value="${escapeHTML(sec.title || '')}" placeholder="e.g. Guest Accommodations">
          </div>
          <div class="form-group col-4">
            <label class="form-label">Text Alignment</label>
            <select class="form-input sec-field" data-key="align">
              <option value="left" ${sec.align === 'left' ? 'selected' : ''}>Left</option>
              <option value="center" ${(!sec.align || sec.align === 'center') ? 'selected' : ''}>Center</option>
              <option value="right" ${sec.align === 'right' ? 'selected' : ''}>Right</option>
            </select>
          </div>
          <div class="form-group col-12">
            <label class="form-label required">Message Content</label>
            <textarea class="form-textarea sec-field" data-key="text" rows="3" placeholder="Inform guests about parking, lodging, or transport...">${escapeHTML(sec.text || '')}</textarea>
          </div>
        `;
        break;
      }

      case 'image': {
        container.innerHTML = `
          <div class="form-group col-12">
            <label class="form-label required">Photo Image URL</label>
            <div class="image-upload-wrapper">
              <div class="image-upload-row">
                <input type="text" class="form-input sec-field img-url-input" data-key="url" value="${escapeHTML(sec.url || '')}" placeholder="/uploads/... or https://...">
                <label class="upload-btn-label">
                  <span>📷 Upload</span>
                  <input type="file" accept="image/png,image/jpeg,image/webp,image/gif" class="uploader-file-input" style="display:none;">
                </label>
              </div>
              ${sec.url ? `<img src="${escapeHTML(sec.url)}" class="image-upload-thumb">` : ''}
            </div>
          </div>
          <div class="form-group col-6">
            <label class="form-label">Caption</label>
            <input type="text" class="form-input sec-field" data-key="caption" value="${escapeHTML(sec.caption || '')}" placeholder="Optional photo caption">
          </div>
          <div class="form-group col-6">
            <label class="form-label">Alt Text</label>
            <input type="text" class="form-input sec-field" data-key="alt" value="${escapeHTML(sec.alt || '')}" placeholder="Accessibility description">
          </div>
        `;
        break;
      }

      case 'carousel': {
        container.innerHTML = `
          <div class="form-group col-6">
            <label class="form-label">Gallery Title</label>
            <input type="text" class="form-input sec-field" data-key="title" value="${escapeHTML(sec.title || '')}" placeholder="e.g. Our Story">
          </div>
          <div class="form-group col-3">
            <label class="form-label">Aspect Ratio</label>
            <select class="form-input sec-field" data-key="aspect_ratio">
              <option value="4-3" ${sec.aspect_ratio === '4-3' ? 'selected' : ''}>4:3 Standard</option>
              <option value="16-9" ${sec.aspect_ratio === '16-9' ? 'selected' : ''}>16:9 Landscape</option>
              <option value="1-1" ${sec.aspect_ratio === '1-1' ? 'selected' : ''}>1:1 Square</option>
              <option value="4-5" ${sec.aspect_ratio === '4-5' ? 'selected' : ''}>4:5 Portrait</option>
            </select>
          </div>
          <div class="form-group col-3">
            <label class="form-label">Fit Mode</label>
            <select class="form-input sec-field" data-key="fit">
              <option value="cover" ${sec.fit === 'cover' ? 'selected' : ''}>Cover (Crop)</option>
              <option value="contain" ${sec.fit === 'contain' ? 'selected' : ''}>Contain</option>
            </select>
          </div>
          <div class="form-group col-12">
            <label class="form-label">Gallery Photos</label>
            <div class="sub-items-stack" id="carousel-items-${idx}"></div>
            <button type="button" class="btn btn-secondary btn-add-sub-item" style="margin-top: 0.5rem; font-size: 0.8125rem;">+ Add Photo to Gallery</button>
          </div>
        `;
        renderCarouselItems(container.querySelector(`#carousel-items-${idx}`), sec, idx);
        container.querySelector('.btn-add-sub-item').addEventListener('click', () => {
          if (!sec.images) sec.images = [];
          sec.images.push({ url: '', caption: '', alt: '' });
          renderCarouselItems(container.querySelector(`#carousel-items-${idx}`), sec, idx);
          triggerPreviewSync();
        });
        break;
      }

      case 'timeline': {
        container.innerHTML = `
          <div class="form-group col-12">
            <label class="form-label">Schedule Section Title</label>
            <input type="text" class="form-input sec-field" data-key="title" value="${escapeHTML(sec.title || '')}" placeholder="Schedule of Events">
          </div>
          <div class="form-group col-12">
            <label class="form-label">Milestones</label>
            <div class="sub-items-stack" id="timeline-items-${idx}"></div>
            <button type="button" class="btn btn-secondary btn-add-sub-item" style="margin-top: 0.5rem; font-size: 0.8125rem;">+ Add Milestone</button>
          </div>
        `;
        renderTimelineItems(container.querySelector(`#timeline-items-${idx}`), sec, idx);
        container.querySelector('.btn-add-sub-item').addEventListener('click', () => {
          if (!sec.items) sec.items = [];
          sec.items.push({ time: '', title: '', description: '', icon: '' });
          renderTimelineItems(container.querySelector(`#timeline-items-${idx}`), sec, idx);
          triggerPreviewSync();
        });
        break;
      }

      case 'dress_code': {
        const hints = (sec.palette_hints && Array.isArray(sec.palette_hints)) ? sec.palette_hints.join(', ') : '';
        container.innerHTML = `
          <div class="form-group col-6">
            <label class="form-label">Section Title</label>
            <input type="text" class="form-input sec-field" data-key="title" value="${escapeHTML(sec.title || '')}" placeholder="Dress Code">
          </div>
          <div class="form-group col-6">
            <label class="form-label required">Attire Name</label>
            <input type="text" class="form-input sec-field" data-key="name" value="${escapeHTML(sec.name || '')}" placeholder="e.g. Garden Formal">
          </div>
          <div class="form-group col-12">
            <label class="form-label">Guideline Description</label>
            <textarea class="form-textarea sec-field" data-key="description" rows="2" placeholder="Suits or cocktail dresses. Block heels recommended for the lawn...">${escapeHTML(sec.description || '')}</textarea>
          </div>
          <div class="form-group col-12">
            <label class="form-label">Suggested Palette Hints (Hex codes, comma-separated)</label>
            <input type="text" class="form-input sec-palette-hints" value="${escapeHTML(hints)}" placeholder="#2A4738, #D4AF37, #EFEBE4">
          </div>
        `;
        const hintsInput = container.querySelector('.sec-palette-hints');
        hintsInput.addEventListener('input', () => {
          const raw = hintsInput.value.trim();
          sec.palette_hints = raw ? raw.split(',').map(s => s.trim()).filter(Boolean) : [];
          triggerPreviewSync();
        });
        break;
      }

      case 'rsvp': {
        const deadlineVal = sec.deadline ? formatDateTimeLocal(sec.deadline) : '';
        container.innerHTML = `
          <div class="form-group col-6" style="flex-direction:row; align-items:center; gap:0.5rem;">
            <input type="checkbox" id="rsvp-enable-${idx}" class="sec-field" data-key="enabled" ${sec.enabled ? 'checked' : ''}>
            <label for="rsvp-enable-${idx}" class="form-label" style="margin-bottom:0;cursor:pointer;">Accept Interactive RSVPs</label>
          </div>
          <div class="form-group col-6">
            <label class="form-label">RSVP Deadline</label>
            <input type="datetime-local" class="form-input sec-deadline-input" value="${deadlineVal}">
          </div>
          <div class="form-group col-4">
            <label class="form-label">Max Party Size</label>
            <input type="number" class="form-input sec-field" data-key="max_party_size" value="${sec.max_party_size || 2}" min="1" max="10">
          </div>
          <div class="form-group col-4" style="flex-direction:row; align-items:center; gap:0.5rem; margin-top:1.5rem;">
            <input type="checkbox" id="rsvp-dietary-${idx}" class="sec-field" data-key="ask_dietary" ${sec.ask_dietary ? 'checked' : ''}>
            <label for="rsvp-dietary-${idx}" class="form-label" style="margin-bottom:0;cursor:pointer;">Ask Dietary Needs</label>
          </div>
          <div class="form-group col-4" style="flex-direction:row; align-items:center; gap:0.5rem; margin-top:1.5rem;">
            <input type="checkbox" id="rsvp-song-${idx}" class="sec-field" data-key="ask_song_request" ${sec.ask_song_request ? 'checked' : ''}>
            <label for="rsvp-song-${idx}" class="form-label" style="margin-bottom:0;cursor:pointer;">Ask Song Request</label>
          </div>
          <div class="form-group col-12">
            <label class="form-label">Custom Note / RSVP Prompt</label>
            <input type="text" class="form-input sec-field" data-key="custom_note" value="${escapeHTML(sec.custom_note || '')}" placeholder="Kindly RSVP by August 15th">
          </div>
        `;
        const deadInput = container.querySelector('.sec-deadline-input');
        deadInput.addEventListener('change', () => {
          sec.deadline = parseDateTimeLocalToISO(deadInput.value);
          triggerPreviewSync();
        });
        break;
      }

      case 'rsvp_external': {
        const deadlineVal = sec.deadline ? formatDateTimeLocal(sec.deadline) : '';
        container.innerHTML = `
          <div class="form-group col-6">
            <label class="form-label">Title</label>
            <input type="text" class="form-input sec-field" data-key="title" value="${escapeHTML(sec.title || '')}" placeholder="RSVP Confirmation">
          </div>
          <div class="form-group col-6">
            <label class="form-label required">External Form URL</label>
            <input type="url" class="form-input sec-field" data-key="form_url" value="${escapeHTML(sec.form_url || '')}" placeholder="https://forms.google.com/...">
          </div>
          <div class="form-group col-6">
            <label class="form-label">Button Label</label>
            <input type="text" class="form-input sec-field" data-key="button_label" value="${escapeHTML(sec.button_label || '')}" placeholder="Open RSVP Form &nearr;">
          </div>
          <div class="form-group col-6">
            <label class="form-label">Deadline</label>
            <input type="datetime-local" class="form-input sec-ext-deadline" value="${deadlineVal}">
          </div>
          <div class="form-group col-12">
            <label class="form-label">Prompt Message</label>
            <input type="text" class="form-input sec-field" data-key="prompt" value="${escapeHTML(sec.prompt || '')}" placeholder="Please click below to register your attendance">
          </div>
        `;
        const deadInput = container.querySelector('.sec-ext-deadline');
        deadInput.addEventListener('change', () => {
          sec.deadline = parseDateTimeLocalToISO(deadInput.value);
          triggerPreviewSync();
        });
        break;
      }

      case 'faqs': {
        container.innerHTML = `
          <div class="form-group col-12">
            <label class="form-label">FAQs Title</label>
            <input type="text" class="form-input sec-field" data-key="title" value="${escapeHTML(sec.title || '')}" placeholder="Information for Guests">
          </div>
          <div class="form-group col-12">
            <label class="form-label">Questions & Answers</label>
            <div class="sub-items-stack" id="faq-items-${idx}"></div>
            <button type="button" class="btn btn-secondary btn-add-sub-item" style="margin-top: 0.5rem; font-size: 0.8125rem;">+ Add Question</button>
          </div>
        `;
        renderFAQItems(container.querySelector(`#faq-items-${idx}`), sec, idx);
        container.querySelector('.btn-add-sub-item').addEventListener('click', () => {
          if (!sec.items) sec.items = [];
          sec.items.push({ question: '', answer: '' });
          renderFAQItems(container.querySelector(`#faq-items-${idx}`), sec, idx);
          triggerPreviewSync();
        });
        break;
      }

      case 'gift_registry': {
        container.innerHTML = `
          <div class="form-group col-12">
            <label class="form-label">Registry Message</label>
            <textarea class="form-textarea sec-field" data-key="message" rows="2" placeholder="Your presence is our present...">${escapeHTML(sec.message || '')}</textarea>
          </div>
          <div class="form-group col-12">
            <label class="form-label">Registry Links</label>
            <div class="sub-items-stack" id="registry-items-${idx}"></div>
            <button type="button" class="btn btn-secondary btn-add-sub-item" style="margin-top: 0.5rem; font-size: 0.8125rem;">+ Add Registry Link</button>
          </div>
        `;
        renderRegistryItems(container.querySelector(`#registry-items-${idx}`), sec, idx);
        container.querySelector('.btn-add-sub-item').addEventListener('click', () => {
          if (!sec.links) sec.links = [];
          sec.links.push({ label: '', url: '' });
          renderRegistryItems(container.querySelector(`#registry-items-${idx}`), sec, idx);
          triggerPreviewSync();
        });
        break;
      }

      case 'closing': {
        container.innerHTML = `
          <div class="form-group col-12">
            <label class="form-label">Heartfelt Message</label>
            <textarea class="form-textarea sec-field" data-key="message" rows="2" placeholder="We cannot wait to celebrate with you!">${escapeHTML(sec.message || '')}</textarea>
          </div>
          <div class="form-group col-6">
            <label class="form-label">Sign-off</label>
            <input type="text" class="form-input sec-field" data-key="signoff" value="${escapeHTML(sec.signoff || '')}" placeholder="With love & gratitude,">
          </div>
          <div class="form-group col-6">
            <label class="form-label">Display Hosts</label>
            <input type="text" class="form-input sec-field" data-key="hosts" value="${escapeHTML(sec.hosts || '')}" placeholder="Sarah & Alexander">
          </div>
        `;
        break;
      }
    }

    // Attach generic change listeners for primitive inputs
    container.querySelectorAll('.sec-field').forEach(input => {
      const key = input.dataset.key;
      if (!key) return;

      const updateVal = () => {
        if (input.type === 'checkbox') {
          sec[key] = input.checked;
        } else if (input.type === 'number') {
          sec[key] = parseInt(input.value, 10) || 0;
        } else {
          sec[key] = input.value;
        }
        triggerPreviewSync();
      };

      input.addEventListener('input', updateVal);
      input.addEventListener('change', updateVal);
    });

    // Wire up image uploaders inside this section
    container.querySelectorAll('.uploader-file-input').forEach(fileInput => {
      fileInput.addEventListener('change', () => {
        if (!fileInput.files || fileInput.files.length === 0) return;
        const file = fileInput.files[0];
        const wrapper = fileInput.closest('.image-upload-wrapper');
        const urlInput = wrapper.querySelector('.img-url-input');

        uploadImage(file, (url) => {
          urlInput.value = url;
          const key = urlInput.dataset.key;
          if (key) sec[key] = url;

          // Update or add thumbnail
          let thumb = wrapper.querySelector('.image-upload-thumb');
          if (!thumb) {
            thumb = document.createElement('img');
            thumb.className = 'image-upload-thumb';
            wrapper.appendChild(thumb);
          }
          thumb.src = url;
          triggerPreviewSync();
        });
      });
    });

    return container;
  }

  // --- Sub-items renderers (Carousel, Timeline, FAQs, Registry) ---

  function renderCarouselItems(container, sec, secIdx) {
    container.innerHTML = '';
    const images = sec.images || [];

    images.forEach((img, i) => {
      const row = document.createElement('div');
      row.className = 'sub-item-card form-grid';
      row.innerHTML = `
        <button type="button" class="sub-item-remove-btn" title="Remove Photo">&times;</button>
        <div class="form-group col-12">
          <label class="form-label required">Photo #${i + 1} URL</label>
          <div class="image-upload-wrapper">
            <div class="image-upload-row">
              <input type="text" class="form-input item-field item-url-input" value="${escapeHTML(img.url || '')}" placeholder="/uploads/... or https://...">
              <label class="upload-btn-label">
                <span>📷 Upload</span>
                <input type="file" accept="image/png,image/jpeg,image/webp,image/gif" class="uploader-file-input" style="display:none;">
              </label>
            </div>
            ${img.url ? `<img src="${escapeHTML(img.url)}" class="image-upload-thumb">` : ''}
          </div>
        </div>
        <div class="form-group col-6">
          <label class="form-label">Caption</label>
          <input type="text" class="form-input item-field item-caption" value="${escapeHTML(img.caption || '')}" placeholder="Caption">
        </div>
        <div class="form-group col-6">
          <label class="form-label">Alt Text</label>
          <input type="text" class="form-input item-field item-alt" value="${escapeHTML(img.alt || '')}" placeholder="Alt text">
        </div>
      `;

      row.querySelector('.sub-item-remove-btn').addEventListener('click', () => {
        sec.images.splice(i, 1);
        renderCarouselItems(container, sec, secIdx);
        triggerPreviewSync();
      });

      const urlInput = row.querySelector('.item-url-input');
      const captionInput = row.querySelector('.item-caption');
      const altInput = row.querySelector('.item-alt');

      const updateImg = () => {
        img.url = urlInput.value.trim();
        img.caption = captionInput.value.trim();
        img.alt = altInput.value.trim();
        triggerPreviewSync();
      };
      urlInput.addEventListener('input', updateImg);
      captionInput.addEventListener('input', updateImg);
      altInput.addEventListener('input', updateImg);

      const fileInput = row.querySelector('.uploader-file-input');
      fileInput.addEventListener('change', () => {
        if (!fileInput.files || fileInput.files.length === 0) return;
        uploadImage(fileInput.files[0], (url) => {
          urlInput.value = url;
          img.url = url;
          const wrapper = row.querySelector('.image-upload-wrapper');
          let thumb = wrapper.querySelector('.image-upload-thumb');
          if (!thumb) {
            thumb = document.createElement('img');
            thumb.className = 'image-upload-thumb';
            wrapper.appendChild(thumb);
          }
          thumb.src = url;
          triggerPreviewSync();
        });
      });

      container.appendChild(row);
    });
  }

  function renderTimelineItems(container, sec, secIdx) {
    container.innerHTML = '';
    const items = sec.items || [];

    items.forEach((item, i) => {
      const row = document.createElement('div');
      row.className = 'sub-item-card form-grid';
      row.innerHTML = `
        <button type="button" class="sub-item-remove-btn" title="Remove Milestone">&times;</button>
        <div class="form-group col-3">
          <label class="form-label required">Time</label>
          <input type="text" class="form-input item-time" value="${escapeHTML(item.time || '')}" placeholder="e.g. 4:00 PM">
        </div>
        <div class="form-group col-5">
          <label class="form-label required">Milestone Title</label>
          <input type="text" class="form-input item-title" value="${escapeHTML(item.title || '')}" placeholder="e.g. Ceremony">
        </div>
        <div class="form-group col-4">
          <label class="form-label">Location / Note</label>
          <input type="text" class="form-input item-desc" value="${escapeHTML(item.description || '')}" placeholder="e.g. Conservatory Lawn">
        </div>
      `;

      row.querySelector('.sub-item-remove-btn').addEventListener('click', () => {
        sec.items.splice(i, 1);
        renderTimelineItems(container, sec, secIdx);
        triggerPreviewSync();
      });

      const tInput = row.querySelector('.item-time');
      const titleInput = row.querySelector('.item-title');
      const descInput = row.querySelector('.item-desc');

      const updateItem = () => {
        item.time = tInput.value.trim();
        item.title = titleInput.value.trim();
        item.description = descInput.value.trim();
        triggerPreviewSync();
      };
      tInput.addEventListener('input', updateItem);
      titleInput.addEventListener('input', updateItem);
      descInput.addEventListener('input', updateItem);

      container.appendChild(row);
    });
  }

  function renderFAQItems(container, sec, secIdx) {
    container.innerHTML = '';
    const items = sec.items || [];

    items.forEach((item, i) => {
      const row = document.createElement('div');
      row.className = 'sub-item-card form-grid';
      row.innerHTML = `
        <button type="button" class="sub-item-remove-btn" title="Remove FAQ">&times;</button>
        <div class="form-group col-12">
          <label class="form-label required">Question</label>
          <input type="text" class="form-input item-q" value="${escapeHTML(item.question || '')}" placeholder="e.g. Are children invited?">
        </div>
        <div class="form-group col-12">
          <label class="form-label required">Answer</label>
          <textarea class="form-textarea item-a" rows="2" placeholder="e.g. We kindly request an adults-only celebration.">${escapeHTML(item.answer || '')}</textarea>
        </div>
      `;

      row.querySelector('.sub-item-remove-btn').addEventListener('click', () => {
        sec.items.splice(i, 1);
        renderFAQItems(container, sec, secIdx);
        triggerPreviewSync();
      });

      const qInput = row.querySelector('.item-q');
      const aInput = row.querySelector('.item-a');

      const updateFAQ = () => {
        item.question = qInput.value.trim();
        item.answer = aInput.value.trim();
        triggerPreviewSync();
      };
      qInput.addEventListener('input', updateFAQ);
      aInput.addEventListener('input', updateFAQ);

      container.appendChild(row);
    });
  }

  function renderRegistryItems(container, sec, secIdx) {
    container.innerHTML = '';
    const links = sec.links || [];

    links.forEach((link, i) => {
      const row = document.createElement('div');
      row.className = 'sub-item-card form-grid';
      row.innerHTML = `
        <button type="button" class="sub-item-remove-btn" title="Remove Link">&times;</button>
        <div class="form-group col-5">
          <label class="form-label required">Store / Fund Label</label>
          <input type="text" class="form-input item-label" value="${escapeHTML(link.label || '')}" placeholder="e.g. Crate & Barrel">
        </div>
        <div class="form-group col-7">
          <label class="form-label required">Registry Link URL</label>
          <input type="url" class="form-input item-url" value="${escapeHTML(link.url || '')}" placeholder="https://...">
        </div>
      `;

      row.querySelector('.sub-item-remove-btn').addEventListener('click', () => {
        sec.links.splice(i, 1);
        renderRegistryItems(container, sec, secIdx);
        triggerPreviewSync();
      });

      const lInput = row.querySelector('.item-label');
      const uInput = row.querySelector('.item-url');

      const updateLink = () => {
        link.label = lInput.value.trim();
        link.url = uInput.value.trim();
        triggerPreviewSync();
      };
      lInput.addEventListener('input', updateLink);
      uInput.addEventListener('input', updateLink);

      container.appendChild(row);
    });
  }

  // --- Section Add / Remove / Reorder ---

  function addSection(type) {
    if (!state.invitation) {
      state.invitation = { sections: [] };
    }
    if (!Array.isArray(state.invitation.sections)) {
      state.invitation.sections = [];
    }
    const newSec = createDefaultSection(type);
    state.invitation.sections.push(newSec);
    renderSections();
    triggerPreviewSync();
  }

  function removeSection(index) {
    if (!state.invitation.sections || index < 0 || index >= state.invitation.sections.length) return;
    state.invitation.sections.splice(index, 1);
    renderSections();
    triggerPreviewSync();
  }

  function moveSection(index, direction) {
    const sections = state.invitation.sections;
    if (!sections) return;
    const targetIdx = index + direction;
    if (targetIdx < 0 || targetIdx >= sections.length) return;

    const temp = sections[index];
    sections[index] = sections[targetIdx];
    sections[targetIdx] = temp;

    renderSections();
    triggerPreviewSync();
  }

  dom.btnAddSection.addEventListener('click', () => {
    const selectedType = dom.selectNewSectionType.value;
    if (selectedType) {
      addSection(selectedType);
    }
  });

  // --- Image Upload Pipeline ---

  function uploadImage(file, onSuccess) {
    if (!file) return;

    const formData = new FormData();
    formData.append('image', file);
    if (state.invitation && state.invitation.slug) {
      formData.append('slug', state.invitation.slug);
    }

    dom.previewSyncStatus.textContent = 'Uploading image...';

    fetch('/api/upload', {
      method: 'POST',
      body: formData,
    })
      .then(res => {
        if (!res.ok) {
          return res.json().then(errData => {
            throw new Error(errData.error || `Upload failed with status ${res.status}`);
          });
        }
        return res.json();
      })
      .then(info => {
        dom.previewSyncStatus.textContent = 'Image uploaded!';
        onSuccess(info.url);
      })
      .catch(err => {
        dom.previewSyncStatus.textContent = 'Image upload error';
        alert(`Failed to upload photo: ${err.message}`);
      });
  }

  // --- Live Preview Synchronization ---

  function triggerPreviewSync() {
    clearTimeout(state.previewDebounceTimer);
    dom.previewSyncStatus.textContent = 'Syncing preview...';
    state.previewDebounceTimer = setTimeout(updateLivePreview, 350);
  }

  function updateLivePreview() {
    const payload = serializeForm();

    fetch('/admin/invitations/preview', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
      .then(res => {
        if (!res.ok) {
          throw new Error('Preview rendering failed');
        }
        return res.text();
      })
      .then(html => {
        dom.previewIframe.srcdoc = html;
        dom.previewSyncStatus.textContent = 'Live preview synced';
      })
      .catch(err => {
        dom.previewSyncStatus.textContent = 'Preview sync paused';
      });
  }

  // Viewport device switcher
  dom.btnDeviceMobile.addEventListener('click', () => {
    dom.btnDeviceMobile.classList.add('active');
    dom.btnDeviceDesktop.classList.remove('active');
    dom.previewWrapper.className = 'preview-device-wrapper device-mobile';
  });

  dom.btnDeviceDesktop.addEventListener('click', () => {
    dom.btnDeviceDesktop.classList.add('active');
    dom.btnDeviceMobile.classList.remove('active');
    dom.previewWrapper.className = 'preview-device-wrapper device-desktop';
  });

  dom.btnRefreshPreview.addEventListener('click', () => {
    updateLivePreview();
  });

  dom.btnOpenPreviewTab.addEventListener('click', () => {
    const slug = dom.fieldSlug.value.trim();
    if (!state.isNew && slug) {
      window.open(`/i/${slug}`, '_blank');
    } else {
      updateLivePreview();
    }
  });

  // --- Form Input Listeners ---

  // Slug auto-generation on new draft
  dom.fieldTitle.addEventListener('input', () => {
    if (state.isNew && !state.manualSlugEdited) {
      const autoSlug = slugify(dom.fieldTitle.value);
      dom.fieldSlug.value = autoSlug;
      dom.slugHintText.textContent = autoSlug || 'slug';
    }
    triggerPreviewSync();
  });

  dom.fieldSlug.addEventListener('input', () => {
    if (state.isNew) {
      state.manualSlugEdited = true;
      dom.slugHintText.textContent = dom.fieldSlug.value || 'slug';
    }
    triggerPreviewSync();
  });

  // Form inputs change trigger preview
  [
    dom.fieldSubtitle,
    dom.fieldHosts,
    dom.fieldDescription,
    dom.fieldDateStart,
    dom.fieldDateEnd,
    dom.fieldTimezone,
    dom.fieldLocName,
    dom.fieldLocAddress,
    dom.fieldLocMapUrl,
    dom.fieldLocDirections,
    dom.fieldPalPrimary,
    dom.fieldPalBg,
    dom.fieldPalAccent,
    dom.fieldPalText,
    dom.fieldPalCardBg,
    dom.fieldCustomCSS,
  ].forEach(el => {
    if (el) {
      el.addEventListener('input', triggerPreviewSync);
      el.addEventListener('change', triggerPreviewSync);
    }
  });

  // --- Save / Submit Handler ---

  function showErrors(errList) {
    dom.validationList.innerHTML = '';
    errList.forEach(msg => {
      const li = document.createElement('li');
      li.textContent = msg;
      dom.validationList.appendChild(li);
    });
    dom.validationAlert.style.display = 'flex';
    dom.validationAlert.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }

  function hideErrors() {
    dom.validationAlert.style.display = 'none';
    dom.validationList.innerHTML = '';
  }

  function showSuccess(msg) {
    dom.successToastMsg.textContent = msg;
    dom.successToast.style.display = 'flex';
    setTimeout(() => {
      dom.successToast.style.display = 'none';
    }, 4500);
  }

  dom.btnSave.addEventListener('click', () => {
    hideErrors();

    // Client-side quick check
    const clientErrors = [];
    if (!dom.fieldTitle.value.trim()) clientErrors.push('Event Title is required.');
    if (!dom.fieldSlug.value.trim()) clientErrors.push('URL Slug is required.');
    if (!dom.fieldDateStart.value) clientErrors.push('Start Date & Time is required.');
    if (!dom.fieldLocName.value.trim()) clientErrors.push('Venue Name is required.');
    if (!dom.fieldLocAddress.value.trim()) clientErrors.push('Physical Street Address is required.');

    if (clientErrors.length > 0) {
      showErrors(clientErrors);
      return;
    }

    const payload = serializeForm();

    // UI Loading state
    dom.btnSave.disabled = true;
    dom.saveSpinner.style.display = 'inline';
    dom.saveLabel.textContent = 'Saving...';

    const url = state.isNew ? '/api/invitations' : `/api/invitations/${encodeURIComponent(state.invitation.slug)}`;
    const method = state.isNew ? 'POST' : 'PUT';

    fetch(url, {
      method: method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
      .then(async res => {
        const data = await res.json().catch(() => ({}));
        if (!res.ok) {
          const errList = [];
          if (data.error) {
            data.error.split(';').forEach(s => errList.push(s.trim()));
          } else {
            errList.push(`Server returned error status ${res.status}`);
          }
          throw errList;
        }
        return data;
      })
      .then(savedInv => {
        showSuccess('Invitation saved successfully!');
        state.invitation = savedInv;

        if (state.isNew) {
          state.isNew = false;
          dom.statusBadge.className = 'badge badge-success';
          dom.statusBadge.textContent = '✓ Saved in Database';
          dom.fieldSlug.readOnly = true;
          dom.pageTitle.textContent = 'Edit: ' + savedInv.title;

          // Update URL without page reload
          window.history.replaceState(null, '', `/admin/invitations/${savedInv.slug}/edit`);
        }

        updateLivePreview();
      })
      .catch(errs => {
        if (Array.isArray(errs)) {
          showErrors(errs);
        } else {
          showErrors([errs.message || 'An unexpected error occurred while saving.']);
        }
      })
      .finally(() => {
        dom.btnSave.disabled = false;
        dom.saveSpinner.style.display = 'none';
        dom.saveLabel.textContent = 'Save Invitation';
      });
  });

  // --- Delete Handler ---

  if (dom.btnDelete) {
    dom.btnDelete.addEventListener('click', () => {
      const slug = state.invitation.slug;
      if (!slug) return;

      const confirmed = window.confirm(
        `Are you sure you want to delete "${state.invitation.title}"?\n\nThis will permanently remove the digital invitation pass and ALL recorded guest RSVPs.`
      );
      if (!confirmed) return;

      fetch(`/api/invitations/${encodeURIComponent(slug)}`, {
        method: 'DELETE',
      })
        .then(res => {
          if (!res.ok) throw new Error('Failed to delete invitation');
          window.location.href = '/admin';
        })
        .catch(err => {
          alert(`Error deleting invitation: ${err.message}`);
        });
    });
  }

  // --- Bootstrap on Page Load ---

  document.addEventListener('DOMContentLoaded', () => {
    const dataEl = document.getElementById('initial-invitation-data');
    if (dataEl && dataEl.textContent.trim()) {
      try {
        state.invitation = JSON.parse(dataEl.textContent.trim());
      } catch (e) {
        console.error('Failed to parse initial invitation data:', e);
      }
    }

    if (!state.invitation) {
      state.invitation = {
        version: '1.0',
        slug: '',
        title: '',
        theme: { id: 'botanical-elegance' },
        sections: [],
      };
    }

    state.isNew = !state.invitation.slug || state.invitation.slug === 'preview';

    populateForm(state.invitation);
    updateLivePreview();
  });
})();
