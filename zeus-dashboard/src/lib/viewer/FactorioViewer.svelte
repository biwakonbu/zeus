<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { TaskItem, TaskStatus, Priority, TimelineItem, TimelineResponse } from '$lib/types/api';
	import { fetchTimeline, fetchDownstream } from '$lib/api/client';
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
	let engineReady = $state(false); // エンジン初期化完了フラグ（$effect の依存関係用）
	let positions: Map<string, NodePosition> = $state(new Map());
	let layoutBounds = $state({ minX: 0, maxX: 0, minY: 0, maxY: 0, width: 0, height: 0 });
	let layers: string[][] = [];

	// 差分更新用のキャッシュ
	let previousTasksHash: string = '';
	let previousTaskIds: Set<string> = new Set();
	let previousDependencyHash: string = '';

	// Visibility/LOD 差分更新用
	let previousVisibleNodeIds: Set<string> = new Set();
	let previousLODLevel: LODLevel | null = null;

	/**
	 * タスクリストのハッシュを計算（浅い比較用）
	 */
	function computeTasksHash(taskList: TaskItem[]): string {
		return taskList.map(t =>
			`${t.id}:${t.status}:${t.progress ?? 0}:${t.priority ?? ''}:${t.assignee ?? ''}`
		).join('|');
	}

	/**
	 * 依存関係のハッシュを計算（構造変更の検出用）
	 */
	function computeDependencyHash(taskList: TaskItem[]): string {
		return taskList.map(t => `${t.id}:${t.dependencies.join(',')}`).sort().join('|');
	}

	/**
	 * タスクが変更されたかチェック
	 * @returns 変更タイプ: 'none' | 'data' | 'structure'
	 */
	function detectTaskChanges(newTasks: TaskItem[]): 'none' | 'data' | 'structure' {
		const newHash = computeTasksHash(newTasks);
		const newDepHash = computeDependencyHash(newTasks);
		const newIds = new Set(newTasks.map(t => t.id));

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

	// 選択中のID一覧
	let selectedIds: string[] = $state([]);

	// 矩形選択用
	let isRectSelecting = $state(false);
	let rectSelectStart: { x: number; y: number } | null = null;
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
						id: n.getTaskId(),
						name: n.getTask().title,
						x: Math.round(n.x),
						y: Math.round(n.y),
						status: n.getTask().status,
						progress: n.getTask().progress ?? 0
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
					taskCount: nodeMap.size,
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
					totalCount: tasks.length
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

		// リサイズ監視
		resizeObserver = new ResizeObserver(() => {
			engine?.resize();
		});
		resizeObserver.observe(containerElement);

		// エンジン準備完了を通知（$effect をトリガー）
		engineReady = true;
	}

	onDestroy(() => {
		window.removeEventListener('keydown', handleKeyDown);
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

	// タスクが変更されたら再レンダリング
	// NOTE: Svelte 5 の $effect は条件分岐内でのみ読み取られた変数を追跡しない
	// そのため tasks と engineReady を条件の外で明示的に読み取り、依存関係として登録する
	$effect(() => {
		const currentTasks = tasks; // 依存関係を明示的に登録
		const ready = engineReady; // エンジン初期化完了を依存関係に追加
		if (ready && engine && layoutEngine) {
			const changeType = detectTaskChanges(currentTasks);

			if (changeType === 'none') {
				// 変更なし - 何もしない
				return;
			}

			if (changeType === 'data') {
				// データのみ変更 - ノードの表示を更新
				updateTaskNodes(currentTasks);
			} else {
				// 構造変更 - フルレンダリング
				renderTasks(currentTasks);
			}

			// タイムラインデータの読み込みをデバウンス（500ms）
			if (timelineLoadTimer) {
				clearTimeout(timelineLoadTimer);
			}
			timelineLoadTimer = setTimeout(() => {
				loadTimelineData();
				timelineLoadTimer = null;
			}, 500);
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
	function updateTaskNodes(taskList: TaskItem[]): void {
		if (!filterManager || !selectionManager) return;

		const updateStart = nowMs();
		let updatedCount = 0;

		// タスクマップを作成
		const taskMap = new Map(taskList.map(t => [t.id, t]));

		// フィルターマネージャーを更新
		filterManager.setTasks(taskList);
		availableAssignees = filterManager.getAvailableAssignees();

		// 既存ノードのデータを更新
		for (const [id, node] of nodeMap) {
			const task = taskMap.get(id);
			if (task) {
				node.updateTask(task);
				updatedCount++;
			}
		}

		// フィルター適用
		visibleTaskIds = new Set(filterManager.getVisibleIds());
		updateVisibility();

		logMetrics('updateTaskNodes', updateStart, {
			totalNodes: nodeMap.size,
			updatedNodes: updatedCount
		}, true);
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
	 * タスクをレンダリング
	 */
	function renderTasks(taskList: TaskItem[]): void {
		if (!engine || !layoutEngine || !spatialIndex || !filterManager || !selectionManager) return;

		const renderStart = nowMs();
		const renderSeq = ++renderSequence;
		let dependencyCount = 0;
		for (const task of taskList) {
			dependencyCount += task.dependencies.length;
		}

		const nodeContainer = engine.getNodeContainer();
		const edgeContainer = engine.getEdgeContainer();
		if (!nodeContainer || !edgeContainer) return;

		// マネージャーにタスクを設定
		filterManager.setTasks(taskList);
		selectionManager.setTasks(taskList);
		availableAssignees = filterManager.getAvailableAssignees();

		// レイアウト計算
		const layoutStart = nowMs();
		const layout = layoutEngine.layout(taskList);
		const layoutMs = nowMs() - layoutStart;
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

		logMetrics('renderTasks', renderStart, {
			seq: renderSeq,
			tasks: taskList.length,
			dependencies: dependencyCount,
			nodes: nodeMap.size,
			edges: edgeFactory.getAll().length,
			layoutMs: Math.round(layoutMs),
			boundsW: Math.round(layout.bounds.width),
			boundsH: Math.round(layout.bounds.height),
			layers: layout.layers.length,
			visibleFilterCount: visibleTaskIds.size,
			viewportScale: Number(currentViewport.scale.toFixed(2))
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
				const passesFilter = visibleTaskIds.size === 0 || visibleTaskIds.has(id);
				node.visible = false; // ビューポート外
			}
		}

		// キャッシュを更新
		previousVisibleNodeIds = currentVisibleIds;

		// エッジの表示/非表示を更新（可視ノードに接続されているもののみ）
		for (const edge of edgeFactory.getAll()) {
			const fromVisible = currentVisibleIds.has(edge.getFromId());
			const toVisible = currentVisibleIds.has(edge.getToId());
			edge.visible = fromVisible || toVisible;
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
			hoverDebounceTimer = setTimeout(async () => {
				await highlightImpactedTasks(taskId);
				hoverDebounceTimer = null;
			}, 300);
		} else {
			clearImpactHighlight();
		}
	}

	/**
	 * 影響を受けるタスクをハイライト
	 */
	async function highlightImpactedTasks(taskId: string): Promise<void> {
		try {
			const response = await fetchDownstream(taskId);

			highlightedDownstream = new Set(response.downstream);
			highlightedUpstream = new Set(response.upstream);

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
		} catch (error) {
			console.debug('Failed to fetch downstream tasks:', error);
			clearImpactHighlight();
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
		<button
			class="control-btn critical-path-toggle"
			class:active={showCriticalPath}
			onclick={toggleCriticalPath}
			title={showCriticalPath ? 'Hide Critical Path' : 'Show Critical Path'}
		>
			<span class="icon">⚡</span>
		</button>
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
		{#if METRICS_ENABLED}
			<span class="status-item metrics-info"> Logs: {metricsEntries.length} </span>
		{/if}
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
