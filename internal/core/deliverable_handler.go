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

// DeliverableHandler は DeliverableEntity エンティティのハンドラー
// 個別ファイル (deliverables/del-{uuid}.yaml) で管理
type DeliverableHandler struct {
	fileStore        FileStore
	sanitizer        *Sanitizer
	objectiveHandler *ObjectiveHandler
}

// NewDeliverableHandler は新しい DeliverableHandler を作成
func NewDeliverableHandler(fs FileStore, objHandler *ObjectiveHandler, _ *IDCounterManager) *DeliverableHandler {
	return &DeliverableHandler{
		fileStore:        fs,
		sanitizer:        NewSanitizer(),
		objectiveHandler: objHandler,
	}
}

// Type はエンティティタイプを返す
func (h *DeliverableHandler) Type() string {
	return "deliverable"
}

// Add は Deliverable を追加
func (h *DeliverableHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
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
	deliverable := &DeliverableEntity{
		ID:     id,
		Title:  sanitizedName,
		Status: DeliverableStatusPlanned,
		Format: DeliverableFormatOther,
		Metadata: Metadata{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// オプション適用
	for _, opt := range opts {
		opt(deliverable)
	}

	// ObjectiveID が設定されている場合、存在確認
	if deliverable.ObjectiveID != "" {
		if err := h.validateObjectiveReference(ctx, deliverable.ObjectiveID); err != nil {
			return nil, err
		}
	}

	// バリデーション
	if err := deliverable.Validate(); err != nil {
		return nil, err
	}

	// ファイル書き込み
	filePath := filepath.Join("deliverables", id+".yaml")
	if err := h.fileStore.WriteYaml(ctx, filePath, deliverable); err != nil {
		return nil, err
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  h.Type(),
	}, nil
}

// List は Deliverable 一覧を取得
func (h *DeliverableHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	deliverables, err := h.getAllDeliverables(ctx)
	if err != nil {
		return nil, err
	}

	// フィルタ適用
	if filter != nil && filter.Status != "" {
		filtered := []*DeliverableEntity{}
		for _, del := range deliverables {
			if string(del.Status) == filter.Status {
				filtered = append(filtered, del)
			}
		}
		deliverables = filtered
	}

	// Limit 適用
	if filter != nil && filter.Limit > 0 && len(deliverables) > filter.Limit {
		deliverables = deliverables[:filter.Limit]
	}

	return &ListResult{
		Entity: h.Type() + "s",
		Items:  []ListItem{},
		Total:  len(deliverables),
	}, nil
}

// Get は Deliverable を取得
func (h *DeliverableHandler) Get(ctx context.Context, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ID バリデーション
	if err := ValidateID("deliverable", id); err != nil {
		return nil, err
	}

	filePath := filepath.Join("deliverables", id+".yaml")
	var deliverable DeliverableEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &deliverable); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrEntityNotFound
		}
		return nil, err
	}

	return &deliverable, nil
}

// Update は Deliverable を更新
func (h *DeliverableHandler) Update(ctx context.Context, id string, update any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("deliverable", id); err != nil {
		return err
	}

	// 既存を取得
	existing, err := h.Get(ctx, id)
	if err != nil {
		return err
	}
	existingDel := existing.(*DeliverableEntity)

	// 更新適用
	if del, ok := update.(*DeliverableEntity); ok {
		del.ID = id // ID は変更不可
		del.Metadata.CreatedAt = existingDel.Metadata.CreatedAt
		del.Metadata.UpdatedAt = Now()

		// ObjectiveID が変更された場合、存在確認
		if del.ObjectiveID != "" && del.ObjectiveID != existingDel.ObjectiveID {
			if err := h.validateObjectiveReference(ctx, del.ObjectiveID); err != nil {
				return err
			}
		}

		// バリデーション
		if err := del.Validate(); err != nil {
			return err
		}

		filePath := filepath.Join("deliverables", id+".yaml")
		return h.fileStore.WriteYaml(ctx, filePath, del)
	}

	return fmt.Errorf("invalid update type: expected *DeliverableEntity")
}

// Delete は Deliverable を削除
func (h *DeliverableHandler) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("deliverable", id); err != nil {
		return err
	}

	// 存在確認
	if _, err := h.Get(ctx, id); err != nil {
		return err
	}

	// ファイル削除
	filePath := filepath.Join("deliverables", id+".yaml")
	return h.fileStore.Delete(ctx, filePath)
}

// generateID は UUID 形式の Deliverable ID を生成
func (h *DeliverableHandler) generateID() string {
	return fmt.Sprintf("del-%s", uuid.New().String()[:8])
}

// getAllDeliverables は全 Deliverable を取得
func (h *DeliverableHandler) getAllDeliverables(ctx context.Context) ([]*DeliverableEntity, error) {
	// deliverables ディレクトリ内のファイル一覧を取得
	files, err := h.fileStore.ListDir(ctx, "deliverables")
	if err != nil {
		if os.IsNotExist(err) {
			return []*DeliverableEntity{}, nil
		}
		return nil, err
	}

	var deliverables []*DeliverableEntity
	for _, file := range files {
		if !strings.HasSuffix(file, ".yaml") {
			continue
		}

		// ID を抽出
		id := strings.TrimSuffix(filepath.Base(file), ".yaml")
		if err := ValidateID("deliverable", id); err != nil {
			// 無効な ID は無視（ログは記録しない）
			continue
		}

		// フルパスを構築
		filePath := filepath.Join("deliverables", file)
		var del DeliverableEntity
		if err := h.fileStore.ReadYaml(ctx, filePath, &del); err != nil {
			// パーミッション不足以外のエラーは報告
			if !os.IsPermission(err) {
				return nil, fmt.Errorf("failed to read deliverable file %s: %w", filePath, err)
			}
			// パーミッションエラーのみ無視（スキップしてログは記録しない）
			continue
		}
		deliverables = append(deliverables, &del)
	}

	// ID でソート
	sort.Slice(deliverables, func(i, j int) bool {
		return deliverables[i].ID < deliverables[j].ID
	})

	return deliverables, nil
}

// validateObjectiveReference は Objective 参照の存在を確認
func (h *DeliverableHandler) validateObjectiveReference(ctx context.Context, objectiveID string) error {
	if h.objectiveHandler == nil {
		return nil // ハンドラーがない場合はスキップ
	}

	_, err := h.objectiveHandler.Get(ctx, objectiveID)
	if err == ErrEntityNotFound {
		return fmt.Errorf("referenced objective not found: %s", objectiveID)
	}
	return err
}

// GetDeliverablesByObjective は指定 Objective に紐づく Deliverable を取得
func (h *DeliverableHandler) GetDeliverablesByObjective(ctx context.Context, objectiveID string) ([]*DeliverableEntity, error) {
	all, err := h.getAllDeliverables(ctx)
	if err != nil {
		return nil, err
	}

	var result []*DeliverableEntity
	for _, del := range all {
		if del.ObjectiveID == objectiveID {
			result = append(result, del)
		}
	}

	return result, nil
}

// Deliverable オプション関数

// WithDeliverableDescription は Deliverable の説明を設定
func WithDeliverableDescription(desc string) EntityOption {
	return func(v any) {
		if del, ok := v.(*DeliverableEntity); ok {
			del.Description = desc
		}
	}
}

// WithDeliverableObjective は Deliverable の紐づく Objective を設定
func WithDeliverableObjective(objectiveID string) EntityOption {
	return func(v any) {
		if del, ok := v.(*DeliverableEntity); ok {
			del.ObjectiveID = objectiveID
		}
	}
}

// WithDeliverableFormat は Deliverable のフォーマットを設定
func WithDeliverableFormat(format DeliverableFormat) EntityOption {
	return func(v any) {
		if del, ok := v.(*DeliverableEntity); ok {
			del.Format = format
		}
	}
}

// WithDeliverableAcceptanceCriteria は Deliverable の受入基準を設定
func WithDeliverableAcceptanceCriteria(criteria []string) EntityOption {
	return func(v any) {
		if del, ok := v.(*DeliverableEntity); ok {
			del.AcceptanceCriteria = criteria
		}
	}
}

// WithDeliverableStatus は Deliverable のステータスを設定
func WithDeliverableStatus(status DeliverableStatus) EntityOption {
	return func(v any) {
		if del, ok := v.(*DeliverableEntity); ok {
			del.Status = status
		}
	}
}

// WithDeliverableProgress は Deliverable の進捗を設定
func WithDeliverableProgress(progress int) EntityOption {
	return func(v any) {
		if del, ok := v.(*DeliverableEntity); ok {
			if progress >= 0 && progress <= 100 {
				del.Progress = progress
			}
		}
	}
}

// WithDeliverableOwner は Deliverable のオーナーを設定
func WithDeliverableOwner(owner string) EntityOption {
	return func(v any) {
		if del, ok := v.(*DeliverableEntity); ok {
			del.Metadata.Owner = owner
		}
	}
}

// WithDeliverableTags は Deliverable のタグを設定
func WithDeliverableTags(tags []string) EntityOption {
	return func(v any) {
		if del, ok := v.(*DeliverableEntity); ok {
			del.Metadata.Tags = tags
		}
	}
}
