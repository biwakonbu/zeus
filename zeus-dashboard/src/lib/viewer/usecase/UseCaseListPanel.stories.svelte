<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import UseCaseListPanel from './UseCaseListPanel.svelte';
	import type { ActorItem, UseCaseItem } from '$lib/types/api';

	const { Story } = defineMeta({
		title: 'Viewer/UseCase/UseCaseListPanel',
		component: UseCaseListPanel,
		tags: ['autodocs'],
		argTypes: {
			selectedActorId: {
				control: 'text',
				description: '選択中の Actor ID'
			},
			selectedUseCaseId: {
				control: 'text',
				description: '選択中の UseCase ID'
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
			id: 'uc-login',
			title: 'ログイン',
			status: 'active',
			actors: [{ actor_id: 'actor-001', role: 'primary' }],
			relations: []
		},
		{
			id: 'uc-search',
			title: '商品を検索する',
			status: 'active',
			actors: [{ actor_id: 'actor-001', role: 'primary' }],
			relations: []
		},
		{
			id: 'uc-order',
			title: '注文を処理する',
			status: 'draft',
			actors: [
				{ actor_id: 'actor-001', role: 'primary' },
				{ actor_id: 'actor-003', role: 'secondary' }
			],
			relations: []
		},
		{
			id: 'uc-report',
			title: 'レポートを生成する',
			status: 'deprecated',
			actors: [{ actor_id: 'actor-004', role: 'primary' }],
			relations: []
		},
		{
			id: 'uc-admin',
			title: 'ユーザー管理',
			status: 'active',
			actors: [{ actor_id: 'actor-002', role: 'primary' }],
			relations: []
		}
	];

	let selectedActorId = $state<string | null>(null);
	let selectedUseCaseId = $state<string | null>(null);

	function handleActorSelect(actor: ActorItem) {
		selectedActorId = actor.id;
		selectedUseCaseId = null;
		console.log('Actor selected:', actor);
	}

	function handleUseCaseSelect(usecase: UseCaseItem) {
		selectedUseCaseId = usecase.id;
		selectedActorId = null;
		console.log('UseCase selected:', usecase);
	}
</script>

<!-- デフォルト -->
<Story name="Default">
	<div
		style="width: 300px; height: 500px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: hidden;"
	>
		<UseCaseListPanel
			actors={sampleActors}
			usecases={sampleUseCases}
			selectedActorId={null}
			selectedUseCaseId={null}
			onActorSelect={handleActorSelect}
			onUseCaseSelect={handleUseCaseSelect}
		/>
	</div>
</Story>

<!-- Actor 選択済み -->
<Story name="ActorSelected">
	<div
		style="width: 300px; height: 500px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: hidden;"
	>
		<UseCaseListPanel
			actors={sampleActors}
			usecases={sampleUseCases}
			selectedActorId="actor-001"
			selectedUseCaseId={null}
			onActorSelect={handleActorSelect}
			onUseCaseSelect={handleUseCaseSelect}
		/>
	</div>
</Story>

<!-- UseCase 選択済み -->
<Story name="UseCaseSelected">
	<div
		style="width: 300px; height: 500px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: hidden;"
	>
		<UseCaseListPanel
			actors={sampleActors}
			usecases={sampleUseCases}
			selectedActorId={null}
			selectedUseCaseId="uc-search"
			onActorSelect={handleActorSelect}
			onUseCaseSelect={handleUseCaseSelect}
		/>
	</div>
</Story>

<!-- 空のリスト -->
<Story name="Empty">
	<div
		style="width: 300px; height: 400px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: hidden;"
	>
		<UseCaseListPanel
			actors={[]}
			usecases={[]}
			selectedActorId={null}
			selectedUseCaseId={null}
			onActorSelect={handleActorSelect}
			onUseCaseSelect={handleUseCaseSelect}
		/>
	</div>
</Story>

<!-- インタラクティブ -->
<Story name="Interactive">
	<div
		style="width: 300px; height: 500px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: hidden;"
	>
		<UseCaseListPanel
			actors={sampleActors}
			usecases={sampleUseCases}
			{selectedActorId}
			{selectedUseCaseId}
			onActorSelect={handleActorSelect}
			onUseCaseSelect={handleUseCaseSelect}
		/>
	</div>
	<p style="color: var(--text-muted); font-size: 12px; margin-top: 8px; padding-left: 4px;">
		選択中: {selectedActorId ?? selectedUseCaseId ?? '（なし）'}
	</p>
</Story>
