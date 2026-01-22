<script lang="ts">
	import { Icon } from '$lib/components/ui';

	// ビュータイプの定義
	export type ViewType = 'graph' | 'usecase';

	interface Props {
		currentView: ViewType;
		onViewChange: (view: ViewType) => void;
		disabledViews?: ViewType[];
	}

	let { currentView, onViewChange, disabledViews = [] }: Props = $props();

	// Lucide Icon 名を使用
	const views: { type: ViewType; label: string; iconName: string; description: string }[] = [
		{
			type: 'graph',
			label: 'Graph',
			iconName: 'Network',
			description: '依存関係グラフ'
		},
		{
			type: 'usecase',
			label: 'UseCase',
			iconName: 'Users',
			description: 'UML ユースケース図'
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
			<span class="view-icon">
				<Icon name={view.iconName} size={14} />
			</span>
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
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.view-label {
		font-size: 12px;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}
</style>
