<script lang="ts">
	// SegmentedTabs - セグメント化タブ UI
	// Factorio 風デザインのタブコンポーネント

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
</script>

<div class="segmented-tabs" role="tablist">
	{#each tabs as tab}
		<button
			class="tab"
			class:active={activeTab === tab.id}
			role="tab"
			aria-selected={activeTab === tab.id}
			tabindex={activeTab === tab.id ? 0 : -1}
			onclick={() => onTabChange(tab.id)}
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
		background: linear-gradient(
			180deg,
			rgba(25, 25, 25, 0.9) 0%,
			rgba(20, 20, 20, 0.95) 100%
		);
		border: 2px solid var(--border-metal);
		border-radius: var(--border-radius-sm);
		overflow: hidden;
		/* インナーシャドウで凹み感 */
		box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.3);
	}

	.tab {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.25rem;
		padding: 0.5rem 0.75rem;
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
