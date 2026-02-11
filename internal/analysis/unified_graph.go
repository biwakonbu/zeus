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

	// 1. ノードを構築（Objective はノードではなくグループとして扱う）
	b.buildNodes(graph)

	// 2. エッジを構築（implements のみ、contributes は廃止）
	b.buildEdges(graph)

	// 3. グループを構築（Objective ベース）
	b.buildGroups(graph)

	// 4. フィルタを適用
	b.applyFilter(graph)

	// 5. 構造層深さを計算
	b.calculateStructuralDepth(graph)

	// 6. 循環を検出
	b.detectCycles(graph)

	// 7. 孤立ノードを検出
	b.detectIsolated(graph)

	// 8. 統計を計算
	b.calculateStats(graph)

	return graph
}

// validateEdgeRule は relation/layer の許容行列を検証する
func validateEdgeRule(fromType, toType EntityType, layer UnifiedEdgeLayer, relation UnifiedEdgeRelation) error {
	// レイヤーと relation の整合性
	switch layer {
	case EdgeLayerStructural:
		if relation != RelationImplements {
			return fmt.Errorf("relation %q is not allowed in structural layer (only 'implements')", relation)
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

	// Objective はノードではなくグループとして扱う（buildGroups で処理）
}

// buildEdges はエッジを構築（implements のみ、contributes は廃止）
func (b *UnifiedGraphBuilder) buildEdges(graph *UnifiedGraph) {
	// implements (structural): Activity -> UseCase
	for _, a := range b.activities {
		if a.UseCaseID != "" {
			b.addStructuralEdge(graph, a.ID, a.UseCaseID, RelationImplements)
		}
	}

	// contributes は廃止: UseCase -> Objective の関係はグループで表現
}

// buildGroups は Objective ベースのグループを構築
func (b *UnifiedGraphBuilder) buildGroups(graph *UnifiedGraph) {
	// UseCase の objective_id から逆引きマップ構築
	usecasesByObjective := make(map[string][]string) // objective_id -> []usecase_id
	for _, u := range b.usecases {
		if u.ObjectiveID != "" {
			usecasesByObjective[u.ObjectiveID] = append(usecasesByObjective[u.ObjectiveID], u.ID)
		}
	}

	// Activity を UseCase 経由で Objective に割り当て
	activitiesByUseCase := make(map[string][]string) // usecase_id -> []activity_id
	for _, a := range b.activities {
		if a.UseCaseID != "" {
			activitiesByUseCase[a.UseCaseID] = append(activitiesByUseCase[a.UseCaseID], a.ID)
		}
	}

	// 各 Objective に対してグループを構築
	for _, obj := range b.objectives {
		seen := make(map[string]bool)
		var nodeIDs []string

		// UseCase を追加
		for _, ucID := range usecasesByObjective[obj.ID] {
			if _, exists := graph.Nodes[ucID]; exists && !seen[ucID] {
				nodeIDs = append(nodeIDs, ucID)
				seen[ucID] = true
			}
			// UseCase に紐づく Activity を追加
			for _, actID := range activitiesByUseCase[ucID] {
				if _, exists := graph.Nodes[actID]; exists && !seen[actID] {
					nodeIDs = append(nodeIDs, actID)
					seen[actID] = true
				}
			}
		}

		sort.Strings(nodeIDs)

		graph.Groups = append(graph.Groups, UnifiedGraphGroup{
			ID:          obj.ID,
			Title:       obj.Title,
			Description: obj.Description,
			Goals:       obj.Goals,
			Status:      obj.Status,
			Owner:       obj.Owner,
			Tags:        obj.Tags,
			NodeIDs:     nodeIDs,
		})
	}

	// グループを ID でソート
	sort.Slice(graph.Groups, func(i, j int) bool {
		return graph.Groups[i].ID < graph.Groups[j].ID
	})
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

	// 3. グループフィルタ（GroupIDs 指定時）
	if len(b.filter.GroupIDs) > 0 {
		groupNodeIDs := make(map[string]bool)
		groupIDSet := make(map[string]bool)
		for _, gid := range b.filter.GroupIDs {
			groupIDSet[gid] = true
		}
		for _, group := range graph.Groups {
			if groupIDSet[group.ID] {
				for _, nid := range group.NodeIDs {
					groupNodeIDs[nid] = true
				}
			}
		}
		for id := range graph.Nodes {
			if !groupNodeIDs[id] {
				toRemove[id] = true
			}
		}
		for id := range toRemove {
			delete(graph.Nodes, id)
		}
		graph.Edges = b.filterEdgesByRule(graph.Nodes, graph.Edges)
	}

	// 4. フォーカスフィルタ
	if b.filter.FocusID != "" && b.filter.FocusDepth > 0 {
		focusID := b.filter.FocusID
		// FocusID が Objective ID の場合: そのグループ内ノードを対象にフォールバック
		if _, exists := graph.Nodes[focusID]; !exists {
			for _, group := range graph.Groups {
				if group.ID == focusID && len(group.NodeIDs) > 0 {
					// グループ内の全ノードを reachable にする
					for id := range graph.Nodes {
						if !slices.Contains(group.NodeIDs, id) {
							delete(graph.Nodes, id)
						}
					}
					graph.Edges = b.filterEdgesByRule(graph.Nodes, graph.Edges)
					focusID = "" // フォーカス探索をスキップ
					break
				}
			}
		}
		if focusID != "" {
			reachable := b.findReachableNodes(graph, focusID, b.filter.FocusDepth)
			for id := range graph.Nodes {
				if !reachable[id] {
					delete(graph.Nodes, id)
				}
			}
			graph.Edges = b.filterEdgesByRule(graph.Nodes, graph.Edges)
		}
	}

	// 5. 構造親子を再構築
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
	connectedNodes := make(map[string]bool)
	for _, edge := range graph.Edges {
		connectedNodes[edge.From] = true
		connectedNodes[edge.To] = true
	}
	for id := range graph.Nodes {
		if !connectedNodes[id] {
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
	stats.GroupCount = len(graph.Groups)
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
	fmt.Fprintf(&sb, "Total Groups: %d\n", g.Stats.GroupCount)
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

	// グループ（Objective）情報
	if len(g.Groups) > 0 {
		sb.WriteString("--- Groups (Objectives) ---\n")
		for _, group := range g.Groups {
			fmt.Fprintf(&sb, "  [%s] %s (%s) - %d nodes\n", group.ID, group.Title, group.Status, len(group.NodeIDs))
		}
		sb.WriteString("\n")
	}

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

	// グループに所属するノード ID を収集
	groupedNodeIDs := make(map[string]bool)
	for _, group := range g.Groups {
		for _, nid := range group.NodeIDs {
			groupedNodeIDs[nid] = true
		}
	}

	// Objective をサブグラフとして出力
	for _, group := range g.Groups {
		clusterID := strings.ReplaceAll(group.ID, "-", "_")
		fmt.Fprintf(&sb, "  subgraph cluster_%s {\n", clusterID)
		fmt.Fprintf(&sb, "    label=\"%s: %s\";\n", group.ID, group.Title)
		sb.WriteString("    style=dashed;\n")
		sb.WriteString("    color=\"#9C27B0\";\n")
		sb.WriteString("    fontcolor=\"#9C27B0\";\n\n")

		for _, nid := range group.NodeIDs {
			node, exists := g.Nodes[nid]
			if !exists {
				continue
			}
			shape, fill := dotNodeStyle(node)
			label := fmt.Sprintf("%s\\n%s", node.ID, node.Title)
			fmt.Fprintf(&sb, "    \"%s\" [label=\"%s\", shape=\"%s\", style=\"filled\", fillcolor=\"%s\"];\n", nid, label, shape, fill)
		}
		sb.WriteString("  }\n\n")
	}

	// グループ外のノード
	sb.WriteString("  // Ungrouped nodes\n")
	ids := make([]string, 0, len(g.Nodes))
	for id := range g.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if groupedNodeIDs[id] {
			continue
		}
		node := g.Nodes[id]
		shape, fill := dotNodeStyle(node)
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

// dotNodeStyle はノードタイプに応じた DOT スタイルを返す
func dotNodeStyle(node *UnifiedGraphNode) (shape string, fill string) {
	switch node.Type {
	case EntityTypeActivity:
		return "box", "#4CAF50"
	case EntityTypeUseCase:
		return "ellipse", "#2196F3"
	default:
		return "box", "#888888"
	}
}

func mermaidArrow(edge UnifiedEdge) string {
	switch edge.Relation {
	case RelationImplements:
		return "==>|implements|"
	default:
		return "-->"
	}
}

// ToMermaid は Mermaid 形式で出力
func (g *UnifiedGraph) ToMermaid() string {
	var sb strings.Builder

	sb.WriteString("```mermaid\ngraph TD\n")

	// グループに所属するノード ID を収集
	groupedNodeIDs := make(map[string]bool)
	for _, group := range g.Groups {
		for _, nid := range group.NodeIDs {
			groupedNodeIDs[nid] = true
		}
	}

	// Objective をサブグラフとして出力
	for _, group := range g.Groups {
		mermaidID := strings.ReplaceAll(group.ID, "-", "_")
		title := escapeMermaidText(group.Title)
		fmt.Fprintf(&sb, "  subgraph %s[\"%s: %s\"]\n", mermaidID, group.ID, title)
		for _, nid := range group.NodeIDs {
			node, exists := g.Nodes[nid]
			if !exists {
				continue
			}
			mermaidNodeDef(&sb, nid, node)
		}
		sb.WriteString("  end\n\n")
	}

	// グループ外のノードを出力
	ids := make([]string, 0, len(g.Nodes))
	for id := range g.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		if groupedNodeIDs[id] {
			continue
		}
		node := g.Nodes[id]
		mermaidNodeDef(&sb, id, node)
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

	for _, id := range ids {
		node := g.Nodes[id]
		class := "activity"
		if node.Type == EntityTypeUseCase {
			class = "usecase"
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

// mermaidNodeDef は Mermaid ノード定義を出力
func mermaidNodeDef(sb *strings.Builder, id string, node *UnifiedGraphNode) {
	title := escapeMermaidText(node.Title)
	switch node.Type {
	case EntityTypeActivity:
		fmt.Fprintf(sb, "  %s([%s: %s])\n", id, id, title)
	case EntityTypeUseCase:
		fmt.Fprintf(sb, "  %s((%s: %s))\n", id, id, title)
	}
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
