package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// ObjectiveHandler は ObjectiveEntity エンティティのハンドラー
// 個別ファイル (objectives/obj-{uuid}.yaml) で管理
type ObjectiveHandler struct {
	fileStore FileStore
	sanitizer *Sanitizer
}

// NewObjectiveHandler は新しい ObjectiveHandler を作成
func NewObjectiveHandler(fs FileStore, _ *IDCounterManager) *ObjectiveHandler {
	return &ObjectiveHandler{
		fileStore: fs,
		sanitizer: NewSanitizer(),
	}
}

// Type はエンティティタイプを返す
func (h *ObjectiveHandler) Type() string {
	return "objective"
}

// Add は Objective を追加
func (h *ObjectiveHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// サニタイズ
	sanitizedName, err := h.sanitizer.SanitizeString("title", name)
	if err != nil {
		return nil, fmt.Errorf("invalid title: %w", err)
	}

	// UUID 形式の ID を生成
	id := h.generateID()

	now := Now()
	objective := &ObjectiveEntity{
		ID:     id,
		Title:  sanitizedName,
		Status: ObjectiveStatusNotStarted,
		Metadata: Metadata{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// オプション適用
	for _, opt := range opts {
		opt(objective)
	}

	// バリデーション
	if err := objective.Validate(); err != nil {
		return nil, err
	}

	// ファイル書き込み
	filePath := filepath.Join("objectives", id+".yaml")
	if err := h.fileStore.WriteYaml(ctx, filePath, objective); err != nil {
		return nil, err
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  h.Type(),
	}, nil
}

// List は Objective 一覧を取得
func (h *ObjectiveHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	objectives, err := h.getAllObjectives(ctx)
	if err != nil {
		return nil, err
	}

	// フィルタ適用
	if filter != nil && filter.Status != "" {
		filtered := []*ObjectiveEntity{}
		for _, obj := range objectives {
			if string(obj.Status) == filter.Status {
				filtered = append(filtered, obj)
			}
		}
		objectives = filtered
	}

	// Limit 適用
	if filter != nil && filter.Limit > 0 && len(objectives) > filter.Limit {
		objectives = objectives[:filter.Limit]
	}

	// ListResult.Items は []ListItem なので、空を返す（互換性維持）
	// 本来は汎用 Entity リスト対応が望ましい
	return &ListResult{
		Entity: h.Type() + "s",
		Items:  []ListItem{},
		Total:  len(objectives),
	}, nil
}

// Get は Objective を取得
func (h *ObjectiveHandler) Get(ctx context.Context, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ID バリデーション
	if err := ValidateID("objective", id); err != nil {
		return nil, err
	}

	filePath := filepath.Join("objectives", id+".yaml")
	var objective ObjectiveEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &objective); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrEntityNotFound
		}
		return nil, err
	}

	return &objective, nil
}

// Update は Objective を更新
func (h *ObjectiveHandler) Update(ctx context.Context, id string, update any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("objective", id); err != nil {
		return err
	}

	// 既存を取得
	existing, err := h.Get(ctx, id)
	if err != nil {
		return err
	}
	existingObj := existing.(*ObjectiveEntity)

	// 更新適用
	if obj, ok := update.(*ObjectiveEntity); ok {
		obj.ID = id // ID は変更不可
		obj.Metadata.CreatedAt = existingObj.Metadata.CreatedAt
		obj.Metadata.UpdatedAt = Now()

		// バリデーション
		if err := obj.Validate(); err != nil {
			return err
		}

		filePath := filepath.Join("objectives", id+".yaml")
		return h.fileStore.WriteYaml(ctx, filePath, obj)
	}

	return fmt.Errorf("invalid update type: expected *ObjectiveEntity")
}

// Delete は Objective を削除
func (h *ObjectiveHandler) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("objective", id); err != nil {
		return err
	}

	// 存在確認
	if _, err := h.Get(ctx, id); err != nil {
		return err
	}

	// ファイル削除
	filePath := filepath.Join("objectives", id+".yaml")
	return h.fileStore.Delete(ctx, filePath)
}

// generateID は UUID 形式の Objective ID を生成
func (h *ObjectiveHandler) generateID() string {
	return fmt.Sprintf("obj-%s", uuid.New().String()[:8])
}

// getAllObjectives は全 Objective を取得
func (h *ObjectiveHandler) getAllObjectives(ctx context.Context) ([]*ObjectiveEntity, error) {
	// objectives ディレクトリ内のファイル一覧を取得
	files, err := h.fileStore.ListDir(ctx, "objectives")
	if err != nil {
		if os.IsNotExist(err) {
			return []*ObjectiveEntity{}, nil
		}
		return nil, err
	}

	var objectives []*ObjectiveEntity
	for _, file := range files {
		if !strings.HasSuffix(file, ".yaml") {
			continue
		}

		// ID を抽出
		id := strings.TrimSuffix(filepath.Base(file), ".yaml")
		if err := ValidateID("objective", id); err != nil {
			// 無効な ID は無視（ログは記録しない）
			continue
		}

		// フルパスを構築
		filePath := filepath.Join("objectives", file)
		var obj ObjectiveEntity
		if err := h.fileStore.ReadYaml(ctx, filePath, &obj); err != nil {
			// パーミッション不足以外のエラーは報告
			if !os.IsPermission(err) {
				return nil, fmt.Errorf("failed to read objective file %s: %w", filePath, err)
			}
			// パーミッションエラーのみ無視（スキップしてログは記録しない）
			continue
		}
		objectives = append(objectives, &obj)
	}

	// ID でソート
	sort.Slice(objectives, func(i, j int) bool {
		return objectives[i].ID < objectives[j].ID
	})

	return objectives, nil
}

// Objective オプション関数

// WithObjectiveDescription は Objective の説明を設定
func WithObjectiveDescription(desc string) EntityOption {
	return func(v any) {
		if obj, ok := v.(*ObjectiveEntity); ok {
			obj.Description = desc
		}
	}
}

// WithObjectiveStatus は Objective のステータスを設定
func WithObjectiveStatus(status ObjectiveStatus) EntityOption {
	return func(v any) {
		if obj, ok := v.(*ObjectiveEntity); ok {
			obj.Status = status
		}
	}
}

// WithObjectiveOwner は Objective のオーナーを設定
func WithObjectiveOwner(owner string) EntityOption {
	return func(v any) {
		if obj, ok := v.(*ObjectiveEntity); ok {
			obj.Owner = owner
		}
	}
}

// WithObjectiveTags は Objective のタグを設定
func WithObjectiveTags(tags []string) EntityOption {
	return func(v any) {
		if obj, ok := v.(*ObjectiveEntity); ok {
			obj.Tags = tags
		}
	}
}
