// ビューワー共通定数定義
// Activity/UseCase 等のビューワー間で共有する定数

/**
 * テキスト解像度（高DPI/Retina対応）
 * - ブラウザ環境では devicePixelRatio を考慮
 * - SSR 環境ではデフォルト値 2 を使用
 * - 最大値を 4 に制限してパフォーマンスを維持
 */
export const TEXT_RESOLUTION =
	typeof window !== 'undefined' ? Math.min(window.devicePixelRatio * 2, 4) : 2;

/**
 * 共通色定義（Factorio テーマ準拠）
 * - PixiJS 用の 0x 形式で定義
 * - design-tokens.ts の色と整合性を維持
 */
export const COMMON_COLORS = {
	// テキスト色
	text: 0xe0e0e0,
	textMuted: 0x888888,
	textDark: 0x1a1a1a,

	// 背景色
	background: 0x1a1a1a,
	backgroundPanel: 0x2d2d2d,
	backgroundHover: 0x3a3a3a,
	backgroundSelected: 0x4a4a4a,

	// ボーダー色
	border: 0x4a4a4a,
	borderHighlight: 0x666666,
	borderHover: 0xff9533,
	borderSelected: 0xff9533,

	// アクセント色（Factorio オレンジ）
	accent: 0xff9533,
	highlighted: 0xff9533,

	// 状態色
	accentGreen: 0x22c55e, // 成功
	accentRed: 0xef4444, // エラー/停止
	accentBlue: 0x3b82f6 // 情報
} as const;

export type CommonColors = typeof COMMON_COLORS;

/**
 * エッジ色定義（3層構造: グロー → 外側 → コア）
 * - 電気回路風の発光感を表現
 * - 背景 0x1a1a1a に対して高コントラストを確保
 *
 * デザイン原則:
 * - コア: 明るい白〜淡いオレンジで視認性を確保
 * - 外側: 暗い縁取りでコアを際立たせる
 * - グロー: 淡いハロー効果で電気配線感を演出
 */
export const EDGE_COLORS = {
	normal: {
		core: 0xcccccc, // 明るいコア（204, 204, 204）- 背景との差 178
		outer: 0x2a2a2a, // 暗い縁取り - 背景より少し明るい程度でコアを際立たせる
		glow: 0x666666, // グロー色
		glowAlpha: 0.5 // グロー強度を上げる
	},
	critical: {
		core: 0xffbb66, // 明るいオレンジ
		outer: 0x4a2a00, // 暗い縁取り
		glow: 0xff9533,
		glowAlpha: 0.6
	},
	blocked: {
		core: 0xff8888, // 明るい赤
		outer: 0x4a1a1a, // 暗い縁取り
		glow: 0xff4444,
		glowAlpha: 0.5
	},
	highlighted: {
		core: 0xffcc88, // 明るいオレンジ
		outer: 0xff9533, // アクセント色の縁取り
		glow: 0xff9533,
		glowAlpha: 0.7
	}
} as const;

/**
 * エッジ幅定義（内側と外側）
 * - 太めの線で視認性を確保
 */
export const EDGE_WIDTHS = {
	normal: { core: 2.5, outer: 6 },
	critical: { core: 3, outer: 7 },
	blocked: { core: 2.5, outer: 6 },
	highlighted: { core: 3.5, outer: 8 }
} as const;

export type EdgeColors = typeof EDGE_COLORS;
export type EdgeWidths = typeof EDGE_WIDTHS;
