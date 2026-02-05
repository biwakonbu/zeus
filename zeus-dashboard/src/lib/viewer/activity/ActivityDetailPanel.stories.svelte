<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ActivityDetailPanel from './ActivityDetailPanel.svelte';
	import type { ActivityItem, ActivityNodeItem } from '$lib/types/api';

	const { Story } = defineMeta({
		title: 'Viewer/Activity/ActivityDetailPanel',
		component: ActivityDetailPanel,
		tags: ['autodocs'],
		argTypes: {}
	});
</script>

<script lang="ts">
	// サンプルデータ
	const sampleActivity: ActivityItem = {
		id: 'act-001',
		title: 'ログイン処理フロー',
		status: 'active',
		description: 'ユーザー認証の一連の処理',
		usecase_id: 'uc-001',
		nodes: [
			{ id: 'n1', type: 'initial', name: '' },
			{ id: 'n2', type: 'action', name: '認証情報入力' },
			{ id: 'n3', type: 'decision', name: '認証成功?' },
			{ id: 'n4', type: 'action', name: 'ダッシュボード表示' },
			{ id: 'n5', type: 'action', name: 'エラー表示' },
			{ id: 'n6', type: 'final', name: '' }
		],
		transitions: [
			{ id: 't1', source: 'n1', target: 'n2' },
			{ id: 't2', source: 'n2', target: 'n3' },
			{ id: 't3', source: 'n3', target: 'n4', guard: 'Yes' },
			{ id: 't4', source: 'n3', target: 'n5', guard: 'No' },
			{ id: 't5', source: 'n4', target: 'n6' },
			{ id: 't6', source: 'n5', target: 'n2' }
		],
		created_at: '2024-01-15T10:00:00Z',
		updated_at: '2024-01-15T10:00:00Z'
	};

	const complexActivity: ActivityItem = {
		id: 'act-002',
		title: '注文処理フロー',
		status: 'active',
		description: '商品の注文から完了までの流れ',
		usecase_id: 'uc-002',
		nodes: [
			{ id: 'n1', type: 'initial', name: '' },
			{ id: 'n2', type: 'action', name: 'カート確認' },
			{ id: 'n3', type: 'action', name: '配送先選択' },
			{ id: 'n4', type: 'fork', name: '' },
			{ id: 'n5', type: 'action', name: '在庫確認' },
			{ id: 'n6', type: 'action', name: '決済処理' },
			{ id: 'n7', type: 'join', name: '' },
			{ id: 'n8', type: 'decision', name: '在庫あり?' },
			{ id: 'n9', type: 'action', name: '注文確定' },
			{ id: 'n10', type: 'action', name: '在庫切れ通知' },
			{ id: 'n11', type: 'merge', name: '' },
			{ id: 'n12', type: 'final', name: '' }
		],
		transitions: [
			{ id: 't1', source: 'n1', target: 'n2' },
			{ id: 't2', source: 'n2', target: 'n3' },
			{ id: 't3', source: 'n3', target: 'n4' },
			{ id: 't4', source: 'n4', target: 'n5' },
			{ id: 't5', source: 'n4', target: 'n6' },
			{ id: 't6', source: 'n5', target: 'n7' },
			{ id: 't7', source: 'n6', target: 'n7' },
			{ id: 't8', source: 'n7', target: 'n8' },
			{ id: 't9', source: 'n8', target: 'n9', guard: 'Yes' },
			{ id: 't10', source: 'n8', target: 'n10', guard: 'No' },
			{ id: 't11', source: 'n9', target: 'n11' },
			{ id: 't12', source: 'n10', target: 'n11' },
			{ id: 't13', source: 'n11', target: 'n12' }
		],
		created_at: '2024-01-16T09:00:00Z',
		updated_at: '2024-01-16T14:30:00Z'
	};

	const draftActivity: ActivityItem = {
		id: 'act-003',
		title: 'レポート生成フロー',
		status: 'draft',
		description: '各種レポートの生成処理（作成中）',
		nodes: [
			{ id: 'n1', type: 'initial', name: '' },
			{ id: 'n2', type: 'action', name: 'データ取得' },
			{ id: 'n3', type: 'final', name: '' }
		],
		transitions: [
			{ id: 't1', source: 'n1', target: 'n2' },
			{ id: 't2', source: 'n2', target: 'n3' }
		],
		created_at: '2024-01-17T11:00:00Z',
		updated_at: '2024-01-17T11:00:00Z'
	};

	let selectedNode = $state<ActivityNodeItem | null>(null);

	function handleClose() {
		console.log('Panel closed');
	}

	function handleNodeClick(node: ActivityNodeItem) {
		selectedNode = node;
		console.log('Node clicked:', node);
	}
</script>

<!-- アクティビティ情報のみ -->
<Story name="ActivityInfo">
	<div
		style="width: 360px; height: 500px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: auto; padding: 12px;"
	>
		<ActivityDetailPanel
			activity={sampleActivity}
			selectedNode={null}
			onClose={handleClose}
			onNodeClick={handleNodeClick}
		/>
	</div>
</Story>

<!-- ノード選択済み（アクション） -->
<Story name="ActionNodeSelected">
	<div
		style="width: 360px; height: 600px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: auto; padding: 12px;"
	>
		<ActivityDetailPanel
			activity={sampleActivity}
			selectedNode={sampleActivity.nodes[1]}
			onClose={handleClose}
			onNodeClick={handleNodeClick}
		/>
	</div>
</Story>

<!-- ノード選択済み（分岐） -->
<Story name="DecisionNodeSelected">
	<div
		style="width: 360px; height: 600px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: auto; padding: 12px;"
	>
		<ActivityDetailPanel
			activity={sampleActivity}
			selectedNode={sampleActivity.nodes[2]}
			onClose={handleClose}
			onNodeClick={handleNodeClick}
		/>
	</div>
</Story>

<!-- 複雑なアクティビティ -->
<Story name="ComplexActivity">
	<div
		style="width: 360px; height: 700px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: auto; padding: 12px;"
	>
		<ActivityDetailPanel
			activity={complexActivity}
			selectedNode={null}
			onClose={handleClose}
			onNodeClick={handleNodeClick}
		/>
	</div>
</Story>

<!-- 複雑なアクティビティ（Fork 選択） -->
<Story name="ForkNodeSelected">
	<div
		style="width: 360px; height: 700px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: auto; padding: 12px;"
	>
		<ActivityDetailPanel
			activity={complexActivity}
			selectedNode={complexActivity.nodes[3]}
			onClose={handleClose}
			onNodeClick={handleNodeClick}
		/>
	</div>
</Story>

<!-- 下書き状態のアクティビティ -->
<Story name="DraftActivity">
	<div
		style="width: 360px; height: 500px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: auto; padding: 12px;"
	>
		<ActivityDetailPanel
			activity={draftActivity}
			selectedNode={null}
			onClose={handleClose}
			onNodeClick={handleNodeClick}
		/>
	</div>
</Story>

<!-- 未選択状態 -->
<Story name="Empty">
	<div
		style="width: 360px; height: 300px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: auto; padding: 12px;"
	>
		<ActivityDetailPanel
			activity={null}
			selectedNode={null}
			onClose={handleClose}
			onNodeClick={handleNodeClick}
		/>
	</div>
</Story>

<!-- インタラクティブ -->
<Story name="Interactive">
	<div
		style="width: 360px; height: 600px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: auto; padding: 12px;"
	>
		<ActivityDetailPanel
			activity={sampleActivity}
			{selectedNode}
			onClose={handleClose}
			onNodeClick={handleNodeClick}
		/>
	</div>
	<p style="color: var(--text-muted); font-size: 12px; margin-top: 8px; padding-left: 4px;">
		選択ノード: {selectedNode?.name || selectedNode?.type || '（なし）'}
	</p>
</Story>
