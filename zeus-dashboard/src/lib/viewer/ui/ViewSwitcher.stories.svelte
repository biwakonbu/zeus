<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ViewSwitcher from './ViewSwitcher.svelte';

	const { Story } = defineMeta({
		title: 'Viewer/ViewSwitcher',
		component: ViewSwitcher,
		tags: ['autodocs'],
		argTypes: {
			currentView: {
				control: 'select',
				options: ['graph', 'wbs']
			},
			disabledViews: {
				control: 'multi-select',
				options: ['graph', 'wbs']
			}
		}
	});
</script>

<script lang="ts">
	import { fn } from '@storybook/test';

	// Action ハンドラー
	const handleViewChange = fn();

	// 状態付きのラッパー
	let currentView: 'graph' | 'wbs' = $state('graph');

	function createHandler(view: 'graph' | 'wbs') {
		currentView = view;
		handleViewChange(view);
	}
</script>

<!-- Graph 選択中 -->
<Story name="GraphSelected">
	<ViewSwitcher currentView="graph" onViewChange={handleViewChange} />
</Story>

<!-- WBS 選択中 -->
<Story name="WBSSelected">
	<ViewSwitcher currentView="wbs" onViewChange={handleViewChange} />
</Story>

<!-- WBS 無効化 -->
<Story name="WBSDisabled">
	<ViewSwitcher currentView="graph" onViewChange={handleViewChange} disabledViews={['wbs']} />
</Story>

<!-- インタラクティブ -->
<Story name="Interactive">
	<div style="display: flex; flex-direction: column; gap: 16px; align-items: center;">
		<ViewSwitcher {currentView} onViewChange={createHandler} />
		<p style="color: var(--text-secondary); font-size: 12px;">
			Current view: <strong style="color: var(--accent-primary);">{currentView}</strong>
		</p>
	</div>
</Story>

<!-- ダークテーマでの表示 -->
<Story name="InContext">
	<div
		style="background: var(--bg-secondary); padding: 16px; border-radius: 8px; display: flex; justify-content: center;"
	>
		<ViewSwitcher currentView="wbs" onViewChange={handleViewChange} />
	</div>
</Story>
