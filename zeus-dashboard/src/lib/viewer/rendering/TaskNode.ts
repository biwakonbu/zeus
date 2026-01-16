// タスクノードの描画クラス
import { Container, Graphics, Text, FederatedPointerEvent } from 'pixi.js';
import type { TaskItem, TaskStatus, Priority } from '$lib/types/api';

// ノードサイズ定数
const NODE_WIDTH = 180;
const NODE_HEIGHT = 70;
const CORNER_RADIUS = 4;
const PROGRESS_BAR_HEIGHT = 8;
const PADDING = 10;

// 色定義（Factorioテーマに準拠）
const COLORS = {
	// ステータス色
	status: {
		completed: 0x44cc44,
		in_progress: 0x4488ff,
		pending: 0x888888,
		blocked: 0xee4444
	},
	// 優先度色
	priority: {
		high: 0xee4444,
		medium: 0xffcc00,
		low: 0x44cc44
	},
	// 基本色
	background: 0x2d2d2d,
	backgroundHover: 0x3a3a3a,
	backgroundSelected: 0x4a4a4a,
	border: 0x4a4a4a,
	borderHighlight: 0xff9533,
	text: 0xffffff,
	textSecondary: 0xb8b8b8,
	textMuted: 0x888888,
	progressBg: 0x1a1a1a
};

// LOD レベル
export enum LODLevel {
	// 最大ズームアウト：色付きの四角のみ
	Macro = 0,
	// 中間：ステータス + ID のみ
	Meso = 1,
	// 最大ズームイン：全情報表示
	Micro = 2
}

/**
 * TaskNode - タスクの視覚的表現
 *
 * 責務:
 * - タスクのグラフィカル表示
 * - インタラクション（クリック、ホバー）
 * - LOD（詳細度）に応じた表示切り替え
 */
export class TaskNode extends Container {
	private task: TaskItem;
	private background: Graphics;
	private statusIndicator: Graphics;
	private idText: Text;
	private titleText: Text;
	private progressBar: Graphics;
	private metaText: Text;

	private isHovered = false;
	private isSelected = false;
	private currentLOD: LODLevel = LODLevel.Micro;

	// イベントコールバック
	private onClickCallback?: (node: TaskNode, event?: FederatedPointerEvent) => void;
	private onHoverCallback?: (node: TaskNode, isHovered: boolean) => void;

	// 進捗率（0-100）- タスク自体には進捗がないので、ステータスから推定
	private progress: number;

	constructor(task: TaskItem) {
		super();

		this.task = task;
		this.progress = this.estimateProgress(task.status);

		// コンポーネント初期化
		this.background = new Graphics();
		this.statusIndicator = new Graphics();
		this.idText = new Text({ text: '', style: { fontSize: 12, fill: COLORS.text, fontFamily: 'IBM Plex Mono, monospace' } });
		this.titleText = new Text({ text: '', style: { fontSize: 11, fill: COLORS.textSecondary, fontFamily: 'IBM Plex Mono, monospace' } });
		this.progressBar = new Graphics();
		this.metaText = new Text({ text: '', style: { fontSize: 10, fill: COLORS.textMuted, fontFamily: 'IBM Plex Mono, monospace' } });

		this.addChild(this.background);
		this.addChild(this.statusIndicator);
		this.addChild(this.idText);
		this.addChild(this.titleText);
		this.addChild(this.progressBar);
		this.addChild(this.metaText);

		// インタラクション設定
		this.eventMode = 'static';
		this.cursor = 'pointer';

		this.on('pointerover', () => this.handleHover(true));
		this.on('pointerout', () => this.handleHover(false));
		this.on('pointertap', (e: FederatedPointerEvent) => this.handleClick(e));

		// 初回描画
		this.draw();
	}

	/**
	 * ステータスから進捗率を推定
	 */
	private estimateProgress(status: TaskStatus): number {
		switch (status) {
			case 'completed': return 100;
			case 'in_progress': return 50;
			case 'pending': return 0;
			case 'blocked': return 0;
			default: return 0;
		}
	}

	/**
	 * ノードを描画
	 */
	draw(): void {
		this.drawBackground();
		this.drawStatusIndicator();
		this.drawTexts();
		this.drawProgressBar();
	}

	/**
	 * 背景を描画
	 */
	private drawBackground(): void {
		this.background.clear();

		let bgColor = COLORS.background;
		let borderColor = COLORS.border;

		if (this.isSelected) {
			bgColor = COLORS.backgroundSelected;
			borderColor = COLORS.borderHighlight;
		} else if (this.isHovered) {
			bgColor = COLORS.backgroundHover;
			borderColor = COLORS.borderHighlight;
		}

		// 背景
		this.background.roundRect(0, 0, NODE_WIDTH, NODE_HEIGHT, CORNER_RADIUS);
		this.background.fill(bgColor);
		this.background.stroke({ width: 2, color: borderColor });

		// 金属フレーム効果（上部ハイライト）
		this.background.moveTo(CORNER_RADIUS, 1);
		this.background.lineTo(NODE_WIDTH - CORNER_RADIUS, 1);
		this.background.stroke({ width: 1, color: 0x666666, alpha: 0.5 });
	}

	/**
	 * ステータスインジケーターを描画
	 */
	private drawStatusIndicator(): void {
		this.statusIndicator.clear();

		const statusColor = COLORS.status[this.task.status] || COLORS.status.pending;

		// 左側のステータスバー
		this.statusIndicator.rect(0, 0, 4, NODE_HEIGHT);
		this.statusIndicator.fill(statusColor);

		// ステータスドット
		this.statusIndicator.circle(PADDING + 6, PADDING + 6, 4);
		this.statusIndicator.fill(statusColor);
	}

	/**
	 * テキストを描画
	 */
	private drawTexts(): void {
		if (this.currentLOD === LODLevel.Macro) {
			// マクロレベルでは非表示
			this.idText.visible = false;
			this.titleText.visible = false;
			this.metaText.visible = false;
			return;
		}

		this.idText.visible = true;

		// ID テキスト
		const shortId = this.task.id.length > 10 ? this.task.id.substring(0, 10) + '...' : this.task.id;
		this.idText.text = shortId;
		this.idText.x = PADDING + 14;
		this.idText.y = PADDING;

		if (this.currentLOD === LODLevel.Meso) {
			// メソレベルではIDのみ
			this.titleText.visible = false;
			this.metaText.visible = false;
			return;
		}

		// マイクロレベルでは全情報表示
		this.titleText.visible = true;
		this.metaText.visible = true;

		// タイトル（省略）
		const maxTitleLength = 18;
		const title = this.task.title.length > maxTitleLength
			? this.task.title.substring(0, maxTitleLength) + '...'
			: this.task.title;
		this.titleText.text = title;
		this.titleText.x = PADDING;
		this.titleText.y = PADDING + 20;

		// メタ情報（担当者）
		const assignee = this.task.assignee || 'unassigned';
		this.metaText.text = `@${assignee}`;
		this.metaText.x = PADDING;
		this.metaText.y = NODE_HEIGHT - PADDING - 10;
	}

	/**
	 * プログレスバーを描画
	 */
	private drawProgressBar(): void {
		this.progressBar.clear();

		if (this.currentLOD !== LODLevel.Micro) {
			this.progressBar.visible = false;
			return;
		}

		this.progressBar.visible = true;

		const barWidth = NODE_WIDTH - PADDING * 2;
		const barY = PADDING + 38;

		// 背景
		this.progressBar.roundRect(PADDING, barY, barWidth, PROGRESS_BAR_HEIGHT, 2);
		this.progressBar.fill(COLORS.progressBg);

		// 進捗
		if (this.progress > 0) {
			const progressWidth = (barWidth * this.progress) / 100;
			const progressColor = COLORS.status[this.task.status] || COLORS.status.pending;

			this.progressBar.roundRect(PADDING, barY, progressWidth, PROGRESS_BAR_HEIGHT, 2);
			this.progressBar.fill(progressColor);
		}
	}

	/**
	 * LODレベルを設定
	 */
	setLOD(level: LODLevel): void {
		if (this.currentLOD !== level) {
			this.currentLOD = level;
			this.draw();
		}
	}

	/**
	 * ホバー処理
	 */
	private handleHover(isHovered: boolean): void {
		this.isHovered = isHovered;
		this.drawBackground();
		this.onHoverCallback?.(this, isHovered);
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
		this.isSelected = selected;
		this.drawBackground();
	}

	/**
	 * タスクデータを更新
	 */
	updateTask(task: TaskItem): void {
		this.task = task;
		this.progress = this.estimateProgress(task.status);
		this.draw();
	}

	/**
	 * タスクIDを取得
	 */
	getTaskId(): string {
		return this.task.id;
	}

	/**
	 * タスクデータを取得
	 */
	getTask(): TaskItem {
		return this.task;
	}

	/**
	 * ノードの幅を取得
	 */
	static getWidth(): number {
		return NODE_WIDTH;
	}

	/**
	 * ノードの高さを取得
	 */
	static getHeight(): number {
		return NODE_HEIGHT;
	}

	/**
	 * イベントリスナーを設定
	 */
	onClick(callback: (node: TaskNode, event?: FederatedPointerEvent) => void): void {
		this.onClickCallback = callback;
	}

	onHover(callback: (node: TaskNode, isHovered: boolean) => void): void {
		this.onHoverCallback = callback;
	}
}
