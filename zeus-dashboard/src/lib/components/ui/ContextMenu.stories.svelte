<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ContextMenu from './ContextMenu.svelte';
	import type { ContextMenuItem } from './ContextMenu.svelte';

	const { Story } = defineMeta({
		title: 'UI/ContextMenu',
		component: ContextMenu,
		tags: ['autodocs'],
		argTypes: {
			x: {
				control: { type: 'number', min: 0, max: 500 },
				description: 'X座標'
			},
			y: {
				control: { type: 'number', min: 0, max: 500 },
				description: 'Y座標'
			}
		}
	});
</script>

<script lang="ts">
	// Story は defineMeta から export される
	const basicItems: ContextMenuItem[] = [
		{ id: 'edit', label: '編集', icon: 'Edit' },
		{ id: 'copy', label: 'コピー', icon: 'Copy', shortcut: '⌘C' },
		{ id: 'delete', label: '削除', icon: 'Trash2', danger: true }
	];

	const fullItems: ContextMenuItem[] = [
		{ id: 'view', label: '詳細を表示', icon: 'ExternalLink' },
		{ id: 'edit', label: '編集', icon: 'Edit', shortcut: '⌘E' },
		{ id: 'copy', label: 'コピー', icon: 'Copy', shortcut: '⌘C' },
		{ id: 'sep1', label: '', separator: true },
		{ id: 'approve', label: '承認', icon: 'CheckCircle' },
		{ id: 'reject', label: '却下', icon: 'XCircle', disabled: true },
		{ id: 'sep2', label: '', separator: true },
		{ id: 'delete', label: '削除', icon: 'Trash2', danger: true, shortcut: '⌫' }
	];

	const simpleItems: ContextMenuItem[] = [
		{ id: 'action1', label: 'アクション1' },
		{ id: 'action2', label: 'アクション2' },
		{ id: 'action3', label: 'アクション3' }
	];

	function handleSelect(id: string) {
		console.log('Selected:', id);
	}

	function handleClose() {
		console.log('Menu closed');
	}
</script>

<!-- 基本 -->
<Story name="Default">
	<div style="position: relative; width: 400px; height: 300px; background: var(--bg-primary);">
		<ContextMenu
			items={basicItems}
			x={50}
			y={50}
			onSelect={handleSelect}
			onClose={handleClose}
		/>
	</div>
</Story>

<!-- フル機能 -->
<Story name="FullFeatures">
	<div style="position: relative; width: 500px; height: 400px; background: var(--bg-primary);">
		<ContextMenu
			items={fullItems}
			x={50}
			y={50}
			onSelect={handleSelect}
			onClose={handleClose}
		/>
	</div>
</Story>

<!-- シンプル（アイコンなし） -->
<Story name="SimpleNoIcons">
	<div style="position: relative; width: 400px; height: 250px; background: var(--bg-primary);">
		<ContextMenu
			items={simpleItems}
			x={50}
			y={50}
			onSelect={handleSelect}
			onClose={handleClose}
		/>
	</div>
</Story>

<!-- 無効化項目 -->
<Story name="WithDisabledItems">
	{@const disabledItems: ContextMenuItem[] = [
		{ id: 'enabled', label: '有効なアクション', icon: 'Zap' },
		{ id: 'disabled1', label: '無効なアクション', icon: 'Settings', disabled: true },
		{ id: 'disabled2', label: '権限がありません', icon: 'AlertTriangle', disabled: true },
		{ id: 'sep', label: '', separator: true },
		{ id: 'another', label: '別のアクション', icon: 'MoreHorizontal' }
	]}
	<div style="position: relative; width: 400px; height: 300px; background: var(--bg-primary);">
		<ContextMenu
			items={disabledItems}
			x={50}
			y={50}
			onSelect={handleSelect}
			onClose={handleClose}
		/>
	</div>
</Story>

<!-- 危険アクション -->
<Story name="DangerActions">
	{@const dangerItems: ContextMenuItem[] = [
		{ id: 'view', label: '詳細を表示', icon: 'ExternalLink' },
		{ id: 'sep', label: '', separator: true },
		{ id: 'archive', label: 'アーカイブ', icon: 'Package', danger: true },
		{ id: 'delete', label: '完全に削除', icon: 'Trash2', danger: true }
	]}
	<div style="position: relative; width: 400px; height: 300px; background: var(--bg-primary);">
		<ContextMenu
			items={dangerItems}
			x={50}
			y={50}
			onSelect={handleSelect}
			onClose={handleClose}
		/>
	</div>
</Story>
