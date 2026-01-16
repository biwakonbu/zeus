<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		title: string;
		icon?: string;
		loading?: boolean;
		error?: string | null;
		children: Snippet;
		headerRight?: Snippet;
	}

	let { title, icon = '', loading = false, error = null, children, headerRight }: Props = $props();
</script>

<div class="panel metal-frame">
	<div class="panel-header">
		<div class="panel-title-wrapper">
			{#if icon}
				<span class="panel-icon">{icon}</span>
			{/if}
			<h2 class="panel-title">{title}</h2>
		</div>
		{#if headerRight}
			<div class="panel-header-right">
				{@render headerRight()}
			</div>
		{/if}
	</div>

	<div class="panel-body">
		{#if loading}
			<div class="panel-loading">
				<div class="loading-spinner"></div>
				<span>Loading...</span>
			</div>
		{:else if error}
			<div class="panel-error">
				<span class="error-icon">&#9888;</span>
				<span class="error-message">{error}</span>
			</div>
		{:else}
			{@render children()}
		{/if}
	</div>
</div>

<style>
	.panel {
		background-color: var(--bg-panel);
		border: 2px solid var(--border-metal);
		border-radius: var(--border-radius-md);
		padding: var(--spacing-lg);
		height: 100%;
	}

	.panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: var(--spacing-md);
		padding-bottom: var(--spacing-sm);
		border-bottom: 1px solid var(--border-dark);
	}

	.panel-title-wrapper {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
	}

	.panel-icon {
		font-size: var(--font-size-lg);
		color: var(--accent-primary);
	}

	.panel-title {
		font-size: var(--font-size-lg);
		font-weight: 600;
		color: var(--accent-primary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin: 0;
	}

	.panel-header-right {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
	}

	.panel-body {
		min-height: 60px;
	}

	.panel-loading {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--spacing-sm);
		color: var(--text-secondary);
		padding: var(--spacing-lg);
	}

	.loading-spinner {
		width: 20px;
		height: 20px;
		border: 2px solid var(--border-metal);
		border-top-color: var(--accent-primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.panel-error {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		color: var(--status-poor);
		padding: var(--spacing-md);
		background-color: rgba(238, 68, 68, 0.1);
		border: 1px solid var(--status-poor);
		border-radius: var(--border-radius-sm);
	}

	.error-icon {
		font-size: var(--font-size-lg);
	}

	.error-message {
		font-size: var(--font-size-sm);
	}
</style>
