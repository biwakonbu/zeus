<script lang="ts">
	import type { IssueSeverity, StaleRecommendation } from '$lib/types/api';

	// Props
	interface Props {
		type: 'issue' | 'stale';
		severity?: IssueSeverity;
		recommendation?: StaleRecommendation;
		count?: number;
		compact?: boolean;
	}
	let { type, severity, recommendation, count, compact = false }: Props = $props();

	// Issue の色とアイコン
	function getIssueStyle(s: IssueSeverity): { color: string; bgColor: string; icon: string } {
		switch (s) {
			case 'error':
				return { color: '#ef4444', bgColor: '#3b1515', icon: '⚠' };
			case 'warning':
				return { color: '#f59e0b', bgColor: '#3b2f15', icon: '⚡' };
			default:
				return { color: '#888', bgColor: '#2a2a2a', icon: '○' };
		}
	}

	// Stale の色とアイコン
	function getStaleStyle(r: StaleRecommendation): { color: string; bgColor: string; icon: string } {
		switch (r) {
			case 'archive':
				return { color: '#8b5cf6', bgColor: '#2d1f4e', icon: '📦' };
			case 'review':
				return { color: '#3b82f6', bgColor: '#1e2d4d', icon: '🔍' };
			case 'delete':
				return { color: '#ef4444', bgColor: '#3b1515', icon: '🗑' };
			default:
				return { color: '#888', bgColor: '#2a2a2a', icon: '○' };
		}
	}

	// スタイル計算
	let style = $derived.by(() => {
		if (type === 'issue' && severity) {
			return getIssueStyle(severity);
		} else if (type === 'stale' && recommendation) {
			return getStaleStyle(recommendation);
		}
		return { color: '#888', bgColor: '#2a2a2a', icon: '○' };
	});

	// ラベル
	function getLabel(): string {
		if (type === 'issue' && severity) {
			return severity === 'error' ? 'エラー' : '警告';
		} else if (type === 'stale' && recommendation) {
			switch (recommendation) {
				case 'archive':
					return 'アーカイブ推奨';
				case 'review':
					return 'レビュー推奨';
				case 'delete':
					return '削除推奨';
			}
		}
		return '';
	}
</script>

{#if compact}
	<span class="badge compact" style="background: {style.bgColor}; color: {style.color};">
		<span class="badge-icon">{style.icon}</span>
		{#if count !== undefined}
			<span class="badge-count">{count}</span>
		{/if}
	</span>
{:else}
	<div class="badge full" style="background: {style.bgColor}; border-color: {style.color};">
		<span class="badge-icon">{style.icon}</span>
		<span class="badge-label" style="color: {style.color};">{getLabel()}</span>
		{#if count !== undefined}
			<span class="badge-count" style="background: {style.color};">{count}</span>
		{/if}
	</div>
{/if}

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		border-radius: 4px;
	}

	.badge.compact {
		padding: 4px 8px;
		font-size: 12px;
	}

	.badge.full {
		padding: 6px 12px;
		border: 1px solid;
	}

	.badge-icon {
		font-size: 14px;
	}

	.badge-label {
		font-size: 12px;
		font-weight: 500;
	}

	.badge-count {
		font-size: 11px;
		font-weight: 700;
		padding: 2px 6px;
		border-radius: 10px;
		color: #1a1a1a;
		min-width: 20px;
		text-align: center;
	}

	.badge.compact .badge-count {
		background: currentColor;
		color: #1a1a1a;
	}
</style>
