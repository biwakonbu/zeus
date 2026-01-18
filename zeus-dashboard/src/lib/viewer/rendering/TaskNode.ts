// タスクノードの描画クラス
import { Container, Graphics, Text, FederatedPointerEvent } from 'pixi.js';
import type { TaskItem, TaskStatus, Priority } from '$lib/types/api';

// ノードサイズ定数
const NODE_WIDTH = 200;
const NODE_HEIGHT = 80;
const CORNER_RADIUS = 6;
const PROGRESS_BAR_HEIGHT = 6;
const PADDING = 10;
const CONTENT_LEFT = 20;  // 左ステータスバー分のオフセット

// テキスト解像度（Retina対応）
const TEXT_RESOLUTION = typeof window !== 'undefined'
	? Math.min(window.devicePixelRatio * 2, 4)  // 最大4xに制限
	: 2;

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
	borderCritical: 0xff9533,
	text: 0xffffff,
	textSecondary: 0xb8b8b8,
	textMuted: 0x888888,
	progressBg: 0x1a1a1a,
	// クリティカルパス用
	criticalGlow: 0xff9533,
	slackBadge: 0x2d5a2d,
	// 影響範囲ハイライト用
	downstreamHighlight: 0xffcc00,  // 下流タスク（黄色）
	upstreamHighlight: 0x44aaff     // 上流タスク（水色）
};

// ハイライトタイプ
export type HighlightType = 'downstream' | 'upstream' | null;

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
	private slackBadge: Graphics;
	private slackText: Text;

	private isHovered = false;
	private isSelected = false;
	private currentLOD: LODLevel = LODLevel.Micro;

	// イベントコールバック
	private onClickCallback?: (node: TaskNode, event?: FederatedPointerEvent) => void;
	private onHoverCallback?: (node: TaskNode, isHovered: boolean) => void;

	// 進捗率（0-100）- タスク自体には進捗がないので、ステータスから推定
	private progress: number;

	// クリティカルパス・スラック情報
	private isOnCriticalPath = false;
	private slack: number | null = null;

	// 影響範囲ハイライト
	private highlightType: HighlightType = null;

	constructor(task: TaskItem) {
		super();

		this.task = task;
		this.progress = this.estimateProgress(task.status);

		// コンポーネント初期化
		this.background = new Graphics();
		this.statusIndicator = new Graphics();
		this.idText = new Text({
			text: '',
			style: { fontSize: 12, fill: COLORS.text, fontFamily: 'IBM Plex Mono, monospace' },
			resolution: TEXT_RESOLUTION
		});
		this.titleText = new Text({
			text: '',
			style: { fontSize: 11, fill: COLORS.textSecondary, fontFamily: 'IBM Plex Mono, monospace' },
			resolution: TEXT_RESOLUTION
		});
		this.progressBar = new Graphics();
		this.metaText = new Text({
			text: '',
			style: { fontSize: 10, fill: COLORS.textMuted, fontFamily: 'IBM Plex Mono, monospace' },
			resolution: TEXT_RESOLUTION
		});
		this.slackBadge = new Graphics();
		this.slackText = new Text({
			text: '',
			style: { fontSize: 9, fill: COLORS.text, fontFamily: 'IBM Plex Mono, monospace' },
			resolution: TEXT_RESOLUTION
		});

		this.addChild(this.background);
		this.addChild(this.statusIndicator);
		this.addChild(this.idText);
		this.addChild(this.titleText);
		this.addChild(this.progressBar);
		this.addChild(this.metaText);
		this.addChild(this.slackBadge);
		this.addChild(this.slackText);

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
		this.drawSlackBadge();
	}

	/**
	 * 背景を描画
	 */
	private drawBackground(): void {
		this.background.clear();

		let bgColor = COLORS.background;
		let borderColor = COLORS.border;
		let borderWidth = 2;

		if (this.isSelected) {
			bgColor = COLORS.backgroundSelected;
			borderColor = COLORS.borderHighlight;
		} else if (this.isHovered) {
			bgColor = COLORS.backgroundHover;
			borderColor = COLORS.borderHighlight;
		} else if (this.highlightType === 'downstream') {
			// 下流タスク（黄色ハイライト）
			borderColor = COLORS.downstreamHighlight;
			borderWidth = 3;
		} else if (this.highlightType === 'upstream') {
			// 上流タスク（水色ハイライト）
			borderColor = COLORS.upstreamHighlight;
			borderWidth = 3;
		} else if (this.isOnCriticalPath) {
			// クリティカルパス上のノードはオレンジボーダー
			borderColor = COLORS.borderCritical;
			borderWidth = 3;
		}

		// 背景
		this.background.roundRect(0, 0, NODE_WIDTH, NODE_HEIGHT, CORNER_RADIUS);
		this.background.fill(bgColor);
		this.background.stroke({ width: borderWidth, color: borderColor });

		// 影響範囲ハイライトのグロー効果
		if (this.highlightType && !this.isSelected && !this.isHovered) {
			const glowColor = this.highlightType === 'downstream'
				? COLORS.downstreamHighlight
				: COLORS.upstreamHighlight;
			this.background.roundRect(-2, -2, NODE_WIDTH + 4, NODE_HEIGHT + 4, CORNER_RADIUS + 2);
			this.background.stroke({ width: 1, color: glowColor, alpha: 0.4 });
		}
		// クリティカルパスの場合はグロー効果
		else if (this.isOnCriticalPath && !this.isSelected && !this.isHovered && !this.highlightType) {
			this.background.roundRect(-2, -2, NODE_WIDTH + 4, NODE_HEIGHT + 4, CORNER_RADIUS + 2);
			this.background.stroke({ width: 1, color: COLORS.criticalGlow, alpha: 0.3 });
		}

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

		// 左側のステータスバー（角丸に合わせて調整）
		this.statusIndicator.roundRect(0, 0, 6, NODE_HEIGHT, { topLeft: CORNER_RADIUS, bottomLeft: CORNER_RADIUS, topRight: 0, bottomRight: 0 });
		this.statusIndicator.fill(statusColor);
	}

	/**
	 * テキストを描画
	 */
	private drawTexts(): void {
		const contentWidth = NODE_WIDTH - CONTENT_LEFT - PADDING;

		if (this.currentLOD === LODLevel.Macro) {
			// マクロレベルでは非表示
			this.idText.visible = false;
			this.titleText.visible = false;
			this.metaText.visible = false;
			return;
		}

		this.idText.visible = true;

		// ID テキスト（上部）
		const maxIdChars = Math.floor(contentWidth / 7);  // 等幅フォントで概算
		const shortId = this.task.id.length > maxIdChars
			? this.task.id.substring(0, maxIdChars - 2) + '..'
			: this.task.id;
		this.idText.text = shortId;
		this.idText.x = CONTENT_LEFT;
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

		// タイトル（中央）
		const maxTitleChars = Math.floor(contentWidth / 6.5);
		const title = this.task.title.length > maxTitleChars
			? this.task.title.substring(0, maxTitleChars - 2) + '..'
			: this.task.title;
		this.titleText.text = title;
		this.titleText.x = CONTENT_LEFT;
		this.titleText.y = PADDING + 16;

		// メタ情報（担当者 - 下部）
		const assignee = this.task.assignee || 'unassigned';
		const maxAssigneeChars = Math.floor(contentWidth / 7);
		const displayAssignee = assignee.length > maxAssigneeChars
			? assignee.substring(0, maxAssigneeChars - 2) + '..'
			: assignee;
		this.metaText.text = `@${displayAssignee}`;
		this.metaText.x = CONTENT_LEFT;
		this.metaText.y = NODE_HEIGHT - PADDING - 12;
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

		const barWidth = NODE_WIDTH - CONTENT_LEFT - PADDING;
		const barY = PADDING + 34;  // タイトルとメタの間

		// 背景
		this.progressBar.roundRect(CONTENT_LEFT, barY, barWidth, PROGRESS_BAR_HEIGHT, 3);
		this.progressBar.fill(COLORS.progressBg);

		// 進捗
		if (this.progress > 0) {
			const progressWidth = (barWidth * this.progress) / 100;
			const progressColor = COLORS.status[this.task.status] || COLORS.status.pending;

			this.progressBar.roundRect(CONTENT_LEFT, barY, progressWidth, PROGRESS_BAR_HEIGHT, 3);
			this.progressBar.fill(progressColor);
		}
	}

	/**
	 * スラックバッジを描画
	 */
	private drawSlackBadge(): void {
		this.slackBadge.clear();
		this.slackText.visible = false;

		// スラック表示条件: 値が設定されていて、Microレベル
		if (this.slack === null || this.currentLOD !== LODLevel.Micro) {
			this.slackBadge.visible = false;
			return;
		}

		this.slackBadge.visible = true;
		this.slackText.visible = true;

		// バッジの位置（右上角）
		const badgeX = NODE_WIDTH - 8;
		const badgeY = -4;
		const badgeWidth = 40;
		const badgeHeight = 16;

		// バッジの色（スラック0はオレンジ、それ以外は緑系）
		const badgeColor = this.slack === 0 ? COLORS.criticalGlow : COLORS.slackBadge;

		// バッジ背景
		this.slackBadge.roundRect(badgeX - badgeWidth + 8, badgeY, badgeWidth, badgeHeight, 4);
		this.slackBadge.fill(badgeColor);
		this.slackBadge.stroke({ width: 1, color: 0x1a1a1a });

		// スラック日数テキスト
		const slackStr = this.slack === 0 ? 'CRIT' : `+${this.slack}d`;
		this.slackText.text = slackStr;
		this.slackText.x = badgeX - badgeWidth + 12;
		this.slackText.y = badgeY + 3;
	}

	/**
	 * LODレベルを設定（軽量化: visibility のみ切り替え）
	 */
	setLOD(level: LODLevel): void {
		if (this.currentLOD === level) return;
		this.currentLOD = level;
		this.updateLODVisibility();
	}

	/**
	 * LODに応じた要素の表示/非表示を更新（draw() より軽量）
	 */
	private updateLODVisibility(): void {
		if (this.currentLOD === LODLevel.Macro) {
			// マクロレベル: テキスト類を全て非表示
			this.idText.visible = false;
			this.titleText.visible = false;
			this.metaText.visible = false;
			this.progressBar.visible = false;
			this.slackBadge.visible = false;
			this.slackText.visible = false;
		} else if (this.currentLOD === LODLevel.Meso) {
			// メソレベル: IDのみ表示
			this.idText.visible = true;
			this.titleText.visible = false;
			this.metaText.visible = false;
			this.progressBar.visible = false;
			this.slackBadge.visible = false;
			this.slackText.visible = false;
		} else {
			// マイクロレベル: 全情報表示
			this.idText.visible = true;
			this.titleText.visible = true;
			this.metaText.visible = true;
			this.progressBar.visible = true;
			// スラックバッジは値がある場合のみ
			if (this.slack !== null) {
				this.slackBadge.visible = true;
				this.slackText.visible = true;
			}
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
	 * クリティカルパス状態を設定
	 */
	setCriticalPath(isOnCriticalPath: boolean): void {
		if (this.isOnCriticalPath !== isOnCriticalPath) {
			this.isOnCriticalPath = isOnCriticalPath;
			this.draw();
		}
	}

	/**
	 * スラック（余裕日数）を設定
	 * @param slack - スラック日数（null, または 0 以上の有限数値）
	 */
	setSlack(slack: number | null): void {
		// null または undefined は常に許可
		if (slack === null || slack === undefined) {
			if (this.slack !== null) {
				this.slack = null;
				this.draw();
			}
			return;
		}

		// 無効な値（負数、Infinity, NaN）は無視してログ出力
		if (!Number.isFinite(slack) || slack < 0) {
			console.warn(`Invalid slack value for task ${this.task.id}: ${slack}`);
			return;
		}

		if (this.slack !== slack) {
			this.slack = slack;
			this.draw();
		}
	}

	/**
	 * 影響範囲ハイライトを設定
	 * @param highlighted - ハイライト状態
	 * @param type - ハイライトタイプ（'downstream' | 'upstream'）
	 */
	setHighlighted(highlighted: boolean, type?: 'downstream' | 'upstream'): void {
		const newType: HighlightType = highlighted ? (type || 'downstream') : null;
		if (this.highlightType !== newType) {
			this.highlightType = newType;
			this.drawBackground();
		}
	}

	/**
	 * ハイライトタイプを取得
	 */
	getHighlightType(): HighlightType {
		return this.highlightType;
	}

	/**
	 * クリティカルパス上にあるかを取得
	 */
	isTaskOnCriticalPath(): boolean {
		return this.isOnCriticalPath;
	}

	/**
	 * スラック値を取得
	 */
	getSlack(): number | null {
		return this.slack;
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
