<script lang="ts">
	// ActivityListPanel - アクティビティ一覧パネル（オーバーレイ用シンプル版）
	import { onDestroy } from 'svelte';
	import type { ActivityItem } from '$lib/types/api';
	import { Icon, SearchInput } from '$lib/components/ui';

	interface Props {
		activities: ActivityItem[];
		selectedActivityId: string | null;
		onActivitySelect: (activity: ActivityItem) => void;
	}
	let { activities, selectedActivityId, onActivitySelect }: Props = $props();

	// 状態
	let searchQuery = $state('');
	let debouncedQuery = $state('');

	// debounce 用タイマー
	let debounceTimer: ReturnType<typeof setTimeout>;

	// 検索入力ハンドラー
	function handleSearchInput(value: string) {
		searchQuery = value;
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => {
			debouncedQuery = value;
		}, 250);
	}

	// 検索クリア
	function handleSearchClear() {
		searchQuery = '';
		debouncedQuery = '';
	}

	// フィルタ済みアイテム
	const filteredActivities = $derived.by(() => {
		const query = debouncedQuery.toLowerCase();
		if (!query) return activities;
		return activities.filter(
			(a) =>
				a.title.toLowerCase().includes(query) ||
				(a.description?.toLowerCase().includes(query) ?? false) ||
				a.id.toLowerCase().includes(query)
		);
	});

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

	// クリーンアップ
	onDestroy(() => {
		clearTimeout(debounceTimer);
	});
</script>

<div class="list-panel-content">
	<!-- 検索 -->
	<div class="search-row">
		<SearchInput
			value={searchQuery}
			placeholder="検索..."
			onInput={handleSearchInput}
			onClear={handleSearchClear}
		/>
	</div>

	<!-- カウント表示 -->
	<div class="count-row">
		<span class="count-label">{filteredActivities.length} / {activities.length}</span>
	</div>

	<!-- リスト -->
	<div class="list-area">
		{#if filteredActivities.length === 0}
			<div class="empty-list">
				<Icon name="Inbox" size={24} />
				<span>該当するアクティビティがありません</span>
			</div>
		{:else}
			{#each filteredActivities as activity (activity.id)}
				{@const isSelected = selectedActivityId === activity.id}
				<button
					class="activity-item"
					class:selected={isSelected}
					onclick={() => onActivitySelect(activity)}
				>
					<div class="activity-header">
						<span class="status-dot" style="background: {getStatusColor(activity.status)}"></span>
						<span class="activity-title">{activity.title}</span>
					</div>
					<div class="activity-meta">
						<span class="activity-id">{activity.id}</span>
						<span class="activity-status">{getStatusLabel(activity.status)}</span>
					</div>
					{#if activity.description}
						<div class="activity-desc">{activity.description}</div>
					{/if}
					<div class="activity-stats">
						<span class="stat">
							<Icon name="Circle" size={10} />
							{activity.nodes.length} ノード
						</span>
						<span class="stat">
							<Icon name="ArrowRight" size={10} />
							{activity.transitions.length} 遷移
						</span>
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
		flex-shrink: 0;
		padding: 4px 0;
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
		padding: 32px;
		color: var(--text-muted);
		text-align: center;
	}

	.empty-list span {
		font-size: 0.75rem;
	}

	.activity-item {
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

	.activity-item:hover {
		background: rgba(0, 0, 0, 0.3);
		border-color: var(--border-hover);
	}

	.activity-item.selected {
		background: rgba(255, 149, 51, 0.15);
		border-color: var(--accent-primary);
	}

	.activity-header {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.activity-title {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--text-primary);
		line-height: 1.3;
	}

	.activity-meta {
		display: flex;
		gap: 8px;
		font-size: 0.6875rem;
		color: var(--text-muted);
	}

	.activity-id {
		font-family: monospace;
		font-size: 0.625rem;
	}

	.activity-status {
		background: rgba(0, 0, 0, 0.3);
		padding: 1px 5px;
		border-radius: 2px;
	}

	.activity-desc {
		font-size: 0.75rem;
		color: var(--text-secondary);
		line-height: 1.4;
		overflow: hidden;
		text-overflow: ellipsis;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
	}

	.activity-stats {
		display: flex;
		gap: 12px;
		font-size: 0.6875rem;
		color: var(--text-muted);
	}

	.stat {
		display: flex;
		align-items: center;
		gap: 4px;
	}
</style>
