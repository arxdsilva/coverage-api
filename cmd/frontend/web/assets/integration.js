const refreshProjects = document.getElementById('refreshProjects');
const projectSelector = document.getElementById('projectSelector');
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
const sidebarCollapsedKey = 'opencoverage.sidebarCollapsed.integration';

refreshProjects.addEventListener('click', () => loadProjects());
projectSelector.addEventListener('change', async (e) => {
  await selectProject(e.target.value);
});
projectSearchInput.addEventListener('input', (e) => {
  filterAndRenderProjects(e.target.value);
});
integrationBranchFilter.addEventListener('change', async () => {
  await loadIntegrationRuns(selectedProjectId);
});
integrationStatusFilter.addEventListener('change', async () => {
  await loadIntegrationRuns(selectedProjectId);
});
integrationReload.addEventListener('click', async () => {
  await loadIntegrationScreen(selectedProjectId);
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
  await loadHeatmap();
});
toggleSidebar.addEventListener('click', () => {
  const shouldCollapse = !appShell.classList.contains('sidebar-collapsed');
  setSidebarCollapsed(shouldCollapse);
});

initializeSidebarState();

(async () => {
  await loadProjects();
  if (getQueryParam('heatmap') === 'open') {
    toggleIntegrationHeatmapOverlay(true);
    await loadHeatmap();
  }
})();

function getQueryParam(name) {
  const params = new URLSearchParams(window.location.search);
  return params.get(name);
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
    filteredProjects = items;
    renderProjectSelector();

    if (!selectedProjectId && projects.length > 0) {
      await selectProject(projects[0].id);
      renderProjectSelector();
    } else if (selectedProjectId) {
      await selectProject(selectedProjectId);
      renderProjectSelector();
    }
  } catch (err) {
    integrationScreenProjectTitle.textContent = 'Failed to load projects';
    integrationScreenProjectMeta.textContent = err.message;
  }
}

function renderProjectSelector() {
  projectSelector.innerHTML = '';

  const emptyOption = document.createElement('option');
  emptyOption.value = '';
  emptyOption.textContent = 'Select a project...';
  projectSelector.appendChild(emptyOption);

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
  if (!term) {
    filteredProjects = projects;
  } else {
    filteredProjects = projects.filter((p) => {
      const name = (p.name || '').toLowerCase();
      const key = (p.projectKey || '').toLowerCase();
      return name.includes(term) || key.includes(term);
    });
  }
  renderProjectSelector();
}

function renderIntegrationBranchFilter(project) {
  const selectedValue = integrationBranchFilter.value;
  integrationBranchFilter.innerHTML = '<option value="">All branches</option>';

  const defaultBranch = project?.defaultBranch || 'main';
  const branches = [defaultBranch, 'develop'];
  for (const branch of branches) {
    const option = document.createElement('option');
    option.value = branch;
    option.textContent = branch;
    integrationBranchFilter.appendChild(option);
  }

  integrationBranchFilter.value = branches.includes(selectedValue) ? selectedValue : '';
}

async function selectProject(projectId) {
  selectedProjectId = projectId;
  projectSearchInput.value = '';

  const project = projects.find((p) => p.id === projectId);
  integrationScreenProjectTitle.textContent = project?.name || project?.projectKey || 'Project';
  integrationScreenProjectMeta.textContent = `${project?.projectKey || ''} - default branch: ${project?.defaultBranch || 'main'}`;

  renderIntegrationBranchFilter(project);
  await loadIntegrationScreen(projectId);
}

async function loadIntegrationScreen(projectId) {
  if (!projectId) {
    integrationRunChain.innerHTML = '<p class="muted">Select a project to view its run chain.</p>';
    integrationRunsBody.innerHTML = '<tr><td colspan="7" class="muted">Select a project first.</td></tr>';
    integrationFailedSpecsBody.innerHTML = '<tr><td colspan="4" class="muted">No run selected.</td></tr>';
    return;
  }

  await Promise.all([loadIntegrationLatestComparison(projectId), loadIntegrationRuns(projectId)]);
}

async function loadIntegrationLatestComparison(projectId) {
  try {
    const res = await fetch(`/api/projects/${projectId}/integration-test-runs/latest-comparison`);
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

async function loadIntegrationRuns(projectId) {
  integrationRunChain.innerHTML = '';
  integrationRunsBody.innerHTML = '';
  selectedIntegrationRunId = null;
  currentIntegrationRunItems = [];

  try {
    const url = new URL(`/api/projects/${projectId}/integration-test-runs`, window.location.origin);
    url.searchParams.set('page', '1');
    url.searchParams.set('pageSize', '20');
    if (integrationBranchFilter.value) {
      url.searchParams.set('branch', integrationBranchFilter.value);
    }
    if (integrationStatusFilter.value) {
      url.searchParams.set('status', integrationStatusFilter.value);
    }

    const res = await fetch(url.toString());
    if (!res.ok) throw new Error(`failed to load integration runs (${res.status})`);
    const data = await res.json();
    const items = data.items || [];

    currentIntegrationRunItems = items;
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

    if (items.length === 0) {
      integrationRunsBody.innerHTML = '<tr><td colspan="7" class="muted">No integration runs found.</td></tr>';
      integrationFailedSpecsBody.innerHTML = '<tr><td colspan="4" class="muted">No run selected.</td></tr>';
      return;
    }

    selectedIntegrationRunId = items[0].id;
    highlightSelectedRunRow();
    renderIntegrationRunChain(items);
    await loadIntegrationRunDetails(projectId, selectedIntegrationRunId);
  } catch (err) {
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
    label.textContent = shortCommit(run.commitSha);

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

      const nameEl = document.createElement('span');
      nameEl.className = 'integration-heatmap-project-name';
      nameEl.textContent = project.projectName || project.projectKey;
      nameEl.title = project.projectKey;
      rowEl.appendChild(nameEl);

      const tilesEl = document.createElement('div');
      tilesEl.className = 'integration-heatmap-tiles';

      for (const run of project.runs || []) {
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
            await selectProject(project.projectId);
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
      }

      rowEl.appendChild(tilesEl);
      groupEl.appendChild(rowEl);
    }

    integrationHeatmap.appendChild(groupEl);
  }
}
