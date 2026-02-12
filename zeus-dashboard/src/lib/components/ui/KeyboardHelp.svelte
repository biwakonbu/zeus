<script lang="ts">
	// キーボードショートカットヘルプモーダル
	import { Icon } from '$lib/components/ui';
	import { shortcutsList, formatShortcutKey } from '$lib/stores/keyboard';
	import { currentView } from '$lib/stores/view';
	import type { ViewType } from '$lib/viewer';

	interface Props {
		onClose: () => void;
	}

	let { onClose }: Props = $props();

	// ビューごとの操作ヒント定義
	const viewHints: Record<ViewType, { description: string; key: string }[]> = {
		vision: [
			{ description: 'Objective 選択', key: 'Click' },
			{ description: '選択解除', key: 'Esc' }
		],
		graph: [
			{ description: 'ズーム', key: 'Scroll' },
			{ description: 'パン（移動）', key: 'Shift+Drag' },
			{ description: 'チェーン選択', key: 'Shift+Click' },
			{ description: '依存関係フィルター', key: 'Alt+Click / Right-Click' }
		],
		usecase: [
			{ description: 'ズーム', key: 'Scroll' },
			{ description: 'パン（移動）', key: 'Drag' },
			{ description: 'Actor/UseCase 選択', key: 'Click' },
			{ description: '関連エンティティ表示', key: 'Click (Filter Mode)' }
		],
		activity: [
			{ description: 'ズーム', key: 'Scroll' },
			{ description: 'パン（移動）', key: 'Drag' },
			{ description: 'ノード選択', key: 'Click' }
		]
	};

	// 現在のビューのラベル
	const viewLabels: Record<ViewType, string> = {
		vision: 'Vision View',
		graph: 'Graph View',
		usecase: 'UseCase View',
		activity: 'Activity View'
	};

	// 現在のビューのヒント
	const currentViewHints = $derived(viewHints[$currentView] || []);
	const currentViewLabel = $derived(viewLabels[$currentView] || 'View');

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			onClose();
		}
	}

	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			onClose();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div class="keyboard-help-backdrop" onclick={handleBackdropClick}>
	<div class="keyboard-help" role="dialog" aria-modal="true" aria-labelledby="keyboard-help-title">
		<div class="help-header">
			<h2 id="keyboard-help-title" class="help-title">
				<Icon name="Keyboard" size={20} />
				キーボードショートカット
			</h2>
			<button class="close-button" onclick={onClose} aria-label="閉じる">
				<Icon name="X" size={16} />
			</button>
		</div>

		<div class="help-content">
			{#each Object.entries($shortcutsList) as [category, shortcuts] (category)}
				<div class="shortcut-category">
					<h3 class="category-title">{category}</h3>
					<div class="shortcut-list">
						{#each shortcuts as shortcut (shortcut.key + (shortcut.modifiers?.join('') ?? ''))}
							<div class="shortcut-item">
								<span class="shortcut-description">{shortcut.description}</span>
								<kbd class="shortcut-key">
									{formatShortcutKey(shortcut.key, shortcut.modifiers)}
								</kbd>
							</div>
						{/each}
					</div>
				</div>
			{/each}

			{#if Object.keys($shortcutsList).length === 0}
				<div class="no-shortcuts">
					<Icon name="Info" size={24} />
					<p>登録されたショートカットはありません</p>
				</div>
			{/if}

			<!-- デフォルトショートカットの説明 -->
			<div class="shortcut-category">
				<h3 class="category-title">共通</h3>
				<div class="shortcut-list">
					<div class="shortcut-item">
						<span class="shortcut-description">ヘルプを表示</span>
						<kbd class="shortcut-key">?</kbd>
					</div>
					<div class="shortcut-item">
						<span class="shortcut-description">検索にフォーカス</span>
						<kbd class="shortcut-key">/</kbd>
					</div>
					<div class="shortcut-item">
						<span class="shortcut-description">閉じる / キャンセル</span>
						<kbd class="shortcut-key">Esc</kbd>
					</div>
				</div>
			</div>

			<!-- 現在のビューのマウス操作 -->
			{#if currentViewHints.length > 0}
				<div class="shortcut-category">
					<h3 class="category-title">{currentViewLabel} - マウス操作</h3>
					<div class="shortcut-list">
						{#each currentViewHints as hint (hint.key)}
							<div class="shortcut-item">
								<span class="shortcut-description">{hint.description}</span>
								<kbd class="shortcut-key">{hint.key}</kbd>
							</div>
						{/each}
					</div>
				</div>
			{/if}
		</div>

		<div class="help-footer">
			<span class="footer-hint">
				<kbd>?</kbd> でこのヘルプを表示
			</span>
		</div>
	</div>
</div>

<style>
	.keyboard-help-backdrop {
		position: fixed;
		inset: 0;
		z-index: 9999;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(0, 0, 0, 0.7);
		backdrop-filter: blur(2px);
		animation: backdrop-fade 0.15s ease-out;
	}

	@keyframes backdrop-fade {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	.keyboard-help {
		width: 90%;
		max-width: 500px;
		max-height: 80vh;
		display: flex;
		flex-direction: column;
		background: var(--bg-panel, #2a2a2a);
		border: 2px solid var(--border-metal, #3a3a3a);
		border-radius: var(--border-radius-md, 8px);
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
		animation: modal-enter 0.2s ease-out;
	}

	@keyframes modal-enter {
		from {
			opacity: 0;
			transform: scale(0.95) translateY(-10px);
		}
		to {
			opacity: 1;
			transform: scale(1) translateY(0);
		}
	}

	.help-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--spacing-md, 16px) var(--spacing-lg, 24px);
		border-bottom: 1px solid var(--border-dark, #333);
	}

	.help-title {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm, 8px);
		margin: 0;
		font-size: var(--font-size-lg, 18px);
		font-weight: 600;
		color: var(--accent-primary, #f59e0b);
	}

	.close-button {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: var(--spacing-xs, 4px);
		background: transparent;
		border: none;
		color: var(--text-muted, #888);
		cursor: pointer;
		border-radius: var(--border-radius-sm, 4px);
		transition:
			color 0.15s ease,
			background-color 0.15s ease;
	}

	.close-button:hover {
		color: var(--text-primary, #e0e0e0);
		background: var(--bg-hover, #3a3a3a);
	}

	.close-button:focus-visible {
		outline: var(--focus-ring-width, 2px) solid var(--focus-ring-color, #f59e0b);
		outline-offset: 1px;
	}

	.help-content {
		flex: 1;
		overflow-y: auto;
		padding: var(--spacing-md, 16px) var(--spacing-lg, 24px);
	}

	.shortcut-category {
		margin-bottom: var(--spacing-lg, 24px);
	}

	.shortcut-category:last-child {
		margin-bottom: 0;
	}

	.category-title {
		margin: 0 0 var(--spacing-sm, 8px) 0;
		font-size: var(--font-size-sm, 13px);
		font-weight: 600;
		color: var(--text-muted, #888);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.shortcut-list {
		display: flex;
		flex-direction: column;
		gap: var(--spacing-xs, 4px);
	}

	.shortcut-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--spacing-xs, 4px) var(--spacing-sm, 8px);
		border-radius: var(--border-radius-sm, 4px);
	}

	.shortcut-item:hover {
		background: var(--bg-hover, #333);
	}

	.shortcut-description {
		font-size: var(--font-size-sm, 13px);
		color: var(--text-primary, #e0e0e0);
	}

	.shortcut-key {
		padding: 2px 8px;
		background: var(--bg-secondary, #252525);
		border: 1px solid var(--border-metal, #3a3a3a);
		border-radius: 3px;
		font-family: var(--font-mono, 'IBM Plex Mono', monospace);
		font-size: var(--font-size-xs, 11px);
		color: var(--text-secondary, #ccc);
		white-space: nowrap;
	}

	.no-shortcuts {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--spacing-sm, 8px);
		padding: var(--spacing-lg, 24px);
		color: var(--text-muted, #888);
		text-align: center;
	}

	.help-footer {
		padding: var(--spacing-sm, 8px) var(--spacing-lg, 24px);
		border-top: 1px solid var(--border-dark, #333);
		text-align: center;
	}

	.footer-hint {
		font-size: var(--font-size-xs, 11px);
		color: var(--text-muted, #666);
	}

	.footer-hint kbd {
		padding: 1px 4px;
		background: var(--bg-secondary, #252525);
		border: 1px solid var(--border-metal, #3a3a3a);
		border-radius: 2px;
		font-family: var(--font-mono, 'IBM Plex Mono', monospace);
		font-size: var(--font-size-xs, 11px);
	}

	@media (prefers-reduced-motion: reduce) {
		.keyboard-help-backdrop {
			animation: none;
		}

		.keyboard-help {
			animation: none;
		}

		.close-button {
			transition: none;
		}
	}
</style>
