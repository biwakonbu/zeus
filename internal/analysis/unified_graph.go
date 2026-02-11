package analysis

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// UnifiedGraphBuilder は統合グラフを構築する
type UnifiedGraphBuilder struct {
	activities       []ActivityInfo
	usecases         []UseCaseInfo
	objectives       []ObjectiveInfo
	filter           *GraphFilter
	validationErrors []string
}

// NewUnifiedGraphBuilder は新しい UnifiedGraphBuilder を作成
func NewUnifiedGraphBuilder() *UnifiedGraphBuilder {
	return &UnifiedGraphBuilder{
		activities: []ActivityInfo{},
		usecases:   []UseCaseInfo{},
		objectives: []ObjectiveInfo{},
		filter:     NewGraphFilter(),
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

// ValidationErrors は構築時に検出された関係定義違反を返す
func (b *UnifiedGraphBuilder) ValidationErrors() []string {
	out := make([]string, len(b.validationErrors))
	copy(out, b.validationErrors)
	return out
}

// Build は統合グラフを構築
func (b *UnifiedGraphBuilder) Build() *UnifiedGraph {
	b.validationErrors = nil

	graph := &UnifiedGraph{
		Nodes:    make(map[string]*UnifiedGraphNode),
		Edges:    []UnifiedEdge{},
		Cycles:   [][]string{},
		Isolated: []string{},
		Stats: UnifiedGraphStats{
			NodesByType:     make(map[EntityType]int),
			EdgesByLayer:    make(map[UnifiedEdgeLayer]int),
			EdgesByRelation: make(map[UnifiedEdgeRelation]int),
		},
	}

	// 1. ノードを構築
	b.buildNodes(graph)

	// 2. エッジを構築
	b.buildEdges(graph)

	// 3. フィルタを適用
	b.applyFilter(graph)

	// 4. 構造層深さを計算
	b.calculateStructuralDepth(graph)

	// 5. 循環を検出
	b.detectCycles(graph)

	// 6. 孤立ノードを検出
	b.detectIsolated(graph)

	// 7. 統計を計算
	b.calculateStats(graph)

	return graph
}

// validateEdgeRule は relation/layer の許容行列を検証する
func validateEdgeRule(fromType, toType EntityType, layer UnifiedEdgeLayer, relation UnifiedEdgeRelation) error {
	// レイヤーと relation の整合性
	switch layer {
	case EdgeLayerStructural:
		if !slices.Contains([]UnifiedEdgeRelation{RelationImplements, RelationContributes}, relation) {
			return fmt.Errorf("relation %q is not allowed in structural layer", relation)
		}
	default:
		return fmt.Errorf("unknown edge layer: %q", layer)
	}

	// relation ごとの許容型
	switch relation {
	case RelationImplements:
		if fromType == EntityTypeActivity && toType == EntityTypeUseCase {
			return nil
		}
		return fmt.Errorf("implements must connect activity->usecase (got %s->%s)", fromType, toType)
	case RelationContributes:
		if fromType == EntityTypeUseCase && toType == EntityTypeObjective {
			return nil
		}
		return fmt.Errorf("contributes must connect usecase->objective (got %s->%s)", fromType, toType)
	default:
		return fmt.Errorf("unknown edge relation: %q", relation)
	}
}

// addEdge は関係ルールを検証した上でエッジを追加
func (b *UnifiedGraphBuilder) addEdge(graph *UnifiedGraph, fromID, toID string, layer UnifiedEdgeLayer, relation UnifiedEdgeRelation) bool {
	fromNode, fromOK := graph.Nodes[fromID]
	toNode, toOK := graph.Nodes[toID]
	if !fromOK || !toOK {
		return false
	}

	if err := validateEdgeRule(fromNode.Type, toNode.Type, layer, relation); err != nil {
		b.validationErrors = append(b.validationErrors,
			fmt.Sprintf("invalid edge %s -> %s (%s/%s): %v", fromID, toID, layer, relation, err))
		return false
	}

	graph.Edges = append(graph.Edges, UnifiedEdge{
		From:     fromID,
		To:       toID,
		Layer:    layer,
		Relation: relation,
	})

	return true
}

// addStructuralEdge は structural エッジを追加し、構造親子関係を更新する
func (b *UnifiedGraphBuilder) addStructuralEdge(graph *UnifiedGraph, childID, parentID string, relation UnifiedEdgeRelation) {
	if !b.addEdge(graph, childID, parentID, EdgeLayerStructural, relation) {
		return
	}

	childNode := graph.Nodes[childID]
	parentNode := graph.Nodes[parentID]
	if childNode == nil || parentNode == nil {
		return
	}
	childNode.StructuralParents = append(childNode.StructuralParents, parentID)
	parentNode.StructuralChildren = append(parentNode.StructuralChildren, childID)
}

// buildNodes はノードを構築
func (b *UnifiedGraphBuilder) buildNodes(graph *UnifiedGraph) {
	// Activity ノード
	for _, a := range b.activities {
		node := &UnifiedGraphNode{
			ID:                 a.ID,
			Type:               EntityTypeActivity,
			Title:              a.Title,
			Status:             a.Status,
			StructuralParents:  []string{},
			StructuralChildren: []string{},
		}
		graph.Nodes[a.ID] = node
	}

	// UseCase ノード
	for _, u := range b.usecases {
		node := &UnifiedGraphNode{
			ID:                 u.ID,
			Type:               EntityTypeUseCase,
			Title:              u.Title,
			Status:             u.Status,
			StructuralParents:  []string{},
			StructuralChildren: []string{},
		}
		graph.Nodes[u.ID] = node
	}

	// Objective ノード
	for _, o := range b.objectives {
		node := &UnifiedGraphNode{
			ID:                 o.ID,
			Type:               EntityTypeObjective,
			Title:              o.Title,
			Status:             o.Status,
			StructuralParents:  []string{},
			StructuralChildren: []string{},
		}
		graph.Nodes[o.ID] = node
	}
}

// buildEdges はエッジを構築
func (b *UnifiedGraphBuilder) buildEdges(graph *UnifiedGraph) {
	// implements (structural): Activity -> UseCase
	for _, a := range b.activities {
		if a.UseCaseID != "" {
			b.addStructuralEdge(graph, a.ID, a.UseCaseID, RelationImplements)
		}
	}

	// contributes (structural): UseCase -> Objective
	for _, u := range b.usecases {
		if u.ObjectiveID != "" {
			b.addStructuralEdge(graph, u.ID, u.ObjectiveID, RelationContributes)
		}
	}
}

// applyFilter はフィルタを適用
func (b *UnifiedGraphBuilder) applyFilter(graph *UnifiedGraph) {
	if b.filter == nil {
		return
	}

	// 1. ノード削除候補を収集
	toRemove := make(map[string]bool)
	for id, node := range graph.Nodes {
		if len(b.filter.IncludeTypes) > 0 && !slices.Contains(b.filter.IncludeTypes, node.Type) {
			toRemove[id] = true
			continue
		}
		if slices.Contains(b.filter.ExcludeTypes, node.Type) {
			toRemove[id] = true
			continue
		}
		if b.filter.HideCompleted && (node.Status == "completed" || node.Status == "deprecated") {
			toRemove[id] = true
			continue
		}
		if b.filter.HideDraft && node.Status == "draft" {
			toRemove[id] = true
			continue
		}
	}

	for id := range toRemove {
		delete(graph.Nodes, id)
	}

	// 2. エッジをフィルタ（存在ノード + レイヤー + relation）
	graph.Edges = b.filterEdgesByRule(graph.Nodes, graph.Edges)

	// 3. フォーカスフィルタ
	if b.filter.FocusID != "" && b.filter.FocusDepth > 0 {
		reachable := b.findReachableNodes(graph, b.filter.FocusID, b.filter.FocusDepth)
		for id := range graph.Nodes {
			if !reachable[id] {
				delete(graph.Nodes, id)
			}
		}
		graph.Edges = b.filterEdgesByRule(graph.Nodes, graph.Edges)
	}

	// 4. 構造親子を再構築
	b.rebuildStructuralAdjacency(graph)
}

func (b *UnifiedGraphBuilder) filterEdgesByRule(nodes map[string]*UnifiedGraphNode, edges []UnifiedEdge) []UnifiedEdge {
	filtered := make([]UnifiedEdge, 0, len(edges))
	for _, edge := range edges {
		if _, fromExists := nodes[edge.From]; !fromExists {
			continue
		}
		if _, toExists := nodes[edge.To]; !toExists {
			continue
		}
		if len(b.filter.IncludeLayers) > 0 && !slices.Contains(b.filter.IncludeLayers, edge.Layer) {
			continue
		}
		if len(b.filter.IncludeRelations) > 0 && !slices.Contains(b.filter.IncludeRelations, edge.Relation) {
			continue
		}
		filtered = append(filtered, edge)
	}
	return filtered
}

func (b *UnifiedGraphBuilder) rebuildStructuralAdjacency(graph *UnifiedGraph) {
	for _, node := range graph.Nodes {
		node.StructuralParents = []string{}
		node.StructuralChildren = []string{}
	}

	for _, edge := range graph.Edges {
		if edge.Layer != EdgeLayerStructural {
			continue
		}
		fromNode := graph.Nodes[edge.From]
		toNode := graph.Nodes[edge.To]
		if fromNode == nil || toNode == nil {
			continue
		}
		fromNode.StructuralParents = append(fromNode.StructuralParents, edge.To)
		toNode.StructuralChildren = append(toNode.StructuralChildren, edge.From)
	}
}

// findReachableNodes はフォーカスノードから指定深さまで到達可能なノードを検索
func (b *UnifiedGraphBuilder) findReachableNodes(graph *UnifiedGraph, focusID string, depth int) map[string]bool {
	reachable := make(map[string]bool)
	if _, exists := graph.Nodes[focusID]; !exists {
		return reachable
	}

	adjacency := make(map[string][]string)
	for _, edge := range graph.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
		adjacency[edge.To] = append(adjacency[edge.To], edge.From)
	}

	type queueItem struct {
		id    string
		depth int
	}
	queue := []queueItem{{id: focusID, depth: 0}}
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

		for _, neighborID := range adjacency[current.id] {
			if !visited[neighborID] {
				queue = append(queue, queueItem{id: neighborID, depth: current.depth + 1})
			}
		}
	}

	return reachable
}

// calculateStructuralDepth は構造層ノードの深さを計算
func (b *UnifiedGraphBuilder) calculateStructuralDepth(graph *UnifiedGraph) {
	for _, node := range graph.Nodes {
		if len(node.StructuralParents) == 0 {
			node.StructuralDepth = 0
		} else {
			node.StructuralDepth = -1
		}
	}

	changed := true
	for changed {
		changed = false
		for _, node := range graph.Nodes {
			if node.StructuralDepth >= 0 {
				for _, childID := range node.StructuralChildren {
					child := graph.Nodes[childID]
					if child != nil && (child.StructuralDepth < 0 || child.StructuralDepth < node.StructuralDepth+1) {
						child.StructuralDepth = node.StructuralDepth + 1
						changed = true
					}
				}
			}
		}
	}

	for _, node := range graph.Nodes {
		if node.StructuralDepth < 0 {
			node.StructuralDepth = 0
		}
	}
}

// detectCycles は循環依存を検出（全レイヤー対象）
func (b *UnifiedGraphBuilder) detectCycles(graph *UnifiedGraph) {
	visited := make(map[string]int) // 0=未訪問, 1=訪問中, 2=訪問済み
	path := []string{}

	adj := make(map[string][]string)
	for _, edge := range graph.Edges {
		adj[edge.From] = append(adj[edge.From], edge.To)
	}

	var dfs func(id string)
	dfs = func(id string) {
		if visited[id] == 2 {
			return
		}
		if visited[id] == 1 {
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
		for _, nextID := range adj[id] {
			dfs(nextID)
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
	for id := range graph.Nodes {
		hasEdge := false
		for _, edge := range graph.Edges {
			if edge.From == id || edge.To == id {
				hasEdge = true
				break
			}
		}
		if !hasEdge {
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

	for _, node := range graph.Nodes {
		stats.NodesByType[node.Type]++
		if node.Type == EntityTypeActivity {
			stats.TotalActivities++
			if node.Status == "deprecated" {
				stats.CompletedActivities++
			}
		}
	}

	for _, edge := range graph.Edges {
		stats.EdgesByLayer[edge.Layer]++
		stats.EdgesByRelation[edge.Relation]++
	}

	for _, node := range graph.Nodes {
		if node.StructuralDepth > stats.MaxStructuralDepth {
			stats.MaxStructuralDepth = node.StructuralDepth
		}
	}
}

func sortEdges(edges []UnifiedEdge) {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		if edges[i].To != edges[j].To {
			return edges[i].To < edges[j].To
		}
		if edges[i].Layer != edges[j].Layer {
			return edges[i].Layer < edges[j].Layer
		}
		return edges[i].Relation < edges[j].Relation
	})
}

// ToText はテキスト形式で出力
func (g *UnifiedGraph) ToText() string {
	var sb strings.Builder

	sb.WriteString("=== Unified Graph (Two-Layer Model) ===\n\n")
	fmt.Fprintf(&sb, "Total Nodes: %d\n", g.Stats.TotalNodes)
	fmt.Fprintf(&sb, "Total Edges: %d\n", g.Stats.TotalEdges)
	fmt.Fprintf(&sb, "Max Structural Depth: %d\n", g.Stats.MaxStructuralDepth)
	if g.Stats.TotalActivities > 0 {
		fmt.Fprintf(&sb, "Deprecated Activities: %d/%d\n", g.Stats.CompletedActivities, g.Stats.TotalActivities)
	}
	sb.WriteString("\n")

	sb.WriteString("--- Nodes by Type ---\n")
	types := make([]string, 0, len(g.Stats.NodesByType))
	for t := range g.Stats.NodesByType {
		types = append(types, string(t))
	}
	sort.Strings(types)
	for _, t := range types {
		fmt.Fprintf(&sb, "  %s: %d\n", t, g.Stats.NodesByType[EntityType(t)])
	}
	sb.WriteString("\n")

	sb.WriteString("--- Nodes ---\n")
	ids := make([]string, 0, len(g.Nodes))
	for id := range g.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		node := g.Nodes[id]
		fmt.Fprintf(&sb, "[%s] %s (%s) - %s [depth=%d]\n", node.Type, node.ID, node.Title, node.Status, node.StructuralDepth)
		if len(node.StructuralParents) > 0 {
			fmt.Fprintf(&sb, "    structural_parents: %v\n", node.StructuralParents)
		}
	}
	sb.WriteString("\n")

	sortEdges(g.Edges)

	sb.WriteString("--- Structural Relations ---\n")
	for _, edge := range g.Edges {
		fmt.Fprintf(&sb, "  %s -[%s]-> %s\n", edge.From, edge.Relation, edge.To)
	}
	sb.WriteString("\n")

	if len(g.Cycles) > 0 {
		sb.WriteString("--- Cycles Detected ---\n")
		for _, cycle := range g.Cycles {
			fmt.Fprintf(&sb, "  %s\n", strings.Join(cycle, " -> "))
		}
		sb.WriteString("\n")
	}

	if len(g.Isolated) > 0 {
		sb.WriteString("--- Isolated Nodes ---\n")
		for _, id := range g.Isolated {
			fmt.Fprintf(&sb, "  %s\n", id)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func dotEdgeStyle(edge UnifiedEdge) (style string, color string, label string) {
	switch edge.Relation {
	case RelationImplements:
		return "bold", "#1f77b4", "implements"
	case RelationContributes:
		return "bold", "#7f3fbf", "contributes"
	default:
		return "solid", "#333333", string(edge.Relation)
	}
}

// ToDot は DOT 形式で出力
func (g *UnifiedGraph) ToDot() string {
	var sb strings.Builder

	sb.WriteString("digraph UnifiedGraph {\n")
	sb.WriteString("  rankdir=TB;\n")
	sb.WriteString("  node [shape=box];\n\n")

	sb.WriteString("  // Node styles by type\n")
	for id, node := range g.Nodes {
		var shape, fill string
		switch node.Type {
		case EntityTypeActivity:
			shape = "box"
			fill = "#4CAF50"
		case EntityTypeUseCase:
			shape = "ellipse"
			fill = "#2196F3"
		case EntityTypeObjective:
			shape = "diamond"
			fill = "#9C27B0"
		default:
			shape = "box"
			fill = "#888888"
		}
		label := fmt.Sprintf("%s\\n%s", node.ID, node.Title)
		fmt.Fprintf(&sb, "  \"%s\" [label=\"%s\", shape=\"%s\", style=\"filled\", fillcolor=\"%s\"];\n", id, label, shape, fill)
	}
	sb.WriteString("\n")

	edges := make([]UnifiedEdge, len(g.Edges))
	copy(edges, g.Edges)
	sortEdges(edges)

	sb.WriteString("  // Edges\n")
	for _, edge := range edges {
		style, color, label := dotEdgeStyle(edge)
		fmt.Fprintf(&sb, "  \"%s\" -> \"%s\" [style=%s, color=\"%s\", label=\"%s\"];\n", edge.From, edge.To, style, color, label)
	}

	sb.WriteString("}\n")
	return sb.String()
}

func mermaidArrow(edge UnifiedEdge) string {
	switch edge.Relation {
	case RelationImplements:
		return "==>|implements|"
	case RelationContributes:
		return "==>|contributes|"
	default:
		return "-->"
	}
}

// ToMermaid は Mermaid 形式で出力
func (g *UnifiedGraph) ToMermaid() string {
	var sb strings.Builder

	sb.WriteString("```mermaid\ngraph TD\n")

	ids := make([]string, 0, len(g.Nodes))
	for id := range g.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		node := g.Nodes[id]
		title := escapeMermaidText(node.Title)
		switch node.Type {
		case EntityTypeActivity:
			fmt.Fprintf(&sb, "  %s([%s: %s])\n", id, id, title)
		case EntityTypeUseCase:
			fmt.Fprintf(&sb, "  %s((%s: %s))\n", id, id, title)
		case EntityTypeObjective:
			fmt.Fprintf(&sb, "  %s{%s: %s}\n", id, id, title)
		}
	}
	sb.WriteString("\n")

	edges := make([]UnifiedEdge, len(g.Edges))
	copy(edges, g.Edges)
	sortEdges(edges)

	for _, edge := range edges {
		fmt.Fprintf(&sb, "  %s %s %s\n", edge.From, mermaidArrow(edge), edge.To)
	}

	sb.WriteString("\n  %% Styles\n")
	sb.WriteString("  classDef activity fill:#4CAF50,stroke:#333,color:#fff\n")
	sb.WriteString("  classDef usecase fill:#2196F3,stroke:#333,color:#fff\n")
	sb.WriteString("  classDef objective fill:#9C27B0,stroke:#333,color:#fff\n")

	for _, id := range ids {
		node := g.Nodes[id]
		class := "activity"
		switch node.Type {
		case EntityTypeUseCase:
			class = "usecase"
		case EntityTypeObjective:
			class = "objective"
		}
		fmt.Fprintf(&sb, "  class %s %s\n", id, class)
	}

	for i, edge := range edges {
		if edge.Layer == EdgeLayerStructural {
			fmt.Fprintf(&sb, "  linkStyle %d stroke:#444,stroke-width:2px\n", i)
		}
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
