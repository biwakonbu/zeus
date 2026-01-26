// タスク間のエッジ（依存関係）描画クラス
// 3層構造（グロー → 外側 → コア）で電気回路風の発光感を表現
import { Graphics } from 'pixi.js';
import { EDGE_COLORS, EDGE_WIDTHS } from '$lib/viewer/shared/constants';

// エッジタイプ
export enum EdgeType {
	// 通常の依存関係
	Normal = 'normal',
	// クリティカルパス上のエッジ
	Critical = 'critical',
	// ブロックされている（依存先が未完了）
	Blocked = 'blocked',
	// ハイライト（選択時）
	Highlighted = 'highlighted'
}

// 矢印設定
const ARROW_SIZE = 12;
const ARROW_ANGLE = Math.PI / 6; // 30度

/**
 * TaskEdge - タスク間の依存関係を視覚化
 *
 * 責務:
 * - 2つのノード間のエッジを描画
 * - エッジタイプに応じたスタイリング
 * - 曲線パスの計算（交差を減らす）
 */
export class TaskEdge extends Graphics {
	private fromId: string;
	private toId: string;
	private edgeType: EdgeType = EdgeType.Normal;

	// 座標
	private fromX: number = 0;
	private fromY: number = 0;
	private toX: number = 0;
	private toY: number = 0;

	constructor(fromId: string, toId: string) {
		super();

		this.fromId = fromId;
		this.toId = toId;
	}

	/**
	 * エッジの両端の座標を設定
	 */
	setEndpoints(fromX: number, fromY: number, toX: number, toY: number): void {
		this.fromX = fromX;
		this.fromY = fromY;
		this.toX = toX;
		this.toY = toY;
		this.draw();
	}

	/**
	 * エッジタイプを設定
	 */
	setType(type: EdgeType): void {
		this.edgeType = type;
		this.draw();
	}

	/**
	 * エッジを描画（3層構造: グロー → 外側 → コア）
	 * 電気回路風の発光感を表現
	 */
	draw(): void {
		this.clear();

		const style = EDGE_COLORS[this.edgeType];
		const widths = EDGE_WIDTHS[this.edgeType];

		// ベジェ曲線のコントロールポイントを計算
		const { cp1x, cp1y, cp2x, cp2y } = this.calculateControlPoints();

		// Step 1: グロー（最外層）- 淡いハロー効果
		// glowAlpha をそのまま使用（乗算なし）
		this.moveTo(this.fromX, this.fromY);
		this.bezierCurveTo(cp1x, cp1y, cp2x, cp2y, this.toX, this.toY);
		this.stroke({ width: widths.outer + 6, color: style.glow, alpha: style.glowAlpha * 0.5 });

		// Step 2: 外側（縁取り）- 暗めの縁取りでコアを際立たせる
		this.moveTo(this.fromX, this.fromY);
		this.bezierCurveTo(cp1x, cp1y, cp2x, cp2y, this.toX, this.toY);
		this.stroke({ width: widths.outer, color: style.outer, alpha: 1.0 });

		// Step 3: コア（内側）- 明るいコア線
		this.moveTo(this.fromX, this.fromY);
		this.bezierCurveTo(cp1x, cp1y, cp2x, cp2y, this.toX, this.toY);
		this.stroke({ width: widths.core, color: style.core, alpha: 1.0 });

		// 矢印を描画
		this.drawArrow(cp2x, cp2y, this.toX, this.toY, style);
	}

	/**
	 * ベジェ曲線のコントロールポイントを計算
	 */
	private calculateControlPoints(): { cp1x: number; cp1y: number; cp2x: number; cp2y: number } {
		const dx = this.toX - this.fromX;
		const dy = this.toY - this.fromY;
		const distance = Math.sqrt(dx * dx + dy * dy);

		// 曲線の強さ（距離に比例）
		const curvature = Math.min(distance * 0.3, 100);

		// Y方向が主な場合（上から下への流れ）
		if (Math.abs(dy) > Math.abs(dx)) {
			const sign = dy > 0 ? 1 : -1;
			return {
				cp1x: this.fromX,
				cp1y: this.fromY + curvature * sign,
				cp2x: this.toX,
				cp2y: this.toY - curvature * sign
			};
		}

		// X方向が主な場合（横方向の流れ）
		const sign = dx > 0 ? 1 : -1;
		return {
			cp1x: this.fromX + curvature * sign,
			cp1y: this.fromY,
			cp2x: this.toX - curvature * sign,
			cp2y: this.toY
		};
	}

	/**
	 * 矢印を描画（3層構造対応）
	 */
	private drawArrow(
		fromX: number,
		fromY: number,
		toX: number,
		toY: number,
		style: (typeof EDGE_COLORS)[keyof typeof EDGE_COLORS]
	): void {
		const widths = EDGE_WIDTHS[this.edgeType];

		// 方向ベクトルを計算
		const dx = toX - fromX;
		const dy = toY - fromY;
		const angle = Math.atan2(dy, dx);

		// 矢印の先端
		const arrowX1 = toX - ARROW_SIZE * Math.cos(angle - ARROW_ANGLE);
		const arrowY1 = toY - ARROW_SIZE * Math.sin(angle - ARROW_ANGLE);
		const arrowX2 = toX - ARROW_SIZE * Math.cos(angle + ARROW_ANGLE);
		const arrowY2 = toY - ARROW_SIZE * Math.sin(angle + ARROW_ANGLE);

		// 矢印も3層で描画

		// Step 1: グロー
		this.moveTo(toX, toY);
		this.lineTo(arrowX1, arrowY1);
		this.moveTo(toX, toY);
		this.lineTo(arrowX2, arrowY2);
		this.stroke({ width: widths.outer + 4, color: style.glow, alpha: style.glowAlpha * 0.5 });

		// Step 2: 外側（塗りつぶし三角形）
		this.moveTo(toX, toY);
		this.lineTo(arrowX1, arrowY1);
		this.lineTo(arrowX2, arrowY2);
		this.closePath();
		this.fill(style.outer);

		// Step 3: コア（内側の明るい三角形）
		const innerScale = 0.7;
		const innerArrowX1 = toX - ARROW_SIZE * innerScale * Math.cos(angle - ARROW_ANGLE);
		const innerArrowY1 = toY - ARROW_SIZE * innerScale * Math.sin(angle - ARROW_ANGLE);
		const innerArrowX2 = toX - ARROW_SIZE * innerScale * Math.cos(angle + ARROW_ANGLE);
		const innerArrowY2 = toY - ARROW_SIZE * innerScale * Math.sin(angle + ARROW_ANGLE);

		this.moveTo(toX, toY);
		this.lineTo(innerArrowX1, innerArrowY1);
		this.lineTo(innerArrowX2, innerArrowY2);
		this.closePath();
		this.fill(style.core);
	}

	/**
	 * From ノード ID を取得
	 */
	getFromId(): string {
		return this.fromId;
	}

	/**
	 * To ノード ID を取得
	 */
	getToId(): string {
		return this.toId;
	}

	/**
	 * エッジの識別キーを生成
	 */
	static createKey(fromId: string, toId: string): string {
		return `${fromId}-->${toId}`;
	}

	/**
	 * このエッジのキーを取得
	 */
	getKey(): string {
		return TaskEdge.createKey(this.fromId, this.toId);
	}
}

/**
 * エッジファクトリー - 複数のエッジを効率的に管理
 * ノード→エッジのインデックスにより O(1) でエッジを取得可能
 */
export class EdgeFactory {
	private edges: Map<string, TaskEdge> = new Map();
	// ノードID → 関連するエッジキーのセット（高速検索用インデックス）
	private nodeToEdges: Map<string, Set<string>> = new Map();

	/**
	 * エッジを作成または取得
	 */
	getOrCreate(fromId: string, toId: string): TaskEdge {
		const key = TaskEdge.createKey(fromId, toId);

		let edge = this.edges.get(key);
		if (!edge) {
			edge = new TaskEdge(fromId, toId);
			this.edges.set(key, edge);

			// インデックスを更新
			this.addToIndex(fromId, key);
			this.addToIndex(toId, key);
		}

		return edge;
	}

	/**
	 * インデックスにエッジを追加
	 */
	private addToIndex(nodeId: string, edgeKey: string): void {
		let edgeSet = this.nodeToEdges.get(nodeId);
		if (!edgeSet) {
			edgeSet = new Set();
			this.nodeToEdges.set(nodeId, edgeSet);
		}
		edgeSet.add(edgeKey);
	}

	/**
	 * インデックスからエッジを削除
	 */
	private removeFromIndex(nodeId: string, edgeKey: string): void {
		const edgeSet = this.nodeToEdges.get(nodeId);
		if (edgeSet) {
			edgeSet.delete(edgeKey);
			if (edgeSet.size === 0) {
				this.nodeToEdges.delete(nodeId);
			}
		}
	}

	/**
	 * エッジを取得
	 */
	get(fromId: string, toId: string): TaskEdge | undefined {
		const key = TaskEdge.createKey(fromId, toId);
		return this.edges.get(key);
	}

	/**
	 * エッジを削除
	 */
	remove(fromId: string, toId: string): boolean {
		const key = TaskEdge.createKey(fromId, toId);
		const edge = this.edges.get(key);
		if (edge) {
			// インデックスから削除
			this.removeFromIndex(fromId, key);
			this.removeFromIndex(toId, key);

			edge.destroy();
			this.edges.delete(key);
			return true;
		}
		return false;
	}

	/**
	 * 全エッジを取得
	 */
	getAll(): TaskEdge[] {
		return Array.from(this.edges.values());
	}

	/**
	 * 全エッジを削除
	 */
	clear(): void {
		for (const edge of this.edges.values()) {
			edge.destroy();
		}
		this.edges.clear();
		this.nodeToEdges.clear();
	}

	/**
	 * ノードに関連する全エッジを取得（O(1) インデックス検索）
	 */
	getEdgesForNode(nodeId: string): TaskEdge[] {
		const edgeKeys = this.nodeToEdges.get(nodeId);
		if (!edgeKeys) return [];

		const result: TaskEdge[] = [];
		for (const key of edgeKeys) {
			const edge = this.edges.get(key);
			if (edge) {
				result.push(edge);
			}
		}
		return result;
	}

	/**
	 * ノードに関連するエッジの数を取得（O(1)）
	 */
	getEdgeCountForNode(nodeId: string): number {
		return this.nodeToEdges.get(nodeId)?.size ?? 0;
	}
}
