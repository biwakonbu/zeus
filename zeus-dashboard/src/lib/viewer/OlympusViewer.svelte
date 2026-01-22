<script lang="ts">
	// WBS ビューワー（改善版）
	// 4 つのビュー（Health, Timeline, Density, Affinity）を提供
	import { onMount } from 'svelte';
	import { fetchWBSAggregated, fetchAffinity } from '$lib/api/client';
	import type { WBSAggregatedResponse, AffinityResponse } from '$lib/types/api';
	import HealthView from './wbs/health/HealthView.svelte';
	import TimelineView from './wbs/timeline/TimelineView.svelte';
	import DensityView from './wbs/density/DensityView.svelte';
	import AffinityView from './wbs/affinity/AffinityView.svelte';
	import WBSSummaryBar from './wbs/WBSSummaryBar.svelte';
	import EntityDetailPanel from './wbs/EntityDetailPanel.svelte';
	import { Icon } from '$lib/components/ui';
	import {
		selectedEntityId,
		selectEntity,
		clearSelection
	} from './wbs/stores/wbsStore';

	// Props
	interface Props {
		onNodeSelect?: (nodeId: string, nodeType: string) => void;
	}
	let { onNodeSelect }: Props = $props();

	// ビュータブ（4視点）
	type ViewTab = 'health' | 'timeline' | 'density' | 'affinity';
	let activeView: ViewTab = $state('health');

	// 状態
	let aggregatedData: WBSAggregatedResponse | null = $state(null);
	let affinityData: AffinityResponse | null = $state(null);
	let loading = $state(true);
	let error: string | null = $state(null);

	// 詳細パネル表示
	let showDetailPanel = $state(false);

	// Store からの選択状態を購読して詳細パネルを制御
	$effect(() => {
		showDetailPanel = $selectedEntityId !== null;
	});

	// データ読み込み
	async function loadData() {
		loading = true;
		error = null;
		try {
			// 並列読み込み
			const [agg, aff] = await Promise.all([
				fetchWBSAggregated(),
				fetchAffinity().catch(() => null)  // Affinity はオプショナル
			]);
			aggregatedData = agg;
			affinityData = aff;
		} catch (e) {
			error = e instanceof Error ? e.message : 'WBS データの読み込みに失敗しました';
		} finally {
			loading = false;
		}
	}

	// ノード選択ハンドラ（各ビューから呼ばれる）
	function handleNodeSelect(nodeId: string, nodeType: string) {
		selectEntity(nodeId, nodeType);
		onNodeSelect?.(nodeId, nodeType);
	}

	// エンティティ選択ハンドラ（詳細パネルから呼ばれる）
	function handleEntitySelect(entityId: string) {
		selectEntity(entityId, null);
	}

	// 詳細パネルを閉じる
	function closeDetailPanel() {
		clearSelection();
	}

	// タブ情報（Phase 7: Affinity タブ追加）
	const tabs: Array<{ id: ViewTab; label: string; icon: string }> = [
		{ id: 'health', label: 'Health', icon: 'Heart' },
		{ id: 'timeline', label: 'Timeline', icon: 'Calendar' },
		{ id: 'density', label: 'Density', icon: 'Flame' },
		{ id: 'affinity', label: 'Affinity', icon: 'GitBranch' }
	];

	onMount(() => {
		loadData();
	});
</script>

<div class="wbs-viewer">
	<!-- ヘッダー & ビュー切り替え -->
	<div class="wbs-header">
		<div class="view-tabs">
			{#each tabs as tab}
				<button
					class="view-tab"
					class:active={activeView === tab.id}
					onclick={() => (activeView = tab.id)}
					aria-pressed={activeView === tab.id}
				>
					<span class="tab-icon"><Icon name={tab.icon} size={14} /></span>
					<span class="tab-label">{tab.label}</span>
				</button>
			{/each}
		</div>

		<div class="header-actions">
			<button class="refresh-btn" onclick={() => loadData()} title="更新" disabled={loading}>
				<span class="icon" class:spinning={loading}><Icon name="RefreshCw" size={14} /></span>
			</button>
		</div>
	</div>

	<!-- メインコンテンツ（ビュー + 詳細パネル） -->
	<div class="main-content">
		<!-- ビューコンテンツ -->
		<div class="view-content" class:with-panel={showDetailPanel}>
			{#if loading && !aggregatedData}
				<div class="loading-state">
					<div class="spinner"></div>
					<span>読み込み中...</span>
				</div>
			{:else if error}
				<div class="error-state">
					<span class="error-icon"><Icon name="AlertTriangle" size={32} /></span>
					<span>{error}</span>
					<button class="retry-btn" onclick={() => loadData()}>再試行</button>
				</div>
			{:else}
				<!-- 4視点ビュー（Phase 7: Affinity 追加） -->
				<div class="view-container">
					{#if activeView === 'health'}
						<HealthView data={aggregatedData} onNodeSelect={handleNodeSelect} />
					{:else if activeView === 'timeline'}
						<TimelineView data={aggregatedData} onNodeSelect={handleNodeSelect} />
					{:else if activeView === 'density'}
						<DensityView data={aggregatedData} onNodeSelect={handleNodeSelect} />
					{:else if activeView === 'affinity'}
						<AffinityView data={affinityData} onNodeSelect={handleNodeSelect} />
					{/if}
				</div>
			{/if}
		</div>

		<!-- エンティティ詳細パネル -->
		{#if showDetailPanel}
			<div class="detail-panel">
				<EntityDetailPanel
					entityId={$selectedEntityId}
					onClose={closeDetailPanel}
					onEntitySelect={handleEntitySelect}
				/>
			</div>
		{/if}
	</div>

	<!-- サマリーバー -->
	<WBSSummaryBar data={aggregatedData} />
</div>

<style>
	.wbs-viewer {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: var(--bg-primary, #1a1a1a);
		color: var(--text-primary, #e0e0e0);
		font-family: var(--font-family, 'IBM Plex Mono', 'JetBrains Mono', monospace);
	}

	/* ヘッダー */
	.wbs-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0 16px;
		background: var(--bg-secondary, #252525);
		border-bottom: 1px solid var(--border-metal, #3a3a3a);
		min-height: 48px;
	}

	.view-tabs {
		display: flex;
		gap: 4px;
	}

	.view-tab {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 10px 16px;
		background: transparent;
		border: none;
		border-bottom: 2px solid transparent;
		color: var(--text-muted, #888);
		font-size: 13px;
		font-family: inherit;
		cursor: pointer;
		transition: all 0.2s;
	}

	.view-tab:hover {
		color: var(--text-secondary, #ccc);
		background: var(--bg-hover, #2a2a2a);
	}

	.view-tab.active {
		color: var(--accent-primary, #f59e0b);
		border-bottom-color: var(--accent-primary, #f59e0b);
	}

	.tab-icon {
		display: flex;
		align-items: center;
		font-size: 14px;
	}

	.header-actions {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.refresh-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 6px 10px;
		background: var(--bg-panel, #333);
		border: 1px solid var(--border-metal, #444);
		color: var(--text-secondary, #ccc);
		border-radius: 4px;
		cursor: pointer;
		font-size: 14px;
		transition: all 0.2s;
	}

	.refresh-btn:hover:not(:disabled) {
		background: var(--bg-hover, #444);
		border-color: var(--accent-primary, #f59e0b);
		color: var(--accent-primary, #f59e0b);
	}

	.refresh-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.refresh-btn .icon {
		display: inline-flex;
	}

	.refresh-btn .icon.spinning {
		animation: spin 1s linear infinite;
	}

	/* メインコンテンツ */
	.main-content {
		flex: 1;
		display: flex;
		overflow: hidden;
	}

	/* ビューコンテンツ */
	.view-content {
		flex: 1;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		transition: width 0.3s ease;
	}

	.view-content.with-panel {
		flex: 2;
	}

	.view-container {
		flex: 1;
		overflow: hidden;
	}

	/* 詳細パネル */
	.detail-panel {
		flex: 1;
		min-width: 320px;
		max-width: 400px;
		border-left: 1px solid var(--border-dark, #333);
		overflow: hidden;
	}

	/* ローディング・エラー状態 */
	.loading-state,
	.error-state {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 16px;
		color: var(--text-muted, #888);
	}

	.spinner {
		width: 32px;
		height: 32px;
		border: 3px solid var(--bg-panel, #333);
		border-top-color: var(--accent-primary, #f59e0b);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.error-icon {
		display: flex;
		color: var(--status-poor, #ef4444);
	}

	.retry-btn {
		padding: 8px 16px;
		background: var(--bg-secondary, #252525);
		border: 1px solid var(--accent-primary, #f59e0b);
		color: var(--accent-primary, #f59e0b);
		border-radius: 4px;
		cursor: pointer;
		font-family: inherit;
		transition: all 0.2s;
	}

	.retry-btn:hover {
		background: var(--accent-primary, #f59e0b);
		color: var(--bg-primary, #1a1a1a);
	}
</style>
