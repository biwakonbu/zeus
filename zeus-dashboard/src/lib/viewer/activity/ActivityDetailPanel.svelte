<script lang="ts">
	// Activity Detail Panel（オーバーレイ用）
	// 3セクション構成: アクティビティ情報、選択中のノード、ノード一覧
	import type { ActivityItem, ActivityNodeItem, ActivityTransitionItem } from '$lib/types/api';
	import { Icon } from '$lib/components/ui';
	import { navigateToEntity } from '$lib/stores/view';

	interface Props {
		activity?: ActivityItem | null;
		selectedNode?: ActivityNodeItem | null;
		onClose?: () => void;
		onNodeClick?: (node: ActivityNodeItem) => void;
	}
	let { activity = null, selectedNode = null, onClose, onNodeClick }: Props = $props();

	// UseCase へ遷移
	function handleUseCaseClick(usecaseId: string) {
		navigateToEntity('usecase', 'usecase', usecaseId);
	}

	// ノードタイプのラベル
	const nodeTypeLabels: Record<string, string> = {
		initial: '開始',
		final: '終了',
		action: 'アクション',
		decision: '分岐',
		merge: '合流',
		fork: '並列分岐',
		join: '並列合流'
	};

	// ノードタイプのアイコン
	function getNodeIcon(type: string): string {
		const icons: Record<string, string> = {
			initial: 'Circle',
			final: 'CircleDot',
			action: 'Square',
			decision: 'Diamond',
			merge: 'Diamond',
			fork: 'Minus',
			join: 'Minus'
		};
		return icons[type] ?? 'HelpCircle';
	}

	// ノードタイプの色
	function getNodeColor(type: string): string {
		const colors: Record<string, string> = {
			initial: 'var(--status-good)',
			final: 'var(--status-poor)',
			action: 'var(--accent-primary)',
			decision: 'var(--status-fair)',
			merge: 'var(--status-fair)',
			fork: '#3b82f6',
			join: '#3b82f6'
		};
		return colors[type] ?? 'var(--text-secondary)';
	}

	// ステータス色取得
	function getStatusColor(status: string): string {
		const colors: Record<string, string> = {
			active: 'var(--status-good)',
			draft: 'var(--status-fair)',
			deprecated: 'var(--text-muted)'
		};
		return colors[status] ?? 'var(--text-secondary)';
	}

	// ステータスラベル
	function getStatusLabel(status: string): string {
		const labels: Record<string, string> = {
			active: 'アクティブ',
			draft: '下書き',
			deprecated: '非推奨'
		};
		return labels[status] ?? status;
	}

	// 入力遷移を取得
	const incomingTransitions = $derived.by((): ActivityTransitionItem[] => {
		if (!selectedNode || !activity) return [];
		return activity.transitions.filter((t) => t.target === selectedNode.id);
	});

	// 出力遷移を取得
	const outgoingTransitions = $derived.by((): ActivityTransitionItem[] => {
		if (!selectedNode || !activity) return [];
		return activity.transitions.filter((t) => t.source === selectedNode.id);
	});

	// ノード名を取得するヘルパー
	function getNodeName(nodeId: string): string {
		if (!activity) return nodeId;
		const foundNode = activity.nodes.find((n) => n.id === nodeId);
		if (!foundNode) return nodeId;
		return foundNode.name || nodeTypeLabels[foundNode.type] || nodeId;
	}
</script>

<div class="detail-content">
	{#if activity}
		<!-- セクション1: アクティビティ情報（常に表示） -->
		<section class="section activity-section">
			<h4 class="section-title">
				<Icon name="FileText" size={12} />
				アクティビティ情報
			</h4>
			<div class="activity-info">
				<div class="info-row">
					<span class="info-label">タイトル</span>
					<span class="info-value title">{activity.title}</span>
				</div>
				<div class="info-row">
					<span class="info-label">ID</span>
					<span class="info-value monospace">{activity.id}</span>
				</div>
				<div class="info-row">
					<span class="info-label">ステータス</span>
					<span class="status-badge" style="background: {getStatusColor(activity.status)}">
						{getStatusLabel(activity.status)}
					</span>
				</div>
				{#if activity.usecase_id}
					<div class="info-row">
						<span class="info-label">UseCase</span>
						<button
							class="usecase-link"
							onclick={() => {
								if (activity?.usecase_id) {
									handleUseCaseClick(activity.usecase_id);
								}
							}}
							title="UseCase ビューで表示"
						>
							<span class="monospace">{activity.usecase_id}</span>
							<Icon name="ExternalLink" size={10} />
						</button>
					</div>
				{/if}
				<div class="info-row stats">
					<span class="stat-item">
						<Icon name="Circle" size={10} />
						ノード数: {activity.nodes.length}
					</span>
					<span class="stat-item">
						<Icon name="ArrowRight" size={10} />
						遷移数: {activity.transitions.length}
					</span>
				</div>
			</div>
		</section>

		<!-- セクション2: 選択中のノード（ノード選択時のみ表示） -->
		{#if selectedNode}
			<section class="section selected-node-section">
				<h4 class="section-title">
					<span class="node-icon" style="color: {getNodeColor(selectedNode.type)}">
						<Icon name={getNodeIcon(selectedNode.type)} size={12} />
					</span>
					選択中のノード
				</h4>
				<div class="selected-node-info">
					<div class="node-header">
						<span class="node-name">{selectedNode.name || nodeTypeLabels[selectedNode.type]}</span>
						<span class="node-type-badge" style="background: {getNodeColor(selectedNode.type)}">
							{nodeTypeLabels[selectedNode.type] ?? selectedNode.type}
						</span>
					</div>
					<div class="node-id">{selectedNode.id}</div>
					<div class="transition-stats">
						<span class="stat">
							<Icon name="ArrowLeft" size={10} />
							入力遷移: {incomingTransitions.length}
						</span>
						<span class="stat">
							<Icon name="ArrowRight" size={10} />
							出力遷移: {outgoingTransitions.length}
						</span>
					</div>

					<!-- 入力遷移詳細 -->
					{#if incomingTransitions.length > 0}
						<div class="transition-detail">
							<span class="transition-label">入力元:</span>
							<ul class="transition-list">
								{#each incomingTransitions as transition}
									<li class="transition-item">
										<span class="transition-node">{getNodeName(transition.source)}</span>
										{#if transition.guard}
											<span class="guard-condition">[{transition.guard}]</span>
										{/if}
									</li>
								{/each}
							</ul>
						</div>
					{/if}

					<!-- 出力遷移詳細 -->
					{#if outgoingTransitions.length > 0}
						<div class="transition-detail">
							<span class="transition-label">出力先:</span>
							<ul class="transition-list">
								{#each outgoingTransitions as transition}
									<li class="transition-item">
										<span class="transition-node">{getNodeName(transition.target)}</span>
										{#if transition.guard}
											<span class="guard-condition">[{transition.guard}]</span>
										{/if}
									</li>
								{/each}
							</ul>
						</div>
					{/if}
				</div>
			</section>
		{/if}

		<!-- セクション3: ノード一覧（常に展開） -->
		<section class="section node-list-section">
			<h4 class="section-title">
				<Icon name="List" size={12} />
				ノード一覧 ({activity.nodes.length})
			</h4>
			<ul class="node-list">
				{#each activity.nodes as node (node.id)}
					{@const isSelected = selectedNode?.id === node.id}
					<li class="node-list-item" class:selected={isSelected}>
						<button
							class="node-button"
							onclick={() => onNodeClick?.(node)}
							title={node.id}
						>
							<span class="node-icon" style="color: {getNodeColor(node.type)}">
								<Icon name={getNodeIcon(node.type)} size={12} />
							</span>
							<span class="node-name">{node.name || nodeTypeLabels[node.type]}</span>
							{#if isSelected}
								<span class="selected-indicator">選択中</span>
							{/if}
						</button>
					</li>
				{/each}
			</ul>
		</section>
	{:else}
		<!-- 未選択状態 -->
		<div class="empty-state">
			<Icon name="Info" size={20} />
			<span>アクティビティを選択してください</span>
		</div>
	{/if}
</div>

<style>
	.detail-content {
		font-size: 0.8125rem;
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

	/* セクション1: アクティビティ情報 */
	.activity-info {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.info-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: 0.75rem;
	}

	.info-row.stats {
		flex-wrap: wrap;
		gap: 8px;
		justify-content: flex-start;
		margin-top: 4px;
	}

	.info-label {
		color: var(--text-muted);
		flex-shrink: 0;
	}

	.info-value {
		color: var(--text-primary);
		text-align: right;
	}

	.info-value.title {
		font-weight: 600;
	}

	.monospace {
		font-family: monospace;
		font-size: 0.6875rem;
		background: rgba(0, 0, 0, 0.3);
		padding: 2px 6px;
		border-radius: 3px;
	}

	.status-badge {
		display: inline-block;
		padding: 2px 6px;
		border-radius: 3px;
		font-size: 0.625rem;
		font-weight: 500;
		color: #1a1a1a;
	}

	.usecase-link {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 3px 8px;
		background: rgba(255, 149, 51, 0.1);
		border: 1px solid rgba(255, 149, 51, 0.3);
		border-radius: 3px;
		color: var(--accent-primary);
		cursor: pointer;
		font-family: inherit;
		font-size: 0.6875rem;
		transition: background 0.15s ease, border-color 0.15s ease;
	}

	.usecase-link:hover {
		background: rgba(255, 149, 51, 0.2);
		border-color: var(--accent-primary);
	}

	.usecase-link .monospace {
		background: transparent;
		padding: 0;
		color: inherit;
	}

	.stat-item {
		display: flex;
		align-items: center;
		gap: 4px;
		color: var(--text-muted);
		font-size: 0.6875rem;
	}

	/* セクション2: 選択中のノード */
	.selected-node-info {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.node-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}

	.node-name {
		font-weight: 600;
		font-size: 0.8125rem;
		color: var(--text-primary);
	}

	.node-type-badge {
		display: inline-block;
		padding: 2px 6px;
		border-radius: 3px;
		font-size: 0.5625rem;
		font-weight: 500;
		color: #1a1a1a;
		text-transform: uppercase;
	}

	.node-id {
		font-family: monospace;
		font-size: 0.625rem;
		color: var(--text-muted);
	}

	.transition-stats {
		display: flex;
		gap: 12px;
		margin-top: 4px;
	}

	.transition-stats .stat {
		display: flex;
		align-items: center;
		gap: 4px;
		font-size: 0.6875rem;
		color: var(--text-secondary);
	}

	.transition-detail {
		margin-top: 8px;
		padding: 8px;
		background: rgba(0, 0, 0, 0.2);
		border-radius: 4px;
	}

	.transition-label {
		display: block;
		font-size: 0.625rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-muted);
		margin-bottom: 4px;
	}

	.transition-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.transition-item {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 0.6875rem;
	}

	.transition-node {
		color: var(--text-primary);
	}

	.guard-condition {
		font-size: 0.625rem;
		color: var(--accent-primary);
		font-style: italic;
	}

	/* セクション3: ノード一覧 */
	.node-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
		max-height: 200px;
		overflow-y: auto;
	}

	.node-list-item {
		border-radius: 3px;
	}

	.node-list-item.selected {
		background: rgba(255, 149, 51, 0.15);
	}

	.node-button {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		padding: 6px 8px;
		background: transparent;
		border: none;
		border-radius: 3px;
		cursor: pointer;
		text-align: left;
		font-family: inherit;
		font-size: 0.75rem;
		color: var(--text-primary);
		transition: background 0.1s ease;
	}

	.node-button:hover {
		background: rgba(255, 255, 255, 0.05);
	}

	.node-list-item.selected .node-button {
		background: transparent;
	}

	.node-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.node-button .node-name {
		flex: 1;
		font-weight: normal;
		font-size: 0.75rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.selected-indicator {
		font-size: 0.5625rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--accent-primary);
		background: rgba(255, 149, 51, 0.2);
		padding: 2px 5px;
		border-radius: 2px;
		flex-shrink: 0;
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
