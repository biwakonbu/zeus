package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// AssumptionHandler は AssumptionEntity エンティティのハンドラー
// 個別ファイル (assumptions/assum-NNN.yaml) で管理
type AssumptionHandler struct {
	fileStore          FileStore
	sanitizer          *Sanitizer
	objectiveHandler   *ObjectiveHandler
	deliverableHandler *DeliverableHandler
}

// NewAssumptionHandler は新しい AssumptionHandler を作成
func NewAssumptionHandler(fs FileStore, objHandler *ObjectiveHandler, delHandler *DeliverableHandler) *AssumptionHandler {
	return &AssumptionHandler{
		fileStore:          fs,
		sanitizer:          NewSanitizer(),
		objectiveHandler:   objHandler,
		deliverableHandler: delHandler,
	}
}

// Type はエンティティタイプを返す
func (h *AssumptionHandler) Type() string {
	return "assumption"
}

// Add は Assumption を追加
func (h *AssumptionHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// サニタイズ
	sanitizedName, err := h.sanitizer.SanitizeString("title", name)
	if err != nil {
		return nil, fmt.Errorf("invalid title: %w", err)
	}

	// 次の ID を生成
	nextNum, err := h.getNextIDNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}
	id := fmt.Sprintf("assum-%03d", nextNum)

	now := Now()
	assumption := &AssumptionEntity{
		ID:     id,
		Title:  sanitizedName,
		Status: AssumptionStatusAssumed,
		Metadata: Metadata{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// オプション適用
	for _, opt := range opts {
		opt(assumption)
	}

	// 参照の存在確認
	if assumption.ObjectiveID != "" {
		if err := h.validateObjectiveReference(ctx, assumption.ObjectiveID); err != nil {
			return nil, err
		}
	}
	if assumption.DeliverableID != "" {
		if err := h.validateDeliverableReference(ctx, assumption.DeliverableID); err != nil {
			return nil, err
		}
	}

	// バリデーション
	if err := assumption.Validate(); err != nil {
		return nil, err
	}

	// ファイル書き込み
	filePath := filepath.Join("assumptions", id+".yaml")
	if err := h.fileStore.WriteYaml(ctx, filePath, assumption); err != nil {
		return nil, err
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  h.Type(),
	}, nil
}

// List は Assumption 一覧を取得
func (h *AssumptionHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	assumptions, err := h.getAllAssumptions(ctx)
	if err != nil {
		return nil, err
	}

	// フィルタ適用
	if filter != nil && filter.Status != "" {
		filtered := []*AssumptionEntity{}
		for _, assum := range assumptions {
			if string(assum.Status) == filter.Status {
				filtered = append(filtered, assum)
			}
		}
		assumptions = filtered
	}

	// Limit 適用
	if filter != nil && filter.Limit > 0 && len(assumptions) > filter.Limit {
		assumptions = assumptions[:filter.Limit]
	}

	return &ListResult{
		Entity: h.Type() + "s",
		Items:  []Task{},
		Total:  len(assumptions),
	}, nil
}

// Get は Assumption を取得
func (h *AssumptionHandler) Get(ctx context.Context, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ID バリデーション
	if err := ValidateID("assumption", id); err != nil {
		return nil, err
	}

	filePath := filepath.Join("assumptions", id+".yaml")
	var assumption AssumptionEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &assumption); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrEntityNotFound
		}
		return nil, err
	}

	return &assumption, nil
}

// Update は Assumption を更新
func (h *AssumptionHandler) Update(ctx context.Context, id string, update any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("assumption", id); err != nil {
		return err
	}

	// 既存を取得
	existing, err := h.Get(ctx, id)
	if err != nil {
		return err
	}
	existingAssum := existing.(*AssumptionEntity)

	// 更新適用
	if assum, ok := update.(*AssumptionEntity); ok {
		assum.ID = id // ID は変更不可
		assum.Metadata.CreatedAt = existingAssum.Metadata.CreatedAt
		assum.Metadata.UpdatedAt = Now()

		// 参照の存在確認
		if assum.ObjectiveID != "" && assum.ObjectiveID != existingAssum.ObjectiveID {
			if err := h.validateObjectiveReference(ctx, assum.ObjectiveID); err != nil {
				return err
			}
		}
		if assum.DeliverableID != "" && assum.DeliverableID != existingAssum.DeliverableID {
			if err := h.validateDeliverableReference(ctx, assum.DeliverableID); err != nil {
				return err
			}
		}

		// バリデーション
		if err := assum.Validate(); err != nil {
			return err
		}

		filePath := filepath.Join("assumptions", id+".yaml")
		return h.fileStore.WriteYaml(ctx, filePath, assum)
	}

	return fmt.Errorf("invalid update type: expected *AssumptionEntity")
}

// Delete は Assumption を削除
func (h *AssumptionHandler) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("assumption", id); err != nil {
		return err
	}

	// 存在確認
	if _, err := h.Get(ctx, id); err != nil {
		return err
	}

	// ファイル削除
	filePath := filepath.Join("assumptions", id+".yaml")
	return h.fileStore.Delete(ctx, filePath)
}

// getNextIDNumber は次の ID 番号を取得
func (h *AssumptionHandler) getNextIDNumber(ctx context.Context) (int, error) {
	assumptions, err := h.getAllAssumptions(ctx)
	if err != nil {
		return 1, nil
	}

	maxNum := 0
	for _, assum := range assumptions {
		var num int
		if _, err := fmt.Sscanf(assum.ID, "assum-%d", &num); err == nil {
			if num > maxNum {
				maxNum = num
			}
		}
	}

	return maxNum + 1, nil
}

// getAllAssumptions は全 Assumption を取得
func (h *AssumptionHandler) getAllAssumptions(ctx context.Context) ([]*AssumptionEntity, error) {
	files, err := h.fileStore.ListDir(ctx, "assumptions")
	if err != nil {
		if os.IsNotExist(err) {
			return []*AssumptionEntity{}, nil
		}
		return nil, err
	}

	var assumptions []*AssumptionEntity
	for _, file := range files {
		if !strings.HasSuffix(file, ".yaml") {
			continue
		}

		id := strings.TrimSuffix(filepath.Base(file), ".yaml")
		if err := ValidateID("assumption", id); err != nil {
			continue
		}

		filePath := filepath.Join("assumptions", file)
		var assum AssumptionEntity
		if err := h.fileStore.ReadYaml(ctx, filePath, &assum); err != nil {
			if !os.IsPermission(err) {
				return nil, fmt.Errorf("failed to read assumption file %s: %w", filePath, err)
			}
			continue
		}
		assumptions = append(assumptions, &assum)
	}

	sort.Slice(assumptions, func(i, j int) bool {
		return assumptions[i].ID < assumptions[j].ID
	})

	return assumptions, nil
}

// ValidateAssumption は Assumption を検証して結果を記録
func (h *AssumptionHandler) ValidateAssumption(ctx context.Context, id string, validation AssumptionValidation, newStatus AssumptionStatus) error {
	existing, err := h.Get(ctx, id)
	if err != nil {
		return err
	}
	assum := existing.(*AssumptionEntity)

	assum.Validation = validation
	assum.Validation.ValidatedAt = Now()
	assum.Status = newStatus
	assum.Metadata.UpdatedAt = Now()

	if err := assum.Validate(); err != nil {
		return err
	}

	filePath := filepath.Join("assumptions", id+".yaml")
	return h.fileStore.WriteYaml(ctx, filePath, assum)
}

// validateObjectiveReference は Objective 参照の存在を確認
func (h *AssumptionHandler) validateObjectiveReference(ctx context.Context, objectiveID string) error {
	if h.objectiveHandler == nil {
		return nil
	}

	_, err := h.objectiveHandler.Get(ctx, objectiveID)
	if err == ErrEntityNotFound {
		return fmt.Errorf("referenced objective not found: %s", objectiveID)
	}
	return err
}

// validateDeliverableReference は Deliverable 参照の存在を確認
func (h *AssumptionHandler) validateDeliverableReference(ctx context.Context, deliverableID string) error {
	if h.deliverableHandler == nil {
		return nil
	}

	_, err := h.deliverableHandler.Get(ctx, deliverableID)
	if err == ErrEntityNotFound {
		return fmt.Errorf("referenced deliverable not found: %s", deliverableID)
	}
	return err
}

// Assumption オプション関数

// WithAssumptionStatus は Assumption のステータスを設定
func WithAssumptionStatus(status AssumptionStatus) EntityOption {
	return func(v any) {
		if assum, ok := v.(*AssumptionEntity); ok {
			assum.Status = status
		}
	}
}

// WithAssumptionObjective は Assumption の Objective を設定
func WithAssumptionObjective(objectiveID string) EntityOption {
	return func(v any) {
		if assum, ok := v.(*AssumptionEntity); ok {
			assum.ObjectiveID = objectiveID
		}
	}
}

// WithAssumptionDeliverable は Assumption の Deliverable を設定
func WithAssumptionDeliverable(deliverableID string) EntityOption {
	return func(v any) {
		if assum, ok := v.(*AssumptionEntity); ok {
			assum.DeliverableID = deliverableID
		}
	}
}

// WithAssumptionDescription は Assumption の説明を設定
func WithAssumptionDescription(desc string) EntityOption {
	return func(v any) {
		if assum, ok := v.(*AssumptionEntity); ok {
			assum.Description = desc
		}
	}
}

// WithAssumptionIfInvalid は Assumption の無効時の対応を設定
func WithAssumptionIfInvalid(ifInvalid string) EntityOption {
	return func(v any) {
		if assum, ok := v.(*AssumptionEntity); ok {
			assum.IfInvalid = ifInvalid
		}
	}
}
