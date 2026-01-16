<script lang="ts">
	// ビュータイプの定義
	export type ViewType = 'graph' | 'wbs' | 'timeline';

	interface Props {
		currentView: ViewType;
		onViewChange: (view: ViewType) => void;
		disabledViews?: ViewType[];
	}

	let { currentView, onViewChange, disabledViews = [] }: Props = $props();

	const views: { type: ViewType; label: string; icon: string; description: string }[] = [
		{
			type: 'graph',
			label: 'Graph',
			icon: '⬡',
			description: '依存関係グラフ'
		},
		{
			type: 'wbs',
			label: 'WBS',
			icon: '▤',
			description: '階層構造'
		},
		{
			type: 'timeline',
			label: 'Timeline',
			icon: '▬',
			description: 'ガントチャート'
		}
	];

	function handleViewChange(view: ViewType) {
		if (!disabledViews.includes(view)) {
			onViewChange(view);
		}
	}
</script>

<div class="view-switcher">
	{#each views as view}
		{@const isActive = currentView === view.type}
		{@const isDisabled = disabledViews.includes(view.type)}
		<button
			class="view-btn"
			class:active={isActive}
			class:disabled={isDisabled}
			onclick={() => handleViewChange(view.type)}
			title={view.description}
			disabled={isDisabled}
		>
			<span class="view-icon">{view.icon}</span>
			<span class="view-label">{view.label}</span>
		</button>
	{/each}
</div>

<style>
	.view-switcher {
		display: flex;
		background: #252525;
		border-radius: 6px;
		padding: 4px;
		gap: 4px;
	}

	.view-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 16px;
		background: transparent;
		border: none;
		color: #888;
		border-radius: 4px;
		cursor: pointer;
		font-family: inherit;
		font-size: 13px;
		transition: all 0.2s;
	}

	.view-btn:hover:not(.disabled) {
		background: #333;
		color: #ccc;
	}

	.view-btn.active {
		background: #f59e0b;
		color: #1a1a1a;
		font-weight: 600;
	}

	.view-btn.disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.view-icon {
		font-size: 14px;
	}

	.view-label {
		font-size: 12px;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}
</style>
