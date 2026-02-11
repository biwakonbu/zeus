<script lang="ts">
	// UseCase View - PixiJS 版
	// ミニマルデザイン: キャンバスが主役、パネルはオーバーレイで必要時のみ表示
	import { onMount, onDestroy } from 'svelte';
	import { get } from 'svelte/store';
	import type {
		UseCaseDiagramResponse,
		ActorItem,
		UseCaseItem,
		SubsystemItem
	} from '$lib/types/api';
	import { fetchUseCaseDiagram, fetchSubsystems, fetchActivities } from '$lib/api/client';
	import { Icon, EmptyState, OverlayPanel } from '$lib/components/ui';
	import { UseCaseEngine, type UseCaseEngineData } from './engine/UseCaseEngine';
	import UseCaseListPanel from './UseCaseListPanel.svelte';
	import UseCaseViewPanel from './UseCaseViewPanel.svelte';
	import {
		updateUseCaseViewState,
		resetUseCaseViewState,
		pendingNavigation,
		clearPendingNavigation
	} from '$lib/stores/view';

	import type { ActivityItem } from '$lib/types/api';

	type Props = {
		boundary?: string;
		activities?: ActivityItem[];
		onActorSelect?: (actor: ActorItem) => void;
		onUseCaseSelect?: (usecase: UseCaseItem) => void;
	};
	let { boundary = '', activities = [], onActorSelect, onUseCaseSelect }: Props = $props();

	// データ状態
	let data: UseCaseDiagramResponse | null = $state(null);
	let loading = $state(true);
	let error: string | null = $state(null);

	// パネル表示状態
	// リストパネルはデフォルト表示（フィルタモードでアクター/ユースケースを選択するため）
	let showListPanel = $state(true);
	let showDetailPanel = $state(false);

	// 選択状態
	let selectedActorId: string | null = $state(null);
	let selectedUseCaseId: string | null = $state(null);

	// 選択されたエンティティ
	const selectedActor = $derived.by((): ActorItem | null => {
		if (!selectedActorId || !data) return null;
		return data.actors.find((a: ActorItem) => a.id === selectedActorId) ?? null;
	});

	const selectedUseCase = $derived.by((): UseCaseItem | null => {
		if (!selectedUseCaseId || !data) return null;
		return data.usecases.find((u: UseCaseItem) => u.id === selectedUseCaseId) ?? null;
	});

	// 何か選択されているか
	const hasSelection = $derived(selectedActor !== null || selectedUseCase !== null);

	// ホバー状態（Tooltip用）
	let hoveredActor: ActorItem | null = $state(null);
	let hoveredUseCase: UseCaseItem | null = $state(null);
	let hoverPosition = $state({ x: 0, y: 0 });

	// Tooltip 位置
	const TOOLTIP_WIDTH = 280;
	const TOOLTIP_HEIGHT = 150;
	const TOOLTIP_OFFSET = 16;

	const tooltipStyle = $derived(() => {
		const viewportWidth = typeof window !== 'undefined' ? window.innerWidth : 1920;
		const viewportHeight = typeof window !== 'undefined' ? window.innerHeight : 1080;
		const flipX = hoverPosition.x + TOOLTIP_WIDTH + TOOLTIP_OFFSET > viewportWidth;
		const flipY = hoverPosition.y + TOOLTIP_HEIGHT + TOOLTIP_OFFSET > viewportHeight;
		const left = flipX
			? hoverPosition.x - TOOLTIP_WIDTH - TOOLTIP_OFFSET
			: hoverPosition.x + TOOLTIP_OFFSET;
		const top = flipY
			? hoverPosition.y - TOOLTIP_HEIGHT - TOOLTIP_OFFSET
			: hoverPosition.y + TOOLTIP_OFFSET;
		return `left: ${Math.max(8, left)}px; top: ${Math.max(8, top)}px;`;
	});

	// PixiJS エンジン
	let engine: UseCaseEngine | null = null;
	let canvasContainer: HTMLDivElement | null = $state(null);
	let currentZoom = $state(1.0);

	// サブシステムデータ
	let subsystems: SubsystemItem[] = $state([]);

	// 直接取得した Activity データ（props からのフォールバック対応）
	let fetchedActivities: ActivityItem[] = $state([]);

	// データ取得（ユースケース図、サブシステム、Activity を並列取得）
	async function loadData() {
		loading = true;
		error = null;
		try {
			const [diagramData, subsystemsResponse, activitiesResponse] = await Promise.all([
				fetchUseCaseDiagram(boundary || undefined),
				fetchSubsystems(),
				fetchActivities()
			]);
			data = diagramData;
			subsystems = subsystemsResponse.subsystems || [];
			fetchedActivities = activitiesResponse.activities || [];
		} catch (e) {
			error = e instanceof Error ? e.message : 'データの取得に失敗しました';
		} finally {
			loading = false;
		}
	}

	// エンジン初期化状態
	let engineInitializing = false;
	let engineReady = $state(false);

	/**
	 * 選択を全解除（ステート変更のみ）
	 * エンジンへの反映は選択同期 $effect が担当する（責務の一元化）
	 * NOTE: 親コールバック (onActorSelect/onUseCaseSelect) は選択時のみ呼ばれる。
	 *       解除時の通知は現在の Props 型 (ActorItem => void) では対応していない。
	 */
	function clearAllSelection(): void {
		selectedActorId = null;
		selectedUseCaseId = null;
		showDetailPanel = false;
	}

	// エンジン初期化（一度だけ実行）
	async function initEngine(): Promise<void> {
		if (!canvasContainer || engineInitializing || engineReady) return;
		engineInitializing = true;

		try {
			engine = new UseCaseEngine();
			await engine.init(canvasContainer);

			// フィルタモードは有効（選択時に関連ノードだけ表示）
			// 初期表示は setData() 後に showAll() で行う
			engine.setFilterMode(true);

			engine.onActorClicked((actor) => {
				if (selectedActorId === actor.id) {
					// 再クリック → 選択解除（トグル）
					clearAllSelection();
					return;
				}
				selectedActorId = actor.id;
				selectedUseCaseId = null;
				showDetailPanel = true;
				onActorSelect?.(actor);
			});

			engine.onActorHovered((actor, event) => {
				hoveredActor = actor;
				hoveredUseCase = null;
				if (event) hoverPosition = { x: event.clientX, y: event.clientY };
			});

			engine.onUseCaseClicked((usecase) => {
				if (selectedUseCaseId === usecase.id) {
					// 再クリック → 選択解除（トグル）
					clearAllSelection();
					return;
				}
				selectedUseCaseId = usecase.id;
				selectedActorId = null;
				showDetailPanel = true;
				onUseCaseSelect?.(usecase);
			});

			engine.onUseCaseHovered((usecase, event) => {
				hoveredUseCase = usecase;
				hoveredActor = null;
				if (event) hoverPosition = { x: event.clientX, y: event.clientY };
			});

			engine.onViewportChanged((viewport) => {
				currentZoom = viewport.scale;
				// ヘッダーの store を更新
				updateUseCaseViewState({ zoom: viewport.scale });
			});

			// エンジン準備完了（$effect をトリガー）
			engineReady = true;
		} catch (e) {
			error = e instanceof Error ? e.message : 'エンジン初期化に失敗しました';
		} finally {
			engineInitializing = false;
		}
	}

	// ズーム操作（ヘッダーから呼び出される）
	function handleZoomIn() {
		engine?.setZoom(currentZoom * 1.2);
	}
	function handleZoomOut() {
		engine?.setZoom(currentZoom / 1.2);
	}
	function handleZoomReset() {
		engine?.centerView();
	}

	// パネル操作
	function toggleListPanel() {
		showListPanel = !showListPanel;
		updateUseCaseViewState({ showListPanel });
	}

	function closeListPanel() {
		showListPanel = false;
		updateUseCaseViewState({ showListPanel: false });
	}

	function closeDetailPanel() {
		// ステート変更のみ。エンジンへの反映は $effect が担当
		clearAllSelection();
	}

	// Actor/UseCase クリック（リストから）
	// トグル動作: 同じ要素を再クリックで選択解除 + 全表示
	function handleActorClick(actor: ActorItem) {
		if (selectedActorId === actor.id) {
			// 再クリック → 選択解除
			clearAllSelection();
			return;
		}
		selectedActorId = actor.id;
		selectedUseCaseId = null;
		showDetailPanel = true;
		engine?.selectActor(actor.id);
		onActorSelect?.(actor);
	}

	function handleUseCaseClick(usecase: UseCaseItem) {
		if (selectedUseCaseId === usecase.id) {
			// 再クリック → 選択解除
			clearAllSelection();
			return;
		}
		selectedUseCaseId = usecase.id;
		selectedActorId = null;
		showDetailPanel = true;
		engine?.selectUseCase(usecase.id);
		onUseCaseSelect?.(usecase);
	}

	// Actor アイコン取得
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

	// ステータス色取得
	function getStatusColor(status: string): string {
		const colors: Record<string, string> = {
			active: 'var(--status-good)',
			draft: 'var(--status-fair)',
			deprecated: 'var(--text-muted)'
		};
		return colors[status] ?? 'var(--text-secondary)';
	}

	// ESC キーで段階的に解除
	// 1. 選択/フィルタがアクティブ → 解除
	// 2. 何もなければリストパネルを閉じる
	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			if (selectedActorId || selectedUseCaseId) {
				// フィルタ/選択がアクティブ → 解除
				clearAllSelection();
			} else if (showDetailPanel) {
				showDetailPanel = false;
			} else if (showListPanel) {
				closeListPanel();
			}
		}
	}

	// Store へのコールバック登録（一度だけ実行）
	let callbacksRegistered = false;
	function registerStoreCallbacks(): void {
		if (callbacksRegistered) return;
		updateUseCaseViewState({
			onZoomIn: handleZoomIn,
			onZoomOut: handleZoomOut,
			onZoomReset: handleZoomReset,
			onToggleListPanel: toggleListPanel
		});
		callbacksRegistered = true;
	}

	// Store へのデータ同期
	function syncStoreData(): void {
		updateUseCaseViewState({
			zoom: currentZoom,
			boundary: data?.boundary || 'System',
			actorCount: data?.actors.length || 0,
			usecaseCount: data?.usecases.length || 0,
			showListPanel
		});
	}

	onMount(() => {
		loadData();
		document.addEventListener('keydown', handleKeydown);
		return () => document.removeEventListener('keydown', handleKeydown);
	});

	// エンジン初期化 Effect（canvasContainer が利用可能になったら一度だけ）
	$effect(() => {
		const container = canvasContainer;
		const ready = engineReady;
		if (container && !ready && !engineInitializing) {
			initEngine();
		}
	});

	// コールバック登録 Effect（engineReady 時に一度だけ）
	$effect(() => {
		if (engineReady) {
			registerStoreCallbacks();
		}
	});

	// データ設定 Effect（エンジン初期化後、data または subsystems が変更されたら）
	// 選択状態の反映は選択同期 Effect に任せる
	// 注意: この Effect は選択同期 Effect より先に定義する必要がある
	//       Svelte 5 では Effect は定義順に実行されるため、
	//       setData() → 選択反映 の順序が保証される
	// NOTE: syncStoreData() は含めない → currentZoom / showListPanel の依存を遮断し、
	//       ズーム操作での $effect 再実行（フルリビルド）を防止する
	$effect(() => {
		const ready = engineReady;
		const currentData = data;
		const currentSubsystems = subsystems;
		if (!ready || !engine || !currentData) return;

		const engineData: UseCaseEngineData = { ...currentData, subsystems: currentSubsystems };
		engine.setData(engineData);

		// pendingNavigation は依存に含めない（遷移中の "一瞬の全表示" を避ける）
		// ナビゲーション中は showAll() をスキップ（ナビゲーション Effect に任せる）
		const nav = get(pendingNavigation);
		if (!nav || nav.view !== 'usecase') {
			engine.showAll();
		}

		// データ設定後に store のデータ部分を同期（コールバック以外）
		updateUseCaseViewState({
			boundary: currentData.boundary || 'System',
			actorCount: currentData.actors.length || 0,
			usecaseCount: currentData.usecases.length || 0
		});
	});

	// 選択状態の同期 Effect（選択状態が変更されたらエンジンに反映）
	// エンジン操作の唯一の責務ポイント（ハンドラは select 時の即時反映のみ例外的に直接呼び出す）
	// 注意: この Effect はデータ設定 Effect の後に定義すること
	//       エンジン側で冪等性が保証されているため重複呼び出しは問題ない
	$effect(() => {
		const ready = engineReady;
		const usecaseId = selectedUseCaseId;
		const actorId = selectedActorId;
		if (!ready || !engine) return;

		if (usecaseId) {
			engine.selectUseCase(usecaseId);
		} else if (actorId) {
			engine.selectActor(actorId);
		} else {
			// 両方 null の場合、選択解除 + フィルタ解除（全表示に戻す）
			engine.clearSelectionVisual();
			engine.showAll();
		}
	});

	// showListPanel が変わったら store を更新
	$effect(() => {
		updateUseCaseViewState({ showListPanel });
	});

	$effect(() => {
		const ready = engineReady;
		const container = canvasContainer;
		if (!container || !ready || !engine) return () => {};
		const resizeObserver = new ResizeObserver(() => engine?.resize());
		resizeObserver.observe(container);
		return () => resizeObserver.disconnect();
	});

	// ナビゲーションによる自動選択 Effect
	// エンジンに直接選択を反映（絞り込みを確実に実行）
	$effect(() => {
		// ローカル変数に代入することで TypeScript の型推論を活用し、
		// 以降のコードで null チェック後の型が確定する
		const nav = $pendingNavigation;
		const ready = engineReady;
		const currentData = data;
		if (!nav || nav.view !== 'usecase' || !ready || !engine || !currentData) return;

		if (nav.entityType === 'usecase' && nav.entityId) {
			// UseCase を選択
			const usecase = currentData.usecases.find((u: UseCaseItem) => u.id === nav.entityId);
			if (usecase) {
				selectedUseCaseId = usecase.id;
				selectedActorId = null;
				showDetailPanel = true;
				engine.selectUseCase(usecase.id);
				onUseCaseSelect?.(usecase);
			} else {
				clearAllSelection();
				engine.clearSelectionVisual();
				engine.showAll();
			}
			clearPendingNavigation();
		} else if (nav.entityType === 'actor' && nav.entityId) {
			// Actor を選択
			const actor = currentData.actors.find((a: ActorItem) => a.id === nav.entityId);
			if (actor) {
				selectedActorId = actor.id;
				selectedUseCaseId = null;
				showDetailPanel = true;
				engine.selectActor(actor.id);
				onActorSelect?.(actor);
			} else {
				clearAllSelection();
				engine.clearSelectionVisual();
				engine.showAll();
			}
			clearPendingNavigation();
		} else {
			// 想定外のナビゲーションは詰まらせない
			clearPendingNavigation();
		}
	});

	onDestroy(() => {
		engine?.destroy();
		engine = null;
		// store をリセット
		resetUseCaseViewState();
	});
</script>

<div class="usecase-view">
	{#if loading}
		<div class="loading-state">
			<Icon name="RefreshCw" size={32} />
			<span>読み込み中...</span>
		</div>
	{:else if error}
		<div class="error-state">
			<Icon name="AlertTriangle" size={32} />
			<span>{error}</span>
			<button class="retry-button" onclick={loadData}>再試行</button>
		</div>
	{:else if !data || (data.actors.length === 0 && data.usecases.length === 0)}
		<EmptyState
			title="ユースケース図がありません"
			description="zeus add actor / zeus add usecase でアクターとユースケースを追加してください"
			icon="ClipboardList"
		/>
	{:else}
		<!-- フルスクリーンキャンバス -->
		<div class="canvas-area">
			<div class="canvas-wrapper" bind:this={canvasContainer}></div>

			<!-- リストパネル（オーバーレイ） -->
			{#if showListPanel}
				<OverlayPanel
					title="要素一覧"
					position="top-left"
					panelId="usecase-list"
					defaultWidthPreset="medium"
					onClose={closeListPanel}
				>
					<UseCaseListPanel
						actors={data.actors}
						usecases={data.usecases}
						{selectedActorId}
						{selectedUseCaseId}
						onActorSelect={handleActorClick}
						onUseCaseSelect={handleUseCaseClick}
					/>
				</OverlayPanel>
			{/if}

			<!-- 詳細パネル（オーバーレイ） -->
			{#if showDetailPanel && hasSelection}
				<OverlayPanel
					title="プロパティ"
					position="top-right"
					panelId="usecase-detail"
					defaultWidthPreset="medium"
					onClose={closeDetailPanel}
				>
					<div class="detail-content">
						<UseCaseViewPanel
							actor={selectedActor}
							usecase={selectedUseCase}
							actors={data.actors}
							usecases={data.usecases}
							activities={fetchedActivities.length > 0 ? fetchedActivities : activities}
							onClose={closeDetailPanel}
						/>
					</div>
				</OverlayPanel>
			{/if}
		</div>

		<!-- ホバー Tooltip -->
		{#if hoveredActor}
			<div class="hover-tooltip" style={tooltipStyle()}>
				<div class="tooltip-header">
					<Icon name={getActorIcon(hoveredActor.type)} size={14} />
					<span>{hoveredActor.title}</span>
				</div>
				<div class="tooltip-meta">
					<span class="tooltip-type">{hoveredActor.type}</span>
					<span class="tooltip-id">{hoveredActor.id}</span>
				</div>
				{#if hoveredActor.description}
					<div class="tooltip-desc">{hoveredActor.description}</div>
				{/if}
			</div>
		{/if}

		{#if hoveredUseCase}
			<div class="hover-tooltip" style={tooltipStyle()}>
				<div class="tooltip-header">
					<span
						class="tooltip-status-dot"
						style="background: {getStatusColor(hoveredUseCase.status)}"
					></span>
					<span>{hoveredUseCase.title}</span>
				</div>
				<div class="tooltip-meta">
					<span class="tooltip-id">{hoveredUseCase.id}</span>
					<span class="tooltip-status">{hoveredUseCase.status}</span>
				</div>
				{#if hoveredUseCase.description}
					<div class="tooltip-desc">{hoveredUseCase.description}</div>
				{/if}
				{#if hoveredUseCase.actors.length > 0}
					<div class="tooltip-actors">{hoveredUseCase.actors.length} actor(s)</div>
				{/if}
			</div>
		{/if}
	{/if}
</div>

<style>
	.usecase-view {
		width: 100%;
		height: 100%;
		position: relative;
		overflow: hidden;
		min-height: 400px;
		background: var(--bg-primary);
		color: var(--text-primary);
	}

	.loading-state,
	.error-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 1rem;
		height: 100%;
		color: var(--text-muted);
	}

	.error-state {
		color: var(--status-poor);
	}

	.retry-button {
		padding: 0.5rem 1rem;
		background: var(--bg-secondary);
		border: 1px solid var(--border-metal);
		color: var(--text-primary);
		cursor: pointer;
		border-radius: 4px;
	}

	.retry-button:hover {
		background: var(--bg-hover);
	}

	/* フルスクリーンキャンバス */
	.canvas-area {
		position: relative;
		width: 100%;
		height: 100%;
		overflow: hidden;
	}

	.canvas-wrapper {
		width: 100%;
		height: 100%;
		background-color: #1a1a1a;
		background-image: radial-gradient(
			circle at 1px 1px,
			rgba(255, 149, 51, 0.08) 1px,
			transparent 0
		);
		background-size: 24px 24px;
	}

	.canvas-wrapper :global(canvas) {
		display: block;
	}

	/* 詳細パネルのコンテンツ（パディング調整） */
	.detail-content {
		padding: 12px;
	}

	/* ホバー Tooltip */
	.hover-tooltip {
		position: fixed;
		z-index: 1000;
		background: rgba(30, 30, 30, 0.95);
		border: 1px solid var(--border-metal);
		border-radius: 6px;
		padding: 10px;
		min-width: 160px;
		max-width: 260px;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
		pointer-events: none;
		backdrop-filter: blur(8px);
	}

	.tooltip-header {
		display: flex;
		align-items: center;
		gap: 6px;
		font-weight: 600;
		font-size: 0.8125rem;
		margin-bottom: 6px;
	}

	.tooltip-status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
	}

	.tooltip-meta {
		display: flex;
		gap: 6px;
		font-size: 0.6875rem;
		color: var(--text-muted);
		margin-bottom: 6px;
	}

	.tooltip-type,
	.tooltip-status {
		background: rgba(0, 0, 0, 0.3);
		padding: 2px 6px;
		border-radius: 3px;
	}

	.tooltip-id {
		font-family: monospace;
		font-size: 0.65rem;
	}

	.tooltip-desc {
		font-size: 0.75rem;
		color: var(--text-secondary);
		line-height: 1.4;
	}

	.tooltip-actors {
		font-size: 0.6875rem;
		color: var(--accent-primary);
		margin-top: 6px;
	}
</style>
