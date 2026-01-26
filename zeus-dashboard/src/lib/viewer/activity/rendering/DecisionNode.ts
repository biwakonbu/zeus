// DecisionNode - UML アクティビティ図の分岐ノード（ひし形）
import { Text } from 'pixi.js';
import type { ActivityNodeItem } from '$lib/types/api';
import { ActivityNodeBase } from './ActivityNodeBase';
import { DECISION_NODE_SIZE, NODE_COLORS, COMMON_COLORS, TEXT_RESOLUTION } from './constants';

/**
 * DecisionNode - 分岐ノード
 *
 * UML 表記: ひし形
 * 条件分岐点を表す
 */
export class DecisionNode extends ActivityNodeBase {
	private nameText: Text | null = null;
	private readonly size: number;

	constructor(nodeData: ActivityNodeItem) {
		super(nodeData);

		this.size = DECISION_NODE_SIZE.width;

		// 名前がある場合はテキストコンポーネント作成
		if (nodeData.name) {
			this.nameText = new Text({
				text: '',
				style: {
					fontSize: 9,
					fill: NODE_COLORS.decision.text,
					fontFamily: 'IBM Plex Mono, monospace',
					align: 'center'
				},
				resolution: TEXT_RESOLUTION
			});
			this.addChild(this.nameText);
		}

		// 初回描画
		this.draw();
	}

	/**
	 * 分岐ノードを描画
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
			// ひし形グロー
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

		// テキスト描画
		this.drawText();
	}

	/**
	 * テキストを描画
	 */
	private drawText(): void {
		if (!this.nameText || !this.nodeData.name) return;

		// 短縮表示（?マーク等）
		const displayText = this.nodeData.name.length > 5 ? '?' : this.nodeData.name;
		this.nameText.text = displayText;

		// 中央配置
		this.nameText.x = (this.size - this.nameText.width) / 2;
		this.nameText.y = (this.size - this.nameText.height) / 2;
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
