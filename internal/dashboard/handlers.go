package dashboard

import (
	"encoding/json"
	"net/http"
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
