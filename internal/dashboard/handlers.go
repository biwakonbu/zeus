package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/biwakonbu/zeus/internal/analysis"
	"github.com/biwakonbu/zeus/internal/core"
)

// StatusResponse はステータス API のレスポンス
type StatusResponse struct {
	Project          ProjectInfo  `json:"project"`
	State            ProjectState `json:"state"`
	PendingApprovals int          `json:"pending_approvals"`
}

// ProjectInfo はプロジェクト情報
type ProjectInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	StartDate   string `json:"start_date"`
}

// ProjectState はプロジェクト状態
type ProjectState struct {
	Health  string    `json:"health"`
	Summary TaskStats `json:"summary"`
}

// TaskStats はタスク統計
type TaskStats struct {
	TotalTasks int `json:"total_tasks"`
	Completed  int `json:"completed"`
	InProgress int `json:"in_progress"`
	Pending    int `json:"pending"`
}

// TasksResponse はタスク一覧 API のレスポンス
type TasksResponse struct {
	Tasks []TaskItem `json:"tasks"`
	Total int        `json:"total"`
}

// TaskItem はタスクアイテム
type TaskItem struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Status       string   `json:"status"`
	Priority     string   `json:"priority"`
	Assignee     string   `json:"assignee"`
	Dependencies []string `json:"dependencies"`

	// Phase 6A: WBS・タイムライン機能用フィールド
	ParentID  string `json:"parent_id,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	DueDate   string `json:"due_date,omitempty"`
	Progress  int    `json:"progress"`
	WBSCode   string `json:"wbs_code,omitempty"`
}

// GraphResponse はグラフ API のレスポンス
type GraphResponse struct {
	Mermaid  string     `json:"mermaid"`
	Stats    GraphStats `json:"stats"`
	Cycles   [][]string `json:"cycles"`
	Isolated []string   `json:"isolated"`
}

// GraphStats はグラフ統計
type GraphStats struct {
	TotalNodes       int `json:"total_nodes"`
	WithDependencies int `json:"with_dependencies"`
	IsolatedCount    int `json:"isolated_count"`
	CycleCount       int `json:"cycle_count"`
	MaxDepth         int `json:"max_depth"`
}

// PredictResponse は予測 API のレスポンス
type PredictResponse struct {
	Completion *CompletionPrediction `json:"completion,omitempty"`
	Risk       *RiskPrediction       `json:"risk,omitempty"`
	Velocity   *VelocityReport       `json:"velocity,omitempty"`
}

// CompletionPrediction は完了予測
type CompletionPrediction struct {
	RemainingTasks    int     `json:"remaining_tasks"`
	AverageVelocity   float64 `json:"average_velocity"`
	EstimatedDate     string  `json:"estimated_date"`
	ConfidenceLevel   int     `json:"confidence_level"`
	MarginDays        int     `json:"margin_days"`
	HasSufficientData bool    `json:"has_sufficient_data"`
}

// RiskPrediction はリスク予測
type RiskPrediction struct {
	OverallLevel string       `json:"overall_level"`
	Factors      []RiskFactor `json:"factors"`
	Score        int          `json:"score"`
}

// RiskFactor はリスク要因
type RiskFactor struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Impact      int    `json:"impact"`
}

// VelocityReport はベロシティレポート
type VelocityReport struct {
	Last7Days     int     `json:"last_7_days"`
	Last14Days    int     `json:"last_14_days"`
	Last30Days    int     `json:"last_30_days"`
	WeeklyAverage float64 `json:"weekly_average"`
	Trend         string  `json:"trend"`
	DataPoints    int     `json:"data_points"`
}

// ErrorResponse はエラーレスポンス
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

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

// handleAPIStatus はステータス API を処理
func (s *Server) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	status, err := s.zeus.Status(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := StatusResponse{
		Project: ProjectInfo{
			ID:          status.Project.ID,
			Name:        status.Project.Name,
			Description: status.Project.Description,
			StartDate:   status.Project.StartDate,
		},
		State: ProjectState{
			Health: string(status.State.Health),
			Summary: TaskStats{
				TotalTasks: status.State.Summary.TotalTasks,
				Completed:  status.State.Summary.Completed,
				InProgress: status.State.Summary.InProgress,
				Pending:    status.State.Summary.Pending,
			},
		},
		PendingApprovals: status.PendingApprovals,
	}

	writeJSON(w, http.StatusOK, response)
}

// handleAPITasks はタスク一覧 API を処理
func (s *Server) handleAPITasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	result, err := s.zeus.List(ctx, "task")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tasks := make([]TaskItem, len(result.Items))
	for i, t := range result.Items {
		tasks[i] = TaskItem{
			ID:           t.ID,
			Title:        t.Title,
			Status:       string(t.Status),
			Priority:     string(t.Priority),
			Assignee:     t.Assignee,
			Dependencies: t.Dependencies,
			ParentID:     t.ParentID,
			StartDate:    t.StartDate,
			DueDate:      t.DueDate,
			Progress:     t.Progress,
			WBSCode:      t.WBSCode,
		}
	}

	response := TasksResponse{
		Tasks: tasks,
		Total: result.Total,
	}

	writeJSON(w, http.StatusOK, response)
}

// handleAPIGraph はグラフ API を処理
func (s *Server) handleAPIGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	graph, err := s.zeus.BuildDependencyGraph(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := GraphResponse{
		Mermaid: graph.ToMermaid(),
		Stats: GraphStats{
			TotalNodes:       graph.Stats.TotalNodes,
			WithDependencies: graph.Stats.WithDependencies,
			IsolatedCount:    graph.Stats.IsolatedCount,
			CycleCount:       graph.Stats.CycleCount,
			MaxDepth:         graph.Stats.MaxDepth,
		},
		Cycles:   graph.Cycles,
		Isolated: graph.Isolated,
	}

	if response.Cycles == nil {
		response.Cycles = [][]string{}
	}
	if response.Isolated == nil {
		response.Isolated = []string{}
	}

	writeJSON(w, http.StatusOK, response)
}

// handleAPIPredict は予測 API を処理
func (s *Server) handleAPIPredict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	result, err := s.zeus.Predict(ctx, "all")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := PredictResponse{}

	if result.Completion != nil {
		response.Completion = &CompletionPrediction{
			RemainingTasks:    result.Completion.RemainingTasks,
			AverageVelocity:   result.Completion.AverageVelocity,
			EstimatedDate:     result.Completion.EstimatedDate,
			ConfidenceLevel:   result.Completion.ConfidenceLevel,
			MarginDays:        result.Completion.MarginDays,
			HasSufficientData: result.Completion.HasSufficientData,
		}
	}

	if result.Risk != nil {
		factors := make([]RiskFactor, len(result.Risk.Factors))
		for i, f := range result.Risk.Factors {
			factors[i] = RiskFactor{
				Name:        f.Name,
				Description: f.Description,
				Impact:      f.Impact,
			}
		}
		response.Risk = &RiskPrediction{
			OverallLevel: string(result.Risk.OverallLevel),
			Factors:      factors,
			Score:        result.Risk.Score,
		}
	}

	if result.Velocity != nil {
		response.Velocity = &VelocityReport{
			Last7Days:     result.Velocity.Last7Days,
			Last14Days:    result.Velocity.Last14Days,
			Last30Days:    result.Velocity.Last30Days,
			WeeklyAverage: result.Velocity.WeeklyAverage,
			Trend:         string(result.Velocity.Trend),
			DataPoints:    result.Velocity.DataPoints,
		}
	}

	writeJSON(w, http.StatusOK, response)
}

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

// BottleneckResponse はボトルネック API のレスポンス
type BottleneckResponse struct {
	Bottlenecks []BottleneckItem  `json:"bottlenecks"`
	Summary     BottleneckSummary `json:"summary"`
}

// BottleneckItem はボトルネック項目
type BottleneckItem struct {
	Type       string   `json:"type"`
	Severity   string   `json:"severity"`
	Entities   []string `json:"entities"`
	Message    string   `json:"message"`
	Impact     string   `json:"impact"`
	Suggestion string   `json:"suggestion"`
}

// BottleneckSummary はボトルネックのサマリー
type BottleneckSummary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Warning  int `json:"warning"`
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

// writeJSON は JSON レスポンスを書き込む
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError はエラーレスポンスを書き込む
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}

// handleSSE は Server-Sent Events 接続を処理
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	// SSE に必要なヘッダーを設定
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// CORS ヘッダー（開発モード時）
	if s.devMode {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	// Flusher を取得
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// クライアント ID を生成（UUID の代わりにシンプルな形式）
	clientID := r.RemoteAddr + "-" + r.Header.Get("X-Request-ID")
	if clientID == r.RemoteAddr+"-" {
		clientID = r.RemoteAddr + "-" + string(rune(s.broadcaster.ClientCount()))
	}

	// クライアントを登録
	client := s.broadcaster.AddClient(clientID)
	defer s.broadcaster.RemoveClient(clientID)

	// 接続確立メッセージを送信
	_, _ = w.Write([]byte("event: connected\ndata: {\"client_id\":\"" + clientID + "\"}\n\n"))
	flusher.Flush()

	// クライアントの切断を検知
	ctx := r.Context()

	// イベントループ
	for {
		select {
		case <-ctx.Done():
			// クライアントが切断
			return
		case event, ok := <-client.Events:
			if !ok {
				// チャネルがクローズ
				return
			}

			// イベントデータを JSON にエンコード
			data, err := FormatSSEMessage(event)
			if err != nil {
				continue
			}

			// SSE 形式で送信
			_, err = w.Write([]byte("event: " + string(event.Type) + "\ndata: " + string(data) + "\n\n"))
			if err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// hasYamlSuffix は .yaml または .yml 拡張子を持つかチェック
func hasYamlSuffix(filename string) bool {
	return strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")
}

// handleAPIBottlenecks はボトルネック API を処理
func (s *Server) handleAPIBottlenecks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	fileStore := s.zeus.FileStore()

	// タスクを取得
	var taskStore core.TaskStore
	if err := fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		taskStore = core.TaskStore{Tasks: []core.Task{}}
	}
	tasks := make([]analysis.TaskInfo, len(taskStore.Tasks))
	for i, t := range taskStore.Tasks {
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

	// Objective を取得
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

	// Deliverable を取得
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

	// Risk を取得
	risks := []analysis.RiskInfo{}
	riskFiles, err := fileStore.ListDir(ctx, "risks")
	if err == nil {
		for _, file := range riskFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			var risk core.RiskEntity
			if err := fileStore.ReadYaml(ctx, "risks/"+file, &risk); err == nil {
				// RiskScore を数値スコアに変換
				score := riskScoreToInt(string(risk.RiskScore))
				risks = append(risks, analysis.RiskInfo{
					ID:          risk.ID,
					Title:       risk.Title,
					Probability: string(risk.Probability),
					Impact:      string(risk.Impact),
					Score:       score,
					Status:      string(risk.Status),
				})
			}
		}
	}

	// ボトルネック分析を実行
	analyzer := analysis.NewBottleneckAnalyzer(tasks, objectives, deliverables, risks, nil)
	result, err := analyzer.Analyze(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ボトルネック分析エラー: "+err.Error())
		return
	}

	// レスポンス変換
	bottlenecks := make([]BottleneckItem, len(result.Bottlenecks))
	for i, b := range result.Bottlenecks {
		entities := b.Entities
		if entities == nil {
			entities = []string{}
		}
		bottlenecks[i] = BottleneckItem{
			Type:       string(b.Type),
			Severity:   string(b.Severity),
			Entities:   entities,
			Message:    b.Message,
			Impact:     b.Impact,
			Suggestion: b.Suggestion,
		}
	}

	response := BottleneckResponse{
		Bottlenecks: bottlenecks,
		Summary: BottleneckSummary{
			Critical: result.Summary.Critical,
			High:     result.Summary.High,
			Medium:   result.Summary.Medium,
			Warning:  result.Summary.Warning,
		},
	}

	writeJSON(w, http.StatusOK, response)
}

// riskScoreToInt は RiskScore 文字列を数値スコアに変換
func riskScoreToInt(score string) int {
	switch score {
	case "critical":
		return 9
	case "high":
		return 6
	case "medium":
		return 4
	case "low":
		return 2
	default:
		return 0
	}
}

// =============================================================================
// WBS Aggregated API（4視点用の集約データ）
// =============================================================================

// WBSAggregatedResponse は WBS 集約 API のレスポンス
type WBSAggregatedResponse struct {
	Progress  *ProgressAggregation `json:"progress"`
	Issues    *IssueAggregation    `json:"issues"`
	Coverage  *CoverageAggregation `json:"coverage"`
	Resources *ResourceAggregation `json:"resources"`
}

// ProgressAggregation は進捗集約データ（ツリーマップ用）
type ProgressAggregation struct {
	Vision        *ProgressNode   `json:"vision,omitempty"`
	Objectives    []*ProgressNode `json:"objectives"`
	TotalProgress int             `json:"total_progress"`
}

// ProgressNode は進捗ツリーのノード
type ProgressNode struct {
	ID            string          `json:"id"`
	Title         string          `json:"title"`
	NodeType      string          `json:"node_type"`
	Progress      int             `json:"progress"`
	Status        string          `json:"status"`
	ChildrenCount int             `json:"children_count"`
	Children      []*ProgressNode `json:"children,omitempty"`
}

// IssueAggregation は問題集中データ（バブルチャート用）
type IssueAggregation struct {
	Items       []*IssueBubble `json:"items"`
	TotalIssues int            `json:"total_issues"`
	MaxSeverity string         `json:"max_severity"`
}

// IssueBubble はバブルチャート用のデータ
type IssueBubble struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	NodeType     string  `json:"node_type"`
	ProblemCount int     `json:"problem_count"`
	RiskCount    int     `json:"risk_count"`
	TotalIssues  int     `json:"total_issues"`
	MaxSeverity  string  `json:"max_severity"`
	RiskScore    float64 `json:"risk_score"`
	Progress     int     `json:"progress"`
}

// CoverageAggregation はカバレッジデータ（サンバースト用）
type CoverageAggregation struct {
	Root          *CoverageNode `json:"root"`
	CoverageScore int           `json:"coverage_score"`
	OrphanedTasks []string      `json:"orphaned_tasks"`
	MissingLinks  []string      `json:"missing_links"`
}

// CoverageNode はサンバースト用のノード
type CoverageNode struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	NodeType  string          `json:"node_type"`
	HasIssue  bool            `json:"has_issue"`
	IssueType string          `json:"issue_type,omitempty"`
	Value     int             `json:"value"`
	Children  []*CoverageNode `json:"children,omitempty"`
}

// ResourceAggregation はリソース配分データ（ヒートマップ用）
type ResourceAggregation struct {
	Assignees  []string         `json:"assignees"`
	Objectives []string         `json:"objectives"`
	Matrix     [][]ResourceCell `json:"matrix"`
}

// ResourceCell はヒートマップのセル
type ResourceCell struct {
	TaskCount    int `json:"task_count"`
	Progress     int `json:"progress"`
	BlockedCount int `json:"blocked_count"`
}

// handleAPIWBSAggregated は WBS 集約 API を処理
func (s *Server) handleAPIWBSAggregated(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	fileStore := s.zeus.FileStore()

	// タスクを取得
	var taskStore core.TaskStore
	if err := fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		taskStore = core.TaskStore{Tasks: []core.Task{}}
	}

	// Objective を取得
	objectives := []core.ObjectiveEntity{}
	objFiles, err := fileStore.ListDir(ctx, "objectives")
	if err == nil {
		for _, file := range objFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			var obj core.ObjectiveEntity
			if err := fileStore.ReadYaml(ctx, "objectives/"+file, &obj); err == nil {
				objectives = append(objectives, obj)
			}
		}
	}

	// Deliverable を取得
	deliverables := []core.DeliverableEntity{}
	delFiles, err := fileStore.ListDir(ctx, "deliverables")
	if err == nil {
		for _, file := range delFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			var del core.DeliverableEntity
			if err := fileStore.ReadYaml(ctx, "deliverables/"+file, &del); err == nil {
				deliverables = append(deliverables, del)
			}
		}
	}

	// Vision を取得
	var vision core.Vision
	_ = fileStore.ReadYaml(ctx, "vision.yaml", &vision)

	// Problem を取得
	problems := []core.ProblemEntity{}
	probFiles, err := fileStore.ListDir(ctx, "problems")
	if err == nil {
		for _, file := range probFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			var prob core.ProblemEntity
			if err := fileStore.ReadYaml(ctx, "problems/"+file, &prob); err == nil {
				problems = append(problems, prob)
			}
		}
	}

	// Risk を取得
	risks := []core.RiskEntity{}
	riskFiles, err := fileStore.ListDir(ctx, "risks")
	if err == nil {
		for _, file := range riskFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			var risk core.RiskEntity
			if err := fileStore.ReadYaml(ctx, "risks/"+file, &risk); err == nil {
				risks = append(risks, risk)
			}
		}
	}

	// 集約データを構築
	response := WBSAggregatedResponse{
		Progress:  buildProgressAggregation(vision, objectives, deliverables, taskStore.Tasks),
		Issues:    buildIssueAggregation(objectives, deliverables, problems, risks),
		Coverage:  buildCoverageAggregation(vision, objectives, deliverables, taskStore.Tasks),
		Resources: buildResourceAggregation(objectives, taskStore.Tasks),
	}

	writeJSON(w, http.StatusOK, response)
}

// buildProgressAggregation は進捗集約データを構築
func buildProgressAggregation(vision core.Vision, objectives []core.ObjectiveEntity, deliverables []core.DeliverableEntity, tasks []core.Task) *ProgressAggregation {
	result := &ProgressAggregation{
		Objectives: []*ProgressNode{},
	}

	// Objective ID → Deliverables マップ
	objDeliverables := make(map[string][]core.DeliverableEntity)
	for _, del := range deliverables {
		objDeliverables[del.ObjectiveID] = append(objDeliverables[del.ObjectiveID], del)
	}

	// Deliverable ID → Tasks マップ
	// （タスクの DeliverableID がないので、ParentID から判定）
	delTasks := make(map[string][]core.Task)
	for _, task := range tasks {
		if task.ParentID != "" {
			delTasks[task.ParentID] = append(delTasks[task.ParentID], task)
		}
	}

	// Objective ごとの進捗計算
	totalProgress := 0
	for _, obj := range objectives {
		objNode := &ProgressNode{
			ID:       obj.ID,
			Title:    obj.Title,
			NodeType: "objective",
			Progress: obj.Progress,
			Status:   string(obj.Status),
			Children: []*ProgressNode{},
		}

		// 関連 Deliverable を追加
		for _, del := range objDeliverables[obj.ID] {
			delNode := &ProgressNode{
				ID:       del.ID,
				Title:    del.Title,
				NodeType: "deliverable",
				Progress: del.Progress,
				Status:   string(del.Status),
				Children: []*ProgressNode{},
			}

			// 関連タスクを追加
			for _, task := range delTasks[del.ID] {
				taskNode := &ProgressNode{
					ID:       task.ID,
					Title:    task.Title,
					NodeType: "task",
					Progress: task.Progress,
					Status:   string(task.Status),
				}
				delNode.Children = append(delNode.Children, taskNode)
			}
			delNode.ChildrenCount = len(delNode.Children)

			objNode.Children = append(objNode.Children, delNode)
		}
		objNode.ChildrenCount = len(objNode.Children)

		result.Objectives = append(result.Objectives, objNode)
		totalProgress += obj.Progress
	}

	// 平均進捗
	if len(objectives) > 0 {
		result.TotalProgress = totalProgress / len(objectives)
	}

	// Vision を追加（存在する場合）
	if vision.Title != "" {
		result.Vision = &ProgressNode{
			ID:            "vision",
			Title:         vision.Title,
			NodeType:      "vision",
			Progress:      result.TotalProgress,
			Status:        "active",
			ChildrenCount: len(objectives),
		}
	}

	return result
}

// buildIssueAggregation は問題集中データを構築
func buildIssueAggregation(objectives []core.ObjectiveEntity, deliverables []core.DeliverableEntity, problems []core.ProblemEntity, risks []core.RiskEntity) *IssueAggregation {
	result := &IssueAggregation{
		Items:       []*IssueBubble{},
		MaxSeverity: "low",
	}

	// エンティティ ID → 問題/リスク集計
	issueMap := make(map[string]*IssueBubble)

	// Objective 用のバブルを作成
	for _, obj := range objectives {
		issueMap[obj.ID] = &IssueBubble{
			ID:       obj.ID,
			Title:    obj.Title,
			NodeType: "objective",
			Progress: obj.Progress,
		}
	}

	// Deliverable 用のバブルを作成
	for _, del := range deliverables {
		issueMap[del.ID] = &IssueBubble{
			ID:       del.ID,
			Title:    del.Title,
			NodeType: "deliverable",
			Progress: del.Progress,
		}
	}

	// Problem を集計
	for _, prob := range problems {
		targetID := prob.ObjectiveID
		if targetID == "" {
			targetID = prob.DeliverableID
		}
		if targetID == "" {
			continue
		}
		if bubble, ok := issueMap[targetID]; ok {
			bubble.ProblemCount++
			bubble.TotalIssues++
			// 深刻度を更新
			if severityRank(string(prob.Severity)) > severityRank(bubble.MaxSeverity) {
				bubble.MaxSeverity = string(prob.Severity)
			}
		}
	}

	// Risk を集計
	for _, risk := range risks {
		targetID := risk.ObjectiveID
		if targetID == "" {
			targetID = risk.DeliverableID
		}
		if targetID == "" {
			continue
		}
		if bubble, ok := issueMap[targetID]; ok {
			bubble.RiskCount++
			bubble.TotalIssues++
			// リスクスコアを加算
			bubble.RiskScore += float64(riskScoreToInt(string(risk.RiskScore)))
			// 深刻度を更新
			riskSeverity := riskScoreToSeverity(string(risk.RiskScore))
			if severityRank(riskSeverity) > severityRank(bubble.MaxSeverity) {
				bubble.MaxSeverity = riskSeverity
			}
		}
	}

	// 結果に追加（問題がある項目のみ）
	for _, bubble := range issueMap {
		if bubble.TotalIssues > 0 {
			result.Items = append(result.Items, bubble)
			result.TotalIssues += bubble.TotalIssues
			if severityRank(bubble.MaxSeverity) > severityRank(result.MaxSeverity) {
				result.MaxSeverity = bubble.MaxSeverity
			}
		}
	}

	return result
}

// severityRank は深刻度のランクを返す
func severityRank(severity string) int {
	switch severity {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// riskScoreToSeverity はリスクスコアを深刻度に変換
func riskScoreToSeverity(score string) string {
	switch score {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "medium":
		return "medium"
	default:
		return "low"
	}
}

// buildCoverageAggregation はカバレッジデータを構築
func buildCoverageAggregation(vision core.Vision, objectives []core.ObjectiveEntity, deliverables []core.DeliverableEntity, tasks []core.Task) *CoverageAggregation {
	result := &CoverageAggregation{
		OrphanedTasks: []string{},
		MissingLinks:  []string{},
	}

	// Objective ID → Deliverables マップ
	objDeliverables := make(map[string][]core.DeliverableEntity)
	for _, del := range deliverables {
		objDeliverables[del.ObjectiveID] = append(objDeliverables[del.ObjectiveID], del)
	}

	// Deliverable ID → Tasks マップ
	delTasks := make(map[string][]core.Task)
	linkedTaskIDs := make(map[string]bool)
	for _, task := range tasks {
		if task.ParentID != "" {
			delTasks[task.ParentID] = append(delTasks[task.ParentID], task)
			linkedTaskIDs[task.ID] = true
		}
	}

	// 孤立タスクを検出
	for _, task := range tasks {
		if !linkedTaskIDs[task.ID] {
			result.OrphanedTasks = append(result.OrphanedTasks, task.ID)
		}
	}

	// ルートノードを構築
	root := &CoverageNode{
		ID:       "vision",
		Title:    vision.Title,
		NodeType: "vision",
		Value:    1,
		Children: []*CoverageNode{},
	}
	if root.Title == "" {
		root.Title = "Project"
	}

	coveredCount := 0
	for _, obj := range objectives {
		objNode := &CoverageNode{
			ID:       obj.ID,
			Title:    obj.Title,
			NodeType: "objective",
			Value:    1,
			Children: []*CoverageNode{},
		}

		// Deliverable なしの Objective をマーク
		dels := objDeliverables[obj.ID]
		if len(dels) == 0 {
			objNode.HasIssue = true
			objNode.IssueType = "no_deliverables"
			result.MissingLinks = append(result.MissingLinks, obj.ID)
		} else {
			coveredCount++
		}

		// Deliverable を追加
		for _, del := range dels {
			delNode := &CoverageNode{
				ID:       del.ID,
				Title:    del.Title,
				NodeType: "deliverable",
				Value:    1,
				Children: []*CoverageNode{},
			}

			// Task なしの Deliverable をマーク
			taskList := delTasks[del.ID]
			if len(taskList) == 0 {
				delNode.HasIssue = true
				delNode.IssueType = "no_tasks"
				result.MissingLinks = append(result.MissingLinks, del.ID)
			}

			// タスクを追加
			for _, task := range taskList {
				taskNode := &CoverageNode{
					ID:       task.ID,
					Title:    task.Title,
					NodeType: "task",
					Value:    1,
				}
				delNode.Children = append(delNode.Children, taskNode)
			}

			objNode.Children = append(objNode.Children, delNode)
		}

		root.Children = append(root.Children, objNode)
	}

	result.Root = root

	// カバレッジスコアを計算
	if len(objectives) > 0 {
		result.CoverageScore = (coveredCount * 100) / len(objectives)
	} else {
		result.CoverageScore = 100
	}

	return result
}

// buildResourceAggregation はリソース配分データを構築
func buildResourceAggregation(objectives []core.ObjectiveEntity, tasks []core.Task) *ResourceAggregation {
	result := &ResourceAggregation{
		Assignees:  []string{},
		Objectives: []string{},
		Matrix:     [][]ResourceCell{},
	}

	// 担当者一覧を収集
	assigneeSet := make(map[string]bool)
	for _, task := range tasks {
		if task.Assignee != "" {
			assigneeSet[task.Assignee] = true
		}
	}
	for assignee := range assigneeSet {
		result.Assignees = append(result.Assignees, assignee)
	}

	// Objective 一覧
	objIDs := make(map[string]int) // ID → index
	for i, obj := range objectives {
		result.Objectives = append(result.Objectives, obj.Title)
		objIDs[obj.ID] = i
	}

	// 担当者がいない場合は空のマトリクスを返す
	if len(result.Assignees) == 0 || len(objectives) == 0 {
		return result
	}

	// マトリクスを初期化
	result.Matrix = make([][]ResourceCell, len(result.Assignees))
	for i := range result.Matrix {
		result.Matrix[i] = make([]ResourceCell, len(objectives))
	}

	// タスク → Objective のマッピング（簡易版: ParentID から推定）
	// 実際には Deliverable → Objective のリンクを辿る必要がある
	assigneeIdx := make(map[string]int)
	for i, a := range result.Assignees {
		assigneeIdx[a] = i
	}

	for _, task := range tasks {
		if task.Assignee == "" {
			continue
		}
		aIdx, ok := assigneeIdx[task.Assignee]
		if !ok {
			continue
		}

		// 仮: 最初の Objective に割り当て（実際はリンクを辿る）
		oIdx := 0
		if task.ParentID != "" {
			if idx, ok := objIDs[task.ParentID]; ok {
				oIdx = idx
			}
		}
		if oIdx >= len(objectives) {
			oIdx = 0
		}

		result.Matrix[aIdx][oIdx].TaskCount++
		result.Matrix[aIdx][oIdx].Progress += task.Progress
		if task.Status == core.TaskStatusBlocked {
			result.Matrix[aIdx][oIdx].BlockedCount++
		}
	}

	// 進捗を平均化
	for i := range result.Matrix {
		for j := range result.Matrix[i] {
			if result.Matrix[i][j].TaskCount > 0 {
				result.Matrix[i][j].Progress /= result.Matrix[i][j].TaskCount
			}
		}
	}

	return result
}

// =============================================================================
// Affinity API（Phase 7: Affinity Canvas）
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

	// Tasks（単一ファイル、直接読み込み）
	wg.Add(1)
	go func() {
		defer wg.Done()
		var taskStore core.TaskStore
		if err := fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
			taskStore = core.TaskStore{Tasks: []core.Task{}}
		}
		result := make([]analysis.TaskInfo, len(taskStore.Tasks))
		for i, t := range taskStore.Tasks {
			completedAt := ""
			if t.Status == core.TaskStatusCompleted {
				completedAt = t.UpdatedAt
			}
			result[i] = analysis.TaskInfo{
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

// =============================================================================
// UML UseCase API
// =============================================================================

// ActorItem はアクター API のアイテム
type ActorItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// ActorsResponse はアクター一覧 API のレスポンス
type ActorsResponse struct {
	Actors []ActorItem `json:"actors"`
	Total  int         `json:"total"`
}

// UseCaseActorRefItem はユースケースアクター参照 API のアイテム
type UseCaseActorRefItem struct {
	ActorID string `json:"actor_id"`
	Role    string `json:"role"`
}

// UseCaseRelationItem はユースケースリレーション API のアイテム
type UseCaseRelationItem struct {
	Type           string `json:"type"`
	TargetID       string `json:"target_id"`
	Condition      string `json:"condition,omitempty"`
	ExtensionPoint string `json:"extension_point,omitempty"`
}

// UseCaseItem はユースケース API のアイテム
type UseCaseItem struct {
	ID          string                `json:"id"`
	Title       string                `json:"title"`
	Description string                `json:"description,omitempty"`
	Status      string                `json:"status"`
	ObjectiveID string                `json:"objective_id,omitempty"`
	Actors      []UseCaseActorRefItem `json:"actors"`
	Relations   []UseCaseRelationItem `json:"relations"`
}

// UseCasesResponse はユースケース一覧 API のレスポンス
type UseCasesResponse struct {
	UseCases []UseCaseItem `json:"usecases"`
	Total    int           `json:"total"`
}

// UseCaseDiagramResponse はユースケース図 API のレスポンス
type UseCaseDiagramResponse struct {
	Actors   []ActorItem   `json:"actors"`
	UseCases []UseCaseItem `json:"usecases"`
	Boundary string        `json:"boundary"`
	Mermaid  string        `json:"mermaid"`
}

// handleAPIActors はアクター一覧 API を処理
func (s *Server) handleAPIActors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	fileStore := s.zeus.FileStore()

	var actorsFile core.ActorsFile
	if err := fileStore.ReadYaml(ctx, "actors.yaml", &actorsFile); err != nil {
		actorsFile = core.ActorsFile{Actors: []core.ActorEntity{}}
	}

	actors := make([]ActorItem, len(actorsFile.Actors))
	for i, a := range actorsFile.Actors {
		actors[i] = ActorItem{
			ID:          a.ID,
			Title:       a.Title,
			Type:        string(a.Type),
			Description: a.Description,
		}
	}

	response := ActorsResponse{
		Actors: actors,
		Total:  len(actors),
	}

	writeJSON(w, http.StatusOK, response)
}

// handleAPIUseCases はユースケース一覧 API を処理
func (s *Server) handleAPIUseCases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	fileStore := s.zeus.FileStore()

	// usecases ディレクトリからファイル一覧を取得
	files, err := fileStore.ListDir(ctx, "usecases")
	if err != nil {
		// ディレクトリが存在しない場合は空リストを返す
		response := UseCasesResponse{
			UseCases: []UseCaseItem{},
			Total:    0,
		}
		writeJSON(w, http.StatusOK, response)
		return
	}

	usecases := make([]UseCaseItem, 0)
	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var uc core.UseCaseEntity
		if err := fileStore.ReadYaml(ctx, "usecases/"+file, &uc); err != nil {
			continue
		}

		// アクター参照の変換
		actors := make([]UseCaseActorRefItem, len(uc.Actors))
		for j, ar := range uc.Actors {
			actors[j] = UseCaseActorRefItem{
				ActorID: ar.ActorID,
				Role:    string(ar.Role),
			}
		}

		// リレーションの変換
		relations := make([]UseCaseRelationItem, len(uc.Relations))
		for j, rel := range uc.Relations {
			relations[j] = UseCaseRelationItem{
				Type:           string(rel.Type),
				TargetID:       rel.TargetID,
				Condition:      rel.Condition,
				ExtensionPoint: rel.ExtensionPoint,
			}
		}

		usecases = append(usecases, UseCaseItem{
			ID:          uc.ID,
			Title:       uc.Title,
			Description: uc.Description,
			Status:      string(uc.Status),
			ObjectiveID: uc.ObjectiveID,
			Actors:      actors,
			Relations:   relations,
		})
	}

	response := UseCasesResponse{
		UseCases: usecases,
		Total:    len(usecases),
	}

	writeJSON(w, http.StatusOK, response)
}

// handleAPIUseCaseDiagram はユースケース図 API を処理
func (s *Server) handleAPIUseCaseDiagram(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	fileStore := s.zeus.FileStore()

	// クエリパラメータからシステム境界名を取得
	boundary := r.URL.Query().Get("boundary")
	if boundary == "" {
		boundary = "System"
	}

	// アクターを取得
	var actorsFile core.ActorsFile
	if err := fileStore.ReadYaml(ctx, "actors.yaml", &actorsFile); err != nil {
		actorsFile = core.ActorsFile{Actors: []core.ActorEntity{}}
	}

	actors := make([]ActorItem, len(actorsFile.Actors))
	for i, a := range actorsFile.Actors {
		actors[i] = ActorItem{
			ID:          a.ID,
			Title:       a.Title,
			Type:        string(a.Type),
			Description: a.Description,
		}
	}

	// ユースケースを取得
	files, _ := fileStore.ListDir(ctx, "usecases")
	usecases := make([]UseCaseItem, 0)
	ucEntities := make([]core.UseCaseEntity, 0)

	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var uc core.UseCaseEntity
		if err := fileStore.ReadYaml(ctx, "usecases/"+file, &uc); err != nil {
			continue
		}

		ucEntities = append(ucEntities, uc)

		// アクター参照の変換
		ucActors := make([]UseCaseActorRefItem, len(uc.Actors))
		for j, ar := range uc.Actors {
			ucActors[j] = UseCaseActorRefItem{
				ActorID: ar.ActorID,
				Role:    string(ar.Role),
			}
		}

		// リレーションの変換
		relations := make([]UseCaseRelationItem, len(uc.Relations))
		for j, rel := range uc.Relations {
			relations[j] = UseCaseRelationItem{
				Type:           string(rel.Type),
				TargetID:       rel.TargetID,
				Condition:      rel.Condition,
				ExtensionPoint: rel.ExtensionPoint,
			}
		}

		usecases = append(usecases, UseCaseItem{
			ID:          uc.ID,
			Title:       uc.Title,
			Description: uc.Description,
			Status:      string(uc.Status),
			ObjectiveID: uc.ObjectiveID,
			Actors:      ucActors,
			Relations:   relations,
		})
	}

	// Mermaid 形式でユースケース図を生成
	mermaid := generateUseCaseMermaid(actorsFile.Actors, ucEntities, boundary)

	response := UseCaseDiagramResponse{
		Actors:   actors,
		UseCases: usecases,
		Boundary: boundary,
		Mermaid:  mermaid,
	}

	writeJSON(w, http.StatusOK, response)
}

// generateUseCaseMermaid は Mermaid 形式でユースケース図を生成
func generateUseCaseMermaid(actors []core.ActorEntity, usecases []core.UseCaseEntity, boundary string) string {
	var sb strings.Builder

	sb.WriteString("flowchart LR\n")

	// アクター定義
	sb.WriteString("    %% Actors\n")
	for _, actor := range actors {
		mermaidID := strings.ReplaceAll(actor.ID, "-", "_")
		typeEmoji := actorTypeEmoji(actor.Type)
		sb.WriteString("    " + mermaidID + "[" + typeEmoji + " " + escapeForMermaidDiagram(actor.Title) + "]\n")
	}

	// システム境界サブグラフ
	sb.WriteString("\n    subgraph boundary[" + escapeForMermaidDiagram(boundary) + "]\n")

	// ユースケース定義
	sb.WriteString("        %% UseCases\n")
	for _, uc := range usecases {
		mermaidID := strings.ReplaceAll(uc.ID, "-", "_")
		sb.WriteString("        " + mermaidID + "((" + escapeForMermaidDiagram(uc.Title) + "))\n")
	}

	sb.WriteString("    end\n")

	// アクターとユースケースの関連
	sb.WriteString("\n    %% Actor-UseCase Relations\n")
	for _, uc := range usecases {
		ucID := strings.ReplaceAll(uc.ID, "-", "_")
		for _, actorRef := range uc.Actors {
			actorID := strings.ReplaceAll(actorRef.ActorID, "-", "_")
			if actorRef.Role == core.ActorRolePrimary {
				sb.WriteString("    " + actorID + " ==> " + ucID + "\n")
			} else {
				sb.WriteString("    " + actorID + " --> " + ucID + "\n")
			}
		}
	}

	// ユースケース間のリレーション
	sb.WriteString("\n    %% UseCase Relations\n")
	for _, uc := range usecases {
		ucID := strings.ReplaceAll(uc.ID, "-", "_")
		for _, rel := range uc.Relations {
			targetID := strings.ReplaceAll(rel.TargetID, "-", "_")
			switch rel.Type {
			case core.RelationTypeInclude:
				sb.WriteString("    " + ucID + " -.->|include| " + targetID + "\n")
			case core.RelationTypeExtend:
				label := "extend"
				if rel.Condition != "" {
					label = "extend [" + rel.Condition + "]"
				}
				sb.WriteString("    " + targetID + " -.->|" + escapeForMermaidDiagram(label) + "| " + ucID + "\n")
			case core.RelationTypeGeneralize:
				sb.WriteString("    " + ucID + " -->|generalize| " + targetID + "\n")
			}
		}
	}

	return sb.String()
}

// actorTypeEmoji はアクタータイプの絵文字を返す
func actorTypeEmoji(t core.ActorType) string {
	switch t {
	case core.ActorTypeHuman:
		return "👤"
	case core.ActorTypeSystem:
		return "🖥️"
	case core.ActorTypeTime:
		return "⏰"
	case core.ActorTypeDevice:
		return "📱"
	case core.ActorTypeExternal:
		return "🌐"
	default:
		return "❓"
	}
}

// escapeForMermaidDiagram は Mermaid 用にエスケープ
func escapeForMermaidDiagram(s string) string {
	s = strings.ReplaceAll(s, "\"", "'")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
