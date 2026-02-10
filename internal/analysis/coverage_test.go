package analysis

import (
	"context"
	"testing"
)

// ===== NewCoverageAnalyzer テスト =====

func TestNewCoverageAnalyzer(t *testing.T) {
	objectives := []ObjectiveInfo{{ID: "obj-001", Title: "目標1"}}
	tasks := []TaskInfo{{ID: "task-001", Title: "タスク1"}}

	analyzer := NewCoverageAnalyzer(objectives, tasks)

	if analyzer == nil {
		t.Fatal("NewCoverageAnalyzer returned nil")
	}
	if len(analyzer.objectives) != 1 {
		t.Errorf("expected 1 objective, got %d", len(analyzer.objectives))
	}
	if len(analyzer.tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(analyzer.tasks))
	}
}

func TestNewCoverageAnalyzer_Empty(t *testing.T) {
	analyzer := NewCoverageAnalyzer(nil, nil)

	if analyzer == nil {
		t.Fatal("NewCoverageAnalyzer returned nil for empty input")
	}
}

// ===== Analyze テスト =====

func TestCoverageAnalyzer_Analyze(t *testing.T) {
	objectives := []ObjectiveInfo{{ID: "obj-001", Title: "目標1"}}
	tasks := []TaskInfo{{ID: "task-001", Title: "タスク1", ParentID: "obj-001"}}

	analyzer := NewCoverageAnalyzer(objectives, tasks)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze returned nil result")
	}
	if result.Issues == nil {
		t.Error("Issues should not be nil")
	}
}

func TestCoverageAnalyzer_Analyze_ContextCancellation(t *testing.T) {
	analyzer := NewCoverageAnalyzer(nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := analyzer.Analyze(ctx)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// ===== Objective カバレッジテスト =====

func TestCoverageAnalyzer_NoTasks(t *testing.T) {
	// Task が紐づいていない Objective
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
	}

	analyzer := NewCoverageAnalyzer(objectives, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	issueCount := 0
	for _, issue := range result.Issues {
		if issue.Type == CoverageIssueNoTasks && issue.EntityID == "obj-001" {
			issueCount++
		}
	}

	if issueCount == 0 {
		t.Error("expected NoTasks issue for objective without tasks")
	}
}

func TestCoverageAnalyzer_ObjectiveCovered(t *testing.T) {
	// Task が紐づいている Objective
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
	}
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", ParentID: "obj-001"},
	}

	analyzer := NewCoverageAnalyzer(objectives, tasks)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.ObjectivesCover != 1 {
		t.Errorf("expected 1 objective covered, got %d", result.ObjectivesCover)
	}
}

// ===== 孤立タスクテスト =====

func TestCoverageAnalyzer_OrphanedTasks(t *testing.T) {
	// 親がいない孤立タスク
	tasks := []TaskInfo{
		{ID: "task-001", Title: "孤立タスク", ParentID: ""},
	}

	analyzer := NewCoverageAnalyzer(nil, tasks)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	orphanCount := 0
	for _, issue := range result.Issues {
		if issue.Type == CoverageIssueOrphaned && issue.EntityID == "task-001" {
			orphanCount++
		}
	}

	if orphanCount == 0 {
		t.Error("expected orphaned issue for task without parent")
	}
}

func TestCoverageAnalyzer_TaskWithParent(t *testing.T) {
	// 親タスクがいるタスクは孤立ではない
	tasks := []TaskInfo{
		{ID: "task-001", Title: "親タスク", ParentID: ""},
		{ID: "task-002", Title: "子タスク", ParentID: "task-001"},
	}

	analyzer := NewCoverageAnalyzer(nil, tasks)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	for _, issue := range result.Issues {
		if issue.Type == CoverageIssueOrphaned && issue.EntityID == "task-002" {
			t.Error("task with parent should not be orphaned")
		}
	}
}

// ===== カバレッジスコアテスト =====

func TestCoverageAnalyzer_CoverageScore_FullCoverage(t *testing.T) {
	// 完全カバレッジ
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
	}
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", ParentID: "obj-001"},
	}

	analyzer := NewCoverageAnalyzer(objectives, tasks)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.CoverageScore != 100 {
		t.Errorf("expected coverage score 100, got %d", result.CoverageScore)
	}
}

func TestCoverageAnalyzer_CoverageScore_NoCoverage(t *testing.T) {
	// カバレッジなし（Objective に Task が紐づいていない）
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
	}

	analyzer := NewCoverageAnalyzer(objectives, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.CoverageScore == 100 {
		t.Error("expected coverage score less than 100 for no coverage")
	}
}

func TestCoverageAnalyzer_CoverageScore_Empty(t *testing.T) {
	// 空のデータ
	analyzer := NewCoverageAnalyzer(nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.CoverageScore != 100 {
		t.Errorf("expected coverage score 100 for empty data, got %d", result.CoverageScore)
	}
}

func TestCoverageAnalyzer_CoverageScore_TasksOnly(t *testing.T) {
	// タスクのみの場合
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", ParentID: ""},
		{ID: "task-002", Title: "タスク2", ParentID: ""},
	}

	analyzer := NewCoverageAnalyzer(nil, tasks)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// 全て孤立なのでスコアは 0
	if result.CoverageScore != 0 {
		t.Errorf("expected coverage score 0 for all orphan tasks, got %d", result.CoverageScore)
	}
}

// ===== Issue 種類テスト =====

func TestCoverageIssueType_Values(t *testing.T) {
	testCases := []struct {
		issueType CoverageIssueType
		expected  string
	}{
		{CoverageIssueNoTasks, "no_tasks"},
		{CoverageIssueOrphaned, "orphaned"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			if string(tc.issueType) != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, string(tc.issueType))
			}
		})
	}
}

func TestCoverageIssueSeverity_Values(t *testing.T) {
	testCases := []struct {
		severity CoverageIssueSeverity
		expected string
	}{
		{CoverageSeverityWarning, "warning"},
		{CoverageSeverityError, "error"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			if string(tc.severity) != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, string(tc.severity))
			}
		})
	}
}

// ===== 複合シナリオテスト =====

func TestCoverageAnalyzer_ComplexScenario(t *testing.T) {
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
		{ID: "obj-002", Title: "目標2"},
	}
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", ParentID: "obj-001"},
		{ID: "task-002", Title: "孤立タスク", ParentID: ""},
	}

	analyzer := NewCoverageAnalyzer(objectives, tasks)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// 問題の種類をカウント
	issueTypes := make(map[CoverageIssueType]int)
	for _, issue := range result.Issues {
		issueTypes[issue.Type]++
	}

	t.Logf("Issues: %v", issueTypes)
	t.Logf("Coverage: %d%%, ObjectivesCover: %d/%d",
		result.CoverageScore, result.ObjectivesCover, result.ObjectivesTotal)

	// obj-002 は Task なし
	if issueTypes[CoverageIssueNoTasks] == 0 {
		t.Error("expected NoTasks issue")
	}
	// task-002 は孤立
	if issueTypes[CoverageIssueOrphaned] == 0 {
		t.Error("expected Orphaned issue")
	}
}

// ===== 統計テスト =====

func TestCoverageAnalyzer_Statistics(t *testing.T) {
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "目標1"},
		{ID: "obj-002", Title: "目標2"},
	}
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", ParentID: "obj-001"},
	}

	analyzer := NewCoverageAnalyzer(objectives, tasks)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.ObjectivesTotal != 2 {
		t.Errorf("expected ObjectivesTotal 2, got %d", result.ObjectivesTotal)
	}
	if result.ObjectivesCover != 1 {
		t.Errorf("expected ObjectivesCover 1, got %d", result.ObjectivesCover)
	}
}
