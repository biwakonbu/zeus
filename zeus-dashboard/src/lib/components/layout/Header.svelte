<script lang="ts">
	import type { ConnectionState } from '$lib/types/api';

	interface Props {
		connectionState?: ConnectionState;
	}

	let { connectionState = 'disconnected' }: Props = $props();

	// 接続状態のラベル
	function getConnectionLabel(state: ConnectionState): string {
		switch (state) {
			case 'connected':
				return 'Connected';
			case 'connecting':
				return 'Connecting...';
			case 'disconnected':
				return 'Disconnected';
			default:
				return 'Unknown';
		}
	}
</script>

<header class="header">
	<div class="header-content">
		<div class="logo">
			<span class="logo-icon">&#9889;</span>
			<h1 class="logo-text">ZEUS</h1>
			<span class="logo-subtitle">Dashboard</span>
		</div>

		<div class="header-status">
			<div class="connection-status">
				<span class="connection-indicator {connectionState}"></span>
				<span class="connection-label">{getConnectionLabel(connectionState)}</span>
			</div>
		</div>
	</div>
</header>

<style>
	.header {
		background-color: var(--bg-secondary);
		border-bottom: 2px solid var(--border-metal);
		padding: var(--spacing-md) var(--spacing-xl);
		position: sticky;
		top: 0;
		z-index: 100;
	}

	.header-content {
		display: flex;
		align-items: center;
		justify-content: space-between;
		max-width: 1600px;
		margin: 0 auto;
	}

	.logo {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
	}

	.logo-icon {
		font-size: 1.5rem;
		color: var(--accent-primary);
	}

	.logo-text {
		font-size: var(--font-size-xl);
		font-weight: 700;
		color: var(--accent-primary);
		letter-spacing: 0.1em;
		text-transform: uppercase;
	}

	.logo-subtitle {
		font-size: var(--font-size-sm);
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding-left: var(--spacing-sm);
		border-left: 1px solid var(--border-metal);
	}

	.header-status {
		display: flex;
		align-items: center;
		gap: var(--spacing-lg);
	}

	.connection-status {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		padding: var(--spacing-xs) var(--spacing-sm);
		background-color: var(--bg-panel);
		border: 1px solid var(--border-metal);
		border-radius: var(--border-radius-sm);
	}

	.connection-label {
		font-size: var(--font-size-xs);
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
</style>
