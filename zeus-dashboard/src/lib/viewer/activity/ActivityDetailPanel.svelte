<script lang="ts">
	// Activity Detail Panel（オーバーレイ用シンプル版）
	// 選択されたノードの詳細を表示
	import type { ActivityItem, ActivityNodeItem, ActivityTransitionItem } from '$lib/types/api';
	import { Icon } from '$lib/components/ui';
	import { navigateToEntity } from '$lib/stores/view';

	interface Props {
		node?: ActivityNodeItem | null;
		activity?: ActivityItem | null;
		onClose?: () => void;
	}
	let { node = null, activity = null, onClose }: Props = $props();

	// UseCase へ遷移
	function handleUseCaseClick(usecaseId: string) {
		navigateToEntity('usecase', 'usecase', usecaseId);
	}

	// ノードタイプのラベル
	const nodeTypeLabels: Record<string, string> = {
		initial: '開始ノード',
		final: '終了ノード',
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

	// 入力遷移を取得
	const incomingTransitions = $derived.by((): ActivityTransitionItem[] => {
		if (!node || !activity) return [];
		return activity.transitions.filter((t) => t.target === node.id);
	});

	// 出力遷移を取得
	const outgoingTransitions = $derived.by((): ActivityTransitionItem[] => {
		if (!node || !activity) return [];
		return activity.transitions.filter((t) => t.source === node.id);
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
	{#if node}
		<!-- ノード詳細 -->
		<div class="entity-header">
			<span class="entity-icon" style="color: {getNodeColor(node.type)}">
				<Icon name={getNodeIcon(node.type)} size={20} />
			</span>
			<div class="entity-info">
				<h3 class="entity-title">{node.name || nodeTypeLabels[node.type]}</h3>
				<span class="entity-id">{node.id}</span>
			</div>
		</div>

		<dl class="detail-list">
			<div class="detail-item">
				<dt>タイプ</dt>
				<dd>
					<span class="badge" style="background: {getNodeColor(node.type)}; color: #1a1a1a;">
						{nodeTypeLabels[node.type] ?? node.type}
					</span>
				</dd>
			</div>

			{#if node.name && node.type !== 'initial' && node.type !== 'final'}
				<div class="detail-item full">
					<dt>名前</dt>
					<dd class="monospace">{node.name}</dd>
				</div>
			{/if}

			<!-- 入力遷移 -->
			{#if incomingTransitions.length > 0}
				<div class="detail-item full">
					<dt>
						<Icon name="ArrowLeft" size={12} />
						入力遷移 ({incomingTransitions.length})
					</dt>
					<dd>
						<ul class="transition-list">
							{#each incomingTransitions as transition}
								<li class="transition-item">
									<span class="transition-source">{getNodeName(transition.source)}</span>
									<Icon name="ArrowRight" size={10} />
									{#if transition.guard}
										<span class="guard-condition">[{transition.guard}]</span>
									{/if}
								</li>
							{/each}
						</ul>
					</dd>
				</div>
			{/if}

			<!-- 出力遷移 -->
			{#if outgoingTransitions.length > 0}
				<div class="detail-item full">
					<dt>
						<Icon name="ArrowRight" size={12} />
						出力遷移 ({outgoingTransitions.length})
					</dt>
					<dd>
						<ul class="transition-list">
							{#each outgoingTransitions as transition}
								<li class="transition-item">
									<Icon name="ArrowRight" size={10} />
									<span class="transition-target">{getNodeName(transition.target)}</span>
									{#if transition.guard}
										<span class="guard-condition">[{transition.guard}]</span>
									{/if}
								</li>
							{/each}
						</ul>
					</dd>
				</div>
			{/if}

			<!-- フロー情報（decision/merge の場合） -->
			{#if node.type === 'decision'}
				<div class="detail-item full info-box">
					<Icon name="Info" size={14} />
					<span>分岐ノードは条件に基づいて制御フローを分岐させます。</span>
				</div>
			{:else if node.type === 'fork'}
				<div class="detail-item full info-box">
					<Icon name="Info" size={14} />
					<span>並列分岐ノードは複数の並行フローを開始します。</span>
				</div>
			{:else if node.type === 'join'}
				<div class="detail-item full info-box">
					<Icon name="Info" size={14} />
					<span>並列合流ノードは複数の並行フローを同期します。</span>
				</div>
			{/if}
		</dl>

		<!-- アクティビティ情報 -->
		{#if activity}
			<div class="activity-info">
				<h4 class="section-title">
					<Icon name="FileText" size={12} />
					アクティビティ情報
				</h4>
				<div class="activity-summary">
					<div class="summary-item">
						<span class="summary-label">タイトル</span>
						<span class="summary-value">{activity.title}</span>
					</div>
					{#if activity.usecase_id}
						<div class="summary-item">
							<span class="summary-label">ユースケース</span>
							<button
								class="link-button"
								onclick={() => {
									if (activity?.usecase_id) {
										handleUseCaseClick(activity.usecase_id);
									}
								}}
								title="UseCase ビューで表示"
							>
								<Icon name="ExternalLink" size={10} />
								<span class="monospace">{activity.usecase_id}</span>
							</button>
						</div>
					{/if}
					<div class="summary-item">
						<span class="summary-label">総ノード数</span>
						<span class="summary-value">{activity.nodes.length}</span>
					</div>
					<div class="summary-item">
						<span class="summary-label">総遷移数</span>
						<span class="summary-value">{activity.transitions.length}</span>
					</div>
				</div>
			</div>
		{/if}
	{:else}
		<!-- 未選択状態 -->
		<div class="empty-state">
			<Icon name="Info" size={20} />
			<span>ノードを選択してください</span>
		</div>
	{/if}
</div>

<style>
	.detail-content {
		font-size: 0.8125rem;
	}

	.entity-header {
		display: flex;
		align-items: flex-start;
		gap: 10px;
		margin-bottom: 12px;
		padding-bottom: 10px;
		border-bottom: 1px solid var(--border-metal);
	}

	.entity-icon {
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.entity-info {
		flex: 1;
		min-width: 0;
	}

	.entity-title {
		margin: 0 0 2px 0;
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text-primary);
		line-height: 1.3;
	}

	.entity-id {
		font-size: 0.6875rem;
		color: var(--text-muted);
		font-family: monospace;
	}

	.detail-list {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 10px;
	}

	.detail-item {
		display: flex;
		flex-direction: column;
		gap: 3px;
	}

	.detail-item.full {
		grid-column: 1 / -1;
	}

	.detail-item dt {
		display: flex;
		align-items: center;
		gap: 4px;
		font-size: 0.625rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-muted);
		font-weight: 500;
	}

	.detail-item dd {
		margin: 0;
		font-size: 0.8125rem;
		color: var(--text-primary);
	}

	.badge {
		display: inline-block;
		padding: 2px 6px;
		border-radius: 3px;
		font-size: 0.6875rem;
		font-weight: 500;
	}

	.monospace {
		font-family: monospace;
		font-size: 0.75rem;
		background: rgba(0, 0, 0, 0.3);
		padding: 2px 6px;
		border-radius: 3px;
	}

	.transition-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.transition-item {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 5px 8px;
		background: rgba(0, 0, 0, 0.25);
		border: 1px solid var(--border-metal);
		border-radius: 4px;
		font-size: 0.75rem;
	}

	.transition-source,
	.transition-target {
		color: var(--text-primary);
	}

	.guard-condition {
		font-size: 0.65rem;
		color: var(--accent-primary);
		font-style: italic;
		margin-left: auto;
	}

	.info-box {
		display: flex;
		align-items: flex-start;
		gap: 8px;
		padding: 8px 10px;
		background: rgba(255, 149, 51, 0.1);
		border: 1px solid rgba(255, 149, 51, 0.2);
		border-radius: 4px;
		font-size: 0.75rem;
		color: var(--text-secondary);
		margin-top: 8px;
	}

	.activity-info {
		margin-top: 16px;
		padding-top: 12px;
		border-top: 1px solid var(--border-metal);
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

	.activity-summary {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.summary-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: 0.75rem;
	}

	.summary-label {
		color: var(--text-muted);
	}

	.summary-value {
		color: var(--text-primary);
	}

	/* リンクボタンスタイル */
	.link-button {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 2px 6px;
		background: rgba(255, 149, 51, 0.1);
		border: 1px solid rgba(255, 149, 51, 0.3);
		border-radius: 3px;
		color: var(--accent-primary);
		cursor: pointer;
		font-family: inherit;
		font-size: 0.75rem;
		transition: background 0.15s ease, border-color 0.15s ease;
	}

	.link-button:hover {
		background: rgba(255, 149, 51, 0.2);
		border-color: var(--accent-primary);
	}

	.link-button .monospace {
		background: transparent;
		padding: 0;
		color: inherit;
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
