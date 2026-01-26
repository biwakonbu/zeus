<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import FilterDropdown from './FilterDropdown.svelte';

	const { Story } = defineMeta({
		title: 'Viewer/UseCase/FilterDropdown',
		component: FilterDropdown,
		tags: ['autodocs'],
		argTypes: {
			selected: {
				control: 'text',
				description: '選択中のオプションID'
			},
			placeholder: {
				control: 'text',
				description: 'プレースホルダー（未選択時）'
			}
		}
	});
</script>

<script lang="ts">
	// Story は defineMeta から export される
	const actorOptions = [
		{ id: 'user', label: 'ユーザー' },
		{ id: 'admin', label: '管理者' },
		{ id: 'system', label: 'システム' },
		{ id: 'external', label: '外部サービス' }
	];

	const statusOptions = [
		{ id: 'draft', label: 'Draft' },
		{ id: 'active', label: 'Active' },
		{ id: 'deprecated', label: 'Deprecated' }
	];

	let selectedActor = $state<string | null>(null);
	let selectedStatus = $state<string | null>(null);

	function handleActorSelect(id: string | null) {
		selectedActor = id;
		console.log('Actor selected:', id);
	}

	function handleStatusSelect(id: string | null) {
		selectedStatus = id;
		console.log('Status selected:', id);
	}
</script>

<!-- デフォルト（未選択） -->
<Story name="Default">
	<div style="width: 250px; padding: 24px; background: var(--bg-primary);">
		<FilterDropdown
			options={actorOptions}
			selected={null}
			placeholder="関連 Actor"
			onSelect={handleActorSelect}
		/>
	</div>
</Story>

<!-- 選択済み -->
<Story name="Selected">
	<div style="width: 250px; padding: 24px; background: var(--bg-primary);">
		<FilterDropdown
			options={actorOptions}
			selected="admin"
			placeholder="関連 Actor"
			onSelect={handleActorSelect}
		/>
	</div>
</Story>

<!-- ステータスフィルタ -->
<Story name="StatusFilter">
	<div style="width: 200px; padding: 24px; background: var(--bg-primary);">
		<FilterDropdown
			options={statusOptions}
			selected={null}
			placeholder="ステータス"
			onSelect={handleStatusSelect}
		/>
	</div>
</Story>

<!-- 長いオプション -->
<Story name="LongOptions">
	{@const longOptions = [
		{ id: 'opt1', label: '非常に長いオプション名がここに入ります' },
		{ id: 'opt2', label: '中程度のオプション' },
		{ id: 'opt3', label: '短い' }
	]}
	<div style="width: 250px; padding: 24px; background: var(--bg-primary);">
		<FilterDropdown
			options={longOptions}
			selected={null}
			placeholder="選択してください"
			onSelect={handleActorSelect}
		/>
	</div>
</Story>

<!-- 複数のドロップダウン -->
<Story name="MultipleDropdowns">
	<div style="display: flex; gap: 16px; padding: 24px; background: var(--bg-primary);">
		<div style="width: 180px;">
			<FilterDropdown
				options={actorOptions}
				selected={selectedActor}
				placeholder="Actor"
				onSelect={handleActorSelect}
			/>
		</div>
		<div style="width: 150px;">
			<FilterDropdown
				options={statusOptions}
				selected={selectedStatus}
				placeholder="Status"
				onSelect={handleStatusSelect}
			/>
		</div>
	</div>
</Story>

<!-- インタラクティブ -->
<Story name="Interactive">
	<div style="width: 250px; padding: 24px; background: var(--bg-primary);">
		<FilterDropdown
			options={actorOptions}
			selected={selectedActor}
			placeholder="関連 Actor"
			onSelect={handleActorSelect}
		/>
		<p style="color: var(--text-muted); font-size: 12px; margin-top: 16px;">
			選択中: {selectedActor ?? '（なし）'}
		</p>
	</div>
</Story>
