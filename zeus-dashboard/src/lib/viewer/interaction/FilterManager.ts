// フィルタリング管理クラス

import type { GraphEdge, GraphNode, EntityStatus } from '$lib/types/api';

/**
 * フィルター条件
 */
export interface FilterCriteria {
	statuses?: EntityStatus[];
	searchText?: string;
	hasDependencies?: boolean;
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
 * FilterManager - ノードのフィルタリングを管理
 *
 * 責務:
 * - ステータス別フィルタ
 * - テキスト検索
 * - フィルター結果のキャッシュ
 */
export class FilterManager {
	private nodes: GraphNode[] = [];
	private edges: GraphEdge[] = [];
	private criteria: FilterCriteria = {};
	private visibleIds: Set<string> = new Set();
	private listeners: ((event: FilterChangeEvent) => void)[] = [];
	private edgeCountByNode: Map<string, number> = new Map();

	/**
	 * グラフデータを設定
	 */
	setGraph(nodes: GraphNode[], edges: GraphEdge[]): void {
		this.nodes = nodes;
		this.edges = edges;
		this.applyFilter();
	}

	/**
	 * ノードデータのみ更新（後方互換）
	 */
	setNodes(nodes: GraphNode[]): void {
		this.setGraph(nodes, this.edges);
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
	toggleStatus(status: EntityStatus): void {
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
	 * 表示されるノードIDを取得
	 */
	getVisibleIds(): string[] {
		return Array.from(this.visibleIds);
	}

	/**
	 * 表示されるノードを取得
	 */
	getVisibleNodes(): GraphNode[] {
		return this.nodes.filter((n) => this.visibleIds.has(n.id));
	}

	/**
	 * 非表示のノードIDを取得
	 */
	getHiddenIds(): string[] {
		return this.nodes.filter((n) => !this.visibleIds.has(n.id)).map((n) => n.id);
	}

	/**
	 * フィルターがアクティブか
	 */
	isActive(): boolean {
		return (
			(this.criteria.statuses?.length ?? 0) > 0 ||
			!!this.criteria.searchText ||
			this.criteria.hasDependencies !== undefined
		);
	}

	/**
	 * 現在のフィルター条件を取得
	 */
	getCriteria(): FilterCriteria {
		return { ...this.criteria };
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
		this.edgeCountByNode.clear();
		for (const edge of this.edges) {
			this.edgeCountByNode.set(edge.from, (this.edgeCountByNode.get(edge.from) ?? 0) + 1);
			this.edgeCountByNode.set(edge.to, (this.edgeCountByNode.get(edge.to) ?? 0) + 1);
		}

		for (const node of this.nodes) {
			if (this.matchesCriteria(node)) {
				this.visibleIds.add(node.id);
			}
		}

		// 変更があった場合のみイベント発火
		const newVisibleIds = Array.from(this.visibleIds);
		const hiddenIds = this.nodes.filter((n) => !this.visibleIds.has(n.id)).map((n) => n.id);

		this.emit({
			criteria: this.criteria,
			visibleIds: newVisibleIds,
			hiddenIds
		});
	}

	/**
	 * ノードがフィルター条件にマッチするか
	 */
	private matchesCriteria(node: GraphNode): boolean {
		// ステータスフィルター
		if (this.criteria.statuses && this.criteria.statuses.length > 0) {
			const nodeStatus = node.status as EntityStatus;
			if (!this.criteria.statuses.includes(nodeStatus)) {
				return false;
			}
		}

		// テキスト検索
		if (this.criteria.searchText) {
			const searchLower = this.criteria.searchText.toLowerCase();
			const titleMatch = node.title.toLowerCase().includes(searchLower);
			const idMatch = node.id.toLowerCase().includes(searchLower);
			if (!titleMatch && !idMatch) {
				return false;
			}
		}

		// 依存関係フィルター
		if (this.criteria.hasDependencies !== undefined) {
			const hasDeps = (this.edgeCountByNode.get(node.id) ?? 0) > 0;
			if (this.criteria.hasDependencies !== hasDeps) {
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
		this.nodes = [];
		this.edges = [];
		this.visibleIds.clear();
		this.listeners = [];
		this.criteria = {};
		this.edgeCountByNode.clear();
	}
}
