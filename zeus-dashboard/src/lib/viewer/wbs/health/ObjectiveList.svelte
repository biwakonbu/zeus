<script lang="ts">
	// Objective 階層リスト
	// 折りたたみ可能な Objective → Deliverable のリスト表示
	import ProgressBar from '../shared/ProgressBar.svelte';
	import type { ProgressNode } from '$lib/types/api';

	interface Props {
		objectives: ProgressNode[];
		selectedId: string | null;
		expandedIds: Set<string>;
		onSelect: (id: string, type: string) => void;
		onToggle: (id: string) => void;
	}
	let { objectives, selectedId, expandedIds, onSelect, onToggle }: Props = $props();

	function handleKeydown(event: KeyboardEvent, id: string, type: string) {
		if (event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			onSelect(id, type);
		}
	}

	function handleToggleKeydown(event: KeyboardEvent, id: string) {
		if (event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			onToggle(id);
		}
	}
</script>

<div class="objective-list">
	{#each objectives as obj (obj.id)}
		<div class="objective-item" class:selected={selectedId === obj.id}>
			<button
				class="toggle-btn"
				class:expanded={expandedIds.has(obj.id)}
				onclick={() => onToggle(obj.id)}
				onkeydown={(e) => handleToggleKeydown(e, obj.id)}
				aria-label={expandedIds.has(obj.id) ? '折りたたむ' : '展開する'}
				aria-expanded={expandedIds.has(obj.id)}
			>
				<span class="toggle-icon">{expandedIds.has(obj.id) ? '▼' : '▶'}</span>
			</button>
			<button
				class="info"
				onclick={() => onSelect(obj.id, 'objective')}
				onkeydown={(e) => handleKeydown(e, obj.id, 'objective')}
			>
				<span class="obj-id">{obj.id}</span>
				<span class="title">{obj.title}</span>
				<div class="progress-wrapper">
					<ProgressBar progress={obj.progress} size="sm" />
					<span class="progress-label">{obj.progress}%</span>
				</div>
			</button>
		</div>

		{#if expandedIds.has(obj.id) && obj.children && obj.children.length > 0}
			<div class="children">
				{#each obj.children as child (child.id)}
					<button
						class="deliverable-item"
						class:selected={selectedId === child.id}
						onclick={() => onSelect(child.id, 'deliverable')}
						onkeydown={(e) => handleKeydown(e, child.id, 'deliverable')}
					>
						<span class="branch">├</span>
						<span class="del-id">{child.id}</span>
						<span class="title">{child.title}</span>
						<div class="progress-wrapper">
							<ProgressBar progress={child.progress} size="sm" />
							<span class="progress-label">{child.progress}%</span>
						</div>
					</button>
				{/each}
			</div>
		{/if}
	{/each}

	{#if objectives.length === 0}
		<div class="empty-state">
			<span class="empty-icon">📋</span>
			<span class="empty-text">Objective がありません</span>
		</div>
	{/if}
</div>

<style>
	.objective-list {
		flex: 1;
		overflow-y: auto;
		padding: 8px 0;
	}

	.objective-item {
		display: flex;
		align-items: center;
		border-bottom: 1px solid var(--border-dark, #333333);
		transition: background-color 0.15s ease;
	}

	.objective-item:hover {
		background-color: var(--bg-hover, #3a3a3a);
	}

	.objective-item.selected {
		background-color: var(--bg-secondary, #242424);
		border-left: 3px solid var(--accent-primary, #ff9533);
	}

	.toggle-btn {
		width: 32px;
		height: 40px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: transparent;
		border: none;
		cursor: pointer;
		color: var(--text-muted, #888888);
		transition: color 0.15s ease;
	}

	.toggle-btn:hover {
		color: var(--accent-primary, #ff9533);
	}

	.toggle-icon {
		font-size: 10px;
		transition: transform 0.15s ease;
	}

	.toggle-btn.expanded .toggle-icon {
		transform: rotate(0deg);
	}

	.info {
		flex: 1;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 16px 10px 0;
		background: transparent;
		border: none;
		cursor: pointer;
		text-align: left;
		color: inherit;
	}

	.obj-id,
	.del-id {
		font-size: 11px;
		font-weight: 500;
		color: var(--text-muted, #888888);
		min-width: 60px;
	}

	.title {
		flex: 1;
		font-size: 13px;
		color: var(--text-primary, #ffffff);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.progress-wrapper {
		display: flex;
		align-items: center;
		gap: 8px;
		min-width: 120px;
	}

	.progress-wrapper :global(.progress-bar) {
		flex: 1;
	}

	.progress-label {
		font-size: 11px;
		font-weight: 500;
		color: var(--text-secondary, #b8b8b8);
		min-width: 36px;
		text-align: right;
	}

	/* 子要素（Deliverable） */
	.children {
		background-color: rgba(0, 0, 0, 0.1);
	}

	.deliverable-item {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 8px 16px 8px 24px;
		border-bottom: 1px solid var(--border-dark, #333333);
		background: transparent;
		border: none;
		border-bottom: 1px solid var(--border-dark, #333333);
		cursor: pointer;
		width: 100%;
		text-align: left;
		color: inherit;
		transition: background-color 0.15s ease;
	}

	.deliverable-item:hover {
		background-color: var(--bg-hover, #3a3a3a);
	}

	.deliverable-item.selected {
		background-color: var(--bg-secondary, #242424);
		border-left: 3px solid var(--accent-primary, #ff9533);
		padding-left: 21px;
	}

	.branch {
		color: var(--border-metal, #4a4a4a);
		font-family: monospace;
	}

	/* 空状態 */
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 48px 16px;
		color: var(--text-muted, #888888);
	}

	.empty-icon {
		font-size: 32px;
		opacity: 0.5;
		margin-bottom: 8px;
	}

	.empty-text {
		font-size: 13px;
	}
</style>
