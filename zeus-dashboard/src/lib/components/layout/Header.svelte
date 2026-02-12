<script lang="ts">
	import type { ConnectionState } from '$lib/types/api';
	import { ViewSwitcher, type ViewType } from '$lib/viewer';
	import {
		currentView,
		setView,
		usecaseViewState,
		graphViewState,
		activityViewState,
		visionViewState
	} from '$lib/stores/view';
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

	function handleGraphToggleListPanel() {
		$graphViewState.onToggleListPanel?.();
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

	// Activity ビューのコントロール
	function handleActivityZoomIn() {
		$activityViewState.onZoomIn?.();
	}

	function handleActivityZoomOut() {
		$activityViewState.onZoomOut?.();
	}

	function handleActivityZoomReset() {
		$activityViewState.onZoomReset?.();
	}

	function handleActivityToggleListPanel() {
		$activityViewState.onToggleListPanel?.();
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
			<!-- Vision ビュー専用コントロール -->
			{#if $currentView === 'vision'}
				<div class="vision-controls">
					<div class="info-badge">
						<Icon name="Eye" size={14} />
						<span>{$visionViewState.objectiveCount} objectives</span>
					</div>
				</div>
			{/if}

			<!-- UseCase ビュー専用コントロール -->
			{#if $currentView === 'usecase'}
				<div class="usecase-controls">
					<!-- Objective セレクター -->
					{#if $usecaseViewState.objectiveOptions.length > 0}
						<select
							class="objective-select"
							value={$usecaseViewState.selectedObjectiveId ?? ''}
							onchange={(e) => {
								const val = e.currentTarget.value;
								$usecaseViewState.onObjectiveChange?.(val || null);
							}}
						>
							<option value="">All Objectives</option>
							{#each $usecaseViewState.objectiveOptions as obj}
								<option value={obj.id}>{obj.title}</option>
							{/each}
						</select>
					{/if}

					<!-- 情報バッジ -->
					<div class="info-badge">
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
					<!-- ノード数バッジ -->
					<div class="info-badge">
						<span>{$graphViewState.visibleCount}/{$graphViewState.nodeCount}</span>
					</div>

					<!-- リストパネルトグル -->
					<button
						class="control-btn"
						class:active={$graphViewState.showListPanel}
						onclick={handleGraphToggleListPanel}
						aria-label="ノード一覧"
						title="ノード一覧 (L)"
					>
						<Icon name="List" size={16} />
					</button>

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

			<!-- Activity ビュー専用コントロール -->
			{#if $currentView === 'activity'}
				<div class="activity-controls">
					<!-- 情報バッジ -->
					<div class="info-badge">
						<Icon name="Workflow" size={14} />
						<span>{$activityViewState.activityCount} activities</span>
					</div>

					<!-- リストパネルトグル -->
					<button
						class="control-btn"
						class:active={$activityViewState.showListPanel}
						onclick={handleActivityToggleListPanel}
						aria-label="リストパネル"
						title="リスト表示 (L)"
					>
						<Icon name="List" size={16} />
					</button>

					<div class="control-separator"></div>

					<!-- ズームコントロール -->
					<button
						class="control-btn"
						onclick={handleActivityZoomOut}
						aria-label="ズームアウト"
						title="ズームアウト"
					>
						<Icon name="Minus" size={16} />
					</button>
					<span class="zoom-display">{Math.round($activityViewState.zoom * 100)}%</span>
					<button
						class="control-btn"
						onclick={handleActivityZoomIn}
						aria-label="ズームイン"
						title="ズームイン"
					>
						<Icon name="Plus" size={16} />
					</button>
					<button
						class="control-btn"
						onclick={handleActivityZoomReset}
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

	/* Vision / UseCase / Graph / Activity コントロール共通 */
	.vision-controls,
	.usecase-controls,
	.graph-controls,
	.activity-controls {
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

	/* フィルターアクティブ状態 */
	.control-btn.filter-active {
		background: rgba(238, 68, 68, 0.2);
		color: #ee4444;
	}

	.control-btn.filter-active:hover {
		background: rgba(238, 68, 68, 0.3);
	}

	/* Objective セレクター */
	.objective-select {
		appearance: none;
		background: var(--bg-secondary);
		border: 1px solid var(--border-metal);
		border-radius: 4px;
		color: var(--text-primary);
		font-size: 0.75rem;
		padding: 4px 24px 4px 8px;
		cursor: pointer;
		max-width: 200px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%23999' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 6px center;
	}

	.objective-select:hover {
		border-color: var(--accent-primary);
	}

	.objective-select:focus {
		outline: none;
		border-color: var(--accent-primary);
		box-shadow: 0 0 0 1px var(--accent-primary);
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
