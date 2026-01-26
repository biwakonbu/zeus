// InitialNode - UML アクティビティ図の開始ノード（黒丸）
import type { ActivityNodeItem } from '$lib/types/api';
import { ActivityNodeBase } from './ActivityNodeBase';
import { TERMINAL_NODE_SIZE, NODE_COLORS, COMMON_COLORS } from './constants';

/**
 * InitialNode - 開始ノード
 *
 * UML 表記: 塗りつぶされた黒丸
 * アクティビティの開始点を表す
 */
export class InitialNode extends ActivityNodeBase {
	private readonly radius: number;

	constructor(nodeData: ActivityNodeItem) {
		super(nodeData);

		this.radius = TERMINAL_NODE_SIZE.initialRadius;

		// 初回描画
		this.draw();
	}

	/**
	 * 開始ノードを描画
	 */
	draw(): void {
		this.background.clear();

		const centerX = this.radius;
		const centerY = this.radius;

		// グロー効果（選択/ホバー時）
		if (this.isSelected || this.isHovered) {
			const glowColor = this.isSelected ? COMMON_COLORS.borderSelected : COMMON_COLORS.borderHover;
			this.background.circle(centerX, centerY, this.radius + 4);
			this.background.fill({ color: glowColor, alpha: 0.2 });
		}

		// メイン円（塗りつぶし）
		this.background.circle(centerX, centerY, this.radius);
		this.background.fill(NODE_COLORS.initial.fill);
		this.background.stroke({ width: this.getBorderWidth(), color: this.getBorderColor() });
	}

	/**
	 * ノード幅を取得
	 */
	getNodeWidth(): number {
		return this.radius * 2;
	}

	/**
	 * ノード高さを取得
	 */
	getNodeHeight(): number {
		return this.radius * 2;
	}
}
