<script lang="ts">
	// FilterDropdown - フィルタドロップダウンコンポーネント
	// 関連 Actor でのフィルタリング用
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

	const selectedLabel = $derived(
		selected ? options.find((o) => o.id === selected)?.label ?? placeholder : placeholder
	);

	function handleToggle() {
		isOpen = !isOpen;
	}

	function handleSelect(id: string | null) {
		onSelect(id);
		isOpen = false;
	}

	// 外部クリックで閉じる
	function handleClickOutside(event: MouseEvent) {
		if (dropdownRef && !dropdownRef.contains(event.target as Node)) {
			isOpen = false;
		}
	}

	// ESC キーで閉じる
	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && isOpen) {
			isOpen = false;
		}
	}

	$effect(() => {
		if (isOpen) {
			document.addEventListener('click', handleClickOutside);
			document.addEventListener('keydown', handleKeydown);
		}
		return () => {
			document.removeEventListener('click', handleClickOutside);
			document.removeEventListener('keydown', handleKeydown);
		};
	});
</script>

<div class="dropdown" bind:this={dropdownRef}>
	<button
		class="dropdown-trigger"
		class:open={isOpen}
		onclick={handleToggle}
		aria-haspopup="listbox"
		aria-expanded={isOpen}
	>
		<span class="trigger-label">{selectedLabel}</span>
		<span class="trigger-icon" class:rotated={isOpen}>
			<Icon name="ChevronDown" size={14} />
		</span>
	</button>

	{#if isOpen}
		<ul class="dropdown-menu" role="listbox">
			<li>
				<button
					class="dropdown-item"
					class:selected={selected === null}
					role="option"
					aria-selected={selected === null}
					onclick={() => handleSelect(null)}
				>
					全て
				</button>
			</li>
			{#each options as option}
				<li>
					<button
						class="dropdown-item"
						class:selected={selected === option.id}
						role="option"
						aria-selected={selected === option.id}
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
		background: linear-gradient(
			180deg,
			rgba(25, 25, 25, 0.9) 0%,
			rgba(20, 20, 20, 0.95) 100%
		);
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
		background: linear-gradient(
			180deg,
			rgba(45, 45, 45, 0.98) 0%,
			rgba(36, 36, 36, 0.95) 100%
		);
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

	.dropdown-item:hover {
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
