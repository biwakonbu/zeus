package core

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
)

// ActivityHandler はアクティビティエンティティのハンドラー
type ActivityHandler struct {
	fileStore      FileStore
	usecaseHandler *UseCaseHandler
}

// NewActivityHandler は ActivityHandler を生成
func NewActivityHandler(fs FileStore, usecaseHandler *UseCaseHandler) *ActivityHandler {
	return &ActivityHandler{
		fileStore:      fs,
		usecaseHandler: usecaseHandler,
	}
}

// Type はエンティティタイプを返す
func (h *ActivityHandler) Type() string {
	return "activity"
}

// Add はアクティビティを追加
func (h *ActivityHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// activities ディレクトリを確保
	if err := h.fileStore.EnsureDir(ctx, "activities"); err != nil {
		return nil, fmt.Errorf("failed to ensure activities directory: %w", err)
	}

	// ID を生成
	id := h.generateActivityID()
	now := Now()

	activity := ActivityEntity{
		ID:     id,
		Title:  name,
		Status: ActivityStatusDraft,
		Metadata: Metadata{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// オプション適用
	for _, opt := range opts {
		opt(&activity)
	}

	// 参照整合性チェック: UseCaseID（任意紐付け）
	if activity.UseCaseID != "" && h.usecaseHandler != nil {
		if _, err := h.usecaseHandler.Get(ctx, activity.UseCaseID); err != nil {
			return nil, fmt.Errorf("referenced usecase not found: %s", activity.UseCaseID)
		}
	}

	// バリデーション
	if err := activity.Validate(); err != nil {
		return nil, err
	}

	// 個別ファイルに保存
	filePath := filepath.Join("activities", id+".yaml")
	if err := h.fileStore.WriteYaml(ctx, filePath, &activity); err != nil {
		return nil, fmt.Errorf("failed to write activity file: %w", err)
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  h.Type(),
	}, nil
}

// List はアクティビティ一覧を取得
func (h *ActivityHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// activities ディレクトリが存在しない場合は空リストを返す
	if !h.fileStore.Exists(ctx, "activities") {
		return &ListResult{
			Entity: h.Type(),
			Items:  []Task{},
			Total:  0,
		}, nil
	}

	// ディレクトリ内のファイルを列挙
	files, err := h.fileStore.ListDir(ctx, "activities")
	if err != nil {
		return nil, fmt.Errorf("failed to list activities directory: %w", err)
	}

	items := make([]Task, 0)
	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var activity ActivityEntity
		if err := h.fileStore.ReadYaml(ctx, filepath.Join("activities", file), &activity); err != nil {
			continue // 読み込み失敗はスキップ
		}
		// Task に変換（ListResult 互換性のため）
		items = append(items, Task{
			ID:        activity.ID,
			Title:     activity.Title,
			Status:    TaskStatus(activity.Status),
			CreatedAt: activity.Metadata.CreatedAt,
			UpdatedAt: activity.Metadata.UpdatedAt,
		})
	}

	return &ListResult{
		Entity: h.Type(),
		Items:  items,
		Total:  len(items),
	}, nil
}

// Get はアクティビティを取得
func (h *ActivityHandler) Get(ctx context.Context, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ID のセキュリティ検証
	if err := ValidateID("activity", id); err != nil {
		return nil, err
	}

	filePath := filepath.Join("activities", id+".yaml")
	if !h.fileStore.Exists(ctx, filePath) {
		return nil, ErrEntityNotFound
	}

	var activity ActivityEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &activity); err != nil {
		return nil, fmt.Errorf("failed to read activity file: %w", err)
	}

	return &activity, nil
}

// Update はアクティビティを更新
func (h *ActivityHandler) Update(ctx context.Context, id string, update any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID のセキュリティ検証
	if err := ValidateID("activity", id); err != nil {
		return err
	}

	filePath := filepath.Join("activities", id+".yaml")
	if !h.fileStore.Exists(ctx, filePath) {
		return ErrEntityNotFound
	}

	var activity ActivityEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &activity); err != nil {
		return fmt.Errorf("failed to read activity file: %w", err)
	}

	// 更新データを適用
	if updateMap, ok := update.(map[string]any); ok {
		if title, exists := updateMap["title"].(string); exists {
			activity.Title = title
		}
		if desc, exists := updateMap["description"].(string); exists {
			activity.Description = desc
		}
		if status, exists := updateMap["status"].(string); exists {
			activity.Status = ActivityStatus(status)
		}
		if usecaseID, exists := updateMap["usecase_id"].(string); exists {
			activity.UseCaseID = usecaseID
		}
	}
	activity.Metadata.UpdatedAt = Now()

	// バリデーション
	if err := activity.Validate(); err != nil {
		return err
	}

	return h.fileStore.WriteYaml(ctx, filePath, &activity)
}

// Delete はアクティビティを削除
func (h *ActivityHandler) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID のセキュリティ検証
	if err := ValidateID("activity", id); err != nil {
		return err
	}

	filePath := filepath.Join("activities", id+".yaml")
	if !h.fileStore.Exists(ctx, filePath) {
		return ErrEntityNotFound
	}

	return h.fileStore.Delete(ctx, filePath)
}

// GetAll は全アクティビティを取得（API用）
func (h *ActivityHandler) GetAll(ctx context.Context) ([]ActivityEntity, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// activities ディレクトリが存在しない場合は空リストを返す
	if !h.fileStore.Exists(ctx, "activities") {
		return []ActivityEntity{}, nil
	}

	// ディレクトリ内のファイルを列挙
	files, err := h.fileStore.ListDir(ctx, "activities")
	if err != nil {
		return nil, fmt.Errorf("failed to list activities directory: %w", err)
	}

	activities := make([]ActivityEntity, 0)
	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var activity ActivityEntity
		if err := h.fileStore.ReadYaml(ctx, filepath.Join("activities", file), &activity); err != nil {
			continue // 読み込み失敗はスキップ
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// AddNode はアクティビティにノードを追加
func (h *ActivityHandler) AddNode(ctx context.Context, activityID string, node ActivityNode) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID のセキュリティ検証
	if err := ValidateID("activity", activityID); err != nil {
		return err
	}

	filePath := filepath.Join("activities", activityID+".yaml")
	if !h.fileStore.Exists(ctx, filePath) {
		return ErrEntityNotFound
	}

	var activity ActivityEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &activity); err != nil {
		return fmt.Errorf("failed to read activity file: %w", err)
	}

	// ノードのバリデーション
	if err := node.Validate(); err != nil {
		return err
	}

	// 重複チェック
	for _, existing := range activity.Nodes {
		if existing.ID == node.ID {
			return fmt.Errorf("node already exists: %s", node.ID)
		}
	}

	// ノードを追加
	activity.Nodes = append(activity.Nodes, node)
	activity.Metadata.UpdatedAt = Now()

	return h.fileStore.WriteYaml(ctx, filePath, &activity)
}

// AddTransition はアクティビティに遷移を追加
func (h *ActivityHandler) AddTransition(ctx context.Context, activityID string, trans ActivityTransition) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID のセキュリティ検証
	if err := ValidateID("activity", activityID); err != nil {
		return err
	}

	filePath := filepath.Join("activities", activityID+".yaml")
	if !h.fileStore.Exists(ctx, filePath) {
		return ErrEntityNotFound
	}

	var activity ActivityEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &activity); err != nil {
		return fmt.Errorf("failed to read activity file: %w", err)
	}

	// 遷移のバリデーション
	if err := trans.Validate(); err != nil {
		return err
	}

	// ソース/ターゲットのノード存在確認
	nodeIDs := make(map[string]bool)
	for _, node := range activity.Nodes {
		nodeIDs[node.ID] = true
	}
	if !nodeIDs[trans.Source] {
		return fmt.Errorf("transition source not found: %s", trans.Source)
	}
	if !nodeIDs[trans.Target] {
		return fmt.Errorf("transition target not found: %s", trans.Target)
	}

	// 重複チェック
	for _, existing := range activity.Transitions {
		if existing.ID == trans.ID {
			return fmt.Errorf("transition already exists: %s", trans.ID)
		}
	}

	// 遷移を追加
	activity.Transitions = append(activity.Transitions, trans)
	activity.Metadata.UpdatedAt = Now()

	return h.fileStore.WriteYaml(ctx, filePath, &activity)
}

// generateActivityID はアクティビティ ID を生成
func (h *ActivityHandler) generateActivityID() string {
	return fmt.Sprintf("act-%s", uuid.New().String()[:8])
}

// ===== EntityOption 関数群 =====

// WithActivityUseCase は UseCase ID を設定
func WithActivityUseCase(usecaseID string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.UseCaseID = usecaseID
		}
	}
}

// WithActivityDescription は説明を設定
func WithActivityDescription(desc string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.Description = desc
		}
	}
}

// WithActivityStatus はステータスを設定
func WithActivityStatus(status ActivityStatus) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.Status = status
		}
	}
}

// WithActivityNodes はノードを設定
func WithActivityNodes(nodes []ActivityNode) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.Nodes = nodes
		}
	}
}

// WithActivityTransitions は遷移を設定
func WithActivityTransitions(transitions []ActivityTransition) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.Transitions = transitions
		}
	}
}

// WithActivityOwner はオーナーを設定
func WithActivityOwner(owner string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.Metadata.Owner = owner
		}
	}
}

// WithActivityTags はタグを設定
func WithActivityTags(tags []string) EntityOption {
	return func(v any) {
		if a, ok := v.(*ActivityEntity); ok {
			a.Metadata.Tags = tags
		}
	}
}
