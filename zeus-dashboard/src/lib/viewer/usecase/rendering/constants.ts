// UseCase ビューワー共通定数

// 共通定数を共有ファイルからインポート
export { TEXT_RESOLUTION, COMMON_COLORS } from '$lib/viewer/shared/constants';

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
		background: 0x2d3d2d, // 暗めの緑
		border: 0x44cc44, // 緑
		glowAlpha: 0.12
	},
	draft: {
		background: 0x3d3520, // 暗めの黄
		border: 0xccaa00, // 黄
		glowAlpha: 0.08
	},
	deprecated: {
		background: 0x2a2a2a, // 暗めのグレー
		border: 0x555555, // グレー
		glowAlpha: 0
	}
};
