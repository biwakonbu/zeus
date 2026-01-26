// UseCaseNode - UML ユースケース図のユースケース（楕円）描画クラス
// 楕円形でユースケースを表現し、ステータスに応じた背景色とラベル表示を行う
import { Container, Graphics, Text, TextStyle, CanvasTextMetrics, FederatedPointerEvent } from 'pixi.js';
import type { UseCaseItem, UseCaseStatus } from '$lib/types/api';
import { TEXT_RESOLUTION, COMMON_COLORS, USECASE_SIZE, USECASE_STATUS_STYLES } from './constants';

// パディング（ID表示用）
const ID_AREA_HEIGHT = 14;

// 色定義（Factorio テーマ準拠）
const COLORS = {
	// 基本色（共通定数から取得）
	backgroundHover: COMMON_COLORS.backgroundHover,
	backgroundSelected: COMMON_COLORS.backgroundSelected,
	borderHover: COMMON_COLORS.borderHover,
	borderSelected: COMMON_COLORS.borderSelected,
	text: COMMON_COLORS.text,
	textMuted: COMMON_COLORS.textMuted
};

/**
 * UseCaseNode - UML ユースケースの視覚的表現
 *
 * 責務:
 * - 楕円形のユースケース描画
 * - テキスト量に応じたサイズ自動調整
 * - ステータスに応じた背景色でのスタイル変更
 * - インタラクション（クリック、ホバー）
 */
export class UseCaseNode extends Container {
	private usecase: UseCaseItem;
	private background: Graphics;
	private titleText: Text;
	private idText: Text;

	// 動的サイズ
	private ellipseWidth: number;
	private ellipseHeight: number;

	private isHovered = false;
	private isSelected = false;

	// イベントコールバック
	private onClickCallback?: (node: UseCaseNode, event?: FederatedPointerEvent) => void;
	private onHoverCallback?: (node: UseCaseNode, isHovered: boolean, event?: MouseEvent) => void;

	constructor(usecase: UseCaseItem) {
		super();

		this.usecase = usecase;

		// サイズ計算
		const size = this.calculateSize();
		this.ellipseWidth = size.width;
		this.ellipseHeight = size.height;

		// コンポーネント初期化
		this.background = new Graphics();
		this.titleText = new Text({
			text: '',
			style: {
				fontSize: 11,
				fill: COLORS.text,
				fontFamily: 'IBM Plex Mono, monospace',
				align: 'center',
				wordWrap: true,
				wordWrapWidth: this.ellipseWidth - USECASE_SIZE.paddingH * 2
			},
			resolution: TEXT_RESOLUTION
		});
		this.idText = new Text({
			text: '',
			style: {
				fontSize: 9,
				fill: COLORS.textMuted,
				fontFamily: 'IBM Plex Mono, monospace',
				align: 'center'
			},
			resolution: TEXT_RESOLUTION
		});

		this.addChild(this.background);
		this.addChild(this.titleText);
		this.addChild(this.idText);

		// インタラクション設定
		this.eventMode = 'static';
		this.cursor = 'pointer';

		this.on('pointerover', (e: FederatedPointerEvent) => this.handleHover(true, e));
		this.on('pointerout', () => this.handleHover(false));
		this.on('pointertap', (e: FederatedPointerEvent) => this.handleClick(e));

		// 初回描画
		this.draw();
	}

	/**
	 * テキスト量に応じたサイズを計算
	 */
	private calculateSize(): { width: number; height: number } {
		// PixiJS TextMetrics でテキスト幅を測定
		const style = new TextStyle({
			fontSize: 11,
			fontFamily: 'IBM Plex Mono, monospace'
		});
		const metrics = CanvasTextMetrics.measureText(this.usecase.title, style);

		// パディングを加算してサイズ決定
		const width = Math.min(
			USECASE_SIZE.maxWidth,
			Math.max(USECASE_SIZE.minWidth, metrics.width + USECASE_SIZE.paddingH * 2)
		);
		const height = Math.min(
			USECASE_SIZE.maxHeight,
			Math.max(USECASE_SIZE.minHeight, metrics.height + USECASE_SIZE.paddingV * 2 + ID_AREA_HEIGHT)
		);

		return { width, height };
	}

	/**
	 * ユースケースを描画
	 */
	draw(): void {
		this.drawBackground();
		this.drawTexts();
	}

	/**
	 * 背景（楕円）を描画
	 * ステータスに応じた背景色・ボーダー色で表現
	 */
	private drawBackground(): void {
		this.background.clear();

		const centerX = this.ellipseWidth / 2;
		const centerY = this.ellipseHeight / 2;

		// ステータススタイルを取得
		const statusStyle = USECASE_STATUS_STYLES[this.usecase.status] || USECASE_STATUS_STYLES.draft;

		let bgColor = statusStyle.background;
		let borderColor = statusStyle.border;
		let borderWidth = 2;
		let glowAlpha = statusStyle.glowAlpha;

		// 選択/ホバー時はオーバーライド
		if (this.isSelected) {
			bgColor = COLORS.backgroundSelected;
			borderColor = COLORS.borderSelected;
			borderWidth = 3;
			glowAlpha = 0.2;
		} else if (this.isHovered) {
			bgColor = COLORS.backgroundHover;
			borderColor = COLORS.borderHover;
			glowAlpha = 0.15;
		}

		// グロー効果
		if (glowAlpha > 0) {
			this.background.ellipse(centerX, centerY, this.ellipseWidth / 2 + 4, this.ellipseHeight / 2 + 4);
			this.background.fill({ color: borderColor, alpha: glowAlpha });
		}

		// メイン楕円
		this.background.ellipse(centerX, centerY, this.ellipseWidth / 2, this.ellipseHeight / 2);
		this.background.fill(bgColor);
		this.background.stroke({ width: borderWidth, color: borderColor });

		// 上部ハイライト（金属感）
		this.background.ellipse(centerX, centerY - 5, this.ellipseWidth / 2 - 15, this.ellipseHeight / 2 - 15);
		this.background.stroke({ width: 1, color: 0x666666, alpha: 0.3 });
	}

	/**
	 * テキストを描画
	 */
	private drawTexts(): void {
		const centerX = this.ellipseWidth / 2;

		// タイトル（中央）- 最大幅で切り詰め
		const maxWidth = this.ellipseWidth - USECASE_SIZE.paddingH;
		let title = this.usecase.title;

		// テキストが最大幅を超える場合は切り詰め
		this.titleText.text = title;
		if (this.titleText.width > maxWidth) {
			while (this.titleText.width > maxWidth && title.length > 3) {
				title = title.substring(0, title.length - 1);
				this.titleText.text = title + '..';
			}
		}

		this.titleText.x = centerX - this.titleText.width / 2;
		this.titleText.y = this.ellipseHeight / 2 - this.titleText.height / 2 - 3;

		// ID（下部）
		this.idText.text = this.usecase.id;
		this.idText.x = centerX - this.idText.width / 2;
		this.idText.y = this.ellipseHeight / 2 + 8;
	}

	/**
	 * ホバー処理
	 */
	private handleHover(isHovered: boolean, event?: FederatedPointerEvent): void {
		this.isHovered = isHovered;
		this.draw();
		const mouseEvent = event?.nativeEvent as MouseEvent | undefined;
		this.onHoverCallback?.(this, isHovered, mouseEvent);
	}

	/**
	 * クリック処理
	 */
	private handleClick(e: FederatedPointerEvent): void {
		e.stopPropagation();
		this.onClickCallback?.(this, e);
	}

	/**
	 * 選択状態を設定
	 */
	setSelected(selected: boolean): void {
		if (this.isSelected !== selected) {
			this.isSelected = selected;
			this.draw();
		}
	}

	/**
	 * ユースケースデータを更新
	 */
	updateUseCase(usecase: UseCaseItem): void {
		this.usecase = usecase;
		// サイズ再計算
		const size = this.calculateSize();
		this.ellipseWidth = size.width;
		this.ellipseHeight = size.height;
		this.draw();
	}

	/**
	 * ユースケースIDを取得
	 */
	getUseCaseId(): string {
		return this.usecase.id;
	}

	/**
	 * ユースケースデータを取得
	 */
	getUseCase(): UseCaseItem {
		return this.usecase;
	}

	/**
	 * インスタンスの幅を取得
	 */
	getWidth(): number {
		return this.ellipseWidth;
	}

	/**
	 * インスタンスの高さを取得
	 */
	getHeight(): number {
		return this.ellipseHeight;
	}

	/**
	 * 最大幅を取得（レイアウト計算用）
	 */
	static getMaxWidth(): number {
		return USECASE_SIZE.maxWidth;
	}

	/**
	 * 最大高さを取得（レイアウト計算用）
	 */
	static getMaxHeight(): number {
		return USECASE_SIZE.maxHeight;
	}

	/**
	 * イベントリスナーを設定
	 */
	onClick(callback: (node: UseCaseNode, event?: FederatedPointerEvent) => void): void {
		this.onClickCallback = callback;
	}

	onHover(callback: (node: UseCaseNode, isHovered: boolean, event?: MouseEvent) => void): void {
		this.onHoverCallback = callback;
	}
}
