<script lang="ts">
	import Panel from '$lib/components/ui/Panel.svelte';
	import Table from '$lib/components/ui/Table.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import { tasks, totalTasks, tasksLoading, tasksError } from '$lib/stores/tasks';
	import type { TaskItem, TaskStatus, Priority } from '$lib/types/api';

	// テーブルカラム定義
	const columns = [
		{ key: 'id', label: 'ID', width: '100px' },
		{ key: 'title', label: 'Title' },
		{ key: 'status', label: 'Status', width: '120px' },
		{ key: 'priority', label: 'Priority', width: '100px' },
		{ key: 'assignee', label: 'Assignee', width: '120px' }
	];

	// ステータスのバッジバリアント
	function getStatusVariant(status: TaskStatus): 'success' | 'info' | 'muted' | 'danger' {
		switch (status) {
			case 'completed':
				return 'success';
			case 'in_progress':
				return 'info';
			case 'pending':
				return 'muted';
			case 'blocked':
				return 'danger';
			default:
				return 'muted';
		}
	}

	// ステータスのラベル
	function getStatusLabel(status: TaskStatus): string {
		switch (status) {
			case 'completed':
				return 'Completed';
			case 'in_progress':
				return 'In Progress';
			case 'pending':
				return 'Pending';
			case 'blocked':
				return 'Blocked';
			default:
				return status;
		}
	}

	// 優先度のバッジバリアント
	function getPriorityVariant(priority: Priority): 'danger' | 'warning' | 'success' | 'muted' {
		switch (priority) {
			case 'high':
				return 'danger';
			case 'medium':
				return 'warning';
			case 'low':
				return 'success';
			default:
				return 'muted';
		}
	}
</script>

<Panel title="Tasks" icon="&#128203;" loading={$tasksLoading} error={$tasksError}>
	{#snippet headerRight()}
		<span class="task-count">{$totalTasks} tasks</span>
	{/snippet}

	<Table {columns} data={$tasks} rowKey="id" emptyMessage="No tasks found">
		{#snippet cellRenderer({ item, column })}
			{@const task = item as TaskItem}
			{#if column.key === 'id'}
				<code class="task-id">{task.id.substring(0, 8)}</code>
			{:else if column.key === 'title'}
				<span class="task-title">{task.title}</span>
			{:else if column.key === 'status'}
				<Badge variant={getStatusVariant(task.status)} size="sm">
					{getStatusLabel(task.status)}
				</Badge>
			{:else if column.key === 'priority'}
				<Badge variant={getPriorityVariant(task.priority)} size="sm">
					{task.priority.toUpperCase()}
				</Badge>
			{:else if column.key === 'assignee'}
				{#if task.assignee}
					<span class="assignee">{task.assignee}</span>
				{:else}
					<span class="unassigned">Unassigned</span>
				{/if}
			{/if}
		{/snippet}
	</Table>
</Panel>

<style>
	.task-count {
		font-size: var(--font-size-sm);
		color: var(--text-secondary);
	}

	.task-id {
		font-size: var(--font-size-xs);
		color: var(--text-muted);
		background-color: var(--bg-secondary);
		padding: 2px 6px;
		border-radius: var(--border-radius-sm);
	}

	.task-title {
		font-weight: 500;
	}

	.assignee {
		color: var(--text-secondary);
	}

	.unassigned {
		color: var(--text-muted);
		font-style: italic;
	}
</style>
