<script lang="ts">
	import Panel from '$lib/components/ui/Panel.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import MermaidGraph from '$lib/components/graph/MermaidGraph.svelte';
	import { mermaidCode, graphStats, hasCycles, hasIsolated, cycles, isolated, graphLoading, graphError } from '$lib/stores/graph';
</script>

<Panel title="Dependency Graph" icon="&#128279;" loading={$graphLoading} error={$graphError}>
	{#snippet headerRight()}
		<div class="graph-badges">
			{#if $hasCycles}
				<Badge variant="danger" size="sm">
					{$cycles.length} Cycles
				</Badge>
			{/if}
			{#if $hasIsolated}
				<Badge variant="warning" size="sm">
					{$isolated.length} Isolated
				</Badge>
			{/if}
		</div>
	{/snippet}

	<div class="graph-content">
		{#if $graphStats}
			<div class="graph-stats">
				<div class="stat-item">
					<span class="stat-label">Nodes</span>
					<span class="stat-value">{$graphStats.total_nodes}</span>
				</div>
				<div class="stat-item">
					<span class="stat-label">Connected</span>
					<span class="stat-value">{$graphStats.with_dependencies}</span>
				</div>
				<div class="stat-item">
					<span class="stat-label">Max Depth</span>
					<span class="stat-value">{$graphStats.max_depth}</span>
				</div>
			</div>
		{/if}

		<div class="graph-container">
			<MermaidGraph code={$mermaidCode} />
		</div>

		{#if $hasCycles}
			<div class="graph-warning">
				<span class="warning-icon">&#9888;</span>
				<span class="warning-text">
					Circular dependencies detected: {$cycles.map(c => c.join(' -> ')).join('; ')}
				</span>
			</div>
		{/if}
	</div>
</Panel>

<style>
	.graph-badges {
		display: flex;
		gap: var(--spacing-xs);
	}

	.graph-content {
		display: flex;
		flex-direction: column;
		gap: var(--spacing-md);
	}

	.graph-stats {
		display: flex;
		gap: var(--spacing-lg);
		padding: var(--spacing-sm) var(--spacing-md);
		background-color: var(--bg-secondary);
		border: 1px solid var(--border-metal);
		border-radius: var(--border-radius-sm);
	}

	.stat-item {
		display: flex;
		align-items: center;
		gap: var(--spacing-xs);
	}

	.stat-label {
		font-size: var(--font-size-xs);
		color: var(--text-muted);
		text-transform: uppercase;
	}

	.stat-value {
		font-size: var(--font-size-sm);
		font-weight: 600;
		color: var(--accent-primary);
	}

	.graph-container {
		background-color: var(--bg-secondary);
		border: 2px solid var(--border-metal);
		border-radius: var(--border-radius-md);
		min-height: 300px;
		overflow: hidden;
	}

	.graph-warning {
		display: flex;
		align-items: flex-start;
		gap: var(--spacing-sm);
		padding: var(--spacing-sm) var(--spacing-md);
		background-color: rgba(238, 68, 68, 0.1);
		border: 1px solid var(--status-poor);
		border-radius: var(--border-radius-sm);
		color: var(--status-poor);
		font-size: var(--font-size-sm);
	}

	.warning-icon {
		flex-shrink: 0;
	}

	.warning-text {
		word-break: break-word;
	}
</style>
