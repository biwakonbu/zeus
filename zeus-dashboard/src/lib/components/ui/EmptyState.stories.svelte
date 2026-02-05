<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import EmptyState from './EmptyState.svelte';

	const { Story } = defineMeta({
		title: 'UI/EmptyState',
		component: EmptyState,
		tags: ['autodocs'],
		argTypes: {
			icon: {
				control: 'text',
				description: 'Lucide Icons のアイコン名'
			},
			title: {
				control: 'text',
				description: 'メインタイトル'
			},
			description: {
				control: 'text',
				description: '説明文'
			},
			actionLabel: {
				control: 'text',
				description: 'アクションボタンのラベル'
			}
		}
	});
</script>

<script lang="ts">
	// Story は defineMeta から export される
</script>

<!-- デフォルト（シンプル） -->
<Story name="Default">
	<div style="background: var(--bg-primary); padding: 24px;">
		<EmptyState title="データがありません" />
	</div>
</Story>

<!-- 説明文付き -->
<Story name="WithDescription">
	<div style="background: var(--bg-primary); padding: 24px;">
		<EmptyState
			title="タスクがありません"
			description="新しいタスクを作成して、プロジェクトを開始しましょう。"
		/>
	</div>
</Story>

<!-- アクションボタン付き -->
<Story name="WithAction">
	<div style="background: var(--bg-primary); padding: 24px;">
		<EmptyState
			title="タスクがありません"
			description="新しいタスクを作成して、プロジェクトを開始しましょう。"
			actionLabel="タスクを作成"
			onAction={() => alert('タスク作成ボタンがクリックされました')}
		/>
	</div>
</Story>

<!-- カスタムアイコン -->
<Story name="CustomIcon">
	<div style="background: var(--bg-primary); padding: 24px;">
		<EmptyState
			icon="Search"
			title="検索結果がありません"
			description="別のキーワードで検索してみてください。"
		/>
	</div>
</Story>

<!-- 各種バリエーション -->
<Story name="Variations">
	<div
		style="display: flex; flex-direction: column; gap: 32px; background: var(--bg-primary); padding: 24px;"
	>
		<div style="border: 1px solid var(--border-metal); border-radius: 8px;">
			<EmptyState
				icon="ClipboardList"
				title="タスクがありません"
				description="zeus add task で新しいタスクを追加してください。"
				actionLabel="ヘルプを見る"
				onAction={() => alert('ヘルプを表示')}
			/>
		</div>
		<div style="border: 1px solid var(--border-metal); border-radius: 8px;">
			<EmptyState
				icon="Target"
				title="目標が設定されていません"
				description="プロジェクトの目標を設定することで、進捗を追跡できます。"
			/>
		</div>
		<div style="border: 1px solid var(--border-metal); border-radius: 8px;">
			<EmptyState
				icon="AlertTriangle"
				title="エラーが発生しました"
				description="データの読み込みに失敗しました。再試行してください。"
				actionLabel="再試行"
				onAction={() => alert('再試行')}
			/>
		</div>
	</div>
</Story>
