// UseCaseEngine - UML ユースケース図の PixiJS エンジン
// アクター、ユースケース、システム境界、関係線のレイアウトと描画を管理
import { Application, Container, Graphics, FederatedPointerEvent } from 'pixi.js';
import type {
	UseCaseDiagramResponse,
	ActorItem,
	UseCaseItem,
	UseCaseActorRef,
	UseCaseRelation
} from '$lib/types/api';
import { ActorNode } from '../rendering/ActorNode';
import { UseCaseNode } from '../rendering/UseCaseNode';
import { SystemBoundary } from '../rendering/SystemBoundary';
import { RelationEdge, ActorUseCaseEdge } from '../rendering/RelationEdge';

// レイアウト定数
const ACTOR_MARGIN = 80;        // アクター間の垂直マージン
const USECASE_MARGIN_X = 40;    // ユースケース間の水平マージン
const USECASE_MARGIN_Y = 30;    // ユースケース間の垂直マージン
const BOUNDARY_PADDING = 60;    // 境界内のパディング
const ACTOR_BOUNDARY_GAP = 100; // アクターとシステム境界の間隔

// ビューポート設定
const MIN_SCALE = 0.3;
const MAX_SCALE = 2.5;
const ZOOM_SPEED = 0.001;

// 設定型
export interface UseCaseEngineConfig {
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
function getDefaultConfig(): UseCaseEngineConfig {
	return {
		backgroundColor: 0x1a1a1a,
		antialias: true,
		resolution: typeof window !== 'undefined' ? (window.devicePixelRatio || 1) : 1
	};
}

/**
 * UseCaseEngine - UML ユースケース図の PixiJS エンジン
 *
 * 責務:
 * - PixiJS Application の初期化/破棄
 * - アクター、ユースケース、境界、関係線の管理
 * - 自動レイアウト計算
 * - ズーム/パン操作
 */
export class UseCaseEngine {
	private app: Application | null = null;
	private worldContainer: Container | null = null;
	private gridContainer: Container | null = null;
	private boundaryContainer: Container | null = null;
	private edgeContainer: Container | null = null;
	private usecaseContainer: Container | null = null;
	private actorContainer: Container | null = null;

	private config: UseCaseEngineConfig;
	private viewport: Viewport = {
		x: 0,
		y: 0,
		width: 0,
		height: 0,
		scale: 1.0
	};

	// ノード管理
	private actorNodes: Map<string, ActorNode> = new Map();
	private usecaseNodes: Map<string, UseCaseNode> = new Map();
	private systemBoundary: SystemBoundary | null = null;
	private relationEdges: Map<string, RelationEdge> = new Map();
	private actorUsecaseEdges: Map<string, ActorUseCaseEdge> = new Map();

	// 位置データ
	private actorPositions: Map<string, { x: number; y: number }> = new Map();
	private usecasePositions: Map<string, { x: number; y: number }> = new Map();

	// 境界サイズ/位置（レイアウト計算用）
	private boundarySize = { width: 400, height: 300 };
	private boundaryPosition = { x: 0, y: 0 };

	// パン操作
	private isPanning = false;
	private lastPanPosition = { x: 0, y: 0 };

	// イベントリスナー（クリーンアップ用に保持）
	private wheelHandler: ((e: WheelEvent) => void) | null = null;

	// イベントコールバック
	private onActorClick?: (actor: ActorItem) => void;
	private onActorHover?: (actor: ActorItem | null, event?: MouseEvent) => void;
	private onUseCaseClick?: (usecase: UseCaseItem) => void;
	private onUseCaseHover?: (usecase: UseCaseItem | null, event?: MouseEvent) => void;
	private onViewportChange?: (viewport: Viewport) => void;

	// 選択状態
	private selectedActorId: string | null = null;
	private selectedUseCaseId: string | null = null;

	// フィルタモード（デフォルトで非表示、選択時に関連要素のみ表示）
	private filterModeEnabled = false;

	// データ（フィルタリング計算用に保持）
	private currentData: UseCaseDiagramResponse | null = null;

	constructor(config: Partial<UseCaseEngineConfig> = {}) {
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
		this.boundaryContainer = new Container();
		this.edgeContainer = new Container();
		this.usecaseContainer = new Container();
		this.actorContainer = new Container();

		// コンテナ追加順序 = 描画順序（後に追加されたものが上に表示）
		// UML図として正しい順序: グリッド → 境界 → エッジ → ユースケース → アクター
		// ノードがエッジの上に表示されることで、線がノード内部を貫通しない
		this.worldContainer.addChild(this.gridContainer);
		this.worldContainer.addChild(this.boundaryContainer);
		this.worldContainer.addChild(this.edgeContainer);
		this.worldContainer.addChild(this.usecaseContainer);
		this.worldContainer.addChild(this.actorContainer);
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

		// マウスホイールでズーム（クリーンアップ用にハンドラーを保持）
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
		if (e.button === 1 || e.button === 2 || e.shiftKey) {
			this.isPanning = true;
			this.lastPanPosition = { x: e.globalX, y: e.globalY };
		}
	}

	/**
	 * パン移動
	 */
	private handlePanMove(e: FederatedPointerEvent): void {
		if (!this.isPanning || !this.worldContainer) return;

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
	}

	/**
	 * グリッド描画
	 */
	private drawGrid(): void {
		if (!this.gridContainer || !this.app) return;

		this.gridContainer.removeChildren();

		const grid = new Graphics();
		const gridSize = this.calculateGridSize();
		const color = 0xff9533;
		const alpha = 0.03;

		const viewWidth = this.app.screen.width / this.viewport.scale;
		const viewHeight = this.app.screen.height / this.viewport.scale;
		const startX = Math.floor(this.viewport.x / gridSize) * gridSize - gridSize;
		const startY = Math.floor(this.viewport.y / gridSize) * gridSize - gridSize;
		const endX = startX + viewWidth + gridSize * 3;
		const endY = startY + viewHeight + gridSize * 3;

		for (let x = startX; x <= endX; x += gridSize) {
			grid.moveTo(x, startY);
			grid.lineTo(x, endY);
		}

		for (let y = startY; y <= endY; y += gridSize) {
			grid.moveTo(startX, y);
			grid.lineTo(endX, y);
		}

		grid.stroke({ width: 1 / this.viewport.scale, color, alpha });
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
	 * データを設定して描画
	 */
	setData(data: UseCaseDiagramResponse): void {
		this.clearAll();
		this.currentData = data;

		if (!data.actors.length && !data.usecases.length) return;

		// レイアウト計算
		this.calculateLayout(data);

		// システム境界を作成
		this.createSystemBoundary(data.boundary);

		// アクターを作成
		for (const actor of data.actors) {
			this.createActorNode(actor);
		}

		// ユースケースを作成
		for (const usecase of data.usecases) {
			this.createUseCaseNode(usecase);
		}

		// 関係線を作成
		this.createEdges(data.usecases);

		// フィルタモード有効時は非表示にする
		if (this.filterModeEnabled) {
			this.hideAll();
		}

		// ビューを中央に配置
		this.centerView();
	}

	/**
	 * レイアウトを計算
	 */
	private calculateLayout(data: UseCaseDiagramResponse): void {
		// アクター配置（左側に縦並び）
		const actorHeight = ActorNode.getHeight();
		let actorY = BOUNDARY_PADDING;

		for (const actor of data.actors) {
			this.actorPositions.set(actor.id, {
				x: BOUNDARY_PADDING,
				y: actorY
			});
			actorY += actorHeight + ACTOR_MARGIN;
		}

		// ユースケース配置（グリッド状）
		const usecaseWidth = UseCaseNode.getWidth();
		const usecaseHeight = UseCaseNode.getHeight();
		const maxCols = 3;
		let col = 0;
		let row = 0;

		const boundaryX = BOUNDARY_PADDING + ActorNode.getWidth() + ACTOR_BOUNDARY_GAP;

		for (const usecase of data.usecases) {
			this.usecasePositions.set(usecase.id, {
				x: boundaryX + BOUNDARY_PADDING + col * (usecaseWidth + USECASE_MARGIN_X),
				y: SystemBoundary.getTitleHeight() + BOUNDARY_PADDING + row * (usecaseHeight + USECASE_MARGIN_Y)
			});

			col++;
			if (col >= maxCols) {
				col = 0;
				row++;
			}
		}

		// 境界サイズを計算
		const usecaseCols = Math.min(data.usecases.length, maxCols);
		const usecaseRows = Math.ceil(data.usecases.length / maxCols) || 1;

		const boundaryWidth = Math.max(
			300,
			usecaseCols * (usecaseWidth + USECASE_MARGIN_X) + BOUNDARY_PADDING * 2
		);
		const boundaryHeight = Math.max(
			200,
			usecaseRows * (usecaseHeight + USECASE_MARGIN_Y) + SystemBoundary.getTitleHeight() + BOUNDARY_PADDING * 2
		);

		// 境界位置を保存
		this.systemBoundary = null;  // 後で作成
		this.boundarySize = { width: boundaryWidth, height: boundaryHeight };
		this.boundaryPosition = { x: boundaryX, y: 0 };
	}

	/**
	 * システム境界を作成
	 */
	private createSystemBoundary(name: string): void {
		if (!this.boundaryContainer) return;

		this.systemBoundary = new SystemBoundary(
			name,
			this.boundarySize.width,
			this.boundarySize.height
		);
		this.systemBoundary.x = this.boundaryPosition.x;
		this.systemBoundary.y = this.boundaryPosition.y;

		this.boundaryContainer.addChild(this.systemBoundary);
	}

	/**
	 * アクターノードを作成
	 */
	private createActorNode(actor: ActorItem): void {
		if (!this.actorContainer) return;

		const node = new ActorNode(actor);
		const position = this.actorPositions.get(actor.id);

		if (position) {
			node.x = position.x;
			node.y = position.y;
		}

		// イベント設定
		node.onClick(() => {
			this.selectActor(actor.id);
			this.onActorClick?.(actor);
		});

		node.onHover((_, isHovered, event) => {
			this.onActorHover?.(isHovered ? actor : null, event);
			this.highlightActorEdges(actor.id, isHovered);
		});

		this.actorNodes.set(actor.id, node);
		this.actorContainer.addChild(node);
	}

	/**
	 * ユースケースノードを作成
	 */
	private createUseCaseNode(usecase: UseCaseItem): void {
		if (!this.usecaseContainer) return;

		const node = new UseCaseNode(usecase);
		const position = this.usecasePositions.get(usecase.id);

		if (position) {
			node.x = position.x;
			node.y = position.y;
		}

		// イベント設定
		node.onClick(() => {
			this.selectUseCase(usecase.id);
			this.onUseCaseClick?.(usecase);
		});

		node.onHover((_, isHovered, event) => {
			this.onUseCaseHover?.(isHovered ? usecase : null, event);
			this.highlightUseCaseEdges(usecase.id, isHovered);
		});

		this.usecaseNodes.set(usecase.id, node);
		this.usecaseContainer.addChild(node);
	}

	/**
	 * エッジを作成
	 */
	private createEdges(usecases: UseCaseItem[]): void {
		if (!this.edgeContainer) return;

		for (const usecase of usecases) {
			// アクターとの関係線
			for (const actorRef of usecase.actors) {
				this.createActorUseCaseEdge(actorRef, usecase.id);
			}

			// ユースケース間の関係線
			for (const relation of usecase.relations) {
				this.createRelationEdge(usecase.id, relation);
			}
		}
	}

	/**
	 * アクター・ユースケース間のエッジを作成
	 */
	private createActorUseCaseEdge(actorRef: UseCaseActorRef, usecaseId: string): void {
		if (!this.edgeContainer) return;

		const actorNode = this.actorNodes.get(actorRef.actor_id);
		const usecaseNode = this.usecaseNodes.get(usecaseId);

		if (!actorNode || !usecaseNode) return;

		const isPrimary = actorRef.role === 'primary';
		const edge = new ActorUseCaseEdge(actorRef.actor_id, usecaseId, isPrimary);

		// エンドポイント設定
		const fromX = actorNode.x + ActorNode.getWidth() / 2;
		const fromY = actorNode.y + ActorNode.getHeight() / 2;
		const toX = usecaseNode.x;
		const toY = usecaseNode.y + UseCaseNode.getHeight() / 2;

		edge.setEndpoints(fromX, fromY, toX, toY);

		this.actorUsecaseEdges.set(edge.getKey(), edge);
		this.edgeContainer.addChild(edge);
	}

	/**
	 * ユースケース間の関係線を作成
	 */
	private createRelationEdge(fromId: string, relation: UseCaseRelation): void {
		if (!this.edgeContainer) return;

		const fromNode = this.usecaseNodes.get(fromId);
		const toNode = this.usecaseNodes.get(relation.target_id);

		if (!fromNode || !toNode) return;

		const edge = new RelationEdge(fromId, relation.target_id, relation.type, relation.condition);

		// エンドポイント設定
		const fromX = fromNode.x + UseCaseNode.getWidth() / 2;
		const fromY = fromNode.y + UseCaseNode.getHeight() / 2;
		const toX = toNode.x + UseCaseNode.getWidth() / 2;
		const toY = toNode.y + UseCaseNode.getHeight() / 2;

		edge.setEndpoints(fromX, fromY, toX, toY);

		this.relationEdges.set(edge.getKey(), edge);
		this.edgeContainer.addChild(edge);
	}

	/**
	 * アクターを選択
	 */
	selectActor(actorId: string): void {
		// 以前の選択を解除
		if (this.selectedActorId) {
			const prevNode = this.actorNodes.get(this.selectedActorId);
			prevNode?.setSelected(false);
		}
		if (this.selectedUseCaseId) {
			const prevNode = this.usecaseNodes.get(this.selectedUseCaseId);
			prevNode?.setSelected(false);
			this.selectedUseCaseId = null;
		}

		// 新しい選択
		this.selectedActorId = actorId;
		const node = this.actorNodes.get(actorId);
		node?.setSelected(true);

		// フィルタモード有効時は関連要素のみ表示
		if (this.filterModeEnabled) {
			this.showRelatedTo(actorId, null);
		}
	}

	/**
	 * ユースケースを選択
	 */
	selectUseCase(usecaseId: string): void {
		// 以前の選択を解除
		if (this.selectedUseCaseId) {
			const prevNode = this.usecaseNodes.get(this.selectedUseCaseId);
			prevNode?.setSelected(false);
		}
		if (this.selectedActorId) {
			const prevNode = this.actorNodes.get(this.selectedActorId);
			prevNode?.setSelected(false);
			this.selectedActorId = null;
		}

		// 新しい選択
		this.selectedUseCaseId = usecaseId;
		const node = this.usecaseNodes.get(usecaseId);
		node?.setSelected(true);

		// フィルタモード有効時は関連要素のみ表示
		if (this.filterModeEnabled) {
			this.showRelatedTo(null, usecaseId);
		}
	}

	/**
	 * 選択を解除
	 */
	clearSelection(): void {
		if (this.selectedActorId) {
			const node = this.actorNodes.get(this.selectedActorId);
			node?.setSelected(false);
			this.selectedActorId = null;
		}
		if (this.selectedUseCaseId) {
			const node = this.usecaseNodes.get(this.selectedUseCaseId);
			node?.setSelected(false);
			this.selectedUseCaseId = null;
		}

		// フィルタモード有効時は非表示に戻す
		if (this.filterModeEnabled) {
			this.hideAll();
		}
	}

	/**
	 * 選択の視覚状態のみ解除（図は維持）
	 * パネルを閉じる際に使用 - 図は表示したまま選択ハイライトのみ解除
	 */
	clearSelectionVisual(): void {
		if (this.selectedActorId) {
			const node = this.actorNodes.get(this.selectedActorId);
			node?.setSelected(false);
			this.selectedActorId = null;
		}
		if (this.selectedUseCaseId) {
			const node = this.usecaseNodes.get(this.selectedUseCaseId);
			node?.setSelected(false);
			this.selectedUseCaseId = null;
		}
		// hideAll() は呼ばない - 図は維持
	}

	/**
	 * アクターに関連するエッジをハイライト
	 */
	private highlightActorEdges(actorId: string, highlight: boolean): void {
		for (const edge of this.actorUsecaseEdges.values()) {
			if (edge.getActorId() === actorId) {
				edge.setHighlighted(highlight);
			}
		}
	}

	/**
	 * ユースケースに関連するエッジをハイライト
	 */
	private highlightUseCaseEdges(usecaseId: string, highlight: boolean): void {
		// アクター関連エッジ
		for (const edge of this.actorUsecaseEdges.values()) {
			if (edge.getUseCaseId() === usecaseId) {
				edge.setHighlighted(highlight);
			}
		}

		// ユースケース間エッジ
		for (const edge of this.relationEdges.values()) {
			if (edge.getFromId() === usecaseId || edge.getToId() === usecaseId) {
				edge.setHighlighted(highlight);
			}
		}
	}

	/**
	 * ビューを中央に配置
	 */
	centerView(): void {
		if (!this.worldContainer || !this.app) return;

		// 全体のバウンディングボックスを計算
		let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;

		for (const node of this.actorNodes.values()) {
			minX = Math.min(minX, node.x);
			minY = Math.min(minY, node.y);
			maxX = Math.max(maxX, node.x + ActorNode.getWidth());
			maxY = Math.max(maxY, node.y + ActorNode.getHeight());
		}

		if (this.systemBoundary) {
			minX = Math.min(minX, this.systemBoundary.x);
			minY = Math.min(minY, this.systemBoundary.y);
			maxX = Math.max(maxX, this.systemBoundary.x + this.systemBoundary.getBoundaryWidth());
			maxY = Math.max(maxY, this.systemBoundary.y + this.systemBoundary.getBoundaryHeight());
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
		for (const node of this.actorNodes.values()) {
			node.destroy();
		}
		this.actorNodes.clear();

		for (const node of this.usecaseNodes.values()) {
			node.destroy();
		}
		this.usecaseNodes.clear();

		// エッジを破棄
		for (const edge of this.actorUsecaseEdges.values()) {
			edge.destroy();
		}
		this.actorUsecaseEdges.clear();

		for (const edge of this.relationEdges.values()) {
			edge.destroy();
		}
		this.relationEdges.clear();

		// 境界を破棄
		if (this.systemBoundary) {
			this.systemBoundary.destroy();
			this.systemBoundary = null;
		}

		// 位置データをクリア
		this.actorPositions.clear();
		this.usecasePositions.clear();

		// 選択状態をクリア
		this.selectedActorId = null;
		this.selectedUseCaseId = null;

		// データをクリア（フィルタモードで使用）
		// Note: setData で新しいデータが設定されるので、ここではクリアしない
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
	 * リサイズ処理
	 */
	resize(): void {
		if (!this.app) return;

		this.viewport.width = this.app.screen.width;
		this.viewport.height = this.app.screen.height;
		this.drawGrid();
	}

	/**
	 * イベントリスナーを設定
	 */
	onActorClicked(callback: (actor: ActorItem) => void): void {
		this.onActorClick = callback;
	}

	onActorHovered(callback: (actor: ActorItem | null, event?: MouseEvent) => void): void {
		this.onActorHover = callback;
	}

	onUseCaseClicked(callback: (usecase: UseCaseItem) => void): void {
		this.onUseCaseClick = callback;
	}

	onUseCaseHovered(callback: (usecase: UseCaseItem | null, event?: MouseEvent) => void): void {
		this.onUseCaseHover = callback;
	}

	onViewportChanged(callback: (viewport: Viewport) => void): void {
		this.onViewportChange = callback;
	}

	/**
	 * フィルタモードを設定
	 * @param enabled true: デフォルト非表示、選択時に関連要素のみ表示
	 */
	setFilterMode(enabled: boolean): void {
		this.filterModeEnabled = enabled;
		if (enabled) {
			// 選択がなければすべて非表示
			if (!this.selectedActorId && !this.selectedUseCaseId) {
				this.hideAll();
			}
		} else {
			// フィルタモード無効時はすべて表示
			this.showAll();
		}
	}

	/**
	 * フィルタモードが有効かどうかを取得
	 */
	isFilterModeEnabled(): boolean {
		return this.filterModeEnabled;
	}

	/**
	 * すべての要素を非表示
	 */
	hideAll(): void {
		// アクターノード
		for (const node of this.actorNodes.values()) {
			node.visible = false;
		}

		// ユースケースノード
		for (const node of this.usecaseNodes.values()) {
			node.visible = false;
		}

		// エッジ
		for (const edge of this.actorUsecaseEdges.values()) {
			edge.visible = false;
		}
		for (const edge of this.relationEdges.values()) {
			edge.visible = false;
		}

		// システム境界も非表示
		if (this.systemBoundary) {
			this.systemBoundary.visible = false;
		}
	}

	/**
	 * すべての要素を表示
	 */
	showAll(): void {
		// アクターノード
		for (const node of this.actorNodes.values()) {
			node.visible = true;
		}

		// ユースケースノード
		for (const node of this.usecaseNodes.values()) {
			node.visible = true;
		}

		// エッジ
		for (const edge of this.actorUsecaseEdges.values()) {
			edge.visible = true;
		}
		for (const edge of this.relationEdges.values()) {
			edge.visible = true;
		}

		// システム境界
		if (this.systemBoundary) {
			this.systemBoundary.visible = true;
		}
	}

	/**
	 * 選択されたエンティティに関連する要素のみを表示
	 * @param actorId 選択されたアクターID（null ならアクター選択なし）
	 * @param usecaseId 選択されたユースケースID（null ならユースケース選択なし）
	 */
	showRelatedTo(actorId: string | null, usecaseId: string | null): void {
		if (!this.currentData) return;

		// まずすべて非表示
		this.hideAll();

		if (!actorId && !usecaseId) {
			// 選択なしの場合は非表示のまま
			return;
		}

		// 表示するIDを収集
		const visibleActorIds = new Set<string>();
		const visibleUseCaseIds = new Set<string>();

		if (actorId) {
			// アクターが選択された場合
			visibleActorIds.add(actorId);

			// このアクターに関連するユースケースを探す
			for (const usecase of this.currentData.usecases) {
				const isRelated = usecase.actors.some(ref => ref.actor_id === actorId);
				if (isRelated) {
					visibleUseCaseIds.add(usecase.id);
				}
			}
		}

		if (usecaseId) {
			// ユースケースが選択された場合
			visibleUseCaseIds.add(usecaseId);

			const usecase = this.currentData.usecases.find(u => u.id === usecaseId);
			if (usecase) {
				// 関連するアクターを追加
				for (const actorRef of usecase.actors) {
					visibleActorIds.add(actorRef.actor_id);
				}

				// 関連するユースケースを追加（include, extend, generalize）
				for (const relation of usecase.relations) {
					visibleUseCaseIds.add(relation.target_id);
				}

				// このユースケースを参照している他のユースケースも追加
				for (const otherUsecase of this.currentData.usecases) {
					const referencesThis = otherUsecase.relations.some(r => r.target_id === usecaseId);
					if (referencesThis) {
						visibleUseCaseIds.add(otherUsecase.id);
					}
				}
			}
		}

		// 表示を適用
		for (const [id, node] of this.actorNodes) {
			node.visible = visibleActorIds.has(id);
		}

		for (const [id, node] of this.usecaseNodes) {
			node.visible = visibleUseCaseIds.has(id);
		}

		// エッジ: 両端が表示されている場合のみ表示
		for (const edge of this.actorUsecaseEdges.values()) {
			const actorVisible = visibleActorIds.has(edge.getActorId());
			const usecaseVisible = visibleUseCaseIds.has(edge.getUseCaseId());
			edge.visible = actorVisible && usecaseVisible;
		}

		for (const edge of this.relationEdges.values()) {
			const fromVisible = visibleUseCaseIds.has(edge.getFromId());
			const toVisible = visibleUseCaseIds.has(edge.getToId());
			edge.visible = fromVisible && toVisible;
		}

		// システム境界: 表示される UseCase が1つでもあれば表示
		if (this.systemBoundary) {
			this.systemBoundary.visible = visibleUseCaseIds.size > 0;
		}
	}

	/**
	 * エンジンを破棄
	 */
	destroy(): void {
		this.clearAll();

		// wheel イベントリスナーを削除
		if (this.app && this.wheelHandler) {
			this.app.canvas.removeEventListener('wheel', this.wheelHandler);
			this.wheelHandler = null;
		}

		// コールバックをクリア
		this.onActorClick = undefined;
		this.onActorHover = undefined;
		this.onUseCaseClick = undefined;
		this.onUseCaseHover = undefined;
		this.onViewportChange = undefined;

		if (this.app) {
			this.app.destroy(true, { children: true, texture: true });
			this.app = null;
		}

		this.worldContainer = null;
		this.gridContainer = null;
		this.boundaryContainer = null;
		this.edgeContainer = null;
		this.usecaseContainer = null;
		this.actorContainer = null;
	}
}
