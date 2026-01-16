<script lang="ts">
	type BadgeVariant = 'default' | 'success' | 'warning' | 'danger' | 'info' | 'muted';

	interface Props {
		variant?: BadgeVariant;
		size?: 'sm' | 'md';
		children: import('svelte').Snippet;
	}

	let { variant = 'default', size = 'md', children }: Props = $props();
</script>

<span class="badge badge-{variant} badge-{size}">
	{@render children()}
</span>

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		font-family: var(--font-family);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		border-radius: var(--border-radius-sm);
		white-space: nowrap;
	}

	.badge-sm {
		font-size: var(--font-size-xs);
		padding: 2px 6px;
	}

	.badge-md {
		font-size: var(--font-size-xs);
		padding: 4px 8px;
	}

	/* バリアント */
	.badge-default {
		background-color: var(--bg-secondary);
		color: var(--text-secondary);
		border: 1px solid var(--border-metal);
	}

	.badge-success {
		background-color: rgba(68, 204, 68, 0.2);
		color: var(--status-good);
		border: 1px solid var(--status-good);
	}

	.badge-warning {
		background-color: rgba(255, 204, 0, 0.2);
		color: var(--status-fair);
		border: 1px solid var(--status-fair);
	}

	.badge-danger {
		background-color: rgba(238, 68, 68, 0.2);
		color: var(--status-poor);
		border: 1px solid var(--status-poor);
	}

	.badge-info {
		background-color: rgba(68, 136, 255, 0.2);
		color: var(--status-info);
		border: 1px solid var(--status-info);
	}

	.badge-muted {
		background-color: rgba(136, 136, 136, 0.2);
		color: var(--text-muted);
		border: 1px solid var(--text-muted);
	}
</style>
