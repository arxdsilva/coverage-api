const orgRefresh = document.getElementById('orgRefresh');
const applyFilters = document.getElementById('applyFilters');
const orgInput = document.getElementById('orgInput');
const repoInput = document.getElementById('repoInput');
const authorInput = document.getElementById('authorInput');
const windowDaysInput = document.getElementById('windowDaysInput');
const sortInput = document.getElementById('sortInput');
const minIdleHoursInput = document.getElementById('minIdleHoursInput');
const minOpenHoursInput = document.getElementById('minOpenHoursInput');
const includeDraftsInput = document.getElementById('includeDraftsInput');

const orgAutoRefreshInterval = document.getElementById('orgAutoRefreshInterval');
const orgAutoRefreshStatus = document.getElementById('orgAutoRefreshStatus');
const orgAutoRefreshProgressBar = document.getElementById('orgAutoRefreshProgressBar');
const orgLastUpdated = document.getElementById('orgLastUpdated');

const kpiHanging = document.getElementById('kpiHanging');
const kpiOpenConsidered = document.getElementById('kpiOpenConsidered');
const kpiReviewEvents = document.getElementById('kpiReviewEvents');
const kpiTopReviewer = document.getElementById('kpiTopReviewer');

const orgLeaderboardBody = document.getElementById('orgLeaderboardBody');
const orgHangingBody = document.getElementById('orgHangingBody');

const appShell = document.getElementById('appShell');
const toggleSidebar = document.getElementById('toggleSidebar');

const sidebarCollapsedKey = 'opencoverage.sidebarCollapsed.orgInsights';
const orgAutoRefreshStorageKey = 'opencoverage.autoRefresh.orgInsights';
const orgDefaultAutoRefreshInterval = '1h';
const allowedWindowDays = Object.freeze(['30', '60', '90']);
const orgAutoRefreshIntervals = Object.freeze({
  off: 0,
  '15s': 15000,
  '30s': 30000,
  '60s': 60000,
  '5m': 300000,
  '1h': 3600000,
});

let refreshTimeoutId = 0;
let refreshInFlight = false;
let refreshCountdownIntervalId = 0;
let nextRefreshAt = 0;
let refreshDurationMs = 0;

orgRefresh.addEventListener('click', async () => {
  await performRefresh('manual');
});

applyFilters.addEventListener('click', async () => {
  syncUrlFromFilters();
  await performRefresh('manual');
});

orgAutoRefreshInterval.addEventListener('change', () => {
  persistAutoRefreshInterval(orgAutoRefreshInterval.value);
  scheduleAutoRefresh();
});

toggleSidebar.addEventListener('click', () => {
  const shouldCollapse = !appShell.classList.contains('sidebar-collapsed');
  setSidebarCollapsed(shouldCollapse);
});

document.addEventListener('visibilitychange', () => {
  if (document.hidden) {
    clearAutoRefresh();
    updateAutoRefreshStatus();
    return;
  }
  scheduleAutoRefresh();
});

initializeSidebarState();
restoreFiltersFromUrl();
initializeAutoRefreshControl();
performRefresh('initial').then(() => {
  scheduleAutoRefresh();
});

function initializeAutoRefreshControl() {
  const persisted = window.localStorage.getItem(orgAutoRefreshStorageKey);
  const nextValue = Object.prototype.hasOwnProperty.call(orgAutoRefreshIntervals, persisted)
    ? persisted
    : orgDefaultAutoRefreshInterval;
  orgAutoRefreshInterval.value = nextValue;
  updateAutoRefreshStatus();
}

function getAutoRefreshIntervalValue() {
  const selectedValue = orgAutoRefreshInterval.value;
  return Object.prototype.hasOwnProperty.call(orgAutoRefreshIntervals, selectedValue)
    ? selectedValue
    : orgDefaultAutoRefreshInterval;
}

function getAutoRefreshIntervalMs() {
  return orgAutoRefreshIntervals[getAutoRefreshIntervalValue()] || 0;
}

function persistAutoRefreshInterval(value) {
  const nextValue = Object.prototype.hasOwnProperty.call(orgAutoRefreshIntervals, value)
    ? value
    : orgDefaultAutoRefreshInterval;
  window.localStorage.setItem(orgAutoRefreshStorageKey, nextValue);
}

function clearAutoRefresh() {
  if (refreshTimeoutId) {
    window.clearTimeout(refreshTimeoutId);
    refreshTimeoutId = 0;
  }
  if (refreshCountdownIntervalId) {
    window.clearInterval(refreshCountdownIntervalId);
    refreshCountdownIntervalId = 0;
  }
}

function setAutoRefreshProgress(progressRatio) {
  if (!orgAutoRefreshProgressBar) return;
  const safeRatio = Math.max(0, Math.min(1, progressRatio));
  orgAutoRefreshProgressBar.style.transform = `scaleX(${safeRatio})`;
}

function formatRemainingTime(ms) {
  const totalSeconds = Math.max(0, Math.ceil(ms / 1000));
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  if (minutes > 0) {
    return `${minutes}m ${seconds}s`;
  }
  return `${seconds}s`;
}

function updateAutoRefreshStatus() {
  const intervalLabel = getAutoRefreshIntervalValue();
  if (intervalLabel === 'off') {
    orgAutoRefreshStatus.textContent = 'Auto refresh is off.';
    setAutoRefreshProgress(0);
    return;
  }

  if (document.hidden) {
    orgAutoRefreshStatus.textContent = `Paused (${intervalLabel}) while tab is hidden.`;
    setAutoRefreshProgress(0);
    return;
  }

  if (refreshInFlight) {
    orgAutoRefreshStatus.textContent = `Refreshing now (${intervalLabel}).`;
    setAutoRefreshProgress(0);
    return;
  }

  if (!nextRefreshAt) {
    orgAutoRefreshStatus.textContent = `Scheduled every ${intervalLabel}.`;
    setAutoRefreshProgress(0);
    return;
  }

  const remainingMs = nextRefreshAt - Date.now();
  orgAutoRefreshStatus.textContent = `Next refresh in ${formatRemainingTime(remainingMs)} (${intervalLabel}).`;
  const denominator = refreshDurationMs || getAutoRefreshIntervalMs() || 1;
  setAutoRefreshProgress(remainingMs / denominator);
}

function scheduleAutoRefresh() {
  clearAutoRefresh();
  nextRefreshAt = 0;
  refreshDurationMs = 0;

  const intervalMs = getAutoRefreshIntervalMs();
  if (!intervalMs || document.hidden) {
    updateAutoRefreshStatus();
    return;
  }

  refreshDurationMs = intervalMs;
  nextRefreshAt = Date.now() + intervalMs;
  updateAutoRefreshStatus();
  setAutoRefreshProgress(1);

  refreshCountdownIntervalId = window.setInterval(() => {
    updateAutoRefreshStatus();
  }, 200);

  refreshTimeoutId = window.setTimeout(async () => {
    if (document.hidden || refreshInFlight) {
      scheduleAutoRefresh();
      return;
    }
    await performRefresh('auto');
  }, intervalMs);
}

function parseRepos(value) {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter((item) => item.length > 0);
}

function syncUrlFromFilters() {
  const params = new URLSearchParams();
  if (orgInput.value.trim()) params.set('org', orgInput.value.trim());
  params.set('windowDays', normalizeWindowDays(windowDaysInput.value));
  if (authorInput.value.trim()) params.set('author', authorInput.value.trim());
  if (sortInput.value) params.set('sort', sortInput.value);
  if (minIdleHoursInput.value) params.set('minIdleHours', minIdleHoursInput.value);
  if (minOpenHoursInput.value) params.set('minOpenHours', minOpenHoursInput.value);
  if (includeDraftsInput.checked) params.set('includeDrafts', 'true');

  const repos = parseRepos(repoInput.value);
  repos.forEach((repo) => params.append('repo', repo));

  const nextUrl = `${window.location.pathname}?${params.toString()}`;
  window.history.replaceState({}, '', nextUrl);
}

function restoreFiltersFromUrl() {
  const params = new URLSearchParams(window.location.search);
  orgInput.value = params.get('org') || '';
  windowDaysInput.value = normalizeWindowDays(params.get('windowDays') || '30');
  authorInput.value = params.get('author') || '';
  sortInput.value = params.get('sort') || 'staleness_desc';
  minIdleHoursInput.value = params.get('minIdleHours') || '48';
  minOpenHoursInput.value = params.get('minOpenHours') || '72';
  includeDraftsInput.checked = params.get('includeDrafts') === 'true';

  const repos = params.getAll('repo');
  repoInput.value = repos.join(', ');
}

async function performRefresh(source = 'manual') {
  if (refreshInFlight) return false;
  const org = orgInput.value.trim();
  if (!org) {
    renderError('Organization is required.');
    return false;
  }

  refreshInFlight = true;
  clearAutoRefresh();
  nextRefreshAt = 0;
  refreshDurationMs = 0;
  updateAutoRefreshStatus();

  if (source === 'manual') {
    orgRefresh.disabled = true;
    orgRefresh.textContent = 'Refreshing...';
  }

  try {
    const params = buildQueryParams();
    const [leaderboardRes, hangingRes] = await Promise.all([
      fetch(`/api/github/orgs/${encodeURIComponent(org)}/reviewers/leaderboard?${params.leaderboard}`),
      fetch(`/api/github/orgs/${encodeURIComponent(org)}/pull-requests/hanging?${params.hanging}`),
    ]);

    if (!leaderboardRes.ok) {
      throw new Error(`failed to load reviewers leaderboard (${leaderboardRes.status})`);
    }
    if (!hangingRes.ok) {
      throw new Error(`failed to load hanging pull requests (${hangingRes.status})`);
    }

    const leaderboard = await leaderboardRes.json();
    const hanging = await hangingRes.json();
    renderData(leaderboard, hanging);
    return true;
  } catch (err) {
    renderError(err.message || 'Failed to refresh org insights.');
    return false;
  } finally {
    if (source === 'manual') {
      orgRefresh.disabled = false;
      orgRefresh.textContent = 'Refresh';
    }

    refreshInFlight = false;
    scheduleAutoRefresh();
  }
}

function buildQueryParams() {
  const repos = parseRepos(repoInput.value);
  const shared = new URLSearchParams();
  repos.forEach((repo) => shared.append('repo', repo));

  const leaderboard = new URLSearchParams(shared.toString());
  leaderboard.set('windowDays', normalizeWindowDays(windowDaysInput.value));

  const hanging = new URLSearchParams(shared.toString());
  hanging.set('limit', '200');
  hanging.set('minIdleHours', minIdleHoursInput.value || '48');
  hanging.set('minOpenHours', minOpenHoursInput.value || '72');
  hanging.set('sort', sortInput.value || 'staleness_desc');
  if (authorInput.value.trim()) {
    hanging.set('author', authorInput.value.trim());
  }
  if (includeDraftsInput.checked) {
    hanging.set('includeDrafts', 'true');
  }

  return {
    leaderboard: leaderboard.toString(),
    hanging: hanging.toString(),
  };
}

function renderData(leaderboard, hanging) {
  renderKpis(leaderboard, hanging);
  renderLeaderboardRows(leaderboard.reviewers || []);
  renderHangingRows(hanging.items || []);

  const generatedAt = hanging.generatedAt || new Date().toISOString();
  orgLastUpdated.textContent = `Updated at ${new Date(generatedAt).toLocaleString()}`;
}

function renderKpis(leaderboard, hanging) {
  const hangingCount = hanging?.summary?.hangingPullRequests;
  const openConsidered = hanging?.summary?.openPullRequestsConsidered;
  const reviewEvents = leaderboard?.summary?.totalReviewEvents;
  const topReviewer = leaderboard?.reviewers?.[0]?.login;

  kpiHanging.textContent = Number.isFinite(hangingCount) ? String(hangingCount) : '-';
  kpiOpenConsidered.textContent = Number.isFinite(openConsidered) ? String(openConsidered) : '-';
  kpiReviewEvents.textContent = Number.isFinite(reviewEvents) ? String(reviewEvents) : '-';
  kpiTopReviewer.textContent = topReviewer || '-';
}

function renderLeaderboardRows(reviewers) {
  if (!reviewers.length) {
    orgLeaderboardBody.innerHTML = '<tr><td colspan="8" class="muted">No review activity found for current filters.</td></tr>';
    return;
  }

  orgLeaderboardBody.innerHTML = reviewers
    .map((reviewer) => {
      const latest = reviewer.latestReviewAt ? new Date(reviewer.latestReviewAt).toLocaleString() : '-';
      return `
        <tr>
          <td>${escapeHtml(reviewer.login || '-')}</td>
          <td>${reviewer.totalReviews ?? '-'}</td>
          <td>${reviewer.approvals ?? '-'}</td>
          <td>${reviewer.changeRequests ?? '-'}</td>
          <td>${reviewer.comments ?? '-'}</td>
          <td>${reviewer.uniquePullRequestsReviewed ?? '-'}</td>
          <td>${reviewer.reposReviewed ?? '-'}</td>
          <td>${escapeHtml(latest)}</td>
        </tr>
      `;
    })
    .join('');
}

function renderHangingRows(items) {
  if (!items.length) {
    orgHangingBody.innerHTML = '<tr><td colspan="8" class="muted">No hanging pull requests for current filters.</td></tr>';
    return;
  }

  orgHangingBody.innerHTML = items
    .map((item) => {
      const reasons = (item.reasons || []).join(', ');
      return `
        <tr>
          <td>${escapeHtml(item.repository || '-')}</td>
          <td>#${item.number} ${escapeHtml(item.title || '')}</td>
          <td>${escapeHtml(item.author || '-')}</td>
          <td>${escapeHtml(reasons || '-')}</td>
          <td>${item.idleHours ?? '-'}h</td>
          <td>${item.ageHours ?? '-'}h</td>
          <td>${escapeHtml(item.reviewState || '-')}</td>
          <td><a href="${escapeHtml(item.url || '#')}" target="_blank" rel="noreferrer">Open</a></td>
        </tr>
      `;
    })
    .join('');
}

function renderError(message) {
  orgLeaderboardBody.innerHTML = `<tr><td colspan="8" class="muted">${escapeHtml(message)}</td></tr>`;
  orgHangingBody.innerHTML = `<tr><td colspan="8" class="muted">${escapeHtml(message)}</td></tr>`;
  orgLastUpdated.textContent = 'Failed to load data.';
  kpiHanging.textContent = '-';
  kpiOpenConsidered.textContent = '-';
  kpiReviewEvents.textContent = '-';
  kpiTopReviewer.textContent = '-';
}

function initializeSidebarState() {
  const persisted = window.localStorage.getItem(sidebarCollapsedKey);
  setSidebarCollapsed(persisted === 'true');
}

function normalizeWindowDays(value) {
  if (allowedWindowDays.includes(String(value))) {
    return String(value);
  }
  return '30';
}

function setSidebarCollapsed(collapsed) {
  appShell.classList.toggle('sidebar-collapsed', collapsed);
  toggleSidebar.textContent = collapsed ? '▸' : '◂';
  toggleSidebar.setAttribute('aria-label', collapsed ? 'Expand sidebar' : 'Collapse sidebar');
  toggleSidebar.setAttribute('title', collapsed ? 'Expand sidebar' : 'Collapse sidebar');
  toggleSidebar.setAttribute('aria-expanded', String(!collapsed));
  window.localStorage.setItem(sidebarCollapsedKey, String(collapsed));
}

function escapeHtml(value) {
  return String(value)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');
}
