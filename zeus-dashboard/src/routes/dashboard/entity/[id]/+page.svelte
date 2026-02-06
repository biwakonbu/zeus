<script lang="ts">
	// Drill-Down View
	// エンティティの詳細を表示するページ
	// 議論結果（round: 20260121-174500_wbsdesign）に基づく

	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { restoreDrillDownState } from '$lib/stores/drillDown';
	import { Icon, ProgressBar } from '$lib/components/ui';

	interface PageData {
		entityId: string;
		returnUrl: string;
	}

	let { data }: { data: PageData } = $props();

	// エンティティデータ（実際の API 実装後に置き換え）
	interface EntityDetail {
		id: string;
		title: string;
		type: 'vision' | 'objective' | 'deliverable' | 'activity' | 'usecase';
		status: string;
		progress: number;
		description?: string;
		relatedEntities: Array<{ id: string; title: string; type: string }>;
		history: Array<{ date: string; action: string }>;
	}

	let entity: EntityDetail | null = $state(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		try {
			// TODO: API 実装後に fetchEntityDetails(data.entityId) に置き換え
			// 仮データを使用
			await new Promise((resolve) => setTimeout(resolve, 200));
			entity = {
				id: data.entityId,
				title: `Entity ${data.entityId}`,
				type: detectEntityType(data.entityId),
				status: 'in_progress',
				progress: 65,
				description:
					'This is a placeholder description for the entity. The actual content will be loaded from the API.',
				relatedEntities: [
					{ id: 'DEL-001', title: 'Related Deliverable 1', type: 'deliverable' },
					{ id: 'ACT-001', title: 'Related Activity 1', type: 'activity' }
				],
				history: [
					{ date: '2026-01-21 10:30', action: 'Progress updated to 65%' },
					{ date: '2026-01-20 15:00', action: 'Status changed to In Progress' }
				]
			};
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load entity';
		} finally {
			loading = false;
		}
	});

	// ID プレフィックスからエンティティタイプを推測
	function detectEntityType(id: string): 'vision' | 'objective' | 'deliverable' | 'activity' | 'usecase' {
		if (id.startsWith('VIS') || id.startsWith('vis')) return 'vision';
		if (id.startsWith('OBJ') || id.startsWith('obj')) return 'objective';
		if (id.startsWith('DEL') || id.startsWith('del')) return 'deliverable';
		if (id.startsWith('UC') || id.startsWith('uc')) return 'usecase';
		return 'activity';
	}

	// 戻るボタンハンドラ
	function handleBack() {
		const savedState = restoreDrillDownState();
		if (savedState) {
			goto(savedState.returnUrl);
			// スクロール位置復元
			if (savedState.viewState.scrollX !== undefined) {
				setTimeout(() => {
					window.scrollTo(savedState.viewState.scrollX!, savedState.viewState.scrollY!);
				}, 0);
			}
		} else {
			goto(data.returnUrl);
		}
	}

	// Escape キーで戻る
	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			handleBack();
		}
	}

	// ステータスに応じた色
	function getStatusColor(status: string): string {
		switch (status) {
			case 'completed':
				return 'var(--status-good)';
			case 'in_progress':
				return 'var(--status-info)';
			case 'blocked':
				return 'var(--status-poor)';
			default:
				return 'var(--text-muted)';
		}
	}

	// タイプに応じたアイコン
	function getTypeIcon(type: string): string {
		switch (type) {
			case 'vision':
				return 'Target';
			case 'objective':
				return 'Flag';
			case 'deliverable':
				return 'Package';
			case 'activity':
				return 'CheckSquare';
			case 'usecase':
				return 'Users';
			default:
				return 'Circle';
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="drill-down-view">
	<!-- ヘッダー -->
	<header class="drill-down-header">
		<button class="back-btn" onclick={handleBack}>
			<Icon name="ArrowLeft" size={20} />
			<span>Back to Dashboard</span>
		</button>
		<div class="header-shortcuts">
			<span class="shortcut-hint">
				<kbd>Esc</kbd> to go back
			</span>
		</div>
	</header>

	<!-- メインコンテンツ -->
	<main class="drill-down-content">
		{#if loading}
			<div class="loading-state">
				<div class="spinner"></div>
				<span>Loading entity...</span>
			</div>
		{:else if error}
			<div class="error-state">
				<Icon name="AlertTriangle" size={32} />
				<span>{error}</span>
				<button class="retry-btn" onclick={() => location.reload()}>Retry</button>
			</div>
		{:else if entity}
			<div class="entity-detail">
				<!-- エンティティヘッダー -->
				<div class="entity-header">
					<div class="entity-type-badge">
						<Icon name={getTypeIcon(entity.type)} size={18} />
						<span>{entity.type.toUpperCase()}</span>
					</div>
					<h1 class="entity-id">{entity.id}</h1>
					<h2 class="entity-title">{entity.title}</h2>
				</div>

				<!-- 統計カード -->
				<div class="entity-stats">
					<div class="stat-card">
						<span class="stat-label">Progress</span>
						<div class="stat-progress">
							<ProgressBar value={entity.progress} />
						</div>
						<span class="stat-value">{entity.progress}%</span>
					</div>
					<div class="stat-card">
						<span class="stat-label">Status</span>
						<span class="stat-value status" style="color: {getStatusColor(entity.status)}">
							{entity.status.replace(/_/g, ' ')}
						</span>
					</div>
				</div>

				<!-- 説明 -->
				{#if entity.description}
					<section class="entity-section">
						<h3 class="section-title">Description</h3>
						<p class="section-content">{entity.description}</p>
					</section>
				{/if}

				<!-- 関連エンティティ -->
				{#if entity.relatedEntities.length > 0}
					<section class="entity-section">
						<h3 class="section-title">Related Entities</h3>
						<ul class="related-list">
							{#each entity.relatedEntities as related}
								<li class="related-item">
									<Icon name={getTypeIcon(related.type)} size={14} />
									<span class="related-id">{related.id}</span>
									<span class="related-title">{related.title}</span>
								</li>
							{/each}
						</ul>
					</section>
				{/if}

				<!-- 履歴 -->
				{#if entity.history.length > 0}
					<section class="entity-section">
						<h3 class="section-title">History</h3>
						<ul class="history-list">
							{#each entity.history as item}
								<li class="history-item">
									<span class="history-date">{item.date}</span>
									<span class="history-action">{item.action}</span>
								</li>
							{/each}
						</ul>
					</section>
				{/if}
			</div>
		{/if}
	</main>
</div>

<style>
	.drill-down-view {
		min-height: 100vh;
		background: var(--bg-primary, #1a1a1a);
		color: var(--text-primary, #fff);
		font-family: var(--font-family);
	}

	/* ヘッダー */
	.drill-down-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 24px;
		background: var(--bg-secondary, #242424);
		border-bottom: 2px solid var(--border-metal, #4a4a4a);
	}

	.back-btn {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 16px;
		background: transparent;
		border: 1px solid var(--border-metal, #4a4a4a);
		border-radius: var(--border-radius-md, 4px);
		color: var(--text-secondary, #b8b8b8);
		cursor: pointer;
		font-family: inherit;
		font-size: 14px;
		transition: all 0.2s;
	}

	.back-btn:hover {
		background: var(--bg-hover, #3a3a3a);
		border-color: var(--accent-primary, #ff9533);
		color: var(--accent-primary, #ff9533);
	}

	.header-shortcuts {
		display: flex;
		align-items: center;
	}

	.shortcut-hint {
		font-size: 12px;
		color: var(--text-muted, #888);
	}

	.shortcut-hint kbd {
		display: inline-block;
		padding: 2px 6px;
		background: var(--bg-panel, #2d2d2d);
		border: 1px solid var(--border-metal, #4a4a4a);
		border-radius: 3px;
		font-size: 11px;
		margin-right: 4px;
	}

	/* メインコンテンツ */
	.drill-down-content {
		padding: 32px;
		max-width: 960px;
		margin: 0 auto;
	}

	/* ローディング・エラー状態 */
	.loading-state,
	.error-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 64px;
		gap: 16px;
		color: var(--text-muted, #888);
	}

	.spinner {
		width: 32px;
		height: 32px;
		border: 3px solid var(--bg-panel, #333);
		border-top-color: var(--accent-primary, #ff9533);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.error-state {
		color: var(--status-poor, #ee4444);
	}

	.retry-btn {
		padding: 8px 16px;
		background: transparent;
		border: 1px solid var(--accent-primary, #ff9533);
		border-radius: var(--border-radius-md, 4px);
		color: var(--accent-primary, #ff9533);
		cursor: pointer;
		font-family: inherit;
	}

	.retry-btn:hover {
		background: var(--accent-primary, #ff9533);
		color: var(--bg-primary, #1a1a1a);
	}

	/* エンティティ詳細 */
	.entity-header {
		margin-bottom: 32px;
	}

	.entity-type-badge {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 4px 12px;
		background: var(--bg-panel, #2d2d2d);
		border-radius: var(--border-radius-md, 4px);
		color: var(--accent-primary, #ff9533);
		font-size: 12px;
		font-weight: 600;
		letter-spacing: 0.1em;
		margin-bottom: 12px;
	}

	.entity-id {
		font-size: 14px;
		color: var(--text-muted, #888);
		margin: 8px 0;
		font-weight: 400;
	}

	.entity-title {
		font-size: 28px;
		font-weight: 600;
		color: var(--text-primary, #fff);
		margin: 0;
	}

	/* 統計カード */
	.entity-stats {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 16px;
		margin-bottom: 32px;
	}

	.stat-card {
		padding: 16px;
		background: var(--bg-panel, #2d2d2d);
		border: 1px solid var(--border-metal, #4a4a4a);
		border-radius: var(--border-radius-md, 4px);
	}

	.stat-label {
		font-size: 12px;
		color: var(--text-muted, #888);
		display: block;
		margin-bottom: 8px;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.stat-progress {
		margin-bottom: 8px;
	}

	.stat-value {
		font-size: 16px;
		font-weight: 600;
		text-transform: capitalize;
	}

	/* セクション */
	.entity-section {
		margin-bottom: 24px;
	}

	.section-title {
		font-size: 14px;
		font-weight: 600;
		color: var(--text-secondary, #b8b8b8);
		margin-bottom: 12px;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding-bottom: 8px;
		border-bottom: 1px solid var(--border-dark, #333);
	}

	.section-content {
		font-size: 14px;
		line-height: 1.6;
		color: var(--text-secondary, #b8b8b8);
	}

	/* 関連エンティティリスト */
	.related-list {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.related-item {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 12px;
		background: var(--bg-secondary, #242424);
		border-radius: var(--border-radius-md, 4px);
		margin-bottom: 8px;
		cursor: pointer;
		transition: background 0.2s;
	}

	.related-item:hover {
		background: var(--bg-hover, #3a3a3a);
	}

	.related-id {
		font-size: 12px;
		color: var(--text-muted, #888);
		font-family: var(--font-family);
	}

	.related-title {
		font-size: 13px;
		color: var(--text-secondary, #b8b8b8);
	}

	/* 履歴リスト */
	.history-list {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.history-item {
		display: flex;
		gap: 16px;
		padding: 8px 0;
		border-bottom: 1px solid var(--border-dark, #333);
		font-size: 13px;
	}

	.history-item:last-child {
		border-bottom: none;
	}

	.history-date {
		color: var(--text-muted, #888);
		min-width: 140px;
	}

	.history-action {
		color: var(--text-secondary, #b8b8b8);
	}

	/* モバイル対応 */
	@media (max-width: 768px) {
		.drill-down-content {
			padding: 16px;
		}

		.entity-stats {
			grid-template-columns: 1fr;
		}

		.entity-title {
			font-size: 22px;
		}
	}
</style>
