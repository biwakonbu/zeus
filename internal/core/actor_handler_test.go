package core

import (
	"context"
	"os"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// テスト用のセットアップ
func setupActorHandlerTest(t *testing.T) (*ActorHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-actor-test")
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
	handler := NewActorHandler(fs)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return handler, zeusPath, cleanup
}

func TestActorHandlerType(t *testing.T) {
	handler, _, cleanup := setupActorHandlerTest(t)
	defer cleanup()

	if handler.Type() != "actor" {
		t.Errorf("expected type 'actor', got %q", handler.Type())
	}
}

func TestActorHandlerAdd(t *testing.T) {
	handler, _, cleanup := setupActorHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクター追加
	result, err := handler.Add(ctx, "Test Actor")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if !result.Success {
		t.Error("Add should succeed")
	}

	if result.Entity != "actor" {
		t.Errorf("expected entity 'actor', got %q", result.Entity)
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
		t.Errorf("expected 1 actor, got %d", listResult.Total)
	}
}

func TestActorHandlerAddWithOptions(t *testing.T) {
	handler, _, cleanup := setupActorHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// オプション付きでアクター追加
	result, err := handler.Add(ctx, "Test Actor with Options",
		WithActorType(ActorTypeSystem),
		WithActorDescription("This is a system actor"),
		WithActorOwner("test-user"),
		WithActorTags([]string{"system", "api"}),
	)
	if err != nil {
		t.Fatalf("Add with options failed: %v", err)
	}

	// アクターを取得して確認
	actorAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	actor := actorAny.(*ActorEntity)
	if actor.Type != ActorTypeSystem {
		t.Errorf("expected type 'system', got %q", actor.Type)
	}

	if actor.Description != "This is a system actor" {
		t.Errorf("expected description 'This is a system actor', got %q", actor.Description)
	}

	if actor.Metadata.Owner != "test-user" {
		t.Errorf("expected owner 'test-user', got %q", actor.Metadata.Owner)
	}

	if len(actor.Metadata.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(actor.Metadata.Tags))
	}
}

func TestActorHandlerAddAllTypes(t *testing.T) {
	handler, _, cleanup := setupActorHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name      string
		actorType ActorType
	}{
		{"Human Actor", ActorTypeHuman},
		{"System Actor", ActorTypeSystem},
		{"Time Actor", ActorTypeTime},
		{"Device Actor", ActorTypeDevice},
		{"External Actor", ActorTypeExternal},
	}

	for _, tt := range tests {
		result, err := handler.Add(ctx, tt.name, WithActorType(tt.actorType))
		if err != nil {
			t.Fatalf("Add %s failed: %v", tt.name, err)
		}

		actorAny, err := handler.Get(ctx, result.ID)
		if err != nil {
			t.Fatalf("Get %s failed: %v", tt.name, err)
		}

		actor := actorAny.(*ActorEntity)
		if actor.Type != tt.actorType {
			t.Errorf("expected type %q, got %q", tt.actorType, actor.Type)
		}
	}

	// 全アクターがリストで確認できる
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 5 {
		t.Errorf("expected 5 actors, got %d", listResult.Total)
	}
}

func TestActorHandlerList(t *testing.T) {
	handler, _, cleanup := setupActorHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 複数アクターを追加
	for i := 0; i < 5; i++ {
		_, err := handler.Add(ctx, "Actor")
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
		t.Errorf("expected 5 actors, got %d", listResult.Total)
	}
}

func TestActorHandlerListEmpty(t *testing.T) {
	handler, _, cleanup := setupActorHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 空のリスト
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 0 {
		t.Errorf("expected 0 actors, got %d", listResult.Total)
	}
}

func TestActorHandlerGet(t *testing.T) {
	handler, _, cleanup := setupActorHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクター追加
	result, err := handler.Add(ctx, "Get Test Actor")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// アクターを取得
	actorAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	actor := actorAny.(*ActorEntity)
	if actor.Title != "Get Test Actor" {
		t.Errorf("expected title 'Get Test Actor', got %q", actor.Title)
	}
}

func TestActorHandlerGetNotFound(t *testing.T) {
	handler, _, cleanup := setupActorHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で取得（有効なフォーマット）
	_, err := handler.Get(ctx, "actor-00000000")
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestActorHandlerUpdate(t *testing.T) {
	handler, _, cleanup := setupActorHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクター追加
	result, err := handler.Add(ctx, "Update Test Actor")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 更新
	updateData := map[string]any{
		"title":       "Updated Title",
		"type":        "system",
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

	updated := updatedAny.(*ActorEntity)
	if updated.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", updated.Title)
	}

	if updated.Type != ActorTypeSystem {
		t.Errorf("expected type 'system', got %q", updated.Type)
	}

	if updated.Description != "Updated description" {
		t.Errorf("expected description 'Updated description', got %q", updated.Description)
	}
}

func TestActorHandlerUpdateNotFound(t *testing.T) {
	handler, _, cleanup := setupActorHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で更新（有効なフォーマット）
	err := handler.Update(ctx, "actor-00000000", map[string]any{"title": "Test"})
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestActorHandlerDelete(t *testing.T) {
	handler, _, cleanup := setupActorHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクター追加
	result, err := handler.Add(ctx, "Delete Test Actor")
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
		t.Error("expected error for deleted actor")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestActorHandlerDeleteNotFound(t *testing.T) {
	handler, _, cleanup := setupActorHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で削除（有効なフォーマット）
	err := handler.Delete(ctx, "actor-00000000")
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestActorHandlerContextCancellation(t *testing.T) {
	handler, _, cleanup := setupActorHandlerTest(t)
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
	_, err = handler.Get(ctx, "actor-00000000")
	if err == nil {
		t.Error("Get should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Update（有効なフォーマット）
	err = handler.Update(ctx, "actor-00000000", map[string]any{})
	if err == nil {
		t.Error("Update should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Delete（有効なフォーマット）
	err = handler.Delete(ctx, "actor-00000000")
	if err == nil {
		t.Error("Delete should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestGenerateActorIDFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-actor-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	zeusPath := tmpDir + "/.zeus"
	fs := yaml.NewFileManager(zeusPath)
	handler := NewActorHandler(fs)

	// ID 生成テスト
	id := handler.generateActorID()

	// プレフィックスが正しいか
	if len(id) < 6 || id[:6] != "actor-" {
		t.Errorf("expected ID to start with 'actor-', got %q", id)
	}

	// 長さが正しいか (actor- + 8文字)
	if len(id) != 14 {
		t.Errorf("expected ID length to be 14, got %d", len(id))
	}
}

func TestActorEntityValidate(t *testing.T) {
	tests := []struct {
		name    string
		actor   ActorEntity
		wantErr bool
	}{
		{
			name: "valid actor",
			actor: ActorEntity{
				ID:    "actor-12345678",
				Title: "Test Actor",
				Type:  ActorTypeHuman,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			actor: ActorEntity{
				Title: "Test Actor",
				Type:  ActorTypeHuman,
			},
			wantErr: true,
		},
		{
			name: "missing title",
			actor: ActorEntity{
				ID:   "actor-12345678",
				Type: ActorTypeHuman,
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			actor: ActorEntity{
				ID:    "actor-12345678",
				Title: "Test Actor",
				Type:  "invalid",
			},
			wantErr: true,
		},
		{
			name: "empty type (defaults to human)",
			actor: ActorEntity{
				ID:    "actor-12345678",
				Title: "Test Actor",
				Type:  "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.actor.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
