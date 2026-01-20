<script lang="ts">
	// コンテキストメニューコンポーネント
	// 右クリックで表示されるメニュー（最大8項目）
	import { Icon } from '$lib/components/ui';

	export interface ContextMenuItem {
		id: string;
		label: string;
		icon?: string;
		shortcut?: string;
		disabled?: boolean;
		danger?: boolean;
		separator?: boolean;
	}

	interface Props {
		items: ContextMenuItem[];
		x: number;
		y: number;
		onSelect: (id: string) => void;
		onClose: () => void;
	}

	let { items, x, y, onSelect, onClose }: Props = $props();

	// 画面外に出ないように位置を調整
	let menuRef: HTMLDivElement | null = $state(null);
	let adjustedX = $derived.by(() => {
		if (!menuRef) return x;
		const rect = menuRef.getBoundingClientRect();
		const viewportWidth = window.innerWidth;
		if (x + rect.width > viewportWidth - 16) {
			return viewportWidth - rect.width - 16;
		}
		return x;
	});
	let adjustedY = $derived.by(() => {
		if (!menuRef) return y;
		const rect = menuRef.getBoundingClientRect();
		const viewportHeight = window.innerHeight;
		if (y + rect.height > viewportHeight - 16) {
			return viewportHeight - rect.height - 16;
		}
		return y;
	});

	function handleItemClick(item: ContextMenuItem) {
		if (item.disabled || item.separator) return;
		onSelect(item.id);
		onClose();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			onClose();
		}
	}

	function handleBackdropClick() {
		onClose();
	}

	function handleBackdropContextMenu(event: MouseEvent) {
		event.preventDefault();
		onClose();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div
	class="context-menu-backdrop"
	onclick={handleBackdropClick}
	oncontextmenu={handleBackdropContextMenu}
></div>

<div
	class="context-menu"
	bind:this={menuRef}
	style="left: {adjustedX}px; top: {adjustedY}px;"
	role="menu"
	aria-label="コンテキストメニュー"
>
	{#each items.slice(0, 8) as item (item.id)}
		{#if item.separator}
			<div class="menu-separator" role="separator"></div>
		{:else}
			<button
				class="menu-item"
				class:disabled={item.disabled}
				class:danger={item.danger}
				role="menuitem"
				disabled={item.disabled}
				onclick={() => handleItemClick(item)}
			>
				{#if item.icon}
					<span class="item-icon">
						<Icon name={item.icon} size={14} />
					</span>
				{/if}
				<span class="item-label">{item.label}</span>
				{#if item.shortcut}
					<span class="item-shortcut">{item.shortcut}</span>
				{/if}
			</button>
		{/if}
	{/each}
</div>

<style>
	.context-menu-backdrop {
		position: fixed;
		inset: 0;
		z-index: 9998;
	}

	.context-menu {
		position: fixed;
		z-index: 9999;
		min-width: 180px;
		max-width: 280px;
		padding: var(--spacing-xs, 4px) 0;
		background: var(--bg-panel, #2a2a2a);
		border: 2px solid var(--border-metal, #3a3a3a);
		border-radius: var(--border-radius-sm, 4px);
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.5);
		animation: context-menu-enter 0.1s ease-out;
	}

	@keyframes context-menu-enter {
		from {
			opacity: 0;
			transform: scale(0.95);
		}
		to {
			opacity: 1;
			transform: scale(1);
		}
	}

	.menu-item {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm, 8px);
		width: 100%;
		padding: var(--spacing-sm, 8px) var(--spacing-md, 16px);
		background: transparent;
		border: none;
		color: var(--text-primary, #e0e0e0);
		font-family: inherit;
		font-size: var(--font-size-sm, 13px);
		text-align: left;
		cursor: pointer;
		transition: background-color 0.1s ease;
	}

	.menu-item:hover:not(.disabled) {
		background: var(--bg-hover, #3a3a3a);
	}

	.menu-item:focus-visible {
		outline: none;
		background: var(--bg-hover, #3a3a3a);
		box-shadow: inset 0 0 0 2px var(--focus-ring-color, #f59e0b);
	}

	.menu-item.disabled {
		color: var(--text-muted, #666);
		cursor: not-allowed;
	}

	.menu-item.danger {
		color: var(--status-poor, #ef4444);
	}

	.menu-item.danger:hover:not(.disabled) {
		background: rgba(239, 68, 68, 0.15);
	}

	.item-icon {
		display: flex;
		align-items: center;
		flex-shrink: 0;
		color: var(--text-muted, #888);
	}

	.menu-item:hover:not(.disabled) .item-icon {
		color: var(--text-secondary, #ccc);
	}

	.menu-item.danger .item-icon {
		color: var(--status-poor, #ef4444);
	}

	.item-label {
		flex: 1;
	}

	.item-shortcut {
		font-size: var(--font-size-xs, 11px);
		color: var(--text-muted, #666);
		font-family: var(--font-mono, 'IBM Plex Mono', monospace);
	}

	.menu-separator {
		height: 1px;
		margin: var(--spacing-xs, 4px) var(--spacing-sm, 8px);
		background: var(--border-dark, #333);
	}

	@media (prefers-reduced-motion: reduce) {
		.context-menu {
			animation: none;
		}

		.menu-item {
			transition: none;
		}
	}
</style>
