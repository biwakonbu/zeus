// ノード間のエッジ描画クラス
// relation/layer に応じた意味的スタイル + インタラクション状態スタイルを統合
import type { GraphEdgeLayer, GraphEdgeRelation } from '$lib/types/api';
import { Graphics } from 'pixi.js';
import { EDGE_COLORS, EDGE_WIDTHS } from '$lib/viewer/shared/constants';

// エッジ状態（インタラクション）
export enum EdgeType {
	Normal = 'normal',
	Critical = 'critical',
	Blocked = 'blocked',
	Highlighted = 'highlighted'
}

// 矢印設定
const ARROW_SIZE = 8;
const ARROW_ANGLE = Math.PI / 6;
const FLOW_DOT_RADIUS = 2.4;
const FLOW_DOT_MIN_SPACING = 110;

interface SemanticStyle {
	core: number;
	outer: number;
	widthCore: number;
	widthOuter: number;
}

interface PolylinePoint {
	x: number;
	y: number;
}

function semanticStyle(layer: GraphEdgeLayer, relation: GraphEdgeRelation): SemanticStyle {
	const base = {
		depends_on: { core: 0xd0d0d0, outer: 0x4a4a4a },
		produces: { core: 0xffb366, outer: 0x8a5c2a },
		parent: { core: 0x88b8ff, outer: 0x355d96 },
		implements: { core: 0x66ccff, outer: 0x226688 },
		contributes: { core: 0xb48dff, outer: 0x5c4090 },
		fulfills: { core: 0x8be08b, outer: 0x2e7a2e }
	} as const;

	const rel = base[relation] ?? base.depends_on;
	if (layer === 'structural') {
		return {
			core: rel.core,
			outer: rel.outer,
			widthCore: 2,
			widthOuter: 4
		};
	}
	return {
		core: rel.core,
		outer: rel.outer,
		widthCore: 1.5,
		widthOuter: 3
	};
}

/**
 * GraphEdge - ノード間関係を視覚化
 */
export class GraphEdge extends Graphics {
	private fromId: string;
	private toId: string;
	private layer: GraphEdgeLayer;
	private relation: GraphEdgeRelation;
	private edgeType: EdgeType = EdgeType.Normal;

	private fromX = 0;
	private fromY = 0;
	private toX = 0;
	private toY = 0;
	private polyline: PolylinePoint[] | null = null;
	private flowPhase = 0;

	constructor(
		fromId: string,
		toId: string,
		layer: GraphEdgeLayer = 'reference',
		relation: GraphEdgeRelation = 'depends_on'
	) {
		super();
		this.fromId = fromId;
		this.toId = toId;
		this.layer = layer;
		this.relation = relation;
	}

	setEndpoints(fromX: number, fromY: number, toX: number, toY: number): void {
		this.fromX = fromX;
		this.fromY = fromY;
		this.toX = toX;
		this.toY = toY;
		this.polyline = null;
		this.draw();
	}

	setPolyline(points: PolylinePoint[]): void {
		const normalized = this.normalizePolyline(points);
		if (normalized.length >= 2) {
			this.polyline = normalized;
			const first = normalized[0];
			const last = normalized[normalized.length - 1];
			this.fromX = first.x;
			this.fromY = first.y;
			this.toX = last.x;
			this.toY = last.y;
		} else {
			this.polyline = null;
		}
		this.draw();
	}

	updateFlowAnimation(phase: number): void {
		if (!this.polyline || this.polyline.length < 2) return;
		if (Math.abs(this.flowPhase - phase) < 0.5) return;
		this.flowPhase = phase;
		this.draw();
	}

	setType(type: EdgeType): void {
		this.edgeType = type;
		this.draw();
	}

	setSemantic(layer: GraphEdgeLayer, relation: GraphEdgeRelation): void {
		this.layer = layer;
		this.relation = relation;
		this.draw();
	}

	getLayer(): GraphEdgeLayer {
		return this.layer;
	}

	getRelation(): GraphEdgeRelation {
		return this.relation;
	}

	draw(): void {
		this.clear();

		let core = 0;
		let outer = 0;
		let widthCore = 0;
		let widthOuter = 0;

		if (this.edgeType === EdgeType.Highlighted) {
			core = EDGE_COLORS.highlighted.core;
			outer = EDGE_COLORS.highlighted.outer;
			widthCore = EDGE_WIDTHS.highlighted.core;
			widthOuter = EDGE_WIDTHS.highlighted.outer;
		} else if (this.edgeType === EdgeType.Blocked) {
			core = EDGE_COLORS.blocked.core;
			outer = EDGE_COLORS.blocked.outer;
			widthCore = EDGE_WIDTHS.blocked.core;
			widthOuter = EDGE_WIDTHS.blocked.outer;
		} else if (this.edgeType === EdgeType.Critical) {
			core = EDGE_COLORS.critical.core;
			outer = EDGE_COLORS.critical.outer;
			widthCore = EDGE_WIDTHS.critical.core;
			widthOuter = EDGE_WIDTHS.critical.outer;
		} else {
			const semantic = semanticStyle(this.layer, this.relation);
			core = semantic.core;
			outer = semantic.outer;
			widthCore = semantic.widthCore;
			widthOuter = semantic.widthOuter;
		}

		if (this.polyline && this.polyline.length >= 2) {
			this.drawPolyline(this.polyline, widthOuter, outer, widthCore, core);
			this.drawFlowDots(this.polyline, core, outer);
			const arrowFrom = this.findArrowSource(this.polyline);
			this.drawArrow(arrowFrom.x, arrowFrom.y, this.toX, this.toY, core, outer);
			return;
		}

		const { cp1x, cp1y, cp2x, cp2y } = this.calculateControlPoints();
		this.moveTo(this.fromX, this.fromY);
		this.bezierCurveTo(cp1x, cp1y, cp2x, cp2y, this.toX, this.toY);
		this.stroke({ width: widthOuter, color: outer, alpha: this.layer === 'reference' ? 0.9 : 1 });

		this.moveTo(this.fromX, this.fromY);
		this.bezierCurveTo(cp1x, cp1y, cp2x, cp2y, this.toX, this.toY);
		this.stroke({ width: widthCore, color: core, alpha: 1 });

		this.drawArrow(cp2x, cp2y, this.toX, this.toY, core, outer);
	}

	private drawPolyline(
		points: PolylinePoint[],
		widthOuter: number,
		outer: number,
		widthCore: number,
		core: number
	): void {
		this.moveTo(points[0].x, points[0].y);
		for (let i = 1; i < points.length; i++) {
			this.lineTo(points[i].x, points[i].y);
		}
		this.stroke({ width: widthOuter, color: outer, alpha: this.layer === 'reference' ? 0.9 : 1 });

		this.moveTo(points[0].x, points[0].y);
		for (let i = 1; i < points.length; i++) {
			this.lineTo(points[i].x, points[i].y);
		}
		this.stroke({ width: widthCore, color: core, alpha: 1 });
	}

	private findArrowSource(points: PolylinePoint[]): PolylinePoint {
		for (let i = points.length - 2; i >= 0; i--) {
			const p = points[i];
			if (p.x !== this.toX || p.y !== this.toY) {
				return p;
			}
		}
		return points[0];
	}

	private normalizePolyline(points: PolylinePoint[]): PolylinePoint[] {
		const deduped: PolylinePoint[] = [];
		for (const point of points) {
			const prev = deduped[deduped.length - 1];
			if (!prev || prev.x !== point.x || prev.y !== point.y) {
				deduped.push(point);
			}
		}
		return deduped;
	}

	private drawFlowDots(points: PolylinePoint[], core: number, outer: number): void {
		const totalLength = this.computePolylineLength(points);
		if (totalLength < FLOW_DOT_MIN_SPACING) return;

		const spacing = Math.max(FLOW_DOT_MIN_SPACING, Math.floor(totalLength / 3));
		const dotCount = Math.max(2, Math.min(4, Math.floor(totalLength / spacing) + 1));
		const offset = ((this.flowPhase % spacing) + spacing) % spacing;

		for (let i = 0; i < dotCount; i++) {
			const distance = (offset + i * spacing) % totalLength;
			const point = this.getPointAtDistance(points, distance);
			this.circle(point.x, point.y, FLOW_DOT_RADIUS + 1);
			this.fill({ color: outer, alpha: 0.45 });
			this.circle(point.x, point.y, FLOW_DOT_RADIUS);
			this.fill({ color: core, alpha: 0.9 });
		}
	}

	private computePolylineLength(points: PolylinePoint[]): number {
		let length = 0;
		for (let i = 1; i < points.length; i++) {
			length += Math.abs(points[i].x - points[i - 1].x) + Math.abs(points[i].y - points[i - 1].y);
		}
		return length;
	}

	private getPointAtDistance(points: PolylinePoint[], distance: number): PolylinePoint {
		if (points.length === 0) return { x: 0, y: 0 };
		if (points.length === 1) return points[0];

		let remaining = distance;
		for (let i = 1; i < points.length; i++) {
			const from = points[i - 1];
			const to = points[i];
			const segmentLength = Math.abs(to.x - from.x) + Math.abs(to.y - from.y);
			if (segmentLength <= 0) continue;
			if (remaining <= segmentLength) {
				const ratio = remaining / segmentLength;
				return {
					x: from.x + (to.x - from.x) * ratio,
					y: from.y + (to.y - from.y) * ratio
				};
			}
			remaining -= segmentLength;
		}

		return points[points.length - 1];
	}

	private calculateControlPoints(): { cp1x: number; cp1y: number; cp2x: number; cp2y: number } {
		const dx = this.toX - this.fromX;
		const dy = this.toY - this.fromY;
		const distance = Math.sqrt(dx * dx + dy * dy);
		const curvature = Math.min(distance * 0.3, 100);

		if (Math.abs(dy) > Math.abs(dx)) {
			const sign = dy > 0 ? 1 : -1;
			return {
				cp1x: this.fromX,
				cp1y: this.fromY + curvature * sign,
				cp2x: this.toX,
				cp2y: this.toY - curvature * sign
			};
		}

		const sign = dx > 0 ? 1 : -1;
		return {
			cp1x: this.fromX + curvature * sign,
			cp1y: this.fromY,
			cp2x: this.toX - curvature * sign,
			cp2y: this.toY
		};
	}

	private drawArrow(fromX: number, fromY: number, toX: number, toY: number, core: number, outer: number): void {
		const dx = toX - fromX;
		const dy = toY - fromY;
		const angle = Math.atan2(dy, dx);

		const arrowX1 = toX - ARROW_SIZE * Math.cos(angle - ARROW_ANGLE);
		const arrowY1 = toY - ARROW_SIZE * Math.sin(angle - ARROW_ANGLE);
		const arrowX2 = toX - ARROW_SIZE * Math.cos(angle + ARROW_ANGLE);
		const arrowY2 = toY - ARROW_SIZE * Math.sin(angle + ARROW_ANGLE);

		this.moveTo(toX, toY);
		this.lineTo(arrowX1, arrowY1);
		this.lineTo(arrowX2, arrowY2);
		this.closePath();
		this.fill(outer);

		const innerScale = 0.7;
		const innerArrowX1 = toX - ARROW_SIZE * innerScale * Math.cos(angle - ARROW_ANGLE);
		const innerArrowY1 = toY - ARROW_SIZE * innerScale * Math.sin(angle - ARROW_ANGLE);
		const innerArrowX2 = toX - ARROW_SIZE * innerScale * Math.cos(angle + ARROW_ANGLE);
		const innerArrowY2 = toY - ARROW_SIZE * innerScale * Math.sin(angle + ARROW_ANGLE);

		this.moveTo(toX, toY);
		this.lineTo(innerArrowX1, innerArrowY1);
		this.lineTo(innerArrowX2, innerArrowY2);
		this.closePath();
		this.fill(core);
	}

	getFromId(): string {
		return this.fromId;
	}

	getToId(): string {
		return this.toId;
	}

	static createKey(
		fromId: string,
		toId: string,
		layer: GraphEdgeLayer,
		relation: GraphEdgeRelation
	): string {
		return `${fromId}-->${toId}::${layer}:${relation}`;
	}

	getKey(): string {
		return GraphEdge.createKey(this.fromId, this.toId, this.layer, this.relation);
	}
}

// 後方互換性エイリアス
export { GraphEdge as TaskEdge };

/**
 * EdgeFactory - 複数のエッジを効率的に管理
 */
export class EdgeFactory {
	private edges: Map<string, GraphEdge> = new Map();
	private nodeToEdges: Map<string, Set<string>> = new Map();

	getOrCreate(
		fromId: string,
		toId: string,
		layer: GraphEdgeLayer,
		relation: GraphEdgeRelation
	): GraphEdge {
		const key = GraphEdge.createKey(fromId, toId, layer, relation);
		let edge = this.edges.get(key);
		if (!edge) {
			edge = new GraphEdge(fromId, toId, layer, relation);
			this.edges.set(key, edge);
			this.addToIndex(fromId, key);
			this.addToIndex(toId, key);
		} else {
			edge.setSemantic(layer, relation);
		}
		return edge;
	}

	private addToIndex(nodeId: string, edgeKey: string): void {
		let edgeSet = this.nodeToEdges.get(nodeId);
		if (!edgeSet) {
			edgeSet = new Set();
			this.nodeToEdges.set(nodeId, edgeSet);
		}
		edgeSet.add(edgeKey);
	}

	private removeFromIndex(nodeId: string, edgeKey: string): void {
		const edgeSet = this.nodeToEdges.get(nodeId);
		if (edgeSet) {
			edgeSet.delete(edgeKey);
			if (edgeSet.size === 0) {
				this.nodeToEdges.delete(nodeId);
			}
		}
	}

	get(
		fromId: string,
		toId: string,
		layer: GraphEdgeLayer,
		relation: GraphEdgeRelation
	): GraphEdge | undefined {
		const key = GraphEdge.createKey(fromId, toId, layer, relation);
		return this.edges.get(key);
	}

	remove(fromId: string, toId: string, layer: GraphEdgeLayer, relation: GraphEdgeRelation): boolean {
		const key = GraphEdge.createKey(fromId, toId, layer, relation);
		const edge = this.edges.get(key);
		if (!edge) return false;

		this.removeFromIndex(fromId, key);
		this.removeFromIndex(toId, key);
		edge.destroy();
		this.edges.delete(key);
		return true;
	}

	getAll(): GraphEdge[] {
		return Array.from(this.edges.values());
	}

	clear(): void {
		for (const edge of this.edges.values()) {
			edge.destroy();
		}
		this.edges.clear();
		this.nodeToEdges.clear();
	}

	getEdgesForNode(nodeId: string): GraphEdge[] {
		const edgeKeys = this.nodeToEdges.get(nodeId);
		if (!edgeKeys) return [];

		const result: GraphEdge[] = [];
		for (const key of edgeKeys) {
			const edge = this.edges.get(key);
			if (edge) result.push(edge);
		}
		return result;
	}

	getEdgeCountForNode(nodeId: string): number {
		return this.nodeToEdges.get(nodeId)?.size ?? 0;
	}
}
