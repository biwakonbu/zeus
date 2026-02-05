<script lang="ts">
	// FilterDropdown - フィルタドロップダウンコンポーネント
	// 関連 Actor でのフィルタリング用（矢印キーナビゲーション対応）
	import { Icon } from '$lib/components/ui';

	type Option = {
		id: string;
		label: string;
	};

	interface Props {
		options: Option[];
		selected: string | null;
		placeholder: string;
		onSelect: (id: string | null) => void;
	}
	let { options, selected, placeholder, onSelect }: Props = $props();

	let isOpen = $state(false);
	let dropdownRef: HTMLDivElement | null = $state(null);
	let triggerRef: HTMLButtonElement | null = $state(null);
	let menuRef: HTMLUListElement | null = $state(null);
	let focusedIndex = $state(-1); // -1 = trigger focused, 0 = 全て, 1+ = options

	// 全オプション（「全て」を含む）
	const allOptions = $derived([
		{ id: null, label: '全て' },
		...options.map((o) => ({ id: o.id, label: o.label }))
	]);

	const selectedLabel = $derived(
		selected ? (options.find((o) => o.id === selected)?.label ?? placeholder) : placeholder
	);

	function handleToggle() {
		isOpen = !isOpen;
		if (isOpen) {
			// 開いたときは現在の選択位置にフォーカス
			const currentIndex = allOptions.findIndex((o) => o.id === selected);
			focusedIndex = currentIndex >= 0 ? currentIndex : 0;
		}
	}

	function handleSelect(id: string | null) {
		onSelect(id);
		isOpen = false;
		focusedIndex = -1;
		triggerRef?.focus();
	}

	// 外部クリックで閉じる
	function handleClickOutside(event: MouseEvent) {
		const target = event.target;
		if (dropdownRef && target instanceof Node && !dropdownRef.contains(target)) {
			isOpen = false;
			focusedIndex = -1;
		}
	}

	// キーボードナビゲーション
	function handleKeydown(event: KeyboardEvent) {
		if (!isOpen) {
			// 閉じているとき
			if (event.key === 'Enter' || event.key === ' ' || event.key === 'ArrowDown') {
				event.preventDefault();
				isOpen = true;
				focusedIndex = 0;
			}
			return;
		}

		// 開いているとき
		switch (event.key) {
			case 'Escape':
				event.preventDefault();
				isOpen = false;
				focusedIndex = -1;
				triggerRef?.focus();
				break;
			case 'ArrowDown':
				event.preventDefault();
				focusedIndex = Math.min(focusedIndex + 1, allOptions.length - 1);
				focusMenuItem(focusedIndex);
				break;
			case 'ArrowUp':
				event.preventDefault();
				focusedIndex = Math.max(focusedIndex - 1, 0);
				focusMenuItem(focusedIndex);
				break;
			case 'Home':
				event.preventDefault();
				focusedIndex = 0;
				focusMenuItem(focusedIndex);
				break;
			case 'End':
				event.preventDefault();
				focusedIndex = allOptions.length - 1;
				focusMenuItem(focusedIndex);
				break;
			case 'Enter':
			case ' ':
				event.preventDefault();
				if (focusedIndex >= 0 && focusedIndex < allOptions.length) {
					handleSelect(allOptions[focusedIndex].id);
				}
				break;
			case 'Tab':
				// Tab キーで閉じる
				isOpen = false;
				focusedIndex = -1;
				break;
		}
	}

	// メニューアイテムにフォーカス
	function focusMenuItem(index: number) {
		if (menuRef) {
			const items = menuRef.querySelectorAll<HTMLButtonElement>('.dropdown-item');
			items[index]?.focus();
		}
	}

	$effect(() => {
		if (isOpen) {
			// 次のフレームでリスナーを追加（トグルクリックを無視するため）
			const timeoutId = setTimeout(() => {
				document.addEventListener('click', handleClickOutside);
			}, 0);
			return () => {
				clearTimeout(timeoutId);
				document.removeEventListener('click', handleClickOutside);
			};
		}
	});
</script>

<div class="dropdown" bind:this={dropdownRef} role="presentation">
	<button
		bind:this={triggerRef}
		class="dropdown-trigger"
		class:open={isOpen}
		onclick={handleToggle}
		onkeydown={handleKeydown}
		aria-haspopup="listbox"
		aria-expanded={isOpen}
	>
		<span class="trigger-label">{selectedLabel}</span>
		<span class="trigger-icon" class:rotated={isOpen}>
			<Icon name="ChevronDown" size={14} />
		</span>
	</button>

	{#if isOpen}
		<ul bind:this={menuRef} class="dropdown-menu" role="listbox">
			{#each allOptions as option, i}
				<li>
					<button
						class="dropdown-item"
						class:selected={selected === option.id}
						class:focused={focusedIndex === i}
						role="option"
						aria-selected={selected === option.id}
						tabindex={focusedIndex === i ? 0 : -1}
						onclick={() => handleSelect(option.id)}
					>
						{option.label}
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>

<style>
	.dropdown {
		position: relative;
	}

	.dropdown-trigger {
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
		padding: 0.5rem 0.75rem;
		background: linear-gradient(180deg, rgba(25, 25, 25, 0.9) 0%, rgba(20, 20, 20, 0.95) 100%);
		border: 2px solid var(--border-metal);
		border-radius: var(--border-radius-sm);
		color: var(--text-primary);
		font-family: var(--font-family);
		font-size: var(--font-size-sm);
		cursor: pointer;
		/* インナーシャドウ */
		box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.3);
		transition:
			border-color var(--transition-select) ease-out,
			box-shadow var(--transition-select) ease-out;
	}

	.dropdown-trigger:hover {
		border-color: var(--border-highlight);
	}

	.dropdown-trigger.open {
		border-color: var(--accent-primary);
		/* グロー効果 */
		box-shadow:
			0 0 10px rgba(255, 149, 51, 0.3),
			inset 0 2px 4px rgba(0, 0, 0, 0.3);
	}

	.dropdown-trigger:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	.trigger-label {
		flex: 1;
		text-align: left;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.trigger-icon {
		display: flex;
		align-items: center;
		color: var(--text-muted);
		transition: transform var(--transition-select) ease-out;
	}

	.trigger-icon.rotated {
		transform: rotate(180deg);
	}

	.dropdown-trigger.open .trigger-icon {
		color: var(--accent-primary);
	}

	.dropdown-menu {
		position: absolute;
		top: calc(100% + 4px);
		left: 0;
		right: 0;
		max-height: 200px;
		overflow-y: auto;
		background: linear-gradient(180deg, rgba(45, 45, 45, 0.98) 0%, rgba(36, 36, 36, 0.95) 100%);
		border: 2px solid var(--border-metal);
		border-radius: var(--border-radius-sm);
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
		z-index: 100;
		list-style: none;
		padding: 0.25rem 0;
		margin: 0;
	}

	.dropdown-item {
		display: block;
		width: 100%;
		padding: 0.5rem 0.75rem;
		background: transparent;
		border: none;
		border-left: 2px solid transparent;
		color: var(--text-primary);
		font-family: var(--font-family);
		font-size: var(--font-size-sm);
		text-align: left;
		cursor: pointer;
		transition:
			background-color var(--transition-select) ease-out,
			border-color var(--transition-select) ease-out;
	}

	.dropdown-item:hover,
	.dropdown-item.focused {
		background: rgba(255, 149, 51, 0.1);
		border-left-color: var(--border-highlight);
	}

	.dropdown-item.selected {
		background: var(--accent-primary);
		color: var(--bg-primary);
		border-left-color: var(--accent-primary);
	}

	.dropdown-item:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: calc(-1 * var(--focus-ring-width));
	}

	/* アニメーション対応 */
	@media (prefers-reduced-motion: reduce) {
		.dropdown-trigger,
		.trigger-icon,
		.dropdown-item {
			transition: none;
		}
	}
</style>
