<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Badge from './Badge.svelte';

	const { Story } = defineMeta({
		title: 'UI/Badge',
		component: Badge,
		tags: ['autodocs'],
		argTypes: {
			variant: {
				control: 'select',
				options: ['default', 'success', 'warning', 'danger', 'info', 'muted']
			},
			size: {
				control: 'radio',
				options: ['sm', 'md']
			}
		}
	});
</script>

<script lang="ts">
	// Story は defineMeta から export される
</script>

<!-- デフォルト -->
<Story name="Default">
	<Badge>Default</Badge>
</Story>

<!-- Success -->
<Story name="Success">
	<Badge variant="success">完了</Badge>
</Story>

<!-- Warning -->
<Story name="Warning">
	<Badge variant="warning">注意</Badge>
</Story>

<!-- Danger -->
<Story name="Danger">
	<Badge variant="danger">危険</Badge>
</Story>

<!-- Info -->
<Story name="Info">
	<Badge variant="info">情報</Badge>
</Story>

<!-- Muted -->
<Story name="Muted">
	<Badge variant="muted">非活性</Badge>
</Story>

<!-- Small Size -->
<Story name="SmallSize">
	<div style="display: flex; gap: 8px; align-items: center;">
		<Badge variant="success" size="sm">SM</Badge>
		<Badge variant="warning" size="sm">Small</Badge>
		<Badge variant="danger" size="sm">Badge</Badge>
	</div>
</Story>

<!-- All Variants -->
<Story name="AllVariants">
	<div style="display: flex; flex-direction: column; gap: 16px;">
		<div>
			<h4 style="color: var(--text-secondary); margin-bottom: 8px; font-size: 12px;">Medium Size (md)</h4>
			<div style="display: flex; gap: 8px; flex-wrap: wrap;">
				<Badge variant="default">Default</Badge>
				<Badge variant="success">Success</Badge>
				<Badge variant="warning">Warning</Badge>
				<Badge variant="danger">Danger</Badge>
				<Badge variant="info">Info</Badge>
				<Badge variant="muted">Muted</Badge>
			</div>
		</div>
		<div>
			<h4 style="color: var(--text-secondary); margin-bottom: 8px; font-size: 12px;">Small Size (sm)</h4>
			<div style="display: flex; gap: 8px; flex-wrap: wrap;">
				<Badge variant="default" size="sm">Default</Badge>
				<Badge variant="success" size="sm">Success</Badge>
				<Badge variant="warning" size="sm">Warning</Badge>
				<Badge variant="danger" size="sm">Danger</Badge>
				<Badge variant="info" size="sm">Info</Badge>
				<Badge variant="muted" size="sm">Muted</Badge>
			</div>
		</div>
	</div>
</Story>

<!-- Status Examples -->
<Story name="StatusExamples">
	<div style="display: flex; flex-direction: column; gap: 8px;">
		<div style="display: flex; align-items: center; gap: 12px;">
			<span style="color: var(--text-secondary); font-size: 12px; width: 80px;">タスク:</span>
			<Badge variant="success">Completed</Badge>
			<Badge variant="warning">In Progress</Badge>
			<Badge variant="muted">Pending</Badge>
			<Badge variant="danger">Blocked</Badge>
		</div>
		<div style="display: flex; align-items: center; gap: 12px;">
			<span style="color: var(--text-secondary); font-size: 12px; width: 80px;">優先度:</span>
			<Badge variant="danger" size="sm">High</Badge>
			<Badge variant="warning" size="sm">Medium</Badge>
			<Badge variant="info" size="sm">Low</Badge>
		</div>
	</div>
</Story>
