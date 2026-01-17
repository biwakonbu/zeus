<script lang="ts">
	import { onMount } from 'svelte';
	import { fetchTimeline } from '$lib/api/client';
	import type { TimelineResponse, TimelineItem, TimelineStats, TaskStatus, Priority } from '$lib/types/api';

	// Props
	interface Props {
		onTaskSelect?: (task: TimelineItem | null) => void;
	}
	let { onTaskSelect }: Props = $props();

	// 状態
	let timelineData: TimelineResponse | null = $state(null);
	let loading = $state(true);
	let error: string | null = $state(null);
	let selectedTaskId: string | null = $state(null);
	let hoveredTaskId: string | null = $state(null);
	let searchQuery = $state('');
	let statusFilter: TaskStatus | 'all' = $state('all');
	let showCriticalPathOnly = $state(false);
	let zoomLevel = $state(1); // 1 = 日単位、2 = 週単位

	// タイムスケール設定
	let dayWidth = $derived(zoomLevel === 1 ? 30 : 5);
	let timeUnit = $derived(zoomLevel === 1 ? 'day' : 'week');

	// フィルターされたアイテム
	let filteredItems = $derived.by(() => {
		if (!timelineData) return [];
		return timelineData.items.filter((item) => {
			const matchesSearch =
				!searchQuery ||
				item.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
				item.task_id.toLowerCase().includes(searchQuery.toLowerCase());
			const matchesStatus = statusFilter === 'all' || item.status === statusFilter;
			const matchesCriticalPath = !showCriticalPathOnly || item.is_on_critical_path;
			return matchesSearch && matchesStatus && matchesCriticalPath;
		});
	});

	// プロジェクト期間の計算
	let projectDays = $derived.by(() => {
		if (!timelineData || !timelineData.project_start || !timelineData.project_end) return 30;
		const start = new Date(timelineData.project_start);
		const end = new Date(timelineData.project_end);
		const diffTime = Math.abs(end.getTime() - start.getTime());
		const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
		return Math.max(diffDays, 7);
	});

	// タイムラインの日付ラベル
	let dateLabels = $derived.by(() => {
		if (!timelineData || !timelineData.project_start) return [];
		const start = new Date(timelineData.project_start);
		const labels: { date: Date; label: string; isWeekStart: boolean }[] = [];

		for (let i = 0; i <= projectDays; i += (zoomLevel === 1 ? 1 : 7)) {
			const date = new Date(start);
			date.setDate(start.getDate() + i);
			const isWeekStart = date.getDay() === 1 || i === 0;
			labels.push({
				date,
				label: zoomLevel === 1
					? `${date.getMonth() + 1}/${date.getDate()}`
					: `${date.getMonth() + 1}/${date.getDate()}`,
				isWeekStart
			});
		}
		return labels;
	});

	// データ読み込み
	async function loadData() {
		loading = true;
		error = null;
		try {
			timelineData = await fetchTimeline();
		} catch (e) {
			error = e instanceof Error ? e.message : 'タイムラインデータの読み込みに失敗しました';
		} finally {
			loading = false;
		}
	}

	// タスクバーの位置計算
	function getBarPosition(item: TimelineItem): { left: number; width: number } {
		if (!timelineData || !timelineData.project_start || !item.start_date || !item.end_date) {
			return { left: 0, width: dayWidth * 3 };
		}

		const projectStart = new Date(timelineData.project_start);
		const taskStart = new Date(item.start_date);
		const taskEnd = new Date(item.end_date);

		const startOffset = Math.floor((taskStart.getTime() - projectStart.getTime()) / (1000 * 60 * 60 * 24));
		const duration = Math.max(1, Math.ceil((taskEnd.getTime() - taskStart.getTime()) / (1000 * 60 * 60 * 24)));

		return {
			left: startOffset * dayWidth,
			width: duration * dayWidth
		};
	}

	// タスク選択
	function selectTask(task: TimelineItem) {
		selectedTaskId = task.task_id;
		onTaskSelect?.(task);
	}

	// ステータスに応じたカラー
	function getStatusColor(status: TaskStatus): string {
		switch (status) {
			case 'completed':
				return '#22c55e';
			case 'in_progress':
				return '#f59e0b';
			case 'blocked':
				return '#ef4444';
			case 'pending':
			default:
				return '#6b7280';
		}
	}

	// 優先度に応じたカラー
	function getPriorityColor(priority: Priority): string {
		switch (priority) {
			case 'high':
				return '#ef4444';
			case 'medium':
				return '#f59e0b';
			case 'low':
			default:
				return '#22c55e';
		}
	}

	// 今日の位置（プロジェクト開始からの日数）
	let todayOffset = $derived.by(() => {
		if (!timelineData || !timelineData.project_start) return -1;
		const projectStart = new Date(timelineData.project_start);
		const today = new Date();
		const diffDays = Math.floor((today.getTime() - projectStart.getTime()) / (1000 * 60 * 60 * 24));
		return diffDays;
	});

	// 日付フォーマット
	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		const date = new Date(dateStr);
		return `${date.getFullYear()}/${date.getMonth() + 1}/${date.getDate()}`;
	}

	onMount(() => {
		loadData();
	});
</script>

<div class="timeline-viewer">
	<!-- ヘッダー -->
	<div class="timeline-header">
		<div class="timeline-title">
			<h2>Timeline / Gantt View</h2>
			{#if timelineData}
				<span class="timeline-stats">
					{timelineData.stats.total_tasks} tasks |
					{timelineData.stats.on_critical_path} on critical path |
					{formatDate(timelineData.project_start)} - {formatDate(timelineData.project_end)}
				</span>
			{/if}
		</div>
		<div class="timeline-controls">
			<button
				class="timeline-btn"
				class:active={showCriticalPathOnly}
				onclick={() => (showCriticalPathOnly = !showCriticalPathOnly)}
				title="クリティカルパスのみ表示"
			>
				<span class="icon">🔥</span>
				<span class="btn-label">Critical</span>
			</button>
			<button
				class="timeline-btn"
				onclick={() => (zoomLevel = zoomLevel === 1 ? 2 : 1)}
				title="ズーム切り替え"
			>
				<span class="icon">{zoomLevel === 1 ? '🔍-' : '🔍+'}</span>
				<span class="btn-label">{zoomLevel === 1 ? 'Week' : 'Day'}</span>
			</button>
			<button class="timeline-btn" onclick={() => loadData()} title="更新">
				<span class="icon">↻</span>
			</button>
		</div>
	</div>

	<!-- フィルター -->
	<div class="timeline-filters">
		<input
			type="text"
			class="timeline-search"
			placeholder="検索..."
			bind:value={searchQuery}
		/>
		<select class="timeline-select" bind:value={statusFilter}>
			<option value="all">全ステータス</option>
			<option value="pending">未着手</option>
			<option value="in_progress">進行中</option>
			<option value="completed">完了</option>
			<option value="blocked">ブロック</option>
		</select>
	</div>

	<!-- タイムラインコンテンツ -->
	<div class="timeline-content">
		{#if loading}
			<div class="timeline-loading">
				<div class="spinner"></div>
				<span>読み込み中...</span>
			</div>
		{:else if error}
			<div class="timeline-error">
				<span class="error-icon">⚠</span>
				<span>{error}</span>
				<button class="timeline-btn retry-btn" onclick={() => loadData()}>再試行</button>
			</div>
		{:else if filteredItems.length === 0}
			<div class="timeline-empty">
				<div class="empty-icon">📅</div>
				<h3 class="empty-title">タイムラインに表示するタスクがありません</h3>
				<p class="empty-description">
					タスクをタイムラインに表示するには、開始日と期限日を設定してください
				</p>
				<div class="empty-guide">
					<p class="guide-label">タスクに日付を設定する方法:</p>
					<div class="guide-code">
						<code>zeus add task "タスク名" --start 2026-01-20 --due 2026-01-31</code>
					</div>
					<p class="guide-hint">
						既存タスクの日付更新については <code>zeus help</code> を参照してください
					</p>
				</div>
			</div>
		{:else}
			<!-- ガントチャート -->
			<div class="gantt-container">
				<!-- タスク名列 -->
				<div class="task-list">
					<div class="task-list-header">TASK</div>
					{#each filteredItems as item}
						{@const isSelected = selectedTaskId === item.task_id}
						{@const isHovered = hoveredTaskId === item.task_id}
						<div
							class="task-row"
							class:selected={isSelected}
							class:hovered={isHovered}
							class:critical={item.is_on_critical_path}
							onclick={() => selectTask(item)}
							onmouseenter={() => (hoveredTaskId = item.task_id)}
							onmouseleave={() => (hoveredTaskId = null)}
							role="button"
							tabindex="0"
							onkeydown={(e) => e.key === 'Enter' && selectTask(item)}
						>
							<span class="task-status" style="color: {getStatusColor(item.status)}">●</span>
							<span class="task-title">{item.title}</span>
							{#if item.is_on_critical_path}
								<span class="critical-badge" title="クリティカルパス">🔥</span>
							{/if}
						</div>
					{/each}
				</div>

				<!-- タイムライングリッド -->
				<div class="timeline-grid-wrapper">
					<!-- ヘッダー（日付） -->
					<div class="timeline-grid-header" style="width: {projectDays * dayWidth}px">
						{#each dateLabels as label, i}
							<div
								class="date-cell"
								class:week-start={label.isWeekStart}
								style="left: {i * (zoomLevel === 1 ? dayWidth : dayWidth * 7)}px; width: {zoomLevel === 1 ? dayWidth : dayWidth * 7}px"
							>
								{label.label}
							</div>
						{/each}
					</div>

					<!-- グリッドとバー -->
					<div class="timeline-grid" style="width: {projectDays * dayWidth}px">
						<!-- 今日の線 -->
						{#if todayOffset >= 0 && todayOffset <= projectDays}
							<div class="today-line" style="left: {todayOffset * dayWidth}px"></div>
						{/if}

						<!-- グリッド線（週の区切り） -->
						{#each dateLabels as label, i}
							{#if label.isWeekStart}
								<div
									class="grid-line"
									style="left: {i * (zoomLevel === 1 ? dayWidth : dayWidth * 7)}px"
								></div>
							{/if}
						{/each}

						<!-- タスクバー -->
						{#each filteredItems as item}
							{@const pos = getBarPosition(item)}
							{@const isSelected = selectedTaskId === item.task_id}
							{@const isHovered = hoveredTaskId === item.task_id}
							<div class="task-bar-row">
								<div
									class="task-bar"
									class:selected={isSelected}
									class:hovered={isHovered}
									class:critical={item.is_on_critical_path}
									style="
										left: {pos.left}px;
										width: {pos.width}px;
										background-color: {getStatusColor(item.status)};
									"
									onclick={() => selectTask(item)}
									onmouseenter={() => (hoveredTaskId = item.task_id)}
									onmouseleave={() => (hoveredTaskId = null)}
									role="button"
									tabindex="0"
									onkeydown={(e) => e.key === 'Enter' && selectTask(item)}
									title="{item.title}
{formatDate(item.start_date)} - {formatDate(item.end_date)}
進捗: {item.progress}%
スラック: {item.slack}日"
								>
									<!-- プログレスバー -->
									<div
										class="progress-fill"
										style="width: {item.progress}%"
									></div>
									<!-- ラベル（バーが十分大きい場合） -->
									{#if pos.width > 60}
										<span class="bar-label">{item.progress}%</span>
									{/if}
								</div>
							</div>
						{/each}
					</div>
				</div>
			</div>
		{/if}
	</div>

	<!-- 統計パネル -->
	{#if timelineData && !loading}
		<div class="timeline-stats-panel">
			<div class="stat-item">
				<span class="stat-label">総タスク</span>
				<span class="stat-value">{timelineData.stats.total_tasks}</span>
			</div>
			<div class="stat-item">
				<span class="stat-label">日付あり</span>
				<span class="stat-value">{timelineData.stats.tasks_with_dates}</span>
			</div>
			<div class="stat-item critical">
				<span class="stat-label">クリティカル</span>
				<span class="stat-value">{timelineData.stats.on_critical_path}</span>
			</div>
			<div class="stat-item">
				<span class="stat-label">平均スラック</span>
				<span class="stat-value">{timelineData.stats.average_slack.toFixed(1)}日</span>
			</div>
			<div class="stat-item" class:warning={timelineData.stats.overdue_tasks > 0}>
				<span class="stat-label">遅延</span>
				<span class="stat-value">{timelineData.stats.overdue_tasks}</span>
			</div>
			<div class="stat-item">
				<span class="stat-label">期限内完了</span>
				<span class="stat-value">{timelineData.stats.completed_on_time}</span>
			</div>
		</div>
	{/if}
</div>

<!-- 選択中タスク詳細パネル -->
{#if selectedTaskId && timelineData}
	{@const selectedTask = filteredItems.find(t => t.task_id === selectedTaskId)}
	{#if selectedTask}
		<div class="task-detail-overlay">
			<div class="task-detail-card">
				<div class="detail-header">
					<h3>{selectedTask.title}</h3>
					<button class="close-btn" onclick={() => { selectedTaskId = null; onTaskSelect?.(null); }}>×</button>
				</div>
				<div class="detail-body">
					<div class="detail-row">
						<span class="label">期間</span>
						<span class="value">{formatDate(selectedTask.start_date)} - {formatDate(selectedTask.end_date)}</span>
					</div>
					<div class="detail-row">
						<span class="label">ステータス</span>
						<span class="value" style="color: {getStatusColor(selectedTask.status)}">{selectedTask.status}</span>
					</div>
					<div class="detail-row">
						<span class="label">進捗</span>
						<div class="progress-bar-container">
							<div class="progress-bar" style="width: {selectedTask.progress}%"></div>
							<span class="progress-text">{selectedTask.progress}%</span>
						</div>
					</div>
					<div class="detail-row">
						<span class="label">優先度</span>
						<span class="value" style="color: {getPriorityColor(selectedTask.priority)}">{selectedTask.priority}</span>
					</div>
					<div class="detail-row">
						<span class="label">担当者</span>
						<span class="value">{selectedTask.assignee || 'Unassigned'}</span>
					</div>
					<div class="detail-row">
						<span class="label">スラック</span>
						<span class="value">{selectedTask.slack}日</span>
					</div>
					{#if selectedTask.is_on_critical_path}
						<div class="critical-warning">
							<span class="icon">🔥</span>
							<span>クリティカルパス上のタスク</span>
						</div>
					{/if}
					{#if selectedTask.dependencies.length > 0}
						<div class="detail-row">
							<span class="label">依存</span>
							<span class="value">{selectedTask.dependencies.length} tasks</span>
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/if}
{/if}

<style>
	.timeline-viewer {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: #1a1a1a;
		color: #e0e0e0;
		font-family: 'JetBrains Mono', 'Fira Code', monospace;
	}

	.timeline-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px 16px;
		background: #252525;
		border-bottom: 1px solid #3a3a3a;
	}

	.timeline-title {
		display: flex;
		align-items: baseline;
		gap: 12px;
	}

	.timeline-title h2 {
		margin: 0;
		font-size: 16px;
		font-weight: 600;
		color: #f59e0b;
	}

	.timeline-stats {
		font-size: 12px;
		color: #888;
	}

	.timeline-controls {
		display: flex;
		gap: 8px;
	}

	.timeline-btn {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 6px 10px;
		background: #333;
		border: 1px solid #444;
		color: #ccc;
		border-radius: 4px;
		cursor: pointer;
		font-size: 12px;
		transition: all 0.2s;
	}

	.timeline-btn:hover {
		background: #444;
		border-color: #f59e0b;
		color: #f59e0b;
	}

	.timeline-btn.active {
		background: #f59e0b;
		border-color: #f59e0b;
		color: #1a1a1a;
	}

	.timeline-btn .icon {
		font-size: 14px;
	}

	.btn-label {
		font-size: 11px;
		text-transform: uppercase;
	}

	.timeline-filters {
		display: flex;
		gap: 8px;
		padding: 8px 16px;
		background: #222;
		border-bottom: 1px solid #333;
	}

	.timeline-search {
		flex: 1;
		padding: 6px 12px;
		background: #1a1a1a;
		border: 1px solid #333;
		color: #e0e0e0;
		border-radius: 4px;
		font-size: 13px;
	}

	.timeline-search:focus {
		outline: none;
		border-color: #f59e0b;
	}

	.timeline-select {
		padding: 6px 12px;
		background: #1a1a1a;
		border: 1px solid #333;
		color: #e0e0e0;
		border-radius: 4px;
		font-size: 13px;
		cursor: pointer;
	}

	.timeline-select:focus {
		outline: none;
		border-color: #f59e0b;
	}

	.timeline-content {
		flex: 1;
		overflow: hidden;
	}

	.timeline-loading,
	.timeline-error,
	.timeline-empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 300px;
		padding: var(--spacing-xl);
		gap: 12px;
		color: #888;
	}

	.timeline-empty .empty-icon {
		font-size: 64px;
		margin-bottom: var(--spacing-md);
	}

	.timeline-empty .empty-title {
		font-size: var(--font-size-lg);
		color: var(--text-primary);
		margin: 0 0 var(--spacing-sm) 0;
	}

	.timeline-empty .empty-description {
		font-size: var(--font-size-md);
		color: var(--text-muted);
		margin: 0 0 var(--spacing-lg) 0;
		max-width: 500px;
		text-align: center;
	}

	.timeline-empty .empty-guide {
		background: var(--bg-secondary);
		border: 1px solid var(--border-dark);
		border-radius: var(--border-radius-md);
		padding: var(--spacing-lg);
		max-width: 600px;
	}

	.timeline-empty .guide-label {
		font-size: var(--font-size-sm);
		color: var(--text-primary);
		font-weight: 600;
		margin: 0 0 var(--spacing-sm) 0;
	}

	.timeline-empty .guide-code {
		background: var(--bg-primary);
		border: 1px solid var(--border-metal);
		border-radius: var(--border-radius-sm);
		padding: var(--spacing-md);
		margin-bottom: var(--spacing-md);
		font-family: 'JetBrains Mono', 'Fira Code', monospace;
	}

	.timeline-empty .guide-code code {
		color: var(--accent-primary);
		font-size: var(--font-size-sm);
	}

	.timeline-empty .guide-hint {
		font-size: var(--font-size-xs);
		color: var(--text-muted);
		margin: 0;
	}

	.timeline-empty .guide-hint code {
		background: var(--bg-primary);
		padding: 2px 6px;
		border-radius: 3px;
		color: var(--accent-secondary);
	}

	.spinner {
		width: 24px;
		height: 24px;
		border: 2px solid #333;
		border-top-color: #f59e0b;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.timeline-error {
		color: #ef4444;
	}

	.error-icon {
		font-size: 24px;
	}

	.retry-btn {
		margin-top: 8px;
	}

	/* ガントチャート */
	.gantt-container {
		display: flex;
		height: 100%;
		overflow: hidden;
	}

	.task-list {
		width: 250px;
		min-width: 250px;
		background: #222;
		border-right: 1px solid #3a3a3a;
		overflow-y: auto;
	}

	.task-list-header {
		padding: 10px 12px;
		font-size: 11px;
		font-weight: 600;
		color: #888;
		text-transform: uppercase;
		background: #2a2a2a;
		border-bottom: 1px solid #3a3a3a;
		position: sticky;
		top: 0;
	}

	.task-row {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 12px;
		height: 36px;
		border-bottom: 1px solid #2a2a2a;
		cursor: pointer;
		transition: background 0.15s;
	}

	.task-row:hover,
	.task-row.hovered {
		background: #2a2a2a;
	}

	.task-row.selected {
		background: #3a3a3a;
		border-left: 3px solid #f59e0b;
		padding-left: 9px;
	}

	.task-row.critical {
		border-left: 3px solid #ef4444;
		padding-left: 9px;
	}

	.task-row.critical.selected {
		border-left-color: #f59e0b;
	}

	.task-status {
		font-size: 10px;
	}

	.task-title {
		flex: 1;
		font-size: 12px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.critical-badge {
		font-size: 12px;
	}

	/* タイムライングリッド */
	.timeline-grid-wrapper {
		flex: 1;
		overflow: auto;
	}

	.timeline-grid-header {
		position: sticky;
		top: 0;
		height: 32px;
		background: #2a2a2a;
		border-bottom: 1px solid #3a3a3a;
		z-index: 10;
	}

	.date-cell {
		position: absolute;
		top: 0;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 10px;
		color: #888;
		border-left: 1px solid #333;
	}

	.date-cell.week-start {
		border-left-color: #444;
		font-weight: 600;
		color: #aaa;
	}

	.timeline-grid {
		position: relative;
		min-height: calc(100% - 32px);
	}

	.grid-line {
		position: absolute;
		top: 0;
		bottom: 0;
		width: 1px;
		background: #333;
	}

	.today-line {
		position: absolute;
		top: 0;
		bottom: 0;
		width: 2px;
		background: #f59e0b;
		z-index: 5;
	}

	.task-bar-row {
		position: relative;
		height: 36px;
		border-bottom: 1px solid #2a2a2a;
	}

	.task-bar {
		position: absolute;
		top: 6px;
		height: 24px;
		border-radius: 4px;
		cursor: pointer;
		transition: all 0.15s;
		overflow: hidden;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.task-bar:hover,
	.task-bar.hovered {
		transform: scaleY(1.1);
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.4);
		z-index: 2;
	}

	.task-bar.selected {
		outline: 2px solid #f59e0b;
		outline-offset: 2px;
		z-index: 3;
	}

	.task-bar.critical {
		outline: 2px dashed #ef4444;
		outline-offset: -2px;
	}

	.progress-fill {
		position: absolute;
		left: 0;
		top: 0;
		height: 100%;
		background: rgba(0, 0, 0, 0.3);
	}

	.bar-label {
		position: relative;
		z-index: 1;
		font-size: 10px;
		font-weight: 600;
		color: #fff;
		text-shadow: 0 0 2px #000;
	}

	/* 統計パネル */
	.timeline-stats-panel {
		display: flex;
		justify-content: center;
		gap: 24px;
		padding: 12px 16px;
		background: #222;
		border-top: 1px solid #333;
	}

	.stat-item {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
	}

	.stat-item.critical .stat-label,
	.stat-item.critical .stat-value {
		color: #ef4444;
	}

	.stat-item.warning .stat-value {
		color: #f59e0b;
	}

	.stat-label {
		font-size: 11px;
		color: #888;
		text-transform: uppercase;
	}

	.stat-value {
		font-size: 16px;
		font-weight: 600;
		color: #f59e0b;
	}

	/* タスク詳細オーバーレイ */
	.task-detail-overlay {
		position: fixed;
		bottom: 20px;
		right: 20px;
		z-index: 100;
	}

	.task-detail-card {
		background: #252525;
		border: 2px solid #3a3a3a;
		border-radius: 8px;
		width: 320px;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
	}

	.detail-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px 16px;
		border-bottom: 1px solid #3a3a3a;
	}

	.detail-header h3 {
		margin: 0;
		font-size: 14px;
		font-weight: 600;
		color: #e0e0e0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		flex: 1;
	}

	.close-btn {
		width: 24px;
		height: 24px;
		background: none;
		border: none;
		color: #888;
		font-size: 18px;
		cursor: pointer;
		transition: color 0.15s;
	}

	.close-btn:hover {
		color: #e0e0e0;
	}

	.detail-body {
		padding: 12px 16px;
	}

	.detail-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 6px 0;
		border-bottom: 1px solid #2a2a2a;
	}

	.detail-row:last-child {
		border-bottom: none;
	}

	.detail-row .label {
		font-size: 11px;
		color: #888;
		text-transform: uppercase;
	}

	.detail-row .value {
		font-size: 13px;
		color: #e0e0e0;
	}

	.progress-bar-container {
		width: 120px;
		height: 16px;
		background: #2a2a2a;
		border-radius: 3px;
		position: relative;
		overflow: hidden;
	}

	.progress-bar {
		height: 100%;
		background: linear-gradient(90deg, #f59e0b, #d97706);
		transition: width 0.3s;
	}

	.progress-text {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		font-size: 10px;
		font-weight: 600;
		color: #fff;
		text-shadow: 0 0 2px #000;
	}

	.critical-warning {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 12px;
		margin-top: 8px;
		background: rgba(239, 68, 68, 0.2);
		border: 1px solid #ef4444;
		border-radius: 4px;
		color: #ef4444;
		font-size: 12px;
	}
</style>
