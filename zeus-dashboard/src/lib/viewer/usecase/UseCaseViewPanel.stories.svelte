<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import UseCaseViewPanel from './UseCaseViewPanel.svelte';
	import type { ActorItem, UseCaseItem, ActivityItem } from '$lib/types/api';

	const { Story } = defineMeta({
		title: 'Viewer/UseCase/UseCaseViewPanel',
		component: UseCaseViewPanel,
		tags: ['autodocs'],
		argTypes: {}
	});
</script>

<script lang="ts">
	// サンプルデータ
	const sampleActors: ActorItem[] = [
		{ id: 'actor-001', title: 'ユーザー', type: 'human', description: '一般ユーザー' },
		{ id: 'actor-002', title: '管理者', type: 'human', description: 'システム管理者' },
		{ id: 'actor-003', title: '決済システム', type: 'system', description: '外部決済サービス' }
	];

	const sampleUseCases: UseCaseItem[] = [
		{
			id: 'uc-001',
			title: 'ログイン',
			status: 'active',
			description: 'ユーザーがシステムにログインする',
			objective_id: 'obj-001',
			actors: [{ actor_id: 'actor-001', role: 'primary' }],
			relations: []
		},
		{
			id: 'uc-002',
			title: '商品を検索する',
			status: 'active',
			actors: [{ actor_id: 'actor-001', role: 'primary' }],
			relations: [{ type: 'include', target_id: 'uc-001' }]
		}
	];

	const useCaseWithScenario: UseCaseItem = {
		id: 'uc-003',
		title: '注文を処理する',
		status: 'active',
		description: 'ユーザーが商品を注文するプロセス',
		objective_id: 'obj-002',
		actors: [
			{ actor_id: 'actor-001', role: 'primary' },
			{ actor_id: 'actor-003', role: 'secondary' }
		],
		relations: [
			{ type: 'include', target_id: 'uc-001' },
			{ type: 'extend', target_id: 'uc-002', condition: 'クーポン適用時' }
		],
		scenario: {
			preconditions: ['ユーザーがログイン済みであること', 'カートに商品が入っていること'],
			trigger: 'ユーザーが「注文する」ボタンをクリック',
			main_flow: [
				'システムは注文内容を確認画面に表示する',
				'ユーザーは配送先を選択する',
				'ユーザーは支払い方法を選択する',
				'システムは決済処理を実行する',
				'システムは注文完了画面を表示する'
			],
			alternative_flows: [
				{
					id: 'A1',
					name: '新規配送先追加',
					condition: 'ユーザーが新しい配送先を追加したい場合',
					steps: [
						'ユーザーは「新規住所を追加」を選択',
						'システムは住所入力フォームを表示',
						'ユーザーは住所情報を入力して保存'
					],
					rejoins_at: 'メインフロー ステップ2'
				}
			],
			exception_flows: [
				{
					id: 'E1',
					name: '決済失敗',
					trigger: '決済処理が失敗した場合',
					steps: [
						'システムはエラーメッセージを表示',
						'ユーザーは別の支払い方法を選択するか、キャンセルする'
					],
					outcome: 'メインフロー ステップ3 に戻る、またはキャンセル'
				}
			],
			postconditions: ['注文が確定している', '在庫が減少している', '確認メールが送信されている']
		}
	};

	const sampleActivities: ActivityItem[] = [
		{
			id: 'act-001',
			title: '注文処理フロー',
			status: 'active',
			usecase_id: 'uc-003',
			nodes: [],
			transitions: [],
			created_at: '2024-01-15T10:00:00Z',
			updated_at: '2024-01-15T10:00:00Z'
		},
		{
			id: 'act-002',
			title: '決済処理フロー',
			status: 'active',
			usecase_id: 'uc-003',
			nodes: [],
			transitions: [],
			created_at: '2024-01-16T09:00:00Z',
			updated_at: '2024-01-16T09:00:00Z'
		}
	];

	function handleClose() {
		console.log('Panel closed');
	}
</script>

<!-- Actor 詳細 -->
<Story name="ActorDetail">
	<div
		style="width: 360px; height: 400px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: auto; padding: 12px;"
	>
		<UseCaseViewPanel
			actor={sampleActors[0]}
			usecase={null}
			actors={sampleActors}
			usecases={sampleUseCases}
			onClose={handleClose}
		/>
	</div>
</Story>

<!-- システムアクター詳細 -->
<Story name="SystemActorDetail">
	<div
		style="width: 360px; height: 400px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: auto; padding: 12px;"
	>
		<UseCaseViewPanel
			actor={sampleActors[2]}
			usecase={null}
			actors={sampleActors}
			usecases={sampleUseCases}
			onClose={handleClose}
		/>
	</div>
</Story>

<!-- UseCase 詳細（シンプル） -->
<Story name="UseCaseSimple">
	<div
		style="width: 360px; height: 500px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: auto; padding: 12px;"
	>
		<UseCaseViewPanel
			actor={null}
			usecase={sampleUseCases[0]}
			actors={sampleActors}
			usecases={sampleUseCases}
			onClose={handleClose}
		/>
	</div>
</Story>

<!-- UseCase 詳細（リレーションあり） -->
<Story name="UseCaseWithRelations">
	<div
		style="width: 360px; height: 500px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: auto; padding: 12px;"
	>
		<UseCaseViewPanel
			actor={null}
			usecase={sampleUseCases[1]}
			actors={sampleActors}
			usecases={sampleUseCases}
			onClose={handleClose}
		/>
	</div>
</Story>

<!-- UseCase 詳細（シナリオあり） -->
<Story name="UseCaseWithScenario">
	<div
		style="width: 400px; height: 700px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: auto; padding: 12px;"
	>
		<UseCaseViewPanel
			actor={null}
			usecase={useCaseWithScenario}
			actors={sampleActors}
			usecases={sampleUseCases}
			activities={sampleActivities}
			onClose={handleClose}
		/>
	</div>
</Story>

<!-- 未選択状態 -->
<Story name="Empty">
	<div
		style="width: 360px; height: 300px; background: var(--bg-panel); border: 1px solid var(--border-metal); border-radius: 8px; overflow: auto; padding: 12px;"
	>
		<UseCaseViewPanel
			actor={null}
			usecase={null}
			actors={sampleActors}
			usecases={sampleUseCases}
			onClose={handleClose}
		/>
	</div>
</Story>
