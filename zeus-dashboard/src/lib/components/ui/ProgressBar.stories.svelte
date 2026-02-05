<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ProgressBar from './ProgressBar.svelte';

	const { Story } = defineMeta({
		title: 'UI/ProgressBar',
		component: ProgressBar,
		tags: ['autodocs'],
		args: {
			value: 50,
			max: 100,
			showLabel: true,
			size: 'md'
		},
		argTypes: {
			value: {
				control: { type: 'range', min: 0, max: 100, step: 1 }
			},
			max: {
				control: { type: 'number', min: 1 }
			},
			showLabel: {
				control: 'boolean'
			},
			size: {
				control: 'radio',
				options: ['sm', 'md', 'lg']
			}
		}
	});
</script>

<script lang="ts">
	import { onMount } from 'svelte';

	// Animated Story 用の状態
	let animatedValue = $state(0);
	let animatedDirection = $state(1);

	onMount(() => {
		const interval = setInterval(() => {
			animatedValue += animatedDirection * 2;
			if (animatedValue >= 100) {
				animatedDirection = -1;
				animatedValue = 100;
			} else if (animatedValue <= 0) {
				animatedDirection = 1;
				animatedValue = 0;
			}
		}, 50);
		return () => clearInterval(interval);
	});
</script>

<!-- 空（0%） -->
<Story name="Empty" args={{ value: 0 }}>
	<div style="width: 300px;">
		<ProgressBar value={0} />
	</div>
</Story>

<!-- 半分（50%） -->
<Story name="Half" args={{ value: 50 }}>
	<div style="width: 300px;">
		<ProgressBar value={50} />
	</div>
</Story>

<!-- 完了（100%） -->
<Story name="Complete" args={{ value: 100 }}>
	<div style="width: 300px;">
		<ProgressBar value={100} />
	</div>
</Story>

<!-- ラベルなし -->
<Story name="NoLabel" args={{ value: 75, showLabel: false }}>
	<div style="width: 300px;">
		<ProgressBar value={75} showLabel={false} />
	</div>
</Story>

<!-- 小サイズ -->
<Story name="SizeSmall" args={{ value: 60, size: 'sm' }}>
	<div style="width: 300px;">
		<span style="color: var(--text-secondary); font-size: 12px; display: block; margin-bottom: 4px;"
			>Small (sm)</span
		>
		<ProgressBar value={60} size="sm" />
	</div>
</Story>

<!-- 中サイズ -->
<Story name="SizeMedium" args={{ value: 60, size: 'md' }}>
	<div style="width: 300px;">
		<span style="color: var(--text-secondary); font-size: 12px; display: block; margin-bottom: 4px;"
			>Medium (md)</span
		>
		<ProgressBar value={60} size="md" />
	</div>
</Story>

<!-- 大サイズ -->
<Story name="SizeLarge" args={{ value: 60, size: 'lg' }}>
	<div style="width: 300px;">
		<span style="color: var(--text-secondary); font-size: 12px; display: block; margin-bottom: 4px;"
			>Large (lg)</span
		>
		<ProgressBar value={60} size="lg" />
	</div>
</Story>

<!-- 低い進捗（赤） -->
<Story name="ColorPoor" args={{ value: 30 }}>
	<div style="width: 300px;">
		<span style="color: var(--text-secondary); font-size: 12px; display: block; margin-bottom: 4px;"
			>0-49%: 赤（Poor）</span
		>
		<ProgressBar value={30} />
	</div>
</Story>

<!-- 中程度の進捗（黄） -->
<Story name="ColorFair" args={{ value: 65 }}>
	<div style="width: 300px;">
		<span style="color: var(--text-secondary); font-size: 12px; display: block; margin-bottom: 4px;"
			>50-79%: 黄（Fair）</span
		>
		<ProgressBar value={65} />
	</div>
</Story>

<!-- 高い進捗（緑） -->
<Story name="ColorGood" args={{ value: 90 }}>
	<div style="width: 300px;">
		<span style="color: var(--text-secondary); font-size: 12px; display: block; margin-bottom: 4px;"
			>80-100%: 緑（Good）</span
		>
		<ProgressBar value={90} />
	</div>
</Story>

<!-- カスタム max 値 -->
<Story name="CustomMax" args={{ value: 3, max: 5 }}>
	<div style="width: 300px;">
		<p style="color: var(--text-secondary); font-size: 12px; margin-bottom: 8px;">
			タスク進捗: 3/5 完了
		</p>
		<ProgressBar value={3} max={5} />
	</div>
</Story>

<!-- インタラクティブ -->
<Story name="Interactive">
	<div style="width: 300px;">
		<ProgressBar value={50} />
	</div>
</Story>

<!-- アニメーション -->
<Story name="Animated">
	<div style="width: 300px;">
		<span
			style="color: var(--text-secondary); font-size: 12px; display: block; margin-bottom: 4px;"
		>
			アニメーション例（自動で値が変化）
		</span>
		<ProgressBar value={animatedValue} />
	</div>
</Story>

<!-- 全サイズ比較 -->
<Story name="AllSizes">
	<div style="display: flex; flex-direction: column; gap: 16px; width: 300px;">
		<div>
			<span
				style="color: var(--text-secondary); font-size: 12px; display: block; margin-bottom: 4px;"
				>Small (sm)</span
			>
			<ProgressBar value={75} size="sm" />
		</div>
		<div>
			<span
				style="color: var(--text-secondary); font-size: 12px; display: block; margin-bottom: 4px;"
				>Medium (md)</span
			>
			<ProgressBar value={75} size="md" />
		</div>
		<div>
			<span
				style="color: var(--text-secondary); font-size: 12px; display: block; margin-bottom: 4px;"
				>Large (lg)</span
			>
			<ProgressBar value={75} size="lg" />
		</div>
	</div>
</Story>

<!-- 全色比較 -->
<Story name="AllColors">
	<div style="display: flex; flex-direction: column; gap: 12px; width: 300px;">
		<div>
			<span
				style="color: var(--text-secondary); font-size: 12px; display: block; margin-bottom: 4px;"
				>0-49%: 赤（危険）</span
			>
			<ProgressBar value={25} />
		</div>
		<div>
			<span
				style="color: var(--text-secondary); font-size: 12px; display: block; margin-bottom: 4px;"
				>50-79%: 黄（注意）</span
			>
			<ProgressBar value={65} />
		</div>
		<div>
			<span
				style="color: var(--text-secondary); font-size: 12px; display: block; margin-bottom: 4px;"
				>80-100%: 緑（良好）</span
			>
			<ProgressBar value={90} />
		</div>
	</div>
</Story>
