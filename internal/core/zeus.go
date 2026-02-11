package core

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
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
	fileStore        FileStore
	stateStore       StateStore
	approvalStore    ApprovalStore
	entityRegistry   *EntityRegistry
	idCounterManager *IDCounterManager

	// UML ハンドラーへの直接アクセス（TASK-006）
	subsystemHandler *SubsystemHandler
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
	if z.idCounterManager == nil {
		z.idCounterManager = NewIDCounterManager(z.fileStore)
	}
	if z.entityRegistry == nil {
		z.entityRegistry = NewEntityRegistry()

		// 10 概念モデルのハンドラー登録（Phase 1）
		z.entityRegistry.Register(NewVisionHandler(z.fileStore))
		objHandler := NewObjectiveHandler(z.fileStore, z.idCounterManager)
		z.entityRegistry.Register(objHandler)

		// 10 概念モデルのハンドラー登録（Phase 2）
		conHandler := NewConsiderationHandler(z.fileStore, objHandler, z.idCounterManager)
		z.entityRegistry.Register(conHandler)
		decHandler := NewDecisionHandler(z.fileStore, conHandler, z.idCounterManager)
		z.entityRegistry.Register(decHandler)
		z.entityRegistry.Register(NewProblemHandler(z.fileStore, objHandler, z.idCounterManager))
		z.entityRegistry.Register(NewRiskHandler(z.fileStore, objHandler, z.idCounterManager))
		z.entityRegistry.Register(NewAssumptionHandler(z.fileStore, objHandler, z.idCounterManager))

		// 10 概念モデルのハンドラー登録（Phase 3）
		z.entityRegistry.Register(NewConstraintHandler(z.fileStore))
		z.entityRegistry.Register(NewQualityHandler(z.fileStore, objHandler, z.idCounterManager))

		// UML ユースケース図のハンドラー登録
		actorHandler := NewActorHandler(z.fileStore)
		z.entityRegistry.Register(actorHandler)

		// UML サブシステムのハンドラー登録（TASK-006）
		z.subsystemHandler = NewSubsystemHandler(z.fileStore)
		z.entityRegistry.Register(z.subsystemHandler)

		usecaseHandler := NewUseCaseHandler(z.fileStore, objHandler, actorHandler, z.idCounterManager)
		z.entityRegistry.Register(usecaseHandler)

		// UML アクティビティ図のハンドラー登録
		z.entityRegistry.Register(NewActivityHandler(z.fileStore, usecaseHandler))
	}

	return z
}

// Subsystems は SubsystemHandler を返す（TASK-006）
func (z *Zeus) Subsystems() *SubsystemHandler {
	return z.subsystemHandler
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

	// 初期状態を記録（空の状態）
	state := z.stateStore.CalculateState([]ListItem{})
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
func (z *Zeus) executeAdd(ctx context.Context, handler EntityHandler, _, name string, opts ...EntityOption) (*AddResult, error) {
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

	// エンティティ名を正規化
	normalizedEntity := entity
	switch normalizedEntity {
	case "", "activities":
		normalizedEntity = "activity"
	case "tasks":
		// 後方互換性: tasks も activity として扱う
		normalizedEntity = "activity"
	}

	// EntityRegistry から適切なハンドラーを取得
	handler, ok := z.entityRegistry.Get(normalizedEntity)
	if !ok {
		return nil, ErrUnknownEntity
	}

	return handler.List(ctx, nil)
}

// Get は指定されたエンティティを取得
func (z *Zeus) Get(ctx context.Context, entity, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// EntityRegistry から適切なハンドラーを取得
	handler, ok := z.entityRegistry.Get(entity)
	if !ok {
		return nil, ErrUnknownEntity
	}

	return handler.Get(ctx, id)
}

// GetRegistry は EntityRegistry を返す
func (z *Zeus) GetRegistry() *EntityRegistry {
	return z.entityRegistry
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
	// Note: tasks/ ディレクトリは非推奨。Activity を使用してください。
	return []string{
		"config", "state", "state/snapshots",
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

	// Activity から状態を計算（Task は非推奨、Activity に統合）
	actHandler := z.GetActivityHandler()
	if actHandler == nil {
		state := z.stateStore.CalculateState([]ListItem{})
		return z.stateStore.SaveCurrentState(ctx, state)
	}

	activities, err := actHandler.GetAll(ctx)
	if err != nil {
		state := z.stateStore.CalculateState([]ListItem{})
		return z.stateStore.SaveCurrentState(ctx, state)
	}

	// Activity を ListItem 形式に変換して状態計算
	tasks := make([]ListItem, len(activities))
	for i, act := range activities {
		tasks[i] = ListItem{
			ID:     act.ID,
			Title:  act.Title,
			Status: activityStatusToItemStatus(act.Status),
		}
	}

	state := z.stateStore.CalculateState(tasks)
	return z.stateStore.SaveCurrentState(ctx, state)
}

// activityStatusToItemStatus は ActivityStatus を ItemStatus に変換
func activityStatusToItemStatus(status ActivityStatus) ItemStatus {
	switch status {
	case ActivityStatusActive:
		return ItemStatusInProgress
	case ActivityStatusDeprecated:
		return ItemStatusCompleted
	case ActivityStatusDraft:
		return ItemStatusPending
	default:
		return ItemStatusPending
	}
}

// GenerateSuggestions はAI提案を生成
func (z *Zeus) GenerateSuggestions(ctx context.Context, status *StatusResult, limit int, impactFilter string) ([]Suggestion, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	suggestions := []Suggestion{}

	// Activity から統計を計算
	draftActivities := 0

	actHandler := z.GetActivityHandler()
	if actHandler != nil {
		activities, err := actHandler.GetAll(ctx)
		if err == nil {
			for _, act := range activities {
				if act.Status == ActivityStatusDraft {
					draftActivities++
				}
			}
		}
	}

	// Draft 状態のアクティビティが多い場合、優先順位付けを提案
	if draftActivities > 5 && (impactFilter == "" || impactFilter == "medium") {
		suggestions = append(suggestions, Suggestion{
			ID:          fmt.Sprintf("sugg-%s", uuid.New().String()[:8]),
			Type:        SuggestionPriorityChange,
			Description: "下書き状態のアクティビティが多いため、Active への移行を検討しましょう",
			Rationale:   fmt.Sprintf("%d件のアクティビティが Draft 状態です", draftActivities),
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

		// 作成された Activity ID を記録
		if suggestion.Type == SuggestionNewTask {
			if suggestion.ActivityData != nil {
				result.CreatedActivityID = suggestion.ActivityData.ID
				result.CreatedTaskID = suggestion.ActivityData.ID // 後方互換性
			}
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
func (z *Zeus) Explain(ctx context.Context, entityID string, includeContext bool) (*ExplainResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// "project" の場合はプロジェクト全体の説明
	if entityID == "project" {
		return z.explainProject(ctx, includeContext)
	}

	// Activity IDの場合
	if len(entityID) >= 4 && entityID[:4] == "act-" {
		return z.explainActivity(ctx, entityID, includeContext)
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
		state.Summary.TotalActivities,
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

// explainActivity は特定 Activity の説明を生成
func (z *Zeus) explainActivity(ctx context.Context, activityID string, includeContext bool) (*ExplainResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Activity を取得
	actHandler := z.GetActivityHandler()
	if actHandler == nil {
		return nil, fmt.Errorf("Activity handler が利用できません")
	}

	act, err := actHandler.Get(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("Activity が見つかりません: %s", activityID)
	}
	activity, ok := act.(*ActivityEntity)
	if !ok {
		return nil, fmt.Errorf("Activity の型が不正です: %s", activityID)
	}

	// 要約を生成
	summary := fmt.Sprintf("「%s」は現在 %s 状態の Activity です。",
		activity.Title, activity.Status)

	// 詳細を生成
	details := ""
	if activity.Description != "" {
		details = activity.Description
	}

	result := &ExplainResult{
		EntityID:    activityID,
		EntityType:  "activity",
		Summary:     summary,
		Details:     details,
		Context:     make(map[string]string),
		Suggestions: []string{},
	}

	// コンテキスト情報を追加
	if includeContext {
		result.Context["status"] = string(activity.Status)
		result.Context["created_at"] = activity.Metadata.CreatedAt
	}

	return result, nil
}

// applySuggestion は個別の提案を適用
// Task/Activity 統合により、Activity を使用
func (z *Zeus) applySuggestion(ctx context.Context, suggestion *Suggestion) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ActivityHandler を取得
	actHandler := z.GetActivityHandler()
	if actHandler == nil {
		return fmt.Errorf("Activity handler が利用できません")
	}

	switch suggestion.Type {
	case SuggestionNewTask:
		// ActivityData から Activity を追加
		if suggestion.ActivityData == nil {
			return fmt.Errorf("new_task タイプの提案に Activity データがありません")
		}

		activity := suggestion.ActivityData

		// Activity を追加
		result, err := actHandler.Add(ctx, activity.Title,
			WithActivityDescription(activity.Description),
			WithActivityStatus(activity.Status),
		)
		if err != nil {
			return fmt.Errorf("Activity の追加に失敗しました: %w", err)
		}

		// 生成された ID を suggestion に反映（レスポンス用）
		suggestion.ActivityData.ID = result.ID
		return nil

	case SuggestionPriorityChange:
		// 情報提供のみ（Activity に Priority フィールドは存在しない）
		fmt.Println("Info: Priority 変更は情報提供のみです（Activity に Priority フィールドはありません）")
		return nil

	case SuggestionDependency:
		// 情報提供のみ（Activity に Dependencies フィールドは存在しない）
		fmt.Println("Info: 依存関係の変更は情報提供のみです（Activity に Dependencies フィールドはありません）")
		return nil

	case SuggestionRiskMitigation:
		// リスク対策は情報提供のみ（Activity 変更なし）
		// ユーザーへの警告として機能し、適用済みとしてマークされる
		return nil

	default:
		return fmt.Errorf("不明な提案タイプ: %s", suggestion.Type)
	}
}

// ===== Phase 4: 高度な分析機能 =====

// activityToAnalysisTaskInfo は Activity を analysis.TaskInfo に変換（後方互換性用）
func activityToAnalysisTaskInfo(activities []ActivityEntity) []analysis.TaskInfo {
	result := make([]analysis.TaskInfo, len(activities))
	for i, a := range activities {
		result[i] = analysis.TaskInfo{
			ID:     a.ID,
			Title:  a.Title,
			Status: string(a.Status),
		}
	}
	return result
}

// toAnalysisActivityInfo は core.ActivityEntity を analysis.ActivityInfo に変換
func toAnalysisActivityInfo(activities []ActivityEntity) []analysis.ActivityInfo {
	result := make([]analysis.ActivityInfo, len(activities))
	for i, a := range activities {
		result[i] = analysis.ActivityInfo{
			ID:        a.ID,
			Title:     a.Title,
			Status:    string(a.Status),
			UseCaseID: a.UseCaseID,
			CreatedAt: a.Metadata.CreatedAt,
			UpdatedAt: a.Metadata.UpdatedAt,
		}
	}
	return result
}

// toAnalysisUseCaseInfo は core.UseCaseEntity を analysis.UseCaseInfo に変換
func toAnalysisUseCaseInfo(usecases []UseCaseEntity) []analysis.UseCaseInfo {
	result := make([]analysis.UseCaseInfo, len(usecases))
	for i, u := range usecases {
		actorIDs := make([]string, len(u.Actors))
		for j, a := range u.Actors {
			actorIDs[j] = a.ActorID
		}
		result[i] = analysis.UseCaseInfo{
			ID:          u.ID,
			Title:       u.Title,
			Status:      string(u.Status),
			ObjectiveID: u.ObjectiveID,
			SubsystemID: u.SubsystemID,
			ActorIDs:    actorIDs,
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
		Summary: report.SummaryStats{
			TotalActivities: state.Summary.TotalActivities,
			Completed:       state.Summary.Completed,
			InProgress:      state.Summary.InProgress,
			Pending:         state.Summary.Pending,
		},
	}
}

// BuildDependencyGraph は依存関係グラフを構築
func (z *Zeus) BuildDependencyGraph(ctx context.Context) (*analysis.DependencyGraph, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Activity 一覧を取得
	actHandler := z.GetActivityHandler()
	if actHandler == nil {
		return &analysis.DependencyGraph{
			Nodes:    make(map[string]*analysis.GraphNode),
			Edges:    []analysis.Edge{},
			Cycles:   [][]string{},
			Isolated: []string{},
			Stats:    analysis.GraphStats{},
		}, nil
	}

	activities, err := actHandler.GetAll(ctx)
	if err != nil {
		return &analysis.DependencyGraph{
			Nodes:    make(map[string]*analysis.GraphNode),
			Edges:    []analysis.Edge{},
			Cycles:   [][]string{},
			Isolated: []string{},
			Stats:    analysis.GraphStats{},
		}, nil
	}

	if len(activities) == 0 {
		return &analysis.DependencyGraph{
			Nodes:    make(map[string]*analysis.GraphNode),
			Edges:    []analysis.Edge{},
			Cycles:   [][]string{},
			Isolated: []string{},
			Stats:    analysis.GraphStats{},
		}, nil
	}

	// Activity を analysis.TaskInfo に変換
	taskInfos := activityToAnalysisTaskInfo(activities)

	// グラフを構築
	builder := analysis.NewGraphBuilder(taskInfos)
	return builder.Build(ctx)
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

	// 分析結果を作成（グラフのみ）
	var analysisResult *analysis.AnalysisResult

	// グラフを取得（Markdown形式用）
	if format == "markdown" {
		graph, _ := z.BuildDependencyGraph(ctx)
		analysisResult = &analysis.AnalysisResult{
			Graph: graph,
		}
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

// hasYamlSuffix は .yaml または .yml 拡張子を持つかチェック
func hasYamlSuffix(filename string) bool {
	return len(filename) > 5 && (filename[len(filename)-5:] == ".yaml" || filename[len(filename)-4:] == ".yml")
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

// ===== Task/Activity 統合: UnifiedGraph 機能 =====

// BuildUnifiedGraph は統合グラフを構築
// Activity, UseCase, Objective を統合した依存関係グラフを返す
func (z *Zeus) BuildUnifiedGraph(ctx context.Context, filter *analysis.GraphFilter) (*analysis.UnifiedGraph, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// 1. 全 Activity を取得
	activities := []ActivityEntity{}
	actFiles, err := z.fileStore.ListDir(ctx, "activities")
	if err == nil {
		for _, file := range actFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			var act ActivityEntity
			if err := z.fileStore.ReadYaml(ctx, filepath.Join("activities", file), &act); err == nil {
				activities = append(activities, act)
			}
		}
	}

	// 2. 全 UseCase を取得
	usecases := []UseCaseEntity{}
	ucFiles, err := z.fileStore.ListDir(ctx, "usecases")
	if err == nil {
		for _, file := range ucFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			var uc UseCaseEntity
			if err := z.fileStore.ReadYaml(ctx, filepath.Join("usecases", file), &uc); err == nil {
				usecases = append(usecases, uc)
			}
		}
	}

	// 3. 全 Objective を取得
	objectives := []analysis.ObjectiveInfo{}
	objFiles, err := z.fileStore.ListDir(ctx, "objectives")
	if err == nil {
		for _, file := range objFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			var obj ObjectiveEntity
			if err := z.fileStore.ReadYaml(ctx, filepath.Join("objectives", file), &obj); err == nil {
				objectives = append(objectives, analysis.ObjectiveInfo{
					ID:          obj.ID,
					Title:       obj.Title,
					Description: obj.Description,
					Goals:       obj.Goals,
					Status:      string(obj.Status),
				})
			}
		}
	}

	// 4. UnifiedGraphBuilder で構築
	builder := analysis.NewUnifiedGraphBuilder().
		WithActivities(toAnalysisActivityInfo(activities)).
		WithUseCases(toAnalysisUseCaseInfo(usecases)).
		WithObjectives(objectives)

	if filter != nil {
		builder = builder.WithFilter(filter)
	}

	graph := builder.Build()
	if errs := builder.ValidationErrors(); len(errs) > 0 {
		return nil, fmt.Errorf("unified graph relation validation failed: %s", strings.Join(errs, "; "))
	}

	return graph, nil
}

// GetActivityHandler は ActivityHandler を返す
func (z *Zeus) GetActivityHandler() *ActivityHandler {
	if handler, ok := z.entityRegistry.Get("activity"); ok {
		if activityHandler, ok := handler.(*ActivityHandler); ok {
			return activityHandler
		}
	}
	return nil
}
