<script context="module" lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import WBSViewer from './WBSViewer.svelte';

	const { Story } = defineMeta({
		title: 'Viewer/WBSViewer',
		component: WBSViewer,
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
	import type { WBSNode } from '$lib/types/api';

	// Action ハンドラー
	const handleNodeSelect = fn();

	// 選択中のノード
	let selectedNode: WBSNode | null = $state(null);

	function handleInteractiveSelect(node: WBSNode | null) {
		selectedNode = node;
		handleNodeSelect(node);
	}
</script>

<!-- デフォルト（MSW でモックデータを返す） -->
<Story name="Default">
	<div style="height: 600px; background: var(--bg-primary);">
		<WBSViewer onNodeSelect={handleNodeSelect} />
	</div>
</Story>

<!-- インタラクティブ -->
<Story name="Interactive">
	<div style="height: 650px; background: var(--bg-primary); position: relative;">
		<WBSViewer onNodeSelect={handleInteractiveSelect} />
		{#if selectedNode}
			<div style="position: fixed; bottom: 20px; right: 20px; background: var(--bg-panel); padding: 16px; border-radius: 8px; border: 2px solid var(--border-metal); max-width: 300px; z-index: 100;">
				<h4 style="color: var(--accent-primary); margin: 0 0 8px 0; font-size: 14px;">選択中のノード</h4>
				<div style="color: var(--text-secondary); font-size: 12px;">
					<p style="margin: 4px 0;"><strong>ID:</strong> {selectedNode.id}</p>
					<p style="margin: 4px 0;"><strong>タイトル:</strong> {selectedNode.title}</p>
					<p style="margin: 4px 0;"><strong>WBS:</strong> {selectedNode.wbs_code}</p>
					<p style="margin: 4px 0;"><strong>進捗:</strong> {selectedNode.progress}%</p>
					<p style="margin: 4px 0;"><strong>ステータス:</strong> {selectedNode.status}</p>
				</div>
			</div>
		{/if}
	</div>
</Story>

<!-- フルスクリーン -->
<Story name="Fullscreen">
	<div style="height: 100vh; background: var(--bg-primary);">
		<WBSViewer onNodeSelect={handleNodeSelect} />
	</div>
</Story>

