// 選択管理クラス

import type { TaskItem } from '$lib/types/api';

/**
 * 選択イベント
 */
export interface SelectionEvent {
	type: 'select' | 'deselect' | 'clear';
	taskIds: string[];
}

/**
 * SelectionManager - タスクの選択状態を管理
 *
 * 責務:
 * - 単一選択/複数選択の管理
 * - 矩形選択のサポート
 * - 依存チェーン選択
 * - 選択イベントの発火
 */
export class SelectionManager {
	private selectedIds: Set<string> = new Set();
	private listeners: ((event: SelectionEvent) => void)[] = [];
	private tasks: Map<string, TaskItem> = new Map();

	/**
	 * タスクデータを設定
	 */
	setTasks(tasks: TaskItem[]): void {
		this.tasks.clear();
		for (const task of tasks) {
			this.tasks.set(task.id, task);
		}
	}

	/**
	 * 単一選択（トグル）
	 */
	toggleSelect(taskId: string, multi = false): void {
		if (!multi) {
			// シングル選択モード
			if (this.selectedIds.has(taskId)) {
				this.selectedIds.delete(taskId);
				this.emit({ type: 'deselect', taskIds: [taskId] });
			} else {
				const oldIds = Array.from(this.selectedIds);
				this.selectedIds.clear();
				this.selectedIds.add(taskId);
				if (oldIds.length > 0) {
					this.emit({ type: 'deselect', taskIds: oldIds });
				}
				this.emit({ type: 'select', taskIds: [taskId] });
			}
		} else {
			// マルチ選択モード（Ctrl/Cmd + クリック）
			if (this.selectedIds.has(taskId)) {
				this.selectedIds.delete(taskId);
				this.emit({ type: 'deselect', taskIds: [taskId] });
			} else {
				this.selectedIds.add(taskId);
				this.emit({ type: 'select', taskIds: [taskId] });
			}
		}
	}

	/**
	 * 選択に追加
	 */
	addToSelection(taskIds: string[]): void {
		const newIds = taskIds.filter((id) => !this.selectedIds.has(id));
		for (const id of newIds) {
			this.selectedIds.add(id);
		}
		if (newIds.length > 0) {
			this.emit({ type: 'select', taskIds: newIds });
		}
	}

	/**
	 * 選択から除外
	 */
	removeFromSelection(taskIds: string[]): void {
		const removedIds = taskIds.filter((id) => this.selectedIds.has(id));
		for (const id of removedIds) {
			this.selectedIds.delete(id);
		}
		if (removedIds.length > 0) {
			this.emit({ type: 'deselect', taskIds: removedIds });
		}
	}

	/**
	 * 矩形範囲で選択（座標はワールド座標）
	 */
	selectByRect(
		rect: { x: number; y: number; width: number; height: number },
		positions: Map<string, { x: number; y: number }>,
		nodeWidth: number,
		nodeHeight: number,
		additive = false
	): void {
		if (!additive) {
			const oldIds = Array.from(this.selectedIds);
			this.selectedIds.clear();
			if (oldIds.length > 0) {
				this.emit({ type: 'deselect', taskIds: oldIds });
			}
		}

		const newIds: string[] = [];
		for (const [id, pos] of positions) {
			// ノードの矩形
			const nodeRect = {
				x: pos.x - nodeWidth / 2,
				y: pos.y - nodeHeight / 2,
				width: nodeWidth,
				height: nodeHeight
			};

			// 矩形が交差するかチェック
			if (this.rectsIntersect(rect, nodeRect)) {
				if (!this.selectedIds.has(id)) {
					this.selectedIds.add(id);
					newIds.push(id);
				}
			}
		}

		if (newIds.length > 0) {
			this.emit({ type: 'select', taskIds: newIds });
		}
	}

	/**
	 * 依存チェーンを選択（上流・下流）
	 */
	selectDependencyChain(taskId: string, direction: 'upstream' | 'downstream' | 'both'): void {
		const chainIds = new Set<string>([taskId]);

		const visited = new Set<string>();
		const queue = [taskId];

		while (queue.length > 0) {
			const currentId = queue.shift()!;
			if (visited.has(currentId)) continue;
			visited.add(currentId);

			const task = this.tasks.get(currentId);
			if (!task) continue;

			// 上流（依存先）
			if (direction === 'upstream' || direction === 'both') {
				for (const depId of task.dependencies) {
					if (!visited.has(depId) && this.tasks.has(depId)) {
						chainIds.add(depId);
						queue.push(depId);
					}
				}
			}

			// 下流（このタスクに依存するタスク）
			if (direction === 'downstream' || direction === 'both') {
				for (const [id, t] of this.tasks) {
					if (t.dependencies.includes(currentId) && !visited.has(id)) {
						chainIds.add(id);
						queue.push(id);
					}
				}
			}
		}

		// 既存選択をクリアして新規選択
		const oldIds = Array.from(this.selectedIds);
		this.selectedIds = chainIds;

		if (oldIds.length > 0) {
			this.emit({ type: 'deselect', taskIds: oldIds });
		}
		this.emit({ type: 'select', taskIds: Array.from(chainIds) });
	}

	/**
	 * 全選択クリア
	 */
	clearSelection(): void {
		if (this.selectedIds.size > 0) {
			const oldIds = Array.from(this.selectedIds);
			this.selectedIds.clear();
			this.emit({ type: 'clear', taskIds: oldIds });
		}
	}

	/**
	 * 全選択
	 */
	selectAll(): void {
		const allIds = Array.from(this.tasks.keys());
		const newIds = allIds.filter((id) => !this.selectedIds.has(id));
		for (const id of allIds) {
			this.selectedIds.add(id);
		}
		if (newIds.length > 0) {
			this.emit({ type: 'select', taskIds: newIds });
		}
	}

	/**
	 * 選択されているか確認
	 */
	isSelected(taskId: string): boolean {
		return this.selectedIds.has(taskId);
	}

	/**
	 * 選択中のID一覧を取得
	 */
	getSelectedIds(): string[] {
		return Array.from(this.selectedIds);
	}

	/**
	 * 選択数を取得
	 */
	getSelectionCount(): number {
		return this.selectedIds.size;
	}

	/**
	 * イベントリスナーを追加
	 */
	onSelectionChange(listener: (event: SelectionEvent) => void): () => void {
		this.listeners.push(listener);
		return () => {
			const index = this.listeners.indexOf(listener);
			if (index >= 0) {
				this.listeners.splice(index, 1);
			}
		};
	}

	/**
	 * イベントを発火
	 */
	private emit(event: SelectionEvent): void {
		for (const listener of this.listeners) {
			listener(event);
		}
	}

	/**
	 * 矩形の交差判定
	 */
	private rectsIntersect(
		a: { x: number; y: number; width: number; height: number },
		b: { x: number; y: number; width: number; height: number }
	): boolean {
		return !(
			a.x > b.x + b.width ||
			a.x + a.width < b.x ||
			a.y > b.y + b.height ||
			a.y + a.height < b.y
		);
	}

	/**
	 * クリーンアップ
	 */
	destroy(): void {
		this.selectedIds.clear();
		this.listeners = [];
		this.tasks.clear();
	}
}
