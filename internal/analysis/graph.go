package analysis

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// GraphBuilder は依存関係グラフを構築
type GraphBuilder struct {
	tasks map[string]*TaskInfo
}

// NewGraphBuilder は新しい GraphBuilder を作成
func NewGraphBuilder(tasks []TaskInfo) *GraphBuilder {
	taskMap := make(map[string]*TaskInfo)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}
	return &GraphBuilder{tasks: taskMap}
}

// Build は依存関係グラフを構築
func (g *GraphBuilder) Build(ctx context.Context) (*DependencyGraph, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	graph := &DependencyGraph{
		Nodes:    make(map[string]*GraphNode),
		Edges:    []Edge{},
		Cycles:   [][]string{},
		Isolated: []string{},
	}

	// ノードを作成
	for id, task := range g.tasks {
		graph.Nodes[id] = &GraphNode{
			Task:     task,
			Children: task.Dependencies,
			Parents:  []string{},
			Depth:    0,
		}
	}

	// 親参照を設定し、エッジを構築
	for id, node := range graph.Nodes {
		for _, depID := range node.Children {
			if depNode, exists := graph.Nodes[depID]; exists {
				depNode.Parents = append(depNode.Parents, id)
			}
			graph.Edges = append(graph.Edges, Edge{From: id, To: depID})
		}
	}

	// 循環依存を検出
	graph.Cycles = g.detectCycles(graph)

	// 孤立ノードを検出
	graph.Isolated = g.findIsolated(graph)

	// 深さを計算
	g.calculateDepths(graph)

	// 統計を計算
	graph.Stats = g.calculateStats(graph)

	return graph, nil
}

// detectCycles は循環依存を検出（DFS）
func (g *GraphBuilder) detectCycles(graph *DependencyGraph) [][]string {
	cycles := [][]string{}
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(nodeID string, path []string) bool
	dfs = func(nodeID string, path []string) bool {
		visited[nodeID] = true
		recStack[nodeID] = true
		path = append(path, nodeID)

		node, exists := graph.Nodes[nodeID]
		if !exists {
			recStack[nodeID] = false
			return false
		}

		for _, childID := range node.Children {
			if !visited[childID] {
				if dfs(childID, path) {
					return true
				}
			} else if recStack[childID] {
				// 循環検出
				cycleStart := -1
				for i, id := range path {
					if id == childID {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := append(path[cycleStart:], childID)
					cycles = append(cycles, cycle)
				}
			}
		}

		recStack[nodeID] = false
		return false
	}

	for nodeID := range graph.Nodes {
		if !visited[nodeID] {
			dfs(nodeID, []string{})
		}
	}

	return cycles
}

// findIsolated は孤立ノードを検出
func (g *GraphBuilder) findIsolated(graph *DependencyGraph) []string {
	isolated := []string{}
	for id, node := range graph.Nodes {
		if len(node.Children) == 0 && len(node.Parents) == 0 {
			isolated = append(isolated, id)
		}
	}
	sort.Strings(isolated)
	return isolated
}

// calculateDepths はノードの深さを計算
func (g *GraphBuilder) calculateDepths(graph *DependencyGraph) {
	// ルートノード（親がないノード）を見つける
	roots := []string{}
	for id, node := range graph.Nodes {
		if len(node.Parents) == 0 {
			roots = append(roots, id)
		}
	}

	// BFS で深さを計算
	visited := make(map[string]bool)
	queue := make([]string, len(roots))
	copy(queue, roots)

	for _, root := range roots {
		graph.Nodes[root].Depth = 0
		visited[root] = true
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		node := graph.Nodes[current]
		for _, childID := range node.Children {
			if childNode, exists := graph.Nodes[childID]; exists {
				newDepth := node.Depth + 1
				if !visited[childID] || childNode.Depth < newDepth {
					childNode.Depth = newDepth
					visited[childID] = true
					queue = append(queue, childID)
				}
			}
		}
	}
}

// calculateStats はグラフの統計を計算
func (g *GraphBuilder) calculateStats(graph *DependencyGraph) GraphStats {
	stats := GraphStats{
		TotalNodes:    len(graph.Nodes),
		IsolatedCount: len(graph.Isolated),
		CycleCount:    len(graph.Cycles),
	}

	maxDepth := 0
	withDeps := 0
	for _, node := range graph.Nodes {
		if len(node.Children) > 0 || len(node.Parents) > 0 {
			withDeps++
		}
		if node.Depth > maxDepth {
			maxDepth = node.Depth
		}
	}
	stats.WithDependencies = withDeps
	stats.MaxDepth = maxDepth

	return stats
}

// ToText は TEXT 形式（ASCII アート）で出力
func (graph *DependencyGraph) ToText() string {
	var sb strings.Builder

	sb.WriteString("Zeus Dependency Graph\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	// ルートノードから表示
	roots := []string{}
	for id, node := range graph.Nodes {
		if len(node.Parents) == 0 && len(node.Children) > 0 {
			roots = append(roots, id)
		}
	}
	sort.Strings(roots)

	printed := make(map[string]bool)
	for _, root := range roots {
		graph.printTree(&sb, root, "", true, printed)
	}

	// 孤立ノードを表示
	if len(graph.Isolated) > 0 {
		sb.WriteString("\nIsolated Tasks (no dependencies):\n")
		for _, id := range graph.Isolated {
			if node, exists := graph.Nodes[id]; exists {
				sb.WriteString(fmt.Sprintf("  %s: %s\n", id, node.Task.Title))
			}
		}
	}

	// 循環依存の警告
	if len(graph.Cycles) > 0 {
		sb.WriteString("\nWarnings:\n")
		for _, cycle := range graph.Cycles {
			sb.WriteString(fmt.Sprintf("  - Circular dependency: %s\n", strings.Join(cycle, " -> ")))
		}
	}

	// 統計
	sb.WriteString(fmt.Sprintf("\nStats:\n"))
	sb.WriteString(fmt.Sprintf("  Total tasks: %d\n", graph.Stats.TotalNodes))
	sb.WriteString(fmt.Sprintf("  With dependencies: %d\n", graph.Stats.WithDependencies))
	sb.WriteString(fmt.Sprintf("  Isolated: %d\n", graph.Stats.IsolatedCount))
	if graph.Stats.CycleCount > 0 {
		sb.WriteString(fmt.Sprintf("  Circular dependencies: %d\n", graph.Stats.CycleCount))
	}

	sb.WriteString(strings.Repeat("=", 60) + "\n")

	return sb.String()
}

// printTree は再帰的にツリーを描画
func (graph *DependencyGraph) printTree(sb *strings.Builder, nodeID, prefix string, isLast bool, printed map[string]bool) {
	if printed[nodeID] {
		return
	}
	printed[nodeID] = true

	node, exists := graph.Nodes[nodeID]
	if !exists {
		return
	}

	connector := "+-"
	if isLast {
		connector = "`-"
	}
	if prefix == "" {
		connector = ""
	}

	sb.WriteString(fmt.Sprintf("%s%s%s: %s\n", prefix, connector, nodeID, node.Task.Title))

	newPrefix := prefix
	if prefix != "" {
		if isLast {
			newPrefix += "  "
		} else {
			newPrefix += "| "
		}
	}

	children := node.Children
	sort.Strings(children)
	for i, childID := range children {
		isLastChild := i == len(children)-1
		graph.printTree(sb, childID, newPrefix+"  ", isLastChild, printed)
	}
}

// ToDot は DOT 形式（Graphviz）で出力
func (graph *DependencyGraph) ToDot() string {
	var sb strings.Builder

	sb.WriteString("digraph ZeusDependencies {\n")
	sb.WriteString("  rankdir=TB;\n")
	sb.WriteString("  node [shape=box, style=rounded];\n\n")

	// ノード定義
	ids := make([]string, 0, len(graph.Nodes))
	for id := range graph.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		node := graph.Nodes[id]
		// ステータスに応じた色分け
		color := "white"
		switch node.Task.Status {
		case TaskStatusCompleted:
			color = "lightgreen"
		case TaskStatusInProgress:
			color = "lightyellow"
		case TaskStatusBlocked:
			color = "lightcoral"
		}
		label := strings.ReplaceAll(node.Task.Title, "\"", "\\\"")
		sb.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\\n(%s)\", fillcolor=%s, style=filled];\n",
			id, label, node.Task.Status, color))
	}

	sb.WriteString("\n")

	// エッジ定義
	for _, edge := range graph.Edges {
		sb.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", edge.From, edge.To))
	}

	sb.WriteString("}\n")

	return sb.String()
}

// GetDownstreamTasks は指定タスクの下流タスク（依存しているタスク）を取得
// taskID から始めて、そのタスクに依存している全てのタスクを再帰的に収集
func (graph *DependencyGraph) GetDownstreamTasks(taskID string) []string {
	downstream := []string{}
	visited := make(map[string]bool)

	var collect func(id string)
	collect = func(id string) {
		node, exists := graph.Nodes[id]
		if !exists {
			return
		}

		// このタスクを親として持つ（依存している）タスクを収集
		for _, parentID := range node.Parents {
			if !visited[parentID] {
				visited[parentID] = true
				downstream = append(downstream, parentID)
				collect(parentID)
			}
		}
	}

	collect(taskID)
	sort.Strings(downstream)
	return downstream
}

// GetUpstreamTasks は指定タスクの上流タスク（依存先）を取得
// taskID から始めて、そのタスクが依存している全てのタスクを再帰的に収集
func (graph *DependencyGraph) GetUpstreamTasks(taskID string) []string {
	upstream := []string{}
	visited := make(map[string]bool)

	var collect func(id string)
	collect = func(id string) {
		node, exists := graph.Nodes[id]
		if !exists {
			return
		}

		// このタスクが依存しているタスクを収集
		for _, childID := range node.Children {
			if !visited[childID] {
				visited[childID] = true
				upstream = append(upstream, childID)
				collect(childID)
			}
		}
	}

	collect(taskID)
	sort.Strings(upstream)
	return upstream
}

// ToMermaid は Mermaid 形式で出力
func (graph *DependencyGraph) ToMermaid() string {
	var sb strings.Builder

	sb.WriteString("```mermaid\n")
	sb.WriteString("graph TD\n")

	// ノード定義とステータスによるスタイリング
	ids := make([]string, 0, len(graph.Nodes))
	for id := range graph.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		node := graph.Nodes[id]
		// Mermaid用にIDをサニタイズ（ハイフンをアンダースコアに）
		safeID := strings.ReplaceAll(id, "-", "_")
		label := strings.ReplaceAll(node.Task.Title, "\"", "'")
		sb.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", safeID, label))
	}

	sb.WriteString("\n")

	// エッジ定義
	for _, edge := range graph.Edges {
		safeFrom := strings.ReplaceAll(edge.From, "-", "_")
		safeTo := strings.ReplaceAll(edge.To, "-", "_")
		sb.WriteString(fmt.Sprintf("    %s --> %s\n", safeFrom, safeTo))
	}

	// スタイル定義
	sb.WriteString("\n")
	for _, id := range ids {
		node := graph.Nodes[id]
		safeID := strings.ReplaceAll(id, "-", "_")
		switch node.Task.Status {
		case TaskStatusCompleted:
			sb.WriteString(fmt.Sprintf("    style %s fill:#90EE90\n", safeID))
		case TaskStatusInProgress:
			sb.WriteString(fmt.Sprintf("    style %s fill:#FFFFE0\n", safeID))
		case TaskStatusBlocked:
			sb.WriteString(fmt.Sprintf("    style %s fill:#F08080\n", safeID))
		}
	}

	sb.WriteString("```\n")

	return sb.String()
}
