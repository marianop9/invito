/**
 * Invito - Modern Mobile Invitation Interactivity
 */

document.addEventListener('DOMContentLoaded', () => {
  initCountdown();
  initHeroScrollCue();
  initRSVPForm();
  initCarousel();
});


/* ==========================================================================
   Countdown & Meta Details Badge (Interactive Toggle & Live Timer)
   ========================================================================== */
function initCountdown() {
  const container = document.getElementById('inv-hero-meta') || document.getElementById('inv-countdown');
  if (!container) return;

  const targetDateStr = container.getAttribute('data-target-date');
  if (!targetDateStr) return;

  const targetDate = new Date(targetDateStr).getTime();
  if (isNaN(targetDate)) return;

  const summaryEl = document.getElementById('cd-summary');
  const metaLabel = document.getElementById('inv-meta-label');
  const countdownWrap = document.getElementById('inv-meta-countdown');
  const hasCountdown = container.getAttribute('data-countdown') === 'true';

  function update() {
    const now = new Date().getTime();
    const distance = targetDate - now;

    if (distance <= 0) {
      if (summaryEl) summaryEl.textContent = 'Today!';
      if (countdownWrap) countdownWrap.innerHTML = '🎉 <strong>Today is the day!</strong>';
      return;
    }

    const days = Math.floor(distance / (1000 * 60 * 60 * 24));
    const hours = Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    const mins = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60));

    if (summaryEl) {
      if (days > 0) {
        summaryEl.textContent = `${days}d ${hours}h`;
      } else {
        summaryEl.textContent = `${hours}h ${mins}m`;
      }
    }
  }

  update();
  setInterval(update, 60000);

  // If countdown is enabled, allow clicking/tapping the badge to toggle between Date/Time and Countdown
  if (hasCountdown && metaLabel && countdownWrap) {
    let showingCountdown = false;
    container.setAttribute('title', 'Click to view countdown');

    container.addEventListener('click', () => {
      showingCountdown = !showingCountdown;
      if (showingCountdown) {
        metaLabel.style.display = 'none';
        countdownWrap.style.display = 'inline-flex';
        container.setAttribute('title', 'Click to view event date');
      } else {
        metaLabel.style.display = 'inline-flex';
        countdownWrap.style.display = 'none';
        container.setAttribute('title', 'Click to view countdown');
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
  const track = document.getElementById('carousel-track');
  if (!track) return;

  const slides = track.querySelectorAll('.inv-carousel-slide');
  if (slides.length <= 1) return;

  const dotsContainer = document.getElementById('carousel-dots');
  const dots = dotsContainer ? dotsContainer.querySelectorAll('.inv-carousel-dot') : [];

  let currentIndex = 0;
  let autoScrollTimer = null;
  let isVisible = false;
  const INTERVAL_MS = 3200; // Briefly display each image

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
