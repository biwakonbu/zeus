<script lang="ts">
	interface Props {
		value: number;
		max?: number;
		showLabel?: boolean;
		size?: 'sm' | 'md' | 'lg';
	}

	let { value, max = 100, showLabel = true, size = 'md' }: Props = $props();

	// パーセント計算
	const percent = $derived(Math.min(100, Math.max(0, (value / max) * 100)));

	// 色を決定
	const barColor = $derived(() => {
		if (percent >= 80) return 'var(--status-good)';
		if (percent >= 50) return 'var(--status-fair)';
		return 'var(--status-poor)';
	});
</script>

<div class="progress-container progress-{size}">
	<div class="progress-bar">
		<div class="progress-fill" style="width: {percent}%; background-color: {barColor()}">
			<div class="progress-shine"></div>
		</div>
	</div>
	{#if showLabel}
		<span class="progress-label">{Math.round(percent)}%</span>
	{/if}
</div>

<style>
	.progress-container {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		width: 100%;
	}

	.progress-bar {
		flex: 1;
		background-color: var(--bg-secondary);
		border: 2px solid var(--border-metal);
		border-radius: var(--border-radius-sm);
		overflow: hidden;
		position: relative;
	}

	/* サイズ */
	.progress-sm .progress-bar {
		height: 8px;
	}

	.progress-md .progress-bar {
		height: 16px;
	}

	.progress-lg .progress-bar {
		height: 24px;
	}

	.progress-fill {
		height: 100%;
		transition: width 0.3s ease;
		position: relative;
		overflow: hidden;
	}

	/* インダストリアル感のある光沢 */
	.progress-shine {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 50%;
		background: linear-gradient(180deg, rgba(255, 255, 255, 0.2) 0%, transparent 100%);
	}

	.progress-label {
		font-size: var(--font-size-sm);
		font-weight: 600;
		color: var(--accent-primary);
		min-width: 3em;
		text-align: right;
		font-variant-numeric: tabular-nums;
	}
</style>
