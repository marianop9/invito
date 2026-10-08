/**
 * Save the Date — Minimal Live Countdown Timer
 * Pads single digits with leading zeros (00:00:00:00) from days to seconds
 */

document.addEventListener('DOMContentLoaded', () => {
  const countdownEl = document.getElementById('std-countdown');
  if (!countdownEl) return;

  const targetDateStr = countdownEl.getAttribute('data-target-date');
  if (!targetDateStr) return;

  const targetDate = new Date(targetDateStr).getTime();
  if (isNaN(targetDate)) return;

  const daysEl = document.getElementById('std-cd-days');
  const hoursEl = document.getElementById('std-cd-hours');
  const minsEl = document.getElementById('std-cd-mins');
  const secsEl = document.getElementById('std-cd-secs');

  function pad(num) {
    return String(Math.max(0, num)).padStart(2, '0');
  }

  function tick() {
    const now = Date.now();
    const distance = targetDate - now;

    if (distance <= 0) {
      if (daysEl) daysEl.textContent = '00';
      if (hoursEl) hoursEl.textContent = '00';
      if (minsEl) minsEl.textContent = '00';
      if (secsEl) secsEl.textContent = '00';
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
  }

  tick();
  setInterval(tick, 1000);
});
