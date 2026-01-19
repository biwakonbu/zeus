<script lang="ts">
	// ヒートマップグリッド
	// CSS Grid でヒートマップセルを配置
	// 色は進捗率から計算
	interface DensityItem {
		id: string;
		title: string;
		taskCount: number;
		progress: number;
	}

	interface Props {
		items: DensityItem[];
		selectedId: string | null;
		sizeMetric: 'tasks' | 'hours';
		onSelect: (id: string, type: string) => void;
	}
	let { items, selectedId, sizeMetric, onSelect }: Props = $props();

	// 進捗率から色を計算
	function getProgressColor(progress: number): string {
		if (progress >= 70) return 'var(--status-good, #44cc44)';
		if (progress >= 40) return 'var(--status-fair, #ffcc00)';
		return 'var(--status-poor, #ee4444)';
	}

	// 進捗率から背景の透明度を計算
	function getOpacity(progress: number): number {
		return 0.3 + (progress / 100) * 0.7;
	}

	function handleKeydown(event: KeyboardEvent, id: string) {
		if (event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			onSelect(id, 'objective');
		}
	}
</script>

<div class="heatmap-grid">
	{#each items as item (item.id)}
		<button
			class="heatmap-cell"
			class:selected={selectedId === item.id}
			style="--cell-color: {getProgressColor(item.progress)}; --cell-opacity: {getOpacity(item.progress)}"
			onclick={() => onSelect(item.id, 'objective')}
			onkeydown={(e) => handleKeydown(e, item.id)}
			aria-label="{item.title} - {item.progress}%"
		>
			<div class="cell-title">{item.title}</div>
			<div class="cell-bar">
				<div class="cell-bar-fill" style="width: {item.progress}%"></div>
			</div>
			<div class="cell-value">{sizeMetric === 'tasks' ? item.taskCount : item.taskCount * 4}</div>
			<div class="cell-progress">{item.progress}%</div>
		</button>
	{/each}

	{#if items.length === 0}
		<div class="empty-cell">
			<span>No data</span>
		</div>
	{/if}
</div>

<style>
	.heatmap-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
		gap: 12px;
		padding: 16px;
	}

	.heatmap-cell {
		aspect-ratio: 1;
		padding: 12px;
		background: rgba(var(--cell-color), var(--cell-opacity));
		background-color: var(--bg-panel, #2d2d2d);
		border: 2px solid var(--border-metal, #4a4a4a);
		border-radius: 4px;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 8px;
		cursor: pointer;
		transition: all 0.2s ease;
		position: relative;
		overflow: hidden;
	}

	.heatmap-cell::before {
		content: '';
		position: absolute;
		inset: 0;
		background-color: var(--cell-color);
		opacity: var(--cell-opacity);
		z-index: 0;
	}

	.heatmap-cell > * {
		position: relative;
		z-index: 1;
	}

	.heatmap-cell:hover {
		border-color: var(--accent-primary, #ff9533);
		transform: scale(1.02);
	}

	.heatmap-cell.selected {
		border-color: var(--accent-primary, #ff9533);
		box-shadow: 0 0 12px rgba(255, 149, 51, 0.4);
	}

	.cell-title {
		font-size: 11px;
		font-weight: 500;
		color: var(--text-primary, #ffffff);
		text-align: center;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		max-width: 100%;
		text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
	}

	.cell-bar {
		width: 100%;
		height: 6px;
		background: rgba(0, 0, 0, 0.3);
		border-radius: 3px;
		overflow: hidden;
	}

	.cell-bar-fill {
		height: 100%;
		background: var(--text-primary, #ffffff);
		opacity: 0.8;
		transition: width 0.3s ease;
	}

	.cell-value {
		font-size: 20px;
		font-weight: 700;
		color: var(--text-primary, #ffffff);
		text-shadow: 0 1px 3px rgba(0, 0, 0, 0.6);
	}

	.cell-progress {
		font-size: 12px;
		font-weight: 600;
		color: var(--text-primary, #ffffff);
		opacity: 0.9;
		text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
	}

	.empty-cell {
		grid-column: 1 / -1;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 48px;
		color: var(--text-muted, #888888);
		font-size: 13px;
	}
</style>
