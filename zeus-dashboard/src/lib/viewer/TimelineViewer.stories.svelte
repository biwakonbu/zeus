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

