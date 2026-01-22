// UseCaseNode - UML ユースケース図のユースケース（楕円）描画クラス
// 楕円形でユースケースを表現し、ステータスに応じた色分けとラベル表示を行う
import { Container, Graphics, Text, FederatedPointerEvent } from 'pixi.js';
import type { UseCaseItem, UseCaseStatus } from '$lib/types/api';
import { TEXT_RESOLUTION, COMMON_COLORS } from './constants';

// サイズ定数
const ELLIPSE_WIDTH = 140;
const ELLIPSE_HEIGHT = 50;
const PADDING = 10;

// ステータスインジケーター定数
const STATUS_INDICATOR_X = 10;
const STATUS_INDICATOR_RADIUS = 5;

// 色定義（Factorio テーマ準拠）
const COLORS = {
	// ステータス別の色
	status: {
		active: 0x44cc44,      // 緑（アクティブ）
		draft: 0xffcc00,       // 黄（下書き）
		deprecated: 0x888888   // グレー（非推奨）
	} as Record<UseCaseStatus, number>,
	// 基本色（共通定数から取得）
	background: COMMON_COLORS.backgroundPanel,
	backgroundHover: COMMON_COLORS.backgroundHover,
	backgroundSelected: COMMON_COLORS.backgroundSelected,
	border: COMMON_COLORS.borderHighlight,
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
 * - ステータスに応じたスタイル変更
 * - インタラクション（クリック、ホバー）
 */
export class UseCaseNode extends Container {
	private usecase: UseCaseItem;
	private background: Graphics;
	private statusIndicator: Graphics;
	private titleText: Text;
	private idText: Text;

	private isHovered = false;
	private isSelected = false;

	// イベントコールバック
	private onClickCallback?: (node: UseCaseNode, event?: FederatedPointerEvent) => void;
	private onHoverCallback?: (node: UseCaseNode, isHovered: boolean, event?: MouseEvent) => void;

	constructor(usecase: UseCaseItem) {
		super();

		this.usecase = usecase;

		// コンポーネント初期化
		this.background = new Graphics();
		this.statusIndicator = new Graphics();
		this.titleText = new Text({
			text: '',
			style: {
				fontSize: 11,
				fill: COLORS.text,
				fontFamily: 'IBM Plex Mono, monospace',
				align: 'center',
				wordWrap: true,
				wordWrapWidth: ELLIPSE_WIDTH - PADDING * 2
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
		this.addChild(this.statusIndicator);
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
	 * ユースケースを描画
	 */
	draw(): void {
		this.drawBackground();
		this.drawStatusIndicator();
		this.drawTexts();
	}

	/**
	 * 背景（楕円）を描画
	 */
	private drawBackground(): void {
		this.background.clear();

		const centerX = ELLIPSE_WIDTH / 2;
		const centerY = ELLIPSE_HEIGHT / 2;

		let bgColor = COLORS.background;
		let borderColor = COLORS.border;
		let borderWidth = 2;

		if (this.isSelected) {
			bgColor = COLORS.backgroundSelected;
			borderColor = COLORS.borderSelected;
			borderWidth = 3;
		} else if (this.isHovered) {
			bgColor = COLORS.backgroundHover;
			borderColor = COLORS.borderHover;
		}

		// グロー効果（選択/ホバー時）
		if (this.isSelected || this.isHovered) {
			this.background.ellipse(centerX, centerY, ELLIPSE_WIDTH / 2 + 4, ELLIPSE_HEIGHT / 2 + 4);
			this.background.fill({ color: borderColor, alpha: 0.15 });
		}

		// メイン楕円
		this.background.ellipse(centerX, centerY, ELLIPSE_WIDTH / 2, ELLIPSE_HEIGHT / 2);
		this.background.fill(bgColor);
		this.background.stroke({ width: borderWidth, color: borderColor });

		// 上部ハイライト（金属感）
		this.background.ellipse(centerX, centerY - 5, ELLIPSE_WIDTH / 2 - 10, ELLIPSE_HEIGHT / 2 - 15);
		this.background.stroke({ width: 1, color: 0x666666, alpha: 0.3 });
	}

	/**
	 * ステータスインジケーターを描画（左端の小円）
	 */
	private drawStatusIndicator(): void {
		this.statusIndicator.clear();

		const statusColor = COLORS.status[this.usecase.status] || COLORS.status.draft;
		const indicatorY = ELLIPSE_HEIGHT / 2;

		// ステータスドット
		this.statusIndicator.circle(STATUS_INDICATOR_X, indicatorY, STATUS_INDICATOR_RADIUS);
		this.statusIndicator.fill(statusColor);
		this.statusIndicator.stroke({ width: 1, color: 0x1a1a1a });
	}

	/**
	 * テキストを描画
	 */
	private drawTexts(): void {
		const centerX = ELLIPSE_WIDTH / 2;

		// タイトル（中央）
		const maxTitleChars = 18;
		const title = this.usecase.title.length > maxTitleChars
			? this.usecase.title.substring(0, maxTitleChars - 2) + '..'
			: this.usecase.title;
		this.titleText.text = title;
		this.titleText.x = centerX - this.titleText.width / 2;
		this.titleText.y = ELLIPSE_HEIGHT / 2 - this.titleText.height / 2 - 3;

		// ID（下部）
		this.idText.text = this.usecase.id;
		this.idText.x = centerX - this.idText.width / 2;
		this.idText.y = ELLIPSE_HEIGHT / 2 + 8;
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
	 * ノードの幅を取得
	 */
	static getWidth(): number {
		return ELLIPSE_WIDTH;
	}

	/**
	 * ノードの高さを取得
	 */
	static getHeight(): number {
		return ELLIPSE_HEIGHT;
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
