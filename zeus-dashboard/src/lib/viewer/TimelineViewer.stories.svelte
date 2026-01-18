<script context="module" lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import TimelineViewer from './TimelineViewer.svelte';

	const { Story } = defineMeta({
		title: 'Viewer/TimelineViewer',
		component: TimelineViewer,
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
	import type { TimelineItem } from '$lib/types/api';

	// Action ハンドラー
	const handleTaskSelect = fn();

	// 選択中のタスク
	let selectedTask: TimelineItem | null = $state(null);

	function handleInteractiveSelect(task: TimelineItem | null) {
		selectedTask = task;
		handleTaskSelect(task);
	}
</script>

<!-- デフォルト（MSW でモックデータを返す） -->
<Story name="Default">
	<div style="height: 600px; background: var(--bg-primary);">
		<TimelineViewer onTaskSelect={handleTaskSelect} />
	</div>
</Story>

<!-- インタラクティブ -->
<Story name="Interactive">
	<div style="height: 650px; background: var(--bg-primary); position: relative;">
		<TimelineViewer onTaskSelect={handleInteractiveSelect} />
		{#if selectedTask}
			<div style="position: fixed; top: 20px; right: 20px; background: var(--bg-panel); padding: 16px; border-radius: 8px; border: 2px solid var(--border-metal); max-width: 300px; z-index: 100;">
				<h4 style="color: var(--accent-primary); margin: 0 0 8px 0; font-size: 14px;">選択中のタスク</h4>
				<div style="color: var(--text-secondary); font-size: 12px;">
					<p style="margin: 4px 0;"><strong>タイトル:</strong> {selectedTask.title}</p>
					<p style="margin: 4px 0;"><strong>期間:</strong> {selectedTask.start_date} - {selectedTask.end_date}</p>
					<p style="margin: 4px 0;"><strong>進捗:</strong> {selectedTask.progress}%</p>
					<p style="margin: 4px 0;"><strong>スラック:</strong> {selectedTask.slack}日</p>
					{#if selectedTask.is_on_critical_path}
						<p style="margin: 4px 0; color: #ef4444;"><strong>クリティカルパス上</strong></p>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</Story>

<!-- フルスクリーン -->
<Story name="Fullscreen">
	<div style="height: 100vh; background: var(--bg-primary);">
		<TimelineViewer onTaskSelect={handleTaskSelect} />
	</div>
</Story>

<!-- 読み込み中状態 -->
<Story
	name="Loading"
	parameters={{
		msw: {
			handlers: [
				{
					url: '/api/timeline',
					method: 'get',
					status: 200,
					delay: 'infinite'
				}
			]
		}
	}}
>
	<div style="height: 500px; background: var(--bg-primary);">
		<TimelineViewer onTaskSelect={handleTaskSelect} />
	</div>
</Story>

<!-- エラー状態 -->
<Story
	name="Error"
	parameters={{
		msw: {
			handlers: [
				{
					url: '/api/timeline',
					method: 'get',
					status: 500,
					response: { error: 'サーバーエラーが発生しました' }
				}
			]
		}
	}}
>
	<div style="height: 500px; background: var(--bg-primary);">
		<TimelineViewer onTaskSelect={handleTaskSelect} />
	</div>
</Story>

<!-- 空状態 -->
<Story
	name="Empty"
	parameters={{
		msw: {
			handlers: [
				{
					url: '/api/timeline',
					method: 'get',
					status: 200,
					response: {
						items: [],
						critical_path: [],
						project_start: '2026-01-01',
						project_end: '2026-01-01',
						total_duration: 0,
						stats: {
							total_tasks: 0,
							tasks_with_dates: 0,
							on_critical_path: 0,
							average_slack: 0,
							overdue_tasks: 0,
							completed_on_time: 0
						}
					}
				}
			]
		}
	}}
>
	<div style="height: 400px; background: var(--bg-primary);">
		<TimelineViewer onTaskSelect={handleTaskSelect} />
	</div>
</Story>

<!-- クリティカルパス表示 -->
<Story
	name="CriticalPath"
	parameters={{
		msw: {
			handlers: [
				{
					url: '/api/timeline',
					method: 'get',
					status: 200,
					response: {
						items: [
							{
								task_id: 'task-1',
								title: 'プロジェクト設計',
								start_date: '2026-01-01',
								end_date: '2026-01-10',
								progress: 100,
								status: 'completed',
								priority: 'high',
								assignee: 'alice',
								is_on_critical_path: true,
								slack: 0,
								dependencies: []
							},
							{
								task_id: 'task-2',
								title: 'データベース設計',
								start_date: '2026-01-11',
								end_date: '2026-01-20',
								progress: 100,
								status: 'completed',
								priority: 'high',
								assignee: 'bob',
								is_on_critical_path: true,
								slack: 0,
								dependencies: ['task-1']
							},
							{
								task_id: 'task-3',
								title: 'API 実装',
								start_date: '2026-01-21',
								end_date: '2026-02-10',
								progress: 60,
								status: 'in_progress',
								priority: 'high',
								assignee: 'alice',
								is_on_critical_path: true,
								slack: 0,
								dependencies: ['task-2']
							},
							{
								task_id: 'task-4',
								title: 'ドキュメント作成',
								start_date: '2026-01-11',
								end_date: '2026-01-25',
								progress: 80,
								status: 'in_progress',
								priority: 'low',
								assignee: 'charlie',
								is_on_critical_path: false,
								slack: 16,
								dependencies: ['task-1']
							},
							{
								task_id: 'task-5',
								title: '統合テスト',
								start_date: '2026-02-11',
								end_date: '2026-02-25',
								progress: 0,
								status: 'pending',
								priority: 'high',
								assignee: 'bob',
								is_on_critical_path: true,
								slack: 0,
								dependencies: ['task-3']
							},
							{
								task_id: 'task-6',
								title: 'リリース準備',
								start_date: '2026-02-26',
								end_date: '2026-02-28',
								progress: 0,
								status: 'pending',
								priority: 'high',
								assignee: 'alice',
								is_on_critical_path: true,
								slack: 0,
								dependencies: ['task-5']
							}
						],
						critical_path: ['task-1', 'task-2', 'task-3', 'task-5', 'task-6'],
						project_start: '2026-01-01',
						project_end: '2026-02-28',
						total_duration: 59,
						stats: {
							total_tasks: 6,
							tasks_with_dates: 6,
							on_critical_path: 5,
							average_slack: 2.67,
							overdue_tasks: 0,
							completed_on_time: 2
						}
					}
				}
			]
		}
	}}
>
	<div style="height: 600px; background: var(--bg-primary);">
		<TimelineViewer onTaskSelect={handleTaskSelect} />
	</div>
</Story>

<!-- 遅延タスク -->
<Story
	name="OverdueTasks"
	parameters={{
		msw: {
			handlers: [
				{
					url: '/api/timeline',
					method: 'get',
					status: 200,
					response: {
						items: [
							{
								task_id: 'task-1',
								title: '要件定義',
								start_date: '2025-12-01',
								end_date: '2025-12-15',
								progress: 100,
								status: 'completed',
								priority: 'high',
								assignee: 'alice',
								is_on_critical_path: true,
								slack: 0,
								dependencies: []
							},
							{
								task_id: 'task-2',
								title: '設計書作成（遅延中）',
								start_date: '2025-12-16',
								end_date: '2026-01-05',
								progress: 70,
								status: 'in_progress',
								priority: 'high',
								assignee: 'bob',
								is_on_critical_path: true,
								slack: -14,
								dependencies: ['task-1']
							},
							{
								task_id: 'task-3',
								title: 'コードレビュー（大幅遅延）',
								start_date: '2025-12-20',
								end_date: '2025-12-28',
								progress: 30,
								status: 'in_progress',
								priority: 'medium',
								assignee: 'charlie',
								is_on_critical_path: false,
								slack: -22,
								dependencies: ['task-1']
							},
							{
								task_id: 'task-4',
								title: 'テスト実装（ブロック中）',
								start_date: '2026-01-06',
								end_date: '2026-01-20',
								progress: 0,
								status: 'blocked',
								priority: 'high',
								assignee: 'alice',
								is_on_critical_path: true,
								slack: -14,
								dependencies: ['task-2']
							},
							{
								task_id: 'task-5',
								title: 'デプロイ準備',
								start_date: '2026-01-21',
								end_date: '2026-01-25',
								progress: 0,
								status: 'pending',
								priority: 'medium',
								assignee: 'bob',
								is_on_critical_path: true,
								slack: -14,
								dependencies: ['task-4']
							}
						],
						critical_path: ['task-1', 'task-2', 'task-4', 'task-5'],
						project_start: '2025-12-01',
						project_end: '2026-01-25',
						total_duration: 56,
						stats: {
							total_tasks: 5,
							tasks_with_dates: 5,
							on_critical_path: 4,
							average_slack: -12.8,
							overdue_tasks: 4,
							completed_on_time: 1
						}
					}
				}
			]
		}
	}}
>
	<div style="height: 600px; background: var(--bg-primary);">
		<TimelineViewer onTaskSelect={handleTaskSelect} />
	</div>
</Story>
