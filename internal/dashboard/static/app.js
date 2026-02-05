// Zeus Dashboard - Frontend Application

// 定数
const REFRESH_INTERVAL = 5000; // 5秒
const API_BASE = '/api';

// 状態
let isLoading = false;
let lastError = null;

// 初期化
document.addEventListener('DOMContentLoaded', init);

function init() {
  // Mermaid 初期化
  mermaid.initialize({
    startOnLoad: false,
    theme: 'default',
    securityLevel: 'loose',
    flowchart: {
      useMaxWidth: true,
      htmlLabels: true
    }
  });

  // 初回データ取得
  fetchData();

  // 自動更新開始
  setInterval(fetchData, REFRESH_INTERVAL);
}

// データ取得
async function fetchData() {
  if (isLoading) return;
  isLoading = true;

  try {
    const [status, tasks, graph, predict] = await Promise.all([
      fetchAPI('/status'),
      fetchAPI('/tasks'),
      fetchAPI('/graph'),
      fetchAPI('/predict')
    ]);

    hideError();
    updateConnectionStatus(true);
    renderDashboard({ status, tasks, graph, predict });
    updateLastUpdated();
    lastError = null;
  } catch (error) {
    console.error('データ取得エラー:', error);
    showError('データの取得に失敗しました: ' + error.message);
    updateConnectionStatus(false);
    lastError = error;
  } finally {
    isLoading = false;
  }
}

// API リクエスト
async function fetchAPI(endpoint) {
  const response = await fetch(API_BASE + endpoint);
  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.message || `HTTP ${response.status}`);
  }
  return response.json();
}

// ダッシュボード描画
function renderDashboard(data) {
  renderOverview(data.status);
  renderStats(data.status);
  renderTasks(data.tasks);
  renderGraph(data.graph);
  renderPrediction(data.predict);
}

// プロジェクト概要
function renderOverview(status) {
  document.getElementById('project-name').textContent = status.project.name || 'Unknown Project';
  document.getElementById('project-description').textContent = status.project.description || '';
  document.getElementById('project-start-date').textContent = status.project.start_date
    ? 'Started: ' + status.project.start_date
    : '';

  const healthBadge = document.getElementById('project-health');
  healthBadge.textContent = status.state.health;
  healthBadge.className = 'health-badge ' + status.state.health.toLowerCase();

  // 進捗率
  const total = status.state.summary.total_activities || 0;
  const completed = status.state.summary.completed || 0;
  const percent = total > 0 ? Math.round((completed / total) * 100) : 0;

  document.getElementById('progress-fill').style.width = percent + '%';
  document.getElementById('progress-text').textContent = percent + '%';
}

// Activity 統計
function renderStats(status) {
  document.getElementById('stat-total').textContent = status.state.summary.total_activities || 0;
  document.getElementById('stat-completed').textContent = status.state.summary.completed || 0;
  document.getElementById('stat-in-progress').textContent = status.state.summary.in_progress || 0;
  document.getElementById('stat-pending').textContent = status.state.summary.pending || 0;

  // 承認待ち
  const pendingApprovals = status.pending_approvals || 0;
  const pendingApprovalsEl = document.getElementById('pending-approvals');
  if (pendingApprovals > 0) {
    document.getElementById('pending-approvals-count').textContent = pendingApprovals;
    pendingApprovalsEl.style.display = 'flex';
  } else {
    pendingApprovalsEl.style.display = 'none';
  }
}

// タスク一覧
function renderTasks(data) {
  const tbody = document.getElementById('tasks-body');

  if (!data.tasks || data.tasks.length === 0) {
    tbody.innerHTML = '<tr><td colspan="5" class="loading">No tasks found</td></tr>';
    return;
  }

  tbody.innerHTML = data.tasks.map(task => `
    <tr>
      <td><code>${escapeHtml(task.id)}</code></td>
      <td>${escapeHtml(task.title)}</td>
      <td><span class="status-badge ${task.status}">${formatStatus(task.status)}</span></td>
      <td><span class="priority-badge ${task.priority || 'medium'}">${task.priority || '-'}</span></td>
      <td>${escapeHtml(task.assignee) || '-'}</td>
    </tr>
  `).join('');
}

// 依存関係グラフ
async function renderGraph(data) {
  // 統計情報
  document.getElementById('graph-nodes').textContent = data.stats.total_nodes || 0;
  document.getElementById('graph-deps').textContent = data.stats.with_dependencies || 0;
  document.getElementById('graph-isolated').textContent = data.stats.isolated_count || 0;

  // 循環依存の警告
  const cyclesWarning = document.getElementById('graph-cycles-warning');
  if (data.cycles && data.cycles.length > 0) {
    document.getElementById('graph-cycles').textContent = data.cycles.length;
    cyclesWarning.style.display = 'inline';
  } else {
    cyclesWarning.style.display = 'none';
  }

  // グラフ描画
  const graphContainer = document.getElementById('graph-container');
  const graphEmpty = document.getElementById('graph-empty');
  const mermaidEl = document.getElementById('mermaid-graph');

  if (data.stats.with_dependencies === 0) {
    graphContainer.style.display = 'none';
    graphEmpty.style.display = 'block';
    return;
  }

  graphContainer.style.display = 'block';
  graphEmpty.style.display = 'none';

  // Mermaid コードから ``` を除去
  let mermaidCode = data.mermaid || '';
  mermaidCode = mermaidCode.replace(/```mermaid\n?/g, '').replace(/```\n?/g, '').trim();

  if (!mermaidCode) {
    graphContainer.style.display = 'none';
    graphEmpty.style.display = 'block';
    return;
  }

  try {
    // 一意の ID を生成
    const id = 'mermaid-' + Date.now();
    const { svg } = await mermaid.render(id, mermaidCode);
    mermaidEl.innerHTML = svg;
  } catch (error) {
    console.error('Mermaid レンダリングエラー:', error);
    mermaidEl.innerHTML = '<pre>' + escapeHtml(mermaidCode) + '</pre>';
  }
}

// 予測分析
function renderPrediction(data) {
  // 完了予測
  if (data.completion) {
    document.getElementById('pred-date').textContent = data.completion.estimated_date || 'N/A';
    document.getElementById('pred-remaining').textContent = data.completion.remaining_tasks || '-';
    document.getElementById('pred-confidence').textContent =
      data.completion.confidence_level ? data.completion.confidence_level + '%' : '-';
    document.getElementById('pred-margin').textContent =
      data.completion.margin_days ? '+/- ' + data.completion.margin_days + ' days' : '-';
  }

  // リスク分析
  if (data.risk) {
    const riskLevel = document.getElementById('pred-risk-level');
    riskLevel.textContent = data.risk.overall_level || '-';
    riskLevel.className = 'prediction-value risk-badge ' + (data.risk.overall_level || '').toLowerCase();

    const factorsEl = document.getElementById('risk-factors');
    if (data.risk.factors && data.risk.factors.length > 0) {
      factorsEl.innerHTML = data.risk.factors.map(f => `
        <div class="risk-factor">
          <strong>${escapeHtml(f.name)}</strong>: ${escapeHtml(f.description)}
        </div>
      `).join('');
    } else {
      factorsEl.innerHTML = '<span class="no-risk">No risk factors detected</span>';
    }
  }

  // ベロシティ
  if (data.velocity) {
    document.getElementById('pred-velocity').textContent =
      data.velocity.weekly_average ? data.velocity.weekly_average.toFixed(1) + ' tasks/week' : '-';
    document.getElementById('vel-7days').textContent = data.velocity.last_7_days || '0';
    document.getElementById('vel-14days').textContent = data.velocity.last_14_days || '0';

    const trendEl = document.getElementById('vel-trend');
    trendEl.textContent = data.velocity.trend || '-';
    trendEl.className = 'detail-value trend-badge ' + (data.velocity.trend || '').toLowerCase();
  }
}

// ユーティリティ関数

function formatStatus(status) {
  const statusMap = {
    'pending': 'Pending',
    'in_progress': 'In Progress',
    'completed': 'Completed',
    'blocked': 'Blocked'
  };
  return statusMap[status] || status;
}

function escapeHtml(text) {
  if (!text) return '';
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

function updateLastUpdated() {
  const now = new Date();
  document.getElementById('last-updated').textContent =
    'Last updated: ' + now.toLocaleTimeString();
}

function updateConnectionStatus(connected) {
  const indicator = document.getElementById('connection-status');
  if (connected) {
    indicator.classList.remove('error');
  } else {
    indicator.classList.add('error');
  }
}

function showError(message) {
  const banner = document.getElementById('error-banner');
  banner.textContent = message;
  banner.style.display = 'block';
}

function hideError() {
  document.getElementById('error-banner').style.display = 'none';
}
