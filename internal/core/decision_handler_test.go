package core

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// テスト用のセットアップ
func setupDecisionHandlerTest(t *testing.T) (*DecisionHandler, *ConsiderationHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-decision-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath+"/decisions", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create decisions dir: %v", err)
	}
	if err := os.MkdirAll(zeusPath+"/considerations", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create considerations dir: %v", err)
	}

	fs := yaml.NewFileManager(zeusPath)
	conHandler := NewConsiderationHandler(fs, nil, nil, nil)
	handler := NewDecisionHandler(fs, conHandler, nil)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return handler, conHandler, zeusPath, cleanup
}

// createTestConsideration はテスト用の Consideration を作成
func createTestConsideration(t *testing.T, conHandler *ConsiderationHandler) string {
	t.Helper()

	ctx := context.Background()
	result, err := conHandler.Add(ctx, "テスト検討事項",
		WithConsiderationOptions([]ConsiderationOption{
			{ID: "opt-1", Title: "オプション1", Description: "説明1"},
			{ID: "opt-2", Title: "オプション2", Description: "説明2"},
		}),
	)
	if err != nil {
		t.Fatalf("failed to create test consideration: %v", err)
	}
	return result.ID
}

func TestDecisionHandler_Type(t *testing.T) {
	handler, _, _, cleanup := setupDecisionHandlerTest(t)
	defer cleanup()

	if handler.Type() != "decision" {
		t.Errorf("expected type 'decision', got %q", handler.Type())
	}
}

func TestDecisionHandler_Add(t *testing.T) {
	handler, conHandler, _, cleanup := setupDecisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Consideration を先に作成
	conID := createTestConsideration(t, conHandler)

	// Decision 追加
	result, err := handler.Add(ctx, "テスト決定",
		WithDecisionConsideration(conID),
		WithDecisionSelected(SelectedOption{OptionID: "opt-1", Title: "オプション1"}),
		WithDecisionRationale("テスト理由"),
	)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if !result.Success {
		t.Error("Add should succeed")
	}

	if result.Entity != "decision" {
		t.Errorf("expected entity 'decision', got %q", result.Entity)
	}

	// ID フォーマット確認 (dec-NNN)
	if !strings.HasPrefix(result.ID, "dec-") {
		t.Errorf("expected ID to start with 'dec-', got %q", result.ID)
	}
}

func TestDecisionHandler_Add_RequiresConsideration(t *testing.T) {
	handler, _, _, cleanup := setupDecisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// ConsiderationID なしで Decision を作成しようとする
	_, err := handler.Add(ctx, "テスト決定")
	if err == nil {
		t.Error("expected error when creating decision without consideration")
		return
	}

	if !strings.Contains(err.Error(), "consideration_id is required") {
		t.Errorf("expected error about consideration_id being required, got: %v", err)
	}
}

func TestDecisionHandler_Add_RequiresExistingConsideration(t *testing.T) {
	handler, _, _, cleanup := setupDecisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない Consideration を参照して Decision を作成
	_, err := handler.Add(ctx, "テスト決定",
		WithDecisionConsideration("con-999"),
		WithDecisionSelected(SelectedOption{OptionID: "opt-1", Title: "オプション1"}),
		WithDecisionRationale("テスト理由"),
	)
	if err == nil {
		t.Error("expected error when referencing non-existent consideration")
		return
	}

	if !strings.Contains(err.Error(), "referenced consideration not found") {
		t.Errorf("expected error about consideration not found, got: %v", err)
	}
}

func TestDecisionHandler_List(t *testing.T) {
	handler, conHandler, _, cleanup := setupDecisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 複数の Decision を作成
	for i := 0; i < 3; i++ {
		conID := createTestConsideration(t, conHandler)
		_, err := handler.Add(ctx, "テスト決定",
			WithDecisionConsideration(conID),
			WithDecisionSelected(SelectedOption{OptionID: "opt-1", Title: "オプション1"}),
			WithDecisionRationale("テスト理由"),
		)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// リスト取得
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 3 {
		t.Errorf("expected 3 decisions, got %d", listResult.Total)
	}
}

func TestDecisionHandler_List_Empty(t *testing.T) {
	handler, _, _, cleanup := setupDecisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 空のリスト取得
	listResult, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResult.Total != 0 {
		t.Errorf("expected 0 decisions, got %d", listResult.Total)
	}
}

func TestDecisionHandler_List_WithLimit(t *testing.T) {
	handler, conHandler, _, cleanup := setupDecisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 5つの Decision を作成
	for i := 0; i < 5; i++ {
		conID := createTestConsideration(t, conHandler)
		_, err := handler.Add(ctx, "テスト決定",
			WithDecisionConsideration(conID),
			WithDecisionSelected(SelectedOption{OptionID: "opt-1", Title: "オプション1"}),
			WithDecisionRationale("テスト理由"),
		)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Limit 付きでリスト取得
	filter := &ListFilter{Limit: 3}
	listResult, err := handler.List(ctx, filter)
	if err != nil {
		t.Fatalf("List with limit failed: %v", err)
	}

	if listResult.Total != 3 {
		t.Errorf("expected 3 decisions with limit, got %d", listResult.Total)
	}
}

func TestDecisionHandler_Get(t *testing.T) {
	handler, conHandler, _, cleanup := setupDecisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Consideration と Decision を作成
	conID := createTestConsideration(t, conHandler)
	addResult, err := handler.Add(ctx, "テスト決定",
		WithDecisionConsideration(conID),
		WithDecisionSelected(SelectedOption{OptionID: "opt-1", Title: "オプション1"}),
		WithDecisionRationale("テスト理由"),
		WithDecisionImpact([]string{"影響1", "影響2"}),
		WithDecisionDecidedBy("test-user"),
	)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Decision を取得
	decAny, err := handler.Get(ctx, addResult.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	dec := decAny.(*DecisionEntity)
	if dec.Title != "テスト決定" {
		t.Errorf("expected title 'テスト決定', got %q", dec.Title)
	}
	if dec.ConsiderationID != conID {
		t.Errorf("expected consideration_id %q, got %q", conID, dec.ConsiderationID)
	}
	if dec.Rationale != "テスト理由" {
		t.Errorf("expected rationale 'テスト理由', got %q", dec.Rationale)
	}
	if len(dec.Impact) != 2 {
		t.Errorf("expected 2 impacts, got %d", len(dec.Impact))
	}
	if dec.DecidedBy != "test-user" {
		t.Errorf("expected decided_by 'test-user', got %q", dec.DecidedBy)
	}
}

func TestDecisionHandler_Get_NotFound(t *testing.T) {
	handler, _, _, cleanup := setupDecisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で取得
	_, err := handler.Get(ctx, "dec-999")
	if err == nil {
		t.Error("expected error for non-existent ID")
		return
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestDecisionHandler_Get_InvalidID(t *testing.T) {
	handler, _, _, cleanup := setupDecisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 無効な ID で取得
	_, err := handler.Get(ctx, "invalid-id")
	if err == nil {
		t.Error("expected error for invalid ID format")
	}
}

// === イミュータブル制約テスト（M1対応）===

func TestDecisionHandler_Update_Forbidden(t *testing.T) {
	handler, conHandler, _, cleanup := setupDecisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Decision を作成
	conID := createTestConsideration(t, conHandler)
	addResult, err := handler.Add(ctx, "テスト決定",
		WithDecisionConsideration(conID),
		WithDecisionSelected(SelectedOption{OptionID: "opt-1", Title: "オプション1"}),
		WithDecisionRationale("テスト理由"),
	)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Update を試みる（イミュータブルのため常にエラー）
	decAny, _ := handler.Get(ctx, addResult.ID)
	dec := decAny.(*DecisionEntity)
	dec.Rationale = "変更した理由"

	err = handler.Update(ctx, addResult.ID, dec)
	if err == nil {
		t.Error("Update should fail for immutable Decision")
		return
	}

	// エラーメッセージに "immutable" が含まれることを確認
	if !strings.Contains(err.Error(), "immutable") {
		t.Errorf("expected error message to contain 'immutable', got: %v", err)
	}
}

func TestDecisionHandler_Delete_Forbidden(t *testing.T) {
	handler, conHandler, _, cleanup := setupDecisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Decision を作成
	conID := createTestConsideration(t, conHandler)
	addResult, err := handler.Add(ctx, "テスト決定",
		WithDecisionConsideration(conID),
		WithDecisionSelected(SelectedOption{OptionID: "opt-1", Title: "オプション1"}),
		WithDecisionRationale("テスト理由"),
	)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Delete を試みる（イミュータブルのため常にエラー）
	err = handler.Delete(ctx, addResult.ID)
	if err == nil {
		t.Error("Delete should fail for immutable Decision")
		return
	}

	// エラーメッセージに "immutable" が含まれることを確認
	if !strings.Contains(err.Error(), "immutable") {
		t.Errorf("expected error message to contain 'immutable', got: %v", err)
	}

	// エラーメッセージに "permanent records" が含まれることを確認（削除不可の理由）
	if !strings.Contains(err.Error(), "permanent") {
		t.Errorf("expected error message to explain decisions are permanent records, got: %v", err)
	}
}

// === Consideration ステータス更新テスト ===

func TestDecisionHandler_Add_UpdatesConsiderationStatus(t *testing.T) {
	handler, conHandler, _, cleanup := setupDecisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Consideration を作成
	conID := createTestConsideration(t, conHandler)

	// 作成前の Consideration ステータスを確認
	conBefore, _ := conHandler.Get(ctx, conID)
	if conBefore.(*ConsiderationEntity).Status != ConsiderationStatusOpen {
		t.Errorf("expected initial status 'open', got %q", conBefore.(*ConsiderationEntity).Status)
	}

	// Decision を作成
	_, err := handler.Add(ctx, "テスト決定",
		WithDecisionConsideration(conID),
		WithDecisionSelected(SelectedOption{OptionID: "opt-1", Title: "オプション1"}),
		WithDecisionRationale("テスト理由"),
	)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 作成後の Consideration ステータスを確認
	conAfter, _ := conHandler.Get(ctx, conID)
	if conAfter.(*ConsiderationEntity).Status != ConsiderationStatusDecided {
		t.Errorf("expected status 'decided' after Decision creation, got %q", conAfter.(*ConsiderationEntity).Status)
	}
}

// === Context キャンセルテスト ===

func TestDecisionHandler_ContextCancellation(t *testing.T) {
	handler, _, _, cleanup := setupDecisionHandlerTest(t)
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
	_, err = handler.Get(ctx, "dec-001")
	if err == nil {
		t.Error("Get should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// === ID シーケンステスト ===

func TestDecisionHandler_IDSequence(t *testing.T) {
	handler, conHandler, _, cleanup := setupDecisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 3つの Decision を作成
	ids := make([]string, 3)
	for i := 0; i < 3; i++ {
		conID := createTestConsideration(t, conHandler)
		result, err := handler.Add(ctx, "テスト決定",
			WithDecisionConsideration(conID),
			WithDecisionSelected(SelectedOption{OptionID: "opt-1", Title: "オプション1"}),
			WithDecisionRationale("テスト理由"),
		)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
		ids[i] = result.ID
	}

	// ID が連続していることを確認
	if ids[0] != "dec-001" {
		t.Errorf("expected first ID 'dec-001', got %q", ids[0])
	}
	if ids[1] != "dec-002" {
		t.Errorf("expected second ID 'dec-002', got %q", ids[1])
	}
	if ids[2] != "dec-003" {
		t.Errorf("expected third ID 'dec-003', got %q", ids[2])
	}
}

// === オプション関数テスト ===

func TestDecisionHandler_Options(t *testing.T) {
	handler, conHandler, _, cleanup := setupDecisionHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	conID := createTestConsideration(t, conHandler)

	result, err := handler.Add(ctx, "テスト決定",
		WithDecisionConsideration(conID),
		WithDecisionSelected(SelectedOption{OptionID: "opt-1", Title: "選択したオプション"}),
		WithDecisionRejected([]RejectedOption{
			{OptionID: "opt-2", Title: "却下オプション", Reason: "理由"},
		}),
		WithDecisionRationale("選択理由"),
		WithDecisionImpact([]string{"影響1", "影響2"}),
		WithDecisionDecidedBy("決定者"),
	)
	if err != nil {
		t.Fatalf("Add with options failed: %v", err)
	}

	decAny, _ := handler.Get(ctx, result.ID)
	dec := decAny.(*DecisionEntity)

	if dec.Selected.OptionID != "opt-1" {
		t.Errorf("expected selected option_id 'opt-1', got %q", dec.Selected.OptionID)
	}
	if dec.Selected.Title != "選択したオプション" {
		t.Errorf("expected selected title '選択したオプション', got %q", dec.Selected.Title)
	}
	if len(dec.Rejected) != 1 {
		t.Errorf("expected 1 rejected option, got %d", len(dec.Rejected))
	}
	if dec.Rejected[0].Reason != "理由" {
		t.Errorf("expected rejected reason '理由', got %q", dec.Rejected[0].Reason)
	}
	if dec.Rationale != "選択理由" {
		t.Errorf("expected rationale '選択理由', got %q", dec.Rationale)
	}
	if len(dec.Impact) != 2 {
		t.Errorf("expected 2 impacts, got %d", len(dec.Impact))
	}
	if dec.DecidedBy != "決定者" {
		t.Errorf("expected decided_by '決定者', got %q", dec.DecidedBy)
	}
}
