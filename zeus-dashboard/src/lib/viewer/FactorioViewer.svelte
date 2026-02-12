<script lang="ts">
	import { onMount, onDestroy, untrack } from 'svelte';
	import type { EntityStatus, GraphNode, GraphEdge, GraphData, UnifiedGraphGroupItem } from '$lib/types/api';
	import { NODE_TYPE_CONFIG } from './config/nodeTypes';
	import { ViewerEngine, type Viewport } from './engine/ViewerEngine';
	import {
		LayoutEngine,
		EDGE_ROUTING_GRID_UNIT,
		type NodePosition,
		type LayoutGroupBounds
	} from './engine/LayoutEngine';
	import { OrthogonalRouter } from './engine/OrthogonalRouter';
	import { SpatialIndex } from './engine/SpatialIndex';
	import { GraphNodeView, LODLevel } from './rendering/GraphNode';
	import { EdgeFactory, EdgeType } from './rendering/GraphEdge';
	import { GraphGroupBoundary } from './rendering/GraphGroupBoundary';
	import { SelectionManager } from './interaction/SelectionManager';
	import { FilterManager, type FilterCriteria } from './interaction/FilterManager';
	import Minimap from './ui/Minimap.svelte';
	import FilterPanel from './ui/FilterPanel.svelte';
	import GraphListPanel from './graph/GraphListPanel.svelte';
	import GraphDetailPanel from './graph/GraphDetailPanel.svelte';
	import { OverlayPanel } from '$lib/components/ui';
	import { updateGraphViewState, resetGraphViewState, pendingNavigation, clearPendingNavigation } from '$lib/stores/view';
	import { Container, Graphics } from 'pixi.js';
	import type { FederatedPointerEvent, Ticker } from 'pixi.js';

	// Props
	interface Props {
		graphData?: GraphData; // GraphNode/Edge データ
		selectedTaskId?: string | null;
		onTaskSelect?: (taskId: string | null) => void;
		onTaskHover?: (taskId: string | null) => void;
	}

	let { graphData, selectedTaskId = null, onTaskSelect, onTaskHover }: Props = $props();

	// 内部で使用する統一された GraphNode リスト
	let graphNodes = $derived(graphData?.nodes ?? []);

	// 内部で使用するエッジリスト
	let graphEdges = $derived(graphData?.edges ?? []);

	// Objective グループデータ
	let graphGroups = $derived(graphData?.groups ?? []);

	// 内部状態
	let containerElement: HTMLDivElement;
	let engine: ViewerEngine | null = null;
	let layoutEngine: LayoutEngine | null = null;
	let orthogonalRouter: OrthogonalRouter | null = null;
	let spatialIndex: SpatialIndex | null = null;
	let selectionManager: SelectionManager | null = null;
	let filterManager: FilterManager | null = null;

	let nodeMap: Map<string, GraphNodeView> = new Map();
	let edgeFactory: EdgeFactory = new EdgeFactory();
	let engineReady = $state(false); // エンジン初期化完了フラグ（$effect の依存関係用）
	let positions: Map<string, NodePosition> = $state(new Map());
	let layoutBounds = $state({ minX: 0, maxX: 0, minY: 0, maxY: 0, width: 0, height: 0 });
	const FLOW_FRAME_INTERVAL_MS = 1000 / 24;
	const FLOW_SPEED_PX_PER_SEC = 180;
	let flowTickerCallback: ((ticker: Ticker) => void) | null = null;
	let lastFlowAnimationMs = 0;

	// 差分更新用のキャッシュ
	let previousTasksHash: string = '';
	let previousTaskIds: Set<string> = new Set();
	let previousEdgeHash: string = '';

	// Visibility/LOD 差分更新用
	let previousVisibleNodeIds: Set<string> = new Set();
	let previousLODLevel: LODLevel | null = null;

	/**
	 * ノードリストのハッシュを計算（浅い比較用）
	 * Note: progress フィールドは削除されたため、status のみで判定
	 */
	function computeNodesHash(nodeList: GraphNode[]): string {
		return nodeList
			.map((n) => `${n.id}:${n.status}:${n.node_type}`)
			.join('|');
	}

	/**
	 * 依存関係のハッシュを計算（構造変更の検出用）
	 */
	function computeEdgeHash(edgeList: GraphEdge[]): string {
		return edgeList
			.map((e) => `${e.from}->${e.to}:${e.layer}:${e.relation}`)
			.sort()
			.join('|');
	}

	/**
	 * ノードが変更されたかチェック
	 * @returns 変更タイプ: 'none' | 'data' | 'structure'
	 */
	function detectNodeChanges(newNodes: GraphNode[], newEdges: GraphEdge[]): 'none' | 'data' | 'structure' {
		const newHash = computeNodesHash(newNodes);
		const newEdgeHash = computeEdgeHash(newEdges);
		const newIds = new Set(newNodes.map((n) => n.id));

		// 構造変更（追加/削除/依存関係変更）をチェック
		if (
			newEdgeHash !== previousEdgeHash ||
			newIds.size !== previousTaskIds.size ||
			!Array.from(newIds).every((id) => previousTaskIds.has(id))
		) {
			previousTasksHash = newHash;
			previousEdgeHash = newEdgeHash;
			previousTaskIds = newIds;
			return 'structure';
		}

		// データ変更（ステータス/進捗など）をチェック
		if (newHash !== previousTasksHash) {
			previousTasksHash = newHash;
			return 'data';
		}

		return 'none';
	}

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

	// ホバー時の影響範囲読み込みデバウンス用タイマー
	let hoverDebounceTimer: ReturnType<typeof setTimeout> | null = null;

	// フィルター状態
	let filterCriteria: FilterCriteria = $state({});
	let visibleTaskIds: Set<string> = $state(new Set());

	// 依存関係フィルター状態
	let dependencyFilterNodeId: string | null = $state(null); // フィルター対象ノードID
	let dependencyFilterIds: Set<string> = $state(new Set()); // 表示対象のノードID
	let originalPositions: Map<string, NodePosition> | null = $state(null); // フィルター前の元の位置

	// キー状態追跡（Chrome MCP ツール対応）
	let isAltKeyPressed = $state(false);

	// 選択中のID一覧
	let selectedIds: string[] = $state([]);

	// 選択中のグループID
	let selectedGroupId: string | null = $state(null);

	// 矩形選択用（将来機能用）
	let _isRectSelecting = $state(false);
	let _rectSelectStart: { x: number; y: number } | null = null;
	let rectSelectGraphics: Graphics | null = null;

	let resizeObserver: ResizeObserver | null = null;


	// UI パネル表示状態（Header 連携用）
	let showListPanel: boolean = $state(true);
	let showDetailPanel: boolean = $state(false);
	let showFilterPanel: boolean = $state(true);
	let showLegend: boolean = $state(true);
	let graphListSearchQuery: string = $state('');

	const selectedGraphNode = $derived.by(() => {
		if (!selectedTaskId) return null;
		return graphNodes.find((node) => node.id === selectedTaskId) ?? null;
	});

	const selectedGroup = $derived.by(() => {
		if (!selectedGroupId) return null;
		return graphGroups.find((g) => g.id === selectedGroupId) ?? null;
	});

	const visibleGraphNodes = $derived.by(() => {
		if (visibleTaskIds.size === 0) return graphNodes;
		return graphNodes.filter((node) => visibleTaskIds.has(node.id));
	});

	// 影響範囲可視化（下流タスクのハイライト）
	let highlightedDownstream: Set<string> = $state(new Set());
	let highlightedUpstream: Set<string> = $state(new Set());
	let showImpactHighlight: boolean = $state(true);
	let includeStructuralInTraversal: boolean = $state(true);

	type MetricsEntry = {
		timestamp: string;
		label: string;
		durationMs: number;
		[key: string]: unknown;
	};

	type MetricsPayload = {
		session_id: string;
		reason: string;
		meta: Record<string, unknown>;
		entries: MetricsEntry[];
	};

	const metricsParams =
		typeof window !== 'undefined' ? new URLSearchParams(window.location.search) : null;
	const METRICS_ENABLED =
		import.meta.env.DEV ||
		import.meta.env.MODE === 'test' ||
		metricsParams?.has('metrics') === true;
	const METRICS_VERBOSE = metricsParams?.has('metricsVerbose') === true;
	const METRICS_SLOW_THRESHOLD_MS = 50;
	const METRICS_MAX_ENTRIES = 2000;
	const METRICS_AUTOSAVE =
		METRICS_ENABLED &&
		(import.meta.env.MODE === 'test' || metricsParams?.has('metricsAutoSave') === true);
	const METRICS_FLUSH_INTERVAL_MS = 5000;
	const METRICS_ENDPOINT = '/api/metrics';
	let renderSequence = 0;
	let metricsEntries: MetricsEntry[] = $state([]);
	let metricsPending: MetricsEntry[] = $state([]);
	let metricsFlushTimer: ReturnType<typeof setInterval> | null = null;
	let isMetricsFlushing = false;
	const metricsSessionId = createMetricsSessionId();

	function nowMs(): number {
		return typeof performance !== 'undefined' && performance.now ? performance.now() : Date.now();
	}

	function getMemorySnapshot(): { usedMB: number; totalMB: number; limitMB: number } | null {
		const perf =
			typeof performance !== 'undefined'
				? (performance as Performance & {
						memory?: {
							usedJSHeapSize: number;
							totalJSHeapSize: number;
							jsHeapSizeLimit: number;
						};
					})
				: null;
		if (!perf?.memory) return null;
		return {
			usedMB: Math.round(perf.memory.usedJSHeapSize / 1024 / 1024),
			totalMB: Math.round(perf.memory.totalJSHeapSize / 1024 / 1024),
			limitMB: Math.round(perf.memory.jsHeapSizeLimit / 1024 / 1024)
		};
	}

	function createMetricsSessionId(): string {
		if (typeof window === 'undefined') return 'server';
		if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
			return crypto.randomUUID();
		}
		const ts = Date.now().toString(36);
		const rand = Math.random().toString(36).slice(2, 10);
		return `session-${ts}-${rand}`;
	}

	function getMetricsMeta(): Record<string, unknown> {
		const meta: Record<string, unknown> = {
			mode: import.meta.env.MODE,
			viewer: 'graph'
		};
		if (typeof location !== 'undefined') {
			meta.url = location.href;
		}
		if (typeof navigator !== 'undefined') {
			meta.userAgent = navigator.userAgent;
		}
		return meta;
	}

	function enqueueMetrics(entry: MetricsEntry): void {
		metricsEntries = [...metricsEntries, entry];
		if (metricsEntries.length > METRICS_MAX_ENTRIES) {
			metricsEntries = metricsEntries.slice(-METRICS_MAX_ENTRIES);
		}

		if (METRICS_AUTOSAVE) {
			metricsPending = [...metricsPending, entry];
			if (metricsPending.length > METRICS_MAX_ENTRIES) {
				metricsPending = metricsPending.slice(-METRICS_MAX_ENTRIES);
			}
		}

		if (typeof window !== 'undefined') {
			(window as Window & { __VIEWER_METRICS__?: MetricsEntry[] }).__VIEWER_METRICS__ =
				metricsEntries;
		}
	}

	function buildMetricsPayload(entries: MetricsEntry[], reason: string): MetricsPayload {
		return {
			session_id: metricsSessionId,
			reason,
			meta: getMetricsMeta(),
			entries
		};
	}

	async function flushMetrics(reason: string, useBeacon = false): Promise<void> {
		if (!METRICS_AUTOSAVE || metricsPending.length === 0) return;
		if (isMetricsFlushing && !useBeacon) return;

		const batch = metricsPending;
		metricsPending = [];

		const payload = buildMetricsPayload(batch, reason);
		const body = JSON.stringify(payload);

		if (
			useBeacon &&
			typeof navigator !== 'undefined' &&
			typeof navigator.sendBeacon === 'function'
		) {
			const ok = navigator.sendBeacon(
				METRICS_ENDPOINT,
				new Blob([body], { type: 'application/json' })
			);
			if (!ok) {
				metricsPending = [...batch, ...metricsPending];
			}
			return;
		}

		isMetricsFlushing = true;
		try {
			const response = await fetch(METRICS_ENDPOINT, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body,
				keepalive: reason === 'unload'
			});
			if (!response.ok) {
				throw new Error(`metrics upload failed: ${response.status}`);
			}
		} catch {
			metricsPending = [...batch, ...metricsPending];
		} finally {
			isMetricsFlushing = false;
		}
	}

	function handleMetricsPageHide(): void {
		void flushMetrics('pagehide', true);
	}

	function handleMetricsVisibilityChange(): void {
		if (typeof document === 'undefined') return;
		if (document.hidden) {
			void flushMetrics('visibility', true);
		}
	}

	function startMetricsAutoSave(): void {
		if (!METRICS_AUTOSAVE || typeof window === 'undefined') return;
		if (!metricsFlushTimer) {
			metricsFlushTimer = setInterval(() => {
				void flushMetrics('interval');
			}, METRICS_FLUSH_INTERVAL_MS);
		}
		window.addEventListener('pagehide', handleMetricsPageHide);
		if (typeof document !== 'undefined') {
			document.addEventListener('visibilitychange', handleMetricsVisibilityChange);
		}
	}

	function stopMetricsAutoSave(): void {
		if (metricsFlushTimer) {
			clearInterval(metricsFlushTimer);
			metricsFlushTimer = null;
		}
		if (typeof window !== 'undefined') {
			window.removeEventListener('pagehide', handleMetricsPageHide);
		}
		if (typeof document !== 'undefined') {
			document.removeEventListener('visibilitychange', handleMetricsVisibilityChange);
		}
	}

	function logMetrics(
		label: string,
		startMs: number,
		data: Record<string, unknown> = {},
		force = false
	): void {
		if (!METRICS_ENABLED) return;
		const durationMs = nowMs() - startMs;
		if (!force && !METRICS_VERBOSE && durationMs < METRICS_SLOW_THRESHOLD_MS) {
			return;
		}
		const durationMsRounded = Math.round(durationMs);
		const payload: Record<string, unknown> = {
			...data
		};
		const memory = getMemorySnapshot();
		if (memory) {
			payload.memoryMB = memory;
		}
		payload.durationMs = durationMsRounded;
		const entry: MetricsEntry = {
			timestamp: new Date().toISOString(),
			label,
			durationMs: durationMsRounded,
			...payload
		};
		enqueueMetrics(entry);
		console.debug(`[ViewerMetrics] ${label}`, payload);
	}

	function downloadMetrics(): void {
		if (!METRICS_ENABLED || metricsEntries.length === 0) return;
		if (typeof document === 'undefined') return;

		const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
		const filename = `zeus-viewer-metrics-${timestamp}.json`;
		const blob = new Blob([JSON.stringify(metricsEntries, null, 2)], { type: 'application/json' });
		const url = URL.createObjectURL(blob);

		const link = document.createElement('a');
		link.href = url;
		link.download = filename;
		link.click();
		URL.revokeObjectURL(url);
	}

	onMount(() => {
		initializeEngine();
		if (METRICS_AUTOSAVE) {
			startMetricsAutoSave();
		}
	});

	function startFlowAnimation(): void {
		if (!engine || flowTickerCallback) return;
		const app = engine.getApp();
		if (!app) return;

		lastFlowAnimationMs = 0;
		flowTickerCallback = (_ticker: Ticker) => {
			const now = nowMs();
			if (now - lastFlowAnimationMs < FLOW_FRAME_INTERVAL_MS) return;
			lastFlowAnimationMs = now;
			const phase = (now / 1000) * FLOW_SPEED_PX_PER_SEC;

			for (const edge of edgeFactory.getAll()) {
				if (!edge.visible) continue;
				edge.updateFlowAnimation(phase);
			}
		};

		app.ticker.add(flowTickerCallback);
	}

	function stopFlowAnimation(): void {
		if (!engine || !flowTickerCallback) return;
		const app = engine.getApp();
		if (app) {
			app.ticker.remove(flowTickerCallback);
		}
		flowTickerCallback = null;
	}

	// E2E テスト用: 開発/テスト環境でのみグローバルにデバッグヘルパーを公開
	$effect(() => {
		if ((import.meta.env.DEV || import.meta.env.MODE === 'test') && engine) {
			const win = window as Window & {
				__VIEWER_ENGINE__?: ViewerEngine;
				__NODE_MAP__?: Map<string, GraphNodeView>;
				__SELECTION_MANAGER__?: SelectionManager;
				__FILTER_MANAGER__?: FilterManager;
				__EDGE_FACTORY__?: EdgeFactory;
			};
			win.__VIEWER_ENGINE__ = engine;
			win.__NODE_MAP__ = nodeMap as Map<string, GraphNodeView>;
			win.__SELECTION_MANAGER__ = selectionManager ?? undefined;
			win.__FILTER_MANAGER__ = filterManager ?? undefined;
			win.__EDGE_FACTORY__ = edgeFactory;
		}
	});

	// E2E テスト用: 統合テスト API（開発/テスト環境、または ?e2e パラメータ付きで公開）
	$effect(() => {
		const isE2EMode =
			import.meta.env.DEV ||
			import.meta.env.MODE === 'test' ||
			new URLSearchParams(window.location.search).has('e2e');

		if (isE2EMode && engine) {
			const win = window as Window & {
				__ZEUS__?: {
					getGraphState: () => unknown;
					getSelectionState: () => unknown;
					getFilterState: () => unknown;
					isReady: () => boolean;
					getVersion: () => string;
				};
			};

			win.__ZEUS__ = {
				// グラフの論理構造を返す
				getGraphState: () => ({
					nodes: Array.from(nodeMap.values()).map((n) => ({
						id: n.getNodeId(),
						name: n.getGraphNode().title,
						x: Math.round(n.x),
						y: Math.round(n.y),
						status: n.getGraphNode().status,
						nodeType: n.getNodeType()
					})),
					edges: edgeFactory.getAll().map((e) => ({
						from: e.getFromId(),
						to: e.getToId()
					})),
					viewport: {
						zoom: currentViewport.scale,
						panX: Math.round(currentViewport.x),
						panY: Math.round(currentViewport.y)
					},
					nodeCount: nodeMap.size,
					edgeCount: edgeFactory.getAll().length
				}),

				// 選択状態を返す
				getSelectionState: () => ({
					selectedIds: selectedIds,
					count: selectedIds.length,
					multiSelect: selectedIds.length > 1
				}),

				// フィルター状態を返す
				getFilterState: () => ({
					criteria: filterCriteria,
					visibleCount: visibleTaskIds.size,
					totalCount: graphNodes.length
				}),

				// 描画完了を待機（エンジンが初期化済み）
				isReady: () => engine !== null,

				// バージョン情報
				getVersion: () => '0.1.0'
			};
		}
	});

	async function initializeEngine() {
		const initStart = nowMs();
		engine = new ViewerEngine();
		layoutEngine = new LayoutEngine();
		orthogonalRouter = new OrthogonalRouter(
			GraphNodeView.getWidth(),
			GraphNodeView.getHeight(),
			EDGE_ROUTING_GRID_UNIT
		);
		selectionManager = new SelectionManager();
		filterManager = new FilterManager();

		await engine.init(containerElement);
		logMetrics(
			'engine.init',
			initStart,
			{
				containerWidth: containerElement.clientWidth,
				containerHeight: containerElement.clientHeight
			},
			true
		);

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
					showDetailPanel = true;
					onTaskSelect?.(event.taskIds[0]);
				} else if (event.type === 'clear' || event.type === 'deselect') {
					if (selectedIds.length === 0) {
						showDetailPanel = false;
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
		window.addEventListener('keyup', handleKeyUp);

		// リサイズ監視
		resizeObserver = new ResizeObserver(() => {
			engine?.resize();
		});
		resizeObserver.observe(containerElement);

		// ブラウザのデフォルトコンテキストメニューを抑制し、右クリック処理
		containerElement.addEventListener('contextmenu', (e) => {
			console.log('[ContextMenu] contextmenu event fired at:', e.clientX, e.clientY);
			e.preventDefault();

			// キャンバス上の座標を取得
			const rect = containerElement.getBoundingClientRect();
			const x = e.clientX - rect.left;
			const y = e.clientY - rect.top;

			// PixiJS でヒットテストを実行
			if (engine) {
				const app = engine.getApp();
				if (!app) return;
				const hitObject = app.renderer.events.rootBoundary.hitTest(x, y);

				// TaskNode を検索（親をたどる）
				let target: Container | null = hitObject;
				while (target && !(target instanceof GraphNodeView)) {
					target = target.parent as Container | null;
				}

				if (target instanceof GraphNodeView) {
					console.log('[ContextMenu] Right-click on node:', target.getNodeId());
					handleNodeContextMenu(target, null as unknown as FederatedPointerEvent);
				}
			}
		});

		// エンジン準備完了を通知（$effect をトリガー）
		engineReady = true;
		startFlowAnimation();
	}

	onDestroy(() => {
		stopFlowAnimation();
		window.removeEventListener('keydown', handleKeyDown);
		window.removeEventListener('keyup', handleKeyUp);
		resizeObserver?.disconnect();

		// タイマーをクリア
		if (hoverDebounceTimer) {
			clearTimeout(hoverDebounceTimer);
			hoverDebounceTimer = null;
		}

		engine?.destroy();
		edgeFactory.clear();
		nodeMap.clear();
		selectionManager?.destroy();
		filterManager?.destroy();
		orthogonalRouter = null;

		if (METRICS_AUTOSAVE) {
			void flushMetrics('destroy', true);
		}
		stopMetricsAutoSave();

		// Store をリセット
		resetGraphViewState();
	});

	// ノードが変更されたら再レンダリング
	// NOTE: Svelte 5 の $effect は条件分岐内でのみ読み取られた変数を追跡しない
	// そのため graphNodes と engineReady を条件の外で明示的に読み取り、依存関係として登録する
	$effect(() => {
		const currentNodes = graphNodes; // 依存関係を明示的に登録
		const currentEdges = graphEdges;
		const ready = engineReady; // エンジン初期化完了を依存関係に追加
		if (ready && engine && layoutEngine) {
			const changeType = detectNodeChanges(currentNodes, currentEdges);

			if (changeType === 'none') {
				// 変更なし - 何もしない
				return;
			}

			if (changeType === 'data') {
				// データのみ変更 - ノードの表示を更新
				updateGraphNodes(currentNodes, currentEdges);
			} else {
				// 構造変更 - フルレンダリング
				renderGraphNodes(currentNodes, currentEdges);
			}
		}
	});


	/**
	 * 既存ノードのデータを更新（差分更新 - レイアウト再計算なし）
	 */
	function updateGraphNodes(nodeList: GraphNode[], edgeList: GraphEdge[]): void {
		if (!filterManager || !selectionManager) return;

		const updateStart = nowMs();
		let updatedCount = 0;

		// ノードマップを作成
		const graphNodeMap = new Map(nodeList.map((n) => [n.id, n]));

		// フィルターマネージャーを更新
		filterManager.setGraph(nodeList, edgeList);
		filterManager.setGroups(graphGroups);
		selectionManager.setGraph(nodeList, edgeList);


		// 既存ノードのデータを更新
		for (const [id, node] of nodeMap) {
			const graphNode = graphNodeMap.get(id);
			if (graphNode) {
				node.updateData(graphNode);
				updatedCount++;
			}
		}

		// フィルター適用
		visibleTaskIds = new Set(filterManager.getVisibleIds());
		updateVisibility();

		logMetrics(
			'updateGraphNodes',
			updateStart,
			{
				totalNodes: nodeMap.size,
				updatedNodes: updatedCount
			},
			true
		);
	}

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

	$effect(() => {
		showDetailPanel = selectedTaskId !== null;
	});

	/**
	 * グラフノードをレンダリング
	 */
	function renderGraphNodes(nodeList: GraphNode[], edgeList: GraphEdge[]): void {
		if (!engine || !layoutEngine || !spatialIndex || !filterManager || !selectionManager) return;

		const renderStart = nowMs();
		const renderSeq = ++renderSequence;
		const relationCount = edgeList.length;

		const nodeContainer = engine.getNodeContainer();
		const edgeContainer = engine.getEdgeContainer();
		const groupContainer = engine.getGroupContainer();
		const groupLabelContainer = engine.getGroupLabelContainer();
		if (!nodeContainer || !edgeContainer || !groupContainer || !groupLabelContainer) return;

		// 構造変更時は依存関係フィルター状態をリセット
		if (dependencyFilterNodeId !== null) {
			dependencyFilterNodeId = null;
			dependencyFilterIds = new Set();
			originalPositions = null;
		}

		// マネージャーに GraphNode を設定
		filterManager.setGraph(nodeList, edgeList);
		filterManager.setGroups(graphGroups);
		selectionManager.setGraph(nodeList, edgeList);


		// レイアウト計算（GraphNode + Groups で）
		const layoutStart = nowMs();
		const layout = layoutEngine.layout(nodeList, edgeList, graphGroups.length > 0 ? graphGroups : undefined);
		const layoutMs = nowMs() - layoutStart;
		positions = layout.positions;
		layoutBounds = layout.bounds;
		const componentCount = layout.groups.length;

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
		renderGroupBoundaries(layout.groups);

		// ノードIDのセットを作成
		const nodeIds = new Set(nodeList.map((n) => n.id));
		nodeMap.clear();

		// ノードを作成（GraphNode を直接渡す）
		for (const gn of nodeList) {
			const pos = positions.get(gn.id);
			if (!pos) continue;

			const node = new GraphNodeView(gn); // GraphNode を渡す
			node.x = pos.x - GraphNodeView.getWidth() / 2;
			node.y = pos.y - GraphNodeView.getHeight() / 2;

			// イベントハンドラ
			node.onClick((n, e) => handleNodeClick(n, e));
			node.onHover((n, isHovered) => handleNodeHover(n, isHovered));
			node.onContextMenu((n, e) => handleNodeContextMenu(n, e));

			// 選択状態を反映
			node.setSelected(selectionManager!.isSelected(gn.id));

			nodeContainer.addChild(node);
			nodeMap.set(gn.id, node);

			// 空間インデックスに追加
			spatialIndex.insert({
				id: gn.id,
				x: node.x,
				y: node.y,
				width: GraphNodeView.getWidth(),
				height: GraphNodeView.getHeight()
			});
		}

		const nodeByID = new Map(nodeList.map((n) => [n.id, n]));
		let orthogonalEdgeCount = 0;
		let routeFallbackCount = 0;
		for (const edgeData of edgeList) {
			if (!nodeIds.has(edgeData.from) || !nodeIds.has(edgeData.to)) continue;
			const fromPos = positions.get(edgeData.from);
			const toPos = positions.get(edgeData.to);
			if (!fromPos || !toPos) continue;

			const edge = edgeFactory.getOrCreate(
				edgeData.from,
				edgeData.to,
				edgeData.layer,
				edgeData.relation
			);
			const route = orthogonalRouter?.route(edgeData.from, edgeData.to, positions);
			if (route && route.points.length >= 2) {
				edge.setPolyline(route.points);
				orthogonalEdgeCount++;
				if (route.usedFallback) routeFallbackCount++;
			} else {
				const endpoints = layoutEngine!.computeEdgeEndpoints(fromPos, toPos);
				edge.setEndpoints(endpoints.fromX, endpoints.fromY, endpoints.toX, endpoints.toY);
			}

			edge.setType(EdgeType.Normal);

			edgeContainer.addChild(edge);
		}

		// フィルター適用
		visibleTaskIds = new Set(filterManager.getVisibleIds());
		previousLODLevel = null; // 新ノードは Micro 初期値なのでリセットし、次の updateLOD() で強制適用
		updateVisibility();

		// ビューを中央に
		if (nodeList.length > 0) {
			const centerX = (layout.bounds.minX + layout.bounds.maxX) / 2;
			const centerY = (layout.bounds.minY + layout.bounds.maxY) / 2;
			engine.panTo(centerX, centerY, false);
		}

		logMetrics(
			'renderGraphNodes',
			renderStart,
			{
				seq: renderSeq,
				nodes: nodeList.length,
				relations: relationCount,
				renderedNodes: nodeMap.size,
				edges: edgeFactory.getAll().length,
				layoutMs: Math.round(layoutMs),
				boundsW: Math.round(layout.bounds.width),
				boundsH: Math.round(layout.bounds.height),
				layers: layout.layers.length,
				groupCount: layout.groups.length,
				componentCount,
				laneCount: componentCount,
				layoutVersion: layout.layoutVersion,
				gridUnit: orthogonalRouter?.getGridUnit() ?? EDGE_ROUTING_GRID_UNIT,
				orthogonalEdgeCount,
				routeFallbackCount,
				visibleFilterCount: visibleTaskIds.size,
				viewportScale: Number(currentViewport.scale.toFixed(2))
			},
			true
		);
	}

	/**
	 * グループ境界を描画
	 */
	function renderGroupBoundaries(groups: LayoutGroupBounds[]): void {
		if (!engine) return;
		const groupContainer = engine.getGroupContainer();
		const groupLabelContainer = engine.getGroupLabelContainer();
		if (!groupContainer || !groupLabelContainer) return;

		// 既存の境界を destroy してからクリア（メモリリーク防止）
		for (const child of groupContainer.children) {
			if (child instanceof GraphGroupBoundary) {
				child.destroy();
			}
		}
		groupContainer.removeChildren();
		groupLabelContainer.removeChildren();
		// 面積が大きいものを先に描画（＝下層）、小さいものを後に描画（＝上層）
		const sortedGroups = [...groups].sort((a, b) => {
			const areaA = a.width * a.height;
			const areaB = b.width * b.height;
			return areaB - areaA;
		});

		for (const group of sortedGroups) {
			const boundary = new GraphGroupBoundary(group);

			// グループクリック/ホバーコールバック
			boundary.onClick((b, event) => handleGroupClick(b, event));
			boundary.onHover((_b, _isHovered) => {
				// ホバー時の追加処理（必要に応じて拡張）
			});

			// 選択状態を反映
			boundary.setSelected(group.groupId === selectedGroupId);

			groupContainer.addChild(boundary);
			groupLabelContainer.addChild(boundary.getLabelContainer());
		}
	}

	/**
	 * グループを選択し、フィルタ・レイアウトを適用する共通処理
	 */
	function selectGroup(groupId: string): void {
		selectedGroupId = groupId;
		if (selectionManager) selectionManager.clearSelection();
		onTaskSelect?.(null);
		showDetailPanel = true;
		filterManager?.updateCriteria({ groupIds: [groupId] });
		updateGroupSelectionState();
		updateGroupVisibility();
		applyGroupFilterLayout();
	}

	/**
	 * グループ選択を解除し、フィルタ・レイアウトを復元する共通処理
	 */
	function deselectGroup(): void {
		selectedGroupId = null;
		updateGroupSelectionState();
		updateGroupVisibility();
		filterManager?.clearCriterion('groupIds');
		clearGroupFilterLayout();
	}

	/**
	 * グループクリック処理
	 * グループ境界クリックは排他的（1グループのみ）選択 + フィルタリング
	 */
	function handleGroupClick(boundary: GraphGroupBoundary, event?: FederatedPointerEvent): void {
		const groupData = boundary.getGroupData();
		const nativeEvent = event?.nativeEvent as PointerEvent | undefined;
		const altKey = isAltKeyPressed || nativeEvent?.altKey;

		if (selectedGroupId === groupData.id) {
			// 再クリック → 選択解除（Alt の有無に関わらず、フィルタがあれば解除）
			deselectGroup();
			showDetailPanel = false;
		} else {
			if (altKey) {
				// Alt+click → フィルタリング（従来の動作）
				selectGroup(groupData.id);
			} else {
				// 通常クリック → 選択 + DetailPanel のみ（フィルタなし）
				// フィルタ中の場合は先に解除
				if (filterManager?.getCriteria().groupIds?.length) {
					filterManager?.clearCriterion('groupIds');
					clearGroupFilterLayout();
				}
				selectedGroupId = groupData.id;
				if (selectionManager) selectionManager.clearSelection();
				onTaskSelect?.(null);
				showDetailPanel = true;
				updateGroupSelectionState();
				updateGroupVisibility();
			}
		}
	}

	/**
	 * グループ境界の選択状態を更新
	 */
	function updateGroupSelectionState(): void {
		if (!engine) return;
		const groupContainer = engine.getGroupContainer();
		if (!groupContainer) return;

		for (const child of groupContainer.children) {
			if (child instanceof GraphGroupBoundary) {
				child.setSelected(child.getGroupId() === selectedGroupId);
			}
		}
	}

	/**
	 * グループフィルタに応じてグループ境界の表示/非表示を制御
	 */
	function updateGroupVisibility(): void {
		if (!engine) return;
		const groupContainer = engine.getGroupContainer();
		if (!groupContainer) return;

		const hasGroupFilter = selectedGroupId !== null;

		// 境界とラベルの表示/非表示を設定
		for (const child of groupContainer.children) {
			if (child instanceof GraphGroupBoundary) {
				const isVisible = !hasGroupFilter || child.getGroupId() === selectedGroupId;
				child.visible = isVisible;
				// getLabelContainer() で直接ラベルの visibility を同期（インデックス依存を排除）
				child.getLabelContainer().visible = isVisible;
			}
		}
	}

	/**
	 * グループフィルタ適用時の再レイアウト
	 * 可視ノードのみでコンパクトに再配置する
	 */
	function applyGroupFilterLayout(): void {
		if (!layoutEngine || !filterManager) return;

		// 元の位置を保存（初回のみ、依存関係フィルタと共有）
		if (!originalPositions) {
			originalPositions = new Map(positions);
		}

		// 可視ノードIDを取得
		const visibleIds = new Set(filterManager.getVisibleIds());
		if (visibleIds.size === 0) return;

		// 可視ノードのみで再レイアウト
		const filteredLayout = layoutEngine.layoutSubset(graphNodes, graphEdges, visibleIds);

		// 位置を更新
		updateNodePositions(filteredLayout.positions);
		layoutBounds = filteredLayout.bounds;
		renderGroupBoundaries(filteredLayout.groups);

		// ビューをフィット
		fitToView();
	}

	/**
	 * グループフィルタ解除時のレイアウト復元
	 * 依存関係フィルタが有効な場合はそちらに管理を任せる
	 */
	function clearGroupFilterLayout(): void {
		// 依存関係フィルタが有効な場合はスキップ（そちらの管理に任せる）
		if (dependencyFilterNodeId !== null) return;

		// 元の位置に戻す
		if (originalPositions) {
			updateNodePositions(originalPositions);

			// layoutBounds/groups を再計算
			if (layoutEngine) {
				const fullLayout = layoutEngine.layout(
					graphNodes,
					graphEdges,
					graphGroups.length > 0 ? graphGroups : undefined
				);
				layoutBounds = fullLayout.bounds;
				renderGroupBoundaries(fullLayout.groups);
			}

			originalPositions = null;

			// レイアウト復元時のみビューをリセット
			fitToView();
		}
	}

	/**
	 * 表示/非表示を更新（仮想化レンダリング - 差分更新対応）
	 */
	function updateVisibility(): void {
		if (!engine || !spatialIndex) return;

		const worldViewport = engine.getWorldViewport(200);

		// 空間インデックスで可視範囲のノードを取得
		const visibleInViewport = spatialIndex.queryRect(worldViewport);
		const currentVisibleIds = new Set<string>();

		// 可視判定を計算
		for (const item of visibleInViewport) {
			const passesFilter = visibleTaskIds.size === 0 || visibleTaskIds.has(item.id);
			if (passesFilter) {
				currentVisibleIds.add(item.id);
			}
		}

		// 差分検出: 新しく見えるようになったノード
		for (const id of currentVisibleIds) {
			if (!previousVisibleNodeIds.has(id)) {
				const node = nodeMap.get(id);
				if (node) {
					node.visible = true;
					// ビューポート外にいた間にLODが変わっている可能性があるため適用
					// Note: renderGraphNodes() 直後は previousLODLevel === null のため
					// setLOD はスキップされ、次の updateLOD() で全ノードに適用される
					if (previousLODLevel !== null) {
						node.setLOD(previousLODLevel);
					}
				}
			}
		}

		// 差分検出: 見えなくなったノード
		for (const id of previousVisibleNodeIds) {
			if (!currentVisibleIds.has(id)) {
				const node = nodeMap.get(id);
				if (node) node.visible = false;
			}
		}

		// 今回非表示のノードを確認（空間インデックス外）
		for (const [id, node] of nodeMap) {
			if (!currentVisibleIds.has(id) && !previousVisibleNodeIds.has(id)) {
				// 初回または空間インデックス外のノード
				node.visible = false; // ビューポート外
			}
		}

		// キャッシュを更新
		previousVisibleNodeIds = currentVisibleIds;

		// エッジの表示/非表示を更新
		for (const edge of edgeFactory.getAll()) {
			const fromId = edge.getFromId();
			const toId = edge.getToId();

			// 依存関係フィルターがアクティブな場合は両端がフィルター対象内であることを確認
			if (dependencyFilterIds.size > 0) {
				const fromInFilter = dependencyFilterIds.has(fromId);
				const toInFilter = dependencyFilterIds.has(toId);
				const fromVisible = currentVisibleIds.has(fromId);
				const toVisible = currentVisibleIds.has(toId);
				edge.visible = fromInFilter && toInFilter && (fromVisible || toVisible);
			} else {
				const fromVisible = currentVisibleIds.has(fromId);
				const toVisible = currentVisibleIds.has(toId);
				edge.visible = fromVisible || toVisible;
			}
		}
	}

	/**
	 * スケールからLODレベルを計算
	 */
	function computeLODLevel(scale: number): LODLevel {
		if (scale < 0.3) {
			return LODLevel.Macro;
		} else if (scale < 0.7) {
			return LODLevel.Meso;
		} else {
			return LODLevel.Micro;
		}
	}

	/**
	 * LODレベルを更新（条件付き - レベルが変わった時のみ）
	 */
	function updateLOD(scale: number): void {
		const lodLevel = computeLODLevel(scale);

		// LODレベルが変わっていない場合はスキップ
		if (lodLevel === previousLODLevel) return;
		previousLODLevel = lodLevel;

		// 可視ノードのみLODを更新
		for (const id of previousVisibleNodeIds) {
			const node = nodeMap.get(id);
			if (node) {
				node.setLOD(lodLevel);
			}
		}
	}

	/**
	 * ノードクリック処理
	 * コンテキストに応じた動作:
	 * - Alt+クリック: 依存関係フィルター（従来互換）
	 * - Shift+クリック: 依存チェーン選択
	 * - Ctrl/Cmd+クリック: マルチ選択（フィルタ変更なし）
	 * - グループフィルタ中: グループ維持 + ノード選択
	 * - フィルタなし/依存フィルタ中: 依存フィルタ適用（トグル）
	 */
	function handleNodeClick(node: GraphNodeView, event?: FederatedPointerEvent): void {
		if (!selectionManager) return;

		const taskId = node.getNodeId();
		const nativeEvent = event?.nativeEvent as PointerEvent | undefined;

		// Alt+クリック: 従来通り依存関係フィルター
		if (isAltKeyPressed || nativeEvent?.altKey) {
			handleNodeContextMenu(node, event!);
			return;
		}

		// Shift+クリック: 依存チェーン選択
		if (nativeEvent?.shiftKey && selectedIds.length > 0) {
			selectionManager.selectDependencyChain(taskId, 'both');
			return;
		}

		// Ctrl/Cmd+クリック: マルチ選択（フィルタ変更なし）
		if (nativeEvent?.ctrlKey || nativeEvent?.metaKey) {
			selectionManager.toggleSelect(taskId, true);
			return;
		}

		// グループフィルタ中 → グループ維持 + ノード選択のみ
		if (selectedGroupId) {
			selectionManager.toggleSelect(taskId, false);
			return;
		}

		// フィルタなし or 依存フィルタ中 → 依存関係フィルタ適用（トグル）
		const applied = applyNodeDependencyFilter(taskId);
		if (applied) {
			selectionManager.clearSelection();
			selectionManager.toggleSelect(taskId, false);
			// 明示的に DetailPanel を開く（イベントリスナーの暗黙フローに依存しない）
			showDetailPanel = true;
			onTaskSelect?.(taskId);
		} else {
			// フィルタ解除時
			selectionManager.clearSelection();
			onTaskSelect?.(null);
			showDetailPanel = false;
		}
	}

	/**
	 * ノードホバー処理（デバウンス付き）
	 */
	async function handleNodeHover(node: GraphNodeView, isHovered: boolean): Promise<void> {
		const taskId = isHovered ? node.getNodeId() : null;
		hoveredTaskId = taskId;
		onTaskHover?.(taskId);

		// 関連エッジをハイライト（即座に実行）
		highlightRelatedEdges(taskId);

		// 影響範囲可視化はデバウンス（300ms）
		if (hoverDebounceTimer) {
			clearTimeout(hoverDebounceTimer);
			hoverDebounceTimer = null;
		}

		if (showImpactHighlight && taskId) {
			hoverDebounceTimer = setTimeout(() => {
				highlightImpactedTasks(taskId);
				hoverDebounceTimer = null;
			}, 300);
		} else {
			clearImpactHighlight();
		}
	}

	/**
	 * 影響を受けるタスクをハイライト
	 * フロントエンドの edges データを使用して依存関係を計算
	 */
	function highlightImpactedTasks(taskId: string): void {
		// フロントエンドで依存関係を計算
		const downstream = getDownstreamNodes(taskId);
		const upstream = getUpstreamNodes(taskId);

		highlightedDownstream = new Set(downstream);
		highlightedUpstream = new Set(upstream);

		// ノードにハイライト状態を設定
		for (const [id, node] of nodeMap) {
			if (highlightedDownstream.has(id)) {
				node.setHighlighted(true, 'downstream');
			} else if (highlightedUpstream.has(id)) {
				node.setHighlighted(true, 'upstream');
			} else if (id !== taskId) {
				node.setHighlighted(false);
			}
		}
	}

	/**
	 * 影響範囲ハイライトをクリア
	 */
	function clearImpactHighlight(): void {
		highlightedDownstream = new Set();
		highlightedUpstream = new Set();

		for (const node of nodeMap.values()) {
			node.setHighlighted(false);
		}
	}

	function getTraversableEdges(includeStructural: boolean = includeStructuralInTraversal) {
		return edgeFactory.getAll().filter((e) => includeStructural || e.getLayer() !== 'structural');
	}

	/**
	 * 関連ノード集合を取得（無向連結成分）
	 *
	 * 依存の向きが途中で切り替わるパターン
	 * （A <- B -> C など）も取りこぼさないため、
	 * Alt+クリックのフィルターは無向探索で算出する。
	 */
	function getConnectedNodes(nodeId: string, includeStructural: boolean = includeStructuralInTraversal): string[] {
		const visited = new Set<string>([nodeId]);
		const queue = [nodeId];
		const traversableEdges = getTraversableEdges(includeStructural);

		while (queue.length > 0) {
			const current = queue.shift()!;
			for (const edge of traversableEdges) {
				if (edge.getFromId() === current) {
					const neighbor = edge.getToId();
					if (!visited.has(neighbor)) {
						visited.add(neighbor);
						queue.push(neighbor);
					}
				} else if (edge.getToId() === current) {
					const neighbor = edge.getFromId();
					if (!visited.has(neighbor)) {
						visited.add(neighbor);
						queue.push(neighbor);
					}
				}
			}
		}

		visited.delete(nodeId);
		return Array.from(visited);
	}

	/**
	 * 依存関係フィルタの適用/トグル（共通処理）
	 * @returns true = フィルタ適用, false = フィルタ解除
	 */
	function applyNodeDependencyFilter(nodeId: string): boolean {
		// 同じノード → トグル解除
		if (dependencyFilterNodeId === nodeId) {
			clearDependencyFilter();
			return false;
		}

		// 接続ノードを計算（既存の getConnectedNodes を使用）
		let relatedNodes = getConnectedNodes(nodeId);
		if (relatedNodes.length === 0 && !includeStructuralInTraversal) {
			const structuralRelated = getConnectedNodes(nodeId, true);
			if (structuralRelated.length > 0) {
				relatedNodes = structuralRelated;
			}
		}

		const filterIds = new Set<string>([nodeId, ...relatedNodes]);
		dependencyFilterNodeId = nodeId;
		dependencyFilterIds = filterIds;
		applyDependencyFilter();
		return true;
	}

	/**
	 * ノード右クリック処理 - 依存関係フィルター
	 * フロントエンドの edges データを使用して依存関係を計算
	 */
	function handleNodeContextMenu(node: GraphNodeView, _event: FederatedPointerEvent | null): void {
		const nodeId = node.getNodeId();
		// グループフィルタ中ならクリア
		if (selectedGroupId) {
			deselectGroup();
		}
		const applied = applyNodeDependencyFilter(nodeId);
		if (!applied) {
			// フィルタ解除時: 選択とパネルもクリア
			if (selectionManager) selectionManager.clearSelection();
			onTaskSelect?.(null);
			showDetailPanel = false;
		}
	}

	/**
	 * 上流ノードを取得（このノードが依存しているノード）
	 * BFS で間接的な依存関係も含める
	 */
	function getUpstreamNodes(nodeId: string): string[] {
		const visited = new Set<string>();
		const queue = [nodeId];
		const traversableEdges = getTraversableEdges();

		while (queue.length > 0) {
			const current = queue.shift()!;
			if (visited.has(current)) continue;
			visited.add(current);

			// このノードが依存しているエッジを探す（edge.from === current）
			for (const edge of traversableEdges) {
				if (edge.getFromId() === current && !visited.has(edge.getToId())) {
					queue.push(edge.getToId());
				}
			}
		}

		visited.delete(nodeId); // 自分自身は除外
		return Array.from(visited);
	}

	/**
	 * 下流ノードを取得（このノードに依存しているノード）
	 * BFS で間接的な依存関係も含める
	 */
	function getDownstreamNodes(nodeId: string): string[] {
		const visited = new Set<string>();
		const queue = [nodeId];
		const traversableEdges = getTraversableEdges();

		while (queue.length > 0) {
			const current = queue.shift()!;
			if (visited.has(current)) continue;
			visited.add(current);

			// このノードに依存しているエッジを探す（edge.to === current）
			for (const edge of traversableEdges) {
				if (edge.getToId() === current && !visited.has(edge.getFromId())) {
					queue.push(edge.getFromId());
				}
			}
		}

		visited.delete(nodeId); // 自分自身は除外
		return Array.from(visited);
	}

	/**
	 * ノード位置を更新し、空間インデックスを再構築
	 */
	function updateNodePositions(newPositions: Map<string, NodePosition>): void {
		positions = newPositions;

		// 各ノードの位置を更新
		for (const [nodeId, pos] of newPositions) {
			const node = nodeMap.get(nodeId);
			if (node) {
				node.x = pos.x - GraphNodeView.getWidth() / 2;
				node.y = pos.y - GraphNodeView.getHeight() / 2;
			}
		}

		// エッジの端点を再計算
		updateEdgeEndpoints(newPositions);

		// 空間インデックスを再構築
		rebuildSpatialIndex(newPositions);
	}

	/**
	 * エッジの端点を位置に基づいて更新
	 */
	function updateEdgeEndpoints(posMap: Map<string, NodePosition>): void {
		for (const edge of edgeFactory.getAll()) {
			const fromPos = posMap.get(edge.getFromId());
			const toPos = posMap.get(edge.getToId());
			if (fromPos && toPos) {
				const route = orthogonalRouter?.route(edge.getFromId(), edge.getToId(), posMap);
				if (route && route.points.length >= 2) {
					edge.setPolyline(route.points);
					continue;
				}
				if (layoutEngine) {
					const endpoints = layoutEngine.computeEdgeEndpoints(fromPos, toPos);
					edge.setEndpoints(endpoints.fromX, endpoints.fromY, endpoints.toX, endpoints.toY);
				}
			}
		}
	}

	/**
	 * 空間インデックスを再構築
	 */
	function rebuildSpatialIndex(posMap: Map<string, NodePosition>): void {
		if (!spatialIndex) return;

		spatialIndex.clear();

		// バウンディングボックスを計算
		let minX = Infinity,
			maxX = -Infinity;
		let minY = Infinity,
			maxY = -Infinity;

		for (const pos of posMap.values()) {
			minX = Math.min(minX, pos.x - GraphNodeView.getWidth() / 2);
			maxX = Math.max(maxX, pos.x + GraphNodeView.getWidth() / 2);
			minY = Math.min(minY, pos.y - GraphNodeView.getHeight() / 2);
			maxY = Math.max(maxY, pos.y + GraphNodeView.getHeight() / 2);
		}

		if (posMap.size > 0) {
			spatialIndex.rebuild({
				x: minX - 500,
				y: minY - 500,
				width: maxX - minX + 1000,
				height: maxY - minY + 1000
			});

			// ノードを空間インデックスに追加
			for (const [nodeId, pos] of posMap) {
				spatialIndex.insert({
					id: nodeId,
					x: pos.x - GraphNodeView.getWidth() / 2,
					y: pos.y - GraphNodeView.getHeight() / 2,
					width: GraphNodeView.getWidth(),
					height: GraphNodeView.getHeight()
				});
			}
		}
	}

	/**
	 * 現在のレイアウトに合わせてビューをフィット
	 */
	function fitToView(): void {
		if (!engine || layoutBounds.width === 0) return;

		const centerX = (layoutBounds.minX + layoutBounds.maxX) / 2;
		const centerY = (layoutBounds.minY + layoutBounds.maxY) / 2;
		engine.panTo(centerX, centerY, false);
	}

	/**
	 * 依存関係フィルターを適用（再レイアウト付き）
	 */
	function applyDependencyFilter(): void {
		if (dependencyFilterIds.size === 0 || !layoutEngine) return;

		// 元の位置を保存（初回のみ）
		if (!originalPositions) {
			originalPositions = new Map(positions);
		}

		// フィルター対象のみで再レイアウト
		const filteredLayout = layoutEngine.layoutSubset(graphNodes, graphEdges, dependencyFilterIds);

		// 可視性を更新
		visibleTaskIds = dependencyFilterIds;
		updateVisibility();

		// 位置を更新
		updateNodePositions(filteredLayout.positions);
		layoutBounds = filteredLayout.bounds;
		renderGroupBoundaries(filteredLayout.groups);

		// ビューをフィルター結果に合わせてフィット
		fitToView();
	}

	/**
	 * 依存関係フィルターを解除（元のレイアウトに戻す）
	 */
	function clearDependencyFilter(): void {
		dependencyFilterNodeId = null;
		dependencyFilterIds = new Set();

		// フィルターマネージャーの状態に戻す
		if (filterManager) {
			visibleTaskIds = new Set(filterManager.getVisibleIds());
		} else {
			visibleTaskIds = new Set();
		}
		updateVisibility();

		// 元の位置に戻す
		if (originalPositions) {
			updateNodePositions(originalPositions);

			// layoutBounds も再計算
			if (layoutEngine) {
				const fullLayout = layoutEngine.layout(
					graphNodes,
					graphEdges,
					graphGroups.length > 0 ? graphGroups : undefined
				);
				layoutBounds = fullLayout.bounds;
				renderGroupBoundaries(fullLayout.groups);
			}

			originalPositions = null;
		}

		// ビューをリセット
		fitToView();
	}

	function handleTraversalToggle(e: Event): void {
		includeStructuralInTraversal = (e.currentTarget as HTMLInputElement).checked;
		if (!dependencyFilterNodeId) return;

		// トグル変更時に現在のフィルターを再計算
		const relatedNodes = getConnectedNodes(dependencyFilterNodeId, includeStructuralInTraversal);
		dependencyFilterIds = new Set([dependencyFilterNodeId, ...relatedNodes]);
		applyDependencyFilter();
	}

	// 前回ハイライトしたエッジを追跡（差分更新用）
	let previousHighlightedEdges: Set<string> = new Set();

	/**
	 * 関連エッジをハイライト（差分更新対応）
	 */
	function highlightRelatedEdges(taskId: string | null): void {
		const currentHighlightedEdges = new Set<string>();

		if (taskId) {
			// インデックスを使って関連エッジのみ取得（O(1)）
			const relatedEdges = edgeFactory.getEdgesForNode(taskId);
			for (const edge of relatedEdges) {
				edge.setType(EdgeType.Highlighted);
				currentHighlightedEdges.add(edge.getKey());
			}
		}

		// 前回ハイライトされていたが今回されていないエッジを元に戻す
		for (const edgeKey of previousHighlightedEdges) {
			if (!currentHighlightedEdges.has(edgeKey)) {
				const edge = edgeFactory.getAll().find((e) => e.getKey() === edgeKey);
				if (edge) {
					edge.setType(EdgeType.Normal);
				}
			}
		}

		previousHighlightedEdges = currentHighlightedEdges;
	}

	/**
	 * キーボードイベント処理
	 */
	function handleKeyDown(e: KeyboardEvent): void {
		// Alt キー状態を追跡
		if (e.key === 'Alt') {
			isAltKeyPressed = true;
		}

		if (!selectionManager) return;

		// Escape: インタラクション由来のフィルタ解除 + 選択クリア
		// NOTE: ユーザーが FilterPanel で設定したフィルタ（ステータス、テキスト検索）は保持する
		if (e.key === 'Escape') {
			// 依存関係フィルタ解除
			if (dependencyFilterNodeId !== null) {
				clearDependencyFilter();
			}
			// グループ選択解除 + グループフィルタ解除
			if (selectedGroupId) {
				deselectGroup();
			}
			// 選択解除
			selectionManager.clearSelection();
			onTaskSelect?.(null);
			showDetailPanel = false;
		}

		// Ctrl/Cmd + A: 全選択
		if ((e.ctrlKey || e.metaKey) && e.key === 'a') {
			e.preventDefault();
			selectionManager.selectAll();
		}
	}

	/**
	 * キーアップイベント処理
	 */
	function handleKeyUp(e: KeyboardEvent): void {
		if (e.key === 'Alt') {
			isAltKeyPressed = false;
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
	function handleStatusToggle(status: EntityStatus): void {
		filterManager?.toggleStatus(status);
	}

	/**
	 * フィルター: 検索テキスト変更
	 */
	function handleSearchChange(text: string): void {
		filterManager?.setSearchText(text);
	}

	/**
	 * フィルター: グループトグル
	 */
	function handleGroupToggle(groupId: string): void {
		if (selectedGroupId === groupId) {
			deselectGroup();
			showDetailPanel = false;
		} else {
			selectGroup(groupId);
		}
	}

	/**
	 * フィルター: クリア
	 */
	function handleFilterClear(): void {
		// グループ選択中ならレイアウト・境界表示もリセット
		if (selectedGroupId) {
			deselectGroup();
			showDetailPanel = false;
		}
		filterManager?.clearFilter();
	}

	/**
	 * 特定タスクにフォーカス（将来機能用）
	 */
	function _focusTask(taskId: string): void {
		const pos = positions.get(taskId);
		if (pos && engine) {
			engine.panTo(pos.x, pos.y);
			engine.setZoom(1.5);
		}
	}

	// 表示されているタスク数（Store 同期用に保持、UI 表示なし）
	let visibleCount = $derived(Array.from(nodeMap.values()).filter((n) => n.visible).length);

	// パネルトグル関数
	function toggleListPanel(): void {
		showListPanel = !showListPanel;
	}

	function closeListPanel(): void {
		showListPanel = false;
	}

	function toggleFilterPanel(): void {
		showFilterPanel = !showFilterPanel;
	}

	function toggleLegend(): void {
		showLegend = !showLegend;
	}

	function closeDetailPanel(): void {
		showDetailPanel = false;
		if (selectedGroupId) {
			deselectGroup();
		}
		if (selectionManager) {
			selectionManager.clearSelection();
		} else {
			onTaskSelect?.(null);
		}
	}

	function handleGraphListSearchChange(value: string): void {
		graphListSearchQuery = value;
	}

	/**
	 * パネルからのグループ選択処理
	 */
	function handleGroupSelectFromPanel(groupId: string): void {
		if (selectedGroupId === groupId) return;
		selectGroup(groupId);
	}

	function handleGraphNodeSelectFromPanel(nodeId: string): void {
		if (selectionManager) {
			if (selectedTaskId !== nodeId || selectedIds.length !== 1) {
				selectionManager.clearSelection();
				selectionManager.toggleSelect(nodeId, false);
			}
		} else {
			onTaskSelect?.(nodeId);
		}

		showDetailPanel = true;
		const pos = positions.get(nodeId);
		if (pos && engine) {
			engine.panTo(pos.x, pos.y, false);
		}
	}

	/**
	 * Store へのコールバック登録（一度だけ実行）
	 */
	let callbacksRegistered = false;
	function registerStoreCallbacks(): void {
		if (callbacksRegistered) return;
		updateGraphViewState({
			onZoomIn: zoomIn,
			onZoomOut: zoomOut,
			onZoomReset: resetZoom,
			onToggleListPanel: toggleListPanel,
			onToggleFilterPanel: toggleFilterPanel,
			onToggleLegend: toggleLegend,
			onClearDependencyFilter: clearDependencyFilter
		});
		callbacksRegistered = true;
	}

	/**
	 * Store へのデータ同期
	 */
	function syncStoreData(): void {
		updateGraphViewState({
			zoom: currentViewport.scale,
			nodeCount: graphNodes.length,
			visibleCount,
			showListPanel,
			showFilterPanel,
			showLegend,
			hasDependencyFilter: dependencyFilterNodeId !== null,
			dependencyFilterNodeId
		});
	}

	// コールバック登録 Effect（engineReady 時に一度だけ）
	$effect(() => {
		if (engineReady) {
			registerStoreCallbacks();
		}
	});

	// Store へのデータ同期 Effect
	$effect(() => {
		// 依存関係を読み取り、変更時に再実行
		if (
			engineReady &&
			(currentViewport.scale !== undefined ||
				graphNodes.length !== undefined ||
				visibleCount !== undefined ||
				showListPanel !== undefined ||
				showFilterPanel !== undefined ||
				showLegend !== undefined ||
				dependencyFilterNodeId !== undefined)
		) {
			syncStoreData();
		}
	});

	// 他ビューからの Graph ビューナビゲーション受信
	$effect(() => {
		const nav = $pendingNavigation;
		if (!nav || nav.view !== 'graph') return;
		// エンジン未準備の場合はナビゲーションを保留（データロード後に再実行される）
		if (!selectionManager || graphNodes.length === 0) return;
		untrack(() => {
			if (nav.entityId) {
				const nodeExists = graphNodes.some((n) => n.id === nav.entityId);
				if (nodeExists) {
					handleGraphNodeSelectFromPanel(nav.entityId);
				}
			}
			clearPendingNavigation();
		});
	});
</script>

<div class="factorio-viewer">
	<!-- キャンバスコンテナ -->
	<div class="canvas-container" bind:this={containerElement}></div>

	<!-- ノード一覧パネル（オーバーレイ） -->
	{#if showListPanel}
		<OverlayPanel
			title="NODE LIST"
			position="top-left"
			panelId="graph-list"
			defaultWidthPreset="medium"
			onClose={closeListPanel}
		>
			<GraphListPanel
				nodes={visibleGraphNodes}
				selectedNodeId={selectedTaskId ?? null}
				searchQuery={graphListSearchQuery}
				visibleCount={visibleGraphNodes.length}
				totalCount={graphNodes.length}
				onNodeSelect={handleGraphNodeSelectFromPanel}
				onSearchChange={handleGraphListSearchChange}
			/>
		</OverlayPanel>
	{/if}

	<!-- 詳細パネル（オーバーレイ） -->
	{#if showDetailPanel && (selectedGraphNode || selectedGroup)}
		<OverlayPanel
			title="PROPERTIES"
			position="top-right"
			panelId="graph-detail"
			defaultWidthPreset="medium"
			onClose={closeDetailPanel}
		>
			<div class="detail-content">
				<GraphDetailPanel
					node={selectedGraphNode}
					nodes={graphNodes}
					edges={graphEdges}
					group={selectedGroup}
					groups={graphGroups}
					onNodeSelect={handleGraphNodeSelectFromPanel}
					onGroupSelect={handleGroupSelectFromPanel}
				/>
			</div>
		</OverlayPanel>
	{/if}

	<!-- フィルターパネル（オーバーレイ） -->
	{#if showFilterPanel}
		<OverlayPanel
			title="FILTER"
			position="bottom-right"
			panelId="graph-filter"
			defaultWidthPreset="narrow"
			onClose={toggleFilterPanel}
		>
			<FilterPanel
				criteria={filterCriteria}
				groups={graphGroups}
				onStatusToggle={handleStatusToggle}
				onSearchChange={handleSearchChange}
				onGroupToggle={handleGroupToggle}
				onClear={handleFilterClear}
			/>
		</OverlayPanel>
	{/if}

	<!-- ミニマップ -->
	<Minimap
		nodes={graphNodes}
		{positions}
		bounds={layoutBounds}
		viewport={currentViewport}
		onNavigate={handleMinimapNavigate}
	/>

	<!-- 凡例（オーバーレイ） -->
	{#if showLegend}
		<OverlayPanel
			title="LEGEND"
			position="bottom-left"
			panelId="graph-legend"
			defaultWidthPreset="narrow"
			onClose={toggleLegend}
		>
			<div class="legend-content">
				<!-- ノードタイプ凡例（NODE_TYPE_CONFIG から生成） -->
				<div class="legend-section">
					<div class="legend-section-title">NODE TYPE</div>
					{#each Object.entries(NODE_TYPE_CONFIG) as [type, config]}
						<div class="legend-item">
							<span class="legend-dot" style:background-color={config.cssColor}></span>
							<span>{type[0].toUpperCase() + type.slice(1)}</span>
						</div>
					{/each}
				</div>
					<div class="legend-section">
						<div class="legend-section-title">STATUS</div>
					<div class="legend-item">
						<span class="legend-dot draft"></span>
						<span>Draft</span>
					</div>
					<div class="legend-item">
						<span class="legend-dot active"></span>
						<span>Active</span>
					</div>
					<div class="legend-item">
						<span class="legend-dot deprecated"></span>
						<span>Deprecated</span>
						</div>
					</div>
					<div class="legend-section">
						<div class="legend-section-title">RELATION</div>
						<div class="legend-item">
							<span class="legend-line structural"></span>
							<span>Structural (parent / implements)</span>
						</div>
					</div>
					<div class="legend-section">
						<div class="legend-section-title">TRAVERSAL</div>
						<label class="legend-toggle">
								<input
									type="checkbox"
									checked={includeStructuralInTraversal}
									onchange={handleTraversalToggle}
								/>
							<span>Include structural edges in impact filter</span>
						</label>
					</div>
				</div>
			</OverlayPanel>
		{/if}
	</div>

<style>
	.factorio-viewer {
		position: relative;
		width: 100%;
		height: 100%;
		min-height: 400px;
		background-color: var(--bg-primary);
		overflow: hidden;
	}

	.canvas-container {
		width: 100%;
		height: 100%;
	}

	.canvas-container :global(canvas) {
		display: block;
	}

	.detail-content {
		padding: 12px;
	}

	/* 凡例（オーバーレイパネル内） */
	.legend-content {
		padding: var(--spacing-sm);
		font-size: var(--font-size-xs);
	}

	.legend-section {
		margin-bottom: var(--spacing-sm);
	}

	.legend-section:last-child {
		margin-bottom: 0;
	}

	.legend-section-title {
		color: var(--accent-primary);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		margin-bottom: var(--spacing-xs);
		font-size: 0.65rem;
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

	.legend-dot.draft {
		background-color: var(--task-pending);
	}

	.legend-dot.active {
		background-color: var(--task-in-progress);
	}

	.legend-dot.deprecated {
		background-color: var(--text-muted);
	}

	.legend-line {
		display: inline-block;
		width: 16px;
		height: 2px;
	}

	.legend-line.structural {
		background-color: #88b8ff;
		height: 3px;
	}

	.legend-toggle {
		display: flex;
		align-items: center;
		gap: var(--spacing-xs);
		color: var(--text-secondary);
		font-size: 0.72rem;
	}

	/* ノードタイプ別の色は NODE_TYPE_CONFIG から inline style で適用 */
</style>
