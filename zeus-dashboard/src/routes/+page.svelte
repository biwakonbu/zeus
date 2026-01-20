<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { FactorioViewer, WBSViewer, TimelineViewer, ViewSwitcher, type ViewType } from '$lib/viewer';
	import { refreshAllData } from '$lib/stores';
	import { setConnected, setDisconnected, setConnecting } from '$lib/stores/connection';
	import { connectSSE, disconnectSSE } from '$lib/api/sse';
	import { fetchWBSAsGraphData } from '$lib/api/client';
	import { tasks } from '$lib/stores/tasks';
	import type { TimelineItem, WBSGraphData } from '$lib/types/api';

	let useSSE = $state(true);
	let pollingInterval: ReturnType<typeof setInterval> | null = null;

	// 現在のビュー
	let currentView: ViewType = $state('graph');

	// WBS グラフデータ（Graph View 用）
	let wbsGraphData: WBSGraphData | undefined = $state(undefined);

	// 選択中のタスク
	let selectedTaskId: string | null = $state(null);

	// WBS で選択されたノード（WBSViewer 内で EntityDetailPanel が処理するため参照のみ）
	// Note: selectedTaskId のみ同期

	// Timeline で選択されたアイテム（将来機能用）
	let _selectedTimelineItem: TimelineItem | null = $state(null);

	onMount(() => {
		// SSE 失敗時のフォールバックハンドラー
		const handleSSEFailed = () => {
			if (useSSE) {
				console.log('[Dashboard] SSE failed, falling back to polling');
				useSSE = false;
				disconnectSSE();
				startPolling();
			}
		};
		window.addEventListener('sse-failed', handleSSEFailed);

		// 初期データを読み込み
		setConnecting();
		refreshAllData()
			.then(() => {
				setConnected();

				// WBS データを取得（Graph View 用）
				fetchWBSAsGraphData()
					.then(data => {
						wbsGraphData = data;
					})
					.catch(err => {
						console.warn('WBS data fetch failed:', err);
					});

				// SSE 接続を試行（ポーリングとは排他的に実行）
				if (useSSE) {
					try {
						connectSSE();
						// SSE が成功した場合はポーリングを開始しない
					} catch (error) {
						// SSE が利用できない場合のみポーリングにフォールバック
						console.log('SSE not available, falling back to polling', error);
						useSSE = false; // SSE を無効化
						startPolling();
					}
				} else {
					startPolling();
				}
			})
			.catch(() => {
				setDisconnected();
				// エラー時もポーリングを開始（SSE は使わない）
				useSSE = false;
				startPolling();
			});

		// クリーンアップ
		return () => {
			window.removeEventListener('sse-failed', handleSSEFailed);
		};
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
	function handleTaskHover(_taskId: string | null) {
		// 必要に応じてツールチップ表示などを追加
	}

	// WBS ノード選択ハンドラ（WBSViewer の onNodeSelect に合わせた型）
	function handleWBSNodeSelect(nodeId: string, _nodeType: string) {
		// WBS詳細パネルは WBSViewer 内の EntityDetailPanel で表示されるため、
		// ここでは selectedTaskId のみ同期
		selectedTaskId = nodeId;
	}

	// Timeline アイテム選択ハンドラ
	function handleTimelineItemSelect(item: TimelineItem | null) {
		_selectedTimelineItem = item;
		// タスク ID も同期
		selectedTaskId = item?.task_id ?? null;
	}

	// ビュー切り替えハンドラ
	function handleViewChange(view: ViewType) {
		currentView = view;
		// ビュー切り替え時に選択をクリア
		selectedTaskId = null;
		_selectedTimelineItem = null;
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
			graphData={wbsGraphData}
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

<!-- WBS View のノード詳細パネルは WBSViewer 内の EntityDetailPanel で表示 -->

<style>
	/* ビュー切り替えヘッダー */
	.view-header {
		display: flex;
		justify-content: center;
		padding: var(--spacing-xs) var(--spacing-md);
		background-color: var(--bg-secondary);
		border-bottom: 1px solid var(--border-dark);
	}

	/* ビューワーコンテナ - 画面最大化 */
	.viewer-container {
		height: calc(100vh - 85px);
		min-height: 400px;
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
