<script lang="ts">
	import type { GraphEdge, GraphNode, GraphNodeType, UnifiedGraphGroupItem } from '$lib/types/api';
	import { Icon } from '$lib/components/ui';
	import { navigateToEntity } from '$lib/stores/view';
	import { getNodeTypeCSSColor, getNodeTypeLabel } from '../config/nodeTypes';

	interface Props {
		node?: GraphNode | null;
		nodes: GraphNode[];
		edges: GraphEdge[];
		group?: UnifiedGraphGroupItem | null;
		groups?: UnifiedGraphGroupItem[];
		onClose?: () => void;
		onNodeSelect?: (nodeId: string) => void;
		onGroupSelect?: (groupId: string) => void;
	}

	type RelationSummaryItem = {
		key: string;
		layer: string;
		relation: string;
		count: number;
	};

	type RelatedGroupInfo = {
		group: UnifiedGraphGroupItem;
		edgeCount: number;
		outgoing: number;
		incoming: number;
	};

	let {
		node = null,
		nodes,
		edges,
		group = null,
		groups = [],
		onClose = undefined,
		onNodeSelect = undefined,
		onGroupSelect = undefined
	}: Props = $props();

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

	// ノードが属する Objective グループを取得
	const nodeObjectiveGroup = $derived.by(() => {
		const currentNode = node;
		if (!currentNode || !groups || groups.length === 0) return null;
		return groups.find((g) => g.node_ids.includes(currentNode.id)) ?? null;
	});

	// グループ内のノード一覧
	const groupMemberNodes = $derived.by(() => {
		if (!group) return [] as GraphNode[];
		const memberIds = new Set(group.node_ids);
		return nodes.filter((n) => memberIds.has(n.id));
	});

	// タイプ別ノード集計
	const groupNodesByType = $derived.by(() => {
		const map = new Map<GraphNodeType, GraphNode[]>();
		for (const n of groupMemberNodes) {
			const arr = map.get(n.node_type) || [];
			arr.push(n);
			map.set(n.node_type, arr);
		}
		return map;
	});

	// ステータス別集計
	const groupStatusCounts = $derived.by(() => {
		const map = new Map<string, number>();
		for (const n of groupMemberNodes) {
			map.set(n.status, (map.get(n.status) || 0) + 1);
		}
		return map;
	});

	// Activity 進捗率
	const groupProgress = $derived.by(() => {
		const activities = groupMemberNodes.filter((n) => n.node_type === 'activity');
		if (activities.length === 0) return null;
		const completed = activities.filter((n) => n.status === 'deprecated').length;
		return {
			completed,
			total: activities.length,
			rate: Math.round((completed / activities.length) * 100)
		};
	});

	// 関連グループ
	const relatedGroups = $derived.by((): RelatedGroupInfo[] => {
		if (!group || !groups || groups.length === 0) return [];
		const memberSet = new Set(group.node_ids);
		// 各グループの node_ids から逆引きマップ構築
		const nodeToGroup = new Map<string, string>();
		for (const g of groups) {
			if (g.id === group.id) continue;
			for (const nid of g.node_ids) nodeToGroup.set(nid, g.id);
		}
		// エッジ走査でグループ間接続を検出
		const stats = new Map<string, { outgoing: number; incoming: number }>();
		for (const edge of edges) {
			const fromIn = memberSet.has(edge.from);
			const toIn = memberSet.has(edge.to);
			if (fromIn === toIn) continue;
			const otherId = fromIn ? nodeToGroup.get(edge.to) : nodeToGroup.get(edge.from);
			if (!otherId) continue;
			const s = stats.get(otherId) || { outgoing: 0, incoming: 0 };
			if (fromIn) s.outgoing++;
			else s.incoming++;
			stats.set(otherId, s);
		}
		// 結果構築
		const groupById = new Map(groups.map((g) => [g.id, g]));
		return Array.from(stats.entries())
			.map(([id, s]) => {
				const g = groupById.get(id);
				if (!g) return null;
				return {
					group: g,
					edgeCount: s.outgoing + s.incoming,
					outgoing: s.outgoing,
					incoming: s.incoming
				};
			})
			.filter((r): r is RelatedGroupInfo => r !== null)
			.sort((a, b) => b.edgeCount - a.edgeCount);
	});
</script>

<div class="detail-content">
	{#if onClose}
		<button class="close-button" onclick={onClose} title="閉じる">
			<Icon name="X" size={16} />
		</button>
	{/if}

	{#if group && !node}
		<!-- グループ選択モード -->
		<section class="section node-section">
			<h4 class="section-title">
				<Icon name="Layers" size={12} />
				Objective
			</h4>
			<div class="node-header">
				<span class="status-dot" style="background: {getStatusColor(group.status)}"></span>
				<div class="node-heading">
					<div class="node-title">{group.title}</div>
					<div class="node-id">{group.id}</div>
				</div>
			</div>

			<div class="node-info-grid">
				<div class="info-item">
					<span class="label">Status</span>
					<span class="value">{getStatusLabel(group.status)}</span>
				</div>
				<div class="info-item">
					<span class="label">Nodes</span>
					<span class="value">{group.node_ids.length}</span>
				</div>
				{#if group.owner}
					<div class="info-item">
						<span class="label">Owner</span>
						<span class="value">{group.owner}</span>
					</div>
				{/if}
			</div>

			{#if group.tags && group.tags.length > 0}
				<div class="tags-row">
					{#each group.tags as tag}
						<span class="tag-badge">{tag}</span>
					{/each}
				</div>
			{/if}

			{#if group.description}
				<div class="group-description">
					<span class="label">Description</span>
					<p class="description-text">{group.description}</p>
				</div>
			{/if}

			{#if group.goals && group.goals.length > 0}
				<div class="group-goals">
					<span class="label">Goals</span>
					<ul class="goals-list">
						{#each group.goals as goal}
							<li>{goal}</li>
						{/each}
					</ul>
				</div>
			{/if}
		</section>

		<!-- グループ統計 -->
		{#if groupMemberNodes.length > 0}
			<section class="section stats-section">
				<h4 class="section-title">
					<Icon name="BarChart3" size={12} />
					グループ統計
				</h4>

				<!-- タイプ別カウント -->
				<div class="stats-grid">
					{#each Array.from(groupNodesByType.entries()) as [type, typeNodes] (type)}
						<div class="stat-item">
							<span class="stat-dot" style="background: {getNodeTypeCSSColor(type)}"></span>
							<span class="stat-label">{getNodeTypeLabel(type)}</span>
							<span class="stat-value">{typeNodes.length}</span>
						</div>
					{/each}
				</div>

				<!-- ステータス別分布 -->
				{#if groupStatusCounts.size > 0}
					<div class="status-distribution">
						<div class="relation-group-title">ステータス分布</div>
						<div class="stats-grid">
							{#each Array.from(groupStatusCounts.entries()) as [status, count] (status)}
								<div class="stat-item">
									<span class="stat-dot" style="background: {getStatusColor(status)}"></span>
									<span class="stat-label">{getStatusLabel(status)}</span>
									<span class="stat-value">{count}</span>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<!-- 進捗バー -->
				{#if groupProgress}
					<div class="progress-section">
						<div class="relation-group-title">Activity 進捗</div>
						<div class="progress-bar-container">
							<div class="progress-bar" style="width: {groupProgress.rate}%"></div>
						</div>
						<div class="progress-text">
							{groupProgress.completed}/{groupProgress.total} ({groupProgress.rate}%)
						</div>
					</div>
				{/if}
			</section>
		{/if}

		<!-- 関連グループ -->
		{#if relatedGroups.length > 0}
			<section class="section relation-section">
				<h4 class="section-title">
					<Icon name="Link" size={12} />
					関連 Objective ({relatedGroups.length})
				</h4>
				<ul class="relation-list">
					{#each relatedGroups as related (related.group.id)}
						<li>
							<button class="relation-item" onclick={() => onGroupSelect?.(related.group.id)}>
								<span class="status-dot" style="background: {getStatusColor(related.group.status)}"></span>
								<span class="relation-title">{related.group.title}</span>
								<span class="edge-count-badge" title="接続数: 出力 {related.outgoing} / 入力 {related.incoming}">
									{related.edgeCount}
								</span>
							</button>
						</li>
					{/each}
				</ul>
			</section>
		{/if}

		<!-- 所属ノード一覧（タイプ別グルーピング） -->
		{#if groupMemberNodes.length > 0}
			<section class="section relation-section">
				<h4 class="section-title">
					<Icon name="GitBranch" size={12} />
					所属ノード ({groupMemberNodes.length})
				</h4>
				{#each Array.from(groupNodesByType.entries()) as [type, typeNodes] (type)}
					<div class="relation-group">
						<div class="relation-group-title">
							<span class="type-indicator" style="background: {getNodeTypeCSSColor(type)}"></span>
							{getNodeTypeLabel(type)} ({typeNodes.length})
						</div>
						<ul class="relation-list">
							{#each typeNodes as member (member.id)}
								<li>
									<button class="relation-item" onclick={() => onNodeSelect?.(member.id)}>
										<span class="status-dot-small" style="background: {getStatusColor(member.status)}"></span>
										<span class="relation-title">{member.title}</span>
										<span class="relation-id">{member.id}</span>
									</button>
								</li>
							{/each}
						</ul>
					</div>
				{/each}
			</section>
		{/if}
	{:else if node}
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

		{#if nodeObjectiveGroup}
			<section class="section objective-section">
				<h4 class="section-title">
					<Icon name="Layers" size={12} />
					Objective
				</h4>
				<div class="objective-card">
					<span class="status-dot" style="background: {getStatusColor(nodeObjectiveGroup.status)}"></span>
					<div class="node-heading">
						<div class="node-title">{nodeObjectiveGroup.title}</div>
						<div class="node-id">{nodeObjectiveGroup.id}</div>
					</div>
				</div>
			</section>
		{/if}

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

	.status-dot-small {
		width: 7px;
		height: 7px;
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

	.tags-row {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
		margin-top: 8px;
	}

	.tag-badge {
		display: inline-flex;
		align-items: center;
		padding: 2px 8px;
		background: rgba(100, 149, 237, 0.15);
		border: 1px solid rgba(100, 149, 237, 0.3);
		border-radius: 10px;
		font-size: 0.625rem;
		color: var(--text-secondary);
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

	/* 統計セクション */
	.stats-grid {
		display: flex;
		flex-direction: column;
		gap: 4px;
		margin-bottom: 8px;
	}

	.stat-item {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 4px 8px;
		background: rgba(0, 0, 0, 0.22);
		border: 1px solid var(--border-metal);
		border-radius: 4px;
	}

	.stat-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.stat-label {
		flex: 1;
		font-size: 0.6875rem;
		color: var(--text-secondary);
	}

	.stat-value {
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--accent-primary);
	}

	.status-distribution {
		margin-top: 4px;
	}

	.progress-section {
		margin-top: 8px;
	}

	.progress-bar-container {
		width: 100%;
		height: 6px;
		background: rgba(0, 0, 0, 0.3);
		border-radius: 3px;
		margin-top: 4px;
		overflow: hidden;
	}

	.progress-bar {
		height: 100%;
		background: var(--task-completed, #4caf50);
		border-radius: 3px;
		transition: width 0.3s ease;
	}

	.progress-text {
		font-size: 0.6875rem;
		color: var(--text-secondary);
		margin-top: 3px;
		text-align: right;
	}

	/* 関連グループ */
	.edge-count-badge {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 20px;
		padding: 1px 5px;
		background: rgba(255, 149, 51, 0.2);
		border-radius: 10px;
		font-size: 0.625rem;
		font-weight: 600;
		color: var(--accent-primary);
	}

	.type-indicator {
		display: inline-block;
		width: 6px;
		height: 6px;
		border-radius: 50%;
	}

	.relation-group {
		margin-bottom: 10px;
	}

	.relation-group:last-child {
		margin-bottom: 0;
	}

	.relation-group-title {
		display: flex;
		align-items: center;
		gap: 4px;
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

	.group-description {
		margin-top: 8px;
	}

	.description-text {
		margin: 4px 0 0 0;
		font-size: 0.75rem;
		color: var(--text-secondary);
		line-height: 1.4;
	}

	.group-goals {
		margin-top: 8px;
	}

	.goals-list {
		margin: 4px 0 0 0;
		padding-left: 16px;
		font-size: 0.75rem;
		color: var(--text-secondary);
		line-height: 1.5;
	}

	.goals-list li {
		margin-bottom: 2px;
	}

	.objective-card {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px;
		background: rgba(0, 0, 0, 0.22);
		border: 1px solid var(--border-metal);
		border-radius: 4px;
	}
</style>
