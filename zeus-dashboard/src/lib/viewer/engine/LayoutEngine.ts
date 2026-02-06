// 自動レイアウトエンジン
import type { GraphEdge, GraphNode } from '$lib/types/api';
import { GraphNodeView } from '../rendering/GraphNode';

// レイアウト設定
const LAYOUT_VERSION = 'grid-orthogonal-v3' as const;
export const LAYOUT_GRID_UNIT = 50;
export const EDGE_ROUTING_GRID_UNIT = Math.max(10, Math.floor(LAYOUT_GRID_UNIT / 5));
const COL_STEP = LAYOUT_GRID_UNIT * 5; // 250px
const ROW_STEP = LAYOUT_GRID_UNIT * 4; // 200px
const LAYER_PADDING_Y = LAYOUT_GRID_UNIT * 2; // 100px
const MIN_NODE_GAP_X = LAYOUT_GRID_UNIT * 5; // 250px
const GROUP_PADDING_X = 80;
const GROUP_PADDING_Y = 70;
const SWEEP_ITERATIONS = 3;
const STRUCTURAL_WEIGHT = 1.0;
const REFERENCE_WEIGHT = 0.35;

const COMPONENT_COLORS = [
	0x88b8ff, 0x66ccff, 0x8be08b, 0xffb366, 0xb48dff, 0xff9ecf, 0x7adfca, 0xf2d37a
] as const;

interface WeightedNeighbor {
	id: string;
	depth: number;
	weight: number;
}

/**
 * ノードの位置情報
 */
export interface NodePosition {
	id: string;
	x: number;
	y: number;
	layer: number;
}

/**
 * グループ境界情報
 */
export interface LayoutGroupBounds {
	groupId: string;
	label: string;
	nodeCount: number;
	minX: number;
	maxX: number;
	minY: number;
	maxY: number;
	width: number;
	height: number;
	color: number;
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
	layers: string[][];
	groups: LayoutGroupBounds[];
	layoutVersion: typeof LAYOUT_VERSION;
}

/**
 * LayoutEngine - ノードの自動配置
 *
 * 責務:
 * - structural depth を基準にレイヤー化
 * - structural 主 / reference 補助の順序最適化
 * - 50px グリッドへのスナップ
 * - structural 連結成分単位の境界算出
 * - レイアウト結果のキャッシュ
 */
export class LayoutEngine {
	private nodeWidth: number;
	private nodeHeight: number;

	private cachedLayout: LayoutResult | null = null;
	private cachedLayoutHash = '';

	constructor() {
		this.nodeWidth = GraphNodeView.getWidth();
		this.nodeHeight = GraphNodeView.getHeight();
	}

	private computeLayoutHash(nodes: GraphNode[], edges: GraphEdge[]): string {
		const nodePart = nodes
			.map((n) => `${n.id}:${n.node_type}:${n.structural_depth ?? 'na'}`)
			.sort()
			.join('|');
		const edgePart = edges
			.map((e) => `${e.from}->${e.to}:${e.layer}:${e.relation}`)
			.sort()
			.join('|');
		return `${LAYOUT_VERSION}#${nodePart}#${edgePart}`;
	}

	layout(nodes: GraphNode[], edges: GraphEdge[]): LayoutResult {
		const hash = this.computeLayoutHash(nodes, edges);
		if (hash === this.cachedLayoutHash && this.cachedLayout) {
			return this.cachedLayout;
		}

		const result = this.computeLayout(nodes, edges);
		this.cachedLayout = result;
		this.cachedLayoutHash = hash;
		return result;
	}

	clearCache(): void {
		this.cachedLayout = null;
		this.cachedLayoutHash = '';
	}

	layoutSubset(nodes: GraphNode[], edges: GraphEdge[], visibleIds: Set<string>): LayoutResult {
		const filteredNodes = nodes.filter((n) => visibleIds.has(n.id));
		const filteredEdges = edges.filter((e) => visibleIds.has(e.from) && visibleIds.has(e.to));
		return this.computeLayout(filteredNodes, filteredEdges);
	}

	private computeLayout(nodes: GraphNode[], edges: GraphEdge[]): LayoutResult {
		const sortedNodes = this.sortNodes(nodes);
		const sortedEdges = this.sortEdges(edges);
		const structuralEdges = sortedEdges.filter((edge) => edge.layer === 'structural');

		const graph = this.buildGraph(sortedNodes, structuralEdges);
		const fallbackLayers = this.computeLayers(sortedNodes, graph);
		this.minimizeCrossings(fallbackLayers, graph);

		const fallbackLayerIndex = this.buildLayerIndex(fallbackLayers);
		const depthByNode = this.resolveDepth(sortedNodes, fallbackLayerIndex);
		const depthLayers = this.buildDepthLayers(sortedNodes, depthByNode, fallbackLayers);
		this.optimizeDepthOrder(depthLayers, sortedEdges, depthByNode);

		const depthOrder = Array.from(depthLayers.keys()).sort((a, b) => a - b);
		const positions = this.computeGridPositions(depthOrder, depthLayers);
		const bounds = this.computeBounds(positions);
		const groups = this.computeComponentBounds(sortedNodes, structuralEdges, positions);

		return {
			positions,
			bounds,
			layers: depthOrder.map((depth) => [...(depthLayers.get(depth) ?? [])]),
			groups,
			layoutVersion: LAYOUT_VERSION
		};
	}

	private sortNodes(nodes: GraphNode[]): GraphNode[] {
		return [...nodes].sort((a, b) => {
			const depthA = this.normalizeDepth(a.structural_depth);
			const depthB = this.normalizeDepth(b.structural_depth);
			if (depthA !== depthB) return depthA - depthB;
			if (a.node_type !== b.node_type) return a.node_type.localeCompare(b.node_type);
			return a.id.localeCompare(b.id);
		});
	}

	private sortEdges(edges: GraphEdge[]): GraphEdge[] {
		return [...edges].sort((a, b) => {
			if (a.from !== b.from) return a.from.localeCompare(b.from);
			if (a.to !== b.to) return a.to.localeCompare(b.to);
			if (a.layer !== b.layer) return a.layer.localeCompare(b.layer);
			return a.relation.localeCompare(b.relation);
		});
	}

	private normalizeDepth(depth: number | undefined): number {
		if (depth === undefined || !Number.isFinite(depth)) return Number.MAX_SAFE_INTEGER;
		return Math.max(0, Math.floor(depth));
	}

	private buildLayerIndex(layers: string[][]): Map<string, number> {
		const layerIndex = new Map<string, number>();
		for (let layer = 0; layer < layers.length; layer++) {
			for (const id of layers[layer]) {
				layerIndex.set(id, layer);
			}
		}
		return layerIndex;
	}

	private resolveDepth(nodes: GraphNode[], layerIndex: Map<string, number>): Map<string, number> {
		const depthByNode = new Map<string, number>();
		for (const node of nodes) {
			let depth = 0;
			if (node.structural_depth !== undefined && Number.isFinite(node.structural_depth)) {
				depth = Math.max(0, Math.floor(node.structural_depth));
			} else {
				const fallback = layerIndex.get(node.id);
				depth = fallback !== undefined ? Math.max(0, fallback) : 0;
			}
			depthByNode.set(node.id, depth);
		}
		return depthByNode;
	}

	private buildDepthLayers(
		nodes: GraphNode[],
		depthByNode: Map<string, number>,
		fallbackLayers: string[][]
	): Map<number, string[]> {
		const depthLayers = new Map<number, string[]>();
		const fallbackOrder = new Map<string, number>();
		let order = 0;
		for (const layer of fallbackLayers) {
			for (const id of layer) {
				fallbackOrder.set(id, order++);
			}
		}

		for (const node of nodes) {
			const depth = depthByNode.get(node.id) ?? 0;
			if (!depthLayers.has(depth)) {
				depthLayers.set(depth, []);
			}
			depthLayers.get(depth)!.push(node.id);
		}

		for (const nodeIDs of depthLayers.values()) {
			nodeIDs.sort((a, b) => {
				const orderA = fallbackOrder.get(a) ?? Number.MAX_SAFE_INTEGER;
				const orderB = fallbackOrder.get(b) ?? Number.MAX_SAFE_INTEGER;
				if (orderA !== orderB) return orderA - orderB;
				return a.localeCompare(b);
			});
		}

		return depthLayers;
	}

	private optimizeDepthOrder(
		depthLayers: Map<number, string[]>,
		edges: GraphEdge[],
		depthByNode: Map<string, number>
	): void {
		const depthOrder = Array.from(depthLayers.keys()).sort((a, b) => a - b);
		if (depthOrder.length <= 1) return;

		const neighbors = this.buildWeightedNeighbors(depthByNode, edges);
		const layerPositions = this.buildLayerPositions(depthLayers);

		for (let iter = 0; iter < SWEEP_ITERATIONS; iter++) {
			for (let i = 1; i < depthOrder.length; i++) {
				this.reorderDepthLayer(depthOrder[i], 'down', depthLayers, layerPositions, neighbors);
			}
			for (let i = depthOrder.length - 2; i >= 0; i--) {
				this.reorderDepthLayer(depthOrder[i], 'up', depthLayers, layerPositions, neighbors);
			}
		}
	}

	private buildWeightedNeighbors(
		depthByNode: Map<string, number>,
		edges: GraphEdge[]
	): Map<string, WeightedNeighbor[]> {
		const neighbors = new Map<string, WeightedNeighbor[]>();
		for (const nodeID of depthByNode.keys()) {
			neighbors.set(nodeID, []);
		}

		for (const edge of edges) {
			const fromDepth = depthByNode.get(edge.from);
			const toDepth = depthByNode.get(edge.to);
			if (fromDepth === undefined || toDepth === undefined) continue;
			if (!neighbors.has(edge.from) || !neighbors.has(edge.to)) continue;

			const weight = edge.layer === 'structural' ? STRUCTURAL_WEIGHT : REFERENCE_WEIGHT;
			neighbors.get(edge.from)!.push({ id: edge.to, depth: toDepth, weight });
			neighbors.get(edge.to)!.push({ id: edge.from, depth: fromDepth, weight });
		}

		return neighbors;
	}

	private buildLayerPositions(depthLayers: Map<number, string[]>): Map<number, Map<string, number>> {
		const positions = new Map<number, Map<string, number>>();
		for (const [depth, nodeIDs] of depthLayers) {
			const indexMap = new Map<string, number>();
			nodeIDs.forEach((id, index) => indexMap.set(id, index));
			positions.set(depth, indexMap);
		}
		return positions;
	}

	private reorderDepthLayer(
		depth: number,
		direction: 'up' | 'down',
		depthLayers: Map<number, string[]>,
		layerPositions: Map<number, Map<string, number>>,
		neighbors: Map<string, WeightedNeighbor[]>
	): void {
		const currentLayer = depthLayers.get(depth);
		if (!currentLayer || currentLayer.length <= 1) return;

		const scored = currentLayer.map((id, index) => {
			const linked = neighbors.get(id) ?? [];
			let sum = 0;
			let totalWeight = 0;

			for (const neighbor of linked) {
				if (direction === 'down' && neighbor.depth >= depth) continue;
				if (direction === 'up' && neighbor.depth <= depth) continue;
				const refPos = layerPositions.get(neighbor.depth)?.get(neighbor.id);
				if (refPos === undefined) continue;

				const depthDiff = Math.abs(depth - neighbor.depth);
				const weighted = neighbor.weight / Math.max(1, depthDiff);
				sum += refPos * weighted;
				totalWeight += weighted;
			}

			return {
				id,
				score: totalWeight > 0 ? sum / totalWeight : index,
				index
			};
		});

		scored.sort((a, b) => {
			if (Math.abs(a.score - b.score) > 1e-9) return a.score - b.score;
			if (a.index !== b.index) return a.index - b.index;
			return a.id.localeCompare(b.id);
		});

		const reordered = scored.map((entry) => entry.id);
		depthLayers.set(depth, reordered);

		const indexMap = new Map<string, number>();
		reordered.forEach((id, index) => indexMap.set(id, index));
		layerPositions.set(depth, indexMap);
	}

	private computeGridPositions(
		depthOrder: number[],
		depthLayers: Map<number, string[]>
	): Map<string, NodePosition> {
		const positions = new Map<string, NodePosition>();

		for (const depth of depthOrder) {
			const nodeIDs = depthLayers.get(depth) ?? [];
			if (nodeIDs.length === 0) continue;

			const width = (nodeIDs.length - 1) * COL_STEP;
			const startX = this.snapToGrid(-width / 2);
			const y = this.snapToGrid(LAYER_PADDING_Y + depth * ROW_STEP);

			for (let index = 0; index < nodeIDs.length; index++) {
				const id = nodeIDs[index];
				const x = this.snapToGrid(startX + index * COL_STEP);
				positions.set(id, {
					id,
					x,
					y,
					layer: depth
				});
			}

			this.resolveLayerOverlap(nodeIDs, positions);
		}

		return positions;
	}

	private resolveLayerOverlap(nodeIDs: string[], positions: Map<string, NodePosition>): void {
		if (nodeIDs.length <= 1) return;

		for (let i = 1; i < nodeIDs.length; i++) {
			const prev = positions.get(nodeIDs[i - 1]);
			const curr = positions.get(nodeIDs[i]);
			if (!prev || !curr) continue;
			if (curr.x - prev.x < MIN_NODE_GAP_X) {
				curr.x = prev.x + MIN_NODE_GAP_X;
			}
		}

		for (let i = nodeIDs.length - 2; i >= 0; i--) {
			const curr = positions.get(nodeIDs[i]);
			const next = positions.get(nodeIDs[i + 1]);
			if (!curr || !next) continue;
			if (next.x - curr.x < MIN_NODE_GAP_X) {
				curr.x = next.x - MIN_NODE_GAP_X;
			}
		}

		for (const id of nodeIDs) {
			const pos = positions.get(id);
			if (pos) {
				pos.x = this.snapToGrid(pos.x);
			}
		}
	}

	private snapToGrid(value: number): number {
		return Math.round(value / LAYOUT_GRID_UNIT) * LAYOUT_GRID_UNIT;
	}

	private computeComponentBounds(
		nodes: GraphNode[],
		structuralEdges: GraphEdge[],
		positions: Map<string, NodePosition>
	): LayoutGroupBounds[] {
		const components = this.computeStructuralComponents(nodes, structuralEdges);
		const nodeByID = new Map(nodes.map((node) => [node.id, node]));
		const groups: LayoutGroupBounds[] = [];

		for (let index = 0; index < components.length; index++) {
			const component = components[index];
			const lanePositions = component
				.map((id) => positions.get(id))
				.filter((pos): pos is NodePosition => pos !== undefined);

			if (lanePositions.length === 0) continue;

			let minX = Infinity;
			let maxX = -Infinity;
			let minY = Infinity;
			let maxY = -Infinity;

			for (const pos of lanePositions) {
				minX = Math.min(minX, pos.x - this.nodeWidth / 2);
				maxX = Math.max(maxX, pos.x + this.nodeWidth / 2);
				minY = Math.min(minY, pos.y - this.nodeHeight / 2);
				maxY = Math.max(maxY, pos.y + this.nodeHeight / 2);
			}

			minX -= GROUP_PADDING_X;
			maxX += GROUP_PADDING_X;
			minY -= GROUP_PADDING_Y;
			maxY += GROUP_PADDING_Y;

			const anchorNode = component
				.map((id) => nodeByID.get(id))
				.filter((node): node is GraphNode => node !== undefined)
				.sort((a, b) => {
					const depthA = this.normalizeDepth(a.structural_depth);
					const depthB = this.normalizeDepth(b.structural_depth);
					if (depthA !== depthB) return depthA - depthB;
					return a.id.localeCompare(b.id);
				})[0];
			const label = anchorNode?.title?.trim() || `Component ${index + 1}`;

			groups.push({
				groupId: `component-${index}`,
				label,
				nodeCount: component.length,
				minX,
				maxX,
				minY,
				maxY,
				width: maxX - minX,
				height: maxY - minY,
				color: COMPONENT_COLORS[index % COMPONENT_COLORS.length]
			});
		}

		return groups;
	}

	private computeStructuralComponents(nodes: GraphNode[], structuralEdges: GraphEdge[]): string[][] {
		const nodeIDs = nodes.map((node) => node.id).sort();
		const adjacency = new Map<string, Set<string>>();
		for (const id of nodeIDs) {
			adjacency.set(id, new Set());
		}

		for (const edge of structuralEdges) {
			if (!adjacency.has(edge.from) || !adjacency.has(edge.to)) continue;
			adjacency.get(edge.from)!.add(edge.to);
			adjacency.get(edge.to)!.add(edge.from);
		}

		const visited = new Set<string>();
		const components: string[][] = [];

		for (const startID of nodeIDs) {
			if (visited.has(startID)) continue;

			const component: string[] = [];
			const stack = [startID];
			visited.add(startID);

			while (stack.length > 0) {
				const current = stack.pop()!;
				component.push(current);

				const neighbors = Array.from(adjacency.get(current) ?? []).sort().reverse();
				for (const next of neighbors) {
					if (visited.has(next)) continue;
					visited.add(next);
					stack.push(next);
				}
			}

			component.sort();
			components.push(component);
		}

		components.sort((a, b) => a[0].localeCompare(b[0]));
		return components;
	}

	private buildGraph(
		nodes: GraphNode[],
		structuralEdges: GraphEdge[]
	): {
		outgoing: Map<string, Set<string>>;
		incoming: Map<string, Set<string>>;
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

	computeEdgeEndpoints(
		fromPos: NodePosition,
		toPos: NodePosition
	): { fromX: number; fromY: number; toX: number; toY: number } {
		if (fromPos.id === toPos.id) {
			return {
				fromX: fromPos.x + this.nodeWidth / 2,
				fromY: fromPos.y,
				toX: toPos.x,
				toY: toPos.y - this.nodeHeight / 2
			};
		}

		const dx = toPos.x - fromPos.x;
		const dy = toPos.y - fromPos.y;

		if (Math.abs(dx) >= Math.abs(dy)) {
			if (dx >= 0) {
				return {
					fromX: fromPos.x + this.nodeWidth / 2,
					fromY: fromPos.y,
					toX: toPos.x - this.nodeWidth / 2,
					toY: toPos.y
				};
			}
			return {
				fromX: fromPos.x - this.nodeWidth / 2,
				fromY: fromPos.y,
				toX: toPos.x + this.nodeWidth / 2,
				toY: toPos.y
			};
		}

		if (dy >= 0) {
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
