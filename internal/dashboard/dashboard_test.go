package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/biwakonbu/zeus/internal/core"
)

// テスト用の一時ディレクトリを作成
func setupTestZeus(t *testing.T) *core.Zeus {
	t.Helper()

	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// Zeus を初期化
	zeus := core.New(tmpDir)
	ctx := context.Background()

	// プロジェクトを初期化
	_, err := zeus.Init(ctx, "simple")
	if err != nil {
		t.Fatalf("Zeus の初期化に失敗: %v", err)
	}

	return zeus
}

func TestNewServer(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 8080)

	if server == nil {
		t.Fatal("NewServer が nil を返しました")
	}

	if server.Port() != 8080 {
		t.Errorf("ポートが正しくありません: got %d, want 8080", server.Port())
	}

	if server.URL() != "http://127.0.0.1:8080" {
		t.Errorf("URL が正しくありません: got %s, want http://127.0.0.1:8080", server.URL())
	}
}

func TestServerStartShutdown(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0) // ポート 0 で動的割り当て

	ctx := context.Background()

	// 起動
	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("サーバーの起動に失敗: %v", err)
	}

	// 少し待機
	time.Sleep(100 * time.Millisecond)

	// 停止
	err = server.Shutdown(ctx)
	if err != nil {
		t.Fatalf("サーバーの停止に失敗: %v", err)
	}
}

func TestHandleAPIStatus(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	// テストサーバーを作成
	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	// リクエスト送信
	resp, err := http.Get(ts.URL + "/api/status")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	// ステータスコード確認
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// レスポンス確認
	var result StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON のデコードに失敗: %v", err)
	}

	if result.Project.Name == "" {
		t.Error("プロジェクト名が空です")
	}
}

func TestHandleAPITasks(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/tasks")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result TasksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON のデコードに失敗: %v", err)
	}

	// 初期状態ではタスクは空
	if result.Tasks == nil {
		t.Error("タスク配列が nil です")
	}
}

func TestHandleAPIGraph(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/graph")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result GraphResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON のデコードに失敗: %v", err)
	}

	// Mermaid 形式の文字列が存在
	if result.Mermaid == "" {
		t.Error("Mermaid が空です")
	}
}

func TestHandleAPIPredict(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/predict")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result PredictResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON のデコードに失敗: %v", err)
	}

	// 予測結果が存在
	if result.Completion == nil && result.Risk == nil && result.Velocity == nil {
		t.Error("予測結果が全て nil です")
	}
}

func TestHandleIndex(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type が正しくありません: got %s", contentType)
	}
}

func TestHandleMethodNotAllowed(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	// POST リクエストを送信
	resp, err := http.Post(ts.URL+"/api/status", "application/json", nil)
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestHandle404(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/nonexistent")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}
