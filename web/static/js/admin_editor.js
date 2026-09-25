/**
 * Invito Admin Invitation Builder & Live Preview Studio
 * Alpine.js Reactive State Controller
 */

(function () {
  'use strict';

  function slugify(text) {
    return (text || '')
      .toString()
      .toLowerCase()
      .trim()
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .replace(/\s+/g, '-')
      .replace(/[^\w\-]+/g, '')
      .replace(/\-\-+/g, '-')
      .replace(/^-+/, '')
      .replace(/-+$/, '');
  }

  function formatDateTimeLocal(isoStr) {
    if (!isoStr) return '';
    try {
      const d = new Date(isoStr);
      if (isNaN(d.getTime())) return '';
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
    if (!valStr) return undefined;
    try {
      const d = new Date(valStr);
      if (isNaN(d.getTime())) return undefined;
      return d.toISOString();
    } catch {
      return undefined;
    }
  }

  function createDefaultSection(type) {
    switch (type) {
      case 'hero':
        return {
          type: 'hero',
          layout: 'full-bleed',
          eyebrow: 'Celebration',
          vibe: '',
          cover_image_url: '',
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
          text: 'Whatever our souls are made of, his and mine are the same.',
          author: 'Emily Brontë',
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
          name: 'Garden Formal',
          description: 'Cocktail attire or suits. Comfortable footwear recommended for outdoors.',
          palette_hints: ['#2A4738', '#D4AF37', '#EFEBE4'],
          _palette_hints_raw: '#2A4738, #D4AF37, #EFEBE4',
        };
      case 'rsvp':
        return {
          type: 'rsvp',
          enabled: true,
          max_party_size: 2,
          ask_dietary: true,
          ask_song_request: false,
          custom_note: 'Kindly reply at your earliest convenience.',
          deadline: undefined,
          _deadline_local: '',
        };
      case 'rsvp_external':
        return {
          type: 'rsvp_external',
          enabled: true,
          title: 'RSVP Confirmation',
          prompt: 'Please confirm your attendance via our external form',
          form_url: 'https://forms.google.com',
          button_label: 'Open RSVP Form ↗',
          custom_note: '',
          deadline: undefined,
          _deadline_local: '',
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
          links: [{ label: 'Honeymoon Registry', url: 'https://www.zola.com' }],
        };
      case 'closing':
        return {
          type: 'closing',
          message: 'We cannot wait to celebrate with you!',
          signoff: 'With love & gratitude,',
          hosts: '',
        };
      default:
        return { type: type };
    }
  }

  function registerInvitationEditor() {
    Alpine.data('invitationEditor', () => ({
      invitation: {
        version: '1.0',
        slug: '',
        title: '',
        theme: { id: 'botanical-elegance', palette_override: {} },
        location: {},
        sections: [],
      },
      isNew: true,
      manualSlug: false,
      hostsRaw: '',
      dateStartLocal: '',
      dateEndLocal: '',
      selectedNewType: 'quote',
      previewDevice: 'mobile',
      syncStatus: 'Live preview synced',
      saving: false,
      errors: [],
      successMessage: '',
      previewTimer: null,

      init() {
        const dataEl = document.getElementById('initial-invitation-data');
        if (dataEl && dataEl.textContent.trim()) {
          try {
            const parsed = JSON.parse(dataEl.textContent.trim());
            if (parsed) this.invitation = parsed;
          } catch (e) {
            console.error('Failed to parse initial invitation data:', e);
          }
        }
        
        this.isNew = !this.invitation.slug || this.invitation.slug === 'preview';
        this.hostsRaw = (this.invitation.hosts || []).join(', ');
        this.dateStartLocal = formatDateTimeLocal(this.invitation.date_start);
        this.dateEndLocal = formatDateTimeLocal(this.invitation.date_end);

        this.normalizeInivitation(this.invitation)

        this.invitation.sections.forEach(sec => this.normalizeSectionForUI(sec));

        // Expose state to window for inspection in DevTools
        window.editor = this;

        // Reactive live preview synchronization
        this.$watch('invitation', () => this.triggerPreviewSync(), { deep: true });
        this.$watch('hostsRaw', () => this.triggerPreviewSync());
        this.$watch('dateStartLocal', () => this.triggerPreviewSync());
        this.$watch('dateEndLocal', () => this.triggerPreviewSync());

        // Render initial preview
        this.updateLivePreview();
      },

      normalizeInivitation(inv) {
        if (!inv.location) inv.location = {};
        if (!inv.theme) inv.theme = { id: 'botanical-elegance' };
        if (!inv.theme.palette_override) inv.theme.palette_override = {};
        if (!inv.sections) inv.sections = [];
        inv.sections.forEach(sec => this.normalizeSectionForUI(sec));
      },

      normalizeSectionForUI(sec) {
        if (!sec._uid) sec._uid = 'sec_' + Math.random().toString(36).slice(2, 9);
        if (sec.type === 'dress_code') {
          sec._palette_hints_raw = (sec.palette_hints || []).join(', ');
        }
        if (sec.type === 'rsvp' || sec.type === 'rsvp_external') {
          sec._deadline_local = formatDateTimeLocal(sec.deadline);
        }
        if (sec.type === 'carousel' && sec.images) {
          sec.images.forEach(img => {
            if (!img._uid) img._uid = 'img_' + Math.random().toString(36).slice(2, 9);
          });
        }
        if (sec.type === 'timeline' && sec.items) {
          sec.items.forEach(item => {
            if (!item._uid) item._uid = 'item_' + Math.random().toString(36).slice(2, 9);
          });
        }
        if (sec.type === 'faqs' && sec.items) {
          sec.items.forEach(item => {
            if (!item._uid) item._uid = 'faq_' + Math.random().toString(36).slice(2, 9);
          });
        }
        if (sec.type === 'gift_registry' && sec.links) {
          sec.links.forEach(link => {
            if (!link._uid) link._uid = 'reg_' + Math.random().toString(36).slice(2, 9);
          });
        }
      },

      onTitleInput() {
        if (this.isNew && !this.manualSlug) {
          this.invitation.slug = slugify(this.invitation.title);
        }
      },

      getSectionLabel(type) {
        const labels = {
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
        return labels[type] || type;
      },

      addSection(type) {
        const sec = createDefaultSection(type);
        this.normalizeSectionForUI(sec);
        this.invitation.sections.push(sec);
      },

      removeSection(idx) {
        this.invitation.sections.splice(idx, 1);
      },

      moveSection(idx, direction) {
        const target = idx + direction;
        if (target < 0 || target >= this.invitation.sections.length) return;
        const [item] = this.invitation.sections.splice(idx, 1);
        this.invitation.sections.splice(target, 0, item);
      },

      addCarouselImage(sec) {
        if (!sec.images) sec.images = [];
        sec.images.push({ _uid: 'img_' + Math.random().toString(36).slice(2, 9), url: '', caption: '', alt: '' });
      },

      addTimelineItem(sec) {
        if (!sec.items) sec.items = [];
        sec.items.push({ _uid: 'item_' + Math.random().toString(36).slice(2, 9), time: '', title: '', description: '', icon: '' });
      },

      addFAQItem(sec) {
        if (!sec.items) sec.items = [];
        sec.items.push({ _uid: 'faq_' + Math.random().toString(36).slice(2, 9), question: '', answer: '' });
      },

      addRegistryLink(sec) {
        if (!sec.links) sec.links = [];
        sec.links.push({ _uid: 'reg_' + Math.random().toString(36).slice(2, 9), label: '', url: '' });
      },

      async uploadFile(event, targetObj, key) {
        const file = event.target.files?.[0];
        if (!file) return;

        const fd = new FormData();
        fd.append('image', file);
        if (this.invitation.slug) fd.append('slug', this.invitation.slug);

        this.syncStatus = 'Uploading image...';
        try {
          const res = await fetch('/api/upload', { method: 'POST', body: fd });
          const data = await res.json();
          if (!res.ok) throw new Error(data.error || 'Upload failed');
          targetObj[key] = data.url;
          this.syncStatus = 'Image uploaded!';
          event.target.value = '';
        } catch (err) {
          alert('Upload failed: ' + err.message);
          this.syncStatus = 'Upload failed';
        }
      },

      preparePayload() {
        const doc = JSON.parse(JSON.stringify(this.invitation));
        doc.slug = (doc.slug || '').trim();
        doc.title = (doc.title || '').trim();
        doc.hosts = this.hostsRaw ? this.hostsRaw.split(',').map(s => s.trim()).filter(Boolean) : undefined;
        doc.date_start = parseDateTimeLocalToISO(this.dateStartLocal) || undefined;
        doc.date_end = parseDateTimeLocalToISO(this.dateEndLocal) || undefined;

        if (doc.theme && doc.theme.palette_override) {
          const pal = doc.theme.palette_override;
          if (!pal.primary && !pal.background && !pal.accent && !pal.text && !pal.card_bg) {
            delete doc.theme.palette_override;
          }
        }

        if (doc.sections) {
          doc.sections.forEach(sec => {
            delete sec._uid;
            if (sec.type === 'dress_code') {
              sec.palette_hints = sec._palette_hints_raw
                ? sec._palette_hints_raw.split(',').map(s => s.trim()).filter(Boolean)
                : [];
              delete sec._palette_hints_raw;
            }
            if (sec.type === 'rsvp' || sec.type === 'rsvp_external') {
              sec.deadline = parseDateTimeLocalToISO(sec._deadline_local) || undefined;
              delete sec._deadline_local;
            }
            if (sec.images) sec.images.forEach(img => delete img._uid);
            if (sec.items) sec.items.forEach(item => delete item._uid);
            if (sec.links) sec.links.forEach(link => delete link._uid);
          });
        }

        return doc;
      },

      triggerPreviewSync() {
        clearTimeout(this.previewTimer);
        this.syncStatus = 'Syncing preview...';
        this.previewTimer = setTimeout(() => this.updateLivePreview(), 300);
      },

      async updateLivePreview() {
        try {
          const payload = this.preparePayload();
          const res = await fetch('/admin/invitations/preview', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
          });
          if (res.ok) {
            const html = await res.text();
            const iframe = document.getElementById('preview-iframe');
            if (iframe) iframe.srcdoc = html;
            this.syncStatus = 'Live preview synced';
          } else {
            this.syncStatus = 'Preview paused';
          }
        } catch {
          this.syncStatus = 'Preview sync paused';
        }
      },

      openPreviewTab() {
        const slug = this.invitation.slug?.trim();
        if (!this.isNew && slug) {
          window.open(`/i/${encodeURIComponent(slug)}`, '_blank');
        } else {
          this.updateLivePreview();
        }
      },

      async save() {
        this.errors = [];
        const clientErrors = [];
        if (!this.invitation.title?.trim()) clientErrors.push('Event Title is required.');
        if (!this.invitation.slug?.trim()) clientErrors.push('URL Slug is required.');
        if (!this.dateStartLocal) clientErrors.push('Start Date & Time is required.');
        if (!this.invitation.location?.name?.trim()) clientErrors.push('Venue Name is required.');
        if (!this.invitation.location?.address?.trim()) clientErrors.push('Physical Street Address is required.');

        if (clientErrors.length > 0) {
          this.errors = clientErrors;
          window.scrollTo({ top: 0, behavior: 'smooth' });
          return;
        }

        this.saving = true;
        const url = this.isNew ? '/api/invitations' : `/api/invitations/${encodeURIComponent(this.invitation.slug)}`;
        const method = this.isNew ? 'POST' : 'PUT';

        try {
          const res = await fetch(url, {
            method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(this.preparePayload()),
          });
          const data = await res.json().catch(() => ({}));
          if (!res.ok) {
            this.errors = (data.error || 'Failed to save invitation').split(';').map(s => s.trim());
            window.scrollTo({ top: 0, behavior: 'smooth' });
            return;
          }

          this.successMessage = 'Invitation saved successfully!';
          this.invitation = data;
          this.normalizeInivitation(this.invitation);

          if (this.isNew) {
            this.isNew = false;
            window.history.replaceState(null, '', `/admin/invitations/${encodeURIComponent(data.slug)}/edit`);
          }

          setTimeout(() => {
            this.successMessage = '';
          }, 4500);
          this.updateLivePreview();
        } catch (err) {
          this.errors = [err.message || 'An unexpected network error occurred'];
        } finally {
          this.saving = false;
        }
      },

      async deleteInvitation() {
        const slug = this.invitation.slug;
        if (!slug) return;
        if (
          !window.confirm(
            `Are you sure you want to delete "${this.invitation.title}"?\n\nThis will permanently delete the digital invitation pass and ALL recorded guest RSVPs.`
          )
        ) {
          return;
        }
        try {
          const res = await fetch(`/api/invitations/${encodeURIComponent(slug)}`, { method: 'DELETE' });
          if (!res.ok) throw new Error('Delete failed');
          window.location.href = '/admin';
        } catch (err) {
          alert('Failed to delete: ' + err.message);
        }
      },
    }));
  }

  if (window.Alpine) {
    registerInvitationEditor();
  } else {
    document.addEventListener('alpine:init', registerInvitationEditor);
  }
})();
