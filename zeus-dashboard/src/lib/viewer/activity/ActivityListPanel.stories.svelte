<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ActivityListPanel from './ActivityListPanel.svelte';
	import type { ActivityItem } from '$lib/types/api';

	const { Story } = defineMeta({
		title: 'Viewer/Activity/ActivityListPanel',
		component: ActivityListPanel,
		tags: ['autodocs'],
		argTypes: {
			selectedActivityId: {
				control: 'text',
				description: '選択中の Activity ID'
			}
		}
	});
</script>

<script lang="ts">
	// サンプルデータ
	const sampleActivities: ActivityItem[] = [
		{
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
				{ id: 't1-1', source: 'n1', target: 'n2' },
				{ id: 't1-2', source: 'n2', target: 'n3' },
				{ id: 't1-3', source: 'n3', target: 'n4', guard: 'Yes' },
				{ id: 't1-4', source: 'n3', target: 'n5', guard: 'No' },
				{ id: 't1-5', source: 'n4', target: 'n6' },
				{ id: 't1-6', source: 'n5', target: 'n2' }
			],
			created_at: '2024-01-15T10:00:00Z',
			updated_at: '2024-01-15T10:00:00Z'
		},
		{
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
				{ id: 'n8', type: 'action', name: '注文確定' },
				{ id: 'n9', type: 'final', name: '' }
			],
			transitions: [
				{ id: 't2-1', source: 'n1', target: 'n2' },
				{ id: 't2-2', source: 'n2', target: 'n3' },
				{ id: 't2-3', source: 'n3', target: 'n4' },
				{ id: 't2-4', source: 'n4', target: 'n5' },
				{ id: 't2-5', source: 'n4', target: 'n6' },
				{ id: 't2-6', source: 'n5', target: 'n7' },
				{ id: 't2-7', source: 'n6', target: 'n7' },
				{ id: 't2-8', source: 'n7', target: 'n8' },
				{ id: 't2-9', source: 'n8', target: 'n9' }
			],
			created_at: '2024-01-16T09:00:00Z',
			updated_at: '2024-01-16T09:00:00Z'
		},
		{
			id: 'act-003',
			title: 'レポート生成フロー',
			status: 'draft',
			description: '各種レポートの生成処理',
			nodes: [
				{ id: 'n1', type: 'initial', name: '' },
				{ id: 'n2', type: 'action', name: 'データ取得' },
				{ id: 'n3', type: 'action', name: 'レポート生成' },
				{ id: 'n4', type: 'final', name: '' }
			],
			transitions: [
				{ id: 't3-1', source: 'n1', target: 'n2' },
				{ id: 't3-2', source: 'n2', target: 'n3' },
				{ id: 't3-3', source: 'n3', target: 'n4' }
			],
			created_at: '2024-01-17T11:00:00Z',
			updated_at: '2024-01-17T11:00:00Z'
		},
		{
			id: 'act-004',
			title: '旧決済フロー',
			status: 'deprecated',
			description: '廃止予定の決済処理',
			nodes: [
				{ id: 'n1', type: 'initial', name: '' },
				{ id: 'n2', type: 'action', name: '決済' },
				{ id: 'n3', type: 'final', name: '' }
			],
			transitions: [
				{ id: 't4-1', source: 'n1', target: 'n2' },
				{ id: 't4-2', source: 'n2', target: 'n3' }
			],
			created_at: '2024-01-10T08:00:00Z',
			updated_at: '2024-01-10T08:00:00Z'
		}
	];

	let selectedActivityId = $state<string | null>(null);

	function handleActivitySelect(activity: ActivityItem) {
		selectedActivityId = activity.id;
		console.log('Activity selected:', activity);
	}
</script>

<!-- デフォルト -->
<Story name="Default">
	<div
		style="width: 300px; height: 500px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: hidden;"
	>
		<ActivityListPanel
			activities={sampleActivities}
			selectedActivityId={null}
			onActivitySelect={handleActivitySelect}
		/>
	</div>
</Story>

<!-- 選択済み -->
<Story name="WithSelection">
	<div
		style="width: 300px; height: 500px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: hidden;"
	>
		<ActivityListPanel
			activities={sampleActivities}
			selectedActivityId="act-002"
			onActivitySelect={handleActivitySelect}
		/>
	</div>
</Story>

<!-- 空のリスト -->
<Story name="Empty">
	<div
		style="width: 300px; height: 400px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: hidden;"
	>
		<ActivityListPanel
			activities={[]}
			selectedActivityId={null}
			onActivitySelect={handleActivitySelect}
		/>
	</div>
</Story>

<!-- 単一アイテム -->
<Story name="SingleItem">
	<div
		style="width: 300px; height: 400px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: hidden;"
	>
		<ActivityListPanel
			activities={[sampleActivities[0]]}
			selectedActivityId={null}
			onActivitySelect={handleActivitySelect}
		/>
	</div>
</Story>

<!-- インタラクティブ -->
<Story name="Interactive">
	<div
		style="width: 300px; height: 500px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: hidden;"
	>
		<ActivityListPanel
			activities={sampleActivities}
			{selectedActivityId}
			onActivitySelect={handleActivitySelect}
		/>
	</div>
	<p style="color: var(--text-muted); font-size: 12px; margin-top: 8px; padding-left: 4px;">
		選択中: {selectedActivityId ?? '（なし）'}
	</p>
</Story>
