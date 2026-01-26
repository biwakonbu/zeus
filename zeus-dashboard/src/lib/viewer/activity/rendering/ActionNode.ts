// ActionNode - UML アクティビティ図のアクションノード（角丸四角形）
import { Text, TextStyle, CanvasTextMetrics } from 'pixi.js';
import type { ActivityNodeItem } from '$lib/types/api';
import { ActivityNodeBase } from './ActivityNodeBase';
import { ACTION_NODE_SIZE, NODE_COLORS, COMMON_COLORS, TEXT_RESOLUTION } from './constants';

/**
 * ActionNode - アクションノード
 *
 * UML 表記: 角丸四角形
 * 具体的なアクション/処理を表す
 */
export class ActionNode extends ActivityNodeBase {
	private nameText: Text;
	private nodeWidth: number;
	private nodeHeight: number;

	constructor(nodeData: ActivityNodeItem) {
		super(nodeData);

		// テキストコンポーネント初期化
		this.nameText = new Text({
			text: '',
			style: {
				fontSize: 11,
				fill: NODE_COLORS.action.text,
				fontFamily: 'IBM Plex Mono, monospace',
				align: 'center',
				wordWrap: true,
				wordWrapWidth: ACTION_NODE_SIZE.maxWidth - ACTION_NODE_SIZE.paddingH * 2
			},
			resolution: TEXT_RESOLUTION
		});
		this.addChild(this.nameText);

		// サイズ計算
		const size = this.calculateSize();
		this.nodeWidth = size.width;
		this.nodeHeight = size.height;

		// 初回描画
		this.draw();
	}

	/**
	 * テキスト量に応じたサイズを計算
	 */
	private calculateSize(): { width: number; height: number } {
		const name = this.nodeData.name || '';
		const style = new TextStyle({
			fontSize: 11,
			fontFamily: 'IBM Plex Mono, monospace'
		});
		const metrics = CanvasTextMetrics.measureText(name, style);

		const width = Math.min(
			ACTION_NODE_SIZE.maxWidth,
			Math.max(ACTION_NODE_SIZE.minWidth, metrics.width + ACTION_NODE_SIZE.paddingH * 2)
		);
		const height = Math.min(
			ACTION_NODE_SIZE.maxHeight,
			Math.max(ACTION_NODE_SIZE.minHeight, metrics.height + ACTION_NODE_SIZE.paddingV * 2)
		);

		return { width, height };
	}

	/**
	 * アクションノードを描画
	 */
	draw(): void {
		this.background.clear();

		const bgColor = this.getBackgroundColor();
		const borderColor = this.getBorderColor();
		const borderWidth = this.getBorderWidth();

		// グロー効果（選択/ホバー時）
		if (this.isSelected || this.isHovered) {
			const glowColor = this.isSelected ? COMMON_COLORS.borderSelected : COMMON_COLORS.borderHover;
			this.background.roundRect(
				-4,
				-4,
				this.nodeWidth + 8,
				this.nodeHeight + 8,
				ACTION_NODE_SIZE.borderRadius + 4
			);
			this.background.fill({ color: glowColor, alpha: 0.2 });
		}

		// メイン背景（角丸四角形）
		this.background.roundRect(0, 0, this.nodeWidth, this.nodeHeight, ACTION_NODE_SIZE.borderRadius);
		this.background.fill(bgColor);
		this.background.stroke({ width: borderWidth, color: borderColor });

		// 上部ハイライト（金属感）
		this.background.roundRect(
			4,
			4,
			this.nodeWidth - 8,
			this.nodeHeight / 3,
			ACTION_NODE_SIZE.borderRadius - 2
		);
		this.background.fill({ color: 0x555555, alpha: 0.2 });

		// テキスト描画
		this.drawText();
	}

	/**
	 * テキストを描画
	 */
	private drawText(): void {
		const name = this.nodeData.name || '';
		const maxWidth = this.nodeWidth - ACTION_NODE_SIZE.paddingH;
		let displayText = name;

		this.nameText.text = displayText;

		// テキストが最大幅を超える場合は切り詰め
		if (this.nameText.width > maxWidth) {
			while (this.nameText.width > maxWidth && displayText.length > 3) {
				displayText = displayText.substring(0, displayText.length - 1);
				this.nameText.text = displayText + '..';
			}
		}

		// 中央配置
		this.nameText.x = (this.nodeWidth - this.nameText.width) / 2;
		this.nameText.y = (this.nodeHeight - this.nameText.height) / 2;
	}

	/**
	 * ノードデータを更新
	 */
	updateNodeData(nodeData: ActivityNodeItem): void {
		super.updateNodeData(nodeData);
		// サイズ再計算
		const size = this.calculateSize();
		this.nodeWidth = size.width;
		this.nodeHeight = size.height;
	}

	/**
	 * ノード幅を取得
	 */
	getNodeWidth(): number {
		return this.nodeWidth;
	}

	/**
	 * ノード高さを取得
	 */
	getNodeHeight(): number {
		return this.nodeHeight;
	}
}
