package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/biwakonbu/zeus/internal/core"
)

// =============================================================================
// Vision API テスト
// =============================================================================

// TestHandleAPIVision_NotFound は vision.yaml が存在しない場合のテスト
func TestHandleAPIVision_NotFound(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/vision")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result VisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON デコードに失敗: %v", err)
	}

	if result.Vision != nil {
		t.Errorf("Vision が nil であるべきです: got %+v", result.Vision)
	}
}

// TestHandleAPIVision は Vision が存在する場合のテスト
func TestHandleAPIVision(t *testing.T) {
	zeus := setupTestZeus(t)
	ctx := context.Background()

	// Vision を追加
	handler, ok := zeus.GetRegistry().Get("vision")
	if !ok {
		t.Fatal("vision ハンドラーが見つかりません")
	}
	vh := handler.(*core.VisionHandler)

	_, err := vh.Add(ctx, "テストビジョン",
		core.WithVisionStatement("テストステートメント"),
		core.WithVisionSuccessCriteria([]string{"基準1", "基準2"}),
	)
	if err != nil {
		t.Fatalf("Vision 追加に失敗: %v", err)
	}

	server := NewServer(zeus, 0)
	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/vision")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result VisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON デコードに失敗: %v", err)
	}

	if result.Vision == nil {
		t.Fatal("Vision が nil です")
	}

	if result.Vision.Title != "テストビジョン" {
		t.Errorf("Vision Title が正しくありません: got %s, want テストビジョン", result.Vision.Title)
	}

	if result.Vision.Statement != "テストステートメント" {
		t.Errorf("Vision Statement が正しくありません: got %s, want テストステートメント", result.Vision.Statement)
	}

	if len(result.Vision.SuccessCriteria) != 2 {
		t.Errorf("SuccessCriteria の数が正しくありません: got %d, want 2", len(result.Vision.SuccessCriteria))
	}

	if result.Vision.Status == "" {
		t.Error("Vision Status が空です")
	}

	if result.Vision.ID == "" {
		t.Error("Vision ID が空です")
	}
}

// TestHandleAPIVision_MethodNotAllowed は POST 時の 405 エラーテスト
func TestHandleAPIVision_MethodNotAllowed(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/vision", "application/json", nil)
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

// =============================================================================
// Objectives API テスト
// =============================================================================

// TestHandleAPIObjectives_Empty は Objective がない場合のテスト
func TestHandleAPIObjectives_Empty(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/objectives")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result ObjectivesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON デコードに失敗: %v", err)
	}

	if result.Total != 0 {
		t.Errorf("Total が正しくありません: got %d, want 0", result.Total)
	}

	if result.Objectives == nil {
		t.Error("Objectives 配列が nil です")
	}
}

// TestHandleAPIObjectives は Objective が存在する場合のテスト
func TestHandleAPIObjectives(t *testing.T) {
	zeus := setupTestZeus(t)
	ctx := context.Background()

	// Objective を追加
	handler, ok := zeus.GetRegistry().Get("objective")
	if !ok {
		t.Fatal("objective ハンドラーが見つかりません")
	}
	oh := handler.(*core.ObjectiveHandler)

	obj, err := oh.Add(ctx, "テスト目標",
		core.WithObjectiveDescription("テスト用の目標です"),
		core.WithObjectiveGoals([]string{"ゴール1", "ゴール2"}),
		core.WithObjectiveOwner("テストオーナー"),
		core.WithObjectiveTags([]string{"tag1", "tag2"}),
	)
	if err != nil {
		t.Fatalf("Objective 追加に失敗: %v", err)
	}

	server := NewServer(zeus, 0)
	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/objectives")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result ObjectivesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON デコードに失敗: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Total が正しくありません: got %d, want 1", result.Total)
	}

	if len(result.Objectives) != 1 {
		t.Fatalf("Objectives の数が正しくありません: got %d, want 1", len(result.Objectives))
	}

	item := result.Objectives[0]

	if item.ID != obj.ID {
		t.Errorf("Objective ID が正しくありません: got %s, want %s", item.ID, obj.ID)
	}

	if item.Title != "テスト目標" {
		t.Errorf("Objective Title が正しくありません: got %s, want テスト目標", item.Title)
	}

	if item.Description != "テスト用の目標です" {
		t.Errorf("Objective Description が正しくありません: got %s, want テスト用の目標です", item.Description)
	}

	if len(item.Goals) != 2 {
		t.Errorf("Goals の数が正しくありません: got %d, want 2", len(item.Goals))
	}

	if item.Owner != "テストオーナー" {
		t.Errorf("Owner が正しくありません: got %s, want テストオーナー", item.Owner)
	}

	if len(item.Tags) != 2 {
		t.Errorf("Tags の数が正しくありません: got %d, want 2", len(item.Tags))
	}

	if item.UseCaseCount != 0 {
		t.Errorf("UseCaseCount が正しくありません: got %d, want 0", item.UseCaseCount)
	}
}

// TestHandleAPIObjectives_WithUseCaseCount は UseCase カウントのテスト
func TestHandleAPIObjectives_WithUseCaseCount(t *testing.T) {
	zeus := setupTestZeus(t)
	ctx := context.Background()

	// Objective を追加
	objHandler, ok := zeus.GetRegistry().Get("objective")
	if !ok {
		t.Fatal("objective ハンドラーが見つかりません")
	}
	oh := objHandler.(*core.ObjectiveHandler)

	obj, err := oh.Add(ctx, "テスト目標")
	if err != nil {
		t.Fatalf("Objective 追加に失敗: %v", err)
	}

	// UseCase を2つ追加
	ucHandler, ok := zeus.GetRegistry().Get("usecase")
	if !ok {
		t.Fatal("usecase ハンドラーが見つかりません")
	}
	uh := ucHandler.(*core.UseCaseHandler)

	_, err = uh.Add(ctx, "ユースケース1", core.WithUseCaseObjective(obj.ID))
	if err != nil {
		t.Fatalf("UseCase 1 追加に失敗: %v", err)
	}

	_, err = uh.Add(ctx, "ユースケース2", core.WithUseCaseObjective(obj.ID))
	if err != nil {
		t.Fatalf("UseCase 2 追加に失敗: %v", err)
	}

	server := NewServer(zeus, 0)
	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/objectives")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result ObjectivesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON デコードに失敗: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Total が正しくありません: got %d, want 1", result.Total)
	}

	if len(result.Objectives) != 1 {
		t.Fatalf("Objectives の数が正しくありません: got %d, want 1", len(result.Objectives))
	}

	if result.Objectives[0].UseCaseCount != 2 {
		t.Errorf("UseCaseCount が正しくありません: got %d, want 2", result.Objectives[0].UseCaseCount)
	}
}

// TestHandleAPIObjectives_MethodNotAllowed は POST 時の 405 エラーテスト
func TestHandleAPIObjectives_MethodNotAllowed(t *testing.T) {
	zeus := setupTestZeus(t)
	server := NewServer(zeus, 0)

	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/objectives", "application/json", nil)
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("ステータスコードが正しくありません: got %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

// TestHandleAPIObjectives_MinimalFields は最小限のフィールドで作成した場合のテスト
func TestHandleAPIObjectives_MinimalFields(t *testing.T) {
	zeus := setupTestZeus(t)
	ctx := context.Background()

	// Goals/Tags を指定せずに Objective を追加
	handler, ok := zeus.GetRegistry().Get("objective")
	if !ok {
		t.Fatal("objective ハンドラーが見つかりません")
	}
	oh := handler.(*core.ObjectiveHandler)

	_, err := oh.Add(ctx, "最小限の目標")
	if err != nil {
		t.Fatalf("Objective 追加に失敗: %v", err)
	}

	server := NewServer(zeus, 0)
	ts := httptest.NewServer(server.handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/objectives")
	if err != nil {
		t.Fatalf("リクエストに失敗: %v", err)
	}
	defer resp.Body.Close()

	var result ObjectivesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("JSON デコードに失敗: %v", err)
	}

	if len(result.Objectives) != 1 {
		t.Fatalf("Objectives の数が正しくありません: got %d, want 1", len(result.Objectives))
	}

	item := result.Objectives[0]

	if item.Title != "最小限の目標" {
		t.Errorf("Title が正しくありません: got %s, want 最小限の目標", item.Title)
	}

	if item.Status == "" {
		t.Error("Status が空です")
	}

	if item.CreatedAt == "" {
		t.Error("CreatedAt が空です")
	}
}
