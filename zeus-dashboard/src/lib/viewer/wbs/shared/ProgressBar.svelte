<script lang="ts">
	// Factorio 風プログレスバー
	// 進捗率に応じて色が変化する
	interface Props {
		progress: number; // 0-100
		size?: 'sm' | 'md' | 'lg';
		showLabel?: boolean;
	}
	let { progress, size = 'md', showLabel = false }: Props = $props();

	const progressLevel = $derived(progress >= 70 ? 'high' : progress >= 40 ? 'mid' : 'low');
	const clampedProgress = $derived(Math.min(100, Math.max(0, progress)));
</script>

<div class="progress-bar progress-bar--{size}">
	<div
		class="progress-bar__fill progress-bar__fill--{progressLevel}"
		style="width: {clampedProgress}%"
	></div>
	{#if showLabel}
		<span class="progress-bar__label">{progress}%</span>
	{/if}
</div>

<style>
	.progress-bar {
		position: relative;
		background: var(--bg-secondary, #242424);
		border: 1px solid var(--border-metal, #4a4a4a);
		border-radius: 2px;
		overflow: hidden;
	}

	.progress-bar--sm {
		height: 6px;
	}
	.progress-bar--md {
		height: 10px;
	}
	.progress-bar--lg {
		height: 16px;
	}

	.progress-bar__fill {
		height: 100%;
		transition:
			width 0.3s ease,
			background-color 0.3s ease;
	}

	.progress-bar__fill--low {
		background: var(--status-poor, #ee4444);
	}
	.progress-bar__fill--mid {
		background: var(--status-fair, #ffcc00);
	}
	.progress-bar__fill--high {
		background: var(--status-good, #44cc44);
	}

	.progress-bar__label {
		position: absolute;
		right: 4px;
		top: 50%;
		transform: translateY(-50%);
		font-size: 10px;
		font-weight: 600;
		color: var(--text-primary, #ffffff);
		text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
	}
</style>
