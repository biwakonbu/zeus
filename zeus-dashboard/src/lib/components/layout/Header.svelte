<script lang="ts">
	import type { ConnectionState } from '$lib/types/api';
	import { ViewSwitcher, type ViewType } from '$lib/viewer';
	import { currentView, setView, usecaseViewState, graphViewState } from '$lib/stores/view';
	import { Icon } from '$lib/components/ui';

	interface Props {
		connectionState?: ConnectionState;
	}

	let { connectionState = 'disconnected' }: Props = $props();

	// 接続状態のラベル
	function getConnectionLabel(state: ConnectionState): string {
		switch (state) {
			case 'connected':
				return 'Connected';
			case 'connecting':
				return 'Connecting...';
			case 'disconnected':
				return 'Disconnected';
			default:
				return 'Unknown';
		}
	}

	// ビュー切り替えハンドラ
	function handleViewChange(view: ViewType) {
		setView(view);
	}

	// UseCase ビューのコントロール
	function handleZoomIn() {
		$usecaseViewState.onZoomIn?.();
	}

	function handleZoomOut() {
		$usecaseViewState.onZoomOut?.();
	}

	function handleZoomReset() {
		$usecaseViewState.onZoomReset?.();
	}

	function handleToggleListPanel() {
		$usecaseViewState.onToggleListPanel?.();
	}

	// Graph ビューのコントロール
	function handleGraphZoomIn() {
		$graphViewState.onZoomIn?.();
	}

	function handleGraphZoomOut() {
		$graphViewState.onZoomOut?.();
	}

	function handleGraphZoomReset() {
		$graphViewState.onZoomReset?.();
	}

	function handleGraphToggleCriticalPath() {
		$graphViewState.onToggleCriticalPath?.();
	}

	function handleGraphToggleFilterPanel() {
		$graphViewState.onToggleFilterPanel?.();
	}

	function handleGraphToggleLegend() {
		$graphViewState.onToggleLegend?.();
	}

	function handleGraphClearDependencyFilter() {
		$graphViewState.onClearDependencyFilter?.();
	}
</script>

<header class="header">
	<div class="header-content">
		<div class="logo">
			<span class="logo-icon">&#9889;</span>
			<h1 class="logo-text">ZEUS</h1>
			<span class="logo-subtitle">Dashboard</span>
		</div>

		<div class="header-center">
			<ViewSwitcher currentView={$currentView} onViewChange={handleViewChange} />
		</div>

		<div class="header-right">
			<!-- UseCase ビュー専用コントロール -->
			{#if $currentView === 'usecase'}
				<div class="usecase-controls">
					<!-- 情報バッジ -->
					<div class="info-badge">
						<Icon name="Target" size={14} />
						<span>{$usecaseViewState.boundary || 'System'}</span>
						<span class="badge-separator">|</span>
						<span>{$usecaseViewState.actorCount}A / {$usecaseViewState.usecaseCount}UC</span>
					</div>

					<!-- リストパネルトグル -->
					<button
						class="control-btn"
						class:active={$usecaseViewState.showListPanel}
						onclick={handleToggleListPanel}
						aria-label="リストパネル"
						title="リスト表示 (L)"
					>
						<Icon name="List" size={16} />
					</button>

					<div class="control-separator"></div>

					<!-- ズームコントロール -->
					<button
						class="control-btn"
						onclick={handleZoomOut}
						aria-label="ズームアウト"
						title="ズームアウト"
					>
						<Icon name="Minus" size={16} />
					</button>
					<span class="zoom-display">{Math.round($usecaseViewState.zoom * 100)}%</span>
					<button
						class="control-btn"
						onclick={handleZoomIn}
						aria-label="ズームイン"
						title="ズームイン"
					>
						<Icon name="Plus" size={16} />
					</button>
					<button
						class="control-btn"
						onclick={handleZoomReset}
						aria-label="リセット"
						title="ビューをリセット"
					>
						<Icon name="Maximize2" size={16} />
					</button>
				</div>
			{/if}

			<!-- Graph ビュー専用コントロール -->
			{#if $currentView === 'graph'}
				<div class="graph-controls">
					<!-- モードバッジ -->
					<div class="info-badge mode-badge" class:wbs-mode={$graphViewState.mode === 'wbs'}>
						<span>{$graphViewState.mode === 'wbs' ? 'WBS' : 'TASK'}</span>
						<span class="badge-separator">|</span>
						<span>{$graphViewState.visibleCount}/{$graphViewState.nodeCount}</span>
					</div>

					<!-- フィルターパネルトグル -->
					<button
						class="control-btn"
						class:active={$graphViewState.showFilterPanel}
						onclick={handleGraphToggleFilterPanel}
						aria-label="フィルターパネル"
						title="フィルター (F)"
					>
						<Icon name="Filter" size={16} />
					</button>

					<!-- 凡例トグル -->
					<button
						class="control-btn"
						class:active={$graphViewState.showLegend}
						onclick={handleGraphToggleLegend}
						aria-label="凡例"
						title="凡例 (G)"
					>
						<Icon name="Info" size={16} />
					</button>

					<!-- クリティカルパストグル -->
					<button
						class="control-btn"
						class:active={$graphViewState.showCriticalPath}
						onclick={handleGraphToggleCriticalPath}
						aria-label="クリティカルパス"
						title="クリティカルパス (C)"
					>
						<Icon name="Zap" size={16} />
					</button>

					<!-- 依存関係フィルタークリア（条件付き表示） -->
					{#if $graphViewState.hasDependencyFilter}
						<button
							class="control-btn filter-active"
							onclick={handleGraphClearDependencyFilter}
							aria-label="フィルター解除"
							title="依存関係フィルターを解除"
						>
							<Icon name="X" size={16} />
						</button>
					{/if}

					<div class="control-separator"></div>

					<!-- ズームコントロール -->
					<button
						class="control-btn"
						onclick={handleGraphZoomOut}
						aria-label="ズームアウト"
						title="ズームアウト"
					>
						<Icon name="Minus" size={16} />
					</button>
					<span class="zoom-display">{Math.round($graphViewState.zoom * 100)}%</span>
					<button
						class="control-btn"
						onclick={handleGraphZoomIn}
						aria-label="ズームイン"
						title="ズームイン"
					>
						<Icon name="Plus" size={16} />
					</button>
					<button
						class="control-btn"
						onclick={handleGraphZoomReset}
						aria-label="リセット"
						title="ビューをリセット"
					>
						<Icon name="Maximize2" size={16} />
					</button>
				</div>
			{/if}

			<!-- 接続状態 -->
			<div class="connection-status">
				<span class="connection-indicator {connectionState}"></span>
				<span class="connection-label">{getConnectionLabel(connectionState)}</span>
			</div>
		</div>
	</div>
</header>

<style>
	.header {
		background-color: var(--bg-secondary);
		border-bottom: 1px solid var(--border-metal);
		padding: var(--spacing-xs) var(--spacing-md);
		position: sticky;
		top: 0;
		z-index: 100;
	}

	.header-content {
		display: flex;
		align-items: center;
		justify-content: space-between;
		max-width: 1600px;
		margin: 0 auto;
		position: relative;
	}

	.logo {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
	}

	.logo-icon {
		font-size: 1.5rem;
		color: var(--accent-primary);
	}

	.logo-text {
		font-size: var(--font-size-xl);
		font-weight: 700;
		color: var(--accent-primary);
		letter-spacing: 0.1em;
		text-transform: uppercase;
	}

	.logo-subtitle {
		font-size: var(--font-size-sm);
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding-left: var(--spacing-sm);
		border-left: 1px solid var(--border-metal);
	}

	.header-center {
		position: absolute;
		left: 50%;
		transform: translateX(-50%);
	}

	.header-right {
		display: flex;
		align-items: center;
		gap: var(--spacing-md);
	}

	/* UseCase / Graph コントロール共通 */
	.usecase-controls,
	.graph-controls {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 4px 8px;
		background: var(--bg-panel);
		border: 1px solid var(--border-metal);
		border-radius: 6px;
	}

	.info-badge {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 0 8px;
		font-size: 0.75rem;
		color: var(--text-secondary);
	}

	.badge-separator {
		color: var(--border-metal);
	}

	.control-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		background: transparent;
		border: none;
		border-radius: 4px;
		color: var(--text-secondary);
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.control-btn:hover {
		background: rgba(255, 149, 51, 0.15);
		color: var(--accent-primary);
	}

	.control-btn.active {
		background: var(--accent-primary);
		color: var(--bg-primary);
	}

	.control-separator {
		width: 1px;
		height: 20px;
		background: var(--border-metal);
		margin: 0 4px;
	}

	.zoom-display {
		font-size: 0.75rem;
		color: var(--text-muted);
		min-width: 40px;
		text-align: center;
		font-variant-numeric: tabular-nums;
	}

	/* Graph ビュー用のモードバッジ */
	.mode-badge {
		font-weight: 600;
		letter-spacing: 0.05em;
	}

	.mode-badge.wbs-mode {
		background: rgba(255, 215, 0, 0.15);
		border-radius: 4px;
		color: #ffd700;
	}

	/* フィルターアクティブ状態 */
	.control-btn.filter-active {
		background: rgba(238, 68, 68, 0.2);
		color: #ee4444;
	}

	.control-btn.filter-active:hover {
		background: rgba(238, 68, 68, 0.3);
	}

	/* 接続状態 */
	.connection-status {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		padding: var(--spacing-xs) var(--spacing-sm);
		background-color: var(--bg-panel);
		border: 1px solid var(--border-metal);
		border-radius: var(--border-radius-sm);
	}

	.connection-label {
		font-size: var(--font-size-xs);
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
</style>
