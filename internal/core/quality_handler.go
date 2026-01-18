package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// QualityHandler は QualityEntity エンティティのハンドラー
// 個別ファイル (quality/qual-NNN.yaml) で管理
type QualityHandler struct {
	fileStore          FileStore
	sanitizer          *Sanitizer
	deliverableHandler *DeliverableHandler
	idCounterManager   *IDCounterManager
}

// NewQualityHandler は新しい QualityHandler を作成
func NewQualityHandler(fs FileStore, delHandler *DeliverableHandler, idMgr *IDCounterManager) *QualityHandler {
	return &QualityHandler{
		fileStore:          fs,
		sanitizer:          NewSanitizer(),
		deliverableHandler: delHandler,
		idCounterManager:   idMgr,
	}
}

// Type はエンティティタイプを返す
func (h *QualityHandler) Type() string {
	return "quality"
}

// Add は Quality を追加
func (h *QualityHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
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
	id := fmt.Sprintf("qual-%03d", nextNum)

	now := Now()
	quality := &QualityEntity{
		ID:    id,
		Title: sanitizedName,
		Metadata: Metadata{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// オプション適用
	for _, opt := range opts {
		opt(quality)
	}

	// DeliverableID の存在確認（必須）
	if quality.DeliverableID == "" {
		return nil, fmt.Errorf("quality deliverable_id is required")
	}
	if err := h.validateDeliverableReference(ctx, quality.DeliverableID); err != nil {
		return nil, err
	}

	// バリデーション
	if err := quality.Validate(); err != nil {
		return nil, err
	}

	// ファイル書き込み
	filePath := filepath.Join("quality", id+".yaml")
	if err := h.fileStore.WriteYaml(ctx, filePath, quality); err != nil {
		return nil, err
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  h.Type(),
	}, nil
}

// List は Quality 一覧を取得
func (h *QualityHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	qualities, err := h.getAllQualities(ctx)
	if err != nil {
		return nil, err
	}

	// Limit 適用
	if filter != nil && filter.Limit > 0 && len(qualities) > filter.Limit {
		qualities = qualities[:filter.Limit]
	}

	return &ListResult{
		Entity: h.Type(),
		Items:  []Task{},
		Total:  len(qualities),
	}, nil
}

// Get は Quality を取得
func (h *QualityHandler) Get(ctx context.Context, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ID バリデーション
	if err := ValidateID("quality", id); err != nil {
		return nil, err
	}

	filePath := filepath.Join("quality", id+".yaml")
	var quality QualityEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &quality); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrEntityNotFound
		}
		return nil, err
	}

	return &quality, nil
}

// Update は Quality を更新
func (h *QualityHandler) Update(ctx context.Context, id string, update any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("quality", id); err != nil {
		return err
	}

	// 既存を取得
	existing, err := h.Get(ctx, id)
	if err != nil {
		return err
	}
	existingQual := existing.(*QualityEntity)

	// 更新適用
	if qual, ok := update.(*QualityEntity); ok {
		qual.ID = id // ID は変更不可
		qual.Metadata.CreatedAt = existingQual.Metadata.CreatedAt
		qual.Metadata.UpdatedAt = Now()

		// DeliverableID が変更された場合、存在確認
		if qual.DeliverableID != "" && qual.DeliverableID != existingQual.DeliverableID {
			if err := h.validateDeliverableReference(ctx, qual.DeliverableID); err != nil {
				return err
			}
		}

		// バリデーション
		if err := qual.Validate(); err != nil {
			return err
		}

		filePath := filepath.Join("quality", id+".yaml")
		return h.fileStore.WriteYaml(ctx, filePath, qual)
	}

	return fmt.Errorf("invalid update type: expected *QualityEntity")
}

// Delete は Quality を削除
func (h *QualityHandler) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("quality", id); err != nil {
		return err
	}

	// 存在確認
	if _, err := h.Get(ctx, id); err != nil {
		return err
	}

	// ファイル削除
	filePath := filepath.Join("quality", id+".yaml")
	return h.fileStore.Delete(ctx, filePath)
}

// getNextIDNumber は次の ID 番号を取得（O(1)）
func (h *QualityHandler) getNextIDNumber(ctx context.Context) (int, error) {
	if h.idCounterManager != nil {
		return h.idCounterManager.GetNextID(ctx, "quality")
	}
	// フォールバック: 従来の O(N) 方式
	return h.getNextIDNumberLegacy(ctx)
}

// getNextIDNumberLegacy は従来の O(N) 方式で次の ID 番号を取得
func (h *QualityHandler) getNextIDNumberLegacy(ctx context.Context) (int, error) {
	qualities, err := h.getAllQualities(ctx)
	if err != nil {
		return 1, nil
	}

	maxNum := 0
	for _, qual := range qualities {
		var num int
		if _, err := fmt.Sscanf(qual.ID, "qual-%d", &num); err == nil {
			if num > maxNum {
				maxNum = num
			}
		}
	}

	return maxNum + 1, nil
}

// getAllQualities は全 Quality を取得
func (h *QualityHandler) getAllQualities(ctx context.Context) ([]*QualityEntity, error) {
	files, err := h.fileStore.ListDir(ctx, "quality")
	if err != nil {
		if os.IsNotExist(err) {
			return []*QualityEntity{}, nil
		}
		return nil, err
	}

	var qualities []*QualityEntity
	for _, file := range files {
		if !strings.HasSuffix(file, ".yaml") {
			continue
		}

		id := strings.TrimSuffix(filepath.Base(file), ".yaml")
		if err := ValidateID("quality", id); err != nil {
			continue
		}

		filePath := filepath.Join("quality", file)
		var qual QualityEntity
		if err := h.fileStore.ReadYaml(ctx, filePath, &qual); err != nil {
			if !os.IsPermission(err) {
				return nil, fmt.Errorf("failed to read quality file %s: %w", filePath, err)
			}
			continue
		}
		qualities = append(qualities, &qual)
	}

	sort.Slice(qualities, func(i, j int) bool {
		return qualities[i].ID < qualities[j].ID
	})

	return qualities, nil
}

// GetQualitiesByDeliverable は指定 Deliverable に紐づく Quality を取得
func (h *QualityHandler) GetQualitiesByDeliverable(ctx context.Context, deliverableID string) ([]*QualityEntity, error) {
	all, err := h.getAllQualities(ctx)
	if err != nil {
		return nil, err
	}

	var result []*QualityEntity
	for _, qual := range all {
		if qual.DeliverableID == deliverableID {
			result = append(result, qual)
		}
	}

	return result, nil
}

// UpdateMetric は Quality の特定 Metric を更新
func (h *QualityHandler) UpdateMetric(ctx context.Context, qualityID, metricID string, current float64, status MetricStatus) error {
	existing, err := h.Get(ctx, qualityID)
	if err != nil {
		return err
	}
	qual := existing.(*QualityEntity)

	found := false
	for i := range qual.Metrics {
		if qual.Metrics[i].ID == metricID {
			qual.Metrics[i].Current = current
			qual.Metrics[i].Status = status
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("metric not found: %s", metricID)
	}

	qual.Metadata.UpdatedAt = Now()
	filePath := filepath.Join("quality", qualityID+".yaml")
	return h.fileStore.WriteYaml(ctx, filePath, qual)
}

// UpdateGate は Quality の特定 Gate を更新
func (h *QualityHandler) UpdateGate(ctx context.Context, qualityID, gateName string, status GateStatus) error {
	existing, err := h.Get(ctx, qualityID)
	if err != nil {
		return err
	}
	qual := existing.(*QualityEntity)

	found := false
	for i := range qual.Gates {
		if qual.Gates[i].Name == gateName {
			qual.Gates[i].Status = status
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("gate not found: %s", gateName)
	}

	qual.Metadata.UpdatedAt = Now()
	filePath := filepath.Join("quality", qualityID+".yaml")
	return h.fileStore.WriteYaml(ctx, filePath, qual)
}

// validateDeliverableReference は Deliverable 参照の存在を確認
func (h *QualityHandler) validateDeliverableReference(ctx context.Context, deliverableID string) error {
	if h.deliverableHandler == nil {
		return nil
	}

	_, err := h.deliverableHandler.Get(ctx, deliverableID)
	if err == ErrEntityNotFound {
		return fmt.Errorf("referenced deliverable not found: %s", deliverableID)
	}
	return err
}

// Quality オプション関数

// WithQualityDeliverable は Quality の Deliverable を設定
func WithQualityDeliverable(deliverableID string) EntityOption {
	return func(v any) {
		if qual, ok := v.(*QualityEntity); ok {
			qual.DeliverableID = deliverableID
		}
	}
}

// WithQualityMetrics は Quality の Metrics を設定
func WithQualityMetrics(metrics []QualityMetric) EntityOption {
	return func(v any) {
		if qual, ok := v.(*QualityEntity); ok {
			qual.Metrics = metrics
		}
	}
}

// WithQualityGates は Quality の Gates を設定
func WithQualityGates(gates []QualityGate) EntityOption {
	return func(v any) {
		if qual, ok := v.(*QualityEntity); ok {
			qual.Gates = gates
		}
	}
}

// WithQualityReviewer は Quality のレビューアを設定
func WithQualityReviewer(reviewer string) EntityOption {
	return func(v any) {
		if qual, ok := v.(*QualityEntity); ok {
			qual.Reviewer = reviewer
		}
	}
}
