// UseCase ビューワー共通定数
// TEXT_RESOLUTION と共通色定義を一元管理

// テキスト解像度（Retina対応）
export const TEXT_RESOLUTION = typeof window !== 'undefined'
	? Math.min(window.devicePixelRatio * 2, 4)
	: 2;

// 共通色定義（Factorio テーマ準拠）
export const COMMON_COLORS = {
	// テキスト色
	text: 0xe0e0e0,
	textMuted: 0x888888,

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

	// アクセント色
	accent: 0xff9533,
	highlighted: 0xff9533
};
