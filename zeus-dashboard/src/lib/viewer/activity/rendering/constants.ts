// Activity 図描画用の定数定義
// UML 2.5 準拠のアクティビティ図ノードスタイル

// テキスト解像度（高DPI対応）
export const TEXT_RESOLUTION = 2;

// 共通カラー（Factorio テーマ準拠）
export const COMMON_COLORS = {
	// 背景色
	background: 0x2a2a2a,
	backgroundHover: 0x3a3a3a,
	backgroundSelected: 0x4a4a4a,

	// ボーダー色
	border: 0x555555,
	borderHover: 0x888888,
	borderSelected: 0xf59e0b,

	// テキスト色
	text: 0xcccccc,
	textMuted: 0x888888,
	textDark: 0x1a1a1a,

	// アクセントカラー
	accent: 0xf59e0b, // Factorio オレンジ
	accentGreen: 0x22c55e, // 成功
	accentRed: 0xef4444, // エラー/停止
	accentBlue: 0x3b82f6 // 情報
};

// 初期/終了ノードサイズ
export const TERMINAL_NODE_SIZE = {
	// 初期ノード（黒丸）
	initialRadius: 12,
	// 終了ノード（二重丸）
	finalOuterRadius: 14,
	finalInnerRadius: 8
};

// アクションノードサイズ
export const ACTION_NODE_SIZE = {
	minWidth: 100,
	maxWidth: 180,
	minHeight: 40,
	maxHeight: 60,
	paddingH: 16,
	paddingV: 10,
	borderRadius: 8 // 角丸
};

// 分岐/合流ノードサイズ（ひし形）
export const DECISION_NODE_SIZE = {
	width: 50,
	height: 50
};

// フォーク/ジョインノードサイズ（太い線）
export const FORK_NODE_SIZE = {
	width: 100,
	height: 6 // 太さ
};

// 遷移エッジスタイル
export const TRANSITION_STYLE = {
	lineWidth: 2,
	arrowSize: 10,
	guardFontSize: 10,
	guardPadding: 4
};

// ノードタイプ別カラー
export const NODE_COLORS = {
	// 初期ノード - 黒丸
	initial: {
		fill: 0x1a1a1a,
		border: 0x555555
	},
	// 終了ノード - 二重丸
	final: {
		fill: 0x1a1a1a,
		border: 0x555555,
		innerFill: 0x1a1a1a
	},
	// アクションノード - 角丸四角形
	action: {
		background: 0x2a2a2a,
		border: 0x555555,
		text: 0xcccccc
	},
	// 分岐/合流ノード - ひし形
	decision: {
		background: 0x2a2a2a,
		border: 0x555555,
		text: 0xcccccc
	},
	// フォーク/ジョインノード - 太い線
	fork: {
		fill: 0x1a1a1a,
		border: 0x555555
	}
};

// ステータス別スタイル
export const ACTIVITY_STATUS_STYLES = {
	draft: {
		background: 0x333333,
		border: 0x555555,
		glowAlpha: 0
	},
	active: {
		background: 0x2a3d2a,
		border: 0x4a7c4a,
		glowAlpha: 0.1
	},
	deprecated: {
		background: 0x3d2a2a,
		border: 0x7c4a4a,
		glowAlpha: 0
	}
};

// レイアウト定数
export const LAYOUT = {
	// ノード間の水平間隔
	horizontalGap: 60,
	// ノード間の垂直間隔
	verticalGap: 80,
	// 初期マージン
	marginTop: 40,
	marginLeft: 40
};
