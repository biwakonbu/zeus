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
	_, err := zeus.Init(ctx)
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

// Note: TestServerBroadcastAllUpdates_TaskDependenciesAlwaysArray は
// Task 非推奨に伴い削除。Activity API を使用してください。

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

// Note: TestHandleAPITasks は /api/tasks 非推奨に伴い削除。
// Activity API (/api/activities) を使用してください。

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

// Note: TestHandleAPIPredict は predict 機能削除に伴い削除

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

// TestHandleSPAFallback は SPA フォールバックルーティングをテストします
// SvelteKit SPA モードでは、不明なルートは index.html を返し、
// フロントエンドルーターが 404 ページを処理します
func TestHandleSPAFallback(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/nonexistent")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	// SPA フォールバック: 不明なルートは index.html を返す
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type が正しくありません: got %s, want text/html; charset=utf-8", contentType)
	}
}

// TestHandleAPIActors は /api/actors エンドポイントをテストします
func TestHandleAPIActors(t *testing.T) {
	zeus := setupTestZeus(t)
	ctx := context.Background()

	// Actor を追加
	handler, ok := zeus.GetRegistry().Get("actor")
	if !ok {
		t.Fatal("actor ハンドラーが見つかりません")
	}
	actorHandler := handler.(*core.ActorHandler)

	_, err := actorHandler.Add(ctx, "テストアクター",
		core.WithActorType(core.ActorTypeHuman),
		core.WithActorDescription("テスト用アクター"),
	)
	if err != nil {
		t.Fatalf("Actor 追加に失敗: %v", err)
	}

	server := NewServer(zeus, 0)
	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/actors")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result struct {
		Actors []struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"actors"`
		Total int `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON デコードに失敗: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Total が正しくありません: got %d, want 1", result.Total)
	}

	if len(result.Actors) != 1 {
		t.Errorf("Actors の数が正しくありません: got %d, want 1", len(result.Actors))
	}

	if result.Actors[0].Title != "テストアクター" {
		t.Errorf("Actor Title が正しくありません: got %s, want テストアクター", result.Actors[0].Title)
	}
}

// TestHandleAPIUseCases は /api/usecases エンドポイントをテストします
func TestHandleAPIUseCases(t *testing.T) {
	zeus := setupTestZeus(t)
	ctx := context.Background()

	// Objective を追加（UseCase の参照先として必要）
	objHandler, ok := zeus.GetRegistry().Get("objective")
	if !ok {
		t.Fatal("objective ハンドラーが見つかりません")
	}
	oh := objHandler.(*core.ObjectiveHandler)
	obj, err := oh.Add(ctx, "テスト目標")
	if err != nil {
		t.Fatalf("Objective 追加に失敗: %v", err)
	}

	// UseCase を追加
	handler, ok := zeus.GetRegistry().Get("usecase")
	if !ok {
		t.Fatal("usecase ハンドラーが見つかりません")
	}
	usecaseHandler := handler.(*core.UseCaseHandler)

	_, err = usecaseHandler.Add(ctx, "テストユースケース",
		core.WithUseCaseObjective(obj.ID),
		core.WithUseCaseDescription("テスト用ユースケース"),
	)
	if err != nil {
		t.Fatalf("UseCase 追加に失敗: %v", err)
	}

	server := NewServer(zeus, 0)
	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/usecases")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result struct {
		Usecases []struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			Status      string `json:"status"`
			Description string `json:"description"`
		} `json:"usecases"`
		Total int `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON デコードに失敗: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Total が正しくありません: got %d, want 1", result.Total)
	}

	if len(result.Usecases) != 1 {
		t.Errorf("Usecases の数が正しくありません: got %d, want 1", len(result.Usecases))
	}

	if result.Usecases[0].Title != "テストユースケース" {
		t.Errorf("UseCase Title が正しくありません: got %s, want テストユースケース", result.Usecases[0].Title)
	}
}

// TestHandleAPIUseCaseDiagram は /api/uml/usecase エンドポイントをテストします
func TestHandleAPIUseCaseDiagram(t *testing.T) {
	zeus := setupTestZeus(t)
	ctx := context.Background()

	// Objective を追加（UseCase の参照先として必要）
	objHandler, ok := zeus.GetRegistry().Get("objective")
	if !ok {
		t.Fatal("objective ハンドラーが見つかりません")
	}
	oh := objHandler.(*core.ObjectiveHandler)
	obj, err := oh.Add(ctx, "テスト目標")
	if err != nil {
		t.Fatalf("Objective 追加に失敗: %v", err)
	}

	// Actor を追加
	actorHandler, ok := zeus.GetRegistry().Get("actor")
	if !ok {
		t.Fatal("actor ハンドラーが見つかりません")
	}
	ah := actorHandler.(*core.ActorHandler)

	actor, err := ah.Add(ctx, "管理者",
		core.WithActorType(core.ActorTypeHuman),
	)
	if err != nil {
		t.Fatalf("Actor 追加に失敗: %v", err)
	}

	// UseCase を追加
	usecaseHandler, ok := zeus.GetRegistry().Get("usecase")
	if !ok {
		t.Fatal("usecase ハンドラーが見つかりません")
	}
	uh := usecaseHandler.(*core.UseCaseHandler)

	_, err = uh.Add(ctx, "ユーザー登録",
		core.WithUseCaseObjective(obj.ID),
		core.WithUseCaseActor(actor.ID, core.ActorRolePrimary),
	)
	if err != nil {
		t.Fatalf("UseCase 追加に失敗: %v", err)
	}

	server := NewServer(zeus, 0)
	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/uml/usecase")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result struct {
		Actors   []interface{} `json:"actors"`
		Usecases []interface{} `json:"usecases"`
		Boundary string        `json:"boundary"`
		Mermaid  string        `json:"mermaid"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON デコードに失敗: %v", err)
	}

	if len(result.Actors) != 1 {
		t.Errorf("Actors の数が正しくありません: got %d, want 1", len(result.Actors))
	}

	if len(result.Usecases) != 1 {
		t.Errorf("Usecases の数が正しくありません: got %d, want 1", len(result.Usecases))
	}

	if result.Mermaid == "" {
		t.Error("Mermaid が空です")
	}
}

// TestHandleAPIActorsEmpty は Actor がない場合の /api/actors をテストします
func TestHandleAPIActorsEmpty(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/actors")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result struct {
		Actors []interface{} `json:"actors"`
		Total  int           `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON デコードに失敗: %v", err)
	}

	if result.Total != 0 {
		t.Errorf("Total が正しくありません: got %d, want 0", result.Total)
	}
}

// TestHandleAPIUseCasesEmpty は UseCase がない場合の /api/usecases をテストします
func TestHandleAPIUseCasesEmpty(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/usecases")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result struct {
		Usecases []interface{} `json:"usecases"`
		Total    int           `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON デコードに失敗: %v", err)
	}

	if result.Total != 0 {
		t.Errorf("Total が正しくありません: got %d, want 0", result.Total)
	}
}

// TestHandleAPIActivities は /api/activities エンドポイントをテストします
func TestHandleAPIActivities(t *testing.T) {
	zeus := setupTestZeus(t)
	ctx := context.Background()

	// Activity を追加
	handler, ok := zeus.GetRegistry().Get("activity")
	if !ok {
		t.Fatal("activity ハンドラーが見つかりません")
	}
	activityHandler := handler.(*core.ActivityHandler)

	_, err := activityHandler.Add(ctx, "テストアクティビティ",
		core.WithActivityDescription("テスト用アクティビティ"),
		core.WithActivityStatus(core.ActivityStatusActive),
	)
	if err != nil {
		t.Fatalf("Activity 追加に失敗: %v", err)
	}

	server := NewServer(zeus, 0)
	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/activities")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result ActivitiesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON デコードに失敗: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Total が正しくありません: got %d, want 1", result.Total)
	}

	if len(result.Activities) != 1 {
		t.Errorf("Activities の数が正しくありません: got %d, want 1", len(result.Activities))
	}

	if result.Activities[0].Title != "テストアクティビティ" {
		t.Errorf("Activity Title が正しくありません: got %s, want テストアクティビティ", result.Activities[0].Title)
	}

	if result.Activities[0].Status != "active" {
		t.Errorf("Activity Status が正しくありません: got %s, want active", result.Activities[0].Status)
	}
}

// TestHandleAPIActivitiesEmpty は Activity がない場合の /api/activities をテストします
func TestHandleAPIActivitiesEmpty(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/activities")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result ActivitiesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON デコードに失敗: %v", err)
	}

	if result.Total != 0 {
		t.Errorf("Total が正しくありません: got %d, want 0", result.Total)
	}

	if result.Activities == nil {
		t.Error("Activities 配列が nil です")
	}
}

// TestHandleAPIActivityDiagram は /api/uml/activity エンドポイントをテストします
func TestHandleAPIActivityDiagram(t *testing.T) {
	zeus := setupTestZeus(t)
	ctx := context.Background()

	// Activity を追加
	handler, ok := zeus.GetRegistry().Get("activity")
	if !ok {
		t.Fatal("activity ハンドラーが見つかりません")
	}
	activityHandler := handler.(*core.ActivityHandler)

	// ノードと遷移を含むアクティビティを作成
	result, err := activityHandler.Add(ctx, "ログインフロー",
		core.WithActivityDescription("ユーザーログインのアクティビティ図"),
		core.WithActivityStatus(core.ActivityStatusActive),
		core.WithActivityNodes([]core.ActivityNode{
			{ID: "node-001", Type: core.ActivityNodeTypeInitial, Name: ""},
			{ID: "node-002", Type: core.ActivityNodeTypeAction, Name: "ログイン画面表示"},
			{ID: "node-003", Type: core.ActivityNodeTypeDecision, Name: "認証成功？"},
			{ID: "node-004", Type: core.ActivityNodeTypeAction, Name: "ダッシュボード表示"},
			{ID: "node-005", Type: core.ActivityNodeTypeFinal, Name: ""},
		}),
		core.WithActivityTransitions([]core.ActivityTransition{
			{ID: "trans-001", Source: "node-001", Target: "node-002"},
			{ID: "trans-002", Source: "node-002", Target: "node-003"},
			{ID: "trans-003", Source: "node-003", Target: "node-004", Guard: "[認証成功]"},
			{ID: "trans-004", Source: "node-004", Target: "node-005"},
		}),
	)
	if err != nil {
		t.Fatalf("Activity 追加に失敗: %v", err)
	}

	server := NewServer(zeus, 0)
	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/uml/activity?id=" + result.ID)
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var diagramResult ActivityDiagramResponse
	if err := json.NewDecoder(resp.Body).Decode(&diagramResult); err != nil {
		t.Fatalf("JSON デコードに失敗: %v", err)
	}

	if diagramResult.Activity == nil {
		t.Fatal("Activity が nil です")
	}

	if diagramResult.Activity.Title != "ログインフロー" {
		t.Errorf("Activity Title が正しくありません: got %s, want ログインフロー", diagramResult.Activity.Title)
	}

	if len(diagramResult.Activity.Nodes) != 5 {
		t.Errorf("Nodes の数が正しくありません: got %d, want 5", len(diagramResult.Activity.Nodes))
	}

	if len(diagramResult.Activity.Transitions) != 4 {
		t.Errorf("Transitions の数が正しくありません: got %d, want 4", len(diagramResult.Activity.Transitions))
	}

	if diagramResult.Mermaid == "" {
		t.Error("Mermaid が空です")
	}

	// Mermaid に flowchart が含まれているか確認
	if !containsString(diagramResult.Mermaid, "flowchart TD") {
		t.Error("Mermaid に flowchart TD が含まれていません")
	}
}

// TestHandleAPIActivityDiagram_NotFound は存在しない Activity の場合の /api/uml/activity をテストします
func TestHandleAPIActivityDiagram_NotFound(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/uml/activity?id=act-notexist")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

// TestHandleAPIActivityDiagram_MissingID は ID パラメータがない場合の /api/uml/activity をテストします
func TestHandleAPIActivityDiagram_MissingID(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/uml/activity")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

// containsString は文字列に部分文字列が含まれるかチェックするヘルパー
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
