/**
 * Admin Dashboard Client Interactions — Invito Planner
 * Instant search & status filtering for RSVP management
 */

document.addEventListener('DOMContentLoaded', () => {
  initRSVPTableFilters();
});

function initRSVPTableFilters() {
  const searchInput = document.getElementById('rsvp-search-input');
  const clearBtn = document.getElementById('rsvp-search-clear');
  const filterPills = document.querySelectorAll('.filter-pill');
  const tableBody = document.getElementById('rsvp-table-body');
  const noMatchRow = document.getElementById('no-matching-rsvps');
  const countIndicator = document.getElementById('rsvp-count-indicator');

  if (!tableBody) return;

  const rows = Array.from(tableBody.querySelectorAll('.rsvp-row'));
  const totalCount = rows.length;

  let currentFilter = 'all';
  let searchQuery = '';

  function applyFilters() {
    let visibleCount = 0;

    rows.forEach(row => {
      const name = (row.dataset.name || '').toLowerCase();
      const email = (row.dataset.email || '').toLowerCase();
      const isAttending = row.dataset.attending === 'true';
      const hasDietary = row.dataset.hasDietary === 'true';

      // 1. Check Search query
      const matchesSearch = !searchQuery || name.includes(searchQuery) || email.includes(searchQuery);

      // 2. Check Status filter
      let matchesFilter = true;
      if (currentFilter === 'attending') {
        matchesFilter = isAttending;
      } else if (currentFilter === 'declined') {
        matchesFilter = !isAttending;
      } else if (currentFilter === 'dietary') {
        matchesFilter = hasDietary;
      }

      const isVisible = matchesSearch && matchesFilter;
      row.style.display = isVisible ? '' : 'none';

      if (isVisible) {
        visibleCount++;
      }
    });

    // Toggle "No matches" row
    if (noMatchRow) {
      noMatchRow.style.display = (visibleCount === 0 && totalCount > 0) ? '' : 'none';
    }

    // Update count indicator
    if (countIndicator) {
      if (searchQuery || currentFilter !== 'all') {
        countIndicator.textContent = `Showing ${visibleCount} of ${totalCount} guests`;
      } else {
        countIndicator.textContent = `${totalCount} total guests`;
      }
    }
  }

  // Search input listener
  if (searchInput) {
    searchInput.addEventListener('input', (e) => {
      searchQuery = e.target.value.trim().toLowerCase();
      if (clearBtn) {
        clearBtn.style.display = searchQuery ? 'block' : 'none';
      }
      applyFilters();
    });
  }

  // Clear button listener
  if (clearBtn && searchInput) {
    clearBtn.addEventListener('click', () => {
      searchInput.value = '';
      searchQuery = '';
      clearBtn.style.display = 'none';
      searchInput.focus();
      applyFilters();
    });
  }

  // Filter pills click listeners
  filterPills.forEach(pill => {
    pill.addEventListener('click', () => {
      filterPills.forEach(p => p.classList.remove('active'));
      pill.classList.add('active');
      currentFilter = pill.dataset.filter || 'all';
      applyFilters();
    });
  });
}
