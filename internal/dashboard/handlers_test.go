package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/biwakonbu/zeus/internal/core"
)

// setupTestZeusWithActivity はテスト用の Zeus を作成し、Activity を追加する
func setupTestZeusWithActivity(t *testing.T) (*core.Zeus, string) {
	t.Helper()

	tmpDir := t.TempDir()
	zeus := core.New(tmpDir)
	ctx := context.Background()

	_, err := zeus.Init(ctx)
	if err != nil {
		t.Fatalf("Zeus の初期化に失敗: %v", err)
	}

	// テスト用 Activity を追加
	result, err := zeus.Add(ctx, "activity", "Test Activity")
	if err != nil {
		t.Fatalf("Activity の追加に失敗: %v", err)
	}

	return zeus, result.ID
}

// setupTestZeusWithMultipleActivities はテスト用の Zeus を作成し、複数 Activity を追加する
func setupTestZeusWithMultipleActivities(t *testing.T) *core.Zeus {
	t.Helper()

	tmpDir := t.TempDir()
	zeus := core.New(tmpDir)
	ctx := context.Background()

	_, err := zeus.Init(ctx)
	if err != nil {
		t.Fatalf("Zeus の初期化に失敗: %v", err)
	}

	// 複数 Activity を追加（エラーをチェック）
	if _, err := zeus.Add(ctx, "activity", "Activity 1"); err != nil {
		t.Fatalf("Activity 1 の追加に失敗: %v", err)
	}
	if _, err := zeus.Add(ctx, "activity", "Activity 2"); err != nil {
		t.Fatalf("Activity 2 の追加に失敗: %v", err)
	}
	if _, err := zeus.Add(ctx, "activity", "Activity 3"); err != nil {
		t.Fatalf("Activity 3 の追加に失敗: %v", err)
	}

	return zeus
}

// TestHandleAPIStatus_MethodNotAllowed は POST 時の 405 エラーテスト
func TestHandleAPIStatus_MethodNotAllowed(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/status", "application/json", nil)
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

// TestHandleAPIGraph_MethodNotAllowed は POST 時の 405 エラーテスト
func TestHandleAPIGraph_MethodNotAllowed(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/graph", "application/json", nil)
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

// TestHandleAPIStatus_WithActivities は Activity がある場合のステータス API テスト
func TestHandleAPIStatus_WithActivities(t *testing.T) {
	zeus := setupTestZeusWithMultipleActivities(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/status")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON のデコードに失敗: %v", err)
	}

	// Activity 数が 3 以上（TotalActivities フィールドは Activity 数を表す）
	if result.State.Summary.TotalActivities < 3 {
		t.Errorf("TotalActivities が正しくありません: got %d, want >= 3", result.State.Summary.TotalActivities)
	}
}

// TestHandleAPIGraph_WithActivities は Activity がある場合のグラフ API テスト
func TestHandleAPIGraph_WithTasks(t *testing.T) {
	zeus := setupTestZeusWithMultipleActivities(t)
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

	// ノード数が 3 以上
	if result.Stats.TotalNodes < 3 {
		t.Errorf("TotalNodes が正しくありません: got %d, want >= 3", result.Stats.TotalNodes)
	}
}

// TestServerDevMode は開発モードのテスト
func TestServerDevMode(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServerWithDevMode(zeus, 0, true)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/status")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// DevMode が設定されていることを確認（devMode フィールドは private なので間接的に確認）
	// CORS ヘッダーの存在で開発モードを確認
	if origin := resp.Header.Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("開発モードの CORS ヘッダーが設定されていません: got %q, want *", origin)
	}
}

// TestSSEHeaders は SSE エンドポイントのヘッダーテスト
func TestSSEHeaders(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	// SSE 接続を開始（すぐにクローズ）
	resp, err := http.Get(ts.URL + "/api/events")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Content-Type が text/event-stream
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("Content-Type が正しくありません: got %s, want text/event-stream", contentType)
	}

	// Cache-Control が no-cache
	cacheControl := resp.Header.Get("Cache-Control")
	if cacheControl != "no-cache" {
		t.Errorf("Cache-Control が正しくありません: got %s, want no-cache", cacheControl)
	}
}

// TestSSEDevModeHeaders は開発モードでの SSE ヘッダーテスト
func TestSSEDevModeHeaders(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServerWithDevMode(zeus, 0, true)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	// SSE 接続を開始（すぐにクローズ）
	resp, err := http.Get(ts.URL + "/api/events")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// 開発モードでは CORS ヘッダーが設定される
	corsHeader := resp.Header.Get("Access-Control-Allow-Origin")
	if corsHeader != "*" {
		t.Errorf("Access-Control-Allow-Origin が正しくありません: got %s, want *", corsHeader)
	}
}
