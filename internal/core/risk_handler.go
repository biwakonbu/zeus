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

// RiskHandler は RiskEntity エンティティのハンドラー
// 個別ファイル (risks/risk-{uuid}.yaml) で管理
// RiskScore は probability × impact から自動計算
type RiskHandler struct {
	fileStore          FileStore
	sanitizer          *Sanitizer
	objectiveHandler   *ObjectiveHandler
	deliverableHandler *DeliverableHandler
}

// NewRiskHandler は新しい RiskHandler を作成
func NewRiskHandler(fs FileStore, objHandler *ObjectiveHandler, delHandler *DeliverableHandler, _ *IDCounterManager) *RiskHandler {
	return &RiskHandler{
		fileStore:          fs,
		sanitizer:          NewSanitizer(),
		objectiveHandler:   objHandler,
		deliverableHandler: delHandler,
	}
}

// Type はエンティティタイプを返す
func (h *RiskHandler) Type() string {
	return "risk"
}

// Add は Risk を追加
// RiskScore は probability と impact から自動計算される
func (h *RiskHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
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
	risk := &RiskEntity{
		ID:          id,
		Title:       sanitizedName,
		Status:      RiskStatusIdentified,
		Probability: RiskProbabilityMedium,
		Impact:      RiskImpactMedium,
		Metadata: Metadata{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// オプション適用
	for _, opt := range opts {
		opt(risk)
	}

	// 参照の存在確認
	if risk.ObjectiveID != "" {
		if err := h.validateObjectiveReference(ctx, risk.ObjectiveID); err != nil {
			return nil, err
		}
	}
	if risk.DeliverableID != "" {
		if err := h.validateDeliverableReference(ctx, risk.DeliverableID); err != nil {
			return nil, err
		}
	}

	// バリデーション（RiskScore は Validate 内で自動計算される）
	if err := risk.Validate(); err != nil {
		return nil, err
	}

	// ファイル書き込み
	filePath := filepath.Join("risks", id+".yaml")
	if err := h.fileStore.WriteYaml(ctx, filePath, risk); err != nil {
		return nil, err
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  h.Type(),
	}, nil
}

// List は Risk 一覧を取得
func (h *RiskHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	risks, err := h.getAllRisks(ctx)
	if err != nil {
		return nil, err
	}

	// フィルタ適用
	if filter != nil && filter.Status != "" {
		filtered := []*RiskEntity{}
		for _, risk := range risks {
			if string(risk.Status) == filter.Status {
				filtered = append(filtered, risk)
			}
		}
		risks = filtered
	}

	// Limit 適用
	if filter != nil && filter.Limit > 0 && len(risks) > filter.Limit {
		risks = risks[:filter.Limit]
	}

	return &ListResult{
		Entity: h.Type() + "s",
		Items:  []Task{},
		Total:  len(risks),
	}, nil
}

// Get は Risk を取得
func (h *RiskHandler) Get(ctx context.Context, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ID バリデーション
	if err := ValidateID("risk", id); err != nil {
		return nil, err
	}

	filePath := filepath.Join("risks", id+".yaml")
	var risk RiskEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &risk); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrEntityNotFound
		}
		return nil, err
	}

	return &risk, nil
}

// Update は Risk を更新
// probability または impact が変更された場合、RiskScore を再計算
func (h *RiskHandler) Update(ctx context.Context, id string, update any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("risk", id); err != nil {
		return err
	}

	// 既存を取得
	existing, err := h.Get(ctx, id)
	if err != nil {
		return err
	}
	existingRisk := existing.(*RiskEntity)

	// 更新適用
	if risk, ok := update.(*RiskEntity); ok {
		risk.ID = id // ID は変更不可
		risk.Metadata.CreatedAt = existingRisk.Metadata.CreatedAt
		risk.Metadata.UpdatedAt = Now()

		// 参照の存在確認
		if risk.ObjectiveID != "" && risk.ObjectiveID != existingRisk.ObjectiveID {
			if err := h.validateObjectiveReference(ctx, risk.ObjectiveID); err != nil {
				return err
			}
		}
		if risk.DeliverableID != "" && risk.DeliverableID != existingRisk.DeliverableID {
			if err := h.validateDeliverableReference(ctx, risk.DeliverableID); err != nil {
				return err
			}
		}

		// バリデーション（RiskScore は Validate 内で自動計算される）
		if err := risk.Validate(); err != nil {
			return err
		}

		filePath := filepath.Join("risks", id+".yaml")
		return h.fileStore.WriteYaml(ctx, filePath, risk)
	}

	return fmt.Errorf("invalid update type: expected *RiskEntity")
}

// Delete は Risk を削除
func (h *RiskHandler) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("risk", id); err != nil {
		return err
	}

	// 存在確認
	if _, err := h.Get(ctx, id); err != nil {
		return err
	}

	// ファイル削除
	filePath := filepath.Join("risks", id+".yaml")
	return h.fileStore.Delete(ctx, filePath)
}

// generateID は UUID 形式の Risk ID を生成
func (h *RiskHandler) generateID() string {
	return fmt.Sprintf("risk-%s", uuid.New().String()[:8])
}

// getAllRisks は全 Risk を取得
func (h *RiskHandler) getAllRisks(ctx context.Context) ([]*RiskEntity, error) {
	files, err := h.fileStore.ListDir(ctx, "risks")
	if err != nil {
		if os.IsNotExist(err) {
			return []*RiskEntity{}, nil
		}
		return nil, err
	}

	var risks []*RiskEntity
	for _, file := range files {
		if !strings.HasSuffix(file, ".yaml") {
			continue
		}

		id := strings.TrimSuffix(filepath.Base(file), ".yaml")
		if err := ValidateID("risk", id); err != nil {
			continue
		}

		filePath := filepath.Join("risks", file)
		var risk RiskEntity
		if err := h.fileStore.ReadYaml(ctx, filePath, &risk); err != nil {
			if !os.IsPermission(err) {
				return nil, fmt.Errorf("failed to read risk file %s: %w", filePath, err)
			}
			continue
		}
		risks = append(risks, &risk)
	}

	sort.Slice(risks, func(i, j int) bool {
		return risks[i].ID < risks[j].ID
	})

	return risks, nil
}

// GetRisksByScore は指定されたスコアの Risk を取得
func (h *RiskHandler) GetRisksByScore(ctx context.Context, score RiskScore) ([]*RiskEntity, error) {
	all, err := h.getAllRisks(ctx)
	if err != nil {
		return nil, err
	}

	var result []*RiskEntity
	for _, risk := range all {
		if risk.RiskScore == score {
			result = append(result, risk)
		}
	}

	return result, nil
}

// GetHighRisks は critical または high スコアの Risk を取得
func (h *RiskHandler) GetHighRisks(ctx context.Context) ([]*RiskEntity, error) {
	all, err := h.getAllRisks(ctx)
	if err != nil {
		return nil, err
	}

	var result []*RiskEntity
	for _, risk := range all {
		if risk.RiskScore == RiskScoreCritical || risk.RiskScore == RiskScoreHigh {
			result = append(result, risk)
		}
	}

	return result, nil
}

// validateObjectiveReference は Objective 参照の存在を確認
func (h *RiskHandler) validateObjectiveReference(ctx context.Context, objectiveID string) error {
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
func (h *RiskHandler) validateDeliverableReference(ctx context.Context, deliverableID string) error {
	if h.deliverableHandler == nil {
		return nil
	}

	_, err := h.deliverableHandler.Get(ctx, deliverableID)
	if err == ErrEntityNotFound {
		return fmt.Errorf("referenced deliverable not found: %s", deliverableID)
	}
	return err
}

// Risk オプション関数

// WithRiskProbability は Risk の発生確率を設定
func WithRiskProbability(probability RiskProbability) EntityOption {
	return func(v any) {
		if risk, ok := v.(*RiskEntity); ok {
			risk.Probability = probability
		}
	}
}

// WithRiskImpact は Risk の影響度を設定
func WithRiskImpact(impact RiskImpact) EntityOption {
	return func(v any) {
		if risk, ok := v.(*RiskEntity); ok {
			risk.Impact = impact
		}
	}
}

// WithRiskStatus は Risk のステータスを設定
func WithRiskStatus(status RiskStatus) EntityOption {
	return func(v any) {
		if risk, ok := v.(*RiskEntity); ok {
			risk.Status = status
		}
	}
}

// WithRiskObjective は Risk の Objective を設定
func WithRiskObjective(objectiveID string) EntityOption {
	return func(v any) {
		if risk, ok := v.(*RiskEntity); ok {
			risk.ObjectiveID = objectiveID
		}
	}
}

// WithRiskDeliverable は Risk の Deliverable を設定
func WithRiskDeliverable(deliverableID string) EntityOption {
	return func(v any) {
		if risk, ok := v.(*RiskEntity); ok {
			risk.DeliverableID = deliverableID
		}
	}
}

// WithRiskDescription は Risk の説明を設定
func WithRiskDescription(desc string) EntityOption {
	return func(v any) {
		if risk, ok := v.(*RiskEntity); ok {
			risk.Description = desc
		}
	}
}

// WithRiskTrigger は Risk のトリガーを設定
func WithRiskTrigger(trigger string) EntityOption {
	return func(v any) {
		if risk, ok := v.(*RiskEntity); ok {
			risk.Trigger = trigger
		}
	}
}

// WithRiskMitigation は Risk の軽減策を設定
func WithRiskMitigation(mitigation RiskMitigation) EntityOption {
	return func(v any) {
		if risk, ok := v.(*RiskEntity); ok {
			risk.Mitigation = mitigation
		}
	}
}

// WithRiskOwner は Risk のオーナーを設定
func WithRiskOwner(owner string) EntityOption {
	return func(v any) {
		if risk, ok := v.(*RiskEntity); ok {
			risk.Owner = owner
		}
	}
}

// WithRiskReviewDate は Risk のレビュー日を設定
func WithRiskReviewDate(reviewDate string) EntityOption {
	return func(v any) {
		if risk, ok := v.(*RiskEntity); ok {
			risk.ReviewDate = reviewDate
		}
	}
}
