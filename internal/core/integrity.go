package core

import (
	"context"
	"fmt"
)

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
}

// CheckAll は全ての整合性チェックを実行
func (c *IntegrityChecker) CheckAll(ctx context.Context) (*IntegrityResult, error) {
	result := &IntegrityResult{
		Valid:           true,
		ReferenceErrors: []*ReferenceError{},
		CycleErrors:     []*CycleError{},
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

	// エラーがあれば Valid = false
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
				Message:    "referenced objective not found",
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
				Message:    "parent objective not found",
			})
		} else if err != nil {
			return nil, err
		}
	}

	return errors, nil
}

// CheckCycles は循環参照をチェック
// - Objective 階層の循環参照を検出
func (c *IntegrityChecker) CheckCycles(ctx context.Context) ([]*CycleError, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if c.objectiveHandler == nil {
		return []*CycleError{}, nil
	}

	objectives, err := c.objectiveHandler.getAllObjectives(ctx)
	if err != nil {
		return nil, err
	}

	// ID → Objective のマップを作成
	objMap := make(map[string]*ObjectiveEntity)
	for _, obj := range objectives {
		objMap[obj.ID] = obj
	}

	var errors []*CycleError
	visited := make(map[string]bool)

	for _, obj := range objectives {
		if visited[obj.ID] {
			continue
		}

		// このノードから親をたどって循環を検出
		cycle := c.detectCycle(obj.ID, objMap, visited)
		if len(cycle) > 0 {
			errors = append(errors, &CycleError{
				EntityType: "objective",
				Cycle:      cycle,
				Message:    "circular parent reference detected",
			})
		}
	}

	return errors, nil
}

// detectCycle は指定されたノードから循環を検出
func (c *IntegrityChecker) detectCycle(startID string, objMap map[string]*ObjectiveEntity, globalVisited map[string]bool) []string {
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
		obj, exists := objMap[current]
		if !exists || obj.ParentID == "" {
			break
		}
		current = obj.ParentID
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
				Message:    "consideration_id is required but missing",
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
					Message:    "referenced consideration not found",
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
				Message:    "deliverable_id is required but missing",
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
					Message:    "referenced deliverable not found",
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
					Message:    "referenced objective not found",
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
					Message:    "referenced deliverable not found",
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
					Message:    "referenced decision not found",
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
					Message:    "referenced objective not found",
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
					Message:    "referenced deliverable not found",
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
					Message:    "referenced objective not found",
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
					Message:    "referenced deliverable not found",
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
					Message:    "referenced objective not found",
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
					Message:    "referenced deliverable not found",
				})
			} else if err != nil {
				return nil, err
			}
		}
	}

	return errors, nil
}
