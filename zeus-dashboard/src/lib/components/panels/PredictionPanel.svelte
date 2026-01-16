<script lang="ts">
	import Panel from '$lib/components/ui/Panel.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import { completion, risk, velocity, predictionLoading, predictionError, hasSufficientData } from '$lib/stores/prediction';
	import type { RiskLevel, VelocityTrend } from '$lib/types/api';

	// リスクレベルのバッジバリアント
	function getRiskVariant(level: RiskLevel): 'success' | 'warning' | 'danger' {
		switch (level) {
			case 'low':
				return 'success';
			case 'medium':
				return 'warning';
			case 'high':
			case 'critical':
				return 'danger';
			default:
				return 'warning';
		}
	}

	// トレンドのアイコン
	function getTrendIcon(trend: VelocityTrend): string {
		switch (trend) {
			case 'increasing':
				return '&#9650;'; // ▲
			case 'decreasing':
				return '&#9660;'; // ▼
			case 'stable':
				return '&#9644;'; // ─
			default:
				return '&#63;'; // ?
		}
	}

	// トレンドの色
	function getTrendColor(trend: VelocityTrend): string {
		switch (trend) {
			case 'increasing':
				return 'var(--status-good)';
			case 'decreasing':
				return 'var(--status-poor)';
			case 'stable':
				return 'var(--status-fair)';
			default:
				return 'var(--text-muted)';
		}
	}
</script>

<Panel title="Predictions" icon="&#128302;" loading={$predictionLoading} error={$predictionError}>
	<div class="prediction-content">
		{#if !$hasSufficientData}
			<div class="insufficient-data">
				<span class="icon">&#9432;</span>
				<span>Insufficient data for accurate predictions</span>
			</div>
		{/if}

		<div class="prediction-grid">
			<!-- 完了予測 -->
			{#if $completion}
				<div class="prediction-card">
					<div class="card-header">
						<span class="card-icon">&#128197;</span>
						<span class="card-title">Completion</span>
					</div>
					<div class="card-content">
						<div class="primary-value">{$completion.estimated_date || 'N/A'}</div>
						<div class="secondary-info">
							<span>{$completion.remaining_tasks} tasks remaining</span>
							<span class="confidence">
								{$completion.confidence_level}% confidence
							</span>
						</div>
					</div>
				</div>
			{/if}

			<!-- リスク分析 -->
			{#if $risk}
				<div class="prediction-card">
					<div class="card-header">
						<span class="card-icon">&#9888;</span>
						<span class="card-title">Risk</span>
						<Badge variant={getRiskVariant($risk.overall_level)} size="sm">
							{$risk.overall_level.toUpperCase()}
						</Badge>
					</div>
					<div class="card-content">
						<div class="primary-value">{$risk.score}/100</div>
						{#if $risk.factors.length > 0}
							<div class="risk-factors">
								{#each $risk.factors.slice(0, 2) as factor}
									<div class="risk-factor">
										<span class="factor-name">{factor.name}</span>
										<span class="factor-impact">Impact: {factor.impact}</span>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				</div>
			{/if}

			<!-- ベロシティ -->
			{#if $velocity}
				<div class="prediction-card">
					<div class="card-header">
						<span class="card-icon">&#128640;</span>
						<span class="card-title">Velocity</span>
						<span class="trend-icon" style="color: {getTrendColor($velocity.trend)}">
							{@html getTrendIcon($velocity.trend)}
						</span>
					</div>
					<div class="card-content">
						<div class="primary-value">{$velocity.weekly_average.toFixed(1)}</div>
						<div class="secondary-info">
							<span>tasks/week avg</span>
						</div>
						<div class="velocity-breakdown">
							<span>7d: {$velocity.last_7_days}</span>
							<span>14d: {$velocity.last_14_days}</span>
							<span>30d: {$velocity.last_30_days}</span>
						</div>
					</div>
				</div>
			{/if}
		</div>
	</div>
</Panel>

<style>
	.prediction-content {
		display: flex;
		flex-direction: column;
		gap: var(--spacing-md);
	}

	.insufficient-data {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		padding: var(--spacing-sm) var(--spacing-md);
		background-color: rgba(255, 204, 0, 0.1);
		border: 1px solid var(--status-fair);
		border-radius: var(--border-radius-sm);
		color: var(--status-fair);
		font-size: var(--font-size-sm);
	}

	.prediction-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: var(--spacing-md);
	}

	@media (max-width: 1024px) {
		.prediction-grid {
			grid-template-columns: 1fr;
		}
	}

	.prediction-card {
		background-color: var(--bg-secondary);
		border: 2px solid var(--border-metal);
		border-radius: var(--border-radius-md);
		padding: var(--spacing-md);
	}

	.card-header {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		margin-bottom: var(--spacing-sm);
	}

	.card-icon {
		font-size: var(--font-size-lg);
	}

	.card-title {
		font-size: var(--font-size-sm);
		font-weight: 600;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		flex: 1;
	}

	.trend-icon {
		font-size: var(--font-size-lg);
	}

	.card-content {
		display: flex;
		flex-direction: column;
		gap: var(--spacing-xs);
	}

	.primary-value {
		font-size: var(--font-size-xl);
		font-weight: 700;
		color: var(--accent-primary);
	}

	.secondary-info {
		display: flex;
		justify-content: space-between;
		font-size: var(--font-size-xs);
		color: var(--text-muted);
	}

	.confidence {
		color: var(--text-secondary);
	}

	.risk-factors {
		display: flex;
		flex-direction: column;
		gap: var(--spacing-xs);
		margin-top: var(--spacing-sm);
		padding-top: var(--spacing-sm);
		border-top: 1px solid var(--border-dark);
	}

	.risk-factor {
		display: flex;
		justify-content: space-between;
		font-size: var(--font-size-xs);
	}

	.factor-name {
		color: var(--text-secondary);
	}

	.factor-impact {
		color: var(--text-muted);
	}

	.velocity-breakdown {
		display: flex;
		gap: var(--spacing-md);
		margin-top: var(--spacing-sm);
		padding-top: var(--spacing-sm);
		border-top: 1px solid var(--border-dark);
		font-size: var(--font-size-xs);
		color: var(--text-muted);
	}
</style>
