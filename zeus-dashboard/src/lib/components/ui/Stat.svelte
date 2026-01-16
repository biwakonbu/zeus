<script lang="ts">
	type StatVariant = 'default' | 'success' | 'warning' | 'danger' | 'info' | 'muted';

	interface Props {
		label: string;
		value: string | number;
		icon?: string;
		variant?: StatVariant;
		subtitle?: string;
	}

	let { label, value, icon = '', variant = 'default', subtitle = '' }: Props = $props();
</script>

<div class="stat stat-{variant}">
	<div class="stat-header">
		{#if icon}
			<span class="stat-icon">{icon}</span>
		{/if}
		<span class="stat-label">{label}</span>
	</div>
	<div class="stat-value">{value}</div>
	{#if subtitle}
		<div class="stat-subtitle">{subtitle}</div>
	{/if}
</div>

<style>
	.stat {
		display: flex;
		flex-direction: column;
		padding: var(--spacing-md);
		background-color: var(--bg-secondary);
		border: 2px solid var(--border-metal);
		border-radius: var(--border-radius-md);
		transition: border-color var(--transition-fast);
	}

	.stat:hover {
		border-color: var(--border-highlight);
	}

	.stat-header {
		display: flex;
		align-items: center;
		gap: var(--spacing-xs);
		margin-bottom: var(--spacing-xs);
	}

	.stat-icon {
		font-size: var(--font-size-sm);
	}

	.stat-label {
		font-size: var(--font-size-xs);
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.stat-value {
		font-size: var(--font-size-2xl);
		font-weight: 700;
		font-variant-numeric: tabular-nums;
	}

	.stat-subtitle {
		font-size: var(--font-size-xs);
		color: var(--text-muted);
		margin-top: var(--spacing-xs);
	}

	/* バリアント */
	.stat-default .stat-value {
		color: var(--text-primary);
	}

	.stat-success .stat-value {
		color: var(--status-good);
	}

	.stat-warning .stat-value {
		color: var(--status-fair);
	}

	.stat-danger .stat-value {
		color: var(--status-poor);
	}

	.stat-info .stat-value {
		color: var(--status-info);
	}

	.stat-muted .stat-value {
		color: var(--text-muted);
	}
</style>
