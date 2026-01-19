<script lang="ts">
	// Timeline View
	// 計画 vs 実績の時間的乖離を可視化するビュー
	import TimelineScale from './TimelineScale.svelte';
	import TimelineBar from './TimelineBar.svelte';
	import { selectedEntityId } from '../stores/wbsStore';
	import type { WBSAggregatedResponse, ProgressNode } from '$lib/types/api';

	interface Props {
		data: WBSAggregatedResponse | null;
		onNodeSelect: (nodeId: string, nodeType: string) => void;
	}
	let { data, onNodeSelect }: Props = $props();

	// スケール状態
	let scale: 'week' | 'month' | 'quarter' = $state('month');

	// タイムラインの範囲を計算
	const objectives = $derived(data?.progress?.objectives ?? []);
	const timelineRange = $derived(calculateTimelineRange(objectives));

	function calculateTimelineRange(objs: ProgressNode[]): { start: Date; end: Date } {
		const now = new Date();
		// デフォルト: 現在から前後 3 ヶ月
		const defaultStart = new Date(now);
		defaultStart.setMonth(defaultStart.getMonth() - 1);
		const defaultEnd = new Date(now);
		defaultEnd.setMonth(defaultEnd.getMonth() + 2);

		return {
			start: defaultStart,
			end: defaultEnd
		};
	}

	function handleScaleChange(newScale: 'week' | 'month' | 'quarter') {
		scale = newScale;
	}

	function handleTodayClick() {
		// Today へスクロール（将来実装）
		console.log('Scroll to today');
	}

	// ステータスを判定
	function getStatus(
		progress: number,
		status: string
	): 'on_track' | 'delayed' | 'ahead' | 'completed' {
		if (status === 'completed') return 'completed';
		if (progress >= 100) return 'completed';
		// 簡易判定（実際には due_date との比較が必要）
		if (progress >= 70) return 'ahead';
		if (progress >= 30) return 'on_track';
		return 'delayed';
	}

	// 仮の日付を生成（実際のデータに start_date, due_date があれば使用）
	function getMockDates(obj: ProgressNode, index: number): { start: Date; end: Date } {
		const start = new Date();
		start.setDate(start.getDate() - 30 + index * 7);
		const end = new Date(start);
		end.setDate(end.getDate() + 30);
		return { start, end };
	}
</script>

<div class="timeline-view">
	<TimelineScale
		{scale}
		startDate={timelineRange.start}
		endDate={timelineRange.end}
		onScaleChange={handleScaleChange}
		onTodayClick={handleTodayClick}
	/>
	<div class="timeline-content">
		{#each objectives as obj, index (obj.id)}
			{@const dates = getMockDates(obj, index)}
			{@const status = getStatus(obj.progress, obj.status)}
			<TimelineBar
				id={obj.id}
				title={obj.title}
				planStart={dates.start}
				planEnd={dates.end}
				progress={obj.progress}
				{status}
				timelineStart={timelineRange.start}
				timelineEnd={timelineRange.end}
				selected={$selectedEntityId === obj.id}
				onClick={() => onNodeSelect(obj.id, 'objective')}
			/>
		{/each}

		{#if objectives.length === 0}
			<div class="empty-state">
				<span class="empty-icon">📅</span>
				<span class="empty-text">タイムラインデータがありません</span>
			</div>
		{/if}
	</div>
</div>

<style>
	.timeline-view {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: var(--bg-primary, #1a1a1a);
	}

	.timeline-content {
		flex: 1;
		overflow-y: auto;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 48px 16px;
		color: var(--text-muted, #888888);
	}

	.empty-icon {
		font-size: 32px;
		opacity: 0.5;
		margin-bottom: 8px;
	}

	.empty-text {
		font-size: 13px;
	}
</style>
