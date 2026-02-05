package core

import (
	"context"
	"fmt"
	"os"
)

// VisionHandler は Vision エンティティのハンドラー
// Vision は単一ファイル (vision.yaml) で管理される特殊なエンティティ
type VisionHandler struct {
	fileStore FileStore
	sanitizer *Sanitizer
}

// NewVisionHandler は新しい VisionHandler を作成
func NewVisionHandler(fs FileStore) *VisionHandler {
	return &VisionHandler{
		fileStore: fs,
		sanitizer: NewSanitizer(),
	}
}

// Type はエンティティタイプを返す
func (h *VisionHandler) Type() string {
	return "vision"
}

// Add は Vision を作成（既存がある場合は更新）
// Vision はプロジェクトに1つしか存在しないため、Add は Create or Update として動作
func (h *VisionHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// サニタイズ
	sanitizedName, err := h.sanitizer.SanitizeString("title", name)
	if err != nil {
		return nil, fmt.Errorf("invalid title: %w", err)
	}

	now := Now()
	id := "vision-001" // Vision は常に同じ ID

	vision := &Vision{
		ID:     id,
		Title:  sanitizedName,
		Status: VisionStatusDraft,
		Metadata: Metadata{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// オプション適用
	for _, opt := range opts {
		opt(vision)
	}

	// バリデーション
	if err := vision.Validate(); err != nil {
		return nil, err
	}

	// 既存の Vision を確認
	existing, _ := h.Get(ctx, id)
	if existing != nil {
		// 既存がある場合は更新（CreatedAt は保持）
		if ev, ok := existing.(*Vision); ok {
			vision.Metadata.CreatedAt = ev.Metadata.CreatedAt
		}
	}

	// ファイル書き込み
	if err := h.fileStore.WriteYaml(ctx, "vision.yaml", vision); err != nil {
		return nil, err
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  h.Type(),
	}, nil
}

// List は Vision 一覧を取得（単一なので配列長は最大1）
func (h *VisionHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	_, err := h.Get(ctx, "vision-001")
	if err != nil {
		if os.IsNotExist(err) || err == ErrEntityNotFound {
			// Vision が存在しない場合は空のリストを返す
			return &ListResult{
				Entity: h.Type(),
				Items:  []ListItem{}, // 互換性のため Task スライスを使用
				Total:  0,
			}, nil
		}
		return nil, err
	}

	// ListResult.Items は []ListItem なので、空を返す
	// Vision は単一エンティティなので Get を使用することを推奨
	return &ListResult{
		Entity: h.Type(),
		Items:  []ListItem{},
		Total:  1,
	}, nil
}

// Get は Vision を取得
func (h *VisionHandler) Get(ctx context.Context, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ID バリデーション
	if err := ValidateID("vision", id); err != nil {
		return nil, err
	}

	var vision Vision
	if err := h.fileStore.ReadYaml(ctx, "vision.yaml", &vision); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrEntityNotFound
		}
		return nil, err
	}

	return &vision, nil
}

// Update は Vision を更新
func (h *VisionHandler) Update(ctx context.Context, id string, update any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("vision", id); err != nil {
		return err
	}

	// 既存の Vision を取得
	existing, err := h.Get(ctx, id)
	if err != nil {
		return err
	}
	existingVision := existing.(*Vision)

	// 更新適用
	if vision, ok := update.(*Vision); ok {
		vision.ID = id // ID は変更不可
		vision.Metadata.CreatedAt = existingVision.Metadata.CreatedAt
		vision.Metadata.UpdatedAt = Now()

		// バリデーション
		if err := vision.Validate(); err != nil {
			return err
		}

		return h.fileStore.WriteYaml(ctx, "vision.yaml", vision)
	}

	return fmt.Errorf("invalid update type: expected *Vision")
}

// Delete は Vision を削除（Vision は削除不可）
func (h *VisionHandler) Delete(ctx context.Context, id string) error {
	return fmt.Errorf("vision cannot be deleted: vision is a core entity that must exist")
}

// GetVision は Vision を直接取得するヘルパー
func (h *VisionHandler) GetVision(ctx context.Context) (*Vision, error) {
	result, err := h.Get(ctx, "vision-001")
	if err != nil {
		return nil, err
	}
	return result.(*Vision), nil
}

// Vision オプション関数

// WithVisionStatement は Vision の Statement を設定
func WithVisionStatement(statement string) EntityOption {
	return func(v any) {
		if vision, ok := v.(*Vision); ok {
			vision.Statement = statement
		}
	}
}

// WithVisionSuccessCriteria は Vision の成功基準を設定
func WithVisionSuccessCriteria(criteria []string) EntityOption {
	return func(v any) {
		if vision, ok := v.(*Vision); ok {
			vision.SuccessCriteria = criteria
		}
	}
}

// WithVisionStatus は Vision のステータスを設定
func WithVisionStatus(status VisionStatus) EntityOption {
	return func(v any) {
		if vision, ok := v.(*Vision); ok {
			vision.Status = status
		}
	}
}

// WithVisionOwner は Vision のオーナーを設定
func WithVisionOwner(owner string) EntityOption {
	return func(v any) {
		if vision, ok := v.(*Vision); ok {
			vision.Metadata.Owner = owner
		}
	}
}

// WithVisionTags は Vision のタグを設定
func WithVisionTags(tags []string) EntityOption {
	return func(v any) {
		if vision, ok := v.(*Vision); ok {
			vision.Metadata.Tags = tags
		}
	}
}
