// グラフノードの描画クラス（WBS 全ノードタイプ対応）
import { Container, Graphics, Text } from 'pixi.js';
import type { FederatedPointerEvent } from 'pixi.js';
import type { GraphNode, GraphNodeType } from '$lib/types/api';
import {
	NODE_TYPE_CONFIG,
	DEFAULT_NODE_TYPE,
	getNodeTypeColors,
	getNodeTypeLabel,
	shouldShowBadge,
	shouldShowAccentLine
} from '../config/nodeTypes';

// ノードサイズ定数
const NODE_WIDTH = 200;
const NODE_HEIGHT = 80;
const CORNER_RADIUS = 6;
const PROGRESS_BAR_HEIGHT = 6;
const PADDING = 10;
const CONTENT_LEFT = 20; // 左ステータスバー分のオフセット

// テキスト解像度（Retina対応）
const TEXT_RESOLUTION =
	typeof window !== 'undefined'
		? Math.min(window.devicePixelRatio * 2, 4) // 最大4xに制限
		: 2;

// 色定義（Factorioテーマに準拠）
const COLORS = {
	// ステータス色
	status: {
		completed: 0x44cc44,
		in_progress: 0x4488ff,
		pending: 0x888888,
		blocked: 0xee4444
	} as Record<string, number>,
	// 優先度色
	priority: {
		high: 0xee4444,
		medium: 0xffcc00,
		low: 0x44cc44
	},
	// ノードタイプ別の色（NODE_TYPE_CONFIG から生成）
	nodeType: Object.fromEntries(
		Object.entries(NODE_TYPE_CONFIG).map(([key, config]) => [key, config.colors])
	) as Record<
		GraphNodeType,
		{
			indicator: number;
			background: number;
			border: number;
			borderHighlight?: number;
			borderShadow?: number;
		}
	>,
	// 基本色
	background: 0x2d2d2d,
	backgroundHover: 0x3a3a3a,
	backgroundSelected: 0x4a4a4a,
	border: 0x4a4a4a,
	borderHighlight: 0xff9533,
	text: 0xffffff,
	textSecondary: 0xb8b8b8,
	textMuted: 0x888888,
	progressBg: 0x1a1a1a,
	progressFrame: 0x555555,
	progressSegment: 0x333333,
	// 進捗グラデーション（0% → 100%）
	progressGradient: {
		low: 0xff6644, // 0-33%: オレンジ/赤
		mid: 0xffcc00, // 34-66%: 黄色
		high: 0x44dd44 // 67-100%: 緑
	},
	// 影響範囲ハイライト用
	downstreamHighlight: 0xffcc00, // 下流タスク（黄色）
	upstreamHighlight: 0x44aaff // 上流タスク（水色）
};

// 金属効果定数（5層ベベル構造用）
// alpha 累積問題を解消: 合計 0.70 に調整（以前は 1.13 で過度に暗かった）
const METAL_EFFECT = {
	// ベベル透明度（控えめに調整）
	outerShadowAlpha: 0.25, // 0.4 → 0.25（影を控えめに）
	innerShadowAlpha: 0.15, // 0.3 → 0.15（内側影も控えめに）
	innerHighlightAlpha: 0.5, // 維持
	outerHighlightAlpha: 0.4, // 維持
	// ベベル幅
	bevelWidth: 1.5,
	innerBevelWidth: 1,
	// 上部ハイライト（金属光沢）- 領域を拡大
	topHighlightAlpha: 0.2, // 0.25 → 0.20（光沢を適度に）
	topHighlightRatio: 0.45, // 0.35 → 0.45（領域拡大）
	// 下部シャドウ（凹み感）- 開始位置を上げて重なりを作る
	bottomShadowAlpha: 0.1, // 0.15 → 0.10（下部影は最小限）
	bottomShadowRatio: 0.4, // 0.3 → 0.40（60% 位置から開始）
	// グロー設定（選択・ハイライト・クリティカルパス時に適用）
	// 依存グラフ視覚効果最適化 Phase 2: ノード数が多い場合の累積効果を抑制
	baseGlowAlpha: 0.06, // 0.12 → 0.06
	hoverGlowAlpha: 0.12, // 0.25 → 0.12
	selectedGlowAlpha: 0.2 // 0.4 → 0.20
} as const;

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
 * GraphNodeView - グラフノード（Vision, Objective, Deliverable, Activity, UseCase）の視覚的表現
 *
 * 責務:
 * - ノードのグラフィカル表示
 * - ノードタイプに応じたスタイル変更
 * - インタラクション（クリック、ホバー）
 * - LOD（詳細度）に応じた表示切り替え
 */
export class GraphNodeView extends Container {
	private graphNode: GraphNode;
	private nodeType: GraphNodeType;
	private background: Graphics;
	private statusIndicator: Graphics;
	private typeIndicator: Graphics; // ノードタイプバッジ
	private typeText: Text; // タイプバッジのラベル文字
	private idText: Text;
	private titleText: Text;
	private progressBar: Graphics;
	private metaText: Text;

	private isHovered = false;
	private isSelected = false;
	private currentLOD: LODLevel = LODLevel.Micro;

	// イベントコールバック
	private onClickCallback?: (node: GraphNodeView, event?: FederatedPointerEvent) => void;
	private onHoverCallback?: (node: GraphNodeView, isHovered: boolean) => void;
	private onContextMenuCallback?: (node: GraphNodeView, event: FederatedPointerEvent) => void;

	// 進捗率（0-100）- ステータスから推定
	private progress: number;

	// 影響範囲ハイライト
	private highlightType: HighlightType = null;

	constructor(data: GraphNode) {
		super();

		this.graphNode = data;
		this.nodeType = this.graphNode.node_type;
		// 進捗率はステータスから推定（progress フィールドは削除済み）
		this.progress = this.estimateProgressFromStatus(this.graphNode.status);

		// コンポーネント初期化
		this.background = new Graphics();
		this.statusIndicator = new Graphics();
		this.typeIndicator = new Graphics();
		this.typeText = new Text({
			text: '',
			style: {
				fontSize: 11,
				fill: 0x1a1a1a,
				fontFamily: 'IBM Plex Mono, monospace',
				fontWeight: 'bold'
			},
			resolution: TEXT_RESOLUTION
		});
		this.typeText.anchor.set(0.5, 0.5);
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

		this.addChild(this.background);
		this.addChild(this.statusIndicator);
		this.addChild(this.typeIndicator);
		this.addChild(this.typeText);
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
		this.on('pointerdown', (e: FederatedPointerEvent) => {
			console.log('[GraphNodeView] pointerdown event, button:', e.button);
			// 右クリック（button === 2）を検出
			if (e.button === 2) {
				console.log('[GraphNodeView] Right-click detected!');
				this.handleContextMenu(e);
			}
		});

		// 初回描画
		this.draw();
	}

	/**
	 * ステータスから進捗率を推定（汎用ステータス対応）
	 */
	private estimateProgressFromStatus(status: string): number {
		// 完了系
		if (
			[
				'completed',
				'approved',
				'delivered',
				'mitigated',
				'verified',
				'resolved',
				'passing'
			].includes(status)
		) {
			return 100;
		}
		// 進行中系
		if (['in_progress', 'in_review', 'investigating', 'mitigating', 'decided'].includes(status)) {
			return 50;
		}
		// 未着手系
		if (
			[
				'pending',
				'not_started',
				'draft',
				'open',
				'identified',
				'unverified',
				'not_checked'
			].includes(status)
		) {
			return 0;
		}
		// ブロック系
		if (
			['blocked', 'on_hold', 'deferred', 'wont_fix', 'invalid', 'accepted', 'failing'].includes(
				status
			)
		) {
			return 0;
		}
		return 0;
	}

	/**
	 * ノードを描画
	 */
	draw(): void {
		this.drawBackground();
		this.drawStatusIndicator();
		this.drawTypeIndicator();
		this.drawTexts();
		this.drawProgressBar();
	}

	/**
	 * 背景を描画（ノードタイプ別の色対応）
	 * 5層ベベル構造 + 3層グローで金属質感を表現
	 */
	private drawBackground(): void {
		this.background.clear();

		// ノードタイプ別の基本色を取得
		const typeColors = COLORS.nodeType[this.nodeType] || COLORS.nodeType[DEFAULT_NODE_TYPE];
		let bgColor = typeColors.background;
		let borderColor = typeColors.border;
		let borderWidth = 2;
		const highlightColor = typeColors.borderHighlight || 0x777777;
		const shadowColor = typeColors.borderShadow || 0x1a1a1a;

		if (this.isSelected) {
			bgColor = COLORS.backgroundSelected;
			borderColor = COLORS.borderHighlight;
		} else if (this.isHovered) {
			bgColor = COLORS.backgroundHover;
			borderColor = COLORS.borderHighlight;
		} else if (this.highlightType === 'downstream') {
			borderColor = COLORS.downstreamHighlight;
			borderWidth = 3;
		} else if (this.highlightType === 'upstream') {
			borderColor = COLORS.upstreamHighlight;
			borderWidth = 3;
		}

		// === 3層グロー ===

		// 最外層グロー（選択・ハイライト時のみ）
		// Note: ホバー時は中間・内側グローのみ表示し、最外層グローは表示しない（視覚ノイズ軽減のため）
		if (this.isSelected || this.highlightType) {
			const baseGlowColor = typeColors.indicator;
			this.background.roundRect(-8, -8, NODE_WIDTH + 16, NODE_HEIGHT + 16, CORNER_RADIUS + 8);
			this.background.fill({ color: baseGlowColor, alpha: METAL_EFFECT.baseGlowAlpha });
		}

		// 中間・内側グロー（ホバー/選択/ハイライト時に強化）
		if (this.isSelected || this.isHovered || this.highlightType) {
			let glowColor = borderColor;
			let glowAlpha: number = METAL_EFFECT.hoverGlowAlpha;

			if (this.isSelected) {
				glowAlpha = METAL_EFFECT.selectedGlowAlpha;
			} else if (this.highlightType) {
				glowColor =
					this.highlightType === 'downstream'
						? COLORS.downstreamHighlight
						: COLORS.upstreamHighlight;
			}

			// 中間グロー層
			this.background.roundRect(-6, -6, NODE_WIDTH + 12, NODE_HEIGHT + 12, CORNER_RADIUS + 6);
			this.background.fill({ color: glowColor, alpha: glowAlpha * 0.6 });

			// 内側グロー層
			this.background.roundRect(-3, -3, NODE_WIDTH + 6, NODE_HEIGHT + 6, CORNER_RADIUS + 3);
			this.background.fill({ color: glowColor, alpha: glowAlpha });
		}

		// === 5層ベベル構造 ===

		// Layer 1: 外側シャドウ（下・右）- 板金の影
		this.background.roundRect(
			METAL_EFFECT.bevelWidth,
			METAL_EFFECT.bevelWidth,
			NODE_WIDTH,
			NODE_HEIGHT,
			CORNER_RADIUS
		);
		this.background.fill({ color: shadowColor, alpha: METAL_EFFECT.outerShadowAlpha });

		// Layer 2: 内側シャドウ（下・右）
		this.background.roundRect(
			METAL_EFFECT.innerBevelWidth,
			METAL_EFFECT.innerBevelWidth,
			NODE_WIDTH - METAL_EFFECT.innerBevelWidth,
			NODE_HEIGHT - METAL_EFFECT.innerBevelWidth,
			CORNER_RADIUS - 1
		);
		this.background.fill({ color: 0x000000, alpha: METAL_EFFECT.innerShadowAlpha });

		// Layer 3: メイン背景
		this.background.roundRect(0, 0, NODE_WIDTH, NODE_HEIGHT, CORNER_RADIUS);
		this.background.fill(bgColor);
		this.background.stroke({ width: borderWidth, color: borderColor });

		// Layer 4: 内側ハイライト（上部金属光沢）
		this.background.roundRect(
			3,
			3,
			NODE_WIDTH - 6,
			NODE_HEIGHT * METAL_EFFECT.topHighlightRatio,
			CORNER_RADIUS - 2
		);
		this.background.fill({ color: highlightColor, alpha: METAL_EFFECT.topHighlightAlpha });

		// Layer 5: 外側ハイライト（上部ボーダー）
		this.background.moveTo(CORNER_RADIUS, 1);
		this.background.lineTo(NODE_WIDTH - CORNER_RADIUS, 1);
		this.background.stroke({
			width: METAL_EFFECT.bevelWidth,
			color: highlightColor,
			alpha: METAL_EFFECT.outerHighlightAlpha
		});

		// 下部シャドウ（凹み感）
		this.background.roundRect(
			3,
			NODE_HEIGHT * (1 - METAL_EFFECT.bottomShadowRatio),
			NODE_WIDTH - 6,
			NODE_HEIGHT * METAL_EFFECT.bottomShadowRatio - 3,
			CORNER_RADIUS - 2
		);
		this.background.fill({ color: 0x000000, alpha: METAL_EFFECT.bottomShadowAlpha });

		// === アクセントライン（ノードタイプ設定に基づく）===
		if (shouldShowAccentLine(this.nodeType)) {
			this.background.moveTo(CORNER_RADIUS + 4, 2);
			this.background.lineTo(NODE_WIDTH - CORNER_RADIUS - 4, 2);
			this.background.stroke({ width: 1, color: typeColors.indicator, alpha: 0.15 });
		}
	}

	/**
	 * ステータスインジケーターを描画
	 */
	private drawStatusIndicator(): void {
		this.statusIndicator.clear();

		const statusColor = COLORS.status[this.graphNode.status] || COLORS.status.pending;

		// 左側のステータスバー（角丸に合わせて調整）
		// PixiJS v8 では roundRect に個別角丸指定はできないため、パスで描画
		const g = this.statusIndicator;
		const w = 6;
		const h = NODE_HEIGHT;
		const r = CORNER_RADIUS;
		g.moveTo(r, 0);
		g.lineTo(w, 0);
		g.lineTo(w, h);
		g.lineTo(r, h);
		g.arcTo(0, h, 0, h - r, r);
		g.lineTo(0, r);
		g.arcTo(0, 0, r, 0, r);
		g.closePath();
		g.fill(statusColor);
	}

	/**
	 * ノードタイプインジケーターを描画（右上バッジ）
	 */
	private drawTypeIndicator(): void {
		this.typeIndicator.clear();

		// ノードタイプ設定に基づいてバッジ表示を判定
		if (!shouldShowBadge(this.nodeType)) {
			this.typeIndicator.visible = false;
			this.typeText.visible = false;
			return;
		}

		this.typeIndicator.visible = true;
		this.typeText.visible = true;

		const typeColors = getNodeTypeColors(this.nodeType);

		const badgeSize = 20;
		const badgeX = NODE_WIDTH - badgeSize - 4;
		const badgeY = 4;

		// バッジ背景（円形）
		this.typeIndicator.circle(badgeX + badgeSize / 2, badgeY + badgeSize / 2, badgeSize / 2);
		this.typeIndicator.fill(typeColors.indicator);
		this.typeIndicator.stroke({ width: 1, color: 0x1a1a1a });

		// ラベル文字（V/O/D/A/U）を円の中央に配置（anchor ベース）
		const label = getNodeTypeLabel(this.nodeType);
		this.typeText.text = label;
		this.typeText.x = badgeX + badgeSize / 2;
		this.typeText.y = badgeY + badgeSize / 2;
	}

	/**
	 * テキストを描画
	 * テキスト内容は LOD に関係なく常に更新し、可視性は updateLODVisibility() で制御
	 */
	private drawTexts(): void {
		const contentWidth = NODE_WIDTH - CONTENT_LEFT - PADDING;

		// ID テキスト（上部）- 常に内容を更新
		const displayId = this.graphNode.id;
		const maxIdChars = Math.floor(contentWidth / 7); // 等幅フォントで概算
		const shortId =
			displayId.length > maxIdChars ? displayId.substring(0, maxIdChars - 2) + '..' : displayId;
		this.idText.text = shortId;
		this.idText.x = CONTENT_LEFT;
		this.idText.y = PADDING;

		// タイトル（中央）- 常に内容を更新
		const maxTitleChars = Math.floor(contentWidth / 6.5);
		const title =
			this.graphNode.title.length > maxTitleChars
				? this.graphNode.title.substring(0, maxTitleChars - 2) + '..'
				: this.graphNode.title;
		this.titleText.text = title;
		this.titleText.x = CONTENT_LEFT;
		this.titleText.y = PADDING + 16;

		// メタ情報（担当者または進捗 - 下部）- 常に内容を更新
		const assignee = this.graphNode.assignee || '';
		const progressPct = `${Math.round(this.progress)}%`;
		const metaInfo = assignee ? `@${assignee}` : progressPct;
		const maxMetaChars = Math.floor(contentWidth / 7);
		const displayMeta =
			metaInfo.length > maxMetaChars ? metaInfo.substring(0, maxMetaChars - 2) + '..' : metaInfo;
		this.metaText.text = displayMeta;
		this.metaText.x = CONTENT_LEFT;
		this.metaText.y = NODE_HEIGHT - PADDING - 12;

		// LOD に応じた可視性制御
		this.updateLODVisibility();
	}

	/**
	 * 進捗率に応じた色を計算（グラデーション）
	 * 0-33%: オレンジ/赤 → 34-66%: 黄色 → 67-100%: 緑
	 */
	private getProgressColor(progress: number): number {
		const { low, mid, high } = COLORS.progressGradient;

		if (progress <= 33) {
			// 0-33%: low から mid へ補間
			const t = progress / 33;
			return this.lerpColor(low, mid, t);
		} else if (progress <= 66) {
			// 34-66%: mid のまま（黄色ゾーン）
			const t = (progress - 33) / 33;
			return this.lerpColor(mid, mid, t);
		} else {
			// 67-100%: mid から high へ補間
			const t = (progress - 66) / 34;
			return this.lerpColor(mid, high, t);
		}
	}

	/**
	 * 2色間の線形補間
	 */
	private lerpColor(color1: number, color2: number, t: number): number {
		const r1 = (color1 >> 16) & 0xff;
		const g1 = (color1 >> 8) & 0xff;
		const b1 = color1 & 0xff;

		const r2 = (color2 >> 16) & 0xff;
		const g2 = (color2 >> 8) & 0xff;
		const b2 = color2 & 0xff;

		const r = Math.round(r1 + (r2 - r1) * t);
		const g = Math.round(g1 + (g2 - g1) * t);
		const b = Math.round(b1 + (b2 - b1) * t);

		return (r << 16) | (g << 8) | b;
	}

	/**
	 * プログレスバーを描画（Factorio 風インダストリアルデザイン）
	 */
	private drawProgressBar(): void {
		this.progressBar.clear();

		if (this.currentLOD !== LODLevel.Micro) {
			this.progressBar.visible = false;
			return;
		}

		this.progressBar.visible = true;

		const barWidth = NODE_WIDTH - CONTENT_LEFT - PADDING;
		const barY = PADDING + 34;
		const segmentCount = 10;
		const segmentWidth = barWidth / segmentCount;
		const segmentGap = 1; // セグメント間の隙間

		// 外枠フレーム（金属感・立体感）
		// 外側の暗い枠
		this.progressBar.roundRect(
			CONTENT_LEFT - 2,
			barY - 2,
			barWidth + 4,
			PROGRESS_BAR_HEIGHT + 4,
			3
		);
		this.progressBar.fill(0x222222);
		// 内側の明るい枠（溝の表現）
		this.progressBar.roundRect(
			CONTENT_LEFT - 1,
			barY - 1,
			barWidth + 2,
			PROGRESS_BAR_HEIGHT + 2,
			2
		);
		this.progressBar.stroke({ width: 1, color: COLORS.progressFrame });

		// 背景（暗いベース）
		this.progressBar.rect(CONTENT_LEFT, barY, barWidth, PROGRESS_BAR_HEIGHT);
		this.progressBar.fill(COLORS.progressBg);

		// セグメント単位で描画（デジタルゲージ風）
		const filledSegments = Math.ceil((this.progress / 100) * segmentCount);

		for (let i = 0; i < segmentCount; i++) {
			const segX = CONTENT_LEFT + i * segmentWidth + segmentGap / 2;
			const segW = segmentWidth - segmentGap;

			if (i < filledSegments && this.progress > 0) {
				// 塗りつぶしセグメント
				// セグメント位置に応じた進捗色を計算
				const segmentProgress = ((i + 1) / segmentCount) * 100;
				const progressColor = this.getProgressColor(Math.min(segmentProgress, this.progress));

				// メインセグメント
				this.progressBar.rect(segX, barY, segW, PROGRESS_BAR_HEIGHT);
				this.progressBar.fill(progressColor);

				// 上部ハイライト（光沢）
				this.progressBar.rect(segX, barY, segW, 2);
				this.progressBar.fill({ color: 0xffffff, alpha: 0.3 });

				// 下部シャドウ（立体感）
				this.progressBar.rect(segX, barY + PROGRESS_BAR_HEIGHT - 1, segW, 1);
				this.progressBar.fill({ color: 0x000000, alpha: 0.3 });
			} else {
				// 未塗りつぶしセグメント（暗いマーカー）
				this.progressBar.rect(segX, barY, segW, PROGRESS_BAR_HEIGHT);
				this.progressBar.fill(0x252525);
			}
		}

		// 100% 完了時のグロー効果
		if (this.progress >= 100) {
			this.progressBar.roundRect(
				CONTENT_LEFT - 1,
				barY - 1,
				barWidth + 2,
				PROGRESS_BAR_HEIGHT + 2,
				2
			);
			this.progressBar.stroke({ width: 1, color: COLORS.progressGradient.high, alpha: 0.5 });
		}
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
		} else if (this.currentLOD === LODLevel.Meso) {
			// メソレベル: IDのみ表示
			this.idText.visible = true;
			this.titleText.visible = false;
			this.metaText.visible = false;
			this.progressBar.visible = false;
		} else {
			// マイクロレベル: 全情報表示
			this.idText.visible = true;
			this.titleText.visible = true;
			this.metaText.visible = true;
			this.progressBar.visible = true;
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
	 * 右クリック処理
	 */
	private handleContextMenu(e: FederatedPointerEvent): void {
		e.stopPropagation();
		this.onContextMenuCallback?.(this, e);
	}

	/**
	 * 選択状態を設定
	 */
	setSelected(selected: boolean): void {
		this.isSelected = selected;
		this.drawBackground();
	}

	/**
	 * ノードデータを更新
	 */
	updateData(data: GraphNode): void {
		this.graphNode = data;
		this.nodeType = this.graphNode.node_type;
		// 進捗率はステータスから推定（progress フィールドは削除済み）
		this.progress = this.estimateProgressFromStatus(this.graphNode.status);
		this.draw();
	}

	/**
	 * 影響範囲ハイライトを設定
	 * @param highlighted - ハイライト状態
	 * @param type - ハイライトタイプ（'downstream' | 'upstream'）
	 */
	setHighlighted(highlighted: boolean, type?: 'downstream' | 'upstream'): void {
		const newType: HighlightType = highlighted ? type || 'downstream' : null;
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
	 * ノードIDを取得
	 */
	getNodeId(): string {
		return this.graphNode.id;
	}

	/**
	 * GraphNode データを取得
	 */
	getGraphNode(): GraphNode {
		return this.graphNode;
	}

	/**
	 * ノードタイプを取得
	 */
	getNodeType(): GraphNodeType {
		return this.nodeType;
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
	onClick(callback: (node: GraphNodeView, event?: FederatedPointerEvent) => void): void {
		this.onClickCallback = callback;
	}

	onHover(callback: (node: GraphNodeView, isHovered: boolean) => void): void {
		this.onHoverCallback = callback;
	}

	onContextMenu(callback: (node: GraphNodeView, event: FederatedPointerEvent) => void): void {
		this.onContextMenuCallback = callback;
	}
}

// 後方互換性のためのエイリアス
export { GraphNodeView as TaskNode };
