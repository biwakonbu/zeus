package core

import (
	"context"
	"fmt"
	"os"
)

// ConstraintHandler は ConstraintEntity エンティティのハンドラー
// 単一ファイル (constraints.yaml) で管理
type ConstraintHandler struct {
	fileStore FileStore
	sanitizer *Sanitizer
}

// NewConstraintHandler は新しい ConstraintHandler を作成
func NewConstraintHandler(fs FileStore) *ConstraintHandler {
	return &ConstraintHandler{
		fileStore: fs,
		sanitizer: NewSanitizer(),
	}
}

// Type はエンティティタイプを返す
func (h *ConstraintHandler) Type() string {
	return "constraint"
}

// Add は Constraint を追加
func (h *ConstraintHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// サニタイズ
	sanitizedName, err := h.sanitizer.SanitizeString("title", name)
	if err != nil {
		return nil, fmt.Errorf("invalid title: %w", err)
	}

	// 既存のファイルを読み込む
	file, err := h.loadConstraintsFile(ctx)
	if err != nil {
		return nil, err
	}

	// 次の ID を生成
	nextNum := h.getNextIDNumber(file.Constraints)
	id := fmt.Sprintf("const-%03d", nextNum)

	constraint := ConstraintEntity{
		ID:       id,
		Title:    sanitizedName,
		Category: ConstraintCategoryTechnical,
	}

	// オプション適用
	for _, opt := range opts {
		opt(&constraint)
	}

	// バリデーション
	if err := constraint.Validate(); err != nil {
		return nil, err
	}

	// 制約を追加
	file.Constraints = append(file.Constraints, constraint)
	file.Metadata.UpdatedAt = Now()

	// ファイル書き込み
	if err := h.fileStore.WriteYaml(ctx, "constraints.yaml", file); err != nil {
		return nil, err
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  h.Type(),
	}, nil
}

// List は Constraint 一覧を取得
func (h *ConstraintHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	file, err := h.loadConstraintsFile(ctx)
	if err != nil {
		return nil, err
	}

	// Limit 適用
	constraints := file.Constraints
	if filter != nil && filter.Limit > 0 && len(constraints) > filter.Limit {
		constraints = constraints[:filter.Limit]
	}

	return &ListResult{
		Entity: h.Type() + "s",
		Items:  []Task{},
		Total:  len(constraints),
	}, nil
}

// Get は Constraint を取得
func (h *ConstraintHandler) Get(ctx context.Context, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ID バリデーション
	if err := ValidateID("constraint", id); err != nil {
		return nil, err
	}

	file, err := h.loadConstraintsFile(ctx)
	if err != nil {
		return nil, err
	}

	for i := range file.Constraints {
		if file.Constraints[i].ID == id {
			return &file.Constraints[i], nil
		}
	}

	return nil, ErrEntityNotFound
}

// Update は Constraint を更新
func (h *ConstraintHandler) Update(ctx context.Context, id string, update any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("constraint", id); err != nil {
		return err
	}

	file, err := h.loadConstraintsFile(ctx)
	if err != nil {
		return err
	}

	// 更新対象を検索
	found := false
	for i := range file.Constraints {
		if file.Constraints[i].ID == id {
			if constraint, ok := update.(*ConstraintEntity); ok {
				constraint.ID = id // ID は変更不可

				// バリデーション
				if err := constraint.Validate(); err != nil {
					return err
				}

				file.Constraints[i] = *constraint
				found = true
				break
			}
			return fmt.Errorf("invalid update type: expected *ConstraintEntity")
		}
	}

	if !found {
		return ErrEntityNotFound
	}

	file.Metadata.UpdatedAt = Now()
	return h.fileStore.WriteYaml(ctx, "constraints.yaml", file)
}

// Delete は Constraint を削除
func (h *ConstraintHandler) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("constraint", id); err != nil {
		return err
	}

	file, err := h.loadConstraintsFile(ctx)
	if err != nil {
		return err
	}

	// 削除対象を検索
	found := false
	newConstraints := make([]ConstraintEntity, 0, len(file.Constraints))
	for _, c := range file.Constraints {
		if c.ID == id {
			found = true
			continue
		}
		newConstraints = append(newConstraints, c)
	}

	if !found {
		return ErrEntityNotFound
	}

	file.Constraints = newConstraints
	file.Metadata.UpdatedAt = Now()
	return h.fileStore.WriteYaml(ctx, "constraints.yaml", file)
}

// loadConstraintsFile は constraints.yaml を読み込む
func (h *ConstraintHandler) loadConstraintsFile(ctx context.Context) (*ConstraintsFile, error) {
	var file ConstraintsFile
	if err := h.fileStore.ReadYaml(ctx, "constraints.yaml", &file); err != nil {
		if os.IsNotExist(err) {
			// ファイルが存在しない場合は空のファイルを作成
			now := Now()
			file = ConstraintsFile{
				Constraints: []ConstraintEntity{},
				Metadata: Metadata{
					CreatedAt: now,
					UpdatedAt: now,
				},
			}
			return &file, nil
		}
		return nil, err
	}
	return &file, nil
}

// getNextIDNumber は次の ID 番号を取得
func (h *ConstraintHandler) getNextIDNumber(constraints []ConstraintEntity) int {
	maxNum := 0
	for _, c := range constraints {
		var num int
		if _, err := fmt.Sscanf(c.ID, "const-%d", &num); err == nil {
			if num > maxNum {
				maxNum = num
			}
		}
	}
	return maxNum + 1
}

// GetAllConstraints は全 Constraint を取得するヘルパー
func (h *ConstraintHandler) GetAllConstraints(ctx context.Context) ([]ConstraintEntity, error) {
	file, err := h.loadConstraintsFile(ctx)
	if err != nil {
		return nil, err
	}
	return file.Constraints, nil
}

// GetConstraintsByCategory は指定されたカテゴリの Constraint を取得
func (h *ConstraintHandler) GetConstraintsByCategory(ctx context.Context, category ConstraintCategory) ([]ConstraintEntity, error) {
	all, err := h.GetAllConstraints(ctx)
	if err != nil {
		return nil, err
	}

	var result []ConstraintEntity
	for _, c := range all {
		if c.Category == category {
			result = append(result, c)
		}
	}
	return result, nil
}

// GetNonNegotiableConstraints は交渉不可の Constraint を取得
func (h *ConstraintHandler) GetNonNegotiableConstraints(ctx context.Context) ([]ConstraintEntity, error) {
	all, err := h.GetAllConstraints(ctx)
	if err != nil {
		return nil, err
	}

	var result []ConstraintEntity
	for _, c := range all {
		if c.NonNegotiable {
			result = append(result, c)
		}
	}
	return result, nil
}

// Constraint オプション関数

// WithConstraintCategory は Constraint のカテゴリを設定
func WithConstraintCategory(category ConstraintCategory) EntityOption {
	return func(v any) {
		if c, ok := v.(*ConstraintEntity); ok {
			c.Category = category
		}
	}
}

// WithConstraintDescription は Constraint の説明を設定
func WithConstraintDescription(desc string) EntityOption {
	return func(v any) {
		if c, ok := v.(*ConstraintEntity); ok {
			c.Description = desc
		}
	}
}

// WithConstraintSource は Constraint の出典を設定
func WithConstraintSource(source string) EntityOption {
	return func(v any) {
		if c, ok := v.(*ConstraintEntity); ok {
			c.Source = source
		}
	}
}

// WithConstraintImpact は Constraint の影響を設定
func WithConstraintImpact(impact []string) EntityOption {
	return func(v any) {
		if c, ok := v.(*ConstraintEntity); ok {
			c.Impact = impact
		}
	}
}

// WithConstraintNonNegotiable は Constraint の交渉不可フラグを設定
func WithConstraintNonNegotiable(nonNegotiable bool) EntityOption {
	return func(v any) {
		if c, ok := v.(*ConstraintEntity); ok {
			c.NonNegotiable = nonNegotiable
		}
	}
}
