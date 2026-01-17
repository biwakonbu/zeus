package core

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// TaskHandler はタスクエンティティのハンドラー
type TaskHandler struct {
	fileStore FileStore
}

// NewTaskHandler は新しい TaskHandler を作成
func NewTaskHandler(fs FileStore) *TaskHandler {
	return &TaskHandler{fileStore: fs}
}

// Type はエンティティタイプを返す
func (h *TaskHandler) Type() string {
	return "task"
}

// Add はタスクを追加
func (h *TaskHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var taskStore TaskStore
	if err := h.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		return nil, err
	}

	id := h.generateTaskID()
	now := Now()

	task := Task{
		ID:            id,
		Title:         name,
		Status:        TaskStatusPending,
		Dependencies:  []string{},
		ApprovalLevel: ApprovalAuto,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// オプション適用
	for _, opt := range opts {
		opt(&task)
	}

	// ParentID が設定されている場合、Dependencies に追加
	if task.ParentID != "" {
		task.Dependencies = append(task.Dependencies, task.ParentID)
	}

	taskStore.Tasks = append(taskStore.Tasks, task)
	if err := h.fileStore.WriteYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		return nil, err
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  h.Type(),
	}, nil
}

// List はタスク一覧を取得
func (h *TaskHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var taskStore TaskStore
	if err := h.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		return nil, err
	}

	tasks := taskStore.Tasks

	// フィルタ適用
	if filter != nil && filter.Status != "" {
		filtered := []Task{}
		for _, t := range tasks {
			if string(t.Status) == filter.Status {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}

	// Limit適用
	if filter != nil && filter.Limit > 0 && len(tasks) > filter.Limit {
		tasks = tasks[:filter.Limit]
	}

	return &ListResult{
		Entity: h.Type() + "s",
		Items:  tasks,
		Total:  len(tasks),
	}, nil
}

// Get はタスクを取得
func (h *TaskHandler) Get(ctx context.Context, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var taskStore TaskStore
	if err := h.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		return nil, err
	}

	for _, task := range taskStore.Tasks {
		if task.ID == id {
			return &task, nil
		}
	}

	return nil, ErrEntityNotFound
}

// Update はタスクを更新
func (h *TaskHandler) Update(ctx context.Context, id string, update any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var taskStore TaskStore
	if err := h.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		return err
	}

	found := false
	for i, task := range taskStore.Tasks {
		if task.ID == id {
			if u, ok := update.(*Task); ok {
				u.UpdatedAt = Now()
				taskStore.Tasks[i] = *u
			}
			found = true
			break
		}
	}

	if !found {
		return ErrEntityNotFound
	}

	return h.fileStore.WriteYaml(ctx, "tasks/active.yaml", &taskStore)
}

// Delete はタスクを削除
func (h *TaskHandler) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var taskStore TaskStore
	if err := h.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		return err
	}

	remaining := []Task{}
	found := false
	for _, task := range taskStore.Tasks {
		if task.ID == id {
			found = true
			continue
		}
		remaining = append(remaining, task)
	}

	if !found {
		return ErrEntityNotFound
	}

	taskStore.Tasks = remaining
	return h.fileStore.WriteYaml(ctx, "tasks/active.yaml", &taskStore)
}

// generateTaskID はユニークなタスク ID を生成
func (h *TaskHandler) generateTaskID() string {
	return fmt.Sprintf("task-%s", uuid.New().String()[:8])
}

// TaskOption はタスク作成オプション

// WithTaskDescription はタスクの説明を設定
func WithTaskDescription(desc string) EntityOption {
	return func(v any) {
		if t, ok := v.(*Task); ok {
			t.Description = desc
		}
	}
}

// WithTaskStatus はタスクのステータスを設定
func WithTaskStatus(status TaskStatus) EntityOption {
	return func(v any) {
		if t, ok := v.(*Task); ok {
			t.Status = status
		}
	}
}

// WithTaskAssignee はタスクの担当者を設定
func WithTaskAssignee(assignee string) EntityOption {
	return func(v any) {
		if t, ok := v.(*Task); ok {
			t.Assignee = assignee
		}
	}
}

// WithTaskEstimateHours はタスクの見積もり時間を設定
func WithTaskEstimateHours(hours float64) EntityOption {
	return func(v any) {
		if t, ok := v.(*Task); ok {
			t.EstimateHours = hours
		}
	}
}

// WithTaskDependencies はタスクの依存関係を設定
func WithTaskDependencies(deps []string) EntityOption {
	return func(v any) {
		if t, ok := v.(*Task); ok {
			t.Dependencies = deps
		}
	}
}

// WithTaskApprovalLevel はタスクの承認レベルを設定
func WithTaskApprovalLevel(level ApprovalLevel) EntityOption {
	return func(v any) {
		if t, ok := v.(*Task); ok {
			t.ApprovalLevel = level
		}
	}
}

// WithTaskParent はタスクの親タスクIDを設定
func WithTaskParent(parentID string) EntityOption {
	return func(v any) {
		if t, ok := v.(*Task); ok {
			t.ParentID = parentID
		}
	}
}

// WithTaskStartDate はタスクの開始日を設定
func WithTaskStartDate(startDate string) EntityOption {
	return func(v any) {
		if t, ok := v.(*Task); ok {
			t.StartDate = startDate
		}
	}
}

// WithTaskDueDate はタスクの期限日を設定
func WithTaskDueDate(dueDate string) EntityOption {
	return func(v any) {
		if t, ok := v.(*Task); ok {
			t.DueDate = dueDate
		}
	}
}

// WithTaskProgress はタスクの進捗率を設定
func WithTaskProgress(progress int) EntityOption {
	return func(v any) {
		if t, ok := v.(*Task); ok {
			if progress >= 0 && progress <= 100 {
				t.Progress = progress
			}
		}
	}
}

// WithTaskWBSCode はタスクのWBSコードを設定
func WithTaskWBSCode(wbsCode string) EntityOption {
	return func(v any) {
		if t, ok := v.(*Task); ok {
			t.WBSCode = wbsCode
		}
	}
}

// WithTaskPriority はタスクの優先度を設定
func WithTaskPriority(priority TaskPriority) EntityOption {
	return func(v any) {
		if t, ok := v.(*Task); ok {
			t.Priority = priority
		}
	}
}
