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
	height: 10 // 太さ（6 → 10 に変更：視認性向上）
};

// 遷移エッジスタイル（矢印改善版）
export const TRANSITION_STYLE = {
	lineWidth: 2,
	arrowSize: 12, // 10 → 12
	arrowAngle: Math.PI / 7, // より鋭角な矢印
	guardFontSize: 11,
	guardPadding: 6,
	// 曲線オプション
	curveThreshold: 20 // この水平距離以上で曲線を使用
};

// ガード条件ラベルスタイル（バッジ風）
export const GUARD_LABEL_STYLE = {
	fontSize: 11,
	paddingH: 8,
	paddingV: 4,
	borderRadius: 4,
	// 通常状態
	background: 0x2a2a2a,
	backgroundAlpha: 0.95,
	border: 0x4a4a4a,
	borderWidth: 1,
	text: 0xe0e0e0,
	// ホバー/選択時
	hoverBackground: 0x3a3a3a,
	hoverBorder: 0xff9533,
	hoverText: 0xffffff,
	selectedBackground: 0x4a3520,
	selectedBorder: 0xff9533,
	selectedText: 0xffffff
} as const;

// ノードタイプ別カラー（Factorio 風強化版 - 5層ベベル・3層グロー対応）
export const NODE_COLORS = {
	// 初期ノード - 黒丸 + オレンジグロー
	initial: {
		fill: 0x1a1a1a,
		border: 0x666666,
		glow: 0xff9533,
		glowAlpha: 0.4,
		// 常時グロー（強化）
		baseGlowAlpha: 0.15 // 0.08 → 0.15
	},
	// 終了ノード - 二重丸 + 赤グロー
	final: {
		fill: 0x1a1a1a,
		border: 0x666666,
		innerFill: 0x2a1a1a, // 微かに赤み
		glow: 0xef4444,
		glowAlpha: 0.3,
		baseGlowAlpha: 0.12 // 0.06 → 0.12
	},
	// アクションノード - 金属質感強化（5層ベベル対応）
	action: {
		background: 0x2d2d2d, // メイン背景
		backgroundGradientTop: 0x3a3a3a, // 上部明るめ
		backgroundGradientBottom: 0x242424, // 下部暗め
		border: 0x5a5a5a,
		borderHighlight: 0x888888, // より明るく
		borderShadow: 0x1a1a1a, // シャドウ用
		text: 0xe0e0e0,
		// 3層グロー設定（強化）
		baseGlow: 0xff9533, // オレンジグロー
		baseGlowAlpha: 0.12, // 0.06 → 0.12（2倍に）
		hoverGlowAlpha: 0.25, // ホバー時
		selectedGlowAlpha: 0.4, // 選択時
		// アクセントライン（上部オレンジ）- 視認性向上
		accentLine: 0xff9533,
		accentLineAlpha: 0.25 // 0.12 → 0.25（明確に視認可能）
	},
	// 分岐/合流ノード - グラデーション追加（3層グロー対応）
	decision: {
		background: 0x2d3530, // 緑みを帯びた色
		backgroundGradient: 0x242d28,
		border: 0x4a6050,
		borderHighlight: 0x6a8070, // ハイライト
		borderShadow: 0x1a2520, // シャドウ
		text: 0xe0e0e0,
		// 3層グロー設定（強化）
		baseGlow: 0x66aa88,
		baseGlowAlpha: 0.12, // 0.06 → 0.12
		hoverGlowAlpha: 0.25,
		selectedGlowAlpha: 0.4,
		// アクセントライン追加
		accentLine: 0x66aa88,
		accentLineAlpha: 0.25
	},
	// フォーク/ジョインノード - より目立つ（金属質感強化）
	fork: {
		fill: 0x252525,
		border: 0x666666,
		highlight: 0x999999, // より明るく
		shadow: 0x111111, // シャドウ追加
		// 3層グロー設定（強化）
		baseGlow: 0x888888,
		baseGlowAlpha: 0.10, // 0.05 → 0.10
		hoverGlowAlpha: 0.2,
		selectedGlowAlpha: 0.35,
		// アクセントライン追加
		accentLine: 0xaaaaaa,
		accentLineAlpha: 0.20
	}
};

/**
 * 金属効果定数（5層ベベル構造用）
 * Layer 1: 外側シャドウ（下・右）
 * Layer 2: 内側シャドウ（下・右）
 * Layer 3: メイン背景
 * Layer 4: 内側ハイライト（上・左）
 * Layer 5: 外側ハイライト（上・左）
 *
 * alpha 累積問題を解消: 合計 0.70 に調整（以前は 1.13 で過度に暗かった）
 */
export const METAL_EFFECT = {
	// ベベル透明度（控えめに調整）
	outerShadowAlpha: 0.25, // 0.4 → 0.25（影を控えめに）
	innerShadowAlpha: 0.15, // 0.3 → 0.15（内側影も控えめに）
	innerHighlightAlpha: 0.5, // 維持
	outerHighlightAlpha: 0.4, // 維持
	// ベベル幅
	bevelWidth: 1.5,
	innerBevelWidth: 1,
	// 上部ハイライト（金属光沢）- 領域を拡大
	topHighlightAlpha: 0.20, // 0.25 → 0.20（光沢を適度に）
	topHighlightRatio: 0.45, // 0.35 → 0.45（領域拡大）
	// 下部シャドウ（凹み感）- 開始位置を上げて重なりを作る
	bottomShadowAlpha: 0.10, // 0.15 → 0.10（下部影は最小限）
	bottomShadowRatio: 0.40 // 0.3 → 0.40（60% 位置から開始、45% のハイライトと 5% 重複）
} as const;

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

/**
 * 遷移エッジ 4層スタイル（電気回路風・高コントラスト版）
 * - 背景 0x1a1a1a に対して明確に視認できるコントラスト
 * - 状態ごとに異なる発光感を表現
 *
 * デザイン原則:
 * - Layer 0: 最外層グロー（広い、常時微弱）
 * - Layer 1: グロー層（中程度）
 * - Layer 2: 外側縁取り（暗い）
 * - Layer 3: コア（明るい中心線）
 */
export const TRANSITION_EDGE_STYLE = {
	normal: {
		core: 0xdddddd, // より明るいコア
		outer: 0x3a3a3a, // より明るい縁取り
		glow: 0x888888, // より明るいグロー
		glowAlpha: 0.6,
		// 最外層グロー（常時表示）
		outerGlow: 0x666666,
		outerGlowAlpha: 0.15
	},
	hover: {
		core: 0xffdd99, // より明るいオレンジ
		outer: 0xff9533, // アクセント色の縁取り
		glow: 0xff9533,
		glowAlpha: 0.7,
		outerGlow: 0xff9533,
		outerGlowAlpha: 0.25
	},
	selected: {
		core: 0xffcc77, // より明るいオレンジ
		outer: 0x4a2a00, // 暗い縁取り
		glow: 0xff9533,
		glowAlpha: 0.8,
		outerGlow: 0xff9533,
		outerGlowAlpha: 0.3
	}
} as const;

/**
 * 遷移エッジ幅定義（4層構造対応・太めで視認性確保）
 */
export const TRANSITION_EDGE_WIDTHS = {
	normal: { core: 3, outer: 7, glow: 12, outerGlow: 20 },
	hover: { core: 3.5, outer: 8, glow: 14, outerGlow: 24 },
	selected: { core: 4, outer: 9, glow: 16, outerGlow: 28 }
} as const;
