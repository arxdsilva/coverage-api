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
const integrationRunsBody = document.getElementById('integrationRunsBody');
const integrationFailedSpecsBody = document.getElementById('integrationFailedSpecsBody');
const appShell = document.getElementById('appShell');
const toggleSidebar = document.getElementById('toggleSidebar');

let projects = [];
let filteredProjects = [];
let selectedProjectId = null;
let selectedIntegrationRunId = null;
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
toggleSidebar.addEventListener('click', () => {
  const shouldCollapse = !appShell.classList.contains('sidebar-collapsed');
  setSidebarCollapsed(shouldCollapse);
});

initializeSidebarState();
loadProjects();

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
  integrationRunsBody.innerHTML = '';
  selectedIntegrationRunId = null;

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

    for (const run of items) {
      const tr = document.createElement('tr');
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
    await loadIntegrationRunDetails(projectId, selectedIntegrationRunId);
  } catch (err) {
    integrationRunsBody.innerHTML = `<tr><td colspan="7" class="muted">${err.message}</td></tr>`;
    integrationFailedSpecsBody.innerHTML = '<tr><td colspan="4" class="muted">Failed to load selected run details.</td></tr>';
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

function escapeHtml(value) {
  return String(value)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');
}
