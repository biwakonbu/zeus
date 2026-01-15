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

//go:embed static/*
var staticFiles embed.FS

// Server はダッシュボード HTTP サーバー
type Server struct {
	zeus   *core.Zeus
	server *http.Server
	port   int
}

// NewServer は新しい Server を作成
func NewServer(zeus *core.Zeus, port int) *Server {
	return &Server{
		zeus: zeus,
		port: port,
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

// handler は http.Handler を構築
func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()

	// 静的ファイルを提供
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		// embed エラーは起動時に発生するべき
		panic(fmt.Sprintf("静的ファイルの読み込みに失敗: %v", err))
	}

	// API エンドポイント
	mux.HandleFunc("/api/status", s.handleAPIStatus)
	mux.HandleFunc("/api/tasks", s.handleAPITasks)
	mux.HandleFunc("/api/graph", s.handleAPIGraph)
	mux.HandleFunc("/api/predict", s.handleAPIPredict)

	// 静的ファイル（index.html, styles.css, app.js）
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// ルートパスは index.html にリダイレクト
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		// index.html を直接提供
		data, err := staticFiles.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	return mux
}
