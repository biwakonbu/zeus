package core

import (
	"context"
	"os"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// テスト用のセットアップ
func setupVisionHandlerTest(t *testing.T) (*VisionHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-vision-test")
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
	handler := NewVisionHandler(fs)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return handler, zeusPath, cleanup
}

func TestVisionHandlerType(t *testing.T) {
	handler, _, cleanup := setupVisionHandlerTest(t)
	defer cleanup()

	if handler.Type() != "vision" {
		t.Errorf("expected type 'vision', got %q", handler.Type())
	}
}

func TestVisionHandlerAdd(t *testing.T) {
	handler, _, cleanup := setupVisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Vision 追加（statement は必須）
	result, err := handler.Add(ctx, "AI駆動プロジェクト管理",
		WithVisionStatement("AIと人間の協調によるプロジェクト管理の実現"),
	)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if !result.Success {
		t.Error("Add should succeed")
	}

	if result.Entity != "vision" {
		t.Errorf("expected entity 'vision', got %q", result.Entity)
	}

	// ID は常に vision-001
	if result.ID != "vision-001" {
		t.Errorf("expected ID 'vision-001', got %q", result.ID)
	}

	// リストで確認
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 1 {
		t.Errorf("expected 1 vision, got %d", listResult.Total)
	}
}

func TestVisionHandlerAddWithOptions(t *testing.T) {
	handler, _, cleanup := setupVisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// オプション付きで Vision 追加
	result, err := handler.Add(ctx, "AI駆動プロジェクト管理",
		WithVisionStatement("AI と人間の協調によるプロジェクト管理の実現"),
		WithVisionSuccessCriteria([]string{
			"全プロジェクトで AI アシスタントを活用",
			"タスク完了率 90% 以上",
		}),
		WithVisionStatus(VisionStatusActive),
		WithVisionOwner("test-user"),
		WithVisionTags([]string{"ai", "pm"}),
	)
	if err != nil {
		t.Fatalf("Add with options failed: %v", err)
	}

	// Vision を取得して確認
	visionAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	vision := visionAny.(*Vision)
	if vision.Statement != "AI と人間の協調によるプロジェクト管理の実現" {
		t.Errorf("expected statement 'AI と人間の協調によるプロジェクト管理の実現', got %q", vision.Statement)
	}

	if vision.Status != VisionStatusActive {
		t.Errorf("expected status 'active', got %q", vision.Status)
	}

	if len(vision.SuccessCriteria) != 2 {
		t.Errorf("expected 2 success criteria, got %d", len(vision.SuccessCriteria))
	}

	if vision.Metadata.Owner != "test-user" {
		t.Errorf("expected owner 'test-user', got %q", vision.Metadata.Owner)
	}

	if len(vision.Metadata.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(vision.Metadata.Tags))
	}
}

func TestVisionHandlerAddUpdatesExisting(t *testing.T) {
	handler, _, cleanup := setupVisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 最初の Vision を作成（statement は必須）
	_, err := handler.Add(ctx, "初期 Vision",
		WithVisionStatement("初期ステートメント"),
	)
	if err != nil {
		t.Fatalf("First add failed: %v", err)
	}

	// 2回目の Add は既存を更新
	result, err := handler.Add(ctx, "更新された Vision",
		WithVisionStatement("新しいステートメント"),
	)
	if err != nil {
		t.Fatalf("Second add failed: %v", err)
	}

	// ID は同じ
	if result.ID != "vision-001" {
		t.Errorf("expected ID 'vision-001', got %q", result.ID)
	}

	// 更新された内容を確認
	visionAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	vision := visionAny.(*Vision)
	if vision.Title != "更新された Vision" {
		t.Errorf("expected title '更新された Vision', got %q", vision.Title)
	}
	if vision.Statement != "新しいステートメント" {
		t.Errorf("expected statement '新しいステートメント', got %q", vision.Statement)
	}

	// リストは常に 1
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 1 {
		t.Errorf("expected 1 vision, got %d", listResult.Total)
	}
}

func TestVisionHandlerListEmpty(t *testing.T) {
	handler, _, cleanup := setupVisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Vision がない場合
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 0 {
		t.Errorf("expected 0 visions, got %d", listResult.Total)
	}
}

func TestVisionHandlerGet(t *testing.T) {
	handler, _, cleanup := setupVisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Vision 追加（statement は必須）
	_, err := handler.Add(ctx, "Get Test Vision",
		WithVisionStatement("テスト用ステートメント"),
	)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Vision を取得
	visionAny, err := handler.Get(ctx, "vision-001")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	vision := visionAny.(*Vision)
	if vision.Title != "Get Test Vision" {
		t.Errorf("expected title 'Get Test Vision', got %q", vision.Title)
	}
}

func TestVisionHandlerGetNotFound(t *testing.T) {
	handler, _, cleanup := setupVisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない Vision で取得
	_, err := handler.Get(ctx, "vision-001")
	if err == nil {
		t.Error("expected error for non-existent vision")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestVisionHandlerGetVision(t *testing.T) {
	handler, _, cleanup := setupVisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Vision 追加（statement は必須）
	_, err := handler.Add(ctx, "Helper Test Vision",
		WithVisionStatement("ヘルパーテスト用ステートメント"),
	)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// GetVision ヘルパーで取得
	vision, err := handler.GetVision(ctx)
	if err != nil {
		t.Fatalf("GetVision failed: %v", err)
	}

	if vision.Title != "Helper Test Vision" {
		t.Errorf("expected title 'Helper Test Vision', got %q", vision.Title)
	}
}

func TestVisionHandlerUpdate(t *testing.T) {
	handler, _, cleanup := setupVisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Vision 追加（statement は必須）
	result, err := handler.Add(ctx, "Update Test Vision",
		WithVisionStatement("更新テスト用ステートメント"),
	)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Vision を取得
	visionAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	vision := visionAny.(*Vision)

	// 更新
	vision.Title = "Updated Title"
	vision.Status = VisionStatusActive
	err = handler.Update(ctx, result.ID, vision)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// 更新を確認
	updatedAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}

	updated := updatedAny.(*Vision)
	if updated.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", updated.Title)
	}

	if updated.Status != VisionStatusActive {
		t.Errorf("expected status 'active', got %q", updated.Status)
	}
}

func TestVisionHandlerUpdateNotFound(t *testing.T) {
	handler, _, cleanup := setupVisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない Vision で更新
	vision := &Vision{ID: "vision-001", Title: "Test"}
	err := handler.Update(ctx, "vision-001", vision)
	if err == nil {
		t.Error("expected error for non-existent vision")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestVisionHandlerDeleteNotAllowed(t *testing.T) {
	handler, _, cleanup := setupVisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Vision 追加（statement は必須）
	_, err := handler.Add(ctx, "Delete Test Vision",
		WithVisionStatement("削除テスト用ステートメント"),
	)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 削除しようとするとエラー
	err = handler.Delete(ctx, "vision-001")
	if err == nil {
		t.Error("expected error for deleting vision")
	}
}

func TestVisionHandlerContextCancellation(t *testing.T) {
	handler, _, cleanup := setupVisionHandlerTest(t)
	defer cleanup()

	// キャンセル済みのコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Add
	_, err := handler.Add(ctx, "Test")
	if err == nil {
		t.Error("Add should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// List
	_, err = handler.List(ctx, nil)
	if err == nil {
		t.Error("List should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Get
	_, err = handler.Get(ctx, "vision-001")
	if err == nil {
		t.Error("Get should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Update
	err = handler.Update(ctx, "vision-001", &Vision{})
	if err == nil {
		t.Error("Update should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestVisionHandlerInvalidID(t *testing.T) {
	handler, _, cleanup := setupVisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 無効な ID で取得
	_, err := handler.Get(ctx, "invalid-id")
	if err == nil {
		t.Error("expected error for invalid ID")
	}

	// 無効な ID で更新
	err = handler.Update(ctx, "invalid-id", &Vision{})
	if err == nil {
		t.Error("expected error for invalid ID")
	}
}

func TestVisionHandlerPreservesCreatedAt(t *testing.T) {
	handler, _, cleanup := setupVisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Vision 追加（statement は必須）
	result, err := handler.Add(ctx, "Initial Vision",
		WithVisionStatement("CreatedAt テスト用ステートメント"),
	)
	if err != nil {
		t.Fatalf("First add failed: %v", err)
	}

	// 最初の CreatedAt を取得
	visionAny, _ := handler.Get(ctx, result.ID)
	original := visionAny.(*Vision)
	originalCreatedAt := original.Metadata.CreatedAt

	// 更新
	original.Title = "Updated Vision"
	err = handler.Update(ctx, result.ID, original)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// CreatedAt が保持されていることを確認
	updatedAny, _ := handler.Get(ctx, result.ID)
	updated := updatedAny.(*Vision)

	if updated.Metadata.CreatedAt != originalCreatedAt {
		t.Errorf("CreatedAt should be preserved, got %v, expected %v",
			updated.Metadata.CreatedAt, originalCreatedAt)
	}
}
