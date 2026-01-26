// FinalNode - UML アクティビティ図の終了ノード（二重丸）
import type { ActivityNodeItem } from '$lib/types/api';
import { ActivityNodeBase } from './ActivityNodeBase';
import { TERMINAL_NODE_SIZE, NODE_COLORS, COMMON_COLORS } from './constants';

/**
 * FinalNode - 終了ノード
 *
 * UML 表記: 二重丸（外側の円と内側の塗りつぶされた円）
 * アクティビティの終了点を表す
 */
export class FinalNode extends ActivityNodeBase {
	private readonly outerRadius: number;
	private readonly innerRadius: number;

	constructor(nodeData: ActivityNodeItem) {
		super(nodeData);

		this.outerRadius = TERMINAL_NODE_SIZE.finalOuterRadius;
		this.innerRadius = TERMINAL_NODE_SIZE.finalInnerRadius;

		// 初回描画
		this.draw();
	}

	/**
	 * 終了ノードを描画
	 */
	draw(): void {
		this.background.clear();

		const centerX = this.outerRadius;
		const centerY = this.outerRadius;

		// グロー効果（選択/ホバー時）
		if (this.isSelected || this.isHovered) {
			const glowColor = this.isSelected ? COMMON_COLORS.borderSelected : COMMON_COLORS.borderHover;
			this.background.circle(centerX, centerY, this.outerRadius + 4);
			this.background.fill({ color: glowColor, alpha: 0.2 });
		}

		// 外側の円（輪郭のみ）
		this.background.circle(centerX, centerY, this.outerRadius);
		this.background.stroke({ width: this.getBorderWidth(), color: this.getBorderColor() });

		// 内側の円（塗りつぶし）
		this.background.circle(centerX, centerY, this.innerRadius);
		this.background.fill(NODE_COLORS.final.innerFill);
	}

	/**
	 * ノード幅を取得
	 */
	getNodeWidth(): number {
		return this.outerRadius * 2;
	}

	/**
	 * ノード高さを取得
	 */
	getNodeHeight(): number {
		return this.outerRadius * 2;
	}
}
