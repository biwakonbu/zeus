// RelationEdge - UML ユースケース図の関係線描画クラス
// include, extend, generalize の各リレーションタイプに対応
import { Graphics, Text } from 'pixi.js';
import type { RelationType } from '$lib/types/api';
import { TEXT_RESOLUTION, COMMON_COLORS } from './constants';

// 色定義（Factorio テーマ準拠）
const COLORS = {
	include: 0x66cc99, // 緑（include）
	extend: 0xffcc00, // 黄（extend）
	generalize: 0x4488ff, // 青（generalize）
	default: COMMON_COLORS.textMuted,
	highlighted: COMMON_COLORS.highlighted
};

// 線スタイル定数
const LINE_WIDTH = 2;
const DASH_LENGTH = 8;
const GAP_LENGTH = 4;
const ARROW_SIZE = 10;
const ARROW_ANGLE = Math.PI / 6; // 30度
const LABEL_OFFSET = 10;

/**
 * RelationEdge - UML リレーションの視覚的表現
 *
 * 責務:
 * - リレーションタイプに応じた線描画（実線/点線）
 * - 矢印の描画（通常/三角）
 * - ステレオタイプラベルの描画（<<include>>, <<extend>>）
 */
export class RelationEdge extends Graphics {
	private fromId: string;
	private toId: string;
	private relationType: RelationType;
	private condition?: string;

	// 座標
	private fromX: number = 0;
	private fromY: number = 0;
	private toX: number = 0;
	private toY: number = 0;

	// ラベル
	private labelText: Text | null = null;

	// 状態
	private isHighlighted = false;

	constructor(fromId: string, toId: string, type: RelationType, condition?: string) {
		super();

		this.fromId = fromId;
		this.toId = toId;
		this.relationType = type;
		this.condition = condition;

		// ラベル作成
		this.createLabel();
	}

	/**
	 * ラベルを作成
	 */
	private createLabel(): void {
		const labelStyle = {
			fontSize: 10,
			fill: COLORS[this.relationType] || COLORS.default,
			fontFamily: 'IBM Plex Mono, monospace',
			fontStyle: 'italic' as const
		};

		let labelContent = '';
		if (this.relationType === 'include') {
			labelContent = '<<include>>';
		} else if (this.relationType === 'extend') {
			labelContent = this.condition ? `<<extend>>\n[${this.condition}]` : '<<extend>>';
		}
		// generalize はラベルなし

		if (labelContent) {
			this.labelText = new Text({
				text: labelContent,
				style: labelStyle,
				resolution: TEXT_RESOLUTION
			});
			this.addChild(this.labelText);
		}
	}

	/**
	 * エッジの両端の座標を設定
	 */
	setEndpoints(fromX: number, fromY: number, toX: number, toY: number): void {
		this.fromX = fromX;
		this.fromY = fromY;
		this.toX = toX;
		this.toY = toY;
		this.draw();
	}

	/**
	 * エッジを描画
	 */
	draw(): void {
		this.clear();

		const color = this.isHighlighted
			? COLORS.highlighted
			: COLORS[this.relationType] || COLORS.default;
		const width = this.isHighlighted ? LINE_WIDTH + 1 : LINE_WIDTH;

		if (this.relationType === 'generalize') {
			// generalize: 実線 + 三角矢印
			this.drawSolidLine(color, width);
			this.drawTriangleArrow(color);
		} else {
			// include/extend: 点線 + 通常矢印
			// パス構築を統合し、stroke は最後に一度だけ呼ぶ
			this.buildDashedLinePath();
			this.buildArrowPath();
			this.stroke({ width, color });
		}

		// ラベル位置更新
		this.updateLabelPosition();
	}

	/**
	 * 実線を描画
	 */
	private drawSolidLine(color: number, width: number): void {
		this.moveTo(this.fromX, this.fromY);
		this.lineTo(this.toX, this.toY);
		this.stroke({ width, color });
	}

	/**
	 * 点線パスを構築（stroke は呼ばない）
	 */
	private buildDashedLinePath(): void {
		const dx = this.toX - this.fromX;
		const dy = this.toY - this.fromY;
		const distance = Math.sqrt(dx * dx + dy * dy);

		if (distance === 0) return;

		const unitX = dx / distance;
		const unitY = dy / distance;

		let drawn = 0;
		let drawing = true;

		// パスを構築（stroke は draw() で一度だけ呼ぶ）
		while (drawn < distance) {
			const segmentLength = drawing ? DASH_LENGTH : GAP_LENGTH;
			const remainingDistance = distance - drawn;
			const actualLength = Math.min(segmentLength, remainingDistance);

			const startX = this.fromX + unitX * drawn;
			const startY = this.fromY + unitY * drawn;
			const endX = this.fromX + unitX * (drawn + actualLength);
			const endY = this.fromY + unitY * (drawn + actualLength);

			if (drawing) {
				this.moveTo(startX, startY);
				this.lineTo(endX, endY);
			}

			drawn += actualLength;
			drawing = !drawing;
		}
	}

	/**
	 * 通常矢印パスを構築（stroke は呼ばない）
	 */
	private buildArrowPath(): void {
		const dx = this.toX - this.fromX;
		const dy = this.toY - this.fromY;
		const angle = Math.atan2(dy, dx);

		const arrowX1 = this.toX - ARROW_SIZE * Math.cos(angle - ARROW_ANGLE);
		const arrowY1 = this.toY - ARROW_SIZE * Math.sin(angle - ARROW_ANGLE);
		const arrowX2 = this.toX - ARROW_SIZE * Math.cos(angle + ARROW_ANGLE);
		const arrowY2 = this.toY - ARROW_SIZE * Math.sin(angle + ARROW_ANGLE);

		this.moveTo(this.toX, this.toY);
		this.lineTo(arrowX1, arrowY1);
		this.moveTo(this.toX, this.toY);
		this.lineTo(arrowX2, arrowY2);
	}

	/**
	 * 三角矢印を描画（generalize用）
	 */
	private drawTriangleArrow(color: number): void {
		const dx = this.toX - this.fromX;
		const dy = this.toY - this.fromY;
		const angle = Math.atan2(dy, dx);

		const arrowX1 = this.toX - ARROW_SIZE * 1.5 * Math.cos(angle - ARROW_ANGLE);
		const arrowY1 = this.toY - ARROW_SIZE * 1.5 * Math.sin(angle - ARROW_ANGLE);
		const arrowX2 = this.toX - ARROW_SIZE * 1.5 * Math.cos(angle + ARROW_ANGLE);
		const arrowY2 = this.toY - ARROW_SIZE * 1.5 * Math.sin(angle + ARROW_ANGLE);

		// 塗りつぶし三角形
		this.moveTo(this.toX, this.toY);
		this.lineTo(arrowX1, arrowY1);
		this.lineTo(arrowX2, arrowY2);
		this.closePath();
		this.fill({ color: 0x2d2d2d }); // 内部は背景色
		this.stroke({ width: LINE_WIDTH, color });
	}

	/**
	 * ラベル位置を更新
	 */
	private updateLabelPosition(): void {
		if (!this.labelText) return;

		// 線の中央に配置
		const midX = (this.fromX + this.toX) / 2;
		const midY = (this.fromY + this.toY) / 2;

		// 線の角度を計算してオフセット
		const dx = this.toX - this.fromX;
		const dy = this.toY - this.fromY;
		const angle = Math.atan2(dy, dx);

		// 線に垂直方向にオフセット
		const offsetX = -Math.sin(angle) * LABEL_OFFSET;
		const offsetY = Math.cos(angle) * LABEL_OFFSET;

		this.labelText.x = midX + offsetX - this.labelText.width / 2;
		this.labelText.y = midY + offsetY - this.labelText.height / 2;
	}

	/**
	 * ハイライト状態を設定
	 */
	setHighlighted(highlighted: boolean): void {
		if (this.isHighlighted !== highlighted) {
			this.isHighlighted = highlighted;
			this.draw();
		}
	}

	/**
	 * From ノード ID を取得
	 */
	getFromId(): string {
		return this.fromId;
	}

	/**
	 * To ノード ID を取得
	 */
	getToId(): string {
		return this.toId;
	}

	/**
	 * リレーションタイプを取得
	 */
	getRelationType(): RelationType {
		return this.relationType;
	}

	/**
	 * エッジの識別キーを生成
	 */
	static createKey(fromId: string, toId: string, type: RelationType): string {
		return `${fromId}--${type}-->${toId}`;
	}

	/**
	 * このエッジのキーを取得
	 */
	getKey(): string {
		return RelationEdge.createKey(this.fromId, this.toId, this.relationType);
	}
}

/**
 * Actor と UseCase の関連線描画クラス
 * シンプルな実線で接続
 */
export class ActorUseCaseEdge extends Graphics {
	private actorId: string;
	private usecaseId: string;
	private isPrimary: boolean;

	// 座標
	private fromX: number = 0;
	private fromY: number = 0;
	private toX: number = 0;
	private toY: number = 0;

	// 状態
	private isHighlighted = false;

	constructor(actorId: string, usecaseId: string, isPrimary: boolean = true) {
		super();

		this.actorId = actorId;
		this.usecaseId = usecaseId;
		this.isPrimary = isPrimary;
	}

	/**
	 * エッジの両端の座標を設定
	 */
	setEndpoints(fromX: number, fromY: number, toX: number, toY: number): void {
		this.fromX = fromX;
		this.fromY = fromY;
		this.toX = toX;
		this.toY = toY;
		this.draw();
	}

	/**
	 * エッジを描画
	 */
	draw(): void {
		this.clear();

		const color = this.isHighlighted ? 0xff9533 : this.isPrimary ? 0xe0e0e0 : 0x888888;
		const width = this.isHighlighted ? 3 : this.isPrimary ? 2 : 1;
		const alpha = this.isPrimary ? 1.0 : 0.6;

		this.moveTo(this.fromX, this.fromY);
		this.lineTo(this.toX, this.toY);
		this.stroke({ width, color, alpha });
	}

	/**
	 * ハイライト状態を設定
	 */
	setHighlighted(highlighted: boolean): void {
		if (this.isHighlighted !== highlighted) {
			this.isHighlighted = highlighted;
			this.draw();
		}
	}

	/**
	 * Actor ID を取得
	 */
	getActorId(): string {
		return this.actorId;
	}

	/**
	 * UseCase ID を取得
	 */
	getUseCaseId(): string {
		return this.usecaseId;
	}

	/**
	 * エッジの識別キーを生成
	 */
	static createKey(actorId: string, usecaseId: string): string {
		return `${actorId}--assoc-->${usecaseId}`;
	}

	/**
	 * このエッジのキーを取得
	 */
	getKey(): string {
		return ActorUseCaseEdge.createKey(this.actorId, this.usecaseId);
	}
}
