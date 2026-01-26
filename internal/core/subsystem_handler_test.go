package core

import (
	"context"
	"os"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// テスト用のセットアップ
func setupSubsystemHandlerTest(t *testing.T) (*SubsystemHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-subsystem-test")
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
	handler := NewSubsystemHandler(fs)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return handler, zeusPath, cleanup
}

func TestSubsystemHandlerType(t *testing.T) {
	handler, _, cleanup := setupSubsystemHandlerTest(t)
	defer cleanup()

	if handler.Type() != "subsystem" {
		t.Errorf("expected type 'subsystem', got %q", handler.Type())
	}
}

func TestSubsystemHandlerAdd(t *testing.T) {
	handler, _, cleanup := setupSubsystemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// サブシステム追加
	result, err := handler.Add(ctx, "Test Subsystem")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if !result.Success {
		t.Error("Add should succeed")
	}

	if result.Entity != "subsystem" {
		t.Errorf("expected entity 'subsystem', got %q", result.Entity)
	}

	if result.ID == "" {
		t.Error("ID should not be empty")
	}

	// リストで確認
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 1 {
		t.Errorf("expected 1 subsystem, got %d", listResult.Total)
	}
}

func TestSubsystemHandlerAddWithOptions(t *testing.T) {
	handler, _, cleanup := setupSubsystemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// オプション付きでサブシステム追加
	result, err := handler.Add(ctx, "Test Subsystem with Options",
		WithSubsystemDescription("This is a test subsystem"),
		WithSubsystemOwner("test-user"),
		WithSubsystemTags([]string{"core", "backend"}),
	)
	if err != nil {
		t.Fatalf("Add with options failed: %v", err)
	}

	// サブシステムを取得して確認
	subsystemAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	subsystem := subsystemAny.(*SubsystemEntity)
	if subsystem.Description != "This is a test subsystem" {
		t.Errorf("expected description 'This is a test subsystem', got %q", subsystem.Description)
	}

	if subsystem.Metadata.Owner != "test-user" {
		t.Errorf("expected owner 'test-user', got %q", subsystem.Metadata.Owner)
	}

	if len(subsystem.Metadata.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(subsystem.Metadata.Tags))
	}
}

func TestSubsystemHandlerList(t *testing.T) {
	handler, _, cleanup := setupSubsystemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 複数サブシステムを追加
	for i := 0; i < 5; i++ {
		_, err := handler.Add(ctx, "Subsystem")
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
		t.Errorf("expected 5 subsystems, got %d", listResult.Total)
	}
}

func TestSubsystemHandlerListEmpty(t *testing.T) {
	handler, _, cleanup := setupSubsystemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 空のリスト
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 0 {
		t.Errorf("expected 0 subsystems, got %d", listResult.Total)
	}
}

func TestSubsystemHandlerGet(t *testing.T) {
	handler, _, cleanup := setupSubsystemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// サブシステム追加
	result, err := handler.Add(ctx, "Get Test Subsystem")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// サブシステムを取得
	subsystemAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	subsystem := subsystemAny.(*SubsystemEntity)
	if subsystem.Name != "Get Test Subsystem" {
		t.Errorf("expected name 'Get Test Subsystem', got %q", subsystem.Name)
	}
}

func TestSubsystemHandlerGetNotFound(t *testing.T) {
	handler, _, cleanup := setupSubsystemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で取得（有効なフォーマット）
	_, err := handler.Get(ctx, "sub-00000000")
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestSubsystemHandlerGetInvalidID(t *testing.T) {
	handler, _, cleanup := setupSubsystemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 無効な ID フォーマットで取得
	_, err := handler.Get(ctx, "invalid-id")
	if err == nil {
		t.Error("expected error for invalid ID format")
	}
}

func TestSubsystemHandlerUpdate(t *testing.T) {
	handler, _, cleanup := setupSubsystemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// サブシステム追加
	result, err := handler.Add(ctx, "Update Test Subsystem")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 更新
	updateData := map[string]any{
		"name":        "Updated Name",
		"description": "Updated description",
	}
	err = handler.Update(ctx, result.ID, updateData)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// 更新を確認
	updatedAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}

	updated := updatedAny.(*SubsystemEntity)
	if updated.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %q", updated.Name)
	}

	if updated.Description != "Updated description" {
		t.Errorf("expected description 'Updated description', got %q", updated.Description)
	}
}

func TestSubsystemHandlerUpdateNotFound(t *testing.T) {
	handler, _, cleanup := setupSubsystemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で更新（有効なフォーマット）
	err := handler.Update(ctx, "sub-00000000", map[string]any{"name": "Test"})
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestSubsystemHandlerDelete(t *testing.T) {
	handler, _, cleanup := setupSubsystemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// サブシステム追加
	result, err := handler.Add(ctx, "Delete Test Subsystem")
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
		t.Error("expected error for deleted subsystem")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestSubsystemHandlerDeleteNotFound(t *testing.T) {
	handler, _, cleanup := setupSubsystemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で削除（有効なフォーマット）
	err := handler.Delete(ctx, "sub-00000000")
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestSubsystemHandlerListAll(t *testing.T) {
	handler, _, cleanup := setupSubsystemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 複数サブシステムを追加
	for i := 0; i < 3; i++ {
		_, err := handler.Add(ctx, "Subsystem")
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// ListAll で取得
	subsystems, err := handler.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}

	if len(subsystems) != 3 {
		t.Errorf("expected 3 subsystems, got %d", len(subsystems))
	}
}

func TestSubsystemHandlerListAllEmpty(t *testing.T) {
	handler, _, cleanup := setupSubsystemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 空の ListAll
	subsystems, err := handler.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}

	if len(subsystems) != 0 {
		t.Errorf("expected 0 subsystems, got %d", len(subsystems))
	}
}

func TestSubsystemHandlerContextCancellation(t *testing.T) {
	handler, _, cleanup := setupSubsystemHandlerTest(t)
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

	// Get（有効なフォーマット）
	_, err = handler.Get(ctx, "sub-00000000")
	if err == nil {
		t.Error("Get should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Update（有効なフォーマット）
	err = handler.Update(ctx, "sub-00000000", map[string]any{})
	if err == nil {
		t.Error("Update should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Delete（有効なフォーマット）
	err = handler.Delete(ctx, "sub-00000000")
	if err == nil {
		t.Error("Delete should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// ListAll
	_, err = handler.ListAll(ctx)
	if err == nil {
		t.Error("ListAll should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestGenerateSubsystemIDFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-subsystem-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	zeusPath := tmpDir + "/.zeus"
	fs := yaml.NewFileManager(zeusPath)
	handler := NewSubsystemHandler(fs)

	// ID 生成テスト
	id := handler.generateSubsystemID()

	// プレフィックスが正しいか
	if len(id) < 4 || id[:4] != "sub-" {
		t.Errorf("expected ID to start with 'sub-', got %q", id)
	}

	// 長さが正しいか (sub- + 8文字)
	if len(id) != 12 {
		t.Errorf("expected ID length to be 12, got %d", len(id))
	}
}

func TestSubsystemEntityValidate(t *testing.T) {
	tests := []struct {
		name      string
		subsystem SubsystemEntity
		wantErr   bool
	}{
		{
			name: "valid subsystem",
			subsystem: SubsystemEntity{
				ID:   "sub-12345678",
				Name: "Test Subsystem",
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			subsystem: SubsystemEntity{
				Name: "Test Subsystem",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			subsystem: SubsystemEntity{
				ID: "sub-12345678",
			},
			wantErr: true,
		},
		{
			name: "invalid ID format",
			subsystem: SubsystemEntity{
				ID:   "invalid-id",
				Name: "Test Subsystem",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.subsystem.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSubsystemHandlerMultipleAddDelete(t *testing.T) {
	handler, _, cleanup := setupSubsystemHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 複数追加
	ids := make([]string, 3)
	for i := 0; i < 3; i++ {
		result, err := handler.Add(ctx, "Subsystem")
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
		ids[i] = result.ID
	}

	// 中間のサブシステムを削除
	err := handler.Delete(ctx, ids[1])
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 残りを確認
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 2 {
		t.Errorf("expected 2 subsystems, got %d", listResult.Total)
	}

	// 削除したサブシステムは取得できない
	_, err = handler.Get(ctx, ids[1])
	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound for deleted subsystem, got %v", err)
	}

	// 残りのサブシステムは取得できる
	_, err = handler.Get(ctx, ids[0])
	if err != nil {
		t.Errorf("Get first subsystem failed: %v", err)
	}

	_, err = handler.Get(ctx, ids[2])
	if err != nil {
		t.Errorf("Get third subsystem failed: %v", err)
	}
}
