// Graph View のグループ境界描画
import { Container, Graphics, Text } from 'pixi.js';
import type { LayoutGroupBounds } from '../engine/LayoutEngine';

const CORNER_RADIUS = 14;
const BORDER_WIDTH = 2;
const TITLE_HEIGHT = 24;
const TITLE_PADDING_X = 12;
const TEXT_RESOLUTION =
	typeof window !== 'undefined' ? Math.min(window.devicePixelRatio * 2, 4) : 2;

/**
 * GraphGroupBoundary - グループ境界の視覚表現
 */
export class GraphGroupBoundary extends Container {
	private bounds: LayoutGroupBounds;
	private graphics: Graphics;
	private labelText: Text;

	constructor(bounds: LayoutGroupBounds) {
		super();

		this.bounds = bounds;
		this.graphics = new Graphics();
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

		this.eventMode = 'none';
		this.addChild(this.graphics);
		this.addChild(this.labelText);
		this.draw();
	}

	update(bounds: LayoutGroupBounds): void {
		this.bounds = bounds;
		this.draw();
	}

	getGroupId(): LayoutGroupBounds['groupId'] {
		return this.bounds.groupId;
	}

	private draw(): void {
		this.graphics.clear();

		const { minX, minY, width, height, color, label } = this.bounds;
		this.labelText.text = label;

		// 背景
		this.graphics.roundRect(minX, minY, width, height, CORNER_RADIUS);
		this.graphics.fill({ color, alpha: 0.08 });
		this.graphics.stroke({ width: BORDER_WIDTH, color, alpha: 0.35 });

		// タイトルバー
		const titleWidth = Math.max(56, this.labelText.width + TITLE_PADDING_X * 2);
		this.graphics.roundRect(minX + 8, minY + 6, titleWidth, TITLE_HEIGHT, 8);
		this.graphics.fill({ color, alpha: 0.22 });
		this.graphics.stroke({ width: 1, color, alpha: 0.5 });

		this.labelText.x = minX + 8 + TITLE_PADDING_X;
		this.labelText.y = minY + 6 + (TITLE_HEIGHT - this.labelText.height) / 2;
	}
}
