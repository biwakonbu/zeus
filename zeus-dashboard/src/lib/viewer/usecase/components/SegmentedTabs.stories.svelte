<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import SegmentedTabs from './SegmentedTabs.svelte';

	const { Story } = defineMeta({
		title: 'Viewer/UseCase/SegmentedTabs',
		component: SegmentedTabs,
		tags: ['autodocs'],
		argTypes: {
			activeTab: {
				control: 'text',
				description: 'アクティブなタブID'
			}
		}
	});
</script>

<script lang="ts">
	// Story は defineMeta から export される
	const basicTabs = [
		{ id: 'actor', label: 'Actor', count: 5 },
		{ id: 'usecase', label: 'UseCase', count: 12 }
	];

	const threeTabs = [
		{ id: 'all', label: '全て', count: 17 },
		{ id: 'actor', label: 'Actor', count: 5 },
		{ id: 'usecase', label: 'UseCase', count: 12 }
	];

	let activeTab = $state('actor');

	function handleTabChange(tabId: string) {
		activeTab = tabId;
		console.log('Tab changed:', tabId);
	}
</script>

<!-- 基本（2タブ） -->
<Story name="Default">
	<div style="padding: 24px; background: var(--bg-primary);">
		<SegmentedTabs tabs={basicTabs} activeTab="actor" onTabChange={handleTabChange} />
	</div>
</Story>

<!-- 3タブ -->
<Story name="ThreeTabs">
	<div style="padding: 24px; background: var(--bg-primary);">
		<SegmentedTabs tabs={threeTabs} activeTab="all" onTabChange={handleTabChange} />
	</div>
</Story>

<!-- UseCase選択時 -->
<Story name="UseCaseActive">
	<div style="padding: 24px; background: var(--bg-primary);">
		<SegmentedTabs tabs={basicTabs} activeTab="usecase" onTabChange={handleTabChange} />
	</div>
</Story>

<!-- カウント0 -->
<Story name="EmptyCount">
	{@const emptyTabs = [
		{ id: 'actor', label: 'Actor', count: 0 },
		{ id: 'usecase', label: 'UseCase', count: 3 }
	]}
	<div style="padding: 24px; background: var(--bg-primary);">
		<SegmentedTabs tabs={emptyTabs} activeTab="usecase" onTabChange={handleTabChange} />
	</div>
</Story>

<!-- 大きなカウント -->
<Story name="LargeCount">
	{@const largeTabs = [
		{ id: 'actor', label: 'Actor', count: 156 },
		{ id: 'usecase', label: 'UseCase', count: 2048 }
	]}
	<div style="padding: 24px; background: var(--bg-primary);">
		<SegmentedTabs tabs={largeTabs} activeTab="actor" onTabChange={handleTabChange} />
	</div>
</Story>

<!-- インタラクティブ -->
<Story name="Interactive">
	<div style="padding: 24px; background: var(--bg-primary);">
		<SegmentedTabs tabs={basicTabs} {activeTab} onTabChange={handleTabChange} />
		<p style="color: var(--text-muted); font-size: 12px; margin-top: 16px;">
			現在のタブ: {activeTab}
		</p>
	</div>
</Story>
