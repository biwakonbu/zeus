// ActorNode - UML ユースケース図のアクター（棒人間）描画クラス
// 棒人間のスタイルで人間アクター、アイコン形式でシステム/時間/デバイス/外部アクターを表現
import { Container, Graphics, Text, FederatedPointerEvent } from 'pixi.js';
import type { ActorItem, ActorType } from '$lib/types/api';
import { TEXT_RESOLUTION, COMMON_COLORS } from './constants';

// サイズ定数
const ACTOR_WIDTH = 60;
const ACTOR_HEIGHT = 100;
const HEAD_RADIUS = 12;
const BODY_LENGTH = 25;
const ARM_LENGTH = 20;
const LEG_LENGTH = 25;
const LINE_WIDTH = 2;

// 色定義（Factorio テーマ準拠）
const COLORS = {
	// アクタータイプ別の色
	actorType: {
		human: 0xff9533,    // オレンジ（人間）
		system: 0x4488ff,   // ブルー（システム）
		time: 0xffcc00,     // イエロー（時間）
		device: 0x66cc99,   // グリーン（デバイス）
		external: 0xcc66ff  // パープル（外部）
	} as Record<ActorType, number>,
	// 基本色（共通定数から取得）
	stroke: COMMON_COLORS.text,
	strokeHover: COMMON_COLORS.borderHover,
	strokeSelected: COMMON_COLORS.borderSelected,
	text: COMMON_COLORS.text,
	textMuted: COMMON_COLORS.textMuted
};

/**
 * ActorNode - UML アクターの視覚的表現
 *
 * 責務:
 * - 棒人間（human）またはアイコン形式の描画
 * - アクタータイプに応じたスタイル変更
 * - インタラクション（クリック、ホバー）
 */
export class ActorNode extends Container {
	private actor: ActorItem;
	private graphics: Graphics;
	private labelText: Text;
	private typeText: Text;

	private isHovered = false;
	private isSelected = false;

	// イベントコールバック
	private onClickCallback?: (node: ActorNode, event?: FederatedPointerEvent) => void;
	private onHoverCallback?: (node: ActorNode, isHovered: boolean, event?: MouseEvent) => void;

	constructor(actor: ActorItem) {
		super();

		this.actor = actor;

		// グラフィックス初期化
		this.graphics = new Graphics();
		this.labelText = new Text({
			text: '',
			style: {
				fontSize: 11,
				fill: COLORS.text,
				fontFamily: 'IBM Plex Mono, monospace',
				align: 'center'
			},
			resolution: TEXT_RESOLUTION
		});
		this.typeText = new Text({
			text: '',
			style: {
				fontSize: 9,
				fill: COLORS.textMuted,
				fontFamily: 'IBM Plex Mono, monospace',
				align: 'center'
			},
			resolution: TEXT_RESOLUTION
		});

		this.addChild(this.graphics);
		this.addChild(this.labelText);
		this.addChild(this.typeText);

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
	 * アクターを描画
	 */
	draw(): void {
		this.graphics.clear();

		const color = this.getColor();
		const strokeWidth = this.isHovered || this.isSelected ? LINE_WIDTH + 1 : LINE_WIDTH;

		// アクタータイプに応じて描画を分岐
		if (this.actor.type === 'human') {
			this.drawStickFigure(color, strokeWidth);
		} else {
			this.drawIcon(color, strokeWidth);
		}

		// ラベル描画
		this.drawLabel();
	}

	/**
	 * 棒人間を描画（human タイプ）
	 */
	private drawStickFigure(color: number, strokeWidth: number): void {
		const g = this.graphics;
		const centerX = ACTOR_WIDTH / 2;
		const headY = HEAD_RADIUS + 5;

		// 選択/ホバー時の背景グロー
		if (this.isSelected || this.isHovered) {
			g.circle(centerX, headY + BODY_LENGTH / 2, 35);
			g.fill({ color: color, alpha: 0.1 });
		}

		// 頭（円）
		g.circle(centerX, headY, HEAD_RADIUS);
		g.stroke({ width: strokeWidth, color });

		// 胴体（縦線）
		const bodyTop = headY + HEAD_RADIUS;
		const bodyBottom = bodyTop + BODY_LENGTH;
		g.moveTo(centerX, bodyTop);
		g.lineTo(centerX, bodyBottom);
		g.stroke({ width: strokeWidth, color });

		// 腕（横線）
		const armY = bodyTop + BODY_LENGTH * 0.3;
		g.moveTo(centerX - ARM_LENGTH, armY);
		g.lineTo(centerX + ARM_LENGTH, armY);
		g.stroke({ width: strokeWidth, color });

		// 左足
		g.moveTo(centerX, bodyBottom);
		g.lineTo(centerX - LEG_LENGTH * 0.6, bodyBottom + LEG_LENGTH);
		g.stroke({ width: strokeWidth, color });

		// 右足
		g.moveTo(centerX, bodyBottom);
		g.lineTo(centerX + LEG_LENGTH * 0.6, bodyBottom + LEG_LENGTH);
		g.stroke({ width: strokeWidth, color });
	}

	/**
	 * アイコン形式で描画（system, time, device, external タイプ）
	 */
	private drawIcon(color: number, strokeWidth: number): void {
		const g = this.graphics;
		const centerX = ACTOR_WIDTH / 2;
		const centerY = 40;
		const iconSize = 24;

		// 選択/ホバー時の背景
		if (this.isSelected || this.isHovered) {
			g.circle(centerX, centerY, iconSize + 8);
			g.fill({ color: color, alpha: 0.15 });
		}

		// 外枠（円形）
		g.circle(centerX, centerY, iconSize);
		g.stroke({ width: strokeWidth, color });

		// タイプ別のアイコン
		switch (this.actor.type) {
			case 'system':
				this.drawSystemIcon(g, centerX, centerY, iconSize * 0.6, color, strokeWidth);
				break;
			case 'time':
				this.drawTimeIcon(g, centerX, centerY, iconSize * 0.6, color, strokeWidth);
				break;
			case 'device':
				this.drawDeviceIcon(g, centerX, centerY, iconSize * 0.6, color, strokeWidth);
				break;
			case 'external':
				this.drawExternalIcon(g, centerX, centerY, iconSize * 0.6, color, strokeWidth);
				break;
		}
	}

	/**
	 * システムアイコン（サーバー風のボックス）
	 */
	private drawSystemIcon(g: Graphics, cx: number, cy: number, size: number, color: number, sw: number): void {
		const halfSize = size / 2;
		// ボックス
		g.rect(cx - halfSize, cy - halfSize * 1.2, size, size * 1.4);
		g.stroke({ width: sw, color });
		// 横線（サーバーラック風）
		g.moveTo(cx - halfSize + 2, cy - halfSize * 0.4);
		g.lineTo(cx + halfSize - 2, cy - halfSize * 0.4);
		g.stroke({ width: sw * 0.7, color });
		g.moveTo(cx - halfSize + 2, cy + halfSize * 0.4);
		g.lineTo(cx + halfSize - 2, cy + halfSize * 0.4);
		g.stroke({ width: sw * 0.7, color });
	}

	/**
	 * 時間アイコン（時計）
	 */
	private drawTimeIcon(g: Graphics, cx: number, cy: number, size: number, color: number, sw: number): void {
		// 時計の針
		g.moveTo(cx, cy);
		g.lineTo(cx, cy - size * 0.7);
		g.stroke({ width: sw, color });
		g.moveTo(cx, cy);
		g.lineTo(cx + size * 0.5, cy);
		g.stroke({ width: sw, color });
		// 中心点
		g.circle(cx, cy, 2);
		g.fill(color);
	}

	/**
	 * デバイスアイコン（スマートフォン風）
	 */
	private drawDeviceIcon(g: Graphics, cx: number, cy: number, size: number, color: number, sw: number): void {
		const w = size * 0.7;
		const h = size * 1.2;
		// 外枠
		g.roundRect(cx - w / 2, cy - h / 2, w, h, 3);
		g.stroke({ width: sw, color });
		// ホームボタン
		g.circle(cx, cy + h / 2 - 4, 2);
		g.fill(color);
	}

	/**
	 * 外部アイコン（地球儀風）
	 */
	private drawExternalIcon(g: Graphics, cx: number, cy: number, size: number, color: number, sw: number): void {
		// 縦線（経線）
		g.ellipse(cx, cy, size * 0.4, size * 0.8);
		g.stroke({ width: sw * 0.7, color });
		// 横線（緯線）
		g.moveTo(cx - size * 0.8, cy);
		g.lineTo(cx + size * 0.8, cy);
		g.stroke({ width: sw * 0.7, color });
	}

	/**
	 * ラベルを描画
	 */
	private drawLabel(): void {
		const centerX = ACTOR_WIDTH / 2;

		// アクター名
		const maxChars = 10;
		const displayName = this.actor.title.length > maxChars
			? this.actor.title.substring(0, maxChars - 1) + '..'
			: this.actor.title;
		this.labelText.text = displayName;
		this.labelText.x = centerX - this.labelText.width / 2;
		this.labelText.y = ACTOR_HEIGHT - 22;

		// タイプ表示
		const typeLabels: Record<ActorType, string> = {
			human: 'Human',
			system: 'System',
			time: 'Time',
			device: 'Device',
			external: 'External'
		};
		this.typeText.text = typeLabels[this.actor.type] || this.actor.type;
		this.typeText.x = centerX - this.typeText.width / 2;
		this.typeText.y = ACTOR_HEIGHT - 10;
	}

	/**
	 * 色を取得
	 */
	private getColor(): number {
		if (this.isSelected) return COLORS.strokeSelected;
		if (this.isHovered) return COLORS.strokeHover;
		return COLORS.actorType[this.actor.type] || COLORS.stroke;
	}

	/**
	 * ホバー処理
	 */
	private handleHover(isHovered: boolean, event?: FederatedPointerEvent): void {
		this.isHovered = isHovered;
		this.draw();
		// FederatedPointerEvent から MouseEvent 情報を抽出
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
	 * アクターデータを更新
	 */
	updateActor(actor: ActorItem): void {
		this.actor = actor;
		this.draw();
	}

	/**
	 * アクターIDを取得
	 */
	getActorId(): string {
		return this.actor.id;
	}

	/**
	 * アクターデータを取得
	 */
	getActor(): ActorItem {
		return this.actor;
	}

	/**
	 * ノードの幅を取得
	 */
	static getWidth(): number {
		return ACTOR_WIDTH;
	}

	/**
	 * ノードの高さを取得
	 */
	static getHeight(): number {
		return ACTOR_HEIGHT;
	}

	/**
	 * イベントリスナーを設定
	 */
	onClick(callback: (node: ActorNode, event?: FederatedPointerEvent) => void): void {
		this.onClickCallback = callback;
	}

	onHover(callback: (node: ActorNode, isHovered: boolean, event?: MouseEvent) => void): void {
		this.onHoverCallback = callback;
	}
}
