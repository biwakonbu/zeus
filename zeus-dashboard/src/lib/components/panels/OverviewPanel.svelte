<script lang="ts">
	import Panel from '$lib/components/ui/Panel.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import ProgressBar from '$lib/components/ui/ProgressBar.svelte';
	import { projectInfo, projectState, health, progressPercent, statusLoading, statusError } from '$lib/stores/status';
	import type { HealthStatus } from '$lib/types/api';

	// 健全性のバッジバリアント
	function getHealthVariant(health: HealthStatus | null): 'success' | 'warning' | 'danger' | 'muted' {
		switch (health) {
			case 'good':
				return 'success';
			case 'fair':
				return 'warning';
			case 'poor':
				return 'danger';
			default:
				return 'muted';
		}
	}

	// 健全性のラベル
	function getHealthLabel(health: HealthStatus | null): string {
		switch (health) {
			case 'good':
				return 'GOOD';
			case 'fair':
				return 'FAIR';
			case 'poor':
				return 'POOR';
			default:
				return 'UNKNOWN';
		}
	}
</script>

<Panel title="Overview" icon="&#128200;" loading={$statusLoading} error={$statusError}>
	{#snippet headerRight()}
		{#if $health}
			<Badge variant={getHealthVariant($health)}>
				{getHealthLabel($health)}
			</Badge>
		{/if}
	{/snippet}

	<div class="overview-content">
		{#if $projectInfo}
			<div class="project-info">
				<h3 class="project-name">{$projectInfo.name}</h3>
				{#if $projectInfo.description}
					<p class="project-description">{$projectInfo.description}</p>
				{/if}
				<div class="project-meta">
					<span class="meta-item">
						<span class="meta-label">ID:</span>
						<span class="meta-value">{$projectInfo.id.substring(0, 8)}...</span>
					</span>
					{#if $projectInfo.start_date}
						<span class="meta-item">
							<span class="meta-label">Started:</span>
							<span class="meta-value">{$projectInfo.start_date}</span>
						</span>
					{/if}
				</div>
			</div>

			<div class="progress-section">
				<div class="progress-header">
					<span class="progress-title">Progress</span>
					<span class="progress-stats">
						{$projectState?.summary.completed ?? 0} / {$projectState?.summary.total_tasks ?? 0} tasks
					</span>
				</div>
				<ProgressBar value={$progressPercent} size="lg" />
			</div>
		{:else}
			<div class="no-data">No project data available</div>
		{/if}
	</div>
</Panel>

<style>
	.overview-content {
		display: flex;
		flex-direction: column;
		gap: var(--spacing-lg);
	}

	.project-info {
		display: flex;
		flex-direction: column;
		gap: var(--spacing-sm);
	}

	.project-name {
		font-size: var(--font-size-xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
	}

	.project-description {
		font-size: var(--font-size-sm);
		color: var(--text-secondary);
		margin: 0;
		line-height: 1.5;
	}

	.project-meta {
		display: flex;
		gap: var(--spacing-lg);
		margin-top: var(--spacing-xs);
	}

	.meta-item {
		display: flex;
		align-items: center;
		gap: var(--spacing-xs);
		font-size: var(--font-size-xs);
	}

	.meta-label {
		color: var(--text-muted);
		text-transform: uppercase;
	}

	.meta-value {
		color: var(--text-secondary);
		font-family: var(--font-family);
	}

	.progress-section {
		display: flex;
		flex-direction: column;
		gap: var(--spacing-sm);
	}

	.progress-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.progress-title {
		font-size: var(--font-size-sm);
		font-weight: 600;
		color: var(--text-primary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.progress-stats {
		font-size: var(--font-size-sm);
		color: var(--text-secondary);
	}

	.no-data {
		color: var(--text-muted);
		font-style: italic;
		text-align: center;
		padding: var(--spacing-lg);
	}
</style>
