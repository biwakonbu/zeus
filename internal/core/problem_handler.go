package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ProblemHandler は ProblemEntity エンティティのハンドラー
// 個別ファイル (problems/prob-NNN.yaml) で管理
type ProblemHandler struct {
	fileStore          FileStore
	sanitizer          *Sanitizer
	objectiveHandler   *ObjectiveHandler
	deliverableHandler *DeliverableHandler
	idCounterManager   *IDCounterManager
}

// NewProblemHandler は新しい ProblemHandler を作成
func NewProblemHandler(fs FileStore, objHandler *ObjectiveHandler, delHandler *DeliverableHandler, idMgr *IDCounterManager) *ProblemHandler {
	return &ProblemHandler{
		fileStore:          fs,
		sanitizer:          NewSanitizer(),
		objectiveHandler:   objHandler,
		deliverableHandler: delHandler,
		idCounterManager:   idMgr,
	}
}

// Type はエンティティタイプを返す
func (h *ProblemHandler) Type() string {
	return "problem"
}

// Add は Problem を追加
func (h *ProblemHandler) Add(ctx context.Context, name string, opts ...EntityOption) (*AddResult, error) {
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
	id := fmt.Sprintf("prob-%03d", nextNum)

	now := Now()
	problem := &ProblemEntity{
		ID:       id,
		Title:    sanitizedName,
		Status:   ProblemStatusOpen,
		Severity: ProblemSeverityMedium,
		Metadata: Metadata{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// オプション適用
	for _, opt := range opts {
		opt(problem)
	}

	// 参照の存在確認
	if problem.ObjectiveID != "" {
		if err := h.validateObjectiveReference(ctx, problem.ObjectiveID); err != nil {
			return nil, err
		}
	}
	if problem.DeliverableID != "" {
		if err := h.validateDeliverableReference(ctx, problem.DeliverableID); err != nil {
			return nil, err
		}
	}

	// バリデーション
	if err := problem.Validate(); err != nil {
		return nil, err
	}

	// ファイル書き込み
	filePath := filepath.Join("problems", id+".yaml")
	if err := h.fileStore.WriteYaml(ctx, filePath, problem); err != nil {
		return nil, err
	}

	return &AddResult{
		Success: true,
		ID:      id,
		Entity:  h.Type(),
	}, nil
}

// List は Problem 一覧を取得
func (h *ProblemHandler) List(ctx context.Context, filter *ListFilter) (*ListResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	problems, err := h.getAllProblems(ctx)
	if err != nil {
		return nil, err
	}

	// フィルタ適用
	if filter != nil && filter.Status != "" {
		filtered := []*ProblemEntity{}
		for _, prob := range problems {
			if string(prob.Status) == filter.Status {
				filtered = append(filtered, prob)
			}
		}
		problems = filtered
	}

	// Limit 適用
	if filter != nil && filter.Limit > 0 && len(problems) > filter.Limit {
		problems = problems[:filter.Limit]
	}

	return &ListResult{
		Entity: h.Type() + "s",
		Items:  []Task{},
		Total:  len(problems),
	}, nil
}

// Get は Problem を取得
func (h *ProblemHandler) Get(ctx context.Context, id string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// ID バリデーション
	if err := ValidateID("problem", id); err != nil {
		return nil, err
	}

	filePath := filepath.Join("problems", id+".yaml")
	var problem ProblemEntity
	if err := h.fileStore.ReadYaml(ctx, filePath, &problem); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrEntityNotFound
		}
		return nil, err
	}

	return &problem, nil
}

// Update は Problem を更新
func (h *ProblemHandler) Update(ctx context.Context, id string, update any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("problem", id); err != nil {
		return err
	}

	// 既存を取得
	existing, err := h.Get(ctx, id)
	if err != nil {
		return err
	}
	existingProb := existing.(*ProblemEntity)

	// 更新適用
	if prob, ok := update.(*ProblemEntity); ok {
		prob.ID = id // ID は変更不可
		prob.Metadata.CreatedAt = existingProb.Metadata.CreatedAt
		prob.Metadata.UpdatedAt = Now()

		// 参照の存在確認
		if prob.ObjectiveID != "" && prob.ObjectiveID != existingProb.ObjectiveID {
			if err := h.validateObjectiveReference(ctx, prob.ObjectiveID); err != nil {
				return err
			}
		}
		if prob.DeliverableID != "" && prob.DeliverableID != existingProb.DeliverableID {
			if err := h.validateDeliverableReference(ctx, prob.DeliverableID); err != nil {
				return err
			}
		}

		// バリデーション
		if err := prob.Validate(); err != nil {
			return err
		}

		filePath := filepath.Join("problems", id+".yaml")
		return h.fileStore.WriteYaml(ctx, filePath, prob)
	}

	return fmt.Errorf("invalid update type: expected *ProblemEntity")
}

// Delete は Problem を削除
func (h *ProblemHandler) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// ID バリデーション
	if err := ValidateID("problem", id); err != nil {
		return err
	}

	// 存在確認
	if _, err := h.Get(ctx, id); err != nil {
		return err
	}

	// ファイル削除
	filePath := filepath.Join("problems", id+".yaml")
	return h.fileStore.Delete(ctx, filePath)
}

// getNextIDNumber は次の ID 番号を取得（O(1)）
func (h *ProblemHandler) getNextIDNumber(ctx context.Context) (int, error) {
	if h.idCounterManager != nil {
		return h.idCounterManager.GetNextID(ctx, "problem")
	}
	// フォールバック: 従来の O(N) 方式
	return h.getNextIDNumberLegacy(ctx)
}

// getNextIDNumberLegacy は従来の O(N) 方式で次の ID 番号を取得
func (h *ProblemHandler) getNextIDNumberLegacy(ctx context.Context) (int, error) {
	problems, err := h.getAllProblems(ctx)
	if err != nil {
		return 1, nil
	}

	maxNum := 0
	for _, prob := range problems {
		var num int
		if _, err := fmt.Sscanf(prob.ID, "prob-%d", &num); err == nil {
			if num > maxNum {
				maxNum = num
			}
		}
	}

	return maxNum + 1, nil
}

// getAllProblems は全 Problem を取得
func (h *ProblemHandler) getAllProblems(ctx context.Context) ([]*ProblemEntity, error) {
	files, err := h.fileStore.ListDir(ctx, "problems")
	if err != nil {
		if os.IsNotExist(err) {
			return []*ProblemEntity{}, nil
		}
		return nil, err
	}

	var problems []*ProblemEntity
	for _, file := range files {
		if !strings.HasSuffix(file, ".yaml") {
			continue
		}

		id := strings.TrimSuffix(filepath.Base(file), ".yaml")
		if err := ValidateID("problem", id); err != nil {
			continue
		}

		filePath := filepath.Join("problems", file)
		var prob ProblemEntity
		if err := h.fileStore.ReadYaml(ctx, filePath, &prob); err != nil {
			if !os.IsPermission(err) {
				return nil, fmt.Errorf("failed to read problem file %s: %w", filePath, err)
			}
			continue
		}
		problems = append(problems, &prob)
	}

	sort.Slice(problems, func(i, j int) bool {
		return problems[i].ID < problems[j].ID
	})

	return problems, nil
}

// GetProblemsBySeverity は指定された重大度の Problem を取得
func (h *ProblemHandler) GetProblemsBySeverity(ctx context.Context, severity ProblemSeverity) ([]*ProblemEntity, error) {
	all, err := h.getAllProblems(ctx)
	if err != nil {
		return nil, err
	}

	var result []*ProblemEntity
	for _, prob := range all {
		if prob.Severity == severity {
			result = append(result, prob)
		}
	}

	return result, nil
}

// validateObjectiveReference は Objective 参照の存在を確認
func (h *ProblemHandler) validateObjectiveReference(ctx context.Context, objectiveID string) error {
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
func (h *ProblemHandler) validateDeliverableReference(ctx context.Context, deliverableID string) error {
	if h.deliverableHandler == nil {
		return nil
	}

	_, err := h.deliverableHandler.Get(ctx, deliverableID)
	if err == ErrEntityNotFound {
		return fmt.Errorf("referenced deliverable not found: %s", deliverableID)
	}
	return err
}

// Problem オプション関数

// WithProblemSeverity は Problem の重大度を設定
func WithProblemSeverity(severity ProblemSeverity) EntityOption {
	return func(v any) {
		if prob, ok := v.(*ProblemEntity); ok {
			prob.Severity = severity
		}
	}
}

// WithProblemStatus は Problem のステータスを設定
func WithProblemStatus(status ProblemStatus) EntityOption {
	return func(v any) {
		if prob, ok := v.(*ProblemEntity); ok {
			prob.Status = status
		}
	}
}

// WithProblemObjective は Problem の Objective を設定
func WithProblemObjective(objectiveID string) EntityOption {
	return func(v any) {
		if prob, ok := v.(*ProblemEntity); ok {
			prob.ObjectiveID = objectiveID
		}
	}
}

// WithProblemDeliverable は Problem の Deliverable を設定
func WithProblemDeliverable(deliverableID string) EntityOption {
	return func(v any) {
		if prob, ok := v.(*ProblemEntity); ok {
			prob.DeliverableID = deliverableID
		}
	}
}

// WithProblemDescription は Problem の説明を設定
func WithProblemDescription(desc string) EntityOption {
	return func(v any) {
		if prob, ok := v.(*ProblemEntity); ok {
			prob.Description = desc
		}
	}
}

// WithProblemImpact は Problem の影響を設定
func WithProblemImpact(impact string) EntityOption {
	return func(v any) {
		if prob, ok := v.(*ProblemEntity); ok {
			prob.Impact = impact
		}
	}
}

// WithProblemRootCause は Problem の根本原因を設定
func WithProblemRootCause(rootCause string) EntityOption {
	return func(v any) {
		if prob, ok := v.(*ProblemEntity); ok {
			prob.RootCause = rootCause
		}
	}
}

// WithProblemPotentialSolutions は Problem の潜在的解決策を設定
func WithProblemPotentialSolutions(solutions []string) EntityOption {
	return func(v any) {
		if prob, ok := v.(*ProblemEntity); ok {
			prob.PotentialSolutions = solutions
		}
	}
}

// WithProblemReportedBy は Problem の報告者を設定
func WithProblemReportedBy(reportedBy string) EntityOption {
	return func(v any) {
		if prob, ok := v.(*ProblemEntity); ok {
			prob.ReportedBy = reportedBy
		}
	}
}

// WithProblemAssignedTo は Problem の担当者を設定
func WithProblemAssignedTo(assignedTo string) EntityOption {
	return func(v any) {
		if prob, ok := v.(*ProblemEntity); ok {
			prob.AssignedTo = assignedTo
		}
	}
}
