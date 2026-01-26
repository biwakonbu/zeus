// DecisionNode - UML アクティビティ図の分岐ノード（ひし形）
import { Text } from 'pixi.js';
import type { ActivityNodeItem } from '$lib/types/api';
import { DiamondNode } from './DiamondNode';
import { NODE_COLORS, TEXT_RESOLUTION } from './constants';

/**
 * DecisionNode - 分岐ノード
 *
 * UML 表記: ひし形
 * 条件分岐点を表す
 *
 * DiamondNode を継承（MergeNode と共通の描画ロジック）
 * 追加機能: 条件名のテキスト表示
 */
export class DecisionNode extends DiamondNode {
	private nameText: Text | null = null;

	constructor(nodeData: ActivityNodeItem) {
		super(nodeData);

		// 名前がある場合はテキストコンポーネント作成
		if (nodeData.name) {
			this.nameText = new Text({
				text: '',
				style: {
					fontSize: 9,
					fill: NODE_COLORS.decision.text,
					fontFamily: 'IBM Plex Mono, monospace',
					align: 'center'
				},
				resolution: TEXT_RESOLUTION
			});
			this.addChild(this.nameText);
		}

		// 初回描画
		this.draw();
	}

	/**
	 * 分岐ノードを描画
	 */
	draw(): void {
		// 基底クラスのひし形描画
		this.drawDiamond();

		// テキスト描画
		this.drawText();
	}

	/**
	 * テキストを描画
	 */
	private drawText(): void {
		if (!this.nameText || !this.nodeData.name) return;

		// 全文表示
		this.nameText.text = this.nodeData.name;

		// ひし形の下部に配置（中央揃え）
		this.nameText.x = (this.size - this.nameText.width) / 2;
		this.nameText.y = this.size + 4;
	}
}
