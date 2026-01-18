package core

import (
	"context"
	"os"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// テスト用のセットアップ
func setupObjectiveHandlerTest(t *testing.T) (*ObjectiveHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-objective-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath+"/objectives", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create zeus dir: %v", err)
	}

	fs := yaml.NewFileManager(zeusPath)
	handler := NewObjectiveHandler(fs, nil)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return handler, zeusPath, cleanup
}

func TestObjectiveHandlerType(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
	defer cleanup()

	if handler.Type() != "objective" {
		t.Errorf("expected type 'objective', got %q", handler.Type())
	}
}

func TestObjectiveHandlerAdd(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Objective 追加
	result, err := handler.Add(ctx, "認証システム実装")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if !result.Success {
		t.Error("Add should succeed")
	}

	if result.Entity != "objective" {
		t.Errorf("expected entity 'objective', got %q", result.Entity)
	}

	if result.ID == "" {
		t.Error("ID should not be empty")
	}

	// ID フォーマット確認 (obj-NNN)
	if len(result.ID) < 4 || result.ID[:4] != "obj-" {
		t.Errorf("expected ID to start with 'obj-', got %q", result.ID)
	}

	// リストで確認
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 1 {
		t.Errorf("expected 1 objective, got %d", listResult.Total)
	}
}

func TestObjectiveHandlerAddWithOptions(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// オプション付きで Objective 追加
	result, err := handler.Add(ctx, "認証システム実装",
		WithObjectiveDescription("JWT を使用した認証システムを実装"),
		WithObjectiveStatus(ObjectiveStatusInProgress),
		WithObjectiveWBSCode("1.1"),
		WithObjectiveProgress(30),
		WithObjectiveDueDate("2026-02-28"),
		WithObjectiveOwner("test-user"),
		WithObjectiveTags([]string{"backend", "security"}),
	)
	if err != nil {
		t.Fatalf("Add with options failed: %v", err)
	}

	// Objective を取得して確認
	objAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	obj := objAny.(*ObjectiveEntity)
	if obj.Description != "JWT を使用した認証システムを実装" {
		t.Errorf("expected description 'JWT を使用した認証システムを実装', got %q", obj.Description)
	}

	if obj.Status != ObjectiveStatusInProgress {
		t.Errorf("expected status 'in_progress', got %q", obj.Status)
	}

	if obj.WBSCode != "1.1" {
		t.Errorf("expected WBS code '1.1', got %q", obj.WBSCode)
	}

	if obj.Progress != 30 {
		t.Errorf("expected progress 30, got %d", obj.Progress)
	}

	if obj.DueDate != "2026-02-28" {
		t.Errorf("expected due date '2026-02-28', got %q", obj.DueDate)
	}

	if obj.Owner != "test-user" {
		t.Errorf("expected owner 'test-user', got %q", obj.Owner)
	}

	if len(obj.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(obj.Tags))
	}
}

func TestObjectiveHandlerAddWithParent(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 親 Objective を作成
	parentResult, err := handler.Add(ctx, "親 Objective")
	if err != nil {
		t.Fatalf("Add parent failed: %v", err)
	}

	// 子 Objective を作成
	childResult, err := handler.Add(ctx, "子 Objective",
		WithObjectiveParent(parentResult.ID),
	)
	if err != nil {
		t.Fatalf("Add child failed: %v", err)
	}

	// 子 Objective を取得して確認
	childAny, err := handler.Get(ctx, childResult.ID)
	if err != nil {
		t.Fatalf("Get child failed: %v", err)
	}

	child := childAny.(*ObjectiveEntity)
	if child.ParentID != parentResult.ID {
		t.Errorf("expected parent_id %q, got %q", parentResult.ID, child.ParentID)
	}
}

func TestObjectiveHandlerAddWithInvalidParent(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない親 ID で追加
	// 注: ObjectiveHandler は Add 時に親の存在チェックをしない（循環参照チェックのみ）
	// 参照整合性は IntegrityChecker で検証する設計
	result, err := handler.Add(ctx, "子 Objective",
		WithObjectiveParent("obj-999"),
	)

	// Add は成功する（参照整合性は IntegrityChecker で検証）
	if err != nil {
		t.Fatalf("Add should succeed even with invalid parent: %v", err)
	}

	if result.ID == "" {
		t.Error("ID should not be empty")
	}

	// 作成された Objective の parent_id を確認
	objAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	obj := objAny.(*ObjectiveEntity)
	if obj.ParentID != "obj-999" {
		t.Errorf("expected parent_id 'obj-999', got %q", obj.ParentID)
	}
}

func TestObjectiveHandlerList(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 複数 Objective を追加
	for i := 0; i < 5; i++ {
		_, err := handler.Add(ctx, "Objective")
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// 全リスト
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 5 {
		t.Errorf("expected 5 objectives, got %d", listResult.Total)
	}
}

func TestObjectiveHandlerListWithFilter(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 異なるステータスの Objective を追加
	_, err := handler.Add(ctx, "Not Started", WithObjectiveStatus(ObjectiveStatusNotStarted))
	if err != nil {
		t.Fatalf("Add not_started failed: %v", err)
	}

	_, err = handler.Add(ctx, "In Progress", WithObjectiveStatus(ObjectiveStatusInProgress))
	if err != nil {
		t.Fatalf("Add in_progress failed: %v", err)
	}

	_, err = handler.Add(ctx, "Completed", WithObjectiveStatus(ObjectiveStatusCompleted))
	if err != nil {
		t.Fatalf("Add completed failed: %v", err)
	}

	// in_progress でフィルタ
	filter := &ListFilter{Status: string(ObjectiveStatusInProgress)}
	listResult, err := handler.List(ctx, filter)
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}

	if listResult.Total != 1 {
		t.Errorf("expected 1 in_progress objective, got %d", listResult.Total)
	}
}

func TestObjectiveHandlerGet(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Objective 追加
	result, err := handler.Add(ctx, "Get Test Objective")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Objective を取得
	objAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	obj := objAny.(*ObjectiveEntity)
	if obj.Title != "Get Test Objective" {
		t.Errorf("expected title 'Get Test Objective', got %q", obj.Title)
	}
}

func TestObjectiveHandlerGetNotFound(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で取得
	_, err := handler.Get(ctx, "obj-999")
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestObjectiveHandlerUpdate(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Objective 追加
	result, err := handler.Add(ctx, "Update Test Objective")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Objective を取得
	objAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	obj := objAny.(*ObjectiveEntity)

	// 更新
	obj.Title = "Updated Title"
	obj.Status = ObjectiveStatusCompleted
	obj.Progress = 100
	err = handler.Update(ctx, result.ID, obj)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// 更新を確認
	updatedAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}

	updated := updatedAny.(*ObjectiveEntity)
	if updated.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", updated.Title)
	}

	if updated.Status != ObjectiveStatusCompleted {
		t.Errorf("expected status 'completed', got %q", updated.Status)
	}

	if updated.Progress != 100 {
		t.Errorf("expected progress 100, got %d", updated.Progress)
	}
}

func TestObjectiveHandlerUpdateNotFound(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で更新
	obj := &ObjectiveEntity{ID: "obj-999", Title: "Test"}
	err := handler.Update(ctx, "obj-999", obj)
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestObjectiveHandlerDelete(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Objective 追加
	result, err := handler.Add(ctx, "Delete Test Objective")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 削除
	err = handler.Delete(ctx, result.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 削除されたことを確認
	_, err = handler.Get(ctx, result.ID)
	if err == nil {
		t.Error("expected error for deleted objective")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestObjectiveHandlerDeleteNotFound(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で削除
	err := handler.Delete(ctx, "obj-999")
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestObjectiveHandlerContextCancellation(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
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
	_, err = handler.Get(ctx, "obj-001")
	if err == nil {
		t.Error("Get should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Update
	err = handler.Update(ctx, "obj-001", &ObjectiveEntity{})
	if err == nil {
		t.Error("Update should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Delete
	err = handler.Delete(ctx, "obj-001")
	if err == nil {
		t.Error("Delete should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestObjectiveHandlerDeleteWithChildren(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 親 Objective を作成
	parentResult, err := handler.Add(ctx, "親 Objective")
	if err != nil {
		t.Fatalf("Add parent failed: %v", err)
	}

	// 子 Objective を作成
	_, err = handler.Add(ctx, "子 Objective",
		WithObjectiveParent(parentResult.ID),
	)
	if err != nil {
		t.Fatalf("Add child failed: %v", err)
	}

	// 子がいる親を削除しようとするとエラー
	err = handler.Delete(ctx, parentResult.ID)
	if err == nil {
		t.Error("expected error when deleting parent with children")
	}
}

func TestObjectiveHandlerProgressValidation(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 正常な進捗 (0-100)
	result, err := handler.Add(ctx, "Valid Progress", WithObjectiveProgress(50))
	if err != nil {
		t.Fatalf("Add with valid progress failed: %v", err)
	}

	objAny, _ := handler.Get(ctx, result.ID)
	obj := objAny.(*ObjectiveEntity)
	if obj.Progress != 50 {
		t.Errorf("expected progress 50, got %d", obj.Progress)
	}
}

func TestObjectiveHandlerIDSequence(t *testing.T) {
	handler, _, cleanup := setupObjectiveHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 複数 Objective を追加
	ids := make([]string, 3)
	for i := 0; i < 3; i++ {
		result, err := handler.Add(ctx, "Objective")
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
		ids[i] = result.ID
	}

	// ID が連続していることを確認
	if ids[0] != "obj-001" {
		t.Errorf("expected first ID 'obj-001', got %q", ids[0])
	}
	if ids[1] != "obj-002" {
		t.Errorf("expected second ID 'obj-002', got %q", ids[1])
	}
	if ids[2] != "obj-003" {
		t.Errorf("expected third ID 'obj-003', got %q", ids[2])
	}
}
