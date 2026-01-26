// ActionNode - UML アクティビティ図のアクションノード（角丸四角形）
// 5層ベベル構造 + 3層グローで金属質感を表現
import { Text, TextStyle, CanvasTextMetrics } from 'pixi.js';
import type { ActivityNodeItem } from '$lib/types/api';
import { ActivityNodeBase } from './ActivityNodeBase';
import { ACTION_NODE_SIZE, NODE_COLORS, COMMON_COLORS, TEXT_RESOLUTION, METAL_EFFECT } from './constants';

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
				wordWrapWidth: 500, // テキスト全文表示のため十分な幅
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

		const width = Math.max(ACTION_NODE_SIZE.minWidth, metrics.width + ACTION_NODE_SIZE.paddingH * 2);
		const height = Math.min(
			ACTION_NODE_SIZE.maxHeight,
			Math.max(ACTION_NODE_SIZE.minHeight, metrics.height + ACTION_NODE_SIZE.paddingV * 2)
		);

		return { width, height };
	}

	/**
	 * アクションノードを描画
	 * 5層ベベル構造 + 3層グローで金属質感を表現
	 *
	 * Layer 0-2: グロー（最外層/中間/内側）
	 * Layer 3: 外側シャドウ（下・右）
	 * Layer 4: メイン背景
	 * Layer 5: 内側ハイライト（上・左）
	 * Layer 6: 外側ハイライト（上部ボーダー）
	 */
	draw(): void {
		this.background.clear();

		const colors = NODE_COLORS.action;
		const bgColor = colors.background;
		const borderColor = this.getBorderColor();
		const borderWidth = this.getBorderWidth();
		const borderRadius = ACTION_NODE_SIZE.borderRadius;

		// === 3層グロー ===

		// 最外層グロー（常時微弱グロー - Factorio らしさ）
		const baseGlowColor = 'baseGlow' in colors ? colors.baseGlow : COMMON_COLORS.accent;
		const baseGlowAlpha = 'baseGlowAlpha' in colors ? colors.baseGlowAlpha : 0.06;
		this.background.roundRect(
			-8,
			-8,
			this.nodeWidth + 16,
			this.nodeHeight + 16,
			borderRadius + 8
		);
		this.background.fill({ color: baseGlowColor, alpha: baseGlowAlpha });

		// 中間グロー（ホバー/選択時に強化）
		if (this.isSelected || this.isHovered) {
			const glowColor = this.isSelected ? COMMON_COLORS.borderSelected : COMMON_COLORS.borderHover;
			const glowAlpha = this.isSelected
				? ('selectedGlowAlpha' in colors ? colors.selectedGlowAlpha : 0.4)
				: ('hoverGlowAlpha' in colors ? colors.hoverGlowAlpha : 0.25);

			// 中間グロー層
			this.background.roundRect(
				-6,
				-6,
				this.nodeWidth + 12,
				this.nodeHeight + 12,
				borderRadius + 6
			);
			this.background.fill({ color: glowColor, alpha: glowAlpha * 0.6 });

			// 内側グロー層
			this.background.roundRect(
				-3,
				-3,
				this.nodeWidth + 6,
				this.nodeHeight + 6,
				borderRadius + 3
			);
			this.background.fill({ color: glowColor, alpha: glowAlpha });
		}

		// === 5層ベベル構造 ===

		// Layer 1: 外側シャドウ（下・右）- 板金の影
		const shadowColor = 'borderShadow' in colors ? colors.borderShadow : 0x1a1a1a;
		this.background.roundRect(
			METAL_EFFECT.bevelWidth,
			METAL_EFFECT.bevelWidth,
			this.nodeWidth,
			this.nodeHeight,
			borderRadius
		);
		this.background.fill({ color: shadowColor, alpha: METAL_EFFECT.outerShadowAlpha });

		// Layer 2: 内側シャドウ（下・右）
		this.background.roundRect(
			METAL_EFFECT.innerBevelWidth,
			METAL_EFFECT.innerBevelWidth,
			this.nodeWidth - METAL_EFFECT.innerBevelWidth,
			this.nodeHeight - METAL_EFFECT.innerBevelWidth,
			borderRadius - 1
		);
		this.background.fill({ color: 0x000000, alpha: METAL_EFFECT.innerShadowAlpha });

		// Layer 3: メイン背景（角丸四角形）
		this.background.roundRect(0, 0, this.nodeWidth, this.nodeHeight, borderRadius);
		this.background.fill(bgColor);
		this.background.stroke({ width: borderWidth, color: borderColor });

		// Layer 4: 内側ハイライト（上・左）- 金属光沢
		this.background.roundRect(
			3,
			3,
			this.nodeWidth - 6,
			this.nodeHeight * METAL_EFFECT.topHighlightRatio,
			borderRadius - 2
		);
		this.background.fill({ color: colors.borderHighlight, alpha: METAL_EFFECT.topHighlightAlpha });

		// Layer 5: 外側ハイライト（上部ボーダー）
		this.background.moveTo(borderRadius, 1);
		this.background.lineTo(this.nodeWidth - borderRadius, 1);
		this.background.stroke({ width: METAL_EFFECT.bevelWidth, color: colors.borderHighlight, alpha: METAL_EFFECT.outerHighlightAlpha });

		// 下部シャドウ（凹み感）
		this.background.roundRect(
			3,
			this.nodeHeight * (1 - METAL_EFFECT.bottomShadowRatio),
			this.nodeWidth - 6,
			this.nodeHeight * METAL_EFFECT.bottomShadowRatio - 3,
			borderRadius - 2
		);
		this.background.fill({ color: 0x000000, alpha: METAL_EFFECT.bottomShadowAlpha });

		// === アクセントライン（上部オレンジ）===
		if ('accentLine' in colors && 'accentLineAlpha' in colors) {
			this.background.moveTo(borderRadius + 4, 2);
			this.background.lineTo(this.nodeWidth - borderRadius - 4, 2);
			this.background.stroke({ width: 1, color: colors.accentLine, alpha: colors.accentLineAlpha });
		}

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
