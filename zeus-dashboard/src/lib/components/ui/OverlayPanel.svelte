<script lang="ts">
	// 共通オーバーレイパネルコンポーネント
	// ビューワー上に浮かぶフローティングパネルを提供
	// 幅切り替え機能付き（3パターン、localStorage 永続化）
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { Icon } from '$lib/components/ui';

	// 幅プリセット定義
	type WidthPreset = 'narrow' | 'medium' | 'wide';
	const WIDTH_PRESETS: Record<WidthPreset, string> = {
		narrow: '280px',
		medium: '360px',
		wide: '460px'
	};

	interface Props {
		title: string;
		position?: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right';
		width?: string;
		maxHeight?: string;
		showCloseButton?: boolean;
		onClose?: () => void;
		children: Snippet;
		// 幅切り替え機能
		panelId?: string; // localStorage キー用（設定すると幅切り替えアイコンが表示）
		defaultWidthPreset?: WidthPreset;
	}

	let {
		title,
		position = 'top-left',
		width = '280px',
		maxHeight = 'calc(100% - 24px)',
		showCloseButton = true,
		onClose,
		children,
		panelId,
		defaultWidthPreset = 'medium'
	}: Props = $props();

	// localStorage キー
	const STORAGE_KEY_PREFIX = 'zeus-panel-width-';

	// 初期幅プリセットを取得（localStorage > defaultWidthPreset）
	function getInitialWidthPreset(): WidthPreset {
		if (typeof window !== 'undefined' && panelId) {
			const saved = localStorage.getItem(`${STORAGE_KEY_PREFIX}${panelId}`);
			if (saved === 'narrow' || saved === 'medium' || saved === 'wide') {
				return saved;
			}
		}
		return defaultWidthPreset;
	}

	// 幅切り替え状態（panelId が指定されている場合のみ使用）
	let currentWidthPreset = $state<WidthPreset>(getInitialWidthPreset());

	// マウント時にフォーカス
	onMount(() => {
		panelRef?.focus();
	});

	// 幅切り替え
	function setWidthPreset(preset: WidthPreset) {
		currentWidthPreset = preset;
		if (panelId) {
			localStorage.setItem(`${STORAGE_KEY_PREFIX}${panelId}`, preset);
		}
	}

	// パネル参照
	let panelRef: HTMLDivElement | null = $state(null);

	// ARIA 用の一意なID
	const titleId = `overlay-title-${Math.random().toString(36).slice(2, 10)}`;

	// 位置に応じたスタイルクラス
	const positionClass = $derived(`overlay-panel position-${position}`);

	// CSS 値のバリデーション（インジェクション対策）
	function sanitizeCSSLength(value: string, fallback: string): string {
		// 許可パターン: 数値+単位 または calc() 関数
		const pattern = /^(\d+(\.\d+)?(px|rem|em|%|vh|vw)|calc\([^)]+\)|auto)$/;
		return pattern.test(value) ? value : fallback;
	}

	// 実際に使用する幅を計算
	const effectiveWidth = $derived.by(() => {
		// panelId が指定されている場合はプリセットを使用
		if (panelId) {
			return WIDTH_PRESETS[currentWidthPreset];
		}
		// それ以外は width props を使用
		return width;
	});

	// バリデート済みインラインスタイル
	const panelStyle = $derived.by(() => {
		const safeWidth = sanitizeCSSLength(effectiveWidth, '280px');
		const safeMaxHeight = sanitizeCSSLength(maxHeight, 'calc(100% - 24px)');
		return `width: ${safeWidth}; max-height: ${safeMaxHeight};`;
	});

	// ESC キーでパネルを閉じる
	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && onClose) {
			event.preventDefault();
			onClose();
		}
	}
</script>

<div
	bind:this={panelRef}
	class={positionClass}
	style={panelStyle}
	role="dialog"
	aria-labelledby={titleId}
	aria-modal="false"
	tabindex="-1"
	onkeydown={handleKeydown}
>
	<div class="overlay-header">
		<span class="overlay-title" id={titleId}>{title}</span>
		<div class="header-actions">
			<!-- 幅切り替えボタン（panelId 指定時のみ表示） -->
			{#if panelId}
				<div class="width-switcher" role="group" aria-label="パネル幅">
					<button
						class="width-btn"
						class:active={currentWidthPreset === 'narrow'}
						onclick={() => setWidthPreset('narrow')}
						title="狭い (280px)"
						aria-pressed={currentWidthPreset === 'narrow'}
					>
						<span class="width-icon narrow">
							<span class="bar"></span>
						</span>
					</button>
					<button
						class="width-btn"
						class:active={currentWidthPreset === 'medium'}
						onclick={() => setWidthPreset('medium')}
						title="標準 (360px)"
						aria-pressed={currentWidthPreset === 'medium'}
					>
						<span class="width-icon medium">
							<span class="bar"></span>
							<span class="bar"></span>
						</span>
					</button>
					<button
						class="width-btn"
						class:active={currentWidthPreset === 'wide'}
						onclick={() => setWidthPreset('wide')}
						title="広い (460px)"
						aria-pressed={currentWidthPreset === 'wide'}
					>
						<span class="width-icon wide">
							<span class="bar"></span>
							<span class="bar"></span>
							<span class="bar"></span>
						</span>
					</button>
				</div>
			{/if}
			{#if showCloseButton && onClose}
				<button class="close-btn" onclick={onClose} aria-label="閉じる">
					<Icon name="X" size={16} />
				</button>
			{/if}
		</div>
	</div>
	<div class="overlay-content">
		{@render children()}
	</div>
</div>

<style>
	/* オーバーレイパネル共通 */
	.overlay-panel {
		position: absolute;
		background: rgba(26, 26, 26, 0.95);
		border: 1px solid var(--border-metal);
		border-radius: 8px;
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(12px);
		overflow: hidden;
		display: flex;
		flex-direction: column;
		z-index: 10;
	}

	/* フォーカス時のアウトライン */
	.overlay-panel:focus {
		outline: none;
	}

	.overlay-panel:focus-visible {
		outline: 2px solid var(--accent-primary);
		outline-offset: 2px;
	}

	/* 位置バリエーション */
	.position-top-left {
		top: 12px;
		left: 12px;
	}

	.position-top-right {
		top: 12px;
		right: 12px;
	}

	.position-bottom-left {
		bottom: 60px; /* ステータスバーの上 */
		left: 12px;
	}

	.position-bottom-right {
		bottom: 60px; /* ステータスバーの上 */
		right: 12px;
	}

	.overlay-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 8px 12px;
		background: rgba(40, 40, 40, 0.8);
		border-bottom: 1px solid var(--border-metal);
	}

	.overlay-title {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--text-primary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.header-actions {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	/* 幅切り替えボタン */
	.width-switcher {
		display: flex;
		align-items: center;
		gap: 2px;
		padding: 2px;
		background: rgba(0, 0, 0, 0.3);
		border-radius: 4px;
	}

	.width-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 22px;
		height: 18px;
		padding: 0;
		background: transparent;
		border: none;
		border-radius: 3px;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.width-btn:hover {
		background: rgba(255, 255, 255, 0.1);
	}

	.width-btn.active {
		background: rgba(255, 149, 51, 0.3);
	}

	.width-btn:focus-visible {
		outline: 2px solid var(--accent-primary);
		outline-offset: 1px;
	}

	/* 幅アイコン（バーで幅を表現） */
	.width-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 2px;
		height: 12px;
	}

	.width-icon .bar {
		width: 3px;
		height: 10px;
		background: var(--text-muted);
		border-radius: 1px;
		transition: background 0.15s ease;
	}

	.width-btn:hover .bar {
		background: var(--text-secondary);
	}

	.width-btn.active .bar {
		background: var(--accent-primary);
	}

	.close-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 24px;
		height: 24px;
		background: transparent;
		border: none;
		border-radius: 4px;
		color: var(--text-muted);
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.close-btn:hover {
		background: rgba(255, 149, 51, 0.15);
		color: var(--accent-primary);
	}

	.close-btn:focus-visible {
		outline: 2px solid var(--accent-primary);
		outline-offset: 1px;
	}

	.overlay-content {
		flex: 1;
		overflow-y: auto;
	}

	/* レスポンシブ（詳細度を上げて !important を回避） */
	@media (max-width: 768px) {
		.overlay-panel.position-top-left,
		.overlay-panel.position-top-right,
		.overlay-panel.position-bottom-left,
		.overlay-panel.position-bottom-right {
			left: 12px;
			right: 12px;
			width: auto;
			max-height: 50%;
		}

		.overlay-panel.position-top-left,
		.overlay-panel.position-top-right {
			top: 12px;
		}

		.overlay-panel.position-bottom-left,
		.overlay-panel.position-bottom-right {
			top: auto;
			bottom: 12px;
		}
	}

	/* アニメーション対応 */
	@media (prefers-reduced-motion: reduce) {
		.close-btn {
			transition: none;
		}
	}
</style>
