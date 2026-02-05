<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import OverlayPanel from './OverlayPanel.svelte';

	const { Story } = defineMeta({
		title: 'UI/OverlayPanel',
		component: OverlayPanel,
		tags: ['autodocs'],
		argTypes: {
			title: {
				control: 'text',
				description: 'パネルタイトル'
			},
			position: {
				control: 'select',
				options: ['top-left', 'top-right', 'bottom-left', 'bottom-right'],
				description: '表示位置'
			},
			width: {
				control: 'text',
				description: '幅（CSS値）'
			},
			maxHeight: {
				control: 'text',
				description: '最大高さ（CSS値）'
			},
			showCloseButton: {
				control: 'boolean',
				description: '閉じるボタンの表示'
			}
		},
		parameters: {
			layout: 'fullscreen'
		}
	});
</script>

<script lang="ts">
	// Story は defineMeta から export される
	function handleClose() {
		console.log('OverlayPanel closed');
	}
</script>

<!-- 左上（デフォルト） -->
<Story name="TopLeft">
	<div style="position: relative; width: 100%; height: 400px; background: var(--bg-primary);">
		<OverlayPanel title="要素一覧" position="top-left" onClose={handleClose}>
			<div style="padding: 12px;">
				<p style="color: var(--text-secondary); font-size: 13px; margin: 0;">
					左上に配置されたオーバーレイパネルです。
				</p>
				<ul
					style="color: var(--text-secondary); font-size: 13px; padding-left: 20px; margin: 12px 0 0;"
				>
					<li>項目 1</li>
					<li>項目 2</li>
					<li>項目 3</li>
				</ul>
			</div>
		</OverlayPanel>
	</div>
</Story>

<!-- 右上 -->
<Story name="TopRight">
	<div style="position: relative; width: 100%; height: 400px; background: var(--bg-primary);">
		<OverlayPanel title="詳細情報" position="top-right" onClose={handleClose}>
			<div style="padding: 12px;">
				<p style="color: var(--text-secondary); font-size: 13px; margin: 0;">
					右上に配置されたオーバーレイパネルです。
				</p>
			</div>
		</OverlayPanel>
	</div>
</Story>

<!-- 左下 -->
<Story name="BottomLeft">
	<div style="position: relative; width: 100%; height: 400px; background: var(--bg-primary);">
		<OverlayPanel title="フィルター" position="bottom-left" onClose={handleClose}>
			<div style="padding: 12px;">
				<p style="color: var(--text-secondary); font-size: 13px; margin: 0;">
					左下に配置されたオーバーレイパネルです。
				</p>
			</div>
		</OverlayPanel>
	</div>
</Story>

<!-- 右下 -->
<Story name="BottomRight">
	<div style="position: relative; width: 100%; height: 400px; background: var(--bg-primary);">
		<OverlayPanel title="設定" position="bottom-right" onClose={handleClose}>
			<div style="padding: 12px;">
				<p style="color: var(--text-secondary); font-size: 13px; margin: 0;">
					右下に配置されたオーバーレイパネルです。
				</p>
			</div>
		</OverlayPanel>
	</div>
</Story>

<!-- 閉じるボタンなし -->
<Story name="NoCloseButton">
	<div style="position: relative; width: 100%; height: 400px; background: var(--bg-primary);">
		<OverlayPanel title="常時表示" position="top-left" showCloseButton={false}>
			<div style="padding: 12px;">
				<p style="color: var(--text-secondary); font-size: 13px; margin: 0;">
					閉じるボタンのないパネルです。
				</p>
			</div>
		</OverlayPanel>
	</div>
</Story>

<!-- カスタムサイズ -->
<Story name="CustomSize">
	<div style="position: relative; width: 100%; height: 500px; background: var(--bg-primary);">
		<OverlayPanel
			title="カスタムサイズ"
			position="top-left"
			width="400px"
			maxHeight="300px"
			onClose={handleClose}
		>
			<div style="padding: 12px;">
				<p style="color: var(--text-secondary); font-size: 13px; margin: 0 0 12px;">
					幅400px、最大高さ300pxのパネルです。
				</p>
				{#each Array(20) as _, i}
					<p style="color: var(--text-muted); font-size: 12px; margin: 4px 0;">
						スクロール可能なコンテンツ {i + 1}
					</p>
				{/each}
			</div>
		</OverlayPanel>
	</div>
</Story>

<!-- 複数パネル -->
<Story name="MultiplePositions">
	<div style="position: relative; width: 100%; height: 500px; background: var(--bg-primary);">
		<OverlayPanel title="左上パネル" position="top-left" width="200px" onClose={handleClose}>
			<div style="padding: 12px;">
				<p style="color: var(--text-secondary); font-size: 13px; margin: 0;">要素一覧</p>
			</div>
		</OverlayPanel>
		<OverlayPanel title="右上パネル" position="top-right" width="200px" onClose={handleClose}>
			<div style="padding: 12px;">
				<p style="color: var(--text-secondary); font-size: 13px; margin: 0;">詳細情報</p>
			</div>
		</OverlayPanel>
	</div>
</Story>

<!-- リッチコンテンツ -->
<Story name="RichContent">
	<div style="position: relative; width: 100%; height: 500px; background: var(--bg-primary);">
		<OverlayPanel title="タスク詳細" position="top-right" width="320px" onClose={handleClose}>
			<div style="padding: 12px;">
				<div style="margin-bottom: 16px;">
					<span style="color: var(--text-muted); font-size: 11px;">ID</span>
					<p style="color: var(--text-primary); font-size: 14px; margin: 4px 0 0;">task-001</p>
				</div>
				<div style="margin-bottom: 16px;">
					<span style="color: var(--text-muted); font-size: 11px;">タイトル</span>
					<p style="color: var(--text-primary); font-size: 14px; margin: 4px 0 0;">
						ダッシュボード実装
					</p>
				</div>
				<div style="margin-bottom: 16px;">
					<span style="color: var(--text-muted); font-size: 11px;">ステータス</span>
					<p style="color: var(--status-info); font-size: 14px; margin: 4px 0 0;">進行中</p>
				</div>
				<div style="margin-bottom: 16px;">
					<span style="color: var(--text-muted); font-size: 11px;">進捗</span>
					<div
						style="margin-top: 8px; height: 8px; background: var(--bg-secondary); border-radius: 4px; overflow: hidden;"
					>
						<div style="width: 65%; height: 100%; background: var(--accent-primary);"></div>
					</div>
					<p style="color: var(--text-secondary); font-size: 12px; margin: 4px 0 0;">65%</p>
				</div>
			</div>
		</OverlayPanel>
	</div>
</Story>
