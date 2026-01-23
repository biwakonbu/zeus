<script lang="ts">
	// SearchInput - 検索入力コンポーネント
	// Factorio 風デザインの検索ボックス
	import { Icon } from '$lib/components/ui';

	interface Props {
		value: string;
		placeholder?: string;
		onInput: (value: string) => void;
		onClear?: () => void;
	}
	let { value, placeholder = '検索...', onInput, onClear }: Props = $props();

	function handleClear() {
		onInput('');
		onClear?.();
	}
</script>

<div class="search-wrapper">
	<span class="search-icon">
		<Icon name="Search" size={16} />
	</span>
	<input
		type="text"
		class="search-input"
		{value}
		{placeholder}
		aria-label="検索"
		oninput={(e) => onInput(e.currentTarget.value)}
	/>
	{#if value}
		<button class="clear-btn" onclick={handleClear} aria-label="クリア">
			<Icon name="X" size={14} />
		</button>
	{/if}
</div>

<style>
	.search-wrapper {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0.75rem;
		background: linear-gradient(
			180deg,
			rgba(25, 25, 25, 0.9) 0%,
			rgba(20, 20, 20, 0.95) 100%
		);
		border: 2px solid var(--border-metal);
		border-radius: var(--border-radius-sm);
		/* インナーシャドウで凹み感 */
		box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.3);
		transition:
			border-color var(--transition-select) ease-out,
			box-shadow var(--transition-select) ease-out;
	}

	.search-wrapper:focus-within {
		border-color: var(--accent-primary);
		/* 強化されたグロー効果 */
		box-shadow:
			0 0 12px rgba(255, 149, 51, 0.4),
			inset 0 2px 4px rgba(0, 0, 0, 0.3);
	}

	.search-icon {
		display: flex;
		align-items: center;
		color: var(--text-muted);
		flex-shrink: 0;
		transition: color var(--transition-select) ease-out;
	}

	.search-wrapper:focus-within .search-icon {
		color: var(--accent-primary);
	}

	.search-input {
		flex: 1;
		min-width: 0;
		background: transparent;
		border: none;
		color: var(--text-primary);
		font-family: var(--font-family);
		font-size: var(--font-size-sm);
	}

	.search-input::placeholder {
		color: var(--text-muted);
	}

	.search-input:focus {
		outline: none;
	}

	.clear-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0.25rem;
		background: transparent;
		border: 1px solid transparent;
		color: var(--text-muted);
		cursor: pointer;
		border-radius: var(--border-radius-sm);
		flex-shrink: 0;
		transition: all var(--transition-select) ease-out;
	}

	.clear-btn:hover {
		color: var(--accent-primary);
		background: rgba(255, 149, 51, 0.15);
		border-color: rgba(255, 149, 51, 0.3);
	}

	.clear-btn:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	/* アニメーション対応 */
	@media (prefers-reduced-motion: reduce) {
		.search-wrapper,
		.search-icon,
		.clear-btn {
			transition: none;
		}
	}
</style>
