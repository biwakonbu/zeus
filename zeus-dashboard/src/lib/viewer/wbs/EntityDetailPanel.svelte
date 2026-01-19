<script lang="ts">
	import { fetchWBS, fetchDownstream } from '$lib/api/client';
	import type {
		WBSNode,
		WBSResponse,
		DownstreamResponse,
		WBSNodeType
	} from '$lib/types/api';

	// Props
	interface Props {
		entityId: string | null;
		onClose?: () => void;
		onEntitySelect?: (entityId: string) => void;
	}
	let { entityId, onClose, onEntitySelect }: Props = $props();

	// 状態
	let wbsData: WBSResponse | null = $state(null);
	let entity: WBSNode | null = $state(null);
	let children: WBSNode[] = $state([]);
	let upstream: string[] = $state([]);
	let downstream: string[] = $state([]);
	let loading = $state(true);
	let error: string | null = $state(null);

	// ノード種別の表示名
	const nodeTypeLabels: Record<WBSNodeType, string> = {
		vision: 'Vision',
		objective: 'Objective',
		deliverable: 'Deliverable',
		task: 'Task'
	};

	// ステータスの色
	const statusColors: Record<string, { bg: string; text: string }> = {
		completed: { bg: '#153b1f', text: '#22c55e' },
		in_progress: { bg: '#3b3515', text: '#eab308' },
		pending: { bg: '#1f2d3b', text: '#3b82f6' },
		blocked: { bg: '#3b1515', text: '#ef4444' },
		not_started: { bg: '#2a2a2a', text: '#888' },
		on_hold: { bg: '#2d1f4e', text: '#8b5cf6' },
		draft: { bg: '#2a2a2a', text: '#888' },
		in_review: { bg: '#3b3515', text: '#eab308' },
		approved: { bg: '#153b1f', text: '#22c55e' },
		delivered: { bg: '#153b1f', text: '#22c55e' }
	};

	// WBS ツリーからエンティティを検索
	function findEntity(nodes: WBSNode[], id: string): WBSNode | null {
		for (const node of nodes) {
			if (node.id === id) return node;
			if (node.children) {
				const found = findEntity(node.children, id);
				if (found) return found;
			}
		}
		return null;
	}

	// 親エンティティを検索
	function findParent(nodes: WBSNode[], childId: string, parent: WBSNode | null = null): WBSNode | null {
		for (const node of nodes) {
			if (node.id === childId) return parent;
			if (node.children) {
				const found = findParent(node.children, childId, node);
				if (found !== null) return found;
			}
		}
		return null;
	}

	// データ取得
	async function loadData() {
		if (!entityId) {
			entity = null;
			loading = false;
			return;
		}

		loading = true;
		error = null;

		try {
			// WBS データ取得
			wbsData = await fetchWBS();

			// エンティティ検索
			entity = findEntity(wbsData.roots, entityId);

			if (entity) {
				// 子エンティティ
				children = entity.children || [];

				// 上流・下流取得
				try {
					const deps = await fetchDownstream(entityId);
					upstream = deps.upstream || [];
					downstream = deps.downstream || [];
				} catch {
					// 依存関係取得エラーは無視
					upstream = [];
					downstream = [];
				}
			} else {
				error = 'エンティティが見つかりません';
			}
		} catch (e) {
			error = e instanceof Error ? e.message : '取得に失敗しました';
		} finally {
			loading = false;
		}
	}

	// entityId 変更時にデータ再取得
	$effect(() => {
		if (entityId) {
			loadData();
		} else {
			entity = null;
			children = [];
			upstream = [];
			downstream = [];
			loading = false;
		}
	});

	// エンティティクリック
	function handleEntityClick(id: string) {
		onEntitySelect?.(id);
	}

	// ステータス色取得
	function getStatusColor(status: string) {
		return statusColors[status] || { bg: '#2a2a2a', text: '#888' };
	}

	// プログレスバー幅
	function getProgressWidth(progress: number): string {
		return `${Math.min(100, Math.max(0, progress))}%`;
	}
</script>

<div class="entity-detail-panel">
	{#if !entityId}
		<div class="empty-state">
			<span class="empty-icon">📋</span>
			<span class="empty-text">エンティティを選択してください</span>
		</div>
	{:else if loading}
		<div class="loading-state">
			<div class="spinner"></div>
			<span>読み込み中...</span>
		</div>
	{:else if error}
		<div class="error-state">
			<span class="error-icon">⚠</span>
			<span>{error}</span>
		</div>
	{:else if entity}
		<!-- ヘッダー -->
		<div class="panel-header">
			<div class="header-main">
				<span class="entity-type-badge">{nodeTypeLabels[entity.node_type]}</span>
				<h2 class="entity-title">{entity.title}</h2>
				{#if onClose}
					<button class="close-button" onclick={onClose}>✕</button>
				{/if}
			</div>
			<div class="header-meta">
				<span class="entity-id">{entity.id}</span>
				{#if entity.wbs_code}
					<span class="wbs-code">WBS: {entity.wbs_code}</span>
				{/if}
			</div>
			<div class="header-status">
				<span
					class="status-badge"
					style="
						background: {getStatusColor(entity.status).bg};
						color: {getStatusColor(entity.status).text};
						border: 1px solid {getStatusColor(entity.status).text};
					"
				>
					{entity.status}
				</span>
				<div class="progress-container">
					<div class="progress-bar">
						<div
							class="progress-fill"
							style="width: {getProgressWidth(entity.progress)};"
						></div>
					</div>
					<span class="progress-text">{entity.progress}%</span>
				</div>
			</div>
		</div>

		<!-- コンテンツ -->
		<div class="panel-content">
			<!-- 基本情報 -->
			{#if entity.assignee || entity.priority}
				<section class="info-section">
					<h3 class="section-title">基本情報</h3>
					<div class="info-grid">
						{#if entity.assignee}
							<div class="info-item">
								<span class="info-label">担当者</span>
								<span class="info-value">{entity.assignee}</span>
							</div>
						{/if}
						{#if entity.priority}
							<div class="info-item">
								<span class="info-label">優先度</span>
								<span class="info-value priority-{entity.priority}">{entity.priority}</span>
							</div>
						{/if}
					</div>
				</section>
			{/if}

			<!-- 子エンティティ -->
			{#if children.length > 0}
				<section class="info-section">
					<h3 class="section-title">
						{#if entity.node_type === 'objective'}
							含まれる成果物 (Deliverables)
						{:else if entity.node_type === 'deliverable'}
							含まれるタスク (Tasks)
						{:else}
							子エンティティ
						{/if}
					</h3>
					<div class="children-list">
						{#each children as child}
							<button
								class="child-item"
								onclick={() => handleEntityClick(child.id)}
							>
								<div class="child-header">
									<span class="child-type">{nodeTypeLabels[child.node_type]}</span>
									<span class="child-title">{child.title}</span>
								</div>
								<div class="child-progress">
									<div class="progress-bar small">
										<div
											class="progress-fill"
											style="width: {getProgressWidth(child.progress)};"
										></div>
									</div>
									<span class="progress-text">{child.progress}%</span>
								</div>
							</button>
						{/each}
					</div>
				</section>
			{/if}

			<!-- 依存関係 -->
			{#if upstream.length > 0 || downstream.length > 0}
				<section class="info-section">
					<h3 class="section-title">依存関係</h3>
					<div class="dependencies">
						{#if upstream.length > 0}
							<div class="dep-group">
								<span class="dep-label">上流 (このエンティティが依存)</span>
								<div class="dep-list">
									{#each upstream as id}
										<button
											class="dep-item upstream"
											onclick={() => handleEntityClick(id)}
										>
											{id}
										</button>
									{/each}
								</div>
							</div>
						{/if}
						{#if downstream.length > 0}
							<div class="dep-group">
								<span class="dep-label">下流 (このエンティティに依存)</span>
								<div class="dep-list">
									{#each downstream as id}
										<button
											class="dep-item downstream"
											onclick={() => handleEntityClick(id)}
										>
											{id}
										</button>
									{/each}
								</div>
							</div>
						{/if}
					</div>
				</section>
			{/if}
		</div>
	{/if}
</div>

<style>
	.entity-detail-panel {
		height: 100%;
		display: flex;
		flex-direction: column;
		background: #1a1a1a;
		color: #e0e0e0;
		overflow: hidden;
	}

	/* 空・ローディング・エラー状態 */
	.empty-state,
	.loading-state,
	.error-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 12px;
		height: 100%;
		color: #888;
	}

	.empty-icon,
	.error-icon {
		font-size: 48px;
		opacity: 0.5;
	}

	.spinner {
		width: 24px;
		height: 24px;
		border: 2px solid #333;
		border-top-color: #f5a623;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.error-state {
		color: #ef4444;
	}

	/* ヘッダー */
	.panel-header {
		padding: 16px 20px;
		border-bottom: 1px solid #333;
		background: #222;
	}

	.header-main {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 8px;
	}

	.entity-type-badge {
		padding: 4px 10px;
		background: #f5a623;
		color: #1a1a1a;
		font-size: 11px;
		font-weight: 700;
		text-transform: uppercase;
		border-radius: 4px;
	}

	.entity-title {
		flex: 1;
		margin: 0;
		font-size: 18px;
		font-weight: 600;
		color: #fff;
	}

	.close-button {
		padding: 4px 8px;
		background: transparent;
		border: none;
		color: #888;
		font-size: 16px;
		cursor: pointer;
	}

	.close-button:hover {
		color: #fff;
	}

	.header-meta {
		display: flex;
		gap: 16px;
		margin-bottom: 12px;
	}

	.entity-id,
	.wbs-code {
		font-size: 12px;
		font-family: 'Fira Code', monospace;
		color: #888;
	}

	.header-status {
		display: flex;
		align-items: center;
		gap: 16px;
	}

	.status-badge {
		padding: 4px 12px;
		font-size: 12px;
		font-weight: 500;
		border-radius: 4px;
	}

	.progress-container {
		flex: 1;
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.progress-bar {
		flex: 1;
		height: 8px;
		background: #333;
		border-radius: 4px;
		overflow: hidden;
	}

	.progress-bar.small {
		height: 4px;
	}

	.progress-fill {
		height: 100%;
		background: linear-gradient(90deg, #f5a623, #f59e0b);
		border-radius: 4px;
		transition: width 0.3s ease;
	}

	.progress-text {
		font-size: 12px;
		color: #888;
		min-width: 36px;
		text-align: right;
	}

	/* コンテンツ */
	.panel-content {
		flex: 1;
		overflow-y: auto;
		padding: 16px 20px;
	}

	.info-section {
		margin-bottom: 24px;
	}

	.section-title {
		margin: 0 0 12px 0;
		padding-bottom: 8px;
		border-bottom: 1px solid #333;
		font-size: 14px;
		font-weight: 600;
		color: #f5a623;
	}

	/* 基本情報グリッド */
	.info-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 12px;
	}

	.info-item {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.info-label {
		font-size: 11px;
		color: #888;
		text-transform: uppercase;
	}

	.info-value {
		font-size: 14px;
		color: #e0e0e0;
	}

	.info-value.priority-high {
		color: #ef4444;
	}

	.info-value.priority-medium {
		color: #eab308;
	}

	.info-value.priority-low {
		color: #22c55e;
	}

	/* 子エンティティリスト */
	.children-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.child-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 10px 14px;
		background: #222;
		border: 1px solid #333;
		border-radius: 6px;
		cursor: pointer;
		transition: all 0.2s;
		text-align: left;
	}

	.child-item:hover {
		background: #2a2a2a;
		border-color: #444;
	}

	.child-header {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.child-type {
		font-size: 10px;
		color: #888;
		text-transform: uppercase;
	}

	.child-title {
		font-size: 13px;
		color: #e0e0e0;
	}

	.child-progress {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 120px;
	}

	/* 依存関係 */
	.dependencies {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.dep-group {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.dep-label {
		font-size: 12px;
		color: #888;
	}

	.dep-list {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}

	.dep-item {
		padding: 4px 10px;
		border: none;
		border-radius: 4px;
		font-size: 12px;
		font-family: 'Fira Code', monospace;
		cursor: pointer;
		transition: opacity 0.2s;
	}

	.dep-item:hover {
		opacity: 0.8;
	}

	.dep-item.upstream {
		background: #1e2d4d;
		color: #3b82f6;
	}

	.dep-item.downstream {
		background: #3b3515;
		color: #eab308;
	}
</style>
