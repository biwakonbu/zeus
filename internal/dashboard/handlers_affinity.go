package dashboard

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/biwakonbu/zeus/internal/analysis"
	"github.com/biwakonbu/zeus/internal/core"
)

// =============================================================================
// Affinity API 型定義（Phase 7: Affinity Canvas）
// =============================================================================

// AffinityResponse は Affinity API のレスポンス
type AffinityResponse struct {
	Nodes    []AffinityNodeResponse    `json:"nodes"`
	Edges    []AffinityEdgeResponse    `json:"edges"`
	Clusters []AffinityClusterResponse `json:"clusters"`
	Weights  AffinityWeightsResponse   `json:"weights"`
	Stats    AffinityStatsResponse     `json:"stats"`
}

// AffinityNodeResponse はノードレスポンス
type AffinityNodeResponse struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	WBSCode  string `json:"wbs_code"`
	Progress int    `json:"progress"`
	Status   string `json:"status"`
}

// AffinityEdgeResponse はエッジレスポンス
type AffinityEdgeResponse struct {
	Source string   `json:"source"`
	Target string   `json:"target"`
	Score  float64  `json:"score"`
	Types  []string `json:"types"`
	Reason string   `json:"reason"`
}

// AffinityClusterResponse はクラスタレスポンス
type AffinityClusterResponse struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

// AffinityWeightsResponse は重みレスポンス
type AffinityWeightsResponse struct {
	ParentChild float64 `json:"parent_child"`
	Sibling     float64 `json:"sibling"`
	WBSAdjacent float64 `json:"wbs_adjacent"`
	Reference   float64 `json:"reference"`
	Category    float64 `json:"category"`
}

// AffinityStatsResponse は統計レスポンス
type AffinityStatsResponse struct {
	TotalNodes     int     `json:"total_nodes"`
	TotalEdges     int     `json:"total_edges"`
	ClusterCount   int     `json:"cluster_count"`
	AvgConnections float64 `json:"avg_connections"`
}

// =============================================================================
// Affinity API ハンドラー
// =============================================================================

// handleAPIAffinity は Affinity API を処理
// クエリパラメータ:
//   - max_siblings: ハブモードに切り替える兄弟数の閾値（デフォルト: 20）
//   - min_score: 最小スコア閾値（デフォルト: 0.0）
//   - max_edges: 最大エッジ数（デフォルト: 0 = 無制限）
func (s *Server) handleAPIAffinity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()

	// クエリパラメータを解析
	options := analysis.DefaultAffinityOptions()
	if v := r.URL.Query().Get("max_siblings"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			options.MaxSiblings = n
		}
	}
	if v := r.URL.Query().Get("min_score"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 {
			options.MinScore = f
		}
	}
	if v := r.URL.Query().Get("max_edges"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			options.MaxEdges = n
		}
	}

	// エンティティを並列に読み込み
	visionInfo, objectives, deliverables, tasks, quality, risks := s.loadAffinityDataParallel(ctx)

	// AffinityCalculator でアフィニティを計算
	calculator := analysis.NewAffinityCalculatorWithOptions(
		visionInfo,
		objectives,
		deliverables,
		tasks,
		quality,
		risks,
		options,
	)

	result, err := calculator.Calculate(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "アフィニティ計算エラー: "+err.Error())
		return
	}

	// レスポンス変換
	nodes := make([]AffinityNodeResponse, len(result.Nodes))
	for i, n := range result.Nodes {
		nodes[i] = AffinityNodeResponse{
			ID:       n.ID,
			Title:    n.Title,
			Type:     n.Type,
			WBSCode:  n.WBSCode,
			Progress: n.Progress,
			Status:   n.Status,
		}
	}

	edges := make([]AffinityEdgeResponse, len(result.Edges))
	for i, e := range result.Edges {
		types := make([]string, len(e.Types))
		for j, t := range e.Types {
			types[j] = string(t)
		}
		edges[i] = AffinityEdgeResponse{
			Source: e.Source,
			Target: e.Target,
			Score:  e.Score,
			Types:  types,
			Reason: e.Reason,
		}
	}

	clusters := make([]AffinityClusterResponse, len(result.Clusters))
	for i, c := range result.Clusters {
		clusters[i] = AffinityClusterResponse{
			ID:      c.ID,
			Name:    c.Name,
			Members: c.Members,
		}
	}

	response := AffinityResponse{
		Nodes:    nodes,
		Edges:    edges,
		Clusters: clusters,
		Weights: AffinityWeightsResponse{
			ParentChild: result.Weights.ParentChild,
			Sibling:     result.Weights.Sibling,
			WBSAdjacent: result.Weights.WBSAdjacent,
			Reference:   result.Weights.Reference,
			Category:    result.Weights.Category,
		},
		Stats: AffinityStatsResponse{
			TotalNodes:     result.Stats.TotalNodes,
			TotalEdges:     result.Stats.TotalEdges,
			ClusterCount:   result.Stats.ClusterCount,
			AvgConnections: result.Stats.AvgConnections,
		},
	}

	writeJSON(w, http.StatusOK, response)
}

// =============================================================================
// Affinity ヘルパー関数
// =============================================================================

// loadAffinityDataParallel はエンティティを並列に読み込む
func (s *Server) loadAffinityDataParallel(ctx context.Context) (
	visionInfo analysis.VisionInfo,
	objectives []analysis.ObjectiveInfo,
	deliverables []analysis.DeliverableInfo,
	tasks []analysis.TaskInfo,
	quality []analysis.QualityInfo,
	risks []analysis.RiskInfo,
) {
	fileStore := s.zeus.FileStore()
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 並列読み込みのワーカー数（セマフォ）
	sem := make(chan struct{}, 10)

	// Vision（単一ファイル、直接読み込み）
	wg.Add(1)
	go func() {
		defer wg.Done()
		var vision core.Vision
		_ = fileStore.ReadYaml(ctx, "vision.yaml", &vision)
		mu.Lock()
		visionInfo = analysis.VisionInfo{
			Title:  vision.Title,
			Status: "active",
		}
		mu.Unlock()
	}()

	// Activities（ディレクトリ内の複数ファイル）
	wg.Add(1)
	go func() {
		defer wg.Done()
		actFiles, err := fileStore.ListDir(ctx, "activities")
		if err != nil {
			return
		}
		result := make([]analysis.TaskInfo, 0, len(actFiles))
		for _, file := range actFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			var act core.ActivityEntity
			if err := fileStore.ReadYaml(ctx, "activities/"+file, &act); err != nil {
				continue
			}
			t := act.ToListItem()
			completedAt := ""
			if t.Status == core.ItemStatusCompleted {
				completedAt = t.UpdatedAt
			}
			result = append(result, analysis.TaskInfo{
				ID:            t.ID,
				Title:         t.Title,
				Status:        string(t.Status),
				Dependencies:  t.Dependencies,
				ParentID:      t.ParentID,
				StartDate:     t.StartDate,
				DueDate:       t.DueDate,
				Progress:      t.Progress,
				WBSCode:       t.WBSCode,
				Priority:      string(t.Priority),
				Assignee:      t.Assignee,
				EstimateHours: t.EstimateHours,
				CreatedAt:     t.CreatedAt,
				UpdatedAt:     t.UpdatedAt,
				CompletedAt:   completedAt,
			})
		}
		mu.Lock()
		tasks = result
		mu.Unlock()
	}()

	// Objectives（ディレクトリ内の複数ファイル）
	wg.Add(1)
	go func() {
		defer wg.Done()
		objFiles, err := fileStore.ListDir(ctx, "objectives")
		if err != nil {
			return
		}
		var objWg sync.WaitGroup
		result := make([]analysis.ObjectiveInfo, 0, len(objFiles))
		for _, file := range objFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			objWg.Add(1)
			file := file
			go func() {
				defer objWg.Done()
				sem <- struct{}{}        // セマフォ取得
				defer func() { <-sem }() // セマフォ解放

				var obj core.ObjectiveEntity
				if err := fileStore.ReadYaml(ctx, "objectives/"+file, &obj); err == nil {
					info := analysis.ObjectiveInfo{
						ID:        obj.ID,
						Title:     obj.Title,
						WBSCode:   obj.WBSCode,
						Progress:  obj.Progress,
						Status:    string(obj.Status),
						ParentID:  obj.ParentID,
						CreatedAt: obj.Metadata.CreatedAt,
						UpdatedAt: obj.Metadata.UpdatedAt,
					}
					mu.Lock()
					result = append(result, info)
					mu.Unlock()
				}
			}()
		}
		objWg.Wait()
		mu.Lock()
		objectives = result
		mu.Unlock()
	}()

	// Deliverables（ディレクトリ内の複数ファイル）
	wg.Add(1)
	go func() {
		defer wg.Done()
		delFiles, err := fileStore.ListDir(ctx, "deliverables")
		if err != nil {
			return
		}
		var delWg sync.WaitGroup
		result := make([]analysis.DeliverableInfo, 0, len(delFiles))
		for _, file := range delFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			delWg.Add(1)
			file := file
			go func() {
				defer delWg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				var del core.DeliverableEntity
				if err := fileStore.ReadYaml(ctx, "deliverables/"+file, &del); err == nil {
					info := analysis.DeliverableInfo{
						ID:          del.ID,
						Title:       del.Title,
						ObjectiveID: del.ObjectiveID,
						Progress:    del.Progress,
						Status:      string(del.Status),
						CreatedAt:   del.Metadata.CreatedAt,
						UpdatedAt:   del.Metadata.UpdatedAt,
					}
					mu.Lock()
					result = append(result, info)
					mu.Unlock()
				}
			}()
		}
		delWg.Wait()
		mu.Lock()
		deliverables = result
		mu.Unlock()
	}()

	// Quality（ディレクトリ内の複数ファイル）
	wg.Add(1)
	go func() {
		defer wg.Done()
		qualFiles, err := fileStore.ListDir(ctx, "quality")
		if err != nil {
			return
		}
		var qualWg sync.WaitGroup
		result := make([]analysis.QualityInfo, 0, len(qualFiles))
		for _, file := range qualFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			qualWg.Add(1)
			file := file
			go func() {
				defer qualWg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				var qual core.QualityEntity
				if err := fileStore.ReadYaml(ctx, "quality/"+file, &qual); err == nil {
					info := analysis.QualityInfo{
						ID:            qual.ID,
						Title:         qual.Title,
						DeliverableID: qual.DeliverableID,
					}
					mu.Lock()
					result = append(result, info)
					mu.Unlock()
				}
			}()
		}
		qualWg.Wait()
		mu.Lock()
		quality = result
		mu.Unlock()
	}()

	// Risks（ディレクトリ内の複数ファイル）
	wg.Add(1)
	go func() {
		defer wg.Done()
		riskFiles, err := fileStore.ListDir(ctx, "risks")
		if err != nil {
			return
		}
		var riskWg sync.WaitGroup
		result := make([]analysis.RiskInfo, 0, len(riskFiles))
		for _, file := range riskFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			riskWg.Add(1)
			file := file
			go func() {
				defer riskWg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				var risk core.RiskEntity
				if err := fileStore.ReadYaml(ctx, "risks/"+file, &risk); err == nil {
					score := riskScoreToInt(string(risk.RiskScore))
					info := analysis.RiskInfo{
						ID:            risk.ID,
						Title:         risk.Title,
						Probability:   string(risk.Probability),
						Impact:        string(risk.Impact),
						Score:         score,
						Status:        string(risk.Status),
						ObjectiveID:   risk.ObjectiveID,
						DeliverableID: risk.DeliverableID,
					}
					mu.Lock()
					result = append(result, info)
					mu.Unlock()
				}
			}()
		}
		riskWg.Wait()
		mu.Lock()
		risks = result
		mu.Unlock()
	}()

	wg.Wait()
	return
}
