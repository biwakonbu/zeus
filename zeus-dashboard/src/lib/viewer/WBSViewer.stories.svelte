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

	// Action ハンドラー
	const handleNodeSelect = fn();

	// 選択中のノード情報
	let selectedNodeId: string | null = $state(null);
	let selectedNodeType: string | null = $state(null);

	function handleInteractiveSelect(nodeId: string, nodeType: string) {
		selectedNodeId = nodeId;
		selectedNodeType = nodeType;
		handleNodeSelect(nodeId, nodeType);
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
		{#if selectedNodeId}
			<div style="position: fixed; bottom: 20px; right: 20px; background: var(--bg-panel); padding: 16px; border-radius: 8px; border: 2px solid var(--border-metal); max-width: 300px; z-index: 100;">
				<h4 style="color: var(--accent-primary); margin: 0 0 8px 0; font-size: 14px;">選択中のノード</h4>
				<div style="color: var(--text-secondary); font-size: 12px;">
					<p style="margin: 4px 0;"><strong>ID:</strong> {selectedNodeId}</p>
					<p style="margin: 4px 0;"><strong>タイプ:</strong> {selectedNodeType}</p>
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

<!-- 読み込み中状態 -->
<Story
	name="Loading"
	parameters={{
		msw: {
			handlers: [
				{
					url: '/api/wbs',
					method: 'get',
					status: 200,
					delay: 'infinite'
				}
			]
		}
	}}
>
	<div style="height: 500px; background: var(--bg-primary);">
		<WBSViewer onNodeSelect={handleNodeSelect} />
	</div>
</Story>

<!-- エラー状態 -->
<Story
	name="Error"
	parameters={{
		msw: {
			handlers: [
				{
					url: '/api/wbs',
					method: 'get',
					status: 500,
					response: { error: 'サーバーエラーが発生しました' }
				}
			]
		}
	}}
>
	<div style="height: 500px; background: var(--bg-primary);">
		<WBSViewer onNodeSelect={handleNodeSelect} />
	</div>
</Story>

<!-- 空状態 -->
<Story
	name="Empty"
	parameters={{
		msw: {
			handlers: [
				{
					url: '/api/wbs',
					method: 'get',
					status: 200,
					response: {
						roots: [],
						max_depth: 0,
						stats: {
							total_nodes: 0,
							root_count: 0,
							leaf_count: 0,
							max_depth: 0,
							avg_progress: 0,
							completed_pct: 0
						}
					}
				}
			]
		}
	}}
>
	<div style="height: 400px; background: var(--bg-primary);">
		<WBSViewer onNodeSelect={handleNodeSelect} />
	</div>
</Story>

<!-- 深い階層（5+ レベル） -->
<Story
	name="DeepHierarchy"
	parameters={{
		msw: {
			handlers: [
				{
					url: '/api/wbs',
					method: 'get',
					status: 200,
					response: {
						roots: [
							{
								id: 'root',
								title: 'プロジェクトルート',
								wbs_code: '1',
								status: 'in_progress',
								progress: 45,
								priority: 'high',
								assignee: 'alice',
								depth: 0,
								children: [
									{
										id: 'l1-1',
										title: 'フェーズ 1',
										wbs_code: '1.1',
										status: 'completed',
										progress: 100,
										priority: 'high',
										assignee: 'bob',
										depth: 1,
										children: [
											{
												id: 'l2-1',
												title: '設計',
												wbs_code: '1.1.1',
												status: 'completed',
												progress: 100,
												priority: 'high',
												assignee: 'alice',
												depth: 2,
												children: [
													{
														id: 'l3-1',
														title: 'アーキテクチャ設計',
														wbs_code: '1.1.1.1',
														status: 'completed',
														progress: 100,
														priority: 'high',
														assignee: 'alice',
														depth: 3,
														children: [
															{
																id: 'l4-1',
																title: 'コンポーネント設計',
																wbs_code: '1.1.1.1.1',
																status: 'completed',
																progress: 100,
																priority: 'medium',
																assignee: 'charlie',
																depth: 4,
																children: [
																	{
																		id: 'l5-1',
																		title: 'インターフェース定義',
																		wbs_code: '1.1.1.1.1.1',
																		status: 'completed',
																		progress: 100,
																		priority: 'medium',
																		assignee: 'charlie',
																		depth: 5
																	}
																]
															}
														]
													}
												]
											}
										]
									},
									{
										id: 'l1-2',
										title: 'フェーズ 2',
										wbs_code: '1.2',
										status: 'in_progress',
										progress: 30,
										priority: 'high',
										assignee: 'bob',
										depth: 1,
										children: [
											{
												id: 'l2-2',
												title: '実装',
												wbs_code: '1.2.1',
												status: 'in_progress',
												progress: 30,
												priority: 'high',
												assignee: 'alice',
												depth: 2
											}
										]
									}
								]
							}
						],
						max_depth: 6,
						stats: {
							total_nodes: 10,
							root_count: 1,
							leaf_count: 3,
							max_depth: 6,
							avg_progress: 65,
							completed_pct: 70
						}
					}
				}
			]
		}
	}}
>
	<div style="height: 700px; background: var(--bg-primary);">
		<WBSViewer onNodeSelect={handleNodeSelect} />
	</div>
</Story>

