<script lang="ts" module>
	/** アイコン名の型（Lucide Icons のコンポーネント名） */
	export type IconName =
		| 'Heart'
		| 'Calendar'
		| 'Flame'
		| 'RefreshCw'
		| 'X'
		| 'AlertTriangle'
		| 'CheckCircle'
		| 'Info'
		| 'XCircle'
		| 'ClipboardList'
		| 'Target'
		| 'BarChart'
		| 'Ruler'
		| 'ZoomIn'
		| 'ZoomOut'
		| 'Keyboard'
		| 'Inbox'
		| 'Settings'
		| 'Zap'
		| 'Circle'
		| 'Package'
		| 'Search'
		| 'Trash2'
		| 'ChevronDown'
		| 'ChevronRight'
		| 'ChevronUp'
		| 'ChevronLeft'
		| 'MoreHorizontal'
		| 'MoreVertical'
		| 'Edit'
		| 'Copy'
		| 'ExternalLink'
		| string;
</script>

<script lang="ts">
	import * as icons from 'lucide-svelte';
	import type { Component } from 'svelte';

	interface Props {
		/** アイコン名（Lucide Icons のコンポーネント名） */
		name: IconName;
		/** アイコンサイズ（px） */
		size?: number;
		/** 線の太さ（Factorio 風: 2.5-3 推奨） */
		strokeWidth?: number;
		/** 追加の CSS クラス */
		class?: string;
		/** グロー効果を適用 */
		glow?: boolean;
		/** aria-label（アクセシビリティ） */
		label?: string;
	}

	let {
		name,
		size = 16,
		strokeWidth = 2.5,
		class: className = '',
		glow = false,
		label
	}: Props = $props();

	// 動的にアイコンコンポーネントを取得
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	const IconComponent = $derived((icons as unknown as Record<string, Component<any>>)[name]);
</script>

{#if IconComponent}
	<span
		class="zeus-icon {className}"
		class:glow
		role={label ? 'img' : 'presentation'}
		aria-label={label}
		aria-hidden={!label}
	>
		<IconComponent {size} {strokeWidth} />
	</span>
{:else}
	<span
		class="zeus-icon-fallback"
		style="width: {size}px; height: {size}px;"
		role="img"
		aria-label={label || 'アイコン'}
	>
		?
	</span>
{/if}

<style>
	.zeus-icon {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		vertical-align: middle;
		flex-shrink: 0;
	}

	/* Factorio 風: 角ばったエッジ */
	.zeus-icon :global(svg) {
		stroke-linecap: square;
		stroke-linejoin: miter;
	}

	/* グロー効果 */
	.zeus-icon.glow {
		filter: drop-shadow(0 0 3px var(--accent-primary, #ff9533));
	}

	/* フォールバック表示 */
	.zeus-icon-fallback {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		background: var(--bg-secondary, #333);
		border-radius: 2px;
		font-size: 10px;
		color: var(--text-muted, #888);
	}
</style>
