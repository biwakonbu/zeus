// DiamondNode - UML アクティビティ図のひし形ノード基底クラス
// DecisionNode（分岐）と MergeNode（合流）の共通実装
// 金属質感 + 3層グロー対応
import type { ActivityNodeItem } from '$lib/types/api';
import { ActivityNodeBase } from './ActivityNodeBase';
import { DECISION_NODE_SIZE, NODE_COLORS, COMMON_COLORS, METAL_EFFECT } from './constants';

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
	 * 5層ベベル構造 + 3層グローで金属質感を表現
	 *
	 * Layer 0-2: グロー（最外層/中間/内側）
	 * Layer 3: 外側シャドウ（下・右）
	 * Layer 4: メイン背景
	 * Layer 5: 内側ハイライト（上・左）
	 */
	protected drawDiamond(): void {
		this.background.clear();

		const centerX = this.size / 2;
		const centerY = this.size / 2;
		const halfSize = this.size / 2;

		const colors = NODE_COLORS.decision;
		const bgColor = colors.background;
		const borderColor = this.isSelected
			? COMMON_COLORS.borderSelected
			: this.isHovered
				? COMMON_COLORS.borderHover
				: colors.border;
		const borderWidth = this.getBorderWidth();

		// === 3層グロー ===

		// 最外層グロー（常時微弱グロー - Factorio らしさ）
		const baseGlowColor = 'baseGlow' in colors ? colors.baseGlow : COMMON_COLORS.accent;
		const baseGlowAlpha = 'baseGlowAlpha' in colors ? colors.baseGlowAlpha : 0.06;
		this.background.moveTo(centerX, centerY - halfSize - 10);
		this.background.lineTo(centerX + halfSize + 10, centerY);
		this.background.lineTo(centerX, centerY + halfSize + 10);
		this.background.lineTo(centerX - halfSize - 10, centerY);
		this.background.closePath();
		this.background.fill({ color: baseGlowColor, alpha: baseGlowAlpha });

		// 中間・内側グロー（ホバー/選択時に強化）
		if (this.isSelected || this.isHovered) {
			const glowColor = this.isSelected ? COMMON_COLORS.borderSelected : COMMON_COLORS.borderHover;
			const glowAlpha = this.isSelected
				? ('selectedGlowAlpha' in colors ? colors.selectedGlowAlpha : 0.4)
				: ('hoverGlowAlpha' in colors ? colors.hoverGlowAlpha : 0.25);

			// 中間グロー層
			this.background.moveTo(centerX, centerY - halfSize - 7);
			this.background.lineTo(centerX + halfSize + 7, centerY);
			this.background.lineTo(centerX, centerY + halfSize + 7);
			this.background.lineTo(centerX - halfSize - 7, centerY);
			this.background.closePath();
			this.background.fill({ color: glowColor, alpha: glowAlpha * 0.6 });

			// 内側グロー層
			this.background.moveTo(centerX, centerY - halfSize - 4);
			this.background.lineTo(centerX + halfSize + 4, centerY);
			this.background.lineTo(centerX, centerY + halfSize + 4);
			this.background.lineTo(centerX - halfSize - 4, centerY);
			this.background.closePath();
			this.background.fill({ color: glowColor, alpha: glowAlpha });
		}

		// === 5層ベベル構造 ===

		// Layer 1: 外側シャドウ（下・右）- 板金の影
		const shadowOffset = METAL_EFFECT.bevelWidth;
		this.background.moveTo(centerX + shadowOffset, centerY - halfSize + shadowOffset);
		this.background.lineTo(centerX + halfSize + shadowOffset, centerY + shadowOffset);
		this.background.lineTo(centerX + shadowOffset, centerY + halfSize + shadowOffset);
		this.background.lineTo(centerX - halfSize + shadowOffset, centerY + shadowOffset);
		this.background.closePath();
		this.background.fill({ color: 0x000000, alpha: METAL_EFFECT.outerShadowAlpha });

		// Layer 2: 内側シャドウ（下・右）- 対称性を維持
		const innerShadowOffset = METAL_EFFECT.innerBevelWidth;
		this.background.moveTo(centerX + innerShadowOffset, centerY - halfSize + innerShadowOffset);
		this.background.lineTo(centerX + halfSize - innerShadowOffset + innerShadowOffset, centerY + innerShadowOffset);
		this.background.lineTo(centerX + innerShadowOffset, centerY + halfSize - innerShadowOffset + innerShadowOffset);
		this.background.lineTo(centerX - halfSize + innerShadowOffset + innerShadowOffset, centerY + innerShadowOffset);
		this.background.closePath();
		this.background.fill({ color: 0x000000, alpha: METAL_EFFECT.innerShadowAlpha });

		// Layer 3: メインひし形
		this.background.moveTo(centerX, centerY - halfSize);
		this.background.lineTo(centerX + halfSize, centerY);
		this.background.lineTo(centerX, centerY + halfSize);
		this.background.lineTo(centerX - halfSize, centerY);
		this.background.closePath();
		this.background.fill(bgColor);
		this.background.stroke({ width: borderWidth, color: borderColor });

		// Layer 4: 上部ハイライト（金属光沢）- 上半分のひし形（対称係数）
		const highlightInset = 4;
		const highlightHalfSize = halfSize - highlightInset;
		this.background.moveTo(centerX, centerY - highlightHalfSize);
		this.background.lineTo(centerX + highlightHalfSize * 0.6, centerY - highlightHalfSize * 0.4);
		this.background.lineTo(centerX - highlightHalfSize * 0.6, centerY - highlightHalfSize * 0.4);
		this.background.closePath();
		const highlightColor = 'borderHighlight' in colors ? colors.borderHighlight : 0x666666;
		this.background.fill({ color: highlightColor, alpha: METAL_EFFECT.topHighlightAlpha });

		// Layer 5: 下部シャドウ（凹み感）- 下半分のひし形（対称係数）
		this.background.moveTo(centerX, centerY + highlightHalfSize);
		this.background.lineTo(centerX + highlightHalfSize * 0.6, centerY + highlightHalfSize * 0.4);
		this.background.lineTo(centerX - highlightHalfSize * 0.6, centerY + highlightHalfSize * 0.4);
		this.background.closePath();
		this.background.fill({ color: 0x000000, alpha: METAL_EFFECT.bottomShadowAlpha });

		// === アクセントライン（上部オレンジ）===
		if ('accentLine' in colors && 'accentLineAlpha' in colors) {
			// 上部の辺に沿ったアクセントライン
			const accentOffset = 3;
			this.background.moveTo(centerX - halfSize * 0.4, centerY - halfSize * 0.6 + accentOffset);
			this.background.lineTo(centerX, centerY - halfSize + accentOffset + 1);
			this.background.lineTo(centerX + halfSize * 0.4, centerY - halfSize * 0.6 + accentOffset);
			this.background.stroke({
				width: 1,
				color: colors.accentLine as number,
				alpha: colors.accentLineAlpha as number
			});
		}
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
