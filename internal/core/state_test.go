package core

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNewStateManager(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	sm := NewStateManager(z.ZeusPath, z.fileStore)
	if sm == nil {
		t.Error("NewStateManager should return non-nil")
	}
}

func TestGetCurrentState_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化前の状態取得（空の状態を返す）
	state, err := z.stateStore.GetCurrentState(ctx)
	if err != nil {
		t.Errorf("GetCurrentState() error = %v", err)
	}
	if state == nil {
		t.Error("GetCurrentState() should return non-nil state")
	}
	if state.Health != HealthUnknown {
		t.Errorf("expected Health 'unknown', got %q", state.Health)
	}
}

func TestSaveCurrentState(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// 状態を保存
	state := &ProjectState{
		Timestamp: Now(),
		Summary: TaskStats{
			TotalTasks: 10,
			Completed:  5,
			InProgress: 3,
			Pending:    2,
		},
		Health: HealthGood,
		Risks:  []string{},
	}

	err = z.stateStore.SaveCurrentState(ctx, state)
	if err != nil {
		t.Errorf("SaveCurrentState() error = %v", err)
	}

	// 読み込んで確認
	loaded, err := z.stateStore.GetCurrentState(ctx)
	if err != nil {
		t.Errorf("GetCurrentState() error = %v", err)
	}
	if loaded.Summary.TotalTasks != 10 {
		t.Errorf("expected TotalTasks 10, got %d", loaded.Summary.TotalTasks)
	}
}

func TestCreateSnapshot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// スナップショット作成
	snapshot, err := z.stateStore.CreateSnapshot(ctx, "test-label")
	if err != nil {
		t.Errorf("CreateSnapshot() error = %v", err)
	}
	if snapshot.Label != "test-label" {
		t.Errorf("expected Label 'test-label', got %q", snapshot.Label)
	}
	if snapshot.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}
}

func TestGetHistory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// 複数のスナップショットを作成（タイムスタンプ競合を避けるためスリープを追加）
	_, err = z.stateStore.CreateSnapshot(ctx, "snapshot-1")
	if err != nil {
		t.Fatalf("CreateSnapshot 1 error = %v", err)
	}
	time.Sleep(1100 * time.Millisecond) // 1.1秒待機（RFC3339は秒単位のため）

	_, err = z.stateStore.CreateSnapshot(ctx, "snapshot-2")
	if err != nil {
		t.Fatalf("CreateSnapshot 2 error = %v", err)
	}
	time.Sleep(1100 * time.Millisecond)

	_, err = z.stateStore.CreateSnapshot(ctx, "snapshot-3")
	if err != nil {
		t.Fatalf("CreateSnapshot 3 error = %v", err)
	}

	// 履歴を取得
	history, err := z.stateStore.GetHistory(ctx, 10)
	if err != nil {
		t.Errorf("GetHistory() error = %v", err)
	}
	if len(history) != 3 {
		t.Errorf("expected 3 snapshots, got %d", len(history))
	}

	// limit のテスト
	history, err = z.stateStore.GetHistory(ctx, 2)
	if err != nil {
		t.Errorf("GetHistory() error = %v", err)
	}
	if len(history) != 2 {
		t.Errorf("expected 2 snapshots with limit, got %d", len(history))
	}
}

func TestGetSnapshot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// スナップショット作成
	created, err := z.stateStore.CreateSnapshot(ctx, "find-me")
	if err != nil {
		t.Fatalf("CreateSnapshot() error = %v", err)
	}

	// スナップショット取得
	found, err := z.stateStore.GetSnapshot(ctx, created.Timestamp)
	if err != nil {
		t.Errorf("GetSnapshot() error = %v", err)
	}
	if found.Label != "find-me" {
		t.Errorf("expected Label 'find-me', got %q", found.Label)
	}
}

func TestGetSnapshot_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// 存在しないスナップショット取得
	_, err = z.stateStore.GetSnapshot(ctx, "non-existent")
	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestRestoreSnapshot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// スナップショット作成
	snapshot, err := z.stateStore.CreateSnapshot(ctx, "restore-test")
	if err != nil {
		t.Fatalf("CreateSnapshot() error = %v", err)
	}

	// 状態を変更
	newState := &ProjectState{
		Timestamp: Now(),
		Summary:   TaskStats{TotalTasks: 100},
		Health:    HealthPoor,
	}
	_ = z.stateStore.SaveCurrentState(ctx, newState)

	// スナップショットから復元
	err = z.stateStore.RestoreSnapshot(ctx, snapshot.Timestamp)
	if err != nil {
		t.Errorf("RestoreSnapshot() error = %v", err)
	}

	// 復元された状態を確認
	restored, _ := z.stateStore.GetCurrentState(ctx)
	if restored.Summary.TotalTasks != 0 {
		t.Errorf("expected restored TotalTasks 0, got %d", restored.Summary.TotalTasks)
	}
}

func TestCalculateState(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	sm := NewStateManager(z.ZeusPath, z.fileStore)

	tasks := []ListItem{
		{ID: "task-1", Status: ItemStatusCompleted},
		{ID: "task-2", Status: ItemStatusCompleted},
		{ID: "task-3", Status: ItemStatusInProgress},
		{ID: "task-4", Status: ItemStatusPending},
		{ID: "task-5", Status: ItemStatusBlocked},
	}

	state := sm.CalculateState(tasks)

	if state.Summary.TotalTasks != 5 {
		t.Errorf("expected TotalTasks 5, got %d", state.Summary.TotalTasks)
	}
	if state.Summary.Completed != 2 {
		t.Errorf("expected Completed 2, got %d", state.Summary.Completed)
	}
	if state.Summary.InProgress != 1 {
		t.Errorf("expected InProgress 1, got %d", state.Summary.InProgress)
	}
	// Pending には blocked も含まれる
	if state.Summary.Pending != 2 {
		t.Errorf("expected Pending 2, got %d", state.Summary.Pending)
	}
}

func TestCalculateHealth(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	sm := NewStateManager(z.ZeusPath, z.fileStore)

	tests := []struct {
		name     string
		tasks    []ListItem
		expected HealthStatus
	}{
		{
			name:     "empty tasks",
			tasks:    []ListItem{},
			expected: HealthUnknown,
		},
		{
			name: "good health (>70% completed)",
			tasks: []ListItem{
				{Status: ItemStatusCompleted},
				{Status: ItemStatusCompleted},
				{Status: ItemStatusCompleted},
				{Status: ItemStatusPending},
			},
			expected: HealthGood,
		},
		{
			name: "fair health (30-70% completed)",
			tasks: []ListItem{
				{Status: ItemStatusCompleted},
				{Status: ItemStatusPending},
				{Status: ItemStatusPending},
			},
			expected: HealthFair,
		},
		{
			name: "poor health (<30% completed)",
			tasks: []ListItem{
				{Status: ItemStatusPending},
				{Status: ItemStatusPending},
				{Status: ItemStatusPending},
				{Status: ItemStatusPending},
			},
			expected: HealthPoor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := sm.CalculateState(tt.tasks)
			if state.Health != tt.expected {
				t.Errorf("expected Health %q, got %q", tt.expected, state.Health)
			}
		})
	}
}

func TestDetectRisks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	sm := NewStateManager(z.ZeusPath, z.fileStore)

	// ブロックされたタスク
	tasks := []ListItem{
		{Status: ItemStatusBlocked},
		{Status: ItemStatusBlocked},
	}
	state := sm.CalculateState(tasks)
	foundBlockedRisk := false
	for _, risk := range state.Risks {
		if risk == "2 task(s) are blocked" {
			foundBlockedRisk = true
		}
	}
	if !foundBlockedRisk {
		t.Errorf("expected blocked risk, got %v", state.Risks)
	}

	// WIP リミット超過
	tasks = []ListItem{
		{Status: ItemStatusInProgress},
		{Status: ItemStatusInProgress},
		{Status: ItemStatusInProgress},
		{Status: ItemStatusInProgress},
		{Status: ItemStatusInProgress},
		{Status: ItemStatusInProgress},
	}
	state = sm.CalculateState(tasks)
	foundWIPRisk := false
	for _, risk := range state.Risks {
		if risk == "Too many tasks in progress (WIP limit exceeded)" {
			foundWIPRisk = true
		}
	}
	if !foundWIPRisk {
		t.Errorf("expected WIP risk, got %v", state.Risks)
	}
}

func TestSanitizeTimestamp(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "2024-01-15T10:30:00+09:00",
			expected: "2024-01-15T10-30-00-09-00",
		},
		{
			input:    "2024-01-15T10:30:00Z",
			expected: "2024-01-15T10-30-00Z",
		},
		{
			input:    "no-special-chars",
			expected: "no-special-chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeTimestamp(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestStateManager_ContextCancellation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = z.stateStore.GetCurrentState(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	err = z.stateStore.SaveCurrentState(ctx, &ProjectState{})
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	_, err = z.stateStore.CreateSnapshot(ctx, "test")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	_, err = z.stateStore.GetHistory(ctx, 10)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	_, err = z.stateStore.GetSnapshot(ctx, "test")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	err = z.stateStore.RestoreSnapshot(ctx, "test")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
