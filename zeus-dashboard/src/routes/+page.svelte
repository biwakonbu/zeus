<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { OverviewPanel, StatsPanel, TasksPanel, PredictionPanel, GraphPanel } from '$lib/components/panels';
	import { FactorioViewer } from '$lib/viewer';
	import { refreshAllData } from '$lib/stores';
	import { setConnected, setDisconnected, setConnecting } from '$lib/stores/connection';
	import { connectSSE, disconnectSSE, getSSEClient } from '$lib/api/sse';
	import { tasks } from '$lib/stores/tasks';

	// ビューモード: 'viewer' (新しいFactorio風) または 'classic' (従来のパネル)
	let viewMode: 'viewer' | 'classic' = $state('viewer');

	let useSSE = $state(true);
	let pollingInterval: ReturnType<typeof setInterval> | null = null;

	// 選択中のタスク
	let selectedTaskId: string | null = $state(null);

	onMount(() => {
		// 初期データを読み込み
		setConnecting();
		refreshAllData()
			.then(() => {
				setConnected();

				// SSE 接続を試行
				if (useSSE) {
					try {
						connectSSE();
					} catch {
						// SSE が利用できない場合はポーリングにフォールバック
						console.log('SSE not available, falling back to polling');
						startPolling();
					}
				} else {
					startPolling();
				}
			})
			.catch(() => {
				setDisconnected();
				startPolling();
			});
	});

	onDestroy(() => {
		if (useSSE) {
			disconnectSSE();
		}
		stopPolling();
	});

	function startPolling() {
		if (pollingInterval) return;

		pollingInterval = setInterval(() => {
			refreshAllData().catch(() => {
				// エラー時は接続状態を更新
			});
		}, 5000);
	}

	function stopPolling() {
		if (pollingInterval) {
			clearInterval(pollingInterval);
			pollingInterval = null;
		}
	}

	// SSE/Polling 切り替え
	function toggleMode() {
		useSSE = !useSSE;

		if (useSSE) {
			stopPolling();
			connectSSE();
		} else {
			disconnectSSE();
			startPolling();
		}
	}

	// ビューモード切り替え
	function switchView(mode: 'viewer' | 'classic') {
		viewMode = mode;
	}

	// タスク選択ハンドラ
	function handleTaskSelect(taskId: string | null) {
		selectedTaskId = taskId;
	}

	// タスクホバーハンドラ
	function handleTaskHover(taskId: string | null) {
		// 必要に応じてツールチップ表示などを追加
	}
</script>

<!-- ビュー切り替えタブ -->
<div class="view-tabs">
	<button
		class="view-tab"
		class:active={viewMode === 'viewer'}
		onclick={() => switchView('viewer')}
	>
		<span class="tab-icon">⬡</span>
		VIEWER
	</button>
	<button
		class="view-tab"
		class:active={viewMode === 'classic'}
		onclick={() => switchView('classic')}
	>
		<span class="tab-icon">▦</span>
		CLASSIC
	</button>
</div>

{#if viewMode === 'viewer'}
	<!-- Factorio風ビューワー -->
	<div class="viewer-container">
		<FactorioViewer
			tasks={$tasks}
			selectedTaskId={selectedTaskId}
			onTaskSelect={handleTaskSelect}
			onTaskHover={handleTaskHover}
		/>
	</div>

	<!-- 選択タスク詳細パネル -->
	{#if selectedTaskId}
		{@const selectedTask = $tasks.find(t => t.id === selectedTaskId)}
		{#if selectedTask}
			<div class="task-detail-panel">
				<div class="panel-header">
					<h3 class="panel-title">TASK DETAIL</h3>
					<button class="close-btn" onclick={() => selectedTaskId = null}>×</button>
				</div>
				<div class="task-detail-content">
					<div class="detail-row">
						<span class="detail-label">ID</span>
						<span class="detail-value">{selectedTask.id}</span>
					</div>
					<div class="detail-row">
						<span class="detail-label">Title</span>
						<span class="detail-value">{selectedTask.title}</span>
					</div>
					<div class="detail-row">
						<span class="detail-label">Status</span>
						<span class="detail-value status-{selectedTask.status}">{selectedTask.status}</span>
					</div>
					<div class="detail-row">
						<span class="detail-label">Priority</span>
						<span class="detail-value priority-{selectedTask.priority}">{selectedTask.priority}</span>
					</div>
					<div class="detail-row">
						<span class="detail-label">Assignee</span>
						<span class="detail-value">{selectedTask.assignee || 'Unassigned'}</span>
					</div>
					{#if selectedTask.dependencies.length > 0}
						<div class="detail-row">
							<span class="detail-label">Dependencies</span>
							<span class="detail-value">{selectedTask.dependencies.length} tasks</span>
						</div>
					{/if}
				</div>
			</div>
		{/if}
	{/if}
{:else}
	<!-- 従来のパネルビュー -->
	<div class="dashboard">
		<div class="dashboard-row">
			<div class="dashboard-col-2">
				<OverviewPanel />
			</div>
			<div class="dashboard-col-2">
				<StatsPanel />
			</div>
		</div>

		<div class="dashboard-row">
			<div class="dashboard-col-full">
				<TasksPanel />
			</div>
		</div>

		<div class="dashboard-row">
			<div class="dashboard-col-2">
				<GraphPanel />
			</div>
			<div class="dashboard-col-2">
				<PredictionPanel />
			</div>
		</div>
	</div>
{/if}

<style>
	/* ビュータブ */
	.view-tabs {
		display: flex;
		gap: var(--spacing-xs);
		margin-bottom: var(--spacing-lg);
	}

	.view-tab {
		display: flex;
		align-items: center;
		gap: var(--spacing-xs);
		padding: var(--spacing-sm) var(--spacing-md);
		font-family: var(--font-family);
		font-size: var(--font-size-sm);
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-secondary);
		background-color: var(--bg-secondary);
		border: 2px solid var(--border-metal);
		border-radius: var(--border-radius-sm);
		cursor: pointer;
		transition: all var(--transition-fast);
	}

	.view-tab:hover {
		background-color: var(--bg-hover);
		border-color: var(--border-highlight);
	}

	.view-tab.active {
		background-color: var(--accent-primary);
		border-color: var(--accent-primary);
		color: var(--bg-primary);
	}

	.tab-icon {
		font-size: 14px;
	}

	/* ビューワーコンテナ */
	.viewer-container {
		height: calc(100vh - 200px);
		min-height: 600px;
	}

	/* タスク詳細パネル */
	.task-detail-panel {
		position: fixed;
		bottom: var(--spacing-xl);
		right: var(--spacing-xl);
		width: 320px;
		background-color: var(--bg-panel);
		border: 2px solid var(--border-metal);
		border-radius: var(--border-radius-md);
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
		z-index: 100;
	}

	.task-detail-panel .panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--spacing-sm) var(--spacing-md);
		border-bottom: 1px solid var(--border-dark);
	}

	.task-detail-panel .panel-title {
		font-size: var(--font-size-sm);
		font-weight: 600;
		color: var(--accent-primary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin: 0;
	}

	.close-btn {
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: transparent;
		border: none;
		color: var(--text-muted);
		font-size: 18px;
		cursor: pointer;
		transition: color var(--transition-fast);
	}

	.close-btn:hover {
		color: var(--text-primary);
	}

	.task-detail-content {
		padding: var(--spacing-md);
	}

	.detail-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--spacing-xs) 0;
		border-bottom: 1px solid var(--border-dark);
	}

	.detail-row:last-child {
		border-bottom: none;
	}

	.detail-label {
		font-size: var(--font-size-xs);
		color: var(--text-muted);
		text-transform: uppercase;
	}

	.detail-value {
		font-size: var(--font-size-sm);
		color: var(--text-primary);
	}

	/* ステータス色 */
	.status-completed {
		color: var(--task-completed);
	}

	.status-in_progress {
		color: var(--task-in-progress);
	}

	.status-pending {
		color: var(--task-pending);
	}

	.status-blocked {
		color: var(--task-blocked);
	}

	/* 優先度色 */
	.priority-high {
		color: var(--priority-high);
	}

	.priority-medium {
		color: var(--priority-medium);
	}

	.priority-low {
		color: var(--priority-low);
	}

	/* 従来のダッシュボード */
	.dashboard {
		display: flex;
		flex-direction: column;
		gap: var(--spacing-lg);
	}

	.dashboard-row {
		display: flex;
		gap: var(--spacing-lg);
	}

	.dashboard-col-2 {
		flex: 1;
		min-width: 0;
	}

	.dashboard-col-full {
		flex: 1;
		min-width: 0;
	}

	@media (max-width: 1024px) {
		.dashboard-row {
			flex-direction: column;
		}

		.viewer-container {
			height: calc(100vh - 180px);
		}
	}
</style>
