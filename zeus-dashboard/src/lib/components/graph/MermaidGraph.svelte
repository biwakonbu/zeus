<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import mermaid from 'mermaid';

	interface Props {
		code: string;
		id?: string;
	}

	let { code, id = 'mermaid-graph' }: Props = $props();

	let containerElement: HTMLDivElement;
	let initialized = $state(false);
	let error = $state<string | null>(null);

	// Mermaid 初期化
	onMount(() => {
		mermaid.initialize({
			startOnLoad: false,
			theme: 'dark',
			themeVariables: {
				darkMode: true,
				background: '#2d2d2d',
				primaryColor: '#ff9533',
				primaryTextColor: '#ffffff',
				primaryBorderColor: '#4a4a4a',
				lineColor: '#666666',
				secondaryColor: '#242424',
				tertiaryColor: '#1a1a1a',
				nodeTextColor: '#ffffff',
				mainBkg: '#2d2d2d',
				nodeBorder: '#ff9533',
				clusterBkg: '#242424',
				clusterBorder: '#4a4a4a',
				titleColor: '#ff9533',
				edgeLabelBackground: '#2d2d2d'
			},
			flowchart: {
				htmlLabels: true,
				curve: 'basis',
				padding: 15,
				nodeSpacing: 50,
				rankSpacing: 50,
				useMaxWidth: true
			},
			securityLevel: 'strict'
		});
		initialized = true;
	});

	// コードが変更されたら再レンダリング
	$effect(() => {
		if (initialized && code && containerElement) {
			renderGraph();
		}
	});

	async function renderGraph() {
		if (!code.trim()) {
			error = null;
			containerElement.innerHTML = '<div class="empty-graph">No graph data available</div>';
			return;
		}

		try {
			error = null;
			const uniqueId = `${id}-${Date.now()}`;
			const { svg } = await mermaid.render(uniqueId, code);
			containerElement.innerHTML = svg;

			// SVG のスタイルを調整
			const svgElement = containerElement.querySelector('svg');
			if (svgElement) {
				svgElement.style.maxWidth = '100%';
				svgElement.style.height = 'auto';
			}
		} catch (err) {
			console.error('Mermaid render error:', err);
			error = err instanceof Error ? err.message : 'Failed to render graph';
			containerElement.innerHTML = '';
		}
	}

	onDestroy(() => {
		if (containerElement) {
			containerElement.innerHTML = '';
		}
	});
</script>

<div class="mermaid-container">
	{#if error}
		<div class="mermaid-error">
			<span class="error-icon">&#9888;</span>
			<span class="error-text">{error}</span>
		</div>
	{/if}
	<div bind:this={containerElement} class="mermaid-graph"></div>
</div>

<style>
	.mermaid-container {
		width: 100%;
		overflow-x: auto;
	}

	.mermaid-graph {
		display: flex;
		justify-content: center;
		min-height: 200px;
		padding: var(--spacing-md);
	}

	.mermaid-graph :global(svg) {
		max-width: 100%;
		height: auto;
	}

	/* Mermaid ノードのカスタムスタイル */
	.mermaid-graph :global(.node rect) {
		fill: var(--bg-secondary) !important;
		stroke: var(--accent-primary) !important;
		stroke-width: 2px !important;
	}

	.mermaid-graph :global(.node polygon) {
		fill: var(--bg-secondary) !important;
		stroke: var(--accent-primary) !important;
		stroke-width: 2px !important;
	}

	.mermaid-graph :global(.edgePath path) {
		stroke: var(--border-highlight) !important;
		stroke-width: 2px !important;
	}

	.mermaid-graph :global(.edgePath marker path) {
		fill: var(--border-highlight) !important;
	}

	.mermaid-error {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		padding: var(--spacing-sm) var(--spacing-md);
		background-color: rgba(238, 68, 68, 0.1);
		border: 1px solid var(--status-poor);
		border-radius: var(--border-radius-sm);
		color: var(--status-poor);
		font-size: var(--font-size-sm);
		margin-bottom: var(--spacing-md);
	}

	.mermaid-graph :global(.empty-graph) {
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--text-muted);
		font-style: italic;
		min-height: 150px;
	}
</style>
