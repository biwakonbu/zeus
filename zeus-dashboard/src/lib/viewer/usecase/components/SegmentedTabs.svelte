<script lang="ts">
	// SegmentedTabs - セグメント化タブ UI
	// Factorio 風デザインのタブコンポーネント（矢印キーナビゲーション対応）

	type TabItem = {
		id: string;
		label: string;
		count: number;
	};

	interface Props {
		tabs: TabItem[];
		activeTab: string;
		onTabChange: (tabId: string) => void;
	}
	let { tabs, activeTab, onTabChange }: Props = $props();

	// ボタン参照配列
	let buttonRefs: HTMLButtonElement[] = $state([]);

	// アクティブタブのインデックス
	const activeIndex = $derived(tabs.findIndex((t) => t.id === activeTab));

	// 矢印キーナビゲーション
	function handleKeydown(event: KeyboardEvent, currentIndex: number) {
		let newIndex: number | null = null;

		switch (event.key) {
			case 'ArrowRight':
			case 'ArrowDown':
				newIndex = currentIndex === tabs.length - 1 ? 0 : currentIndex + 1;
				break;
			case 'ArrowLeft':
			case 'ArrowUp':
				newIndex = currentIndex === 0 ? tabs.length - 1 : currentIndex - 1;
				break;
			case 'Home':
				newIndex = 0;
				break;
			case 'End':
				newIndex = tabs.length - 1;
				break;
			default:
				return; // その他のキーは無視
		}

		if (newIndex !== null) {
			event.preventDefault();
			onTabChange(tabs[newIndex].id);
			buttonRefs[newIndex]?.focus();
		}
	}
</script>

<div class="segmented-tabs" role="tablist">
	{#each tabs as tab, i}
		<button
			bind:this={buttonRefs[i]}
			class="tab"
			class:active={activeTab === tab.id}
			role="tab"
			aria-selected={activeTab === tab.id}
			tabindex={activeTab === tab.id ? 0 : -1}
			onclick={() => onTabChange(tab.id)}
			onkeydown={(e) => handleKeydown(e, i)}
		>
			<span class="tab-label">{tab.label}</span>
			<span class="tab-count">({tab.count})</span>
		</button>
	{/each}
</div>

<style>
	.segmented-tabs {
		display: flex;
		gap: 0;
		min-width: fit-content; /* 内容に応じた最小幅を確保 */
		background: linear-gradient(180deg, rgba(25, 25, 25, 0.9) 0%, rgba(20, 20, 20, 0.95) 100%);
		border: 2px solid var(--border-metal);
		border-radius: var(--border-radius-sm);
		overflow: hidden;
		/* インナーシャドウで凹み感 */
		box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.3);
	}

	.tab {
		flex: 0 0 auto; /* 均等配分ではなく内容に応じたサイズ */
		min-width: 60px; /* 各タブの最小幅 */
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.25rem;
		padding: 0.625rem 0.875rem; /* 縦幅拡大: 約10px x 14px */
		background: transparent;
		border: none;
		border-right: 1px solid var(--border-metal);
		color: var(--text-secondary);
		font-family: var(--font-family);
		font-size: var(--font-size-sm);
		font-weight: 500;
		cursor: pointer;
		transition:
			background-color var(--transition-select) ease-out,
			color var(--transition-select) ease-out,
			box-shadow var(--transition-select) ease-out;
	}

	.tab:last-child {
		border-right: none;
	}

	.tab:hover:not(.active) {
		background: rgba(255, 149, 51, 0.1);
		color: var(--text-primary);
	}

	.tab.active {
		background: var(--accent-primary);
		color: var(--bg-primary);
		/* 選択時のグロー効果 */
		box-shadow:
			0 0 12px rgba(255, 149, 51, 0.5),
			inset 0 1px 0 rgba(255, 255, 255, 0.2);
	}

	.tab:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: calc(-1 * var(--focus-ring-width));
		z-index: 1;
	}

	.tab-label {
		white-space: nowrap;
	}

	.tab-count {
		font-size: var(--font-size-xs);
		opacity: 0.8;
	}

	/* アニメーション対応 */
	@media (prefers-reduced-motion: reduce) {
		.tab {
			transition: none;
		}
	}
</style>
