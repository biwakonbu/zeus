<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import KeyboardHelp from './KeyboardHelp.svelte';

	const { Story } = defineMeta({
		title: 'UI/KeyboardHelp',
		component: KeyboardHelp,
		tags: ['autodocs'],
		parameters: {
			layout: 'fullscreen'
		}
	});
</script>

<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { keyboardStore } from '$lib/stores/keyboard';

	// Story は defineMeta から export される
	function handleClose() {
		console.log('KeyboardHelp closed');
	}

	// ショートカット登録関数
	let unregisterFns: (() => void)[] = [];

	function registerSampleShortcuts() {
		// サンプルショートカットを登録
		unregisterFns.push(
			keyboardStore.register({
				key: 'k',
				modifiers: ['cmd'],
				description: 'コマンドパレットを開く',
				category: 'ナビゲーション',
				action: () => console.log('Command palette')
			})
		);

		unregisterFns.push(
			keyboardStore.register({
				key: 'p',
				modifiers: ['cmd', 'shift'],
				description: 'プロジェクト検索',
				category: 'ナビゲーション',
				action: () => console.log('Project search')
			})
		);

		unregisterFns.push(
			keyboardStore.register({
				key: 'g',
				description: 'グラフビューに切り替え',
				category: 'ビュー',
				action: () => console.log('Switch to graph')
			})
		);

		unregisterFns.push(
			keyboardStore.register({
				key: 'w',
				description: 'WBSビューに切り替え',
				category: 'ビュー',
				action: () => console.log('Switch to WBS')
			})
		);

		unregisterFns.push(
			keyboardStore.register({
				key: 't',
				description: 'タイムラインビューに切り替え',
				category: 'ビュー',
				action: () => console.log('Switch to timeline')
			})
		);

		unregisterFns.push(
			keyboardStore.register({
				key: '+',
				modifiers: ['cmd'],
				description: 'ズームイン',
				category: '操作',
				action: () => console.log('Zoom in')
			})
		);

		unregisterFns.push(
			keyboardStore.register({
				key: '-',
				modifiers: ['cmd'],
				description: 'ズームアウト',
				category: '操作',
				action: () => console.log('Zoom out')
			})
		);

		unregisterFns.push(
			keyboardStore.register({
				key: '0',
				modifiers: ['cmd'],
				description: 'ズームリセット',
				category: '操作',
				action: () => console.log('Reset zoom')
			})
		);
	}

	function cleanup() {
		unregisterFns.forEach((fn) => fn());
		unregisterFns = [];
	}
</script>

<!-- デフォルト（サンプルショートカット付き） -->
<Story name="Default">
	{#snippet children()}
		{@const _ = (onMount(() => { registerSampleShortcuts(); return cleanup; }), undefined)}
		<div style="min-height: 500px; background: var(--bg-primary);">
			<KeyboardHelp onClose={handleClose} />
		</div>
	{/snippet}
</Story>

<!-- ショートカットなし -->
<Story name="NoShortcuts">
	{#snippet children()}
		{@const _ = (onMount(() => cleanup), undefined)}
		<div style="min-height: 500px; background: var(--bg-primary);">
			<KeyboardHelp onClose={handleClose} />
		</div>
	{/snippet}
</Story>
