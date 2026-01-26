<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Icon from './Icon.svelte';

	const { Story } = defineMeta({
		title: 'UI/Icon',
		component: Icon,
		tags: ['autodocs'],
		argTypes: {
			name: {
				control: 'text',
				description: 'Lucide Icons のアイコン名'
			},
			size: {
				control: { type: 'number', min: 12, max: 64 },
				description: 'アイコンサイズ（px）'
			},
			strokeWidth: {
				control: { type: 'number', min: 1, max: 4, step: 0.5 },
				description: '線の太さ'
			},
			glow: {
				control: 'boolean',
				description: 'グロー効果'
			}
		}
	});
</script>

<script lang="ts">
	// Story は defineMeta から export される
</script>

<!-- デフォルト -->
<Story name="Default">
	<div style="display: flex; gap: 16px; padding: 24px; background: var(--bg-primary);">
		<Icon name="Heart" />
		<Icon name="Settings" />
		<Icon name="AlertTriangle" />
	</div>
</Story>

<!-- サイズバリエーション -->
<Story name="Sizes">
	<div style="display: flex; align-items: center; gap: 24px; padding: 24px; background: var(--bg-primary);">
		<div style="text-align: center;">
			<Icon name="Heart" size={12} />
			<p style="color: var(--text-muted); font-size: 10px; margin-top: 4px;">12px</p>
		</div>
		<div style="text-align: center;">
			<Icon name="Heart" size={16} />
			<p style="color: var(--text-muted); font-size: 10px; margin-top: 4px;">16px</p>
		</div>
		<div style="text-align: center;">
			<Icon name="Heart" size={24} />
			<p style="color: var(--text-muted); font-size: 10px; margin-top: 4px;">24px</p>
		</div>
		<div style="text-align: center;">
			<Icon name="Heart" size={32} />
			<p style="color: var(--text-muted); font-size: 10px; margin-top: 4px;">32px</p>
		</div>
		<div style="text-align: center;">
			<Icon name="Heart" size={48} />
			<p style="color: var(--text-muted); font-size: 10px; margin-top: 4px;">48px</p>
		</div>
	</div>
</Story>

<!-- グロー効果 -->
<Story name="GlowEffect">
	<div style="display: flex; gap: 24px; padding: 24px; background: var(--bg-primary);">
		<div style="text-align: center;">
			<Icon name="Zap" size={32} />
			<p style="color: var(--text-muted); font-size: 10px; margin-top: 4px;">通常</p>
		</div>
		<div style="text-align: center;">
			<Icon name="Zap" size={32} glow />
			<p style="color: var(--text-muted); font-size: 10px; margin-top: 4px;">グロー</p>
		</div>
	</div>
</Story>

<!-- ストローク幅 -->
<Story name="StrokeWidths">
	<div style="display: flex; gap: 24px; padding: 24px; background: var(--bg-primary);">
		<div style="text-align: center;">
			<Icon name="Circle" size={32} strokeWidth={1} />
			<p style="color: var(--text-muted); font-size: 10px; margin-top: 4px;">1</p>
		</div>
		<div style="text-align: center;">
			<Icon name="Circle" size={32} strokeWidth={1.5} />
			<p style="color: var(--text-muted); font-size: 10px; margin-top: 4px;">1.5</p>
		</div>
		<div style="text-align: center;">
			<Icon name="Circle" size={32} strokeWidth={2} />
			<p style="color: var(--text-muted); font-size: 10px; margin-top: 4px;">2</p>
		</div>
		<div style="text-align: center;">
			<Icon name="Circle" size={32} strokeWidth={2.5} />
			<p style="color: var(--text-muted); font-size: 10px; margin-top: 4px;">2.5 (デフォルト)</p>
		</div>
		<div style="text-align: center;">
			<Icon name="Circle" size={32} strokeWidth={3} />
			<p style="color: var(--text-muted); font-size: 10px; margin-top: 4px;">3</p>
		</div>
	</div>
</Story>

<!-- カラー付き -->
<Story name="WithColors">
	<div style="display: flex; gap: 16px; padding: 24px; background: var(--bg-primary);">
		<span style="color: var(--status-good);">
			<Icon name="CheckCircle" size={24} />
		</span>
		<span style="color: var(--status-warning);">
			<Icon name="AlertTriangle" size={24} />
		</span>
		<span style="color: var(--status-poor);">
			<Icon name="XCircle" size={24} />
		</span>
		<span style="color: var(--status-info);">
			<Icon name="Info" size={24} />
		</span>
		<span style="color: var(--accent-primary);">
			<Icon name="Flame" size={24} />
		</span>
	</div>
</Story>

<!-- よく使うアイコン一覧 -->
<Story name="CommonIcons">
	<div style="padding: 24px; background: var(--bg-primary);">
		<h4 style="color: var(--text-secondary); margin-bottom: 16px; font-size: 12px;">ナビゲーション</h4>
		<div style="display: flex; flex-wrap: wrap; gap: 16px; margin-bottom: 24px;">
			{#each ['Heart', 'Calendar', 'Flame', 'RefreshCw', 'X', 'ChevronDown', 'ChevronRight', 'ChevronUp', 'ChevronLeft'] as iconName}
				<div style="text-align: center; width: 60px;">
					<Icon name={iconName} size={20} />
					<p style="color: var(--text-muted); font-size: 9px; margin-top: 4px; word-break: break-all;">{iconName}</p>
				</div>
			{/each}
		</div>

		<h4 style="color: var(--text-secondary); margin-bottom: 16px; font-size: 12px;">状態</h4>
		<div style="display: flex; flex-wrap: wrap; gap: 16px; margin-bottom: 24px;">
			{#each ['AlertTriangle', 'CheckCircle', 'Info', 'XCircle'] as iconName}
				<div style="text-align: center; width: 60px;">
					<Icon name={iconName} size={20} />
					<p style="color: var(--text-muted); font-size: 9px; margin-top: 4px; word-break: break-all;">{iconName}</p>
				</div>
			{/each}
		</div>

		<h4 style="color: var(--text-secondary); margin-bottom: 16px; font-size: 12px;">アクション</h4>
		<div style="display: flex; flex-wrap: wrap; gap: 16px; margin-bottom: 24px;">
			{#each ['ClipboardList', 'Target', 'BarChart', 'Ruler', 'ZoomIn', 'ZoomOut', 'Edit', 'Copy', 'Trash2', 'Search', 'ExternalLink'] as iconName}
				<div style="text-align: center; width: 60px;">
					<Icon name={iconName} size={20} />
					<p style="color: var(--text-muted); font-size: 9px; margin-top: 4px; word-break: break-all;">{iconName}</p>
				</div>
			{/each}
		</div>

		<h4 style="color: var(--text-secondary); margin-bottom: 16px; font-size: 12px;">UI</h4>
		<div style="display: flex; flex-wrap: wrap; gap: 16px;">
			{#each ['Keyboard', 'Inbox', 'Settings', 'Zap', 'Package', 'MoreHorizontal', 'MoreVertical'] as iconName}
				<div style="text-align: center; width: 60px;">
					<Icon name={iconName} size={20} />
					<p style="color: var(--text-muted); font-size: 9px; margin-top: 4px; word-break: break-all;">{iconName}</p>
				</div>
			{/each}
		</div>
	</div>
</Story>

<!-- フォールバック -->
<Story name="Fallback">
	<div style="display: flex; gap: 16px; padding: 24px; background: var(--bg-primary);">
		<div style="text-align: center;">
			<Icon name="Heart" size={24} />
			<p style="color: var(--text-muted); font-size: 10px; margin-top: 4px;">有効なアイコン</p>
		</div>
		<div style="text-align: center;">
			<Icon name="InvalidIconName" size={24} />
			<p style="color: var(--text-muted); font-size: 10px; margin-top: 4px;">無効なアイコン（フォールバック）</p>
		</div>
	</div>
</Story>
