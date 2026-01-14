package core

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/biwakonbu/zeus/internal/generator"
	"github.com/biwakonbu/zeus/internal/yaml"
	"github.com/google/uuid"
)

// Zeus はメインアプリケーション構造体
type Zeus struct {
	ProjectPath string
	ZeusPath    string
	ClaudePath  string

	// インターフェースに依存（DI対応）
	fileStore      FileStore
	stateStore     StateStore
	approvalStore  ApprovalStore
	entityRegistry *EntityRegistry
}

// Option は Zeus の設定オプション
type Option func(*Zeus)

// WithFileStore は FileStore を設定
func WithFileStore(fs FileStore) Option {
	return func(z *Zeus) {
		z.fileStore = fs
	}
}

// WithStateStore は StateStore を設定
func WithStateStore(ss StateStore) Option {
	return func(z *Zeus) {
		z.stateStore = ss
	}
}

// WithApprovalStore は ApprovalStore を設定
func WithApprovalStore(as ApprovalStore) Option {
	return func(z *Zeus) {
		z.approvalStore = as
	}
}

// WithEntityRegistry は EntityRegistry を設定
func WithEntityRegistry(er *EntityRegistry) Option {
	return func(z *Zeus) {
		z.entityRegistry = er
	}
}

// New は新しい Zeus インスタンスを作成
func New(projectPath string, opts ...Option) *Zeus {
	zeusPath := filepath.Join(projectPath, ".zeus")

	z := &Zeus{
		ProjectPath: projectPath,
		ZeusPath:    zeusPath,
		ClaudePath:  filepath.Join(projectPath, ".claude"),
	}

	// オプション適用
	for _, opt := range opts {
		opt(z)
	}

	// デフォルト実装の設定
	if z.fileStore == nil {
		z.fileStore = yaml.NewFileManager(zeusPath)
	}
	if z.stateStore == nil {
		z.stateStore = NewStateManager(zeusPath, z.fileStore)
	}
	if z.approvalStore == nil {
		z.approvalStore = NewApprovalManager(zeusPath, z.fileStore)
	}
	if z.entityRegistry == nil {
		z.entityRegistry = NewEntityRegistry()
		z.entityRegistry.Register(NewTaskHandler(z.fileStore))
	}

	return z
}

// Init はプロジェクトを初期化
func (z *Zeus) Init(ctx context.Context, level string) (*InitResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ディレクトリ構造を作成
	dirs := z.getDirectoryStructure(level)
	for _, dir := range dirs {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if err := z.fileStore.EnsureDir(ctx, dir); err != nil {
				return nil, err
			}
		}
	}

	// zeus.yaml を生成
	config := z.generateInitialConfig()
	if err := z.fileStore.WriteYaml(ctx, "zeus.yaml", config); err != nil {
		return nil, err
	}

	// 初期タスクストアを作成
	taskStore := &TaskStore{Tasks: []Task{}}
	if err := z.fileStore.WriteYaml(ctx, "tasks/active.yaml", taskStore); err != nil {
		return nil, err
	}
	if err := z.fileStore.WriteYaml(ctx, "tasks/backlog.yaml", taskStore); err != nil {
		return nil, err
	}

	// 初期状態を記録
	state := z.stateStore.CalculateState(taskStore.Tasks)
	if err := z.stateStore.SaveCurrentState(ctx, state); err != nil {
		return nil, err
	}

	// Claude Code 連携ファイルを生成（standard/advanced レベルの場合）
	if level == "standard" || level == "advanced" {
		gen := generator.NewGenerator(z.ProjectPath)
		if err := gen.GenerateAll(ctx, config.Project.Name, level); err != nil {
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
func (z *Zeus) Status(ctx context.Context) (*StatusResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var config ZeusConfig
	if err := z.fileStore.ReadYaml(ctx, "zeus.yaml", &config); err != nil {
		return nil, ErrConfigNotFound
	}

	state, err := z.stateStore.GetCurrentState(ctx)
	if err != nil {
		return nil, err
	}

	// 承認待ちアイテム数を取得
	pending, _ := z.approvalStore.GetPending(ctx)
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
func (z *Zeus) Add(ctx context.Context, entity, name string) (*AddResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// EntityRegistry から適切なハンドラーを取得
	handler, ok := z.entityRegistry.Get(entity)
	if !ok {
		return nil, ErrUnknownEntity
	}

	result, err := handler.Add(ctx, name)
	if err != nil {
		return nil, err
	}

	// 状態を更新
	if err := z.updateState(ctx); err != nil {
		return nil, err
	}

	return result, nil
}

// List はエンティティ一覧を取得
func (z *Zeus) List(ctx context.Context, entity string) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// "tasks" を "task" に正規化
	normalizedEntity := entity
	if normalizedEntity == "" || normalizedEntity == "tasks" {
		normalizedEntity = "task"
	}

	// EntityRegistry から適切なハンドラーを取得
	handler, ok := z.entityRegistry.Get(normalizedEntity)
	if !ok {
		return nil, ErrUnknownEntity
	}

	return handler.List(ctx, nil)
}

// Pending は承認待ちアイテムを取得
func (z *Zeus) Pending(ctx context.Context) ([]PendingApproval, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return z.approvalStore.GetPending(ctx)
}

// Approve はアイテムを承認
func (z *Zeus) Approve(ctx context.Context, id string) (*ApprovalResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return z.approvalStore.Approve(ctx, id)
}

// Reject はアイテムを却下
func (z *Zeus) Reject(ctx context.Context, id, reason string) (*ApprovalResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return z.approvalStore.Reject(ctx, id, reason)
}

// CreateSnapshot はスナップショットを作成
func (z *Zeus) CreateSnapshot(ctx context.Context, label string) (*Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return z.stateStore.CreateSnapshot(ctx, label)
}

// GetHistory は履歴を取得
func (z *Zeus) GetHistory(ctx context.Context, limit int) ([]Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return z.stateStore.GetHistory(ctx, limit)
}

// RestoreSnapshot はスナップショットから復元
func (z *Zeus) RestoreSnapshot(ctx context.Context, timestamp string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return z.stateStore.RestoreSnapshot(ctx, timestamp)
}

// FileStore はFileStoreを返す（テスト用）
func (z *Zeus) FileStore() FileStore {
	return z.fileStore
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

func (z *Zeus) updateState(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var taskStore TaskStore
	if err := z.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		return err
	}

	state := z.stateStore.CalculateState(taskStore.Tasks)
	return z.stateStore.SaveCurrentState(ctx, state)
}
