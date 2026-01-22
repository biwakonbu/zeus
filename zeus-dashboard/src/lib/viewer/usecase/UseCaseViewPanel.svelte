<script lang="ts">
	// UseCase View Panel
	// 選択された Actor または UseCase の詳細を表示
	import type { ActorItem, UseCaseItem } from '$lib/types/api';
	import { Icon, Panel } from '$lib/components/ui';

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
		primary: '主アクター',
		secondary: '副アクター'
	};

	// Relation タイプのラベル
	const relationLabels: Record<string, string> = {
		include: '<<include>>',
		extend: '<<extend>>',
		generalize: '<<generalize>>'
	};

	// Actor タイプのアイコン
	function getActorIcon(type: string): string {
		switch (type) {
			case 'human':
				return 'User';
			case 'system':
				return 'Server';
			case 'time':
				return 'Clock';
			case 'device':
				return 'Smartphone';
			case 'external':
				return 'Globe';
			default:
				return 'HelpCircle';
		}
	}

	// ステータスカラー
	function getStatusColor(status: string): string {
		switch (status) {
			case 'active':
				return 'var(--status-good)';
			case 'draft':
				return 'var(--status-fair)';
			case 'deprecated':
				return 'var(--text-muted)';
			default:
				return 'var(--text-secondary)';
		}
	}
</script>

<div class="usecase-panel">
	{#if actor}
		<!-- Actor 詳細 -->
		<Panel title="アクター詳細">
			<div class="panel-header">
				<Icon name={getActorIcon(actor.type)} size={24} />
				<div class="header-info">
					<h3>{actor.title}</h3>
					<span class="id">{actor.id}</span>
				</div>
				{#if onClose}
					<button class="close-btn" onclick={onClose} aria-label="閉じる">
						<Icon name="X" size={16} />
					</button>
				{/if}
			</div>

			<dl class="detail-list">
				<div class="detail-item">
					<dt>タイプ</dt>
					<dd>
						<span class="badge badge-type">{actorTypeLabels[actor.type] ?? actor.type}</span>
					</dd>
				</div>
				{#if actor.description}
					<div class="detail-item full">
						<dt>説明</dt>
						<dd class="description">{actor.description}</dd>
					</div>
				{/if}
			</dl>
		</Panel>
	{:else if usecase}
		<!-- UseCase 詳細 -->
		<Panel title="ユースケース詳細">
			<div class="panel-header">
				<div
					class="status-indicator"
					style="background: {getStatusColor(usecase.status)}"
				></div>
				<div class="header-info">
					<h3>{usecase.title}</h3>
					<span class="id">{usecase.id}</span>
				</div>
				{#if onClose}
					<button class="close-btn" onclick={onClose} aria-label="閉じる">
						<Icon name="X" size={16} />
					</button>
				{/if}
			</div>

			<dl class="detail-list">
				<div class="detail-item">
					<dt>ステータス</dt>
					<dd>
						<span
							class="badge"
							style="background: {getStatusColor(usecase.status)}"
						>
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
										<Icon name="User" size={14} />
										<span>{actorRef.actor_id}</span>
										<span class="role-badge">{roleLabels[actorRef.role] ?? actorRef.role}</span>
									</li>
								{/each}
							</ul>
						</dd>
					</div>
				{/if}

				{#if usecase.relations && usecase.relations.length > 0}
					<div class="detail-item full">
						<dt>ユースケース関係</dt>
						<dd>
							<ul class="relation-list">
								{#each usecase.relations as relation}
									<li class="relation-item">
										<span class="relation-type">{relationLabels[relation.type] ?? relation.type}</span>
										<span class="target-id">{relation.target_id}</span>
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
		</Panel>
	{:else}
		<!-- 未選択状態 -->
		<Panel title="詳細">
			<div class="empty-state">
				<Icon name="Info" size={24} />
				<p>アクターまたはユースケースを選択してください</p>
			</div>
		</Panel>
	{/if}
</div>

<style>
	.usecase-panel {
		height: 100%;
		overflow-y: auto;
	}

	.panel-header {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		margin-bottom: 1rem;
		padding-bottom: 0.75rem;
		border-bottom: 1px solid var(--border-primary);
	}

	.header-info {
		flex: 1;
	}

	.header-info h3 {
		margin: 0 0 0.25rem 0;
		font-size: 1rem;
		font-weight: 600;
		color: var(--text-primary);
	}

	.header-info .id {
		font-size: 0.75rem;
		color: var(--text-muted);
		font-family: monospace;
	}

	.status-indicator {
		width: 12px;
		height: 12px;
		border-radius: 50%;
		flex-shrink: 0;
		margin-top: 4px;
	}

	.close-btn {
		background: transparent;
		border: none;
		color: var(--text-muted);
		cursor: pointer;
		padding: 0.25rem;
		border-radius: 4px;
		transition: background 0.15s ease;
	}

	.close-btn:hover {
		background: var(--bg-tertiary);
		color: var(--text-primary);
	}

	.detail-list {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.75rem;
	}

	.detail-item {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.detail-item.full {
		grid-column: 1 / -1;
	}

	.detail-item dt {
		font-size: 0.7rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-muted);
		font-weight: 500;
	}

	.detail-item dd {
		margin: 0;
		font-size: 0.875rem;
		color: var(--text-primary);
	}

	.badge {
		display: inline-block;
		padding: 0.125rem 0.5rem;
		border-radius: 2px;
		font-size: 0.75rem;
		font-weight: 500;
	}

	.badge-type {
		background: var(--bg-tertiary);
		color: var(--text-primary);
	}

	.monospace {
		font-family: monospace;
		font-size: 0.8rem;
		background: var(--bg-primary);
		padding: 0.25rem 0.5rem;
		border-radius: 2px;
	}

	.description {
		line-height: 1.5;
		color: var(--text-secondary);
	}

	.relation-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.relation-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem;
		background: var(--bg-primary);
		border-radius: 4px;
		font-size: 0.85rem;
	}

	.role-badge {
		margin-left: auto;
		font-size: 0.7rem;
		color: var(--text-muted);
		background: var(--bg-tertiary);
		padding: 0.125rem 0.375rem;
		border-radius: 2px;
	}

	.relation-type {
		font-family: monospace;
		font-size: 0.75rem;
		color: var(--accent-primary);
	}

	.target-id {
		font-family: monospace;
		font-size: 0.8rem;
	}

	.condition {
		font-size: 0.75rem;
		color: var(--text-muted);
		font-style: italic;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.75rem;
		padding: 2rem;
		color: var(--text-muted);
		text-align: center;
	}

	.empty-state p {
		margin: 0;
		font-size: 0.875rem;
	}
</style>
