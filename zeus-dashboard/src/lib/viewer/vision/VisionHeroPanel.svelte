<script lang="ts">
	import type { Vision } from '$lib/types/api';
	import { getVisionStatusColor, getVisionStatusLabel } from '$lib/utils/status';

	interface Props {
		vision: Vision;
	}

	let { vision }: Props = $props();
</script>

<div class="vision-hero">
	<div class="vision-header">
		<h2 class="vision-title">{vision.title}</h2>
		<span class="status-badge" style="--badge-color: {getVisionStatusColor(vision.status)}">
			{getVisionStatusLabel(vision.status)}
		</span>
	</div>

	<p class="vision-statement">{vision.statement}</p>

	{#if vision.success_criteria && vision.success_criteria.length > 0}
		<div class="criteria-section">
			<h3 class="criteria-heading">Success Criteria</h3>
			<ul class="criteria-list">
				{#each vision.success_criteria as criterion, i (i)}
					<li class="criteria-item">
						<span class="criteria-check">&#9633;</span>
						<span>{criterion}</span>
					</li>
				{/each}
			</ul>
		</div>
	{/if}
</div>

<style>
	.vision-hero {
		padding: 24px;
		background: linear-gradient(135deg, rgba(245, 158, 11, 0.08) 0%, rgba(245, 158, 11, 0.02) 100%);
		border: 1px solid rgba(245, 158, 11, 0.2);
		border-radius: 8px;
	}

	.vision-header {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 12px;
	}

	.vision-title {
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
	}

	.status-badge {
		display: inline-flex;
		align-items: center;
		padding: 2px 10px;
		font-size: 0.6875rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		border-radius: 12px;
		background: color-mix(in srgb, var(--badge-color) 20%, transparent);
		color: var(--badge-color);
		white-space: nowrap;
	}

	.vision-statement {
		font-size: 0.9375rem;
		line-height: 1.6;
		color: var(--text-secondary);
		margin: 0 0 16px;
		white-space: pre-line;
	}

	.criteria-section {
		border-top: 1px solid var(--border-metal);
		padding-top: 12px;
	}

	.criteria-heading {
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-muted);
		margin: 0 0 8px;
	}

	.criteria-list {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.criteria-item {
		display: flex;
		align-items: flex-start;
		gap: 8px;
		font-size: 0.8125rem;
		color: var(--text-secondary);
		line-height: 1.4;
	}

	.criteria-check {
		flex-shrink: 0;
		color: var(--accent-primary);
		font-size: 0.875rem;
	}
</style>
