<script lang="ts">
	// Activity View - PixiJS 版
	// ミニマルデザイン: キャンバスが主役、パネルはオーバーレイで必要時のみ表示
	import { onMount, onDestroy } from 'svelte';
	import type { ActivitiesResponse, ActivityItem, ActivityNodeItem } from '$lib/types/api';
	import { fetchActivities, fetchActivityDiagram } from '$lib/api/client';
	import { Icon, EmptyState, OverlayPanel } from '$lib/components/ui';
	import { ActivityEngine } from './engine/ActivityEngine';
	import ActivityListPanel from './ActivityListPanel.svelte';
	import ActivityDetailPanel from './ActivityDetailPanel.svelte';
	import {
		updateActivityViewState,
		resetActivityViewState,
		pendingNavigation,
		clearPendingNavigation
	} from '$lib/stores/view';

	type Props = {
		onActivitySelect?: (activity: ActivityItem) => void;
		onNodeSelect?: (node: ActivityNodeItem) => void;
	};
	let { onActivitySelect, onNodeSelect }: Props = $props();

	// データ状態
	let activitiesData: ActivitiesResponse | null = $state(null);
	let currentActivity: ActivityItem | null = $state(null);
	let loading = $state(true);
	let error: string | null = $state(null);

	// パネル表示状態
	// リストパネルはデフォルト表示（アクティビティを選択するため）
	let showListPanel = $state(true);
	let showDetailPanel = $state(false);

	// 選択状態
	let selectedActivityId: string | null = $state(null);
	let selectedNodeId: string | null = $state(null);

	// 選択されたエンティティ
	const selectedActivity = $derived.by((): ActivityItem | null => {
		if (!selectedActivityId || !activitiesData) return null;
		return activitiesData.activities.find((a: ActivityItem) => a.id === selectedActivityId) ?? null;
	});

	const selectedNode = $derived.by((): ActivityNodeItem | null => {
		if (!selectedNodeId || !currentActivity) return null;
		return currentActivity.nodes.find((n: ActivityNodeItem) => n.id === selectedNodeId) ?? null;
	});

	// アクティビティが選択されているか（詳細パネル表示条件）
	const hasActivitySelection = $derived(currentActivity !== null);

	// ホバー状態（Tooltip用）
	let hoveredNode: ActivityNodeItem | null = $state(null);
	let hoverPosition = $state({ x: 0, y: 0 });

	// Tooltip 位置
	const TOOLTIP_WIDTH = 220;
	const TOOLTIP_HEIGHT = 100;
	const TOOLTIP_OFFSET = 16;

	const tooltipStyle = $derived(() => {
		const viewportWidth = typeof window !== 'undefined' ? window.innerWidth : 1920;
		const viewportHeight = typeof window !== 'undefined' ? window.innerHeight : 1080;
		const flipX = hoverPosition.x + TOOLTIP_WIDTH + TOOLTIP_OFFSET > viewportWidth;
		const flipY = hoverPosition.y + TOOLTIP_HEIGHT + TOOLTIP_OFFSET > viewportHeight;
		const left = flipX
			? hoverPosition.x - TOOLTIP_WIDTH - TOOLTIP_OFFSET
			: hoverPosition.x + TOOLTIP_OFFSET;
		const top = flipY
			? hoverPosition.y - TOOLTIP_HEIGHT - TOOLTIP_OFFSET
			: hoverPosition.y + TOOLTIP_OFFSET;
		return `left: ${Math.max(8, left)}px; top: ${Math.max(8, top)}px;`;
	});

	// PixiJS エンジン
	let engine: ActivityEngine | null = null;
	let canvasContainer: HTMLDivElement | null = $state(null);
	let currentZoom = $state(1.0);

	// データ取得
	async function loadData() {
		loading = true;
		error = null;
		try {
			activitiesData = await fetchActivities();
			// 最初のアクティビティを自動選択
			if (activitiesData.activities.length > 0) {
				await selectActivity(activitiesData.activities[0].id);
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'データの取得に失敗しました';
		} finally {
			loading = false;
		}
	}

	// アクティビティ選択
	async function selectActivity(activityId: string) {
		if (selectedActivityId === activityId) return;

		try {
			const response = await fetchActivityDiagram(activityId);
			if (response.activity) {
				selectedActivityId = activityId;
				currentActivity = response.activity;
				selectedNodeId = null;

				// エンジンにデータを設定
				if (engineInitialized && engine) {
					engine.setData(currentActivity);
				}

				// 詳細パネルを自動表示
				showDetailPanel = true;

				syncStoreState();
				onActivitySelect?.(currentActivity);
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'アクティビティの取得に失敗しました';
		}
	}

	// エンジン初期化状態
	let engineInitializing = false;
	let engineInitialized = false;

	// エンジン初期化（一度だけ実行）
	async function initEngine(): Promise<void> {
		if (!canvasContainer || engineInitializing || engineInitialized) return;
		engineInitializing = true;

		try {
			engine = new ActivityEngine();
			await engine.init(canvasContainer);

			engine.onNodeClicked((node) => {
				selectedNodeId = node.id;
				showDetailPanel = true;
				onNodeSelect?.(node);
			});

			engine.onNodeHovered((node, event) => {
				hoveredNode = node;
				if (event) hoverPosition = { x: event.clientX, y: event.clientY };
			});

			engine.onViewportChanged((viewport) => {
				currentZoom = viewport.scale;
				updateActivityViewState({ zoom: viewport.scale });
			});

			engineInitialized = true;

			// 初期化完了後にデータがあれば設定
			if (currentActivity) {
				engine.setData(currentActivity);
				syncStoreState();
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'エンジン初期化に失敗しました';
		} finally {
			engineInitializing = false;
		}
	}

	// ズーム操作（ヘッダーから呼び出される）
	function handleZoomIn() {
		engine?.setZoom(currentZoom * 1.2);
	}
	function handleZoomOut() {
		engine?.setZoom(currentZoom / 1.2);
	}
	function handleZoomReset() {
		engine?.centerView();
	}

	// パネル操作
	function toggleListPanel() {
		showListPanel = !showListPanel;
		updateActivityViewState({ showListPanel });
	}

	function closeListPanel() {
		showListPanel = false;
		updateActivityViewState({ showListPanel: false });
	}

	function closeDetailPanel() {
		showDetailPanel = false;
		selectedNodeId = null;
		// 視覚的な選択状態のみ解除
		engine?.clearSelection();
	}

	// ノードクリック（詳細パネルから）
	function handleNodeClickFromPanel(node: ActivityNodeItem) {
		selectedNodeId = node.id;
		// エンジンにも選択状態を反映
		engine?.selectNode(node.id);
		onNodeSelect?.(node);
	}

	// アクティビティクリック（リストから）
	function handleActivityClick(activity: ActivityItem) {
		selectActivity(activity.id);
	}

	// ノードタイプのラベル
	function getNodeTypeLabel(type: string): string {
		const labels: Record<string, string> = {
			initial: '開始',
			final: '終了',
			action: 'アクション',
			decision: '分岐',
			merge: '合流',
			fork: '並列分岐',
			join: '並列合流'
		};
		return labels[type] ?? type;
	}

	// ESC キーでパネルを閉じる
	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			if (showDetailPanel) closeDetailPanel();
			else if (showListPanel) closeListPanel();
		}
	}

	// ヘッダーの store を更新
	function syncStoreState() {
		updateActivityViewState({
			zoom: currentZoom,
			activityCount: activitiesData?.activities.length || 0,
			selectedActivityId,
			showListPanel,
			onZoomIn: handleZoomIn,
			onZoomOut: handleZoomOut,
			onZoomReset: handleZoomReset,
			onToggleListPanel: toggleListPanel
		});
	}

	onMount(() => {
		loadData();
		document.addEventListener('keydown', handleKeydown);
		return () => document.removeEventListener('keydown', handleKeydown);
	});

	// エンジン初期化 Effect（canvasContainer が利用可能になったら一度だけ）
	$effect(() => {
		if (canvasContainer && !engineInitialized && !engineInitializing) {
			initEngine();
		}
	});

	// showListPanel が変わったら store を更新
	$effect(() => {
		updateActivityViewState({ showListPanel });
	});

	$effect(() => {
		if (!canvasContainer || !engine) return () => {};
		const resizeObserver = new ResizeObserver(() => engine?.resize());
		resizeObserver.observe(canvasContainer);
		return () => resizeObserver.disconnect();
	});

	// ナビゲーションによる自動選択 Effect
	$effect(() => {
		const nav = $pendingNavigation;
		if (!nav || nav.view !== 'activity' || !activitiesData) return;

		if (nav.entityType === 'activity' && nav.entityId) {
			// Activity を選択
			const activity = activitiesData.activities.find((a: ActivityItem) => a.id === nav.entityId);
			if (activity) {
				// 非同期処理を適切にハンドリング
				(async () => {
					await selectActivity(activity.id);
					clearPendingNavigation();
				})();
			} else {
				clearPendingNavigation();
			}
		}
	});

	onDestroy(() => {
		engine?.destroy();
		engine = null;
		// store をリセット
		resetActivityViewState();
	});
</script>

<div class="activity-view">
	{#if loading}
		<div class="loading-state">
			<Icon name="RefreshCw" size={32} />
			<span>読み込み中...</span>
		</div>
	{:else if error}
		<div class="error-state">
			<Icon name="AlertTriangle" size={32} />
			<span>{error}</span>
			<button class="retry-button" onclick={loadData}>再試行</button>
		</div>
	{:else if !activitiesData || activitiesData.activities.length === 0}
		<EmptyState
			title="アクティビティ図がありません"
			description=".zeus/activities/ ディレクトリに YAML ファイルを追加してください"
			icon="Workflow"
		/>
	{:else}
		<!-- フルスクリーンキャンバス -->
		<div class="canvas-area">
			<div class="canvas-wrapper" bind:this={canvasContainer}></div>

			<!-- リストパネル（オーバーレイ） -->
			{#if showListPanel}
				<OverlayPanel
					title="アクティビティ一覧"
					position="top-left"
					width="280px"
					onClose={closeListPanel}
				>
					<ActivityListPanel
						activities={activitiesData.activities}
						{selectedActivityId}
						onActivitySelect={handleActivityClick}
					/>
				</OverlayPanel>
			{/if}

			<!-- 詳細パネル（オーバーレイ） -->
			{#if showDetailPanel && hasActivitySelection}
				<OverlayPanel
					title="プロパティ"
					position="top-right"
					width="280px"
					onClose={closeDetailPanel}
				>
					<div class="detail-content">
						<ActivityDetailPanel
							activity={currentActivity}
							selectedNode={selectedNode}
							onClose={closeDetailPanel}
							onNodeClick={handleNodeClickFromPanel}
						/>
					</div>
				</OverlayPanel>
			{/if}
		</div>

		<!-- ホバー Tooltip -->
		{#if hoveredNode}
			<div class="hover-tooltip" style={tooltipStyle()}>
				<div class="tooltip-header">
					<span class="tooltip-type-badge">{getNodeTypeLabel(hoveredNode.type)}</span>
					{#if hoveredNode.name}
						<span class="tooltip-name">{hoveredNode.name}</span>
					{/if}
				</div>
				<div class="tooltip-meta">
					<span class="tooltip-id">{hoveredNode.id}</span>
				</div>
			</div>
		{/if}
	{/if}
</div>

<style>
	.activity-view {
		width: 100%;
		height: 100%;
		background: var(--bg-primary);
		color: var(--text-primary);
	}

	.loading-state,
	.error-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 1rem;
		height: 100%;
		color: var(--text-muted);
	}

	.error-state {
		color: var(--status-poor);
	}

	.retry-button {
		padding: 0.5rem 1rem;
		background: var(--bg-secondary);
		border: 1px solid var(--border-metal);
		color: var(--text-primary);
		cursor: pointer;
		border-radius: 4px;
	}

	.retry-button:hover {
		background: var(--bg-hover);
	}

	/* フルスクリーンキャンバス */
	.canvas-area {
		position: relative;
		width: 100%;
		height: 100%;
		overflow: hidden;
	}

	.canvas-wrapper {
		width: 100%;
		height: 100%;
		background-color: #1a1a1a;
		background-image:
			radial-gradient(circle at 1px 1px, rgba(255, 149, 51, 0.08) 1px, transparent 0);
		background-size: 24px 24px;
	}

	.canvas-wrapper :global(canvas) {
		display: block;
	}

	/* 詳細パネルのコンテンツ（パディング調整） */
	.detail-content {
		padding: 12px;
	}

	/* ホバー Tooltip */
	.hover-tooltip {
		position: fixed;
		z-index: 1000;
		background: rgba(30, 30, 30, 0.95);
		border: 1px solid var(--border-metal);
		border-radius: 6px;
		padding: 10px;
		min-width: 140px;
		max-width: 220px;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
		pointer-events: none;
		backdrop-filter: blur(8px);
	}

	.tooltip-header {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 6px;
	}

	.tooltip-type-badge {
		font-size: 0.625rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		background: rgba(255, 149, 51, 0.2);
		color: var(--accent-primary);
		padding: 2px 6px;
		border-radius: 3px;
	}

	.tooltip-name {
		font-weight: 600;
		font-size: 0.8125rem;
		color: var(--text-primary);
	}

	.tooltip-meta {
		display: flex;
		gap: 6px;
		font-size: 0.6875rem;
		color: var(--text-muted);
	}

	.tooltip-id {
		font-family: monospace;
		font-size: 0.65rem;
	}
</style>
