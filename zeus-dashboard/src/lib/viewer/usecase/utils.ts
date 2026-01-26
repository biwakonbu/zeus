// UseCase ビューワー共通ユーティリティ
// Actor タイプアイコンとステータス色の変換関数を一元管理

import type { ActorType, SubsystemItem } from '$lib/types/api';

/**
 * Actor タイプに対応する Lucide アイコン名を取得
 * @param type - Actor タイプ
 * @returns Lucide アイコン名
 */
export function getActorIcon(type: ActorType): string {
	const icons: Record<ActorType, string> = {
		human: 'User',
		system: 'Server',
		time: 'Clock',
		device: 'Smartphone',
		external: 'Globe'
	};
	return icons[type] ?? 'HelpCircle';
}

/**
 * UseCase ステータスに対応する CSS カスタムプロパティを取得
 * @param status - UseCase ステータス
 * @returns CSS カスタムプロパティ名
 */
export function getStatusColor(status: string): string {
	switch (status) {
		case 'active':
			return 'var(--status-good)';
		case 'draft':
			return 'var(--status-fair)';
		case 'deprecated':
			return 'var(--text-muted)';
		default:
			return 'var(--text-secondary)';
	}
}

// =============================================================================
// サブシステム関連ユーティリティ（TASK-019）
// =============================================================================

/**
 * 未分類サブシステム用の固定色
 */
export const UNCATEGORIZED_COLOR = 0x555555;

/**
 * 未分類サブシステムの仮想アイテム
 */
export const UNCATEGORIZED_SUBSYSTEM: SubsystemItem = {
	id: '__uncategorized__',
	name: '未分類'
};

/**
 * HSL から 16進数カラーに変換
 * @param h - Hue (0-360)
 * @param s - Saturation (0-100)
 * @param l - Lightness (0-100)
 * @returns 16進数カラー値（0xRRGGBB）
 */
export function hslToHex(h: number, s: number, l: number): number {
	// 正規化
	const sNorm = s / 100;
	const lNorm = l / 100;

	const c = (1 - Math.abs(2 * lNorm - 1)) * sNorm;
	const x = c * (1 - Math.abs(((h / 60) % 2) - 1));
	const m = lNorm - c / 2;

	let r = 0,
		g = 0,
		b = 0;

	if (h < 60) {
		r = c;
		g = x;
		b = 0;
	} else if (h < 120) {
		r = x;
		g = c;
		b = 0;
	} else if (h < 180) {
		r = 0;
		g = c;
		b = x;
	} else if (h < 240) {
		r = 0;
		g = x;
		b = c;
	} else if (h < 300) {
		r = x;
		g = 0;
		b = c;
	} else {
		r = c;
		g = 0;
		b = x;
	}

	// RGB を 0-255 に変換
	const rInt = Math.round((r + m) * 255);
	const gInt = Math.round((g + m) * 255);
	const bInt = Math.round((b + m) * 255);

	// 16進数値に結合
	return (rInt << 16) | (gInt << 8) | bInt;
}

/**
 * サブシステム ID からカラーを生成
 * DJB2 ハッシュアルゴリズムを使用して一貫したカラーを生成
 *
 * @param subsystemId - サブシステム ID
 * @returns 16進数カラー値（0xRRGGBB）
 */
export function generateSubsystemColor(subsystemId: string): number {
	// 未分類の場合は固定色を返す
	if (subsystemId === UNCATEGORIZED_SUBSYSTEM.id) {
		return UNCATEGORIZED_COLOR;
	}

	// DJB2 ハッシュアルゴリズム
	let hash = 5381;
	for (let i = 0; i < subsystemId.length; i++) {
		hash = ((hash << 5) + hash) ^ subsystemId.charCodeAt(i);
	}

	// Hue: 30-330（オレンジ系を避けて Factorio テーマと区別）
	const hue = (Math.abs(hash) % 300) + 30;
	const saturation = 45;
	const lightness = 40;

	return hslToHex(hue, saturation, lightness);
}

/**
 * サブシステム境界の背景色（半透明版）を取得
 * @param color - 16進数カラー値
 * @param alpha - 透明度（0-1）
 * @returns RGBA カラー文字列
 */
export function getSubsystemBackgroundColor(color: number, alpha: number = 0.15): string {
	const r = (color >> 16) & 0xff;
	const g = (color >> 8) & 0xff;
	const b = color & 0xff;
	return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

/**
 * サブシステム境界のボーダー色を取得
 * @param color - 16進数カラー値
 * @returns RGBA カラー文字列
 */
export function getSubsystemBorderColor(color: number): string {
	const r = (color >> 16) & 0xff;
	const g = (color >> 8) & 0xff;
	const b = color & 0xff;
	return `rgba(${r}, ${g}, ${b}, 0.6)`;
}
