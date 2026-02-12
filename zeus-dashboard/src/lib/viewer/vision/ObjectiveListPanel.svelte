<script lang="ts">
	import type { Objective } from '$lib/types/api';
	import { getObjectiveStatusColor, getObjectiveStatusLabel } from '$lib/utils/status';

	interface Props {
		objectives: Objective[];
		selectedId: string | null;
		onSelect: (id: string) => void;
	}

	let { objectives, selectedId, onSelect }: Props = $props();

	// 検索
	let searchQuery = $state('');

	const filteredObjectives = $derived.by(() => {
		if (!searchQuery) return objectives;
		const q = searchQuery.toLowerCase();
		return objectives.filter(
			(obj) =>
				obj.title.toLowerCase().includes(q) ||
				obj.id.toLowerCase().includes(q) ||
				(obj.description && obj.description.toLowerCase().includes(q))
		);
	});
</script>

<div class="objective-list">
	<!-- 検索バー -->
	<div class="search-bar">
		<input
			type="text"
			class="search-input"
			placeholder="Objective を検索..."
			aria-label="Objective を検索"
			bind:value={searchQuery}
		/>
	</div>

	<!-- 一覧 -->
	<div class="list-items">
		{#each filteredObjectives as obj (obj.id)}
			<button
				class="objective-card"
				class:selected={selectedId === obj.id}
				onclick={() => onSelect(obj.id)}
			>
				<div class="card-header">
					<span class="card-id">{obj.id}</span>
					<span class="status-dot" style="background: {getObjectiveStatusColor(obj.status)}"
						title={getObjectiveStatusLabel(obj.status)}
						role="img"
						aria-label="Status: {getObjectiveStatusLabel(obj.status)}"
					></span>
				</div>
				<div class="card-title">{obj.title}</div>
				<div class="card-footer">
					{#if obj.usecase_count > 0}
						<span class="usecase-badge">{obj.usecase_count} UC</span>
					{/if}
					{#if obj.tags && obj.tags.length > 0}
						{#each obj.tags.slice(0, 2) as tag (tag)}
							<span class="tag">{tag}</span>
						{/each}
					{/if}
				</div>
			</button>
		{/each}

		{#if filteredObjectives.length === 0}
			<div class="empty-message">
				{#if searchQuery}
					該当する Objective がありません
				{:else}
					Objective が定義されていません
				{/if}
			</div>
		{/if}
	</div>
</div>

<style>
	.objective-list {
		display: flex;
		flex-direction: column;
		height: 100%;
	}

	.search-bar {
		padding: 8px 12px;
		border-bottom: 1px solid var(--border-metal);
	}

	.search-input {
		width: 100%;
		padding: 6px 10px;
		background: var(--bg-primary);
		border: 1px solid var(--border-metal);
		border-radius: 4px;
		color: var(--text-primary);
		font-size: 0.8125rem;
		font-family: inherit;
		outline: none;
	}

	.search-input:focus {
		border-color: var(--accent-primary);
	}

	.search-input::placeholder {
		color: var(--text-muted);
	}

	.list-items {
		flex: 1;
		overflow-y: auto;
		padding: 8px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.objective-card {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 10px 12px;
		background: var(--bg-primary);
		border: 1px solid var(--border-metal);
		border-radius: 6px;
		cursor: pointer;
		text-align: left;
		font-family: inherit;
		color: var(--text-primary);
		transition: all 0.15s ease;
		width: 100%;
	}

	.objective-card:hover {
		border-color: var(--accent-primary);
		background: rgba(245, 158, 11, 0.05);
	}

	.objective-card.selected {
		border-color: var(--accent-primary);
		background: rgba(245, 158, 11, 0.1);
	}

	.card-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.card-id {
		font-family: monospace;
		font-size: 0.6875rem;
		color: var(--text-muted);
	}

	.status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.card-title {
		font-size: 0.8125rem;
		font-weight: 600;
		line-height: 1.3;
	}

	.card-footer {
		display: flex;
		align-items: center;
		gap: 6px;
		flex-wrap: wrap;
	}

	.usecase-badge {
		font-size: 0.625rem;
		font-weight: 600;
		padding: 1px 6px;
		background: rgba(245, 158, 11, 0.15);
		color: var(--accent-primary);
		border-radius: 8px;
	}

	.tag {
		font-size: 0.625rem;
		padding: 1px 6px;
		background: var(--bg-secondary);
		color: var(--text-muted);
		border-radius: 8px;
	}

	.empty-message {
		padding: 24px;
		text-align: center;
		font-size: 0.8125rem;
		color: var(--text-muted);
	}
</style>
