const selectedProjectName = document.getElementById('selectedProjectName');
const selectedProjectMeta = document.getElementById('selectedProjectMeta');
const packagesBody = document.getElementById('packagesBody');
const runsBody = document.getElementById('runsBody');
const currentCoverage = document.getElementById('currentCoverage');
const previousCoverage = document.getElementById('previousCoverage');
const deltaCoverage = document.getElementById('deltaCoverage');
const thresholdPercent = document.getElementById('thresholdPercent');
const thresholdStatus = document.getElementById('thresholdStatus');
const refreshProjects = document.getElementById('refreshProjects');
const openHeatmap = document.getElementById('openHeatmap');
const closeHeatmap = document.getElementById('closeHeatmap');
const heatmapOverlay = document.getElementById('heatmapOverlay');
const heatmapGrid = document.getElementById('heatmapGrid');
const projectSelector = document.getElementById('projectSelector');
const projectSearchInput = document.getElementById('projectSearchInput');
const compareCard = document.getElementById('compareCard');
const compareSummary = document.getElementById('compareSummary');
const compareCurrent = document.getElementById('compareCurrent');
const compareBaseline = document.getElementById('compareBaseline');
const compareRunType = document.getElementById('compareRunType');
const branchSelector = document.getElementById('branchSelector');

let projects = [];
let allProjects = [];
let filteredProjects = [];
let selectedProjectId = null;
let selectedBranch = '';
let availableBranches = [];
let heatmapItems = [];

refreshProjects.addEventListener('click', () => loadProjects());
openHeatmap.addEventListener('click', () => {
  const isOpen = heatmapOverlay.classList.contains('open');
  toggleHeatmapOverlay(!isOpen);
});
closeHeatmap.addEventListener('click', () => toggleHeatmapOverlay(false));
projectSelector.addEventListener('change', async (e) => {
  await selectProject(e.target.value);
});
projectSearchInput.addEventListener('input', (e) => {
  filterAndRenderProjects(e.target.value);
});
branchSelector.addEventListener('change', async (e) => {
  selectedBranch = e.target.value;
  await loadLatestComparison(selectedProjectId);
});

function toggleHeatmapOverlay(open) {
  heatmapOverlay.classList.toggle('open', open);
  heatmapOverlay.setAttribute('aria-hidden', String(!open));
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
    allProjects = items;
    filteredProjects = items;
    renderProjectSelector();

    if (!selectedProjectId && projects.length > 0) {
      await selectProject(projects[0].id);
     renderProjectSelector(); // Update dropdown to show selected project
     } else if (selectedProjectId) {
       await selectProject(selectedProjectId);
       renderProjectSelector(); // Update dropdown to show selected project
     }

    await loadHeatmap();
  } catch (err) {
    selectedProjectName.textContent = 'Failed to load projects';
    selectedProjectMeta.textContent = err.message;
    heatmapGrid.innerHTML = `<p class="muted">${err.message}</p>`;
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
    filteredProjects = allProjects;
  } else {
    filteredProjects = allProjects.filter(p => {
      const name = (p.name || '').toLowerCase();
      const key = (p.projectKey || '').toLowerCase();
      return name.includes(term) || key.includes(term);
    });
  }
  renderProjectSelector();
}

async function selectProject(projectId) {
  selectedProjectId = projectId;
  selectedBranch = '';
  projectSearchInput.value = '';
  renderHeatmap();

  const project = projects.find((p) => p.id === projectId) || allProjects.find((p) => p.id === projectId);
  selectedProjectName.textContent = project?.name || project?.projectKey || 'Project';
  selectedProjectMeta.textContent = `${project?.projectKey || ''} - default branch: ${project?.defaultBranch || 'main'} - threshold: ${pct(project?.globalThresholdPercent)}`;

  await Promise.all([loadBranches(projectId), loadRecentRuns(projectId)]);
  await loadLatestComparison(projectId);
}

async function loadHeatmap() {
  heatmapItems = [];
  renderHeatmapLoading();

  const sourceProjects = allProjects.length > 0 ? allProjects : projects;

  const items = await Promise.all(
    sourceProjects.map(async (project) => {
      try {
        const res = await fetch(`/api/projects/${project.id}/coverage-runs/latest-comparison`);
        if (!res.ok) throw new Error(`failed to load latest comparison (${res.status})`);
        const data = await res.json();
        return { project, comparison: data.comparison || null };
      } catch (err) {
        return { project, comparison: null, error: err.message };
      }
    }),
  );

  heatmapItems = items;
  renderHeatmap();
}

function renderHeatmapLoading() {
  heatmapGrid.innerHTML = '<p class="muted">Building heatmap...</p>';
}

function renderHeatmap() {
  heatmapGrid.innerHTML = '';

  if (projects.length === 0) {
    heatmapGrid.innerHTML = '<p class="muted">No projects found.</p>';
    return;
  }

  if (heatmapItems.length === 0) {
    renderHeatmapLoading();
    return;
  }

  const sorted = [...heatmapItems].sort((a, b) => {
    const ac = Number(a.comparison?.currentTotalCoveragePercent);
    const bc = Number(b.comparison?.currentTotalCoveragePercent);
    return (Number.isFinite(bc) ? bc : -1) - (Number.isFinite(ac) ? ac : -1);
  });

  for (const item of sorted) {
    const tile = buildHeatmapTile(item);
    heatmapGrid.appendChild(tile);
  }
}

function buildHeatmapTile(item) {
  const btn = document.createElement('button');
  const project = item.project;
  const name = project.name || project.projectKey;
  const current = Number(item.comparison?.currentTotalCoveragePercent);
  const delta = Number(item.comparison?.deltaPercent);
  const threshold = item.comparison?.thresholdStatus;
  const thresholdValue = Number(item.comparison?.thresholdPercent);

  const trendClass = heatTrendClass(current, delta, threshold);
  const size = tileSizeForCoverage(current);

  btn.className = `heat-tile ${trendClass} ${selectedProjectId === project.id ? 'selected' : ''}`;
  btn.style.setProperty('--col-span', String(size.col));
  btn.style.setProperty('--row-span', String(size.row));

  const deltaText = Number.isFinite(delta) ? signedPct(delta) : '-';
  btn.innerHTML = `
    <span class="heat-name">${name}</span>
    <span class="heat-value">${Number.isFinite(current) ? pct(current) : '-'}</span>
    <span class="heat-threshold">Threshold ${Number.isFinite(thresholdValue) ? pct(thresholdValue) : '-'}</span>
    <span class="heat-delta">${deltaText}</span>
  `;
  btn.title = `${name} | threshold=${threshold || 'unknown'} | delta=${deltaText}`;
  btn.addEventListener('click', async () => selectProject(project.id));
  return btn;
}

function heatTrendClass(current, delta, threshold) {
  if (!Number.isFinite(current)) return 'neutral';
  if ((Number.isFinite(delta) && delta < 0) || threshold === 'failed') return 'down';
  if ((Number.isFinite(delta) && delta >= 0) || threshold === 'passed') return 'up';
  return 'neutral';
}

function tileSizeForCoverage(current) {
  if (!Number.isFinite(current)) return { col: 3, row: 2 };
  if (current >= 90) return { col: 6, row: 4 };
  if (current >= 80) return { col: 5, row: 4 };
  if (current >= 70) return { col: 4, row: 3 };
  if (current >= 60) return { col: 4, row: 2 };
  return { col: 3, row: 2 };
}

async function loadBranches(projectId) {
  try {
    const res = await fetch(`/api/projects/${projectId}/branches`);
    if (!res.ok) throw new Error(`failed to load branches (${res.status})`);
    const data = await res.json();
    availableBranches = data.branches || [];
    renderBranchSelector();
  } catch (err) {
    console.error('Error loading branches:', err);
    branchSelector.innerHTML = '<option value="">Error loading branches</option>';
  }
}

function renderBranchSelector() {
  branchSelector.innerHTML = '';
  
  // Add empty option for latest run
  const emptyOption = document.createElement('option');
  emptyOption.value = '';
  emptyOption.textContent = 'Latest Run (All Branches)';
  branchSelector.appendChild(emptyOption);
  
  // Add each branch as an option
  for (const branch of availableBranches) {
    const option = document.createElement('option');
    option.value = branch;
    option.textContent = branch;
    branchSelector.appendChild(option);
  }
  
  branchSelector.value = selectedBranch;
}

async function loadLatestComparison(projectId) {
  packagesBody.innerHTML = '';
  try {
    const url = new URL(`/api/projects/${projectId}/coverage-runs/latest-comparison`, window.location.origin);
    if (selectedBranch) {
      url.searchParams.set('branch', selectedBranch);
    }
    const res = await fetch(url.toString());
    if (!res.ok) throw new Error(`failed to load latest comparison (${res.status})`);
    const data = await res.json();

    currentCoverage.textContent = pct(data.comparison.currentTotalCoveragePercent);
    previousCoverage.textContent = data.comparison.previousTotalCoveragePercent == null ? '-' : pct(data.comparison.previousTotalCoveragePercent);
    deltaCoverage.textContent = data.comparison.deltaPercent == null ? '-' : signedPct(data.comparison.deltaPercent);
    thresholdPercent.textContent = pct(data.comparison.thresholdPercent);

    thresholdStatus.textContent = data.comparison.thresholdStatus || '-';
    thresholdStatus.className = `value ${data.comparison.thresholdStatus === 'passed' ? 'passed' : 'failed'}`;

    const project = projects.find((p) => p.id === projectId) || allProjects.find((p) => p.id === projectId);
    const defaultBranch = project?.defaultBranch || 'main';
    const runBranch = data.run?.branch || '-';
    const runType = data.run?.triggerType || 'unknown';
    const baselineSource = data.comparison?.baselineSource || 'latest_default_branch';
    const isPRComparison = runType === 'pr' && runBranch !== defaultBranch;

    compareSummary.textContent = isPRComparison
      ? `PR branch ${runBranch} is being compared against default branch ${defaultBranch}.`
      : `Latest ${runType.toUpperCase()} run on ${runBranch} is compared against default branch ${defaultBranch}.`;
    compareCurrent.textContent = runBranch;
    compareBaseline.textContent = `${defaultBranch} (${baselineSource})`;
    compareRunType.textContent = runType.toUpperCase();
    compareCard.classList.toggle('pr-mode', isPRComparison);

    for (const p of data.packages || []) {
      const tr = document.createElement('tr');
      tr.innerHTML = `
        <td class="code">${p.importPath}</td>
        <td>${pct(p.currentCoveragePercent)}</td>
        <td>${p.previousCoveragePercent == null ? '-' : pct(p.previousCoveragePercent)}</td>
        <td>${p.deltaPercent == null ? '-' : signedPct(p.deltaPercent)}</td>
        <td class="${directionClass(p.direction)}">${p.direction || '-'}</td>
      `;
      packagesBody.appendChild(tr);
    }
  } catch (err) {
    currentCoverage.textContent = '-';
    previousCoverage.textContent = '-';
    deltaCoverage.textContent = '-';
    thresholdPercent.textContent = '-';
    thresholdStatus.textContent = 'error';
    thresholdStatus.className = 'value failed';
    compareSummary.textContent = err.message;
    compareCurrent.textContent = '-';
    compareBaseline.textContent = '-';
    compareRunType.textContent = '-';
    compareCard.classList.remove('pr-mode');

    const tr = document.createElement('tr');
    tr.innerHTML = `<td colspan="5" class="muted">${err.message}</td>`;
    packagesBody.appendChild(tr);
  }
}

async function loadRecentRuns(projectId) {
  runsBody.innerHTML = '';
  try {
    const res = await fetch(`/api/projects/${projectId}/coverage-runs?page=1&pageSize=20`);
    if (!res.ok) throw new Error(`failed to load runs (${res.status})`);
    const data = await res.json();

    for (const run of data.items || []) {
      const tr = document.createElement('tr');
      tr.innerHTML = `
        <td class="code">${run.id}</td>
        <td>${run.branch}</td>
        <td class="code">${run.commitSha}</td>
        <td>${pct(run.totalCoveragePercent)}</td>
        <td>${new Date(run.runTimestamp).toLocaleString()}</td>
      `;
      runsBody.appendChild(tr);
    }

    if ((data.items || []).length === 0) {
      const tr = document.createElement('tr');
      tr.innerHTML = '<td colspan="5" class="muted">No runs found.</td>';
      runsBody.appendChild(tr);
    }
  } catch (err) {
    const tr = document.createElement('tr');
    tr.innerHTML = `<td colspan="5" class="muted">${err.message}</td>`;
    runsBody.appendChild(tr);
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

function directionClass(direction) {
  if (direction === 'up') return 'up';
  if (direction === 'down') return 'down';
  return 'equal';
}

loadProjects();
