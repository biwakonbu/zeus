<script lang="ts">
	import type { WBSAggregatedResponse } from '$lib/types/api';

	// Props
	interface Props {
		data: WBSAggregatedResponse | null;
	}
	let { data }: Props = $props();

	// 進捗率の色を決定
	function getProgressColor(progress: number): string {
		if (progress >= 80) return '#22c55e'; // 緑
		if (progress >= 50) return '#eab308'; // 黄
		if (progress >= 20) return '#f97316'; // オレンジ
		return '#ef4444'; // 赤
	}

	// 健全性スコアの色
	function getHealthColor(score: number): string {
		if (score >= 80) return '#22c55e';
		if (score >= 60) return '#eab308';
		if (score >= 40) return '#f97316';
		return '#ef4444';
	}

	// 派生データ
	let visionProgress = $derived(data?.progress?.total_progress ?? 0);
	let objectivesCompleted = $derived(
		data?.progress?.objectives.filter((o) => o.status === 'completed').length ?? 0
	);
	let objectivesTotal = $derived(data?.progress?.objectives.length ?? 0);
	let totalIssues = $derived(data?.issues?.total_issues ?? 0);
	let coverageScore = $derived(data?.coverage?.coverage_score ?? 0);
</script>

<div class="summary-bar">
	{#if data}
		<!-- Vision 進捗 -->
		<div class="summary-item">
			<span class="item-icon">🎯</span>
			<span class="item-label">Vision進捗</span>
			<span class="item-value" style="color: {getProgressColor(visionProgress)}">
				{visionProgress.toFixed(0)}%
			</span>
			<div class="mini-progress-bar">
				<div
					class="mini-progress-fill"
					style="width: {visionProgress}%; background: {getProgressColor(visionProgress)};"
				></div>
			</div>
		</div>

		<!-- Objectives 完了率 -->
		<div class="summary-item">
			<span class="item-icon">📊</span>
			<span class="item-label">Objectives</span>
			<span class="item-value">
				{objectivesCompleted}/{objectivesTotal}
				<span class="item-unit">完了</span>
			</span>
		</div>

		<!-- Issues 件数 -->
		<div class="summary-item" class:has-issues={totalIssues > 0}>
			<span class="item-icon">{totalIssues > 0 ? '⚠' : '✅'}</span>
			<span class="item-label">未解決Issues</span>
			<span class="item-value" class:warning={totalIssues > 0}>
				{totalIssues}件
			</span>
		</div>

		<!-- カバレッジスコア -->
		<div class="summary-item">
			<span class="item-icon">📐</span>
			<span class="item-label">カバレッジ</span>
			<span class="item-value" style="color: {getHealthColor(coverageScore)}">
				{coverageScore.toFixed(0)}%
			</span>
		</div>
	{:else}
		<div class="loading-placeholder">
			<span>データ読み込み中...</span>
		</div>
	{/if}
</div>

<style>
	.summary-bar {
		display: flex;
		align-items: center;
		gap: 24px;
		padding: 12px 20px;
		background: #1e1e1e;
		border-top: 1px solid #333;
		font-size: 13px;
	}

	.summary-item {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.item-icon {
		font-size: 14px;
	}

	.item-label {
		color: #888;
	}

	.item-value {
		font-weight: 600;
		color: #e0e0e0;
	}

	.item-value.warning {
		color: #f97316;
	}

	.item-unit {
		font-weight: 400;
		color: #888;
		font-size: 12px;
	}

	.summary-item.has-issues {
		padding: 4px 10px;
		background: rgba(249, 115, 22, 0.1);
		border-radius: 4px;
		border: 1px solid rgba(249, 115, 22, 0.3);
	}

	/* ミニプログレスバー */
	.mini-progress-bar {
		width: 60px;
		height: 4px;
		background: #333;
		border-radius: 2px;
		overflow: hidden;
	}

	.mini-progress-fill {
		height: 100%;
		transition: width 0.3s ease;
	}

	/* ローディング */
	.loading-placeholder {
		color: #666;
		font-style: italic;
	}
</style>
