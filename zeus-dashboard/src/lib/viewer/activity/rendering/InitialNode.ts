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

		// 常時グロー効果（開始点を強調）
		// 外側から内側へ段階的なグロー
		for (let i = 3; i >= 1; i--) {
			this.background.circle(centerX, centerY, this.radius + i * 3);
			this.background.fill({
				color: NODE_COLORS.initial.glow,
				alpha: NODE_COLORS.initial.glowAlpha / (i + 1)
			});
		}

		// 選択/ホバー時の追加グロー
		if (this.isSelected || this.isHovered) {
			const glowColor = this.isSelected ? COMMON_COLORS.borderSelected : COMMON_COLORS.borderHover;
			this.background.circle(centerX, centerY, this.radius + 6);
			this.background.fill({ color: glowColor, alpha: 0.3 });
		}

		// メイン円（塗りつぶし）
		this.background.circle(centerX, centerY, this.radius);
		this.background.fill(NODE_COLORS.initial.fill);
		this.background.stroke({ width: this.getBorderWidth(), color: this.getBorderColor() });

		// 内側ハイライト（金属感）
		this.background.circle(centerX - 2, centerY - 2, this.radius * 0.4);
		this.background.fill({ color: 0x444444, alpha: 0.4 });
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
