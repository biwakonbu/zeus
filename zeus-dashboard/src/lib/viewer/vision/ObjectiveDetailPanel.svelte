<script lang="ts">
	import type { Objective, UseCaseItem } from '$lib/types/api';
	import { navigateToEntity } from '$lib/stores/view';
	import { Icon } from '$lib/components/ui';
	import { getObjectiveStatusColor, getObjectiveStatusLabel, formatDate } from '$lib/utils/status';

	interface Props {
		objective: Objective;
		relatedUseCases: UseCaseItem[];
	}

	let { objective, relatedUseCases }: Props = $props();

	// UseCase Viewer へ遷移
	function handleUseCaseClick(usecaseId: string) {
		navigateToEntity('usecase', 'usecase', usecaseId);
	}
</script>

<div class="detail-panel">
	<!-- ヘッダー -->
	<div class="detail-header">
		<span class="detail-id">{objective.id}</span>
		<span class="status-badge" style="--badge-color: {getObjectiveStatusColor(objective.status)}">
			{getObjectiveStatusLabel(objective.status)}
		</span>
	</div>

	<h3 class="detail-title">{objective.title}</h3>

	<!-- Description -->
	{#if objective.description}
		<div class="detail-section">
			<h4 class="section-label">Description</h4>
			<p class="detail-description">{objective.description}</p>
		</div>
	{/if}

	<!-- Goals -->
	{#if objective.goals && objective.goals.length > 0}
		<div class="detail-section">
			<h4 class="section-label">Goals</h4>
			<ul class="goals-list">
				{#each objective.goals as goal, i (i)}
					<li class="goal-item">{goal}</li>
				{/each}
			</ul>
		</div>
	{/if}

	<!-- Owner -->
	{#if objective.owner}
		<div class="detail-section">
			<h4 class="section-label">Owner</h4>
			<span class="detail-value">{objective.owner}</span>
		</div>
	{/if}

	<!-- Tags -->
	{#if objective.tags && objective.tags.length > 0}
		<div class="detail-section">
			<h4 class="section-label">Tags</h4>
			<div class="tags-list">
				{#each objective.tags as tag (tag)}
					<span class="tag">{tag}</span>
				{/each}
			</div>
		</div>
	{/if}

	<!-- UseCase 図を表示 -->
	{#if relatedUseCases.length > 0}
		<div class="detail-section">
			<button
				class="usecase-diagram-link"
				onclick={() => navigateToEntity('usecase', 'objective', objective.id)}
			>
				<Icon name="ClipboardList" size={14} />
				<span>UseCase 図を表示</span>
			</button>
		</div>
	{/if}

	<!-- 関連 UseCase -->
	<div class="detail-section">
		<h4 class="section-label">UseCase ({relatedUseCases.length})</h4>
		{#if relatedUseCases.length > 0}
			<div class="usecase-list">
				{#each relatedUseCases as uc (uc.id)}
					<button class="usecase-link" onclick={() => handleUseCaseClick(uc.id)}>
						<span class="usecase-id">{uc.id}</span>
						<span class="usecase-title">{uc.title}</span>
						<span class="usecase-nav">
							<Icon name="ExternalLink" size={12} />
						</span>
					</button>
				{/each}
			</div>
		{:else}
			<p class="empty-text">UseCase が紐付けられていません</p>
		{/if}
	</div>

	<!-- メタデータ -->
	<div class="detail-section metadata">
		<span class="meta-item">Created: {formatDate(objective.created_at)}</span>
		<span class="meta-item">Updated: {formatDate(objective.updated_at)}</span>
	</div>
</div>

<style>
	.detail-panel {
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.detail-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.detail-id {
		font-family: monospace;
		font-size: 0.75rem;
		color: var(--text-muted);
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
	}

	.detail-title {
		font-size: 1rem;
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
		line-height: 1.3;
	}

	.detail-section {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.section-label {
		font-size: 0.6875rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-muted);
		margin: 0;
	}

	.detail-description {
		font-size: 0.8125rem;
		line-height: 1.5;
		color: var(--text-secondary);
		margin: 0;
		white-space: pre-line;
	}

	.detail-value {
		font-size: 0.8125rem;
		color: var(--text-secondary);
	}

	.goals-list {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.goal-item {
		font-size: 0.8125rem;
		color: var(--text-secondary);
		line-height: 1.4;
		padding-left: 12px;
		position: relative;
	}

	.goal-item::before {
		content: '-';
		position: absolute;
		left: 0;
		color: var(--accent-primary);
	}

	.tags-list {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
	}

	.tag {
		font-size: 0.6875rem;
		padding: 2px 8px;
		background: var(--bg-secondary);
		color: var(--text-muted);
		border-radius: 8px;
	}

	.usecase-diagram-link {
		display: flex;
		align-items: center;
		gap: 6px;
		width: 100%;
		padding: 8px 12px;
		background: rgba(245, 158, 11, 0.1);
		border: 1px solid rgba(245, 158, 11, 0.3);
		border-radius: 6px;
		color: var(--accent-primary);
		font-family: inherit;
		font-size: 0.8125rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.usecase-diagram-link:hover {
		background: rgba(245, 158, 11, 0.2);
		border-color: var(--accent-primary);
	}

	.usecase-list {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.usecase-link {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 8px;
		background: var(--bg-primary);
		border: 1px solid var(--border-metal);
		border-radius: 4px;
		cursor: pointer;
		font-family: inherit;
		color: var(--text-primary);
		text-align: left;
		transition: all 0.15s ease;
		width: 100%;
	}

	.usecase-link:hover {
		border-color: var(--accent-primary);
		background: rgba(245, 158, 11, 0.05);
	}

	.usecase-id {
		font-family: monospace;
		font-size: 0.6875rem;
		color: var(--text-muted);
		flex-shrink: 0;
	}

	.usecase-title {
		font-size: 0.8125rem;
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.usecase-nav {
		flex-shrink: 0;
		color: var(--text-muted);
		display: flex;
		align-items: center;
	}

	.empty-text {
		font-size: 0.8125rem;
		color: var(--text-muted);
		margin: 0;
		font-style: italic;
	}

	.metadata {
		border-top: 1px solid var(--border-metal);
		padding-top: 8px;
		flex-direction: row;
		gap: 16px;
	}

	.meta-item {
		font-size: 0.6875rem;
		color: var(--text-muted);
	}
</style>
