package core

import (
	"context"
	"os"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// テスト用のセットアップ
func setupUseCaseHandlerTest(t *testing.T) (*UseCaseHandler, *ActorHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-usecase-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath+"/usecases", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create zeus dir: %v", err)
	}

	fs := yaml.NewFileManager(zeusPath)
	actorHandler := NewActorHandler(fs)
	usecaseHandler := NewUseCaseHandler(fs, nil, actorHandler, nil)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return usecaseHandler, actorHandler, zeusPath, cleanup
}

func TestUseCaseHandlerType(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	if handler.Type() != "usecase" {
		t.Errorf("expected type 'usecase', got %q", handler.Type())
	}
}

func TestUseCaseHandlerAdd(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// ユースケース追加（ObjectiveID は必須）
	result, err := handler.Add(ctx, "Test UseCase", WithUseCaseObjective("obj-test0001"))
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if !result.Success {
		t.Error("Add should succeed")
	}

	if result.Entity != "usecase" {
		t.Errorf("expected entity 'usecase', got %q", result.Entity)
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
		t.Errorf("expected 1 usecase, got %d", listResult.Total)
	}
}

func TestUseCaseHandlerAddWithOptions(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// オプション付きでユースケース追加（ObjectiveID は必須）
	result, err := handler.Add(ctx, "Test UseCase with Options",
		WithUseCaseObjective("obj-test0001"),
		WithUseCaseDescription("This is a test usecase"),
		WithUseCaseStatus(UseCaseStatusActive),
		WithUseCaseOwner("test-user"),
		WithUseCaseTags([]string{"auth", "security"}),
	)
	if err != nil {
		t.Fatalf("Add with options failed: %v", err)
	}

	// ユースケースを取得して確認
	usecaseAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	usecase := usecaseAny.(*UseCaseEntity)
	if usecase.Description != "This is a test usecase" {
		t.Errorf("expected description 'This is a test usecase', got %q", usecase.Description)
	}

	if usecase.Status != UseCaseStatusActive {
		t.Errorf("expected status 'active', got %q", usecase.Status)
	}

	if usecase.Metadata.Owner != "test-user" {
		t.Errorf("expected owner 'test-user', got %q", usecase.Metadata.Owner)
	}

	if len(usecase.Metadata.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(usecase.Metadata.Tags))
	}
}

func TestUseCaseHandlerAddAllStatuses(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name   string
		status UseCaseStatus
	}{
		{"Draft UseCase", UseCaseStatusDraft},
		{"Active UseCase", UseCaseStatusActive},
		{"Deprecated UseCase", UseCaseStatusDeprecated},
	}

	for _, tt := range tests {
		result, err := handler.Add(ctx, tt.name, WithUseCaseObjective("obj-test0001"), WithUseCaseStatus(tt.status))
		if err != nil {
			t.Fatalf("Add %s failed: %v", tt.name, err)
		}

		usecaseAny, err := handler.Get(ctx, result.ID)
		if err != nil {
			t.Fatalf("Get %s failed: %v", tt.name, err)
		}

		usecase := usecaseAny.(*UseCaseEntity)
		if usecase.Status != tt.status {
			t.Errorf("expected status %q, got %q", tt.status, usecase.Status)
		}
	}

	// 全ユースケースがリストで確認できる
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 3 {
		t.Errorf("expected 3 usecases, got %d", listResult.Total)
	}
}

func TestUseCaseHandlerList(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 複数ユースケースを追加
	for range 5 {
		_, err := handler.Add(ctx, "UseCase", WithUseCaseObjective("obj-test0001"))
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
		t.Errorf("expected 5 usecases, got %d", listResult.Total)
	}
}

func TestUseCaseHandlerListEmpty(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 空のリスト
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 0 {
		t.Errorf("expected 0 usecases, got %d", listResult.Total)
	}
}

func TestUseCaseHandlerGet(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// ユースケース追加
	result, err := handler.Add(ctx, "Get Test UseCase", WithUseCaseObjective("obj-test0001"))
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// ユースケースを取得
	usecaseAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	usecase := usecaseAny.(*UseCaseEntity)
	if usecase.Title != "Get Test UseCase" {
		t.Errorf("expected title 'Get Test UseCase', got %q", usecase.Title)
	}
}

func TestUseCaseHandlerGetNotFound(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で取得（有効なフォーマット）
	_, err := handler.Get(ctx, "uc-00000000")
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestUseCaseHandlerUpdate(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// ユースケース追加
	result, err := handler.Add(ctx, "Update Test UseCase", WithUseCaseObjective("obj-test0001"))
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 更新
	updateData := map[string]any{
		"title":       "Updated Title",
		"status":      "active",
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

	updated := updatedAny.(*UseCaseEntity)
	if updated.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", updated.Title)
	}

	if updated.Status != UseCaseStatusActive {
		t.Errorf("expected status 'active', got %q", updated.Status)
	}

	if updated.Description != "Updated description" {
		t.Errorf("expected description 'Updated description', got %q", updated.Description)
	}
}

func TestUseCaseHandlerUpdateNotFound(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で更新（有効なフォーマット）
	err := handler.Update(ctx, "uc-00000000", map[string]any{"title": "Test"})
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestUseCaseHandlerDelete(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// ユースケース追加
	result, err := handler.Add(ctx, "Delete Test UseCase", WithUseCaseObjective("obj-test0001"))
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
		t.Error("expected error for deleted usecase")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestUseCaseHandlerDeleteNotFound(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で削除（有効なフォーマット）
	err := handler.Delete(ctx, "uc-00000000")
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestUseCaseHandlerAddRelation(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 2つのユースケースを追加
	result1, err := handler.Add(ctx, "Login UseCase", WithUseCaseObjective("obj-test0001"))
	if err != nil {
		t.Fatalf("Add UseCase1 failed: %v", err)
	}

	result2, err := handler.Add(ctx, "Authenticate UseCase", WithUseCaseObjective("obj-test0001"))
	if err != nil {
		t.Fatalf("Add UseCase2 failed: %v", err)
	}

	// 関係を追加 (Login includes Authenticate)
	rel := UseCaseRelation{
		Type:     RelationTypeInclude,
		TargetID: result2.ID,
	}
	err = handler.AddRelation(ctx, result1.ID, rel)
	if err != nil {
		t.Fatalf("AddRelation failed: %v", err)
	}

	// 関係を確認
	usecaseAny, err := handler.Get(ctx, result1.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	usecase := usecaseAny.(*UseCaseEntity)
	if len(usecase.Relations) != 1 {
		t.Errorf("expected 1 relation, got %d", len(usecase.Relations))
	}

	if usecase.Relations[0].Type != RelationTypeInclude {
		t.Errorf("expected relation type 'include', got %q", usecase.Relations[0].Type)
	}

	if usecase.Relations[0].TargetID != result2.ID {
		t.Errorf("expected target ID %q, got %q", result2.ID, usecase.Relations[0].TargetID)
	}
}

func TestUseCaseHandlerAddRelationExtend(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 2つのユースケースを追加
	result1, err := handler.Add(ctx, "Base UseCase", WithUseCaseObjective("obj-test0001"))
	if err != nil {
		t.Fatalf("Add UseCase1 failed: %v", err)
	}

	result2, err := handler.Add(ctx, "Extended UseCase", WithUseCaseObjective("obj-test0001"))
	if err != nil {
		t.Fatalf("Add UseCase2 failed: %v", err)
	}

	// extend 関係を追加
	rel := UseCaseRelation{
		Type:           RelationTypeExtend,
		TargetID:       result1.ID,
		ExtensionPoint: "additional_validation",
		Condition:      "when validation fails",
	}
	err = handler.AddRelation(ctx, result2.ID, rel)
	if err != nil {
		t.Fatalf("AddRelation failed: %v", err)
	}

	// 関係を確認
	usecaseAny, err := handler.Get(ctx, result2.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	usecase := usecaseAny.(*UseCaseEntity)
	if len(usecase.Relations) != 1 {
		t.Errorf("expected 1 relation, got %d", len(usecase.Relations))
	}

	if usecase.Relations[0].Type != RelationTypeExtend {
		t.Errorf("expected relation type 'extend', got %q", usecase.Relations[0].Type)
	}

	if usecase.Relations[0].ExtensionPoint != "additional_validation" {
		t.Errorf("expected extension point 'additional_validation', got %q", usecase.Relations[0].ExtensionPoint)
	}

	if usecase.Relations[0].Condition != "when validation fails" {
		t.Errorf("expected condition 'when validation fails', got %q", usecase.Relations[0].Condition)
	}
}

func TestUseCaseHandlerAddRelationTargetNotFound(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// ユースケースを追加
	result, err := handler.Add(ctx, "Test UseCase", WithUseCaseObjective("obj-test0001"))
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 存在しないターゲットへの関係を追加（有効なフォーマット）
	rel := UseCaseRelation{
		Type:     RelationTypeInclude,
		TargetID: "uc-00000000",
	}
	err = handler.AddRelation(ctx, result.ID, rel)
	if err == nil {
		t.Error("expected error for non-existent target")
	}
}

func TestUseCaseHandlerAddActor(t *testing.T) {
	handler, actorHandler, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクターを追加
	actorResult, err := actorHandler.Add(ctx, "User", WithActorType(ActorTypeHuman))
	if err != nil {
		t.Fatalf("Add Actor failed: %v", err)
	}

	// ユースケースを追加
	usecaseResult, err := handler.Add(ctx, "Login UseCase", WithUseCaseObjective("obj-test0001"))
	if err != nil {
		t.Fatalf("Add UseCase failed: %v", err)
	}

	// アクター参照を追加
	actorRef := UseCaseActorRef{
		ActorID: actorResult.ID,
		Role:    ActorRolePrimary,
	}
	err = handler.AddActor(ctx, usecaseResult.ID, actorRef)
	if err != nil {
		t.Fatalf("AddActor failed: %v", err)
	}

	// 確認
	usecaseAny, err := handler.Get(ctx, usecaseResult.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	usecase := usecaseAny.(*UseCaseEntity)
	if len(usecase.Actors) != 1 {
		t.Errorf("expected 1 actor, got %d", len(usecase.Actors))
	}

	if usecase.Actors[0].ActorID != actorResult.ID {
		t.Errorf("expected actor ID %q, got %q", actorResult.ID, usecase.Actors[0].ActorID)
	}

	if usecase.Actors[0].Role != ActorRolePrimary {
		t.Errorf("expected role 'primary', got %q", usecase.Actors[0].Role)
	}
}

func TestUseCaseHandlerAddActorDuplicate(t *testing.T) {
	handler, actorHandler, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// アクターを追加
	actorResult, err := actorHandler.Add(ctx, "User", WithActorType(ActorTypeHuman))
	if err != nil {
		t.Fatalf("Add Actor failed: %v", err)
	}

	// ユースケースを追加
	usecaseResult, err := handler.Add(ctx, "Login UseCase", WithUseCaseObjective("obj-test0001"))
	if err != nil {
		t.Fatalf("Add UseCase failed: %v", err)
	}

	// アクター参照を追加（1回目）
	actorRef := UseCaseActorRef{
		ActorID: actorResult.ID,
		Role:    ActorRolePrimary,
	}
	err = handler.AddActor(ctx, usecaseResult.ID, actorRef)
	if err != nil {
		t.Fatalf("AddActor (first) failed: %v", err)
	}

	// 同じアクターを再度追加（エラーになるはず）
	err = handler.AddActor(ctx, usecaseResult.ID, actorRef)
	if err == nil {
		t.Error("expected error for duplicate actor")
	}
}

func TestUseCaseHandlerAddActorNotFound(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// ユースケースを追加（ObjectiveID は必須）
	usecaseResult, err := handler.Add(ctx, "Test UseCase", WithUseCaseObjective("obj-test0001"))
	if err != nil {
		t.Fatalf("Add UseCase failed: %v", err)
	}

	// 存在しないアクターを追加（有効なフォーマット）
	actorRef := UseCaseActorRef{
		ActorID: "actor-00000000",
		Role:    ActorRolePrimary,
	}
	err = handler.AddActor(ctx, usecaseResult.ID, actorRef)
	if err == nil {
		t.Error("expected error for non-existent actor")
	}
}

func TestUseCaseHandlerContextCancellation(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
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
	_, err = handler.Get(ctx, "uc-00000000")
	if err == nil {
		t.Error("Get should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Update（有効なフォーマット）
	err = handler.Update(ctx, "uc-00000000", map[string]any{})
	if err == nil {
		t.Error("Update should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Delete（有効なフォーマット）
	err = handler.Delete(ctx, "uc-00000000")
	if err == nil {
		t.Error("Delete should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestGenerateUseCaseIDFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-usecase-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	zeusPath := tmpDir + "/.zeus"
	fs := yaml.NewFileManager(zeusPath)
	handler := NewUseCaseHandler(fs, nil, nil, nil)

	// ID 生成テスト
	id := handler.generateUseCaseID()

	// プレフィックスが正しいか
	if len(id) < 3 || id[:3] != "uc-" {
		t.Errorf("expected ID to start with 'uc-', got %q", id)
	}

	// 長さが正しいか (uc- + 8文字)
	if len(id) != 11 {
		t.Errorf("expected ID length to be 11, got %d", len(id))
	}
}

func TestUseCaseEntityValidate(t *testing.T) {
	tests := []struct {
		name    string
		usecase UseCaseEntity
		wantErr bool
	}{
		{
			name: "valid usecase",
			usecase: UseCaseEntity{
				ID:          "uc-12345678",
				Title:       "Test UseCase",
				ObjectiveID: "obj-test0001",
				Status:      UseCaseStatusDraft,
			},
			wantErr: false,
		},
		{
			name: "missing objective_id",
			usecase: UseCaseEntity{
				ID:     "uc-12345678",
				Title:  "Test UseCase",
				Status: UseCaseStatusDraft,
			},
			wantErr: true,
		},
		{
			name: "missing ID",
			usecase: UseCaseEntity{
				Title:  "Test UseCase",
				Status: UseCaseStatusDraft,
			},
			wantErr: true,
		},
		{
			name: "missing title",
			usecase: UseCaseEntity{
				ID:     "uc-12345678",
				Status: UseCaseStatusDraft,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			usecase: UseCaseEntity{
				ID:     "uc-12345678",
				Title:  "Test UseCase",
				Status: "invalid",
			},
			wantErr: true,
		},
		{
			name: "empty status (defaults to draft)",
			usecase: UseCaseEntity{
				ID:          "uc-12345678",
				Title:       "Test UseCase",
				ObjectiveID: "obj-test0001",
				Status:      "",
			},
			wantErr: false,
		},
		{
			name: "valid with actors",
			usecase: UseCaseEntity{
				ID:          "uc-12345678",
				Title:       "Test UseCase",
				ObjectiveID: "obj-test0001",
				Status:      UseCaseStatusDraft,
				Actors: []UseCaseActorRef{
					{ActorID: "actor-12345678", Role: ActorRolePrimary},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid actor role",
			usecase: UseCaseEntity{
				ID:     "uc-12345678",
				Title:  "Test UseCase",
				Status: UseCaseStatusDraft,
				Actors: []UseCaseActorRef{
					{ActorID: "actor-12345678", Role: "invalid"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing actor_id",
			usecase: UseCaseEntity{
				ID:     "uc-12345678",
				Title:  "Test UseCase",
				Status: UseCaseStatusDraft,
				Actors: []UseCaseActorRef{
					{ActorID: "", Role: ActorRolePrimary},
				},
			},
			wantErr: true,
		},
		{
			name: "valid with relations",
			usecase: UseCaseEntity{
				ID:          "uc-12345678",
				Title:       "Test UseCase",
				ObjectiveID: "obj-test0001",
				Status:      UseCaseStatusDraft,
				Relations: []UseCaseRelation{
					{Type: RelationTypeInclude, TargetID: "uc-target123"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid relation type",
			usecase: UseCaseEntity{
				ID:     "uc-12345678",
				Title:  "Test UseCase",
				Status: UseCaseStatusDraft,
				Relations: []UseCaseRelation{
					{Type: "invalid", TargetID: "uc-target123"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing relation target_id",
			usecase: UseCaseEntity{
				ID:     "uc-12345678",
				Title:  "Test UseCase",
				Status: UseCaseStatusDraft,
				Relations: []UseCaseRelation{
					{Type: RelationTypeInclude, TargetID: ""},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.usecase.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUseCaseHandlerWithScenario(t *testing.T) {
	handler, _, _, cleanup := setupUseCaseHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// シナリオ付きでユースケース追加（ObjectiveID は必須）
	mainFlow := []string{
		"1. ユーザーがログイン画面を開く",
		"2. ユーザーが認証情報を入力する",
		"3. システムが認証を検証する",
		"4. システムがセッションを作成する",
	}
	result, err := handler.Add(ctx, "Login UseCase",
		WithUseCaseObjective("obj-test0001"),
		WithUseCaseScenario(mainFlow),
	)
	if err != nil {
		t.Fatalf("Add with scenario failed: %v", err)
	}

	// 確認
	usecaseAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	usecase := usecaseAny.(*UseCaseEntity)
	if len(usecase.Scenario.MainFlow) != 4 {
		t.Errorf("expected 4 main flow steps, got %d", len(usecase.Scenario.MainFlow))
	}

	if usecase.Scenario.MainFlow[0] != "1. ユーザーがログイン画面を開く" {
		t.Errorf("unexpected first step: %q", usecase.Scenario.MainFlow[0])
	}
}
