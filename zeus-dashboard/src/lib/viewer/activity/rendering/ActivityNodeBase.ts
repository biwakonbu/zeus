// ActivityNodeBase - アクティビティ図ノードの基底クラス
// 共通の状態管理とインタラクション処理を提供
import { Container, Graphics } from 'pixi.js';
import type { FederatedPointerEvent } from 'pixi.js';
import type { ActivityNodeItem } from '$lib/types/api';
import { COMMON_COLORS } from './constants';

/**
 * ActivityNodeBase - 全アクティビティノードの基底クラス
 *
 * 責務:
 * - 共通の状態管理（選択、ホバー）
 * - インタラクションイベント処理
 * - 子クラスへの描画委譲
 */
export abstract class ActivityNodeBase extends Container {
	protected nodeData: ActivityNodeItem;
	protected background: Graphics;

	protected isHovered = false;
	protected isSelected = false;

	// イベントコールバック
	private onClickCallback?: (node: ActivityNodeBase, event?: FederatedPointerEvent) => void;
	private onHoverCallback?: (node: ActivityNodeBase, isHovered: boolean, event?: MouseEvent) => void;

	constructor(nodeData: ActivityNodeItem) {
		super();

		this.nodeData = nodeData;

		// 背景 Graphics を初期化
		this.background = new Graphics();
		this.addChild(this.background);

		// インタラクション設定
		this.eventMode = 'static';
		this.cursor = 'pointer';

		this.on('pointerover', (e: FederatedPointerEvent) => this.handleHover(true, e));
		this.on('pointerout', () => this.handleHover(false));
		this.on('pointertap', (e: FederatedPointerEvent) => this.handleClick(e));
	}

	/**
	 * ノードを描画（子クラスで実装）
	 */
	abstract draw(): void;

	/**
	 * ノードの幅を取得（子クラスで実装）
	 */
	abstract getNodeWidth(): number;

	/**
	 * ノードの高さを取得（子クラスで実装）
	 */
	abstract getNodeHeight(): number;

	/**
	 * ホバー時の背景色を取得
	 */
	protected getBackgroundColor(): number {
		if (this.isSelected) {
			return COMMON_COLORS.backgroundSelected;
		} else if (this.isHovered) {
			return COMMON_COLORS.backgroundHover;
		}
		return COMMON_COLORS.background;
	}

	/**
	 * ホバー時のボーダー色を取得
	 */
	protected getBorderColor(): number {
		if (this.isSelected) {
			return COMMON_COLORS.borderSelected;
		} else if (this.isHovered) {
			return COMMON_COLORS.borderHover;
		}
		return COMMON_COLORS.border;
	}

	/**
	 * ボーダー幅を取得
	 */
	protected getBorderWidth(): number {
		return this.isSelected ? 3 : 2;
	}

	/**
	 * ホバー処理
	 */
	private handleHover(isHovered: boolean, event?: FederatedPointerEvent): void {
		this.isHovered = isHovered;
		this.draw();
		const mouseEvent = event?.nativeEvent as MouseEvent | undefined;
		this.onHoverCallback?.(this, isHovered, mouseEvent);
	}

	/**
	 * クリック処理
	 */
	private handleClick(e: FederatedPointerEvent): void {
		e.stopPropagation();
		this.onClickCallback?.(this, e);
	}

	/**
	 * 選択状態を設定
	 */
	setSelected(selected: boolean): void {
		if (this.isSelected !== selected) {
			this.isSelected = selected;
			this.draw();
		}
	}

	/**
	 * ノードIDを取得
	 */
	getNodeId(): string {
		return this.nodeData.id;
	}

	/**
	 * ノードタイプを取得
	 */
	getNodeType(): string {
		return this.nodeData.type;
	}

	/**
	 * ノードデータを取得
	 */
	getNodeData(): ActivityNodeItem {
		return this.nodeData;
	}

	/**
	 * ノードデータを更新
	 */
	updateNodeData(nodeData: ActivityNodeItem): void {
		this.nodeData = nodeData;
		this.draw();
	}

	/**
	 * イベントリスナーを設定
	 */
	onClick(callback: (node: ActivityNodeBase, event?: FederatedPointerEvent) => void): void {
		this.onClickCallback = callback;
	}

	onHover(callback: (node: ActivityNodeBase, isHovered: boolean, event?: MouseEvent) => void): void {
		this.onHoverCallback = callback;
	}

	/**
	 * 中心座標を取得（エッジ接続用）
	 */
	getCenterX(): number {
		return this.x + this.getNodeWidth() / 2;
	}

	getCenterY(): number {
		return this.y + this.getNodeHeight() / 2;
	}
}
