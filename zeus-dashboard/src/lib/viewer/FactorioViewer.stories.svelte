<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import FactorioViewer from './FactorioViewer.svelte';

	const { Story } = defineMeta({
		title: 'Viewer/FactorioViewer',
		component: FactorioViewer,
		tags: ['autodocs'],
		parameters: {
			layout: 'fullscreen',
			docs: {
				story: {
					iframeHeight: 600
				}
			}
		}
	});
</script>

<script lang="ts">
	import { fn } from '@storybook/test';
	import type { GraphNode, GraphEdge } from '$lib/types/api';

	// グラフデータ型
	interface GraphData {
		nodes: GraphNode[];
		edges: GraphEdge[];
	}

	interface StoryNode extends GraphNode {
		dependencies: string[];
	}

	// Action ハンドラー
	const handleTaskSelect = fn();
	const handleTaskHover = fn();

	// モックノード（少数）
	const simpleNodes: StoryNode[] = [
		{
			id: 'task-1',
			title: 'プロジェクト設計',
			node_type: 'activity',
			status: 'deprecated',
			dependencies: []
		},
		{
			id: 'task-2',
			title: 'データベース設計',
			node_type: 'activity',
			status: 'deprecated',
			dependencies: ['task-1']
		},
		{
			id: 'task-3',
			title: 'API 実装',
			node_type: 'activity',
			status: 'active',
			dependencies: ['task-2']
		},
		{
			id: 'task-4',
			title: 'フロントエンド実装',
			node_type: 'activity',
			status: 'draft',
			dependencies: ['task-2']
		},
		{
			id: 'task-5',
			title: '統合テスト',
			node_type: 'activity',
			status: 'draft',
			dependencies: ['task-3', 'task-4']
		}
	];

	// より多くのノード
	const complexNodes: StoryNode[] = [
		// レイヤー1
		{
			id: 't1',
			title: 'プロジェクト立ち上げ',
			node_type: 'activity',
			status: 'deprecated',
			dependencies: []
		},
		// レイヤー2
		{
			id: 't2',
			title: '要件定義',
			node_type: 'activity',
			status: 'deprecated',
			dependencies: ['t1']
		},
		{
			id: 't3',
			title: 'チーム編成',
			node_type: 'activity',
			status: 'deprecated',
			dependencies: ['t1']
		},
		// レイヤー3
		{
			id: 't4',
			title: 'アーキテクチャ設計',
			node_type: 'activity',
			status: 'deprecated',
			dependencies: ['t2']
		},
		{
			id: 't5',
			title: 'UI/UX デザイン',
			node_type: 'activity',
			status: 'active',
			dependencies: ['t2']
		},
		{
			id: 't6',
			title: 'インフラ設計',
			node_type: 'activity',
			status: 'deprecated',
			dependencies: ['t2', 't3']
		},
		// レイヤー4
		{
			id: 't7',
			title: 'バックエンド開発',
			node_type: 'activity',
			status: 'active',
			dependencies: ['t4']
		},
		{
			id: 't8',
			title: 'フロントエンド開発',
			node_type: 'activity',
			status: 'draft',
			dependencies: ['t4', 't5']
		},
		{
			id: 't9',
			title: 'CI/CD 構築',
			node_type: 'activity',
			status: 'active',
			dependencies: ['t6']
		},
		// レイヤー5
		{
			id: 't10',
			title: 'API 統合',
			node_type: 'activity',
			status: 'draft',
			dependencies: ['t7', 't8']
		},
		{
			id: 't11',
			title: 'パフォーマンス最適化',
			node_type: 'activity',
			status: 'draft',
			dependencies: ['t7']
		},
		// レイヤー6
		{
			id: 't12',
			title: '結合テスト',
			node_type: 'activity',
			status: 'draft',
			dependencies: ['t10', 't9']
		},
		{
			id: 't13',
			title: 'セキュリティ監査',
			node_type: 'activity',
			status: 'draft',
			dependencies: ['t10']
		},
		// レイヤー7
		{
			id: 't14',
			title: 'ステージングデプロイ',
			node_type: 'activity',
			status: 'draft',
			dependencies: ['t12', 't13']
		},
		// レイヤー8
		{
			id: 't15',
			title: '本番リリース',
			node_type: 'activity',
			status: 'draft',
			dependencies: ['t14']
		}
	];

	// 空のノード
	const emptyNodes: StoryNode[] = [];

	// GraphNode 配列を GraphData に変換するヘルパー
	function toGraphData(nodes: StoryNode[]): GraphData {
		const nodeById = new Map(nodes.map((node) => [node.id, node]));
		const edges: GraphEdge[] = [];

		for (const node of nodes) {
			for (const dep of node.dependencies) {
				const depNode = nodeById.get(dep);
				if (!depNode) continue;

				// structural は child -> parent（レイアウト用）
				edges.push({
					from: node.id,
					to: dep,
					layer: 'structural',
					relation: 'parent'
				});
			}
		}

		const graphNodes: GraphNode[] = nodes.map(({ dependencies: _dependencies, ...node }) => node);
		return { nodes: graphNodes, edges };
	}

	// 選択中のタスクID
	let selectedTaskId: string | null = $state(null);

	function handleInteractiveSelect(taskId: string | null) {
		selectedTaskId = taskId;
		handleTaskSelect(taskId);
	}
</script>

<!-- デフォルト（シンプルなノード） -->
<Story name="Default">
	<div style="height: 600px; background: var(--bg-primary);">
		<FactorioViewer
			graphData={toGraphData(simpleNodes)}
			onTaskSelect={handleTaskSelect}
			onTaskHover={handleTaskHover}
		/>
	</div>
</Story>

<!-- 複雑なグラフ -->
<Story name="ComplexGraph">
	<div style="height: 700px; background: var(--bg-primary);">
		<FactorioViewer
			graphData={toGraphData(complexNodes)}
			onTaskSelect={handleTaskSelect}
			onTaskHover={handleTaskHover}
		/>
	</div>
</Story>

<!-- ノードなし -->
<Story name="Empty">
	<div style="height: 400px; background: var(--bg-primary);">
		<FactorioViewer
			graphData={toGraphData(emptyNodes)}
			onTaskSelect={handleTaskSelect}
			onTaskHover={handleTaskHover}
		/>
	</div>
</Story>

<!-- ノード選択済み -->
<Story name="WithSelection">
	<div style="height: 600px; background: var(--bg-primary);">
		<FactorioViewer
			graphData={toGraphData(simpleNodes)}
			selectedTaskId="task-3"
			onTaskSelect={handleTaskSelect}
			onTaskHover={handleTaskHover}
		/>
	</div>
</Story>

<!-- インタラクティブ -->
<Story name="Interactive">
	<div style="height: 650px; background: var(--bg-primary); position: relative;">
		<FactorioViewer
			graphData={toGraphData(complexNodes)}
			{selectedTaskId}
			onTaskSelect={handleInteractiveSelect}
			onTaskHover={handleTaskHover}
		/>
		<div
			style="position: absolute; top: 60px; right: 60px; background: var(--bg-panel); padding: 12px; border-radius: 4px; border: 1px solid var(--border-metal);"
		>
			<p style="color: var(--text-secondary); font-size: 11px; margin: 0 0 4px 0;">
				選択中のノード:
			</p>
			<p style="color: var(--accent-primary); font-size: 12px; margin: 0;">
				{selectedTaskId || 'なし'}
			</p>
		</div>
	</div>
</Story>

<!-- 全ステータスのノード -->
<Story name="AllStatuses">
	{@const allStatusNodes: StoryNode[] = [
		{ id: 'deprecated-1', title: '非推奨タスク 1', node_type: 'activity', status: 'deprecated', dependencies: [] },
		{ id: 'deprecated-2', title: '非推奨タスク 2', node_type: 'activity', status: 'deprecated', dependencies: ['deprecated-1'] },
		{ id: 'active-1', title: 'アクティブタスク', node_type: 'activity', status: 'active', dependencies: ['deprecated-2'] },
		{ id: 'draft-1', title: '下書きタスク', node_type: 'activity', status: 'draft', dependencies: ['deprecated-2'] },
		{ id: 'draft-2', title: '下書きタスク 2', node_type: 'activity', status: 'draft', dependencies: ['active-1', 'draft-1'] }
	]}
	<div style="height: 500px; background: var(--bg-primary);">
		<FactorioViewer
			graphData={toGraphData(allStatusNodes)}
			onTaskSelect={handleTaskSelect}
			onTaskHover={handleTaskHover}
		/>
	</div>
</Story>

<!-- 大規模グラフ（100+ ノード） -->
<Story name="LargeGraph">
	{@const generateLargeNodes = () => {
		const nodes: StoryNode[] = [];
		const statuses = ['deprecated', 'active', 'draft'] as const;

		// 120 ノードを生成（12 レイヤー × 10 ノード）
		for (let layer = 0; layer < 12; layer++) {
			for (let node = 0; node < 10; node++) {
				const id = `task-${layer}-${node}`;
				const deps: string[] = [];

				// 前レイヤーからランダムに 1-3 個の依存を追加
				if (layer > 0) {
					const numDeps = Math.min(3, Math.floor(Math.random() * 3) + 1);
					for (let d = 0; d < numDeps; d++) {
						const depNode = Math.floor(Math.random() * 10);
						const depId = `task-${layer - 1}-${depNode}`;
						if (!deps.includes(depId)) deps.push(depId);
					}
				}

				const statusIdx =
					layer < 4 ? 0 : layer < 8 ? Math.min(layer - 4, 1) : Math.min(layer - 8, 3);
				nodes.push({
					id,
					title: `タスク ${layer + 1}-${node + 1}`,
					node_type: 'activity',
					status: statuses[statusIdx],
					dependencies: deps
				});
			}
		}
		return nodes;
	}}
	{@const largeNodes = generateLargeNodes()}
	<div style="height: 800px; background: var(--bg-primary);">
		<FactorioViewer
			graphData={toGraphData(largeNodes)}
			onTaskSelect={handleTaskSelect}
			onTaskHover={handleTaskHover}
		/>
	</div>
</Story>

<!-- タイプ別グラフ（階層表示） -->
<Story name="TypedGraph">
	{@const typedNodes: StoryNode[] = [
		// Vision
		{ id: 'vision-1', title: 'プロジェクト管理の革新', node_type: 'vision', status: 'active', dependencies: [] },
		// Objectives
		{ id: 'obj-1', title: 'Phase 1: MVP 開発', node_type: 'objective', status: 'deprecated', dependencies: ['vision-1'] },
		{ id: 'obj-2', title: 'Phase 2: 標準機能', node_type: 'objective', status: 'active', dependencies: ['vision-1'] },
		{ id: 'obj-3', title: 'Phase 3: AI 統合', node_type: 'objective', status: 'draft', dependencies: ['vision-1'] },
		// Activities
		{ id: 'task-1', title: 'CLI 基盤実装', node_type: 'activity', status: 'deprecated', dependencies: ['obj-1'] },
		{ id: 'task-2', title: 'YAML パーサー', node_type: 'activity', status: 'deprecated', dependencies: ['obj-1'] },
		{ id: 'task-3', title: 'UI 実装', node_type: 'activity', status: 'active', dependencies: ['obj-2'] },
		{ id: 'task-4', title: 'API 連携', node_type: 'activity', status: 'draft', dependencies: ['obj-2'] }
	]}
	<div style="height: 700px; background: var(--bg-primary);">
		<FactorioViewer
			graphData={toGraphData(typedNodes)}
			onTaskSelect={handleTaskSelect}
			onTaskHover={handleTaskHover}
		/>
	</div>
</Story>
