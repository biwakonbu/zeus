<script lang="ts">
	// メトリクスパネル
	// 3 つの主要メトリクス（Coverage, Balance, Health）を横並びで表示
	import MetricCard from './MetricCard.svelte';
	import StatusBadge from '../shared/StatusBadge.svelte';

	interface Props {
		coverage: number; // 網羅度 (0-100)
		balance: number; // バランス (0-100)
		overallHealth: number; // 総合健全性 (0-100)
		healthStatus: 'good' | 'fair' | 'poor';
	}
	let { coverage, balance, overallHealth, healthStatus }: Props = $props();
</script>

<div class="metrics-panel">
	<div class="metrics-panel__cards">
		<MetricCard label="Coverage" value={coverage} />
		<MetricCard label="Balance" value={balance} />
		<MetricCard label="Health" value={overallHealth} status={healthStatus} />
	</div>
	<div class="metrics-panel__summary">
		<span class="summary-label">Overall:</span>
		<span class="summary-value">{overallHealth}%</span>
		<StatusBadge status={healthStatus} />
	</div>
</div>

<style>
	.metrics-panel {
		background: var(--bg-secondary, #242424);
		border-bottom: 1px solid var(--border-metal, #4a4a4a);
		padding: 16px;
	}

	.metrics-panel__cards {
		display: flex;
		gap: 12px;
		flex-wrap: wrap;
	}

	.metrics-panel__summary {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-top: 12px;
		padding-top: 12px;
		border-top: 1px solid var(--border-dark, #333333);
	}

	.summary-label {
		font-size: 12px;
		color: var(--text-muted, #888888);
	}

	.summary-value {
		font-size: 14px;
		font-weight: 600;
		color: var(--text-primary, #ffffff);
	}
</style>
