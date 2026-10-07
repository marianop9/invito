/**
 * Invito - Modern Mobile Invitation Interactivity
 */

document.addEventListener('DOMContentLoaded', () => {
  initCountdown();
  initHeroScrollCue();
  initRSVPForm();
  initCarousel();
  initMusicAndSplash();
});


/* ==========================================================================
   Countdown & Meta Details Badge (Interactive Toggle & Live Timer)
   ========================================================================== */
function initCountdown() {
  const countdownEl = document.getElementById('inv-countdown');
  const metaContainer = document.getElementById('inv-hero-meta');
  const container = countdownEl || metaContainer;
  if (!container) return;

  const targetDateStr = container.getAttribute('data-target-date');
  if (!targetDateStr) return;

  const targetDate = new Date(targetDateStr).getTime();
  if (isNaN(targetDate)) return;

  const daysEl = document.getElementById('cd-days');
  const hoursEl = document.getElementById('cd-hours');
  const minsEl = document.getElementById('cd-mins');
  const secsEl = document.getElementById('cd-secs');

  const summaryEl = document.getElementById('cd-summary');
  const metaLabel = document.getElementById('inv-meta-label');
  const countdownWrap = document.getElementById('inv-meta-countdown');

  function pad(num) {
    return String(num).padStart(2, '0');
  }

  function update() {
    const now = new Date().getTime();
    const distance = targetDate - now;

    if (distance <= 0) {
      if (daysEl) daysEl.textContent = '00';
      if (hoursEl) hoursEl.textContent = '00';
      if (minsEl) minsEl.textContent = '00';
      if (secsEl) secsEl.textContent = '00';
      if (summaryEl) summaryEl.textContent = '¡Hoy!';
      if (countdownWrap) countdownWrap.innerHTML = '🎉 <strong>¡Llegó el gran día!</strong>';
      return;
    }

    const days = Math.floor(distance / (1000 * 60 * 60 * 24));
    const hours = Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    const mins = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60));
    const secs = Math.floor((distance % (1000 * 60)) / 1000);

    if (daysEl) daysEl.textContent = pad(days);
    if (hoursEl) hoursEl.textContent = pad(hours);
    if (minsEl) minsEl.textContent = pad(mins);
    if (secsEl) secsEl.textContent = pad(secs);

    if (summaryEl) {
      if (days > 0) {
        summaryEl.textContent = `${days}d ${hours}h`;
      } else {
        summaryEl.textContent = `${hours}h ${mins}m`;
      }
    }
  }

  update();
  setInterval(update, 1000);

  // If legacy countdown toggle exists on meta container
  if (metaContainer && metaContainer.getAttribute('data-countdown') === 'true' && metaLabel && countdownWrap) {
    let showingCountdown = false;
    metaContainer.setAttribute('title', 'Tocar para ver cuenta regresiva');

    metaContainer.addEventListener('click', () => {
      showingCountdown = !showingCountdown;
      if (showingCountdown) {
        metaLabel.style.display = 'none';
        countdownWrap.style.display = 'inline-flex';
        metaContainer.setAttribute('title', 'Tocar para ver fecha');
      } else {
        metaLabel.style.display = 'inline-flex';
        countdownWrap.style.display = 'none';
        metaContainer.setAttribute('title', 'Tocar para ver cuenta regresiva');
      }
    });
  }
}

/* ==========================================================================
   Hero Micro Scroll Cue (Smooth Scroll to Next Content Block)
   ========================================================================== */
function initHeroScrollCue() {
  const cue = document.getElementById('hero-scroll-cue');
  if (!cue) return;

  cue.addEventListener('click', (e) => {
    e.preventDefault();
    const hero = document.getElementById('section-hero');
    const target = (hero && hero.nextElementSibling) ? hero.nextElementSibling : document.getElementById('section-details');
    if (target) {
      const targetTop = target.getBoundingClientRect().top + window.pageYOffset;
      smoothWindowScrollTo(targetTop, 650);
    }
  });
}

function smoothWindowScrollTo(targetY, duration = 650) {
  const startY = window.pageYOffset || document.documentElement.scrollTop;
  const change = targetY - startY;
  if (Math.abs(change) < 2) {
    window.scrollTo(0, targetY);
    return;
  }

  const startTime = performance.now();

  function easeInOutCubic(t) {
    return t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2;
  }

  function step(currentTime) {
    const elapsed = currentTime - startTime;
    const progress = Math.min(elapsed / duration, 1);
    const ease = easeInOutCubic(progress);

    window.scrollTo(0, startY + change * ease);

    if (progress < 1) {
      requestAnimationFrame(step);
    } else {
      window.scrollTo(0, targetY);
    }
  }

  requestAnimationFrame(step);
}

/* ==========================================================================
   RSVP Form Handling & Smooth Feedback
   ========================================================================== */
function initRSVPForm() {
  const form = document.getElementById('rsvp-form');
  if (!form) return;

  const toggleOptions = form.querySelectorAll('.toggle-option');
  const guestsGroup = document.getElementById('rsvp-guests-group');
  const dietaryGroup = document.getElementById('rsvp-dietary-group');

  // Handle Radio Option Styles
  toggleOptions.forEach(label => {
    const input = label.querySelector('input[type="radio"]');
    if (!input) return;

    input.addEventListener('change', () => {
      toggleOptions.forEach(l => l.classList.remove('active'));
      label.classList.add('active');

      const isAttending = input.value === 'true';
      if (guestsGroup) guestsGroup.style.display = isAttending ? 'flex' : 'none';
      if (dietaryGroup) dietaryGroup.style.display = isAttending ? 'flex' : 'none';
    });
  });

  // Handle Submission
  form.addEventListener('submit', async (e) => {
    e.preventDefault();

    const submitBtn = document.getElementById('btn-submit-rsvp');
    const originalBtnText = submitBtn ? submitBtn.innerText : 'Confirm Attendance';
    if (submitBtn) {
      submitBtn.disabled = true;
      submitBtn.innerText = 'Submitting...';
    }

    const formData = new FormData(form);
    const payload = {
      name: formData.get('name'),
      email: formData.get('email'),
      attending: formData.get('attending') === 'true',
      guest_count: parseInt(formData.get('guest_count') || '1', 10),
      dietary_needs: formData.get('dietary_needs') || '',
      song_request: formData.get('song_request') || '',
      personal_message: formData.get('personal_message') || ''
    };

    try {
      const response = await fetch(form.action, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
        body: JSON.stringify(payload)
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to submit RSVP');
      }

      // Render clean compact success state
      const card = document.getElementById('rsvp-card-inner');
      if (card) {
        const guestName = payload.name.split(' ')[0] || 'Friend';
        card.innerHTML = `
          <div class="rsvp-success-compact">
            <div class="rsvp-success-check">✓</div>
            <h3 style="font-size: 1.125rem; font-weight: 700; color: var(--inv-text); margin-bottom: 0.25rem;">
              Response Recorded, ${escapeHTML(guestName)}!
            </h3>
            <p style="font-size: 0.8125rem; color: var(--inv-text-muted);">
              ${payload.attending 
                ? 'We look forward to seeing you at the celebration.' 
                : 'Thank you for letting us know. You will be missed!'}
            </p>
          </div>
        `;
      }
    } catch (err) {
      alert('Error submitting RSVP: ' + err.message);
      if (submitBtn) {
        submitBtn.disabled = false;
        submitBtn.innerText = originalBtnText;
      }
    }
  });
}

function escapeHTML(str) {
  const p = document.createElement('p');
  p.textContent = str;
  return p.innerHTML;
}

/* ==========================================================================
   Auto-Scrolling Photo Carousel (Viewport-Aware & Infinite Wrap)
   ========================================================================== */
function initCarousel() {
  const containers = document.querySelectorAll('[data-carousel="inv-carousel"], .inv-carousel');
  containers.forEach(container => setupSingleCarousel(container));
}

function setupSingleCarousel(container) {
  const track = container.querySelector('.inv-carousel-track');
  if (!track) return;

  const slides = track.querySelectorAll('.inv-carousel-slide');
  if (slides.length <= 1) return;

  const dotsContainer = container.querySelector('.inv-carousel-dots');
  const dots = dotsContainer ? dotsContainer.querySelectorAll('.inv-carousel-dot') : [];

  let currentIndex = 0;
  let autoScrollTimer = null;
  let isVisible = false;
  const INTERVAL_MS = 2500; // Briefly display each image

  function getActiveIndex() {
    const trackRect = track.getBoundingClientRect();
    const trackCenter = trackRect.left + trackRect.width / 2;
    let closestIndex = 0;
    let minDistance = Infinity;

    slides.forEach((slide, idx) => {
      const slideRect = slide.getBoundingClientRect();
      const slideCenter = slideRect.left + slideRect.width / 2;
      const dist = Math.abs(trackCenter - slideCenter);
      if (dist < minDistance) {
        minDistance = dist;
        closestIndex = idx;
      }
    });

    return closestIndex;
  }

  function updateActiveSlide(activeIndex) {
    slides.forEach((slide, idx) => {
      if (idx === activeIndex) {
        slide.classList.add('is-active');
      } else {
        slide.classList.remove('is-active');
      }
    });
    dots.forEach((dot, idx) => {
      if (idx === activeIndex) {
        dot.classList.add('active');
        dot.setAttribute('aria-selected', 'true');
      } else {
        dot.classList.remove('active');
        dot.setAttribute('aria-selected', 'false');
      }
    });
  }

  function smoothTrackScrollTo(element, targetX, duration = 550) {
    const startX = element.scrollLeft;
    const change = targetX - startX;
    if (Math.abs(change) < 2) {
      element.scrollLeft = targetX;
      return;
    }

    if (element._scrollAnimId) {
      cancelAnimationFrame(element._scrollAnimId);
      element._scrollAnimId = null;
    }

    // Disable CSS scroll-snap during programmatic slide animation so browser snap doesn't pop
    element.style.scrollSnapType = 'none';

    const startTime = performance.now();

    function easeOutCubic(t) {
      return 1 - Math.pow(1 - t, 3);
    }

    function step(currentTime) {
      const elapsed = currentTime - startTime;
      const progress = Math.min(elapsed / duration, 1);
      const ease = easeOutCubic(progress);

      element.scrollLeft = startX + change * ease;

      if (progress < 1) {
        element._scrollAnimId = requestAnimationFrame(step);
      } else {
        element.scrollLeft = targetX;
        element._scrollAnimId = null;
        // Re-enable scroll-snap for touch gestures
        setTimeout(() => {
          if (!element._scrollAnimId) {
            element.style.scrollSnapType = '';
          }
        }, 30);
      }
    }

    element._scrollAnimId = requestAnimationFrame(step);
  }

  function scrollToSlide(index, animated = true) {
    if (index < 0 || index >= slides.length) return;
    currentIndex = index;
    const targetSlide = slides[index];
    const offset = targetSlide.offsetLeft - (track.clientWidth - targetSlide.clientWidth) / 2;
    const targetLeft = Math.max(0, offset);

    if (animated) {
      smoothTrackScrollTo(track, targetLeft, 550);
    } else {
      track.scrollLeft = targetLeft;
    }
    updateActiveSlide(index);
  }

  function nextSlide() {
    const nextIndex = (currentIndex + 1) % slides.length; // Wrap around to first image
    scrollToSlide(nextIndex, true);
  }

  function startAutoScroll() {
    stopAutoScroll();
    if (isVisible && !document.hidden) {
      autoScrollTimer = setInterval(nextSlide, INTERVAL_MS);
    }
  }

  function stopAutoScroll() {
    if (autoScrollTimer) {
      clearInterval(autoScrollTimer);
      autoScrollTimer = null;
    }
  }

  function resetAutoScroll() {
    stopAutoScroll();
    startAutoScroll();
  }

  // Dot button clicks
  dots.forEach(dot => {
    dot.addEventListener('click', () => {
      const targetIndex = parseInt(dot.getAttribute('data-dot-index'), 10);
      if (!isNaN(targetIndex)) {
        scrollToSlide(targetIndex, true);
        resetAutoScroll();
      }
    });
  });

  // Track scroll synchronization
  let scrollTimeout;
  track.addEventListener('scroll', () => {
    // If programmatic smooth scroll is active, ignore to avoid interrupting
    if (track._scrollAnimId) return;

    if (scrollTimeout) cancelAnimationFrame(scrollTimeout);
    scrollTimeout = requestAnimationFrame(() => {
      const activeIdx = getActiveIndex();
      if (activeIdx !== currentIndex) {
        currentIndex = activeIdx;
        updateActiveSlide(activeIdx);
      }
    });
  }, { passive: true });

  // Pause on user interaction (hover or touch)
  track.addEventListener('mouseenter', stopAutoScroll);
  track.addEventListener('mouseleave', () => {
    if (isVisible && !document.hidden) startAutoScroll();
  });
  track.addEventListener('touchstart', () => {
    if (track._scrollAnimId) {
      cancelAnimationFrame(track._scrollAnimId);
      track._scrollAnimId = null;
      track.style.scrollSnapType = '';
    }
    stopAutoScroll();
  }, { passive: true });
  track.addEventListener('touchend', () => {
    setTimeout(() => {
      if (isVisible && !document.hidden) startAutoScroll();
    }, 1200);
  }, { passive: true });

  // Pause when tab / window is hidden, resume when active
  document.addEventListener('visibilitychange', () => {
    if (document.hidden) {
      stopAutoScroll();
    } else {
      startAutoScroll();
    }
  });

  // Observe visibility in viewport to trigger auto-scroll only when visible
  if ('IntersectionObserver' in window) {
    const observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          isVisible = true;
          startAutoScroll();
        } else {
          isVisible = false;
          stopAutoScroll();
        }
      });
    }, { threshold: 0.35 });

    observer.observe(track);
  } else {
    // Fallback if IntersectionObserver not available
    isVisible = true;
    startAutoScroll();
  }
}

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
    audio.addEventListener('ended', () => {
      setPlayingState(false);
    });

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

