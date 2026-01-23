<script lang="ts">
	// UseCase View Panel（オーバーレイ用シンプル版）
	// 選択された Actor または UseCase の詳細を表示
	import type { ActorItem, UseCaseItem } from '$lib/types/api';
	import { Icon } from '$lib/components/ui';

	interface Props {
		actor?: ActorItem | null;
		usecase?: UseCaseItem | null;
		onClose?: () => void;
	}
	let { actor = null, usecase = null, onClose }: Props = $props();

	// Actor タイプのラベル
	const actorTypeLabels: Record<string, string> = {
		human: '人間',
		system: 'システム',
		time: '時間',
		device: 'デバイス',
		external: '外部'
	};

	// UseCase ステータスのラベル
	const statusLabels: Record<string, string> = {
		draft: '下書き',
		active: 'アクティブ',
		deprecated: '非推奨'
	};

	// Role のラベル
	const roleLabels: Record<string, string> = {
		primary: '主',
		secondary: '副'
	};

	// Relation タイプのラベル
	const relationLabels: Record<string, string> = {
		include: 'include',
		extend: 'extend',
		generalize: 'generalize'
	};

	// Actor タイプのアイコン
	function getActorIcon(type: string): string {
		const icons: Record<string, string> = {
			human: 'User',
			system: 'Server',
			time: 'Clock',
			device: 'Smartphone',
			external: 'Globe'
		};
		return icons[type] ?? 'HelpCircle';
	}

	// ステータスカラー
	function getStatusColor(status: string): string {
		const colors: Record<string, string> = {
			active: 'var(--status-good)',
			draft: 'var(--status-fair)',
			deprecated: 'var(--text-muted)'
		};
		return colors[status] ?? 'var(--text-secondary)';
	}
</script>

<div class="detail-content">
	{#if actor}
		<!-- Actor 詳細 -->
		<div class="entity-header">
			<span class="entity-icon">
				<Icon name={getActorIcon(actor.type)} size={20} />
			</span>
			<div class="entity-info">
				<h3 class="entity-title">{actor.title}</h3>
				<span class="entity-id">{actor.id}</span>
			</div>
		</div>

		<dl class="detail-list">
			<div class="detail-item">
				<dt>タイプ</dt>
				<dd>
					<span class="badge">{actorTypeLabels[actor.type] ?? actor.type}</span>
				</dd>
			</div>
			{#if actor.description}
				<div class="detail-item full">
					<dt>説明</dt>
					<dd class="description">{actor.description}</dd>
				</div>
			{/if}
		</dl>
	{:else if usecase}
		<!-- UseCase 詳細 -->
		<div class="entity-header">
			<span class="status-dot" style="background: {getStatusColor(usecase.status)}"></span>
			<div class="entity-info">
				<h3 class="entity-title">{usecase.title}</h3>
				<span class="entity-id">{usecase.id}</span>
			</div>
		</div>

		<dl class="detail-list">
			<div class="detail-item">
				<dt>ステータス</dt>
				<dd>
					<span class="badge" style="background: {getStatusColor(usecase.status)}; color: #1a1a1a;">
						{statusLabels[usecase.status] ?? usecase.status}
					</span>
				</dd>
			</div>

			{#if usecase.objective_id}
				<div class="detail-item">
					<dt>目標</dt>
					<dd class="monospace">{usecase.objective_id}</dd>
				</div>
			{/if}

			{#if usecase.description}
				<div class="detail-item full">
					<dt>説明</dt>
					<dd class="description">{usecase.description}</dd>
				</div>
			{/if}

			{#if usecase.actors && usecase.actors.length > 0}
				<div class="detail-item full">
					<dt>関連アクター</dt>
					<dd>
						<ul class="relation-list">
							{#each usecase.actors as actorRef}
								<li class="relation-item">
									<Icon name="User" size={12} />
									<span class="relation-name">{actorRef.actor_id}</span>
									<span class="role-badge">{roleLabels[actorRef.role] ?? actorRef.role}</span>
								</li>
							{/each}
						</ul>
					</dd>
				</div>
			{/if}

			{#if usecase.relations && usecase.relations.length > 0}
				<div class="detail-item full">
					<dt>関係</dt>
					<dd>
						<ul class="relation-list">
							{#each usecase.relations as relation}
								<li class="relation-item">
									<span class="relation-type">{relationLabels[relation.type] ?? relation.type}</span>
									<span class="relation-name">{relation.target_id}</span>
									{#if relation.condition}
										<span class="condition">[{relation.condition}]</span>
									{/if}
								</li>
							{/each}
						</ul>
					</dd>
				</div>
			{/if}
		</dl>
	{:else}
		<!-- 未選択状態 -->
		<div class="empty-state">
			<Icon name="Info" size={20} />
			<span>要素を選択してください</span>
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
		color: var(--text-secondary);
	}

	.status-dot {
		width: 10px;
		height: 10px;
		border-radius: 50%;
		flex-shrink: 0;
		margin-top: 5px;
		box-shadow: 0 0 6px currentColor;
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
		background: var(--bg-tertiary);
	}

	.monospace {
		font-family: monospace;
		font-size: 0.75rem;
		background: rgba(0, 0, 0, 0.3);
		padding: 2px 6px;
		border-radius: 3px;
	}

	.description {
		font-size: 0.75rem;
		line-height: 1.5;
		color: var(--text-secondary);
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
		gap: 6px;
		padding: 5px 8px;
		background: rgba(0, 0, 0, 0.25);
		border: 1px solid var(--border-metal);
		border-radius: 4px;
		font-size: 0.75rem;
	}

	.relation-name {
		font-family: monospace;
		font-size: 0.7rem;
	}

	.role-badge {
		margin-left: auto;
		font-size: 0.625rem;
		color: var(--accent-primary);
		background: rgba(255, 149, 51, 0.15);
		padding: 1px 5px;
		border-radius: 2px;
	}

	.relation-type {
		font-size: 0.625rem;
		color: var(--accent-primary);
		padding: 1px 5px;
		background: rgba(255, 149, 51, 0.15);
		border-radius: 2px;
	}

	.condition {
		font-size: 0.65rem;
		color: var(--text-muted);
		font-style: italic;
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
