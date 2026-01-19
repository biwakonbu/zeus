<script lang="ts">
	// Health View
	// プロジェクト全体の健全性を一目で把握するためのビュー
	// メトリクスパネル + 階層リストで構成
	import MetricsPanel from './MetricsPanel.svelte';
	import ObjectiveList from './ObjectiveList.svelte';
	import { selectedEntityId, expandedIds, toggleExpand } from '../stores/wbsStore';
	import type { WBSAggregatedResponse, ProgressNode } from '$lib/types/api';

	interface Props {
		data: WBSAggregatedResponse | null;
		onNodeSelect: (nodeId: string, nodeType: string) => void;
	}
	let { data, onNodeSelect }: Props = $props();

	// メトリクス計算
	const coverage = $derived(data?.coverage?.coverage_score ?? 0);
	const objectives = $derived(data?.progress?.objectives ?? []);
	const balance = $derived(calculateBalance(objectives));
	const overallHealth = $derived(Math.round(coverage * 0.6 + balance * 0.4));
	const healthStatus = $derived<'good' | 'fair' | 'poor'>(
		overallHealth >= 70 ? 'good' : overallHealth >= 40 ? 'fair' : 'poor'
	);

	/**
	 * バランス（進捗の均一度）を計算
	 * 標準偏差が小さいほどバランスが良い
	 */
	function calculateBalance(objs: ProgressNode[]): number {
		if (objs.length === 0) return 0;
		const progresses = objs.map((o) => o.progress);
		const mean = progresses.reduce((a, b) => a + b, 0) / progresses.length;
		const variance =
			progresses.reduce((sum, p) => sum + Math.pow(p - mean, 2), 0) / progresses.length;
		const stdDev = Math.sqrt(variance);
		// 標準偏差を 0-100 のスコアに変換（stdDev が小さいほど高スコア）
		// stdDev が 50 以上なら 0、0 なら 100
		return Math.max(0, Math.min(100, Math.round(100 - stdDev * 2)));
	}

	function handleToggle(id: string) {
		toggleExpand(id);
	}
</script>

<div class="health-view">
	<MetricsPanel {coverage} {balance} {overallHealth} {healthStatus} />
	<ObjectiveList
		{objectives}
		selectedId={$selectedEntityId}
		expandedIds={$expandedIds}
		onSelect={onNodeSelect}
		onToggle={handleToggle}
	/>
</div>

<style>
	.health-view {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: var(--bg-primary, #1a1a1a);
	}
</style>
