<script lang="ts">
	import type { EntityStatus, Priority } from '$lib/types/api';
	import type { FilterCriteria } from '../interaction/FilterManager';
	import { SearchInput } from '$lib/components/ui';

	// Props
	interface Props {
		criteria: FilterCriteria;
		availableAssignees: string[];
		onStatusToggle: (status: EntityStatus) => void;
		onPriorityToggle: (priority: Priority) => void;
		onAssigneeToggle: (assignee: string) => void;
		onSearchChange: (text: string) => void;
		onClear: () => void;
	}

	let {
		criteria,
		availableAssignees,
		onStatusToggle,
		onPriorityToggle,
		onAssigneeToggle,
		onSearchChange,
		onClear
	}: Props = $props();

	// ステータス一覧
	const statuses: { value: EntityStatus; label: string; color: string }[] = [
		{ value: 'completed', label: 'Completed', color: 'var(--task-completed)' },
		{ value: 'in_progress', label: 'In Progress', color: 'var(--task-in-progress)' },
		{ value: 'pending', label: 'Pending', color: 'var(--task-pending)' },
		{ value: 'blocked', label: 'Blocked', color: 'var(--task-blocked)' }
	];

	// 優先度一覧
	const priorities: { value: Priority; label: string; color: string }[] = [
		{ value: 'high', label: 'High', color: 'var(--priority-high)' },
		{ value: 'medium', label: 'Medium', color: 'var(--priority-medium)' },
		{ value: 'low', label: 'Low', color: 'var(--priority-low)' }
	];

	// フィルターがアクティブか
	let isActive = $derived(
		(criteria.statuses?.length ?? 0) > 0 ||
			(criteria.priorities?.length ?? 0) > 0 ||
			(criteria.assignees?.length ?? 0) > 0 ||
			!!criteria.searchText
	);

	// 検索入力（criteria.searchText に同期）
	let searchValue = $derived(criteria.searchText || '');

	function handleSearchInput(value: string) {
		onSearchChange(value);
	}

	function isStatusActive(status: EntityStatus): boolean {
		return criteria.statuses?.includes(status) ?? false;
	}

	function isPriorityActive(priority: Priority): boolean {
		return criteria.priorities?.includes(priority) ?? false;
	}

	function isAssigneeActive(assignee: string): boolean {
		return criteria.assignees?.includes(assignee) ?? false;
	}
</script>

<div class="filter-content">
	<!-- 検索 -->
	<div class="filter-section">
		<span class="filter-label">Search</span>
		<SearchInput
			value={searchValue}
			placeholder="Task ID or title..."
			compact
			onInput={handleSearchInput}
		/>
	</div>

	<!-- ステータス -->
	<div class="filter-section" role="group" aria-label="Status filter">
		<span class="filter-label">Status</span>
		<div class="filter-chips">
			{#each statuses as status}
				<button
					class="filter-chip"
					class:active={isStatusActive(status.value)}
					style="--chip-color: {status.color}"
					onclick={() => onStatusToggle(status.value)}
				>
					<span class="chip-dot"></span>
					{status.label}
				</button>
			{/each}
		</div>
	</div>

	<!-- 優先度 -->
	<div class="filter-section" role="group" aria-label="Priority filter">
		<span class="filter-label">Priority</span>
		<div class="filter-chips">
			{#each priorities as priority}
				<button
					class="filter-chip"
					class:active={isPriorityActive(priority.value)}
					style="--chip-color: {priority.color}"
					onclick={() => onPriorityToggle(priority.value)}
				>
					<span class="chip-dot"></span>
					{priority.label}
				</button>
			{/each}
		</div>
	</div>

	<!-- 担当者 -->
	{#if availableAssignees.length > 0}
		<div class="filter-section" role="group" aria-label="Assignee filter">
			<span class="filter-label">Assignee</span>
			<div class="filter-chips">
				{#each availableAssignees as assignee}
					<button
						class="filter-chip"
						class:active={isAssigneeActive(assignee)}
						onclick={() => onAssigneeToggle(assignee)}
					>
						{assignee}
					</button>
				{/each}
			</div>
		</div>
	{/if}

	<!-- クリアボタン -->
	{#if isActive}
		<button class="filter-clear" onclick={onClear}>Clear All Filters</button>
	{/if}
</div>

<style>
	.filter-content {
		padding: var(--spacing-sm);
	}

	.filter-section {
		margin-bottom: var(--spacing-sm);
	}

	.filter-section:last-child {
		margin-bottom: 0;
	}

	.filter-label {
		display: block;
		font-size: 10px;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: var(--spacing-xs);
	}

	.filter-chips {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
	}

	.filter-chip {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 4px 8px;
		background-color: var(--bg-primary);
		border: 1px solid var(--border-dark);
		border-radius: var(--border-radius-sm);
		color: var(--text-secondary);
		font-size: 10px;
		cursor: pointer;
		transition: all var(--transition-fast);
	}

	.filter-chip:hover {
		border-color: var(--border-highlight);
		color: var(--text-primary);
	}

	.filter-chip.active {
		background-color: var(--chip-color, var(--accent-primary));
		border-color: var(--chip-color, var(--accent-primary));
		color: var(--bg-primary);
	}

	.chip-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background-color: var(--chip-color, var(--text-muted));
	}

	.filter-chip.active .chip-dot {
		background-color: var(--bg-primary);
	}

	.filter-clear {
		width: 100%;
		padding: var(--spacing-xs) var(--spacing-sm);
		margin-top: var(--spacing-sm);
		background-color: transparent;
		border: 1px solid var(--border-dark);
		border-radius: var(--border-radius-sm);
		color: var(--text-muted);
		font-size: 10px;
		cursor: pointer;
		transition: all var(--transition-fast);
	}

	.filter-clear:hover {
		border-color: var(--status-poor);
		color: var(--status-poor);
	}
</style>
