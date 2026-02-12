<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { Vision, Objective, UseCaseItem } from '$lib/types/api';
	import { fetchVision, fetchObjectives, fetchUseCases } from '$lib/api/client';
	import { Icon, EmptyState } from '$lib/components/ui';
	import {
		updateVisionViewState,
		resetVisionViewState,
		pendingNavigation,
		clearPendingNavigation
	} from '$lib/stores/view';
	import VisionHeroPanel from './VisionHeroPanel.svelte';
	import ObjectiveListPanel from './ObjectiveListPanel.svelte';
	import ObjectiveDetailPanel from './ObjectiveDetailPanel.svelte';

	// データ状態
	let vision: Vision | null = $state(null);
	let objectives: Objective[] = $state([]);
	let usecases: UseCaseItem[] = $state([]);
	let loading = $state(true);
	let error: string | null = $state(null);

	// 選択状態
	let selectedObjectiveId: string | null = $state(null);

	// 選択された Objective
	const selectedObjective = $derived.by((): Objective | null => {
		if (!selectedObjectiveId) return null;
		return objectives.find((o) => o.id === selectedObjectiveId) ?? null;
	});

	// 選択された Objective に紐付く UseCase
	const relatedUseCases = $derived.by((): UseCaseItem[] => {
		if (!selectedObjectiveId) return [];
		return usecases.filter((uc) => uc.objective_id === selectedObjectiveId);
	});

	// データ取得
	async function loadData() {
		loading = true;
		error = null;
		try {
			const [visionRes, objectivesRes, usecasesRes] = await Promise.all([
				fetchVision(),
				fetchObjectives(),
				fetchUseCases()
			]);
			vision = visionRes.vision;
			objectives = objectivesRes.objectives || [];
			usecases = usecasesRes.usecases || [];

			updateVisionViewState({
				objectiveCount: objectives.length,
				showListPanel: true
			});

			// 最初の Objective を自動選択
			if (objectives.length > 0 && !selectedObjectiveId) {
				selectedObjectiveId = objectives[0].id;
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'データの取得に失敗しました';
		} finally {
			loading = false;
		}
	}

	// Objective 選択
	function handleObjectiveSelect(id: string) {
		selectedObjectiveId = id;
		updateVisionViewState({ selectedObjectiveId: id });
	}

	// ESC キーで選択解除
	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			selectedObjectiveId = null;
			updateVisionViewState({ selectedObjectiveId: null });
		}
	}

	onMount(() => {
		loadData();
		document.addEventListener('keydown', handleKeydown);
	});

	// ナビゲーションによる自動選択
	$effect(() => {
		const nav = $pendingNavigation;
		if (!nav || nav.view !== 'vision') return;

		if (nav.entityType === 'objective' && nav.entityId) {
			const obj = objectives.find((o) => o.id === nav.entityId);
			if (obj) {
				selectedObjectiveId = obj.id;
				updateVisionViewState({ selectedObjectiveId: obj.id });
			}
			clearPendingNavigation();
		}
	});

	onDestroy(() => {
		document.removeEventListener('keydown', handleKeydown);
		resetVisionViewState();
	});
</script>

<div class="vision-view">
	{#if loading}
		<div class="loading-state">
			<Icon name="RefreshCw" size={32} />
			<span>読み込み中...</span>
		</div>
	{:else if error}
		<div class="error-state">
			<Icon name="AlertTriangle" size={32} />
			<span>{error}</span>
			<button class="retry-button" onclick={loadData}>再試行</button>
		</div>
	{:else if !vision && objectives.length === 0}
		<EmptyState
			title="Vision/Objective が定義されていません"
			description=".zeus/vision.yaml と .zeus/objectives/ にファイルを追加してください"
			icon="Eye"
		/>
	{:else}
		<div class="vision-content">
			<!-- Vision Hero パネル -->
			{#if vision}
				<div class="hero-area">
					<VisionHeroPanel {vision} />
				</div>
			{/if}

			<!-- Objective エリア: リスト + 詳細 -->
			{#if objectives.length > 0}
				<div class="objectives-area">
					<div class="objectives-list-col">
						<div class="col-header">
							<h3 class="col-title">Objectives</h3>
							<span class="col-count">{objectives.length}</span>
						</div>
						<ObjectiveListPanel
							{objectives}
							selectedId={selectedObjectiveId}
							onSelect={handleObjectiveSelect}
						/>
					</div>
					<div class="objectives-detail-col">
						{#if selectedObjective}
							<ObjectiveDetailPanel
								objective={selectedObjective}
								{relatedUseCases}
							/>
						{:else}
							<div class="no-selection">
								<Icon name="MousePointer" size={24} />
								<span>Objective を選択してください</span>
							</div>
						{/if}
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.vision-view {
		width: 100%;
		height: 100%;
		overflow-y: auto;
		background: var(--bg-primary);
		color: var(--text-primary);
	}

	.loading-state,
	.error-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 1rem;
		height: 100%;
		color: var(--text-muted);
	}

	.loading-state :global(svg) {
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}

	.error-state {
		color: var(--status-poor);
	}

	.retry-button {
		padding: 0.5rem 1rem;
		background: var(--bg-secondary);
		border: 1px solid var(--border-metal);
		color: var(--text-primary);
		cursor: pointer;
		border-radius: 4px;
		font-family: inherit;
	}

	.retry-button:hover {
		background: var(--bg-hover);
	}

	.vision-content {
		max-width: 1200px;
		margin: 0 auto;
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	.hero-area {
		width: 100%;
	}

	.objectives-area {
		display: grid;
		grid-template-columns: minmax(240px, 320px) 1fr;
		gap: 16px;
		min-height: 400px;
	}

	@media (max-width: 768px) {
		.objectives-area {
			grid-template-columns: 1fr;
		}
	}

	.objectives-list-col {
		display: flex;
		flex-direction: column;
		background: var(--bg-secondary);
		border: 1px solid var(--border-metal);
		border-radius: 8px;
		overflow: hidden;
	}

	.col-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 16px;
		border-bottom: 1px solid var(--border-metal);
	}

	.col-title {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.col-count {
		font-size: 0.6875rem;
		font-weight: 600;
		padding: 1px 8px;
		background: rgba(245, 158, 11, 0.15);
		color: var(--accent-primary);
		border-radius: 8px;
	}

	.objectives-detail-col {
		background: var(--bg-secondary);
		border: 1px solid var(--border-metal);
		border-radius: 8px;
		overflow-y: auto;
	}

	.no-selection {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 8px;
		height: 100%;
		min-height: 200px;
		color: var(--text-muted);
		font-size: 0.8125rem;
	}
</style>
