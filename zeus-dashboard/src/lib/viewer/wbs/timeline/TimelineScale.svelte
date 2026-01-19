<script lang="ts">
	// タイムラインスケール
	// 月/週/四半期の時間軸ヘッダーを表示
	interface Props {
		scale: 'week' | 'month' | 'quarter';
		startDate: Date;
		endDate: Date;
		onScaleChange: (scale: 'week' | 'month' | 'quarter') => void;
		onTodayClick: () => void;
	}
	let { scale, startDate, endDate, onScaleChange, onTodayClick }: Props = $props();

	// スケールに応じた期間ラベルを生成
	const periods = $derived(generatePeriods(startDate, endDate, scale));

	function generatePeriods(
		start: Date,
		end: Date,
		scaleType: 'week' | 'month' | 'quarter'
	): { label: string; width: number }[] {
		const result: { label: string; width: number }[] = [];
		const current = new Date(start);
		const totalDays = Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24));

		if (scaleType === 'month') {
			while (current <= end) {
				const monthStart = new Date(current.getFullYear(), current.getMonth(), 1);
				const monthEnd = new Date(current.getFullYear(), current.getMonth() + 1, 0);
				const daysInMonth = monthEnd.getDate();
				const label = current.toLocaleDateString('ja-JP', { month: 'short' });
				result.push({
					label,
					width: Math.round((daysInMonth / totalDays) * 100)
				});
				current.setMonth(current.getMonth() + 1);
			}
		} else if (scaleType === 'week') {
			while (current <= end) {
				const weekNum = getWeekNumber(current);
				result.push({
					label: `W${weekNum}`,
					width: Math.round((7 / totalDays) * 100)
				});
				current.setDate(current.getDate() + 7);
			}
		} else {
			// quarter
			while (current <= end) {
				const quarter = Math.floor(current.getMonth() / 3) + 1;
				const year = current.getFullYear();
				result.push({
					label: `Q${quarter} ${year}`,
					width: Math.round((90 / totalDays) * 100)
				});
				current.setMonth(current.getMonth() + 3);
			}
		}

		return result;
	}

	function getWeekNumber(date: Date): number {
		const firstDayOfYear = new Date(date.getFullYear(), 0, 1);
		const pastDaysOfYear = (date.getTime() - firstDayOfYear.getTime()) / 86400000;
		return Math.ceil((pastDaysOfYear + firstDayOfYear.getDay() + 1) / 7);
	}
</script>

<div class="timeline-scale">
	<div class="scale-controls">
		<span class="scale-label">TIMELINE</span>
		<div class="scale-buttons">
			<button
				class="scale-btn"
				class:active={scale === 'week'}
				onclick={() => onScaleChange('week')}
			>
				W
			</button>
			<button
				class="scale-btn"
				class:active={scale === 'month'}
				onclick={() => onScaleChange('month')}
			>
				M
			</button>
			<button
				class="scale-btn"
				class:active={scale === 'quarter'}
				onclick={() => onScaleChange('quarter')}
			>
				Q
			</button>
		</div>
		<button class="today-btn" onclick={onTodayClick}> Today </button>
	</div>
	<div class="scale-header">
		{#each periods as period}
			<div class="period" style="flex: {period.width}">
				{period.label}
			</div>
		{/each}
	</div>
</div>

<style>
	.timeline-scale {
		background: var(--bg-secondary, #242424);
		border-bottom: 1px solid var(--border-metal, #4a4a4a);
	}

	.scale-controls {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 8px 16px;
		border-bottom: 1px solid var(--border-dark, #333333);
	}

	.scale-label {
		font-size: 12px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: var(--accent-primary, #ff9533);
		margin-right: auto;
	}

	.scale-buttons {
		display: flex;
		gap: 2px;
		background: var(--bg-panel, #2d2d2d);
		border-radius: 4px;
		padding: 2px;
	}

	.scale-btn {
		padding: 4px 10px;
		font-size: 11px;
		font-weight: 600;
		background: transparent;
		border: none;
		border-radius: 2px;
		color: var(--text-muted, #888888);
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.scale-btn:hover {
		color: var(--text-primary, #ffffff);
	}

	.scale-btn.active {
		background: var(--accent-primary, #ff9533);
		color: var(--bg-primary, #1a1a1a);
	}

	.today-btn {
		padding: 4px 12px;
		font-size: 11px;
		font-weight: 500;
		background: var(--bg-panel, #2d2d2d);
		border: 1px solid var(--border-metal, #4a4a4a);
		border-radius: 4px;
		color: var(--text-secondary, #b8b8b8);
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.today-btn:hover {
		border-color: var(--accent-primary, #ff9533);
		color: var(--accent-primary, #ff9533);
	}

	.scale-header {
		display: flex;
		padding: 8px 16px;
		font-size: 11px;
		color: var(--text-muted, #888888);
	}

	.period {
		text-align: center;
		border-left: 1px solid var(--border-dark, #333333);
		padding: 0 4px;
	}

	.period:first-child {
		border-left: none;
	}
</style>
