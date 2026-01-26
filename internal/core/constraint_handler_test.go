package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// setupConstraintHandlerTest は ConstraintHandler テストのセットアップを行う
func setupConstraintHandlerTest(t *testing.T) (*ConstraintHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-constraint-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create zeus dir: %v", err)
	}

	fs := yaml.NewFileManager(zeusPath)
	handler := NewConstraintHandler(fs)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return handler, zeusPath, cleanup
}

// ===== Type() テスト =====

func TestConstraintHandler_Type(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	if got := handler.Type(); got != "constraint" {
		t.Errorf("Type() = %q, want %q", got, "constraint")
	}
}

// ===== Add() テスト =====

func TestConstraintHandler_Add(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "レスポンスタイム 100ms 以内")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if !result.Success {
		t.Error("Add() result.Success = false, want true")
	}
	if result.ID != "const-001" {
		t.Errorf("Add() result.ID = %q, want %q", result.ID, "const-001")
	}
	if result.Entity != "constraint" {
		t.Errorf("Add() result.Entity = %q, want %q", result.Entity, "constraint")
	}

	// 作成されたエンティティを取得して確認
	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() after Add error = %v", err)
	}

	constraint := entity.(*ConstraintEntity)
	if constraint.Title != "レスポンスタイム 100ms 以内" {
		t.Errorf("Title = %q, want %q", constraint.Title, "レスポンスタイム 100ms 以内")
	}
	if constraint.Category != ConstraintCategoryTechnical {
		t.Errorf("Category = %q, want %q", constraint.Category, ConstraintCategoryTechnical)
	}
}

func TestConstraintHandler_Add_WithAllOptions(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "GDPR 準拠必須",
		WithConstraintCategory(ConstraintCategoryLegal),
		WithConstraintDescription("EU 一般データ保護規則に完全準拠すること"),
		WithConstraintSource("法務部"),
		WithConstraintImpact([]string{"データ保存方法の見直し", "ユーザー同意フローの実装"}),
		WithConstraintNonNegotiable(true),
	)
	if err != nil {
		t.Fatalf("Add() with options error = %v", err)
	}

	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	constraint := entity.(*ConstraintEntity)
	if constraint.Category != ConstraintCategoryLegal {
		t.Errorf("Category = %q, want %q", constraint.Category, ConstraintCategoryLegal)
	}
	if constraint.Description != "EU 一般データ保護規則に完全準拠すること" {
		t.Errorf("Description = %q, want correct value", constraint.Description)
	}
	if constraint.Source != "法務部" {
		t.Errorf("Source = %q, want %q", constraint.Source, "法務部")
	}
	if len(constraint.Impact) != 2 {
		t.Errorf("Impact count = %d, want 2", len(constraint.Impact))
	}
	if !constraint.NonNegotiable {
		t.Error("NonNegotiable = false, want true")
	}
}

func TestConstraintHandler_Add_InvalidInput(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 空の名前
	_, err := handler.Add(ctx, "")
	if err == nil {
		t.Error("Add() with empty name should return error")
	}
}

// ===== List() テスト =====

func TestConstraintHandler_List(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 複数の Constraint を追加
	for i := 0; i < 3; i++ {
		_, err := handler.Add(ctx, "制約条件"+string(rune('A'+i)))
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

func TestConstraintHandler_List_Empty(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
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

func TestConstraintHandler_List_WithLimit(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 5つの Constraint を追加
	for i := 0; i < 5; i++ {
		_, err := handler.Add(ctx, "制約条件"+string(rune('A'+i)))
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

// ===== Get() テスト =====

func TestConstraintHandler_Get(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "テスト制約条件")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	constraint := entity.(*ConstraintEntity)
	if constraint.ID != result.ID {
		t.Errorf("Get() ID = %q, want %q", constraint.ID, result.ID)
	}
	if constraint.Title != "テスト制約条件" {
		t.Errorf("Get() Title = %q, want %q", constraint.Title, "テスト制約条件")
	}
}

func TestConstraintHandler_Get_NotFound(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	_, err := handler.Get(ctx, "const-999")
	if err != ErrEntityNotFound {
		t.Errorf("Get() for non-existent ID error = %v, want ErrEntityNotFound", err)
	}
}

func TestConstraintHandler_Get_InvalidID(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	_, err := handler.Get(ctx, "invalid-id")
	if err == nil {
		t.Error("Get() with invalid ID should return error")
	}
}

// ===== Update() テスト =====

func TestConstraintHandler_Update(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "更新前の制約条件")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// 更新データを作成
	updated := &ConstraintEntity{
		ID:            result.ID,
		Title:         "更新後の制約条件",
		Category:      ConstraintCategoryBusiness,
		Description:   "制約の詳細説明",
		NonNegotiable: true,
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

	constraint := entity.(*ConstraintEntity)
	if constraint.Title != "更新後の制約条件" {
		t.Errorf("Title after Update = %q, want %q", constraint.Title, "更新後の制約条件")
	}
	if constraint.Category != ConstraintCategoryBusiness {
		t.Errorf("Category after Update = %q, want %q", constraint.Category, ConstraintCategoryBusiness)
	}
	if !constraint.NonNegotiable {
		t.Error("NonNegotiable after Update = false, want true")
	}
}

func TestConstraintHandler_Update_NotFound(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	updated := &ConstraintEntity{
		ID:    "const-999",
		Title: "存在しない制約",
	}

	err := handler.Update(ctx, "const-999", updated)
	if err != ErrEntityNotFound {
		t.Errorf("Update() for non-existent ID error = %v, want ErrEntityNotFound", err)
	}
}

func TestConstraintHandler_Update_InvalidType(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "制約条件")
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

func TestConstraintHandler_Delete(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "削除予定の制約条件")
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

func TestConstraintHandler_Delete_NotFound(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	err := handler.Delete(ctx, "const-999")
	if err != ErrEntityNotFound {
		t.Errorf("Delete() for non-existent ID error = %v, want ErrEntityNotFound", err)
	}
}

// ===== GetAllConstraints() テスト =====

func TestConstraintHandler_GetAllConstraints(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 3つの Constraint を追加
	for i := 0; i < 3; i++ {
		_, err := handler.Add(ctx, "制約条件"+string(rune('A'+i)))
		if err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	}

	all, err := handler.GetAllConstraints(ctx)
	if err != nil {
		t.Fatalf("GetAllConstraints() error = %v", err)
	}

	if len(all) != 3 {
		t.Errorf("GetAllConstraints() count = %d, want 3", len(all))
	}
}

func TestConstraintHandler_GetAllConstraints_Empty(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	all, err := handler.GetAllConstraints(ctx)
	if err != nil {
		t.Fatalf("GetAllConstraints() error = %v", err)
	}

	if len(all) != 0 {
		t.Errorf("GetAllConstraints() count = %d, want 0", len(all))
	}
}

// ===== GetConstraintsByCategory() テスト =====

func TestConstraintHandler_GetConstraintsByCategory(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 異なるカテゴリの Constraint を追加
	_, err := handler.Add(ctx, "技術制約 1", WithConstraintCategory(ConstraintCategoryTechnical))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	_, err = handler.Add(ctx, "ビジネス制約", WithConstraintCategory(ConstraintCategoryBusiness))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	_, err = handler.Add(ctx, "技術制約 2", WithConstraintCategory(ConstraintCategoryTechnical))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	_, err = handler.Add(ctx, "法的制約", WithConstraintCategory(ConstraintCategoryLegal))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Technical カテゴリを取得
	technical, err := handler.GetConstraintsByCategory(ctx, ConstraintCategoryTechnical)
	if err != nil {
		t.Fatalf("GetConstraintsByCategory() error = %v", err)
	}

	if len(technical) != 2 {
		t.Errorf("GetConstraintsByCategory(Technical) count = %d, want 2", len(technical))
	}

	// Business カテゴリを取得
	business, err := handler.GetConstraintsByCategory(ctx, ConstraintCategoryBusiness)
	if err != nil {
		t.Fatalf("GetConstraintsByCategory() error = %v", err)
	}

	if len(business) != 1 {
		t.Errorf("GetConstraintsByCategory(Business) count = %d, want 1", len(business))
	}
}

// ===== GetNonNegotiableConstraints() テスト =====

func TestConstraintHandler_GetNonNegotiableConstraints(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 交渉可能/不可の Constraint を追加
	_, err := handler.Add(ctx, "交渉可能 1", WithConstraintNonNegotiable(false))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	_, err = handler.Add(ctx, "交渉不可 1", WithConstraintNonNegotiable(true))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	_, err = handler.Add(ctx, "交渉不可 2", WithConstraintNonNegotiable(true))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	_, err = handler.Add(ctx, "交渉可能 2", WithConstraintNonNegotiable(false))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	nonNegotiable, err := handler.GetNonNegotiableConstraints(ctx)
	if err != nil {
		t.Fatalf("GetNonNegotiableConstraints() error = %v", err)
	}

	if len(nonNegotiable) != 2 {
		t.Errorf("GetNonNegotiableConstraints() count = %d, want 2", len(nonNegotiable))
	}

	// 全て NonNegotiable = true であることを確認
	for _, c := range nonNegotiable {
		if !c.NonNegotiable {
			t.Errorf("GetNonNegotiableConstraints() returned negotiable constraint: %s", c.ID)
		}
	}
}

func TestConstraintHandler_GetNonNegotiableConstraints_Empty(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 交渉可能なもののみ追加
	_, err := handler.Add(ctx, "交渉可能", WithConstraintNonNegotiable(false))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	nonNegotiable, err := handler.GetNonNegotiableConstraints(ctx)
	if err != nil {
		t.Fatalf("GetNonNegotiableConstraints() error = %v", err)
	}

	if len(nonNegotiable) != 0 {
		t.Errorf("GetNonNegotiableConstraints() count = %d, want 0", len(nonNegotiable))
	}
}

// ===== Context キャンセルテスト =====

func TestConstraintHandler_ContextCancellation(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
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
	_, err = handler.Get(ctx, "const-001")
	if err == nil {
		t.Error("Get() with cancelled context should return error")
	}

	// Update
	err = handler.Update(ctx, "const-001", &ConstraintEntity{})
	if err == nil {
		t.Error("Update() with cancelled context should return error")
	}

	// Delete
	err = handler.Delete(ctx, "const-001")
	if err == nil {
		t.Error("Delete() with cancelled context should return error")
	}
}

// ===== ID 採番テスト =====

func TestConstraintHandler_IDSequence(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 連続追加で ID が順番に採番されることを確認
	expectedIDs := []string{"const-001", "const-002", "const-003"}

	for i, expected := range expectedIDs {
		result, err := handler.Add(ctx, "制約条件"+string(rune('A'+i)))
		if err != nil {
			t.Fatalf("Add() #%d error = %v", i+1, err)
		}
		if result.ID != expected {
			t.Errorf("Add() #%d ID = %q, want %q", i+1, result.ID, expected)
		}
	}
}

// ===== ファイルパステスト（単一ファイル管理）=====

func TestConstraintHandler_SingleFilePath(t *testing.T) {
	handler, zeusPath, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "テスト制約条件")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// 単一ファイル constraints.yaml が作成されたか確認
	expectedPath := filepath.Join(zeusPath, "constraints.yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file at %q does not exist", expectedPath)
	}

	// 追加で constraint を作成しても同じファイルに格納される
	_, err = handler.Add(ctx, "二つ目の制約条件")
	if err != nil {
		t.Fatalf("Add() second error = %v", err)
	}

	// ファイルは1つだけであることを確認（constraints.yaml）
	entries, err := os.ReadDir(zeusPath)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	yamlCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() == "constraints.yaml" {
			yamlCount++
		}
	}
	if yamlCount != 1 {
		t.Errorf("Expected 1 constraints.yaml file, found %d", yamlCount)
	}

	// 両方の constraint が取得できることを確認
	all, err := handler.GetAllConstraints(ctx)
	if err != nil {
		t.Fatalf("GetAllConstraints() error = %v", err)
	}
	if len(all) != 2 {
		t.Errorf("GetAllConstraints() count = %d, want 2", len(all))
	}

	// 最初に作成した constraint が取得できることを確認
	entity, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if entity.(*ConstraintEntity).Title != "テスト制約条件" {
		t.Errorf("Title = %q, want %q", entity.(*ConstraintEntity).Title, "テスト制約条件")
	}
}

// ===== 全カテゴリテスト =====

func TestConstraintHandler_AllCategories(t *testing.T) {
	handler, _, cleanup := setupConstraintHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	categories := []ConstraintCategory{
		ConstraintCategoryTechnical,
		ConstraintCategoryBusiness,
		ConstraintCategoryLegal,
		ConstraintCategoryResource,
	}

	for _, category := range categories {
		t.Run(string(category), func(t *testing.T) {
			result, err := handler.Add(ctx, "制約 "+string(category), WithConstraintCategory(category))
			if err != nil {
				t.Fatalf("Add() error = %v", err)
			}

			entity, err := handler.Get(ctx, result.ID)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}

			constraint := entity.(*ConstraintEntity)
			if constraint.Category != category {
				t.Errorf("Category = %q, want %q", constraint.Category, category)
			}
		})
	}
}
