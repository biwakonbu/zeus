package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

func TestGenerateApprovalID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "approval-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// FileStore を作成して ApprovalManager に渡す
	fs := yaml.NewFileManager(tmpDir)
	am := NewApprovalManager(tmpDir, fs)

	// ID 生成テスト
	id1 := am.generateApprovalID()
	id2 := am.generateApprovalID()

	// プレフィックスが正しいか
	if !strings.HasPrefix(id1, "approval-") {
		t.Errorf("expected ID to start with 'approval-', got %q", id1)
	}

	// UUID ベースのため、2つの ID が異なるはず
	if id1 == id2 {
		t.Errorf("generated IDs should be unique, but got same: %q", id1)
	}

	// ID の長さが適切か (approval- + 8文字)
	if len(id1) != 17 {
		t.Errorf("expected ID length to be 17, got %d", len(id1))
	}
}

func TestApprovalNotPendingError(t *testing.T) {
	err := &ApprovalNotPendingError{
		ID:            "approval-12345678",
		CurrentStatus: ApprovalStatusApproved,
	}

	// Error() メッセージ検証
	msg := err.Error()
	if !strings.Contains(msg, "approval-12345678") {
		t.Errorf("error message should contain ID, got %q", msg)
	}
	if !strings.Contains(msg, "approved") {
		t.Errorf("error message should contain status, got %q", msg)
	}

	// Is() 検証
	if !err.Is(ErrApprovalNotPending) {
		t.Error("ApprovalNotPendingError should match ErrApprovalNotPending")
	}
}

func TestApprovalManager_Create(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "approval-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := yaml.NewFileManager(tmpDir)
	am := NewApprovalManager(tmpDir, fs)
	ctx := context.Background()

	// 承認待ちアイテムを作成
	approval, err := am.Create(ctx, "task_create", "テストタスク作成", ApprovalApprove, "task-123", map[string]string{"name": "テスト"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if approval == nil {
		t.Fatal("expected non-nil approval")
	}
	if approval.Type != "task_create" {
		t.Errorf("expected Type 'task_create', got %q", approval.Type)
	}
	if approval.Status != ApprovalStatusPending {
		t.Errorf("expected Status 'pending', got %q", approval.Status)
	}
	if approval.EntityID != "task-123" {
		t.Errorf("expected EntityID 'task-123', got %q", approval.EntityID)
	}
}

func TestApprovalManager_Get(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "approval-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := yaml.NewFileManager(tmpDir)
	am := NewApprovalManager(tmpDir, fs)
	ctx := context.Background()

	// 承認待ちアイテムを作成
	created, err := am.Create(ctx, "task_create", "テスト", ApprovalApprove, "", nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 作成したアイテムを取得
	got, err := am.Get(ctx, created.ID)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, got.ID)
	}

	// 存在しないアイテムを取得
	_, err = am.Get(ctx, "non-existent")
	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestApprovalManager_GetPending(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "approval-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := yaml.NewFileManager(tmpDir)
	am := NewApprovalManager(tmpDir, fs)
	ctx := context.Background()

	// 空の状態で取得
	pending, err := am.GetPending(ctx)
	if err != nil {
		t.Errorf("GetPending() error = %v", err)
	}
	if len(pending) != 0 {
		t.Errorf("expected 0 pending items, got %d", len(pending))
	}

	// 複数のアイテムを作成
	_, _ = am.Create(ctx, "task_create", "テスト1", ApprovalApprove, "", nil)
	_, _ = am.Create(ctx, "task_create", "テスト2", ApprovalApprove, "", nil)

	pending, err = am.GetPending(ctx)
	if err != nil {
		t.Errorf("GetPending() error = %v", err)
	}
	if len(pending) != 2 {
		t.Errorf("expected 2 pending items, got %d", len(pending))
	}
}

func TestApprovalManager_GetAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "approval-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := yaml.NewFileManager(tmpDir)
	am := NewApprovalManager(tmpDir, fs)
	ctx := context.Background()

	// 複数のアイテムを作成
	_, _ = am.Create(ctx, "task_create", "テスト1", ApprovalApprove, "", nil)
	_, _ = am.Create(ctx, "task_update", "テスト2", ApprovalNotify, "", nil)

	all, err := am.GetAll(ctx)
	if err != nil {
		t.Errorf("GetAll() error = %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 items, got %d", len(all))
	}
}

func TestApprovalManager_Approve(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "approval-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := yaml.NewFileManager(tmpDir)
	am := NewApprovalManager(tmpDir, fs)
	ctx := context.Background()

	// 承認待ちアイテムを作成
	created, err := am.Create(ctx, "task_create", "承認テスト", ApprovalApprove, "", nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 承認
	result, err := am.Approve(ctx, created.ID)
	if err != nil {
		t.Errorf("Approve() error = %v", err)
	}
	if !result.Success {
		t.Error("expected Success to be true")
	}
	if result.Status != ApprovalStatusApproved {
		t.Errorf("expected Status 'approved', got %q", result.Status)
	}

	// 承認済みファイルが存在するか確認
	approvedPath := filepath.Join(tmpDir, "approvals", "approved", created.ID+".yaml")
	if _, err := os.Stat(approvedPath); os.IsNotExist(err) {
		t.Error("approved file should exist")
	}

	// ペンディングリストから削除されているか確認
	pending, _ := am.GetPending(ctx)
	for _, p := range pending {
		if p.ID == created.ID {
			t.Error("approved item should not be in pending list")
		}
	}
}

func TestApprovalManager_Approve_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "approval-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := yaml.NewFileManager(tmpDir)
	am := NewApprovalManager(tmpDir, fs)
	ctx := context.Background()

	// 存在しないアイテムを承認
	_, err = am.Approve(ctx, "non-existent")
	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestApprovalManager_Approve_AlreadyApproved(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "approval-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := yaml.NewFileManager(tmpDir)
	am := NewApprovalManager(tmpDir, fs)
	ctx := context.Background()

	// 承認待ちアイテムを作成して承認
	created, _ := am.Create(ctx, "task_create", "テスト", ApprovalApprove, "", nil)
	_, _ = am.Approve(ctx, created.ID)

	// 再度承認しようとするとエラー（キューから削除されているので NotFound）
	_, err = am.Approve(ctx, created.ID)
	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound for already approved item, got %v", err)
	}
}

func TestApprovalManager_Reject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "approval-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := yaml.NewFileManager(tmpDir)
	am := NewApprovalManager(tmpDir, fs)
	ctx := context.Background()

	// 承認待ちアイテムを作成
	created, err := am.Create(ctx, "task_create", "却下テスト", ApprovalApprove, "", nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 却下
	result, err := am.Reject(ctx, created.ID, "テスト用却下理由")
	if err != nil {
		t.Errorf("Reject() error = %v", err)
	}
	if !result.Success {
		t.Error("expected Success to be true")
	}
	if result.Status != ApprovalStatusRejected {
		t.Errorf("expected Status 'rejected', got %q", result.Status)
	}

	// 却下済みファイルが存在するか確認
	rejectedPath := filepath.Join(tmpDir, "approvals", "rejected", created.ID+".yaml")
	if _, err := os.Stat(rejectedPath); os.IsNotExist(err) {
		t.Error("rejected file should exist")
	}

	// ペンディングリストから削除されているか確認
	pending, _ := am.GetPending(ctx)
	for _, p := range pending {
		if p.ID == created.ID {
			t.Error("rejected item should not be in pending list")
		}
	}
}

func TestApprovalManager_Reject_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "approval-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := yaml.NewFileManager(tmpDir)
	am := NewApprovalManager(tmpDir, fs)
	ctx := context.Background()

	// 存在しないアイテムを却下
	_, err = am.Reject(ctx, "non-existent", "理由")
	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestApprovalManager_DetermineApprovalLevel(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "approval-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := yaml.NewFileManager(tmpDir)
	am := NewApprovalManager(tmpDir, fs)

	tests := []struct {
		name       string
		actionType string
		mode       string
		expected   ApprovalLevel
	}{
		// strict モード
		{"strict_task_create", "task_create", "strict", ApprovalApprove},
		{"strict_task_update", "task_update", "strict", ApprovalApprove},
		{"strict_suggestion", "suggestion", "strict", ApprovalApprove},
		{"strict_other", "other", "strict", ApprovalNotify},

		// loose モード
		{"loose_task_create", "task_create", "loose", ApprovalAuto},
		{"loose_suggestion", "suggestion", "loose", ApprovalNotify},

		// デフォルトモード
		{"default_suggestion", "suggestion", "", ApprovalApprove},
		{"default_task_update", "task_update", "", ApprovalNotify},
		{"default_task_create", "task_create", "", ApprovalAuto},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := &Settings{ApprovalMode: tt.mode}
			level := am.DetermineApprovalLevel(tt.actionType, settings)
			if level != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, level)
			}
		})
	}
}

func TestApprovalManager_ContextCancellation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "approval-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := yaml.NewFileManager(tmpDir)
	am := NewApprovalManager(tmpDir, fs)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// キャンセル済みコンテキストでのテスト
	_, err = am.GetPending(ctx)
	if err != context.Canceled {
		t.Errorf("GetPending: expected context.Canceled, got %v", err)
	}

	_, err = am.GetAll(ctx)
	if err != context.Canceled {
		t.Errorf("GetAll: expected context.Canceled, got %v", err)
	}

	_, err = am.Get(ctx, "test")
	if err != context.Canceled {
		t.Errorf("Get: expected context.Canceled, got %v", err)
	}

	_, err = am.Create(ctx, "test", "test", ApprovalAuto, "", nil)
	if err != context.Canceled {
		t.Errorf("Create: expected context.Canceled, got %v", err)
	}

	_, err = am.Approve(ctx, "test")
	if err != context.Canceled {
		t.Errorf("Approve: expected context.Canceled, got %v", err)
	}

	_, err = am.Reject(ctx, "test", "reason")
	if err != context.Canceled {
		t.Errorf("Reject: expected context.Canceled, got %v", err)
	}
}
