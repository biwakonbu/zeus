// SyncBarNode - UML アクティビティ図の同期バー基底クラス
// ForkNode（並列分岐）と JoinNode（並列合流）の共通実装
import type { ActivityNodeItem } from '$lib/types/api';
import { ActivityNodeBase } from './ActivityNodeBase';
import { FORK_NODE_SIZE, NODE_COLORS, COMMON_COLORS, METAL_EFFECT } from './constants';

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
	 * 5層ベベル構造 + 3層グローで金属質感を表現
	 *
	 * Layer 0-2: グロー（最外層/中間/内側）
	 * Layer 3: 外側シャドウ（下・右）
	 * Layer 4: メイン背景
	 * Layer 5: 内側ハイライト（上・左）
	 */
	draw(): void {
		this.background.clear();

		const colors = NODE_COLORS.fork;
		const borderColor = this.isSelected
			? COMMON_COLORS.borderSelected
			: this.isHovered
				? COMMON_COLORS.borderHover
				: colors.border;

		// === 3層グロー ===

		// 最外層グロー（常時微弱グロー - Factorio らしさ）
		const baseGlowColor = 'baseGlow' in colors ? colors.baseGlow : COMMON_COLORS.accent;
		const baseGlowAlpha = 'baseGlowAlpha' in colors ? colors.baseGlowAlpha : 0.06;
		this.background.roundRect(-10, -8, this.barWidth + 20, this.barHeight + 16, 6);
		this.background.fill({ color: baseGlowColor, alpha: baseGlowAlpha });

		// 中間・内側グロー（ホバー/選択時に強化）
		if (this.isSelected || this.isHovered) {
			const glowColor = this.isSelected ? COMMON_COLORS.borderSelected : COMMON_COLORS.borderHover;
			const glowAlpha = this.isSelected
				? 'selectedGlowAlpha' in colors
					? colors.selectedGlowAlpha
					: 0.4
				: 'hoverGlowAlpha' in colors
					? colors.hoverGlowAlpha
					: 0.25;

			// 中間グロー層
			this.background.roundRect(-7, -6, this.barWidth + 14, this.barHeight + 12, 5);
			this.background.fill({ color: glowColor, alpha: glowAlpha * 0.6 });

			// 内側グロー層
			this.background.roundRect(-4, -4, this.barWidth + 8, this.barHeight + 8, 4);
			this.background.fill({ color: glowColor, alpha: glowAlpha });
		}

		// === 5層ベベル構造 ===

		// Layer 1: 外側シャドウ（下・右）- 板金の影
		this.background.roundRect(
			METAL_EFFECT.bevelWidth,
			METAL_EFFECT.bevelWidth,
			this.barWidth,
			this.barHeight,
			2
		);
		this.background.fill({ color: 0x000000, alpha: METAL_EFFECT.outerShadowAlpha });

		// Layer 2: 内側シャドウ（下・右）
		this.background.roundRect(
			METAL_EFFECT.innerBevelWidth,
			METAL_EFFECT.innerBevelWidth,
			this.barWidth - METAL_EFFECT.innerBevelWidth,
			this.barHeight - METAL_EFFECT.innerBevelWidth,
			2
		);
		this.background.fill({ color: 0x000000, alpha: METAL_EFFECT.innerShadowAlpha });

		// Layer 3: メインバー（太い横線）
		this.background.roundRect(0, 0, this.barWidth, this.barHeight, 2);
		this.background.fill(colors.fill);
		this.background.stroke({ width: 1, color: borderColor });

		// Layer 4: 上部ハイライト（金属光沢）
		this.background.roundRect(
			2,
			1,
			this.barWidth - 4,
			this.barHeight * METAL_EFFECT.topHighlightRatio,
			1
		);
		this.background.fill({ color: colors.highlight, alpha: METAL_EFFECT.topHighlightAlpha });

		// Layer 5: 下部シャドウ（凹み感）
		this.background.roundRect(
			2,
			this.barHeight * (1 - METAL_EFFECT.bottomShadowRatio),
			this.barWidth - 4,
			this.barHeight * METAL_EFFECT.bottomShadowRatio - 1,
			1
		);
		this.background.fill({ color: 0x000000, alpha: METAL_EFFECT.bottomShadowAlpha });

		// === 左右端のアクセント（金属パネルの縁）- 幅を縮小してバランス改善 ===
		// 左端ハイライト（2px に縮小）
		this.background.roundRect(0, 0, 2, this.barHeight, 1);
		this.background.fill({
			color: colors.highlight,
			alpha: METAL_EFFECT.outerHighlightAlpha * 0.7
		});

		// 右端シャドウ（2px に縮小）
		this.background.roundRect(this.barWidth - 2, 0, 2, this.barHeight, 1);
		this.background.fill({ color: 0x000000, alpha: METAL_EFFECT.bottomShadowAlpha * 0.8 });

		// === アクセントライン（上部オレンジ）===
		if ('accentLine' in colors && 'accentLineAlpha' in colors) {
			// 上部中央にアクセントライン
			this.background.moveTo(this.barWidth * 0.2, 1);
			this.background.lineTo(this.barWidth * 0.8, 1);
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
		return this.barWidth;
	}

	/**
	 * ノード高さを取得
	 */
	getNodeHeight(): number {
		return this.barHeight;
	}
}
