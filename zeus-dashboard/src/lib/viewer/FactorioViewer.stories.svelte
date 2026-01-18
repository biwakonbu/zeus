<script context="module" lang="ts">
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
	import type { TaskItem } from '$lib/types/api';

	// Action ハンドラー
	const handleTaskSelect = fn();
	const handleTaskHover = fn();

	// モックタスク（少数）
	const simpleTasks: TaskItem[] = [
		{
			id: 'task-1',
			title: 'プロジェクト設計',
			status: 'completed',
			priority: 'high',
			assignee: 'alice',
			dependencies: [],
			progress: 100
		},
		{
			id: 'task-2',
			title: 'データベース設計',
			status: 'completed',
			priority: 'high',
			assignee: 'bob',
			dependencies: ['task-1'],
			progress: 100
		},
		{
			id: 'task-3',
			title: 'API 実装',
			status: 'in_progress',
			priority: 'high',
			assignee: 'alice',
			dependencies: ['task-2'],
			progress: 60
		},
		{
			id: 'task-4',
			title: 'フロントエンド実装',
			status: 'pending',
			priority: 'medium',
			assignee: 'charlie',
			dependencies: ['task-2'],
			progress: 0
		},
		{
			id: 'task-5',
			title: '統合テスト',
			status: 'blocked',
			priority: 'high',
			assignee: 'bob',
			dependencies: ['task-3', 'task-4'],
			progress: 0
		}
	];

	// より多くのタスク
	const complexTasks: TaskItem[] = [
		// レイヤー1
		{ id: 't1', title: 'プロジェクト立ち上げ', status: 'completed', priority: 'high', assignee: 'alice', dependencies: [], progress: 100 },
		// レイヤー2
		{ id: 't2', title: '要件定義', status: 'completed', priority: 'high', assignee: 'bob', dependencies: ['t1'], progress: 100 },
		{ id: 't3', title: 'チーム編成', status: 'completed', priority: 'medium', assignee: 'charlie', dependencies: ['t1'], progress: 100 },
		// レイヤー3
		{ id: 't4', title: 'アーキテクチャ設計', status: 'completed', priority: 'high', assignee: 'alice', dependencies: ['t2'], progress: 100 },
		{ id: 't5', title: 'UI/UX デザイン', status: 'in_progress', priority: 'medium', assignee: 'charlie', dependencies: ['t2'], progress: 75 },
		{ id: 't6', title: 'インフラ設計', status: 'completed', priority: 'medium', assignee: 'bob', dependencies: ['t2', 't3'], progress: 100 },
		// レイヤー4
		{ id: 't7', title: 'バックエンド開発', status: 'in_progress', priority: 'high', assignee: 'alice', dependencies: ['t4'], progress: 45 },
		{ id: 't8', title: 'フロントエンド開発', status: 'pending', priority: 'high', assignee: 'charlie', dependencies: ['t4', 't5'], progress: 0 },
		{ id: 't9', title: 'CI/CD 構築', status: 'in_progress', priority: 'medium', assignee: 'bob', dependencies: ['t6'], progress: 80 },
		// レイヤー5
		{ id: 't10', title: 'API 統合', status: 'pending', priority: 'high', assignee: 'alice', dependencies: ['t7', 't8'], progress: 0 },
		{ id: 't11', title: 'パフォーマンス最適化', status: 'blocked', priority: 'medium', assignee: 'bob', dependencies: ['t7'], progress: 0 },
		// レイヤー6
		{ id: 't12', title: '結合テスト', status: 'pending', priority: 'high', assignee: 'bob', dependencies: ['t10', 't9'], progress: 0 },
		{ id: 't13', title: 'セキュリティ監査', status: 'pending', priority: 'high', assignee: 'charlie', dependencies: ['t10'], progress: 0 },
		// レイヤー7
		{ id: 't14', title: 'ステージングデプロイ', status: 'pending', priority: 'medium', assignee: 'bob', dependencies: ['t12', 't13'], progress: 0 },
		// レイヤー8
		{ id: 't15', title: '本番リリース', status: 'pending', priority: 'high', assignee: 'alice', dependencies: ['t14'], progress: 0 }
	];

	// 空のタスク
	const emptyTasks: TaskItem[] = [];

	// 選択中のタスクID
	let selectedTaskId: string | null = $state(null);

	function handleInteractiveSelect(taskId: string | null) {
		selectedTaskId = taskId;
		handleTaskSelect(taskId);
	}
</script>

<!-- デフォルト（シンプルなタスク） -->
<Story name="Default">
	<div style="height: 600px; background: var(--bg-primary);">
		<FactorioViewer
			tasks={simpleTasks}
			onTaskSelect={handleTaskSelect}
			onTaskHover={handleTaskHover}
		/>
	</div>
</Story>

<!-- 複雑なタスクグラフ -->
<Story name="ComplexGraph">
	<div style="height: 700px; background: var(--bg-primary);">
		<FactorioViewer
			tasks={complexTasks}
			onTaskSelect={handleTaskSelect}
			onTaskHover={handleTaskHover}
		/>
	</div>
</Story>

<!-- タスクなし -->
<Story name="Empty">
	<div style="height: 400px; background: var(--bg-primary);">
		<FactorioViewer
			tasks={emptyTasks}
			onTaskSelect={handleTaskSelect}
			onTaskHover={handleTaskHover}
		/>
	</div>
</Story>

<!-- タスク選択済み -->
<Story name="WithSelection">
	<div style="height: 600px; background: var(--bg-primary);">
		<FactorioViewer
			tasks={simpleTasks}
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
			tasks={complexTasks}
			{selectedTaskId}
			onTaskSelect={handleInteractiveSelect}
			onTaskHover={handleTaskHover}
		/>
		<div style="position: absolute; top: 60px; right: 60px; background: var(--bg-panel); padding: 12px; border-radius: 4px; border: 1px solid var(--border-metal);">
			<p style="color: var(--text-secondary); font-size: 11px; margin: 0 0 4px 0;">選択中のタスク:</p>
			<p style="color: var(--accent-primary); font-size: 12px; margin: 0;">
				{selectedTaskId || 'なし'}
			</p>
		</div>
	</div>
</Story>

<!-- 全ステータスのタスク -->
<Story name="AllStatuses">
	{@const allStatusTasks = [
		{ id: 'completed-1', title: '完了タスク 1', status: 'completed' as const, priority: 'high' as const, assignee: 'alice', dependencies: [], progress: 100 },
		{ id: 'completed-2', title: '完了タスク 2', status: 'completed' as const, priority: 'medium' as const, assignee: 'bob', dependencies: ['completed-1'], progress: 100 },
		{ id: 'in_progress-1', title: '進行中タスク', status: 'in_progress' as const, priority: 'high' as const, assignee: 'charlie', dependencies: ['completed-2'], progress: 50 },
		{ id: 'pending-1', title: '未着手タスク', status: 'pending' as const, priority: 'medium' as const, assignee: 'alice', dependencies: ['completed-2'], progress: 0 },
		{ id: 'blocked-1', title: 'ブロック中タスク', status: 'blocked' as const, priority: 'high' as const, assignee: 'bob', dependencies: ['in_progress-1', 'pending-1'], progress: 0 }
	]}
	<div style="height: 500px; background: var(--bg-primary);">
		<FactorioViewer
			tasks={allStatusTasks}
			onTaskSelect={handleTaskSelect}
			onTaskHover={handleTaskHover}
		/>
	</div>
</Story>

