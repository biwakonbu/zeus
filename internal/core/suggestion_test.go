package core

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// TestGenerateSuggestions_EmptyProject は空のプロジェクトでの提案生成をテスト
func TestGenerateSuggestions_EmptyProject(t *testing.T) {
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

	// Status を取得
	status, err := z.Status(ctx)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	// 提案を生成（空のプロジェクトでは提案がないはず）
	suggestions, err := z.GenerateSuggestions(ctx, status, 5, "")
	if err != nil {
		t.Fatalf("GenerateSuggestions failed: %v", err)
	}

	// 空のプロジェクトでは提案がないはず
	if len(suggestions) != 0 {
		t.Errorf("expected no suggestions for empty project, got %d", len(suggestions))
	}
}

// TestGenerateSuggestions_ManyPendingTasks は保留中タスクが多い場合のテスト
func TestGenerateSuggestions_ManyPendingTasks(t *testing.T) {
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

	// 6件以上の保留中タスクを追加（閾値は5）
	for i := 1; i <= 7; i++ {
		_, err := z.Add(ctx, "task", "Test Task")
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Status を取得
	status, err := z.Status(ctx)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	// 提案を生成
	suggestions, err := z.GenerateSuggestions(ctx, status, 5, "")
	if err != nil {
		t.Fatalf("GenerateSuggestions failed: %v", err)
	}

	// 保留中タスクに対する提案があるはず
	if len(suggestions) == 0 {
		t.Error("expected suggestions for many pending tasks")
	}

	// medium impact の提案があるか確認
	foundMedium := false
	for _, s := range suggestions {
		if s.Impact == ImpactMedium && s.Type == SuggestionPriorityChange {
			foundMedium = true
			break
		}
	}
	if !foundMedium {
		t.Error("expected medium-impact priority change suggestion")
	}
}

// TestGenerateSuggestions_ImpactFilter は impact フィルタのテスト
func TestGenerateSuggestions_ImpactFilter(t *testing.T) {
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

	// 7件の保留中タスクを追加
	for i := 1; i <= 7; i++ {
		_, err := z.Add(ctx, "task", "Test Task")
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	status, err := z.Status(ctx)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	// high フィルタでは medium 提案は含まれない
	suggestions, err := z.GenerateSuggestions(ctx, status, 5, "high")
	if err != nil {
		t.Fatalf("GenerateSuggestions failed: %v", err)
	}

	for _, s := range suggestions {
		if s.Impact != ImpactHigh {
			t.Errorf("expected only high impact suggestions, got %s", s.Impact)
		}
	}
}

// TestApplySuggestion_NotFound は存在しない提案の適用をテスト
func TestApplySuggestion_NotFound(t *testing.T) {
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

	// 存在しない提案を適用
	_, err = z.ApplySuggestion(ctx, "nonexistent-id", false, false)
	if err == nil {
		t.Error("expected error for nonexistent suggestion")
	}
	if !strings.Contains(err.Error(), "生成されていません") && !strings.Contains(err.Error(), "見つかりません") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// TestApplySuggestion_DryRun は dry-run モードのテスト
func TestApplySuggestion_DryRun(t *testing.T) {
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

	// 7件のタスクを追加して提案を生成
	for i := 1; i <= 7; i++ {
		z.Add(ctx, "task", "Test Task")
	}

	status, _ := z.Status(ctx)
	suggestions, err := z.GenerateSuggestions(ctx, status, 5, "")
	if err != nil {
		t.Fatalf("GenerateSuggestions failed: %v", err)
	}

	if len(suggestions) == 0 {
		t.Skip("no suggestions generated")
	}

	// dry-run で適用
	result, err := z.ApplySuggestion(ctx, suggestions[0].ID, false, true)
	if err != nil {
		t.Fatalf("ApplySuggestion dry-run failed: %v", err)
	}

	if result.Applied != 1 {
		t.Errorf("expected 1 applied in dry-run, got %d", result.Applied)
	}

	// dry-run 後も提案のステータスは pending のまま
	var store SuggestionStore
	if err := z.fileStore.ReadYaml(ctx, "suggestions/active.yaml", &store); err != nil {
		t.Fatalf("failed to read suggestions: %v", err)
	}

	for _, s := range store.Suggestions {
		if s.ID == suggestions[0].ID && s.Status != SuggestionPending {
			t.Errorf("expected status to remain pending after dry-run, got %s", s.Status)
		}
	}
}

// TestApplySuggestion_AllFlag は --all フラグのテスト
func TestApplySuggestion_AllFlag(t *testing.T) {
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

	// タスクを追加して提案を生成
	for i := 1; i <= 7; i++ {
		z.Add(ctx, "task", "Test Task")
	}

	status, _ := z.Status(ctx)
	suggestions, _ := z.GenerateSuggestions(ctx, status, 5, "")

	if len(suggestions) == 0 {
		t.Skip("no suggestions generated")
	}

	// --all で適用
	result, err := z.ApplySuggestion(ctx, "", true, false)
	if err != nil {
		t.Fatalf("ApplySuggestion --all failed: %v", err)
	}

	// RiskMitigation は適用可能、PriorityChange は Phase 3 で失敗
	if result.Applied+result.Failed != len(suggestions) {
		t.Errorf("expected total processed = %d, got applied=%d, failed=%d",
			len(suggestions), result.Applied, result.Failed)
	}
}

// TestTaskValidate はタスクの検証をテスト
func TestTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    Task
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid task",
			task: Task{
				ID:        "task-1",
				Title:     "Test Task",
				Status:    TaskStatusPending,
				CreatedAt: Now(),
				UpdatedAt: Now(),
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			task: Task{
				Title:  "Test Task",
				Status: TaskStatusPending,
			},
			wantErr: true,
			errMsg:  "task ID is required",
		},
		{
			name: "missing title",
			task: Task{
				ID:     "task-1",
				Status: TaskStatusPending,
			},
			wantErr: true,
			errMsg:  "task title is required",
		},
		{
			name: "missing status",
			task: Task{
				ID:    "task-1",
				Title: "Test Task",
			},
			wantErr: true,
			errMsg:  "task status is required",
		},
		{
			name: "negative estimate hours",
			task: Task{
				ID:            "task-1",
				Title:         "Test Task",
				Status:        TaskStatusPending,
				EstimateHours: -5,
			},
			wantErr: true,
			errMsg:  "estimate_hours must be non-negative",
		},
		{
			name: "invalid approval level",
			task: Task{
				ID:            "task-1",
				Title:         "Test Task",
				Status:        TaskStatusPending,
				ApprovalLevel: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid approval level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, want error containing %s", err, tt.errMsg)
			}
		})
	}
}

// TestSuggestionValidate は提案の検証をテスト
func TestSuggestionValidate(t *testing.T) {
	validTask := &Task{
		ID:        "task-1",
		Title:     "Test Task",
		Status:    TaskStatusPending,
		CreatedAt: Now(),
		UpdatedAt: Now(),
	}

	tests := []struct {
		name    string
		sugg    Suggestion
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid risk mitigation",
			sugg: Suggestion{
				ID:          "sugg-1",
				Type:        SuggestionRiskMitigation,
				Description: "Risk description",
				Impact:      ImpactHigh,
			},
			wantErr: false,
		},
		{
			name: "valid new task",
			sugg: Suggestion{
				ID:          "sugg-1",
				Type:        SuggestionNewTask,
				Description: "New task suggestion",
				Impact:      ImpactMedium,
				TaskData:    validTask,
			},
			wantErr: false,
		},
		{
			name: "new task without task data",
			sugg: Suggestion{
				ID:          "sugg-1",
				Type:        SuggestionNewTask,
				Description: "New task suggestion",
				Impact:      ImpactMedium,
			},
			wantErr: true,
			errMsg:  "new_task suggestion must have TaskData",
		},
		{
			name: "priority change without target",
			sugg: Suggestion{
				ID:          "sugg-1",
				Type:        SuggestionPriorityChange,
				Description: "Priority change",
				Impact:      ImpactMedium,
			},
			wantErr: true,
			errMsg:  "priority_change suggestion must have TargetTaskID",
		},
		{
			name: "missing ID",
			sugg: Suggestion{
				Type:        SuggestionRiskMitigation,
				Description: "Risk",
				Impact:      ImpactHigh,
			},
			wantErr: true,
			errMsg:  "suggestion ID is required",
		},
		{
			name: "missing description",
			sugg: Suggestion{
				ID:     "sugg-1",
				Type:   SuggestionRiskMitigation,
				Impact: ImpactHigh,
			},
			wantErr: true,
			errMsg:  "suggestion description is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sugg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, want error containing %s", err, tt.errMsg)
			}
		})
	}
}

// TestGenerateSuggestionsContextTimeout はコンテキストタイムアウトをテスト
func TestGenerateSuggestionsContextTimeout(t *testing.T) {
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

	status, _ := z.Status(ctx)

	// キャンセル済みのコンテキスト
	canceledCtx, cancel := context.WithCancel(ctx)
	cancel()

	_, err = z.GenerateSuggestions(canceledCtx, status, 5, "")
	if err == nil {
		t.Error("expected error for canceled context")
	}
}

// TestApplySuggestionContextTimeout はコンテキストタイムアウトをテスト
func TestApplySuggestionContextTimeout(t *testing.T) {
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

	// タイムアウト済みのコンテキスト
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	_, err = z.ApplySuggestion(timeoutCtx, "test-id", false, false)
	if err == nil {
		t.Error("expected error for timed out context")
	}
}

// TestSuggestionValidate_AdditionalCases は追加の検証ケースをテスト
func TestSuggestionValidate_AdditionalCases(t *testing.T) {
	tests := []struct {
		name    string
		sugg    Suggestion
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing impact",
			sugg: Suggestion{
				ID:          "sugg-1",
				Type:        SuggestionRiskMitigation,
				Description: "Test description",
			},
			wantErr: true,
			errMsg:  "suggestion impact is required",
		},
		{
			name: "priority change without new priority",
			sugg: Suggestion{
				ID:           "sugg-1",
				Type:         SuggestionPriorityChange,
				Description:  "Priority change",
				Impact:       ImpactMedium,
				TargetTaskID: "task-1",
			},
			wantErr: true,
			errMsg:  "priority_change suggestion must have NewPriority",
		},
		{
			name: "dependency without target task ID",
			sugg: Suggestion{
				ID:           "sugg-1",
				Type:         SuggestionDependency,
				Description:  "Add dependency",
				Impact:       ImpactLow,
				Dependencies: []string{"task-1"},
			},
			wantErr: true,
			errMsg:  "dependency suggestion must have TargetTaskID",
		},
		{
			name: "dependency without dependencies",
			sugg: Suggestion{
				ID:           "sugg-1",
				Type:         SuggestionDependency,
				Description:  "Add dependency",
				Impact:       ImpactLow,
				TargetTaskID: "task-1",
			},
			wantErr: true,
			errMsg:  "dependency suggestion must have at least one dependency",
		},
		{
			name: "valid dependency suggestion",
			sugg: Suggestion{
				ID:           "sugg-1",
				Type:         SuggestionDependency,
				Description:  "Add dependency",
				Impact:       ImpactLow,
				TargetTaskID: "task-1",
				Dependencies: []string{"task-2"},
			},
			wantErr: false,
		},
		{
			name: "valid priority change",
			sugg: Suggestion{
				ID:           "sugg-1",
				Type:         SuggestionPriorityChange,
				Description:  "Priority change",
				Impact:       ImpactMedium,
				TargetTaskID: "task-1",
				NewPriority:  "high",
			},
			wantErr: false,
		},
		{
			name: "unknown suggestion type",
			sugg: Suggestion{
				ID:          "sugg-1",
				Type:        "unknown_type",
				Description: "Unknown type",
				Impact:      ImpactLow,
			},
			wantErr: true,
			errMsg:  "unknown suggestion type",
		},
		{
			name: "new task with invalid task data",
			sugg: Suggestion{
				ID:          "sugg-1",
				Type:        SuggestionNewTask,
				Description: "New task",
				Impact:      ImpactMedium,
				TaskData: &Task{
					ID: "task-1",
					// Title がない
					Status: TaskStatusPending,
				},
			},
			wantErr: true,
			errMsg:  "invalid task data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sugg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, want error containing %s", err, tt.errMsg)
			}
		})
	}
}

// TestApplySuggestion_NewTask は new_task タイプの提案適用をテスト
func TestApplySuggestion_NewTask(t *testing.T) {
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

	// 手動で new_task 提案を作成
	suggestion := &Suggestion{
		ID:          "sugg-new-task-1",
		Type:        SuggestionNewTask,
		Description: "Add new task for testing",
		Impact:      ImpactMedium,
		Status:      SuggestionPending,
		TaskData: &Task{
			ID:        "task-new-1",
			Title:     "New Task from Suggestion",
			Status:    TaskStatusPending,
			CreatedAt: Now(),
			UpdatedAt: Now(),
		},
	}

	// 提案をストアに保存
	store := &SuggestionStore{
		Suggestions: []Suggestion{*suggestion},
	}
	if err := z.fileStore.WriteYaml(ctx, "suggestions/active.yaml", store); err != nil {
		t.Fatalf("failed to write suggestion: %v", err)
	}

	// 提案を適用
	result, err := z.ApplySuggestion(ctx, suggestion.ID, false, false)
	if err != nil {
		t.Fatalf("ApplySuggestion failed: %v", err)
	}

	if result.Applied != 1 {
		t.Errorf("expected 1 applied, got %d", result.Applied)
	}

	// タスクが追加されたか確認
	var taskStore TaskStore
	if err := z.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		t.Fatalf("failed to read tasks: %v", err)
	}

	found := false
	for _, task := range taskStore.Tasks {
		if task.ID == "task-new-1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected new task to be added")
	}
}

// TestApplySuggestion_PriorityChange は priority_change タイプの提案適用をテスト
func TestApplySuggestion_PriorityChange(t *testing.T) {
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
	_, err = z.Add(ctx, "task", "Test Task")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// タスク ID を取得
	var taskStore TaskStore
	if err := z.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		t.Fatalf("failed to read tasks: %v", err)
	}
	if len(taskStore.Tasks) == 0 {
		t.Fatal("no tasks found")
	}
	taskID := taskStore.Tasks[0].ID

	// priority_change 提案を作成
	suggestion := &Suggestion{
		ID:           "sugg-priority-1",
		Type:         SuggestionPriorityChange,
		Description:  "Change priority to high",
		Impact:       ImpactMedium,
		Status:       SuggestionPending,
		TargetTaskID: taskID,
		NewPriority:  "high",
	}

	// 提案をストアに保存
	store := &SuggestionStore{
		Suggestions: []Suggestion{*suggestion},
	}
	if err := z.fileStore.WriteYaml(ctx, "suggestions/active.yaml", store); err != nil {
		t.Fatalf("failed to write suggestion: %v", err)
	}

	// 提案を適用
	result, err := z.ApplySuggestion(ctx, suggestion.ID, false, false)
	if err != nil {
		t.Fatalf("ApplySuggestion failed: %v", err)
	}

	if result.Applied != 1 {
		t.Errorf("expected 1 applied, got %d", result.Applied)
	}

	// タスクの優先度が変更されたか確認
	if err := z.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		t.Fatalf("failed to read tasks: %v", err)
	}

	for _, task := range taskStore.Tasks {
		if task.ID == taskID {
			if task.Priority != PriorityHigh {
				t.Errorf("expected priority high, got %s", task.Priority)
			}
			break
		}
	}
}

// TestApplySuggestion_Dependency は dependency タイプの提案適用をテスト
func TestApplySuggestion_Dependency(t *testing.T) {
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

	// 2つのタスクを追加
	_, err = z.Add(ctx, "task", "Test Task 1")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	_, err = z.Add(ctx, "task", "Test Task 2")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// タスク ID を取得
	var taskStore TaskStore
	if err := z.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		t.Fatalf("failed to read tasks: %v", err)
	}
	if len(taskStore.Tasks) < 2 {
		t.Fatal("not enough tasks")
	}
	task1ID := taskStore.Tasks[0].ID
	task2ID := taskStore.Tasks[1].ID

	// dependency 提案を作成
	suggestion := &Suggestion{
		ID:           "sugg-dep-1",
		Type:         SuggestionDependency,
		Description:  "Add dependency",
		Impact:       ImpactLow,
		Status:       SuggestionPending,
		TargetTaskID: task1ID,
		Dependencies: []string{task2ID},
	}

	// 提案をストアに保存
	store := &SuggestionStore{
		Suggestions: []Suggestion{*suggestion},
	}
	if err := z.fileStore.WriteYaml(ctx, "suggestions/active.yaml", store); err != nil {
		t.Fatalf("failed to write suggestion: %v", err)
	}

	// 提案を適用
	result, err := z.ApplySuggestion(ctx, suggestion.ID, false, false)
	if err != nil {
		t.Fatalf("ApplySuggestion failed: %v", err)
	}

	if result.Applied != 1 {
		t.Errorf("expected 1 applied, got %d", result.Applied)
	}

	// 依存関係が追加されたか確認
	if err := z.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		t.Fatalf("failed to read tasks: %v", err)
	}

	for _, task := range taskStore.Tasks {
		if task.ID == task1ID {
			found := false
			for _, dep := range task.Dependencies {
				if dep == task2ID {
					found = true
					break
				}
			}
			if !found {
				t.Error("expected dependency to be added")
			}
			break
		}
	}
}

// TestApplySuggestion_TargetTaskNotFound はターゲットタスクが見つからない場合をテスト
func TestApplySuggestion_TargetTaskNotFound(t *testing.T) {
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

	// priority_change 提案を作成（存在しないタスク）
	suggestion := &Suggestion{
		ID:           "sugg-priority-1",
		Type:         SuggestionPriorityChange,
		Description:  "Change priority to high",
		Impact:       ImpactMedium,
		Status:       SuggestionPending,
		TargetTaskID: "nonexistent-task",
		NewPriority:  "high",
	}

	// 提案をストアに保存
	store := &SuggestionStore{
		Suggestions: []Suggestion{*suggestion},
	}
	if err := z.fileStore.WriteYaml(ctx, "suggestions/active.yaml", store); err != nil {
		t.Fatalf("failed to write suggestion: %v", err)
	}

	// 提案を適用（--all で実行）
	result, err := z.ApplySuggestion(ctx, "", true, false)
	if err != nil {
		t.Fatalf("ApplySuggestion failed: %v", err)
	}

	// ターゲットが見つからないので失敗
	if result.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", result.Failed)
	}
}

// TestApplySuggestion_NewTaskWithoutTaskData は TaskData がない new_task をテスト
func TestApplySuggestion_NewTaskWithoutTaskData(t *testing.T) {
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

	// new_task 提案を作成（TaskData なし）
	suggestion := &Suggestion{
		ID:          "sugg-new-task-1",
		Type:        SuggestionNewTask,
		Description: "Add new task",
		Impact:      ImpactMedium,
		Status:      SuggestionPending,
		// TaskData がない
	}

	// 提案をストアに保存
	store := &SuggestionStore{
		Suggestions: []Suggestion{*suggestion},
	}
	if err := z.fileStore.WriteYaml(ctx, "suggestions/active.yaml", store); err != nil {
		t.Fatalf("failed to write suggestion: %v", err)
	}

	// 提案を適用（--all で実行）
	result, err := z.ApplySuggestion(ctx, "", true, false)
	if err != nil {
		t.Fatalf("ApplySuggestion failed: %v", err)
	}

	// TaskData がないので失敗
	if result.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", result.Failed)
	}
}
