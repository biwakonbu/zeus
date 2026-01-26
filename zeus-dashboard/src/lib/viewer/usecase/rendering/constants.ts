// UseCase ビューワー共通定数
// TEXT_RESOLUTION と共通色定義を一元管理

// テキスト解像度（Retina対応）
export const TEXT_RESOLUTION = typeof window !== 'undefined'
	? Math.min(window.devicePixelRatio * 2, 4)
	: 2;

// ユースケースサイズ制約（テキスト量に応じた自動調整用）
export const USECASE_SIZE = {
	minWidth: 120,
	minHeight: 50,
	maxWidth: 220,
	maxHeight: 70,
	paddingH: 24,
	paddingV: 16
};

// ステータス別スタイル（背景色によるステータス表現）
export const USECASE_STATUS_STYLES = {
	active: {
		background: 0x2d3d2d,    // 暗めの緑
		border: 0x44cc44,        // 緑
		glowAlpha: 0.12
	},
	draft: {
		background: 0x3d3520,    // 暗めの黄
		border: 0xccaa00,        // 黄
		glowAlpha: 0.08
	},
	deprecated: {
		background: 0x2a2a2a,    // 暗めのグレー
		border: 0x555555,        // グレー
		glowAlpha: 0
	}
};

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
