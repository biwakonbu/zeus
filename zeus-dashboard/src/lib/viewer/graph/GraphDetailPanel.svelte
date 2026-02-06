<script lang="ts">
	import type { GraphEdge, GraphNode } from '$lib/types/api';
	import { Icon } from '$lib/components/ui';
	import { navigateToEntity } from '$lib/stores/view';
	import { getNodeTypeCSSColor, getNodeTypeLabel } from '../config/nodeTypes';

	interface Props {
		node?: GraphNode | null;
		nodes: GraphNode[];
		edges: GraphEdge[];
		onClose?: () => void;
		onNodeSelect?: (nodeId: string) => void;
	}

	type RelationSummaryItem = {
		key: string;
		layer: string;
		relation: string;
		count: number;
	};

	let { node = null, nodes, edges, onClose = undefined, onNodeSelect = undefined }: Props = $props();

	const nodeById = $derived.by(() => new Map(nodes.map((item) => [item.id, item])));

	const incomingEdges = $derived.by(() => {
		if (!node) return [] as GraphEdge[];
		return edges.filter((edge) => edge.to === node.id);
	});

	const outgoingEdges = $derived.by(() => {
		if (!node) return [] as GraphEdge[];
		return edges.filter((edge) => edge.from === node.id);
	});

	function toUniqueNodes(ids: string[], map: Map<string, GraphNode>): GraphNode[] {
		const seen = new Set<string>();
		const result: GraphNode[] = [];

		for (const id of ids) {
			if (seen.has(id)) continue;
			seen.add(id);
			const found = map.get(id);
			if (found) result.push(found);
		}

		return result;
	}

	const upstreamNodes = $derived.by(() => toUniqueNodes(outgoingEdges.map((edge) => edge.to), nodeById));
	const downstreamNodes = $derived.by(() =>
		toUniqueNodes(incomingEdges.map((edge) => edge.from), nodeById)
	);

	function summarizeRelations(targetEdges: GraphEdge[]): RelationSummaryItem[] {
		const summary = new Map<string, RelationSummaryItem>();

		for (const edge of targetEdges) {
			const key = `${edge.layer}:${edge.relation}`;
			const prev = summary.get(key);
			if (prev) {
				prev.count += 1;
				continue;
			}
			summary.set(key, {
				key,
				layer: edge.layer,
				relation: edge.relation,
				count: 1
			});
		}

		return Array.from(summary.values()).sort((a, b) => {
			if (a.count !== b.count) return b.count - a.count;
			return a.key.localeCompare(b.key);
		});
	}

	const incomingRelationSummary = $derived.by(() => summarizeRelations(incomingEdges));
	const outgoingRelationSummary = $derived.by(() => summarizeRelations(outgoingEdges));

	function getStatusColor(status: string): string {
		const colors: Record<string, string> = {
			completed: 'var(--task-completed)',
			in_progress: 'var(--task-in-progress)',
			pending: 'var(--task-pending)',
			blocked: 'var(--task-blocked)',
			active: 'var(--status-good)',
			draft: 'var(--status-fair)',
			deprecated: 'var(--text-muted)'
		};
		return colors[status] ?? 'var(--text-secondary)';
	}

	function getStatusLabel(status: string): string {
		const labels: Record<string, string> = {
			completed: '完了',
			in_progress: '進行中',
			pending: '待機',
			blocked: 'ブロック',
			active: 'アクティブ',
			draft: '下書き',
			deprecated: '非推奨'
		};
		return labels[status] ?? status;
	}

	function handleNavigate(): void {
		if (!node) return;
		if (node.node_type === 'activity') {
			navigateToEntity('activity', 'activity', node.id);
		} else if (node.node_type === 'usecase') {
			navigateToEntity('usecase', 'usecase', node.id);
		}
	}

	const canNavigateToEntityView = $derived.by(
		() => node?.node_type === 'activity' || node?.node_type === 'usecase'
	);
</script>

<div class="detail-content">
	{#if onClose}
		<button class="close-button" onclick={onClose} title="閉じる">
			<Icon name="X" size={16} />
		</button>
	{/if}

	{#if node}
		<section class="section node-section">
			<h4 class="section-title">
				<Icon name="FileText" size={12} />
				ノード情報
			</h4>
			<div class="node-header">
				<span class="status-dot" style="background: {getStatusColor(node.status)}"></span>
				<div class="node-heading">
					<div class="node-title">{node.title}</div>
					<div class="node-id">{node.id}</div>
				</div>
				<span class="type-badge" style="background: {getNodeTypeCSSColor(node.node_type)}">
					{getNodeTypeLabel(node.node_type)}
				</span>
			</div>

			<div class="node-info-grid">
				<div class="info-item">
					<span class="label">Type</span>
					<span class="value">{node.node_type}</span>
				</div>
				<div class="info-item">
					<span class="label">Status</span>
					<span class="value">{getStatusLabel(node.status)}</span>
				</div>
				<div class="info-item">
					<span class="label">Priority</span>
					<span class="value">{node.priority ?? '-'}</span>
				</div>
				<div class="info-item">
					<span class="label">Assignee</span>
					<span class="value">{node.assignee ?? '-'}</span>
				</div>
				<div class="info-item">
					<span class="label">Depth</span>
					<span class="value">{node.structural_depth ?? '-'}</span>
				</div>
				<div class="info-item">
					<span class="label">Relations</span>
					<span class="value">{incomingEdges.length + outgoingEdges.length}</span>
				</div>
			</div>

			{#if canNavigateToEntityView}
				<div class="action-row">
					<button class="jump-button" onclick={handleNavigate}>
						<Icon name="ExternalLink" size={12} />
						<span>{node.node_type === 'activity' ? 'Activity ビューで開く' : 'UseCase ビューで開く'}</span>
					</button>
				</div>
			{/if}
		</section>

		<section class="section relation-section">
			<h4 class="section-title">
				<Icon name="GitBranch" size={12} />
				上流/下流ノード
			</h4>

			<div class="relation-group">
				<div class="relation-group-title">上流 ({upstreamNodes.length})</div>
				{#if upstreamNodes.length === 0}
					<div class="empty">上流ノードなし</div>
				{:else}
					<ul class="relation-list">
						{#each upstreamNodes as upstream (upstream.id)}
							<li>
								<button class="relation-item" onclick={() => onNodeSelect?.(upstream.id)}>
									<span
										class="relation-type"
										style="background: {getNodeTypeCSSColor(upstream.node_type)}"
									>
										{getNodeTypeLabel(upstream.node_type)}
									</span>
									<span class="relation-title">{upstream.title}</span>
									<span class="relation-id">{upstream.id}</span>
								</button>
							</li>
						{/each}
					</ul>
				{/if}
			</div>

			<div class="relation-group">
				<div class="relation-group-title">下流 ({downstreamNodes.length})</div>
				{#if downstreamNodes.length === 0}
					<div class="empty">下流ノードなし</div>
				{:else}
					<ul class="relation-list">
						{#each downstreamNodes as downstream (downstream.id)}
							<li>
								<button class="relation-item" onclick={() => onNodeSelect?.(downstream.id)}>
									<span
										class="relation-type"
										style="background: {getNodeTypeCSSColor(downstream.node_type)}"
									>
										{getNodeTypeLabel(downstream.node_type)}
									</span>
									<span class="relation-title">{downstream.title}</span>
									<span class="relation-id">{downstream.id}</span>
								</button>
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		</section>

		<section class="section summary-section">
			<h4 class="section-title">
				<Icon name="BarChart3" size={12} />
				Relation 内訳
			</h4>

			<div class="summary-group">
				<div class="relation-group-title">Incoming</div>
				{#if incomingRelationSummary.length === 0}
					<div class="empty">データなし</div>
				{:else}
					<ul class="summary-list">
						{#each incomingRelationSummary as item (item.key)}
							<li class="summary-item">
								<span class="summary-key">{item.layer} / {item.relation}</span>
								<span class="summary-count">{item.count}</span>
							</li>
						{/each}
					</ul>
				{/if}
			</div>

			<div class="summary-group">
				<div class="relation-group-title">Outgoing</div>
				{#if outgoingRelationSummary.length === 0}
					<div class="empty">データなし</div>
				{:else}
					<ul class="summary-list">
						{#each outgoingRelationSummary as item (item.key)}
							<li class="summary-item">
								<span class="summary-key">{item.layer} / {item.relation}</span>
								<span class="summary-count">{item.count}</span>
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		</section>
	{:else}
		<div class="empty-state">
			<Icon name="Info" size={20} />
			<span>ノードを選択してください</span>
		</div>
	{/if}
</div>

<style>
	.detail-content {
		font-size: 0.8125rem;
		position: relative;
	}

	.close-button {
		position: absolute;
		top: 0;
		right: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		background: rgba(0, 0, 0, 0.3);
		border: 1px solid var(--border-metal);
		border-radius: 4px;
		color: var(--text-secondary);
		cursor: pointer;
		transition:
			background 0.15s ease,
			color 0.15s ease;
	}

	.close-button:hover {
		background: rgba(255, 100, 100, 0.2);
		color: var(--text-primary);
	}

	.section {
		margin-bottom: 16px;
		padding-bottom: 12px;
		border-bottom: 1px solid var(--border-metal);
	}

	.section:last-child {
		margin-bottom: 0;
		padding-bottom: 0;
		border-bottom: none;
	}

	.section-title {
		display: flex;
		align-items: center;
		gap: 6px;
		margin: 0 0 10px 0;
		font-size: 0.6875rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-muted);
		font-weight: 600;
	}

	.node-header {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 10px;
	}

	.status-dot {
		width: 9px;
		height: 9px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.node-heading {
		flex: 1;
		min-width: 0;
	}

	.node-title {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text-primary);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.node-id {
		font-size: 0.6875rem;
		font-family: monospace;
		color: var(--text-muted);
	}

	.type-badge {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 34px;
		padding: 2px 6px;
		border-radius: 3px;
		font-size: 0.625rem;
		font-weight: 600;
		color: #1a1a1a;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.node-info-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 8px;
	}

	.info-item {
		display: flex;
		flex-direction: column;
		gap: 3px;
	}

	.label {
		font-size: 0.625rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-muted);
	}

	.value {
		font-size: 0.75rem;
		color: var(--text-primary);
		word-break: break-word;
	}

	.action-row {
		margin-top: 10px;
	}

	.jump-button {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 5px 10px;
		background: rgba(255, 149, 51, 0.15);
		border: 1px solid rgba(255, 149, 51, 0.35);
		border-radius: 4px;
		color: var(--accent-primary);
		cursor: pointer;
		font-family: inherit;
		font-size: 0.75rem;
	}

	.jump-button:hover {
		background: rgba(255, 149, 51, 0.2);
		border-color: var(--accent-primary);
	}

	.relation-group {
		margin-bottom: 10px;
	}

	.relation-group:last-child {
		margin-bottom: 0;
	}

	.relation-group-title {
		font-size: 0.6875rem;
		font-weight: 600;
		color: var(--text-secondary);
		margin-bottom: 6px;
	}

	.relation-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.relation-item {
		display: flex;
		align-items: center;
		gap: 7px;
		width: 100%;
		padding: 6px 8px;
		background: rgba(0, 0, 0, 0.22);
		border: 1px solid var(--border-metal);
		border-radius: 4px;
		color: var(--text-primary);
		cursor: pointer;
		font-family: inherit;
		font-size: 0.75rem;
		text-align: left;
	}

	.relation-item:hover {
		background: rgba(0, 0, 0, 0.3);
		border-color: var(--border-hover);
	}

	.relation-type {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 28px;
		padding: 1px 5px;
		border-radius: 3px;
		font-size: 0.5625rem;
		font-weight: 600;
		color: #1a1a1a;
		text-transform: uppercase;
	}

	.relation-title {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.relation-id {
		font-family: monospace;
		font-size: 0.625rem;
		color: var(--text-muted);
	}

	.summary-group {
		margin-bottom: 10px;
	}

	.summary-group:last-child {
		margin-bottom: 0;
	}

	.summary-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.summary-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		padding: 5px 8px;
		background: rgba(0, 0, 0, 0.22);
		border: 1px solid var(--border-metal);
		border-radius: 4px;
	}

	.summary-key {
		font-size: 0.6875rem;
		color: var(--text-secondary);
	}

	.summary-count {
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--accent-primary);
	}

	.empty {
		font-size: 0.6875rem;
		color: var(--text-muted);
		padding: 4px 0;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 8px;
		padding: 24px;
		color: var(--text-muted);
		text-align: center;
	}

	.empty-state span {
		font-size: 0.75rem;
	}
</style>
