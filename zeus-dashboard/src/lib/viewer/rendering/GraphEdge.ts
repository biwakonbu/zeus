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

interface SemanticStyle {
	core: number;
	outer: number;
	widthCore: number;
	widthOuter: number;
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

		const { cp1x, cp1y, cp2x, cp2y } = this.calculateControlPoints();

		this.moveTo(this.fromX, this.fromY);
		this.bezierCurveTo(cp1x, cp1y, cp2x, cp2y, this.toX, this.toY);
		this.stroke({ width: widthOuter, color: outer, alpha: this.layer === 'reference' ? 0.9 : 1 });

		this.moveTo(this.fromX, this.fromY);
		this.bezierCurveTo(cp1x, cp1y, cp2x, cp2y, this.toX, this.toY);
		this.stroke({ width: widthCore, color: core, alpha: 1 });

		this.drawArrow(cp2x, cp2y, this.toX, this.toY, core, outer);
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
