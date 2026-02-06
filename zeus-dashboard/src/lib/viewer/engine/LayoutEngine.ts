// 自動レイアウトエンジン
import type { GraphEdge, GraphNode } from '$lib/types/api';
import { GraphNodeView } from '../rendering/GraphNode';

// レイアウト設定
const HORIZONTAL_SPACING = 250; // ノード間の水平距離
const VERTICAL_SPACING = 120; // ノード間の垂直距離
const LAYER_PADDING = 50; // レイヤー間のパディング

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
 * LayoutEngine - ノードの自動配置
 *
 * 責務:
 * - structural エッジのみでレイヤー分け
 * - 交差最小化のための並び替え
 * - 座標計算
 * - レイアウト結果のキャッシュ
 */
export class LayoutEngine {
	private nodeWidth: number;
	private nodeHeight: number;

	// レイアウトキャッシュ
	private cachedLayout: LayoutResult | null = null;
	private cachedLayoutHash = '';

	constructor() {
		this.nodeWidth = GraphNodeView.getWidth();
		this.nodeHeight = GraphNodeView.getHeight();
	}

	/**
	 * レイアウトハッシュを計算（構造変更の検出用）
	 */
	private computeLayoutHash(nodes: GraphNode[], structuralEdges: GraphEdge[]): string {
		const nodePart = nodes
			.map((n) => n.id)
			.sort()
			.join('|');
		const edgePart = structuralEdges
			.map((e) => `${e.from}->${e.to}:${e.relation}`)
			.sort()
			.join('|');
		return `${nodePart}#${edgePart}`;
	}

	/**
	 * ノードをレイアウト（キャッシュ対応）
	 */
	layout(nodes: GraphNode[], structuralEdges: GraphEdge[]): LayoutResult {
		const hash = this.computeLayoutHash(nodes, structuralEdges);

		if (hash === this.cachedLayoutHash && this.cachedLayout) {
			return this.cachedLayout;
		}

		const graph = this.buildGraph(nodes, structuralEdges);
		const layers = this.computeLayers(nodes, graph);
		this.minimizeCrossings(layers, graph);
		const positions = this.computePositions(layers);
		const bounds = this.computeBounds(positions);

		const result = { positions, bounds, layers };
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
	 */
	layoutSubset(nodes: GraphNode[], structuralEdges: GraphEdge[], visibleIds: Set<string>): LayoutResult {
		const filteredNodes = nodes.filter((n) => visibleIds.has(n.id));
		const filteredEdges = structuralEdges.filter((e) => visibleIds.has(e.from) && visibleIds.has(e.to));

		const graph = this.buildGraph(filteredNodes, filteredEdges);
		const layers = this.computeLayers(filteredNodes, graph);
		this.minimizeCrossings(layers, graph);
		const positions = this.computePositions(layers);
		const bounds = this.computeBounds(positions);

		return { positions, bounds, layers };
	}

	/**
	 * 構造層グラフを構築
	 */
	private buildGraph(
		nodes: GraphNode[],
		structuralEdges: GraphEdge[]
	): {
		outgoing: Map<string, Set<string>>; // id -> 親側ノード
		incoming: Map<string, Set<string>>; // id -> 子側ノード
	} {
		const outgoing = new Map<string, Set<string>>();
		const incoming = new Map<string, Set<string>>();

		for (const node of nodes) {
			outgoing.set(node.id, new Set());
			incoming.set(node.id, new Set());
		}

		const nodeIds = new Set(nodes.map((n) => n.id));
		for (const edge of structuralEdges) {
			if (!nodeIds.has(edge.from) || !nodeIds.has(edge.to)) continue;
			outgoing.get(edge.from)?.add(edge.to);
			incoming.get(edge.to)?.add(edge.from);
		}

		return { outgoing, incoming };
	}

	/**
	 * トポロジカルソートでレイヤーを計算
	 */
	private computeLayers(
		nodes: GraphNode[],
		graph: { outgoing: Map<string, Set<string>>; incoming: Map<string, Set<string>> }
	): string[][] {
		const layers: string[][] = [];
		const nodeLayer = new Map<string, number>();
		const remaining = new Set(nodes.map((n) => n.id));

		const rootNodes = nodes
			.filter((n) => {
				const deps = graph.outgoing.get(n.id) ?? new Set();
				return deps.size === 0;
			})
			.map((n) => n.id);

		if (rootNodes.length === 0 && remaining.size > 0) {
			const firstNode = remaining.values().next().value;
			if (firstNode) rootNodes.push(firstNode);
		}

		let currentLayer = 0;
		let currentNodes = rootNodes;

		while (currentNodes.length > 0) {
			layers[currentLayer] = currentNodes;

			for (const nodeID of currentNodes) {
				nodeLayer.set(nodeID, currentLayer);
				remaining.delete(nodeID);
			}

			const nextNodes: string[] = [];
			for (const nodeID of remaining) {
				const deps = graph.outgoing.get(nodeID) ?? new Set();
				const allDepsProcessed = Array.from(deps).every((d) => nodeLayer.has(d));
				if (allDepsProcessed) {
					nextNodes.push(nodeID);
				}
			}

			if (nextNodes.length === 0 && remaining.size > 0) {
				const forcedNode = remaining.values().next().value;
				if (forcedNode) nextNodes.push(forcedNode);
			}

			currentNodes = nextNodes;
			currentLayer++;
		}

		return layers;
	}

	/**
	 * 交差最小化のためにレイヤー内のノードを並び替え
	 */
	private minimizeCrossings(
		layers: string[][],
		graph: { outgoing: Map<string, Set<string>>; incoming: Map<string, Set<string>> }
	): void {
		for (let iter = 0; iter < 3; iter++) {
			for (let i = 1; i < layers.length; i++) {
				this.reorderLayer(layers, i, graph, 'down');
			}
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

		const refPositions = new Map<string, number>();
		refLayer.forEach((id, index) => refPositions.set(id, index));

		const barycenters: { id: string; value: number }[] = [];
		for (const nodeID of layer) {
			const connections =
				direction === 'down'
					? graph.outgoing.get(nodeID) ?? new Set()
					: graph.incoming.get(nodeID) ?? new Set();

			let sum = 0;
			let count = 0;
			for (const connID of connections) {
				const pos = refPositions.get(connID);
				if (pos !== undefined) {
					sum += pos;
					count++;
				}
			}

			const barycenter = count > 0 ? sum / count : layer.indexOf(nodeID);
			barycenters.push({ id: nodeID, value: barycenter });
		}

		barycenters.sort((a, b) => a.value - b.value);
		layers[layerIndex] = barycenters.map((b) => b.id);
	}

	/**
	 * 座標を計算
	 */
	private computePositions(layers: string[][]): Map<string, NodePosition> {
		const positions = new Map<string, NodePosition>();

		for (let layerIndex = 0; layerIndex < layers.length; layerIndex++) {
			const layer = layers[layerIndex];
			const layerWidth = layer.length * HORIZONTAL_SPACING;
			const y = layerIndex * VERTICAL_SPACING + LAYER_PADDING;
			const startX = -(layerWidth / 2) + HORIZONTAL_SPACING / 2;

			for (let i = 0; i < layer.length; i++) {
				const nodeID = layer[i];
				const x = startX + i * HORIZONTAL_SPACING;
				positions.set(nodeID, { id: nodeID, x, y, layer: layerIndex });
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
		const dy = toPos.y - fromPos.y;

		if (dy > 0) {
			return {
				fromX: fromPos.x,
				fromY: fromPos.y + this.nodeHeight / 2,
				toX: toPos.x,
				toY: toPos.y - this.nodeHeight / 2
			};
		}

		return {
			fromX: fromPos.x,
			fromY: fromPos.y - this.nodeHeight / 2,
			toX: toPos.x,
			toY: toPos.y + this.nodeHeight / 2
		};
	}
}
