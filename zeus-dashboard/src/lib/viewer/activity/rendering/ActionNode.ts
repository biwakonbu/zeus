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

		// テキストコンポーネント初期化（より明るいテキスト）
		this.nameText = new Text({
			text: '',
			style: {
				fontSize: 11,
				fill: NODE_COLORS.action.text, // 0xe0e0e0
				fontFamily: 'IBM Plex Mono, monospace',
				align: 'center',
				wordWrap: true,
				wordWrapWidth: ACTION_NODE_SIZE.maxWidth - ACTION_NODE_SIZE.paddingH * 2,
				// テキストにも軽いシャドウ（PixiJS v8 形式）
				dropShadow: {
					color: 0x000000,
					alpha: 0.3,
					distance: 1,
					blur: 1
				}
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

		const bgColor = NODE_COLORS.action.background;
		const borderColor = this.getBorderColor();
		const borderWidth = this.getBorderWidth();
		const borderRadius = ACTION_NODE_SIZE.borderRadius;

		// グロー効果（選択/ホバー時）
		if (this.isSelected || this.isHovered) {
			const glowColor = this.isSelected ? COMMON_COLORS.borderSelected : COMMON_COLORS.borderHover;
			this.background.roundRect(
				-4,
				-4,
				this.nodeWidth + 8,
				this.nodeHeight + 8,
				borderRadius + 4
			);
			this.background.fill({ color: glowColor, alpha: 0.25 });
		}

		// インナーシャドウ効果（外側に暗い縁）
		this.background.roundRect(2, 2, this.nodeWidth - 4, this.nodeHeight - 4, borderRadius - 1);
		this.background.fill({ color: 0x000000, alpha: 0.15 });

		// メイン背景（角丸四角形）
		this.background.roundRect(0, 0, this.nodeWidth, this.nodeHeight, borderRadius);
		this.background.fill(bgColor);
		this.background.stroke({ width: borderWidth, color: borderColor });

		// 上部ハイライト（金属光沢）- より強く
		this.background.roundRect(
			3,
			3,
			this.nodeWidth - 6,
			this.nodeHeight * 0.35,
			borderRadius - 2
		);
		this.background.fill({ color: NODE_COLORS.action.borderHighlight, alpha: 0.25 });

		// 下部シャドウ（凹み感）
		this.background.roundRect(
			3,
			this.nodeHeight * 0.65,
			this.nodeWidth - 6,
			this.nodeHeight * 0.3,
			borderRadius - 2
		);
		this.background.fill({ color: 0x000000, alpha: 0.12 });

		// 上部ボーダーハイライト（金属の縁）
		this.background.moveTo(borderRadius, 1);
		this.background.lineTo(this.nodeWidth - borderRadius, 1);
		this.background.stroke({ width: 1, color: NODE_COLORS.action.borderHighlight, alpha: 0.4 });

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
