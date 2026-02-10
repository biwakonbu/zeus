<script lang="ts">
	import type { GraphNode } from '$lib/types/api';
	import { Icon, SearchInput } from '$lib/components/ui';
	import { getNodeTypeCSSColor, getNodeTypeLabel } from '../config/nodeTypes';

	interface Props {
		nodes: GraphNode[];
		selectedNodeId: string | null;
		searchQuery: string;
		visibleCount: number;
		totalCount: number;
		onNodeSelect: (nodeId: string) => void;
		onSearchChange: (value: string) => void;
	}

	let {
		nodes,
		selectedNodeId,
		searchQuery,
		visibleCount,
		totalCount,
		onNodeSelect,
		onSearchChange
	}: Props = $props();

	const filteredNodes = $derived.by(() => {
		const query = searchQuery.trim().toLowerCase();
		if (!query) return nodes;

		return nodes.filter((node) => {
			const fields = [node.title, node.id, node.node_type, node.status];
			return fields.some((field) => field.toLowerCase().includes(query));
		});
	});

	function getStatusColor(status: string): string {
		const colors: Record<string, string> = {
			draft: 'var(--task-pending)',
			active: 'var(--task-in-progress)',
			deprecated: 'var(--task-completed)',
			// Objective 用ステータス
			not_started: 'var(--task-pending)',
			in_progress: 'var(--task-in-progress)',
			completed: 'var(--task-completed)',
			on_hold: 'var(--task-on-hold)'
		};
		return colors[status] ?? 'var(--text-secondary)';
	}

	function getStatusLabel(status: string): string {
		const labels: Record<string, string> = {
			draft: '下書き',
			active: 'アクティブ',
			deprecated: '非推奨',
			// Objective 用ステータス
			not_started: '未着手',
			in_progress: '進行中',
			completed: '完了',
			on_hold: '保留'
		};
		return labels[status] ?? status;
	}
</script>

<div class="list-panel-content">
	<div class="search-row">
		<SearchInput
			value={searchQuery}
			placeholder="ノードを検索..."
			onInput={onSearchChange}
		/>
	</div>

	<div class="count-row">
		<span class="count-label">表示: {visibleCount} / 全体: {totalCount}</span>
		{#if filteredNodes.length !== nodes.length}
			<span class="count-label">検索結果: {filteredNodes.length}</span>
		{/if}
	</div>

	<div class="list-area">
		{#if filteredNodes.length === 0}
			<div class="empty-list">
				<Icon name="Inbox" size={24} />
				<span>表示可能なノードがありません</span>
			</div>
		{:else}
			{#each filteredNodes as node (node.id)}
				{@const isSelected = selectedNodeId === node.id}
				<button class="node-item" class:selected={isSelected} onclick={() => onNodeSelect(node.id)}>
					<div class="node-header">
						<span class="type-badge" style="background: {getNodeTypeCSSColor(node.node_type)}">
							{getNodeTypeLabel(node.node_type)}
						</span>
						<span class="status-dot" style="background: {getStatusColor(node.status)}"></span>
						<span class="node-title">{node.title}</span>
					</div>
					<div class="node-meta">
						<span class="node-id">{node.id}</span>
						<span class="node-type">{node.node_type}</span>
						<span class="node-status">{getStatusLabel(node.status)}</span>
					</div>
				</button>
			{/each}
		{/if}
	</div>
</div>

<style>
	.list-panel-content {
		display: flex;
		flex-direction: column;
		gap: 8px;
		height: 100%;
		padding: 8px;
		overflow: hidden;
	}

	.search-row {
		flex-shrink: 0;
	}

	.count-row {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
		flex-shrink: 0;
	}

	.count-label {
		font-size: 0.6875rem;
		color: var(--text-muted);
	}

	.list-area {
		flex: 1;
		overflow-y: auto;
		margin: 0 -8px -8px;
		padding: 0 8px 8px;
	}

	.empty-list {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 8px;
		height: 100%;
		padding: 24px;
		color: var(--text-muted);
		text-align: center;
	}

	.empty-list span {
		font-size: 0.75rem;
	}

	.node-item {
		display: flex;
		flex-direction: column;
		gap: 6px;
		width: 100%;
		margin-bottom: 6px;
		padding: 10px;
		background: rgba(0, 0, 0, 0.2);
		border: 1px solid var(--border-metal);
		border-radius: 4px;
		cursor: pointer;
		text-align: left;
		font-family: inherit;
		transition:
			background 0.15s ease,
			border-color 0.15s ease;
	}

	.node-item:hover {
		background: rgba(0, 0, 0, 0.3);
		border-color: var(--border-hover);
	}

	.node-item.selected {
		background: rgba(255, 149, 51, 0.15);
		border-color: var(--accent-primary);
	}

	.node-header {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.type-badge {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 32px;
		padding: 2px 6px;
		border-radius: 3px;
		font-size: 0.625rem;
		font-weight: 600;
		color: #1a1a1a;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.status-dot {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.node-title {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--text-primary);
		line-height: 1.3;
	}

	.node-meta {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
		font-size: 0.6875rem;
		color: var(--text-muted);
	}

	.node-id {
		font-family: monospace;
		font-size: 0.625rem;
	}

	.node-type,
	.node-status {
		padding: 1px 5px;
		background: rgba(0, 0, 0, 0.3);
		border-radius: 2px;
	}
</style>
