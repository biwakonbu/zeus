<script lang="ts">
	import Panel from '$lib/components/ui/Panel.svelte';
	import Stat from '$lib/components/ui/Stat.svelte';
	import { projectState, statusLoading, statusError } from '$lib/stores/status';
</script>

<Panel title="Statistics" icon="&#128202;" loading={$statusLoading} error={$statusError}>
	<div class="stats-grid">
		<Stat
			label="Total Tasks"
			value={$projectState?.summary.total_tasks ?? 0}
			icon="&#128203;"
			variant="default"
		/>
		<Stat
			label="Completed"
			value={$projectState?.summary.completed ?? 0}
			icon="&#9989;"
			variant="success"
		/>
		<Stat
			label="In Progress"
			value={$projectState?.summary.in_progress ?? 0}
			icon="&#9881;"
			variant="info"
		/>
		<Stat
			label="Pending"
			value={$projectState?.summary.pending ?? 0}
			icon="&#9200;"
			variant="muted"
		/>
	</div>
</Panel>

<style>
	.stats-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: var(--spacing-md);
	}

	@media (max-width: 1024px) {
		.stats-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 640px) {
		.stats-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
