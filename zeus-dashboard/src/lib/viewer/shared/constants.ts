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
 * エッジ色定義（2層構造: 外側縁取り → コア）
 * - シンプルで視認性重視の設計
 * - 背景 0x1a1a1a に対して高コントラストを確保
 *
 * デザイン原則:
 * - コア: 明るい色で視認性を確保
 * - 外側: 暗い縁取りでコアを際立たせる
 */
export const EDGE_COLORS = {
	normal: {
		core: 0xcccccc, // やや暗めのコア（204, 204, 204）
		outer: 0x555555 // 縁取り
	},
	critical: {
		core: 0xffaa44, // オレンジコア
		outer: 0x885522 // 暗いオレンジ縁取り
	},
	highlighted: {
		core: 0xffcc66, // 明るいオレンジ
		outer: 0xff9533 // アクセント色
	}
} as const;

/**
 * エッジ幅定義（2層構造）
 * - 細めの線でシンプルに
 * - 描画回数: 4回 → 2回（50%削減）
 * - 最大幅: 20px → 5px（75%削減）
 */
export const EDGE_WIDTHS = {
	normal: { core: 1.5, outer: 3 },
	critical: { core: 2, outer: 4 },
	highlighted: { core: 2.5, outer: 5 }
} as const;

export type EdgeColors = typeof EDGE_COLORS;
export type EdgeWidths = typeof EDGE_WIDTHS;
