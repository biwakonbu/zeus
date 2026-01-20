<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import * as d3 from 'd3';
	import type { ProgressAggregation } from '$lib/types/api';

	// Props
	interface Props {
		data: ProgressAggregation | null;
		onNodeSelect?: (nodeId: string, nodeType: string) => void;
	}
	let { data, onNodeSelect }: Props = $props();

	// DOM 参照
	let containerEl: HTMLDivElement;
	let width = $state(0);
	let height = $state(0);

	// 進捗率に基づく色を返す
	function getProgressColor(progress: number): string {
		if (progress >= 80) return '#22c55e'; // 緑
		if (progress >= 60) return '#4ade80'; // 明るい緑
		if (progress >= 40) return '#eab308'; // 黄
		if (progress >= 20) return '#f97316'; // オレンジ
		return '#ef4444'; // 赤
	}

	// ステータスに基づく境界色
	function getStatusBorderColor(status: string): string {
		switch (status) {
			case 'completed':
				return '#22c55e';
			case 'in_progress':
				return '#3b82f6';
			case 'not_started':
				return '#666';
			case 'on_hold':
				return '#f97316';
			default:
				return '#444';
		}
	}

	// ツリーマップノード型
	interface TreemapNode {
		name: string;
		id: string;
		progress: number;
		status: string;
		value?: number;
		children?: TreemapNode[];
	}

	// ツリーマップデータ構造に変換
	function buildHierarchy(aggData: ProgressAggregation): d3.HierarchyNode<TreemapNode> {
		const root: TreemapNode = {
			name: aggData.vision?.title || 'Vision',
			id: aggData.vision?.id || 'vision',
			progress: aggData.total_progress,
			status: aggData.vision?.status || 'in_progress',
			children: aggData.objectives.map((obj) => ({
				name: obj.title,
				id: obj.id,
				progress: obj.progress,
				status: obj.status,
				value: Math.max(obj.children_count, 1), // 最低1
				children: obj.children?.map((child) => ({
					name: child.title,
					id: child.id,
					progress: child.progress,
					status: child.status,
					value: Math.max(child.children_count, 1)
				}))
			}))
		};

		return d3.hierarchy(root).sum((d) => d.value || 1);
	}

	// ツリーマップを描画
	function render() {
		if (!containerEl || !data || width === 0 || height === 0) return;

		// 既存の SVG をクリア
		d3.select(containerEl).selectAll('svg').remove();

		const hierarchy = buildHierarchy(data);

		// ツリーマップレイアウト
		const treemap = d3
			.treemap<TreemapNode>()
			.size([width, height])
			.paddingOuter(4)
			.paddingTop(24)
			.paddingInner(2)
			.round(true);

		const root = treemap(hierarchy);

		// SVG 作成
		const svg = d3
			.select(containerEl)
			.append('svg')
			.attr('width', width)
			.attr('height', height)
			.style('font-family', "'Inter', sans-serif");

		// ノードグループ
		const nodes = svg
			.selectAll('g')
			.data(root.descendants().filter((d) => d.depth > 0))
			.join('g')
			.attr('transform', (d) => `translate(${d.x0},${d.y0})`);

		// 背景矩形
		nodes
			.append('rect')
			.attr('width', (d) => Math.max(0, d.x1 - d.x0))
			.attr('height', (d) => Math.max(0, d.y1 - d.y0))
			.attr('fill', (d) => {
				const progress = d.data.progress;
				const baseColor = getProgressColor(progress);
				// 深さに応じて透明度を調整
				const alpha = d.depth === 1 ? 0.8 : 0.6;
				return d3.color(baseColor)?.copy({ opacity: alpha })?.formatRgb() || baseColor;
			})
			.attr('stroke', (d) => getStatusBorderColor(d.data.status))
			.attr('stroke-width', 2)
			.attr('rx', 4)
			.style('cursor', 'pointer')
			.on('click', (event, d) => {
				event.stopPropagation();
				onNodeSelect?.(d.data.id, d.depth === 1 ? 'objective' : 'deliverable');
			})
			.on('mouseenter', function () {
				d3.select(this).attr('stroke-width', 3).attr('stroke', '#fff');
			})
			.on('mouseleave', function (_event, d) {
				d3.select(this).attr('stroke-width', 2).attr('stroke', getStatusBorderColor(d.data.status));
			});

		// プログレスバー背景
		nodes
			.filter((d) => d.x1 - d.x0 > 40 && d.y1 - d.y0 > 30)
			.append('rect')
			.attr('x', 4)
			.attr('y', (d) => Math.max(0, d.y1 - d.y0 - 10))
			.attr('width', (d) => Math.max(0, d.x1 - d.x0 - 8))
			.attr('height', 6)
			.attr('fill', 'rgba(0,0,0,0.3)')
			.attr('rx', 3);

		// プログレスバー
		nodes
			.filter((d) => d.x1 - d.x0 > 40 && d.y1 - d.y0 > 30)
			.append('rect')
			.attr('x', 4)
			.attr('y', (d) => Math.max(0, d.y1 - d.y0 - 10))
			.attr('width', (d) => Math.max(0, ((d.x1 - d.x0 - 8) * d.data.progress) / 100))
			.attr('height', 6)
			.attr('fill', '#fff')
			.attr('rx', 3)
			.style('opacity', 0.8);

		// タイトルテキスト
		nodes
			.filter((d) => d.x1 - d.x0 > 50 && d.y1 - d.y0 > 25)
			.append('text')
			.attr('x', 6)
			.attr('y', 16)
			.attr('fill', '#fff')
			.attr('font-size', (d) => (d.depth === 1 ? '12px' : '10px'))
			.attr('font-weight', (d) => (d.depth === 1 ? '600' : '400'))
			.text((d) => {
				const maxLen = Math.floor((d.x1 - d.x0) / 8);
				const name = d.data.name;
				return name.length > maxLen ? name.slice(0, maxLen - 1) + '…' : name;
			})
			.style('pointer-events', 'none')
			.style('text-shadow', '0 1px 2px rgba(0,0,0,0.5)');

		// 進捗率テキスト
		nodes
			.filter((d) => d.x1 - d.x0 > 60 && d.y1 - d.y0 > 40)
			.append('text')
			.attr('x', 6)
			.attr('y', 30)
			.attr('fill', 'rgba(255,255,255,0.8)')
			.attr('font-size', '10px')
			.text((d) => `${d.data.progress}%`)
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
</script>

<div class="progress-view" bind:this={containerEl}>
	{#if !data}
		<div class="empty-state">
			<span class="empty-icon">📊</span>
			<span class="empty-text">データがありません</span>
		</div>
	{/if}
</div>

<style>
	.progress-view {
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
</style>
