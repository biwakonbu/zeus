// TransitionEdge - UML アクティビティ図の遷移（矢印）描画クラス
// 2層構造（外側縁取り → コア）でシンプルに視認性を確保
import { Container, Graphics, Text } from 'pixi.js';
import type { ActivityTransitionItem } from '$lib/types/api';
import {
	TRANSITION_STYLE,
	TEXT_RESOLUTION,
	TRANSITION_EDGE_STYLE,
	TRANSITION_EDGE_WIDTHS,
	GUARD_LABEL_STYLE
} from './constants';
import { formatGuardCondition } from '$lib/viewer/shared/utils';

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
	private guardContainer: Container | null = null;
	private guardBackground: Graphics | null = null;
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

		// ガード条件がある場合はバッジ風コンポーネント作成
		if (transition.guard) {
			this.guardContainer = new Container();

			// 背景（角丸矩形）
			this.guardBackground = new Graphics();
			this.guardContainer.addChild(this.guardBackground);

			// テキスト（UML 準拠: 角括弧で囲む）
			const guardLabel = formatGuardCondition(transition.guard);
			this.guardText = new Text({
				text: guardLabel,
				style: {
					fontSize: GUARD_LABEL_STYLE.fontSize,
					fill: GUARD_LABEL_STYLE.text,
					fontFamily: 'IBM Plex Mono, monospace',
					fontWeight: '500'
				},
				resolution: TEXT_RESOLUTION
			});
			this.guardContainer.addChild(this.guardText);

			this.addChild(this.guardContainer);
		}

		// インタラクション設定
		this.eventMode = 'static';
		this.cursor = 'pointer';

		// 初回描画
		this.draw();
	}

	/**
	 * エッジを描画（2層構造: 外側縁取り → コア）
	 * シンプルで視認性を確保
	 */
	draw(): void {
		this.clear();

		const style = this.getEdgeStyle();
		const widths = this.getEdgeWidths();

		// 水平方向のオフセットを計算
		const dx = Math.abs(this.targetPoint.x - this.sourcePoint.x);
		const dy = this.targetPoint.y - this.sourcePoint.y;

		// 曲線か直線かを判定
		const useCurve = dx > TRANSITION_STYLE.curveThreshold && dy > 0;
		const midY = (this.sourcePoint.y + this.targetPoint.y) / 2;

		// Layer 1: 外側（縁取り）- 暗めの縁取りでコアを際立たせる
		this.drawPath(useCurve, midY);
		this.stroke({ width: widths.outer, color: style.outer, alpha: 1.0 });

		// Layer 2: コア（内側）- 明るいコア線
		this.drawPath(useCurve, midY);
		this.stroke({ width: widths.core, color: style.core, alpha: 1.0 });

		// 矢印を描画（2層構造対応）
		this.drawArrow(style, widths);

		// ガード条件ラベルを配置
		this.positionGuardText();
	}

	/**
	 * パスを描画（曲線または直線）
	 */
	private drawPath(useCurve: boolean, midY: number): void {
		if (useCurve) {
			// ベジェ曲線で描画
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
	}

	/**
	 * 矢印を描画（2層構造対応）
	 */
	private drawArrow(
		style: (typeof TRANSITION_EDGE_STYLE)[keyof typeof TRANSITION_EDGE_STYLE],
		_widths: (typeof TRANSITION_EDGE_WIDTHS)[keyof typeof TRANSITION_EDGE_WIDTHS]
	): void {
		// 矢印サイズを縮小（12px → 8px）
		const arrowSize = 8;
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

		// Layer 1: 縁取り（外側の暗い部分）
		this.moveTo(tipX, tipY);
		this.lineTo(leftX, leftY);
		this.lineTo(rightX, rightY);
		this.closePath();
		this.fill(style.outer);

		// Layer 2: コア（内側の明るい部分、70%スケール）
		const innerScale = 0.7;
		const innerLeftX = tipX - arrowSize * innerScale * Math.cos(angle - arrowAngle);
		const innerLeftY = tipY - arrowSize * innerScale * Math.sin(angle - arrowAngle);
		const innerRightX = tipX - arrowSize * innerScale * Math.cos(angle + arrowAngle);
		const innerRightY = tipY - arrowSize * innerScale * Math.sin(angle + arrowAngle);

		this.moveTo(tipX, tipY);
		this.lineTo(innerLeftX, innerLeftY);
		this.lineTo(innerRightX, innerRightY);
		this.closePath();
		this.fill(style.core);
	}

	/**
	 * ガード条件ラベルを配置（バッジ風）
	 */
	private positionGuardText(): void {
		if (!this.guardContainer || !this.guardBackground || !this.guardText) return;

		// 線の中点に配置
		const midX = (this.sourcePoint.x + this.targetPoint.x) / 2;
		const midY = (this.sourcePoint.y + this.targetPoint.y) / 2;

		// 線から少しオフセット
		const dx = this.targetPoint.x - this.sourcePoint.x;
		const dy = this.targetPoint.y - this.sourcePoint.y;
		const angle = Math.atan2(dy, dx);
		const offset = 16;

		// テキストサイズに基づいて背景サイズを計算
		const textWidth = this.guardText.width;
		const textHeight = this.guardText.height;
		const bgWidth = textWidth + GUARD_LABEL_STYLE.paddingH * 2;
		const bgHeight = textHeight + GUARD_LABEL_STYLE.paddingV * 2;

		// 状態に応じたスタイルを取得
		const style = this.getGuardLabelStyle();

		// 背景を描画
		this.guardBackground.clear();

		// 背景（角丸矩形）
		this.guardBackground.roundRect(0, 0, bgWidth, bgHeight, GUARD_LABEL_STYLE.borderRadius);
		this.guardBackground.fill({
			color: style.background,
			alpha: GUARD_LABEL_STYLE.backgroundAlpha
		});
		this.guardBackground.stroke({ width: GUARD_LABEL_STYLE.borderWidth, color: style.border });

		// テキストを中央配置
		this.guardText.x = GUARD_LABEL_STYLE.paddingH;
		this.guardText.y = GUARD_LABEL_STYLE.paddingV;
		this.guardText.style.fill = style.text;

		// コンテナを線の法線方向にオフセットして中央配置
		this.guardContainer.x = midX + offset * Math.cos(angle + Math.PI / 2) - bgWidth / 2;
		this.guardContainer.y = midY + offset * Math.sin(angle + Math.PI / 2) - bgHeight / 2;
	}

	/**
	 * ガードラベルのスタイルを取得（状態に応じた色）
	 */
	private getGuardLabelStyle(): { background: number; border: number; text: number } {
		if (this.isSelected) {
			return {
				background: GUARD_LABEL_STYLE.selectedBackground,
				border: GUARD_LABEL_STYLE.selectedBorder,
				text: GUARD_LABEL_STYLE.selectedText
			};
		} else if (this.isHovered) {
			return {
				background: GUARD_LABEL_STYLE.hoverBackground,
				border: GUARD_LABEL_STYLE.hoverBorder,
				text: GUARD_LABEL_STYLE.hoverText
			};
		}
		return {
			background: GUARD_LABEL_STYLE.background,
			border: GUARD_LABEL_STYLE.border,
			text: GUARD_LABEL_STYLE.text
		};
	}

	/**
	 * エッジスタイルを取得（状態に応じた3層スタイル）
	 */
	private getEdgeStyle(): (typeof TRANSITION_EDGE_STYLE)[keyof typeof TRANSITION_EDGE_STYLE] {
		if (this.isSelected) {
			return TRANSITION_EDGE_STYLE.selected;
		} else if (this.isHovered) {
			return TRANSITION_EDGE_STYLE.hover;
		}
		return TRANSITION_EDGE_STYLE.normal;
	}

	/**
	 * エッジ幅を取得（状態に応じた幅）
	 */
	private getEdgeWidths(): (typeof TRANSITION_EDGE_WIDTHS)[keyof typeof TRANSITION_EDGE_WIDTHS] {
		if (this.isSelected) {
			return TRANSITION_EDGE_WIDTHS.selected;
		} else if (this.isHovered) {
			return TRANSITION_EDGE_WIDTHS.hover;
		}
		return TRANSITION_EDGE_WIDTHS.normal;
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
