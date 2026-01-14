package mocks_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/biwakonbu/zeus/internal/core"
	"github.com/biwakonbu/zeus/internal/testing/mocks"
)

// TestMockStateStore_GetCurrentState は MockStateStore の GetCurrentState のテスト
func TestMockStateStore_GetCurrentState(t *testing.T) {
	mockState := mocks.NewMockStateStore()
	ctx := context.Background()

	state, err := mockState.GetCurrentState(ctx)
	if err != nil {
		t.Fatalf("GetCurrentState failed: %v", err)
	}

	if state.Summary.TotalTasks != 0 {
		t.Errorf("Expected 0 tasks, got %d", state.Summary.TotalTasks)
	}
}

// TestMockApprovalStore_GetPending は MockApprovalStore の GetPending のテスト
func TestMockApprovalStore_GetPending(t *testing.T) {
	mockApproval := mocks.NewMockApprovalStore()
	ctx := context.Background()

	approvals, err := mockApproval.GetPending(ctx)
	if err != nil {
		t.Fatalf("GetPending failed: %v", err)
	}

	if len(approvals) != 0 {
		t.Errorf("Expected 0 approvals, got %d", len(approvals))
	}
}

// TestMockStateStoreErrorInjection はエラー注入のテスト
func TestMockStateStoreErrorInjection(t *testing.T) {
	mockState := mocks.NewMockStateStore()
	mockState.GetCurrentStateError = fmt.Errorf("injected error")

	ctx := context.Background()
	_, err := mockState.GetCurrentState(ctx)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != "injected error" {
		t.Errorf("Expected 'injected error', got '%s'", err.Error())
	}
}

// TestMockApprovalStoreErrorInjection はエラー注入のテスト
func TestMockApprovalStoreErrorInjection(t *testing.T) {
	mockApproval := mocks.NewMockApprovalStore()
	mockApproval.GetPendingError = fmt.Errorf("injected error")

	ctx := context.Background()
	_, err := mockApproval.GetPending(ctx)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestMockStateStoreContextCancellation はコンテキストキャンセルのテスト
func TestMockStateStoreContextCancellation(t *testing.T) {
	mockState := mocks.NewMockStateStore()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 即座にキャンセル

	_, err := mockState.GetCurrentState(ctx)

	if err == nil {
		t.Fatal("Expected context canceled error, got nil")
	}
}

// TestMockApprovalStoreContextCancellation はコンテキストキャンセルのテスト
func TestMockApprovalStoreContextCancellation(t *testing.T) {
	mockApproval := mocks.NewMockApprovalStore()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 即座にキャンセル

	_, err := mockApproval.GetPending(ctx)

	if err == nil {
		t.Fatal("Expected context canceled error, got nil")
	}
}

// TestMockStateStoreSnapshotOperations はスナップショット操作のテスト
func TestMockStateStoreSnapshotOperations(t *testing.T) {
	mockState := mocks.NewMockStateStore()

	ctx := context.Background()

	// スナップショット作成
	snapshot, err := mockState.CreateSnapshot(ctx, "test-snapshot")
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	// スナップショット取得
	retrieved, err := mockState.GetSnapshot(ctx, snapshot.Timestamp)
	if err != nil {
		t.Fatalf("GetSnapshot failed: %v", err)
	}

	if retrieved.Label != "test-snapshot" {
		t.Errorf("Expected label 'test-snapshot', got '%s'", retrieved.Label)
	}

	// 履歴取得
	history, err := mockState.GetHistory(ctx, 10)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 snapshot in history, got %d", len(history))
	}
}

// TestMockApprovalStoreApprovalFlow は承認フローのテスト
func TestMockApprovalStoreApprovalFlow(t *testing.T) {
	mockApproval := mocks.NewMockApprovalStore()

	ctx := context.Background()

	// 承認作成
	approval, err := mockApproval.Create(ctx, "task_create", "Test approval", core.ApprovalApprove, "task-1", nil)
	if err != nil {
		t.Fatalf("Create approval failed: %v", err)
	}

	// Pending確認
	pending, err := mockApproval.GetPending(ctx)
	if err != nil {
		t.Fatalf("GetPending failed: %v", err)
	}

	if len(pending) != 1 {
		t.Errorf("Expected 1 pending approval, got %d", len(pending))
	}

	// 承認実行
	result, err := mockApproval.Approve(ctx, approval.ID)
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected approval to succeed")
	}

	if result.Status != core.ApprovalStatusApproved {
		t.Errorf("Expected status 'approved', got '%s'", result.Status)
	}

	// 承認後のPending確認
	pending, err = mockApproval.GetPending(ctx)
	if err != nil {
		t.Fatalf("GetPending after approval failed: %v", err)
	}

	if len(pending) != 0 {
		t.Errorf("Expected 0 pending approvals after approval, got %d", len(pending))
	}
}

// TestMockApprovalStoreRejectFlow は却下フローのテスト
func TestMockApprovalStoreRejectFlow(t *testing.T) {
	mockApproval := mocks.NewMockApprovalStore()

	ctx := context.Background()

	// 承認作成
	approval, err := mockApproval.Create(ctx, "task_delete", "Test reject", core.ApprovalApprove, "task-2", nil)
	if err != nil {
		t.Fatalf("Create approval failed: %v", err)
	}

	// 却下実行
	result, err := mockApproval.Reject(ctx, approval.ID, "Not needed")
	if err != nil {
		t.Fatalf("Reject failed: %v", err)
	}

	if result.Success {
		t.Error("Expected rejection to not succeed")
	}

	if result.Status != core.ApprovalStatusRejected {
		t.Errorf("Expected status 'rejected', got '%s'", result.Status)
	}

	// 却下後のPending確認
	pending, err := mockApproval.GetPending(ctx)
	if err != nil {
		t.Fatalf("GetPending after rejection failed: %v", err)
	}

	if len(pending) != 0 {
		t.Errorf("Expected 0 pending approvals after rejection, got %d", len(pending))
	}
}

// TestMockStateStoreCalculateState はCalculateStateのテスト
func TestMockStateStoreCalculateState(t *testing.T) {
	mockState := mocks.NewMockStateStore()

	tasks := []core.Task{
		{ID: "1", Status: core.TaskStatusCompleted},
		{ID: "2", Status: core.TaskStatusInProgress},
		{ID: "3", Status: core.TaskStatusPending},
		{ID: "4", Status: core.TaskStatusCompleted},
	}

	state := mockState.CalculateState(tasks)

	if state.Summary.TotalTasks != 4 {
		t.Errorf("Expected 4 total tasks, got %d", state.Summary.TotalTasks)
	}

	if state.Summary.Completed != 2 {
		t.Errorf("Expected 2 completed tasks, got %d", state.Summary.Completed)
	}

	if state.Summary.InProgress != 1 {
		t.Errorf("Expected 1 in-progress task, got %d", state.Summary.InProgress)
	}

	if state.Summary.Pending != 1 {
		t.Errorf("Expected 1 pending task, got %d", state.Summary.Pending)
	}

	if state.Health != core.HealthFair {
		t.Errorf("Expected health 'fair' (50%% completion), got '%s'", state.Health)
	}
}

// TestMockStateStoreContextTimeout はコンテキストタイムアウトのテスト
func TestMockStateStoreContextTimeout(t *testing.T) {
	mockState := mocks.NewMockStateStore()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // タイムアウトを確実に発生させる

	_, err := mockState.GetCurrentState(ctx)

	if err == nil {
		t.Fatal("Expected context deadline exceeded error, got nil")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}
