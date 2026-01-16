package dashboard

import (
	"encoding/json"
	"net/http"

	"github.com/biwakonbu/zeus/internal/analysis"
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
	wbsTree, err := s.zeus.BuildWBSTree(ctx)
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

// writeJSON は JSON レスポンスを書き込む
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
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

