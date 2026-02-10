// ノードタイプの色・ラベル・表示設定の一元管理
// 新しいノードタイプを追加する場合はここだけ変更すればよい
import type { GraphNodeType } from '$lib/types/api';

export interface NodeTypeColors {
	indicator: number;
	background: number;
	border: number;
	borderHighlight: number;
	borderShadow: number;
}

export interface NodeTypeConfig {
	label: string; // バッジラベル（1文字）
	cssColor: string; // CSS 色コード
	colors: NodeTypeColors; // PixiJS 描画用色
	showBadge: boolean; // タイプバッジを表示するか
	showAccentLine: boolean; // 上部アクセントラインを表示するか
}

// ノードタイプ設定マスター定義
export const NODE_TYPE_CONFIG: Record<GraphNodeType, NodeTypeConfig> = {
	vision: {
		label: 'VIS',
		cssColor: '#ffd700',
		colors: {
			indicator: 0xffd700, // ゴールド - 最上位の目標
			background: 0x3d3520,
			border: 0xffd700,
			borderHighlight: 0xffee88,
			borderShadow: 0x2a2510
		},
		showBadge: true,
		showAccentLine: true
	},
	objective: {
		label: 'OBJ',
		cssColor: '#6699ff',
		colors: {
			indicator: 0x6699ff, // ブルー - 目標
			background: 0x2d3550,
			border: 0x6699ff,
			borderHighlight: 0x99bbff,
			borderShadow: 0x1a2030
		},
		showBadge: true,
		showAccentLine: true
	},
	activity: {
		label: 'ACT',
		cssColor: '#cc8844',
		colors: {
			indicator: 0xcc8844, // アンバー - Activity
			background: 0x352d20,
			border: 0xcc8844,
			borderHighlight: 0xddaa66,
			borderShadow: 0x221a10
		},
		showBadge: true,
		showAccentLine: true
	},
	usecase: {
		label: 'UC',
		cssColor: '#9966cc',
		colors: {
			indicator: 0x9966cc, // パープル - ユースケース
			background: 0x352d45,
			border: 0x9966cc,
			borderHighlight: 0xbb99ee,
			borderShadow: 0x201a30
		},
		showBadge: true,
		showAccentLine: true
	}
};

// デフォルトのノードタイプ（未知の型のフォールバック）
export const DEFAULT_NODE_TYPE: GraphNodeType = 'activity';

// PixiJS 色定義の取得
export function getNodeTypeColors(nodeType: GraphNodeType): NodeTypeColors {
	return NODE_TYPE_CONFIG[nodeType]?.colors ?? NODE_TYPE_CONFIG[DEFAULT_NODE_TYPE].colors;
}

// CSS 色の取得
export function getNodeTypeCSSColor(nodeType: GraphNodeType): string {
	return NODE_TYPE_CONFIG[nodeType]?.cssColor ?? NODE_TYPE_CONFIG[DEFAULT_NODE_TYPE].cssColor;
}

// ラベルの取得
export function getNodeTypeLabel(nodeType: GraphNodeType): string {
	return NODE_TYPE_CONFIG[nodeType]?.label ?? '?';
}

// バッジ表示するかの判定
export function shouldShowBadge(nodeType: GraphNodeType): boolean {
	return NODE_TYPE_CONFIG[nodeType]?.showBadge ?? true;
}

// アクセントライン表示するかの判定
export function shouldShowAccentLine(nodeType: GraphNodeType): boolean {
	return NODE_TYPE_CONFIG[nodeType]?.showAccentLine ?? true;
}
