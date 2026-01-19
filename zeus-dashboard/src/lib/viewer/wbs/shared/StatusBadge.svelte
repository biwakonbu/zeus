<script lang="ts">
	// ステータス表示バッジ
	// Factorio 風のデザインでステータスを視覚的に表示
	interface Props {
		status: 'good' | 'fair' | 'poor' | string;
		label?: string;
	}
	let { status, label }: Props = $props();

	const displayLabel = $derived(label || status.toUpperCase());
</script>

<span class="status-badge status-badge--{status}">
	{displayLabel}
</span>

<style>
	.status-badge {
		display: inline-block;
		padding: 2px 8px;
		border-radius: 2px;
		font-size: 11px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		border: 1px solid transparent;
	}

	.status-badge--good {
		background: var(--status-good, #44cc44);
		color: var(--bg-primary, #1a1a1a);
		border-color: rgba(68, 204, 68, 0.5);
	}

	.status-badge--fair {
		background: var(--status-fair, #ffcc00);
		color: var(--bg-primary, #1a1a1a);
		border-color: rgba(255, 204, 0, 0.5);
	}

	.status-badge--poor {
		background: var(--status-poor, #ee4444);
		color: var(--text-primary, #ffffff);
		border-color: rgba(238, 68, 68, 0.5);
	}

	/* 追加ステータス（汎用） */
	.status-badge--info {
		background: var(--status-info, #4488ff);
		color: var(--text-primary, #ffffff);
		border-color: rgba(68, 136, 255, 0.5);
	}
</style>
