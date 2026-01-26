// DiamondNode - UML アクティビティ図のひし形ノード基底クラス
// DecisionNode（分岐）と MergeNode（合流）の共通実装
import type { ActivityNodeItem } from '$lib/types/api';
import { ActivityNodeBase } from './ActivityNodeBase';
import { DECISION_NODE_SIZE, NODE_COLORS, COMMON_COLORS } from './constants';

/**
 * DiamondNode - ひし形ノード基底クラス
 *
 * UML 表記: ひし形
 * Decision（分岐）と Merge（合流）で共有される描画ロジックを提供
 */
export abstract class DiamondNode extends ActivityNodeBase {
	protected readonly size: number;

	constructor(nodeData: ActivityNodeItem) {
		super(nodeData);

		this.size = DECISION_NODE_SIZE.width;
	}

	/**
	 * ひし形を描画
	 */
	protected drawDiamond(): void {
		this.background.clear();

		const centerX = this.size / 2;
		const centerY = this.size / 2;
		const halfSize = this.size / 2;

		const bgColor = NODE_COLORS.decision.background;
		const borderColor = this.isSelected
			? COMMON_COLORS.borderSelected
			: this.isHovered
				? COMMON_COLORS.borderHover
				: NODE_COLORS.decision.border;
		const borderWidth = this.getBorderWidth();

		// グロー効果（選択/ホバー時）
		if (this.isSelected || this.isHovered) {
			const glowColor = this.isSelected ? COMMON_COLORS.borderSelected : COMMON_COLORS.borderHover;
			// ひし形グロー
			this.background.moveTo(centerX, centerY - halfSize - 5);
			this.background.lineTo(centerX + halfSize + 5, centerY);
			this.background.lineTo(centerX, centerY + halfSize + 5);
			this.background.lineTo(centerX - halfSize - 5, centerY);
			this.background.closePath();
			this.background.fill({ color: glowColor, alpha: 0.25 });
		}

		// 下部シャドウ（立体感）
		this.background.moveTo(centerX, centerY - halfSize + 2);
		this.background.lineTo(centerX + halfSize - 2, centerY);
		this.background.lineTo(centerX, centerY + halfSize - 2);
		this.background.lineTo(centerX - halfSize + 2, centerY);
		this.background.closePath();
		this.background.fill({ color: 0x000000, alpha: 0.15 });

		// メインひし形
		this.background.moveTo(centerX, centerY - halfSize);
		this.background.lineTo(centerX + halfSize, centerY);
		this.background.lineTo(centerX, centerY + halfSize);
		this.background.lineTo(centerX - halfSize, centerY);
		this.background.closePath();
		this.background.fill(bgColor);
		this.background.stroke({ width: borderWidth, color: borderColor });

		// 上部ハイライト（三角形部分）
		const highlightSize = halfSize * 0.5;
		this.background.moveTo(centerX, centerY - halfSize + 5);
		this.background.lineTo(centerX + highlightSize, centerY - 3);
		this.background.lineTo(centerX - highlightSize, centerY - 3);
		this.background.closePath();
		this.background.fill({ color: 0x555555, alpha: 0.2 });
	}

	/**
	 * ノード幅を取得
	 */
	getNodeWidth(): number {
		return this.size;
	}

	/**
	 * ノード高さを取得
	 */
	getNodeHeight(): number {
		return this.size;
	}
}
