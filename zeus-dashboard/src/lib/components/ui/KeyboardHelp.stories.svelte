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
	// Story は defineMeta から export される
	function handleClose() {
		console.log('KeyboardHelp closed');
	}
</script>

<!-- デフォルト（サンプルショートカット付き） -->
<Story name="Default">
	<div style="min-height: 500px; background: var(--bg-primary);">
		<KeyboardHelp onClose={handleClose} />
	</div>
</Story>

<!-- ショートカットなし -->
<Story name="NoShortcuts">
	<div style="min-height: 500px; background: var(--bg-primary);">
		<KeyboardHelp onClose={handleClose} />
	</div>
</Story>
