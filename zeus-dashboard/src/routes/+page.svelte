<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { FactorioViewer } from '$lib/viewer';
	import { UseCaseView } from '$lib/viewer/usecase';
	import { ActivityView } from '$lib/viewer/activity';
	import { refreshAllData, currentView } from '$lib/stores';
	import { setConnected, setDisconnected, setConnecting } from '$lib/stores/connection';
	import { connectSSE, disconnectSSE } from '$lib/api/sse';
	import { fetchActivities, fetchUnifiedGraph } from '$lib/api/client';
	import { NODE_TYPE_CONFIG, DEFAULT_NODE_TYPE } from '$lib/viewer/config/nodeTypes';
	import type {
		ActivityItem,
		GraphNode,
		GraphEdge,
		UnifiedGraphResponse,
		GraphNodeType
	} from '$lib/types/api';

	// グラフデータ型（GraphNode/Edge の組み合わせ）
	interface GraphData {
		nodes: GraphNode[];
		edges: GraphEdge[];
	}

	// バックエンドの型をフロントエンドの GraphNodeType に安全にマッピング
	function mapNodeType(backendType: string): GraphNodeType {
		if (backendType in NODE_TYPE_CONFIG) {
			return backendType as GraphNodeType;
		}
		return DEFAULT_NODE_TYPE;
	}

	// UnifiedGraph を GraphData に変換するヘルパー
	function convertUnifiedGraphToGraphData(unified: UnifiedGraphResponse): GraphData {
		const nodes: GraphNode[] = unified.nodes.map((n) => ({
			id: n.id,
			title: n.title,
			node_type: mapNodeType(n.type),
			status: n.status,
			priority: n.priority,
			assignee: n.assignee,
			dependencies: n.parents || []
		}));

		const edges: GraphEdge[] = unified.edges.map((e) => ({
			from: e.source,
			to: e.target
		}));

		return { nodes, edges };
	}

	let useSSE = $state(true);
	let pollingInterval: ReturnType<typeof setInterval> | null = null;

	// グラフデータ（Graph View 用）- UnifiedGraph API から取得
	let graphData = $state<GraphData>({ nodes: [], edges: [] });

	// Activity データ（UseCase View の関連 Activity 表示用）
	let activitiesData: ActivityItem[] = $state([]);

	// 選択中のタスク
	let selectedTaskId: string | null = $state(null);

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

				// UnifiedGraph からグラフデータを取得（Graph View 用）
				fetchUnifiedGraph()
					.then((data) => {
						graphData = convertUnifiedGraphToGraphData(data);
					})
					.catch((err) => {
						console.warn('UnifiedGraph fetch failed:', err);
					});

				// Activity データを取得（UseCase View の関連 Activity 表示用）
				fetchActivities()
					.then((data) => {
						activitiesData = data.activities || [];
					})
					.catch((err) => {
						console.warn('Activities data fetch failed:', err);
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
</script>

<!-- ビューワーコンテナ -->
<div class="viewer-container">
	{#if $currentView === 'graph'}
		<FactorioViewer
			{graphData}
			{selectedTaskId}
			onTaskSelect={handleTaskSelect}
			onTaskHover={handleTaskHover}
		/>
	{:else if $currentView === 'usecase'}
		<UseCaseView activities={activitiesData} />
	{:else if $currentView === 'activity'}
		<ActivityView />
	{/if}
</div>

<!-- 選択ノード詳細パネル（Graph View） -->
{#if $currentView === 'graph' && selectedTaskId && graphData}
	{@const selectedNode = graphData.nodes.find((n) => n.id === selectedTaskId)}
	{#if selectedNode}
		<div class="task-detail-panel">
			<div class="panel-header">
				<h3 class="panel-title">NODE DETAIL</h3>
				<button class="close-btn" onclick={() => (selectedTaskId = null)}>x</button>
			</div>
			<div class="task-detail-content">
				<div class="detail-row">
					<span class="detail-label">ID</span>
					<span class="detail-value">{selectedNode.id}</span>
				</div>
				<div class="detail-row">
					<span class="detail-label">Title</span>
					<span class="detail-value">{selectedNode.title}</span>
				</div>
				<div class="detail-row">
					<span class="detail-label">Type</span>
					<span class="detail-value node-type-{selectedNode.node_type}"
						>{selectedNode.node_type}</span
					>
				</div>
				<div class="detail-row">
					<span class="detail-label">Status</span>
					<span class="detail-value status-{selectedNode.status}">{selectedNode.status}</span>
				</div>
				{#if selectedNode.priority}
					<div class="detail-row">
						<span class="detail-label">Priority</span>
						<span class="detail-value priority-{selectedNode.priority}"
							>{selectedNode.priority}</span
						>
					</div>
				{/if}
				<div class="detail-row">
					<span class="detail-label">Assignee</span>
					<span class="detail-value">{selectedNode.assignee || 'Unassigned'}</span>
				</div>
				{#if selectedNode.dependencies.length > 0}
					<div class="detail-row">
						<span class="detail-label">Dependencies</span>
						<span class="detail-value">{selectedNode.dependencies.length} nodes</span>
					</div>
				{/if}
			</div>
		</div>
	{/if}
{/if}

<!-- UseCase View / Activity View は内部で詳細パネルを管理 -->

<style>
	/* ビューワーコンテナ - 画面最大化 */
	.viewer-container {
		height: 100%;
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

	/* ノードタイプ色（NODE_TYPE_CONFIG と同期） */
	.node-type-vision {
		color: #ffd700;
	}

	.node-type-objective {
		color: #6699ff;
	}

	.node-type-deliverable {
		color: #66cc99;
	}

	.node-type-activity {
		color: #cc8844;
	}

	.node-type-usecase {
		color: #9966cc;
	}
</style>
