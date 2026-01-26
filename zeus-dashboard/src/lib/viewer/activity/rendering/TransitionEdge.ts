// TransitionEdge - UML アクティビティ図の遷移（矢印）描画クラス
import { Graphics, Text } from 'pixi.js';
import type { ActivityTransitionItem } from '$lib/types/api';
import { TRANSITION_STYLE, COMMON_COLORS, TEXT_RESOLUTION } from './constants';

/**
 * 接続ポイント（ノードの境界上の点）
 */
interface ConnectionPoint {
	x: number;
	y: number;
}

/**
 * TransitionEdge - 遷移エッジ
 *
 * 責務:
 * - ノード間の遷移（矢印）を描画
 * - ガード条件のラベル表示
 * - ホバー/選択状態の視覚的フィードバック
 */
export class TransitionEdge extends Graphics {
	private transition: ActivityTransitionItem;
	private sourcePoint: ConnectionPoint;
	private targetPoint: ConnectionPoint;
	private guardText: Text | null = null;

	private isHovered = false;
	private isSelected = false;

	constructor(
		transition: ActivityTransitionItem,
		sourcePoint: ConnectionPoint,
		targetPoint: ConnectionPoint
	) {
		super();

		this.transition = transition;
		this.sourcePoint = sourcePoint;
		this.targetPoint = targetPoint;

		// ガード条件がある場合はテキストコンポーネント作成
		if (transition.guard) {
			this.guardText = new Text({
				text: transition.guard,
				style: {
					fontSize: TRANSITION_STYLE.guardFontSize,
					fill: COMMON_COLORS.textMuted,
					fontFamily: 'IBM Plex Mono, monospace'
				},
				resolution: TEXT_RESOLUTION
			});
			this.addChild(this.guardText);
		}

		// インタラクション設定
		this.eventMode = 'static';
		this.cursor = 'pointer';

		// 初回描画
		this.draw();
	}

	/**
	 * エッジを描画
	 */
	draw(): void {
		this.clear();

		const color = this.getLineColor();
		const width = this.getLineWidth();

		// 水平方向のオフセットを計算
		const dx = Math.abs(this.targetPoint.x - this.sourcePoint.x);
		const dy = this.targetPoint.y - this.sourcePoint.y;

		// 水平オフセットが大きく、かつ下方向への遷移の場合は曲線を使用
		if (dx > TRANSITION_STYLE.curveThreshold && dy > 0) {
			// ベジェ曲線で描画
			const midY = (this.sourcePoint.y + this.targetPoint.y) / 2;
			this.moveTo(this.sourcePoint.x, this.sourcePoint.y);
			this.bezierCurveTo(
				this.sourcePoint.x,
				midY,
				this.targetPoint.x,
				midY,
				this.targetPoint.x,
				this.targetPoint.y
			);
		} else {
			// 直線で描画
			this.moveTo(this.sourcePoint.x, this.sourcePoint.y);
			this.lineTo(this.targetPoint.x, this.targetPoint.y);
		}
		this.stroke({ width, color });

		// 矢印を描画
		this.drawArrow(color);

		// ガード条件ラベルを配置
		this.positionGuardText();
	}

	/**
	 * 矢印を描画（改善版：より鋭角、縁取り付き）
	 */
	private drawArrow(color: number): void {
		const arrowSize = TRANSITION_STYLE.arrowSize;
		const arrowAngle = TRANSITION_STYLE.arrowAngle;
		const dx = this.targetPoint.x - this.sourcePoint.x;
		const dy = this.targetPoint.y - this.sourcePoint.y;
		const angle = Math.atan2(dy, dx);

		// 矢印の頂点（ターゲットポイント）
		const tipX = this.targetPoint.x;
		const tipY = this.targetPoint.y;

		// 矢印の左右の点（より鋭角に）
		const leftX = tipX - arrowSize * Math.cos(angle - arrowAngle);
		const leftY = tipY - arrowSize * Math.sin(angle - arrowAngle);
		const rightX = tipX - arrowSize * Math.cos(angle + arrowAngle);
		const rightY = tipY - arrowSize * Math.sin(angle + arrowAngle);

		// 三角形を描画（塗りつぶし + 縁取り）
		this.moveTo(tipX, tipY);
		this.lineTo(leftX, leftY);
		this.lineTo(rightX, rightY);
		this.closePath();
		this.fill(color);
		this.stroke({ width: 1, color: 0x888888, alpha: 0.5 });
	}

	/**
	 * ガード条件ラベルを配置
	 */
	private positionGuardText(): void {
		if (!this.guardText) return;

		// 線の中点に配置
		const midX = (this.sourcePoint.x + this.targetPoint.x) / 2;
		const midY = (this.sourcePoint.y + this.targetPoint.y) / 2;

		// 線から少しオフセット
		const dx = this.targetPoint.x - this.sourcePoint.x;
		const dy = this.targetPoint.y - this.sourcePoint.y;
		const angle = Math.atan2(dy, dx);
		const offset = 12;

		// 線の法線方向にオフセット
		this.guardText.x = midX + offset * Math.cos(angle + Math.PI / 2) - this.guardText.width / 2;
		this.guardText.y = midY + offset * Math.sin(angle + Math.PI / 2) - this.guardText.height / 2;
	}

	/**
	 * 線の色を取得
	 */
	private getLineColor(): number {
		if (this.isSelected) {
			return COMMON_COLORS.borderSelected;
		} else if (this.isHovered) {
			return COMMON_COLORS.borderHover;
		}
		return COMMON_COLORS.border;
	}

	/**
	 * 線の太さを取得
	 */
	private getLineWidth(): number {
		if (this.isSelected) {
			return TRANSITION_STYLE.lineWidth + 1;
		} else if (this.isHovered) {
			return TRANSITION_STYLE.lineWidth + 0.5;
		}
		return TRANSITION_STYLE.lineWidth;
	}

	/**
	 * 接続ポイントを更新
	 */
	updatePoints(sourcePoint: ConnectionPoint, targetPoint: ConnectionPoint): void {
		this.sourcePoint = sourcePoint;
		this.targetPoint = targetPoint;
		this.draw();
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
	 * ホバー状態を設定
	 */
	setHovered(hovered: boolean): void {
		if (this.isHovered !== hovered) {
			this.isHovered = hovered;
			this.draw();
		}
	}

	/**
	 * 遷移IDを取得
	 */
	getTransitionId(): string {
		return this.transition.id;
	}

	/**
	 * ソースノードIDを取得
	 */
	getSourceId(): string {
		return this.transition.source;
	}

	/**
	 * ターゲットノードIDを取得
	 */
	getTargetId(): string {
		return this.transition.target;
	}

	/**
	 * 遷移データを取得
	 */
	getTransition(): ActivityTransitionItem {
		return this.transition;
	}
}
