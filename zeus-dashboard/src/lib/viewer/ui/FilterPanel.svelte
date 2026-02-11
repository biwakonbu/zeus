<script lang="ts">
	import type { EntityStatus, UnifiedGraphGroupItem } from '$lib/types/api';
	import type { FilterCriteria } from '../interaction/FilterManager';
	import { SearchInput } from '$lib/components/ui';

	// Props
	interface Props {
		criteria: FilterCriteria;
		groups?: UnifiedGraphGroupItem[];
		onStatusToggle: (status: EntityStatus) => void;
		onSearchChange: (text: string) => void;
		onGroupToggle?: (groupId: string) => void;
		onClear: () => void;
	}

	let {
		criteria,
		groups = [],
		onStatusToggle,
		onSearchChange,
		onGroupToggle,
		onClear
	}: Props = $props();

	// ステータス一覧
	const statuses: { value: EntityStatus; label: string; color: string }[] = [
		{ value: 'draft', label: 'Draft', color: 'var(--task-pending)' },
		{ value: 'active', label: 'Active', color: 'var(--task-in-progress)' },
		{ value: 'deprecated', label: 'Deprecated', color: 'var(--text-muted)' }
	];

	// フィルターがアクティブか
	let isActive = $derived(
		(criteria.statuses?.length ?? 0) > 0 ||
			!!criteria.searchText ||
			(criteria.groupIds?.length ?? 0) > 0
	);

	// 検索入力（criteria.searchText に同期）
	let searchValue = $derived(criteria.searchText || '');

	function handleSearchInput(value: string) {
		onSearchChange(value);
	}

	function isStatusActive(status: EntityStatus): boolean {
		return criteria.statuses?.includes(status) ?? false;
	}

	function isGroupActive(groupId: string): boolean {
		return criteria.groupIds?.includes(groupId) ?? false;
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

	<!-- Objective グループ -->
	{#if groups && groups.length > 0}
		<div class="filter-section" role="group" aria-label="Group filter">
			<span class="filter-label">Objective</span>
			<div class="filter-chips">
				{#each groups as group}
					<button
						class="filter-chip"
						class:active={isGroupActive(group.id)}
						style="--chip-color: var(--accent-primary)"
						onclick={() => onGroupToggle?.(group.id)}
					>
						<span class="chip-dot"></span>
						{group.title}
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
