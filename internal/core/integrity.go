package core

import (
	"context"
	"errors"
	"fmt"
)

// 整合性チェック用エラーメッセージ定数
const (
	// 参照エラーメッセージ
	ErrMsgReferencedObjectiveNotFound      = "referenced objective not found"
	ErrMsgReferencedDeliverableNotFound    = "referenced deliverable not found"
	ErrMsgReferencedConsiderationNotFound  = "referenced consideration not found"
	ErrMsgReferencedDecisionNotFound       = "referenced decision not found"
	ErrMsgReferencedSubsystemNotFound      = "referenced subsystem not found"
	ErrMsgReferencedActorNotFound          = "referenced actor not found"
	ErrMsgReferencedUseCaseNotFound        = "referenced usecase not found"
	ErrMsgReferencedActivityNotFound       = "referenced dependency activity not found"
	ErrMsgReferencedParentActivityNotFound = "referenced parent activity not found"
	ErrMsgParentObjectiveNotFound          = "parent objective not found"

	// 必須フィールド欠損メッセージ
	ErrMsgObjectiveIDRequired     = "objective_id is required but missing"
	ErrMsgDeliverableIDRequired   = "deliverable_id is required but missing"
	ErrMsgConsiderationIDRequired = "consideration_id is required but missing"

	// 無効なID形式メッセージ
	ErrMsgInvalidSubsystemIDFormat   = "invalid subsystem ID format"
	ErrMsgInvalidActorIDFormat       = "invalid actor ID format"
	ErrMsgInvalidUseCaseIDFormat     = "invalid usecase ID format"
	ErrMsgInvalidDeliverableIDFormat = "invalid deliverable ID format"

	// 循環参照メッセージ
	ErrMsgCircularParentReference    = "circular parent reference detected"
	ErrMsgCircularDependencyDetected = "circular dependency detected"

	// RelatedDeliverables 用メッセージ
	ErrMsgRelatedDeliverableNotFound      = "referenced deliverable in related_deliverables not found"
	ErrMsgInvalidRelatedDeliverableFormat = "invalid deliverable ID format in related_deliverables"
)

// ParentResolver は親 ID を解決する関数型（汎用循環参照検出用）
type ParentResolver func(id string) string

// IntegrityChecker は参照整合性をチェックする
type IntegrityChecker struct {
	objectiveHandler     *ObjectiveHandler
	deliverableHandler   *DeliverableHandler
	considerationHandler *ConsiderationHandler
	decisionHandler      *DecisionHandler
	problemHandler       *ProblemHandler
	riskHandler          *RiskHandler
	assumptionHandler    *AssumptionHandler
	qualityHandler       *QualityHandler
	usecaseHandler       *UseCaseHandler
	subsystemHandler     *SubsystemHandler
	activityHandler      *ActivityHandler
	actorHandler         *ActorHandler
}

// NewIntegrityChecker は新しい IntegrityChecker を作成
func NewIntegrityChecker(objHandler *ObjectiveHandler, delHandler *DeliverableHandler) *IntegrityChecker {
	return &IntegrityChecker{
		objectiveHandler:   objHandler,
		deliverableHandler: delHandler,
	}
}

// SetConsiderationHandler は ConsiderationHandler を設定
func (c *IntegrityChecker) SetConsiderationHandler(h *ConsiderationHandler) {
	c.considerationHandler = h
}

// SetDecisionHandler は DecisionHandler を設定
func (c *IntegrityChecker) SetDecisionHandler(h *DecisionHandler) {
	c.decisionHandler = h
}

// SetProblemHandler は ProblemHandler を設定
func (c *IntegrityChecker) SetProblemHandler(h *ProblemHandler) {
	c.problemHandler = h
}

// SetRiskHandler は RiskHandler を設定
func (c *IntegrityChecker) SetRiskHandler(h *RiskHandler) {
	c.riskHandler = h
}

// SetAssumptionHandler は AssumptionHandler を設定
func (c *IntegrityChecker) SetAssumptionHandler(h *AssumptionHandler) {
	c.assumptionHandler = h
}

// SetQualityHandler は QualityHandler を設定
func (c *IntegrityChecker) SetQualityHandler(h *QualityHandler) {
	c.qualityHandler = h
}

// SetUseCaseHandler は UseCaseHandler を設定
func (c *IntegrityChecker) SetUseCaseHandler(h *UseCaseHandler) {
	c.usecaseHandler = h
}

// SetSubsystemHandler は SubsystemHandler を設定
func (c *IntegrityChecker) SetSubsystemHandler(h *SubsystemHandler) {
	c.subsystemHandler = h
}

// SetActivityHandler は ActivityHandler を設定
func (c *IntegrityChecker) SetActivityHandler(h *ActivityHandler) {
	c.activityHandler = h
}

// SetActorHandler は ActorHandler を設定
func (c *IntegrityChecker) SetActorHandler(h *ActorHandler) {
	c.actorHandler = h
}

// ReferenceError は参照エラーを表す
type ReferenceError struct {
	SourceType string // "deliverable" or "objective"
	SourceID   string
	TargetType string // "objective"
	TargetID   string
	Message    string
}

// Error は error インターフェースを実装
func (e *ReferenceError) Error() string {
	return fmt.Sprintf("%s %s → %s %s: %s", e.SourceType, e.SourceID, e.TargetType, e.TargetID, e.Message)
}

// CycleError は循環参照エラーを表す
type CycleError struct {
	EntityType string   // "objective"
	Cycle      []string // 循環パスの ID リスト
	Message    string
}

// Error は error インターフェースを実装
func (e *CycleError) Error() string {
	return fmt.Sprintf("%s cycle detected: %v - %s", e.EntityType, e.Cycle, e.Message)
}

// IntegrityResult は整合性チェックの結果
type IntegrityResult struct {
	Valid           bool
	ReferenceErrors []*ReferenceError
	CycleErrors     []*CycleError
	Warnings        []*ReferenceWarning // 警告（エラーではないが注意が必要）
}

// ReferenceWarning は参照警告を表す（エラーではないが注意が必要な参照問題）
type ReferenceWarning struct {
	SourceType string
	SourceID   string
	TargetType string
	TargetID   string
	Message    string
}

// Warning は警告メッセージを返す
func (w *ReferenceWarning) Warning() string {
	return fmt.Sprintf("%s %s → %s %s: %s", w.SourceType, w.SourceID, w.TargetType, w.TargetID, w.Message)
}

// CheckAll は全ての整合性チェックを実行
func (c *IntegrityChecker) CheckAll(ctx context.Context) (*IntegrityResult, error) {
	result := &IntegrityResult{
		Valid:           true,
		ReferenceErrors: []*ReferenceError{},
		CycleErrors:     []*CycleError{},
		Warnings:        []*ReferenceWarning{},
	}

	// 参照チェック
	refErrors, err := c.CheckReferences(ctx)
	if err != nil {
		return nil, fmt.Errorf("reference check failed: %w", err)
	}
	result.ReferenceErrors = refErrors

	// 循環参照チェック
	cycleErrors, err := c.CheckCycles(ctx)
	if err != nil {
		return nil, fmt.Errorf("cycle check failed: %w", err)
	}
	result.CycleErrors = cycleErrors

	// 警告チェック（UseCase → Subsystem 参照など）
	warnings, err := c.CheckWarnings(ctx)
	if err != nil {
		return nil, fmt.Errorf("warning check failed: %w", err)
	}
	result.Warnings = warnings

	// エラーがあれば Valid = false（警告は Valid に影響しない）
	if len(result.ReferenceErrors) > 0 || len(result.CycleErrors) > 0 {
		result.Valid = false
	}

	return result, nil
}

// CheckReferences は参照整合性をチェック
// - Deliverable → Objective 参照
// - Objective → Objective (parent) 参照
// - Decision → Consideration 参照（必須）
// - Quality → Deliverable 参照（必須）
// - UseCase → Objective 参照（必須）
// - Consideration → Objective/Deliverable 参照（任意）
// - Problem → Objective/Deliverable 参照（任意）
// - Risk → Objective/Deliverable 参照（任意）
// - Assumption → Objective/Deliverable 参照（任意）
// - Consideration ← Decision 逆参照（削除時チェック用）
func (c *IntegrityChecker) CheckReferences(ctx context.Context) ([]*ReferenceError, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var errors []*ReferenceError

	// Deliverable → Objective 参照チェック
	delErrors, err := c.checkDeliverableReferences(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, delErrors...)

	// Objective → Objective (parent) 参照チェック
	objErrors, err := c.checkObjectiveParentReferences(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, objErrors...)

	// Decision → Consideration 参照チェック（必須）
	decErrors, err := c.checkDecisionReferences(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, decErrors...)

	// Quality → Deliverable 参照チェック（必須）
	qualErrors, err := c.checkQualityReferences(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, qualErrors...)

	// UseCase → Objective 参照チェック（必須）
	ucErrors, err := c.checkUseCaseObjectiveReferences(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, ucErrors...)

	// Consideration → Objective/Deliverable 参照チェック
	conErrors, err := c.checkConsiderationReferences(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, conErrors...)

	// Problem → Objective/Deliverable 参照チェック
	probErrors, err := c.checkProblemReferences(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, probErrors...)

	// Risk → Objective/Deliverable 参照チェック
	riskErrors, err := c.checkRiskReferences(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, riskErrors...)

	// Assumption → Objective/Deliverable 参照チェック
	assumErrors, err := c.checkAssumptionReferences(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, assumErrors...)

	return errors, nil
}

// checkDeliverableReferences は Deliverable から Objective への参照をチェック
func (c *IntegrityChecker) checkDeliverableReferences(ctx context.Context) ([]*ReferenceError, error) {
	if c.deliverableHandler == nil {
		return []*ReferenceError{}, nil
	}

	deliverables, err := c.deliverableHandler.getAllDeliverables(ctx)
	if err != nil {
		return nil, err
	}

	var errors []*ReferenceError
	for _, del := range deliverables {
		if del.ObjectiveID == "" {
			continue // 参照なしは OK
		}

		// Objective の存在確認
		_, err := c.objectiveHandler.Get(ctx, del.ObjectiveID)
		if err == ErrEntityNotFound {
			errors = append(errors, &ReferenceError{
				SourceType: "deliverable",
				SourceID:   del.ID,
				TargetType: "objective",
				TargetID:   del.ObjectiveID,
				Message:    ErrMsgReferencedObjectiveNotFound,
			})
		} else if err != nil {
			return nil, err
		}
	}

	return errors, nil
}

// checkObjectiveParentReferences は Objective の親参照をチェック
func (c *IntegrityChecker) checkObjectiveParentReferences(ctx context.Context) ([]*ReferenceError, error) {
	if c.objectiveHandler == nil {
		return []*ReferenceError{}, nil
	}

	objectives, err := c.objectiveHandler.getAllObjectives(ctx)
	if err != nil {
		return nil, err
	}

	var errors []*ReferenceError
	for _, obj := range objectives {
		if obj.ParentID == "" {
			continue // 親なしは OK
		}

		// 親 Objective の存在確認
		_, err := c.objectiveHandler.Get(ctx, obj.ParentID)
		if err == ErrEntityNotFound {
			errors = append(errors, &ReferenceError{
				SourceType: "objective",
				SourceID:   obj.ID,
				TargetType: "objective",
				TargetID:   obj.ParentID,
				Message:    ErrMsgParentObjectiveNotFound,
			})
		} else if err != nil {
			return nil, err
		}
	}

	return errors, nil
}

// CheckCycles は循環参照をチェック
// - Objective 階層の循環参照を検出
// - Activity 階層（ParentID）の循環参照を検出
// - Activity 依存関係（Dependencies）の循環参照を検出
func (c *IntegrityChecker) CheckCycles(ctx context.Context) ([]*CycleError, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var errors []*CycleError

	// Objective 階層の循環参照チェック
	objCycles, err := c.checkObjectiveCycles(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, objCycles...)

	// Activity 階層（ParentID）の循環参照チェック
	actParentCycles, err := c.checkActivityParentCycles(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, actParentCycles...)

	// Activity 依存関係（Dependencies）の循環参照チェック
	actDepCycles, err := c.checkActivityDependencyCycles(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, actDepCycles...)

	return errors, nil
}

// checkObjectiveCycles は Objective 階層の循環参照をチェック
func (c *IntegrityChecker) checkObjectiveCycles(ctx context.Context) ([]*CycleError, error) {
	if c.objectiveHandler == nil {
		return []*CycleError{}, nil
	}

	objectives, err := c.objectiveHandler.getAllObjectives(ctx)
	if err != nil {
		return nil, err
	}

	// ID → ParentID のマップを作成
	parentMap := make(map[string]string)
	for _, obj := range objectives {
		parentMap[obj.ID] = obj.ParentID
	}

	// ParentResolver を作成
	resolver := func(id string) string {
		return parentMap[id]
	}

	var errors []*CycleError
	visited := make(map[string]bool)

	for _, obj := range objectives {
		if visited[obj.ID] || obj.ParentID == "" {
			continue
		}

		// 汎用循環検出関数を使用
		cycle := c.detectCycleGeneric(obj.ID, resolver, visited)
		if len(cycle) > 0 {
			errors = append(errors, &CycleError{
				EntityType: "objective",
				Cycle:      cycle,
				Message:    ErrMsgCircularParentReference,
			})
		}
	}

	return errors, nil
}

// checkActivityParentCycles は Activity 階層（ParentID）の循環参照をチェック
func (c *IntegrityChecker) checkActivityParentCycles(ctx context.Context) ([]*CycleError, error) {
	if c.activityHandler == nil {
		return []*CycleError{}, nil
	}

	activities, err := c.activityHandler.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// ID → ParentID のマップを作成
	parentMap := make(map[string]string)
	for _, act := range activities {
		parentMap[act.ID] = act.ParentID
	}

	// ParentResolver を作成
	resolver := func(id string) string {
		return parentMap[id]
	}

	var errors []*CycleError
	visited := make(map[string]bool)

	for _, act := range activities {
		if visited[act.ID] || act.ParentID == "" {
			continue
		}

		// 汎用循環検出関数を使用
		cycle := c.detectCycleGeneric(act.ID, resolver, visited)
		if len(cycle) > 0 {
			errors = append(errors, &CycleError{
				EntityType: "activity",
				Cycle:      cycle,
				Message:    ErrMsgCircularParentReference,
			})
		}
	}

	return errors, nil
}

// checkActivityDependencyCycles は Activity 依存関係（Dependencies）の循環参照をチェック
func (c *IntegrityChecker) checkActivityDependencyCycles(ctx context.Context) ([]*CycleError, error) {
	if c.activityHandler == nil {
		return []*CycleError{}, nil
	}

	// ActivityHandler の DetectDependencyCycles を使用
	cycles, err := c.activityHandler.DetectDependencyCycles(ctx)
	if err != nil {
		return nil, err
	}

	var errors []*CycleError
	for _, cycle := range cycles {
		errors = append(errors, &CycleError{
			EntityType: "activity",
			Cycle:      cycle,
			Message:    ErrMsgCircularDependencyDetected,
		})
	}

	return errors, nil
}

// detectCycleGeneric は汎用的な循環参照検出関数
// ParentResolver を使用して親 ID を解決し、循環を検出する
func (c *IntegrityChecker) detectCycleGeneric(startID string, resolver ParentResolver, globalVisited map[string]bool) []string {
	localVisited := make(map[string]bool)
	path := []string{}

	current := startID
	for current != "" {
		// グローバルに訪問済みなら、このパスは既にチェック済み
		if globalVisited[current] {
			break
		}

		// ローカルで訪問済みなら循環
		if localVisited[current] {
			// 循環パスを構築
			cycleStart := -1
			for i, id := range path {
				if id == current {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				return append(path[cycleStart:], current)
			}
			return path
		}

		localVisited[current] = true
		path = append(path, current)

		// 次の親へ
		parentID := resolver(current)
		if parentID == "" {
			break
		}
		current = parentID
	}

	// 訪問したノードをグローバルに記録
	for id := range localVisited {
		globalVisited[id] = true
	}

	return nil // 循環なし
}

// checkDecisionReferences は Decision から Consideration への参照をチェック（必須）
func (c *IntegrityChecker) checkDecisionReferences(ctx context.Context) ([]*ReferenceError, error) {
	if c.decisionHandler == nil {
		return []*ReferenceError{}, nil
	}

	decisions, err := c.decisionHandler.getAllDecisions(ctx)
	if err != nil {
		return nil, err
	}

	var errors []*ReferenceError
	for _, dec := range decisions {
		// ConsiderationID は必須
		if dec.ConsiderationID == "" {
			errors = append(errors, &ReferenceError{
				SourceType: "decision",
				SourceID:   dec.ID,
				TargetType: "consideration",
				TargetID:   "",
				Message:    ErrMsgConsiderationIDRequired,
			})
			continue
		}

		// Consideration の存在確認
		if c.considerationHandler != nil {
			_, err := c.considerationHandler.Get(ctx, dec.ConsiderationID)
			if err == ErrEntityNotFound {
				errors = append(errors, &ReferenceError{
					SourceType: "decision",
					SourceID:   dec.ID,
					TargetType: "consideration",
					TargetID:   dec.ConsiderationID,
					Message:    ErrMsgReferencedConsiderationNotFound,
				})
			} else if err != nil {
				return nil, err
			}
		}
	}

	return errors, nil
}

// checkQualityReferences は Quality から Deliverable への参照をチェック（必須）
func (c *IntegrityChecker) checkQualityReferences(ctx context.Context) ([]*ReferenceError, error) {
	if c.qualityHandler == nil {
		return []*ReferenceError{}, nil
	}

	qualities, err := c.qualityHandler.getAllQualities(ctx)
	if err != nil {
		return nil, err
	}

	var errors []*ReferenceError
	for _, qual := range qualities {
		// DeliverableID は必須
		if qual.DeliverableID == "" {
			errors = append(errors, &ReferenceError{
				SourceType: "quality",
				SourceID:   qual.ID,
				TargetType: "deliverable",
				TargetID:   "",
				Message:    ErrMsgDeliverableIDRequired,
			})
			continue
		}

		// Deliverable の存在確認
		if c.deliverableHandler != nil {
			_, err := c.deliverableHandler.Get(ctx, qual.DeliverableID)
			if err == ErrEntityNotFound {
				errors = append(errors, &ReferenceError{
					SourceType: "quality",
					SourceID:   qual.ID,
					TargetType: "deliverable",
					TargetID:   qual.DeliverableID,
					Message:    ErrMsgReferencedDeliverableNotFound,
				})
			} else if err != nil {
				return nil, err
			}
		}
	}

	return errors, nil
}

// checkConsiderationReferences は Consideration から Objective/Deliverable への参照をチェック
func (c *IntegrityChecker) checkConsiderationReferences(ctx context.Context) ([]*ReferenceError, error) {
	if c.considerationHandler == nil {
		return []*ReferenceError{}, nil
	}

	considerations, err := c.considerationHandler.getAllConsiderations(ctx)
	if err != nil {
		return nil, err
	}

	var errors []*ReferenceError
	for _, con := range considerations {
		// ObjectiveID のチェック（任意）
		if con.ObjectiveID != "" && c.objectiveHandler != nil {
			_, err := c.objectiveHandler.Get(ctx, con.ObjectiveID)
			if err == ErrEntityNotFound {
				errors = append(errors, &ReferenceError{
					SourceType: "consideration",
					SourceID:   con.ID,
					TargetType: "objective",
					TargetID:   con.ObjectiveID,
					Message:    ErrMsgReferencedObjectiveNotFound,
				})
			} else if err != nil {
				return nil, err
			}
		}

		// DeliverableID のチェック（任意）
		if con.DeliverableID != "" && c.deliverableHandler != nil {
			_, err := c.deliverableHandler.Get(ctx, con.DeliverableID)
			if err == ErrEntityNotFound {
				errors = append(errors, &ReferenceError{
					SourceType: "consideration",
					SourceID:   con.ID,
					TargetType: "deliverable",
					TargetID:   con.DeliverableID,
					Message:    ErrMsgReferencedDeliverableNotFound,
				})
			} else if err != nil {
				return nil, err
			}
		}

		// DecisionID のチェック（任意）
		if con.DecisionID != "" && c.decisionHandler != nil {
			_, err := c.decisionHandler.Get(ctx, con.DecisionID)
			if err == ErrEntityNotFound {
				errors = append(errors, &ReferenceError{
					SourceType: "consideration",
					SourceID:   con.ID,
					TargetType: "decision",
					TargetID:   con.DecisionID,
					Message:    ErrMsgReferencedDecisionNotFound,
				})
			} else if err != nil {
				return nil, err
			}
		}
	}

	return errors, nil
}

// checkProblemReferences は Problem から Objective/Deliverable への参照をチェック
func (c *IntegrityChecker) checkProblemReferences(ctx context.Context) ([]*ReferenceError, error) {
	if c.problemHandler == nil {
		return []*ReferenceError{}, nil
	}

	problems, err := c.problemHandler.getAllProblems(ctx)
	if err != nil {
		return nil, err
	}

	var errors []*ReferenceError
	for _, prob := range problems {
		// ObjectiveID のチェック（任意）
		if prob.ObjectiveID != "" && c.objectiveHandler != nil {
			_, err := c.objectiveHandler.Get(ctx, prob.ObjectiveID)
			if err == ErrEntityNotFound {
				errors = append(errors, &ReferenceError{
					SourceType: "problem",
					SourceID:   prob.ID,
					TargetType: "objective",
					TargetID:   prob.ObjectiveID,
					Message:    ErrMsgReferencedObjectiveNotFound,
				})
			} else if err != nil {
				return nil, err
			}
		}

		// DeliverableID のチェック（任意）
		if prob.DeliverableID != "" && c.deliverableHandler != nil {
			_, err := c.deliverableHandler.Get(ctx, prob.DeliverableID)
			if err == ErrEntityNotFound {
				errors = append(errors, &ReferenceError{
					SourceType: "problem",
					SourceID:   prob.ID,
					TargetType: "deliverable",
					TargetID:   prob.DeliverableID,
					Message:    ErrMsgReferencedDeliverableNotFound,
				})
			} else if err != nil {
				return nil, err
			}
		}
	}

	return errors, nil
}

// checkRiskReferences は Risk から Objective/Deliverable への参照をチェック
func (c *IntegrityChecker) checkRiskReferences(ctx context.Context) ([]*ReferenceError, error) {
	if c.riskHandler == nil {
		return []*ReferenceError{}, nil
	}

	risks, err := c.riskHandler.getAllRisks(ctx)
	if err != nil {
		return nil, err
	}

	var errors []*ReferenceError
	for _, risk := range risks {
		// ObjectiveID のチェック（任意）
		if risk.ObjectiveID != "" && c.objectiveHandler != nil {
			_, err := c.objectiveHandler.Get(ctx, risk.ObjectiveID)
			if err == ErrEntityNotFound {
				errors = append(errors, &ReferenceError{
					SourceType: "risk",
					SourceID:   risk.ID,
					TargetType: "objective",
					TargetID:   risk.ObjectiveID,
					Message:    ErrMsgReferencedObjectiveNotFound,
				})
			} else if err != nil {
				return nil, err
			}
		}

		// DeliverableID のチェック（任意）
		if risk.DeliverableID != "" && c.deliverableHandler != nil {
			_, err := c.deliverableHandler.Get(ctx, risk.DeliverableID)
			if err == ErrEntityNotFound {
				errors = append(errors, &ReferenceError{
					SourceType: "risk",
					SourceID:   risk.ID,
					TargetType: "deliverable",
					TargetID:   risk.DeliverableID,
					Message:    ErrMsgReferencedDeliverableNotFound,
				})
			} else if err != nil {
				return nil, err
			}
		}
	}

	return errors, nil
}

// checkAssumptionReferences は Assumption から Objective/Deliverable への参照をチェック
func (c *IntegrityChecker) checkAssumptionReferences(ctx context.Context) ([]*ReferenceError, error) {
	if c.assumptionHandler == nil {
		return []*ReferenceError{}, nil
	}

	assumptions, err := c.assumptionHandler.getAllAssumptions(ctx)
	if err != nil {
		return nil, err
	}

	var errors []*ReferenceError
	for _, assum := range assumptions {
		// ObjectiveID のチェック（任意）
		if assum.ObjectiveID != "" && c.objectiveHandler != nil {
			_, err := c.objectiveHandler.Get(ctx, assum.ObjectiveID)
			if err == ErrEntityNotFound {
				errors = append(errors, &ReferenceError{
					SourceType: "assumption",
					SourceID:   assum.ID,
					TargetType: "objective",
					TargetID:   assum.ObjectiveID,
					Message:    ErrMsgReferencedObjectiveNotFound,
				})
			} else if err != nil {
				return nil, err
			}
		}

		// DeliverableID のチェック（任意）
		if assum.DeliverableID != "" && c.deliverableHandler != nil {
			_, err := c.deliverableHandler.Get(ctx, assum.DeliverableID)
			if err == ErrEntityNotFound {
				errors = append(errors, &ReferenceError{
					SourceType: "assumption",
					SourceID:   assum.ID,
					TargetType: "deliverable",
					TargetID:   assum.DeliverableID,
					Message:    ErrMsgReferencedDeliverableNotFound,
				})
			} else if err != nil {
				return nil, err
			}
		}
	}

	return errors, nil
}

// checkUseCaseObjectiveReferences は UseCase から Objective への参照をチェック（必須）
func (c *IntegrityChecker) checkUseCaseObjectiveReferences(ctx context.Context) ([]*ReferenceError, error) {
	if c.usecaseHandler == nil {
		return []*ReferenceError{}, nil
	}

	usecases, err := c.usecaseHandler.getAllUseCases(ctx)
	if err != nil {
		return nil, err
	}

	var errors []*ReferenceError
	for _, uc := range usecases {
		// ObjectiveID は必須
		if uc.ObjectiveID == "" {
			errors = append(errors, &ReferenceError{
				SourceType: "usecase",
				SourceID:   uc.ID,
				TargetType: "objective",
				TargetID:   "",
				Message:    ErrMsgObjectiveIDRequired,
			})
			continue
		}

		// Objective の存在確認
		if c.objectiveHandler != nil {
			_, err := c.objectiveHandler.Get(ctx, uc.ObjectiveID)
			if err == ErrEntityNotFound {
				errors = append(errors, &ReferenceError{
					SourceType: "usecase",
					SourceID:   uc.ID,
					TargetType: "objective",
					TargetID:   uc.ObjectiveID,
					Message:    ErrMsgReferencedObjectiveNotFound,
				})
			} else if err != nil {
				return nil, err
			}
		}
	}

	return errors, nil
}

// CheckWarnings は警告レベルの参照問題をチェック
// - UseCase → Subsystem 参照（任意、存在しないサブシステムへの参照は警告）
// - UseCase → Actor 参照（任意、存在しないアクターへの参照は警告）
// - Activity → UseCase 参照（任意、存在しないユースケースへの参照は警告）
// - Activity → Activity (Dependencies) 参照（任意、存在しないアクティビティへの参照は警告）
// - Activity → Activity (ParentID) 参照（任意、存在しない親への参照は警告）
// - Activity → Deliverable (RelatedDeliverables) 参照（推奨、存在しない成果物への参照は警告）
// - Activity.Node → Deliverable (DeliverableIDs) 参照（任意、存在しない成果物への参照は警告）
func (c *IntegrityChecker) CheckWarnings(ctx context.Context) ([]*ReferenceWarning, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var warnings []*ReferenceWarning

	// UseCase → Subsystem 参照チェック
	usecaseSubsystemWarnings, err := c.checkUseCaseSubsystemReferences(ctx)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, usecaseSubsystemWarnings...)

	// UseCase → Actor 参照チェック
	usecaseActorWarnings, err := c.checkUseCaseActorReferences(ctx)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, usecaseActorWarnings...)

	// Activity 関連チェック（一度だけ取得して再利用）
	activityWarnings, err := c.checkAllActivityReferences(ctx)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, activityWarnings...)

	return warnings, nil
}

// checkAllActivityReferences は Activity 関連の全ての参照チェックを一度の GetAll で実行
// パフォーマンス最適化: Activity を一度だけ取得し、複数のチェックで再利用
func (c *IntegrityChecker) checkAllActivityReferences(ctx context.Context) ([]*ReferenceWarning, error) {
	if c.activityHandler == nil {
		return []*ReferenceWarning{}, nil
	}

	// Activity を一度だけ取得
	activities, err := c.activityHandler.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Activity ID のセットを作成（Dependencies/ParentID チェック用）
	activityIDs := make(map[string]bool)
	for _, act := range activities {
		activityIDs[act.ID] = true
	}

	var warnings []*ReferenceWarning

	// Activity → UseCase 参照チェック
	ucWarnings := c.checkActivityUseCaseReferencesWithData(ctx, activities)
	warnings = append(warnings, ucWarnings...)

	// Activity → Activity (Dependencies) 参照チェック
	depWarnings := c.checkActivityDependencyReferencesWithData(activities, activityIDs)
	warnings = append(warnings, depWarnings...)

	// Activity → Activity (ParentID) 参照チェック
	parentWarnings := c.checkActivityParentReferencesWithData(activities, activityIDs)
	warnings = append(warnings, parentWarnings...)

	// Activity → Deliverable 参照チェック
	delWarnings := c.checkActivityDeliverableReferencesWithData(ctx, activities)
	warnings = append(warnings, delWarnings...)

	return warnings, nil
}

// checkUseCaseSubsystemReferences は UseCase から Subsystem への参照をチェック（警告レベル）
// SubsystemID が設定されているが、該当の Subsystem が存在しない場合は警告を出す
// 無効な ID 形式（ValidationError）も警告として扱う
func (c *IntegrityChecker) checkUseCaseSubsystemReferences(ctx context.Context) ([]*ReferenceWarning, error) {
	if c.usecaseHandler == nil {
		return []*ReferenceWarning{}, nil
	}

	usecases, err := c.usecaseHandler.getAllUseCases(ctx)
	if err != nil {
		return nil, err
	}

	var warnings []*ReferenceWarning
	for _, uc := range usecases {
		// SubsystemID が未設定なら OK（任意フィールド）
		if uc.SubsystemID == "" {
			continue
		}

		// SubsystemHandler が未設定なら警告チェックをスキップ
		if c.subsystemHandler == nil {
			continue
		}

		// Subsystem の存在確認
		_, err := c.subsystemHandler.Get(ctx, uc.SubsystemID)
		if err == ErrEntityNotFound {
			// サブシステムが存在しない
			warnings = append(warnings, &ReferenceWarning{
				SourceType: "usecase",
				SourceID:   uc.ID,
				TargetType: "subsystem",
				TargetID:   uc.SubsystemID,
				Message:    ErrMsgReferencedSubsystemNotFound,
			})
		} else if err != nil {
			// ValidationError（無効な ID 形式）も警告として扱う
			var validationErr *ValidationError
			if errors.As(err, &validationErr) {
				warnings = append(warnings, &ReferenceWarning{
					SourceType: "usecase",
					SourceID:   uc.ID,
					TargetType: "subsystem",
					TargetID:   uc.SubsystemID,
					Message:    ErrMsgInvalidSubsystemIDFormat,
				})
			} else {
				// その他のエラーはそのまま返す
				return nil, err
			}
		}
	}

	return warnings, nil
}

// checkUseCaseActorReferences は UseCase から Actor への参照をチェック（警告レベル）
func (c *IntegrityChecker) checkUseCaseActorReferences(ctx context.Context) ([]*ReferenceWarning, error) {
	if c.usecaseHandler == nil {
		return []*ReferenceWarning{}, nil
	}

	usecases, err := c.usecaseHandler.getAllUseCases(ctx)
	if err != nil {
		return nil, err
	}

	var warnings []*ReferenceWarning
	for _, uc := range usecases {
		// ActorHandler が未設定なら警告チェックをスキップ
		if c.actorHandler == nil {
			continue
		}

		for _, actorRef := range uc.Actors {
			if actorRef.ActorID == "" {
				continue
			}

			// Actor の存在確認
			_, err := c.actorHandler.Get(ctx, actorRef.ActorID)
			if err == ErrEntityNotFound {
				warnings = append(warnings, &ReferenceWarning{
					SourceType: "usecase",
					SourceID:   uc.ID,
					TargetType: "actor",
					TargetID:   actorRef.ActorID,
					Message:    ErrMsgReferencedActorNotFound,
				})
			} else if err != nil {
				var validationErr *ValidationError
				if errors.As(err, &validationErr) {
					warnings = append(warnings, &ReferenceWarning{
						SourceType: "usecase",
						SourceID:   uc.ID,
						TargetType: "actor",
						TargetID:   actorRef.ActorID,
						Message:    ErrMsgInvalidActorIDFormat,
					})
				} else {
					return nil, err
				}
			}
		}
	}

	return warnings, nil
}

// ===== 最適化されたヘルパーメソッド（事前取得データを使用） =====

// checkActivityUseCaseReferencesWithData は Activity から UseCase への参照をチェック（事前取得データ使用）
func (c *IntegrityChecker) checkActivityUseCaseReferencesWithData(ctx context.Context, activities []ActivityEntity) []*ReferenceWarning {
	var warnings []*ReferenceWarning
	for _, act := range activities {
		// UseCaseID が未設定なら OK（任意フィールド）
		if act.UseCaseID == "" {
			continue
		}

		// UseCaseHandler が未設定なら警告チェックをスキップ
		if c.usecaseHandler == nil {
			continue
		}

		// UseCase の存在確認
		_, err := c.usecaseHandler.Get(ctx, act.UseCaseID)
		if err == ErrEntityNotFound {
			warnings = append(warnings, &ReferenceWarning{
				SourceType: "activity",
				SourceID:   act.ID,
				TargetType: "usecase",
				TargetID:   act.UseCaseID,
				Message:    ErrMsgReferencedUseCaseNotFound,
			})
		} else if err != nil {
			var validationErr *ValidationError
			if errors.As(err, &validationErr) {
				warnings = append(warnings, &ReferenceWarning{
					SourceType: "activity",
					SourceID:   act.ID,
					TargetType: "usecase",
					TargetID:   act.UseCaseID,
					Message:    ErrMsgInvalidUseCaseIDFormat,
				})
			}
		}
	}
	return warnings
}

// checkActivityDependencyReferencesWithData は Activity から Activity (Dependencies) への参照をチェック（事前取得データ使用）
func (c *IntegrityChecker) checkActivityDependencyReferencesWithData(activities []ActivityEntity, activityIDs map[string]bool) []*ReferenceWarning {
	var warnings []*ReferenceWarning
	for _, act := range activities {
		for _, depID := range act.Dependencies {
			if depID == "" {
				continue
			}

			// 依存先 Activity の存在確認
			if !activityIDs[depID] {
				warnings = append(warnings, &ReferenceWarning{
					SourceType: "activity",
					SourceID:   act.ID,
					TargetType: "activity",
					TargetID:   depID,
					Message:    ErrMsgReferencedActivityNotFound,
				})
			}
		}
	}
	return warnings
}

// checkActivityParentReferencesWithData は Activity から Activity (ParentID) への参照をチェック（事前取得データ使用）
func (c *IntegrityChecker) checkActivityParentReferencesWithData(activities []ActivityEntity, activityIDs map[string]bool) []*ReferenceWarning {
	var warnings []*ReferenceWarning
	for _, act := range activities {
		// ParentID が未設定なら OK（任意フィールド）
		if act.ParentID == "" {
			continue
		}

		// 親 Activity の存在確認
		if !activityIDs[act.ParentID] {
			warnings = append(warnings, &ReferenceWarning{
				SourceType: "activity",
				SourceID:   act.ID,
				TargetType: "activity",
				TargetID:   act.ParentID,
				Message:    ErrMsgReferencedParentActivityNotFound,
			})
		}
	}
	return warnings
}

// checkActivityDeliverableReferencesWithData は Activity から Deliverable への参照をチェック（事前取得データ使用）
func (c *IntegrityChecker) checkActivityDeliverableReferencesWithData(ctx context.Context, activities []ActivityEntity) []*ReferenceWarning {
	// DeliverableHandler が未設定なら警告チェックをスキップ
	if c.deliverableHandler == nil {
		return []*ReferenceWarning{}
	}

	var warnings []*ReferenceWarning
	for _, act := range activities {
		// RelatedDeliverables のチェック
		for _, delID := range act.RelatedDeliverables {
			if delID == "" {
				continue
			}

			_, err := c.deliverableHandler.Get(ctx, delID)
			if err == ErrEntityNotFound {
				warnings = append(warnings, &ReferenceWarning{
					SourceType: "activity",
					SourceID:   act.ID,
					TargetType: "deliverable",
					TargetID:   delID,
					Message:    ErrMsgRelatedDeliverableNotFound,
				})
			} else if err != nil {
				var validationErr *ValidationError
				if errors.As(err, &validationErr) {
					warnings = append(warnings, &ReferenceWarning{
						SourceType: "activity",
						SourceID:   act.ID,
						TargetType: "deliverable",
						TargetID:   delID,
						Message:    ErrMsgInvalidRelatedDeliverableFormat,
					})
				}
			}
		}

		// Node.DeliverableIDs のチェック
		for _, node := range act.Nodes {
			for _, delID := range node.DeliverableIDs {
				if delID == "" {
					continue
				}

				_, err := c.deliverableHandler.Get(ctx, delID)
				if err == ErrEntityNotFound {
					warnings = append(warnings, &ReferenceWarning{
						SourceType: "activity",
						SourceID:   act.ID,
						TargetType: "deliverable",
						TargetID:   delID,
						Message:    fmt.Sprintf("referenced deliverable in node %s not found", node.ID),
					})
				} else if err != nil {
					var validationErr *ValidationError
					if errors.As(err, &validationErr) {
						warnings = append(warnings, &ReferenceWarning{
							SourceType: "activity",
							SourceID:   act.ID,
							TargetType: "deliverable",
							TargetID:   delID,
							Message:    fmt.Sprintf("invalid deliverable ID format in node %s", node.ID),
						})
					}
				}
			}
		}
	}
	return warnings
}
