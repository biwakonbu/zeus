<script lang="ts">
	// UseCase View Panel（オーバーレイ用シンプル版）
	// 選択された Actor または UseCase の詳細を表示
	import type { ActorItem, UseCaseItem, ActivityItem } from '$lib/types/api';
	import { Icon } from '$lib/components/ui';
	import { navigateToEntity } from '$lib/stores/view';

	interface Props {
		actor?: ActorItem | null;
		usecase?: UseCaseItem | null;
		actors?: ActorItem[];
		usecases?: UseCaseItem[];
		activities?: ActivityItem[];
		onClose?: () => void;
	}
	let { actor = null, usecase = null, actors = [], usecases = [], activities = [], onClose }: Props = $props();

	// 関連 Activity を取得
	const relatedActivities = $derived.by((): ActivityItem[] => {
		if (!usecase) return [];
		return activities.filter((a) => a.usecase_id === usecase.id);
	});

	// Activity へ遷移
	function handleActivityClick(activityId: string) {
		navigateToEntity('activity', 'activity', activityId);
	}

	// 折りたたみ状態
	let alternativeFlowExpanded = $state<Record<string, boolean>>({});
	let exceptionFlowExpanded = $state<Record<string, boolean>>({});

	function toggleAlternativeFlow(id: string) {
		alternativeFlowExpanded[id] = !alternativeFlowExpanded[id];
	}

	function toggleExceptionFlow(id: string) {
		exceptionFlowExpanded[id] = !exceptionFlowExpanded[id];
	}

	// シナリオが存在するか確認
	function hasScenario(uc: UseCaseItem | null): boolean {
		if (!uc?.scenario) return false;
		const s = uc.scenario;
		return !!(
			(s.preconditions && s.preconditions.length > 0) ||
			s.trigger ||
			(s.main_flow && s.main_flow.length > 0) ||
			(s.alternative_flows && s.alternative_flows.length > 0) ||
			(s.exception_flows && s.exception_flows.length > 0) ||
			(s.postconditions && s.postconditions.length > 0)
		);
	}

	// 名前解決ヘルパー関数
	function getActorName(actorId: string): string {
		const found = actors.find((a) => a.id === actorId);
		return found ? found.title : actorId;
	}

	function getUseCaseName(usecaseId: string): string {
		const found = usecases.find((u) => u.id === usecaseId);
		return found ? found.title : usecaseId;
	}

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
	{#if onClose}
		<button class="close-button" onclick={onClose} title="閉じる">
			<Icon name="X" size={16} />
		</button>
	{/if}
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
									<span class="relation-name">{getActorName(actorRef.actor_id)}</span>
									<span class="relation-id">({actorRef.actor_id})</span>
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
									<span class="relation-name">{getUseCaseName(relation.target_id)}</span>
									<span class="relation-id">({relation.target_id})</span>
									{#if relation.condition}
										<span class="condition">[{relation.condition}]</span>
									{/if}
								</li>
							{/each}
						</ul>
					</dd>
				</div>
			{/if}

			<!-- シナリオセクション -->
			{#if hasScenario(usecase)}
				<div class="detail-item full scenario-section">
					<dt>
						<Icon name="FileText" size={12} />
						シナリオ
					</dt>
					<dd class="scenario-content">
						<!-- 事前条件 -->
						{#if usecase.scenario?.preconditions && usecase.scenario.preconditions.length > 0}
							<div class="scenario-group">
								<h4 class="scenario-heading">事前条件</h4>
								<ul class="scenario-list">
									{#each usecase.scenario.preconditions as condition}
										<li>{condition}</li>
									{/each}
								</ul>
							</div>
						{/if}

						<!-- トリガー -->
						{#if usecase.scenario?.trigger}
							<div class="scenario-group">
								<h4 class="scenario-heading">トリガー</h4>
								<p class="scenario-trigger">{usecase.scenario.trigger}</p>
							</div>
						{/if}

						<!-- メインフロー -->
						{#if usecase.scenario?.main_flow && usecase.scenario.main_flow.length > 0}
							<div class="scenario-group">
								<h4 class="scenario-heading">メインフロー</h4>
								<ol class="scenario-flow-list">
									{#each usecase.scenario.main_flow as step}
										<li>{step}</li>
									{/each}
								</ol>
							</div>
						{/if}

						<!-- 代替フロー -->
						{#if usecase.scenario?.alternative_flows && usecase.scenario.alternative_flows.length > 0}
							<div class="scenario-group">
								<h4 class="scenario-heading">代替フロー</h4>
								{#each usecase.scenario.alternative_flows as altFlow}
									<div class="flow-card">
										<button
											class="flow-header"
											onclick={() => toggleAlternativeFlow(altFlow.id)}
											aria-expanded={alternativeFlowExpanded[altFlow.id] ?? false}
										>
											<Icon
												name={alternativeFlowExpanded[altFlow.id] ? 'ChevronDown' : 'ChevronRight'}
												size={12}
											/>
											<span class="flow-id">{altFlow.id}</span>
											<span class="flow-name">{altFlow.name}</span>
										</button>
										{#if alternativeFlowExpanded[altFlow.id]}
											<div class="flow-body">
												<div class="flow-condition">
													<strong>条件:</strong> {altFlow.condition}
												</div>
												<ol class="flow-steps">
													{#each altFlow.steps as step}
														<li>{step}</li>
													{/each}
												</ol>
												{#if altFlow.rejoins_at}
													<div class="flow-rejoins">
														<Icon name="CornerDownRight" size={12} />
														{altFlow.rejoins_at}
													</div>
												{/if}
											</div>
										{/if}
									</div>
								{/each}
							</div>
						{/if}

						<!-- 例外フロー -->
						{#if usecase.scenario?.exception_flows && usecase.scenario.exception_flows.length > 0}
							<div class="scenario-group">
								<h4 class="scenario-heading">例外フロー</h4>
								{#each usecase.scenario.exception_flows as excFlow}
									<div class="flow-card exception">
										<button
											class="flow-header"
											onclick={() => toggleExceptionFlow(excFlow.id)}
											aria-expanded={exceptionFlowExpanded[excFlow.id] ?? false}
										>
											<Icon
												name={exceptionFlowExpanded[excFlow.id] ? 'ChevronDown' : 'ChevronRight'}
												size={12}
											/>
											<span class="flow-id">{excFlow.id}</span>
											<span class="flow-name">{excFlow.name}</span>
										</button>
										{#if exceptionFlowExpanded[excFlow.id]}
											<div class="flow-body">
												<div class="flow-trigger">
													<strong>発生条件:</strong> {excFlow.trigger}
												</div>
												<ol class="flow-steps">
													{#each excFlow.steps as step}
														<li>{step}</li>
													{/each}
												</ol>
												{#if excFlow.outcome}
													<div class="flow-outcome">
														<Icon name="ArrowRight" size={12} />
														{excFlow.outcome}
													</div>
												{/if}
											</div>
										{/if}
									</div>
								{/each}
							</div>
						{/if}

						<!-- 事後条件 -->
						{#if usecase.scenario?.postconditions && usecase.scenario.postconditions.length > 0}
							<div class="scenario-group">
								<h4 class="scenario-heading">事後条件</h4>
								<ul class="scenario-list">
									{#each usecase.scenario.postconditions as condition}
										<li>{condition}</li>
									{/each}
								</ul>
							</div>
						{/if}
					</dd>
				</div>
			{/if}

			<!-- 関連アクティビティセクション -->
			{#if relatedActivities.length > 0}
				<div class="detail-item full activity-section">
					<dt>
						<Icon name="Workflow" size={12} />
						関連アクティビティ ({relatedActivities.length})
					</dt>
					<dd>
						<ul class="activity-list">
							{#each relatedActivities as activity}
								<li class="activity-item">
									<button
										class="activity-link"
										onclick={() => handleActivityClick(activity.id)}
										title="Activity ビューで表示"
									>
										<Icon name="ExternalLink" size={10} />
										<span class="activity-title">{activity.title}</span>
										<span class="activity-id">{activity.id}</span>
									</button>
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
		transition: background 0.15s ease, color 0.15s ease;
	}

	.close-button:hover {
		background: rgba(255, 100, 100, 0.2);
		color: var(--text-primary);
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
		font-size: 0.75rem;
		color: var(--text-primary);
	}

	.relation-id {
		font-family: monospace;
		font-size: 0.65rem;
		color: var(--text-muted);
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

	/* シナリオセクション */
	.scenario-section {
		margin-top: 12px;
		padding-top: 12px;
		border-top: 1px solid var(--border-metal);
	}

	.scenario-section dt {
		display: flex;
		align-items: center;
		gap: 4px;
		color: var(--accent-primary);
	}

	.scenario-content {
		display: flex;
		flex-direction: column;
		gap: 12px;
		margin-top: 8px;
	}

	.scenario-group {
		background: rgba(0, 0, 0, 0.2);
		border: 1px solid var(--border-metal);
		border-radius: 4px;
		padding: 10px;
	}

	.scenario-heading {
		margin: 0 0 8px 0;
		font-size: 0.6875rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-muted);
		font-weight: 600;
	}

	.scenario-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.scenario-list li {
		font-size: 0.75rem;
		color: var(--text-secondary);
		padding-left: 12px;
		position: relative;
	}

	.scenario-list li::before {
		content: '•';
		position: absolute;
		left: 0;
		color: var(--accent-primary);
	}

	.scenario-trigger {
		margin: 0;
		font-size: 0.75rem;
		color: var(--text-primary);
		background: rgba(255, 149, 51, 0.1);
		padding: 6px 10px;
		border-radius: 3px;
		border-left: 2px solid var(--accent-primary);
	}

	.scenario-flow-list {
		margin: 0;
		padding: 0 0 0 20px;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.scenario-flow-list li {
		font-size: 0.75rem;
		color: var(--text-secondary);
	}

	/* 代替・例外フローカード */
	.flow-card {
		background: rgba(0, 0, 0, 0.15);
		border: 1px solid var(--border-metal);
		border-radius: 4px;
		overflow: hidden;
		margin-bottom: 6px;
	}

	.flow-card:last-child {
		margin-bottom: 0;
	}

	.flow-card.exception {
		border-color: rgba(255, 100, 100, 0.3);
	}

	.flow-header {
		display: flex;
		align-items: center;
		gap: 6px;
		width: 100%;
		padding: 8px 10px;
		background: transparent;
		border: none;
		cursor: pointer;
		color: var(--text-primary);
		text-align: left;
		font-family: inherit;
		transition: background 0.15s ease;
	}

	.flow-header:hover {
		background: rgba(255, 255, 255, 0.05);
	}

	.flow-id {
		font-family: monospace;
		font-size: 0.625rem;
		color: var(--accent-primary);
		background: rgba(255, 149, 51, 0.15);
		padding: 1px 5px;
		border-radius: 2px;
	}

	.flow-name {
		font-size: 0.75rem;
		font-weight: 500;
	}

	.flow-body {
		padding: 8px 10px 10px 28px;
		border-top: 1px solid var(--border-metal);
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.flow-condition,
	.flow-trigger {
		font-size: 0.7rem;
		color: var(--text-secondary);
	}

	.flow-condition strong,
	.flow-trigger strong {
		color: var(--text-muted);
		font-weight: 500;
	}

	.flow-steps {
		margin: 0;
		padding: 0 0 0 16px;
		display: flex;
		flex-direction: column;
		gap: 3px;
	}

	.flow-steps li {
		font-size: 0.7rem;
		color: var(--text-secondary);
	}

	.flow-rejoins,
	.flow-outcome {
		display: flex;
		align-items: center;
		gap: 4px;
		font-size: 0.65rem;
		color: var(--text-muted);
		font-style: italic;
	}

	/* 関連アクティビティセクション */
	.activity-section {
		margin-top: 12px;
		padding-top: 12px;
		border-top: 1px solid var(--border-metal);
	}

	.activity-section dt {
		display: flex;
		align-items: center;
		gap: 4px;
		color: var(--accent-primary);
	}

	.activity-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
		margin-top: 8px;
	}

	.activity-item {
		display: block;
	}

	.activity-link {
		display: flex;
		align-items: center;
		gap: 6px;
		width: 100%;
		padding: 6px 8px;
		background: rgba(255, 149, 51, 0.1);
		border: 1px solid rgba(255, 149, 51, 0.3);
		border-radius: 4px;
		color: var(--text-primary);
		cursor: pointer;
		font-family: inherit;
		font-size: 0.75rem;
		text-align: left;
		transition: background 0.15s ease, border-color 0.15s ease;
	}

	.activity-link:hover {
		background: rgba(255, 149, 51, 0.2);
		border-color: var(--accent-primary);
	}

	.activity-title {
		flex: 1;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.activity-id {
		font-family: monospace;
		font-size: 0.65rem;
		color: var(--text-muted);
		flex-shrink: 0;
	}
</style>
