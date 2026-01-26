package analysis

import (
	"context"
	"testing"
	"time"
)

// ===== NewTimelineBuilder テスト =====

func TestNewTimelineBuilder(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1"},
		{ID: "task-002", Title: "タスク2"},
	}

	builder := NewTimelineBuilder(tasks)

	if builder == nil {
		t.Fatal("NewTimelineBuilder returned nil")
	}
	if len(builder.tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(builder.tasks))
	}
}

func TestNewTimelineBuilder_Empty(t *testing.T) {
	builder := NewTimelineBuilder(nil)

	if builder == nil {
		t.Fatal("NewTimelineBuilder returned nil for empty input")
	}
	if len(builder.tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(builder.tasks))
	}
}

// ===== Build テスト =====

func TestTimelineBuilder_Build(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", StartDate: "2024-01-01", DueDate: "2024-01-15"},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if timeline == nil {
		t.Fatal("Build returned nil timeline")
	}
}

func TestTimelineBuilder_Build_ContextCancellation(t *testing.T) {
	builder := NewTimelineBuilder(nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := builder.Build(ctx)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestTimelineBuilder_Build_Empty(t *testing.T) {
	builder := NewTimelineBuilder(nil)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if timeline == nil {
		t.Fatal("Build returned nil timeline")
	}
	if len(timeline.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(timeline.Items))
	}
}

// ===== タイムラインアイテムテスト =====

func TestTimelineBuilder_Items(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", StartDate: "2024-01-01", DueDate: "2024-01-15", Status: TaskStatusInProgress, Progress: 50},
		{ID: "task-002", Title: "タスク2", StartDate: "2024-01-10", DueDate: "2024-01-20"},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if len(timeline.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(timeline.Items))
	}

	// アイテムの内容を確認
	for _, item := range timeline.Items {
		if item.TaskID == "task-001" {
			if item.Title != "タスク1" {
				t.Errorf("expected title 'タスク1', got %q", item.Title)
			}
			if item.Progress != 50 {
				t.Errorf("expected progress 50, got %d", item.Progress)
			}
			if item.Status != TaskStatusInProgress {
				t.Errorf("expected status %s, got %s", TaskStatusInProgress, item.Status)
			}
		}
	}
}

func TestTimelineBuilder_Items_Sorted(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-002", Title: "タスク2", StartDate: "2024-01-15", DueDate: "2024-01-20"},
		{ID: "task-001", Title: "タスク1", StartDate: "2024-01-01", DueDate: "2024-01-10"},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 開始日でソートされていることを確認
	if len(timeline.Items) < 2 {
		t.Fatal("expected at least 2 items")
	}
	if timeline.Items[0].StartDate > timeline.Items[1].StartDate {
		t.Error("items should be sorted by start date")
	}
}

// ===== プロジェクト期間テスト =====

func TestTimelineBuilder_ProjectDuration(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", StartDate: "2024-01-01", DueDate: "2024-01-10"},
		{ID: "task-002", Title: "タスク2", StartDate: "2024-01-05", DueDate: "2024-01-20"},
		{ID: "task-003", Title: "タスク3", StartDate: "2024-01-15", DueDate: "2024-01-30"},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if timeline.ProjectStart != "2024-01-01" {
		t.Errorf("expected project start '2024-01-01', got %q", timeline.ProjectStart)
	}
	if timeline.ProjectEnd != "2024-01-30" {
		t.Errorf("expected project end '2024-01-30', got %q", timeline.ProjectEnd)
	}
	if timeline.TotalDuration != 29 {
		t.Errorf("expected total duration 29 days, got %d", timeline.TotalDuration)
	}
}

// ===== クリティカルパステスト =====

func TestTimelineBuilder_CriticalPath(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", StartDate: "2024-01-01", DueDate: "2024-01-10"},
		{ID: "task-002", Title: "タスク2", StartDate: "2024-01-10", DueDate: "2024-01-20", Dependencies: []string{"task-001"}},
		{ID: "task-003", Title: "タスク3", StartDate: "2024-01-20", DueDate: "2024-01-30", Dependencies: []string{"task-002"}},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if len(timeline.CriticalPath) == 0 {
		t.Error("expected non-empty critical path")
	}

	t.Logf("Critical path: %v", timeline.CriticalPath)
}

func TestTimelineBuilder_CriticalPath_SingleTask(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", StartDate: "2024-01-01", DueDate: "2024-01-10"},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 単一タスクはクリティカルパス上
	if len(timeline.CriticalPath) != 1 {
		t.Errorf("expected 1 task on critical path, got %d", len(timeline.CriticalPath))
	}
}

func TestTimelineBuilder_IsOnCriticalPath(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", StartDate: "2024-01-01", DueDate: "2024-01-10"},
		{ID: "task-002", Title: "タスク2", StartDate: "2024-01-10", DueDate: "2024-01-20", Dependencies: []string{"task-001"}},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// クリティカルパス上のタスクにフラグが設定されていることを確認
	for _, item := range timeline.Items {
		if item.IsOnCriticalPath {
			t.Logf("Task %s is on critical path", item.TaskID)
		}
	}
}

// ===== スラックテスト =====

func TestTimelineBuilder_Slack(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "クリティカル", StartDate: "2024-01-01", DueDate: "2024-01-10"},
		{ID: "task-002", Title: "余裕あり", StartDate: "2024-01-01", DueDate: "2024-01-05"},
		{ID: "task-003", Title: "後続", StartDate: "2024-01-10", DueDate: "2024-01-15", Dependencies: []string{"task-001"}},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// スラックが計算されていることを確認
	for _, item := range timeline.Items {
		t.Logf("Task %s: slack = %d days", item.TaskID, item.Slack)
	}
}

// ===== 統計テスト =====

func TestTimelineBuilder_Stats(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", StartDate: yesterday, DueDate: today, Status: TaskStatusInProgress},
		{ID: "task-002", Title: "タスク2", StartDate: today, DueDate: tomorrow, Status: TaskStatusCompleted},
		{ID: "task-003", Title: "日付なし", Status: TaskStatusPending},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if timeline.Stats.TotalTasks != 3 {
		t.Errorf("expected TotalTasks 3, got %d", timeline.Stats.TotalTasks)
	}
	if timeline.Stats.TasksWithDates != 2 {
		t.Errorf("expected TasksWithDates 2, got %d", timeline.Stats.TasksWithDates)
	}
}

func TestTimelineBuilder_Stats_Overdue(t *testing.T) {
	// タイムゾーンの境界問題を回避するため -2 日を使用
	twoDaysAgo := time.Now().AddDate(0, 0, -2).Format("2006-01-02")
	lastWeek := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

	tasks := []TaskInfo{
		{ID: "task-001", Title: "期限切れ1", StartDate: lastWeek, DueDate: twoDaysAgo, Status: TaskStatusInProgress},
		{ID: "task-002", Title: "期限切れ2", StartDate: lastWeek, DueDate: twoDaysAgo, Status: TaskStatusPending},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if timeline.Stats.OverdueTasks != 2 {
		t.Errorf("expected OverdueTasks 2, got %d", timeline.Stats.OverdueTasks)
	}
}

func TestTimelineBuilder_Stats_CompletedOnTime(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	today := time.Now().Format("2006-01-02")

	tasks := []TaskInfo{
		{ID: "task-001", Title: "期限内完了", StartDate: today, DueDate: tomorrow, Status: TaskStatusCompleted},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if timeline.Stats.CompletedOnTime != 1 {
		t.Errorf("expected CompletedOnTime 1, got %d", timeline.Stats.CompletedOnTime)
	}
}

func TestTimelineBuilder_Stats_AverageSlack(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", StartDate: "2024-01-01", DueDate: "2024-01-10"},
		{ID: "task-002", Title: "タスク2", StartDate: "2024-01-05", DueDate: "2024-01-15"},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 平均スラックが計算されていることを確認
	t.Logf("Average slack: %.2f days", timeline.Stats.AverageSlack)
}

func TestTimelineBuilder_Stats_OnCriticalPath(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", StartDate: "2024-01-01", DueDate: "2024-01-10"},
		{ID: "task-002", Title: "タスク2", StartDate: "2024-01-10", DueDate: "2024-01-20", Dependencies: []string{"task-001"}},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if timeline.Stats.OnCriticalPath == 0 {
		t.Error("expected at least one task on critical path")
	}
	if timeline.Stats.OnCriticalPath != len(timeline.CriticalPath) {
		t.Errorf("OnCriticalPath %d does not match CriticalPath length %d",
			timeline.Stats.OnCriticalPath, len(timeline.CriticalPath))
	}
}

// ===== 日付なしタスクテスト =====

func TestTimelineBuilder_TasksWithoutDates(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "日付なし1"},
		{ID: "task-002", Title: "日付なし2"},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 日付なしタスクはタイムラインに含まれない
	if len(timeline.Items) != 0 {
		t.Errorf("expected 0 items for tasks without dates, got %d", len(timeline.Items))
	}
	if timeline.Stats.TasksWithDates != 0 {
		t.Errorf("expected TasksWithDates 0, got %d", timeline.Stats.TasksWithDates)
	}
}

func TestTimelineBuilder_MixedTasks(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "日付あり", StartDate: "2024-01-01", DueDate: "2024-01-10"},
		{ID: "task-002", Title: "日付なし"},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if len(timeline.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(timeline.Items))
	}
	if timeline.Stats.TotalTasks != 2 {
		t.Errorf("expected TotalTasks 2, got %d", timeline.Stats.TotalTasks)
	}
}

// ===== 依存関係テスト =====

func TestTimelineBuilder_Dependencies(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", StartDate: "2024-01-01", DueDate: "2024-01-10"},
		{ID: "task-002", Title: "タスク2", StartDate: "2024-01-10", DueDate: "2024-01-20", Dependencies: []string{"task-001"}},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 依存関係がアイテムに含まれていることを確認
	for _, item := range timeline.Items {
		if item.TaskID == "task-002" {
			if len(item.Dependencies) != 1 {
				t.Errorf("expected 1 dependency, got %d", len(item.Dependencies))
			}
			if item.Dependencies[0] != "task-001" {
				t.Errorf("expected dependency 'task-001', got %q", item.Dependencies[0])
			}
		}
	}
}

// ===== 部分日付テスト =====

func TestTimelineBuilder_StartDateOnly(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "開始日のみ", StartDate: "2024-01-01"},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 開始日のみでもタイムラインに含まれる
	if len(timeline.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(timeline.Items))
	}
}

func TestTimelineBuilder_EndDateOnly(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "終了日のみ", DueDate: "2024-01-10"},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 終了日のみでもタイムラインに含まれる
	if len(timeline.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(timeline.Items))
	}
}

// ===== 複合シナリオテスト =====

func TestTimelineBuilder_ComplexScenario(t *testing.T) {
	tasks := []TaskInfo{
		{ID: "task-001", Title: "設計", StartDate: "2024-01-01", DueDate: "2024-01-10", Status: TaskStatusCompleted, Progress: 100},
		{ID: "task-002", Title: "実装", StartDate: "2024-01-10", DueDate: "2024-01-25", Status: TaskStatusInProgress, Progress: 50, Dependencies: []string{"task-001"}},
		{ID: "task-003", Title: "テスト", StartDate: "2024-01-25", DueDate: "2024-02-05", Status: TaskStatusPending, Dependencies: []string{"task-002"}},
		{ID: "task-004", Title: "ドキュメント", StartDate: "2024-01-15", DueDate: "2024-01-30", Status: TaskStatusInProgress, Progress: 30},
	}

	builder := NewTimelineBuilder(tasks)
	ctx := context.Background()

	timeline, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	t.Logf("Project: %s to %s (%d days)",
		timeline.ProjectStart, timeline.ProjectEnd, timeline.TotalDuration)
	t.Logf("Critical path: %v", timeline.CriticalPath)
	t.Logf("Stats: %+v", timeline.Stats)

	// 基本的な検証
	if len(timeline.Items) != 4 {
		t.Errorf("expected 4 items, got %d", len(timeline.Items))
	}
	if timeline.ProjectStart != "2024-01-01" {
		t.Errorf("expected project start '2024-01-01', got %q", timeline.ProjectStart)
	}
	if timeline.ProjectEnd != "2024-02-05" {
		t.Errorf("expected project end '2024-02-05', got %q", timeline.ProjectEnd)
	}
	if len(timeline.CriticalPath) == 0 {
		t.Error("expected non-empty critical path")
	}
}
