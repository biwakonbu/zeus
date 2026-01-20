<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import * as d3 from 'd3';
	import type { CoverageAggregation, CoverageNode } from '$lib/types/api';

	// Props
	interface Props {
		data: CoverageAggregation | null;
		onNodeSelect?: (nodeId: string, nodeType: string) => void;
	}
	let { data, onNodeSelect }: Props = $props();

	// DOM 参照
	let containerEl: HTMLDivElement;
	let width = $state(0);
	let height = $state(0);

	// ノードタイプに基づく基本色
	function getNodeTypeColor(nodeType: string, hasIssue: boolean): string {
		if (hasIssue) {
			return '#ef4444'; // 問題あり
		}
		switch (nodeType) {
			case 'vision':
				return '#f5a623';
			case 'objective':
				return '#3b82f6';
			case 'deliverable':
				return '#22c55e';
			case 'task':
				return '#8b5cf6';
			default:
				return '#666';
		}
	}

	// 問題タイプに基づく色
	function getIssueTypeColor(issueType: string | undefined): string {
		switch (issueType) {
			case 'no_deliverables':
				return '#eab308'; // 黄（Deliverable なし）
			case 'no_tasks':
				return '#ef4444'; // 赤（Task なし）
			case 'orphaned':
				return '#9ca3af'; // 灰（孤立）
			default:
				return 'transparent';
		}
	}

	// サンバーストを描画
	function render() {
		if (!containerEl || !data?.root || width === 0 || height === 0) return;

		// 既存の SVG をクリア
		d3.select(containerEl).selectAll('svg').remove();

		const radius = Math.min(width, height) / 2 - 20;

		// 階層データに変換
		const hierarchy = d3
			.hierarchy<CoverageNode>(data.root)
			.sum((d) => d.value || 1)
			.sort((a, b) => (b.value || 0) - (a.value || 0));

		// パーティションレイアウト
		const partition = d3.partition<CoverageNode>().size([2 * Math.PI, radius]);

		const root = partition(hierarchy);

		// アーク生成器
		const arc = d3
			.arc<d3.HierarchyRectangularNode<CoverageNode>>()
			.startAngle((d) => d.x0)
			.endAngle((d) => d.x1)
			.padAngle((d) => Math.min((d.x1 - d.x0) / 2, 0.005))
			.padRadius(radius / 2)
			.innerRadius((d) => d.y0)
			.outerRadius((d) => d.y1 - 1);

		// SVG 作成
		const svg = d3
			.select(containerEl)
			.append('svg')
			.attr('width', width)
			.attr('height', height)
			.style('font-family', "'Inter', sans-serif");

		// 中心にグループを配置
		const g = svg.append('g').attr('transform', `translate(${width / 2},${height / 2})`);

		// アークパス
		g.selectAll('path')
			.data(root.descendants().filter((d) => d.depth > 0))
			.join('path')
			.attr('d', arc)
			.attr('fill', (d) => {
				const color = getNodeTypeColor(d.data.node_type, d.data.has_issue);
				// 深さに応じて明度を調整
				const lightness = 1 - d.depth * 0.15;
				const c = d3.color(color);
				if (c) {
					return c.brighter(lightness).formatHex();
				}
				return color;
			})
			.attr('stroke', (d) => {
				if (d.data.has_issue) {
					return getIssueTypeColor(d.data.issue_type);
				}
				return '#1a1a1a';
			})
			.attr('stroke-width', (d) => (d.data.has_issue ? 3 : 1))
			.style('cursor', 'pointer')
			.on('click', (event, d) => {
				event.stopPropagation();
				onNodeSelect?.(d.data.id, d.data.node_type);
			})
			.on('mouseenter', function (event, d) {
				d3.select(this).attr('stroke', '#fff').attr('stroke-width', 2);

				// ツールチップ表示
				const [x, y] = arc.centroid(d);
				const tooltip = d3.select(containerEl).select('.tooltip');
				tooltip
					.style('display', 'block')
					.style('left', `${width / 2 + x + 20}px`)
					.style('top', `${height / 2 + y}px`)
					.html(
						`
						<div class="tooltip-title">${d.data.title}</div>
						<div class="tooltip-row">
							<span class="label">タイプ:</span>
							<span class="value">${d.data.node_type}</span>
						</div>
						${
							d.data.has_issue
								? `
						<div class="tooltip-row issue">
							<span class="label">問題:</span>
							<span class="value">${d.data.issue_type || '不明'}</span>
						</div>
						`
								: ''
						}
					`
					);
			})
			.on('mouseleave', function (event, d) {
				d3.select(this)
					.attr('stroke', d.data.has_issue ? getIssueTypeColor(d.data.issue_type) : '#1a1a1a')
					.attr('stroke-width', d.data.has_issue ? 3 : 1);
				d3.select(containerEl).select('.tooltip').style('display', 'none');
			});

		// 中心にカバレッジスコアを表示
		g.append('text')
			.attr('text-anchor', 'middle')
			.attr('dy', '-0.2em')
			.attr('fill', '#e0e0e0')
			.attr('font-size', '24px')
			.attr('font-weight', '700')
			.text(`${data.coverage_score.toFixed(0)}%`);

		g.append('text')
			.attr('text-anchor', 'middle')
			.attr('dy', '1.2em')
			.attr('fill', '#888')
			.attr('font-size', '12px')
			.text('カバレッジ');

		// 凡例
		const legend = svg.append('g').attr('transform', `translate(20, ${height - 100})`);

		const legendItems = [
			{ color: '#eab308', label: 'Deliverable なし' },
			{ color: '#ef4444', label: 'Task なし' },
			{ color: '#9ca3af', label: '孤立エンティティ' }
		];

		legendItems.forEach((item, i) => {
			const row = legend.append('g').attr('transform', `translate(0, ${i * 20})`);

			row.append('rect').attr('width', 12).attr('height', 12).attr('fill', item.color).attr('rx', 2);

			row
				.append('text')
				.attr('x', 18)
				.attr('y', 10)
				.attr('fill', '#888')
				.attr('font-size', '11px')
				.text(item.label);
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
	let noData = $derived(!data?.root);
</script>

<div class="coverage-view" bind:this={containerEl}>
	<!-- ツールチップ -->
	<div class="tooltip"></div>

	{#if noData}
		<div class="empty-state">
			<span class="empty-icon">📐</span>
			<span class="empty-text">データがありません</span>
		</div>
	{/if}
</div>

<style>
	.coverage-view {
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
		font-size: 14px;
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

	:global(.tooltip-row.issue .value) {
		color: #ef4444;
	}
</style>
