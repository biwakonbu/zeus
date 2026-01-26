<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import RichTooltip from './RichTooltip.svelte';
	import type { TooltipEntity } from './types';

	const { Story } = defineMeta({
		title: 'UI/RichTooltip',
		component: RichTooltip,
		tags: ['autodocs'],
		argTypes: {
			visible: {
				control: 'boolean',
				description: '表示状態'
			},
			position: {
				description: '表示位置（{x, y}）'
			}
		}
	});
</script>

<script lang="ts">
	// Story は defineMeta から export される

	// サンプルエンティティ
	const taskEntity: TooltipEntity = {
		id: 'task-001',
		title: 'ダッシュボードの実装',
		type: 'task',
		status: 'in_progress',
		progress: 65,
		lastUpdate: '2024-01-15 14:30'
	};

	const completedEntity: TooltipEntity = {
		id: 'task-002',
		title: 'API エンドポイントの設計',
		type: 'task',
		status: 'completed',
		progress: 100,
		lastUpdate: '2024-01-14 10:00'
	};

	const blockedEntity: TooltipEntity = {
		id: 'task-003',
		title: '外部サービスとの連携実装',
		type: 'task',
		status: 'blocked',
		progress: 30,
		lastUpdate: '2024-01-13 16:45'
	};

	const objectiveEntity: TooltipEntity = {
		id: 'obj-001',
		title: 'ユーザー体験の改善',
		type: 'objective',
		status: 'in_progress',
		progress: 45,
		lastUpdate: '2024-01-15 09:00'
	};

	const deliverableEntity: TooltipEntity = {
		id: 'del-001',
		title: 'フロントエンドアプリケーション',
		type: 'deliverable',
		status: 'in_progress',
		progress: 80
	};

	const visionEntity: TooltipEntity = {
		id: 'vision',
		title: '最高のプロジェクト管理ツールを作る',
		type: 'vision',
		status: 'in_progress',
		progress: 25
	};
</script>

<!-- 進行中タスク -->
<Story name="InProgressTask">
	<div style="position: relative; width: 100%; height: 300px; background: var(--bg-primary); padding: 24px;">
		<RichTooltip visible={true} entity={taskEntity} position={{ x: 50, y: 50 }} />
	</div>
</Story>

<!-- 完了タスク -->
<Story name="CompletedTask">
	<div style="position: relative; width: 100%; height: 300px; background: var(--bg-primary); padding: 24px;">
		<RichTooltip visible={true} entity={completedEntity} position={{ x: 50, y: 50 }} />
	</div>
</Story>

<!-- ブロックされたタスク -->
<Story name="BlockedTask">
	<div style="position: relative; width: 100%; height: 300px; background: var(--bg-primary); padding: 24px;">
		<RichTooltip visible={true} entity={blockedEntity} position={{ x: 50, y: 50 }} />
	</div>
</Story>

<!-- Objective -->
<Story name="Objective">
	<div style="position: relative; width: 100%; height: 300px; background: var(--bg-primary); padding: 24px;">
		<RichTooltip visible={true} entity={objectiveEntity} position={{ x: 50, y: 50 }} />
	</div>
</Story>

<!-- Deliverable -->
<Story name="Deliverable">
	<div style="position: relative; width: 100%; height: 300px; background: var(--bg-primary); padding: 24px;">
		<RichTooltip visible={true} entity={deliverableEntity} position={{ x: 50, y: 50 }} />
	</div>
</Story>

<!-- Vision -->
<Story name="Vision">
	<div style="position: relative; width: 100%; height: 300px; background: var(--bg-primary); padding: 24px;">
		<RichTooltip visible={true} entity={visionEntity} position={{ x: 50, y: 50 }} />
	</div>
</Story>

<!-- 非表示 -->
<Story name="Hidden">
	<div style="position: relative; width: 100%; height: 200px; background: var(--bg-primary); padding: 24px;">
		<p style="color: var(--text-secondary);">visible=false の場合、ツールチップは表示されません。</p>
		<RichTooltip visible={false} entity={taskEntity} position={{ x: 50, y: 50 }} />
	</div>
</Story>

<!-- 各ステータス一覧 -->
<Story name="AllStatuses">
	{@const pendingEntity: TooltipEntity = {
		id: 'task-pending',
		title: '待機中のタスク',
		type: 'task',
		status: 'pending',
		progress: 0
	}}
	{@const onHoldEntity: TooltipEntity = {
		id: 'task-hold',
		title: '保留中のタスク',
		type: 'task',
		status: 'on_hold',
		progress: 50
	}}
	<div style="display: flex; flex-wrap: wrap; gap: 16px; padding: 24px; background: var(--bg-primary);">
		<div style="position: relative; width: 360px; height: 260px; border: 1px solid var(--border-metal); border-radius: 8px;">
			<p style="color: var(--text-muted); font-size: 10px; padding: 8px;">in_progress</p>
			<RichTooltip visible={true} entity={taskEntity} position={{ x: 16, y: 32 }} />
		</div>
		<div style="position: relative; width: 360px; height: 260px; border: 1px solid var(--border-metal); border-radius: 8px;">
			<p style="color: var(--text-muted); font-size: 10px; padding: 8px;">completed</p>
			<RichTooltip visible={true} entity={completedEntity} position={{ x: 16, y: 32 }} />
		</div>
		<div style="position: relative; width: 360px; height: 260px; border: 1px solid var(--border-metal); border-radius: 8px;">
			<p style="color: var(--text-muted); font-size: 10px; padding: 8px;">blocked</p>
			<RichTooltip visible={true} entity={blockedEntity} position={{ x: 16, y: 32 }} />
		</div>
		<div style="position: relative; width: 360px; height: 260px; border: 1px solid var(--border-metal); border-radius: 8px;">
			<p style="color: var(--text-muted); font-size: 10px; padding: 8px;">pending</p>
			<RichTooltip visible={true} entity={pendingEntity} position={{ x: 16, y: 32 }} />
		</div>
		<div style="position: relative; width: 360px; height: 260px; border: 1px solid var(--border-metal); border-radius: 8px;">
			<p style="color: var(--text-muted); font-size: 10px; padding: 8px;">on_hold</p>
			<RichTooltip visible={true} entity={onHoldEntity} position={{ x: 16, y: 32 }} />
		</div>
	</div>
</Story>
