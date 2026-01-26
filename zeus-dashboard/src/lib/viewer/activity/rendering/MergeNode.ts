// MergeNode - UML アクティビティ図の合流ノード（ひし形）
import type { ActivityNodeItem } from '$lib/types/api';
import { ActivityNodeBase } from './ActivityNodeBase';
import { DECISION_NODE_SIZE, COMMON_COLORS } from './constants';

/**
 * MergeNode - 合流ノード
 *
 * UML 表記: ひし形（Decision と同じ形状）
 * 複数の制御フローが合流する点を表す
 * Decision と視覚的には同じだが、意味的に区別される
 */
export class MergeNode extends ActivityNodeBase {
	private readonly size: number;

	constructor(nodeData: ActivityNodeItem) {
		super(nodeData);

		this.size = DECISION_NODE_SIZE.width;

		// 初回描画
		this.draw();
	}

	/**
	 * 合流ノードを描画
	 */
	draw(): void {
		this.background.clear();

		const centerX = this.size / 2;
		const centerY = this.size / 2;
		const halfSize = this.size / 2;

		const bgColor = this.getBackgroundColor();
		const borderColor = this.getBorderColor();
		const borderWidth = this.getBorderWidth();

		// グロー効果（選択/ホバー時）
		if (this.isSelected || this.isHovered) {
			const glowColor = this.isSelected ? COMMON_COLORS.borderSelected : COMMON_COLORS.borderHover;
			this.background.moveTo(centerX, centerY - halfSize - 4);
			this.background.lineTo(centerX + halfSize + 4, centerY);
			this.background.lineTo(centerX, centerY + halfSize + 4);
			this.background.lineTo(centerX - halfSize - 4, centerY);
			this.background.closePath();
			this.background.fill({ color: glowColor, alpha: 0.2 });
		}

		// メインひし形
		this.background.moveTo(centerX, centerY - halfSize);
		this.background.lineTo(centerX + halfSize, centerY);
		this.background.lineTo(centerX, centerY + halfSize);
		this.background.lineTo(centerX - halfSize, centerY);
		this.background.closePath();
		this.background.fill(bgColor);
		this.background.stroke({ width: borderWidth, color: borderColor });
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
