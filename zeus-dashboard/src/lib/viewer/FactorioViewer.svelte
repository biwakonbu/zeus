<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { TaskItem, TaskStatus, Priority, GraphNode, WBSGraphData, TimelineItem } from '$lib/types/api';
	import { fetchTimeline } from '$lib/api/client';
	import { ViewerEngine, type Viewport } from './engine/ViewerEngine';
	import { LayoutEngine, type NodePosition } from './engine/LayoutEngine';
	import { SpatialIndex } from './engine/SpatialIndex';
	import { TaskNode, LODLevel } from './rendering/TaskNode';
	import { EdgeFactory, EdgeType } from './rendering/TaskEdge';
	import { SelectionManager } from './interaction/SelectionManager';
	import { FilterManager, type FilterCriteria } from './interaction/FilterManager';
	import Minimap from './ui/Minimap.svelte';
	import FilterPanel from './ui/FilterPanel.svelte';
	import { Container, Graphics, FederatedPointerEvent } from 'pixi.js';

	// Props
	interface Props {
		tasks?: TaskItem[];
		graphData?: WBSGraphData;  // WBS モード用の GraphNode/Edge データ
		selectedTaskId?: string | null;
		onTaskSelect?: (taskId: string | null) => void;
		onTaskHover?: (taskId: string | null) => void;
	}

	let {
		tasks = [],
		graphData,
		selectedTaskId = null,
		onTaskSelect,
		onTaskHover
	}: Props = $props();

	// WBS モード判定: graphData が提供されていれば WBS モード
	let isWBSMode = $derived(!!graphData && graphData.nodes.length > 0);

	/**
	 * TaskItem[] から GraphNode[] への変換（後方互換性用）
	 */
	function tasksToGraphNodes(taskList: TaskItem[]): GraphNode[] {
		return taskList.map(t => ({
			id: t.id,
			title: t.title,
			node_type: 'task' as const,
			status: t.status,
			progress: t.progress ?? 0,
			priority: t.priority,
			assignee: t.assignee || undefined,
			wbs_code: t.wbs_code,
			dependencies: t.dependencies
		}));
	}

	// 内部で使用する統一された GraphNode リスト
	let graphNodes = $derived(
		isWBSMode ? graphData!.nodes : tasksToGraphNodes(tasks)
	);

	// 内部で使用するエッジリスト
	let graphEdges = $derived(
		isWBSMode ? graphData!.edges : []
	);

	// 内部状態
	let containerElement: HTMLDivElement;
	let engine: ViewerEngine | null = null;
	let layoutEngine: LayoutEngine | null = null;
	let spatialIndex: SpatialIndex | null = null;
	let selectionManager: SelectionManager | null = null;
	let filterManager: FilterManager | null = null;

	let nodeMap: Map<string, TaskNode> = new Map();
	let edgeFactory: EdgeFactory = new EdgeFactory();
	let engineReady = $state(false); // エンジン初期化完了フラグ（$effect の依存関係用）
	let positions: Map<string, NodePosition> = $state(new Map());
	let layoutBounds = $state({ minX: 0, maxX: 0, minY: 0, maxY: 0, width: 0, height: 0 });

	// 差分更新用のキャッシュ
	let previousTasksHash: string = '';
	let previousTaskIds: Set<string> = new Set();
	let previousDependencyHash: string = '';

	// Visibility/LOD 差分更新用
	let previousVisibleNodeIds: Set<string> = new Set();
	let previousLODLevel: LODLevel | null = null;

	/**
	 * ノードリストのハッシュを計算（浅い比較用）
	 */
	function computeNodesHash(nodeList: GraphNode[]): string {
		return nodeList.map(n =>
			`${n.id}:${n.status}:${n.progress ?? 0}:${n.priority ?? ''}:${n.assignee ?? ''}:${n.node_type}`
		).join('|');
	}

	/**
	 * 依存関係のハッシュを計算（構造変更の検出用）
	 */
	function computeDependencyHash(nodeList: GraphNode[]): string {
		return nodeList.map(n => `${n.id}:${n.dependencies.join(',')}`).sort().join('|');
	}

	/**
	 * ノードが変更されたかチェック
	 * @returns 変更タイプ: 'none' | 'data' | 'structure'
	 */
	function detectNodeChanges(newNodes: GraphNode[]): 'none' | 'data' | 'structure' {
		const newHash = computeNodesHash(newNodes);
		const newDepHash = computeDependencyHash(newNodes);
		const newIds = new Set(newNodes.map(n => n.id));

		// 構造変更（追加/削除/依存関係変更）をチェック
		if (newDepHash !== previousDependencyHash ||
			newIds.size !== previousTaskIds.size ||
			!Array.from(newIds).every(id => previousTaskIds.has(id))) {
			previousTasksHash = newHash;
			previousDependencyHash = newDepHash;
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
	let availableAssignees: string[] = $state([]);
	let visibleTaskIds: Set<string> = $state(new Set());

	// 依存関係フィルター状態
	let dependencyFilterNodeId: string | null = $state(null);  // フィルター対象ノードID
	let dependencyFilterIds: Set<string> = $state(new Set());  // 表示対象のノードID
	let originalPositions: Map<string, NodePosition> | null = $state(null);  // フィルター前の元の位置

	// キー状態追跡（Chrome MCP ツール対応）
	let isAltKeyPressed = $state(false);

	// 選択中のID一覧
	let selectedIds: string[] = $state([]);

	// 矩形選択用（将来機能用）
	let _isRectSelecting = $state(false);
	let _rectSelectStart: { x: number; y: number } | null = null;
	let rectSelectGraphics: Graphics | null = null;

	let resizeObserver: ResizeObserver | null = null;

	// クリティカルパス情報
	let criticalPathIds: Set<string> = $state(new Set());
	let timelineItems: Map<string, TimelineItem> = $state(new Map());
	let showCriticalPath: boolean = $state(true);
	let isLoadingTimeline: boolean = $state(false);
	let timelineLoadError: string | null = $state(null);

	// 影響範囲可視化（下流タスクのハイライト）
	let highlightedDownstream: Set<string> = $state(new Set());
	let highlightedUpstream: Set<string> = $state(new Set());
	let showImpactHighlight: boolean = $state(true);

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

	const metricsParams = typeof window !== 'undefined' ? new URLSearchParams(window.location.search) : null;
	const METRICS_ENABLED = import.meta.env.DEV || import.meta.env.MODE === 'test' || metricsParams?.has('metrics') === true;
	const METRICS_VERBOSE = metricsParams?.has('metricsVerbose') === true;
	const METRICS_SLOW_THRESHOLD_MS = 50;
	const METRICS_MAX_ENTRIES = 2000;
	const METRICS_AUTOSAVE = METRICS_ENABLED && (import.meta.env.MODE === 'test' || metricsParams?.has('metricsAutoSave') === true);
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
		const perf = typeof performance !== 'undefined'
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
			(window as Window & { __VIEWER_METRICS__?: MetricsEntry[] }).__VIEWER_METRICS__ = metricsEntries;
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

		if (useBeacon && typeof navigator !== 'undefined' && typeof navigator.sendBeacon === 'function') {
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

	// E2E テスト用: 開発/テスト環境でのみグローバルにデバッグヘルパーを公開
	$effect(() => {
		if ((import.meta.env.DEV || import.meta.env.MODE === 'test') && engine) {
			const win = window as Window & {
				__VIEWER_ENGINE__?: ViewerEngine;
				__NODE_MAP__?: Map<string, TaskNode>;
				__SELECTION_MANAGER__?: SelectionManager;
				__FILTER_MANAGER__?: FilterManager;
				__EDGE_FACTORY__?: EdgeFactory;
				__CRITICAL_PATH_IDS__?: Set<string>;
			};
			win.__VIEWER_ENGINE__ = engine;
			win.__NODE_MAP__ = nodeMap;
			win.__SELECTION_MANAGER__ = selectionManager ?? undefined;
			win.__FILTER_MANAGER__ = filterManager ?? undefined;
			win.__EDGE_FACTORY__ = edgeFactory;
			win.__CRITICAL_PATH_IDS__ = criticalPathIds;
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
					getCriticalPathState: () => unknown;
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
						progress: n.getGraphNode().progress ?? 0,
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
					edgeCount: edgeFactory.getAll().length,
					mode: isWBSMode ? 'wbs' : 'task'
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

				// クリティカルパス状態を返す
				getCriticalPathState: () => ({
					enabled: showCriticalPath,
					criticalPathIds: Array.from(criticalPathIds),
					criticalPathCount: criticalPathIds.size
				}),

				// 描画完了を待機（アニメーション中でなく、エンジンが初期化済み）
				isReady: () => engine !== null && !isLoadingTimeline,

				// バージョン情報
				getVersion: () => '0.1.0'
			};
		}
	});

	async function initializeEngine() {
		const initStart = nowMs();
		engine = new ViewerEngine();
		layoutEngine = new LayoutEngine();
		selectionManager = new SelectionManager();
		filterManager = new FilterManager();

		await engine.init(containerElement);
		logMetrics('engine.init', initStart, {
			containerWidth: containerElement.clientWidth,
			containerHeight: containerElement.clientHeight
		}, true);

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
				while (target && !(target instanceof TaskNode)) {
					target = target.parent as Container | null;
				}

				if (target instanceof TaskNode) {
					console.log('[ContextMenu] Right-click on node:', target.getNodeId());
					handleNodeContextMenu(target, null as unknown as FederatedPointerEvent);
				}
			}
		});

		// エンジン準備完了を通知（$effect をトリガー）
		engineReady = true;
	}

	onDestroy(() => {
		window.removeEventListener('keydown', handleKeyDown);
		window.removeEventListener('keyup', handleKeyUp);
		resizeObserver?.disconnect();

		// タイマーをクリア
		if (timelineLoadTimer) {
			clearTimeout(timelineLoadTimer);
			timelineLoadTimer = null;
		}
		if (hoverDebounceTimer) {
			clearTimeout(hoverDebounceTimer);
			hoverDebounceTimer = null;
		}

		engine?.destroy();
		edgeFactory.clear();
		nodeMap.clear();
		selectionManager?.destroy();
		filterManager?.destroy();

		if (METRICS_AUTOSAVE) {
			void flushMetrics('destroy', true);
		}
		stopMetricsAutoSave();
	});

	// タイムライン読み込みのデバウンス用タイマー
	let timelineLoadTimer: ReturnType<typeof setTimeout> | null = null;

	// ノードが変更されたら再レンダリング
	// NOTE: Svelte 5 の $effect は条件分岐内でのみ読み取られた変数を追跡しない
	// そのため graphNodes と engineReady を条件の外で明示的に読み取り、依存関係として登録する
	$effect(() => {
		const currentNodes = graphNodes; // 依存関係を明示的に登録
		const ready = engineReady; // エンジン初期化完了を依存関係に追加
		if (ready && engine && layoutEngine) {
			const changeType = detectNodeChanges(currentNodes);

			if (changeType === 'none') {
				// 変更なし - 何もしない
				return;
			}

			if (changeType === 'data') {
				// データのみ変更 - ノードの表示を更新
				updateGraphNodes(currentNodes);
			} else {
				// 構造変更 - フルレンダリング
				renderGraphNodes(currentNodes);
			}

			// タイムラインデータの読み込みをデバウンス（500ms）- Task モードのみ
			if (!isWBSMode) {
				if (timelineLoadTimer) {
					clearTimeout(timelineLoadTimer);
				}
				timelineLoadTimer = setTimeout(() => {
					loadTimelineData();
					timelineLoadTimer = null;
				}, 500);
			}
		}
	});

	/**
	 * タイムラインデータを読み込み、クリティカルパス情報を更新
	 */
	async function loadTimelineData(): Promise<void> {
		if (isLoadingTimeline) return; // 既に読み込み中の場合はスキップ

		const timelineStart = nowMs();
		let result: 'ok' | 'invalid' | 'error' = 'ok';
		let itemsCount = 0;
		let validItems = 0;
		let invalidItems = 0;
		let criticalCount = 0;
		let errorMessage: string | null = null;

		isLoadingTimeline = true;
		timelineLoadError = null;

		try {
			const timeline = await fetchTimeline();

			// レスポンスの検証
			if (!timeline) {
				console.debug('Timeline response is empty');
				timelineLoadError = 'Timeline data is empty';
				result = 'invalid';
				errorMessage = timelineLoadError;
				resetCriticalPath();
				return;
			}

			// critical_path は配列として存在する必要がある
			if (!Array.isArray(timeline.critical_path)) {
				const errMsg = 'Invalid timeline response: critical_path is not an array';
				console.warn(errMsg);
				timelineLoadError = errMsg;
				result = 'invalid';
				errorMessage = errMsg;
				resetCriticalPath();
				return;
			}

			// items は配列として存在する必要がある
			if (!Array.isArray(timeline.items)) {
				const errMsg = 'Invalid timeline response: items is not an array';
				console.warn(errMsg);
				timelineLoadError = errMsg;
				result = 'invalid';
				errorMessage = errMsg;
				resetCriticalPath();
				return;
			}

			itemsCount = timeline.items.length;

			// クリティカルパスのIDセットを更新（文字列型チェック）
			const validCriticalPaths = timeline.critical_path.filter((id) => typeof id === 'string');
			criticalPathIds = new Set(validCriticalPaths);
			criticalCount = validCriticalPaths.length;

			// タイムラインアイテムをマップに格納（スラック値を検証）
			const itemMap = new Map<string, TimelineItem>();
			for (const item of timeline.items) {
				// 必須フィールドの存在確認
				if (!item.task_id || typeof item.task_id !== 'string') {
					console.warn('Invalid timeline item: missing or invalid task_id', item);
					invalidItems++;
					continue;
				}

				// スラック値のバリデーション（null, undefined, 非負整数のみ許可）
				if (item.slack !== null && item.slack !== undefined) {
					if (typeof item.slack !== 'number' || item.slack < 0 || !Number.isFinite(item.slack)) {
						console.warn(`Invalid slack value for task ${item.task_id}:`, item.slack);
						item.slack = null;
					}
				}

				itemMap.set(item.task_id, item);
				validItems++;
			}
			timelineItems = itemMap;
			timelineLoadError = null;

			// ノードとエッジにクリティカルパス情報を適用
			applyCriticalPathInfo();
		} catch (error) {
			// API エラーや予期しないエラーをログに記録
			const errMsg = error instanceof Error ? error.message : 'Unknown error loading timeline data';
			console.warn('Error loading timeline data:', error);
			timelineLoadError = errMsg;
			result = 'error';
			errorMessage = errMsg;
			resetCriticalPath();
		} finally {
			isLoadingTimeline = false;
			logMetrics('timeline.load', timelineStart, {
				result,
				items: itemsCount,
				validItems,
				invalidItems,
				criticalPath: criticalCount,
				error: errorMessage
			}, true);
		}
	}

	/**
	 * クリティカルパス情報をリセット
	 */
	function resetCriticalPath(): void {
		criticalPathIds = new Set();
		timelineItems = new Map();

		// ノードとエッジをリセット
		for (const node of nodeMap.values()) {
			node.setCriticalPath(false);
			node.setSlack(null);
		}
		for (const edge of edgeFactory.getAll()) {
			if (edge.getToId() && edge.getFromId()) {
				edge.setType(EdgeType.Normal);
			}
		}
	}

	/**
	 * クリティカルパス情報をノードとエッジに適用
	 */
	function applyCriticalPathInfo(): void {
		if (!showCriticalPath) {
			// クリティカルパス表示がオフの場合はクリア
			for (const node of nodeMap.values()) {
				node.setCriticalPath(false);
				node.setSlack(null);
			}
			for (const edge of edgeFactory.getAll()) {
				if (edge.getToId() && edge.getFromId()) {
					edge.setType(EdgeType.Normal);
				}
			}
			return;
		}

		// ノードにクリティカルパス・スラック情報を設定
		for (const [id, node] of nodeMap) {
			const isOnCritical = criticalPathIds.has(id);
			node.setCriticalPath(isOnCritical);

			// スラック値を設定
			const timelineItem = timelineItems.get(id);
			if (timelineItem) {
				node.setSlack(timelineItem.slack);
			} else {
				node.setSlack(null);
			}
		}

		// エッジにクリティカルパス情報を設定
		for (const edge of edgeFactory.getAll()) {
			const fromId = edge.getFromId();
			const toId = edge.getToId();

			// 両端がクリティカルパス上にある場合、エッジもクリティカル
			if (criticalPathIds.has(fromId) && criticalPathIds.has(toId)) {
				edge.setType(EdgeType.Critical);
			}
		}
	}

	/**
	 * 既存ノードのデータを更新（差分更新 - レイアウト再計算なし）
	 */
	function updateGraphNodes(nodeList: GraphNode[]): void {
		if (!filterManager || !selectionManager) return;

		const updateStart = nowMs();
		let updatedCount = 0;

		// ノードマップを作成
		const graphNodeMap = new Map(nodeList.map(n => [n.id, n]));

		// フィルターマネージャーを更新（TaskItem 互換形式で）
		const taskItems = graphNodesToTaskItems(nodeList);
		filterManager.setTasks(taskItems);
		availableAssignees = filterManager.getAvailableAssignees();

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

		logMetrics('updateGraphNodes', updateStart, {
			totalNodes: nodeMap.size,
			updatedNodes: updatedCount
		}, true);
	}

	/**
	 * GraphNode[] から TaskItem[] 互換形式に変換（フィルター・選択マネージャー用）
	 */
	function graphNodesToTaskItems(nodeList: GraphNode[]): TaskItem[] {
		return nodeList.map(n => ({
			id: n.id,
			title: n.title,
			status: (n.status === 'completed' || n.status === 'in_progress' || n.status === 'pending' || n.status === 'blocked')
				? n.status as TaskStatus
				: 'pending' as TaskStatus,
			priority: (n.priority === 'high' || n.priority === 'medium' || n.priority === 'low')
				? n.priority as Priority
				: 'medium' as Priority,
			assignee: n.assignee || '',
			dependencies: n.dependencies,
			progress: n.progress,
			wbs_code: n.wbs_code
		}));
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

	/**
	 * グラフノードをレンダリング
	 */
	function renderGraphNodes(nodeList: GraphNode[]): void {
		if (!engine || !layoutEngine || !spatialIndex || !filterManager || !selectionManager) return;

		const renderStart = nowMs();
		const renderSeq = ++renderSequence;
		let dependencyCount = 0;
		for (const gn of nodeList) {
			dependencyCount += gn.dependencies.length;
		}

		const nodeContainer = engine.getNodeContainer();
		const edgeContainer = engine.getEdgeContainer();
		if (!nodeContainer || !edgeContainer) return;

		// 構造変更時は依存関係フィルター状態をリセット
		if (dependencyFilterNodeId !== null) {
			dependencyFilterNodeId = null;
			dependencyFilterIds = new Set();
			originalPositions = null;
		}

		// TaskItem 互換形式に変換してマネージャーに設定
		const taskItems = graphNodesToTaskItems(nodeList);
		filterManager.setTasks(taskItems);
		selectionManager.setTasks(taskItems);
		availableAssignees = filterManager.getAvailableAssignees();

		// レイアウト計算（TaskItem 互換形式で）
		const layoutStart = nowMs();
		const layout = layoutEngine.layout(taskItems);
		const layoutMs = nowMs() - layoutStart;
		positions = layout.positions;
		layoutBounds = layout.bounds;
		// layers は将来のグループ化機能用に保持可能だが、現在は不使用

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

		// ノードIDのセットを作成
		const nodeIds = new Set(nodeList.map((n) => n.id));
		nodeMap.clear();

		// ノードを作成（GraphNode を直接渡す）
		for (const gn of nodeList) {
			const pos = positions.get(gn.id);
			if (!pos) continue;

			const node = new TaskNode(gn);  // GraphNode を渡す
			node.x = pos.x - TaskNode.getWidth() / 2;
			node.y = pos.y - TaskNode.getHeight() / 2;

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
				width: TaskNode.getWidth(),
				height: TaskNode.getHeight()
			});
		}

		// エッジを作成（WBS モードの場合は graphEdges を使用）
		if (isWBSMode && graphEdges.length > 0) {
			// WBS モード: graphEdges から直接エッジを作成
			for (const edge of graphEdges) {
				if (!nodeIds.has(edge.from) || !nodeIds.has(edge.to)) continue;

				const fromPos = positions.get(edge.from);
				const toPos = positions.get(edge.to);
				if (!fromPos || !toPos) continue;

				const taskEdge = edgeFactory.getOrCreate(edge.from, edge.to);
				const endpoints = layoutEngine!.computeEdgeEndpoints(fromPos, toPos);
				taskEdge.setEndpoints(endpoints.fromX, endpoints.fromY, endpoints.toX, endpoints.toY);
				taskEdge.setType(EdgeType.Normal);

				edgeContainer.addChild(taskEdge);
			}
		} else {
			// Task モード: dependencies からエッジを作成
			for (const gn of nodeList) {
				const toPos = positions.get(gn.id);
				if (!toPos) continue;

				for (const depId of gn.dependencies) {
					if (!nodeIds.has(depId)) continue;

					const fromPos = positions.get(depId);
					if (!fromPos) continue;

					const edge = edgeFactory.getOrCreate(depId, gn.id);
					const endpoints = layoutEngine!.computeEdgeEndpoints(fromPos, toPos);
					edge.setEndpoints(endpoints.fromX, endpoints.fromY, endpoints.toX, endpoints.toY);

					// エッジタイプを設定
					const depNode = nodeList.find((n) => n.id === depId);
					if (depNode) {
						if (depNode.status !== 'completed' && gn.status === 'blocked') {
							edge.setType(EdgeType.Blocked);
						} else {
							edge.setType(EdgeType.Normal);
						}
					}

					edgeContainer.addChild(edge);
				}
			}
		}

		// フィルター適用
		visibleTaskIds = new Set(filterManager.getVisibleIds());
		updateVisibility();

		// ビューを中央に
		if (nodeList.length > 0) {
			const centerX = (layout.bounds.minX + layout.bounds.maxX) / 2;
			const centerY = (layout.bounds.minY + layout.bounds.maxY) / 2;
			engine.panTo(centerX, centerY, false);
		}

		logMetrics('renderGraphNodes', renderStart, {
			seq: renderSeq,
			nodes: nodeList.length,
			dependencies: dependencyCount,
			renderedNodes: nodeMap.size,
			edges: edgeFactory.getAll().length,
			layoutMs: Math.round(layoutMs),
			boundsW: Math.round(layout.bounds.width),
			boundsH: Math.round(layout.bounds.height),
			layers: layout.layers.length,
			visibleFilterCount: visibleTaskIds.size,
			viewportScale: Number(currentViewport.scale.toFixed(2)),
			mode: isWBSMode ? 'wbs' : 'task'
		}, true);
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
				if (node) node.visible = true;
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
	 */
	function handleNodeClick(node: TaskNode, event?: FederatedPointerEvent): void {
		if (!selectionManager) return;

		const taskId = node.getTaskId();
		// FederatedPointerEvent から nativeEvent 経由でキー情報を取得
		const nativeEvent = event?.nativeEvent as PointerEvent | undefined;
		const isMulti = nativeEvent?.ctrlKey || nativeEvent?.metaKey || nativeEvent?.shiftKey;

		// Alt+クリック: 依存関係フィルター（グローバルキー状態を使用）
		if (isAltKeyPressed || nativeEvent?.altKey) {
			console.log('[NodeClick] Alt+click detected, triggering filter');
			handleNodeContextMenu(node, event!);
			return;
		}

		if (nativeEvent?.shiftKey && selectedIds.length > 0) {
			// Shift+クリック: 依存チェーン選択
			selectionManager.selectDependencyChain(taskId, 'both');
		} else {
			selectionManager.toggleSelect(taskId, isMulti);
		}
	}

	/**
	 * ノードホバー処理（デバウンス付き）
	 */
	async function handleNodeHover(node: TaskNode, isHovered: boolean): Promise<void> {
		const taskId = isHovered ? node.getTaskId() : null;
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

	/**
	 * ノード右クリック処理 - 依存関係フィルター
	 * フロントエンドの edges データを使用して依存関係を計算
	 */
	function handleNodeContextMenu(node: TaskNode, _event: FederatedPointerEvent | null): void {
		console.log('[ContextMenu] Right-click detected on node:', node.getNodeId());
		const nodeId = node.getNodeId();

		// 同じノードを再度右クリックしたらフィルターを解除
		if (dependencyFilterNodeId === nodeId) {
			clearDependencyFilter();
			return;
		}

		// フロントエンドで依存関係を計算
		const upstream = getUpstreamNodes(nodeId);
		const downstream = getDownstreamNodes(nodeId);

		console.log('[DependencyFilter] Node:', nodeId, 'Upstream:', upstream.length, 'Downstream:', downstream.length);

		// 表示対象: 選択ノード + 上流 + 下流
		const filterIds = new Set<string>([
			nodeId,
			...upstream,
			...downstream
		]);

		dependencyFilterNodeId = nodeId;
		dependencyFilterIds = filterIds;

		// visibleTaskIds を更新してフィルターを適用
		applyDependencyFilter();
	}

	/**
	 * 上流ノードを取得（このノードが依存しているノード）
	 * BFS で間接的な依存関係も含める
	 */
	function getUpstreamNodes(nodeId: string): string[] {
		const visited = new Set<string>();
		const queue = [nodeId];

		while (queue.length > 0) {
			const current = queue.shift()!;
			if (visited.has(current)) continue;
			visited.add(current);

			// このノードが依存しているエッジを探す（edge.to === current）
			for (const edge of edgeFactory.getAll()) {
				if (edge.getToId() === current && !visited.has(edge.getFromId())) {
					queue.push(edge.getFromId());
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

		while (queue.length > 0) {
			const current = queue.shift()!;
			if (visited.has(current)) continue;
			visited.add(current);

			// このノードに依存しているエッジを探す（edge.from === current）
			for (const edge of edgeFactory.getAll()) {
				if (edge.getFromId() === current && !visited.has(edge.getToId())) {
					queue.push(edge.getToId());
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
				node.x = pos.x - TaskNode.getWidth() / 2;
				node.y = pos.y - TaskNode.getHeight() / 2;
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
			if (fromPos && toPos && layoutEngine) {
				const endpoints = layoutEngine.computeEdgeEndpoints(fromPos, toPos);
				edge.setEndpoints(endpoints.fromX, endpoints.fromY, endpoints.toX, endpoints.toY);
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
		let minX = Infinity, maxX = -Infinity;
		let minY = Infinity, maxY = -Infinity;

		for (const pos of posMap.values()) {
			minX = Math.min(minX, pos.x - TaskNode.getWidth() / 2);
			maxX = Math.max(maxX, pos.x + TaskNode.getWidth() / 2);
			minY = Math.min(minY, pos.y - TaskNode.getHeight() / 2);
			maxY = Math.max(maxY, pos.y + TaskNode.getHeight() / 2);
		}

		if (posMap.size > 0) {
			spatialIndex.rebuild({
				x: minX - 500,
				y: minY - 500,
				width: (maxX - minX) + 1000,
				height: (maxY - minY) + 1000
			});

			// ノードを空間インデックスに追加
			for (const [nodeId, pos] of posMap) {
				spatialIndex.insert({
					id: nodeId,
					x: pos.x - TaskNode.getWidth() / 2,
					y: pos.y - TaskNode.getHeight() / 2,
					width: TaskNode.getWidth(),
					height: TaskNode.getHeight()
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
		const taskItems = graphNodesToTaskItems(graphNodes);
		const filteredLayout = layoutEngine.layoutSubset(taskItems, dependencyFilterIds);

		// 可視性を更新
		visibleTaskIds = dependencyFilterIds;
		updateVisibility();

		// 位置を更新
		updateNodePositions(filteredLayout.positions);
		layoutBounds = filteredLayout.bounds;

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
				const taskItems = graphNodesToTaskItems(graphNodes);
				const fullLayout = layoutEngine.layout(taskItems);
				layoutBounds = fullLayout.bounds;
			}

			originalPositions = null;
		}

		// ビューをリセット
		fitToView();
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
				const edge = edgeFactory.getAll().find(e => e.getKey() === edgeKey);
				if (edge) {
					const fromId = edge.getFromId();
					const toId = edge.getToId();
					if (showCriticalPath && criticalPathIds.has(fromId) && criticalPathIds.has(toId)) {
						edge.setType(EdgeType.Critical);
					} else {
						edge.setType(EdgeType.Normal);
					}
				}
			}
		}

		previousHighlightedEdges = currentHighlightedEdges;
	}

	/**
	 * クリティカルパス表示を切り替え
	 */
	function toggleCriticalPath(): void {
		showCriticalPath = !showCriticalPath;
		applyCriticalPathInfo();
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
	 * 特定タスクにフォーカス（将来機能用）
	 */
	function _focusTask(taskId: string): void {
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
		<button
			class="control-btn critical-path-toggle"
			class:active={showCriticalPath}
			onclick={toggleCriticalPath}
			title={showCriticalPath ? 'Hide Critical Path' : 'Show Critical Path'}
		>
			<span class="icon">⚡</span>
		</button>
		{#if dependencyFilterNodeId}
			<button
				class="control-btn filter-reset-btn"
				onclick={clearDependencyFilter}
				title="Reset Dependency Filter"
			>
				<span class="icon">✕</span>
			</button>
		{/if}
		{#if METRICS_ENABLED}
			<button class="control-btn metrics-btn" onclick={downloadMetrics} title="Download metrics log">
				<span class="icon">DL</span>
			</button>
		{/if}
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
		nodes={graphNodes}
		{isWBSMode}
		{positions}
		bounds={layoutBounds}
		viewport={currentViewport}
		onNavigate={handleMinimapNavigate}
	/>

	<!-- ステータスバー -->
	<div class="status-bar">
		<span class="status-item mode-indicator" class:wbs-mode={isWBSMode}>
			{isWBSMode ? 'WBS' : 'TASK'}
		</span>
		<span class="status-item">
			Zoom: {(currentViewport.scale * 100).toFixed(0)}%
		</span>
		<span class="status-item"> Nodes: {visibleCount}/{graphNodes.length} </span>
		{#if METRICS_ENABLED}
			<span class="status-item metrics-info"> Logs: {metricsEntries.length} </span>
		{/if}
		{#if selectedIds.length > 0}
			<span class="status-item selection-info"> Selected: {selectedIds.length} </span>
		{/if}
		{#if dependencyFilterNodeId}
			<span class="status-item dependency-filter-info">
				Filtered: {dependencyFilterNodeId}
				<button class="inline-reset-btn" onclick={clearDependencyFilter}>×</button>
			</span>
		{/if}
		{#if hoveredTaskId}
			<span class="status-item hover-info">
				{hoveredTaskId}
			</span>
		{/if}
	</div>

	<!-- 凡例 -->
	<div class="legend">
		{#if isWBSMode}
			<!-- WBS モード: ノードタイプ凡例 -->
			<div class="legend-title">NODE TYPE</div>
			<div class="legend-item">
				<span class="legend-dot vision"></span>
				<span>Vision</span>
			</div>
			<div class="legend-item">
				<span class="legend-dot objective"></span>
				<span>Objective</span>
			</div>
			<div class="legend-item">
				<span class="legend-dot deliverable"></span>
				<span>Deliverable</span>
			</div>
			<div class="legend-item">
				<span class="legend-dot task"></span>
				<span>Task</span>
			</div>
			<div class="legend-divider"></div>
		{/if}
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
		{#if showCriticalPath && criticalPathIds.size > 0}
			<div class="legend-divider"></div>
			<div class="legend-title">CRITICAL PATH</div>
			<div class="legend-item">
				<span class="legend-line critical"></span>
				<span>Critical</span>
			</div>
			<div class="legend-item">
				<span class="legend-badge slack-zero">CRIT</span>
				<span>No Slack</span>
			</div>
			<div class="legend-item">
				<span class="legend-badge slack-positive">+3d</span>
				<span>Slack Days</span>
			</div>
		{/if}
	</div>

	<!-- 操作ヒント -->
	<div class="hints">
		<div class="hint-item">Scroll: Zoom</div>
		<div class="hint-item">Shift+Drag: Pan</div>
		<div class="hint-item">Shift+Click: Chain Select</div>
		<div class="hint-item">Alt+Click / Right-Click: Filter Dependencies</div>
	</div>
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

	/* モードインジケーター */
	.mode-indicator {
		font-weight: 600;
		padding: 2px 8px;
		border-radius: 4px;
		background-color: rgba(136, 136, 136, 0.3);
		color: var(--text-secondary);
		letter-spacing: 0.05em;
	}

	.mode-indicator.wbs-mode {
		background-color: rgba(255, 215, 0, 0.2);
		color: #ffd700;
		border: 1px solid rgba(255, 215, 0, 0.4);
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

	/* WBS ノードタイプ別の色（TaskNode.ts と同期） */
	.legend-dot.vision {
		background-color: #ffd700;  /* ゴールド - 最上位の目標 */
	}

	.legend-dot.objective {
		background-color: #6699ff;  /* ブルー - 目標 */
	}

	.legend-dot.deliverable {
		background-color: #66cc99;  /* グリーン - 成果物 */
	}

	.legend-dot.task {
		background-color: #888888;  /* グレー - タスク */
	}

	.legend-divider {
		height: 1px;
		background-color: var(--border-dark);
		margin: var(--spacing-xs) 0;
	}

	.legend-line {
		width: 16px;
		height: 3px;
		border-radius: 1px;
	}

	.legend-line.critical {
		background-color: var(--accent-primary);
	}

	.legend-badge {
		font-size: 8px;
		padding: 1px 4px;
		border-radius: 2px;
		font-weight: 600;
	}

	.legend-badge.slack-zero {
		background-color: var(--accent-primary);
		color: var(--bg-primary);
	}

	.legend-badge.slack-positive {
		background-color: #2d5a2d;
		color: var(--text-primary);
	}

	/* クリティカルパストグル */
	.critical-path-toggle.active {
		background-color: var(--accent-primary);
		border-color: var(--accent-primary);
		color: var(--bg-primary);
	}

	.critical-path-toggle.active:hover {
		background-color: var(--accent-secondary);
	}

	/* フィルターリセットボタン */
	.filter-reset-btn {
		background-color: rgba(238, 68, 68, 0.3);
		border-color: #ee4444;
	}

	.filter-reset-btn:hover {
		background-color: rgba(238, 68, 68, 0.5);
	}

	/* インラインリセットボタン */
	.inline-reset-btn {
		background: none;
		border: none;
		color: var(--text-muted);
		cursor: pointer;
		margin-left: 4px;
		padding: 0 4px;
		font-size: 12px;
	}

	.inline-reset-btn:hover {
		color: var(--text-primary);
	}

	/* 依存関係フィルター状態表示 */
	.dependency-filter-info {
		color: var(--accent-primary);
		background-color: rgba(255, 149, 51, 0.2);
		padding: 2px 8px;
		border-radius: 4px;
	}

	.metrics-btn {
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.02em;
	}

	.metrics-info {
		color: var(--accent-primary);
	}

	/* 操作ヒント */
	.hints {
		position: absolute;
		bottom: 40px;
		left: var(--spacing-md);
		display: flex;
		gap: var(--spacing-md);
		font-size: 11px;
		color: var(--text-secondary);
	}

	.hint-item {
		background-color: rgba(0, 0, 0, 0.7);
		padding: 3px 8px;
		border-radius: var(--border-radius-sm);
	}
</style>
