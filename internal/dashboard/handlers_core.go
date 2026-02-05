package dashboard

import (
	"net/http"

	"github.com/biwakonbu/zeus/internal/core"
)

// =============================================================================
// Core API 型定義
// =============================================================================

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

// =============================================================================
// Core API ハンドラー
// =============================================================================

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

// =============================================================================
// Tasks API ハンドラー（Activity を TaskItem 形式で返す）
// =============================================================================

// TasksResponse はタスク一覧 API のレスポンス
type TasksResponse struct {
	Tasks []TaskItem `json:"tasks"`
	Total int        `json:"total"`
}

// TaskItem はタスク項目
type TaskItem struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Status       string   `json:"status"`
	Priority     string   `json:"priority"`
	Assignee     string   `json:"assignee"`
	Dependencies []string `json:"dependencies"`
	ParentID     string   `json:"parent_id,omitempty"`
	StartDate    string   `json:"start_date,omitempty"`
	DueDate      string   `json:"due_date,omitempty"`
	Progress     int      `json:"progress"`
	WBSCode      string   `json:"wbs_code,omitempty"`
}

// handleAPITasks はタスク一覧 API を処理（Activity を TaskItem 形式で返す）
func (s *Server) handleAPITasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	fileStore := s.zeus.FileStore()

	// Activity を取得
	files, err := fileStore.ListDir(ctx, "activities")
	if err != nil {
		// ディレクトリが存在しない場合は空のレスポンス
		response := TasksResponse{
			Tasks: []TaskItem{},
			Total: 0,
		}
		writeJSON(w, http.StatusOK, response)
		return
	}

	items := []TaskItem{}
	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}

		var act core.ActivityEntity
		if err := fileStore.ReadYaml(ctx, "activities/"+file, &act); err != nil {
			continue
		}

		// Simple モードの Activity のみ TaskItem に変換
		if len(act.Nodes) == 0 {
			items = append(items, TaskItem{
				ID:           act.ID,
				Title:        act.Title,
				Status:       string(act.Status),
				Priority:     string(act.Priority),
				Assignee:     act.Assignee,
				Dependencies: act.Dependencies,
				ParentID:     act.ParentID,
				StartDate:    act.StartDate,
				DueDate:      act.DueDate,
				Progress:     act.Progress,
				WBSCode:      act.WBSCode,
			})
		}
	}

	response := TasksResponse{
		Tasks: items,
		Total: len(items),
	}

	writeJSON(w, http.StatusOK, response)
}
