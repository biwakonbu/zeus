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
		_, err := z.Add(ctx, "activity", "Test Activity")
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
		_, err := z.Add(ctx, "activity", "Test Activity")
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
		z.Add(ctx, "activity", "Test Activity")
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
		z.Add(ctx, "activity", "Test Activity")
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
		task    ListItem
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid task",
			task: ListItem{
				ID:        "task-1",
				Title:     "Test Task",
				Status:    ItemStatusPending,
				CreatedAt: Now(),
				UpdatedAt: Now(),
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			task: ListItem{
				Title:  "Test Task",
				Status: ItemStatusPending,
			},
			wantErr: true,
			errMsg:  "item ID is required",
		},
		{
			name: "missing title",
			task: ListItem{
				ID:     "task-1",
				Status: ItemStatusPending,
			},
			wantErr: true,
			errMsg:  "item title is required",
		},
		{
			name: "missing status",
			task: ListItem{
				ID:    "task-1",
				Title: "Test Task",
			},
			wantErr: true,
			errMsg:  "item status is required",
		},
		{
			name: "invalid approval level",
			task: ListItem{
				ID:            "task-1",
				Title:         "Test Task",
				Status:        ItemStatusPending,
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
	validTask := &ListItem{
		ID:        "task-1",
		Title:     "Test Task",
		Status:    ItemStatusPending,
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
			errMsg:  "new_task suggestion must have ActivityData or TaskData",
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
				TaskData: &ListItem{
					ID: "task-1",
					// Title がない
					Status: ItemStatusPending,
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

	// ActivityData を使用して new_task 提案を作成
	now := Now()
	suggestion := &Suggestion{
		ID:          "sugg-new-task-1",
		Type:        SuggestionNewTask,
		Description: "Add new activity for testing",
		Impact:      ImpactMedium,
		Status:      SuggestionPending,
		ActivityData: &ActivityEntity{
			ID:     "act-a1b2c3d4", // UUID 形式（8桁16進数）
			Title:  "New Activity from Suggestion",
			Status: ActivityStatusDraft,
			Metadata: Metadata{
				CreatedAt: now,
				UpdatedAt: now,
			},
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

	// Activity が追加されたか確認
	actHandler := z.GetActivityHandler()
	if actHandler == nil {
		t.Fatal("activity handler is nil")
	}

	listResult, err := actHandler.List(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list activities: %v", err)
	}

	found := false
	for _, item := range listResult.Items {
		if item.Title == "New Activity from Suggestion" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected new activity to be added")
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

	// Activity を追加
	result, err := z.Add(ctx, "activity", "Test Activity")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	activityID := result.ID

	// priority_change 提案を作成
	suggestion := &Suggestion{
		ID:           "sugg-priority-1",
		Type:         SuggestionPriorityChange,
		Description:  "Change priority to high",
		Impact:       ImpactMedium,
		Status:       SuggestionPending,
		TargetTaskID: activityID,
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
	applyResult, err := z.ApplySuggestion(ctx, suggestion.ID, false, false)
	if err != nil {
		t.Fatalf("ApplySuggestion failed: %v", err)
	}

	if applyResult.Applied != 1 {
		t.Errorf("expected 1 applied, got %d", applyResult.Applied)
	}

	// Activity の優先度変更は ActivityEntity.Priority フィールドが削除されたため確認不要
	// ApplySuggestion が正常に完了したことで十分
	_ = activityID
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

	// 2つの Activity を追加
	result1, err := z.Add(ctx, "activity", "Test Activity 1")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	result2, err := z.Add(ctx, "activity", "Test Activity 2")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	act1ID := result1.ID
	act2ID := result2.ID

	// dependency 提案を作成
	suggestion := &Suggestion{
		ID:           "sugg-dep-1",
		Type:         SuggestionDependency,
		Description:  "Add dependency",
		Impact:       ImpactLow,
		Status:       SuggestionPending,
		TargetTaskID: act1ID,
		Dependencies: []string{act2ID},
	}

	// 提案をストアに保存
	store := &SuggestionStore{
		Suggestions: []Suggestion{*suggestion},
	}
	if err := z.fileStore.WriteYaml(ctx, "suggestions/active.yaml", store); err != nil {
		t.Fatalf("failed to write suggestion: %v", err)
	}

	// 提案を適用
	applyResult, err := z.ApplySuggestion(ctx, suggestion.ID, false, false)
	if err != nil {
		t.Fatalf("ApplySuggestion failed: %v", err)
	}

	if applyResult.Applied != 1 {
		t.Errorf("expected 1 applied, got %d", applyResult.Applied)
	}

	// ActivityEntity.Dependencies フィールドが削除されたため、依存関係の直接確認は不要
	// ApplySuggestion が正常に完了したことで十分
	_ = act1ID
	_ = act2ID
}

// TestApplySuggestion_PriorityChangeInformational は priority_change が情報提供のみで成功することをテスト
func TestApplySuggestion_PriorityChangeInformational(t *testing.T) {
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

	// priority_change 提案を作成（存在しない Activity）
	suggestion := &Suggestion{
		ID:           "sugg-priority-1",
		Type:         SuggestionPriorityChange,
		Description:  "Change priority to high",
		Impact:       ImpactMedium,
		Status:       SuggestionPending,
		TargetTaskID: "act-nonexistent",
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

	// priority_change は情報提供のみ（Activity に Priority フィールドは存在しない）のため成功扱い
	if result.Applied != 1 {
		t.Errorf("expected 1 applied (informational), got %d", result.Applied)
	}
}

// TestApplySuggestion_NewTaskWithoutActivityData は ActivityData がない new_task をテスト
func TestApplySuggestion_NewTaskWithoutActivityData(t *testing.T) {
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

	// new_task 提案を作成（ActivityData なし、TaskData もなし）
	suggestion := &Suggestion{
		ID:          "sugg-new-task-1",
		Type:        SuggestionNewTask,
		Description: "Add new activity",
		Impact:      ImpactMedium,
		Status:      SuggestionPending,
		// ActivityData も TaskData もなし
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

	// ActivityData がないので失敗
	if result.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", result.Failed)
	}
}
