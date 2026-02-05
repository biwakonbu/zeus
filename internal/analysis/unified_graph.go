package analysis

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// UnifiedGraphBuilder は統合グラフを構築する
type UnifiedGraphBuilder struct {
	activities   []ActivityInfo
	usecases     []UseCaseInfo
	deliverables []DeliverableInfo
	objectives   []ObjectiveInfo
	filter       *GraphFilter
}

// NewUnifiedGraphBuilder は新しい UnifiedGraphBuilder を作成
func NewUnifiedGraphBuilder() *UnifiedGraphBuilder {
	return &UnifiedGraphBuilder{
		activities:   []ActivityInfo{},
		usecases:     []UseCaseInfo{},
		deliverables: []DeliverableInfo{},
		objectives:   []ObjectiveInfo{},
		filter:       NewGraphFilter(),
	}
}

// WithActivities は Activity 情報を設定
func (b *UnifiedGraphBuilder) WithActivities(activities []ActivityInfo) *UnifiedGraphBuilder {
	b.activities = activities
	return b
}

// WithUseCases は UseCase 情報を設定
func (b *UnifiedGraphBuilder) WithUseCases(usecases []UseCaseInfo) *UnifiedGraphBuilder {
	b.usecases = usecases
	return b
}

// WithDeliverables は Deliverable 情報を設定
func (b *UnifiedGraphBuilder) WithDeliverables(deliverables []DeliverableInfo) *UnifiedGraphBuilder {
	b.deliverables = deliverables
	return b
}

// WithObjectives は Objective 情報を設定
func (b *UnifiedGraphBuilder) WithObjectives(objectives []ObjectiveInfo) *UnifiedGraphBuilder {
	b.objectives = objectives
	return b
}

// WithFilter はフィルタを設定
func (b *UnifiedGraphBuilder) WithFilter(filter *GraphFilter) *UnifiedGraphBuilder {
	if filter != nil {
		b.filter = filter
	}
	return b
}

// Build は統合グラフを構築
func (b *UnifiedGraphBuilder) Build() *UnifiedGraph {
	graph := &UnifiedGraph{
		Nodes:    make(map[string]*UnifiedGraphNode),
		Edges:    []UnifiedEdge{},
		Cycles:   [][]string{},
		Isolated: []string{},
		Stats: UnifiedGraphStats{
			NodesByType: make(map[EntityType]int),
			EdgesByType: make(map[UnifiedEdgeType]int),
		},
	}

	// 1. ノードを構築
	b.buildNodes(graph)

	// 2. エッジを構築
	b.buildEdges(graph)

	// 3. フィルタを適用
	b.applyFilter(graph)

	// 4. 深さを計算
	b.calculateDepth(graph)

	// 5. 循環を検出
	b.detectCycles(graph)

	// 6. 孤立ノードを検出
	b.detectIsolated(graph)

	// 7. 統計を計算
	b.calculateStats(graph)

	return graph
}

// buildNodes はノードを構築
func (b *UnifiedGraphBuilder) buildNodes(graph *UnifiedGraph) {
	// Activity ノード
	for _, a := range b.activities {
		node := &UnifiedGraphNode{
			ID:       a.ID,
			Type:     EntityTypeActivity,
			Title:    a.Title,
			Status:   a.Status,
			Priority: a.Priority,
			Assignee: a.Assignee,
			Mode:     a.Mode,
			Parents:  []string{},
			Children: []string{},
		}
		graph.Nodes[a.ID] = node
	}

	// UseCase ノード
	for _, u := range b.usecases {
		node := &UnifiedGraphNode{
			ID:       u.ID,
			Type:     EntityTypeUseCase,
			Title:    u.Title,
			Status:   u.Status,
			Parents:  []string{},
			Children: []string{},
		}
		graph.Nodes[u.ID] = node
	}

	// Deliverable ノード
	for _, d := range b.deliverables {
		node := &UnifiedGraphNode{
			ID:       d.ID,
			Type:     EntityTypeDeliverable,
			Title:    d.Title,
			Status:   d.Status,
			Parents:  []string{},
			Children: []string{},
		}
		graph.Nodes[d.ID] = node
	}

	// Objective ノード
	for _, o := range b.objectives {
		node := &UnifiedGraphNode{
			ID:       o.ID,
			Type:     EntityTypeObjective,
			Title:    o.Title,
			Status:   o.Status,
			Parents:  []string{},
			Children: []string{},
		}
		graph.Nodes[o.ID] = node
	}
}

// addHierarchyEdge は階層関係のエッジを追加し、Parents/Children を更新する
// 深さ計算（calculateDepth）に反映するため、階層関係では Parents/Children を更新する
func (b *UnifiedGraphBuilder) addHierarchyEdge(graph *UnifiedGraph, childID, parentID string, edgeType UnifiedEdgeType, label string) {
	childNode, childOk := graph.Nodes[childID]
	parentNode, parentOk := graph.Nodes[parentID]
	if !childOk || !parentOk {
		return
	}

	edge := UnifiedEdge{
		From:  childID,
		To:    parentID,
		Type:  edgeType,
		Label: label,
	}
	graph.Edges = append(graph.Edges, edge)

	// Parents/Children を更新（深さ計算に反映）
	childNode.Parents = append(childNode.Parents, parentID)
	parentNode.Children = append(parentNode.Children, childID)
}

// buildEdges はエッジを構築
func (b *UnifiedGraphBuilder) buildEdges(graph *UnifiedGraph) {
	// Activity の依存関係エッジ
	for _, a := range b.activities {
		// 依存関係
		for _, depID := range a.Dependencies {
			if _, exists := graph.Nodes[depID]; exists {
				edge := UnifiedEdge{
					From: a.ID,
					To:   depID,
					Type: EdgeTypeDependency,
				}
				graph.Edges = append(graph.Edges, edge)
				// ノードの Parents/Children を更新
				graph.Nodes[a.ID].Parents = append(graph.Nodes[a.ID].Parents, depID)
				graph.Nodes[depID].Children = append(graph.Nodes[depID].Children, a.ID)
			}
		}

		// 親子関係
		if a.ParentID != "" {
			if _, exists := graph.Nodes[a.ParentID]; exists {
				edge := UnifiedEdge{
					From: a.ID,
					To:   a.ParentID,
					Type: EdgeTypeParent,
				}
				graph.Edges = append(graph.Edges, edge)
			}
		}

		// Activity → UseCase 関連（階層関係: Activity は UseCase の下に配置）
		if a.UseCaseID != "" {
			b.addHierarchyEdge(graph, a.ID, a.UseCaseID, EdgeTypeRelates, "implements")
		}

		// Activity → Deliverable 関連
		// NOTE: produces 関係は「成果物への出力」を示し、階層関係ではないため
		// Parents/Children は更新しない（深さ計算に影響させない）
		for _, delID := range a.RelatedDeliverables {
			if _, exists := graph.Nodes[delID]; exists {
				edge := UnifiedEdge{
					From:  a.ID,
					To:    delID,
					Type:  EdgeTypeRelates,
					Label: "produces",
				}
				graph.Edges = append(graph.Edges, edge)
			}
		}
	}

	// UseCase → Objective 関連（階層関係: UseCase は Objective の下に配置）
	for _, u := range b.usecases {
		if u.ObjectiveID != "" {
			b.addHierarchyEdge(graph, u.ID, u.ObjectiveID, EdgeTypeContributes, "contributes")
		}
	}

	// Deliverable → Objective 関連（階層関係: Deliverable は Objective の下に配置）
	for _, d := range b.deliverables {
		if d.ObjectiveID != "" {
			b.addHierarchyEdge(graph, d.ID, d.ObjectiveID, EdgeTypeContributes, "fulfills")
		}
	}

	// Objective の親子関係
	for _, o := range b.objectives {
		if o.ParentID != "" {
			if _, exists := graph.Nodes[o.ParentID]; exists {
				edge := UnifiedEdge{
					From: o.ID,
					To:   o.ParentID,
					Type: EdgeTypeParent,
				}
				graph.Edges = append(graph.Edges, edge)
			}
		}
	}
}

// applyFilter はフィルタを適用
func (b *UnifiedGraphBuilder) applyFilter(graph *UnifiedGraph) {
	if b.filter == nil {
		return
	}

	// 削除するノード ID を収集
	toRemove := make(map[string]bool)

	for id, node := range graph.Nodes {
		// タイプフィルタ
		if len(b.filter.IncludeTypes) > 0 {
			if !slices.Contains(b.filter.IncludeTypes, node.Type) {
				toRemove[id] = true
				continue
			}
		}

		if slices.Contains(b.filter.ExcludeTypes, node.Type) {
			toRemove[id] = true
		}

		// 完了済みフィルタ
		if b.filter.HideCompleted {
			if node.Status == "completed" {
				toRemove[id] = true
				continue
			}
		}

		// ドラフトフィルタ
		if b.filter.HideDraft {
			if node.Status == "draft" {
				toRemove[id] = true
				continue
			}
		}
	}

	// フォーカスフィルタ
	if b.filter.FocusID != "" && b.filter.FocusDepth > 0 {
		reachable := b.findReachableNodes(graph, b.filter.FocusID, b.filter.FocusDepth)
		for id := range graph.Nodes {
			if !reachable[id] {
				toRemove[id] = true
			}
		}
	}

	// ノードを削除
	for id := range toRemove {
		delete(graph.Nodes, id)
	}

	// エッジをフィルタ（両端がグラフに存在するもののみ残す）
	filteredEdges := []UnifiedEdge{}
	for _, edge := range graph.Edges {
		if _, fromExists := graph.Nodes[edge.From]; !fromExists {
			continue
		}
		if _, toExists := graph.Nodes[edge.To]; !toExists {
			continue
		}
		filteredEdges = append(filteredEdges, edge)
	}
	graph.Edges = filteredEdges

	// Parents/Children を更新
	for _, node := range graph.Nodes {
		filteredParents := []string{}
		for _, p := range node.Parents {
			if _, exists := graph.Nodes[p]; exists {
				filteredParents = append(filteredParents, p)
			}
		}
		node.Parents = filteredParents

		filteredChildren := []string{}
		for _, c := range node.Children {
			if _, exists := graph.Nodes[c]; exists {
				filteredChildren = append(filteredChildren, c)
			}
		}
		node.Children = filteredChildren
	}
}

// findReachableNodes はフォーカスノードから指定深さまで到達可能なノードを検索
// パフォーマンス改善: 隣接リストを事前構築して O(V + E) で探索
func (b *UnifiedGraphBuilder) findReachableNodes(graph *UnifiedGraph, focusID string, depth int) map[string]bool {
	reachable := make(map[string]bool)
	if _, exists := graph.Nodes[focusID]; !exists {
		return reachable
	}

	// 隣接リストを事前構築（双方向）
	adjacency := make(map[string][]string)
	for _, edge := range graph.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
		adjacency[edge.To] = append(adjacency[edge.To], edge.From)
	}

	// BFS で探索
	type queueItem struct {
		id    string
		depth int
	}
	queue := []queueItem{{focusID, 0}}
	visited := make(map[string]bool)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.id] {
			continue
		}
		visited[current.id] = true
		reachable[current.id] = true

		if current.depth >= depth {
			continue
		}

		node := graph.Nodes[current.id]
		if node == nil {
			continue
		}

		// 親と子を探索
		for _, parentID := range node.Parents {
			if !visited[parentID] {
				queue = append(queue, queueItem{parentID, current.depth + 1})
			}
		}
		for _, childID := range node.Children {
			if !visited[childID] {
				queue = append(queue, queueItem{childID, current.depth + 1})
			}
		}

		// 隣接リストから接続ノードを探索
		for _, neighborID := range adjacency[current.id] {
			if !visited[neighborID] {
				queue = append(queue, queueItem{neighborID, current.depth + 1})
			}
		}
	}

	return reachable
}

// calculateDepth はノードの深さを計算
func (b *UnifiedGraphBuilder) calculateDepth(graph *UnifiedGraph) {
	// 依存関係がないノードを起点として深さを計算
	for _, node := range graph.Nodes {
		if len(node.Parents) == 0 {
			node.Depth = 0
		} else {
			node.Depth = -1 // 未計算
		}
	}

	// 繰り返し深さを更新
	changed := true
	for changed {
		changed = false
		for _, node := range graph.Nodes {
			if node.Depth >= 0 {
				for _, childID := range node.Children {
					child := graph.Nodes[childID]
					if child != nil && (child.Depth < 0 || child.Depth < node.Depth+1) {
						child.Depth = node.Depth + 1
						changed = true
					}
				}
			}
		}
	}

	// 未計算のノード（循環の一部）には深さ 0 を設定
	for _, node := range graph.Nodes {
		if node.Depth < 0 {
			node.Depth = 0
		}
	}
}

// detectCycles は循環依存を検出
func (b *UnifiedGraphBuilder) detectCycles(graph *UnifiedGraph) {
	// 訪問状態: 0=未訪問, 1=訪問中, 2=訪問済み
	visited := make(map[string]int)
	var path []string

	var dfs func(id string)
	dfs = func(id string) {
		if visited[id] == 2 {
			return
		}
		if visited[id] == 1 {
			// 循環検出
			cycleStart := -1
			for i, pid := range path {
				if pid == id {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := make([]string, len(path)-cycleStart)
				copy(cycle, path[cycleStart:])
				cycle = append(cycle, id)
				graph.Cycles = append(graph.Cycles, cycle)
			}
			return
		}

		visited[id] = 1
		path = append(path, id)

		node := graph.Nodes[id]
		if node != nil {
			for _, parentID := range node.Parents {
				dfs(parentID)
			}
		}

		path = path[:len(path)-1]
		visited[id] = 2
	}

	for id := range graph.Nodes {
		if visited[id] == 0 {
			dfs(id)
		}
	}
}

// detectIsolated は孤立ノードを検出
func (b *UnifiedGraphBuilder) detectIsolated(graph *UnifiedGraph) {
	for id, node := range graph.Nodes {
		hasEdge := false
		for _, edge := range graph.Edges {
			if edge.From == id || edge.To == id {
				hasEdge = true
				break
			}
		}
		if !hasEdge && len(node.Parents) == 0 && len(node.Children) == 0 {
			graph.Isolated = append(graph.Isolated, id)
		}
	}
	sort.Strings(graph.Isolated)
}

// calculateStats は統計を計算
func (b *UnifiedGraphBuilder) calculateStats(graph *UnifiedGraph) {
	stats := &graph.Stats
	stats.TotalNodes = len(graph.Nodes)
	stats.TotalEdges = len(graph.Edges)
	stats.IsolatedCount = len(graph.Isolated)
	stats.CycleCount = len(graph.Cycles)

	// タイプ別ノード数
	for _, node := range graph.Nodes {
		stats.NodesByType[node.Type]++
		if node.Type == EntityTypeActivity {
			stats.TotalActivities++
			if node.Status == "completed" {
				stats.CompletedActivities++
			}
		}
	}

	// タイプ別エッジ数
	for _, edge := range graph.Edges {
		stats.EdgesByType[edge.Type]++
	}

	// 最大深さ
	for _, node := range graph.Nodes {
		if node.Depth > stats.MaxDepth {
			stats.MaxDepth = node.Depth
		}
	}
}

// ToText はテキスト形式で出力
func (g *UnifiedGraph) ToText() string {
	var sb strings.Builder

	sb.WriteString("=== Unified Graph ===\n\n")

	// 統計情報
	fmt.Fprintf(&sb, "Total Nodes: %d\n", g.Stats.TotalNodes)
	fmt.Fprintf(&sb, "Total Edges: %d\n", g.Stats.TotalEdges)
	fmt.Fprintf(&sb, "Max Depth: %d\n", g.Stats.MaxDepth)
	if g.Stats.TotalActivities > 0 {
		fmt.Fprintf(&sb, "Activity Progress: %d/%d\n", g.Stats.CompletedActivities, g.Stats.TotalActivities)
	}
	sb.WriteString("\n")

	// ノード別
	sb.WriteString("--- Nodes by Type ---\n")
	for t, count := range g.Stats.NodesByType {
		fmt.Fprintf(&sb, "  %s: %d\n", t, count)
	}
	sb.WriteString("\n")

	// ノード一覧
	sb.WriteString("--- Nodes ---\n")
	// ID でソート
	ids := make([]string, 0, len(g.Nodes))
	for id := range g.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		node := g.Nodes[id]
		fmt.Fprintf(&sb, "[%s] %s (%s) - %s", node.Type, node.ID, node.Title, node.Status)
		if node.Assignee != "" {
			fmt.Fprintf(&sb, " @%s", node.Assignee)
		}
		sb.WriteString("\n")
		if len(node.Parents) > 0 {
			fmt.Fprintf(&sb, "    depends on: %v\n", node.Parents)
		}
	}
	sb.WriteString("\n")

	// 循環依存
	if len(g.Cycles) > 0 {
		sb.WriteString("--- Cycles Detected ---\n")
		for _, cycle := range g.Cycles {
			fmt.Fprintf(&sb, "  %s\n", strings.Join(cycle, " -> "))
		}
		sb.WriteString("\n")
	}

	// 孤立ノード
	if len(g.Isolated) > 0 {
		sb.WriteString("--- Isolated Nodes ---\n")
		for _, id := range g.Isolated {
			fmt.Fprintf(&sb, "  %s\n", id)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// ToDot は DOT 形式で出力
func (g *UnifiedGraph) ToDot() string {
	var sb strings.Builder

	sb.WriteString("digraph UnifiedGraph {\n")
	sb.WriteString("  rankdir=TB;\n")
	sb.WriteString("  node [shape=box];\n\n")

	// ノードのスタイル定義
	sb.WriteString("  // Node styles by type\n")
	for id, node := range g.Nodes {
		var style, color string
		switch node.Type {
		case EntityTypeActivity:
			if node.Mode == "flow" {
				style = "rounded"
			} else {
				style = "filled,rounded"
			}
			color = "#4CAF50"
		case EntityTypeUseCase:
			style = "ellipse"
			color = "#2196F3"
		case EntityTypeDeliverable:
			style = "folder"
			color = "#FF9800"
		case EntityTypeObjective:
			style = "diamond"
			color = "#9C27B0"
		}

		label := fmt.Sprintf("%s\\n%s", node.ID, node.Title)

		fmt.Fprintf(&sb, "  \"%s\" [label=\"%s\", style=\"%s\", fillcolor=\"%s\"];\n",
			id, label, style, color)
	}
	sb.WriteString("\n")

	// エッジ
	sb.WriteString("  // Edges\n")
	for _, edge := range g.Edges {
		var style, color string
		switch edge.Type {
		case EdgeTypeDependency:
			style = "solid"
			color = "black"
		case EdgeTypeParent:
			style = "dashed"
			color = "gray"
		case EdgeTypeRelates:
			style = "dotted"
			color = "blue"
		case EdgeTypeContributes:
			style = "bold"
			color = "purple"
		}

		label := ""
		if edge.Label != "" {
			label = fmt.Sprintf(", label=\"%s\"", edge.Label)
		}

		fmt.Fprintf(&sb, "  \"%s\" -> \"%s\" [style=%s, color=\"%s\"%s];\n",
			edge.From, edge.To, style, color, label)
	}

	sb.WriteString("}\n")
	return sb.String()
}

// ToMermaid は Mermaid 形式で出力
func (g *UnifiedGraph) ToMermaid() string {
	var sb strings.Builder

	sb.WriteString("```mermaid\ngraph TD\n")

	// ノード定義
	for id, node := range g.Nodes {
		var shape string
		title := escapeMermaidText(node.Title)
		switch node.Type {
		case EntityTypeActivity:
			if node.Mode == "flow" {
				shape = fmt.Sprintf("  %s([%s: %s])", id, id, title)
			} else {
				shape = fmt.Sprintf("  %s[%s: %s]", id, id, title)
			}
		case EntityTypeUseCase:
			shape = fmt.Sprintf("  %s((%s: %s))", id, id, title)
		case EntityTypeDeliverable:
			shape = fmt.Sprintf("  %s[/%s: %s/]", id, id, title)
		case EntityTypeObjective:
			shape = fmt.Sprintf("  %s{%s: %s}", id, id, title)
		}
		sb.WriteString(shape + "\n")
	}

	sb.WriteString("\n")

	// エッジ定義
	for _, edge := range g.Edges {
		var arrow string
		switch edge.Type {
		case EdgeTypeDependency:
			arrow = "-->"
		case EdgeTypeParent:
			arrow = "-.->|parent|"
		case EdgeTypeRelates:
			if edge.Label != "" {
				arrow = fmt.Sprintf("-.->|%s|", edge.Label)
			} else {
				arrow = "-.->"
			}
		case EdgeTypeContributes:
			if edge.Label != "" {
				arrow = fmt.Sprintf("==>|%s|", edge.Label)
			} else {
				arrow = "==>"
			}
		}
		fmt.Fprintf(&sb, "  %s %s %s\n", edge.From, arrow, edge.To)
	}

	// スタイル
	sb.WriteString("\n  %% Styles\n")
	sb.WriteString("  classDef activity fill:#4CAF50,stroke:#333,color:#fff\n")
	sb.WriteString("  classDef usecase fill:#2196F3,stroke:#333,color:#fff\n")
	sb.WriteString("  classDef deliverable fill:#FF9800,stroke:#333,color:#fff\n")
	sb.WriteString("  classDef objective fill:#9C27B0,stroke:#333,color:#fff\n")

	// クラス適用
	for id, node := range g.Nodes {
		var class string
		switch node.Type {
		case EntityTypeActivity:
			class = "activity"
		case EntityTypeUseCase:
			class = "usecase"
		case EntityTypeDeliverable:
			class = "deliverable"
		case EntityTypeObjective:
			class = "objective"
		}
		fmt.Fprintf(&sb, "  class %s %s\n", id, class)
	}

	sb.WriteString("```\n")
	return sb.String()
}

// escapeMermaidText は Mermaid 構文で問題になる特殊文字をエスケープ
func escapeMermaidText(s string) string {
	replacer := strings.NewReplacer(
		"[", "#91;",
		"]", "#93;",
		"{", "#123;",
		"}", "#125;",
		"(", "#40;",
		")", "#41;",
		"<", "#60;",
		">", "#62;",
		"|", "#124;",
		"\"", "#34;",
	)
	return replacer.Replace(s)
}
