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
			structural_depth: n.structural_depth
		}));

		const edges: GraphEdge[] = unified.edges.map((e) => ({
			from: e.source,
			to: e.target,
			layer: e.layer,
			relation: e.relation
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

<style>
	/* ビューワーコンテナ - 画面最大化 */
	.viewer-container {
		height: 100%;
	}
</style>
