package analysis

import (
	"context"
	"testing"
	"time"
)

// ===== NewStaleAnalyzer テスト =====

func TestNewStaleAnalyzer(t *testing.T) {
	tasks := []TaskInfo{{ID: "task-001", Title: "タスク1"}}
	objectives := []ObjectiveInfo{{ID: "obj-001", Title: "目標1"}}

	analyzer := NewStaleAnalyzer(tasks, objectives, nil)

	if analyzer == nil {
		t.Fatal("NewStaleAnalyzer returned nil")
	}
	if len(analyzer.tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(analyzer.tasks))
	}
	if len(analyzer.objectives) != 1 {
		t.Errorf("expected 1 objective, got %d", len(analyzer.objectives))
	}
}

func TestNewStaleAnalyzer_WithConfig(t *testing.T) {
	config := &StaleAnalyzerConfig{
		CompletedStaleDays: 60,
		BlockedStaleDays:   21,
		NoProgressDays:     30,
	}

	analyzer := NewStaleAnalyzer(nil, nil, config)

	if analyzer.config.CompletedStaleDays != 60 {
		t.Errorf("expected CompletedStaleDays 60, got %d", analyzer.config.CompletedStaleDays)
	}
	if analyzer.config.BlockedStaleDays != 21 {
		t.Errorf("expected BlockedStaleDays 21, got %d", analyzer.config.BlockedStaleDays)
	}
	if analyzer.config.NoProgressDays != 30 {
		t.Errorf("expected NoProgressDays 30, got %d", analyzer.config.NoProgressDays)
	}
}

func TestNewStaleAnalyzer_DefaultConfig(t *testing.T) {
	analyzer := NewStaleAnalyzer(nil, nil, nil)

	if analyzer.config.CompletedStaleDays != DefaultStaleConfig.CompletedStaleDays {
		t.Errorf("expected default CompletedStaleDays %d, got %d",
			DefaultStaleConfig.CompletedStaleDays, analyzer.config.CompletedStaleDays)
	}
	if analyzer.config.BlockedStaleDays != DefaultStaleConfig.BlockedStaleDays {
		t.Errorf("expected default BlockedStaleDays %d, got %d",
			DefaultStaleConfig.BlockedStaleDays, analyzer.config.BlockedStaleDays)
	}
	if analyzer.config.NoProgressDays != DefaultStaleConfig.NoProgressDays {
		t.Errorf("expected default NoProgressDays %d, got %d",
			DefaultStaleConfig.NoProgressDays, analyzer.config.NoProgressDays)
	}
}

// ===== Analyze テスト =====

func TestStaleAnalyzer_Analyze(t *testing.T) {
	analyzer := NewStaleAnalyzer(nil, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze returned nil result")
	}
	if result.StaleEntities == nil {
		t.Error("StaleEntities should not be nil")
	}
}

func TestStaleAnalyzer_Analyze_ContextCancellation(t *testing.T) {
	analyzer := NewStaleAnalyzer(nil, nil, nil)

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

// ===== 完了後陳腐化テスト =====

func TestStaleAnalyzer_CompletedOld_Task(t *testing.T) {
	// 31日前に完了したタスク
	completedAt := time.Now().AddDate(0, 0, -31).Format(time.RFC3339)
	tasks := []TaskInfo{
		{ID: "task-001", Title: "古いタスク", Status: TaskStatusCompleted, CompletedAt: completedAt},
	}

	analyzer := NewStaleAnalyzer(tasks, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	staleCount := 0
	for _, entity := range result.StaleEntities {
		if entity.Type == StaleTypeCompletedOld && entity.EntityID == "task-001" {
			staleCount++
			if entity.Recommendation != StaleRecommendArchive {
				t.Errorf("expected Archive recommendation, got %s", entity.Recommendation)
			}
		}
	}

	if staleCount == 0 {
		t.Error("expected completed old task to be detected")
	}
}

func TestStaleAnalyzer_CompletedOld_Objective(t *testing.T) {
	// 31日前に完了した Objective
	updatedAt := time.Now().AddDate(0, 0, -31).Format(time.RFC3339)
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "古い目標", Status: "completed", UpdatedAt: updatedAt},
	}

	analyzer := NewStaleAnalyzer(nil, objectives, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	staleCount := 0
	for _, entity := range result.StaleEntities {
		if entity.Type == StaleTypeCompletedOld && entity.EntityID == "obj-001" {
			staleCount++
		}
	}

	if staleCount == 0 {
		t.Error("expected completed old objective to be detected")
	}
}

// ===== ブロック長期化テスト =====

func TestStaleAnalyzer_BlockedLong(t *testing.T) {
	// 15日前からブロック状態
	updatedAt := time.Now().AddDate(0, 0, -15).Format(time.RFC3339)
	tasks := []TaskInfo{
		{ID: "task-001", Title: "長期ブロックタスク", Status: TaskStatusBlocked, UpdatedAt: updatedAt},
	}

	analyzer := NewStaleAnalyzer(tasks, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	staleCount := 0
	for _, entity := range result.StaleEntities {
		if entity.Type == StaleTypeBlockedLong && entity.EntityID == "task-001" {
			staleCount++
			if entity.Recommendation != StaleRecommendReview {
				t.Errorf("expected Review recommendation, got %s", entity.Recommendation)
			}
		}
	}

	if staleCount == 0 {
		t.Error("expected blocked long task to be detected")
	}
}

func TestStaleAnalyzer_BlockedLong_NotDetectedIfRecent(t *testing.T) {
	// 5日前からブロック状態（まだ陳腐化していない）
	updatedAt := time.Now().AddDate(0, 0, -5).Format(time.RFC3339)
	tasks := []TaskInfo{
		{ID: "task-001", Title: "最近ブロックタスク", Status: TaskStatusBlocked, UpdatedAt: updatedAt},
	}

	analyzer := NewStaleAnalyzer(tasks, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	for _, entity := range result.StaleEntities {
		if entity.Type == StaleTypeBlockedLong && entity.EntityID == "task-001" {
			t.Error("recently blocked task should not be detected as stale")
		}
	}
}

// ===== 孤立タスクテスト =====

func TestStaleAnalyzer_Orphaned(t *testing.T) {
	// 完了済みで孤立したタスク
	tasks := []TaskInfo{
		{ID: "task-001", Title: "孤立完了タスク", Status: TaskStatusCompleted, ParentID: ""},
	}

	analyzer := NewStaleAnalyzer(tasks, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	orphanCount := 0
	for _, entity := range result.StaleEntities {
		if entity.Type == StaleTypeOrphaned && entity.EntityID == "task-001" {
			orphanCount++
			if entity.Recommendation != StaleRecommendReview {
				t.Errorf("expected Review recommendation, got %s", entity.Recommendation)
			}
		}
	}

	if orphanCount == 0 {
		t.Error("expected orphaned completed task to be detected")
	}
}

func TestStaleAnalyzer_Orphaned_WithDependencies(t *testing.T) {
	// 依存関係があるタスクは孤立ではない
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク1", Status: TaskStatusCompleted, ParentID: "", Dependencies: []string{"task-002"}},
	}

	analyzer := NewStaleAnalyzer(tasks, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	for _, entity := range result.StaleEntities {
		if entity.Type == StaleTypeOrphaned && entity.EntityID == "task-001" {
			t.Error("task with dependencies should not be orphaned")
		}
	}
}

func TestStaleAnalyzer_Orphaned_Referenced(t *testing.T) {
	// 他から参照されているタスクは孤立ではない
	tasks := []TaskInfo{
		{ID: "task-001", Title: "参照されるタスク", Status: TaskStatusCompleted, ParentID: ""},
		{ID: "task-002", Title: "参照するタスク", Dependencies: []string{"task-001"}},
	}

	analyzer := NewStaleAnalyzer(tasks, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	for _, entity := range result.StaleEntities {
		if entity.Type == StaleTypeOrphaned && entity.EntityID == "task-001" {
			t.Error("referenced task should not be orphaned")
		}
	}
}

// ===== 推奨アクションカウントテスト =====

func TestStaleAnalyzer_RecommendationCounts(t *testing.T) {
	completedAt := time.Now().AddDate(0, 0, -31).Format(time.RFC3339)
	blockedAt := time.Now().AddDate(0, 0, -15).Format(time.RFC3339)

	tasks := []TaskInfo{
		// Archive 推奨
		{ID: "task-001", Title: "古いタスク", Status: TaskStatusCompleted, CompletedAt: completedAt},
		// Review 推奨
		{ID: "task-002", Title: "ブロックタスク", Status: TaskStatusBlocked, UpdatedAt: blockedAt},
	}

	analyzer := NewStaleAnalyzer(tasks, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.ArchiveCount == 0 {
		t.Error("expected at least one archive recommendation")
	}
	if result.ReviewCount == 0 {
		t.Error("expected at least one review recommendation")
	}
	if result.TotalStale != len(result.StaleEntities) {
		t.Errorf("TotalStale %d does not match StaleEntities count %d",
			result.TotalStale, len(result.StaleEntities))
	}
}

// ===== 設定可能日数テスト =====

func TestStaleAnalyzer_ConfigurableDays(t *testing.T) {
	// 10日前に完了（デフォルトでは陳腐化しないが、カスタム設定で陳腐化）
	completedAt := time.Now().AddDate(0, 0, -10).Format(time.RFC3339)
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク", Status: TaskStatusCompleted, CompletedAt: completedAt},
	}

	config := &StaleAnalyzerConfig{
		CompletedStaleDays: 7, // 7日で陳腐化
	}

	analyzer := NewStaleAnalyzer(tasks, nil, config)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	staleCount := 0
	for _, entity := range result.StaleEntities {
		if entity.Type == StaleTypeCompletedOld && entity.EntityID == "task-001" {
			staleCount++
		}
	}

	if staleCount == 0 {
		t.Error("expected task to be stale with custom config")
	}
}

// ===== 陳腐化タイプテスト =====

func TestStaleType_Values(t *testing.T) {
	testCases := []struct {
		staleType StaleType
		expected  string
	}{
		{StaleTypeCompletedOld, "completed_old"},
		{StaleTypeOrphaned, "orphaned"},
		{StaleTypeBlockedLong, "blocked_long"},
		{StaleTypeNoProgress, "no_progress"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			if string(tc.staleType) != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, string(tc.staleType))
			}
		})
	}
}

func TestStaleRecommendation_Values(t *testing.T) {
	testCases := []struct {
		recommendation StaleRecommendation
		expected       string
	}{
		{StaleRecommendArchive, "archive"},
		{StaleRecommendReview, "review"},
		{StaleRecommendDelete, "delete"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			if string(tc.recommendation) != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, string(tc.recommendation))
			}
		})
	}
}

// ===== 日付パーステスト =====

func TestStaleAnalyzer_ParseDate_RFC3339(t *testing.T) {
	analyzer := NewStaleAnalyzer(nil, nil, nil)

	dateStr := "2024-01-15T10:30:00Z"
	result := analyzer.parseDate(dateStr)

	if result == nil {
		t.Error("expected non-nil result for RFC3339 format")
	}
}

func TestStaleAnalyzer_ParseDate_DateOnly(t *testing.T) {
	analyzer := NewStaleAnalyzer(nil, nil, nil)

	dateStr := "2024-01-15"
	result := analyzer.parseDate(dateStr)

	if result == nil {
		t.Error("expected non-nil result for date-only format")
	}
}

func TestStaleAnalyzer_ParseDate_Empty(t *testing.T) {
	analyzer := NewStaleAnalyzer(nil, nil, nil)

	result := analyzer.parseDate("")

	if result != nil {
		t.Error("expected nil result for empty string")
	}
}

func TestStaleAnalyzer_ParseDate_Invalid(t *testing.T) {
	analyzer := NewStaleAnalyzer(nil, nil, nil)

	result := analyzer.parseDate("invalid-date")

	if result != nil {
		t.Error("expected nil result for invalid date")
	}
}

// ===== DaysStale テスト =====

func TestStaleAnalyzer_DaysStale(t *testing.T) {
	// 45日前に完了
	completedAt := time.Now().AddDate(0, 0, -45).Format(time.RFC3339)
	tasks := []TaskInfo{
		{ID: "task-001", Title: "タスク", Status: TaskStatusCompleted, CompletedAt: completedAt},
	}

	analyzer := NewStaleAnalyzer(tasks, nil, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	for _, entity := range result.StaleEntities {
		if entity.EntityID == "task-001" {
			if entity.DaysStale < 44 || entity.DaysStale > 46 {
				t.Errorf("expected DaysStale around 45, got %d", entity.DaysStale)
			}
		}
	}
}

// ===== 複合シナリオテスト =====

func TestStaleAnalyzer_ComplexScenario(t *testing.T) {
	completedAt := time.Now().AddDate(0, 0, -35).Format(time.RFC3339)
	blockedAt := time.Now().AddDate(0, 0, -20).Format(time.RFC3339)
	objUpdatedAt := time.Now().AddDate(0, 0, -40).Format(time.RFC3339)

	tasks := []TaskInfo{
		// CompletedOld
		{ID: "task-001", Title: "古いタスク", Status: TaskStatusCompleted, CompletedAt: completedAt},
		// BlockedLong
		{ID: "task-002", Title: "ブロックタスク", Status: TaskStatusBlocked, UpdatedAt: blockedAt},
		// Orphaned
		{ID: "task-003", Title: "孤立タスク", Status: TaskStatusCompleted, ParentID: ""},
	}
	objectives := []ObjectiveInfo{
		{ID: "obj-001", Title: "古い目標", Status: "completed", UpdatedAt: objUpdatedAt},
	}

	analyzer := NewStaleAnalyzer(tasks, objectives, nil)
	ctx := context.Background()

	result, err := analyzer.Analyze(ctx)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	staleTypes := make(map[StaleType]int)
	for _, entity := range result.StaleEntities {
		staleTypes[entity.Type]++
	}

	t.Logf("Stale types: %v", staleTypes)
	t.Logf("Total: %d, Archive: %d, Review: %d, Delete: %d",
		result.TotalStale, result.ArchiveCount, result.ReviewCount, result.DeleteCount)

	// 各タイプが検出されることを確認
	if staleTypes[StaleTypeCompletedOld] < 2 {
		t.Error("expected at least 2 completed old entities")
	}
	if staleTypes[StaleTypeBlockedLong] == 0 {
		t.Error("expected blocked long entity")
	}
}
