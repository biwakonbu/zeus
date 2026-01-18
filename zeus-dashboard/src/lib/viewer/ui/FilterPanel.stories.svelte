<script context="module" lang="ts">
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
			availableAssignees: ['alice', 'bob', 'charlie'],
			onStatusToggle: () => {},
			onPriorityToggle: () => {},
			onAssigneeToggle: () => {},
			onSearchChange: () => {},
			onClear: () => {}
		}
	});
</script>

<script lang="ts">
	import { fn } from '@storybook/test';
	import type { TaskStatus, Priority } from '$lib/types/api';

	// Action ハンドラー
	const handleStatusToggle = fn();
	const handlePriorityToggle = fn();
	const handleAssigneeToggle = fn();
	const handleSearchChange = fn();
	const handleClear = fn();

	// ステータスフィルターあり
	const withStatusFilter: FilterCriteria = {
		statuses: ['in_progress', 'pending']
	};

	// 完全なフィルター
	const fullFilter: FilterCriteria = {
		statuses: ['in_progress'],
		priorities: ['high'],
		assignees: ['alice'],
		searchText: 'API'
	};

	// インタラクティブ用の状態
	let interactiveCriteria: FilterCriteria = $state({});

	function toggleStatus(status: TaskStatus) {
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

	function togglePriority(priority: Priority) {
		const priorities = interactiveCriteria.priorities || [];
		const index = priorities.indexOf(priority);
		if (index >= 0) {
			priorities.splice(index, 1);
		} else {
			priorities.push(priority);
		}
		interactiveCriteria = {
			...interactiveCriteria,
			priorities: priorities.length > 0 ? [...priorities] : undefined
		};
		handlePriorityToggle(priority);
	}

	function toggleAssignee(assignee: string) {
		const assignees = interactiveCriteria.assignees || [];
		const index = assignees.indexOf(assignee);
		if (index >= 0) {
			assignees.splice(index, 1);
		} else {
			assignees.push(assignee);
		}
		interactiveCriteria = {
			...interactiveCriteria,
			assignees: assignees.length > 0 ? [...assignees] : undefined
		};
		handleAssigneeToggle(assignee);
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
<Story name="NoFilter" args={{
	criteria: {},
	onStatusToggle: handleStatusToggle,
	onPriorityToggle: handlePriorityToggle,
	onAssigneeToggle: handleAssigneeToggle,
	onSearchChange: handleSearchChange,
	onClear: handleClear
}} let:args>
	<div class="filter-story-wrapper">
		<FilterPanel {...args} />
	</div>
</Story>

<!-- ステータスフィルターあり -->
<Story name="WithStatusFilter" args={{
	criteria: withStatusFilter,
	onStatusToggle: handleStatusToggle,
	onPriorityToggle: handlePriorityToggle,
	onAssigneeToggle: handleAssigneeToggle,
	onSearchChange: handleSearchChange,
	onClear: handleClear
}} let:args>
	<div class="filter-story-wrapper">
		<FilterPanel {...args} />
	</div>
</Story>

<!-- フルフィルター -->
<Story name="FullFilter" args={{
	criteria: fullFilter,
	onStatusToggle: handleStatusToggle,
	onPriorityToggle: handlePriorityToggle,
	onAssigneeToggle: handleAssigneeToggle,
	onSearchChange: handleSearchChange,
	onClear: handleClear
}} let:args>
	<div class="filter-story-wrapper">
		<FilterPanel {...args} />
	</div>
</Story>

<!-- 担当者なし -->
<Story name="NoAssignees" args={{
	criteria: {},
	availableAssignees: [],
	onStatusToggle: handleStatusToggle,
	onPriorityToggle: handlePriorityToggle,
	onAssigneeToggle: handleAssigneeToggle,
	onSearchChange: handleSearchChange,
	onClear: handleClear
}} let:args>
	<div class="filter-story-wrapper">
		<FilterPanel {...args} />
	</div>
</Story>

<!-- インタラクティブ -->
<Story name="Interactive" args={{
	criteria: interactiveCriteria,
	onStatusToggle: toggleStatus,
	onPriorityToggle: togglePriority,
	onAssigneeToggle: toggleAssignee,
	onSearchChange: changeSearch,
	onClear: clearAll
}}>
	<div class="filter-story-wrapper">
		<FilterPanel
			criteria={interactiveCriteria}
			availableAssignees={['alice', 'bob', 'charlie']}
			onStatusToggle={toggleStatus}
			onPriorityToggle={togglePriority}
			onAssigneeToggle={toggleAssignee}
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
