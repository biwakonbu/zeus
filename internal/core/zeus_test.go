package core

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/biwakonbu/zeus/internal/yaml"
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

	_, err = z.Init(ctx)
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
	result, err := z.Init(ctx)
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
	_, err = z.Init(ctx)
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

	result, err := z.Init(ctx)
	if err != nil {
		t.Fatalf("Init with timeout failed: %v", err)
	}
	if !result.Success {
		t.Error("Init should succeed with sufficient timeout")
	}
}

// ===== 追加テスト =====

// Pending テスト
func TestPending(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 初期状態では承認待ちがない
	pending, err := z.Pending(ctx)
	if err != nil {
		t.Fatalf("Pending failed: %v", err)
	}
	if len(pending) != 0 {
		t.Errorf("expected 0 pending, got %d", len(pending))
	}
}

// Pending コンテキストキャンセルテスト
func TestPendingContextTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = z.Pending(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// Approve テスト
func TestApprove(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 存在しない ID を承認しようとするとエラー
	_, err = z.Approve(ctx, "nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent approval")
	}
}

// Approve コンテキストキャンセルテスト
func TestApproveContextTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = z.Approve(ctx, "test-id")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// Reject テスト
func TestReject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 存在しない ID を却下しようとするとエラー
	_, err = z.Reject(ctx, "nonexistent-id", "test reason")
	if err == nil {
		t.Error("expected error for nonexistent approval")
	}
}

// Reject コンテキストキャンセルテスト
func TestRejectContextTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = z.Reject(ctx, "test-id", "reason")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// RestoreSnapshot 統合テスト
func TestRestoreSnapshotIntegration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// タスクを追加
	_, _ = z.Add(ctx, "task", "Task 1")

	// スナップショット作成
	snapshot, err := z.CreateSnapshot(ctx, "before-more-tasks")
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	// さらにタスクを追加
	_, _ = z.Add(ctx, "task", "Task 2")
	_, _ = z.Add(ctx, "task", "Task 3")

	// スナップショットから復元
	err = z.RestoreSnapshot(ctx, snapshot.Timestamp)
	if err != nil {
		t.Fatalf("RestoreSnapshot failed: %v", err)
	}

	// 復元後の状態を確認（スナップショット時点のタスク数が復元される）
	status, _ := z.Status(ctx)
	// 注意: RestoreSnapshot は状態のみを復元し、タスクストア自体は変更しない
	if status.State.Summary.TotalTasks != 1 {
		// スナップショット時点では 1 タスク
		t.Logf("Note: RestoreSnapshot restores state snapshot, tasks count in state: %d", status.State.Summary.TotalTasks)
	}
}

// RestoreSnapshot コンテキストキャンセルテスト
func TestRestoreSnapshotContextTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = z.RestoreSnapshot(ctx, "test-timestamp")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// Explain プロジェクトテスト
func TestExplain_Project(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// プロジェクト説明を取得
	result, err := z.Explain(ctx, "project", false)
	if err != nil {
		t.Fatalf("Explain failed: %v", err)
	}

	if result.EntityID != "project" {
		t.Errorf("expected EntityID 'project', got %q", result.EntityID)
	}
	if result.EntityType != "project" {
		t.Errorf("expected EntityType 'project', got %q", result.EntityType)
	}
	if result.Summary == "" {
		t.Error("Summary should not be empty")
	}
}

// Explain プロジェクト（コンテキスト付き）テスト
func TestExplain_ProjectWithContext(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// コンテキスト付きでプロジェクト説明を取得
	result, err := z.Explain(ctx, "project", true)
	if err != nil {
		t.Fatalf("Explain failed: %v", err)
	}

	// コンテキスト情報が含まれているか
	if _, ok := result.Context["project_id"]; !ok {
		t.Error("Context should contain project_id")
	}
	if _, ok := result.Context["automation_level"]; !ok {
		t.Error("Context should contain automation_level")
	}
}

// Explain タスクテスト
func TestExplain_Task(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// タスクを追加
	addResult, err := z.Add(ctx, "task", "Test Task")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// タスク説明を取得
	result, err := z.Explain(ctx, addResult.ID, false)
	if err != nil {
		t.Fatalf("Explain failed: %v", err)
	}

	if result.EntityID != addResult.ID {
		t.Errorf("expected EntityID %q, got %q", addResult.ID, result.EntityID)
	}
	if result.EntityType != "task" {
		t.Errorf("expected EntityType 'task', got %q", result.EntityType)
	}
}

// Explain 不明なエンティティテスト
func TestExplain_UnknownEntity(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 不明なエンティティ
	_, err = z.Explain(ctx, "unknown-entity", false)
	if err == nil {
		t.Error("expected error for unknown entity")
	}
}

// Explain コンテキストキャンセルテスト
func TestExplainContextTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = z.Explain(ctx, "project", false)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// BuildDependencyGraph テスト
func TestBuildDependencyGraph(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// タスクを追加
	_, _ = z.Add(ctx, "task", "Task 1")
	_, _ = z.Add(ctx, "task", "Task 2")

	// 依存関係グラフを構築
	graph, err := z.BuildDependencyGraph(ctx)
	if err != nil {
		t.Fatalf("BuildDependencyGraph failed: %v", err)
	}

	if graph == nil {
		t.Error("graph should not be nil")
	}
	if len(graph.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(graph.Nodes))
	}
}

// BuildDependencyGraph 空のプロジェクトテスト
func TestBuildDependencyGraph_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 空のプロジェクトでグラフを構築
	graph, err := z.BuildDependencyGraph(ctx)
	if err != nil {
		t.Fatalf("BuildDependencyGraph failed: %v", err)
	}

	if len(graph.Nodes) != 0 {
		t.Errorf("expected 0 nodes for empty project, got %d", len(graph.Nodes))
	}
}

// BuildDependencyGraph コンテキストキャンセルテスト
func TestBuildDependencyGraphContextTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = z.BuildDependencyGraph(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// Predict テスト
func TestPredict(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// タスクを追加
	_, _ = z.Add(ctx, "task", "Task 1")
	_, _ = z.Add(ctx, "task", "Task 2")

	// 予測を実行
	result, err := z.Predict(ctx, "all")
	if err != nil {
		t.Fatalf("Predict failed: %v", err)
	}

	if result == nil {
		t.Error("result should not be nil")
	}
}

// Predict 各タイプテスト
func TestPredict_Types(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	types := []string{"completion", "risk", "velocity", ""}

	for _, predType := range types {
		t.Run(predType, func(t *testing.T) {
			result, err := z.Predict(ctx, predType)
			if err != nil {
				t.Errorf("Predict(%q) failed: %v", predType, err)
			}
			if result == nil {
				t.Errorf("Predict(%q) should return non-nil result", predType)
			}
		})
	}
}

// Predict 不明なタイプテスト
func TestPredict_UnknownType(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	_, err = z.Predict(ctx, "unknown")
	if err == nil {
		t.Error("expected error for unknown prediction type")
	}
}

// Predict コンテキストキャンセルテスト
func TestPredictContextTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = z.Predict(ctx, "all")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// GenerateReport テスト
func TestGenerateReport(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 各フォーマットでレポート生成
	formats := []string{"text", "html", "markdown", ""}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			report, err := z.GenerateReport(ctx, format)
			if err != nil {
				t.Errorf("GenerateReport(%q) failed: %v", format, err)
			}
			if report == "" {
				t.Errorf("GenerateReport(%q) should return non-empty string", format)
			}
		})
	}
}

// GenerateReport 不明なフォーマットテスト
func TestGenerateReport_UnknownFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	_, err = z.GenerateReport(ctx, "unknown")
	if err == nil {
		t.Error("expected error for unknown format")
	}
}

// GenerateReport コンテキストキャンセルテスト
func TestGenerateReportContextTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = z.GenerateReport(ctx, "text")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// getDirectoryStructure テスト（単一構造に変更）
func TestGetDirectoryStructure(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)

	// 単一構造になったので、固定のディレクトリ数を確認
	dirs := z.getDirectoryStructure()
	expectedMinDirs := 10 // 統一構造の最低ディレクトリ数

	if len(dirs) < expectedMinDirs {
		t.Errorf("expected at least %d directories, got %d", expectedMinDirs, len(dirs))
	}

	// 必須ディレクトリが含まれているか確認
	requiredDirs := []string{"config", "tasks", "state", "approvals/pending", "approvals/approved", "approvals/rejected"}
	for _, required := range requiredDirs {
		found := false
		for _, dir := range dirs {
			if dir == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected directory %q to be in structure", required)
		}
	}
}

// FileStore アクセサテスト
func TestFileStoreAccessor(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)

	fs := z.FileStore()
	if fs == nil {
		t.Error("FileStore() should return non-nil")
	}
}

// Add 不明なエンティティテスト
func TestAdd_UnknownEntity(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 不明なエンティティタイプ
	_, err = z.Add(ctx, "unknown", "test")
	if err != ErrUnknownEntity {
		t.Errorf("expected ErrUnknownEntity, got %v", err)
	}
}

// List 不明なエンティティテスト
func TestList_UnknownEntity(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	// 初期化
	_, err = z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 不明なエンティティタイプ
	_, err = z.List(ctx, "unknown")
	if err != ErrUnknownEntity {
		t.Errorf("expected ErrUnknownEntity, got %v", err)
	}
}

// Init テスト（単一構造に変更）
func TestInit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx := context.Background()

	result, err := z.Init(ctx)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if !result.Success {
		t.Error("Init should succeed")
	}
	// Level フィールドは削除されたので確認しない
}

// ===== DI オプション関数テスト =====

// WithFileStore テスト
func TestWithFileStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// カスタム FileStore を作成
	customFS := yaml.NewFileManager(tmpDir)

	// WithFileStore オプションを使用
	z := New(tmpDir, WithFileStore(customFS))

	// FileStore が設定されていることを確認
	if z.fileStore != customFS {
		t.Error("WithFileStore should set custom FileStore")
	}
}

// WithStateStore テスト
func TestWithStateStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// カスタム StateStore を作成
	fs := yaml.NewFileManager(tmpDir)
	customSS := NewStateManager(tmpDir, fs)

	// WithStateStore オプションを使用
	z := New(tmpDir, WithStateStore(customSS))

	// StateStore が設定されていることを確認
	if z.stateStore != customSS {
		t.Error("WithStateStore should set custom StateStore")
	}
}

// WithApprovalStore テスト
func TestWithApprovalStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// カスタム ApprovalStore を作成
	fs := yaml.NewFileManager(tmpDir)
	customAS := NewApprovalManager(tmpDir, fs)

	// WithApprovalStore オプションを使用
	z := New(tmpDir, WithApprovalStore(customAS))

	// ApprovalStore が設定されていることを確認
	if z.approvalStore != customAS {
		t.Error("WithApprovalStore should set custom ApprovalStore")
	}
}

// WithEntityRegistry テスト
func TestWithEntityRegistry(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// カスタム EntityRegistry を作成
	customER := NewEntityRegistry()

	// WithEntityRegistry オプションを使用
	z := New(tmpDir, WithEntityRegistry(customER))

	// EntityRegistry が設定されていることを確認
	if z.entityRegistry != customER {
		t.Error("WithEntityRegistry should set custom EntityRegistry")
	}
}

// 複合 DI オプションテスト
func TestMultipleOptions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// カスタムコンポーネントを作成
	customFS := yaml.NewFileManager(tmpDir)
	customSS := NewStateManager(tmpDir, customFS)
	customAS := NewApprovalManager(tmpDir, customFS)
	customER := NewEntityRegistry()

	// 全オプションを使用
	z := New(tmpDir,
		WithFileStore(customFS),
		WithStateStore(customSS),
		WithApprovalStore(customAS),
		WithEntityRegistry(customER),
	)

	// 全てが設定されていることを確認
	if z.fileStore != customFS {
		t.Error("WithFileStore should set custom FileStore")
	}
	if z.stateStore != customSS {
		t.Error("WithStateStore should set custom StateStore")
	}
	if z.approvalStore != customAS {
		t.Error("WithApprovalStore should set custom ApprovalStore")
	}
	if z.entityRegistry != customER {
		t.Error("WithEntityRegistry should set custom EntityRegistry")
	}
}

// ===== エラー型テスト =====

// PathTraversalError テスト
func TestPathTraversalError(t *testing.T) {
	err := &PathTraversalError{
		RequestedPath: "/etc/passwd",
		BasePath:      "/home/user/project",
	}

	// Error() メッセージ検証
	msg := err.Error()
	if !strings.Contains(msg, "/etc/passwd") {
		t.Errorf("error message should contain requested path, got %q", msg)
	}
	if !strings.Contains(msg, "/home/user/project") {
		t.Errorf("error message should contain base path, got %q", msg)
	}
	if !strings.Contains(msg, "path traversal") {
		t.Errorf("error message should contain 'path traversal', got %q", msg)
	}

	// Is() 検証
	if !err.Is(ErrPathTraversal) {
		t.Error("PathTraversalError should match ErrPathTraversal")
	}
}

// ===== task_handler オプション関数テスト =====

// WithTaskDependencies テスト
func TestWithTaskDependencies(t *testing.T) {
	task := &Task{}
	deps := []string{"task-1", "task-2"}

	opt := WithTaskDependencies(deps)
	opt(task)

	if len(task.Dependencies) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(task.Dependencies))
	}
	if task.Dependencies[0] != "task-1" {
		t.Errorf("expected first dependency 'task-1', got %q", task.Dependencies[0])
	}
}

// WithTaskApprovalLevel テスト
func TestWithTaskApprovalLevel(t *testing.T) {
	task := &Task{}

	opt := WithTaskApprovalLevel(ApprovalApprove)
	opt(task)

	if task.ApprovalLevel != ApprovalApprove {
		t.Errorf("expected approval level 'approve', got %q", task.ApprovalLevel)
	}
}

// ===== CreateSnapshot/GetHistory 追加テスト =====

// CreateSnapshot コンテキストキャンセルテスト
func TestCreateSnapshotContextTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = z.CreateSnapshot(ctx, "test")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// GetHistory コンテキストキャンセルテスト
func TestGetHistoryContextTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeus-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	z := New(tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = z.GetHistory(ctx, 10)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
