<script lang="ts">
	// メトリクスカード
	// 単一のメトリクス（Coverage, Balance, Health など）を表示
	interface Props {
		label: string;
		value: number;
		unit?: string;
		status?: 'good' | 'fair' | 'poor';
	}
	let { label, value, unit = '%', status }: Props = $props();

	// 値からステータスを自動判定（status が未指定の場合）
	const computedStatus = $derived(status || (value >= 70 ? 'good' : value >= 40 ? 'fair' : 'poor'));
</script>

<div class="metric-card metric-card--{computedStatus}">
	<div class="metric-card__label">{label}</div>
	<div class="metric-card__value">
		{value}<span class="metric-card__unit">{unit}</span>
	</div>
</div>

<style>
	.metric-card {
		background: var(--bg-panel, #2d2d2d);
		border: 2px solid var(--border-metal, #4a4a4a);
		border-radius: 4px;
		padding: 12px 16px;
		min-width: 100px;
		text-align: center;
		transition: all 0.2s ease;
	}

	.metric-card:hover {
		border-color: var(--border-highlight, #666666);
	}

	.metric-card__label {
		font-size: 10px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: var(--text-muted, #888888);
		margin-bottom: 4px;
	}

	.metric-card__value {
		font-size: 28px;
		font-weight: 700;
		line-height: 1;
	}

	.metric-card__unit {
		font-size: 14px;
		font-weight: 500;
		margin-left: 2px;
	}

	/* ステータスに応じた値の色 */
	.metric-card--good .metric-card__value {
		color: var(--status-good, #44cc44);
	}

	.metric-card--fair .metric-card__value {
		color: var(--status-fair, #ffcc00);
	}

	.metric-card--poor .metric-card__value {
		color: var(--status-poor, #ee4444);
	}
</style>
