// Activity 図描画用の定数定義
// UML 2.5 準拠のアクティビティ図ノードスタイル

// 共通定数を共有ファイルからインポート
export { TEXT_RESOLUTION, COMMON_COLORS } from '$lib/viewer/shared/constants';

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

// 遷移エッジスタイル（矢印改善版）
export const TRANSITION_STYLE = {
	lineWidth: 2,
	arrowSize: 12, // 10 → 12
	arrowAngle: Math.PI / 7, // より鋭角な矢印
	guardFontSize: 10,
	guardPadding: 4,
	// 曲線オプション
	curveThreshold: 20 // この水平距離以上で曲線を使用
};

// ノードタイプ別カラー（Factorio 風強化版）
export const NODE_COLORS = {
	// 初期ノード - 黒丸 + オレンジグロー
	initial: {
		fill: 0x1a1a1a,
		border: 0x666666,
		glow: 0xff9533,
		glowAlpha: 0.4
	},
	// 終了ノード - 二重丸 + 赤グロー
	final: {
		fill: 0x1a1a1a,
		border: 0x666666,
		innerFill: 0x2a1a1a, // 微かに赤み
		glow: 0xef4444,
		glowAlpha: 0.3
	},
	// アクションノード - 金属質感強化
	action: {
		background: 0x2d2d2d, // 少し明るく
		backgroundGradientTop: 0x3a3a3a, // 上部明るめ
		backgroundGradientBottom: 0x242424, // 下部暗め
		border: 0x5a5a5a,
		borderHighlight: 0x777777, // 上部ハイライト用
		text: 0xe0e0e0 // より明るく
	},
	// 分岐/合流ノード - グラデーション追加
	decision: {
		background: 0x2d3530, // 緑みを帯びた色
		backgroundGradient: 0x242d28,
		border: 0x4a6050,
		text: 0xe0e0e0
	},
	// フォーク/ジョインノード - より目立つ
	fork: {
		fill: 0x252525,
		border: 0x666666,
		highlight: 0x888888
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

// レイアウト定数（間隔拡大版）
export const LAYOUT = {
	// ノード間の水平間隔
	horizontalGap: 100, // 60 → 100
	// ノード間の垂直間隔
	verticalGap: 120, // 80 → 120
	// 初期マージン
	marginTop: 60, // 40 → 60
	marginLeft: 60, // 40 → 60
	// 最小全体幅
	minTotalWidth: 600 // 追加
};
