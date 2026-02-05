<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Minimap from './Minimap.svelte';
	import type { GraphNode } from '$lib/types/api';
	import type { NodePosition, LayoutResult } from '../engine/LayoutEngine';
	import type { Viewport } from '../engine/ViewerEngine';

	// モックノード
	const mockNodes: GraphNode[] = [
		{
			id: 'task-1',
			title: 'タスク 1',
			node_type: 'task',
			status: 'completed',
			priority: 'high',
			assignee: 'alice',
			dependencies: []
		},
		{
			id: 'task-2',
			title: 'タスク 2',
			node_type: 'task',
			status: 'completed',
			priority: 'medium',
			assignee: 'bob',
			dependencies: ['task-1']
		},
		{
			id: 'task-3',
			title: 'タスク 3',
			node_type: 'task',
			status: 'in_progress',
			priority: 'high',
			assignee: 'alice',
			dependencies: ['task-2']
		},
		{
			id: 'task-4',
			title: 'タスク 4',
			node_type: 'task',
			status: 'pending',
			priority: 'medium',
			assignee: 'charlie',
			dependencies: ['task-2']
		},
		{
			id: 'task-5',
			title: 'タスク 5',
			node_type: 'task',
			status: 'blocked',
			priority: 'low',
			assignee: 'bob',
			dependencies: ['task-3', 'task-4']
		}
	];

	// モック位置情報
	const mockPositions: Map<string, NodePosition> = new Map([
		['task-1', { id: 'task-1', x: 100, y: 100, layer: 0 }],
		['task-2', { id: 'task-2', x: 300, y: 100, layer: 1 }],
		['task-3', { id: 'task-3', x: 500, y: 50, layer: 2 }],
		['task-4', { id: 'task-4', x: 500, y: 150, layer: 2 }],
		['task-5', { id: 'task-5', x: 700, y: 100, layer: 3 }]
	]);

	// モック境界
	const mockBounds: LayoutResult['bounds'] = {
		minX: 50,
		minY: 0,
		maxX: 800,
		maxY: 200,
		width: 750,
		height: 200
	};

	// 通常ビューポート
	const normalViewport: Viewport = {
		x: 100,
		y: 50,
		width: 400,
		height: 300,
		scale: 1
	};

	const { Story } = defineMeta({
		title: 'Viewer/Minimap',
		component: Minimap,
		tags: ['autodocs'],
		parameters: {
			layout: 'padded'
		},
		args: {
			nodes: mockNodes,
			positions: mockPositions,
			bounds: mockBounds,
			viewport: normalViewport,
			onNavigate: () => {}
		}
	});
</script>

<script lang="ts">
	import { fn } from '@storybook/test';

	// Action ハンドラー
	const handleNavigate = fn();

	// ズームアウトビューポート
	const zoomedOutViewport: Viewport = {
		x: 50,
		y: 0,
		width: 800,
		height: 600,
		scale: 0.5
	};

	// ズームインビューポート
	const zoomedInViewport: Viewport = {
		x: 400,
		y: 80,
		width: 200,
		height: 150,
		scale: 2
	};

	// インタラクティブ用の状態
	let interactiveViewport: Viewport = $state({ ...normalViewport });

	function handleInteractiveNavigate(x: number, y: number) {
		interactiveViewport = {
			...interactiveViewport,
			x: x - interactiveViewport.width / interactiveViewport.scale / 2,
			y: y - interactiveViewport.height / interactiveViewport.scale / 2
		};
		handleNavigate(x, y);
	}
</script>

<!-- デフォルト -->
<Story name="Default">
	<div class="minimap-story-wrapper">
		<Minimap
			nodes={mockNodes}
			positions={mockPositions}
			bounds={mockBounds}
			viewport={normalViewport}
			onNavigate={handleNavigate}
		/>
	</div>
</Story>

<!-- ズームアウト時 -->
<Story name="ZoomedOut">
	<div class="minimap-story-wrapper">
		<Minimap
			nodes={mockNodes}
			positions={mockPositions}
			bounds={mockBounds}
			viewport={zoomedOutViewport}
			onNavigate={handleNavigate}
		/>
	</div>
</Story>

<!-- ズームイン時 -->
<Story name="ZoomedIn">
	<div class="minimap-story-wrapper">
		<Minimap
			nodes={mockNodes}
			positions={mockPositions}
			bounds={mockBounds}
			viewport={zoomedInViewport}
			onNavigate={handleNavigate}
		/>
	</div>
</Story>

<!-- 空のマップ -->
<Story name="Empty">
	<div class="minimap-story-wrapper">
		<Minimap
			nodes={[]}
			positions={new Map()}
			bounds={{ minX: 0, minY: 0, maxX: 100, maxY: 100, width: 100, height: 100 }}
			viewport={normalViewport}
			onNavigate={handleNavigate}
		/>
	</div>
</Story>

<!-- インタラクティブ -->
<Story name="Interactive" args={{ onNavigate: handleInteractiveNavigate }}>
	<div class="minimap-story-wrapper">
		<Minimap
			nodes={mockNodes}
			positions={mockPositions}
			bounds={mockBounds}
			viewport={interactiveViewport}
			onNavigate={handleInteractiveNavigate}
		/>
		<div class="viewport-info">
			<p style="color: #888; font-size: 11px; margin-bottom: 4px;">クリックでビューポートを移動</p>
			<pre style="color: #f5a623; font-size: 10px;">
x: {Math.round(interactiveViewport.x)}, y: {Math.round(interactiveViewport.y)}
			</pre>
		</div>
	</div>
</Story>

<!-- 状態色の凡例 -->
<Story name="StatusLegend">
	<div style="display: flex; gap: 24px; align-items: flex-start;">
		<div class="minimap-story-wrapper">
			<Minimap
				nodes={mockNodes}
				positions={mockPositions}
				bounds={mockBounds}
				viewport={normalViewport}
				onNavigate={handleNavigate}
			/>
		</div>
		<div style="padding: 12px; background: #2d2d2d; border-radius: 4px;">
			<p style="color: #888; font-size: 11px; margin-bottom: 8px; text-transform: uppercase;">
				Status Colors
			</p>
			<div style="display: flex; flex-direction: column; gap: 6px;">
				<div style="display: flex; align-items: center; gap: 8px;">
					<span style="width: 8px; height: 8px; border-radius: 50%; background: #22c55e;"></span>
					<span style="color: #888; font-size: 11px;">Completed</span>
				</div>
				<div style="display: flex; align-items: center; gap: 8px;">
					<span style="width: 8px; height: 8px; border-radius: 50%; background: #3b82f6;"></span>
					<span style="color: #888; font-size: 11px;">In Progress</span>
				</div>
				<div style="display: flex; align-items: center; gap: 8px;">
					<span style="width: 8px; height: 8px; border-radius: 50%; background: #f5a623;"></span>
					<span style="color: #888; font-size: 11px;">Pending</span>
				</div>
				<div style="display: flex; align-items: center; gap: 8px;">
					<span style="width: 8px; height: 8px; border-radius: 50%; background: #ef4444;"></span>
					<span style="color: #888; font-size: 11px;">Blocked</span>
				</div>
			</div>
		</div>
	</div>
</Story>

<style>
	/* Storybook 用ラッパー：Minimap の position: absolute を相対的に表示 */
	.minimap-story-wrapper {
		position: relative;
		min-height: 180px;
		width: 220px;
		background: #1a1a1a;
		padding: 16px;
		border: 1px solid #444;
	}

	/* Minimap の position を Storybook 表示用に上書き */
	.minimap-story-wrapper :global(.minimap) {
		position: static;
	}

	.viewport-info {
		margin-top: 16px;
		padding: 12px;
		background: #2d2d2d;
		border-radius: 4px;
	}
</style>
