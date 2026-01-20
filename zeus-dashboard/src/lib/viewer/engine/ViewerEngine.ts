// PixiJS ベースのビューワーエンジン
import { Application, Container, Graphics, FederatedPointerEvent } from 'pixi.js';

// 型定義
export interface ViewerConfig {
	backgroundColor: number;
	antialias: boolean;
	resolution: number;
}

export interface Viewport {
	x: number;
	y: number;
	width: number;
	height: number;
	scale: number;
}

export interface NodePosition {
	id: string;
	x: number;
	y: number;
}

// デフォルト設定（SSR 対応: window は実行時に参照）
const DEFAULT_CONFIG: ViewerConfig = {
	backgroundColor: 0x1a1a1a,
	antialias: true,
	resolution: 1 // 実行時に getDefaultConfig() で上書き
};

// クライアントサイドでのデフォルト設定を取得
function getDefaultConfig(): ViewerConfig {
	return {
		...DEFAULT_CONFIG,
		resolution: typeof window !== 'undefined' ? (window.devicePixelRatio || 1) : 1
	};
}

// ズーム設定
const MIN_SCALE = 0.1;
const MAX_SCALE = 3.0;
const ZOOM_SPEED = 0.001;

/**
 * ViewerEngine - PixiJS ベースのキャンバス管理
 *
 * 責務:
 * - PixiJS Application の初期化/破棄
 * - ビューポートの管理（パン/ズーム）
 * - コンテナ階層の管理
 */
export class ViewerEngine {
	private app: Application | null = null;
	private worldContainer: Container | null = null;
	private gridContainer: Container | null = null;
	private edgeContainer: Container | null = null;
	private nodeContainer: Container | null = null;

	private config: ViewerConfig;
	private viewport: Viewport = {
		x: 0,
		y: 0,
		width: 0,
		height: 0,
		scale: 1.0
	};

	// パン操作用の状態
	private isPanning = false;
	private lastPanPosition = { x: 0, y: 0 };

	// イベントコールバック
	private onViewportChange?: (viewport: Viewport) => void;
	private onNodeClick?: (nodeId: string) => void;
	private onNodeHover?: (nodeId: string | null) => void;

	constructor(config: Partial<ViewerConfig> = {}) {
		this.config = { ...getDefaultConfig(), ...config };
	}

	/**
	 * エンジンを初期化し、指定したコンテナにアタッチ
	 */
	async init(container: HTMLElement): Promise<void> {
		// PixiJS Application を作成
		this.app = new Application();

		await this.app.init({
			background: this.config.backgroundColor,
			antialias: this.config.antialias,
			resolution: this.config.resolution,
			autoDensity: true,
			resizeTo: container
		});

		// キャンバスをコンテナに追加
		container.appendChild(this.app.canvas as HTMLCanvasElement);

		// コンテナ階層を構築
		this.worldContainer = new Container();
		this.gridContainer = new Container();
		this.edgeContainer = new Container();
		this.nodeContainer = new Container();

		this.worldContainer.addChild(this.gridContainer);
		this.worldContainer.addChild(this.edgeContainer);
		this.worldContainer.addChild(this.nodeContainer);
		this.app.stage.addChild(this.worldContainer);

		// インタラクションを設定
		this.setupInteraction();

		// グリッド描画
		this.drawGrid();

		// ビューポート初期化
		this.viewport.width = container.clientWidth;
		this.viewport.height = container.clientHeight;
		this.centerView();
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
		this.app.canvas.addEventListener('wheel', (e: WheelEvent) => {
			e.preventDefault();
			this.handleZoom(e);
		}, { passive: false });

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
		if (!this.worldContainer) return;

		// マウス位置を取得
		const rect = this.app!.canvas.getBoundingClientRect();
		const mouseX = e.clientX - rect.left;
		const mouseY = e.clientY - rect.top;

		// 現在のワールド座標を計算
		const worldX = (mouseX - this.worldContainer.x) / this.viewport.scale;
		const worldY = (mouseY - this.worldContainer.y) / this.viewport.scale;

		// スケール変更
		const delta = -e.deltaY * ZOOM_SPEED;
		const newScale = Math.min(MAX_SCALE, Math.max(MIN_SCALE, this.viewport.scale * (1 + delta)));

		// スケールが変わった場合のみ更新
		if (newScale !== this.viewport.scale) {
			this.viewport.scale = newScale;
			this.worldContainer.scale.set(newScale);

			// マウス位置を中心にズーム
			this.worldContainer.x = mouseX - worldX * newScale;
			this.worldContainer.y = mouseY - worldY * newScale;

			this.viewport.x = -this.worldContainer.x / newScale;
			this.viewport.y = -this.worldContainer.y / newScale;

			// グリッド再描画（LOD対応）
			this.drawGrid();

			this.emitViewportChange();
		}
	}

	/**
	 * パン開始
	 */
	private handlePanStart(e: FederatedPointerEvent): void {
		// 右クリックまたは中クリックの場合のみパン
		// 左クリックはノード選択用に残す
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

		this.emitViewportChange();
	}

	/**
	 * パン終了
	 */
	private handlePanEnd(): void {
		this.isPanning = false;
	}

	/**
	 * グリッド描画（LOD対応）
	 */
	private drawGrid(): void {
		if (!this.gridContainer || !this.app) return;

		this.gridContainer.removeChildren();

		const grid = new Graphics();
		const gridSize = this.calculateGridSize();
		const color = 0xff9533;
		const alpha = 0.03;

		// 表示範囲を計算
		const viewWidth = this.app.screen.width / this.viewport.scale;
		const viewHeight = this.app.screen.height / this.viewport.scale;
		const startX = Math.floor(this.viewport.x / gridSize) * gridSize - gridSize;
		const startY = Math.floor(this.viewport.y / gridSize) * gridSize - gridSize;
		const endX = startX + viewWidth + gridSize * 3;
		const endY = startY + viewHeight + gridSize * 3;

		// 縦線
		for (let x = startX; x <= endX; x += gridSize) {
			grid.moveTo(x, startY);
			grid.lineTo(x, endY);
		}

		// 横線
		for (let y = startY; y <= endY; y += gridSize) {
			grid.moveTo(startX, y);
			grid.lineTo(endX, y);
		}

		grid.stroke({ width: 1 / this.viewport.scale, color, alpha });

		this.gridContainer.addChild(grid);
	}

	/**
	 * ズームレベルに応じたグリッドサイズを計算
	 */
	private calculateGridSize(): number {
		if (this.viewport.scale < 0.3) return 200;
		if (this.viewport.scale < 0.7) return 100;
		if (this.viewport.scale < 1.5) return 50;
		return 25;
	}

	/**
	 * ビューを中心に配置
	 */
	centerView(): void {
		if (!this.worldContainer || !this.app) return;

		this.worldContainer.x = this.app.screen.width / 2;
		this.worldContainer.y = this.app.screen.height / 2;

		this.viewport.x = -this.worldContainer.x / this.viewport.scale;
		this.viewport.y = -this.worldContainer.y / this.viewport.scale;

		this.emitViewportChange();
	}

	/**
	 * 特定座標にビューを移動
	 */
	panTo(x: number, y: number, animate = true): void {
		if (!this.worldContainer || !this.app) return;

		const targetX = this.app.screen.width / 2 - x * this.viewport.scale;
		const targetY = this.app.screen.height / 2 - y * this.viewport.scale;

		if (animate) {
			// TODO: アニメーション実装
			this.worldContainer.x = targetX;
			this.worldContainer.y = targetY;
		} else {
			this.worldContainer.x = targetX;
			this.worldContainer.y = targetY;
		}

		this.viewport.x = -this.worldContainer.x / this.viewport.scale;
		this.viewport.y = -this.worldContainer.y / this.viewport.scale;

		this.emitViewportChange();
	}

	/**
	 * ズームレベルを設定
	 */
	setZoom(scale: number, _animate = true): void {
		if (!this.worldContainer || !this.app) return;

		const newScale = Math.min(MAX_SCALE, Math.max(MIN_SCALE, scale));

		if (newScale !== this.viewport.scale) {
			// 画面中心を基準にズーム
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
			this.emitViewportChange();
		}
	}

	/**
	 * ノードコンテナを取得
	 */
	getNodeContainer(): Container | null {
		return this.nodeContainer;
	}

	/**
	 * エッジコンテナを取得
	 */
	getEdgeContainer(): Container | null {
		return this.edgeContainer;
	}

	/**
	 * PixiJS Application インスタンスを取得
	 * ヒットテスト等の低レベル操作に使用
	 */
	getApp(): Application | null {
		return this.app;
	}

	/**
	 * 現在のビューポートを取得
	 */
	getViewport(): Viewport {
		return { ...this.viewport };
	}

	/**
	 * イベントリスナーを設定
	 */
	onViewportChanged(callback: (viewport: Viewport) => void): void {
		this.onViewportChange = callback;
	}

	onNodeClicked(callback: (nodeId: string) => void): void {
		this.onNodeClick = callback;
	}

	onNodeHovered(callback: (nodeId: string | null) => void): void {
		this.onNodeHover = callback;
	}

	/**
	 * ビューポート変更イベントを発火
	 */
	private emitViewportChange(): void {
		this.onViewportChange?.(this.getViewport());
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
	 * ワールド座標でのビューポート矩形を取得（仮想化レンダリング用）
	 */
	getWorldViewport(margin = 100): { x: number; y: number; width: number; height: number } {
		if (!this.app) {
			return { x: 0, y: 0, width: 0, height: 0 };
		}

		const screenWidth = this.app.screen.width;
		const screenHeight = this.app.screen.height;

		return {
			x: this.viewport.x - margin / this.viewport.scale,
			y: this.viewport.y - margin / this.viewport.scale,
			width: (screenWidth + margin * 2) / this.viewport.scale,
			height: (screenHeight + margin * 2) / this.viewport.scale
		};
	}

	/**
	 * ワールド座標をスクリーン座標に変換
	 */
	worldToScreen(worldX: number, worldY: number): { x: number; y: number } {
		if (!this.worldContainer) {
			return { x: 0, y: 0 };
		}
		return {
			x: worldX * this.viewport.scale + this.worldContainer.x,
			y: worldY * this.viewport.scale + this.worldContainer.y
		};
	}

	/**
	 * スクリーン座標をワールド座標に変換
	 */
	screenToWorld(screenX: number, screenY: number): { x: number; y: number } {
		if (!this.worldContainer) {
			return { x: 0, y: 0 };
		}
		return {
			x: (screenX - this.worldContainer.x) / this.viewport.scale,
			y: (screenY - this.worldContainer.y) / this.viewport.scale
		};
	}

	/**
	 * ワールドコンテナを取得（矩形選択用）
	 */
	getWorldContainer(): Container | null {
		return this.worldContainer;
	}

	/**
	 * ステージを取得（イベント登録用）
	 */
	getStage(): Container | null {
		return this.app?.stage ?? null;
	}

	/**
	 * エンジンを破棄
	 */
	destroy(): void {
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
