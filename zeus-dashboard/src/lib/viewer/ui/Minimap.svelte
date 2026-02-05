<script lang="ts">
	import type { GraphNode, GraphNodeType } from '$lib/types/api';
	import type { NodePosition, LayoutResult } from '../engine/LayoutEngine';
	import type { Viewport } from '../engine/ViewerEngine';

	// Props
	interface Props {
		nodes: GraphNode[];
		isWBSMode?: boolean;
		positions: Map<string, NodePosition>;
		bounds: LayoutResult['bounds'];
		viewport: Viewport;
		onNavigate?: (x: number, y: number) => void;
	}

	let { nodes, isWBSMode = false, positions, bounds, viewport, onNavigate }: Props = $props();

	// ミニマップサイズ
	const MINIMAP_WIDTH = 180;
	const MINIMAP_HEIGHT = 120;
	const PADDING = 10;

	// スケール計算
	let scale = $derived.by(() => {
		if (bounds.width === 0 || bounds.height === 0) return 1;
		const scaleX = (MINIMAP_WIDTH - PADDING * 2) / bounds.width;
		const scaleY = (MINIMAP_HEIGHT - PADDING * 2) / bounds.height;
		return Math.min(scaleX, scaleY, 1);
	});

	// ビューポート矩形（ミニマップ座標系）
	let viewRect = $derived.by(() => {
		const vx = (viewport.x - bounds.minX) * scale + PADDING;
		const vy = (viewport.y - bounds.minY) * scale + PADDING;
		const vw = (viewport.width / viewport.scale) * scale;
		const vh = (viewport.height / viewport.scale) * scale;
		return { x: vx, y: vy, width: vw, height: vh };
	});

	// ノードを描画用に変換
	let nodeRects = $derived.by(() => {
		const rects: { x: number; y: number; status: string; nodeType: GraphNodeType }[] = [];
		for (const node of nodes) {
			const pos = positions.get(node.id);
			if (!pos) continue;
			rects.push({
				x: (pos.x - bounds.minX) * scale + PADDING,
				y: (pos.y - bounds.minY) * scale + PADDING,
				status: node.status,
				nodeType: node.node_type
			});
		}
		return rects;
	});

	// クリックでナビゲート
	function handleClick(e: MouseEvent) {
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		const clickX = e.clientX - rect.left;
		const clickY = e.clientY - rect.top;

		// ミニマップ座標をワールド座標に変換
		const worldX = (clickX - PADDING) / scale + bounds.minX;
		const worldY = (clickY - PADDING) / scale + bounds.minY;

		onNavigate?.(worldX, worldY);
	}

	// ドラッグでビューポート移動
	let isDragging = $state(false);

	function handleMouseDown(e: MouseEvent) {
		isDragging = true;
		handleClick(e);
	}

	function handleMouseMove(e: MouseEvent) {
		if (!isDragging) return;
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		const clickX = e.clientX - rect.left;
		const clickY = e.clientY - rect.top;

		const worldX = (clickX - PADDING) / scale + bounds.minX;
		const worldY = (clickY - PADDING) / scale + bounds.minY;

		onNavigate?.(worldX, worldY);
	}

	function handleMouseUp() {
		isDragging = false;
	}

	// ステータス色マッピング
	function getStatusColor(status: string): string {
		switch (status) {
			case 'completed':
				return 'var(--task-completed)';
			case 'in_progress':
				return 'var(--task-in-progress)';
			case 'pending':
				return 'var(--task-pending)';
			case 'blocked':
				return 'var(--task-blocked)';
			default:
				return 'var(--text-muted)';
		}
	}

	// ノードタイプ色マッピング（TaskNode.ts と同期）
	function getNodeTypeColor(nodeType: GraphNodeType): string {
		switch (nodeType) {
			case 'vision':
				return '#ffd700';  // ゴールド
			case 'objective':
				return '#6699ff';  // ブルー
			case 'deliverable':
				return '#66cc99';  // グリーン
			case 'task':
			default:
				return '#888888';  // グレー
		}
	}

	// モードに応じた色を取得
	function getNodeColor(node: { status: string; nodeType: GraphNodeType }): string {
		return isWBSMode ? getNodeTypeColor(node.nodeType) : getStatusColor(node.status);
	}
</script>

<div
	class="minimap"
	onmousedown={handleMouseDown}
	onmousemove={handleMouseMove}
	onmouseup={handleMouseUp}
	onmouseleave={handleMouseUp}
	role="button"
	tabindex="0"
>
	<div class="minimap-title">MAP</div>
	<svg width={MINIMAP_WIDTH} height={MINIMAP_HEIGHT} class="minimap-svg">
		<!-- 背景 -->
		<rect x="0" y="0" width={MINIMAP_WIDTH} height={MINIMAP_HEIGHT} fill="var(--bg-primary)" />

		<!-- ノード -->
		{#each nodeRects as node}
			<circle cx={node.x} cy={node.y} r="3" fill={getNodeColor(node)} opacity="0.8" />
		{/each}

		<!-- ビューポート領域 -->
		<rect
			x={viewRect.x}
			y={viewRect.y}
			width={viewRect.width}
			height={viewRect.height}
			fill="none"
			stroke="var(--accent-primary)"
			stroke-width="1.5"
			opacity="0.8"
		/>
	</svg>
</div>

<style>
	.minimap {
		position: absolute;
		bottom: 40px;
		right: var(--spacing-md);
		background-color: rgba(45, 45, 45, 0.95);
		border: 2px solid var(--border-metal);
		border-radius: var(--border-radius-sm);
		overflow: hidden;
		cursor: pointer;
		user-select: none;
	}

	.minimap:hover {
		border-color: var(--accent-primary);
	}

	.minimap-title {
		padding: 4px 8px;
		font-size: 10px;
		font-weight: 600;
		color: var(--accent-primary);
		text-transform: uppercase;
		letter-spacing: 0.1em;
		background-color: rgba(0, 0, 0, 0.3);
		border-bottom: 1px solid var(--border-dark);
	}

	.minimap-svg {
		display: block;
	}
</style>
