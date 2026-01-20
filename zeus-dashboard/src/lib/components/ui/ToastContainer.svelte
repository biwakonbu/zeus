<script lang="ts">
	// Toast コンテナ
	// 全てのトースト通知を画面右上に表示
	import { toastStore } from '$lib/stores/toast';
	import Toast from './Toast.svelte';

	function handleDismiss(id: string) {
		toastStore.remove(id);
	}
</script>

<div class="toast-container" aria-label="通知" role="region">
	{#each $toastStore.toasts as toast (toast.id)}
		<Toast
			id={toast.id}
			type={toast.type}
			message={toast.message}
			dismissible={toast.dismissible}
			onDismiss={handleDismiss}
		/>
	{/each}
</div>

<style>
	.toast-container {
		position: fixed;
		top: var(--spacing-lg, 24px);
		right: var(--spacing-lg, 24px);
		z-index: 9999;
		display: flex;
		flex-direction: column;
		gap: var(--spacing-sm, 8px);
		pointer-events: none;
	}

	.toast-container > :global(*) {
		pointer-events: auto;
	}

	/* モバイル対応 */
	@media (max-width: 768px) {
		.toast-container {
			top: auto;
			bottom: var(--spacing-lg, 24px);
			left: var(--spacing-md, 16px);
			right: var(--spacing-md, 16px);
		}
	}
</style>
