// Drill-Down Mode 状態管理ストア
// ダッシュボードから詳細ページへの遷移時に状態を保持・復元する

import { writable, get } from 'svelte/store';

/**
 * Drill-Down 遷移時に保存する状態
 */
export interface DrillDownState {
	/** 戻り先 URL */
	returnUrl: string;
	/** ビューの状態 */
	viewState: {
		/** スクロール X 位置 */
		scrollX?: number;
		/** スクロール Y 位置 */
		scrollY?: number;
		/** ズームレベル（Graph View 等） */
		zoomLevel?: number;
		/** アクティブなビュータブ（WBS Viewer） */
		activeView?: string;
	};
}

// 内部ストア
const drillDownState = writable<DrillDownState | null>(null);

/**
 * Drill-Down 遷移前に状態を保存
 */
export function saveDrillDownState(state: DrillDownState): void {
	drillDownState.set(state);
}

/**
 * 保存された状態を取得してクリア
 * @returns 保存された状態、なければ null
 */
export function restoreDrillDownState(): DrillDownState | null {
	const state = get(drillDownState);
	drillDownState.set(null);
	return state;
}

/**
 * 保存された状態があるかチェック
 */
export function hasSavedState(): boolean {
	return get(drillDownState) !== null;
}

/**
 * 状態をクリア（明示的にリセットしたい場合）
 */
export function clearDrillDownState(): void {
	drillDownState.set(null);
}

// ストア自体もエクスポート（購読用）
export { drillDownState };
