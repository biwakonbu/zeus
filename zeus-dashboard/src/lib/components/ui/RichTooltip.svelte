<script lang="ts">
	// Rich Tooltip コンポーネント
	// 議論結果（round: 20260121-174500_wbsdesign）に基づく
	// サイズ: 320x220px、表示遅延: 500ms、位置自動調整

	import { fly } from 'svelte/transition';
	import { tokens, getEntityIcon } from '$lib/theme/design-tokens';
	import Icon from './Icon.svelte';
	import ProgressBar from './ProgressBar.svelte';
	import type { TooltipEntity } from './types';

	interface Props {
		visible: boolean;
		entity: TooltipEntity | null;
		position: { x: number; y: number };
	}

	let { visible, entity, position }: Props = $props();

	// 位置計算（ビューポート端で自動フリップ）
	const adjustedPosition = $derived(calculatePosition(position.x, position.y));

	function calculatePosition(x: number, y: number): { x: number; y: number } {
		const W = tokens.tooltip.width;
		const H = tokens.tooltip.height;
		const OFFSET = tokens.tooltip.offset;

		let newX = x + OFFSET;
		let newY = y + OFFSET;

		if (typeof window !== 'undefined') {
			// 右端でフリップ
			if (newX + W > window.innerWidth) {
				newX = x - W - OFFSET;
			}
			// 左端で補正
			if (newX < 0) {
				newX = OFFSET;
			}
			// 下端でフリップ
			if (newY + H > window.innerHeight) {
				newY = y - H - OFFSET;
			}
			// 上端で補正
			if (newY < 0) {
				newY = OFFSET;
			}
		}

		return { x: newX, y: newY };
	}

	// ステータスに応じた色
	function getStatusColor(status: string): string {
		switch (status) {
			case 'completed':
				return 'var(--status-good)';
			case 'in_progress':
				return 'var(--status-info)';
			case 'blocked':
				return 'var(--status-poor)';
			case 'on_hold':
				return 'var(--status-fair)';
			default:
				return 'var(--text-muted)';
		}
	}

	// ステータス表示名
	function formatStatus(status: string): string {
		return status.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
	}
</script>

{#if visible && entity}
	<div
		class="rich-tooltip"
		style="left: {adjustedPosition.x}px; top: {adjustedPosition.y}px;"
		transition:fly={{ y: 10, duration: 100 }}
		role="tooltip"
		aria-live="polite"
	>
		<!-- ヘッダー：タイプ表示 -->
		<div class="tooltip-header">
			<Icon name={getEntityIcon(entity.type)} size={16} />
			<span class="tooltip-type">{entity.type.toUpperCase()}</span>
		</div>

		<!-- ID -->
		<div class="tooltip-id">{entity.id}</div>

		<!-- タイトル -->
		<div class="tooltip-title">{entity.title}</div>

		<!-- プログレスバー -->
		<div class="tooltip-progress">
			<div class="progress-bar-container">
				<ProgressBar value={entity.progress} size="sm" showLabel={false} />
			</div>
			<span class="progress-label">{entity.progress}%</span>
		</div>

		<!-- メタ情報 -->
		<div class="tooltip-meta">
			<div class="meta-item">
				<span class="meta-label">Status:</span>
				<span class="meta-value" style="color: {getStatusColor(entity.status)}">
					{formatStatus(entity.status)}
				</span>
			</div>
			{#if entity.lastUpdate}
				<div class="meta-item">
					<span class="meta-label">Updated:</span>
					<span class="meta-value">{entity.lastUpdate}</span>
				</div>
			{/if}
		</div>

		<!-- ヒント -->
		<div class="tooltip-hint">
			<span>Double-click to view details</span>
		</div>
	</div>
{/if}

<style>
	.rich-tooltip {
		position: fixed;
		z-index: 1000;
		width: var(--tooltip-width, 320px);
		min-height: 180px;
		max-height: var(--tooltip-height, 220px);
		padding: var(--spacing-md, 12px);
		background: var(--bg-panel, #2d2d2d);
		border: 2px solid var(--border-metal, #4a4a4a);
		border-radius: var(--border-radius-md, 4px);
		box-shadow: var(--shadow-tooltip, 0 4px 20px rgba(0, 0, 0, 0.5));
		pointer-events: none;
		font-family: var(--font-family);
	}

	.tooltip-header {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 8px;
		color: var(--accent-primary, #ff9533);
	}

	.tooltip-type {
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.1em;
	}

	.tooltip-id {
		font-size: 12px;
		color: var(--text-muted, #888);
		margin-bottom: 4px;
		font-family: var(--font-family);
	}

	.tooltip-title {
		font-size: 14px;
		font-weight: 500;
		color: var(--text-primary, #fff);
		margin-bottom: 12px;
		line-height: 1.4;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.tooltip-progress {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 12px;
	}

	.progress-bar-container {
		flex: 1;
	}

	.progress-label {
		font-size: 12px;
		font-weight: 600;
		color: var(--text-primary);
		min-width: 36px;
		text-align: right;
	}

	.tooltip-meta {
		display: flex;
		flex-direction: column;
		gap: 4px;
		margin-bottom: 8px;
	}

	.meta-item {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 12px;
	}

	.meta-label {
		color: var(--text-muted);
	}

	.meta-value {
		color: var(--text-secondary);
	}

	.tooltip-hint {
		font-size: 10px;
		color: var(--text-muted);
		text-align: center;
		padding-top: 8px;
		border-top: 1px solid var(--border-dark, #333);
	}
</style>
