<script context="module" lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Header from './Header.svelte';

	const { Story } = defineMeta({
		title: 'Layout/Header',
		component: Header,
		tags: ['autodocs'],
		parameters: {
			layout: 'fullscreen'
		},
		args: {
			connectionState: 'connected'
		},
		argTypes: {
			connectionState: {
				control: 'select',
				options: ['connected', 'connecting', 'disconnected']
			}
		}
	});
</script>

<script lang="ts">
	// Story は defineMeta から export される
</script>

<!-- 接続済み -->
<Story name="Connected" args={{ connectionState: 'connected' }}>
	<Header connectionState="connected" />
</Story>

<!-- 接続中 -->
<Story name="Connecting" args={{ connectionState: 'connecting' }}>
	<Header connectionState="connecting" />
</Story>

<!-- 切断済み -->
<Story name="Disconnected" args={{ connectionState: 'disconnected' }}>
	<Header connectionState="disconnected" />
</Story>

<!-- 全状態比較 -->
<Story name="AllStates">
	<div style="display: flex; flex-direction: column; gap: 16px;">
		<div>
			<p style="color: var(--text-secondary); font-size: 12px; margin-bottom: 4px;">Connected</p>
			<Header connectionState="connected" />
		</div>
		<div>
			<p style="color: var(--text-secondary); font-size: 12px; margin-bottom: 4px;">Connecting</p>
			<Header connectionState="connecting" />
		</div>
		<div>
			<p style="color: var(--text-secondary); font-size: 12px; margin-bottom: 4px;">Disconnected</p>
			<Header connectionState="disconnected" />
		</div>
	</div>
</Story>

<!-- インタラクティブ -->
<Story name="Interactive">
	<Header connectionState="connected" />
</Story>
