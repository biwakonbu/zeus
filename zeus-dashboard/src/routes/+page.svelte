<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { FactorioViewer, WBSViewer, TimelineViewer, ViewSwitcher, type ViewType } from '$lib/viewer';
	import { refreshAllData } from '$lib/stores';
	import { setConnected, setDisconnected, setConnecting } from '$lib/stores/connection';
	import { connectSSE, disconnectSSE } from '$lib/api/sse';
	import { tasks } from '$lib/stores/tasks';
	import type { WBSNode, TimelineItem } from '$lib/types/api';

	let useSSE = $state(true);
	let pollingInterval: ReturnType<typeof setInterval> | null = null;

	// 現在のビュー
	let currentView: ViewType = $state('graph');

	// 選択中のタスク
	let selectedTaskId: string | null = $state(null);

	// WBS で選択されたノード
	let selectedWBSNode: WBSNode | null = $state(null);

	// Timeline で選択されたアイテム
	let selectedTimelineItem: TimelineItem | null = $state(null);

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

	// タスク選択ハンドラ
	function handleTaskSelect(taskId: string | null) {
		selectedTaskId = taskId;
	}

	// タスクホバーハンドラ
	function handleTaskHover(taskId: string | null) {
		// 必要に応じてツールチップ表示などを追加
	}

	// WBS ノード選択ハンドラ
	function handleWBSNodeSelect(node: WBSNode | null) {
		selectedWBSNode = node;
		// タスク ID も同期
		selectedTaskId = node?.id ?? null;
	}

	// Timeline アイテム選択ハンドラ
	function handleTimelineItemSelect(item: TimelineItem | null) {
		selectedTimelineItem = item;
		// タスク ID も同期
		selectedTaskId = item?.task_id ?? null;
	}

	// ビュー切り替えハンドラ
	function handleViewChange(view: ViewType) {
		currentView = view;
		// ビュー切り替え時に選択をクリア
		selectedTaskId = null;
		selectedWBSNode = null;
		selectedTimelineItem = null;
	}
</script>

<!-- ビュー切り替えヘッダー -->
<div class="view-header">
	<ViewSwitcher
		{currentView}
		onViewChange={handleViewChange}
	/>
</div>

<!-- ビューワーコンテナ -->
<div class="viewer-container">
	{#if currentView === 'graph'}
		<FactorioViewer
			tasks={$tasks}
			selectedTaskId={selectedTaskId}
			onTaskSelect={handleTaskSelect}
			onTaskHover={handleTaskHover}
		/>
	{:else if currentView === 'wbs'}
		<WBSViewer onNodeSelect={handleWBSNodeSelect} />
	{:else if currentView === 'timeline'}
		<TimelineViewer onTaskSelect={handleTimelineItemSelect} />
	{/if}
</div>

<!-- 選択タスク詳細パネル（Graph View） -->
{#if currentView === 'graph' && selectedTaskId}
	{@const selectedTask = $tasks.find(t => t.id === selectedTaskId)}
	{#if selectedTask}
		<div class="task-detail-panel">
			<div class="panel-header">
				<h3 class="panel-title">TASK DETAIL</h3>
				<button class="close-btn" onclick={() => selectedTaskId = null}>x</button>
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

<!-- 選択ノード詳細パネル（WBS View） -->
{#if currentView === 'wbs' && selectedWBSNode}
	<div class="task-detail-panel">
		<div class="panel-header">
			<h3 class="panel-title">WBS NODE DETAIL</h3>
			<button class="close-btn" onclick={() => { selectedWBSNode = null; selectedTaskId = null; }}>x</button>
		</div>
		<div class="task-detail-content">
			{#if selectedWBSNode.wbs_code}
				<div class="detail-row">
					<span class="detail-label">WBS Code</span>
					<span class="detail-value wbs-code">{selectedWBSNode.wbs_code}</span>
				</div>
			{/if}
			<div class="detail-row">
				<span class="detail-label">Title</span>
				<span class="detail-value">{selectedWBSNode.title}</span>
			</div>
			<div class="detail-row">
				<span class="detail-label">Status</span>
				<span class="detail-value status-{selectedWBSNode.status}">{selectedWBSNode.status}</span>
			</div>
			<div class="detail-row">
				<span class="detail-label">Progress</span>
				<div class="progress-detail">
					<div class="progress-bar-detail">
						<div class="progress-fill" style="width: {selectedWBSNode.progress}%"></div>
					</div>
					<span class="progress-value">{selectedWBSNode.progress}%</span>
				</div>
			</div>
			<div class="detail-row">
				<span class="detail-label">Priority</span>
				<span class="detail-value priority-{selectedWBSNode.priority}">{selectedWBSNode.priority}</span>
			</div>
			<div class="detail-row">
				<span class="detail-label">Assignee</span>
				<span class="detail-value">{selectedWBSNode.assignee || 'Unassigned'}</span>
			</div>
			<div class="detail-row">
				<span class="detail-label">Depth</span>
				<span class="detail-value">{selectedWBSNode.depth}</span>
			</div>
			{#if selectedWBSNode.children && selectedWBSNode.children.length > 0}
				<div class="detail-row">
					<span class="detail-label">Children</span>
					<span class="detail-value">{selectedWBSNode.children.length} subtasks</span>
				</div>
			{/if}
		</div>
	</div>
{/if}

<style>
	/* ビュー切り替えヘッダー */
	.view-header {
		display: flex;
		justify-content: center;
		padding: var(--spacing-sm) var(--spacing-md);
		background-color: var(--bg-secondary);
		border-bottom: 1px solid var(--border-dark);
	}

	/* ビューワーコンテナ */
	.viewer-container {
		height: calc(100vh - 180px);
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

	/* WBS コード */
	.wbs-code {
		font-family: 'JetBrains Mono', 'Fira Code', monospace;
		background: var(--bg-secondary);
		padding: 2px 8px;
		border-radius: var(--border-radius-sm);
		color: var(--accent-primary);
	}

	/* プログレスバー詳細 */
	.progress-detail {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		flex: 1;
	}

	.progress-bar-detail {
		flex: 1;
		height: 8px;
		background: var(--bg-secondary);
		border-radius: 4px;
		overflow: hidden;
	}

	.progress-fill {
		height: 100%;
		background: linear-gradient(90deg, var(--accent-primary), var(--accent-secondary));
		transition: width 0.3s;
	}

	.progress-value {
		font-size: var(--font-size-sm);
		color: var(--accent-primary);
		font-weight: 600;
		min-width: 40px;
		text-align: right;
	}

	@media (max-width: 1024px) {
		.viewer-container {
			height: calc(100vh - 160px);
		}
	}
</style>
