package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DecisionHandler は DecisionEntity エンティティのハンドラー
// 個別ファイル (decisions/dec-NNN.yaml) で管理
// Decision はイミュータブル：一度作成されると更新不可
type DecisionHandler struct {
	fileStore            FileStore
	sanitizer            *Sanitizer
	considerationHandler *ConsiderationHandler
}

// NewDecisionHandler は新しい DecisionHandler を作成
func NewDecisionHandler(fs FileStore, conHandler *ConsiderationHandler) *DecisionHandler {
	return &DecisionHandler{
		fileStore:            fs,
		sanitizer:            NewSanitizer(),
		considerationHandler: conHandler,
	}
}

// Type はエンティティタイプを返す
func (h *DecisionHandler) Type() string {
	return "decision"
}

// Add は Decision を追加
// Decision 作成時に紐づく Consideration のステータスを decided に更新
func (h *DecisionHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
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
	id := fmt.Sprintf("dec-%03d", nextNum)

	now := Now()
	decision := &DecisionEntity{
		ID:        id,
		Title:     sanitizedName,
		DecidedAt: now,
	}

	// オプション適用
	for _, opt := range opts {
		opt(decision)
	}

	// ConsiderationID の存在確認
	if decision.ConsiderationID == "" {
		return nil, fmt.Errorf("decision consideration_id is required")
	}
	if err := h.validateConsiderationReference(ctx, decision.ConsiderationID); err != nil {
		return nil, err
	}

	// バリデーション
	if err := decision.Validate(); err != nil {
		return nil, err
	}

	// ファイル書き込み
	filePath := filepath.Join("decisions", id+".yaml")
	if err := h.fileStore.WriteYaml(ctx, filePath, decision); err != nil {
		return nil, err
	}

	// Consideration のステータスを decided に更新
	if h.considerationHandler != nil {
		if err := h.considerationHandler.SetDecisionID(ctx, decision.ConsiderationID, id); err != nil {
			// ロールバック: Decision ファイルを削除
			_ = h.fileStore.Delete(ctx, filePath)
			return nil, fmt.Errorf("failed to update consideration status: %w", err)
		}
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  h.Type(),
	}, nil
}

// List は Decision 一覧を取得
func (h *DecisionHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	decisions, err := h.getAllDecisions(ctx)
	if err != nil {
		return nil, err
	}

	// Limit 適用
	if filter != nil && filter.Limit > 0 && len(decisions) > filter.Limit {
		decisions = decisions[:filter.Limit]
	}

	return &ListResult{
		Entity: h.Type() + "s",
		Items:  []Task{},
		Total:  len(decisions),
	}, nil
}

// Get は Decision を取得
func (h *DecisionHandler) Get(ctx context.Context, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ID バリデーション
	if err := ValidateID("decision", id); err != nil {
		return nil, err
	}

	filePath := filepath.Join("decisions", id+".yaml")
	var decision DecisionEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &decision); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrEntityNotFound
		}
		return nil, err
	}

	return &decision, nil
}

// Update は Decision を更新（イミュータブルのため常にエラー）
func (h *DecisionHandler) Update(ctx context.Context, id string, update any) error {
	return fmt.Errorf("decision is immutable: cannot update decision %s", id)
}

// Delete は Decision を削除（イミュータブルのため常にエラー）
// Decision は一度作成されると変更・削除ができません。
// 誤った Decision がある場合は、新しい Consideration を作成して正しい Decision を記録してください。
func (h *DecisionHandler) Delete(ctx context.Context, id string) error {
	return fmt.Errorf("decision is immutable: cannot delete decision %s (decisions are permanent records)", id)
}

// getNextIDNumber は次の ID 番号を取得
func (h *DecisionHandler) getNextIDNumber(ctx context.Context) (int, error) {
	decisions, err := h.getAllDecisions(ctx)
	if err != nil {
		return 1, nil
	}

	maxNum := 0
	for _, dec := range decisions {
		var num int
		if _, err := fmt.Sscanf(dec.ID, "dec-%d", &num); err == nil {
			if num > maxNum {
				maxNum = num
			}
		}
	}

	return maxNum + 1, nil
}

// getAllDecisions は全 Decision を取得
func (h *DecisionHandler) getAllDecisions(ctx context.Context) ([]*DecisionEntity, error) {
	files, err := h.fileStore.ListDir(ctx, "decisions")
	if err != nil {
		if os.IsNotExist(err) {
			return []*DecisionEntity{}, nil
		}
		return nil, err
	}

	var decisions []*DecisionEntity
	for _, file := range files {
		if !strings.HasSuffix(file, ".yaml") {
			continue
		}

		id := strings.TrimSuffix(filepath.Base(file), ".yaml")
		if err := ValidateID("decision", id); err != nil {
			continue
		}

		filePath := filepath.Join("decisions", file)
		var dec DecisionEntity
		if err := h.fileStore.ReadYaml(ctx, filePath, &dec); err != nil {
			if !os.IsPermission(err) {
				return nil, fmt.Errorf("failed to read decision file %s: %w", filePath, err)
			}
			continue
		}
		decisions = append(decisions, &dec)
	}

	sort.Slice(decisions, func(i, j int) bool {
		return decisions[i].ID < decisions[j].ID
	})

	return decisions, nil
}

// validateConsiderationReference は Consideration 参照の存在を確認
func (h *DecisionHandler) validateConsiderationReference(ctx context.Context, considerationID string) error {
	if h.considerationHandler == nil {
		return nil
	}

	_, err := h.considerationHandler.Get(ctx, considerationID)
	if err == ErrEntityNotFound {
		return fmt.Errorf("referenced consideration not found: %s", considerationID)
	}
	return err
}

// Decision オプション関数

// WithDecisionConsideration は Decision の Consideration を設定
func WithDecisionConsideration(considerationID string) EntityOption {
	return func(v any) {
		if dec, ok := v.(*DecisionEntity); ok {
			dec.ConsiderationID = considerationID
		}
	}
}

// WithDecisionSelected は Decision の選択されたオプションを設定
func WithDecisionSelected(selected SelectedOption) EntityOption {
	return func(v any) {
		if dec, ok := v.(*DecisionEntity); ok {
			dec.Selected = selected
		}
	}
}

// WithDecisionRejected は Decision の却下されたオプションを設定
func WithDecisionRejected(rejected []RejectedOption) EntityOption {
	return func(v any) {
		if dec, ok := v.(*DecisionEntity); ok {
			dec.Rejected = rejected
		}
	}
}

// WithDecisionRationale は Decision の理由を設定
func WithDecisionRationale(rationale string) EntityOption {
	return func(v any) {
		if dec, ok := v.(*DecisionEntity); ok {
			dec.Rationale = rationale
		}
	}
}

// WithDecisionImpact は Decision の影響を設定
func WithDecisionImpact(impact []string) EntityOption {
	return func(v any) {
		if dec, ok := v.(*DecisionEntity); ok {
			dec.Impact = impact
		}
	}
}

// WithDecisionDecidedBy は Decision の決定者を設定
func WithDecisionDecidedBy(decidedBy string) EntityOption {
	return func(v any) {
		if dec, ok := v.(*DecisionEntity); ok {
			dec.DecidedBy = decidedBy
		}
	}
}
