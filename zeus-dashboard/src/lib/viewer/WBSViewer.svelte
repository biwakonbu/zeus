<script lang="ts">
	import { onMount } from 'svelte';
	import { fetchWBS } from '$lib/api/client';
	import type { WBSResponse, WBSNode, WBSStats, TaskStatus, Priority } from '$lib/types/api';

	// Props
	interface Props {
		onNodeSelect?: (node: WBSNode | null) => void;
	}
	let { onNodeSelect }: Props = $props();

	// 状態
	let wbsData: WBSResponse | null = $state(null);
	let loading = $state(true);
	let error: string | null = $state(null);
	let expandedNodes: Set<string> = $state(new Set());
	let selectedNodeId: string | null = $state(null);
	let searchQuery = $state('');
	let statusFilter: TaskStatus | 'all' = $state('all');
	let priorityFilter: Priority | 'all' = $state('all');

	// フィルターされたルートノード
	let filteredRoots = $derived.by(() => {
		if (!wbsData) return [];
		return filterNodes(wbsData.roots);
	});

	// ノードのフィルタリング（再帰的）
	function filterNodes(nodes: WBSNode[]): WBSNode[] {
		const result: WBSNode[] = [];
		for (const node of nodes) {
			const filteredChildren = node.children ? filterNodes(node.children) : undefined;
			const matchesSearch =
				!searchQuery ||
				node.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
				node.wbs_code.toLowerCase().includes(searchQuery.toLowerCase());
			const matchesStatus = statusFilter === 'all' || node.status === statusFilter;
			const matchesPriority = priorityFilter === 'all' || node.priority === priorityFilter;

			// 子がマッチするか、自身がマッチする場合は含める
			const hasMatchingChildren = filteredChildren && filteredChildren.length > 0;
			if (hasMatchingChildren || (matchesSearch && matchesStatus && matchesPriority)) {
				result.push({ ...node, children: filteredChildren });
			}
		}
		return result;
	}

	// データ読み込み
	async function loadData() {
		loading = true;
		error = null;
		try {
			wbsData = await fetchWBS();
			// デフォルトで最初のレベルを展開
			if (wbsData.roots) {
				wbsData.roots.forEach((root) => expandedNodes.add(root.id));
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'WBS データの読み込みに失敗しました';
		} finally {
			loading = false;
		}
	}

	// 展開/折りたたみの切り替え
	function toggleExpand(nodeId: string) {
		if (expandedNodes.has(nodeId)) {
			expandedNodes.delete(nodeId);
		} else {
			expandedNodes.add(nodeId);
		}
		expandedNodes = new Set(expandedNodes);
	}

	// ノード選択
	function selectNode(node: WBSNode) {
		selectedNodeId = node.id;
		onNodeSelect?.(node);
	}

	// すべて展開
	function expandAll() {
		if (!wbsData) return;
		const collectIds = (nodes: WBSNode[]): string[] => {
			return nodes.flatMap((n) => [n.id, ...(n.children ? collectIds(n.children) : [])]);
		};
		expandedNodes = new Set(collectIds(wbsData.roots));
	}

	// すべて折りたたむ
	function collapseAll() {
		expandedNodes = new Set();
	}

	// ステータスに応じたアイコンとカラー
	function getStatusInfo(status: TaskStatus): { icon: string; color: string; label: string } {
		switch (status) {
			case 'completed':
				return { icon: '✓', color: '#22c55e', label: '完了' };
			case 'in_progress':
				return { icon: '●', color: '#f59e0b', label: '進行中' };
			case 'blocked':
				return { icon: '✗', color: '#ef4444', label: 'ブロック' };
			case 'pending':
			default:
				return { icon: '○', color: '#6b7280', label: '未着手' };
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

	onMount(() => {
		loadData();
	});
</script>

<div class="wbs-viewer">
	<!-- ヘッダー -->
	<div class="wbs-header">
		<div class="wbs-title">
			<h2>WBS Structure</h2>
			{#if wbsData}
				<span class="wbs-stats">
					{wbsData.stats.total_nodes} tasks | Depth: {wbsData.stats.max_depth} | {wbsData.stats
						.completed_pct}% complete
				</span>
			{/if}
		</div>
		<div class="wbs-controls">
			<button class="wbs-btn" onclick={() => expandAll()} title="すべて展開">
				<span class="icon">⊞</span>
			</button>
			<button class="wbs-btn" onclick={() => collapseAll()} title="すべて折りたたむ">
				<span class="icon">⊟</span>
			</button>
			<button class="wbs-btn" onclick={() => loadData()} title="更新">
				<span class="icon">↻</span>
			</button>
		</div>
	</div>

	<!-- フィルター -->
	<div class="wbs-filters">
		<input
			type="text"
			class="wbs-search"
			placeholder="検索..."
			bind:value={searchQuery}
		/>
		<select class="wbs-select" bind:value={statusFilter}>
			<option value="all">全ステータス</option>
			<option value="pending">未着手</option>
			<option value="in_progress">進行中</option>
			<option value="completed">完了</option>
			<option value="blocked">ブロック</option>
		</select>
		<select class="wbs-select" bind:value={priorityFilter}>
			<option value="all">全優先度</option>
			<option value="high">高</option>
			<option value="medium">中</option>
			<option value="low">低</option>
		</select>
	</div>

	<!-- ツリー表示 -->
	<div class="wbs-tree-container">
		{#if loading}
			<div class="wbs-loading">
				<div class="spinner"></div>
				<span>読み込み中...</span>
			</div>
		{:else if error}
			<div class="wbs-error">
				<span class="error-icon">⚠</span>
				<span>{error}</span>
				<button class="wbs-btn retry-btn" onclick={() => loadData()}>再試行</button>
			</div>
		{:else if filteredRoots.length === 0}
			<div class="wbs-empty">
				<span>表示するタスクがありません</span>
			</div>
		{:else}
			<div class="wbs-tree">
				{#each filteredRoots as node}
					{@render treeNode(node, 0)}
				{/each}
			</div>
		{/if}
	</div>

	<!-- 統計パネル -->
	{#if wbsData && !loading}
		<div class="wbs-stats-panel">
			<div class="stat-item">
				<span class="stat-label">ルート</span>
				<span class="stat-value">{wbsData.stats.root_count}</span>
			</div>
			<div class="stat-item">
				<span class="stat-label">リーフ</span>
				<span class="stat-value">{wbsData.stats.leaf_count}</span>
			</div>
			<div class="stat-item">
				<span class="stat-label">平均進捗</span>
				<span class="stat-value">{wbsData.stats.avg_progress}%</span>
			</div>
		</div>
	{/if}
</div>

<!-- ツリーノードの再帰的レンダリング -->
{#snippet treeNode(node: WBSNode, depth: number)}
	{@const hasChildren = node.children && node.children.length > 0}
	{@const isExpanded = expandedNodes.has(node.id)}
	{@const isSelected = selectedNodeId === node.id}
	{@const statusInfo = getStatusInfo(node.status)}
	{@const priorityColor = getPriorityColor(node.priority)}

	<div class="tree-node" style="--depth: {depth}">
		<div
			class="node-row"
			class:selected={isSelected}
			class:has-children={hasChildren}
			onclick={() => selectNode(node)}
			onkeydown={(e) => e.key === 'Enter' && selectNode(node)}
			role="treeitem"
			tabindex="0"
			aria-selected={isSelected}
			aria-expanded={hasChildren ? isExpanded : undefined}
		>
			<!-- 展開ボタン -->
			<button
				class="expand-btn"
				class:invisible={!hasChildren}
				onclick={(e) => {
					e.stopPropagation();
					toggleExpand(node.id);
				}}
				aria-label={isExpanded ? '折りたたむ' : '展開'}
			>
				{#if hasChildren}
					<span class="expand-icon" class:expanded={isExpanded}>▶</span>
				{/if}
			</button>

			<!-- WBS コード -->
			{#if node.wbs_code}
				<span class="wbs-code">{node.wbs_code}</span>
			{/if}

			<!-- タイトル -->
			<span class="node-title">{node.title}</span>

			<!-- プログレスバー -->
			<div class="progress-bar-container">
				<div class="progress-bar" style="width: {node.progress}%"></div>
				<span class="progress-text">{node.progress}%</span>
			</div>

			<!-- ステータス -->
			<span class="status-badge" style="color: {statusInfo.color}" title={statusInfo.label}>
				{statusInfo.icon}
			</span>

			<!-- 優先度インジケーター -->
			<span class="priority-indicator" style="background-color: {priorityColor}" title={node.priority}></span>

			<!-- 担当者 -->
			{#if node.assignee}
				<span class="assignee" title={node.assignee}>
					{node.assignee.slice(0, 2).toUpperCase()}
				</span>
			{/if}
		</div>

		<!-- 子ノード -->
		{#if hasChildren && isExpanded}
			<div class="children">
				{#each node.children as child}
					{@render treeNode(child, depth + 1)}
				{/each}
			</div>
		{/if}
	</div>
{/snippet}

<style>
	.wbs-viewer {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: #1a1a1a;
		color: #e0e0e0;
		font-family: 'JetBrains Mono', 'Fira Code', monospace;
	}

	.wbs-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px 16px;
		background: #252525;
		border-bottom: 1px solid #3a3a3a;
	}

	.wbs-title {
		display: flex;
		align-items: baseline;
		gap: 12px;
	}

	.wbs-title h2 {
		margin: 0;
		font-size: 16px;
		font-weight: 600;
		color: #f59e0b;
	}

	.wbs-stats {
		font-size: 12px;
		color: #888;
	}

	.wbs-controls {
		display: flex;
		gap: 8px;
	}

	.wbs-btn {
		padding: 6px 10px;
		background: #333;
		border: 1px solid #444;
		color: #ccc;
		border-radius: 4px;
		cursor: pointer;
		font-size: 14px;
		transition: all 0.2s;
	}

	.wbs-btn:hover {
		background: #444;
		border-color: #f59e0b;
		color: #f59e0b;
	}

	.wbs-btn .icon {
		font-size: 14px;
	}

	.wbs-filters {
		display: flex;
		gap: 8px;
		padding: 8px 16px;
		background: #222;
		border-bottom: 1px solid #333;
	}

	.wbs-search {
		flex: 1;
		padding: 6px 12px;
		background: #1a1a1a;
		border: 1px solid #333;
		color: #e0e0e0;
		border-radius: 4px;
		font-size: 13px;
	}

	.wbs-search:focus {
		outline: none;
		border-color: #f59e0b;
	}

	.wbs-select {
		padding: 6px 12px;
		background: #1a1a1a;
		border: 1px solid #333;
		color: #e0e0e0;
		border-radius: 4px;
		font-size: 13px;
		cursor: pointer;
	}

	.wbs-select:focus {
		outline: none;
		border-color: #f59e0b;
	}

	.wbs-tree-container {
		flex: 1;
		overflow: auto;
		padding: 8px;
	}

	.wbs-loading,
	.wbs-error,
	.wbs-empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 200px;
		gap: 12px;
		color: #888;
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

	.wbs-error {
		color: #ef4444;
	}

	.error-icon {
		font-size: 24px;
	}

	.retry-btn {
		margin-top: 8px;
	}

	.wbs-tree {
		padding: 4px 0;
	}

	.tree-node {
		margin-left: calc(var(--depth) * 20px);
	}

	.node-row {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 8px;
		border-radius: 4px;
		cursor: pointer;
		transition: background 0.15s;
	}

	.node-row:hover {
		background: #2a2a2a;
	}

	.node-row.selected {
		background: #3a3a3a;
		border-left: 3px solid #f59e0b;
		padding-left: 5px;
	}

	.expand-btn {
		width: 20px;
		height: 20px;
		padding: 0;
		background: none;
		border: none;
		color: #888;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.expand-btn.invisible {
		visibility: hidden;
	}

	.expand-btn:hover {
		color: #f59e0b;
	}

	.expand-icon {
		font-size: 10px;
		transition: transform 0.2s;
	}

	.expand-icon.expanded {
		transform: rotate(90deg);
	}

	.wbs-code {
		font-size: 11px;
		color: #f59e0b;
		background: #2a2a2a;
		padding: 2px 6px;
		border-radius: 3px;
		font-weight: 600;
		min-width: 50px;
		text-align: center;
	}

	.node-title {
		flex: 1;
		font-size: 13px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.progress-bar-container {
		width: 80px;
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

	.status-badge {
		font-size: 14px;
		width: 20px;
		text-align: center;
	}

	.priority-indicator {
		width: 8px;
		height: 8px;
		border-radius: 50%;
	}

	.assignee {
		font-size: 11px;
		background: #3a3a3a;
		color: #ccc;
		padding: 2px 6px;
		border-radius: 3px;
		font-weight: 500;
	}

	.children {
		border-left: 1px dashed #444;
		margin-left: 10px;
	}

	.wbs-stats-panel {
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
</style>
