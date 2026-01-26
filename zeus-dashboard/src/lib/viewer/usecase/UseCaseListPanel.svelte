<script lang="ts">
	// UseCaseListPanel - 統合リストパネル（オーバーレイ用シンプル版）
	// Actor/UseCase をタブ、検索、フィルタで管理
	import { onDestroy } from 'svelte';
	import type { ActorItem, UseCaseItem } from '$lib/types/api';
	import SegmentedTabs from './components/SegmentedTabs.svelte';
	import SearchInput from './components/SearchInput.svelte';
	import FilterDropdown from './components/FilterDropdown.svelte';
	import GroupedList from './components/GroupedList.svelte';

	// リストアイテム型
	type ActorListItem = ActorItem & { itemType: 'actor' };
	type UseCaseListItem = UseCaseItem & { itemType: 'usecase' };
	type ListItem = ActorListItem | UseCaseListItem;

	interface Props {
		actors: ActorItem[];
		usecases: UseCaseItem[];
		selectedActorId: string | null;
		selectedUseCaseId: string | null;
		onActorSelect: (actor: ActorItem) => void;
		onUseCaseSelect: (usecase: UseCaseItem) => void;
	}
	let {
		actors,
		usecases,
		selectedActorId,
		selectedUseCaseId,
		onActorSelect,
		onUseCaseSelect
	}: Props = $props();

	// 状態
	let activeTab = $state<'all' | 'actor' | 'usecase'>('all');
	let searchQuery = $state('');
	let debouncedQuery = $state('');
	let filterActorId = $state<string | null>(null);

	// debounce 用タイマー
	let debounceTimer: ReturnType<typeof setTimeout>;

	// 検索入力ハンドラー
	function handleSearch(value: string) {
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

	// タブ切り替え
	function handleTabChange(tabId: string) {
		activeTab = tabId as typeof activeTab;
		// 検索・フィルタをリセット（選択は維持）
		searchQuery = '';
		debouncedQuery = '';
		filterActorId = null;
	}

	// タブデータ
	const tabs = $derived([
		{ id: 'all', label: '全て', count: actors.length + usecases.length },
		{ id: 'actor', label: 'Actor', count: actors.length },
		{ id: 'usecase', label: 'UseCase', count: usecases.length }
	]);

	// フィルタ用 Actor オプション
	const filterOptions = $derived(actors.map((a) => ({ id: a.id, label: a.title })));

	// フィルタ表示条件（UseCase/全て タブのみ）
	const showFilter = $derived(activeTab === 'all' || activeTab === 'usecase');

	// フィルタ済みアイテム（$derived.by で関数呼び出し不要）
	const filteredItems = $derived.by(() => {
		const query = debouncedQuery.toLowerCase();

		// ベースデータ
		let actorList: ActorListItem[] = actors.map((a) => ({ ...a, itemType: 'actor' as const }));
		let usecaseList: UseCaseListItem[] = usecases.map((u) => ({
			...u,
			itemType: 'usecase' as const
		}));

		// Actor フィルタ（UseCase のみに適用）
		if (filterActorId) {
			usecaseList = usecaseList.filter((uc) =>
				uc.actors.some((a) => a.actor_id === filterActorId)
			);
		}

		// 検索フィルタ
		if (query) {
			actorList = actorList.filter(
				(a) =>
					a.title.toLowerCase().includes(query) ||
					(a.description?.toLowerCase().includes(query) ?? false)
			);
			usecaseList = usecaseList.filter(
				(u) =>
					u.title.toLowerCase().includes(query) ||
					(u.description?.toLowerCase().includes(query) ?? false)
			);
		}

		// タブフィルタ
		if (activeTab === 'actor') return actorList;
		if (activeTab === 'usecase') return usecaseList;
		return [...actorList, ...usecaseList] as ListItem[];
	});

	// グループ化フラグ（「全て」タブのみ）
	const shouldGroupBy = $derived(activeTab === 'all');

	// 選択 ID
	const selectedId = $derived(selectedActorId ?? selectedUseCaseId);

	// アイテム選択ハンドラー
	function handleSelect(item: ListItem) {
		if (item.itemType === 'actor') {
			onActorSelect(item);
		} else {
			onUseCaseSelect(item);
		}
	}

	// クリーンアップ
	onDestroy(() => {
		clearTimeout(debounceTimer);
	});
</script>

<div class="list-panel-content">
	<!-- タブ -->
	<div class="control-row">
		<SegmentedTabs {tabs} {activeTab} onTabChange={handleTabChange} />
	</div>

	<!-- 検索 -->
	<div class="control-row">
		<SearchInput
			value={searchQuery}
			placeholder="検索..."
			onInput={handleSearch}
			onClear={handleSearchClear}
		/>
	</div>

	<!-- フィルタ（条件付き表示） -->
	{#if showFilter && filterOptions.length > 0}
		<div class="control-row">
			<FilterDropdown
				options={filterOptions}
				selected={filterActorId}
				placeholder="Actor: 全て"
				onSelect={(id) => (filterActorId = id)}
			/>
		</div>
	{/if}

	<!-- リスト -->
	<div class="list-area">
		<GroupedList
			items={filteredItems}
			groupBy={shouldGroupBy}
			{selectedId}
			{actors}
			onSelect={handleSelect}
		/>
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

	.control-row {
		flex-shrink: 0;
	}

	.list-area {
		flex: 1;
		overflow-y: auto;
		margin: 0 -8px -8px;
	}
</style>
