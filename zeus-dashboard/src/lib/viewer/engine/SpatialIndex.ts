// Quadtree 空間インデックス - 大量ノードの高速クエリ

/**
 * 矩形領域
 */
export interface Rect {
	x: number;
	y: number;
	width: number;
	height: number;
}

/**
 * 空間インデックスに格納するアイテム
 */
export interface SpatialItem {
	id: string;
	x: number;
	y: number;
	width: number;
	height: number;
}

/**
 * Quadtree ノード
 */
class QuadNode {
	private bounds: Rect;
	private items: SpatialItem[] = [];
	private children: QuadNode[] | null = null;
	private maxItems: number;
	private maxDepth: number;
	private depth: number;

	constructor(bounds: Rect, maxItems = 10, maxDepth = 8, depth = 0) {
		this.bounds = bounds;
		this.maxItems = maxItems;
		this.maxDepth = maxDepth;
		this.depth = depth;
	}

	/**
	 * アイテムを挿入
	 */
	insert(item: SpatialItem): boolean {
		// バウンドに含まれない場合は挿入しない
		if (!this.containsItem(item)) {
			return false;
		}

		// 子ノードがある場合は子に挿入を試みる
		if (this.children) {
			for (const child of this.children) {
				if (child.insert(item)) {
					return true;
				}
			}
			// どの子にも入らなかった場合は自身に格納
			this.items.push(item);
			return true;
		}

		// 自身に格納
		this.items.push(item);

		// 分割条件を満たしたら分割
		if (this.items.length > this.maxItems && this.depth < this.maxDepth) {
			this.subdivide();
		}

		return true;
	}

	/**
	 * 領域内のアイテムを検索
	 */
	query(range: Rect): SpatialItem[] {
		const result: SpatialItem[] = [];

		// 範囲が交差しない場合は空を返す
		if (!this.intersects(range)) {
			return result;
		}

		// 自身のアイテムをチェック
		for (const item of this.items) {
			if (this.itemIntersectsRange(item, range)) {
				result.push(item);
			}
		}

		// 子ノードを再帰的に検索
		if (this.children) {
			for (const child of this.children) {
				result.push(...child.query(range));
			}
		}

		return result;
	}

	/**
	 * 全アイテムをクリア
	 */
	clear(): void {
		this.items = [];
		this.children = null;
	}

	/**
	 * 全アイテム数を取得
	 */
	count(): number {
		let total = this.items.length;
		if (this.children) {
			for (const child of this.children) {
				total += child.count();
			}
		}
		return total;
	}

	/**
	 * 4分割
	 */
	private subdivide(): void {
		const { x, y, width, height } = this.bounds;
		const halfW = width / 2;
		const halfH = height / 2;

		this.children = [
			// 左上
			new QuadNode(
				{ x, y, width: halfW, height: halfH },
				this.maxItems,
				this.maxDepth,
				this.depth + 1
			),
			// 右上
			new QuadNode(
				{ x: x + halfW, y, width: halfW, height: halfH },
				this.maxItems,
				this.maxDepth,
				this.depth + 1
			),
			// 左下
			new QuadNode(
				{ x, y: y + halfH, width: halfW, height: halfH },
				this.maxItems,
				this.maxDepth,
				this.depth + 1
			),
			// 右下
			new QuadNode(
				{ x: x + halfW, y: y + halfH, width: halfW, height: halfH },
				this.maxItems,
				this.maxDepth,
				this.depth + 1
			)
		];

		// 既存アイテムを子に再配置
		const oldItems = this.items;
		this.items = [];

		for (const item of oldItems) {
			let inserted = false;
			for (const child of this.children) {
				if (child.insert(item)) {
					inserted = true;
					break;
				}
			}
			// どの子にも入らなかった場合は自身に残す
			if (!inserted) {
				this.items.push(item);
			}
		}
	}

	/**
	 * アイテムがバウンド内に含まれるか
	 */
	private containsItem(item: SpatialItem): boolean {
		return (
			item.x >= this.bounds.x &&
			item.x + item.width <= this.bounds.x + this.bounds.width &&
			item.y >= this.bounds.y &&
			item.y + item.height <= this.bounds.y + this.bounds.height
		);
	}

	/**
	 * 範囲が交差するか
	 */
	private intersects(range: Rect): boolean {
		return !(
			range.x > this.bounds.x + this.bounds.width ||
			range.x + range.width < this.bounds.x ||
			range.y > this.bounds.y + this.bounds.height ||
			range.y + range.height < this.bounds.y
		);
	}

	/**
	 * アイテムが範囲と交差するか
	 */
	private itemIntersectsRange(item: SpatialItem, range: Rect): boolean {
		return !(
			item.x > range.x + range.width ||
			item.x + item.width < range.x ||
			item.y > range.y + range.height ||
			item.y + item.height < range.y
		);
	}
}

/**
 * SpatialIndex - Quadtree ベースの空間インデックス
 *
 * 責務:
 * - 大量ノードの効率的な管理
 * - ビューポート内のノードを高速に取得
 * - 点・矩形でのクエリ
 */
export class SpatialIndex {
	private root: QuadNode;
	private items: Map<string, SpatialItem> = new Map();

	constructor(bounds: Rect, maxItems = 10, maxDepth = 8) {
		this.root = new QuadNode(bounds, maxItems, maxDepth);
	}

	/**
	 * アイテムを挿入
	 */
	insert(item: SpatialItem): void {
		this.items.set(item.id, item);
		this.root.insert(item);
	}

	/**
	 * 複数アイテムを一括挿入
	 */
	insertAll(items: SpatialItem[]): void {
		for (const item of items) {
			this.insert(item);
		}
	}

	/**
	 * 矩形範囲内のアイテムを取得
	 */
	queryRect(range: Rect): SpatialItem[] {
		return this.root.query(range);
	}

	/**
	 * ビューポート内のアイテムを取得（マージン付き）
	 */
	queryViewport(viewport: Rect, margin = 100): SpatialItem[] {
		const expandedRange: Rect = {
			x: viewport.x - margin,
			y: viewport.y - margin,
			width: viewport.width + margin * 2,
			height: viewport.height + margin * 2
		};
		return this.root.query(expandedRange);
	}

	/**
	 * 点を含むアイテムを取得
	 */
	queryPoint(x: number, y: number): SpatialItem[] {
		return this.root.query({
			x,
			y,
			width: 1,
			height: 1
		});
	}

	/**
	 * アイテムを取得
	 */
	get(id: string): SpatialItem | undefined {
		return this.items.get(id);
	}

	/**
	 * 全アイテムをクリア
	 */
	clear(): void {
		this.items.clear();
		this.root.clear();
	}

	/**
	 * 再構築（バウンドが変わった場合）
	 */
	rebuild(newBounds: Rect): void {
		const oldItems = Array.from(this.items.values());
		this.root = new QuadNode(newBounds);
		this.items.clear();
		this.insertAll(oldItems);
	}

	/**
	 * アイテム総数
	 */
	count(): number {
		return this.items.size;
	}
}
