<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { TaskItem, TaskStatus, Priority } from '$lib/types/api';
	import { ViewerEngine, type Viewport } from './engine/ViewerEngine';
	import { LayoutEngine, type NodePosition } from './engine/LayoutEngine';
	import { SpatialIndex, type SpatialItem } from './engine/SpatialIndex';
	import { TaskNode, LODLevel } from './rendering/TaskNode';
	import { TaskEdge, EdgeFactory, EdgeType } from './rendering/TaskEdge';
	import { SelectionManager } from './interaction/SelectionManager';
	import { FilterManager, type FilterCriteria } from './interaction/FilterManager';
	import Minimap from './ui/Minimap.svelte';
	import FilterPanel from './ui/FilterPanel.svelte';
	import { Graphics } from 'pixi.js';

	// Props
	interface Props {
		tasks: TaskItem[];
		selectedTaskId?: string | null;
		onTaskSelect?: (taskId: string | null) => void;
		onTaskHover?: (taskId: string | null) => void;
	}

	let {
		tasks = [],
		selectedTaskId = null,
		onTaskSelect,
		onTaskHover
	}: Props = $props();

	// 内部状態
	let containerElement: HTMLDivElement;
	let engine: ViewerEngine | null = null;
	let layoutEngine: LayoutEngine | null = null;
	let spatialIndex: SpatialIndex | null = null;
	let selectionManager: SelectionManager | null = null;
	let filterManager: FilterManager | null = null;

	let nodeMap: Map<string, TaskNode> = new Map();
	let edgeFactory: EdgeFactory = new EdgeFactory();
	let positions: Map<string, NodePosition> = $state(new Map());
	let layoutBounds = $state({ minX: 0, maxX: 0, minY: 0, maxY: 0, width: 0, height: 0 });
	let layers: string[][] = [];

	// ビューポート情報（UI表示用）
	let currentViewport: Viewport = $state({
		x: 0,
		y: 0,
		width: 0,
		height: 0,
		scale: 1.0
	});

	// ホバー中のタスク
	let hoveredTaskId: string | null = $state(null);

	// フィルター状態
	let filterCriteria: FilterCriteria = $state({});
	let availableAssignees: string[] = $state([]);
	let visibleTaskIds: Set<string> = $state(new Set());

	// 選択中のID一覧
	let selectedIds: string[] = $state([]);

	// 矩形選択用
	let isRectSelecting = $state(false);
	let rectSelectStart: { x: number; y: number } | null = null;
	let rectSelectGraphics: Graphics | null = null;

	let resizeObserver: ResizeObserver | null = null;

	onMount(() => {
		initializeEngine();
	});

	async function initializeEngine() {
		engine = new ViewerEngine();
		layoutEngine = new LayoutEngine();
		selectionManager = new SelectionManager();
		filterManager = new FilterManager();

		await engine.init(containerElement);

		// 空間インデックスを初期化（十分な大きさのバウンド）
		spatialIndex = new SpatialIndex({
			x: -10000,
			y: -10000,
			width: 20000,
			height: 20000
		});

		// イベントリスナー設定
		engine.onViewportChanged((viewport) => {
			currentViewport = viewport;
			updateLOD(viewport.scale);
			updateVisibility();
		});

		// 選択マネージャーのイベント
		selectionManager.onSelectionChange((event) => {
			selectedIds = selectionManager!.getSelectedIds();

			// ノードの選択状態を更新
			for (const [id, node] of nodeMap) {
				node.setSelected(selectionManager!.isSelected(id));
			}

			// 最初の選択IDを親に通知
			if (event.taskIds.length > 0) {
				if (event.type === 'select') {
					onTaskSelect?.(event.taskIds[0]);
				} else if (event.type === 'clear' || event.type === 'deselect') {
					if (selectedIds.length === 0) {
						onTaskSelect?.(null);
					}
				}
			}
		});

		// フィルターマネージャーのイベント
		filterManager.onFilterChange((event) => {
			filterCriteria = event.criteria;
			visibleTaskIds = new Set(event.visibleIds);
			updateVisibility();
		});

		// 矩形選択用のグラフィックス
		const worldContainer = engine.getWorldContainer();
		if (worldContainer) {
			rectSelectGraphics = new Graphics();
			worldContainer.addChild(rectSelectGraphics);
		}

		// キーボードイベント
		window.addEventListener('keydown', handleKeyDown);

		// リサイズ監視
		resizeObserver = new ResizeObserver(() => {
			engine?.resize();
		});
		resizeObserver.observe(containerElement);

		// 初回レンダリング
		renderTasks(tasks);
	}

	onDestroy(() => {
		window.removeEventListener('keydown', handleKeyDown);
		resizeObserver?.disconnect();
		engine?.destroy();
		edgeFactory.clear();
		nodeMap.clear();
		selectionManager?.destroy();
		filterManager?.destroy();
	});

	// タスクが変更されたら再レンダリング
	$effect(() => {
		if (engine && layoutEngine) {
			renderTasks(tasks);
		}
	});

	// 外部からの選択状態変更を反映
	$effect(() => {
		if (selectionManager && selectedTaskId !== undefined) {
			const currentSelected = selectionManager.getSelectedIds();
			if (selectedTaskId === null && currentSelected.length > 0) {
				selectionManager.clearSelection();
			} else if (selectedTaskId && !currentSelected.includes(selectedTaskId)) {
				selectionManager.clearSelection();
				selectionManager.toggleSelect(selectedTaskId);
			}
		}
	});

	/**
	 * タスクをレンダリング
	 */
	function renderTasks(taskList: TaskItem[]): void {
		if (!engine || !layoutEngine || !spatialIndex || !filterManager || !selectionManager) return;

		const nodeContainer = engine.getNodeContainer();
		const edgeContainer = engine.getEdgeContainer();
		if (!nodeContainer || !edgeContainer) return;

		// マネージャーにタスクを設定
		filterManager.setTasks(taskList);
		selectionManager.setTasks(taskList);
		availableAssignees = filterManager.getAvailableAssignees();

		// レイアウト計算
		const layout = layoutEngine.layout(taskList);
		positions = layout.positions;
		layoutBounds = layout.bounds;
		layers = layout.layers;

		// 空間インデックスをクリアして再構築
		spatialIndex.clear();
		if (layout.bounds.width > 0 && layout.bounds.height > 0) {
			spatialIndex.rebuild({
				x: layout.bounds.minX - 500,
				y: layout.bounds.minY - 500,
				width: layout.bounds.width + 1000,
				height: layout.bounds.height + 1000
			});
		}

		// 既存のノードをクリア
		nodeContainer.removeChildren();
		edgeFactory.clear();
		edgeContainer.removeChildren();

		// タスクIDのセットを作成
		const taskIds = new Set(taskList.map((t) => t.id));
		nodeMap.clear();

		// ノードを作成
		for (const task of taskList) {
			const pos = positions.get(task.id);
			if (!pos) continue;

			const node = new TaskNode(task);
			node.x = pos.x - TaskNode.getWidth() / 2;
			node.y = pos.y - TaskNode.getHeight() / 2;

			// イベントハンドラ
			node.onClick((n, e) => handleNodeClick(n, e));
			node.onHover((n, isHovered) => handleNodeHover(n, isHovered));

			// 選択状態を反映
			node.setSelected(selectionManager!.isSelected(task.id));

			nodeContainer.addChild(node);
			nodeMap.set(task.id, node);

			// 空間インデックスに追加
			spatialIndex.insert({
				id: task.id,
				x: node.x,
				y: node.y,
				width: TaskNode.getWidth(),
				height: TaskNode.getHeight()
			});
		}

		// エッジを作成
		for (const task of taskList) {
			const toPos = positions.get(task.id);
			if (!toPos) continue;

			for (const depId of task.dependencies) {
				// 依存先がタスクリストに存在する場合のみエッジを作成
				if (!taskIds.has(depId)) continue;

				const fromPos = positions.get(depId);
				if (!fromPos) continue;

				const edge = edgeFactory.getOrCreate(depId, task.id);

				// エッジの端点を計算
				const endpoints = layoutEngine!.computeEdgeEndpoints(fromPos, toPos);
				edge.setEndpoints(endpoints.fromX, endpoints.fromY, endpoints.toX, endpoints.toY);

				// エッジタイプを設定
				const depTask = taskList.find((t) => t.id === depId);
				if (depTask) {
					if (depTask.status !== 'completed' && task.status === 'blocked') {
						edge.setType(EdgeType.Blocked);
					} else {
						edge.setType(EdgeType.Normal);
					}
				}

				edgeContainer.addChild(edge);
			}
		}

		// フィルター適用
		visibleTaskIds = new Set(filterManager.getVisibleIds());
		updateVisibility();

		// ビューを中央に
		if (taskList.length > 0) {
			const centerX = (layout.bounds.minX + layout.bounds.maxX) / 2;
			const centerY = (layout.bounds.minY + layout.bounds.maxY) / 2;
			engine.panTo(centerX, centerY, false);
		}
	}

	/**
	 * 表示/非表示を更新（仮想化レンダリング）
	 */
	function updateVisibility(): void {
		if (!engine || !spatialIndex) return;

		const worldViewport = engine.getWorldViewport(200);

		// 空間インデックスで可視範囲のノードを取得
		const visibleInViewport = spatialIndex.queryRect(worldViewport);
		const visibleInViewportIds = new Set(visibleInViewport.map((item) => item.id));

		// ノードの表示/非表示を更新
		for (const [id, node] of nodeMap) {
			const isInViewport = visibleInViewportIds.has(id);
			const passesFilter = visibleTaskIds.size === 0 || visibleTaskIds.has(id);

			// ビューポート内かつフィルターを通過したノードのみ表示
			node.visible = isInViewport && passesFilter;
		}

		// エッジの表示/非表示を更新
		for (const edge of edgeFactory.getAll()) {
			const fromVisible = nodeMap.get(edge.getFromId())?.visible ?? false;
			const toVisible = nodeMap.get(edge.getToId())?.visible ?? false;
			edge.visible = fromVisible || toVisible;
		}
	}

	/**
	 * LODレベルを更新
	 */
	function updateLOD(scale: number): void {
		let lodLevel: LODLevel;

		if (scale < 0.3) {
			lodLevel = LODLevel.Macro;
		} else if (scale < 0.7) {
			lodLevel = LODLevel.Meso;
		} else {
			lodLevel = LODLevel.Micro;
		}

		for (const node of nodeMap.values()) {
			node.setLOD(lodLevel);
		}
	}

	/**
	 * ノードクリック処理
	 */
	function handleNodeClick(node: TaskNode, event?: PointerEvent): void {
		if (!selectionManager) return;

		const taskId = node.getTaskId();
		const isMulti = event?.ctrlKey || event?.metaKey || event?.shiftKey;

		if (event?.shiftKey && selectedIds.length > 0) {
			// Shift+クリック: 依存チェーン選択
			selectionManager.selectDependencyChain(taskId, 'both');
		} else {
			selectionManager.toggleSelect(taskId, isMulti);
		}
	}

	/**
	 * ノードホバー処理
	 */
	function handleNodeHover(node: TaskNode, isHovered: boolean): void {
		const taskId = isHovered ? node.getTaskId() : null;
		hoveredTaskId = taskId;
		onTaskHover?.(taskId);

		// 関連エッジをハイライト
		highlightRelatedEdges(taskId);
	}

	/**
	 * 関連エッジをハイライト
	 */
	function highlightRelatedEdges(taskId: string | null): void {
		for (const edge of edgeFactory.getAll()) {
			if (taskId && (edge.getFromId() === taskId || edge.getToId() === taskId)) {
				edge.setType(EdgeType.Highlighted);
			} else {
				// 元のタイプに戻す
				edge.setType(EdgeType.Normal);
			}
		}
	}

	/**
	 * キーボードイベント処理
	 */
	function handleKeyDown(e: KeyboardEvent): void {
		if (!selectionManager) return;

		// Escape: 選択クリア
		if (e.key === 'Escape') {
			selectionManager.clearSelection();
		}

		// Ctrl/Cmd + A: 全選択
		if ((e.ctrlKey || e.metaKey) && e.key === 'a') {
			e.preventDefault();
			selectionManager.selectAll();
		}
	}

	/**
	 * ズームイン
	 */
	function zoomIn(): void {
		engine?.setZoom(currentViewport.scale * 1.2);
	}

	/**
	 * ズームアウト
	 */
	function zoomOut(): void {
		engine?.setZoom(currentViewport.scale / 1.2);
	}

	/**
	 * ズームリセット
	 */
	function resetZoom(): void {
		engine?.setZoom(1.0);
		engine?.centerView();
	}

	/**
	 * ミニマップからのナビゲーション
	 */
	function handleMinimapNavigate(x: number, y: number): void {
		engine?.panTo(x, y, false);
	}

	/**
	 * フィルター: ステータストグル
	 */
	function handleStatusToggle(status: TaskStatus): void {
		filterManager?.toggleStatus(status);
	}

	/**
	 * フィルター: 優先度トグル
	 */
	function handlePriorityToggle(priority: Priority): void {
		filterManager?.togglePriority(priority);
	}

	/**
	 * フィルター: 担当者トグル
	 */
	function handleAssigneeToggle(assignee: string): void {
		filterManager?.toggleAssignee(assignee);
	}

	/**
	 * フィルター: 検索テキスト変更
	 */
	function handleSearchChange(text: string): void {
		filterManager?.setSearchText(text);
	}

	/**
	 * フィルター: クリア
	 */
	function handleFilterClear(): void {
		filterManager?.clearFilter();
	}

	/**
	 * 特定タスクにフォーカス
	 */
	function focusTask(taskId: string): void {
		const pos = positions.get(taskId);
		if (pos && engine) {
			engine.panTo(pos.x, pos.y);
			engine.setZoom(1.5);
		}
	}

	// 表示されているタスク数
	let visibleCount = $derived(
		Array.from(nodeMap.values()).filter((n) => n.visible).length
	);
</script>

<div class="factorio-viewer">
	<!-- キャンバスコンテナ -->
	<div class="canvas-container" bind:this={containerElement}></div>

	<!-- ビューコントロール -->
	<div class="view-controls">
		<button class="control-btn" onclick={zoomIn} title="Zoom In">
			<span class="icon">+</span>
		</button>
		<button class="control-btn" onclick={zoomOut} title="Zoom Out">
			<span class="icon">-</span>
		</button>
		<button class="control-btn" onclick={resetZoom} title="Reset View">
			<span class="icon">⊙</span>
		</button>
	</div>

	<!-- フィルターパネル -->
	<FilterPanel
		criteria={filterCriteria}
		{availableAssignees}
		onStatusToggle={handleStatusToggle}
		onPriorityToggle={handlePriorityToggle}
		onAssigneeToggle={handleAssigneeToggle}
		onSearchChange={handleSearchChange}
		onClear={handleFilterClear}
	/>

	<!-- ミニマップ -->
	<Minimap
		{tasks}
		{positions}
		bounds={layoutBounds}
		viewport={currentViewport}
		onNavigate={handleMinimapNavigate}
	/>

	<!-- ステータスバー -->
	<div class="status-bar">
		<span class="status-item">
			Zoom: {(currentViewport.scale * 100).toFixed(0)}%
		</span>
		<span class="status-item"> Tasks: {visibleCount}/{tasks.length} </span>
		{#if selectedIds.length > 0}
			<span class="status-item selection-info"> Selected: {selectedIds.length} </span>
		{/if}
		{#if hoveredTaskId}
			<span class="status-item hover-info">
				{hoveredTaskId}
			</span>
		{/if}
	</div>

	<!-- 凡例 -->
	<div class="legend">
		<div class="legend-title">STATUS</div>
		<div class="legend-item">
			<span class="legend-dot completed"></span>
			<span>Completed</span>
		</div>
		<div class="legend-item">
			<span class="legend-dot in-progress"></span>
			<span>In Progress</span>
		</div>
		<div class="legend-item">
			<span class="legend-dot pending"></span>
			<span>Pending</span>
		</div>
		<div class="legend-item">
			<span class="legend-dot blocked"></span>
			<span>Blocked</span>
		</div>
	</div>

	<!-- 操作ヒント -->
	<div class="hints">
		<div class="hint-item">Scroll: Zoom</div>
		<div class="hint-item">Shift+Drag: Pan</div>
		<div class="hint-item">Shift+Click: Chain Select</div>
	</div>
</div>

<style>
	.factorio-viewer {
		position: relative;
		width: 100%;
		height: 100%;
		min-height: 600px;
		background-color: var(--bg-primary);
		border: 2px solid var(--border-metal);
		border-radius: var(--border-radius-md);
		overflow: hidden;
	}

	.canvas-container {
		width: 100%;
		height: 100%;
	}

	.canvas-container :global(canvas) {
		display: block;
	}

	/* ビューコントロール */
	.view-controls {
		position: absolute;
		top: var(--spacing-md);
		right: var(--spacing-md);
		display: flex;
		flex-direction: column;
		gap: var(--spacing-xs);
	}

	.control-btn {
		width: 36px;
		height: 36px;
		display: flex;
		align-items: center;
		justify-content: center;
		background-color: var(--bg-panel);
		border: 2px solid var(--border-metal);
		border-radius: var(--border-radius-sm);
		color: var(--text-primary);
		font-size: 18px;
		cursor: pointer;
		transition: all var(--transition-fast);
	}

	.control-btn:hover {
		background-color: var(--bg-hover);
		border-color: var(--accent-primary);
	}

	.control-btn .icon {
		line-height: 1;
	}

	/* ステータスバー */
	.status-bar {
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		display: flex;
		gap: var(--spacing-lg);
		padding: var(--spacing-sm) var(--spacing-md);
		background-color: rgba(26, 26, 26, 0.9);
		border-top: 1px solid var(--border-dark);
		font-size: var(--font-size-xs);
		color: var(--text-secondary);
	}

	.status-item {
		display: flex;
		align-items: center;
		gap: var(--spacing-xs);
	}

	.hover-info {
		color: var(--accent-primary);
	}

	.selection-info {
		color: var(--status-info);
	}

	/* 凡例 */
	.legend {
		position: absolute;
		top: var(--spacing-md);
		left: var(--spacing-md);
		padding: var(--spacing-sm);
		background-color: rgba(45, 45, 45, 0.9);
		border: 1px solid var(--border-metal);
		border-radius: var(--border-radius-sm);
		font-size: var(--font-size-xs);
	}

	.legend-title {
		color: var(--accent-primary);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		margin-bottom: var(--spacing-xs);
	}

	.legend-item {
		display: flex;
		align-items: center;
		gap: var(--spacing-xs);
		color: var(--text-secondary);
		margin-bottom: 2px;
	}

	.legend-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
	}

	.legend-dot.completed {
		background-color: var(--task-completed);
	}

	.legend-dot.in-progress {
		background-color: var(--task-in-progress);
	}

	.legend-dot.pending {
		background-color: var(--task-pending);
	}

	.legend-dot.blocked {
		background-color: var(--status-poor);
	}

	/* 操作ヒント */
	.hints {
		position: absolute;
		bottom: 40px;
		left: var(--spacing-md);
		display: flex;
		gap: var(--spacing-md);
		font-size: 10px;
		color: var(--text-muted);
	}

	.hint-item {
		background-color: rgba(0, 0, 0, 0.5);
		padding: 2px 6px;
		border-radius: var(--border-radius-sm);
	}
</style>
