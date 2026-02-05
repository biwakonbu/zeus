// ActorNode - UML ユースケース図のアクター描画クラス
// UML 標準のシンプルな棒人間とステレオタイプ付き長方形
import { Container, Graphics, Text } from 'pixi.js';
import type { FederatedPointerEvent } from 'pixi.js';
import type { ActorItem, ActorType } from '$lib/types/api';
import { TEXT_RESOLUTION, COMMON_COLORS } from './constants';

// サイズ定数
const ACTOR_WIDTH = 60;
const ACTOR_HEIGHT = 90;
const LINE_WIDTH = 2;

// Human（棒人間）定数
const HUMAN = {
	headRadius: 8,
	bodyLength: 20,
	armLength: 16,
	legLength: 18
};

// 非人間アクター（長方形）定数
const BOX = {
	width: 50,
	height: 40,
	centerY: 35
};

// 色定義（Factorio テーマ準拠）
const COLORS = {
	// 基本色（共通定数から取得）
	stroke: COMMON_COLORS.text,
	strokeHover: COMMON_COLORS.borderHover,
	strokeSelected: COMMON_COLORS.borderSelected,
	text: COMMON_COLORS.text,
	textMuted: COMMON_COLORS.textMuted,
	background: COMMON_COLORS.backgroundPanel
};

/**
 * ActorNode - UML アクターの視覚的表現
 *
 * 責務:
 * - UML 標準のアクターアイコン描画
 * - Human: 棒人間（stick figure）
 * - 非人間: ステレオタイプ付き長方形
 * - インタラクション（クリック、ホバー）
 */
export class ActorNode extends Container {
	private actor: ActorItem;
	private graphics: Graphics;
	private labelText: Text;
	private stereotypeText: Text;

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
		this.stereotypeText = new Text({
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
		this.addChild(this.stereotypeText);
		this.addChild(this.labelText);

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

		// Human は棒人間、それ以外はステレオタイプ付き長方形
		if (this.actor.type === 'human') {
			this.drawStickFigure(color, strokeWidth);
		} else {
			this.drawStereotypeBox(color, strokeWidth);
		}

		// ラベル描画
		this.drawLabel();
	}

	/**
	 * UML 標準の棒人間を描画
	 */
	private drawStickFigure(color: number, sw: number): void {
		const g = this.graphics;
		const cx = ACTOR_WIDTH / 2;
		const headY = HUMAN.headRadius + 5;

		// 選択/ホバー時の背景グロー
		if (this.isSelected || this.isHovered) {
			g.circle(cx, headY + HUMAN.bodyLength / 2, 30);
			g.fill({ color: color, alpha: 0.1 });
		}

		// 頭（円）
		g.circle(cx, headY, HUMAN.headRadius);
		g.stroke({ width: sw, color });

		// 体（縦線）
		const bodyTop = headY + HUMAN.headRadius;
		const bodyBottom = bodyTop + HUMAN.bodyLength;
		g.moveTo(cx, bodyTop);
		g.lineTo(cx, bodyBottom);
		g.stroke({ width: sw, color });

		// 腕（横線、T字型）
		const armY = bodyTop + HUMAN.bodyLength * 0.3;
		g.moveTo(cx - HUMAN.armLength, armY);
		g.lineTo(cx + HUMAN.armLength, armY);
		g.stroke({ width: sw, color });

		// 左足
		g.moveTo(cx, bodyBottom);
		g.lineTo(cx - HUMAN.legLength * 0.6, bodyBottom + HUMAN.legLength);
		g.stroke({ width: sw, color });

		// 右足
		g.moveTo(cx, bodyBottom);
		g.lineTo(cx + HUMAN.legLength * 0.6, bodyBottom + HUMAN.legLength);
		g.stroke({ width: sw, color });
	}

	/**
	 * UML 標準のステレオタイプ付き長方形を描画
	 */
	private drawStereotypeBox(color: number, sw: number): void {
		const g = this.graphics;
		const cx = ACTOR_WIDTH / 2;
		const cy = BOX.centerY;

		// 選択/ホバー時の背景グロー
		if (this.isSelected || this.isHovered) {
			g.roundRect(
				cx - BOX.width / 2 - 5,
				cy - BOX.height / 2 - 5,
				BOX.width + 10,
				BOX.height + 10,
				4
			);
			g.fill({ color: color, alpha: 0.1 });
		}

		// 長方形
		g.rect(cx - BOX.width / 2, cy - BOX.height / 2, BOX.width, BOX.height);
		g.fill({ color: COLORS.background, alpha: 0.8 });
		g.stroke({ width: sw, color });

		// ステレオタイプをボックス内に表示
		const stereotypes: Record<ActorType, string> = {
			human: '',
			system: '«system»',
			time: '«time»',
			device: '«device»',
			external: '«external»'
		};
		const stereotype = stereotypes[this.actor.type];
		if (stereotype) {
			this.stereotypeText.text = stereotype;
			this.stereotypeText.x = cx - this.stereotypeText.width / 2;
			this.stereotypeText.y = cy - BOX.height / 2 + 5;
		} else {
			this.stereotypeText.text = '';
		}
	}

	/**
	 * ラベルを描画
	 */
	private drawLabel(): void {
		const centerX = ACTOR_WIDTH / 2;

		// アクター名
		const maxChars = 10;
		const displayName =
			this.actor.title.length > maxChars
				? this.actor.title.substring(0, maxChars - 1) + '..'
				: this.actor.title;
		this.labelText.text = displayName;
		this.labelText.x = centerX - this.labelText.width / 2;
		this.labelText.y = ACTOR_HEIGHT - 14;
	}

	/**
	 * 色を取得
	 */
	private getColor(): number {
		if (this.isSelected) return COLORS.strokeSelected;
		if (this.isHovered) return COLORS.strokeHover;
		return COLORS.stroke;
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
