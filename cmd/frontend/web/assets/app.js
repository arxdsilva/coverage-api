const projectList = document.getElementById('projectList');
const selectedProjectName = document.getElementById('selectedProjectName');
const selectedProjectMeta = document.getElementById('selectedProjectMeta');
const packagesBody = document.getElementById('packagesBody');
const runsBody = document.getElementById('runsBody');
const currentCoverage = document.getElementById('currentCoverage');
const previousCoverage = document.getElementById('previousCoverage');
const deltaCoverage = document.getElementById('deltaCoverage');
const thresholdStatus = document.getElementById('thresholdStatus');
const refreshProjects = document.getElementById('refreshProjects');

let projects = [];
let selectedProjectId = null;

refreshProjects.addEventListener('click', () => loadProjects());

async function loadProjects() {
  try {
    const res = await fetch('/api/projects');
    if (!res.ok) throw new Error(`failed to load projects (${res.status})`);
    const data = await res.json();
    projects = data.items || [];
    renderProjectList();

    if (!selectedProjectId && projects.length > 0) {
      selectProject(projects[0].id);
    } else if (selectedProjectId) {
      await selectProject(selectedProjectId);
    }
  } catch (err) {
    selectedProjectName.textContent = 'Failed to load projects';
    selectedProjectMeta.textContent = err.message;
  }
}

function renderProjectList() {
  projectList.innerHTML = '';

  if (projects.length === 0) {
    const li = document.createElement('li');
    li.textContent = 'No projects found.';
    li.className = 'muted';
    projectList.appendChild(li);
    return;
  }

  for (const project of projects) {
    const li = document.createElement('li');
    const btn = document.createElement('button');
    btn.className = selectedProjectId === project.id ? 'active' : '';
    btn.innerHTML = `<strong>${project.name || project.projectKey}</strong><small>${project.id}</small>`;
    btn.addEventListener('click', () => selectProject(project.id));
    li.appendChild(btn);
    projectList.appendChild(li);
  }
}

async function selectProject(projectId) {
  selectedProjectId = projectId;
  renderProjectList();

  const project = projects.find((p) => p.id === projectId);
  selectedProjectName.textContent = project?.name || project?.projectKey || 'Project';
  selectedProjectMeta.textContent = `${project?.projectKey || ''} - default branch: ${project?.defaultBranch || 'main'}`;

  await Promise.all([loadLatestComparison(projectId), loadRecentRuns(projectId)]);
}

async function loadLatestComparison(projectId) {
  packagesBody.innerHTML = '';
  try {
    const res = await fetch(`/api/projects/${projectId}/coverage-runs/latest-comparison`);
    if (!res.ok) throw new Error(`failed to load latest comparison (${res.status})`);
    const data = await res.json();

    currentCoverage.textContent = pct(data.comparison.currentTotalCoveragePercent);
    previousCoverage.textContent = data.comparison.previousTotalCoveragePercent == null ? '-' : pct(data.comparison.previousTotalCoveragePercent);
    deltaCoverage.textContent = data.comparison.deltaPercent == null ? '-' : signedPct(data.comparison.deltaPercent);

    thresholdStatus.textContent = data.comparison.thresholdStatus || '-';
    thresholdStatus.className = `value ${data.comparison.thresholdStatus === 'passed' ? 'passed' : 'failed'}`;

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
    thresholdStatus.textContent = 'error';
    thresholdStatus.className = 'value failed';

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
