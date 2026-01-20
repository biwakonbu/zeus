<script lang="ts">
	// EmptyState コンポーネント
	// データがない場合の統一的な表示
	import { Icon } from '$lib/components/ui';

	interface Props {
		icon?: string;
		title: string;
		description?: string;
		actionLabel?: string;
		onAction?: () => void;
	}

	let {
		icon = 'Inbox',
		title,
		description = '',
		actionLabel = '',
		onAction
	}: Props = $props();
</script>

<div class="empty-state">
	<div class="empty-icon">
		<Icon name={icon} size={48} />
	</div>
	<h3 class="empty-title">{title}</h3>
	{#if description}
		<p class="empty-description">{description}</p>
	{/if}
	{#if actionLabel && onAction}
		<button class="empty-action" onclick={onAction}>
			{actionLabel}
		</button>
	{/if}
</div>

<style>
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--spacing-xxl, 48px) var(--spacing-lg, 24px);
		text-align: center;
		min-height: 200px;
	}

	.empty-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 80px;
		height: 80px;
		margin-bottom: var(--spacing-lg, 24px);
		background: var(--bg-secondary, #252525);
		border: 2px solid var(--border-metal, #3a3a3a);
		border-radius: 50%;
		color: var(--text-muted, #888);
		opacity: 0.6;
	}

	.empty-title {
		margin: 0 0 var(--spacing-sm, 8px) 0;
		font-size: var(--font-size-lg, 18px);
		font-weight: 600;
		color: var(--text-primary, #e0e0e0);
	}

	.empty-description {
		margin: 0 0 var(--spacing-lg, 24px) 0;
		font-size: var(--font-size-sm, 13px);
		color: var(--text-muted, #888);
		max-width: 300px;
		line-height: 1.5;
	}

	.empty-action {
		padding: var(--spacing-sm, 8px) var(--spacing-lg, 24px);
		background: var(--bg-secondary, #252525);
		border: 2px solid var(--accent-primary, #f59e0b);
		border-radius: var(--border-radius-sm, 4px);
		color: var(--accent-primary, #f59e0b);
		font-family: inherit;
		font-size: var(--font-size-sm, 13px);
		font-weight: 500;
		cursor: pointer;
		transition: background-color 0.15s ease, color 0.15s ease;
	}

	.empty-action:hover {
		background: var(--accent-primary, #f59e0b);
		color: var(--bg-primary, #1a1a1a);
	}

	.empty-action:focus-visible {
		outline: var(--focus-ring-width, 2px) solid var(--focus-ring-color, #f59e0b);
		outline-offset: var(--focus-ring-offset, 2px);
	}

	@media (prefers-reduced-motion: reduce) {
		.empty-action {
			transition: none;
		}
	}
</style>
