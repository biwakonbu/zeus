package core

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestGenerateTaskID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)

	// ID 生成テスト
	id1 := z.generateTaskID()
	id2 := z.generateTaskID()

	// プレフィックスが正しいか
	if !strings.HasPrefix(id1, "task-") {
		t.Errorf("expected ID to start with 'task-', got %q", id1)
	}

	// UUID ベースのため、2つの ID が異なるはず
	if id1 == id2 {
		t.Errorf("generated IDs should be unique, but got same: %q", id1)
	}

	// ID の長さが適切か (task- + 8文字)
	if len(id1) != 13 {
		t.Errorf("expected ID length to be 13, got %d", len(id1))
	}
}

func TestGenerateTaskIDUniqueness(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)

	// 1000個の ID を生成して重複がないか確認
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := z.generateTaskID()
		if ids[id] {
			t.Errorf("duplicate ID generated: %q", id)
		}
		ids[id] = true
	}
}


// DI テスト: デフォルト実装が使用されることを確認
func TestZeusDefaultImplementations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)

	// デフォルト実装が設定されていることを確認
	if z.fileStore == nil {
		t.Error("fileStore should have default implementation")
	}

	if z.stateStore == nil {
		t.Error("stateStore should have default implementation")
	}

	if z.approvalStore == nil {
		t.Error("approvalStore should have default implementation")
	}

	if z.entityRegistry == nil {
		t.Error("entityRegistry should have default implementation")
	}
}

// Context タイムアウトテスト: Init
func TestInitContextTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)

	// 既にキャンセルされたコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = z.Init(ctx, "simple")
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// Context タイムアウトテスト: Status
func TestStatusContextTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)

	// 既にキャンセルされたコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = z.Status(ctx)
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// Context タイムアウトテスト: Add
func TestAddContextTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)

	// 既にキャンセルされたコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = z.Add(ctx, "task", "test")
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// Context タイムアウトテスト: List
func TestListContextTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)

	// 既にキャンセルされたコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = z.List(ctx, "task")
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// 統合テスト: Init から Add, List まで
func TestZeusIntegration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// Init
	result, err := z.Init(ctx, "simple")
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if !result.Success {
		t.Error("Init should succeed")
	}

	// Status
	status, err := z.Status(ctx)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if status.State.Summary.TotalTasks != 0 {
		t.Errorf("expected 0 tasks, got %d", status.State.Summary.TotalTasks)
	}

	// Add task
	addResult, err := z.Add(ctx, "task", "Test Task 1")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if !addResult.Success {
		t.Error("Add should succeed")
	}
	if addResult.Entity != "task" {
		t.Errorf("expected entity 'task', got %q", addResult.Entity)
	}

	// List tasks
	listResult, err := z.List(ctx, "tasks")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if listResult.Total != 1 {
		t.Errorf("expected 1 task, got %d", listResult.Total)
	}

	// Add another task
	_, err = z.Add(ctx, "task", "Test Task 2")
	if err != nil {
		t.Fatalf("Add second task failed: %v", err)
	}

	// List again
	listResult, err = z.List(ctx, "tasks")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if listResult.Total != 2 {
		t.Errorf("expected 2 tasks, got %d", listResult.Total)
	}
}

// スナップショットテスト
func TestZeusSnapshot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// Init
	_, err = z.Init(ctx, "standard")
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Create snapshot
	snapshot, err := z.CreateSnapshot(ctx, "test-snapshot")
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}
	if snapshot.Label != "test-snapshot" {
		t.Errorf("expected label 'test-snapshot', got %q", snapshot.Label)
	}

	// Get history
	history, err := z.GetHistory(ctx, 10)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("expected 1 snapshot, got %d", len(history))
	}
}

// タイムアウトを使ったテスト
func TestZeusWithTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)

	// 十分なタイムアウトでの操作
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := z.Init(ctx, "simple")
	if err != nil {
		t.Fatalf("Init with timeout failed: %v", err)
	}
	if !result.Success {
		t.Error("Init should succeed with sufficient timeout")
	}
}
