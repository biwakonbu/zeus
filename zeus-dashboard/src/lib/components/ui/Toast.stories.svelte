<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Toast from './Toast.svelte';

	const { Story } = defineMeta({
		title: 'UI/Toast',
		component: Toast,
		tags: ['autodocs'],
		argTypes: {
			type: {
				control: 'select',
				options: ['info', 'success', 'warning', 'error'],
				description: 'トーストのタイプ'
			},
			message: {
				control: 'text',
				description: 'メッセージ内容'
			},
			dismissible: {
				control: 'boolean',
				description: '閉じるボタンの表示'
			}
		}
	});
</script>

<script lang="ts">
	// Story は defineMeta から export される
	function handleDismiss(id: string) {
		console.log('Dismissed:', id);
	}
</script>

<!-- Info -->
<Story name="Info">
	<div style="padding: 24px; background: var(--bg-primary);">
		<Toast
			id="toast-info"
			type="info"
			message="情報メッセージです。"
			dismissible={true}
			onDismiss={handleDismiss}
		/>
	</div>
</Story>

<!-- Success -->
<Story name="Success">
	<div style="padding: 24px; background: var(--bg-primary);">
		<Toast
			id="toast-success"
			type="success"
			message="操作が正常に完了しました。"
			dismissible={true}
			onDismiss={handleDismiss}
		/>
	</div>
</Story>

<!-- Warning -->
<Story name="Warning">
	<div style="padding: 24px; background: var(--bg-primary);">
		<Toast
			id="toast-warning"
			type="warning"
			message="注意: この操作は取り消せません。"
			dismissible={true}
			onDismiss={handleDismiss}
		/>
	</div>
</Story>

<!-- Error -->
<Story name="Error">
	<div style="padding: 24px; background: var(--bg-primary);">
		<Toast
			id="toast-error"
			type="error"
			message="エラーが発生しました。再試行してください。"
			dismissible={true}
			onDismiss={handleDismiss}
		/>
	</div>
</Story>

<!-- 閉じるボタンなし -->
<Story name="NotDismissible">
	<div style="padding: 24px; background: var(--bg-primary);">
		<Toast
			id="toast-no-dismiss"
			type="info"
			message="このトーストは自動的に消えます。"
			dismissible={false}
			onDismiss={handleDismiss}
		/>
	</div>
</Story>

<!-- 長いメッセージ -->
<Story name="LongMessage">
	<div style="padding: 24px; background: var(--bg-primary);">
		<Toast
			id="toast-long"
			type="warning"
			message="これは長いメッセージの例です。トースト通知は複数行にわたる場合でも、適切にレイアウトされます。重要な情報を伝えるために必要な文章量を使用できます。"
			dismissible={true}
			onDismiss={handleDismiss}
		/>
	</div>
</Story>

<!-- 全バリエーション -->
<Story name="AllVariants">
	<div
		style="display: flex; flex-direction: column; gap: 12px; padding: 24px; background: var(--bg-primary);"
	>
		<Toast
			id="toast-all-info"
			type="info"
			message="情報: プロジェクトが更新されました。"
			dismissible={true}
			onDismiss={handleDismiss}
		/>
		<Toast
			id="toast-all-success"
			type="success"
			message="成功: タスクが完了しました。"
			dismissible={true}
			onDismiss={handleDismiss}
		/>
		<Toast
			id="toast-all-warning"
			type="warning"
			message="警告: 期限が近づいています。"
			dismissible={true}
			onDismiss={handleDismiss}
		/>
		<Toast
			id="toast-all-error"
			type="error"
			message="エラー: 接続に失敗しました。"
			dismissible={true}
			onDismiss={handleDismiss}
		/>
	</div>
</Story>
