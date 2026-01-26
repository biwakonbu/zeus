<script lang="ts">
	// UseCase View - PixiJS 版
	// ミニマルデザイン: キャンバスが主役、パネルはオーバーレイで必要時のみ表示
	import { onMount, onDestroy } from 'svelte';
	import type { UseCaseDiagramResponse, ActorItem, UseCaseItem } from '$lib/types/api';
	import { fetchUseCaseDiagram } from '$lib/api/client';
	import { Icon, EmptyState, OverlayPanel } from '$lib/components/ui';
	import { UseCaseEngine } from './engine/UseCaseEngine';
	import UseCaseListPanel from './UseCaseListPanel.svelte';
	import UseCaseViewPanel from './UseCaseViewPanel.svelte';
	import { updateUseCaseViewState, resetUseCaseViewState } from '$lib/stores/view';

	type Props = {
		boundary?: string;
		onActorSelect?: (actor: ActorItem) => void;
		onUseCaseSelect?: (usecase: UseCaseItem) => void;
	};
	let { boundary = '', onActorSelect, onUseCaseSelect }: Props = $props();

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

	// データ取得
	async function loadData() {
		loading = true;
		error = null;
		try {
			data = await fetchUseCaseDiagram(boundary || undefined);
		} catch (e) {
			error = e instanceof Error ? e.message : 'データの取得に失敗しました';
		} finally {
			loading = false;
		}
	}

	// エンジン初期化状態
	let engineInitializing = false;
	let engineInitialized = false;

	// エンジン初期化（一度だけ実行）
	async function initEngine(): Promise<void> {
		if (!canvasContainer || engineInitializing || engineInitialized) return;
		engineInitializing = true;

		try {
			engine = new UseCaseEngine();
			await engine.init(canvasContainer);

			// デフォルトでフィルタモードを有効化（選択するまで非表示）
			engine.setFilterMode(true);

			engine.onActorClicked((actor) => {
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

			engineInitialized = true;

			// 初期化完了後にデータがあれば設定
			if (data) {
				engine.setData(data);
				syncStoreState();
			}
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
		showDetailPanel = false;
		selectedActorId = null;
		selectedUseCaseId = null;
		// 視覚的な選択状態のみ解除（図は消さない）
		engine?.clearSelectionVisual();
	}

	// Actor/UseCase クリック（リストから）
	function handleActorClick(actor: ActorItem) {
		selectedActorId = actor.id;
		selectedUseCaseId = null;
		showDetailPanel = true;
		engine?.selectActor(actor.id);
		onActorSelect?.(actor);
	}

	function handleUseCaseClick(usecase: UseCaseItem) {
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

	// ESC キーでパネルを閉じる
	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			if (showDetailPanel) closeDetailPanel();
			else if (showListPanel) closeListPanel();
		}
	}

	// ヘッダーの store を更新
	function syncStoreState() {
		updateUseCaseViewState({
			zoom: currentZoom,
			boundary: data?.boundary || 'System',
			actorCount: data?.actors.length || 0,
			usecaseCount: data?.usecases.length || 0,
			showListPanel,
			onZoomIn: handleZoomIn,
			onZoomOut: handleZoomOut,
			onZoomReset: handleZoomReset,
			onToggleListPanel: toggleListPanel
		});
	}

	onMount(() => {
		loadData();
		document.addEventListener('keydown', handleKeydown);
		return () => document.removeEventListener('keydown', handleKeydown);
	});

	// エンジン初期化 Effect（canvasContainer が利用可能になったら一度だけ）
	$effect(() => {
		if (canvasContainer && !engineInitialized && !engineInitializing) {
			initEngine();
		}
	});

	// データ設定 Effect（エンジン初期化後、data が変更されたら）
	$effect(() => {
		if (engineInitialized && engine && data) {
			engine.setData(data);
			syncStoreState();
		}
	});

	// showListPanel が変わったら store を更新
	$effect(() => {
		updateUseCaseViewState({ showListPanel });
	});

	$effect(() => {
		if (!canvasContainer || !engine) return () => {};
		const resizeObserver = new ResizeObserver(() => engine?.resize());
		resizeObserver.observe(canvasContainer);
		return () => resizeObserver.disconnect();
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
					width="320px"
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
					width="300px"
					onClose={closeDetailPanel}
				>
					<div class="detail-content">
						<UseCaseViewPanel
							actor={selectedActor}
							usecase={selectedUseCase}
							actors={data.actors}
							usecases={data.usecases}
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
					<span class="tooltip-status-dot" style="background: {getStatusColor(hoveredUseCase.status)}"></span>
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
		background-image:
			radial-gradient(circle at 1px 1px, rgba(255, 149, 51, 0.08) 1px, transparent 0);
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
