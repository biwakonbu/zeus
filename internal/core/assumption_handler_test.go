package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// setupAssumptionHandlerTest は AssumptionHandler テストのセットアップを行う
func setupAssumptionHandlerTest(t *testing.T) (*AssumptionHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-assumption-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath+"/assumptions", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create zeus dir: %v", err)
	}

	fs := yaml.NewFileManager(zeusPath)
	handler := NewAssumptionHandler(fs, nil, nil, nil)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return handler, zeusPath, cleanup
}

// ===== Type() テスト =====

func TestAssumptionHandler_Type(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	if got := handler.Type(); got != "assumption" {
		t.Errorf("Type() = %q, want %q", got, "assumption")
	}
}

// ===== Add() テスト =====

func TestAssumptionHandler_Add(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "既存システムとの互換性")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if !result.Success {
		t.Error("Add() result.Success = false, want true")
	}
	if result.ID != "assum-001" {
		t.Errorf("Add() result.ID = %q, want %q", result.ID, "assum-001")
	}
	if result.Entity != "assumption" {
		t.Errorf("Add() result.Entity = %q, want %q", result.Entity, "assumption")
	}

	// 作成されたエンティティを取得して確認
	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() after Add error = %v", err)
	}

	assum := entity.(*AssumptionEntity)
	if assum.Title != "既存システムとの互換性" {
		t.Errorf("Title = %q, want %q", assum.Title, "既存システムとの互換性")
	}
	if assum.Status != AssumptionStatusAssumed {
		t.Errorf("Status = %q, want %q", assum.Status, AssumptionStatusAssumed)
	}
}

func TestAssumptionHandler_Add_WithAllOptions(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "顧客は既存ワークフローを維持",
		WithAssumptionStatus(AssumptionStatusValidated),
		WithAssumptionDescription("顧客のワークフロー変更が最小限であることを前提とする"),
		WithAssumptionIfInvalid("UI の大幅な見直しが必要"),
	)
	if err != nil {
		t.Fatalf("Add() with options error = %v", err)
	}

	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	assum := entity.(*AssumptionEntity)
	if assum.Status != AssumptionStatusValidated {
		t.Errorf("Status = %q, want %q", assum.Status, AssumptionStatusValidated)
	}
	if assum.Description != "顧客のワークフロー変更が最小限であることを前提とする" {
		t.Errorf("Description = %q, want correct value", assum.Description)
	}
	if assum.IfInvalid != "UI の大幅な見直しが必要" {
		t.Errorf("IfInvalid = %q, want correct value", assum.IfInvalid)
	}
}

func TestAssumptionHandler_Add_InvalidInput(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 空の名前
	_, err := handler.Add(ctx, "")
	if err == nil {
		t.Error("Add() with empty name should return error")
	}
}

// ===== List() テスト =====

func TestAssumptionHandler_List(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 複数の Assumption を追加
	for i := 0; i < 3; i++ {
		_, err := handler.Add(ctx, "前提条件"+string(rune('A'+i)))
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

func TestAssumptionHandler_List_Empty(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
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

func TestAssumptionHandler_List_WithLimit(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 5つの Assumption を追加
	for i := 0; i < 5; i++ {
		_, err := handler.Add(ctx, "前提条件"+string(rune('A'+i)))
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

func TestAssumptionHandler_List_WithStatusFilter(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 異なるステータスの Assumption を追加
	_, err := handler.Add(ctx, "前提 1", WithAssumptionStatus(AssumptionStatusAssumed))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	_, err = handler.Add(ctx, "前提 2", WithAssumptionStatus(AssumptionStatusValidated))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	_, err = handler.Add(ctx, "前提 3", WithAssumptionStatus(AssumptionStatusAssumed))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	result, err := handler.List(ctx, &ListFilter{Status: string(AssumptionStatusAssumed)})
	if err != nil {
		t.Fatalf("List() with status filter error = %v", err)
	}

	if result.Total != 2 {
		t.Errorf("List() with status filter Total = %d, want 2", result.Total)
	}
}

// ===== Get() テスト =====

func TestAssumptionHandler_Get(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "テスト前提条件")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	assum := entity.(*AssumptionEntity)
	if assum.ID != result.ID {
		t.Errorf("Get() ID = %q, want %q", assum.ID, result.ID)
	}
	if assum.Title != "テスト前提条件" {
		t.Errorf("Get() Title = %q, want %q", assum.Title, "テスト前提条件")
	}
}

func TestAssumptionHandler_Get_NotFound(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	_, err := handler.Get(ctx, "assum-999")
	if err != ErrEntityNotFound {
		t.Errorf("Get() for non-existent ID error = %v, want ErrEntityNotFound", err)
	}
}

func TestAssumptionHandler_Get_InvalidID(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	_, err := handler.Get(ctx, "invalid-id")
	if err == nil {
		t.Error("Get() with invalid ID should return error")
	}
}

// ===== Update() テスト =====

func TestAssumptionHandler_Update(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "更新前の前提条件")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// 更新データを作成
	updated := &AssumptionEntity{
		ID:          result.ID,
		Title:       "更新後の前提条件",
		Status:      AssumptionStatusInvalidated,
		Description: "前提が無効であることが判明",
		IfInvalid:   "代替策を検討",
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

	assum := entity.(*AssumptionEntity)
	if assum.Title != "更新後の前提条件" {
		t.Errorf("Title after Update = %q, want %q", assum.Title, "更新後の前提条件")
	}
	if assum.Status != AssumptionStatusInvalidated {
		t.Errorf("Status after Update = %q, want %q", assum.Status, AssumptionStatusInvalidated)
	}
	if assum.Description != "前提が無効であることが判明" {
		t.Errorf("Description after Update = %q, want correct value", assum.Description)
	}
}

func TestAssumptionHandler_Update_NotFound(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	updated := &AssumptionEntity{
		ID:    "assum-999",
		Title: "存在しない前提",
	}

	err := handler.Update(ctx, "assum-999", updated)
	if err != ErrEntityNotFound {
		t.Errorf("Update() for non-existent ID error = %v, want ErrEntityNotFound", err)
	}
}

func TestAssumptionHandler_Update_InvalidType(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "前提条件")
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

func TestAssumptionHandler_Delete(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "削除予定の前提条件")
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

func TestAssumptionHandler_Delete_NotFound(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	err := handler.Delete(ctx, "assum-999")
	if err != ErrEntityNotFound {
		t.Errorf("Delete() for non-existent ID error = %v, want ErrEntityNotFound", err)
	}
}

// ===== ValidateAssumption() テスト =====

func TestAssumptionHandler_ValidateAssumption(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "検証対象の前提条件")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// 検証を実行
	validation := AssumptionValidation{
		Method: "ユーザーインタビュー",
		Result: "前提が正しいことを確認",
	}

	err = handler.ValidateAssumption(ctx, result.ID, validation, AssumptionStatusValidated)
	if err != nil {
		t.Fatalf("ValidateAssumption() error = %v", err)
	}

	// 検証結果を確認
	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() after ValidateAssumption error = %v", err)
	}

	assum := entity.(*AssumptionEntity)
	if assum.Status != AssumptionStatusValidated {
		t.Errorf("Status after validation = %q, want %q", assum.Status, AssumptionStatusValidated)
	}
	if assum.Validation.Method != "ユーザーインタビュー" {
		t.Errorf("Validation.Method = %q, want correct value", assum.Validation.Method)
	}
	if assum.Validation.Result != "前提が正しいことを確認" {
		t.Errorf("Validation.Result = %q, want correct value", assum.Validation.Result)
	}
	if assum.Validation.ValidatedAt == "" {
		t.Error("Validation.ValidatedAt should not be empty")
	}
}

func TestAssumptionHandler_ValidateAssumption_Invalidate(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "無効化される前提条件")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// 無効化検証を実行
	validation := AssumptionValidation{
		Method: "実地調査",
		Result: "前提が誤りであることが判明",
	}

	err = handler.ValidateAssumption(ctx, result.ID, validation, AssumptionStatusInvalidated)
	if err != nil {
		t.Fatalf("ValidateAssumption() error = %v", err)
	}

	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() after ValidateAssumption error = %v", err)
	}

	assum := entity.(*AssumptionEntity)
	if assum.Status != AssumptionStatusInvalidated {
		t.Errorf("Status after invalidation = %q, want %q", assum.Status, AssumptionStatusInvalidated)
	}
}

func TestAssumptionHandler_ValidateAssumption_NotFound(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	validation := AssumptionValidation{
		Method: "テスト",
		Result: "結果",
	}

	err := handler.ValidateAssumption(ctx, "assum-999", validation, AssumptionStatusValidated)
	if err != ErrEntityNotFound {
		t.Errorf("ValidateAssumption() for non-existent ID error = %v, want ErrEntityNotFound", err)
	}
}

// ===== Context キャンセルテスト =====

func TestAssumptionHandler_ContextCancellation(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
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
	_, err = handler.Get(ctx, "assum-001")
	if err == nil {
		t.Error("Get() with cancelled context should return error")
	}

	// Update
	err = handler.Update(ctx, "assum-001", &AssumptionEntity{})
	if err == nil {
		t.Error("Update() with cancelled context should return error")
	}

	// Delete
	err = handler.Delete(ctx, "assum-001")
	if err == nil {
		t.Error("Delete() with cancelled context should return error")
	}
}

// ===== ID 採番テスト =====

func TestAssumptionHandler_IDSequence(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 連続追加で ID が順番に採番されることを確認
	expectedIDs := []string{"assum-001", "assum-002", "assum-003"}

	for i, expected := range expectedIDs {
		result, err := handler.Add(ctx, "前提条件"+string(rune('A'+i)))
		if err != nil {
			t.Fatalf("Add() #%d error = %v", i+1, err)
		}
		if result.ID != expected {
			t.Errorf("Add() #%d ID = %q, want %q", i+1, result.ID, expected)
		}
	}
}

// ===== ファイルパステスト =====

func TestAssumptionHandler_FilePath(t *testing.T) {
	handler, zeusPath, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "テスト前提条件")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// ファイルが正しいパスに作成されたか確認
	expectedPath := filepath.Join(zeusPath, "assumptions", result.ID+".yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file at %q does not exist", expectedPath)
	}
}

// ===== メタデータテスト =====

func TestAssumptionHandler_Metadata(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
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

	assum := entity.(*AssumptionEntity)
	if assum.Metadata.CreatedAt == "" {
		t.Error("CreatedAt should not be empty")
	}
	if assum.Metadata.UpdatedAt == "" {
		t.Error("UpdatedAt should not be empty")
	}
}

// ===== 全ステータステスト =====

func TestAssumptionHandler_AllStatusLevels(t *testing.T) {
	handler, _, cleanup := setupAssumptionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	statuses := []AssumptionStatus{
		AssumptionStatusAssumed,
		AssumptionStatusValidated,
		AssumptionStatusInvalidated,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			result, err := handler.Add(ctx, "前提 "+string(status), WithAssumptionStatus(status))
			if err != nil {
				t.Fatalf("Add() error = %v", err)
			}

			entity, err := handler.Get(ctx, result.ID)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}

			assum := entity.(*AssumptionEntity)
			if assum.Status != status {
				t.Errorf("Status = %q, want %q", assum.Status, status)
			}
		})
	}
}
