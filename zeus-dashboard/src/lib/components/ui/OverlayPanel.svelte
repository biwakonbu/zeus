<script lang="ts">
	// 共通オーバーレイパネルコンポーネント
	// ビューワー上に浮かぶフローティングパネルを提供
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { Icon } from '$lib/components/ui';

	interface Props {
		title: string;
		position?: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right';
		width?: string;
		maxHeight?: string;
		showCloseButton?: boolean;
		onClose?: () => void;
		children: Snippet;
	}

	let {
		title,
		position = 'top-left',
		width = '280px',
		maxHeight = 'calc(100% - 24px)',
		showCloseButton = true,
		onClose,
		children
	}: Props = $props();

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

	// バリデート済みインラインスタイル
	const panelStyle = $derived.by(() => {
		const safeWidth = sanitizeCSSLength(width, '280px');
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

	// マウント時にフォーカス
	onMount(() => {
		panelRef?.focus();
	});
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
		{#if showCloseButton && onClose}
			<button class="close-btn" onclick={onClose} aria-label="閉じる">
				<Icon name="X" size={16} />
			</button>
		{/if}
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
