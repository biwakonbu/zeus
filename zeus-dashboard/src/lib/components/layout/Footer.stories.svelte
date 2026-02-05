<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Footer from './Footer.svelte';

	const { Story } = defineMeta({
		title: 'Layout/Footer',
		component: Footer,
		tags: ['autodocs'],
		parameters: {
			layout: 'fullscreen'
		}
	});
</script>

<script lang="ts">
	// Story は defineMeta から export される
</script>

<!-- デフォルト -->
<Story name="Default">
	<div style="min-height: 100px; display: flex; flex-direction: column;">
		<div style="flex: 1;"></div>
		<Footer />
	</div>
</Story>

<!-- ページレイアウト内での表示 -->
<Story name="InPageLayout">
	<div
		style="min-height: 300px; display: flex; flex-direction: column; background: var(--bg-primary);"
	>
		<div style="flex: 1; display: flex; align-items: center; justify-content: center;">
			<p style="color: var(--text-muted);">メインコンテンツエリア</p>
		</div>
		<Footer />
	</div>
</Story>
