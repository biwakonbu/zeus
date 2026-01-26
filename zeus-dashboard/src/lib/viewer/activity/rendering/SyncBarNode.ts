// SyncBarNode - UML アクティビティ図の同期バー基底クラス
// ForkNode（並列分岐）と JoinNode（並列合流）の共通実装
import type { ActivityNodeItem } from '$lib/types/api';
import { ActivityNodeBase } from './ActivityNodeBase';
import { FORK_NODE_SIZE, NODE_COLORS, COMMON_COLORS } from './constants';

/**
 * SyncBarNode - 同期バー基底クラス
 *
 * UML 表記: 太い横線（同期バー）
 * Fork（分岐）と Join（合流）で共有される描画ロジックを提供
 */
export abstract class SyncBarNode extends ActivityNodeBase {
	protected readonly barWidth: number;
	protected readonly barHeight: number;

	constructor(nodeData: ActivityNodeItem) {
		super(nodeData);

		this.barWidth = FORK_NODE_SIZE.width;
		this.barHeight = FORK_NODE_SIZE.height;

		// 初回描画
		this.draw();
	}

	/**
	 * 同期バーを描画
	 */
	draw(): void {
		this.background.clear();

		const borderColor = this.isSelected
			? COMMON_COLORS.borderSelected
			: this.isHovered
				? COMMON_COLORS.borderHover
				: NODE_COLORS.fork.border;

		// グロー効果（選択/ホバー時）
		if (this.isSelected || this.isHovered) {
			const glowColor = this.isSelected ? COMMON_COLORS.borderSelected : COMMON_COLORS.borderHover;
			this.background.roundRect(-5, -5, this.barWidth + 10, this.barHeight + 10, 5);
			this.background.fill({ color: glowColor, alpha: 0.25 });
		}

		// 下部シャドウ
		this.background.roundRect(2, 2, this.barWidth - 4, this.barHeight, 2);
		this.background.fill({ color: 0x000000, alpha: 0.2 });

		// メインバー（太い横線）
		this.background.roundRect(0, 0, this.barWidth, this.barHeight, 2);
		this.background.fill(NODE_COLORS.fork.fill);
		this.background.stroke({ width: 1, color: borderColor });

		// 上部ハイライト（金属光沢）
		this.background.roundRect(2, 1, this.barWidth - 4, this.barHeight * 0.4, 1);
		this.background.fill({ color: NODE_COLORS.fork.highlight, alpha: 0.3 });

		// 左右の端にアクセント
		this.background.roundRect(0, 0, 3, this.barHeight, 1);
		this.background.fill({ color: NODE_COLORS.fork.highlight, alpha: 0.2 });
		this.background.roundRect(this.barWidth - 3, 0, 3, this.barHeight, 1);
		this.background.fill({ color: NODE_COLORS.fork.highlight, alpha: 0.2 });
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
