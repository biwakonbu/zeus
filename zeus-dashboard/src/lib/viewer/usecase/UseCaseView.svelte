<script lang="ts">
	// UseCase View - PixiJS 版
	// UML ユースケース図を PixiJS で表示するビュー
	import { onMount, onDestroy } from 'svelte';
	import type {
		UseCaseDiagramResponse,
		ActorItem,
		UseCaseItem
	} from '$lib/types/api';
	import { fetchUseCaseDiagram } from '$lib/api/client';
	import { Icon, EmptyState, Panel } from '$lib/components/ui';
	import { UseCaseEngine } from './engine/UseCaseEngine';

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

	// 選択状態
	let selectedActorId: string | null = $state(null);
	let selectedUseCaseId: string | null = $state(null);

	// ホバー状態（Tooltip用）
	let hoveredActor: ActorItem | null = $state(null);
	let hoveredUseCase: UseCaseItem | null = $state(null);
	let hoverPosition = $state({ x: 0, y: 0 });

	// Tooltip 位置（ビューポートオーバーフロー防止）
	const TOOLTIP_WIDTH = 280;
	const TOOLTIP_HEIGHT = 150;
	const TOOLTIP_OFFSET = 16;

	const tooltipStyle = $derived(() => {
		const viewportWidth = typeof window !== 'undefined' ? window.innerWidth : 1920;
		const viewportHeight = typeof window !== 'undefined' ? window.innerHeight : 1080;

		// 右端・下端に近い場合は反対側に表示
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

	// ズーム状態
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

	// エンジン初期化
	async function initEngine() {
		if (!canvasContainer || engine) return;

		engine = new UseCaseEngine();
		await engine.init(canvasContainer);

		// イベントリスナー設定
		engine.onActorClicked((actor) => {
			selectedActorId = actor.id;
			selectedUseCaseId = null;
			onActorSelect?.(actor);
		});

		engine.onActorHovered((actor, event) => {
			hoveredActor = actor;
			hoveredUseCase = null;
			if (event) {
				hoverPosition = { x: event.clientX, y: event.clientY };
			}
		});

		engine.onUseCaseClicked((usecase) => {
			selectedUseCaseId = usecase.id;
			selectedActorId = null;
			onUseCaseSelect?.(usecase);
		});

		engine.onUseCaseHovered((usecase, event) => {
			hoveredUseCase = usecase;
			hoveredActor = null;
			if (event) {
				hoverPosition = { x: event.clientX, y: event.clientY };
			}
		});

		engine.onViewportChanged((viewport) => {
			currentZoom = viewport.scale;
		});

		// データがあれば描画
		if (data) {
			engine.setData(data);
		}
	}

	// Actor クリック処理（サイドバーから）
	function handleActorClick(actor: ActorItem) {
		selectedActorId = actor.id;
		selectedUseCaseId = null;
		engine?.selectActor(actor.id);
		onActorSelect?.(actor);
	}

	// UseCase クリック処理（サイドバーから）
	function handleUseCaseClick(usecase: UseCaseItem) {
		selectedUseCaseId = usecase.id;
		selectedActorId = null;
		engine?.selectUseCase(usecase.id);
		onUseCaseSelect?.(usecase);
	}

	// ズーム操作
	function handleZoomIn() {
		engine?.setZoom(currentZoom * 1.2);
	}

	function handleZoomOut() {
		engine?.setZoom(currentZoom / 1.2);
	}

	function handleZoomReset() {
		engine?.centerView();
	}

	// Actor タイプのアイコン名を取得
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

	// UseCase ステータスの色を取得
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

	// マウント時
	onMount(() => {
		loadData();
	});

	// データ変更時にエンジンを初期化/更新
	$effect(() => {
		if (data && canvasContainer) {
			if (!engine) {
				// async 関数をエラーハンドリング付きで呼び出し
				initEngine().catch((e) => {
					error = e instanceof Error ? e.message : 'エンジン初期化に失敗しました';
				});
			} else {
				engine.setData(data);
			}
		}
	});

	// リサイズ対応
	$effect(() => {
		// 早期 return でクリーンアップ関数が常に返されるようにする
		if (!canvasContainer || !engine) {
			return () => {};
		}

		const resizeObserver = new ResizeObserver(() => {
			engine?.resize();
		});
		resizeObserver.observe(canvasContainer);

		return () => resizeObserver.disconnect();
	});

	// クリーンアップ
	onDestroy(() => {
		engine?.destroy();
		engine = null;
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
		<div class="usecase-layout">
			<!-- 左サイドバー: Actor リスト -->
			<aside class="actor-sidebar">
				<Panel title="アクター ({data.actors.length})">
					<ul class="actor-list">
						{#each data.actors as actor}
							<li>
								<button
									class="actor-item"
									class:selected={selectedActorId === actor.id}
									onclick={() => handleActorClick(actor)}
								>
									<Icon name={getActorIcon(actor.type)} size={16} />
									<span class="actor-title">{actor.title}</span>
									<span class="actor-type">{actor.type}</span>
								</button>
							</li>
						{/each}
					</ul>
				</Panel>
			</aside>

			<!-- メインエリア: PixiJS キャンバス -->
			<main class="diagram-area">
				<div class="diagram-header">
					<h2>
						<Icon name="Target" size={20} />
						{data.boundary || 'System'}
					</h2>
					<div class="diagram-controls">
						<button class="zoom-btn" onclick={handleZoomOut} aria-label="ズームアウト">
							<Icon name="ZoomOut" size={16} />
						</button>
						<span class="zoom-level">{Math.round(currentZoom * 100)}%</span>
						<button class="zoom-btn" onclick={handleZoomIn} aria-label="ズームイン">
							<Icon name="ZoomIn" size={16} />
						</button>
						<button class="zoom-btn" onclick={handleZoomReset} aria-label="リセット">
							<Icon name="Maximize" size={16} />
						</button>
					</div>
					<div class="diagram-stats">
						<span>{data.actors.length} actors</span>
						<span>{data.usecases.length} usecases</span>
					</div>
				</div>
				<div class="canvas-wrapper" bind:this={canvasContainer}></div>
			</main>

			<!-- 右サイドバー: UseCase リスト -->
			<aside class="usecase-sidebar">
				<Panel title="ユースケース ({data.usecases.length})">
					<ul class="usecase-list">
						{#each data.usecases as usecase}
							<li>
								<button
									class="usecase-item"
									class:selected={selectedUseCaseId === usecase.id}
									onclick={() => handleUseCaseClick(usecase)}
								>
									<span
										class="usecase-status"
										style="background: {getStatusColor(usecase.status)}"
									></span>
									<span class="usecase-title">{usecase.title}</span>
									<span class="usecase-id">{usecase.id}</span>
								</button>
							</li>
						{/each}
					</ul>
				</Panel>
			</aside>
		</div>

		<!-- ホバーTooltip -->
		{#if hoveredActor}
			<div
				class="hover-tooltip"
				style={tooltipStyle()}
			>
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
			<div
				class="hover-tooltip"
				style={tooltipStyle()}
			>
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
					<div class="tooltip-actors">
						{hoveredUseCase.actors.length} actor(s) connected
					</div>
				{/if}
			</div>
		{/if}
	{/if}
</div>

<style>
	.usecase-view {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
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

	.usecase-layout {
		display: grid;
		grid-template-columns: 220px 1fr 260px;
		gap: 0;
		height: 100%;
		overflow: hidden;
	}

	.actor-sidebar,
	.usecase-sidebar {
		overflow-y: auto;
		background: var(--bg-secondary);
		border-right: 1px solid var(--border-metal);
	}

	.usecase-sidebar {
		border-right: none;
		border-left: 1px solid var(--border-metal);
	}

	.actor-list,
	.usecase-list {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.actor-item,
	.usecase-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		width: 100%;
		padding: 0.625rem 0.75rem;
		background: transparent;
		border: none;
		border-bottom: 1px solid var(--border-dark);
		color: var(--text-primary);
		cursor: pointer;
		text-align: left;
		transition: background var(--transition-fast);
	}

	.actor-item:hover,
	.usecase-item:hover {
		background: var(--bg-hover);
	}

	.actor-item.selected,
	.usecase-item.selected {
		background: var(--accent-primary);
		color: var(--bg-primary);
	}

	.actor-title,
	.usecase-title {
		flex: 1;
		font-weight: 500;
		font-size: 0.8125rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.actor-type {
		font-size: 0.6875rem;
		color: var(--text-muted);
		background: var(--bg-primary);
		padding: 0.125rem 0.375rem;
		border-radius: 2px;
	}

	.usecase-status {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.usecase-id {
		font-size: 0.625rem;
		color: var(--text-muted);
		font-family: var(--font-family);
	}

	.diagram-area {
		display: flex;
		flex-direction: column;
		background: var(--bg-primary);
		overflow: hidden;
	}

	.diagram-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.625rem 1rem;
		background: var(--bg-secondary);
		border-bottom: 1px solid var(--border-metal);
	}

	.diagram-header h2 {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin: 0;
		font-size: 0.9375rem;
		font-weight: 600;
	}

	.diagram-controls {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.zoom-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		background: var(--bg-primary);
		border: 1px solid var(--border-metal);
		border-radius: 4px;
		color: var(--text-secondary);
		cursor: pointer;
		transition: all var(--transition-fast);
	}

	.zoom-btn:hover {
		background: var(--bg-hover);
		color: var(--accent-primary);
		border-color: var(--accent-primary);
	}

	.zoom-level {
		font-size: 0.75rem;
		color: var(--text-muted);
		min-width: 40px;
		text-align: center;
		font-family: var(--font-family);
	}

	.diagram-stats {
		display: flex;
		gap: 1rem;
		font-size: 0.75rem;
		color: var(--text-muted);
	}

	.canvas-wrapper {
		flex: 1;
		overflow: hidden;
		position: relative;
	}

	.canvas-wrapper :global(canvas) {
		display: block;
	}

	/* ホバーTooltip */
	.hover-tooltip {
		position: fixed;
		z-index: 1000;
		background: var(--bg-panel);
		border: 1px solid var(--border-metal);
		border-radius: 4px;
		padding: 0.625rem;
		min-width: 180px;
		max-width: 280px;
		box-shadow: var(--shadow-tooltip);
		pointer-events: none;
	}

	.tooltip-header {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		font-weight: 600;
		font-size: 0.8125rem;
		margin-bottom: 0.375rem;
	}

	.tooltip-status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
	}

	.tooltip-meta {
		display: flex;
		gap: 0.5rem;
		font-size: 0.6875rem;
		color: var(--text-muted);
		margin-bottom: 0.375rem;
	}

	.tooltip-type,
	.tooltip-status {
		background: var(--bg-secondary);
		padding: 0.125rem 0.375rem;
		border-radius: 2px;
	}

	.tooltip-id {
		font-family: var(--font-family);
	}

	.tooltip-desc {
		font-size: 0.75rem;
		color: var(--text-secondary);
		line-height: 1.4;
		margin-top: 0.375rem;
	}

	.tooltip-actors {
		font-size: 0.6875rem;
		color: var(--accent-primary);
		margin-top: 0.375rem;
	}

	/* レスポンシブ対応 */
	@media (max-width: 1024px) {
		.usecase-layout {
			grid-template-columns: 180px 1fr 200px;
		}
	}

	@media (max-width: 768px) {
		.usecase-layout {
			grid-template-columns: 1fr;
			grid-template-rows: auto 1fr auto;
		}

		.actor-sidebar,
		.usecase-sidebar {
			max-height: 180px;
			border-right: none;
			border-left: none;
			border-bottom: 1px solid var(--border-metal);
		}

		.usecase-sidebar {
			border-bottom: none;
			border-top: 1px solid var(--border-metal);
		}
	}
</style>
