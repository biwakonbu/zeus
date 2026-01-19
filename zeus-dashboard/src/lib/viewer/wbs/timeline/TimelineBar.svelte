<script lang="ts">
	// タイムラインバー
	// 計画（薄色）と実績（濃色）の 2 段バーを表示
	interface Props {
		id: string;
		title: string;
		planStart: Date;
		planEnd: Date;
		actualStart?: Date;
		actualEnd?: Date;
		progress: number;
		status: 'on_track' | 'delayed' | 'ahead' | 'completed';
		timelineStart: Date;
		timelineEnd: Date;
		selected?: boolean;
		onClick: () => void;
	}
	let {
		id,
		title,
		planStart,
		planEnd,
		actualStart,
		actualEnd,
		progress,
		status,
		timelineStart,
		timelineEnd,
		selected = false,
		onClick
	}: Props = $props();

	// タイムライン全体の日数
	const totalDays = $derived(
		Math.ceil((timelineEnd.getTime() - timelineStart.getTime()) / (1000 * 60 * 60 * 24))
	);

	// 計画バーの位置と幅（%）
	const planLeft = $derived(
		Math.max(
			0,
			((planStart.getTime() - timelineStart.getTime()) / (1000 * 60 * 60 * 24) / totalDays) * 100
		)
	);
	const planWidth = $derived(
		Math.min(
			100 - planLeft,
			((planEnd.getTime() - planStart.getTime()) / (1000 * 60 * 60 * 24) / totalDays) * 100
		)
	);

	// 実績バーの位置と幅（%）
	const actualLeft = $derived(
		actualStart
			? Math.max(
					0,
					((actualStart.getTime() - timelineStart.getTime()) / (1000 * 60 * 60 * 24) / totalDays) *
						100
				)
			: planLeft
	);
	const actualWidth = $derived(() => {
		if (!actualStart) return (planWidth * progress) / 100;
		const endDate = actualEnd || new Date();
		const width =
			((endDate.getTime() - actualStart.getTime()) / (1000 * 60 * 60 * 24) / totalDays) * 100;
		return Math.min(100 - actualLeft, width);
	});

	// ステータスラベル
	const statusLabel = $derived(
		status === 'on_track'
			? 'ON TRACK'
			: status === 'delayed'
				? 'DELAYED'
				: status === 'ahead'
					? 'AHEAD'
					: 'COMPLETED'
	);
</script>

<button class="timeline-bar" class:selected onclick={onClick}>
	<div class="bar-header">
		<span class="bar-id">{id}</span>
		<span class="bar-title">{title}</span>
	</div>
	<div class="bar-content">
		<!-- 計画バー -->
		<div class="bar-row">
			<span class="bar-label">Plan:</span>
			<div class="bar-track">
				<div class="bar-fill bar-fill--plan" style="left: {planLeft}%; width: {planWidth}%"></div>
			</div>
		</div>
		<!-- 実績バー -->
		<div class="bar-row">
			<span class="bar-label">Real:</span>
			<div class="bar-track">
				<div
					class="bar-fill bar-fill--actual bar-fill--{status}"
					style="left: {actualLeft}%; width: {actualWidth}%"
				></div>
			</div>
		</div>
	</div>
	<div class="bar-footer">
		<span class="progress-text">{progress}%</span>
		<span class="status-text status-text--{status}">{statusLabel}</span>
	</div>
</button>

<style>
	.timeline-bar {
		display: block;
		width: 100%;
		padding: 12px 16px;
		background: transparent;
		border: none;
		border-bottom: 1px solid var(--border-dark, #333333);
		cursor: pointer;
		text-align: left;
		transition: background-color 0.15s ease;
	}

	.timeline-bar:hover {
		background-color: var(--bg-hover, #3a3a3a);
	}

	.timeline-bar.selected {
		background-color: var(--bg-secondary, #242424);
		border-left: 3px solid var(--accent-primary, #ff9533);
		padding-left: 13px;
	}

	.bar-header {
		display: flex;
		gap: 12px;
		margin-bottom: 8px;
	}

	.bar-id {
		font-size: 11px;
		font-weight: 500;
		color: var(--text-muted, #888888);
	}

	.bar-title {
		font-size: 13px;
		color: var(--text-primary, #ffffff);
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.bar-content {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.bar-row {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.bar-label {
		font-size: 10px;
		color: var(--text-muted, #888888);
		min-width: 32px;
	}

	.bar-track {
		flex: 1;
		height: 8px;
		background: var(--bg-panel, #2d2d2d);
		border-radius: 2px;
		position: relative;
		overflow: hidden;
	}

	.bar-fill {
		position: absolute;
		top: 0;
		height: 100%;
		border-radius: 2px;
		transition:
			width 0.3s ease,
			left 0.3s ease;
	}

	.bar-fill--plan {
		background: var(--border-metal, #4a4a4a);
		opacity: 0.6;
	}

	.bar-fill--actual {
		background: var(--accent-primary, #ff9533);
	}

	.bar-fill--on_track {
		background: var(--status-good, #44cc44);
	}

	.bar-fill--delayed {
		background: var(--status-poor, #ee4444);
	}

	.bar-fill--ahead {
		background: var(--status-info, #4488ff);
	}

	.bar-fill--completed {
		background: var(--status-good, #44cc44);
	}

	.bar-footer {
		display: flex;
		justify-content: space-between;
		margin-top: 8px;
	}

	.progress-text {
		font-size: 12px;
		font-weight: 600;
		color: var(--text-secondary, #b8b8b8);
	}

	.status-text {
		font-size: 10px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.status-text--on_track {
		color: var(--status-good, #44cc44);
	}

	.status-text--delayed {
		color: var(--status-poor, #ee4444);
	}

	.status-text--ahead {
		color: var(--status-info, #4488ff);
	}

	.status-text--completed {
		color: var(--status-good, #44cc44);
	}
</style>
