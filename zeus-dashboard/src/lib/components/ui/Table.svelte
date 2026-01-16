<script lang="ts" generics="T">
	import type { Snippet } from 'svelte';

	interface Column<T> {
		key: keyof T | string;
		label: string;
		width?: string;
	}

	interface Props<T> {
		columns: Column<T>[];
		data: T[];
		emptyMessage?: string;
		rowKey?: keyof T;
		cellRenderer?: Snippet<[{ item: T; column: Column<T> }]>;
	}

	let {
		columns,
		data,
		emptyMessage = 'No data available',
		rowKey,
		cellRenderer
	}: Props<T> = $props();

	// セルの値を取得
	function getCellValue(item: T, key: string): unknown {
		const keys = key.split('.');
		let value: unknown = item;
		for (const k of keys) {
			if (value && typeof value === 'object' && k in value) {
				value = (value as Record<string, unknown>)[k];
			} else {
				return '';
			}
		}
		return value;
	}

	// 行のキーを取得
	function getRowKey(item: T, index: number): string {
		if (rowKey && item && typeof item === 'object' && rowKey in item) {
			return String((item as Record<string, unknown>)[rowKey as string]);
		}
		return String(index);
	}
</script>

<div class="table-container">
	<table class="table">
		<thead>
			<tr>
				{#each columns as column}
					<th style={column.width ? `width: ${column.width}` : ''}>
						{column.label}
					</th>
				{/each}
			</tr>
		</thead>
		<tbody>
			{#if data.length === 0}
				<tr class="empty-row">
					<td colspan={columns.length}>
						<div class="empty-message">{emptyMessage}</div>
					</td>
				</tr>
			{:else}
				{#each data as item, index (getRowKey(item, index))}
					<tr>
						{#each columns as column}
							<td>
								{#if cellRenderer}
									{@render cellRenderer({ item, column })}
								{:else}
									{getCellValue(item, String(column.key))}
								{/if}
							</td>
						{/each}
					</tr>
				{/each}
			{/if}
		</tbody>
	</table>
</div>

<style>
	.table-container {
		overflow-x: auto;
	}

	.table {
		width: 100%;
		border-collapse: collapse;
		font-size: var(--font-size-sm);
	}

	.table th {
		text-align: left;
		padding: var(--spacing-sm) var(--spacing-md);
		background-color: var(--bg-secondary);
		color: var(--accent-primary);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		font-size: var(--font-size-xs);
		border-bottom: 2px solid var(--border-metal);
	}

	.table td {
		padding: var(--spacing-sm) var(--spacing-md);
		border-bottom: 1px solid var(--border-dark);
		color: var(--text-primary);
	}

	.table tbody tr {
		transition: background-color var(--transition-fast);
	}

	.table tbody tr:hover {
		background-color: var(--bg-hover);
	}

	.empty-row td {
		text-align: center;
	}

	.empty-message {
		padding: var(--spacing-lg);
		color: var(--text-muted);
		font-style: italic;
	}
</style>
