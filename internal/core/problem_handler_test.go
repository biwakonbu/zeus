package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// setupProblemHandlerTest は ProblemHandler テストのセットアップを行う
func setupProblemHandlerTest(t *testing.T) (*ProblemHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-problem-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath+"/problems", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create zeus dir: %v", err)
	}

	fs := yaml.NewFileManager(zeusPath)
	handler := NewProblemHandler(fs, nil, nil, nil)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return handler, zeusPath, cleanup
}

// setupProblemHandlerTestWithReferences は参照ハンドラー付きのセットアップを行う
func setupProblemHandlerTestWithReferences(t *testing.T) (*ProblemHandler, *ObjectiveHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-problem-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// 必要なディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	for _, dir := range []string{"problems", "objectives"} {
		if err := os.MkdirAll(filepath.Join(zeusPath, dir), 0755); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("failed to create %s dir: %v", dir, err)
		}
	}

	fs := yaml.NewFileManager(zeusPath)
	objHandler := NewObjectiveHandler(fs, nil)
	handler := NewProblemHandler(fs, objHandler, nil, nil)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return handler, objHandler, zeusPath, cleanup
}

// ===== Type() テスト =====

func TestProblemHandler_Type(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	if got := handler.Type(); got != "problem" {
		t.Errorf("Type() = %q, want %q", got, "problem")
	}
}

// ===== Add() テスト =====

func TestProblemHandler_Add(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "ビルドが遅い")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if !result.Success {
		t.Error("Add() result.Success = false, want true")
	}
	if result.ID != "prob-001" {
		t.Errorf("Add() result.ID = %q, want %q", result.ID, "prob-001")
	}
	if result.Entity != "problem" {
		t.Errorf("Add() result.Entity = %q, want %q", result.Entity, "problem")
	}

	// 作成されたエンティティを取得して確認
	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() after Add error = %v", err)
	}

	prob := entity.(*ProblemEntity)
	if prob.Title != "ビルドが遅い" {
		t.Errorf("Title = %q, want %q", prob.Title, "ビルドが遅い")
	}
	if prob.Status != ProblemStatusOpen {
		t.Errorf("Status = %q, want %q", prob.Status, ProblemStatusOpen)
	}
	if prob.Severity != ProblemSeverityMedium {
		t.Errorf("Severity = %q, want %q", prob.Severity, ProblemSeverityMedium)
	}
}

func TestProblemHandler_Add_WithAllOptions(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "本番環境の障害",
		WithProblemSeverity(ProblemSeverityCritical),
		WithProblemStatus(ProblemStatusInProgress),
		WithProblemDescription("本番環境でサービスがダウンしている"),
		WithProblemImpact("ユーザーがサービスを利用できない"),
		WithProblemRootCause("データベース接続の枯渇"),
		WithProblemPotentialSolutions([]string{"コネクションプールの拡大", "クエリの最適化"}),
		WithProblemReportedBy("監視システム"),
		WithProblemAssignedTo("インフラチーム"),
	)
	if err != nil {
		t.Fatalf("Add() with options error = %v", err)
	}

	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	prob := entity.(*ProblemEntity)
	if prob.Severity != ProblemSeverityCritical {
		t.Errorf("Severity = %q, want %q", prob.Severity, ProblemSeverityCritical)
	}
	if prob.Status != ProblemStatusInProgress {
		t.Errorf("Status = %q, want %q", prob.Status, ProblemStatusInProgress)
	}
	if prob.Description != "本番環境でサービスがダウンしている" {
		t.Errorf("Description = %q, want correct value", prob.Description)
	}
	if prob.Impact != "ユーザーがサービスを利用できない" {
		t.Errorf("Impact = %q, want correct value", prob.Impact)
	}
	if prob.RootCause != "データベース接続の枯渇" {
		t.Errorf("RootCause = %q, want correct value", prob.RootCause)
	}
	if len(prob.PotentialSolutions) != 2 {
		t.Errorf("PotentialSolutions length = %d, want 2", len(prob.PotentialSolutions))
	}
	if prob.ReportedBy != "監視システム" {
		t.Errorf("ReportedBy = %q, want %q", prob.ReportedBy, "監視システム")
	}
	if prob.AssignedTo != "インフラチーム" {
		t.Errorf("AssignedTo = %q, want %q", prob.AssignedTo, "インフラチーム")
	}
}

func TestProblemHandler_Add_InvalidInput(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 空の名前
	_, err := handler.Add(ctx, "")
	if err == nil {
		t.Error("Add() with empty name should return error")
	}
}

func TestProblemHandler_Add_WithObjectiveReference(t *testing.T) {
	handler, objHandler, _, cleanup := setupProblemHandlerTestWithReferences(t)
	defer cleanup()

	ctx := context.Background()

	// Objective を先に作成
	objResult, err := objHandler.Add(ctx, "プロジェクト目標")
	if err != nil {
		t.Fatalf("ObjectiveHandler.Add() error = %v", err)
	}

	result, err := handler.Add(ctx, "目標に関連する問題",
		WithProblemObjective(objResult.ID),
	)
	if err != nil {
		t.Fatalf("Add() with objective error = %v", err)
	}

	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	prob := entity.(*ProblemEntity)
	if prob.ObjectiveID != objResult.ID {
		t.Errorf("ObjectiveID = %q, want %q", prob.ObjectiveID, objResult.ID)
	}
}

func TestProblemHandler_Add_WithInvalidObjectiveReference(t *testing.T) {
	handler, _, _, cleanup := setupProblemHandlerTestWithReferences(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない Objective を参照
	_, err := handler.Add(ctx, "問題",
		WithProblemObjective("obj-999"),
	)
	if err == nil {
		t.Error("Add() with invalid objective reference should return error")
	}
}

// ===== List() テスト =====

func TestProblemHandler_List(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 複数の Problem を追加
	for i := 0; i < 3; i++ {
		_, err := handler.Add(ctx, "問題"+string(rune('A'+i)))
		if err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	}

	result, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if result.Total != 3 {
		t.Errorf("List() Total = %d, want 3", result.Total)
	}
}

func TestProblemHandler_List_Empty(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if result.Total != 0 {
		t.Errorf("List() Total = %d, want 0", result.Total)
	}
}

func TestProblemHandler_List_WithLimit(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 5つの Problem を追加
	for i := 0; i < 5; i++ {
		_, err := handler.Add(ctx, "問題"+string(rune('A'+i)))
		if err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	}

	result, err := handler.List(ctx, &ListFilter{Limit: 3})
	if err != nil {
		t.Fatalf("List() with limit error = %v", err)
	}

	if result.Total != 3 {
		t.Errorf("List() with limit Total = %d, want 3", result.Total)
	}
}

func TestProblemHandler_List_WithStatusFilter(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 異なるステータスの Problem を追加
	_, err := handler.Add(ctx, "オープンな問題", WithProblemStatus(ProblemStatusOpen))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	_, err = handler.Add(ctx, "解決済みの問題", WithProblemStatus(ProblemStatusResolved))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	_, err = handler.Add(ctx, "もう一つオープンな問題", WithProblemStatus(ProblemStatusOpen))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	result, err := handler.List(ctx, &ListFilter{Status: string(ProblemStatusOpen)})
	if err != nil {
		t.Fatalf("List() with status filter error = %v", err)
	}

	if result.Total != 2 {
		t.Errorf("List() with status filter Total = %d, want 2", result.Total)
	}
}

// ===== Get() テスト =====

func TestProblemHandler_Get(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "テスト問題")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	prob := entity.(*ProblemEntity)
	if prob.ID != result.ID {
		t.Errorf("Get() ID = %q, want %q", prob.ID, result.ID)
	}
	if prob.Title != "テスト問題" {
		t.Errorf("Get() Title = %q, want %q", prob.Title, "テスト問題")
	}
}

func TestProblemHandler_Get_NotFound(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	_, err := handler.Get(ctx, "prob-999")
	if err != ErrEntityNotFound {
		t.Errorf("Get() for non-existent ID error = %v, want ErrEntityNotFound", err)
	}
}

func TestProblemHandler_Get_InvalidID(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	_, err := handler.Get(ctx, "invalid-id")
	if err == nil {
		t.Error("Get() with invalid ID should return error")
	}
}

// ===== Update() テスト =====

func TestProblemHandler_Update(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "更新前の問題")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// 更新データを作成
	updated := &ProblemEntity{
		ID:       result.ID,
		Title:    "更新後の問題",
		Status:   ProblemStatusResolved,
		Severity: ProblemSeverityHigh,
	}

	err = handler.Update(ctx, result.ID, updated)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// 更新結果を確認
	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() after Update error = %v", err)
	}

	prob := entity.(*ProblemEntity)
	if prob.Title != "更新後の問題" {
		t.Errorf("Title after Update = %q, want %q", prob.Title, "更新後の問題")
	}
	if prob.Status != ProblemStatusResolved {
		t.Errorf("Status after Update = %q, want %q", prob.Status, ProblemStatusResolved)
	}
	if prob.Severity != ProblemSeverityHigh {
		t.Errorf("Severity after Update = %q, want %q", prob.Severity, ProblemSeverityHigh)
	}
}

func TestProblemHandler_Update_NotFound(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	updated := &ProblemEntity{
		ID:    "prob-999",
		Title: "存在しない問題",
	}

	err := handler.Update(ctx, "prob-999", updated)
	if err != ErrEntityNotFound {
		t.Errorf("Update() for non-existent ID error = %v, want ErrEntityNotFound", err)
	}
}

func TestProblemHandler_Update_InvalidType(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "問題")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// 間違った型で更新
	err = handler.Update(ctx, result.ID, "wrong type")
	if err == nil {
		t.Error("Update() with wrong type should return error")
	}
}

// ===== Delete() テスト =====

func TestProblemHandler_Delete(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "削除予定の問題")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	err = handler.Delete(ctx, result.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// 削除後に取得を試みる
	_, err = handler.Get(ctx, result.ID)
	if err != ErrEntityNotFound {
		t.Errorf("Get() after Delete error = %v, want ErrEntityNotFound", err)
	}
}

func TestProblemHandler_Delete_NotFound(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	err := handler.Delete(ctx, "prob-999")
	if err != ErrEntityNotFound {
		t.Errorf("Delete() for non-existent ID error = %v, want ErrEntityNotFound", err)
	}
}

// ===== GetProblemsBySeverity() テスト =====

func TestProblemHandler_GetProblemsBySeverity(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 異なる重大度の Problem を追加
	_, err := handler.Add(ctx, "クリティカルな問題 1", WithProblemSeverity(ProblemSeverityCritical))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	_, err = handler.Add(ctx, "軽微な問題", WithProblemSeverity(ProblemSeverityLow))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	_, err = handler.Add(ctx, "クリティカルな問題 2", WithProblemSeverity(ProblemSeverityCritical))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// クリティカルな問題を取得
	critical, err := handler.GetProblemsBySeverity(ctx, ProblemSeverityCritical)
	if err != nil {
		t.Fatalf("GetProblemsBySeverity() error = %v", err)
	}

	if len(critical) != 2 {
		t.Errorf("GetProblemsBySeverity(Critical) count = %d, want 2", len(critical))
	}

	// 軽微な問題を取得
	low, err := handler.GetProblemsBySeverity(ctx, ProblemSeverityLow)
	if err != nil {
		t.Fatalf("GetProblemsBySeverity() error = %v", err)
	}

	if len(low) != 1 {
		t.Errorf("GetProblemsBySeverity(Low) count = %d, want 1", len(low))
	}
}

func TestProblemHandler_GetProblemsBySeverity_Empty(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 問題を追加（High のみ）
	_, err := handler.Add(ctx, "高優先度の問題", WithProblemSeverity(ProblemSeverityHigh))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Critical を検索（該当なし）
	critical, err := handler.GetProblemsBySeverity(ctx, ProblemSeverityCritical)
	if err != nil {
		t.Fatalf("GetProblemsBySeverity() error = %v", err)
	}

	if len(critical) != 0 {
		t.Errorf("GetProblemsBySeverity(Critical) count = %d, want 0", len(critical))
	}
}

// ===== Context キャンセルテスト =====

func TestProblemHandler_ContextCancellation(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	// キャンセル済みコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Add
	_, err := handler.Add(ctx, "テスト")
	if err == nil {
		t.Error("Add() with cancelled context should return error")
	}

	// List
	_, err = handler.List(ctx, nil)
	if err == nil {
		t.Error("List() with cancelled context should return error")
	}

	// Get
	_, err = handler.Get(ctx, "prob-001")
	if err == nil {
		t.Error("Get() with cancelled context should return error")
	}

	// Update
	err = handler.Update(ctx, "prob-001", &ProblemEntity{})
	if err == nil {
		t.Error("Update() with cancelled context should return error")
	}

	// Delete
	err = handler.Delete(ctx, "prob-001")
	if err == nil {
		t.Error("Delete() with cancelled context should return error")
	}
}

// ===== ID 採番テスト =====

func TestProblemHandler_IDSequence(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 連続追加で ID が順番に採番されることを確認
	expectedIDs := []string{"prob-001", "prob-002", "prob-003"}

	for i, expected := range expectedIDs {
		result, err := handler.Add(ctx, "問題"+string(rune('A'+i)))
		if err != nil {
			t.Fatalf("Add() #%d error = %v", i+1, err)
		}
		if result.ID != expected {
			t.Errorf("Add() #%d ID = %q, want %q", i+1, result.ID, expected)
		}
	}
}

// ===== ファイルパステスト =====

func TestProblemHandler_FilePath(t *testing.T) {
	handler, zeusPath, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "テスト問題")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// ファイルが正しいパスに作成されたか確認
	expectedPath := filepath.Join(zeusPath, "problems", result.ID+".yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file at %q does not exist", expectedPath)
	}
}

// ===== メタデータテスト =====

func TestProblemHandler_Metadata(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "メタデータテスト")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	prob := entity.(*ProblemEntity)
	if prob.Metadata.CreatedAt == "" {
		t.Error("CreatedAt should not be empty")
	}
	if prob.Metadata.UpdatedAt == "" {
		t.Error("UpdatedAt should not be empty")
	}
}

// ===== 全重大度レベルテスト =====

func TestProblemHandler_AllSeverityLevels(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	severities := []ProblemSeverity{
		ProblemSeverityCritical,
		ProblemSeverityHigh,
		ProblemSeverityMedium,
		ProblemSeverityLow,
	}

	for _, severity := range severities {
		t.Run(string(severity), func(t *testing.T) {
			result, err := handler.Add(ctx, "問題 "+string(severity), WithProblemSeverity(severity))
			if err != nil {
				t.Fatalf("Add() error = %v", err)
			}

			entity, err := handler.Get(ctx, result.ID)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}

			prob := entity.(*ProblemEntity)
			if prob.Severity != severity {
				t.Errorf("Severity = %q, want %q", prob.Severity, severity)
			}
		})
	}
}

// ===== 全ステータステスト =====

func TestProblemHandler_AllStatusLevels(t *testing.T) {
	handler, _, cleanup := setupProblemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	statuses := []ProblemStatus{
		ProblemStatusOpen,
		ProblemStatusInProgress,
		ProblemStatusResolved,
		ProblemStatusWontFix,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			result, err := handler.Add(ctx, "問題 "+string(status), WithProblemStatus(status))
			if err != nil {
				t.Fatalf("Add() error = %v", err)
			}

			entity, err := handler.Get(ctx, result.ID)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}

			prob := entity.(*ProblemEntity)
			if prob.Status != status {
				t.Errorf("Status = %q, want %q", prob.Status, status)
			}
		})
	}
}
