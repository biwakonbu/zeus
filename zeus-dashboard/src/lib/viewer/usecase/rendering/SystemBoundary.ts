// SystemBoundary - UML ユースケース図のシステム境界（四角形）描画クラス
// ユースケースを包含するシステム境界を角丸四角形で表現
import { Container, Graphics, Text } from 'pixi.js';
import { TEXT_RESOLUTION, COMMON_COLORS } from './constants';

// サイズ定数
const PADDING = 40;
const TITLE_HEIGHT = 30;
const CORNER_RADIUS = 8;
const BORDER_WIDTH = 2;

// 色定義（Factorio テーマ準拠）
const COLORS = {
	background: COMMON_COLORS.background,
	border: COMMON_COLORS.border,
	borderHighlight: COMMON_COLORS.borderHighlight,
	titleBg: COMMON_COLORS.backgroundPanel,
	text: COMMON_COLORS.text,
	textMuted: COMMON_COLORS.textMuted
};

/**
 * SystemBoundary - UML システム境界の視覚的表現
 *
 * 責務:
 * - システム境界の四角形描画
 * - タイトルの表示
 * - 内部領域のサイズ管理
 */
export class SystemBoundary extends Container {
	private boundaryName: string;
	private graphics: Graphics;
	private titleText: Text;

	private boundaryWidth: number;
	private boundaryHeight: number;

	constructor(name: string, width: number = 400, height: number = 300) {
		super();

		this.boundaryName = name || 'System';
		this.boundaryWidth = width;
		this.boundaryHeight = height;

		// コンポーネント初期化
		this.graphics = new Graphics();
		this.titleText = new Text({
			text: '',
			style: {
				fontSize: 14,
				fill: COLORS.text,
				fontFamily: 'IBM Plex Mono, monospace',
				fontWeight: 'bold'
			},
			resolution: TEXT_RESOLUTION
		});

		this.addChild(this.graphics);
		this.addChild(this.titleText);

		// 初回描画
		this.draw();
	}

	/**
	 * システム境界を描画
	 */
	draw(): void {
		this.graphics.clear();

		const g = this.graphics;

		// 背景
		g.roundRect(0, 0, this.boundaryWidth, this.boundaryHeight, CORNER_RADIUS);
		g.fill({ color: COLORS.background, alpha: 0.5 });

		// 外枠（金属フレーム風）
		// 暗い外枠
		g.roundRect(0, 0, this.boundaryWidth, this.boundaryHeight, CORNER_RADIUS);
		g.stroke({ width: BORDER_WIDTH + 1, color: COLORS.border });

		// 明るい内枠
		g.roundRect(1, 1, this.boundaryWidth - 2, this.boundaryHeight - 2, CORNER_RADIUS - 1);
		g.stroke({ width: BORDER_WIDTH, color: COLORS.borderHighlight, alpha: 0.5 });

		// タイトル背景
		g.roundRect(0, 0, this.boundaryWidth, TITLE_HEIGHT, CORNER_RADIUS);
		g.fill(COLORS.titleBg);

		// タイトルと本体の境界線
		g.moveTo(0, TITLE_HEIGHT);
		g.lineTo(this.boundaryWidth, TITLE_HEIGHT);
		g.stroke({ width: 1, color: COLORS.border });

		// 上部ハイライト（金属感）
		g.moveTo(CORNER_RADIUS, 1);
		g.lineTo(this.boundaryWidth - CORNER_RADIUS, 1);
		g.stroke({ width: 1, color: COLORS.borderHighlight, alpha: 0.3 });

		// タイトルテキスト
		this.titleText.text = this.boundaryName;
		this.titleText.x = PADDING / 2;
		this.titleText.y = (TITLE_HEIGHT - this.titleText.height) / 2;
	}

	/**
	 * サイズを設定
	 */
	setSize(width: number, height: number): void {
		this.boundaryWidth = width;
		this.boundaryHeight = height;
		this.draw();
	}

	/**
	 * 名前を設定
	 */
	setName(name: string): void {
		this.boundaryName = name;
		this.draw();
	}

	/**
	 * 内部領域の左上座標を取得（ユースケース配置用）
	 */
	getContentOffset(): { x: number; y: number } {
		return {
			x: PADDING,
			y: TITLE_HEIGHT + PADDING
		};
	}

	/**
	 * 内部領域のサイズを取得
	 */
	getContentSize(): { width: number; height: number } {
		return {
			width: this.boundaryWidth - PADDING * 2,
			height: this.boundaryHeight - TITLE_HEIGHT - PADDING * 2
		};
	}

	/**
	 * 境界の幅を取得
	 */
	getBoundaryWidth(): number {
		return this.boundaryWidth;
	}

	/**
	 * 境界の高さを取得
	 */
	getBoundaryHeight(): number {
		return this.boundaryHeight;
	}

	/**
	 * パディング値を取得
	 */
	static getPadding(): number {
		return PADDING;
	}

	/**
	 * タイトル高さを取得
	 */
	static getTitleHeight(): number {
		return TITLE_HEIGHT;
	}
}
