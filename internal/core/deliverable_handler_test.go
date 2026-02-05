package core

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// テスト用のセットアップ（Deliverable + Objective ハンドラー）
func setupDeliverableHandlerTest(t *testing.T) (*DeliverableHandler, *ObjectiveHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-deliverable-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath+"/deliverables", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create deliverables dir: %v", err)
	}
	if err := os.MkdirAll(zeusPath+"/objectives", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create objectives dir: %v", err)
	}

	fs := yaml.NewFileManager(zeusPath)
	objHandler := NewObjectiveHandler(fs, nil)
	delHandler := NewDeliverableHandler(fs, objHandler, nil)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return delHandler, objHandler, zeusPath, cleanup
}

func TestDeliverableHandlerType(t *testing.T) {
	handler, _, _, cleanup := setupDeliverableHandlerTest(t)
	defer cleanup()

	if handler.Type() != "deliverable" {
		t.Errorf("expected type 'deliverable', got %q", handler.Type())
	}
}

func TestDeliverableHandlerAdd(t *testing.T) {
	handler, objHandler, _, cleanup := setupDeliverableHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 先に Objective を作成（Deliverable は objective_id が必須）
	objResult, err := objHandler.Add(ctx, "認証システム実装")
	if err != nil {
		t.Fatalf("Add objective failed: %v", err)
	}

	// Deliverable 追加
	result, err := handler.Add(ctx, "API設計書",
		WithDeliverableObjective(objResult.ID),
	)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if !result.Success {
		t.Error("Add should succeed")
	}

	if result.Entity != "deliverable" {
		t.Errorf("expected entity 'deliverable', got %q", result.Entity)
	}

	if result.ID == "" {
		t.Error("ID should not be empty")
	}

	// ID フォーマット確認 (del-NNN)
	if len(result.ID) < 4 || result.ID[:4] != "del-" {
		t.Errorf("expected ID to start with 'del-', got %q", result.ID)
	}

	// リストで確認
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 1 {
		t.Errorf("expected 1 deliverable, got %d", listResult.Total)
	}
}

func TestDeliverableHandlerAddWithOptions(t *testing.T) {
	handler, objHandler, _, cleanup := setupDeliverableHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 先に Objective を作成
	objResult, err := objHandler.Add(ctx, "認証システム実装")
	if err != nil {
		t.Fatalf("Add objective failed: %v", err)
	}

	// オプション付きで Deliverable 追加
	result, err := handler.Add(ctx, "API設計書",
		WithDeliverableDescription("REST API の設計書"),
		WithDeliverableObjective(objResult.ID),
		WithDeliverableFormat(DeliverableFormatDocument),
		WithDeliverableStatus(DeliverableStatusInProgress),
		WithDeliverableAcceptanceCriteria([]string{"全エンドポイントを網羅", "レスポンス形式を定義"}),
		WithDeliverableOwner("test-user"),
		WithDeliverableTags([]string{"docs", "api"}),
	)
	if err != nil {
		t.Fatalf("Add with options failed: %v", err)
	}

	// Deliverable を取得して確認
	delAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	del := delAny.(*DeliverableEntity)
	if del.Description != "REST API の設計書" {
		t.Errorf("expected description 'REST API の設計書', got %q", del.Description)
	}

	if del.ObjectiveID != objResult.ID {
		t.Errorf("expected objective_id %q, got %q", objResult.ID, del.ObjectiveID)
	}

	if del.Format != DeliverableFormatDocument {
		t.Errorf("expected format 'document', got %q", del.Format)
	}

	if del.Status != DeliverableStatusInProgress {
		t.Errorf("expected status 'in_progress', got %q", del.Status)
	}

	if len(del.AcceptanceCriteria) != 2 {
		t.Errorf("expected 2 acceptance criteria, got %d", len(del.AcceptanceCriteria))
	}

	if del.Metadata.Owner != "test-user" {
		t.Errorf("expected owner 'test-user', got %q", del.Metadata.Owner)
	}

	if len(del.Metadata.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(del.Metadata.Tags))
	}
}

func TestDeliverableHandlerAddWithInvalidObjective(t *testing.T) {
	handler, _, _, cleanup := setupDeliverableHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない Objective ID で追加
	_, err := handler.Add(ctx, "API設計書",
		WithDeliverableObjective("obj-999"),
	)

	if err == nil {
		t.Error("expected error for invalid objective")
	}
}

func TestDeliverableHandlerList(t *testing.T) {
	handler, objHandler, _, cleanup := setupDeliverableHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 先に Objective を作成
	objResult, err := objHandler.Add(ctx, "テスト Objective")
	if err != nil {
		t.Fatalf("Add objective failed: %v", err)
	}

	// 複数 Deliverable を追加
	for i := range 5 {
		_, err := handler.Add(ctx, "Deliverable",
			WithDeliverableObjective(objResult.ID),
		)
		if err != nil {
			t.Fatalf("Add[%d] failed: %v", i, err)
		}
	}

	// 全リスト
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 5 {
		t.Errorf("expected 5 deliverables, got %d", listResult.Total)
	}
}

func TestDeliverableHandlerListWithFilter(t *testing.T) {
	handler, objHandler, _, cleanup := setupDeliverableHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 先に Objective を作成
	objResult, err := objHandler.Add(ctx, "テスト Objective")
	if err != nil {
		t.Fatalf("Add objective failed: %v", err)
	}

	// 異なるステータスの Deliverable を追加
	_, err = handler.Add(ctx, "Planned",
		WithDeliverableObjective(objResult.ID),
		WithDeliverableStatus(DeliverableStatusPlanned),
	)
	if err != nil {
		t.Fatalf("Add planned failed: %v", err)
	}

	_, err = handler.Add(ctx, "In Progress",
		WithDeliverableObjective(objResult.ID),
		WithDeliverableStatus(DeliverableStatusInProgress),
	)
	if err != nil {
		t.Fatalf("Add in_progress failed: %v", err)
	}

	_, err = handler.Add(ctx, "Completed",
		WithDeliverableObjective(objResult.ID),
		WithDeliverableStatus(DeliverableStatusCompleted),
	)
	if err != nil {
		t.Fatalf("Add completed failed: %v", err)
	}

	// in_progress でフィルタ
	filter := &ListFilter{Status: string(DeliverableStatusInProgress)}
	listResult, err := handler.List(ctx, filter)
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}

	if listResult.Total != 1 {
		t.Errorf("expected 1 in_progress deliverable, got %d", listResult.Total)
	}
}

func TestDeliverableHandlerGet(t *testing.T) {
	handler, objHandler, _, cleanup := setupDeliverableHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 先に Objective を作成
	objResult, err := objHandler.Add(ctx, "テスト Objective")
	if err != nil {
		t.Fatalf("Add objective failed: %v", err)
	}

	// Deliverable 追加
	result, err := handler.Add(ctx, "Get Test Deliverable",
		WithDeliverableObjective(objResult.ID),
	)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Deliverable を取得
	delAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	del := delAny.(*DeliverableEntity)
	if del.Title != "Get Test Deliverable" {
		t.Errorf("expected title 'Get Test Deliverable', got %q", del.Title)
	}
}

func TestDeliverableHandlerGetNotFound(t *testing.T) {
	handler, _, _, cleanup := setupDeliverableHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で取得
	_, err := handler.Get(ctx, "del-999")
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestDeliverableHandlerUpdate(t *testing.T) {
	handler, objHandler, _, cleanup := setupDeliverableHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 先に Objective を作成
	objResult, err := objHandler.Add(ctx, "テスト Objective")
	if err != nil {
		t.Fatalf("Add objective failed: %v", err)
	}

	// Deliverable 追加
	result, err := handler.Add(ctx, "Update Test Deliverable",
		WithDeliverableObjective(objResult.ID),
	)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Deliverable を取得
	delAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	del := delAny.(*DeliverableEntity)

	// 更新
	del.Title = "Updated Title"
	del.Status = DeliverableStatusCompleted
	err = handler.Update(ctx, result.ID, del)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// 更新を確認
	updatedAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}

	updated := updatedAny.(*DeliverableEntity)
	if updated.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", updated.Title)
	}

	if updated.Status != DeliverableStatusCompleted {
		t.Errorf("expected status 'completed', got %q", updated.Status)
	}
}

func TestDeliverableHandlerUpdateNotFound(t *testing.T) {
	handler, _, _, cleanup := setupDeliverableHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で更新
	del := &DeliverableEntity{ID: "del-999", Title: "Test"}
	err := handler.Update(ctx, "del-999", del)
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestDeliverableHandlerDelete(t *testing.T) {
	handler, objHandler, _, cleanup := setupDeliverableHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 先に Objective を作成
	objResult, err := objHandler.Add(ctx, "テスト Objective")
	if err != nil {
		t.Fatalf("Add objective failed: %v", err)
	}

	// Deliverable 追加
	result, err := handler.Add(ctx, "Delete Test Deliverable",
		WithDeliverableObjective(objResult.ID),
	)
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
		t.Error("expected error for deleted deliverable")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestDeliverableHandlerDeleteNotFound(t *testing.T) {
	handler, _, _, cleanup := setupDeliverableHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で削除
	err := handler.Delete(ctx, "del-999")
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestDeliverableHandlerContextCancellation(t *testing.T) {
	handler, _, _, cleanup := setupDeliverableHandlerTest(t)
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
	_, err = handler.Get(ctx, "del-001")
	if err == nil {
		t.Error("Get should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Update
	err = handler.Update(ctx, "del-001", &DeliverableEntity{})
	if err == nil {
		t.Error("Update should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Delete
	err = handler.Delete(ctx, "del-001")
	if err == nil {
		t.Error("Delete should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestDeliverableHandlerGetByObjective(t *testing.T) {
	handler, objHandler, _, cleanup := setupDeliverableHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Objective を作成
	objResult, err := objHandler.Add(ctx, "認証システム実装")
	if err != nil {
		t.Fatalf("Add objective failed: %v", err)
	}

	// 同じ Objective に紐づく Deliverable を 3 つ作成
	for i := range 3 {
		_, err := handler.Add(ctx, "Deliverable",
			WithDeliverableObjective(objResult.ID),
		)
		if err != nil {
			t.Fatalf("Add deliverable[%d] failed: %v", i, err)
		}
	}

	// 別の Objective を作成して、そこに紐づく Deliverable を 2 つ作成
	objResult2, err := objHandler.Add(ctx, "別の Objective")
	if err != nil {
		t.Fatalf("Add objective2 failed: %v", err)
	}
	for i := range 2 {
		_, err := handler.Add(ctx, "Other Deliverable",
			WithDeliverableObjective(objResult2.ID),
		)
		if err != nil {
			t.Fatalf("Add other deliverable[%d] failed: %v", i, err)
		}
	}

	// Objective に紐づく Deliverable を取得
	deliverables, err := handler.GetDeliverablesByObjective(ctx, objResult.ID)
	if err != nil {
		t.Fatalf("GetDeliverablesByObjective failed: %v", err)
	}

	if len(deliverables) != 3 {
		t.Errorf("expected 3 deliverables for objective, got %d", len(deliverables))
	}
}

func TestDeliverableHandlerFormats(t *testing.T) {
	handler, objHandler, _, cleanup := setupDeliverableHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 先に Objective を作成
	objResult, err := objHandler.Add(ctx, "テスト Objective")
	if err != nil {
		t.Fatalf("Add objective failed: %v", err)
	}

	testCases := []struct {
		name   string
		format DeliverableFormat
	}{
		{"Document", DeliverableFormatDocument},
		{"Code", DeliverableFormatCode},
		{"Data", DeliverableFormatData},
		{"Design", DeliverableFormatDesign},
		{"Presentation", DeliverableFormatPresentation},
		{"Other", DeliverableFormatOther},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := handler.Add(ctx, tc.name,
				WithDeliverableObjective(objResult.ID),
				WithDeliverableFormat(tc.format),
			)
			if err != nil {
				t.Fatalf("Add failed: %v", err)
			}

			delAny, err := handler.Get(ctx, result.ID)
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}

			del := delAny.(*DeliverableEntity)
			if del.Format != tc.format {
				t.Errorf("expected format %q, got %q", tc.format, del.Format)
			}
		})
	}
}

func TestDeliverableHandlerIDSequence(t *testing.T) {
	handler, objHandler, _, cleanup := setupDeliverableHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 先に Objective を作成
	objResult, err := objHandler.Add(ctx, "テスト Objective")
	if err != nil {
		t.Fatalf("Add objective failed: %v", err)
	}

	// 複数 Deliverable を追加
	ids := make([]string, 3)
	for i := range 3 {
		result, err := handler.Add(ctx, "Deliverable",
			WithDeliverableObjective(objResult.ID),
		)
		if err != nil {
			t.Fatalf("Add[%d] failed: %v", i, err)
		}
		ids[i] = result.ID
	}

	// 全ての ID がユニークであることを確認
	seen := make(map[string]bool)
	for i, id := range ids {
		if seen[id] {
			t.Errorf("duplicate ID found: %q", id)
		}
		seen[id] = true

		// プレフィックスが正しいことを確認
		if !strings.HasPrefix(id, "del-") {
			t.Errorf("ID[%d] = %q, expected prefix 'del-'", i, id)
		}
	}

	// ID 数が正しいことを確認
	if len(seen) != 3 {
		t.Errorf("expected 3 unique IDs, got %d", len(seen))
	}
}
