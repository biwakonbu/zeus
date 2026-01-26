// ForkNode - UML アクティビティ図の並列分岐ノード（太い横線）
import type { ActivityNodeItem } from '$lib/types/api';
import { ActivityNodeBase } from './ActivityNodeBase';
import { FORK_NODE_SIZE, NODE_COLORS, COMMON_COLORS } from './constants';

/**
 * ForkNode - 並列分岐ノード
 *
 * UML 表記: 太い横線（同期バー）
 * 1つの制御フローが複数の並列フローに分岐する点を表す
 */
export class ForkNode extends ActivityNodeBase {
	private readonly barWidth: number;
	private readonly barHeight: number;

	constructor(nodeData: ActivityNodeItem) {
		super(nodeData);

		this.barWidth = FORK_NODE_SIZE.width;
		this.barHeight = FORK_NODE_SIZE.height;

		// 初回描画
		this.draw();
	}

	/**
	 * 並列分岐ノードを描画
	 */
	draw(): void {
		this.background.clear();

		const borderColor = this.getBorderColor();

		// グロー効果（選択/ホバー時）
		if (this.isSelected || this.isHovered) {
			const glowColor = this.isSelected ? COMMON_COLORS.borderSelected : COMMON_COLORS.borderHover;
			this.background.roundRect(-4, -4, this.barWidth + 8, this.barHeight + 8, 4);
			this.background.fill({ color: glowColor, alpha: 0.2 });
		}

		// メインバー（太い横線）
		this.background.roundRect(0, 0, this.barWidth, this.barHeight, 2);
		this.background.fill(NODE_COLORS.fork.fill);
		this.background.stroke({ width: 1, color: borderColor });
	}

	/**
	 * ノード幅を取得
	 */
	getNodeWidth(): number {
		return this.barWidth;
	}

	/**
	 * ノード高さを取得
	 */
	getNodeHeight(): number {
		return this.barHeight;
	}
}
