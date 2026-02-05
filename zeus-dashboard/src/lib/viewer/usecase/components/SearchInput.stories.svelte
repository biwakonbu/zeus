<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import SearchInput from './SearchInput.svelte';

	const { Story } = defineMeta({
		title: 'Viewer/UseCase/SearchInput',
		component: SearchInput,
		tags: ['autodocs'],
		argTypes: {
			value: {
				control: 'text',
				description: '検索文字列'
			},
			placeholder: {
				control: 'text',
				description: 'プレースホルダーテキスト'
			}
		}
	});
</script>

<script lang="ts">
	// Story は defineMeta から export される
	let searchValue = $state('');

	function handleInput(value: string) {
		searchValue = value;
		console.log('Search input:', value);
	}

	function handleClear() {
		console.log('Search cleared');
	}
</script>

<!-- デフォルト（空） -->
<Story name="Default">
	<div style="width: 300px; padding: 24px; background: var(--bg-primary);">
		<SearchInput value="" placeholder="検索..." onInput={handleInput} onClear={handleClear} />
	</div>
</Story>

<!-- 入力値あり -->
<Story name="WithValue">
	<div style="width: 300px; padding: 24px; background: var(--bg-primary);">
		<SearchInput
			value="ユースケース"
			placeholder="検索..."
			onInput={handleInput}
			onClear={handleClear}
		/>
	</div>
</Story>

<!-- カスタムプレースホルダー -->
<Story name="CustomPlaceholder">
	<div style="width: 300px; padding: 24px; background: var(--bg-primary);">
		<SearchInput
			value=""
			placeholder="Actor / UseCase を検索..."
			onInput={handleInput}
			onClear={handleClear}
		/>
	</div>
</Story>

<!-- 幅バリエーション -->
<Story name="WidthVariations">
	<div
		style="display: flex; flex-direction: column; gap: 16px; padding: 24px; background: var(--bg-primary);"
	>
		<div style="width: 200px;">
			<p style="color: var(--text-muted); font-size: 10px; margin-bottom: 4px;">200px</p>
			<SearchInput value="" placeholder="検索..." onInput={handleInput} />
		</div>
		<div style="width: 300px;">
			<p style="color: var(--text-muted); font-size: 10px; margin-bottom: 4px;">300px</p>
			<SearchInput value="" placeholder="検索..." onInput={handleInput} />
		</div>
		<div style="width: 100%;">
			<p style="color: var(--text-muted); font-size: 10px; margin-bottom: 4px;">100%</p>
			<SearchInput value="" placeholder="検索..." onInput={handleInput} />
		</div>
	</div>
</Story>

<!-- インタラクティブ -->
<Story name="Interactive">
	<div style="width: 300px; padding: 24px; background: var(--bg-primary);">
		<SearchInput
			value={searchValue}
			placeholder="入力してみてください..."
			onInput={handleInput}
			onClear={handleClear}
		/>
		<p style="color: var(--text-muted); font-size: 12px; margin-top: 12px;">
			入力値: "{searchValue}"
		</p>
	</div>
</Story>
