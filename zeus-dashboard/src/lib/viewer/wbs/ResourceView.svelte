<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import * as d3 from 'd3';
	import type { ResourceAggregation, ResourceCell } from '$lib/types/api';

	// Props
	interface Props {
		data: ResourceAggregation | null;
		onCellSelect?: (assignee: string, objective: string) => void;
	}
	let { data, onCellSelect }: Props = $props();

	// DOM 参照
	let containerEl: HTMLDivElement;
	let width = $state(0);
	let height = $state(0);

	// マージン設定
	const margin = { top: 60, right: 20, bottom: 20, left: 120 };

	// タスク数に基づく色スケール
	function getTaskCountColor(count: number): string {
		if (count === 0) return '#1a1a1a';
		if (count <= 2) return '#1e3a5f';
		if (count <= 5) return '#2563eb';
		if (count <= 10) return '#3b82f6';
		return '#60a5fa';
	}

	// 進捗率に基づく色（テキスト用）
	function getProgressColor(progress: number): string {
		if (progress >= 80) return '#22c55e';
		if (progress >= 50) return '#eab308';
		if (progress >= 20) return '#f97316';
		return '#ef4444';
	}

	// ヒートマップを描画
	function render() {
		if (!containerEl || !data || width === 0 || height === 0) return;

		// 既存の SVG をクリア
		d3.select(containerEl).selectAll('svg').remove();

		const { assignees, objectives, matrix } = data;

		if (assignees.length === 0 || objectives.length === 0) return;

		// 描画領域の計算
		const innerWidth = width - margin.left - margin.right;
		const innerHeight = height - margin.top - margin.bottom;

		const cellWidth = Math.min(60, innerWidth / assignees.length);
		const cellHeight = Math.min(40, innerHeight / objectives.length);

		// SVG 作成
		const svg = d3
			.select(containerEl)
			.append('svg')
			.attr('width', width)
			.attr('height', height)
			.style('font-family', "'Inter', sans-serif");

		const g = svg.append('g').attr('transform', `translate(${margin.left},${margin.top})`);

		// 横軸（担当者）
		const xScale = d3
			.scaleBand()
			.domain(assignees)
			.range([0, cellWidth * assignees.length])
			.padding(0.05);

		// 縦軸（Objective）
		const yScale = d3
			.scaleBand()
			.domain(objectives)
			.range([0, cellHeight * objectives.length])
			.padding(0.05);

		// 横軸ラベル
		g.append('g')
			.selectAll('text')
			.data(assignees)
			.join('text')
			.attr('x', (d) => (xScale(d) || 0) + xScale.bandwidth() / 2)
			.attr('y', -10)
			.attr('text-anchor', 'middle')
			.attr('fill', '#888')
			.attr('font-size', '11px')
			.text((d) => (d.length > 8 ? d.slice(0, 8) + '…' : d));

		// 縦軸ラベル
		g.append('g')
			.selectAll('text')
			.data(objectives)
			.join('text')
			.attr('x', -8)
			.attr('y', (d) => (yScale(d) || 0) + yScale.bandwidth() / 2)
			.attr('text-anchor', 'end')
			.attr('dominant-baseline', 'middle')
			.attr('fill', '#888')
			.attr('font-size', '11px')
			.text((d) => (d.length > 12 ? d.slice(0, 12) + '…' : d));

		// セルデータを展開
		const cellData: Array<{
			assignee: string;
			objective: string;
			cell: ResourceCell;
			row: number;
			col: number;
		}> = [];

		matrix.forEach((row, rowIndex) => {
			row.forEach((cell, colIndex) => {
				cellData.push({
					assignee: assignees[colIndex],
					objective: objectives[rowIndex],
					cell,
					row: rowIndex,
					col: colIndex
				});
			});
		});

		// セル描画
		const cells = g
			.selectAll('g.cell')
			.data(cellData)
			.join('g')
			.attr('class', 'cell')
			.attr(
				'transform',
				(d) => `translate(${xScale(d.assignee) || 0},${yScale(d.objective) || 0})`
			);

		// セル背景
		cells
			.append('rect')
			.attr('width', xScale.bandwidth())
			.attr('height', yScale.bandwidth())
			.attr('fill', (d) => getTaskCountColor(d.cell.task_count))
			.attr('stroke', '#333')
			.attr('stroke-width', 1)
			.attr('rx', 2)
			.style('cursor', (d) => (d.cell.task_count > 0 ? 'pointer' : 'default'))
			.on('click', (event, d) => {
				if (d.cell.task_count > 0) {
					event.stopPropagation();
					onCellSelect?.(d.assignee, d.objective);
				}
			})
			.on('mouseenter', function (event, d) {
				if (d.cell.task_count > 0) {
					d3.select(this).attr('stroke', '#fff').attr('stroke-width', 2);

					// ツールチップ表示
					const tooltip = d3.select(containerEl).select('.tooltip');
					tooltip
						.style('display', 'block')
						.style('left', `${margin.left + (xScale(d.assignee) || 0) + xScale.bandwidth() + 10}px`)
						.style('top', `${margin.top + (yScale(d.objective) || 0)}px`)
						.html(
							`
							<div class="tooltip-title">${d.assignee}</div>
							<div class="tooltip-subtitle">${d.objective}</div>
							<div class="tooltip-row">
								<span class="label">タスク数:</span>
								<span class="value">${d.cell.task_count}</span>
							</div>
							<div class="tooltip-row">
								<span class="label">進捗:</span>
								<span class="value" style="color: ${getProgressColor(d.cell.progress)}">${d.cell.progress}%</span>
							</div>
							${
								d.cell.blocked_count > 0
									? `
							<div class="tooltip-row blocked">
								<span class="label">ブロック:</span>
								<span class="value">${d.cell.blocked_count}</span>
							</div>
							`
									: ''
							}
						`
						);
				}
			})
			.on('mouseleave', function () {
				d3.select(this).attr('stroke', '#333').attr('stroke-width', 1);
				d3.select(containerEl).select('.tooltip').style('display', 'none');
			});

		// タスク数テキスト（セルが十分大きい場合）
		cells
			.filter((d) => d.cell.task_count > 0 && xScale.bandwidth() > 30 && yScale.bandwidth() > 20)
			.append('text')
			.attr('x', xScale.bandwidth() / 2)
			.attr('y', yScale.bandwidth() / 2)
			.attr('text-anchor', 'middle')
			.attr('dominant-baseline', 'middle')
			.attr('fill', '#fff')
			.attr('font-size', '12px')
			.attr('font-weight', '600')
			.text((d) => d.cell.task_count)
			.style('pointer-events', 'none');

		// ブロックマーカー（セルが十分大きい場合）
		cells
			.filter(
				(d) => d.cell.blocked_count > 0 && xScale.bandwidth() > 40 && yScale.bandwidth() > 25
			)
			.append('circle')
			.attr('cx', xScale.bandwidth() - 8)
			.attr('cy', 8)
			.attr('r', 5)
			.attr('fill', '#ef4444')
			.style('pointer-events', 'none');

		// 凡例
		const legend = svg.append('g').attr('transform', `translate(${width - 150}, 20)`);

		legend
			.append('text')
			.attr('fill', '#888')
			.attr('font-size', '11px')
			.attr('font-weight', '500')
			.text('タスク数');

		const legendItems = [
			{ count: 0, label: '0' },
			{ count: 2, label: '1-2' },
			{ count: 5, label: '3-5' },
			{ count: 10, label: '6-10' },
			{ count: 15, label: '11+' }
		];

		legendItems.forEach((item, i) => {
			legend
				.append('rect')
				.attr('x', i * 24)
				.attr('y', 16)
				.attr('width', 20)
				.attr('height', 12)
				.attr('fill', getTaskCountColor(item.count))
				.attr('stroke', '#333')
				.attr('rx', 2);
		});
	}

	// リサイズ監視
	let resizeObserver: ResizeObserver | null = null;

	onMount(() => {
		if (containerEl) {
			resizeObserver = new ResizeObserver((entries) => {
				for (const entry of entries) {
					width = entry.contentRect.width;
					height = entry.contentRect.height;
				}
			});
			resizeObserver.observe(containerEl);
		}
	});

	onDestroy(() => {
		resizeObserver?.disconnect();
	});

	// データまたはサイズ変更時に再描画
	$effect(() => {
		if (data && width > 0 && height > 0) {
			render();
		}
	});

	// データなし判定
	let noData = $derived(!data || data.assignees.length === 0 || data.objectives.length === 0);
</script>

<div class="resource-view" bind:this={containerEl}>
	<!-- ツールチップ -->
	<div class="tooltip"></div>

	{#if noData}
		<div class="empty-state">
			<span class="empty-icon">👥</span>
			<span class="empty-text">リソースデータがありません</span>
			<span class="empty-subtext">担当者が割り当てられたタスクがありません</span>
		</div>
	{/if}
</div>

<style>
	.resource-view {
		width: 100%;
		height: 100%;
		background: #1a1a1a;
		overflow: hidden;
		position: relative;
	}

	.empty-state {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		color: #666;
	}

	.empty-icon {
		font-size: 48px;
		opacity: 0.5;
	}

	.empty-text {
		font-size: 16px;
		font-weight: 500;
	}

	.empty-subtext {
		font-size: 12px;
		opacity: 0.7;
	}

	/* ツールチップ */
	.tooltip {
		display: none;
		position: absolute;
		background: #2a2a2a;
		border: 1px solid #444;
		border-radius: 6px;
		padding: 12px;
		font-size: 12px;
		color: #e0e0e0;
		z-index: 100;
		pointer-events: none;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
		max-width: 250px;
	}

	:global(.tooltip-title) {
		font-weight: 600;
		margin-bottom: 4px;
	}

	:global(.tooltip-subtitle) {
		font-size: 11px;
		color: #888;
		margin-bottom: 8px;
		padding-bottom: 6px;
		border-bottom: 1px solid #444;
	}

	:global(.tooltip-row) {
		display: flex;
		justify-content: space-between;
		gap: 12px;
		margin-top: 4px;
	}

	:global(.tooltip-row .label) {
		color: #888;
	}

	:global(.tooltip-row .value) {
		font-weight: 500;
	}

	:global(.tooltip-row.blocked .value) {
		color: #ef4444;
	}
</style>
