<script lang="ts">
	// GroupedList - グループ化リストコンポーネント
	// Actor/UseCase をグループ化して表示
	import { Icon } from '$lib/components/ui';
	import type { ActorItem, UseCaseItem } from '$lib/types/api';
	import { getActorIcon, getStatusColor } from '../utils';

	// リストアイテム型
	type ActorListItem = ActorItem & { itemType: 'actor' };
	type UseCaseListItem = UseCaseItem & { itemType: 'usecase' };
	type ListItem = ActorListItem | UseCaseListItem;

	interface Props {
		items: ListItem[];
		groupBy: boolean;
		selectedId: string | null;
		actors?: ActorItem[];
		onSelect: (item: ListItem) => void;
	}
	let { items, groupBy, selectedId, actors = [], onSelect }: Props = $props();

	// グループ化（$derived.by で関数呼び出し不要）
	const grouped = $derived.by(() => {
		if (!groupBy) {
			return [{ key: 'all', label: '', items }];
		}
		const actorItems = items.filter((i): i is ActorListItem => i.itemType === 'actor');
		const usecaseItems = items.filter((i): i is UseCaseListItem => i.itemType === 'usecase');
		return [
			{ key: 'actor', label: 'Actor', items: actorItems },
			{ key: 'usecase', label: 'UseCase', items: usecaseItems }
		].filter((g) => g.items.length > 0);
	});

	// アクター名解決
	function getActorNames(actorRefs: UseCaseListItem['actors']): string {
		if (!actorRefs || actorRefs.length === 0) return '';
		const names = actorRefs
			.map((ref) => actors.find((a) => a.id === ref.actor_id)?.title ?? ref.actor_id)
			.join(', ');
		return `関連: ${names}`;
	}

	// 型ガード
	function isActor(item: ListItem): item is ActorListItem {
		return item.itemType === 'actor';
	}
</script>

<div class="grouped-list">
	{#each grouped as group}
		{#if group.label}
			<div class="group-header">
				<span class="group-label">{group.label}</span>
				<span class="group-count">({group.items.length})</span>
			</div>
		{/if}

		<ul class="list">
			{#each group.items as item}
				<li>
					<button
						class="list-item"
						class:selected={selectedId === item.id}
						onclick={() => onSelect(item)}
					>
						{#if isActor(item)}
							<!-- Actor アイテム -->
							<span class="item-icon">
								<Icon name={getActorIcon(item.type)} size={16} />
							</span>
							<div class="item-content">
								<span class="item-title">{item.title}</span>
								<span class="item-meta">{item.type}</span>
							</div>
						{:else}
							<!-- UseCase アイテム -->
							<span class="item-status" style="background: {getStatusColor(item.status)}"></span>
							<div class="item-content">
								<span class="item-title">{item.title}</span>
								{#if item.actors.length > 0}
									<span class="item-relations">{getActorNames(item.actors)}</span>
								{/if}
							</div>
						{/if}
					</button>
				</li>
			{/each}
		</ul>
	{/each}

	{#if items.length === 0}
		<div class="empty-state">
			<Icon name="Inbox" size={24} />
			<span>該当するアイテムがありません</span>
		</div>
	{/if}
</div>

<style>
	.grouped-list {
		display: flex;
		flex-direction: column;
		gap: 0;
		overflow-y: auto;
	}

	.group-header {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0.75rem;
		background: linear-gradient(
			180deg,
			rgba(50, 50, 50, 0.9) 0%,
			rgba(40, 40, 40, 0.8) 100%
		);
		border-bottom: 1px solid var(--border-metal);
		position: sticky;
		top: 0;
		z-index: 1;
	}

	.group-label {
		font-weight: 600;
		font-size: var(--font-size-sm);
		color: var(--accent-primary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		text-shadow: 0 0 8px rgba(255, 149, 51, 0.3);
	}

	.group-count {
		font-size: var(--font-size-xs);
		color: var(--text-muted);
	}

	.list {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.list-item {
		display: flex;
		align-items: flex-start;
		gap: 0.5rem;
		width: 100%;
		padding: 0.625rem 0.75rem;
		background: transparent;
		border: none;
		border-bottom: 1px solid var(--border-dark);
		border-left: 2px solid transparent;
		color: var(--text-primary);
		font-family: var(--font-family);
		cursor: pointer;
		text-align: left;
		transition:
			background-color var(--transition-select) ease-out,
			border-color var(--transition-select) ease-out,
			box-shadow var(--transition-select) ease-out;
	}

	.list-item:hover:not(.selected) {
		background: rgba(255, 149, 51, 0.08);
		border-left-color: var(--border-highlight);
	}

	.list-item.selected {
		background: var(--accent-primary);
		color: var(--bg-primary);
		border-left-color: var(--accent-primary);
		/* 選択時のグロー効果 */
		box-shadow: 0 0 12px rgba(255, 149, 51, 0.5);
	}

	.list-item.selected .item-meta,
	.list-item.selected .item-relations {
		color: var(--bg-primary);
		opacity: 0.85;
		background: rgba(0, 0, 0, 0.2);
	}

	.list-item:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: calc(-1 * var(--focus-ring-width));
		z-index: 1;
	}

	.item-icon {
		display: flex;
		align-items: center;
		flex-shrink: 0;
		margin-top: 0.125rem;
	}

	.item-status {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		flex-shrink: 0;
		margin-top: 0.375rem;
		/* ステータスドットのグロー */
		box-shadow: 0 0 6px currentColor;
	}

	.item-content {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 0.125rem;
	}

	.item-title {
		font-size: var(--font-size-sm);
		font-weight: 500;
		line-height: 1.3;
		/* 折り返し許可（省略禁止） */
		word-wrap: break-word;
		overflow-wrap: break-word;
	}

	.item-meta {
		font-size: var(--font-size-xs);
		color: var(--text-muted);
		background: rgba(0, 0, 0, 0.3);
		padding: 0.125rem 0.375rem;
		border-radius: 2px;
		width: fit-content;
		border: 1px solid var(--border-dark);
	}

	.item-relations {
		font-size: var(--font-size-xs);
		color: var(--text-muted);
		line-height: 1.3;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		padding: 2rem;
		color: var(--text-muted);
		text-align: center;
	}

	.empty-state span {
		font-size: var(--font-size-sm);
	}

	/* アニメーション対応 */
	@media (prefers-reduced-motion: reduce) {
		.list-item {
			transition: none;
		}
	}
</style>
