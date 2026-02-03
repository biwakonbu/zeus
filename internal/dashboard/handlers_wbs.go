package dashboard

import (
	"net/http"

	"github.com/biwakonbu/zeus/internal/analysis"
	"github.com/biwakonbu/zeus/internal/core"
)

// =============================================================================
// WBS API 型定義
// =============================================================================

// WBSResponse はWBS API のレスポンス
type WBSResponse struct {
	Roots    []*WBSNode `json:"roots"`
	MaxDepth int        `json:"max_depth"`
	Stats    WBSStats   `json:"stats"`
}

// WBSNode はWBS階層のノード
type WBSNode struct {
	ID       string     `json:"id"`
	Title    string     `json:"title"`
	NodeType string     `json:"node_type"` // vision, objective, deliverable, task
	WBSCode  string     `json:"wbs_code"`
	Status   string     `json:"status"`
	Progress int        `json:"progress"`
	Priority string     `json:"priority"`
	Assignee string     `json:"assignee"`
	Children []*WBSNode `json:"children,omitempty"`
	Depth    int        `json:"depth"`
}

// WBSStats はWBS統計
type WBSStats struct {
	TotalNodes   int `json:"total_nodes"`
	RootCount    int `json:"root_count"`
	LeafCount    int `json:"leaf_count"`
	MaxDepth     int `json:"max_depth"`
	AvgProgress  int `json:"avg_progress"`
	CompletedPct int `json:"completed_pct"`
}

// TimelineResponse はタイムライン API のレスポンス
type TimelineResponse struct {
	Items         []TimelineItem `json:"items"`
	CriticalPath  []string       `json:"critical_path"`
	ProjectStart  string         `json:"project_start"`
	ProjectEnd    string         `json:"project_end"`
	TotalDuration int            `json:"total_duration"`
	Stats         TimelineStats  `json:"stats"`
}

// TimelineItem はタイムライン上のアイテム
type TimelineItem struct {
	TaskID           string   `json:"task_id"`
	Title            string   `json:"title"`
	StartDate        string   `json:"start_date"`
	EndDate          string   `json:"end_date"`
	Progress         int      `json:"progress"`
	Status           string   `json:"status"`
	Priority         string   `json:"priority"`
	Assignee         string   `json:"assignee"`
	IsOnCriticalPath bool     `json:"is_on_critical_path"`
	Slack            int      `json:"slack"`
	Dependencies     []string `json:"dependencies"`
}

// TimelineStats はタイムライン統計
type TimelineStats struct {
	TotalTasks      int     `json:"total_tasks"`
	TasksWithDates  int     `json:"tasks_with_dates"`
	OnCriticalPath  int     `json:"on_critical_path"`
	AverageSlack    float64 `json:"average_slack"`
	OverdueTasks    int     `json:"overdue_tasks"`
	CompletedOnTime int     `json:"completed_on_time"`
}

// DownstreamResponse は下流タスク API のレスポンス
type DownstreamResponse struct {
	TaskID     string   `json:"task_id"`
	Downstream []string `json:"downstream"`
	Upstream   []string `json:"upstream"`
	Count      int      `json:"count"`
}

// WBSAnalysisResponse はWBS分析 API のレスポンス
type WBSAnalysisResponse struct {
	Coverage *CoverageAnalysisResult `json:"coverage"`
	Stale    *StaleAnalysisResult    `json:"stale"`
	Summary  AnalysisSummary         `json:"summary"`
}

// CoverageAnalysisResult はカバレッジ分析結果
type CoverageAnalysisResult struct {
	Issues          []CoverageIssue `json:"issues"`
	CoverageScore   int             `json:"coverage_score"`
	ObjectivesCover int             `json:"objectives_covered"`
	ObjectivesTotal int             `json:"objectives_total"`
	DeliverablesOk  int             `json:"deliverables_ok"`
	DeliverablesErr int             `json:"deliverables_err"`
}

// CoverageIssue はカバレッジ問題
type CoverageIssue struct {
	Type        string `json:"type"`
	EntityID    string `json:"entity_id"`
	EntityTitle string `json:"entity_title"`
	EntityType  string `json:"entity_type"`
	Severity    string `json:"severity"`
	Message     string `json:"message"`
}

// StaleAnalysisResult は陳腐化分析結果
type StaleAnalysisResult struct {
	StaleEntities []StaleEntity `json:"stale_entities"`
	TotalStale    int           `json:"total_stale"`
	ArchiveCount  int           `json:"archive_count"`
	ReviewCount   int           `json:"review_count"`
	DeleteCount   int           `json:"delete_count"`
}

// StaleEntity は陳腐化エンティティ
type StaleEntity struct {
	Type           string `json:"type"`
	EntityID       string `json:"entity_id"`
	EntityTitle    string `json:"entity_title"`
	EntityType     string `json:"entity_type"`
	Recommendation string `json:"recommendation"`
	Message        string `json:"message"`
	DaysStale      int    `json:"days_stale"`
}

// AnalysisSummary は分析サマリー
type AnalysisSummary struct {
	TotalObjectives   int    `json:"total_objectives"`
	CoveredObjectives int    `json:"covered_objectives"`
	OrphanedCount     int    `json:"orphaned_count"`
	StaleCount        int    `json:"stale_count"`
	OverallHealth     string `json:"overall_health"`
	HealthScore       int    `json:"health_score"`
}

// =============================================================================
// WBS API ハンドラー
// =============================================================================

// handleAPIWBS はWBS API を処理
func (s *Server) handleAPIWBS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	// 10概念モデル対応の WBS 階層を構築
	wbsTree, err := s.zeus.BuildWBSGraph(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// analysis.WBSNode から dashboard.WBSNode に変換
	response := WBSResponse{
		Roots:    convertWBSNodes(wbsTree.Roots),
		MaxDepth: wbsTree.MaxDepth,
		Stats: WBSStats{
			TotalNodes:   wbsTree.Stats.TotalNodes,
			RootCount:    wbsTree.Stats.RootCount,
			LeafCount:    wbsTree.Stats.LeafCount,
			MaxDepth:     wbsTree.Stats.MaxDepth,
			AvgProgress:  wbsTree.Stats.AvgProgress,
			CompletedPct: wbsTree.Stats.CompletedPct,
		},
	}

	if response.Roots == nil {
		response.Roots = []*WBSNode{}
	}

	writeJSON(w, http.StatusOK, response)
}

// convertWBSNodes は analysis.WBSNode を dashboard.WBSNode に変換
func convertWBSNodes(nodes []*analysis.WBSNode) []*WBSNode {
	if nodes == nil {
		return nil
	}

	result := make([]*WBSNode, len(nodes))
	for i, n := range nodes {
		result[i] = &WBSNode{
			ID:       n.ID,
			Title:    n.Title,
			NodeType: string(n.Type),
			WBSCode:  n.WBSCode,
			Status:   n.Status,
			Progress: n.Progress,
			Priority: n.Priority,
			Assignee: n.Assignee,
			Children: convertWBSNodes(n.Children),
			Depth:    n.Depth,
		}
	}
	return result
}

// handleAPITimeline はタイムライン API を処理
func (s *Server) handleAPITimeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	timeline, err := s.zeus.BuildTimeline(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// analysis.TimelineItem から dashboard.TimelineItem に変換
	items := make([]TimelineItem, len(timeline.Items))
	for i, item := range timeline.Items {
		deps := item.Dependencies
		if deps == nil {
			deps = []string{}
		}
		items[i] = TimelineItem{
			TaskID:           item.TaskID,
			Title:            item.Title,
			StartDate:        item.StartDate,
			EndDate:          item.EndDate,
			Progress:         item.Progress,
			Status:           item.Status,
			Priority:         item.Priority,
			Assignee:         item.Assignee,
			IsOnCriticalPath: item.IsOnCriticalPath,
			Slack:            item.Slack,
			Dependencies:     deps,
		}
	}

	criticalPath := timeline.CriticalPath
	if criticalPath == nil {
		criticalPath = []string{}
	}

	response := TimelineResponse{
		Items:         items,
		CriticalPath:  criticalPath,
		ProjectStart:  timeline.ProjectStart,
		ProjectEnd:    timeline.ProjectEnd,
		TotalDuration: timeline.TotalDuration,
		Stats: TimelineStats{
			TotalTasks:      timeline.Stats.TotalTasks,
			TasksWithDates:  timeline.Stats.TasksWithDates,
			OnCriticalPath:  timeline.Stats.OnCriticalPath,
			AverageSlack:    timeline.Stats.AverageSlack,
			OverdueTasks:    timeline.Stats.OverdueTasks,
			CompletedOnTime: timeline.Stats.CompletedOnTime,
		},
	}

	writeJSON(w, http.StatusOK, response)
}

// handleAPIDownstream は下流タスク API を処理
func (s *Server) handleAPIDownstream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	// クエリパラメータからタスクIDを取得
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "task_id パラメータが必要です")
		return
	}

	ctx := r.Context()
	graph, err := s.zeus.BuildDependencyGraph(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// タスクが存在するか確認
	if _, exists := graph.Nodes[taskID]; !exists {
		writeError(w, http.StatusNotFound, "指定されたタスクが見つかりません: "+taskID)
		return
	}

	// 下流（このタスクに依存しているタスク）と上流（このタスクが依存しているタスク）を取得
	downstream := graph.GetDownstreamTasks(taskID)
	upstream := graph.GetUpstreamTasks(taskID)

	if downstream == nil {
		downstream = []string{}
	}
	if upstream == nil {
		upstream = []string{}
	}

	response := DownstreamResponse{
		TaskID:     taskID,
		Downstream: downstream,
		Upstream:   upstream,
		Count:      len(downstream),
	}

	writeJSON(w, http.StatusOK, response)
}

// handleAPIWBSAnalysis はWBS分析 API を処理
func (s *Server) handleAPIWBSAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	fileStore := s.zeus.FileStore()

	// タスクを取得（tasks/active.yaml から直接読み込み）
	var taskStore core.TaskStore
	if err := fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		taskStore = core.TaskStore{Tasks: []core.Task{}}
	}
	tasks := make([]analysis.TaskInfo, len(taskStore.Tasks))
	for i, t := range taskStore.Tasks {
		// CompletedAt は core.Task にないので、完了状態の場合は UpdatedAt を使用
		completedAt := ""
		if t.Status == core.TaskStatusCompleted {
			completedAt = t.UpdatedAt
		}
		tasks[i] = analysis.TaskInfo{
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
		}
	}

	// Objective を取得（objectives/ ディレクトリから直接読み込み）
	objectives := []analysis.ObjectiveInfo{}
	objFiles, err := fileStore.ListDir(ctx, "objectives")
	if err == nil {
		for _, file := range objFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			var obj core.ObjectiveEntity
			if err := fileStore.ReadYaml(ctx, "objectives/"+file, &obj); err == nil {
				objectives = append(objectives, analysis.ObjectiveInfo{
					ID:        obj.ID,
					Title:     obj.Title,
					WBSCode:   obj.WBSCode,
					Progress:  obj.Progress,
					Status:    string(obj.Status),
					ParentID:  obj.ParentID,
					CreatedAt: obj.Metadata.CreatedAt,
					UpdatedAt: obj.Metadata.UpdatedAt,
				})
			}
		}
	}

	// Deliverable を取得（deliverables/ ディレクトリから直接読み込み）
	deliverables := []analysis.DeliverableInfo{}
	delFiles, err := fileStore.ListDir(ctx, "deliverables")
	if err == nil {
		for _, file := range delFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			var del core.DeliverableEntity
			if err := fileStore.ReadYaml(ctx, "deliverables/"+file, &del); err == nil {
				deliverables = append(deliverables, analysis.DeliverableInfo{
					ID:          del.ID,
					Title:       del.Title,
					ObjectiveID: del.ObjectiveID,
					Progress:    del.Progress,
					Status:      string(del.Status),
					CreatedAt:   del.Metadata.CreatedAt,
					UpdatedAt:   del.Metadata.UpdatedAt,
				})
			}
		}
	}

	// カバレッジ分析
	coverageAnalyzer := analysis.NewCoverageAnalyzer(objectives, deliverables, tasks)
	coverageResult, err := coverageAnalyzer.Analyze(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "カバレッジ分析エラー: "+err.Error())
		return
	}

	// 陳腐化分析
	staleAnalyzer := analysis.NewStaleAnalyzer(tasks, objectives, deliverables, nil)
	staleResult, err := staleAnalyzer.Analyze(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "陳腐化分析エラー: "+err.Error())
		return
	}

	// カバレッジ問題の変換
	coverageIssues := make([]CoverageIssue, len(coverageResult.Issues))
	for i, issue := range coverageResult.Issues {
		coverageIssues[i] = CoverageIssue{
			Type:        string(issue.Type),
			EntityID:    issue.EntityID,
			EntityTitle: issue.EntityTitle,
			EntityType:  issue.EntityType,
			Severity:    string(issue.Severity),
			Message:     issue.Message,
		}
	}

	// 陳腐化エンティティの変換
	staleEntities := make([]StaleEntity, len(staleResult.StaleEntities))
	for i, entity := range staleResult.StaleEntities {
		staleEntities[i] = StaleEntity{
			Type:           string(entity.Type),
			EntityID:       entity.EntityID,
			EntityTitle:    entity.EntityTitle,
			EntityType:     entity.EntityType,
			Recommendation: string(entity.Recommendation),
			Message:        entity.Message,
			DaysStale:      entity.DaysStale,
		}
	}

	// 孤立エンティティのカウント
	orphanedCount := 0
	for _, issue := range coverageResult.Issues {
		if issue.Type == analysis.CoverageIssueOrphaned {
			orphanedCount++
		}
	}

	// 健全性スコアと状態を計算
	healthScore := coverageResult.CoverageScore
	if staleResult.TotalStale > 0 {
		// 陳腐化エンティティがあると健全性スコアを下げる
		penalty := min(staleResult.TotalStale*5, 30)
		healthScore -= penalty
		if healthScore < 0 {
			healthScore = 0
		}
	}

	overallHealth := "good"
	if healthScore < 50 {
		overallHealth = "poor"
	} else if healthScore < 80 {
		overallHealth = "fair"
	}

	response := WBSAnalysisResponse{
		Coverage: &CoverageAnalysisResult{
			Issues:          coverageIssues,
			CoverageScore:   coverageResult.CoverageScore,
			ObjectivesCover: coverageResult.ObjectivesCover,
			ObjectivesTotal: coverageResult.ObjectivesTotal,
			DeliverablesOk:  coverageResult.DeliverablesOk,
			DeliverablesErr: coverageResult.DeliverablesErr,
		},
		Stale: &StaleAnalysisResult{
			StaleEntities: staleEntities,
			TotalStale:    staleResult.TotalStale,
			ArchiveCount:  staleResult.ArchiveCount,
			ReviewCount:   staleResult.ReviewCount,
			DeleteCount:   staleResult.DeleteCount,
		},
		Summary: AnalysisSummary{
			TotalObjectives:   coverageResult.ObjectivesTotal,
			CoveredObjectives: coverageResult.ObjectivesCover,
			OrphanedCount:     orphanedCount,
			StaleCount:        staleResult.TotalStale,
			OverallHealth:     overallHealth,
			HealthScore:       healthScore,
		},
	}

	writeJSON(w, http.StatusOK, response)
}
