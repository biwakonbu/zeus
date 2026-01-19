<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import * as d3 from 'd3';
	import type { IssueAggregation, IssueBubble } from '$lib/types/api';

	// Props
	interface Props {
		data: IssueAggregation | null;
		onNodeSelect?: (nodeId: string, nodeType: string) => void;
	}
	let { data, onNodeSelect }: Props = $props();

	// DOM 参照
	let containerEl: HTMLDivElement;
	let width = $state(0);
	let height = $state(0);

	// 深刻度に基づく色
	function getSeverityColor(severity: string): string {
		switch (severity) {
			case 'critical':
				return '#ef4444';
			case 'high':
				return '#f97316';
			case 'medium':
				return '#eab308';
			case 'low':
				return '#22c55e';
			default:
				return '#666';
		}
	}

	// ノードタイプアイコン
	function getNodeTypeIcon(nodeType: string): string {
		switch (nodeType) {
			case 'vision':
				return '🎯';
			case 'objective':
				return '📊';
			case 'deliverable':
				return '📦';
			default:
				return '📄';
		}
	}

	// バブルチャートを描画
	function render() {
		if (!containerEl || !data || width === 0 || height === 0) return;

		// 既存の SVG をクリア
		d3.select(containerEl).selectAll('svg').remove();

		// 問題がない場合
		if (data.items.length === 0) {
			return;
		}

		// バブルレイアウト用データ
		const bubbleData = data.items.map((item) => ({
			...item,
			value: item.total_issues + 1 // 最低サイズを確保
		}));

		// 階層データに変換（D3 pack用）
		const hierarchy = d3
			.hierarchy({ children: bubbleData } as { children: IssueBubble[] })
			.sum((d: any) => d.value || 0);

		// パックレイアウト
		const pack = d3
			.pack<{ children: IssueBubble[] }>()
			.size([width - 20, height - 20])
			.padding(8);

		const root = pack(hierarchy);

		// SVG 作成
		const svg = d3
			.select(containerEl)
			.append('svg')
			.attr('width', width)
			.attr('height', height)
			.style('font-family', "'Inter', sans-serif");

		// バブルグループ
		const bubbles = svg
			.selectAll('g.bubble')
			.data(root.leaves())
			.join('g')
			.attr('class', 'bubble')
			.attr('transform', (d) => `translate(${d.x + 10},${d.y + 10})`);

		// バブル円
		bubbles
			.append('circle')
			.attr('r', (d) => d.r)
			.attr('fill', (d) => {
				const item = d.data as unknown as IssueBubble;
				const color = getSeverityColor(item.max_severity);
				return d3.color(color)?.copy({ opacity: 0.7 })?.formatRgb() || color;
			})
			.attr('stroke', (d) => {
				const item = d.data as unknown as IssueBubble;
				return getSeverityColor(item.max_severity);
			})
			.attr('stroke-width', 2)
			.style('cursor', 'pointer')
			.on('click', (event, d) => {
				event.stopPropagation();
				const item = d.data as unknown as IssueBubble;
				onNodeSelect?.(item.id, item.node_type);
			})
			.on('mouseenter', function (event, d) {
				d3.select(this).attr('stroke-width', 3).attr('stroke', '#fff');

				// ツールチップ表示
				const item = d.data as unknown as IssueBubble;
				const tooltip = d3.select(containerEl).select('.tooltip');
				tooltip
					.style('display', 'block')
					.style('left', `${d.x + 10 + d.r + 10}px`)
					.style('top', `${d.y + 10}px`)
					.html(
						`
						<div class="tooltip-title">${getNodeTypeIcon(item.node_type)} ${item.title}</div>
						<div class="tooltip-row">
							<span class="label">Problem:</span>
							<span class="value">${item.problem_count}件</span>
						</div>
						<div class="tooltip-row">
							<span class="label">Risk:</span>
							<span class="value">${item.risk_count}件</span>
						</div>
						<div class="tooltip-row">
							<span class="label">リスクスコア:</span>
							<span class="value">${item.risk_score.toFixed(1)}</span>
						</div>
						<div class="tooltip-row">
							<span class="label">進捗:</span>
							<span class="value">${item.progress}%</span>
						</div>
					`
					);
			})
			.on('mouseleave', function (event, d) {
				const item = d.data as unknown as IssueBubble;
				d3.select(this).attr('stroke-width', 2).attr('stroke', getSeverityColor(item.max_severity));
				d3.select(containerEl).select('.tooltip').style('display', 'none');
			});

		// 問題件数テキスト（大きいバブルのみ）
		bubbles
			.filter((d) => d.r > 25)
			.append('text')
			.attr('text-anchor', 'middle')
			.attr('dy', '0.35em')
			.attr('fill', '#fff')
			.attr('font-size', (d) => Math.min(d.r / 2, 16) + 'px')
			.attr('font-weight', '700')
			.text((d) => {
				const item = d.data as unknown as IssueBubble;
				return item.total_issues;
			})
			.style('pointer-events', 'none')
			.style('text-shadow', '0 1px 2px rgba(0,0,0,0.5)');

		// タイトルテキスト（さらに大きいバブルのみ）
		bubbles
			.filter((d) => d.r > 40)
			.append('text')
			.attr('text-anchor', 'middle')
			.attr('dy', (d) => d.r / 2 + 'px')
			.attr('fill', 'rgba(255,255,255,0.8)')
			.attr('font-size', '10px')
			.text((d) => {
				const item = d.data as unknown as IssueBubble;
				const maxLen = Math.floor(d.r / 4);
				return item.title.length > maxLen ? item.title.slice(0, maxLen) + '…' : item.title;
			})
			.style('pointer-events', 'none');
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

	// 問題なし判定
	let noIssues = $derived(data && data.items.length === 0);
</script>

<div class="issue-view" bind:this={containerEl}>
	<!-- ツールチップ -->
	<div class="tooltip"></div>

	{#if !data}
		<div class="empty-state">
			<span class="empty-icon">🔍</span>
			<span class="empty-text">データがありません</span>
		</div>
	{:else if noIssues}
		<div class="empty-state success">
			<span class="empty-icon">✅</span>
			<span class="empty-text">問題は見つかりませんでした</span>
			<span class="empty-subtext">すべてのエンティティが正常です</span>
		</div>
	{/if}
</div>

<style>
	.issue-view {
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

	.empty-state.success {
		color: #22c55e;
	}

	.empty-icon {
		font-size: 48px;
		opacity: 0.7;
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
</style>
