<script lang="ts">
	import { onMount } from 'svelte';

	let currentTime = $state('');

	onMount(() => {
		function updateTime() {
			const now = new Date();
			currentTime = now.toLocaleTimeString('ja-JP', {
				hour: '2-digit',
				minute: '2-digit',
				second: '2-digit'
			});
		}

		updateTime();
		const interval = setInterval(updateTime, 1000);

		return () => clearInterval(interval);
	});
</script>

<footer class="footer">
	<div class="footer-content">
		<span class="footer-text">Zeus</span>
		<span class="footer-separator">|</span>
		<span class="footer-text text-muted">v0.1.0</span>
		<span class="footer-separator">|</span>
		<span class="footer-time">{currentTime}</span>
	</div>
</footer>

<style>
	.footer {
		background-color: var(--bg-secondary);
		border-top: 1px solid var(--border-metal);
		padding: var(--spacing-xs) var(--spacing-md);
		margin-top: auto;
	}

	.footer-content {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--spacing-sm);
	}

	.footer-text {
		font-size: var(--font-size-xs);
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.footer-separator {
		color: var(--border-metal);
		font-size: var(--font-size-xs);
	}

	.footer-time {
		font-size: var(--font-size-xs);
		color: var(--accent-primary);
		font-variant-numeric: tabular-nums;
	}

	.text-muted {
		color: var(--text-muted);
	}
</style>
