package core

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/biwakonbu/zeus/internal/generator"
	"github.com/biwakonbu/zeus/internal/yaml"
	"github.com/google/uuid"
)

// Zeus はメインアプリケーション構造体
type Zeus struct {
	ProjectPath     string
	ZeusPath        string
	ClaudePath      string
	FileManager     *yaml.FileManager
	StateManager    *StateManager
	ApprovalManager *ApprovalManager
}

// New は新しい Zeus インスタンスを作成
func New(projectPath string) *Zeus {
	zeusPath := filepath.Join(projectPath, ".zeus")
	return &Zeus{
		ProjectPath:     projectPath,
		ZeusPath:        zeusPath,
		ClaudePath:      filepath.Join(projectPath, ".claude"),
		FileManager:     yaml.NewFileManager(zeusPath),
		StateManager:    NewStateManager(zeusPath),
		ApprovalManager: NewApprovalManager(zeusPath),
	}
}

// Init はプロジェクトを初期化
func (z *Zeus) Init(level string) (*InitResult, error) {
	// ディレクトリ構造を作成
	dirs := z.getDirectoryStructure(level)
	for _, dir := range dirs {
		if err := z.FileManager.EnsureDir(dir); err != nil {
			return nil, err
		}
	}

	// zeus.yaml を生成
	config := z.generateInitialConfig()
	if err := z.FileManager.WriteYaml("zeus.yaml", config); err != nil {
		return nil, err
	}

	// 初期タスクストアを作成
	taskStore := &TaskStore{Tasks: []Task{}}
	if err := z.FileManager.WriteYaml("tasks/active.yaml", taskStore); err != nil {
		return nil, err
	}
	if err := z.FileManager.WriteYaml("tasks/backlog.yaml", taskStore); err != nil {
		return nil, err
	}

	// 初期状態を記録
	state := z.calculateState(taskStore)
	if err := z.FileManager.WriteYaml("state/current.yaml", state); err != nil {
		return nil, err
	}

	// Claude Code 連携ファイルを生成（standard/advanced レベルの場合）
	if level == "standard" || level == "advanced" {
		gen := generator.NewGenerator(z.ProjectPath)
		if err := gen.GenerateAll(config.Project.Name, level); err != nil {
			// 生成に失敗しても初期化は続行
			fmt.Printf("Warning: Failed to generate Claude Code files: %v\n", err)
		}
	}

	return &InitResult{
		Success:    true,
		Level:      level,
		ZeusPath:   z.ZeusPath,
		ClaudePath: z.ClaudePath,
	}, nil
}

// Status はプロジェクトステータスを取得
func (z *Zeus) Status() (*StatusResult, error) {
	var config ZeusConfig
	if err := z.FileManager.ReadYaml("zeus.yaml", &config); err != nil {
		return nil, ErrConfigNotFound
	}

	state, err := z.getCurrentState()
	if err != nil {
		return nil, err
	}

	// 承認待ちアイテム数を取得
	pending, _ := z.ApprovalManager.GetPending()
	pendingCount := len(pending)

	return &StatusResult{
		Project:          config.Project,
		State:            *state,
		PendingApprovals: pendingCount,
	}, nil
}

// generateTaskID はユニークなタスク ID を生成
// UUID v4 を使用して衝突を防止
func (z *Zeus) generateTaskID() string {
	return fmt.Sprintf("task-%s", uuid.New().String()[:8])
}

// Add はエンティティを追加
func (z *Zeus) Add(entity, name string) (*AddResult, error) {
	if entity != "task" {
		return nil, ErrUnknownEntity
	}

	var taskStore TaskStore
	if err := z.FileManager.ReadYaml("tasks/active.yaml", &taskStore); err != nil {
		return nil, err
	}

	// UUID ベースの ID 生成（衝突防止）
	id := z.generateTaskID()
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

	taskStore.Tasks = append(taskStore.Tasks, task)
	if err := z.FileManager.WriteYaml("tasks/active.yaml", &taskStore); err != nil {
		return nil, err
	}

	// 状態を更新
	if err := z.updateState(); err != nil {
		return nil, err
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  entity,
	}, nil
}

// List はエンティティ一覧を取得
func (z *Zeus) List(entity string) (*ListResult, error) {
	if entity != "" && entity != "tasks" {
		return nil, ErrUnknownEntity
	}

	var taskStore TaskStore
	if err := z.FileManager.ReadYaml("tasks/active.yaml", &taskStore); err != nil {
		return nil, err
	}

	return &ListResult{
		Entity: "tasks",
		Items:  taskStore.Tasks,
		Total:  len(taskStore.Tasks),
	}, nil
}

// Pending は承認待ちアイテムを取得
func (z *Zeus) Pending() ([]PendingApproval, error) {
	return z.ApprovalManager.GetPending()
}

// Approve はアイテムを承認
func (z *Zeus) Approve(id string) (*ApprovalResult, error) {
	return z.ApprovalManager.Approve(id)
}

// Reject はアイテムを却下
func (z *Zeus) Reject(id, reason string) (*ApprovalResult, error) {
	return z.ApprovalManager.Reject(id, reason)
}

// CreateSnapshot はスナップショットを作成
func (z *Zeus) CreateSnapshot(label string) (*Snapshot, error) {
	return z.StateManager.CreateSnapshot(label)
}

// GetHistory は履歴を取得
func (z *Zeus) GetHistory(limit int) ([]Snapshot, error) {
	return z.StateManager.GetHistory(limit)
}

// RestoreSnapshot はスナップショットから復元
func (z *Zeus) RestoreSnapshot(timestamp string) error {
	return z.StateManager.RestoreSnapshot(timestamp)
}

// Private methods

func (z *Zeus) getDirectoryStructure(level string) []string {
	switch level {
	case "simple":
		return []string{"tasks", "state", "backups"}
	case "standard":
		return []string{
			"config", "tasks", "tasks/_archive", "state", "state/snapshots",
			"entities", "approvals/pending", "approvals/approved", "approvals/rejected",
			"logs", "analytics", "backups",
		}
	case "advanced":
		return []string{
			"config", "tasks", "tasks/_archive", "state", "state/snapshots",
			"entities", "approvals/pending", "approvals/approved", "approvals/rejected",
			"logs", "analytics", "graph", "views", ".local", "backups",
		}
	default:
		return []string{"tasks", "state", "backups"}
	}
}

func (z *Zeus) generateInitialConfig() *ZeusConfig {
	return &ZeusConfig{
		Version: "1.0",
		Project: ProjectInfo{
			ID:          fmt.Sprintf("zeus-%d", time.Now().Unix()),
			Name:        "New Zeus Project",
			Description: "Project managed by Zeus",
			StartDate:   Today(),
		},
		Objectives: []Objective{},
		Settings: Settings{
			AutomationLevel: "standard",
			ApprovalMode:    "default",
			AIProvider:      "claude-code",
		},
	}
}

func (z *Zeus) getCurrentState() (*ProjectState, error) {
	var state ProjectState
	if err := z.FileManager.ReadYaml("state/current.yaml", &state); err != nil {
		return z.getEmptyState(), nil
	}
	return &state, nil
}

func (z *Zeus) calculateState(taskStore *TaskStore) *ProjectState {
	stats := TaskStats{
		TotalTasks: len(taskStore.Tasks),
	}

	for _, task := range taskStore.Tasks {
		switch task.Status {
		case TaskStatusCompleted:
			stats.Completed++
		case TaskStatusInProgress:
			stats.InProgress++
		case TaskStatusPending:
			stats.Pending++
		}
	}

	return &ProjectState{
		Timestamp: Now(),
		Summary:   stats,
		Health:    z.calculateHealth(&stats),
		Risks:     []string{},
	}
}

func (z *Zeus) calculateHealth(stats *TaskStats) HealthStatus {
	if stats.TotalTasks == 0 {
		return HealthUnknown
	}
	progress := float64(stats.Completed) / float64(stats.TotalTasks)
	if progress < 0.3 {
		return HealthPoor
	}
	if progress < 0.7 {
		return HealthFair
	}
	return HealthGood
}

func (z *Zeus) getEmptyState() *ProjectState {
	return &ProjectState{
		Timestamp: Now(),
		Summary:   TaskStats{},
		Health:    HealthUnknown,
		Risks:     []string{},
	}
}

func (z *Zeus) updateState() error {
	var taskStore TaskStore
	if err := z.FileManager.ReadYaml("tasks/active.yaml", &taskStore); err != nil {
		return err
	}

	state := z.calculateState(&taskStore)
	return z.FileManager.WriteYaml("state/current.yaml", state)
}
