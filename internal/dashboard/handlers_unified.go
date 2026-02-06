package dashboard

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/biwakonbu/zeus/internal/analysis"
)

// =============================================================================
// UnifiedGraph API 型定義
// =============================================================================

// UnifiedGraphNodeItem は UnifiedGraph ノードの API アイテム
type UnifiedGraphNodeItem struct {
	ID                 string   `json:"id"`
	Type               string   `json:"type"`
	Title              string   `json:"title"`
	Status             string   `json:"status"`
	StructuralDepth    int      `json:"structural_depth"`
	Mode               string   `json:"mode,omitempty"`
	Assignee           string   `json:"assignee,omitempty"`
	Priority           string   `json:"priority,omitempty"`
	StructuralParents  []string `json:"structural_parents,omitempty"`
	StructuralChildren []string `json:"structural_children,omitempty"`
}

// UnifiedGraphEdgeItem は UnifiedGraph エッジの API アイテム
type UnifiedGraphEdgeItem struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Layer    string `json:"layer"`
	Relation string `json:"relation"`
}

// UnifiedGraphStatsItem は UnifiedGraph 統計の API アイテム
type UnifiedGraphStatsItem struct {
	TotalNodes          int            `json:"total_nodes"`
	TotalEdges          int            `json:"total_edges"`
	TotalActivities     int            `json:"total_activities"`
	CompletedActivities int            `json:"completed_activities"`
	MaxStructuralDepth  int            `json:"max_structural_depth"`
	CycleCount          int            `json:"cycle_count"`
	IsolatedCount       int            `json:"isolated_count"`
	NodesByType         map[string]int `json:"nodes_by_type,omitempty"`
	EdgesByLayer        map[string]int `json:"edges_by_layer,omitempty"`
	EdgesByRelation     map[string]int `json:"edges_by_relation,omitempty"`
}

// UnifiedGraphResponse は UnifiedGraph API のレスポンス
type UnifiedGraphResponse struct {
	Nodes    []UnifiedGraphNodeItem `json:"nodes"`
	Edges    []UnifiedGraphEdgeItem `json:"edges"`
	Stats    UnifiedGraphStatsItem  `json:"stats"`
	Cycles   [][]string             `json:"cycles"`
	Isolated []string               `json:"isolated"`
	Mermaid  string                 `json:"mermaid"`

	// フィルター情報
	Filter *UnifiedGraphFilterInfo `json:"filter,omitempty"`
}

// UnifiedGraphFilterInfo はフィルター情報
type UnifiedGraphFilterInfo struct {
	FocusID          string   `json:"focus_id,omitempty"`
	Depth            int      `json:"depth,omitempty"`
	IncludeTypes     []string `json:"include_types,omitempty"`
	IncludeLayers    []string `json:"include_layers,omitempty"`
	IncludeRelations []string `json:"include_relations,omitempty"`
	HideCompleted    bool     `json:"hide_completed"`
	HideDraft        bool     `json:"hide_draft"`
}

// =============================================================================
// UnifiedGraph API ハンドラー
// =============================================================================

// handleAPIUnifiedGraph は UnifiedGraph API を処理
func (s *Server) handleAPIUnifiedGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()

	// クエリパラメータからフィルターを構築
	filter := buildGraphFilter(r)

	// UnifiedGraph を構築
	graph, err := s.zeus.BuildUnifiedGraph(ctx, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// レスポンスを構築
	response := convertUnifiedGraphToResponse(graph, filter)

	writeJSON(w, http.StatusOK, response)
}

// buildGraphFilter はクエリパラメータから GraphFilter を構築
func buildGraphFilter(r *http.Request) *analysis.GraphFilter {
	filter := analysis.NewGraphFilter()

	// focus パラメータ
	focusID := r.URL.Query().Get("focus")
	if focusID != "" {
		// depth パラメータ
		depth := 3 // デフォルト
		if depthStr := r.URL.Query().Get("depth"); depthStr != "" {
			if d, err := strconv.Atoi(depthStr); err == nil && d > 0 {
				depth = d
			}
		}
		filter = filter.WithFocus(focusID, depth)
	}

	// types パラメータ
	typesStr := r.URL.Query().Get("types")
	if typesStr != "" {
		types := parseEntityTypesFromQuery(typesStr)
		if len(types) > 0 {
			filter = filter.WithIncludeTypes(types...)
		}
	}
	// layers パラメータ
	layersStr := r.URL.Query().Get("layers")
	if layersStr != "" {
		layers := parseEdgeLayersFromQuery(layersStr)
		if len(layers) > 0 {
			filter = filter.WithIncludeLayers(layers...)
		}
	}
	// relations パラメータ
	relationsStr := r.URL.Query().Get("relations")
	if relationsStr != "" {
		relations := parseEdgeRelationsFromQuery(relationsStr)
		if len(relations) > 0 {
			filter = filter.WithIncludeRelations(relations...)
		}
	}

	// hide-completed パラメータ
	if r.URL.Query().Get("hide-completed") == "true" {
		filter = filter.WithHideCompleted(true)
	}

	// hide-draft パラメータ
	if r.URL.Query().Get("hide-draft") == "true" {
		filter = filter.WithHideDraft(true)
	}

	return filter
}

// parseEntityTypesFromQuery はクエリパラメータからエンティティタイプを解析
func parseEntityTypesFromQuery(typesStr string) []analysis.EntityType {
	var types []analysis.EntityType
	for _, t := range strings.Split(typesStr, ",") {
		t = strings.TrimSpace(strings.ToLower(t))
		switch t {
		case "activity":
			types = append(types, analysis.EntityTypeActivity)
		case "usecase":
			types = append(types, analysis.EntityTypeUseCase)
		case "deliverable":
			types = append(types, analysis.EntityTypeDeliverable)
		case "objective":
			types = append(types, analysis.EntityTypeObjective)
		}
	}
	return types
}

func parseEdgeLayersFromQuery(layersStr string) []analysis.UnifiedEdgeLayer {
	var layers []analysis.UnifiedEdgeLayer
	for _, l := range strings.Split(layersStr, ",") {
		l = strings.TrimSpace(strings.ToLower(l))
		switch l {
		case string(analysis.EdgeLayerStructural):
			layers = append(layers, analysis.EdgeLayerStructural)
		case string(analysis.EdgeLayerReference):
			layers = append(layers, analysis.EdgeLayerReference)
		}
	}
	return layers
}

func parseEdgeRelationsFromQuery(relationsStr string) []analysis.UnifiedEdgeRelation {
	var relations []analysis.UnifiedEdgeRelation
	for _, r := range strings.Split(relationsStr, ",") {
		r = strings.TrimSpace(strings.ToLower(r))
		switch r {
		case string(analysis.RelationParent):
			relations = append(relations, analysis.RelationParent)
		case string(analysis.RelationDependsOn):
			relations = append(relations, analysis.RelationDependsOn)
		case string(analysis.RelationImplements):
			relations = append(relations, analysis.RelationImplements)
		case string(analysis.RelationContributes):
			relations = append(relations, analysis.RelationContributes)
		case string(analysis.RelationFulfills):
			relations = append(relations, analysis.RelationFulfills)
		case string(analysis.RelationProduces):
			relations = append(relations, analysis.RelationProduces)
		}
	}
	return relations
}

// convertUnifiedGraphToResponse は UnifiedGraph をレスポンスに変換
func convertUnifiedGraphToResponse(graph *analysis.UnifiedGraph, filter *analysis.GraphFilter) UnifiedGraphResponse {
	// ノードの変換（map → slice）
	nodes := make([]UnifiedGraphNodeItem, 0, len(graph.Nodes))
	for _, node := range graph.Nodes {
		item := UnifiedGraphNodeItem{
			ID:                 node.ID,
			Type:               string(node.Type),
			Title:              node.Title,
			Status:             node.Status,
			StructuralDepth:    node.StructuralDepth,
			Mode:               node.Mode,
			Assignee:           node.Assignee,
			Priority:           node.Priority,
			StructuralParents:  node.StructuralParents,
			StructuralChildren: node.StructuralChildren,
		}
		nodes = append(nodes, item)
	}

	// エッジの変換
	edges := make([]UnifiedGraphEdgeItem, len(graph.Edges))
	for i, edge := range graph.Edges {
		edges[i] = UnifiedGraphEdgeItem{
			Source:   edge.From,
			Target:   edge.To,
			Layer:    string(edge.Layer),
			Relation: string(edge.Relation),
		}
	}

	// NodesByType の変換（EntityType → string）
	nodesByType := make(map[string]int)
	for k, v := range graph.Stats.NodesByType {
		nodesByType[string(k)] = v
	}

	// EdgesByLayer の変換
	edgesByLayer := make(map[string]int)
	for k, v := range graph.Stats.EdgesByLayer {
		edgesByLayer[string(k)] = v
	}

	// EdgesByRelation の変換
	edgesByRelation := make(map[string]int)
	for k, v := range graph.Stats.EdgesByRelation {
		edgesByRelation[string(k)] = v
	}

	// 統計の変換
	stats := UnifiedGraphStatsItem{
		TotalNodes:          graph.Stats.TotalNodes,
		TotalEdges:          graph.Stats.TotalEdges,
		TotalActivities:     graph.Stats.TotalActivities,
		CompletedActivities: graph.Stats.CompletedActivities,
		MaxStructuralDepth:  graph.Stats.MaxStructuralDepth,
		CycleCount:          graph.Stats.CycleCount,
		IsolatedCount:       graph.Stats.IsolatedCount,
		NodesByType:         nodesByType,
		EdgesByLayer:        edgesByLayer,
		EdgesByRelation:     edgesByRelation,
	}

	// Cycles, Isolated の nil チェック
	cycles := graph.Cycles
	if cycles == nil {
		cycles = [][]string{}
	}
	isolated := graph.Isolated
	if isolated == nil {
		isolated = []string{}
	}

	// フィルター情報の構築
	var filterInfo *UnifiedGraphFilterInfo
	if filter != nil && (filter.FocusID != "" ||
		len(filter.IncludeTypes) > 0 ||
		len(filter.IncludeLayers) > 0 ||
		len(filter.IncludeRelations) > 0 ||
		filter.HideCompleted || filter.HideDraft) {
		types := make([]string, len(filter.IncludeTypes))
		for i, t := range filter.IncludeTypes {
			types[i] = string(t)
		}
		layers := make([]string, len(filter.IncludeLayers))
		for i, l := range filter.IncludeLayers {
			layers[i] = string(l)
		}
		relations := make([]string, len(filter.IncludeRelations))
		for i, rel := range filter.IncludeRelations {
			relations[i] = string(rel)
		}
		filterInfo = &UnifiedGraphFilterInfo{
			FocusID:          filter.FocusID,
			Depth:            filter.FocusDepth,
			IncludeTypes:     types,
			IncludeLayers:    layers,
			IncludeRelations: relations,
			HideCompleted:    filter.HideCompleted,
			HideDraft:        filter.HideDraft,
		}
	}

	return UnifiedGraphResponse{
		Nodes:    nodes,
		Edges:    edges,
		Stats:    stats,
		Cycles:   cycles,
		Isolated: isolated,
		Mermaid:  graph.ToMermaid(),
		Filter:   filterInfo,
	}
}
