<script lang="ts">
	// 現在時刻を表示（1秒ごとに更新）
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
		<div class="footer-left">
			<span class="footer-text">Zeus Project Management System</span>
			<span class="footer-separator">|</span>
			<span class="footer-text text-muted">v0.1.0</span>
		</div>

		<div class="footer-right">
			<span class="footer-time">{currentTime}</span>
		</div>
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
		justify-content: space-between;
		max-width: 1600px;
		margin: 0 auto;
	}

	.footer-left {
		display: flex;
		align-items: center;
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
	}

	.footer-right {
		display: flex;
		align-items: center;
		gap: var(--spacing-md);
	}

	.footer-time {
		font-size: var(--font-size-sm);
		color: var(--accent-primary);
		font-variant-numeric: tabular-nums;
	}

	.text-muted {
		color: var(--text-muted);
	}
</style>
