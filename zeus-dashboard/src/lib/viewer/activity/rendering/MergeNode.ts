// MergeNode - UML アクティビティ図の合流ノード（ひし形）
import type { ActivityNodeItem } from '$lib/types/api';
import { DiamondNode } from './DiamondNode';

/**
 * MergeNode - 合流ノード
 *
 * UML 表記: ひし形（Decision と同じ形状）
 * 複数の制御フローが合流する点を表す
 * Decision と視覚的には同じだが、意味的に区別される
 *
 * DiamondNode を継承（DecisionNode と共通の描画ロジック）
 */
export class MergeNode extends DiamondNode {
	constructor(nodeData: ActivityNodeItem) {
		super(nodeData);

		// 初回描画
		this.draw();
	}

	/**
	 * 合流ノードを描画
	 */
	draw(): void {
		// 基底クラスのひし形描画のみ
		this.drawDiamond();
	}
}
