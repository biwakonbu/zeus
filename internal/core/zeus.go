package core

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/biwakonbu/zeus/internal/analysis"
	"github.com/biwakonbu/zeus/internal/generator"
	"github.com/biwakonbu/zeus/internal/report"
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
func (z *Zeus) Init(ctx context.Context) (*InitResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ディレクトリ構造を作成
	dirs := z.getDirectoryStructure()
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

	// Claude Code 連携ファイルを常に生成
	gen := generator.NewGenerator(z.ProjectPath)
	if err := gen.GenerateAll(ctx, config.Project.Name); err != nil {
		fmt.Printf("Warning: Claude Code ファイル生成に失敗: %v\n", err)
	}

	return &InitResult{
		Success:    true,
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
// automation_level に応じて承認フローと連携
func (z *Zeus) Add(ctx context.Context, entity, name string, opts ...EntityOption) (*AddResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// EntityRegistry から適切なハンドラーを取得
	handler, ok := z.entityRegistry.Get(entity)
	if !ok {
		return nil, ErrUnknownEntity
	}

	// 設定を読み込んで承認レベルを判定
	var config ZeusConfig
	if err := z.fileStore.ReadYaml(ctx, "zeus.yaml", &config); err != nil {
		// 設定読み込み失敗時は auto として扱う
		config.Settings.ApprovalMode = "loose"
		config.Settings.AutomationLevel = "auto"
	}

	// auto レベルは常に即時実行
	if config.Settings.AutomationLevel == "auto" {
		return z.executeAdd(ctx, handler, entity, name, opts...)
	}

	// notify/approve: 承認レベルを判定
	approvalLevel := z.approvalStore.(*ApprovalManager).DetermineApprovalLevel("task_create", &config.Settings)

	switch approvalLevel {
	case ApprovalAuto:
		// 自動承認: 即時実行
		return z.executeAdd(ctx, handler, entity, name, opts...)

	case ApprovalNotify:
		// 通知付き実行: 実行してログに記録
		result, err := z.executeAdd(ctx, handler, entity, name, opts...)
		if err != nil {
			return nil, err
		}
		// TODO: 通知ログの記録（将来の機能拡張）
		return result, nil

	case ApprovalApprove:
		// 明示的承認が必要: 承認待ちキューに追加
		approval, err := z.approvalStore.(*ApprovalManager).Create(
			ctx,
			"task_create",
			fmt.Sprintf("%s '%s' の追加", entity, name),
			approvalLevel,
			"", // entityID は承認後に決定
			map[string]string{"entity": entity, "name": name},
		)
		if err != nil {
			return nil, fmt.Errorf("承認待ちキューへの追加に失敗しました: %w", err)
		}

		return &AddResult{
			Success:       true,
			ID:            "", // 承認後に決定
			Entity:        entity,
			NeedsApproval: true,
			ApprovalID:    approval.ID,
		}, nil

	default:
		// デフォルトは即時実行
		return z.executeAdd(ctx, handler, entity, name, opts...)
	}
}

// executeAdd は実際のエンティティ追加を実行
func (z *Zeus) executeAdd(ctx context.Context, handler EntityHandler, entity, name string, opts ...EntityOption) (*AddResult, error) {
	result, err := handler.Add(ctx, name, opts...)
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

func (z *Zeus) getDirectoryStructure() []string {
	// 全機能が使用可能な統一構造
	return []string{
		"config", "tasks", "tasks/_archive", "state", "state/snapshots",
		"entities", "approvals/pending", "approvals/approved", "approvals/rejected",
		"logs", "analytics", "backups",
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
			AutomationLevel: "auto",
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

// GenerateSuggestions はAI提案を生成
// 注意: 現在はルールベースの簡易実装。Phase 3 で Claude Code AI 統合予定。
func (z *Zeus) GenerateSuggestions(ctx context.Context, status *StatusResult, limit int, impactFilter string) ([]Suggestion, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	suggestions := []Suggestion{}

	// プロジェクト状態を分析（status から統計を取得可能な場合は再利用）
	pendingTasks := status.State.Summary.Pending
	blockedTasks := 0

	// ブロックされたタスク数はstateに含まれていないため、別途取得
	var taskStore TaskStore
	if err := z.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err == nil {
		for _, task := range taskStore.Tasks {
			if task.Status == TaskStatusBlocked {
				blockedTasks++
			}
		}
	}

	// ブロックされたタスクがある場合、リスク対策を提案
	if blockedTasks > 0 && (impactFilter == "" || impactFilter == "high") {
		suggestions = append(suggestions, Suggestion{
			ID:          fmt.Sprintf("sugg-%s", uuid.New().String()[:8]),
			Type:        SuggestionRiskMitigation,
			Description: fmt.Sprintf("%d件のブロックされたタスクを解決する必要があります", blockedTasks),
			Rationale:   "ブロックされたタスクはプロジェクト全体の進行を妨げます",
			Impact:      ImpactHigh,
			Status:      SuggestionPending,
			CreatedAt:   Now(),
		})
	}

	// 保留中のタスクが多い場合、優先順位付けを提案
	if pendingTasks > 5 && (impactFilter == "" || impactFilter == "medium") {
		suggestions = append(suggestions, Suggestion{
			ID:          fmt.Sprintf("sugg-%s", uuid.New().String()[:8]),
			Type:        SuggestionPriorityChange,
			Description: "保留中のタスクが多いため、優先順位を明確にしましょう",
			Rationale:   fmt.Sprintf("%d件のタスクが保留中です", pendingTasks),
			Impact:      ImpactMedium,
			Status:      SuggestionPending,
			CreatedAt:   Now(),
		})
	}

	// limit を適用
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	// 提案を保存
	if len(suggestions) > 0 {
		if err := z.saveSuggestions(ctx, suggestions); err != nil {
			return nil, fmt.Errorf("提案の保存に失敗しました: %w", err)
		}
	}

	return suggestions, nil
}

// ApplySuggestion は提案を適用
// 部分的な成功をサポート: 一部の提案が失敗しても、成功した分は適用される
func (z *Zeus) ApplySuggestion(ctx context.Context, suggestionID string, applyAll bool, dryRun bool) (*ApplyResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &ApplyResult{
		AppliedIDs: []string{},
		FailedIDs:  []string{},
	}

	// 提案を読み込み
	var store SuggestionStore
	if err := z.fileStore.ReadYaml(ctx, "suggestions/active.yaml", &store); err != nil {
		if !z.fileStore.Exists(ctx, "suggestions/active.yaml") {
			return nil, fmt.Errorf("提案がまだ生成されていません。zeus suggest を実行してください")
		}
		return nil, fmt.Errorf("提案の読み込み失敗: %w", err)
	}

	// 引数検証: --all と提案ID の両方が指定された場合
	if applyAll && suggestionID != "" {
		return nil, fmt.Errorf("--all フラグと提案IDを同時に指定することはできません")
	}

	// 適用対象を特定
	toApply := []int{} // ストア内のインデックスを保持
	if applyAll {
		for i, s := range store.Suggestions {
			if s.Status == SuggestionPending {
				toApply = append(toApply, i)
			}
		}
	} else {
		found := false
		for i, s := range store.Suggestions {
			if s.ID == suggestionID {
				toApply = append(toApply, i)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("提案 %s が見つかりません", suggestionID)
		}
	}

	// Dry-run モード: 検証のみ
	if dryRun {
		for _, idx := range toApply {
			suggestion := &store.Suggestions[idx]
			if suggestion.Status != SuggestionPending {
				result.Skipped++
				continue
			}
			result.Applied++
			result.AppliedIDs = append(result.AppliedIDs, suggestion.ID)
		}
		return result, nil
	}

	// 適用実行: 部分的な成功をサポート
	for _, idx := range toApply {
		suggestion := &store.Suggestions[idx]
		if suggestion.Status != SuggestionPending {
			result.Skipped++
			continue
		}

		// 提案の検証
		if err := suggestion.Validate(); err != nil {
			result.Failed++
			result.FailedIDs = append(result.FailedIDs, suggestion.ID)
			continue
		}

		// 適用実行
		if err := z.applySuggestion(ctx, suggestion); err != nil {
			// エラーをログに記録するが、他の提案の適用は続行
			result.Failed++
			result.FailedIDs = append(result.FailedIDs, suggestion.ID)
			continue
		}

		// 成功: ステータスを更新
		store.Suggestions[idx].Status = SuggestionApplied
		store.Suggestions[idx].UpdatedAt = Now()

		if suggestion.Type == SuggestionNewTask && suggestion.TaskData != nil {
			result.CreatedTaskID = suggestion.TaskData.ID
		}

		result.Applied++
		result.AppliedIDs = append(result.AppliedIDs, suggestion.ID)
	}

	// 成功した提案がある場合、ストアを保存
	if result.Applied > 0 {
		if err := z.fileStore.WriteYaml(ctx, "suggestions/active.yaml", &store); err != nil {
			return result, fmt.Errorf("提案は適用されましたが、ストア保存に失敗しました: %w", err)
		}
	}

	return result, nil
}

// saveSuggestions は提案を保存
// ディレクトリ作成を先に行い、既存提案を適切に読み込む
func (z *Zeus) saveSuggestions(ctx context.Context, suggestions []Suggestion) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ディレクトリを先に確保
	if err := z.fileStore.EnsureDir(ctx, "suggestions"); err != nil {
		return fmt.Errorf("提案ディレクトリの作成に失敗しました: %w", err)
	}

	// 既存の提案を読み込む（ファイルが存在する場合のみ）
	var store SuggestionStore
	if z.fileStore.Exists(ctx, "suggestions/active.yaml") {
		if err := z.fileStore.ReadYaml(ctx, "suggestions/active.yaml", &store); err != nil {
			return fmt.Errorf("既存の提案の読み込みに失敗しました: %w", err)
		}
	}

	// 新しい提案を追加
	store.Suggestions = append(store.Suggestions, suggestions...)

	// 保存
	if err := z.fileStore.WriteYaml(ctx, "suggestions/active.yaml", &store); err != nil {
		return fmt.Errorf("提案の保存に失敗しました: %w", err)
	}

	return nil
}

// Explain はエンティティの詳細説明を生成
// 注意: 現在はルールベースの簡易実装。Phase 3 で AI ベースに拡張予定。
func (z *Zeus) Explain(ctx context.Context, entityID string, includeContext bool) (*ExplainResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// "project" の場合はプロジェクト全体の説明
	if entityID == "project" {
		return z.explainProject(ctx, includeContext)
	}

	// タスクIDの場合
	if len(entityID) >= 5 && entityID[:5] == "task-" {
		return z.explainTask(ctx, entityID, includeContext)
	}

	return nil, fmt.Errorf("不明なエンティティ: %s", entityID)
}

// explainProject はプロジェクト全体の説明を生成
func (z *Zeus) explainProject(ctx context.Context, includeContext bool) (*ExplainResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// 設定を読み込み
	var config ZeusConfig
	if err := z.fileStore.ReadYaml(ctx, "zeus.yaml", &config); err != nil {
		return nil, ErrConfigNotFound
	}

	// 状態を取得
	state, err := z.stateStore.GetCurrentState(ctx)
	if err != nil {
		return nil, err
	}

	// 要約を生成
	summary := fmt.Sprintf("%s は %s に開始されたプロジェクトです。",
		config.Project.Name, config.Project.StartDate)

	// 詳細を生成
	details := fmt.Sprintf("現在の健全性: %s\nタスク: 全 %d 件 (完了: %d, 進行中: %d, 保留: %d)",
		state.Health,
		state.Summary.TotalTasks,
		state.Summary.Completed,
		state.Summary.InProgress,
		state.Summary.Pending)

	result := &ExplainResult{
		EntityID:    "project",
		EntityType:  "project",
		Summary:     summary,
		Details:     details,
		Context:     make(map[string]string),
		Suggestions: []string{},
	}

	// コンテキスト情報を追加
	if includeContext {
		result.Context["project_id"] = config.Project.ID
		result.Context["automation_level"] = config.Settings.AutomationLevel
		result.Context["ai_provider"] = config.Settings.AIProvider
	}

	// 改善提案を生成
	if state.Summary.Pending > 5 {
		result.Suggestions = append(result.Suggestions,
			"保留中のタスクが多いです。優先順位を見直すことをお勧めします。")
	}
	if state.Health == HealthPoor {
		result.Suggestions = append(result.Suggestions,
			"プロジェクトの健全性が低下しています。リスク要因を確認してください。")
	}

	return result, nil
}

// explainTask は特定タスクの説明を生成
func (z *Zeus) explainTask(ctx context.Context, taskID string, includeContext bool) (*ExplainResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// タスクストアを読み込み
	var taskStore TaskStore
	if err := z.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		return nil, err
	}

	// タスクを検索
	var targetTask *Task
	for i := range taskStore.Tasks {
		if taskStore.Tasks[i].ID == taskID {
			targetTask = &taskStore.Tasks[i]
			break
		}
	}

	if targetTask == nil {
		return nil, fmt.Errorf("タスクが見つかりません: %s", taskID)
	}

	// 要約を生成
	summary := fmt.Sprintf("「%s」は現在 %s 状態のタスクです。",
		targetTask.Title, targetTask.Status)

	// 詳細を生成
	details := ""
	if targetTask.Description != "" {
		details = targetTask.Description
	}
	if targetTask.EstimateHours > 0 {
		details += fmt.Sprintf("\n見積もり工数: %.1f 時間", targetTask.EstimateHours)
	}
	if targetTask.ActualHours > 0 {
		details += fmt.Sprintf("\n実績工数: %.1f 時間", targetTask.ActualHours)
	}

	result := &ExplainResult{
		EntityID:    taskID,
		EntityType:  "task",
		Summary:     summary,
		Details:     details,
		Context:     make(map[string]string),
		Suggestions: []string{},
	}

	// コンテキスト情報を追加
	if includeContext {
		result.Context["status"] = string(targetTask.Status)
		result.Context["approval_level"] = string(targetTask.ApprovalLevel)
		result.Context["created_at"] = targetTask.CreatedAt
		if targetTask.Assignee != "" {
			result.Context["assignee"] = targetTask.Assignee
		}
		if len(targetTask.Dependencies) > 0 {
			result.Context["dependencies"] = fmt.Sprintf("%v", targetTask.Dependencies)
		}
	}

	// 改善提案を生成
	if targetTask.Status == TaskStatusBlocked {
		result.Suggestions = append(result.Suggestions,
			"このタスクはブロックされています。依存関係を確認してください。")
	}
	if targetTask.EstimateHours > 0 && targetTask.ActualHours > targetTask.EstimateHours*1.5 {
		result.Suggestions = append(result.Suggestions,
			"実績が見積もりを大幅に超過しています。タスク分割を検討してください。")
	}

	return result, nil
}

// applySuggestion は個別の提案を適用
func (z *Zeus) applySuggestion(ctx context.Context, suggestion *Suggestion) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	switch suggestion.Type {
	case SuggestionNewTask:
		if suggestion.TaskData == nil {
			return fmt.Errorf("new_task タイプの提案にタスクデータがありません")
		}
		// タスクデータを検証
		if err := suggestion.TaskData.Validate(); err != nil {
			return fmt.Errorf("タスクデータが無効です: %w", err)
		}
		// タスクを追加
		var taskStore TaskStore
		if err := z.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
			return fmt.Errorf("タスクストアの読み込みに失敗しました: %w", err)
		}
		taskStore.Tasks = append(taskStore.Tasks, *suggestion.TaskData)
		if err := z.fileStore.WriteYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
			return fmt.Errorf("タスクストアの保存に失敗しました: %w", err)
		}
		return nil

	case SuggestionPriorityChange:
		if suggestion.TargetTaskID == "" {
			return fmt.Errorf("priority_change タイプにターゲットタスクIDがありません")
		}
		if suggestion.NewPriority == "" {
			return fmt.Errorf("priority_change タイプに新しい優先度がありません")
		}
		// タスクストアを読み込み
		var taskStore TaskStore
		if err := z.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
			return fmt.Errorf("タスクストアの読み込みに失敗しました: %w", err)
		}
		// 対象タスクを検索して更新
		found := false
		for i := range taskStore.Tasks {
			if taskStore.Tasks[i].ID == suggestion.TargetTaskID {
				taskStore.Tasks[i].Priority = TaskPriority(suggestion.NewPriority)
				taskStore.Tasks[i].UpdatedAt = Now()
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("タスクが見つかりません: %s", suggestion.TargetTaskID)
		}
		if err := z.fileStore.WriteYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
			return fmt.Errorf("タスクストアの保存に失敗しました: %w", err)
		}
		return nil

	case SuggestionDependency:
		if suggestion.TargetTaskID == "" {
			return fmt.Errorf("dependency タイプにターゲットタスクIDがありません")
		}
		if len(suggestion.Dependencies) == 0 {
			return fmt.Errorf("dependency タイプに依存関係がありません")
		}
		// タスクストアを読み込み
		var taskStore TaskStore
		if err := z.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
			return fmt.Errorf("タスクストアの読み込みに失敗しました: %w", err)
		}
		// 対象タスクを検索して更新
		found := false
		for i := range taskStore.Tasks {
			if taskStore.Tasks[i].ID == suggestion.TargetTaskID {
				// 既存の依存関係に追加（重複を避ける）
				existingDeps := make(map[string]bool)
				for _, dep := range taskStore.Tasks[i].Dependencies {
					existingDeps[dep] = true
				}
				for _, newDep := range suggestion.Dependencies {
					if !existingDeps[newDep] {
						taskStore.Tasks[i].Dependencies = append(taskStore.Tasks[i].Dependencies, newDep)
					}
				}
				taskStore.Tasks[i].UpdatedAt = Now()
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("タスクが見つかりません: %s", suggestion.TargetTaskID)
		}
		if err := z.fileStore.WriteYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
			return fmt.Errorf("タスクストアの保存に失敗しました: %w", err)
		}
		return nil

	case SuggestionRiskMitigation:
		// リスク対策は情報提供のみ（タスク変更なし）
		// ユーザーへの警告として機能し、適用済みとしてマークされる
		return nil

	default:
		return fmt.Errorf("不明な提案タイプ: %s", suggestion.Type)
	}
}

// ===== Phase 4: 高度な分析機能 =====

// toAnalysisTaskInfo は core.Task を analysis.TaskInfo に変換
func toAnalysisTaskInfo(tasks []Task) []analysis.TaskInfo {
	result := make([]analysis.TaskInfo, len(tasks))
	for i, t := range tasks {
		result[i] = analysis.TaskInfo{
			ID:            t.ID,
			Title:         t.Title,
			Status:        string(t.Status),
			Dependencies:  t.Dependencies,
			ParentID:      t.ParentID,
			StartDate:     t.StartDate,
			DueDate:       t.DueDate,
			Progress:      t.Progress,
			WBSCode:       t.WBSCode,
			Priority:      string(t.Priority),
			Assignee:      t.Assignee,
			EstimateHours: t.EstimateHours,
		}
	}
	return result
}

// toAnalysisProjectState は core.ProjectState を analysis.ProjectState に変換
func toAnalysisProjectState(state *ProjectState) *analysis.ProjectState {
	return &analysis.ProjectState{
		Health: string(state.Health),
		Summary: analysis.TaskStats{
			TotalTasks: state.Summary.TotalTasks,
			Completed:  state.Summary.Completed,
			InProgress: state.Summary.InProgress,
			Pending:    state.Summary.Pending,
		},
	}
}

// toAnalysisSnapshots は core.Snapshot を analysis.Snapshot に変換
func toAnalysisSnapshots(snapshots []Snapshot) []analysis.Snapshot {
	result := make([]analysis.Snapshot, len(snapshots))
	for i, s := range snapshots {
		result[i] = analysis.Snapshot{
			Timestamp: s.Timestamp,
			State: analysis.ProjectState{
				Health: string(s.State.Health),
				Summary: analysis.TaskStats{
					TotalTasks: s.State.Summary.TotalTasks,
					Completed:  s.State.Summary.Completed,
					InProgress: s.State.Summary.InProgress,
					Pending:    s.State.Summary.Pending,
				},
			},
		}
	}
	return result
}

// toReportConfig は core.ZeusConfig を report.ZeusConfig に変換
func toReportConfig(config *ZeusConfig) *report.ZeusConfig {
	return &report.ZeusConfig{
		Project: report.ProjectInfo{
			ID:          config.Project.ID,
			Name:        config.Project.Name,
			Description: config.Project.Description,
			StartDate:   config.Project.StartDate,
		},
	}
}

// toReportProjectState は core.ProjectState を report.ProjectState に変換
func toReportProjectState(state *ProjectState) *report.ProjectState {
	return &report.ProjectState{
		Health: string(state.Health),
		Summary: report.TaskStats{
			TotalTasks: state.Summary.TotalTasks,
			Completed:  state.Summary.Completed,
			InProgress: state.Summary.InProgress,
			Pending:    state.Summary.Pending,
		},
	}
}

// BuildDependencyGraph は依存関係グラフを構築
func (z *Zeus) BuildDependencyGraph(ctx context.Context) (*analysis.DependencyGraph, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// タスク一覧を取得
	var taskStore TaskStore
	if err := z.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		return nil, fmt.Errorf("タスクの読み込みに失敗しました: %w", err)
	}

	if len(taskStore.Tasks) == 0 {
		return &analysis.DependencyGraph{
			Nodes:    make(map[string]*analysis.GraphNode),
			Edges:    []analysis.Edge{},
			Cycles:   [][]string{},
			Isolated: []string{},
			Stats:    analysis.GraphStats{},
		}, nil
	}

	// core.Task を analysis.TaskInfo に変換
	taskInfos := toAnalysisTaskInfo(taskStore.Tasks)

	// グラフを構築
	builder := analysis.NewGraphBuilder(taskInfos)
	return builder.Build(ctx)
}

// Predict は予測分析を実行
func (z *Zeus) Predict(ctx context.Context, predType string) (*analysis.AnalysisResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// 現在の状態を取得
	state, err := z.stateStore.GetCurrentState(ctx)
	if err != nil {
		return nil, fmt.Errorf("状態の取得に失敗しました: %w", err)
	}

	// 履歴を取得
	history, err := z.stateStore.GetHistory(ctx, 30) // 最大30件
	if err != nil {
		history = []Snapshot{} // エラー時は空のスライス
	}

	// タスク一覧を取得
	var taskStore TaskStore
	if err := z.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		taskStore = TaskStore{Tasks: []Task{}}
	}

	// 型変換
	analysisState := toAnalysisProjectState(state)
	analysisHistory := toAnalysisSnapshots(history)
	analysisTaskInfos := toAnalysisTaskInfo(taskStore.Tasks)

	// Predictor を作成
	predictor := analysis.NewPredictor(analysisState, analysisHistory, analysisTaskInfos)

	result := &analysis.AnalysisResult{}

	// 予測タイプに応じて分析を実行
	switch predType {
	case "completion":
		completion, err := predictor.PredictCompletion(ctx)
		if err != nil {
			return nil, err
		}
		result.Completion = completion

	case "risk":
		risk, err := predictor.PredictRisk(ctx)
		if err != nil {
			return nil, err
		}
		result.Risk = risk

	case "velocity":
		velocity, err := predictor.CalculateVelocity(ctx)
		if err != nil {
			return nil, err
		}
		result.Velocity = velocity

	case "all", "":
		// 全ての予測を実行
		completion, _ := predictor.PredictCompletion(ctx)
		result.Completion = completion

		risk, _ := predictor.PredictRisk(ctx)
		result.Risk = risk

		velocity, _ := predictor.CalculateVelocity(ctx)
		result.Velocity = velocity

	default:
		return nil, fmt.Errorf("不明な予測タイプ: %s", predType)
	}

	return result, nil
}

// GenerateReport はレポートを生成
func (z *Zeus) GenerateReport(ctx context.Context, format string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	// 設定を取得
	var config ZeusConfig
	if err := z.fileStore.ReadYaml(ctx, "zeus.yaml", &config); err != nil {
		return "", ErrConfigNotFound
	}

	// 状態を取得
	state, err := z.stateStore.GetCurrentState(ctx)
	if err != nil {
		return "", fmt.Errorf("状態の取得に失敗しました: %w", err)
	}

	// 分析結果を取得
	analysisResult, _ := z.Predict(ctx, "all")

	// グラフを取得（Markdown形式用）
	if format == "markdown" {
		graph, _ := z.BuildDependencyGraph(ctx)
		if analysisResult == nil {
			analysisResult = &analysis.AnalysisResult{}
		}
		analysisResult.Graph = graph
	}

	// 型変換
	reportConfig := toReportConfig(&config)
	reportState := toReportProjectState(state)

	// レポートを生成
	gen := report.NewGenerator(reportConfig, reportState, analysisResult)

	switch format {
	case "text", "":
		return gen.GenerateText(ctx)
	case "html":
		return gen.GenerateHTML(ctx)
	case "markdown":
		return gen.GenerateMarkdown(ctx)
	default:
		return "", fmt.Errorf("不明なレポート形式: %s", format)
	}
}

// ===== Phase 6B: WBS 機能 =====

// BuildWBSTree はWBS階層ツリーを構築
func (z *Zeus) BuildWBSTree(ctx context.Context) (*analysis.WBSTree, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// タスク一覧を取得
	var taskStore TaskStore
	if err := z.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		return nil, fmt.Errorf("タスクの読み込みに失敗しました: %w", err)
	}

	if len(taskStore.Tasks) == 0 {
		return &analysis.WBSTree{
			Roots:    []*analysis.WBSNode{},
			MaxDepth: 0,
			Stats:    analysis.WBSStats{},
		}, nil
	}

	// core.Task を analysis.TaskInfo に変換
	taskInfos := toAnalysisTaskInfo(taskStore.Tasks)

	// WBSツリーを構築
	builder := analysis.NewWBSBuilder(taskInfos)
	return builder.Build(ctx)
}

// ===== Phase 6C: タイムライン機能 =====

// BuildTimeline はタイムラインを構築
func (z *Zeus) BuildTimeline(ctx context.Context) (*analysis.Timeline, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// タスク一覧を取得
	var taskStore TaskStore
	if err := z.fileStore.ReadYaml(ctx, "tasks/active.yaml", &taskStore); err != nil {
		return nil, fmt.Errorf("タスクの読み込みに失敗しました: %w", err)
	}

	if len(taskStore.Tasks) == 0 {
		return &analysis.Timeline{
			Items:        []analysis.TimelineItem{},
			CriticalPath: []string{},
			Stats:        analysis.TimelineStats{},
		}, nil
	}

	// core.Task を analysis.TaskInfo に変換
	taskInfos := toAnalysisTaskInfo(taskStore.Tasks)

	// タイムラインを構築
	builder := analysis.NewTimelineBuilder(taskInfos)
	return builder.Build(ctx)
}

// ===== Claude Code 連携ファイル更新 =====

// UpdateClaudeFiles は Claude Code 連携ファイルを最新テンプレートで再生成
func (z *Zeus) UpdateClaudeFiles(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// 設定からプロジェクト名を取得
	var config ZeusConfig
	if err := z.fileStore.ReadYaml(ctx, "zeus.yaml", &config); err != nil {
		return fmt.Errorf("zeus.yaml の読み込みに失敗: %w", err)
	}

	// Claude Code 連携ファイルを生成
	gen := generator.NewGenerator(z.ProjectPath)
	if err := gen.GenerateAll(ctx, config.Project.Name); err != nil {
		return fmt.Errorf("Claude Code ファイル生成に失敗: %w", err)
	}

	return nil
}
