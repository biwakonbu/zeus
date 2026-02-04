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

// TestHandleAPIWBS は WBS API の正常系テスト
func TestHandleAPIWBS(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/wbs")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result WBSResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON のデコードに失敗: %v", err)
	}

	// Roots は nil ではなく空配列
	if result.Roots == nil {
		t.Error("Roots が nil です")
	}
}

// TestHandleAPIWBS_WithActivities は Activity がある場合の WBS API テスト
func TestHandleAPIWBS_WithActivities(t *testing.T) {
	zeus := setupTestZeusWithMultipleActivities(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/wbs")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result WBSResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON のデコードに失敗: %v", err)
	}

	// Stats.TotalNodes が 3 以上（Activity 3 件を追加したため）
	if result.Stats.TotalNodes < 3 {
		t.Errorf("TotalNodes が正しくありません: got %d, want >= 3", result.Stats.TotalNodes)
	}
}

// TestHandleAPIWBS_MethodNotAllowed は POST 時の 405 エラーテスト
func TestHandleAPIWBS_MethodNotAllowed(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/wbs", "application/json", nil)
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

// TestHandleAPITimeline はタイムライン API の正常系テスト
func TestHandleAPITimeline(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/timeline")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result TimelineResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON のデコードに失敗: %v", err)
	}

	// Items は nil ではなく空配列
	if result.Items == nil {
		t.Error("Items が nil です")
	}

	// CriticalPath は nil ではなく空配列
	if result.CriticalPath == nil {
		t.Error("CriticalPath が nil です")
	}
}

// TestHandleAPITimeline_WithTasks はタスクがある場合のタイムライン API テスト
func TestHandleAPITimeline_WithTasks(t *testing.T) {
	zeus := setupTestZeusWithMultipleActivities(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/timeline")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result TimelineResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON のデコードに失敗: %v", err)
	}

	// Stats.TotalTasks が 3 以上
	if result.Stats.TotalTasks < 3 {
		t.Errorf("TotalTasks が正しくありません: got %d, want >= 3", result.Stats.TotalTasks)
	}
}

// TestHandleAPITimeline_MethodNotAllowed は POST 時の 405 エラーテスト
func TestHandleAPITimeline_MethodNotAllowed(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/timeline", "application/json", nil)
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

// TestHandleAPIDownstream は Downstream API の正常系テスト
func TestHandleAPIDownstream(t *testing.T) {
	zeus, activityID := setupTestZeusWithActivity(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/downstream?task_id=" + activityID)
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result DownstreamResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON のデコードに失敗: %v", err)
	}

	if result.TaskID != activityID {
		t.Errorf("TaskID が正しくありません: got %s, want %s", result.TaskID, activityID)
	}

	// Downstream は nil ではなく空配列
	if result.Downstream == nil {
		t.Error("Downstream が nil です")
	}

	// Upstream は nil ではなく空配列
	if result.Upstream == nil {
		t.Error("Upstream が nil です")
	}
}

// TestHandleAPIDownstream_MissingParam は task_id なしの 400 エラーテスト
func TestHandleAPIDownstream_MissingParam(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/downstream")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	// エラーレスポンスを検証
	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("JSON のデコードに失敗: %v", err)
	}

	if errResp.Message == "" {
		t.Error("エラーメッセージが空です")
	}
}

// TestHandleAPIDownstream_NotFound は存在しないタスクの 404 エラーテスト
func TestHandleAPIDownstream_NotFound(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/downstream?task_id=nonexistent-task")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusNotFound)
	}

	// エラーレスポンスを検証
	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("JSON のデコードに失敗: %v", err)
	}

	if errResp.Message == "" {
		t.Error("エラーメッセージが空です")
	}
}

// TestHandleAPIDownstream_MethodNotAllowed は POST 時の 405 エラーテスト
func TestHandleAPIDownstream_MethodNotAllowed(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/downstream", "application/json", nil)
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
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

// Note: TestHandleAPITasks_MethodNotAllowed は /api/tasks 廃止に伴い削除

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

// TestHandleAPIPredict_MethodNotAllowed は POST 時の 405 エラーテスト
func TestHandleAPIPredict_MethodNotAllowed(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/predict", "application/json", nil)
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

	// Activity 数が 3 以上（TotalTasks フィールドは Activity 数を表す）
	if result.State.Summary.TotalTasks < 3 {
		t.Errorf("TotalTasks が正しくありません: got %d, want >= 3", result.State.Summary.TotalTasks)
	}
}

// Note: TestHandleAPITasks_WithTasks と TestHandleAPITasks_DependenciesAlwaysArray は
// /api/tasks 廃止に伴い削除。Activity API (/api/activities) を使用してください。

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

// TestHandleAPIPredict_WithTasks はタスクがある場合の予測 API テスト
func TestHandleAPIPredict_WithTasks(t *testing.T) {
	zeus := setupTestZeusWithMultipleActivities(t)
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
