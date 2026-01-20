<script lang="ts">
	// 個別の Toast コンポーネント
	import { Icon } from '$lib/components/ui';
	import type { ToastType } from '$lib/stores/toast';

	interface Props {
		id: string;
		type: ToastType;
		message: string;
		dismissible: boolean;
		onDismiss: (id: string) => void;
	}

	let { id, type, message, dismissible, onDismiss }: Props = $props();

	// タイプ別アイコン
	const iconMap: Record<ToastType, string> = {
		info: 'Info',
		success: 'CheckCircle',
		warning: 'AlertTriangle',
		error: 'XCircle'
	};

	function handleDismiss() {
		onDismiss(id);
	}
</script>

<div
	class="toast toast-{type}"
	role="alert"
	aria-live={type === 'error' ? 'assertive' : 'polite'}
>
	<span class="toast-icon">
		<Icon name={iconMap[type]} size={18} />
	</span>
	<span class="toast-message">{message}</span>
	{#if dismissible}
		<button class="toast-dismiss" onclick={handleDismiss} aria-label="閉じる">
			<Icon name="X" size={14} />
		</button>
	{/if}
</div>

<style>
	.toast {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm, 8px);
		padding: var(--spacing-sm, 8px) var(--spacing-md, 16px);
		background: var(--bg-panel, #2a2a2a);
		border: 2px solid var(--border-metal, #3a3a3a);
		border-radius: var(--border-radius-sm, 4px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
		font-size: var(--font-size-sm, 13px);
		min-width: 280px;
		max-width: 400px;
		animation: toast-enter 0.2s ease-out;
	}

	@keyframes toast-enter {
		from {
			opacity: 0;
			transform: translateY(-8px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	/* タイプ別スタイル */
	.toast-info {
		border-left: 4px solid var(--status-info, #3b82f6);
	}

	.toast-info .toast-icon {
		color: var(--status-info, #3b82f6);
	}

	.toast-success {
		border-left: 4px solid var(--status-good, #22c55e);
	}

	.toast-success .toast-icon {
		color: var(--status-good, #22c55e);
	}

	.toast-warning {
		border-left: 4px solid var(--status-warning, #eab308);
	}

	.toast-warning .toast-icon {
		color: var(--status-warning, #eab308);
	}

	.toast-error {
		border-left: 4px solid var(--status-poor, #ef4444);
	}

	.toast-error .toast-icon {
		color: var(--status-poor, #ef4444);
	}

	.toast-icon {
		display: flex;
		align-items: center;
		flex-shrink: 0;
	}

	.toast-message {
		flex: 1;
		color: var(--text-primary, #e0e0e0);
		line-height: 1.4;
	}

	.toast-dismiss {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 4px;
		background: transparent;
		border: none;
		color: var(--text-muted, #888);
		cursor: pointer;
		border-radius: var(--border-radius-sm, 4px);
		transition: color 0.15s ease, background-color 0.15s ease;
	}

	.toast-dismiss:hover {
		color: var(--text-primary, #e0e0e0);
		background: var(--bg-hover, #3a3a3a);
	}

	.toast-dismiss:focus-visible {
		outline: var(--focus-ring-width, 2px) solid var(--focus-ring-color, #f59e0b);
		outline-offset: 1px;
	}

	@media (prefers-reduced-motion: reduce) {
		.toast {
			animation: none;
		}

		.toast-dismiss {
			transition: none;
		}
	}
</style>
