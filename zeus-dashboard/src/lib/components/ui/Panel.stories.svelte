<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Panel from './Panel.svelte';

	const { Story } = defineMeta({
		title: 'UI/Panel',
		component: Panel,
		tags: ['autodocs'],
		argTypes: {
			title: {
				control: 'text'
			},
			icon: {
				control: 'text'
			},
			loading: {
				control: 'boolean'
			},
			error: {
				control: 'text'
			}
		}
	});
</script>

<script lang="ts">
	import Badge from './Badge.svelte';
	import ProgressBar from './ProgressBar.svelte';
</script>

<!-- 基本 -->
<Story name="Default">
	<div style="width: 400px;">
		<Panel title="プロジェクト概要">
			<p style="color: var(--text-secondary);">
				パネルの基本的な使用例です。コンテンツはこのように表示されます。
			</p>
		</Panel>
	</div>
</Story>

<!-- アイコン付き -->
<Story name="WithIcon">
	<div style="width: 400px;">
		<Panel title="タスク一覧" icon="📋">
			<p style="color: var(--text-secondary);">
				アイコンを指定するとタイトルの左側に表示されます。
			</p>
		</Panel>
	</div>
</Story>

<!-- ローディング状態 -->
<Story name="Loading">
	<div style="width: 400px;">
		<Panel title="データ取得中" icon="⏳" loading={true}>
			<p>このコンテンツは表示されません</p>
		</Panel>
	</div>
</Story>

<!-- エラー状態 -->
<Story name="Error">
	<div style="width: 400px;">
		<Panel title="接続エラー" icon="🔌" error="サーバーへの接続に失敗しました。再試行してください。">
			<p>このコンテンツは表示されません</p>
		</Panel>
	</div>
</Story>

<!-- ヘッダー右側にコンテンツ -->
<Story name="WithHeaderRight">
	<div style="width: 400px;">
		<Panel title="プロジェクト進捗" icon="📊">
			{#snippet headerRight()}
				<Badge variant="success">稼働中</Badge>
			{/snippet}
			<div style="display: flex; flex-direction: column; gap: 12px;">
				<div style="display: flex; justify-content: space-between; color: var(--text-secondary); font-size: 14px;">
					<span>完了タスク</span>
					<span>12 / 20</span>
				</div>
				<ProgressBar value={60} />
			</div>
		</Panel>
	</div>
</Story>

<!-- 複雑なコンテンツ -->
<Story name="ComplexContent">
	<div style="width: 400px;">
		<Panel title="タスク詳細" icon="📝">
			{#snippet headerRight()}
				<Badge variant="warning" size="sm">IN PROGRESS</Badge>
			{/snippet}
			<div style="display: flex; flex-direction: column; gap: 12px;">
				<div style="display: flex; justify-content: space-between; align-items: center; padding-bottom: 8px; border-bottom: 1px solid var(--border-dark);">
					<span style="color: var(--text-secondary); font-size: 12px;">タスク ID</span>
					<span style="color: var(--accent-primary); font-family: monospace;">task-42</span>
				</div>
				<div style="display: flex; justify-content: space-between; align-items: center; padding-bottom: 8px; border-bottom: 1px solid var(--border-dark);">
					<span style="color: var(--text-secondary); font-size: 12px;">担当者</span>
					<span style="color: var(--text-primary);">alice</span>
				</div>
				<div style="display: flex; justify-content: space-between; align-items: center; padding-bottom: 8px; border-bottom: 1px solid var(--border-dark);">
					<span style="color: var(--text-secondary); font-size: 12px;">優先度</span>
					<Badge variant="danger" size="sm">HIGH</Badge>
				</div>
				<div>
					<span style="color: var(--text-secondary); font-size: 12px; display: block; margin-bottom: 8px;">進捗</span>
					<ProgressBar value={60} size="sm" />
				</div>
			</div>
		</Panel>
	</div>
</Story>

<!-- 全状態比較 -->
<Story name="AllStates">
	<div style="display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; width: 800px;">
		<Panel title="通常状態" icon="✓">
			<p style="color: var(--text-secondary); font-size: 14px;">正常に表示されています。</p>
		</Panel>
		<Panel title="ローディング" icon="⏳" loading={true}>
			<p>表示されません</p>
		</Panel>
		<Panel title="エラー" icon="⚠" error="エラーが発生しました">
			<p>表示されません</p>
		</Panel>
		<Panel title="ヘッダー右" icon="📊">
			{#snippet headerRight()}
				<Badge variant="info" size="sm">INFO</Badge>
			{/snippet}
			<p style="color: var(--text-secondary); font-size: 14px;">ヘッダー右にコンテンツを追加できます。</p>
		</Panel>
	</div>
</Story>
