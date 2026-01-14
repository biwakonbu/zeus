package core

import (
	"context"
	"os"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// テスト用のセットアップ
func setupTaskHandlerTest(t *testing.T) (*TaskHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-task-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath+"/tasks", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create zeus dir: %v", err)
	}

	// 初期タスクファイルを作成
	fs := yaml.NewFileManager(zeusPath)
	taskStore := &TaskStore{Tasks: []Task{}}
	ctx := context.Background()
	if err := fs.WriteYaml(ctx, "tasks/active.yaml", taskStore); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create initial task file: %v", err)
	}

	handler := NewTaskHandler(fs)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return handler, zeusPath, cleanup
}

func TestTaskHandlerType(t *testing.T) {
	handler, _, cleanup := setupTaskHandlerTest(t)
	defer cleanup()

	if handler.Type() != "task" {
		t.Errorf("expected type 'task', got %q", handler.Type())
	}
}

func TestTaskHandlerAdd(t *testing.T) {
	handler, _, cleanup := setupTaskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// タスク追加
	result, err := handler.Add(ctx, "Test Task")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if !result.Success {
		t.Error("Add should succeed")
	}

	if result.Entity != "task" {
		t.Errorf("expected entity 'task', got %q", result.Entity)
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
		t.Errorf("expected 1 task, got %d", listResult.Total)
	}
}

func TestTaskHandlerAddWithOptions(t *testing.T) {
	handler, _, cleanup := setupTaskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// オプション付きでタスク追加
	result, err := handler.Add(ctx, "Test Task with Options",
		WithTaskDescription("Test description"),
		WithTaskStatus(TaskStatusInProgress),
		WithTaskAssignee("test-user"),
		WithTaskEstimateHours(8.0),
	)
	if err != nil {
		t.Fatalf("Add with options failed: %v", err)
	}

	// タスクを取得して確認
	taskAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	task := taskAny.(*Task)
	if task.Description != "Test description" {
		t.Errorf("expected description 'Test description', got %q", task.Description)
	}

	if task.Status != TaskStatusInProgress {
		t.Errorf("expected status 'in_progress', got %q", task.Status)
	}

	if task.Assignee != "test-user" {
		t.Errorf("expected assignee 'test-user', got %q", task.Assignee)
	}

	if task.EstimateHours != 8.0 {
		t.Errorf("expected estimate 8.0, got %f", task.EstimateHours)
	}
}

func TestTaskHandlerList(t *testing.T) {
	handler, _, cleanup := setupTaskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 複数タスクを追加
	for i := 0; i < 5; i++ {
		_, err := handler.Add(ctx, "Task")
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
		t.Errorf("expected 5 tasks, got %d", listResult.Total)
	}
}

func TestTaskHandlerListWithFilter(t *testing.T) {
	handler, _, cleanup := setupTaskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 異なるステータスのタスクを追加
	_, err := handler.Add(ctx, "Pending Task", WithTaskStatus(TaskStatusPending))
	if err != nil {
		t.Fatalf("Add pending failed: %v", err)
	}

	_, err = handler.Add(ctx, "In Progress Task", WithTaskStatus(TaskStatusInProgress))
	if err != nil {
		t.Fatalf("Add in-progress failed: %v", err)
	}

	_, err = handler.Add(ctx, "Completed Task", WithTaskStatus(TaskStatusCompleted))
	if err != nil {
		t.Fatalf("Add completed failed: %v", err)
	}

	// pending でフィルタ
	filter := &ListFilter{Status: string(TaskStatusPending)}
	listResult, err := handler.List(ctx, filter)
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}

	if listResult.Total != 1 {
		t.Errorf("expected 1 pending task, got %d", listResult.Total)
	}
}

func TestTaskHandlerListWithLimit(t *testing.T) {
	handler, _, cleanup := setupTaskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 5タスクを追加
	for i := 0; i < 5; i++ {
		_, err := handler.Add(ctx, "Task")
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Limit=3 でリスト
	filter := &ListFilter{Limit: 3}
	listResult, err := handler.List(ctx, filter)
	if err != nil {
		t.Fatalf("List with limit failed: %v", err)
	}

	if listResult.Total != 3 {
		t.Errorf("expected 3 tasks with limit, got %d", listResult.Total)
	}
}

func TestTaskHandlerGet(t *testing.T) {
	handler, _, cleanup := setupTaskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// タスク追加
	result, err := handler.Add(ctx, "Get Test Task")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// タスクを取得
	taskAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	task := taskAny.(*Task)
	if task.Title != "Get Test Task" {
		t.Errorf("expected title 'Get Test Task', got %q", task.Title)
	}
}

func TestTaskHandlerGetNotFound(t *testing.T) {
	handler, _, cleanup := setupTaskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で取得
	_, err := handler.Get(ctx, "non-existent-id")
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestTaskHandlerUpdate(t *testing.T) {
	handler, _, cleanup := setupTaskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// タスク追加
	result, err := handler.Add(ctx, "Update Test Task")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// タスクを取得
	taskAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	task := taskAny.(*Task)

	// 更新
	task.Title = "Updated Title"
	task.Status = TaskStatusCompleted
	err = handler.Update(ctx, result.ID, task)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// 更新を確認
	updatedAny, err := handler.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}

	updated := updatedAny.(*Task)
	if updated.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", updated.Title)
	}

	if updated.Status != TaskStatusCompleted {
		t.Errorf("expected status 'completed', got %q", updated.Status)
	}
}

func TestTaskHandlerUpdateNotFound(t *testing.T) {
	handler, _, cleanup := setupTaskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で更新
	task := &Task{ID: "non-existent-id", Title: "Test"}
	err := handler.Update(ctx, "non-existent-id", task)
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestTaskHandlerDelete(t *testing.T) {
	handler, _, cleanup := setupTaskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// タスク追加
	result, err := handler.Add(ctx, "Delete Test Task")
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
		t.Error("expected error for deleted task")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestTaskHandlerDeleteNotFound(t *testing.T) {
	handler, _, cleanup := setupTaskHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 存在しない ID で削除
	err := handler.Delete(ctx, "non-existent-id")
	if err == nil {
		t.Error("expected error for non-existent ID")
	}

	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestTaskHandlerContextCancellation(t *testing.T) {
	handler, _, cleanup := setupTaskHandlerTest(t)
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
	_, err = handler.Get(ctx, "test-id")
	if err == nil {
		t.Error("Get should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Update
	err = handler.Update(ctx, "test-id", &Task{})
	if err == nil {
		t.Error("Update should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Delete
	err = handler.Delete(ctx, "test-id")
	if err == nil {
		t.Error("Delete should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestGenerateTaskIDFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-task-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	zeusPath := tmpDir + "/.zeus"
	fs := yaml.NewFileManager(zeusPath)
	handler := NewTaskHandler(fs)

	// ID 生成テスト
	id := handler.generateTaskID()

	// プレフィックスが正しいか
	if len(id) < 5 || id[:5] != "task-" {
		t.Errorf("expected ID to start with 'task-', got %q", id)
	}

	// 長さが正しいか (task- + 8文字)
	if len(id) != 13 {
		t.Errorf("expected ID length to be 13, got %d", len(id))
	}
}
