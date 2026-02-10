<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import FilterPanel from './FilterPanel.svelte';
	import type { FilterCriteria } from '../interaction/FilterManager';

	const { Story } = defineMeta({
		title: 'Viewer/FilterPanel',
		component: FilterPanel,
		tags: ['autodocs'],
		parameters: {
			layout: 'padded'
		},
		args: {
			criteria: {} as FilterCriteria,
			onStatusToggle: () => {},
			onSearchChange: () => {},
			onClear: () => {}
		}
	});
</script>

<script lang="ts">
	import { fn } from '@storybook/test';
	import type { EntityStatus } from '$lib/types/api';

	// Action ハンドラー
	const handleStatusToggle = fn();
	const handleSearchChange = fn();
	const handleClear = fn();

	// ステータスフィルターあり
	const withStatusFilter: FilterCriteria = {
		statuses: ['active']
	};

	// 完全なフィルター
	const fullFilter: FilterCriteria = {
		statuses: ['active'],
		searchText: 'API'
	};

	// インタラクティブ用の状態
	let interactiveCriteria: FilterCriteria = $state({});

	function toggleStatus(status: EntityStatus) {
		const statuses = interactiveCriteria.statuses || [];
		const index = statuses.indexOf(status);
		if (index >= 0) {
			statuses.splice(index, 1);
		} else {
			statuses.push(status);
		}
		interactiveCriteria = {
			...interactiveCriteria,
			statuses: statuses.length > 0 ? [...statuses] : undefined
		};
		handleStatusToggle(status);
	}

	function changeSearch(text: string) {
		interactiveCriteria = {
			...interactiveCriteria,
			searchText: text || undefined
		};
		handleSearchChange(text);
	}

	function clearAll() {
		interactiveCriteria = {};
		handleClear();
	}
</script>

<!-- フィルターなし -->
<Story name="NoFilter">
	<div class="filter-story-wrapper">
		<FilterPanel
			criteria={{}}
			onStatusToggle={handleStatusToggle}
			onSearchChange={handleSearchChange}
			onClear={handleClear}
		/>
	</div>
</Story>

<!-- ステータスフィルターあり -->
<Story name="WithStatusFilter">
	<div class="filter-story-wrapper">
		<FilterPanel
			criteria={withStatusFilter}
			onStatusToggle={handleStatusToggle}
			onSearchChange={handleSearchChange}
			onClear={handleClear}
		/>
	</div>
</Story>

<!-- フルフィルター -->
<Story name="FullFilter">
	<div class="filter-story-wrapper">
		<FilterPanel
			criteria={fullFilter}
			onStatusToggle={handleStatusToggle}
			onSearchChange={handleSearchChange}
			onClear={handleClear}
		/>
	</div>
</Story>

<!-- インタラクティブ -->
<Story
	name="Interactive"
	args={{
		criteria: interactiveCriteria,
		onStatusToggle: toggleStatus,
		onSearchChange: changeSearch,
		onClear: clearAll
	}}
>
	<div class="filter-story-wrapper">
		<FilterPanel
			criteria={interactiveCriteria}
			onStatusToggle={toggleStatus}
			onSearchChange={changeSearch}
			onClear={clearAll}
		/>
		<div class="criteria-display">
			<p style="color: #888; font-size: 11px; margin-bottom: 4px;">Current Criteria:</p>
			<pre style="color: #f5a623; font-size: 10px; overflow-x: auto;">
{JSON.stringify(interactiveCriteria, null, 2)}
			</pre>
		</div>
	</div>
</Story>

<style>
	/* Storybook 用ラッパー：FilterPanel の position: absolute を相対的に表示 */
	.filter-story-wrapper {
		position: relative;
		min-height: 350px;
		width: 320px;
		background: #1a1a1a;
		padding: 16px;
		border: 1px solid #444;
	}

	/* FilterPanel の position を Storybook 表示用に上書き */
	.filter-story-wrapper :global(.filter-panel) {
		position: static;
		max-width: 100%;
	}

	.criteria-display {
		margin-top: 16px;
		padding: 12px;
		background: #2d2d2d;
		border-radius: 4px;
	}
</style>
