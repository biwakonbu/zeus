/**
 * 差分更新ユーティリティ
 *
 * タスクリストの変更を効率的に検出するための関数群
 */

import type { TaskItem } from '$lib/types/api';

/**
 * 変更タイプ
 * - none: 変更なし
 * - data: データのみ変更（ステータス、進捗等）
 * - structure: 構造変更（ノードの追加/削除、依存関係変更）
 */
export type ChangeType = 'none' | 'data' | 'structure';

/**
 * 差分検出器の状態
 */
export interface DiffDetectorState {
	previousTasksHash: string;
	previousTaskIds: Set<string>;
	previousDependencyHash: string;
}

/**
 * 初期状態を作成
 */
export function createInitialState(): DiffDetectorState {
	return {
		previousTasksHash: '',
		previousTaskIds: new Set(),
		previousDependencyHash: ''
	};
}

/**
 * タスクのデータハッシュを計算（ステータス、進捗等のみ）
 *
 * 構造に関係しない属性のみを含める
 */
export function computeTasksHash(tasks: TaskItem[]): string {
	return tasks
		.map(
			(t) =>
				`${t.id}:${t.status}:${t.progress ?? 0}:${t.priority ?? ''}:${t.assignee ?? ''}`
		)
		.join('|');
}

/**
 * 依存関係のハッシュを計算
 */
export function computeDependencyHash(tasks: TaskItem[]): string {
	return tasks
		.map((t) => `${t.id}:${t.dependencies.join(',')}`)
		.sort()
		.join('|');
}

/**
 * タスクリストの変更タイプを検出
 *
 * @param tasks 新しいタスクリスト
 * @param state 前回の状態
 * @returns 変更タイプと更新された状態
 */
export function detectTaskChanges(
	tasks: TaskItem[],
	state: DiffDetectorState
): { changeType: ChangeType; newState: DiffDetectorState } {
	const newTaskIds = new Set(tasks.map((t) => t.id));
	const newDependencyHash = computeDependencyHash(tasks);
	const newTasksHash = computeTasksHash(tasks);

	// 構造変更の検出
	// 1. タスク数が変わった
	// 2. タスクIDが変わった
	// 3. 依存関係が変わった
	const hasStructuralChange =
		newTaskIds.size !== state.previousTaskIds.size ||
		![...newTaskIds].every((id) => state.previousTaskIds.has(id)) ||
		newDependencyHash !== state.previousDependencyHash;

	const newState: DiffDetectorState = {
		previousTasksHash: newTasksHash,
		previousTaskIds: newTaskIds,
		previousDependencyHash: newDependencyHash
	};

	if (hasStructuralChange) {
		return { changeType: 'structure', newState };
	}

	// データ変更の検出
	if (newTasksHash !== state.previousTasksHash) {
		return { changeType: 'data', newState };
	}

	return { changeType: 'none', newState };
}

/**
 * 追加・削除されたタスクを特定
 *
 * @param newTasks 新しいタスクリスト
 * @param previousTaskIds 前回のタスクID集合
 * @returns 追加されたタスクと削除されたタスクID
 */
export function identifyTaskDiff(
	newTasks: TaskItem[],
	previousTaskIds: Set<string>
): {
	added: TaskItem[];
	removed: string[];
	unchanged: TaskItem[];
} {
	const newTaskIds = new Set(newTasks.map((t) => t.id));

	const added = newTasks.filter((t) => !previousTaskIds.has(t.id));
	const removed = [...previousTaskIds].filter((id) => !newTaskIds.has(id));
	const unchanged = newTasks.filter((t) => previousTaskIds.has(t.id));

	return { added, removed, unchanged };
}

/**
 * 可視性の差分を計算
 *
 * @param currentVisibleIds 現在の可視ノードID
 * @param previousVisibleIds 前回の可視ノードID
 * @returns 新しく見えるようになったIDと見えなくなったID
 */
export function computeVisibilityDiff(
	currentVisibleIds: Set<string>,
	previousVisibleIds: Set<string>
): {
	becameVisible: string[];
	becameHidden: string[];
} {
	const becameVisible: string[] = [];
	const becameHidden: string[] = [];

	// 新しく見えるようになったノード
	for (const id of currentVisibleIds) {
		if (!previousVisibleIds.has(id)) {
			becameVisible.push(id);
		}
	}

	// 見えなくなったノード
	for (const id of previousVisibleIds) {
		if (!currentVisibleIds.has(id)) {
			becameHidden.push(id);
		}
	}

	return { becameVisible, becameHidden };
}
