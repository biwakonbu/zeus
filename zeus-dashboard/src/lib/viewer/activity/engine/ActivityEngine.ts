// ActivityEngine - UML アクティビティ図の PixiJS エンジン
// ノード、遷移のレイアウトと描画を管理
import { Application, Container, Graphics } from 'pixi.js';
import type { FederatedPointerEvent } from 'pixi.js';
import type { ActivityItem, ActivityNodeItem, ActivityTransitionItem } from '$lib/types/api';
import { ActivityNodeBase } from '../rendering/ActivityNodeBase';
import { InitialNode } from '../rendering/InitialNode';
import { FinalNode } from '../rendering/FinalNode';
import { ActionNode } from '../rendering/ActionNode';
import { DecisionNode } from '../rendering/DecisionNode';
import { MergeNode } from '../rendering/MergeNode';
import { ForkNode } from '../rendering/ForkNode';
import { JoinNode } from '../rendering/JoinNode';
import { TransitionEdge } from '../rendering/TransitionEdge';
import { LAYOUT, COMMON_COLORS } from '../rendering/constants';

// ビューポート設定
const MIN_SCALE = 0.3;
const MAX_SCALE = 2.5;
const ZOOM_SPEED = 0.001;

// ドラッグ閾値（px）- これ以上動いたらドラッグとみなす
const DRAG_THRESHOLD = 5;

// アニメーション設定
const ANIMATION_DURATION = 250; // ms（応答性とスムーズさのバランス）

// イージング関数（easeOutCubic: 滑らかな減速）
function easeOutCubic(t: number): number {
	return 1 - Math.pow(1 - t, 3);
}

// 設定型
export interface ActivityEngineConfig {
	backgroundColor: number;
	antialias: boolean;
	resolution: number;
}

// ビューポート型
export interface Viewport {
	x: number;
	y: number;
	width: number;
	height: number;
	scale: number;
}

// デフォルト設定
function getDefaultConfig(): ActivityEngineConfig {
	return {
		backgroundColor: 0x1a1a1a,
		antialias: true,
		resolution: typeof window !== 'undefined' ? window.devicePixelRatio || 1 : 1
	};
}

/**
 * ActivityEngine - UML アクティビティ図の PixiJS エンジン
 *
 * 責務:
 * - PixiJS Application の初期化/破棄
 * - ノード（initial, final, action, decision, merge, fork, join）の管理
 * - 遷移エッジの管理
 * - Sugiyama 風の階層レイアウト
 * - ズーム/パン操作
 */
export class ActivityEngine {
	private app: Application | null = null;
	private worldContainer: Container | null = null;
	private gridContainer: Container | null = null;
	private edgeContainer: Container | null = null;
	private nodeContainer: Container | null = null;

	private config: ActivityEngineConfig;
	private viewport: Viewport = {
		x: 0,
		y: 0,
		width: 0,
		height: 0,
		scale: 1.0
	};

	// ノード管理
	private nodes: Map<string, ActivityNodeBase> = new Map();
	private edges: Map<string, TransitionEdge> = new Map();

	// 位置データ（レイアウト計算用）
	private nodePositions: Map<string, { x: number; y: number }> = new Map();

	// パン操作
	private isPanning = false;
	private potentialPan = false; // 左クリック開始後、閾値超えるまで
	private panStartPosition = { x: 0, y: 0 };
	private lastPanPosition = { x: 0, y: 0 };

	// イベントリスナー（クリーンアップ用に保持）
	private wheelHandler: ((e: WheelEvent) => void) | null = null;

	// アニメーション状態
	private animationState: {
		active: boolean;
		startTime: number;
		startX: number;
		startY: number;
		targetX: number;
		targetY: number;
	} | null = null;
	private tickerCallback: (() => void) | null = null;

	// イベントコールバック
	private onNodeClick?: (node: ActivityNodeItem) => void;
	private onNodeHover?: (node: ActivityNodeItem | null, event?: MouseEvent) => void;
	private onViewportChange?: (viewport: Viewport) => void;

	// 選択状態
	private selectedNodeId: string | null = null;

	// 現在のアクティビティデータ
	private currentActivity: ActivityItem | null = null;

	constructor(config: Partial<ActivityEngineConfig> = {}) {
		this.config = { ...getDefaultConfig(), ...config };
	}

	/**
	 * エンジンを初期化
	 */
	async init(container: HTMLElement): Promise<void> {
		this.app = new Application();

		await this.app.init({
			background: this.config.backgroundColor,
			antialias: this.config.antialias,
			resolution: this.config.resolution,
			autoDensity: true,
			resizeTo: container
		});

		container.appendChild(this.app.canvas as HTMLCanvasElement);

		// コンテナ階層を構築
		this.worldContainer = new Container();
		this.gridContainer = new Container();
		this.edgeContainer = new Container();
		this.nodeContainer = new Container();

		// コンテナ追加順序 = 描画順序（後に追加されたものが上に表示）
		this.worldContainer.addChild(this.gridContainer);
		this.worldContainer.addChild(this.edgeContainer);
		this.worldContainer.addChild(this.nodeContainer);
		this.app.stage.addChild(this.worldContainer);

		// インタラクション設定
		this.setupInteraction();

		// グリッド描画
		this.drawGrid();

		// ビューポート初期化
		this.viewport.width = container.clientWidth;
		this.viewport.height = container.clientHeight;
	}

	/**
	 * インタラクションを設定
	 */
	private setupInteraction(): void {
		if (!this.app) return;

		const stage = this.app.stage;
		stage.eventMode = 'static';
		stage.hitArea = this.app.screen;

		// マウスホイールでズーム
		this.wheelHandler = (e: WheelEvent) => {
			e.preventDefault();
			this.handleZoom(e);
		};
		this.app.canvas.addEventListener('wheel', this.wheelHandler, { passive: false });

		// パン操作
		stage.on('pointerdown', (e: FederatedPointerEvent) => this.handlePanStart(e));
		stage.on('pointermove', (e: FederatedPointerEvent) => this.handlePanMove(e));
		stage.on('pointerup', () => this.handlePanEnd());
		stage.on('pointerupoutside', () => this.handlePanEnd());
	}

	/**
	 * ズーム処理
	 */
	private handleZoom(e: WheelEvent): void {
		if (!this.worldContainer || !this.app) return;

		const rect = this.app.canvas.getBoundingClientRect();
		const mouseX = e.clientX - rect.left;
		const mouseY = e.clientY - rect.top;

		const worldX = (mouseX - this.worldContainer.x) / this.viewport.scale;
		const worldY = (mouseY - this.worldContainer.y) / this.viewport.scale;

		const delta = -e.deltaY * ZOOM_SPEED;
		const newScale = Math.min(MAX_SCALE, Math.max(MIN_SCALE, this.viewport.scale * (1 + delta)));

		if (newScale !== this.viewport.scale) {
			this.viewport.scale = newScale;
			this.worldContainer.scale.set(newScale);

			this.worldContainer.x = mouseX - worldX * newScale;
			this.worldContainer.y = mouseY - worldY * newScale;

			this.viewport.x = -this.worldContainer.x / newScale;
			this.viewport.y = -this.worldContainer.y / newScale;

			this.drawGrid();
			this.onViewportChange?.(this.getViewport());
		}
	}

	/**
	 * パン開始
	 */
	private handlePanStart(e: FederatedPointerEvent): void {
		// 左ボタン: 閾値超えてからパン開始（クリックと区別）
		// 中・右ボタン: 即座にパン開始
		if (e.button === 0) {
			this.potentialPan = true;
			this.panStartPosition = { x: e.globalX, y: e.globalY };
			this.lastPanPosition = { x: e.globalX, y: e.globalY };
		} else if (e.button === 1 || e.button === 2) {
			this.isPanning = true;
			this.lastPanPosition = { x: e.globalX, y: e.globalY };
		}
	}

	/**
	 * パン移動
	 */
	private handlePanMove(e: FederatedPointerEvent): void {
		if (!this.worldContainer) return;

		// 左クリック開始後、閾値をチェックしてパン開始判定
		if (this.potentialPan && !this.isPanning) {
			const dx = e.globalX - this.panStartPosition.x;
			const dy = e.globalY - this.panStartPosition.y;
			const distance = Math.sqrt(dx * dx + dy * dy);

			if (distance >= DRAG_THRESHOLD) {
				this.isPanning = true;
			}
		}

		if (!this.isPanning) return;

		const dx = e.globalX - this.lastPanPosition.x;
		const dy = e.globalY - this.lastPanPosition.y;

		this.worldContainer.x += dx;
		this.worldContainer.y += dy;

		this.viewport.x = -this.worldContainer.x / this.viewport.scale;
		this.viewport.y = -this.worldContainer.y / this.viewport.scale;

		this.lastPanPosition = { x: e.globalX, y: e.globalY };
		this.onViewportChange?.(this.getViewport());
	}

	/**
	 * パン終了
	 */
	private handlePanEnd(): void {
		this.isPanning = false;
		this.potentialPan = false;
	}

	/**
	 * グリッド描画（強化版：大グリッド + 小ドット）
	 */
	private drawGrid(): void {
		if (!this.gridContainer || !this.app) return;

		this.gridContainer.removeChildren();

		const grid = new Graphics();
		const gridSize = this.calculateGridSize();
		const color = COMMON_COLORS.accent;
		const alpha = 0.06; // 0.03 → 0.06 で少し見やすく

		const viewWidth = this.app.screen.width / this.viewport.scale;
		const viewHeight = this.app.screen.height / this.viewport.scale;
		const startX = Math.floor(this.viewport.x / gridSize) * gridSize - gridSize;
		const startY = Math.floor(this.viewport.y / gridSize) * gridSize - gridSize;
		const endX = startX + viewWidth + gridSize * 3;
		const endY = startY + viewHeight + gridSize * 3;

		// 大グリッド線
		for (let x = startX; x <= endX; x += gridSize) {
			grid.moveTo(x, startY);
			grid.lineTo(x, endY);
		}

		for (let y = startY; y <= endY; y += gridSize) {
			grid.moveTo(startX, y);
			grid.lineTo(endX, y);
		}

		grid.stroke({ width: 1 / this.viewport.scale, color, alpha });

		// 小ドット（細かいグリッド）
		const dotSize = gridSize / 4;
		const dotRadius = 0.5 / this.viewport.scale;

		// パフォーマンスのため、ズームアウト時はドットを描画しない
		if (this.viewport.scale >= 0.5) {
			for (let x = startX; x <= endX; x += dotSize) {
				for (let y = startY; y <= endY; y += dotSize) {
					// 大グリッドとの交点はスキップ
					if (x % gridSize !== 0 || y % gridSize !== 0) {
						grid.circle(x, y, dotRadius);
						grid.fill({ color, alpha: 0.04 });
					}
				}
			}
		}

		this.gridContainer.addChild(grid);
	}

	/**
	 * グリッドサイズ計算
	 */
	private calculateGridSize(): number {
		if (this.viewport.scale < 0.5) return 100;
		if (this.viewport.scale < 1.0) return 50;
		return 25;
	}

	/**
	 * アクティビティデータを設定して描画
	 */
	setData(activity: ActivityItem | null): void {
		this.clearAll();

		if (!activity || !activity.nodes.length) {
			this.currentActivity = null;
			return;
		}

		this.currentActivity = activity;

		// ノードを作成
		for (const node of activity.nodes) {
			this.createNode(node);
		}

		// Sugiyama 風の階層レイアウト計算
		this.calculateLayout(activity);

		// 位置を適用
		this.applyPositions();

		// 遷移エッジを作成（位置確定後）
		for (const transition of activity.transitions) {
			this.createEdge(transition);
		}

		// ビューを中央に配置
		this.centerView();
	}

	/**
	 * ノードを作成
	 */
	private createNode(nodeData: ActivityNodeItem): void {
		if (!this.nodeContainer) return;

		let node: ActivityNodeBase;

		switch (nodeData.type) {
			case 'initial':
				node = new InitialNode(nodeData);
				break;
			case 'final':
				node = new FinalNode(nodeData);
				break;
			case 'action':
				node = new ActionNode(nodeData);
				break;
			case 'decision':
				node = new DecisionNode(nodeData);
				break;
			case 'merge':
				node = new MergeNode(nodeData);
				break;
			case 'fork':
				node = new ForkNode(nodeData);
				break;
			case 'join':
				node = new JoinNode(nodeData);
				break;
			default:
				// 未知のタイプはアクションとして扱う
				node = new ActionNode(nodeData);
		}

		// イベント設定
		node.onClick(() => {
			this.selectNode(nodeData.id);
			this.onNodeClick?.(nodeData);
		});

		node.onHover((_, isHovered, event) => {
			this.onNodeHover?.(isHovered ? nodeData : null, event);
		});

		this.nodes.set(nodeData.id, node);
		this.nodeContainer.addChild(node);
	}

	/**
	 * Sugiyama 風の階層レイアウト計算
	 * initial → action/decision → final の流れで上から下に配置
	 */
	private calculateLayout(activity: ActivityItem): void {
		// 各ノードの深さ（レベル）を計算
		const levels = this.calculateNodeLevels(activity);

		// レベルごとにノードをグループ化
		const levelGroups: Map<number, string[]> = new Map();
		for (const [nodeId, level] of levels) {
			if (!levelGroups.has(level)) {
				levelGroups.set(level, []);
			}
			levelGroups.get(level)!.push(nodeId);
		}

		// 各レベルのノードを配置
		const sortedLevels = Array.from(levelGroups.keys()).sort((a, b) => a - b);

		// 動的な全体幅計算
		let maxLevelWidth = 0;
		for (const nodeIds of levelGroups.values()) {
			let levelWidth = 0;
			for (const nodeId of nodeIds) {
				const node = this.nodes.get(nodeId);
				if (node) {
					levelWidth += node.getNodeWidth();
				}
			}
			levelWidth += (nodeIds.length - 1) * LAYOUT.horizontalGap;
			maxLevelWidth = Math.max(maxLevelWidth, levelWidth);
		}
		const totalWidth = Math.max(LAYOUT.minTotalWidth, maxLevelWidth + LAYOUT.marginLeft * 2);

		let currentY = LAYOUT.marginTop;

		for (const level of sortedLevels) {
			const nodeIds = levelGroups.get(level)!;

			// このレベルの最大高さと幅を計算
			let maxHeight = 0;
			let levelWidth = 0;

			for (const nodeId of nodeIds) {
				const node = this.nodes.get(nodeId);
				if (node) {
					maxHeight = Math.max(maxHeight, node.getNodeHeight());
					levelWidth += node.getNodeWidth();
				}
			}

			levelWidth += (nodeIds.length - 1) * LAYOUT.horizontalGap;

			// 水平方向に中央揃えで配置（動的な全体幅に基づく）
			let currentX = LAYOUT.marginLeft + (totalWidth - LAYOUT.marginLeft * 2 - levelWidth) / 2;
			if (currentX < LAYOUT.marginLeft) currentX = LAYOUT.marginLeft;

			for (const nodeId of nodeIds) {
				const node = this.nodes.get(nodeId);
				if (node) {
					this.nodePositions.set(nodeId, {
						x: currentX,
						y: currentY + (maxHeight - node.getNodeHeight()) / 2
					});
					currentX += node.getNodeWidth() + LAYOUT.horizontalGap;
				}
			}

			currentY += maxHeight + LAYOUT.verticalGap;
		}
	}

	/**
	 * トポロジカルソートでノードの深さを計算
	 */
	private calculateNodeLevels(activity: ActivityItem): Map<string, number> {
		const levels: Map<string, number> = new Map();
		const inDegree: Map<string, number> = new Map();
		const outEdges: Map<string, string[]> = new Map();

		// 初期化
		for (const node of activity.nodes) {
			inDegree.set(node.id, 0);
			outEdges.set(node.id, []);
		}

		// エッジの情報を収集
		for (const transition of activity.transitions) {
			const current = inDegree.get(transition.target) || 0;
			inDegree.set(transition.target, current + 1);

			const edges = outEdges.get(transition.source) || [];
			edges.push(transition.target);
			outEdges.set(transition.source, edges);
		}

		// BFS でレベルを計算
		const queue: string[] = [];

		// 入次数 0 のノード（初期ノード）から開始
		for (const node of activity.nodes) {
			if ((inDegree.get(node.id) || 0) === 0) {
				queue.push(node.id);
				levels.set(node.id, 0);
			}
		}

		while (queue.length > 0) {
			const nodeId = queue.shift()!;
			const currentLevel = levels.get(nodeId) || 0;

			const targets = outEdges.get(nodeId) || [];
			for (const targetId of targets) {
				// ターゲットのレベルは、現在のレベル + 1 以上
				const existingLevel = levels.get(targetId);
				const newLevel = currentLevel + 1;

				if (existingLevel === undefined || existingLevel < newLevel) {
					levels.set(targetId, newLevel);
				}

				// 入次数を減らす
				const degree = (inDegree.get(targetId) || 1) - 1;
				inDegree.set(targetId, degree);

				if (degree === 0) {
					queue.push(targetId);
				}
			}
		}

		// レベルが設定されていないノード（孤立ノード）は 0 に設定
		for (const node of activity.nodes) {
			if (!levels.has(node.id)) {
				levels.set(node.id, 0);
			}
		}

		return levels;
	}

	/**
	 * 計算した位置をノードに適用
	 */
	private applyPositions(): void {
		for (const [id, pos] of this.nodePositions) {
			const node = this.nodes.get(id);
			if (node) {
				node.x = pos.x;
				node.y = pos.y;
			}
		}
	}

	/**
	 * 遷移エッジを作成
	 */
	private createEdge(transition: ActivityTransitionItem): void {
		if (!this.edgeContainer) return;

		const sourceNode = this.nodes.get(transition.source);
		const targetNode = this.nodes.get(transition.target);

		if (!sourceNode || !targetNode) return;

		// 接続ポイントを計算（ソースの下端からターゲットの上端へ）
		const sourcePoint = {
			x: sourceNode.getCenterX(),
			y: sourceNode.y + sourceNode.getNodeHeight()
		};

		const targetPoint = {
			x: targetNode.getCenterX(),
			y: targetNode.y
		};

		const edge = new TransitionEdge(transition, sourcePoint, targetPoint);
		this.edges.set(transition.id, edge);
		this.edgeContainer.addChild(edge);
	}

	/**
	 * ノードを選択
	 */
	selectNode(nodeId: string): void {
		// 以前の選択を解除
		if (this.selectedNodeId) {
			const prevNode = this.nodes.get(this.selectedNodeId);
			prevNode?.setSelected(false);
		}

		// 新しい選択
		this.selectedNodeId = nodeId;
		const node = this.nodes.get(nodeId);
		node?.setSelected(true);
	}

	/**
	 * 選択を解除
	 */
	clearSelection(): void {
		if (this.selectedNodeId) {
			const node = this.nodes.get(this.selectedNodeId);
			node?.setSelected(false);
			this.selectedNodeId = null;
		}
	}

	/**
	 * ビューを中央に配置
	 */
	centerView(): void {
		if (!this.worldContainer || !this.app) return;

		// 全体のバウンディングボックスを計算
		let minX = Infinity,
			minY = Infinity,
			maxX = -Infinity,
			maxY = -Infinity;

		for (const node of this.nodes.values()) {
			minX = Math.min(minX, node.x);
			minY = Math.min(minY, node.y);
			maxX = Math.max(maxX, node.x + node.getNodeWidth());
			maxY = Math.max(maxY, node.y + node.getNodeHeight());
		}

		if (!isFinite(minX)) {
			// ノードがない場合
			this.worldContainer.x = this.app.screen.width / 2;
			this.worldContainer.y = this.app.screen.height / 2;
			return;
		}

		const contentWidth = maxX - minX;
		const contentHeight = maxY - minY;
		const centerX = minX + contentWidth / 2;
		const centerY = minY + contentHeight / 2;

		// スケールを調整して全体が収まるようにする
		const scaleX = (this.app.screen.width * 0.8) / contentWidth;
		const scaleY = (this.app.screen.height * 0.8) / contentHeight;
		const newScale = Math.min(1.0, Math.min(scaleX, scaleY));

		this.viewport.scale = newScale;
		this.worldContainer.scale.set(newScale);

		this.worldContainer.x = this.app.screen.width / 2 - centerX * newScale;
		this.worldContainer.y = this.app.screen.height / 2 - centerY * newScale;

		this.viewport.x = -this.worldContainer.x / newScale;
		this.viewport.y = -this.worldContainer.y / newScale;

		this.drawGrid();
		this.onViewportChange?.(this.getViewport());
	}

	/**
	 * すべてクリア
	 */
	private clearAll(): void {
		// ノードを破棄
		for (const node of this.nodes.values()) {
			node.destroy();
		}
		this.nodes.clear();

		// エッジを破棄
		for (const edge of this.edges.values()) {
			edge.destroy();
		}
		this.edges.clear();

		// 位置データをクリア
		this.nodePositions.clear();

		// 選択状態をクリア
		this.selectedNodeId = null;
	}

	/**
	 * ビューポートを取得
	 */
	getViewport(): Viewport {
		return { ...this.viewport };
	}

	/**
	 * ズームを設定
	 */
	setZoom(scale: number): void {
		if (!this.worldContainer || !this.app) return;

		const newScale = Math.min(MAX_SCALE, Math.max(MIN_SCALE, scale));

		if (newScale !== this.viewport.scale) {
			const centerX = this.app.screen.width / 2;
			const centerY = this.app.screen.height / 2;
			const worldX = (centerX - this.worldContainer.x) / this.viewport.scale;
			const worldY = (centerY - this.worldContainer.y) / this.viewport.scale;

			this.viewport.scale = newScale;
			this.worldContainer.scale.set(newScale);

			this.worldContainer.x = centerX - worldX * newScale;
			this.worldContainer.y = centerY - worldY * newScale;

			this.viewport.x = -this.worldContainer.x / newScale;
			this.viewport.y = -this.worldContainer.y / newScale;

			this.drawGrid();
			this.onViewportChange?.(this.getViewport());
		}
	}

	/**
	 * ズームイン
	 */
	zoomIn(): void {
		this.setZoom(this.viewport.scale * 1.2);
	}

	/**
	 * ズームアウト
	 */
	zoomOut(): void {
		this.setZoom(this.viewport.scale / 1.2);
	}

	/**
	 * ズームリセット
	 */
	resetZoom(): void {
		this.centerView();
	}

	/**
	 * 特定座標にビューを移動
	 * @param x ワールド座標 X
	 * @param y ワールド座標 Y
	 * @param animate アニメーション有効化（デフォルト: false）
	 */
	panTo(x: number, y: number, animate = false): void {
		if (!this.worldContainer || !this.app) return;

		const targetX = this.app.screen.width / 2 - x * this.viewport.scale;
		const targetY = this.app.screen.height / 2 - y * this.viewport.scale;

		if (animate) {
			this.startPanAnimation(targetX, targetY);
		} else {
			this.worldContainer.x = targetX;
			this.worldContainer.y = targetY;

			this.viewport.x = -this.worldContainer.x / this.viewport.scale;
			this.viewport.y = -this.worldContainer.y / this.viewport.scale;

			this.drawGrid();
			this.onViewportChange?.(this.getViewport());
		}
	}

	/**
	 * パンアニメーションを開始
	 */
	private startPanAnimation(targetX: number, targetY: number): void {
		if (!this.worldContainer || !this.app) return;

		// 既存のアニメーションを停止
		this.stopPanAnimation();

		const startX = this.worldContainer.x;
		const startY = this.worldContainer.y;

		// 移動距離が小さい場合はアニメーションをスキップ
		const distance = Math.sqrt(
			Math.pow(targetX - startX, 2) + Math.pow(targetY - startY, 2)
		);
		if (distance < 1) return;

		this.animationState = {
			active: true,
			startTime: performance.now(),
			startX,
			startY,
			targetX,
			targetY
		};

		// Ticker でアニメーション更新
		this.tickerCallback = () => this.updatePanAnimation();
		this.app.ticker.add(this.tickerCallback);
	}

	/**
	 * パンアニメーションを更新
	 * パフォーマンス最適化: アニメーション中はグリッド再描画をスキップし、完了時のみ更新
	 */
	private updatePanAnimation(): void {
		if (!this.animationState || !this.worldContainer) {
			this.stopPanAnimation();
			return;
		}

		const elapsed = performance.now() - this.animationState.startTime;
		const progress = Math.min(1, elapsed / ANIMATION_DURATION);
		const easedProgress = easeOutCubic(progress);

		const { startX, startY, targetX, targetY } = this.animationState;
		this.worldContainer.x = startX + (targetX - startX) * easedProgress;
		this.worldContainer.y = startY + (targetY - startY) * easedProgress;

		this.viewport.x = -this.worldContainer.x / this.viewport.scale;
		this.viewport.y = -this.worldContainer.y / this.viewport.scale;

		// ビューポート変更を通知（グリッドはスキップ）
		this.onViewportChange?.(this.getViewport());

		// アニメーション完了時のみグリッドを更新
		if (progress >= 1) {
			this.drawGrid();
			this.stopPanAnimation();
		}
	}

	/**
	 * パンアニメーションを停止
	 */
	private stopPanAnimation(): void {
		if (this.tickerCallback && this.app) {
			this.app.ticker.remove(this.tickerCallback);
			this.tickerCallback = null;
		}
		this.animationState = null;
	}

	/**
	 * 特定ノードにビューをフォーカス（スムーススクロール）
	 * @param nodeId フォーカスするノードの ID
	 * @param animate アニメーション有効化（デフォルト: true）
	 */
	focusNode(nodeId: string, animate = true): void {
		const node = this.nodes.get(nodeId);
		if (!node) return;

		// ノードの中心座標を計算
		const centerX = node.getCenterX();
		const centerY = node.y + node.getNodeHeight() / 2;

		this.panTo(centerX, centerY, animate);
	}

	/**
	 * リサイズ処理
	 */
	resize(): void {
		if (!this.app) return;

		this.viewport.width = this.app.screen.width;
		this.viewport.height = this.app.screen.height;
		this.drawGrid();
	}

	/**
	 * 現在のアクティビティを取得
	 */
	getActivity(): ActivityItem | null {
		return this.currentActivity;
	}

	/**
	 * イベントリスナーを設定
	 */
	onNodeClicked(callback: (node: ActivityNodeItem) => void): void {
		this.onNodeClick = callback;
	}

	onNodeHovered(callback: (node: ActivityNodeItem | null, event?: MouseEvent) => void): void {
		this.onNodeHover = callback;
	}

	onViewportChanged(callback: (viewport: Viewport) => void): void {
		this.onViewportChange = callback;
	}

	/**
	 * エンジンを破棄
	 */
	destroy(): void {
		this.clearAll();

		// アニメーションを停止
		this.stopPanAnimation();

		// データをクリア
		this.currentActivity = null;

		// wheel イベントリスナーを削除
		if (this.app && this.wheelHandler) {
			this.app.canvas.removeEventListener('wheel', this.wheelHandler);
			this.wheelHandler = null;
		}

		// コールバックをクリア
		this.onNodeClick = undefined;
		this.onNodeHover = undefined;
		this.onViewportChange = undefined;

		if (this.app) {
			this.app.destroy(true, { children: true, texture: true });
			this.app = null;
		}

		this.worldContainer = null;
		this.gridContainer = null;
		this.edgeContainer = null;
		this.nodeContainer = null;
	}
}
