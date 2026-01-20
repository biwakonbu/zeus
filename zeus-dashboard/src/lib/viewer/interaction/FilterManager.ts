// フィルタリング管理クラス

import type { TaskItem, TaskStatus, Priority } from '$lib/types/api';

/**
 * フィルター条件
 */
export interface FilterCriteria {
	statuses?: TaskStatus[];
	priorities?: Priority[];
	assignees?: string[];
	searchText?: string;
	hasDependencies?: boolean;
	isBlocked?: boolean;
}

/**
 * フィルター変更イベント
 */
export interface FilterChangeEvent {
	criteria: FilterCriteria;
	visibleIds: string[];
	hiddenIds: string[];
}

/**
 * FilterManager - タスクのフィルタリングを管理
 *
 * 責務:
 * - ステータス別フィルタ
 * - 優先度別フィルタ
 * - 担当者別フィルタ
 * - テキスト検索
 * - フィルター結果のキャッシュ
 */
export class FilterManager {
	private tasks: TaskItem[] = [];
	private criteria: FilterCriteria = {};
	private visibleIds: Set<string> = new Set();
	private listeners: ((event: FilterChangeEvent) => void)[] = [];

	/**
	 * タスクデータを設定
	 */
	setTasks(tasks: TaskItem[]): void {
		this.tasks = tasks;
		this.applyFilter();
	}

	/**
	 * フィルター条件を設定
	 */
	setCriteria(criteria: FilterCriteria): void {
		this.criteria = { ...criteria };
		this.applyFilter();
	}

	/**
	 * フィルター条件を更新（マージ）
	 */
	updateCriteria(partial: Partial<FilterCriteria>): void {
		this.criteria = { ...this.criteria, ...partial };
		this.applyFilter();
	}

	/**
	 * ステータスフィルターをトグル
	 */
	toggleStatus(status: TaskStatus): void {
		const statuses = this.criteria.statuses || [];
		const index = statuses.indexOf(status);
		if (index >= 0) {
			statuses.splice(index, 1);
		} else {
			statuses.push(status);
		}
		this.criteria.statuses = statuses.length > 0 ? statuses : undefined;
		this.applyFilter();
	}

	/**
	 * 優先度フィルターをトグル
	 */
	togglePriority(priority: Priority): void {
		const priorities = this.criteria.priorities || [];
		const index = priorities.indexOf(priority);
		if (index >= 0) {
			priorities.splice(index, 1);
		} else {
			priorities.push(priority);
		}
		this.criteria.priorities = priorities.length > 0 ? priorities : undefined;
		this.applyFilter();
	}

	/**
	 * 担当者フィルターをトグル
	 */
	toggleAssignee(assignee: string): void {
		const assignees = this.criteria.assignees || [];
		const index = assignees.indexOf(assignee);
		if (index >= 0) {
			assignees.splice(index, 1);
		} else {
			assignees.push(assignee);
		}
		this.criteria.assignees = assignees.length > 0 ? assignees : undefined;
		this.applyFilter();
	}

	/**
	 * 検索テキストを設定
	 */
	setSearchText(text: string): void {
		this.criteria.searchText = text || undefined;
		this.applyFilter();
	}

	/**
	 * フィルターをクリア
	 */
	clearFilter(): void {
		this.criteria = {};
		this.applyFilter();
	}

	/**
	 * 特定条件をクリア
	 */
	clearCriterion(key: keyof FilterCriteria): void {
		delete this.criteria[key];
		this.applyFilter();
	}

	/**
	 * タスクが表示されるか確認
	 */
	isVisible(taskId: string): boolean {
		return this.visibleIds.has(taskId);
	}

	/**
	 * 表示されるタスクIDを取得
	 */
	getVisibleIds(): string[] {
		return Array.from(this.visibleIds);
	}

	/**
	 * 表示されるタスクを取得
	 */
	getVisibleTasks(): TaskItem[] {
		return this.tasks.filter((t) => this.visibleIds.has(t.id));
	}

	/**
	 * 非表示のタスクIDを取得
	 */
	getHiddenIds(): string[] {
		return this.tasks.filter((t) => !this.visibleIds.has(t.id)).map((t) => t.id);
	}

	/**
	 * フィルターがアクティブか
	 */
	isActive(): boolean {
		return (
			(this.criteria.statuses?.length ?? 0) > 0 ||
			(this.criteria.priorities?.length ?? 0) > 0 ||
			(this.criteria.assignees?.length ?? 0) > 0 ||
			!!this.criteria.searchText ||
			this.criteria.hasDependencies !== undefined ||
			this.criteria.isBlocked !== undefined
		);
	}

	/**
	 * 現在のフィルター条件を取得
	 */
	getCriteria(): FilterCriteria {
		return { ...this.criteria };
	}

	/**
	 * 利用可能な担当者リストを取得
	 */
	getAvailableAssignees(): string[] {
		const assignees = new Set<string>();
		for (const task of this.tasks) {
			if (task.assignee) {
				assignees.add(task.assignee);
			}
		}
		return Array.from(assignees).sort();
	}

	/**
	 * イベントリスナーを追加
	 */
	onFilterChange(listener: (event: FilterChangeEvent) => void): () => void {
		this.listeners.push(listener);
		return () => {
			const index = this.listeners.indexOf(listener);
			if (index >= 0) {
				this.listeners.splice(index, 1);
			}
		};
	}

	/**
	 * フィルターを適用
	 */
	private applyFilter(): void {
		this.visibleIds.clear();

		for (const task of this.tasks) {
			if (this.matchesCriteria(task)) {
				this.visibleIds.add(task.id);
			}
		}

		// 変更があった場合のみイベント発火
		const newVisibleIds = Array.from(this.visibleIds);
		const hiddenIds = this.tasks.filter((t) => !this.visibleIds.has(t.id)).map((t) => t.id);

		this.emit({
			criteria: this.criteria,
			visibleIds: newVisibleIds,
			hiddenIds
		});
	}

	/**
	 * タスクがフィルター条件にマッチするか
	 */
	private matchesCriteria(task: TaskItem): boolean {
		// ステータスフィルター
		if (this.criteria.statuses && this.criteria.statuses.length > 0) {
			if (!this.criteria.statuses.includes(task.status)) {
				return false;
			}
		}

		// 優先度フィルター
		if (this.criteria.priorities && this.criteria.priorities.length > 0) {
			if (!this.criteria.priorities.includes(task.priority)) {
				return false;
			}
		}

		// 担当者フィルター
		if (this.criteria.assignees && this.criteria.assignees.length > 0) {
			if (!this.criteria.assignees.includes(task.assignee)) {
				return false;
			}
		}

		// テキスト検索
		if (this.criteria.searchText) {
			const searchLower = this.criteria.searchText.toLowerCase();
			const titleMatch = task.title.toLowerCase().includes(searchLower);
			const idMatch = task.id.toLowerCase().includes(searchLower);
			const assigneeMatch = task.assignee?.toLowerCase().includes(searchLower) ?? false;
			if (!titleMatch && !idMatch && !assigneeMatch) {
				return false;
			}
		}

		// 依存関係フィルター
		if (this.criteria.hasDependencies !== undefined) {
			const hasDeps = task.dependencies.length > 0;
			if (this.criteria.hasDependencies !== hasDeps) {
				return false;
			}
		}

		// ブロック状態フィルター
		if (this.criteria.isBlocked !== undefined) {
			const isBlocked = task.status === 'blocked';
			if (this.criteria.isBlocked !== isBlocked) {
				return false;
			}
		}

		return true;
	}

	/**
	 * イベントを発火
	 */
	private emit(event: FilterChangeEvent): void {
		for (const listener of this.listeners) {
			listener(event);
		}
	}

	/**
	 * クリーンアップ
	 */
	destroy(): void {
		this.tasks = [];
		this.visibleIds.clear();
		this.listeners = [];
		this.criteria = {};
	}
}
