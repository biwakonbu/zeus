<script lang="ts">
	// Density View
	// 作業量の分布をヒートマップで可視化するビュー
	import HeatmapGrid from './HeatmapGrid.svelte';
	import { selectedEntityId } from '../stores/wbsStore';
	import type { WBSAggregatedResponse, ProgressNode } from '$lib/types/api';
	import { Icon } from '$lib/components/ui';

	interface Props {
		data: WBSAggregatedResponse | null;
		onNodeSelect: (nodeId: string, nodeType: string) => void;
	}
	let { data, onNodeSelect }: Props = $props();

	// サイズ指標の状態
	let sizeMetric: 'tasks' | 'hours' = $state('tasks');

	// Objective をヒートマップ用に変換
	interface DensityItem {
		id: string;
		title: string;
		taskCount: number;
		progress: number;
	}

	const items = $derived<DensityItem[]>(
		(data?.progress?.objectives ?? []).map((obj: ProgressNode) => ({
			id: obj.id,
			title: obj.title,
			taskCount: obj.children_count,
			progress: obj.progress
		}))
	);

	function handleSizeMetricChange(metric: 'tasks' | 'hours') {
		sizeMetric = metric;
	}
</script>

<div class="density-view">
	<div class="density-header">
		<span class="density-label">DENSITY</span>
		<div class="size-selector">
			<span class="size-label">Size:</span>
			<div class="size-buttons">
				<button
					class="size-btn"
					class:active={sizeMetric === 'tasks'}
					onclick={() => handleSizeMetricChange('tasks')}
				>
					Tasks
				</button>
				<button
					class="size-btn"
					class:active={sizeMetric === 'hours'}
					onclick={() => handleSizeMetricChange('hours')}
				>
					Hours
				</button>
			</div>
		</div>
	</div>
	<div class="density-content">
		<HeatmapGrid {items} selectedId={$selectedEntityId} {sizeMetric} onSelect={onNodeSelect} />

		{#if items.length === 0}
			<div class="empty-state">
				<span class="empty-icon"><Icon name="Flame" size={32} /></span>
				<span class="empty-text">Density データがありません</span>
			</div>
		{/if}
	</div>
</div>

<style>
	.density-view {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: var(--bg-primary, #1a1a1a);
	}

	.density-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 16px;
		background: var(--bg-secondary, #242424);
		border-bottom: 1px solid var(--border-metal, #4a4a4a);
	}

	.density-label {
		font-size: 12px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: var(--accent-primary, #ff9533);
	}

	.size-selector {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.size-label {
		font-size: 11px;
		color: var(--text-muted, #888888);
	}

	.size-buttons {
		display: flex;
		gap: 2px;
		background: var(--bg-panel, #2d2d2d);
		border-radius: 4px;
		padding: 2px;
	}

	.size-btn {
		padding: 4px 10px;
		font-size: 11px;
		font-weight: 500;
		background: transparent;
		border: none;
		border-radius: 2px;
		color: var(--text-muted, #888888);
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.size-btn:hover {
		color: var(--text-primary, #ffffff);
	}

	.size-btn.active {
		background: var(--accent-primary, #ff9533);
		color: var(--bg-primary, #1a1a1a);
	}

	.density-content {
		flex: 1;
		overflow-y: auto;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 48px 16px;
		color: var(--text-muted, #888888);
	}

	.empty-icon {
		font-size: 32px;
		opacity: 0.5;
		margin-bottom: 8px;
	}

	.empty-text {
		font-size: 13px;
	}
</style>
