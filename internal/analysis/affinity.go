// Package analysis は Zeus の高度な分析機能を提供する。
package analysis

import (
	"context"
	"sort"
)

// AffinityType は関連の種類
type AffinityType string

const (
	AffinityParentChild AffinityType = "parent-child"
	AffinitySibling     AffinityType = "sibling"
	AffinityReference   AffinityType = "reference"
	AffinityCategory    AffinityType = "category"
)

// AffinityNode はアフィニティキャンバス用のノード
type AffinityNode struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

// AffinityEdge はノード間の関連
type AffinityEdge struct {
	Source string         `json:"source"`
	Target string         `json:"target"`
	Score  float64        `json:"score"`
	Types  []AffinityType `json:"types"`
	Reason string         `json:"reason"`
}

// AffinityWeights は関連タイプの重み
type AffinityWeights struct {
	ParentChild float64 `json:"parent_child"`
	Sibling     float64 `json:"sibling"`
	Reference   float64 `json:"reference"`
	Category    float64 `json:"category"`
}

// AffinityCluster はノードのクラスタ
type AffinityCluster struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

// AffinityStats は統計情報
type AffinityStats struct {
	TotalNodes     int     `json:"total_nodes"`
	TotalEdges     int     `json:"total_edges"`
	ClusterCount   int     `json:"cluster_count"`
	AvgConnections float64 `json:"avg_connections"`
	FilteredEdges  int     `json:"filtered_edges,omitempty"` // フィルタリングで除外されたエッジ数
	UsedHubMode    bool    `json:"used_hub_mode,omitempty"`  // ハブモードを使用したか
}

// AffinityOptions は計算オプション
type AffinityOptions struct {
	// MaxSiblings はハブモードに切り替える兄弟数の閾値（デフォルト: 20）
	MaxSiblings int `json:"max_siblings"`
	// MinScore はエッジを含める最小スコア閾値（デフォルト: 0.0）
	MinScore float64 `json:"min_score"`
	// MaxEdges は最大エッジ数（デフォルト: 0 = 無制限）
	MaxEdges int `json:"max_edges"`
}

// DefaultAffinityOptions はデフォルトのオプションを返す
func DefaultAffinityOptions() AffinityOptions {
	return AffinityOptions{
		MaxSiblings: 20,
		MinScore:    0.0,
		MaxEdges:    0,
	}
}

// AffinityResult は計算結果
type AffinityResult struct {
	Nodes    []AffinityNode    `json:"nodes"`
	Edges    []AffinityEdge    `json:"edges"`
	Clusters []AffinityCluster `json:"clusters"`
	Weights  AffinityWeights   `json:"weights"`
	Stats    AffinityStats     `json:"stats"`
}

// AffinityCalculator は類似度を計算
type AffinityCalculator struct {
	vision       VisionInfo
	objectives   []ObjectiveInfo
	deliverables []DeliverableInfo
	tasks        []TaskInfo
	quality      []QualityInfo
	risks        []RiskInfo
	options      AffinityOptions
	usedHubMode  bool // ハブモードを使用したかどうか
}

// NewAffinityCalculator はコンストラクタ
func NewAffinityCalculator(
	vision VisionInfo,
	objectives []ObjectiveInfo,
	deliverables []DeliverableInfo,
	tasks []TaskInfo,
	quality []QualityInfo,
	risks []RiskInfo,
) *AffinityCalculator {
	return &AffinityCalculator{
		vision:       vision,
		objectives:   objectives,
		deliverables: deliverables,
		tasks:        tasks,
		quality:      quality,
		risks:        risks,
		options:      DefaultAffinityOptions(),
	}
}

// NewAffinityCalculatorWithOptions はオプション付きコンストラクタ
func NewAffinityCalculatorWithOptions(
	vision VisionInfo,
	objectives []ObjectiveInfo,
	deliverables []DeliverableInfo,
	tasks []TaskInfo,
	quality []QualityInfo,
	risks []RiskInfo,
	options AffinityOptions,
) *AffinityCalculator {
	// デフォルト値を設定
	if options.MaxSiblings <= 0 {
		options.MaxSiblings = 20
	}
	return &AffinityCalculator{
		vision:       vision,
		objectives:   objectives,
		deliverables: deliverables,
		tasks:        tasks,
		quality:      quality,
		risks:        risks,
		options:      options,
	}
}

// Calculate はアフィニティを計算
func (ac *AffinityCalculator) Calculate(ctx context.Context) (*AffinityResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ノードを構築
	nodes := ac.buildNodes()

	// エッジを検出
	edges := ac.detectAllEdges()

	// 重みを計算
	weights := ac.CalculateWeights()

	// スコアを計算
	edges = ac.calculateScores(edges, weights)

	// エッジをフィルタリング
	originalEdgeCount := len(edges)
	edges = ac.filterEdges(edges)
	filteredCount := originalEdgeCount - len(edges)

	// クラスタを構築
	clusters := ac.buildClusters(edges)

	// 統計を計算
	stats := ac.calculateStats(nodes, edges, clusters)
	stats.FilteredEdges = filteredCount
	stats.UsedHubMode = ac.usedHubMode

	return &AffinityResult{
		Nodes:    nodes,
		Edges:    edges,
		Clusters: clusters,
		Weights:  weights,
		Stats:    stats,
	}, nil
}

// filterEdges はスコア閾値と最大数でエッジをフィルタリング
func (ac *AffinityCalculator) filterEdges(edges []AffinityEdge) []AffinityEdge {
	// スコア閾値でフィルタリング
	if ac.options.MinScore > 0 {
		filtered := make([]AffinityEdge, 0, len(edges))
		for _, e := range edges {
			if e.Score >= ac.options.MinScore {
				filtered = append(filtered, e)
			}
		}
		edges = filtered
	}

	// 最大エッジ数でフィルタリング（スコア上位を優先）
	if ac.options.MaxEdges > 0 && len(edges) > ac.options.MaxEdges {
		// スコアでソート（降順）
		sort.Slice(edges, func(i, j int) bool {
			return edges[i].Score > edges[j].Score
		})
		edges = edges[:ac.options.MaxEdges]
	}

	return edges
}

// buildNodes は全エンティティからノードを構築
func (ac *AffinityCalculator) buildNodes() []AffinityNode {
	nodes := []AffinityNode{}

	// Vision
	if ac.vision.Title != "" {
		nodes = append(nodes, AffinityNode{
			ID:     "vision",
			Title:  ac.vision.Title,
			Type:   "vision",
			Status: ac.vision.Status,
		})
	}

	// Objectives
	for _, obj := range ac.objectives {
		nodes = append(nodes, AffinityNode{
			ID:     obj.ID,
			Title:  obj.Title,
			Type:   "objective",
			Status: obj.Status,
		})
	}

	// Deliverables
	for _, del := range ac.deliverables {
		nodes = append(nodes, AffinityNode{
			ID:     del.ID,
			Title:  del.Title,
			Type:   "deliverable",
			Status: del.Status,
		})
	}

	// Tasks
	for _, task := range ac.tasks {
		nodes = append(nodes, AffinityNode{
			ID:     task.ID,
			Title:  task.Title,
			Type:   "task",
			Status: task.Status,
		})
	}

	return nodes
}

// detectAllEdges は全関連タイプのエッジを検出
func (ac *AffinityCalculator) detectAllEdges() []AffinityEdge {
	edgeMap := make(map[string]*AffinityEdge) // source-target -> edge

	// 親子関係
	for _, e := range ac.detectParentChild() {
		key := e.Source + "-" + e.Target
		if existing, ok := edgeMap[key]; ok {
			existing.Types = append(existing.Types, e.Types...)
		} else {
			edgeCopy := e
			edgeMap[key] = &edgeCopy
		}
	}

	// 兄弟関係
	for _, e := range ac.detectSibling() {
		key := e.Source + "-" + e.Target
		if existing, ok := edgeMap[key]; ok {
			existing.Types = append(existing.Types, e.Types...)
		} else {
			edgeCopy := e
			edgeMap[key] = &edgeCopy
		}
	}

	// 参照関係
	for _, e := range ac.detectReference() {
		key := e.Source + "-" + e.Target
		if existing, ok := edgeMap[key]; ok {
			existing.Types = append(existing.Types, e.Types...)
		} else {
			edgeCopy := e
			edgeMap[key] = &edgeCopy
		}
	}

	// マップからスライスに変換
	edges := make([]AffinityEdge, 0, len(edgeMap))
	for _, e := range edgeMap {
		edges = append(edges, *e)
	}

	return edges
}

// detectParentChild は親子関係を検出
func (ac *AffinityCalculator) detectParentChild() []AffinityEdge {
	edges := []AffinityEdge{}

	// Vision -> Objective
	if ac.vision.Title != "" {
		for _, obj := range ac.objectives {
			if obj.ParentID == "" { // トップレベル Objective
				edges = append(edges, AffinityEdge{
					Source: "vision",
					Target: obj.ID,
					Types:  []AffinityType{AffinityParentChild},
					Reason: "Vision 直下の Objective",
				})
			}
		}
	}

	// Objective -> Objective (階層)
	for _, obj := range ac.objectives {
		if obj.ParentID != "" {
			edges = append(edges, AffinityEdge{
				Source: obj.ParentID,
				Target: obj.ID,
				Types:  []AffinityType{AffinityParentChild},
				Reason: "親 Objective",
			})
		}
	}

	// Objective -> Deliverable
	for _, del := range ac.deliverables {
		if del.ObjectiveID != "" {
			edges = append(edges, AffinityEdge{
				Source: del.ObjectiveID,
				Target: del.ID,
				Types:  []AffinityType{AffinityParentChild},
				Reason: "Objective の成果物",
			})
		}
	}

	// Deliverable -> Task (ParentID)
	for _, task := range ac.tasks {
		if task.ParentID != "" {
			edges = append(edges, AffinityEdge{
				Source: task.ParentID,
				Target: task.ID,
				Types:  []AffinityType{AffinityParentChild},
				Reason: "親タスク",
			})
		}
	}

	return edges
}

// detectSibling は兄弟関係を検出
// 兄弟数が閾値（MaxSiblings）を超える場合はハブモードに切り替え
func (ac *AffinityCalculator) detectSibling() []AffinityEdge {
	edges := []AffinityEdge{}

	// 同じ親を持つ Deliverable
	objDeliverables := make(map[string][]string)
	for _, del := range ac.deliverables {
		if del.ObjectiveID != "" {
			objDeliverables[del.ObjectiveID] = append(objDeliverables[del.ObjectiveID], del.ID)
		}
	}
	for objID, delIDs := range objDeliverables {
		edges = append(edges, ac.createSiblingEdges(delIDs, objID)...)
	}

	// 同じ親を持つ Task
	parentTasks := make(map[string][]string)
	for _, task := range ac.tasks {
		if task.ParentID != "" {
			parentTasks[task.ParentID] = append(parentTasks[task.ParentID], task.ID)
		}
	}
	for parentID, taskIDs := range parentTasks {
		edges = append(edges, ac.createSiblingEdges(taskIDs, parentID)...)
	}

	return edges
}

// createSiblingEdges は兄弟エッジを作成（ハブモード対応）
// ids が閾値を超える場合は全ペアではなくハブノードを介した接続に切り替え
func (ac *AffinityCalculator) createSiblingEdges(ids []string, parentID string) []AffinityEdge {
	if len(ids) < 2 {
		return nil
	}

	edges := []AffinityEdge{}

	// 閾値を超える場合はハブモード
	if len(ids) > ac.options.MaxSiblings {
		ac.usedHubMode = true
		// 最初の要素をハブとして使用
		hub := ids[0]
		for _, id := range ids[1:] {
			edges = append(edges, AffinityEdge{
				Source: hub,
				Target: id,
				Types:  []AffinityType{AffinitySibling},
				Reason: "同じ " + parentID + " に属する（ハブ経由）",
			})
		}
		return edges
	}

	// 通常モード: 全ペア生成
	for i := range len(ids) {
		for j := i + 1; j < len(ids); j++ {
			edges = append(edges, AffinityEdge{
				Source: ids[i],
				Target: ids[j],
				Types:  []AffinityType{AffinitySibling},
				Reason: "同じ " + parentID + " に属する",
			})
		}
	}
	return edges
}

// detectReference は参照関係を検出
func (ac *AffinityCalculator) detectReference() []AffinityEdge {
	edges := []AffinityEdge{}

	// Quality -> Deliverable
	for _, q := range ac.quality {
		if q.DeliverableID != "" {
			edges = append(edges, AffinityEdge{
				Source: q.ID,
				Target: q.DeliverableID,
				Types:  []AffinityType{AffinityReference},
				Reason: "品質基準の対象",
			})
		}
	}

	// Risk -> Objective / Deliverable
	for _, r := range ac.risks {
		if r.ObjectiveID != "" {
			edges = append(edges, AffinityEdge{
				Source: r.ID,
				Target: r.ObjectiveID,
				Types:  []AffinityType{AffinityReference},
				Reason: "リスクの対象",
			})
		}
		if r.DeliverableID != "" {
			edges = append(edges, AffinityEdge{
				Source: r.ID,
				Target: r.DeliverableID,
				Types:  []AffinityType{AffinityReference},
				Reason: "リスクの対象",
			})
		}
	}

	return edges
}

// CalculateWeights はプロジェクト特性から重みを計算
func (ac *AffinityCalculator) CalculateWeights() AffinityWeights {
	totalEntities := len(ac.objectives) + len(ac.deliverables) + len(ac.tasks)
	if totalEntities == 0 {
		return AffinityWeights{
			ParentChild: 1.0,
			Sibling:     0.7,
			Reference:   0.5,
			Category:    0.3,
		}
	}

	// 参照関係の比率
	refCount := len(ac.quality) + len(ac.risks)
	refRatio := float64(refCount) / float64(totalEntities)

	// 平均兄弟数
	objDeliverables := make(map[string]int)
	for _, del := range ac.deliverables {
		objDeliverables[del.ObjectiveID]++
	}
	totalSiblings := 0
	for _, count := range objDeliverables {
		totalSiblings += count
	}
	avgSiblings := 0.0
	if len(objDeliverables) > 0 {
		avgSiblings = float64(totalSiblings) / float64(len(objDeliverables))
	}

	// 重みを計算
	siblingWeight := 0.7 - (avgSiblings * 0.05)
	if siblingWeight < 0.5 {
		siblingWeight = 0.5
	}
	if siblingWeight > 0.8 {
		siblingWeight = 0.8
	}

	refWeight := 0.4 + (refRatio * 0.3)
	if refWeight > 0.7 {
		refWeight = 0.7
	}

	return AffinityWeights{
		ParentChild: 1.0,
		Sibling:     siblingWeight,
		Reference:   refWeight,
		Category:    0.3,
	}
}

// calculateScores はエッジのスコアを計算
func (ac *AffinityCalculator) calculateScores(edges []AffinityEdge, weights AffinityWeights) []AffinityEdge {
	for i := range edges {
		score := 0.0
		for _, t := range edges[i].Types {
			switch t {
			case AffinityParentChild:
				score += weights.ParentChild
			case AffinitySibling:
				score += weights.Sibling
			case AffinityReference:
				score += weights.Reference
			case AffinityCategory:
				score += weights.Category
			}
		}
		// 正規化（最大 1.0）
		if score > 1.0 {
			score = 1.0
		}
		edges[i].Score = score
	}
	return edges
}

// buildClusters はノードをクラスタリング
// O(O + D) の事前インデックスを使用して高速化
func (ac *AffinityCalculator) buildClusters(_ []AffinityEdge) []AffinityCluster {
	clusters := []AffinityCluster{}

	// 事前インデックス: Objective ID → Deliverable IDs
	// O(D) で構築
	objToDeliverables := make(map[string][]string, len(ac.objectives))
	for _, del := range ac.deliverables {
		if del.ObjectiveID != "" {
			objToDeliverables[del.ObjectiveID] = append(objToDeliverables[del.ObjectiveID], del.ID)
		}
	}

	// Objective ごとにクラスタを作成
	// O(O) でクラスタ生成（O(1) ルックアップ）
	for _, obj := range ac.objectives {
		members := []string{obj.ID}

		// インデックスから O(1) で取得
		if delIDs, ok := objToDeliverables[obj.ID]; ok {
			members = append(members, delIDs...)
		}

		if len(members) > 1 {
			clusters = append(clusters, AffinityCluster{
				ID:      "cluster-" + obj.ID,
				Name:    obj.Title,
				Members: members,
			})
		}
	}

	return clusters
}

// calculateStats は統計情報を計算
func (ac *AffinityCalculator) calculateStats(nodes []AffinityNode, edges []AffinityEdge, clusters []AffinityCluster) AffinityStats {
	// 平均接続数
	connectionCount := make(map[string]int)
	for _, e := range edges {
		connectionCount[e.Source]++
		connectionCount[e.Target]++
	}
	totalConnections := 0
	for _, count := range connectionCount {
		totalConnections += count
	}
	avgConnections := 0.0
	if len(nodes) > 0 {
		avgConnections = float64(totalConnections) / float64(len(nodes))
	}

	return AffinityStats{
		TotalNodes:     len(nodes),
		TotalEdges:     len(edges),
		ClusterCount:   len(clusters),
		AvgConnections: avgConnections,
	}
}
