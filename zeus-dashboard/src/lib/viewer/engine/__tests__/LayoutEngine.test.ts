import { describe, expect, it, vi } from 'vitest';
import type { GraphEdge, GraphNode, GraphNodeType } from '$lib/types/api';

vi.mock('../../rendering/GraphNode', () => ({
	GraphNodeView: {
		getWidth: () => 200,
		getHeight: () => 80
	}
}));

import { LayoutEngine } from '../LayoutEngine';

function createNode(
	id: string,
	nodeType: GraphNodeType,
	structuralDepth?: number,
	overrides: Partial<GraphNode> = {}
): GraphNode {
	return {
		id,
		title: id,
		node_type: nodeType,
		status: 'pending',
		priority: 'medium',
		assignee: 'user',
		structural_depth: structuralDepth,
		...overrides
	};
}

function overlaps(a: { x: number; y: number }, b: { x: number; y: number }): boolean {
	const NODE_WIDTH = 200;
	const NODE_HEIGHT = 80;
	return Math.abs(a.x - b.x) < NODE_WIDTH && Math.abs(a.y - b.y) < NODE_HEIGHT;
}

function shuffled<T>(items: T[]): T[] {
	const copy = [...items];
	for (let i = copy.length - 1; i > 0; i--) {
		const j = (i * 37) % (i + 1);
		[copy[i], copy[j]] = [copy[j], copy[i]];
	}
	return copy;
}

describe('LayoutEngine（grid-orthogonal-v3）', () => {
	it('全ノード座標が 50px グリッドにスナップされる', () => {
		const nodes: GraphNode[] = [
			createNode('v1', 'vision', 0),
			createNode('o1', 'objective', 1),
			createNode('d1', 'deliverable', 2),
			createNode('u1', 'usecase', 2),
			createNode('a1', 'activity', 3)
		];
		const edges: GraphEdge[] = [
			{ from: 'o1', to: 'v1', layer: 'structural', relation: 'contributes' },
			{ from: 'd1', to: 'o1', layer: 'structural', relation: 'fulfills' },
			{ from: 'u1', to: 'o1', layer: 'structural', relation: 'contributes' },
			{ from: 'a1', to: 'u1', layer: 'structural', relation: 'implements' }
		];

		const engine = new LayoutEngine();
		const result = engine.layout(nodes, edges);

		expect(result.layoutVersion).toBe('grid-orthogonal-v3');
		for (const pos of result.positions.values()) {
			expect(Math.abs(pos.x % 50)).toBe(0);
			expect(Math.abs(pos.y % 50)).toBe(0);
		}
	});

	it('structural_depth がある場合は深さを優先して配置する', () => {
		const nodes: GraphNode[] = [
			createNode('a', 'activity', 3),
			createNode('b', 'activity', 0),
			createNode('c', 'activity')
		];
		const edges: GraphEdge[] = [
			{ from: 'a', to: 'b', layer: 'structural', relation: 'parent' },
			{ from: 'c', to: 'b', layer: 'structural', relation: 'parent' }
		];

		const engine = new LayoutEngine();
		const result = engine.layout(nodes, edges);

		expect(result.positions.get('b')?.layer).toBe(0);
		expect(result.positions.get('a')?.layer).toBe(3);
		expect((result.positions.get('c')?.layer ?? -1) >= 1).toBe(true);
		expect(result.positions.get('b')!.y).toBeLessThan(result.positions.get('a')!.y);
	});

	it('入力順序をシャッフルしても同一座標を返す（決定性）', () => {
		const nodes: GraphNode[] = [
			createNode('v1', 'vision', 0),
			createNode('o1', 'objective', 1),
			createNode('d1', 'deliverable', 2),
			createNode('u1', 'usecase', 2),
			createNode('a1', 'activity', 3),
			createNode('a2', 'activity', 3),
			createNode('a3', 'activity', 4)
		];
		const edges: GraphEdge[] = [
			{ from: 'o1', to: 'v1', layer: 'structural', relation: 'contributes' },
			{ from: 'd1', to: 'o1', layer: 'structural', relation: 'fulfills' },
			{ from: 'u1', to: 'o1', layer: 'structural', relation: 'contributes' },
			{ from: 'a1', to: 'u1', layer: 'structural', relation: 'implements' },
			{ from: 'a2', to: 'u1', layer: 'reference', relation: 'depends_on' },
			{ from: 'a3', to: 'a1', layer: 'reference', relation: 'depends_on' }
		];

		const engineA = new LayoutEngine();
		const resultA = engineA.layout(nodes, edges);

		const engineB = new LayoutEngine();
		const resultB = engineB.layout(shuffled(nodes), shuffled(edges));

		expect(resultA.positions.size).toBe(resultB.positions.size);
		for (const node of nodes) {
			const a = resultA.positions.get(node.id);
			const b = resultB.positions.get(node.id);
			expect(a).toBeDefined();
			expect(b).toBeDefined();
			expect(a?.x).toBe(b?.x);
			expect(a?.y).toBe(b?.y);
			expect(a?.layer).toBe(b?.layer);
		}
	});

	it('全ノードに座標が割り当てられる', () => {
		const nodes: GraphNode[] = [];
		const types: GraphNodeType[] = ['vision', 'objective', 'deliverable', 'usecase', 'activity'];
		for (let i = 0; i < 80; i++) {
			nodes.push(createNode(`n-${i}`, types[i % types.length], i % 6));
		}

		const engine = new LayoutEngine();
		const result = engine.layout(nodes, []);

		expect(result.positions.size).toBe(nodes.length);
		for (const node of nodes) {
			expect(result.positions.has(node.id)).toBe(true);
		}
	});

	it('ノード同士が重ならない（AABB 非交差）', () => {
		const nodes: GraphNode[] = [];
		for (let depth = 0; depth < 6; depth++) {
			for (let i = 0; i < 30; i++) {
				nodes.push(createNode(`n-${depth}-${i}`, 'activity', depth));
			}
		}

		const engine = new LayoutEngine();
		const result = engine.layout(nodes, []);
		const entries = Array.from(result.positions.values());

		for (let i = 0; i < entries.length; i++) {
			for (let j = i + 1; j < entries.length; j++) {
				expect(overlaps(entries[i], entries[j])).toBe(false);
			}
		}
	});

	it('layoutSubset は可視ノードに一致する groups を返す', () => {
		const nodes: GraphNode[] = [
			createNode('a', 'activity', 0),
			createNode('b', 'activity', 1),
			createNode('c', 'objective', 0),
			createNode('d', 'objective', 1)
		];
		const edges: GraphEdge[] = [
			{ from: 'b', to: 'a', layer: 'structural', relation: 'parent' },
			{ from: 'd', to: 'c', layer: 'structural', relation: 'parent' },
			{ from: 'b', to: 'd', layer: 'reference', relation: 'depends_on' }
		];

		const visible = new Set<string>(['a', 'b', 'd']);
		const engine = new LayoutEngine();
		const subset = engine.layoutSubset(nodes, edges, visible);

		expect(subset.positions.size).toBe(3);
		expect(Array.from(subset.positions.keys()).sort()).toEqual(['a', 'b', 'd']);
		expect(subset.groups.length).toBe(2);
		expect(subset.groups.map((g) => g.nodeCount).sort((x, y) => x - y)).toEqual([1, 2]);
		for (const group of subset.groups) {
			expect(group.groupId.startsWith('component-')).toBe(true);
		}
	});
});
