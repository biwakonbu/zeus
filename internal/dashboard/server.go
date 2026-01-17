// Package dashboard は Zeus のWeb ダッシュボード機能を提供する。
// HTTP サーバーを起動し、プロジェクト状態をブラウザで可視化する。
package dashboard

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/biwakonbu/zeus/internal/core"
)

//go:embed build/*
var staticFiles embed.FS

// Server はダッシュボード HTTP サーバー
type Server struct {
	zeus        *core.Zeus
	server      *http.Server
	port        int
	devMode     bool
	broadcaster *SSEBroadcaster
}

// NewServer は新しい Server を作成
func NewServer(zeus *core.Zeus, port int) *Server {
	return &Server{
		zeus:        zeus,
		port:        port,
		devMode:     false,
		broadcaster: NewSSEBroadcaster(),
	}
}

// NewServerWithDevMode は開発モードで新しい Server を作成
func NewServerWithDevMode(zeus *core.Zeus, port int, devMode bool) *Server {
	return &Server{
		zeus:        zeus,
		port:        port,
		devMode:     devMode,
		broadcaster: NewSSEBroadcaster(),
	}
}

// Start はサーバーを起動
// 127.0.0.1 にバインドしてローカルアクセスのみ許可
func (s *Server) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	mux := s.handler()

	s.server = &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", s.port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// サーバーを goroutine で起動
	errChan := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	// 少し待ってエラーを確認
	select {
	case err := <-errChan:
		return err
	case <-time.After(100 * time.Millisecond):
		// 起動成功
		return nil
	}
}

// Shutdown はサーバーを停止
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	// グレースフルシャットダウン（5秒タイムアウト）
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.server.Shutdown(shutdownCtx)
}

// URL はサーバーの URL を返す
func (s *Server) URL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", s.port)
}

// Port はサーバーのポート番号を返す
func (s *Server) Port() int {
	return s.port
}

// DevMode は開発モードかどうかを返す
func (s *Server) DevMode() bool {
	return s.devMode
}

// corsMiddleware は CORS ヘッダーを追加するミドルウェア（開発モード用）
func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.devMode {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		next(w, r)
	}
}

// handler は http.Handler を構築
func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()

	// API エンドポイント（CORS 対応）
	mux.HandleFunc("/api/status", s.corsMiddleware(s.handleAPIStatus))
	mux.HandleFunc("/api/tasks", s.corsMiddleware(s.handleAPITasks))
	mux.HandleFunc("/api/graph", s.corsMiddleware(s.handleAPIGraph))
	mux.HandleFunc("/api/predict", s.corsMiddleware(s.handleAPIPredict))
	mux.HandleFunc("/api/wbs", s.corsMiddleware(s.handleAPIWBS))
	mux.HandleFunc("/api/timeline", s.corsMiddleware(s.handleAPITimeline))
	mux.HandleFunc("/api/downstream", s.corsMiddleware(s.handleAPIDownstream))
	mux.HandleFunc("/api/metrics", s.corsMiddleware(s.handleAPIMetrics))
	mux.HandleFunc("/api/events", s.handleSSE) // SSE エンドポイント

	// 静的ファイルを提供（本番モード）
	if !s.devMode {
		buildFS, err := fs.Sub(staticFiles, "build")
		if err != nil {
			// build ディレクトリが存在しない場合は従来の static を使用
			staticFS, err := fs.Sub(staticFiles, "static")
			if err != nil {
				panic(fmt.Sprintf("静的ファイルの読み込みに失敗: %v", err))
			}
			mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
			// ルートパスは index.html
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/" {
					http.NotFound(w, r)
					return
				}
				data, err := staticFiles.ReadFile("static/index.html")
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write(data)
			})
		} else {
			// SvelteKit ビルド成果物を配信
			fileServer := http.FileServer(http.FS(buildFS))
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				// まずファイルの存在を確認
				path := r.URL.Path
				if path == "/" {
					path = "/index.html"
				}

				// ファイルが存在するか確認
				_, err := fs.Stat(buildFS, path[1:]) // 先頭の / を除去
				if err != nil {
					// ファイルが存在しない場合は index.html を返す（SPA 対応）
					data, err := fs.ReadFile(buildFS, "index.html")
					if err != nil {
						http.NotFound(w, r)
						return
					}
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.Write(data)
					return
				}

				fileServer.ServeHTTP(w, r)
			})
		}
	}

	return mux
}

// BroadcastAllUpdates は全データの更新を SSE クライアントに通知
func (s *Server) BroadcastAllUpdates(ctx context.Context) {
	// ステータス
	if status, err := s.zeus.Status(ctx); err == nil {
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
		s.broadcaster.BroadcastStatus(response)
	}

	// タスク
	if result, err := s.zeus.List(ctx, "task"); err == nil {
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
		s.broadcaster.BroadcastTask(response)
	}

	// グラフ
	if graph, err := s.zeus.BuildDependencyGraph(ctx); err == nil {
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
		s.broadcaster.BroadcastGraph(response)
	}

	// 予測
	if result, err := s.zeus.Predict(ctx, "all"); err == nil {
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
		s.broadcaster.BroadcastPrediction(response)
	}
}

// Broadcaster は SSEBroadcaster を返す
func (s *Server) Broadcaster() *SSEBroadcaster {
	return s.broadcaster
}
