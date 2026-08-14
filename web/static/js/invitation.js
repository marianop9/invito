/**
 * Invito - Modern Mobile Invitation Interactivity
 */

document.addEventListener('DOMContentLoaded', () => {
  initCountdown();
  initRSVPForm();
});

/* ==========================================================================
   Countdown Timer (Compact Summary)
   ========================================================================== */
function initCountdown() {
  const container = document.getElementById('inv-countdown');
  if (!container) return;

  const targetDateStr = container.getAttribute('data-target-date');
  if (!targetDateStr) return;

  const targetDate = new Date(targetDateStr).getTime();
  if (isNaN(targetDate)) return;

  const summaryEl = document.getElementById('cd-summary');

  function update() {
    const now = new Date().getTime();
    const distance = targetDate - now;

    if (distance <= 0) {
      container.innerHTML = '<span>🎉 Today is the day!</span>';
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
