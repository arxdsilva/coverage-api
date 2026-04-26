const refreshProjects = document.getElementById('refreshProjects');
const projectSelector = document.getElementById('projectSelector');
const projectGroupFilter = document.getElementById('projectGroupFilter');
const projectSearchInput = document.getElementById('projectSearchInput');
const integrationScreenProjectTitle = document.getElementById('integrationScreenProjectTitle');
const integrationScreenProjectMeta = document.getElementById('integrationScreenProjectMeta');
const integrationStatus = document.getElementById('integrationStatus');
const integrationPassRate = document.getElementById('integrationPassRate');
const integrationFailedSpecsCount = document.getElementById('integrationFailedSpecsCount');
const integrationDelta = document.getElementById('integrationDelta');
const integrationDuration = document.getElementById('integrationDuration');
const integrationBranchFilter = document.getElementById('integrationBranchFilter');
const integrationStatusFilter = document.getElementById('integrationStatusFilter');
const integrationReload = document.getElementById('integrationReload');
const integrationAutoRefreshInterval = document.getElementById('integrationAutoRefreshInterval');
const integrationAutoRefreshStatus = document.getElementById('integrationAutoRefreshStatus');
const integrationAutoRefreshProgressBar = document.getElementById('integrationAutoRefreshProgressBar');
const integrationRunChain = document.getElementById('integrationRunChain');
const integrationRunsBody = document.getElementById('integrationRunsBody');
const integrationFailedSpecsBody = document.getElementById('integrationFailedSpecsBody');
const openIntegrationHeatmap = document.getElementById('openIntegrationHeatmap');
const closeIntegrationHeatmap = document.getElementById('closeIntegrationHeatmap');
const integrationHeatmapOverlay = document.getElementById('integrationHeatmapOverlay');
const heatmapBranchFilter = document.getElementById('heatmapBranchFilter');
const heatmapStatusFilter = document.getElementById('heatmapStatusFilter');
const heatmapReload = document.getElementById('heatmapReload');
const integrationHeatmap = document.getElementById('integrationHeatmap');
const appShell = document.getElementById('appShell');
const toggleSidebar = document.getElementById('toggleSidebar');

let projects = [];
let filteredProjects = [];
let selectedProjectId = null;
let selectedIntegrationRunId = null;
let currentIntegrationRunItems = [];
const allGroupsFilterValue = '__all__';
const ungroupedFilterValue = '__ungrouped__';
const sidebarCollapsedKey = 'opencoverage.sidebarCollapsed.integration';
const integrationAutoRefreshStorageKey = 'opencoverage.autoRefresh.integration';
const integrationDefaultAutoRefreshInterval = '60s';
const integrationAutoRefreshIntervals = Object.freeze({
  off: 0,
  '15s': 15000,
  '30s': 30000,
  '60s': 60000,
  '5m': 300000,
});
let integrationRefreshTimeoutId = 0;
let integrationRefreshInFlight = false;
let integrationRefreshCountdownIntervalId = 0;
let integrationNextRefreshAt = 0;
let integrationRefreshDurationMs = 0;

refreshProjects.addEventListener('click', async () => {
  await performIntegrationRefresh('manual');
});
integrationAutoRefreshInterval.addEventListener('change', () => {
  persistIntegrationAutoRefreshInterval(integrationAutoRefreshInterval.value);
  scheduleIntegrationAutoRefresh();
});
projectSelector.addEventListener('change', async (e) => {
  await selectProject(e.target.value);
});
projectGroupFilter.addEventListener('change', async () => {
  filterAndRenderProjects(projectSearchInput.value);
  await ensureSelectedProjectIsVisible();
});
projectSearchInput.addEventListener('input', (e) => {
  filterAndRenderProjects(e.target.value);
});
integrationBranchFilter.addEventListener('change', async () => {
  await loadIntegrationScreen(selectedProjectId, { preferredRunId: null });
});
integrationStatusFilter.addEventListener('change', async () => {
  await loadIntegrationRuns(selectedProjectId);
});
integrationReload.addEventListener('click', async () => {
  await runWithButtonBusy(integrationReload, 'Reload', 'Reloading...', async () => {
    await loadIntegrationScreen(selectedProjectId, { preferredRunId: selectedIntegrationRunId });
  });
});
openIntegrationHeatmap.addEventListener('click', async () => {
  const isOpen = integrationHeatmapOverlay.classList.contains('open');
  toggleIntegrationHeatmapOverlay(!isOpen);
  if (!isOpen) {
    await loadHeatmap();
  }
});
closeIntegrationHeatmap.addEventListener('click', () => toggleIntegrationHeatmapOverlay(false));
heatmapBranchFilter.addEventListener('change', async () => {
  await loadHeatmap();
});
heatmapStatusFilter.addEventListener('change', async () => {
  await loadHeatmap();
});
heatmapReload.addEventListener('click', async () => {
  await runWithButtonBusy(heatmapReload, 'Reload', 'Reloading...', async () => {
    await loadHeatmap();
  });
});
toggleSidebar.addEventListener('click', () => {
  const shouldCollapse = !appShell.classList.contains('sidebar-collapsed');
  setSidebarCollapsed(shouldCollapse);
});
document.addEventListener('visibilitychange', () => {
  if (document.hidden) {
    clearIntegrationAutoRefresh();
    updateIntegrationAutoRefreshStatus();
    return;
  }
  scheduleIntegrationAutoRefresh();
});

initializeSidebarState();
initializeIntegrationAutoRefreshControl();

(async () => {
  await performIntegrationRefresh('initial');
  if (getQueryParam('heatmap') === 'open') {
    toggleIntegrationHeatmapOverlay(true);
    await loadHeatmap();
  }
})();

function getQueryParam(name) {
  const params = new URLSearchParams(window.location.search);
  return params.get(name);
}

function initializeIntegrationAutoRefreshControl() {
  const persisted = window.localStorage.getItem(integrationAutoRefreshStorageKey);
  const nextValue = Object.prototype.hasOwnProperty.call(integrationAutoRefreshIntervals, persisted)
    ? persisted
    : integrationDefaultAutoRefreshInterval;
  integrationAutoRefreshInterval.value = nextValue;
  updateIntegrationAutoRefreshStatus();
}

function getIntegrationAutoRefreshIntervalValue() {
  const selectedValue = integrationAutoRefreshInterval.value;
  return Object.prototype.hasOwnProperty.call(integrationAutoRefreshIntervals, selectedValue)
    ? selectedValue
    : integrationDefaultAutoRefreshInterval;
}

function getIntegrationAutoRefreshIntervalMs() {
  return integrationAutoRefreshIntervals[getIntegrationAutoRefreshIntervalValue()] || 0;
}

function persistIntegrationAutoRefreshInterval(value) {
  const nextValue = Object.prototype.hasOwnProperty.call(integrationAutoRefreshIntervals, value)
    ? value
    : integrationDefaultAutoRefreshInterval;
  window.localStorage.setItem(integrationAutoRefreshStorageKey, nextValue);
}

function clearIntegrationAutoRefresh() {
  if (!integrationRefreshTimeoutId) return;
  window.clearTimeout(integrationRefreshTimeoutId);
  integrationRefreshTimeoutId = 0;
}

function setIntegrationAutoRefreshProgress(progressRatio) {
  if (!integrationAutoRefreshProgressBar) return;
  const safeRatio = Math.max(0, Math.min(1, progressRatio));
  integrationAutoRefreshProgressBar.style.transform = `scaleX(${safeRatio})`;
}

function clearIntegrationCountdownTicker() {
  if (!integrationRefreshCountdownIntervalId) return;
  window.clearInterval(integrationRefreshCountdownIntervalId);
  integrationRefreshCountdownIntervalId = 0;
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

function updateIntegrationAutoRefreshStatus() {
  if (!integrationAutoRefreshStatus) return;

  const intervalLabel = getIntegrationAutoRefreshIntervalValue();
  if (intervalLabel === 'off') {
    integrationAutoRefreshStatus.textContent = 'Auto refresh is off.';
    setIntegrationAutoRefreshProgress(0);
    return;
  }

  if (document.hidden) {
    integrationAutoRefreshStatus.textContent = `Paused (${intervalLabel}) while tab is hidden.`;
    setIntegrationAutoRefreshProgress(0);
    return;
  }

  if (integrationRefreshInFlight) {
    integrationAutoRefreshStatus.textContent = `Refreshing now (${intervalLabel}).`;
    setIntegrationAutoRefreshProgress(0);
    return;
  }

  if (!integrationNextRefreshAt) {
    integrationAutoRefreshStatus.textContent = `Scheduled every ${intervalLabel}.`;
    setIntegrationAutoRefreshProgress(0);
    return;
  }

  const remainingMs = integrationNextRefreshAt - Date.now();
  integrationAutoRefreshStatus.textContent = `Next refresh in ${formatRemainingTime(remainingMs)} (${intervalLabel}).`;
  const denominator = integrationRefreshDurationMs || getIntegrationAutoRefreshIntervalMs() || 1;
  setIntegrationAutoRefreshProgress(remainingMs / denominator);
}

function scheduleIntegrationAutoRefresh() {
  clearIntegrationAutoRefresh();
  clearIntegrationCountdownTicker();
  integrationNextRefreshAt = 0;
  integrationRefreshDurationMs = 0;

  const intervalMs = getIntegrationAutoRefreshIntervalMs();
  if (!intervalMs || document.hidden) {
    updateIntegrationAutoRefreshStatus();
    return;
  }

  integrationRefreshDurationMs = intervalMs;
  integrationNextRefreshAt = Date.now() + intervalMs;
  setIntegrationAutoRefreshProgress(1);
  updateIntegrationAutoRefreshStatus();
  integrationRefreshCountdownIntervalId = window.setInterval(() => {
    updateIntegrationAutoRefreshStatus();
  }, 200);

  integrationRefreshTimeoutId = window.setTimeout(async () => {
    if (document.hidden || integrationRefreshInFlight) {
      scheduleIntegrationAutoRefresh();
      return;
    }

    await performIntegrationRefresh('auto');
  }, intervalMs);
}

function setIntegrationRefreshButtonBusy(busy) {
  refreshProjects.disabled = busy;
  refreshProjects.textContent = busy ? 'Refreshing...' : 'Refresh';
}

async function runWithButtonBusy(button, idleText, busyText, action) {
  if (integrationRefreshInFlight) {
    return false;
  }

  integrationRefreshInFlight = true;
  clearIntegrationAutoRefresh();
  clearIntegrationCountdownTicker();
  integrationNextRefreshAt = 0;
  integrationRefreshDurationMs = 0;
  updateIntegrationAutoRefreshStatus();
  button.disabled = true;
  button.textContent = busyText;
  try {
    await action();
    return true;
  } finally {
    button.disabled = false;
    button.textContent = idleText;
    integrationRefreshInFlight = false;
    scheduleIntegrationAutoRefresh();
  }
}

async function performIntegrationRefresh(source = 'manual') {
  if (integrationRefreshInFlight) {
    return false;
  }

  integrationRefreshInFlight = true;
  clearIntegrationAutoRefresh();
  clearIntegrationCountdownTicker();
  integrationNextRefreshAt = 0;
  integrationRefreshDurationMs = 0;
  updateIntegrationAutoRefreshStatus();

  const heatmapWasOpen = integrationHeatmapOverlay.classList.contains('open');

  if (source === 'manual') {
    setIntegrationRefreshButtonBusy(true);
  }

  try {
    await loadProjects();

    if (heatmapWasOpen) {
      await loadHeatmap();
    }

    return true;
  } finally {
    if (source === 'manual') {
      setIntegrationRefreshButtonBusy(false);
    }

    integrationRefreshInFlight = false;
    scheduleIntegrationAutoRefresh();
  }
}

function toggleIntegrationHeatmapOverlay(open) {
  integrationHeatmapOverlay.classList.toggle('open', open);
  integrationHeatmapOverlay.setAttribute('aria-hidden', String(!open));
}

function initializeSidebarState() {
  const persisted = window.localStorage.getItem(sidebarCollapsedKey);
  setSidebarCollapsed(persisted === 'true');
}

function setSidebarCollapsed(collapsed) {
  appShell.classList.toggle('sidebar-collapsed', collapsed);
  toggleSidebar.textContent = collapsed ? '▸' : '◂';
  toggleSidebar.setAttribute('aria-label', collapsed ? 'Expand sidebar' : 'Collapse sidebar');
  toggleSidebar.setAttribute('title', collapsed ? 'Expand sidebar' : 'Collapse sidebar');
  toggleSidebar.setAttribute('aria-expanded', String(!collapsed));
  window.localStorage.setItem(sidebarCollapsedKey, String(collapsed));
}

async function loadProjects() {
  try {
    const pageSize = 100;
    let page = 1;
    let totalPages = 1;
    const items = [];

    while (page <= totalPages) {
      const res = await fetch(`/api/projects?page=${page}&pageSize=${pageSize}`);
      if (!res.ok) throw new Error(`failed to load projects (${res.status})`);
      const data = await res.json();
      items.push(...(data.items || []));
      totalPages = Math.max(1, data.pagination?.totalPages || 1);
      page += 1;
    }

    projects = items;
    renderProjectGroupFilter();
    filterAndRenderProjects(projectSearchInput.value);

    const nextSelectedProjectId = filteredProjects.some((project) => project.id === selectedProjectId)
      ? selectedProjectId
      : (filteredProjects[0]?.id || null);

    if (!nextSelectedProjectId) {
      selectedProjectId = null;
      selectedIntegrationRunId = null;
      if (items.length === 0) {
        integrationScreenProjectTitle.textContent = 'No projects found';
        integrationScreenProjectMeta.textContent = 'Upload integration runs to populate this view.';
        integrationRunChain.innerHTML = '<p class="muted">No integration runs found.</p>';
        integrationRunsBody.innerHTML = '<tr><td colspan="7" class="muted">No integration runs found.</td></tr>';
      } else {
        integrationScreenProjectTitle.textContent = 'No projects for current filter';
        integrationScreenProjectMeta.textContent = 'Adjust group and search filters to select a project.';
        integrationRunChain.innerHTML = '<p class="muted">No projects match current filters.</p>';
        integrationRunsBody.innerHTML = '<tr><td colspan="7" class="muted">No projects match current filters.</td></tr>';
      }
      integrationFailedSpecsBody.innerHTML = '<tr><td colspan="4" class="muted">No run selected.</td></tr>';
      integrationStatus.textContent = '-';
      integrationStatus.className = 'value';
      integrationPassRate.textContent = '-';
      integrationFailedSpecsCount.textContent = '-';
      integrationDelta.textContent = '-';
      integrationDuration.textContent = '-';
      renderProjectSelector();
    } else if (nextSelectedProjectId === selectedProjectId) {
      await selectProject(nextSelectedProjectId, { preferredRunId: selectedIntegrationRunId });
      renderProjectSelector();
    } else {
      await selectProject(nextSelectedProjectId);
      renderProjectSelector();
    }
  } catch (err) {
    integrationScreenProjectTitle.textContent = 'Failed to load projects';
    integrationScreenProjectMeta.textContent = err.message;
  }
}

function getProjectGroupValue(project) {
  const rawGroup = typeof project?.group === 'string' ? project.group.trim() : '';
  return rawGroup || ungroupedFilterValue;
}

function renderProjectGroupFilter() {
  const selectedValue = projectGroupFilter.value || allGroupsFilterValue;
  const groupValues = Array.from(new Set(projects.map((project) => getProjectGroupValue(project))));
  groupValues.sort((a, b) => {
    if (a === ungroupedFilterValue) return 1;
    if (b === ungroupedFilterValue) return -1;
    return a.localeCompare(b);
  });

  projectGroupFilter.innerHTML = '';

  const allOption = document.createElement('option');
  allOption.value = allGroupsFilterValue;
  allOption.textContent = 'All groups';
  projectGroupFilter.appendChild(allOption);

  for (const groupValue of groupValues) {
    const option = document.createElement('option');
    option.value = groupValue;
    option.textContent = groupValue === ungroupedFilterValue ? 'Ungrouped' : groupValue;
    projectGroupFilter.appendChild(option);
  }

  projectGroupFilter.value = [allGroupsFilterValue, ...groupValues].includes(selectedValue)
    ? selectedValue
    : allGroupsFilterValue;
}

function renderProjectSelector() {
  projectSelector.innerHTML = '';

  const emptyOption = document.createElement('option');
  emptyOption.value = '';
  emptyOption.textContent = 'Select a project...';
  projectSelector.appendChild(emptyOption);

  if (filteredProjects.length === 0) {
    const noResultsOption = document.createElement('option');
    noResultsOption.value = '';
    noResultsOption.textContent = 'No projects match current filters';
    noResultsOption.disabled = true;
    projectSelector.appendChild(noResultsOption);
  }

  for (const project of filteredProjects) {
    const option = document.createElement('option');
    option.value = project.id;
    option.textContent = `${project.name || project.projectKey} (${project.projectKey})`;
    projectSelector.appendChild(option);
  }

  projectSelector.value = selectedProjectId || '';
}

function filterAndRenderProjects(searchTerm) {
  const term = searchTerm.toLowerCase();
  const selectedGroup = projectGroupFilter.value || allGroupsFilterValue;
  filteredProjects = projects.filter((p) => {
    const groupMatches = selectedGroup === allGroupsFilterValue
      || getProjectGroupValue(p) === selectedGroup;
    if (!groupMatches) return false;
    if (!term) return true;

    const name = (p.name || '').toLowerCase();
    const key = (p.projectKey || '').toLowerCase();
    return name.includes(term) || key.includes(term);
  });

  renderProjectSelector();
}

async function ensureSelectedProjectIsVisible() {
  const selectedVisible = filteredProjects.some((project) => project.id === selectedProjectId);
  if (selectedVisible) {
    renderProjectSelector();
    return;
  }

  const nextProjectId = filteredProjects[0]?.id || null;
  if (!nextProjectId) {
    selectedProjectId = null;
    selectedIntegrationRunId = null;
    integrationScreenProjectTitle.textContent = 'No projects for current filter';
    integrationScreenProjectMeta.textContent = 'Adjust group and search filters to select a project.';
    integrationRunChain.innerHTML = '<p class="muted">No projects match current filters.</p>';
    integrationRunsBody.innerHTML = '<tr><td colspan="7" class="muted">No projects match current filters.</td></tr>';
    integrationFailedSpecsBody.innerHTML = '<tr><td colspan="4" class="muted">No run selected.</td></tr>';
    integrationStatus.textContent = '-';
    integrationStatus.className = 'value';
    integrationPassRate.textContent = '-';
    integrationFailedSpecsCount.textContent = '-';
    integrationDelta.textContent = '-';
    integrationDuration.textContent = '-';
    renderProjectSelector();
    return;
  }

  await selectProject(nextProjectId);
  renderProjectSelector();
}

function renderIntegrationBranchFilter(project, branches = []) {
  const selectedValue = integrationBranchFilter.value;
  integrationBranchFilter.innerHTML = '';

  const defaultBranch = project?.defaultBranch || 'main';
  const orderedBranches = Array.from(new Set([defaultBranch, ...branches.filter(Boolean)]));
  for (const branch of orderedBranches) {
    const option = document.createElement('option');
    option.value = branch;
    option.textContent = branch;
    integrationBranchFilter.appendChild(option);
  }

  integrationBranchFilter.value = orderedBranches.includes(selectedValue)
    ? selectedValue
    : (orderedBranches[0] || defaultBranch);
}

async function loadIntegrationBranches(projectId, defaultBranch) {
  try {
    const res = await fetch(`/api/projects/${projectId}/branches`);
    if (!res.ok) throw new Error(`failed to load branches (${res.status})`);
    const data = await res.json();
    const branches = Array.isArray(data.branches) ? data.branches.filter(Boolean) : [];
    return Array.from(new Set([defaultBranch, ...branches]));
  } catch (err) {
    return [defaultBranch];
  }
}

async function selectProject(projectId, options = {}) {
  const { preferredRunId = null } = options;

  if (!projectId) {
    selectedProjectId = null;
    selectedIntegrationRunId = null;
    integrationScreenProjectTitle.textContent = 'Select a project';
    integrationScreenProjectMeta.textContent = 'Choose a project from the left menu.';
    await loadIntegrationScreen(null);
    renderProjectSelector();
    return;
  }

  selectedProjectId = projectId;

  const project = projects.find((p) => p.id === projectId);
  integrationScreenProjectTitle.textContent = project?.name || project?.projectKey || 'Project';
  integrationScreenProjectMeta.textContent = `${project?.projectKey || ''} - default branch: ${project?.defaultBranch || 'main'}`;

  const defaultBranch = project?.defaultBranch || 'main';
  const branches = await loadIntegrationBranches(projectId, defaultBranch);
  renderIntegrationBranchFilter(project, branches);
  await loadIntegrationScreen(projectId, { preferredRunId });
}

async function loadIntegrationScreen(projectId, options = {}) {
  const { preferredRunId = selectedIntegrationRunId } = options;

  if (!projectId) {
    integrationRunChain.innerHTML = '<p class="muted">Select a project to view its run chain.</p>';
    integrationRunsBody.innerHTML = '<tr><td colspan="7" class="muted">Select a project first.</td></tr>';
    integrationFailedSpecsBody.innerHTML = '<tr><td colspan="4" class="muted">No run selected.</td></tr>';
    return;
  }

  await Promise.all([loadIntegrationLatestComparison(projectId), loadIntegrationRuns(projectId, preferredRunId)]);
}

async function loadIntegrationLatestComparison(projectId) {
  try {
    const url = new URL(`/api/projects/${projectId}/integration-test-runs/latest-comparison`, window.location.origin);
    if (integrationBranchFilter.value) {
      url.searchParams.set('branch', integrationBranchFilter.value);
    }

    const res = await fetch(url.toString());
    if (!res.ok) throw new Error(`failed to load integration comparison (${res.status})`);
    const data = await res.json();

    integrationStatus.textContent = (data.run?.status || '-').toUpperCase();
    integrationStatus.className = `value ${data.run?.status === 'passed' ? 'passed' : 'failed'}`;
    integrationPassRate.textContent = pct(data.run?.passRatePercent);
    integrationFailedSpecsCount.textContent = String(data.run?.failedSpecs ?? '-');
    integrationDelta.textContent = data.comparison?.deltaPercent == null ? '-' : signedPct(data.comparison.deltaPercent);
    integrationDuration.textContent = data.run?.durationMs == null ? '-' : `${Math.round(data.run.durationMs / 1000)}s`;
  } catch (err) {
    integrationStatus.textContent = 'ERROR';
    integrationStatus.className = 'value failed';
    integrationPassRate.textContent = '-';
    integrationFailedSpecsCount.textContent = '-';
    integrationDelta.textContent = '-';
    integrationDuration.textContent = '-';
  }
}

async function loadIntegrationRuns(projectId, preferredRunId = null) {
  integrationRunChain.innerHTML = '';
  integrationRunsBody.innerHTML = '';
  currentIntegrationRunItems = [];

  const retainedRunId = preferredRunId || selectedIntegrationRunId;

  try {
    const url = new URL(`/api/projects/${projectId}/integration-test-runs`, window.location.origin);
    url.searchParams.set('page', '1');
    url.searchParams.set('pageSize', '20');
    const project = projects.find((p) => p.id === projectId);
    const selectedBranch = integrationBranchFilter.value || project?.defaultBranch || 'main';
    url.searchParams.set('branch', selectedBranch);
    if (integrationStatusFilter.value) {
      url.searchParams.set('status', integrationStatusFilter.value);
    }

    const res = await fetch(url.toString());
    if (!res.ok) throw new Error(`failed to load integration runs (${res.status})`);
    const data = await res.json();
    const items = data.items || [];

    currentIntegrationRunItems = items;

    if (items.length === 0) {
      selectedIntegrationRunId = null;
      integrationRunChain.innerHTML = '<p class="muted">No integration runs found for current filters.</p>';
      integrationRunsBody.innerHTML = '<tr><td colspan="7" class="muted">No integration runs found.</td></tr>';
      integrationFailedSpecsBody.innerHTML = '<tr><td colspan="4" class="muted">No run selected.</td></tr>';
      return;
    }

    const nextSelectedRunId = retainedRunId && items.some((run) => run.id === retainedRunId)
      ? retainedRunId
      : items[0].id;

    selectedIntegrationRunId = nextSelectedRunId;
    renderIntegrationRunChain(items);

    for (const run of items) {
      const tr = document.createElement('tr');
      tr.dataset.runId = run.id;
      tr.innerHTML = `
        <td class="code">${run.id}</td>
        <td>${run.branch}</td>
        <td class="code">${run.commitSha}</td>
        <td class="${run.status === 'passed' ? 'up' : 'down'}">${run.status}</td>
        <td>${pct(run.passRatePercent)}</td>
        <td>${run.failedSpecs}</td>
        <td>${new Date(run.runTimestamp).toLocaleString()}</td>
      `;
      tr.addEventListener('click', async () => {
        selectedIntegrationRunId = run.id;
        highlightSelectedRunRow();
        renderIntegrationRunChain(items);
        await loadIntegrationRunDetails(projectId, run.id);
      });
      integrationRunsBody.appendChild(tr);
    }

    highlightSelectedRunRow();
    await loadIntegrationRunDetails(projectId, selectedIntegrationRunId);
  } catch (err) {
    selectedIntegrationRunId = null;
    integrationRunChain.innerHTML = `<p class="muted">${err.message}</p>`;
    integrationRunsBody.innerHTML = `<tr><td colspan="7" class="muted">${err.message}</td></tr>`;
    integrationFailedSpecsBody.innerHTML = '<tr><td colspan="4" class="muted">Failed to load selected run details.</td></tr>';
  }
}

function renderIntegrationRunChain(items) {
  if (!Array.isArray(items) || items.length === 0) {
    integrationRunChain.innerHTML = '<p class="muted">No integration runs found for current filters.</p>';
    return;
  }

  const track = document.createElement('div');
  track.className = 'integration-run-chain-track';

  items.forEach((run, index) => {
    const item = document.createElement('div');
    item.className = 'integration-chain-item';

    const button = document.createElement('button');
    button.type = 'button';
    button.className = `integration-chain-node ${run.status === 'passed' ? 'passed' : 'failed'}`;
    if (selectedIntegrationRunId === run.id) {
      button.classList.add('selected');
    }
    button.title = `${run.status.toUpperCase()} | ${formatDateTime(run.runTimestamp)} | ${pct(run.passRatePercent)}`;
    button.setAttribute('aria-label', `Run ${run.id}, ${run.status}, pass rate ${pct(run.passRatePercent)}`);
    button.addEventListener('click', async () => {
      selectedIntegrationRunId = run.id;
      highlightSelectedRunRow();
      renderIntegrationRunChain(items);
      await loadIntegrationRunDetails(selectedProjectId, run.id);
    });

    const label = document.createElement('p');
    label.className = 'integration-chain-label';
    label.textContent = `${shortCommit(run.commitSha)} · ${formatChainDate(run.runTimestamp)}`;

    item.appendChild(button);
    item.appendChild(label);
    track.appendChild(item);

    if (index < items.length - 1) {
      const connector = document.createElement('span');
      connector.className = 'integration-chain-connector';
      track.appendChild(connector);
    }
  });

  integrationRunChain.innerHTML = '';
  integrationRunChain.appendChild(track);
}

function formatChainDate(value) {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '-';

  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function highlightSelectedRunRow() {
  const rows = integrationRunsBody.querySelectorAll('tr[data-run-id]');
  for (const row of rows) {
    row.classList.toggle('selected-row', row.dataset.runId === selectedIntegrationRunId);
  }
}

async function loadIntegrationRunDetails(projectId, runId) {
  integrationFailedSpecsBody.innerHTML = '';
  try {
    const res = await fetch(`/api/projects/${projectId}/integration-test-runs/${runId}`);
    if (!res.ok) throw new Error(`failed to load integration run details (${res.status})`);
    const data = await res.json();
    const failedSpecs = data.failedSpecs || [];

    if (failedSpecs.length === 0) {
      integrationFailedSpecsBody.innerHTML = '<tr><td colspan="4" class="muted">No failed specs for this run.</td></tr>';
      return;
    }

    for (const failed of failedSpecs) {
      const tr = document.createElement('tr');
      tr.innerHTML = `
        <td class="code">${escapeHtml(failed.specPath || '-')}</td>
        <td>${escapeHtml(failed.failureMessage || '-')}</td>
        <td class="code">${escapeHtml(failed.file || '-')}</td>
        <td>${failed.line || '-'}</td>
      `;
      integrationFailedSpecsBody.appendChild(tr);
    }
  } catch (err) {
    integrationFailedSpecsBody.innerHTML = `<tr><td colspan="4" class="muted">${err.message}</td></tr>`;
  }
}

function pct(v) {
  if (v == null || Number.isNaN(v)) return '-';
  return `${Number(v).toFixed(2)}%`;
}

function signedPct(v) {
  const n = Number(v);
  if (Number.isNaN(n)) return '-';
  return `${n > 0 ? '+' : ''}${n.toFixed(2)}%`;
}

function shortCommit(commitSha) {
  if (!commitSha) return '-';
  return String(commitSha).slice(0, 7);
}

function formatDateTime(value) {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '-';
  return date.toLocaleString();
}

function escapeHtml(value) {
  return String(value)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');
}

async function loadHeatmap() {
  integrationHeatmap.innerHTML = '<p class="muted">Loading heatmap…</p>';
  try {
    const url = new URL('/api/integration-test-runs/heatmap', window.location.origin);
    url.searchParams.set('runsPerProject', '10');
    if (heatmapBranchFilter.value) url.searchParams.set('branch', heatmapBranchFilter.value);
    if (heatmapStatusFilter.value) url.searchParams.set('status', heatmapStatusFilter.value);

    const res = await fetch(url.toString());
    if (!res.ok) throw new Error(`heatmap request failed (${res.status})`);
    const data = await res.json();
    renderHeatmap(data.groups || []);
  } catch (err) {
    integrationHeatmap.innerHTML = `<p class="muted">${escapeHtml(err.message)}</p>`;
  }
}

function renderHeatmap(groups) {
  integrationHeatmap.innerHTML = '';

  if (groups.length === 0) {
    integrationHeatmap.innerHTML = '<p class="muted">No integration runs found.</p>';
    return;
  }

  for (const group of groups) {
    const groupEl = document.createElement('div');
    groupEl.className = 'integration-heatmap-group';

    const groupLabel = document.createElement('p');
    groupLabel.className = 'integration-heatmap-group-name';
    groupLabel.textContent = group.groupName || 'Ungrouped';
    groupEl.appendChild(groupLabel);

    for (const project of group.projects || []) {
      const rowEl = document.createElement('div');
      rowEl.className = 'integration-heatmap-project-row';
      const newestRun = Array.isArray(project.runs) && project.runs.length > 0 ? project.runs[0] : null;
      if (newestRun?.status === 'passed') {
        rowEl.classList.add('newest-passed');
      } else if (newestRun?.status === 'failed') {
        rowEl.classList.add('newest-failed');
      }

      const nameEl = document.createElement('span');
      nameEl.className = 'integration-heatmap-project-name';
      nameEl.textContent = project.projectName || project.projectKey;
      nameEl.title = project.projectKey;
      rowEl.appendChild(nameEl);

      const tilesEl = document.createElement('div');
      tilesEl.className = 'integration-heatmap-tiles';

      const runs = project.runs || [];
      runs.forEach((run, index) => {
        const tile = document.createElement('button');
        tile.type = 'button';
        tile.className = `integration-heatmap-tile ${run.status === 'passed' ? 'passed' : 'failed'}`;
        if (selectedProjectId === project.projectId && selectedIntegrationRunId === run.id) {
          tile.classList.add('selected');
        }
        tile.title = [
          project.projectName || project.projectKey,
          group.groupName ? `Group: ${group.groupName}` : null,
          `Branch: ${run.branch}`,
          `Commit: ${shortCommit(run.commitSha)}`,
          `${formatDateTime(run.runTimestamp)}`,
          `Status: ${run.status.toUpperCase()}`,
          `Pass rate: ${pct(run.passRatePercent)}`,
        ].filter(Boolean).join('\n');
        tile.setAttribute('aria-label', `${project.projectName || project.projectKey} — ${run.status} — ${pct(run.passRatePercent)}`);

        tile.addEventListener('click', async () => {
          if (selectedProjectId !== project.projectId) {
            // Switching project: select it (loads runs for that project)
            projectSelector.value = project.projectId;
            await selectProject(project.projectId, { preferredRunId: run.id });
            renderProjectSelector();
          } else {
            // Same project: synchronize run selection
            selectedIntegrationRunId = run.id;
            highlightSelectedRunRow();
            renderIntegrationRunChain(currentIntegrationRunItems);
            await loadIntegrationRunDetails(project.projectId, run.id);
          }
          // Re-render heatmap to update selected tile highlight
          renderHeatmap(groups);
        });

        tilesEl.appendChild(tile);

        if (index < runs.length - 1) {
          const arrow = document.createElement('span');
          arrow.className = 'integration-heatmap-arrow';
          arrow.textContent = '←';
          arrow.title = 'Oldest to newest';
          arrow.setAttribute('aria-hidden', 'true');
          tilesEl.appendChild(arrow);
        }
      });

      rowEl.appendChild(tilesEl);
      groupEl.appendChild(rowEl);
    }

    integrationHeatmap.appendChild(groupEl);
  }
}
