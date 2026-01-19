<script lang="ts">
	// Props
	interface Props {
		score: number; // 0-100
		label: string;
		subLabel?: string;
		size?: 'small' | 'medium' | 'large';
	}
	let { score, label, subLabel, size = 'medium' }: Props = $props();

	// スコアに応じた色
	function getScoreColor(s: number): string {
		if (s >= 80) return '#22c55e'; // 緑
		if (s >= 60) return '#f59e0b'; // 黄
		if (s >= 40) return '#f97316'; // オレンジ
		return '#ef4444'; // 赤
	}

	let color = $derived(getScoreColor(score));
	let circumference = $derived(2 * Math.PI * 45);
	let dashOffset = $derived(circumference - (score / 100) * circumference);
</script>

<div class="coverage-indicator {size}">
	<div class="gauge">
		<svg viewBox="0 0 100 100">
			<!-- 背景円 -->
			<circle class="gauge-bg" cx="50" cy="50" r="45" />
			<!-- 進捗円 -->
			<circle
				class="gauge-fill"
				cx="50"
				cy="50"
				r="45"
				style="stroke: {color}; stroke-dasharray: {circumference}; stroke-dashoffset: {dashOffset};"
			/>
		</svg>
		<div class="gauge-content">
			<span class="gauge-value" style="color: {color}">{score}</span>
			<span class="gauge-percent">%</span>
		</div>
	</div>
	<div class="indicator-labels">
		<span class="indicator-label">{label}</span>
		{#if subLabel}
			<span class="indicator-sublabel">{subLabel}</span>
		{/if}
	</div>
</div>

<style>
	.coverage-indicator {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
	}

	.gauge {
		position: relative;
	}

	.coverage-indicator.small .gauge {
		width: 60px;
		height: 60px;
	}

	.coverage-indicator.medium .gauge {
		width: 80px;
		height: 80px;
	}

	.coverage-indicator.large .gauge {
		width: 100px;
		height: 100px;
	}

	.gauge svg {
		width: 100%;
		height: 100%;
		transform: rotate(-90deg);
	}

	.gauge-bg {
		fill: none;
		stroke: #333;
		stroke-width: 6;
	}

	.gauge-fill {
		fill: none;
		stroke-width: 6;
		stroke-linecap: round;
		transition:
			stroke-dashoffset 0.5s ease,
			stroke 0.3s ease;
	}

	.gauge-content {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		display: flex;
		align-items: baseline;
	}

	.coverage-indicator.small .gauge-value {
		font-size: 16px;
	}
	.coverage-indicator.small .gauge-percent {
		font-size: 10px;
	}

	.coverage-indicator.medium .gauge-value {
		font-size: 20px;
	}
	.coverage-indicator.medium .gauge-percent {
		font-size: 12px;
	}

	.coverage-indicator.large .gauge-value {
		font-size: 26px;
	}
	.coverage-indicator.large .gauge-percent {
		font-size: 14px;
	}

	.gauge-value {
		font-weight: 700;
	}

	.gauge-percent {
		color: #888;
		font-weight: 500;
	}

	.indicator-labels {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2px;
	}

	.indicator-label {
		font-size: 12px;
		color: #ccc;
		font-weight: 500;
	}

	.indicator-sublabel {
		font-size: 10px;
		color: #888;
	}
</style>
