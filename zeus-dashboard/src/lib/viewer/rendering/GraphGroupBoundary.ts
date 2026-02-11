// Graph View のグループ境界描画
import { Container, Graphics, Text } from 'pixi.js';
import type { FederatedPointerEvent } from 'pixi.js';
import type { LayoutGroupBounds } from '../engine/LayoutEngine';

const CORNER_RADIUS = 14;
const BORDER_WIDTH = 2;
const TITLE_HEIGHT = 24;
const TITLE_PADDING_X = 12;
const TEXT_RESOLUTION =
	typeof window !== 'undefined' ? Math.min(window.devicePixelRatio * 2, 4) : 2;

/**
 * グループデータ（外部に公開する情報）
 */
export interface GroupData {
	id: string;
	title: string;
	description?: string;
	goals?: string[];
	status?: string;
	nodeCount: number;
	nodeIds?: string[];
}

/**
 * GraphGroupBoundary - グループ境界の視覚表現（選択・ホバー対応）
 */
export class GraphGroupBoundary extends Container {
	private bounds: LayoutGroupBounds;
	private boundaryGraphics: Graphics;
	private labelContainer: Container;
	private titleGraphics: Graphics;
	private labelText: Text;
	private selected = false;
	private hovered = false;

	private onClickCallback: ((group: GraphGroupBoundary, event?: FederatedPointerEvent) => void) | null = null;
	private onHoverCallback: ((group: GraphGroupBoundary, isHovered: boolean) => void) | null = null;

	constructor(bounds: LayoutGroupBounds) {
		super();

		this.bounds = bounds;
		this.boundaryGraphics = new Graphics();
		this.labelContainer = new Container();
		this.titleGraphics = new Graphics();
		this.labelText = new Text({
			text: '',
			style: {
				fontSize: 10,
				fill: 0xffffff,
				fontFamily: 'IBM Plex Mono, monospace',
				fontWeight: 'bold'
			},
			resolution: TEXT_RESOLUTION
		});

		this.eventMode = 'static';
		this.cursor = 'pointer';
		this.labelContainer.eventMode = 'none';
		this.addChild(this.boundaryGraphics);
		this.labelContainer.addChild(this.titleGraphics);
		this.labelContainer.addChild(this.labelText);

		// インタラクションイベント
		this.on('pointerover', this.handlePointerOver, this);
		this.on('pointerout', this.handlePointerOut, this);
		this.on('pointertap', this.handlePointerTap, this);

		this.draw();
	}

	update(bounds: LayoutGroupBounds): void {
		this.bounds = bounds;
		this.draw();
	}

	getGroupId(): LayoutGroupBounds['groupId'] {
		return this.bounds.groupId;
	}

	getLabelContainer(): Container {
		return this.labelContainer;
	}

	getGroupData(): GroupData {
		return {
			id: this.bounds.groupId,
			title: this.bounds.label,
			description: this.bounds.description,
			goals: this.bounds.goals,
			status: this.bounds.status,
			nodeCount: this.bounds.nodeCount,
			nodeIds: this.bounds.nodeIds
		};
	}

	setSelected(selected: boolean): void {
		if (this.selected === selected) return;
		this.selected = selected;
		this.draw();
	}

	isSelected(): boolean {
		return this.selected;
	}

	onClick(callback: (group: GraphGroupBoundary, event?: FederatedPointerEvent) => void): void {
		this.onClickCallback = callback;
	}

	onHover(callback: (group: GraphGroupBoundary, isHovered: boolean) => void): void {
		this.onHoverCallback = callback;
	}

	private handlePointerOver(): void {
		this.hovered = true;
		this.draw();
		this.onHoverCallback?.(this, true);
	}

	private handlePointerOut(): void {
		this.hovered = false;
		this.draw();
		this.onHoverCallback?.(this, false);
	}

	private handlePointerTap(event: FederatedPointerEvent): void {
		this.onClickCallback?.(this, event);
	}

	destroy(): void {
		this.off('pointerover', this.handlePointerOver, this);
		this.off('pointerout', this.handlePointerOut, this);
		this.off('pointertap', this.handlePointerTap, this);
		this.onClickCallback = null;
		this.onHoverCallback = null;
		super.destroy({ children: true });
	}

	private draw(): void {
		this.boundaryGraphics.clear();
		this.titleGraphics.clear();

		const { minX, minY, width, height, color, label } = this.bounds;
		this.labelText.text = label;

		// 状態に応じたスタイル
		let fillAlpha = 0.08;
		let borderAlpha = 0.35;
		let borderWidth = BORDER_WIDTH;

		if (this.selected) {
			fillAlpha = 0.15;
			borderAlpha = 0.8;
			borderWidth = 3;
		} else if (this.hovered) {
			fillAlpha = 0.12;
			borderAlpha = 0.6;
		}

		// 背景
		this.boundaryGraphics.roundRect(minX, minY, width, height, CORNER_RADIUS);
		this.boundaryGraphics.fill({ color, alpha: fillAlpha });
		this.boundaryGraphics.stroke({ width: borderWidth, color, alpha: borderAlpha });

		// タイトルバー
		const titleWidth = Math.max(56, this.labelText.width + TITLE_PADDING_X * 2);
		this.titleGraphics.roundRect(minX + 8, minY + 6, titleWidth, TITLE_HEIGHT, 8);
		this.titleGraphics.fill({ color, alpha: 0.22 });
		this.titleGraphics.stroke({ width: 1, color, alpha: 0.5 });

		this.labelText.x = minX + 8 + TITLE_PADDING_X;
		this.labelText.y = minY + 6 + (TITLE_HEIGHT - this.labelText.height) / 2;
	}
}
