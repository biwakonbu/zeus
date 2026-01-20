<script lang="ts">
	// Affinity View
	// エンティティ間の関連性をフォースグラフで可視化するビュー
	// パフォーマンス最適化済み: ノードIDマップ、上限制限、収束判定、LOD
	import { onMount, onDestroy } from 'svelte';
	import type { AffinityResponse, AffinityNode, AffinityEdge } from '$lib/types/api';
	import { Icon } from '$lib/components/ui';

	interface Props {
		data: AffinityResponse | null;
		onNodeSelect: (nodeId: string, nodeType: string) => void;
	}
	let { data, onNodeSelect }: Props = $props();

	// パフォーマンス設定（Phase 3 最適化: ノード数削減で O(N²) を 91% 削減）
	const MAX_NODES = 50; // 150 → 50（11,175 → 1,225 ペア計算）
	const MAX_EDGES = 150; // 500 → 150
	const MIN_VELOCITY_THRESHOLD = 0.5; // 収束判定閾値
	const MAX_ITERATIONS = 100; // 最大イテレーション数

	// SVG コンテナ参照
	let svgContainer: SVGSVGElement | null = $state(null);

	// レイアウト状態
	interface LayoutNode extends AffinityNode {
		x: number;
		y: number;
		vx: number;
		vy: number;
		fx: number | null;
		fy: number | null;
	}

	let layoutNodes: LayoutNode[] = $state([]);
	let hoveredNodeId: string | null = $state(null);
	let selectedNodeId: string | null = $state(null);
	let animationFrame: number | null = null;
	let isSimulating = $state(false); // シミュレーション中フラグ
	let useStaticLayout = $state(true); // 静的レイアウトモード（デフォルト ON でフリーズ回避）

	// ノードIDマップ（O(1)検索用）
	let nodeMap: Map<string, LayoutNode> = $state(new Map());

	// 制限適用後のエッジ
	let limitedEdges: AffinityEdge[] = $state([]);

	// ビューポート状態
	let viewBox = $state({ x: 0, y: 0, width: 800, height: 600 });
	let zoom = $state(1);
	let isPanning = $state(false);
	let panStart = { x: 0, y: 0, viewX: 0, viewY: 0 };

	// フィルター状態
	let showEdges = $state(true);
	let minEdgeScore = $state(0.3);
	let showEdgesDuringAnimation = $state(false); // アニメーション中のエッジ表示

	// ノードタイプ色マッピング
	const nodeColors: Record<string, string> = {
		vision: '#f59e0b',
		objective: '#3b82f6',
		deliverable: '#10b981',
		task: '#8b5cf6'
	};

	// ノードサイズマッピング
	const nodeSizes: Record<string, number> = {
		vision: 24,
		objective: 18,
		deliverable: 14,
		task: 10
	};

	// エッジタイプ色マッピング
	const edgeColors: Record<string, string> = {
		'parent-child': '#f59e0b',
		sibling: '#3b82f6',
		'wbs-adjacent': '#10b981',
		reference: '#ec4899',
		category: '#8b5cf6'
	};

	// ノード優先度計算（重要なノードを残す）
	function getNodePriority(node: AffinityNode): number {
		const typeWeight: Record<string, number> = {
			vision: 100,
			objective: 50,
			deliverable: 30,
			task: 10
		};
		return (typeWeight[node.type] ?? 0) + (node.progress ?? 0);
	}

	// データ変更時にレイアウトを初期化（上限適用）
	$effect(() => {
		if (data?.nodes && data?.edges) {
			// ノード上限適用（優先度でソート）
			const sortedNodes = [...data.nodes]
				.sort((a, b) => getNodePriority(b) - getNodePriority(a))
				.slice(0, MAX_NODES);

			// 有効なノードIDセット
			const nodeIdSet = new Set(sortedNodes.map((n) => n.id));

			// エッジ上限適用（両端が有効なノードのもののみ、スコア順）
			limitedEdges = data.edges
				.filter((e) => nodeIdSet.has(e.source) && nodeIdSet.has(e.target))
				.sort((a, b) => b.score - a.score)
				.slice(0, MAX_EDGES);

			initializeLayout(sortedNodes);
		}
	});

	// ノードIDマップを更新
	$effect(() => {
		nodeMap = new Map(layoutNodes.map((n) => [n.id, n]));
	});

	// レイアウト初期化（Phase 3: 静的モードで即座に表示）
	function initializeLayout(nodes: AffinityNode[]) {
		const centerX = viewBox.width / 2;
		const centerY = viewBox.height / 2;
		const radius = Math.min(viewBox.width, viewBox.height) * 0.35;

		layoutNodes = nodes.map((node, i) => {
			// 初期配置は円形（静的モードでもこれで表示完了）
			const angle = (2 * Math.PI * i) / nodes.length;
			return {
				...node,
				x: centerX + radius * Math.cos(angle),
				y: centerY + radius * Math.sin(angle),
				vx: 0,
				vy: 0,
				fx: null,
				fy: null
			};
		});

		// 静的モードではシミュレーションをスキップ（フリーズ完全回避）
		if (!useStaticLayout) {
			startSimulation();
		}
	}

	// フォースシミュレーション（Phase 3: 非同期開始で UI ブロック回避）
	function startSimulation() {
		if (animationFrame) {
			cancelAnimationFrame(animationFrame);
		}

		let iteration = 0;
		const alpha = 0.3;
		const alphaDecay = 0.02;

		isSimulating = true;

		function tick() {
			if (iteration >= MAX_ITERATIONS || limitedEdges.length === 0) {
				isSimulating = false;
				// 最終状態でリアクティビティをトリガー
				layoutNodes = layoutNodes;
				return;
			}

			const currentAlpha = alpha * Math.pow(1 - alphaDecay, iteration);
			if (currentAlpha < 0.001) {
				isSimulating = false;
				layoutNodes = layoutNodes;
				return;
			}

			// フォース計算
			applyForces(currentAlpha);

			// 速度チェック - 収束したら早期終了
			let totalVelocity = 0;
			for (const node of layoutNodes) {
				totalVelocity += Math.abs(node.vx) + Math.abs(node.vy);
			}
			if (totalVelocity < MIN_VELOCITY_THRESHOLD && iteration > 20) {
				isSimulating = false;
				layoutNodes = layoutNodes;
				return;
			}

			// 位置更新（ミューテーション方式 - 配列再生成なし）
			for (const node of layoutNodes) {
				if (node.fx === null) node.x += node.vx;
				if (node.fy === null) node.y += node.vy;
				node.vx *= 0.6; // 減衰
				node.vy *= 0.6;
			}

			// リアクティビティトリガー（配列参照は同じだが更新を通知）
			layoutNodes = layoutNodes;

			iteration++;
			animationFrame = requestAnimationFrame(tick);
		}

		// Phase 3: 非同期開始 - 初期描画を先に完了させてからシミュレーション開始
		// 2フレーム待つことで UI スレッドのブロックを回避
		requestAnimationFrame(() => {
			requestAnimationFrame(tick);
		});
	}

	// フォース適用（最適化版 - nodeMap で O(1) 検索）
	function applyForces(alpha: number) {
		if (limitedEdges.length === 0) return;

		const centerX = viewBox.width / 2;
		const centerY = viewBox.height / 2;

		// ノード間反発力（O(N²) - ノード数上限で制御）
		for (let i = 0; i < layoutNodes.length; i++) {
			for (let j = i + 1; j < layoutNodes.length; j++) {
				const dx = layoutNodes[j].x - layoutNodes[i].x;
				const dy = layoutNodes[j].y - layoutNodes[i].y;
				const dist = Math.sqrt(dx * dx + dy * dy) || 1;
				const repulsion = (500 * alpha) / (dist * dist);

				const fx = (dx / dist) * repulsion;
				const fy = (dy / dist) * repulsion;

				layoutNodes[i].vx -= fx;
				layoutNodes[i].vy -= fy;
				layoutNodes[j].vx += fx;
				layoutNodes[j].vy += fy;
			}
		}

		// エッジ引力（最適化: Map で O(1) 検索、O(E×N) → O(E)）
		for (const edge of limitedEdges) {
			const source = nodeMap.get(edge.source);
			const target = nodeMap.get(edge.target);
			if (!source || !target) continue;

			const dx = target.x - source.x;
			const dy = target.y - source.y;
			const dist = Math.sqrt(dx * dx + dy * dy) || 1;
			const attraction = edge.score * 0.1 * alpha * dist;

			const fx = (dx / dist) * attraction;
			const fy = (dy / dist) * attraction;

			source.vx += fx;
			source.vy += fy;
			target.vx -= fx;
			target.vy -= fy;
		}

		// 中心への引力
		for (const node of layoutNodes) {
			const dx = centerX - node.x;
			const dy = centerY - node.y;
			node.vx += dx * 0.01 * alpha;
			node.vy += dy * 0.01 * alpha;
		}

		// 境界制約
		for (const node of layoutNodes) {
			const margin = 50;
			node.x = Math.max(margin, Math.min(viewBox.width - margin, node.x));
			node.y = Math.max(margin, Math.min(viewBox.height - margin, node.y));
		}
	}

	// ノードクリック
	function handleNodeClick(node: LayoutNode) {
		selectedNodeId = node.id;
		onNodeSelect(node.id, node.type);
	}

	// ノードドラッグ
	let draggedNode: LayoutNode | null = null;

	function handleNodeMouseDown(event: MouseEvent, node: LayoutNode) {
		event.stopPropagation();
		draggedNode = node;
		node.fx = node.x;
		node.fy = node.y;
	}

	function handleMouseMove(event: MouseEvent) {
		if (draggedNode && svgContainer) {
			const rect = svgContainer.getBoundingClientRect();
			const x = ((event.clientX - rect.left) / rect.width) * viewBox.width + viewBox.x;
			const y = ((event.clientY - rect.top) / rect.height) * viewBox.height + viewBox.y;

			draggedNode.fx = x;
			draggedNode.fy = y;
			draggedNode.x = x;
			draggedNode.y = y;

			// レイアウト配列を更新してリアクティビティをトリガー
			layoutNodes = [...layoutNodes];
		} else if (isPanning && svgContainer) {
			const dx = (event.clientX - panStart.x) / zoom;
			const dy = (event.clientY - panStart.y) / zoom;
			viewBox = {
				...viewBox,
				x: panStart.viewX - dx,
				y: panStart.viewY - dy
			};
		}
	}

	function handleMouseUp() {
		if (draggedNode) {
			draggedNode.fx = null;
			draggedNode.fy = null;
			draggedNode = null;
			// 静的モードではドラッグ後もシミュレーションを開始しない
			if (!useStaticLayout) {
				startSimulation();
			}
		}
		isPanning = false;
	}

	// 手動シミュレーション開始（Simulate ボタン用）
	function handleSimulateClick() {
		useStaticLayout = false;
		startSimulation();
	}

	// パン開始
	function handleSvgMouseDown(event: MouseEvent) {
		if (event.target === svgContainer) {
			isPanning = true;
			panStart = {
				x: event.clientX,
				y: event.clientY,
				viewX: viewBox.x,
				viewY: viewBox.y
			};
		}
	}

	// ズーム
	function handleWheel(event: WheelEvent) {
		event.preventDefault();
		const delta = event.deltaY > 0 ? 0.9 : 1.1;
		const newZoom = Math.max(0.5, Math.min(3, zoom * delta));

		if (svgContainer) {
			const rect = svgContainer.getBoundingClientRect();
			const mouseX = event.clientX - rect.left;
			const mouseY = event.clientY - rect.top;

			// マウス位置を中心にズーム
			const worldX = viewBox.x + (mouseX / rect.width) * viewBox.width;
			const worldY = viewBox.y + (mouseY / rect.height) * viewBox.height;

			const newWidth = viewBox.width / (newZoom / zoom);
			const newHeight = viewBox.height / (newZoom / zoom);

			viewBox = {
				x: worldX - (mouseX / rect.width) * newWidth,
				y: worldY - (mouseY / rect.height) * newHeight,
				width: newWidth,
				height: newHeight
			};
		}

		zoom = newZoom;
	}

	// フィルター済みエッジ（LOD: ズームアウト時は高スコアのみ）
	const filteredEdges = $derived.by(() => {
		const scoreFiltered = limitedEdges.filter((e) => e.score >= minEdgeScore);
		// LOD: ズームが 0.7 未満の場合、高スコアエッジのみ表示
		if (zoom < 0.7) {
			return scoreFiltered.filter((e) => e.score > 0.5);
		}
		return scoreFiltered;
	});

	// アニメーション中のエッジ表示判定
	const shouldShowEdges = $derived(showEdges && (!isSimulating || showEdgesDuringAnimation));

	// 統計情報
	const stats = $derived(data?.stats ?? null);
	const clusters = $derived(data?.clusters ?? []);

	onMount(() => {
		// viewBox を SVG の実際のサイズに合わせる
		if (svgContainer) {
			const rect = svgContainer.getBoundingClientRect();
			viewBox = { x: 0, y: 0, width: rect.width, height: rect.height };
		}
	});

	onDestroy(() => {
		if (animationFrame) {
			cancelAnimationFrame(animationFrame);
		}
	});
</script>

<div class="affinity-view">
	<!-- ヘッダー -->
	<div class="affinity-header">
		<span class="affinity-label">AFFINITY</span>

		<div class="controls">
			<!-- シミュレーションボタン / 状態インジケーター -->
			{#if isSimulating}
				<span class="sim-indicator">Simulating...</span>
			{:else if useStaticLayout}
				<button class="simulate-btn" onclick={handleSimulateClick}>
					<Icon name="RefreshCw" size={14} />
					<span>Simulate</span>
				</button>
			{/if}

			<!-- エッジ表示トグル -->
			<label class="control-item">
				<input type="checkbox" bind:checked={showEdges} />
				<span>Edges</span>
			</label>

			<!-- アニメーション中エッジ表示 -->
			<label class="control-item">
				<input type="checkbox" bind:checked={showEdgesDuringAnimation} />
				<span>Anim Edges</span>
			</label>

			<!-- スコアフィルター -->
			<div class="control-item">
				<span class="control-label">Min Score:</span>
				<input
					type="range"
					min="0"
					max="1"
					step="0.1"
					bind:value={minEdgeScore}
					class="score-slider"
				/>
				<span class="score-value">{minEdgeScore.toFixed(1)}</span>
			</div>
		</div>
	</div>

	<!-- メインキャンバス -->
	<div class="canvas-container">
		{#if !data}
			<div class="empty-state">
				<span class="empty-icon"><Icon name="GitBranch" size={32} /></span>
				<span class="empty-text">Affinity データを読み込み中...</span>
			</div>
		{:else if layoutNodes.length === 0}
			<div class="empty-state">
				<span class="empty-icon"><Icon name="GitBranch" size={32} /></span>
				<span class="empty-text">表示するノードがありません</span>
			</div>
		{:else}
			<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
			<svg
				bind:this={svgContainer}
				class="affinity-canvas"
				viewBox="{viewBox.x} {viewBox.y} {viewBox.width} {viewBox.height}"
				onmousedown={handleSvgMouseDown}
				onmousemove={handleMouseMove}
				onmouseup={handleMouseUp}
				onmouseleave={handleMouseUp}
				onwheel={handleWheel}
				role="img"
				aria-label="Affinity Graph"
			>
				<!-- 背景グリッド -->
				<defs>
					<pattern id="grid" width="50" height="50" patternUnits="userSpaceOnUse">
						<path
							d="M 50 0 L 0 0 0 50"
							fill="none"
							stroke="var(--border-subtle, #2a2a2a)"
							stroke-width="0.5"
						/>
					</pattern>
				</defs>
				<rect width="100%" height="100%" fill="url(#grid)" />

				<!-- エッジ（最適化: nodeMap で O(1) 検索、アニメーション中は非表示オプション） -->
				{#if shouldShowEdges}
					<g class="edges">
						{#each filteredEdges as edge}
							{@const source = nodeMap.get(edge.source)}
							{@const target = nodeMap.get(edge.target)}
							{#if source && target}
								<line
									x1={source.x}
									y1={source.y}
									x2={target.x}
									y2={target.y}
									stroke={edgeColors[edge.types[0]] ?? '#666'}
									stroke-width={1 + edge.score * 2}
									stroke-opacity={0.3 + edge.score * 0.5}
									class="edge"
									class:highlighted={hoveredNodeId === edge.source ||
										hoveredNodeId === edge.target}
								/>
							{/if}
						{/each}
					</g>
				{/if}

				<!-- ノード -->
				<g class="nodes">
					{#each layoutNodes as node}
						{@const size = nodeSizes[node.type] ?? 10}
						{@const color = nodeColors[node.type] ?? '#888'}
						<!-- svelte-ignore a11y_no_static_element_interactions -->
						<g
							class="node"
							class:selected={selectedNodeId === node.id}
							class:hovered={hoveredNodeId === node.id}
							transform="translate({node.x}, {node.y})"
							onmousedown={(e) => handleNodeMouseDown(e, node)}
							onmouseenter={() => (hoveredNodeId = node.id)}
							onmouseleave={() => (hoveredNodeId = null)}
							onclick={() => handleNodeClick(node)}
						>
							<!-- ノード本体 -->
							<circle r={size} fill={color} stroke="#1a1a1a" stroke-width="2" />

							<!-- 進捗リング -->
							{#if node.progress > 0}
								<circle
									r={size + 3}
									fill="none"
									stroke={color}
									stroke-width="2"
									stroke-dasharray="{(2 * Math.PI * (size + 3) * node.progress) / 100} {2 *
										Math.PI *
										(size + 3)}"
									stroke-linecap="round"
									transform="rotate(-90)"
									opacity="0.6"
								/>
							{/if}

							<!-- ラベル（ホバー時表示） -->
							{#if hoveredNodeId === node.id || selectedNodeId === node.id}
								<text y={size + 14} text-anchor="middle" class="node-label">
									{node.title.length > 20 ? node.title.slice(0, 20) + '...' : node.title}
								</text>
							{/if}
						</g>
					{/each}
				</g>
			</svg>
		{/if}
	</div>

	<!-- サイドパネル：統計 & クラスタ -->
	{#if stats}
		<div class="stats-panel">
			<div class="stats-header">Statistics</div>
			<div class="stats-grid">
				<div class="stat-item">
					<span class="stat-value">{layoutNodes.length}</span>
					<span class="stat-label">Nodes</span>
					{#if stats.total_nodes > MAX_NODES}
						<span class="stat-limit">/ {stats.total_nodes}</span>
					{/if}
				</div>
				<div class="stat-item">
					<span class="stat-value">{limitedEdges.length}</span>
					<span class="stat-label">Edges</span>
					{#if stats.total_edges > MAX_EDGES}
						<span class="stat-limit">/ {stats.total_edges}</span>
					{/if}
				</div>
				<div class="stat-item">
					<span class="stat-value">{stats.cluster_count}</span>
					<span class="stat-label">Clusters</span>
				</div>
				<div class="stat-item">
					<span class="stat-value">{stats.avg_connections.toFixed(1)}</span>
					<span class="stat-label">Avg Conn</span>
				</div>
			</div>

			{#if clusters.length > 0}
				<div class="clusters-section">
					<div class="clusters-header">Clusters</div>
					<div class="clusters-list">
						{#each clusters as cluster}
							<div class="cluster-item">
								<span class="cluster-name">{cluster.name}</span>
								<span class="cluster-count">{cluster.members.length}</span>
							</div>
						{/each}
					</div>
				</div>
			{/if}
		</div>
	{/if}

	<!-- 凡例 -->
	<div class="legend">
		<div class="legend-title">Node Types</div>
		<div class="legend-items">
			{#each Object.entries(nodeColors) as [type, color]}
				<div class="legend-item">
					<span class="legend-dot" style="background: {color}"></span>
					<span class="legend-label">{type}</span>
				</div>
			{/each}
		</div>
	</div>
</div>

<style>
	.affinity-view {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: var(--bg-primary, #1a1a1a);
		position: relative;
	}

	/* ヘッダー */
	.affinity-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 16px;
		background: var(--bg-secondary, #242424);
		border-bottom: 1px solid var(--border-metal, #4a4a4a);
	}

	.affinity-label {
		font-size: 12px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: var(--accent-primary, #ff9533);
	}

	.controls {
		display: flex;
		align-items: center;
		gap: 16px;
	}

	.control-item {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
		color: var(--text-muted, #888);
	}

	.control-item input[type='checkbox'] {
		accent-color: var(--accent-primary, #ff9533);
	}

	.control-label {
		color: var(--text-muted, #888);
	}

	.score-slider {
		width: 80px;
		accent-color: var(--accent-primary, #ff9533);
	}

	.score-value {
		min-width: 24px;
		text-align: right;
		color: var(--text-secondary, #ccc);
	}

	.sim-indicator {
		font-size: 11px;
		color: var(--accent-primary, #ff9533);
		animation: pulse 1s ease-in-out infinite;
	}

	.simulate-btn {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 4px 10px;
		background: var(--bg-secondary, #242424);
		border: 1px solid var(--accent-primary, #ff9533);
		border-radius: 3px;
		color: var(--accent-primary, #ff9533);
		font-size: 11px;
		font-family: inherit;
		cursor: pointer;
		transition: background 0.15s ease;
	}

	.simulate-btn:hover {
		background: var(--accent-primary, #ff9533);
		color: var(--bg-primary, #1a1a1a);
	}

	@keyframes pulse {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.5;
		}
	}

	/* キャンバス */
	.canvas-container {
		flex: 1;
		overflow: hidden;
		position: relative;
	}

	.affinity-canvas {
		width: 100%;
		height: 100%;
		cursor: grab;
	}

	.affinity-canvas:active {
		cursor: grabbing;
	}

	/* エッジ */
	.edge {
		transition: stroke-opacity 0.2s;
	}

	.edge.highlighted {
		stroke-opacity: 1 !important;
		stroke-width: 3 !important;
	}

	/* ノード */
	.node {
		cursor: pointer;
		transition: transform 0.15s ease;
	}

	.node:hover circle,
	.node.hovered circle {
		filter: brightness(1.2);
	}

	.node.selected circle {
		stroke: var(--accent-primary, #ff9533);
		stroke-width: 3;
	}

	.node-label {
		font-size: 11px;
		fill: var(--text-primary, #e0e0e0);
		font-family: var(--font-family, 'IBM Plex Mono', monospace);
		pointer-events: none;
	}

	/* 統計パネル */
	.stats-panel {
		position: absolute;
		top: 60px;
		right: 16px;
		background: var(--bg-panel, #2d2d2d);
		border: 1px solid var(--border-metal, #4a4a4a);
		border-radius: 4px;
		padding: 12px;
		min-width: 160px;
	}

	.stats-header,
	.clusters-header {
		font-size: 11px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-muted, #888);
		margin-bottom: 8px;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 8px;
	}

	.stat-item {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 6px;
		background: var(--bg-secondary, #242424);
		border-radius: 2px;
	}

	.stat-value {
		font-size: 16px;
		font-weight: 600;
		color: var(--text-primary, #e0e0e0);
	}

	.stat-label {
		font-size: 9px;
		color: var(--text-muted, #888);
		text-transform: uppercase;
	}

	.stat-limit {
		font-size: 8px;
		color: var(--text-muted, #666);
	}

	.clusters-section {
		margin-top: 12px;
		padding-top: 12px;
		border-top: 1px solid var(--border-subtle, #3a3a3a);
	}

	.clusters-list {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.cluster-item {
		display: flex;
		justify-content: space-between;
		padding: 4px 8px;
		background: var(--bg-secondary, #242424);
		border-radius: 2px;
		font-size: 11px;
	}

	.cluster-name {
		color: var(--text-secondary, #ccc);
	}

	.cluster-count {
		color: var(--accent-primary, #ff9533);
		font-weight: 600;
	}

	/* 凡例 */
	.legend {
		position: absolute;
		bottom: 16px;
		left: 16px;
		background: var(--bg-panel, #2d2d2d);
		border: 1px solid var(--border-metal, #4a4a4a);
		border-radius: 4px;
		padding: 8px 12px;
	}

	.legend-title {
		font-size: 10px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-muted, #888);
		margin-bottom: 6px;
	}

	.legend-items {
		display: flex;
		gap: 12px;
	}

	.legend-item {
		display: flex;
		align-items: center;
		gap: 4px;
	}

	.legend-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
	}

	.legend-label {
		font-size: 10px;
		color: var(--text-secondary, #ccc);
		text-transform: capitalize;
	}

	/* 空状態 */
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		color: var(--text-muted, #888);
	}

	.empty-icon {
		opacity: 0.5;
		margin-bottom: 8px;
	}

	.empty-text {
		font-size: 13px;
	}
</style>
