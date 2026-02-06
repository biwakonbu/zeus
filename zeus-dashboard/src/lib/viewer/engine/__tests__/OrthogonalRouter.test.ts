import { describe, expect, it } from 'vitest';
import { EDGE_ROUTING_GRID_UNIT, type NodePosition } from '../LayoutEngine';
import { OrthogonalRouter } from '../OrthogonalRouter';

function toPos(id: string, x: number, y: number, layer = 0): NodePosition {
	return { id, x, y, layer };
}

function makeMap(items: NodePosition[]): Map<string, NodePosition> {
	return new Map(items.map((item) => [item.id, item]));
}

function isOrthogonal(points: { x: number; y: number }[]): boolean {
	for (let i = 1; i < points.length; i++) {
		if (points[i - 1].x !== points[i].x && points[i - 1].y !== points[i].y) {
			return false;
		}
	}
	return true;
}

describe('OrthogonalRouter', () => {
	it('ノード接点は常に辺へ垂直に接続される', () => {
		const router = new OrthogonalRouter(200, 80, EDGE_ROUTING_GRID_UNIT);

		const horizontal = makeMap([toPos('a', 0, 0), toPos('b', 500, 100)]);
		const routeH = router.route('a', 'b', horizontal).points;
		expect(routeH.length).toBeGreaterThanOrEqual(3);
		expect(routeH[0].y).toBe(routeH[1].y);
		expect(routeH[0].x).not.toBe(routeH[1].x);
		expect(routeH[routeH.length - 1].y).toBe(routeH[routeH.length - 2].y);
		expect(routeH[routeH.length - 1].x).not.toBe(routeH[routeH.length - 2].x);

		const vertical = makeMap([toPos('c', 0, 0), toPos('d', 100, 500)]);
		const routeV = router.route('c', 'd', vertical).points;
		expect(routeV.length).toBeGreaterThanOrEqual(3);
		expect(routeV[0].x).toBe(routeV[1].x);
		expect(routeV[0].y).not.toBe(routeV[1].y);
		expect(routeV[routeV.length - 1].x).toBe(routeV[routeV.length - 2].x);
		expect(routeV[routeV.length - 1].y).not.toBe(routeV[routeV.length - 2].y);
	});

	it('全セグメントが水平または垂直になる', () => {
		const router = new OrthogonalRouter(200, 80, EDGE_ROUTING_GRID_UNIT);
		const positions = makeMap([toPos('a', 0, 0), toPos('b', 500, 250)]);

		const result = router.route('a', 'b', positions);
		expect(result.points.length).toBeGreaterThanOrEqual(2);
		expect(isOrthogonal(result.points)).toBe(true);
	});

	it('折点が配線グリッドに揃う', () => {
		const router = new OrthogonalRouter(200, 80, EDGE_ROUTING_GRID_UNIT);
		const positions = makeMap([toPos('a', 0, 0), toPos('b', 300, 300)]);

		const result = router.route('a', 'b', positions);
		expect(result.points.length).toBeGreaterThanOrEqual(3);

		for (let i = 1; i < result.points.length - 1; i++) {
			const p = result.points[i];
			expect(Math.abs(p.x % EDGE_ROUTING_GRID_UNIT)).toBe(0);
			expect(Math.abs(p.y % EDGE_ROUTING_GRID_UNIT)).toBe(0);
		}
	});

	it('単純障害物ケースでノード矩形を回避する', () => {
		const router = new OrthogonalRouter(200, 80, EDGE_ROUTING_GRID_UNIT);
		const positions = makeMap([toPos('from', 0, 0), toPos('to', 600, 0), toPos('block', 300, 0)]);

		const result = router.route('from', 'to', positions);
		expect(result.collisionCount).toBe(0);
		expect(result.usedFallback).toBe(false);
		expect(result.points.length).toBeGreaterThanOrEqual(3);
	});

	it('同一入力なら同一経路を返す（決定性）', () => {
		const router = new OrthogonalRouter(200, 80, EDGE_ROUTING_GRID_UNIT);
		const positions = makeMap([
			toPos('a', 0, 0),
			toPos('b', 500, 100),
			toPos('c', 250, 100),
			toPos('d', 250, -100)
		]);

		const routeA = router.route('a', 'b', positions);
		const routeB = router.route('a', 'b', positions);
		expect(routeA.usedFallback).toBe(routeB.usedFallback);
		expect(routeA.collisionCount).toBe(routeB.collisionCount);
		expect(routeA.points).toEqual(routeB.points);
	});

	it('高密度で全候補衝突時にフォールバックを返す', () => {
		const router = new OrthogonalRouter(200, 80, EDGE_ROUTING_GRID_UNIT);
		const blockers: NodePosition[] = [];
		for (let y = -300; y <= 300; y += EDGE_ROUTING_GRID_UNIT) {
			blockers.push(toPos(`block-${y}`, 300, y));
		}

		const positions = makeMap([toPos('from', 0, 0), toPos('to', 600, 0), ...blockers]);
		const result = router.route('from', 'to', positions);

		expect(result.usedFallback).toBe(true);
		expect(result.points.length).toBeGreaterThanOrEqual(2);
		expect(isOrthogonal(result.points)).toBe(true);
	});
});
