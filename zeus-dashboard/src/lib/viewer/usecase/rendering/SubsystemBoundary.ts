// SubsystemBoundary - UML ユースケース図のサブシステム境界描画クラス
// TASK-020: サブシステムごとにユースケースをグループ化して表示
// SystemBoundary を参考に、カラー可変対応を追加
import { Container, Graphics, Text } from 'pixi.js';
import { TEXT_RESOLUTION, COMMON_COLORS } from './constants';
import type { SubsystemItem } from '$lib/types/api';
import { generateSubsystemColor, UNCATEGORIZED_SUBSYSTEM } from '../utils';

// サイズ定数
const CORNER_RADIUS = 12;
const BORDER_WIDTH = 2;
const TITLE_HEIGHT = 28;
const PADDING = 20;
const SHADOW_OFFSET = 4;
const SHADOW_ALPHA = 0.3;

/**
 * SubsystemBoundary - UML サブシステム境界の視覚的表現
 *
 * 責務:
 * - サブシステム境界の四角形描画（角丸矩形）
 * - サブシステムごとのカラー表現
 * - タイトルの表示（左上配置、UML 準拠）
 * - 内部のユースケースを囲む
 */
export class SubsystemBoundary extends Container {
	private subsystem: SubsystemItem;
	private graphics: Graphics;
	private titleText: Text;
	private color: number;

	private boundaryWidth: number;
	private boundaryHeight: number;

	constructor(subsystem: SubsystemItem, width: number = 300, height: number = 200) {
		super();

		this.subsystem = subsystem;
		this.boundaryWidth = width;
		this.boundaryHeight = height;

		// サブシステム ID からカラーを生成
		this.color = generateSubsystemColor(subsystem.id);

		// コンポーネント初期化
		this.graphics = new Graphics();
		this.titleText = new Text({
			text: '',
			style: {
				fontSize: 12,
				fill: COMMON_COLORS.text,
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
	 * サブシステム境界を描画
	 */
	draw(): void {
		this.graphics.clear();

		const g = this.graphics;

		// 影の描画
		g.roundRect(
			SHADOW_OFFSET,
			SHADOW_OFFSET,
			this.boundaryWidth,
			this.boundaryHeight,
			CORNER_RADIUS
		);
		g.fill({ color: 0x000000, alpha: SHADOW_ALPHA });

		// 背景（サブシステムカラーの半透明）
		g.roundRect(0, 0, this.boundaryWidth, this.boundaryHeight, CORNER_RADIUS);
		g.fill({ color: this.color, alpha: 0.15 });

		// 外枠（サブシステムカラー）
		g.roundRect(0, 0, this.boundaryWidth, this.boundaryHeight, CORNER_RADIUS);
		g.stroke({ width: BORDER_WIDTH, color: this.color, alpha: 0.6 });

		// タイトルバー背景
		// UML 準拠: 左上に配置
		const titleWidth = this.calculateTitleWidth();
		g.roundRect(0, 0, titleWidth + PADDING * 2, TITLE_HEIGHT, CORNER_RADIUS);
		g.fill({ color: this.color, alpha: 0.4 });

		// タイトルバー下境界
		g.moveTo(0, TITLE_HEIGHT);
		g.lineTo(titleWidth + PADDING * 2, TITLE_HEIGHT);
		g.stroke({ width: 1, color: this.color, alpha: 0.4 });

		// ステレオタイプマーカー（UML サブシステム表記）
		// 右上の小さなマーカー
		const markerSize = 8;
		g.moveTo(this.boundaryWidth - markerSize - 8, 8);
		g.lineTo(this.boundaryWidth - 8, 8);
		g.lineTo(this.boundaryWidth - 8, 8 + markerSize);
		g.stroke({ width: 1, color: this.color, alpha: 0.6 });

		// タイトルテキスト
		this.titleText.text = this.subsystem.name;
		this.titleText.x = PADDING;
		this.titleText.y = (TITLE_HEIGHT - this.titleText.height) / 2;
	}

	/**
	 * タイトルテキストの幅を計算
	 */
	private calculateTitleWidth(): number {
		// 仮のテキストで幅を測定
		const tempText = new Text({
			text: this.subsystem.name,
			style: {
				fontSize: 12,
				fontFamily: 'IBM Plex Mono, monospace',
				fontWeight: 'bold'
			},
			resolution: TEXT_RESOLUTION
		});
		const width = tempText.width;
		tempText.destroy();
		return Math.min(width, this.boundaryWidth - PADDING * 3);
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
	 * サブシステムを更新
	 */
	setSubsystem(subsystem: SubsystemItem): void {
		this.subsystem = subsystem;
		this.color = generateSubsystemColor(subsystem.id);
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
	 * サブシステム ID を取得
	 */
	getSubsystemId(): string {
		return this.subsystem.id;
	}

	/**
	 * サブシステムが「未分類」かどうか
	 */
	isUncategorized(): boolean {
		return this.subsystem.id === UNCATEGORIZED_SUBSYSTEM.id;
	}

	/**
	 * 色を取得
	 */
	getColor(): number {
		return this.color;
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

	/**
	 * 角丸半径を取得
	 */
	static getCornerRadius(): number {
		return CORNER_RADIUS;
	}

	/**
	 * ボーダー幅を取得
	 */
	static getBorderWidth(): number {
		return BORDER_WIDTH;
	}

	/**
	 * 影のオフセットを取得
	 */
	static getShadowOffset(): number {
		return SHADOW_OFFSET;
	}
}
