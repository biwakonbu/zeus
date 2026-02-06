import { EDGE_ROUTING_GRID_UNIT, type NodePosition } from './LayoutEngine';

export interface RoutePoint {
	x: number;
	y: number;
}

export interface OrthogonalRouteResult {
	points: RoutePoint[];
	usedFallback: boolean;
	collisionCount: number;
}

interface Rect {
	minX: number;
	maxX: number;
	minY: number;
	maxY: number;
}

interface CandidateRoute {
	points: RoutePoint[];
	collisionCount: number;
	bends: number;
	manhattanLength: number;
	score: number;
}

type PortSide = 'left' | 'right' | 'top' | 'bottom';

interface RoutePort {
	point: RoutePoint;
	side: PortSide;
	normalX: number;
	normalY: number;
}

const DEFAULT_GRID_UNIT = EDGE_ROUTING_GRID_UNIT;
const OBSTACLE_PADDING = 12;
const STUB_LENGTH_IN_GRID = 1;

export class OrthogonalRouter {
	private nodeWidth: number;
	private nodeHeight: number;
	private gridUnit: number;

	constructor(nodeWidth: number, nodeHeight: number, gridUnit = DEFAULT_GRID_UNIT) {
		this.nodeWidth = nodeWidth;
		this.nodeHeight = nodeHeight;
		this.gridUnit = gridUnit;
	}

	getGridUnit(): number {
		return this.gridUnit;
	}

	route(
		fromId: string,
		toId: string,
		positions: Map<string, NodePosition>
	): OrthogonalRouteResult {
		const fromPos = positions.get(fromId);
		const toPos = positions.get(toId);
		if (!fromPos || !toPos) {
			return { points: [], usedFallback: true, collisionCount: 0 };
		}

		if (fromId === toId) {
			const loopPath = this.buildSelfLoop(fromPos);
			return {
				points: loopPath,
				usedFallback: false,
				collisionCount: 0
			};
		}

		const { fromPort, toPort } = this.pickPorts(fromPos, toPos);
		const fromStub = this.extendFromPort(fromPort);
		const toStub = this.extendFromPort(toPort);
		const obstacles = this.buildObstacles(positions, fromId, toId);
		const offsetSequence = this.buildOffsetSequence();

		const primaryCandidates: CandidateRoute[] = [];
		for (const offset of offsetSequence) {
			const midX = this.snapToGrid((fromStub.x + toStub.x) / 2) + offset * this.gridUnit;
			const midY = this.snapToGrid((fromStub.y + toStub.y) / 2) + offset * this.gridUnit;

			if (this.isHorizontalChannelValid(midX, fromPort, fromStub, toPort, toStub)) {
				primaryCandidates.push(
					this.evaluateRoute(
						[
							fromPort.point,
							fromStub,
							{ x: midX, y: fromStub.y },
							{ x: midX, y: toStub.y },
							toStub,
							toPort.point
						],
						obstacles
					)
				);
			}
			if (this.isVerticalChannelValid(midY, fromPort, fromStub, toPort, toStub)) {
				primaryCandidates.push(
					this.evaluateRoute(
						[
							fromPort.point,
							fromStub,
							{ x: fromStub.x, y: midY },
							{ x: toStub.x, y: midY },
							toStub,
							toPort.point
						],
						obstacles
					)
				);
			}
		}

		const bestPrimary = this.pickBest(primaryCandidates);
		if (bestPrimary && bestPrimary.collisionCount === 0) {
			return {
				points: bestPrimary.points,
				usedFallback: false,
				collisionCount: 0
			};
		}

		const fallbackCandidates: CandidateRoute[] = [];
		if (this.isHorizontalChannelValid(toStub.x, fromPort, fromStub, toPort, toStub)) {
			fallbackCandidates.push(
				this.evaluateRoute(
					[fromPort.point, fromStub, { x: toStub.x, y: fromStub.y }, toStub, toPort.point],
					obstacles
				)
			);
		}
		if (this.isVerticalChannelValid(toStub.y, fromPort, fromStub, toPort, toStub)) {
			fallbackCandidates.push(
				this.evaluateRoute(
					[fromPort.point, fromStub, { x: fromStub.x, y: toStub.y }, toStub, toPort.point],
					obstacles
				)
			);
		}
		if (fromStub.x === toStub.x || fromStub.y === toStub.y) {
			fallbackCandidates.push(
				this.evaluateRoute([fromPort.point, fromStub, toStub, toPort.point], obstacles)
			);
		}

		const bestFallback = this.pickBest(fallbackCandidates);
		if (bestFallback) {
			return {
				points: bestFallback.points,
				usedFallback: true,
				collisionCount: bestFallback.collisionCount
			};
		}

		if (bestPrimary) {
			return {
				points: bestPrimary.points,
				usedFallback: true,
				collisionCount: bestPrimary.collisionCount
			};
		}

		const hardFallback = [fromPort.point, fromStub, toStub, toPort.point];
		return {
			points: this.normalizePath(hardFallback),
			usedFallback: true,
			collisionCount: this.countCollisions(hardFallback, obstacles)
		};
	}

	private buildSelfLoop(pos: NodePosition): RoutePoint[] {
		const halfWidth = this.nodeWidth / 2;
		const halfHeight = this.nodeHeight / 2;
		const loopTopY = this.snapToGrid(pos.y - halfHeight - this.gridUnit);
		const loopRightX = this.snapToGrid(pos.x + halfWidth + this.gridUnit);

		return this.normalizePath([
			{ x: pos.x + halfWidth, y: pos.y },
			{ x: loopRightX, y: pos.y },
			{ x: loopRightX, y: loopTopY },
			{ x: this.snapToGrid(pos.x), y: loopTopY },
			{ x: pos.x, y: pos.y - halfHeight }
		]);
	}

	private pickPorts(
		fromPos: NodePosition,
		toPos: NodePosition
	): { fromPort: RoutePort; toPort: RoutePort } {
		const dx = toPos.x - fromPos.x;
		const dy = toPos.y - fromPos.y;

		if (Math.abs(dx) >= Math.abs(dy)) {
			if (dx >= 0) {
				return {
					fromPort: {
						point: { x: fromPos.x + this.nodeWidth / 2, y: fromPos.y },
						side: 'right',
						normalX: 1,
						normalY: 0
					},
					toPort: {
						point: { x: toPos.x - this.nodeWidth / 2, y: toPos.y },
						side: 'left',
						normalX: -1,
						normalY: 0
					}
				};
			}
			return {
				fromPort: {
					point: { x: fromPos.x - this.nodeWidth / 2, y: fromPos.y },
					side: 'left',
					normalX: -1,
					normalY: 0
				},
				toPort: {
					point: { x: toPos.x + this.nodeWidth / 2, y: toPos.y },
					side: 'right',
					normalX: 1,
					normalY: 0
				}
			};
		}

		if (dy >= 0) {
			return {
				fromPort: {
					point: { x: fromPos.x, y: fromPos.y + this.nodeHeight / 2 },
					side: 'bottom',
					normalX: 0,
					normalY: 1
				},
				toPort: {
					point: { x: toPos.x, y: toPos.y - this.nodeHeight / 2 },
					side: 'top',
					normalX: 0,
					normalY: -1
				}
			};
		}

		return {
			fromPort: {
				point: { x: fromPos.x, y: fromPos.y - this.nodeHeight / 2 },
				side: 'top',
				normalX: 0,
				normalY: -1
			},
			toPort: {
				point: { x: toPos.x, y: toPos.y + this.nodeHeight / 2 },
				side: 'bottom',
				normalX: 0,
				normalY: 1
			}
		};
	}

	private extendFromPort(port: RoutePort): RoutePoint {
		const distance = this.gridUnit * STUB_LENGTH_IN_GRID;
		return {
			x: this.snapToGrid(port.point.x + port.normalX * distance),
			y: this.snapToGrid(port.point.y + port.normalY * distance)
		};
	}

	private isHorizontalChannelValid(
		midX: number,
		fromPort: RoutePort,
		fromStub: RoutePoint,
		toPort: RoutePort,
		toStub: RoutePoint
	): boolean {
		const fromOkay =
			fromPort.side === 'right'
				? midX >= fromStub.x
				: fromPort.side === 'left'
					? midX <= fromStub.x
					: true;
		const toOkay =
			toPort.side === 'right'
				? midX >= toStub.x
				: toPort.side === 'left'
					? midX <= toStub.x
					: true;
		return fromOkay && toOkay;
	}

	private isVerticalChannelValid(
		midY: number,
		fromPort: RoutePort,
		fromStub: RoutePoint,
		toPort: RoutePort,
		toStub: RoutePoint
	): boolean {
		const fromOkay =
			fromPort.side === 'bottom'
				? midY >= fromStub.y
				: fromPort.side === 'top'
					? midY <= fromStub.y
					: true;
		const toOkay =
			toPort.side === 'bottom'
				? midY >= toStub.y
				: toPort.side === 'top'
					? midY <= toStub.y
					: true;
		return fromOkay && toOkay;
	}

	private buildObstacles(
		positions: Map<string, NodePosition>,
		fromId: string,
		toId: string
	): Rect[] {
		const rects: Rect[] = [];
		for (const [id, pos] of positions) {
			if (id === fromId || id === toId) continue;
			rects.push({
				minX: pos.x - this.nodeWidth / 2 - OBSTACLE_PADDING,
				maxX: pos.x + this.nodeWidth / 2 + OBSTACLE_PADDING,
				minY: pos.y - this.nodeHeight / 2 - OBSTACLE_PADDING,
				maxY: pos.y + this.nodeHeight / 2 + OBSTACLE_PADDING
			});
		}
		return rects;
	}

	private evaluateRoute(points: RoutePoint[], obstacles: Rect[]): CandidateRoute {
		const normalized = this.normalizePath(points);
		const collisionCount = this.countCollisions(normalized, obstacles);
		const bends = Math.max(0, normalized.length - 2);
		const manhattanLength = this.computeLength(normalized);
		const score = collisionCount * 100000 + bends * 100 + manhattanLength;
		return {
			points: normalized,
			collisionCount,
			bends,
			manhattanLength,
			score
		};
	}

	private pickBest(candidates: CandidateRoute[]): CandidateRoute | null {
		if (candidates.length === 0) return null;
		const sorted = [...candidates].sort((a, b) => {
			if (a.score !== b.score) return a.score - b.score;
			if (a.collisionCount !== b.collisionCount) return a.collisionCount - b.collisionCount;
			if (a.bends !== b.bends) return a.bends - b.bends;
			if (a.manhattanLength !== b.manhattanLength) return a.manhattanLength - b.manhattanLength;
			return a.points.length - b.points.length;
		});
		return sorted[0] ?? null;
	}

	private buildOffsetSequence(): number[] {
		const minSearchPixels = this.nodeWidth / 2 + OBSTACLE_PADDING + this.gridUnit * 2;
		const maxOffsetStep = Math.min(12, Math.max(5, Math.ceil(minSearchPixels / this.gridUnit)));
		const sequence: number[] = [0];
		for (let step = 1; step <= maxOffsetStep; step++) {
			sequence.push(-step, step);
		}
		return sequence;
	}

	private normalizePath(points: RoutePoint[]): RoutePoint[] {
		const snapped = points.map((point, index) => ({
			x: index === 0 || index === points.length - 1 ? point.x : this.snapToGrid(point.x),
			y: index === 0 || index === points.length - 1 ? point.y : this.snapToGrid(point.y)
		}));

		const deduped: RoutePoint[] = [];
		for (const point of snapped) {
			const prev = deduped[deduped.length - 1];
			if (!prev || prev.x !== point.x || prev.y !== point.y) {
				deduped.push(point);
			}
		}

		const simplified: RoutePoint[] = [];
		for (const point of deduped) {
			simplified.push(point);
			while (simplified.length >= 3) {
				const a = simplified[simplified.length - 3];
				const b = simplified[simplified.length - 2];
				const c = simplified[simplified.length - 1];
				if ((a.x === b.x && b.x === c.x) || (a.y === b.y && b.y === c.y)) {
					simplified.splice(simplified.length - 2, 1);
				} else {
					break;
				}
			}
		}

		if (simplified.length < 2 && deduped.length >= 2) {
			return [deduped[0], deduped[deduped.length - 1]];
		}
		return simplified;
	}

	private countCollisions(points: RoutePoint[], obstacles: Rect[]): number {
		if (points.length < 2) return 0;
		let collisions = 0;

		for (let i = 1; i < points.length; i++) {
			const a = points[i - 1];
			const b = points[i];
			for (const rect of obstacles) {
				if (this.segmentIntersectsRect(a, b, rect)) {
					collisions++;
				}
			}
		}

		return collisions;
	}

	private computeLength(points: RoutePoint[]): number {
		if (points.length < 2) return 0;
		let length = 0;
		for (let i = 1; i < points.length; i++) {
			length += Math.abs(points[i].x - points[i - 1].x) + Math.abs(points[i].y - points[i - 1].y);
		}
		return length;
	}

	private segmentIntersectsRect(a: RoutePoint, b: RoutePoint, rect: Rect): boolean {
		if (a.x === b.x) {
			const x = a.x;
			if (x < rect.minX || x > rect.maxX) return false;
			const minY = Math.min(a.y, b.y);
			const maxY = Math.max(a.y, b.y);
			return maxY >= rect.minY && minY <= rect.maxY;
		}

		if (a.y === b.y) {
			const y = a.y;
			if (y < rect.minY || y > rect.maxY) return false;
			const minX = Math.min(a.x, b.x);
			const maxX = Math.max(a.x, b.x);
			return maxX >= rect.minX && minX <= rect.maxX;
		}

		return false;
	}

	private snapToGrid(value: number): number {
		return Math.round(value / this.gridUnit) * this.gridUnit;
	}
}
