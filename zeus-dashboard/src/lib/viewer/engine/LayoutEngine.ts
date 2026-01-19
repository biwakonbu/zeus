// 自動レイアウトエンジン
import type { TaskItem } from '$lib/types/api';
import { TaskNode } from '../rendering/TaskNode';

// レイアウト設定
const HORIZONTAL_SPACING = 250; // ノード間の水平距離
const VERTICAL_SPACING = 120;   // ノード間の垂直距離
const LAYER_PADDING = 50;       // レイヤー間のパディング

/**
 * ノードの位置情報
 */
export interface NodePosition {
	id: string;
	x: number;
	y: number;
	layer: number; // トポロジカルソートでの深さ
}

/**
 * レイアウト結果
 */
export interface LayoutResult {
	positions: Map<string, NodePosition>;
	bounds: {
		minX: number;
		maxX: number;
		minY: number;
		maxY: number;
		width: number;
		height: number;
	};
	layers: string[][]; // 各レイヤーのノードID
}

/**
 * LayoutEngine - タスクノードの自動配置
 *
 * 責務:
 * - トポロジカルソートによるレイヤー分け
 * - 交差最小化のための並び替え
 * - 座標計算
 * - レイアウト結果のキャッシュ
 */
export class LayoutEngine {
	private nodeWidth: number;
	private nodeHeight: number;

	// レイアウトキャッシュ
	private cachedLayout: LayoutResult | null = null;
	private cachedLayoutHash: string = '';

	constructor() {
		this.nodeWidth = TaskNode.getWidth();
		this.nodeHeight = TaskNode.getHeight();
	}

	/**
	 * レイアウトハッシュを計算（構造変更の検出用）
	 */
	private computeLayoutHash(tasks: TaskItem[]): string {
		// タスクIDと依存関係のみでハッシュを計算（ステータス等は無視）
		return tasks
			.map(t => `${t.id}:${t.dependencies.sort().join(',')}`)
			.sort()
			.join('|');
	}

	/**
	 * タスクリストをレイアウト（キャッシュ対応）
	 */
	layout(tasks: TaskItem[]): LayoutResult {
		const hash = this.computeLayoutHash(tasks);

		// キャッシュが有効な場合はそのまま返す
		if (hash === this.cachedLayoutHash && this.cachedLayout) {
			return this.cachedLayout;
		}

		// 依存関係グラフを構築
		const graph = this.buildGraph(tasks);

		// トポロジカルソートでレイヤーを決定
		const layers = this.computeLayers(tasks, graph);

		// 各レイヤー内でノードを並び替え（交差最小化）
		this.minimizeCrossings(layers, graph);

		// 座標を計算
		const positions = this.computePositions(layers);

		// バウンディングボックスを計算
		const bounds = this.computeBounds(positions);

		const result = { positions, bounds, layers };

		// キャッシュを更新
		this.cachedLayout = result;
		this.cachedLayoutHash = hash;

		return result;
	}

	/**
	 * キャッシュをクリア
	 */
	clearCache(): void {
		this.cachedLayout = null;
		this.cachedLayoutHash = '';
	}

	/**
	 * 部分レイアウト（フィルター時用）
	 * 指定されたノードのみをレイアウトする（キャッシュを使わない）
	 */
	layoutSubset(tasks: TaskItem[], visibleIds: Set<string>): LayoutResult {
		// フィルター対象のタスクのみ抽出
		const filteredTasks = tasks.filter(t => visibleIds.has(t.id));

		// 依存関係も可視ノード内のみに制限
		const adjustedTasks = filteredTasks.map(t => ({
			...t,
			dependencies: t.dependencies.filter(d => visibleIds.has(d))
		}));

		// キャッシュを使わず新しいレイアウトを計算
		const graph = this.buildGraph(adjustedTasks);
		const layers = this.computeLayers(adjustedTasks, graph);
		this.minimizeCrossings(layers, graph);
		const positions = this.computePositions(layers);
		const bounds = this.computeBounds(positions);

		return { positions, bounds, layers };
	}

	/**
	 * 依存関係グラフを構築
	 */
	private buildGraph(tasks: TaskItem[]): {
		outgoing: Map<string, Set<string>>; // id -> 依存先のセット
		incoming: Map<string, Set<string>>; // id -> 依存元のセット
	} {
		const outgoing = new Map<string, Set<string>>();
		const incoming = new Map<string, Set<string>>();

		// 全ノードを初期化
		for (const task of tasks) {
			outgoing.set(task.id, new Set());
			incoming.set(task.id, new Set());
		}

		// エッジを追加
		const taskIds = new Set(tasks.map(t => t.id));
		for (const task of tasks) {
			for (const depId of task.dependencies) {
				// 依存先が存在する場合のみ追加
				if (taskIds.has(depId)) {
					outgoing.get(task.id)?.add(depId);
					incoming.get(depId)?.add(task.id);
				}
			}
		}

		return { outgoing, incoming };
	}

	/**
	 * トポロジカルソートでレイヤーを計算
	 * レイヤー0 = 依存なし（ルート）、レイヤーN = 深さN
	 */
	private computeLayers(
		tasks: TaskItem[],
		graph: { outgoing: Map<string, Set<string>>; incoming: Map<string, Set<string>> }
	): string[][] {
		const layers: string[][] = [];
		const nodeLayer = new Map<string, number>();
		const remaining = new Set(tasks.map(t => t.id));

		// 依存のないノードをレイヤー0に
		const rootNodes = tasks
			.filter(t => t.dependencies.length === 0 || !t.dependencies.some(d => remaining.has(d)))
			.map(t => t.id);

		if (rootNodes.length === 0 && remaining.size > 0) {
			// 循環依存がある場合、任意のノードをルートに
			const firstNode = remaining.values().next().value;
			if (firstNode) {
				rootNodes.push(firstNode);
			}
		}

		let currentLayer = 0;
		let currentNodes = rootNodes;

		while (currentNodes.length > 0) {
			layers[currentLayer] = currentNodes;

			for (const nodeId of currentNodes) {
				nodeLayer.set(nodeId, currentLayer);
				remaining.delete(nodeId);
			}

			// 次のレイヤーを計算
			const nextNodes: string[] = [];
			for (const nodeId of remaining) {
				const deps = graph.outgoing.get(nodeId) || new Set();
				// 全ての依存が処理済みならこのレイヤーに追加
				const allDepsProcessed = Array.from(deps).every(d => nodeLayer.has(d));
				if (allDepsProcessed) {
					nextNodes.push(nodeId);
				}
			}

			// 循環依存で進めない場合、残りを強制的に次のレイヤーに
			if (nextNodes.length === 0 && remaining.size > 0) {
				const forcedNode = remaining.values().next().value;
				if (forcedNode) {
					nextNodes.push(forcedNode);
				}
			}

			currentNodes = nextNodes;
			currentLayer++;
		}

		return layers;
	}

	/**
	 * 交差最小化のためにレイヤー内のノードを並び替え
	 * バリセンター法を使用
	 */
	private minimizeCrossings(
		layers: string[][],
		graph: { outgoing: Map<string, Set<string>>; incoming: Map<string, Set<string>> }
	): void {
		// 複数回イテレーション
		for (let iter = 0; iter < 3; iter++) {
			// 上から下へ
			for (let i = 1; i < layers.length; i++) {
				this.reorderLayer(layers, i, graph, 'down');
			}

			// 下から上へ
			for (let i = layers.length - 2; i >= 0; i--) {
				this.reorderLayer(layers, i, graph, 'up');
			}
		}
	}

	/**
	 * 単一レイヤーを並び替え
	 */
	private reorderLayer(
		layers: string[][],
		layerIndex: number,
		graph: { outgoing: Map<string, Set<string>>; incoming: Map<string, Set<string>> },
		direction: 'up' | 'down'
	): void {
		const layer = layers[layerIndex];
		const refLayer = direction === 'down' ? layers[layerIndex - 1] : layers[layerIndex + 1];

		if (!refLayer || refLayer.length === 0) return;

		// 参照レイヤーでの位置マップ
		const refPositions = new Map<string, number>();
		refLayer.forEach((id, index) => refPositions.set(id, index));

		// 各ノードのバリセンター（重心）を計算
		const barycenters: { id: string; value: number }[] = [];

		for (const nodeId of layer) {
			const connections = direction === 'down'
				? graph.outgoing.get(nodeId) || new Set()
				: graph.incoming.get(nodeId) || new Set();

			let sum = 0;
			let count = 0;

			for (const connId of connections) {
				const pos = refPositions.get(connId);
				if (pos !== undefined) {
					sum += pos;
					count++;
				}
			}

			const barycenter = count > 0 ? sum / count : layer.indexOf(nodeId);
			barycenters.push({ id: nodeId, value: barycenter });
		}

		// バリセンターでソート
		barycenters.sort((a, b) => a.value - b.value);
		layers[layerIndex] = barycenters.map(b => b.id);
	}

	/**
	 * 座標を計算
	 */
	private computePositions(layers: string[][]): Map<string, NodePosition> {
		const positions = new Map<string, NodePosition>();

		// 全体の幅を計算（最大レイヤー幅）
		const maxLayerWidth = Math.max(...layers.map(l => l.length));

		for (let layerIndex = 0; layerIndex < layers.length; layerIndex++) {
			const layer = layers[layerIndex];
			const layerWidth = layer.length * HORIZONTAL_SPACING;

			// Y座標はレイヤーインデックスに基づく
			const y = layerIndex * VERTICAL_SPACING + LAYER_PADDING;

			// X座標は中央揃え
			const startX = -(layerWidth / 2) + HORIZONTAL_SPACING / 2;

			for (let i = 0; i < layer.length; i++) {
				const nodeId = layer[i];
				const x = startX + i * HORIZONTAL_SPACING;

				positions.set(nodeId, {
					id: nodeId,
					x,
					y,
					layer: layerIndex
				});
			}
		}

		return positions;
	}

	/**
	 * バウンディングボックスを計算
	 */
	private computeBounds(positions: Map<string, NodePosition>): LayoutResult['bounds'] {
		if (positions.size === 0) {
			return { minX: 0, maxX: 0, minY: 0, maxY: 0, width: 0, height: 0 };
		}

		let minX = Infinity;
		let maxX = -Infinity;
		let minY = Infinity;
		let maxY = -Infinity;

		for (const pos of positions.values()) {
			minX = Math.min(minX, pos.x - this.nodeWidth / 2);
			maxX = Math.max(maxX, pos.x + this.nodeWidth / 2);
			minY = Math.min(minY, pos.y - this.nodeHeight / 2);
			maxY = Math.max(maxY, pos.y + this.nodeHeight / 2);
		}

		return {
			minX,
			maxX,
			minY,
			maxY,
			width: maxX - minX,
			height: maxY - minY
		};
	}

	/**
	 * エッジの接続点を計算
	 */
	computeEdgeEndpoints(
		fromPos: NodePosition,
		toPos: NodePosition
	): { fromX: number; fromY: number; toX: number; toY: number } {
		// ノードの中心から接続
		// 下方向への接続は下端から、上方向への接続は上端から
		const dy = toPos.y - fromPos.y;

		if (dy > 0) {
			// from が上、to が下
			return {
				fromX: fromPos.x,
				fromY: fromPos.y + this.nodeHeight / 2,
				toX: toPos.x,
				toY: toPos.y - this.nodeHeight / 2
			};
		} else {
			// from が下、to が上
			return {
				fromX: fromPos.x,
				fromY: fromPos.y - this.nodeHeight / 2,
				toX: toPos.x,
				toY: toPos.y + this.nodeHeight / 2
			};
		}
	}
}
