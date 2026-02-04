package core

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// テスト用のセットアップ（ConsiderationHandler）
func setupConsiderationHandlerTest(t *testing.T) (*ConsiderationHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-consideration-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath+"/considerations", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create zeus dir: %v", err)
	}

	fs := yaml.NewFileManager(zeusPath)
	handler := NewConsiderationHandler(fs, nil, nil, nil)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return handler, zeusPath, cleanup
}

// ===== Type テスト =====

func TestConsiderationHandler_Type(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	if handler.Type() != "consideration" {
		t.Errorf("expected type 'consideration', got %q", handler.Type())
	}
}

// ===== Add テスト =====

func TestConsiderationHandler_Add(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "データベース選定")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if !result.Success {
		t.Error("Add should succeed")
	}

	if result.Entity != "consideration" {
		t.Errorf("expected entity 'consideration', got %q", result.Entity)
	}

	if result.ID == "" {
		t.Error("ID should not be empty")
	}

	// ID フォーマット確認 (con-NNN)
	if len(result.ID) < 4 || result.ID[:4] != "con-" {
		t.Errorf("expected ID to start with 'con-', got %q", result.ID)
	}
}

func TestConsiderationHandler_Add_WithOptions(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	options := []ConsiderationOption{
		{ID: "opt-001", Title: "PostgreSQL", Pros: []string{"ACID準拠"}, Cons: []string{"スケーリングが難しい"}},
		{ID: "opt-002", Title: "MySQL", Pros: []string{"高速"}, Cons: []string{"機能が限定的"}},
	}

	result, err := handler.Add(ctx, "データベース選定",
		WithConsiderationContext("バックエンドのデータ永続化層"),
		WithConsiderationOptions(options),
		WithConsiderationRaisedBy("開発チーム"),
		WithConsiderationDueDate("2026-03-01"),
	)
	if err != nil {
		t.Fatalf("Add with options failed: %v", err)
	}

	// 取得して確認
	conAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	con := conAny.(*ConsiderationEntity)
	if con.Context != "バックエンドのデータ永続化層" {
		t.Errorf("expected context 'バックエンドのデータ永続化層', got %q", con.Context)
	}

	if len(con.Options) != 2 {
		t.Errorf("expected 2 options, got %d", len(con.Options))
	}

	if con.RaisedBy != "開発チーム" {
		t.Errorf("expected raised_by '開発チーム', got %q", con.RaisedBy)
	}

	if con.DueDate != "2026-03-01" {
		t.Errorf("expected due_date '2026-03-01', got %q", con.DueDate)
	}

	if con.Status != ConsiderationStatusOpen {
		t.Errorf("expected status 'open', got %q", con.Status)
	}
}

func TestConsiderationHandler_Add_InvalidInput(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 空のタイトル
	_, err := handler.Add(ctx, "")
	if err == nil {
		t.Error("Add with empty title should fail")
	}
}

// ===== List テスト =====

func TestConsiderationHandler_List(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 3件追加
	for i := 0; i < 3; i++ {
		_, err := handler.Add(ctx, "検討事項")
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	result, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("expected 3 considerations, got %d", result.Total)
	}

	if result.Entity != "considerations" {
		t.Errorf("expected entity 'considerations', got %q", result.Entity)
	}
}

func TestConsiderationHandler_List_Empty(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.List(ctx, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if result.Total != 0 {
		t.Errorf("expected 0 considerations, got %d", result.Total)
	}
}

func TestConsiderationHandler_List_WithLimit(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 5件追加
	for i := 0; i < 5; i++ {
		_, err := handler.Add(ctx, "検討事項")
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	result, err := handler.List(ctx, &ListFilter{Limit: 2})
	if err != nil {
		t.Fatalf("List with limit failed: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected 2 considerations (limited), got %d", result.Total)
	}
}

func TestConsiderationHandler_List_WithStatusFilter(t *testing.T) {
	handler, zeusPath, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 異なるステータスの Consideration を直接作成
	openCon := &ConsiderationEntity{
		ID:       "con-001",
		Title:    "Open",
		Status:   ConsiderationStatusOpen,
		Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	decidedCon := &ConsiderationEntity{
		ID:       "con-002",
		Title:    "Decided",
		Status:   ConsiderationStatusDecided,
		Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	deferredCon := &ConsiderationEntity{
		ID:       "con-003",
		Title:    "Deferred",
		Status:   ConsiderationStatusDeferred,
		Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}

	for _, con := range []*ConsiderationEntity{openCon, decidedCon, deferredCon} {
		if err := fs.WriteYaml(ctx, "considerations/"+con.ID+".yaml", con); err != nil {
			t.Fatalf("Write consideration failed: %v", err)
		}
	}

	// open のみフィルタ
	result, err := handler.List(ctx, &ListFilter{Status: "open"})
	if err != nil {
		t.Fatalf("List with status filter failed: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("expected 1 open consideration, got %d", result.Total)
	}
}

// ===== Get テスト =====

func TestConsiderationHandler_Get(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "データベース選定")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	conAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	con := conAny.(*ConsiderationEntity)
	if con.Title != "データベース選定" {
		t.Errorf("expected title 'データベース選定', got %q", con.Title)
	}

	if con.ID != result.ID {
		t.Errorf("expected ID %q, got %q", result.ID, con.ID)
	}
}

func TestConsiderationHandler_Get_NotFound(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	_, err := handler.Get(ctx, "con-999")
	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestConsiderationHandler_Get_InvalidID(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	_, err := handler.Get(ctx, "invalid-id")
	if err == nil {
		t.Error("Get with invalid ID should fail")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

// ===== Update テスト =====

func TestConsiderationHandler_Update(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 追加
	result, err := handler.Add(ctx, "データベース選定")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 更新
	updated := &ConsiderationEntity{
		Title:   "データベース選定（更新）",
		Context: "更新されたコンテキスト",
		Status:  ConsiderationStatusDeferred,
	}

	if err := handler.Update(ctx, result.ID, updated); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// 取得して確認
	conAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	con := conAny.(*ConsiderationEntity)
	if con.Title != "データベース選定（更新）" {
		t.Errorf("expected updated title, got %q", con.Title)
	}

	if con.Context != "更新されたコンテキスト" {
		t.Errorf("expected updated context, got %q", con.Context)
	}

	if con.Status != ConsiderationStatusDeferred {
		t.Errorf("expected status 'deferred', got %q", con.Status)
	}
}

func TestConsiderationHandler_Update_NotFound(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	updated := &ConsiderationEntity{
		Title: "更新",
	}

	err := handler.Update(ctx, "con-999", updated)
	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestConsiderationHandler_Update_InvalidType(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "検討事項")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 不正な型で更新
	err = handler.Update(ctx, result.ID, "invalid")
	if err == nil {
		t.Error("Update with invalid type should fail")
	}
}

// ===== Delete テスト =====

func TestConsiderationHandler_Delete(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "検討事項")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// 削除
	if err := handler.Delete(ctx, result.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 取得できないことを確認
	_, err = handler.Get(ctx, result.ID)
	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound after delete, got %v", err)
	}
}

func TestConsiderationHandler_Delete_NotFound(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	err := handler.Delete(ctx, "con-999")
	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestConsiderationHandler_Delete_ReferencedByDecision(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-consideration-decision-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath+"/considerations", 0755); err != nil {
		t.Fatalf("failed to create considerations dir: %v", err)
	}
	if err := os.MkdirAll(zeusPath+"/decisions", 0755); err != nil {
		t.Fatalf("failed to create decisions dir: %v", err)
	}

	fs := yaml.NewFileManager(zeusPath)
	conHandler := NewConsiderationHandler(fs, nil, nil, nil)

	ctx := context.Background()

	// Consideration を作成
	conResult, err := conHandler.Add(ctx, "検討事項")
	if err != nil {
		t.Fatalf("Add consideration failed: %v", err)
	}

	// Decision を直接作成（Consideration を参照）
	dec := &DecisionEntity{
		ID:              "dec-001",
		Title:           "決定事項",
		ConsiderationID: conResult.ID,
		Selected:        SelectedOption{OptionID: "opt-001", Title: "選択肢"},
		Rationale:       "理由",
		DecidedAt:       Now(),
	}
	if err := fs.WriteYaml(ctx, "decisions/dec-001.yaml", dec); err != nil {
		t.Fatalf("Write decision failed: %v", err)
	}

	// 削除を試みる（M3: 逆参照整合性により失敗するはず）
	err = conHandler.Delete(ctx, conResult.ID)
	if err == nil {
		t.Error("Delete should fail when referenced by Decision")
	}

	// エラーメッセージに Decision ID が含まれることを確認
	if err != nil && !contains(err.Error(), "dec-001") {
		t.Errorf("error should mention referencing decision, got: %v", err)
	}
}

// ===== SetDecisionID テスト =====

func TestConsiderationHandler_SetDecisionID(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	result, err := handler.Add(ctx, "検討事項")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// DecisionID を設定
	if err := handler.SetDecisionID(ctx, result.ID, "dec-001"); err != nil {
		t.Fatalf("SetDecisionID failed: %v", err)
	}

	// 取得して確認
	conAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	con := conAny.(*ConsiderationEntity)
	if con.DecisionID != "dec-001" {
		t.Errorf("expected DecisionID 'dec-001', got %q", con.DecisionID)
	}

	// ステータスが decided に変更されていることを確認
	if con.Status != ConsiderationStatusDecided {
		t.Errorf("expected status 'decided', got %q", con.Status)
	}
}

func TestConsiderationHandler_SetDecisionID_NotFound(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	err := handler.SetDecisionID(ctx, "con-999", "dec-001")
	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

// ===== Context キャンセルテスト =====

func TestConsiderationHandler_ContextCancellation(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	// キャンセル済みのコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Add
	_, err := handler.Add(ctx, "検討事項")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Add should fail with cancelled context, got %v", err)
	}

	// List
	_, err = handler.List(ctx, nil)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("List should fail with cancelled context, got %v", err)
	}

	// Get
	_, err = handler.Get(ctx, "con-001")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Get should fail with cancelled context, got %v", err)
	}

	// Update
	err = handler.Update(ctx, "con-001", &ConsiderationEntity{})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Update should fail with cancelled context, got %v", err)
	}

	// Delete
	err = handler.Delete(ctx, "con-001")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Delete should fail with cancelled context, got %v", err)
	}
}

// ===== ID シーケンステスト =====

func TestConsiderationHandler_IDSequence(t *testing.T) {
	handler, _, cleanup := setupConsiderationHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 全ての ID がユニークでプレフィックスが正しいことを確認
	seen := make(map[string]bool)
	for i := 0; i < 3; i++ {
		result, err := handler.Add(ctx, "検討事項")
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
		if seen[result.ID] {
			t.Errorf("duplicate ID found: %q", result.ID)
		}
		seen[result.ID] = true

		// プレフィックスが正しいことを確認
		if !strings.HasPrefix(result.ID, "con-") {
			t.Errorf("Add() #%d ID = %q, expected prefix 'con-'", i+1, result.ID)
		}
	}

	// ID 数が正しいことを確認
	if len(seen) != 3 {
		t.Errorf("expected 3 unique IDs, got %d", len(seen))
	}
}

// contains は文字列に部分文字列が含まれるかをチェック（integrity_test.go と同じ）
// Note: この関数は既に integrity_test.go で定義されているが、
// パッケージ内の他のテストファイルからも参照可能
