// WBS ビューワー用状態管理
// ビュー間で選択状態と展開状態を共有する
import { writable, derived } from 'svelte/store';

// 選択中のエンティティ
export const selectedEntityId = writable<string | null>(null);
export const selectedEntityType = writable<string | null>(null);

// 展開済みの項目（折りたたみ用）
export const expandedIds = writable<Set<string>>(new Set());

// アクティブなビュー
export type ViewType = 'health' | 'timeline' | 'density';
export const activeView = writable<ViewType>('health');

// 派生状態: 選択があるかどうか
export const hasSelection = derived(selectedEntityId, ($id) => $id !== null);

/**
 * エンティティを選択する
 */
export function selectEntity(id: string | null, type: string | null = null): void {
	selectedEntityId.set(id);
	selectedEntityType.set(type);
}

/**
 * 選択をクリアする
 */
export function clearSelection(): void {
	selectedEntityId.set(null);
	selectedEntityType.set(null);
}

/**
 * 項目の展開/折りたたみをトグルする
 */
export function toggleExpand(id: string): void {
	expandedIds.update((ids) => {
		const newIds = new Set(ids);
		if (newIds.has(id)) {
			newIds.delete(id);
		} else {
			newIds.add(id);
		}
		return newIds;
	});
}

/**
 * 項目を展開する
 */
export function expand(id: string): void {
	expandedIds.update((ids) => {
		const newIds = new Set(ids);
		newIds.add(id);
		return newIds;
	});
}

/**
 * 項目を折りたたむ
 */
export function collapse(id: string): void {
	expandedIds.update((ids) => {
		const newIds = new Set(ids);
		newIds.delete(id);
		return newIds;
	});
}

/**
 * 全て展開する
 */
export function expandAll(ids: string[]): void {
	expandedIds.set(new Set(ids));
}

/**
 * 全て折りたたむ
 */
export function collapseAll(): void {
	expandedIds.set(new Set());
}

/**
 * ビューを切り替える
 */
export function setActiveView(view: ViewType): void {
	activeView.set(view);
}
