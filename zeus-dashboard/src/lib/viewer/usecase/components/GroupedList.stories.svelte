<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import GroupedList from './GroupedList.svelte';
	import type { ActorItem, UseCaseItem } from '$lib/types/api';

	const { Story } = defineMeta({
		title: 'Viewer/UseCase/GroupedList',
		component: GroupedList,
		tags: ['autodocs'],
		argTypes: {
			groupBy: {
				control: 'boolean',
				description: 'グループ化表示'
			}
		}
	});
</script>

<script lang="ts">
	// サンプルデータ
	const sampleActors: ActorItem[] = [
		{ id: 'actor-001', title: 'ユーザー', type: 'human', description: '一般ユーザー' },
		{ id: 'actor-002', title: '管理者', type: 'human', description: 'システム管理者' },
		{ id: 'actor-003', title: '決済システム', type: 'system', description: '外部決済サービス' },
		{ id: 'actor-004', title: 'バッチ処理', type: 'time', description: '定期実行タスク' }
	];

	const sampleUseCases: UseCaseItem[] = [
		{
			id: 'uc-001',
			title: 'ログイン',
			status: 'active',
			actors: [{ actor_id: 'actor-001', role: 'primary' }],
			relations: []
		},
		{
			id: 'uc-002',
			title: '商品を検索する',
			status: 'active',
			actors: [{ actor_id: 'actor-001', role: 'primary' }],
			relations: []
		},
		{
			id: 'uc-003',
			title: '注文を処理する',
			status: 'draft',
			actors: [
				{ actor_id: 'actor-001', role: 'primary' },
				{ actor_id: 'actor-003', role: 'secondary' }
			],
			relations: []
		},
		{
			id: 'uc-004',
			title: 'レポートを生成する',
			status: 'deprecated',
			actors: [{ actor_id: 'actor-004', role: 'primary' }],
			relations: []
		}
	];

	// リストアイテム化
	const actorItems = sampleActors.map((a) => ({ ...a, itemType: 'actor' as const }));
	const useCaseItems = sampleUseCases.map((u) => ({ ...u, itemType: 'usecase' as const }));
	const allItems = [...actorItems, ...useCaseItems];

	let selectedId = $state<string | null>(null);

	function handleSelect(item: { id: string }) {
		selectedId = item.id;
		console.log('Selected:', item);
	}
</script>

<!-- Actor のみ -->
<Story name="ActorsOnly">
	<div style="width: 300px; height: 400px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: hidden;">
		<GroupedList
			items={actorItems}
			groupBy={false}
			selectedId={null}
			actors={sampleActors}
			onSelect={handleSelect}
		/>
	</div>
</Story>

<!-- UseCase のみ -->
<Story name="UseCasesOnly">
	<div style="width: 300px; height: 400px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: hidden;">
		<GroupedList
			items={useCaseItems}
			groupBy={false}
			selectedId={null}
			actors={sampleActors}
			onSelect={handleSelect}
		/>
	</div>
</Story>

<!-- グループ化表示 -->
<Story name="Grouped">
	<div style="width: 300px; height: 500px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: hidden;">
		<GroupedList
			items={allItems}
			groupBy={true}
			selectedId={null}
			actors={sampleActors}
			onSelect={handleSelect}
		/>
	</div>
</Story>

<!-- 選択状態 -->
<Story name="WithSelection">
	<div style="width: 300px; height: 400px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: hidden;">
		<GroupedList
			items={useCaseItems}
			groupBy={false}
			selectedId="uc-002"
			actors={sampleActors}
			onSelect={handleSelect}
		/>
	</div>
</Story>

<!-- 空リスト -->
<Story name="Empty">
	<div style="width: 300px; height: 300px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: hidden;">
		<GroupedList
			items={[]}
			groupBy={false}
			selectedId={null}
			actors={[]}
			onSelect={handleSelect}
		/>
	</div>
</Story>

<!-- インタラクティブ -->
<Story name="Interactive">
	<div style="width: 300px; height: 500px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: hidden;">
		<GroupedList
			items={allItems}
			groupBy={true}
			{selectedId}
			actors={sampleActors}
			onSelect={handleSelect}
		/>
	</div>
	<p style="color: var(--text-muted); font-size: 12px; margin-top: 8px; padding-left: 4px;">
		選択中: {selectedId ?? '（なし）'}
	</p>
</Story>
