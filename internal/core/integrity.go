package core

import (
	"context"
	"fmt"
)

// IntegrityChecker は参照整合性をチェックする
type IntegrityChecker struct {
	objectiveHandler   *ObjectiveHandler
	deliverableHandler *DeliverableHandler
}

// NewIntegrityChecker は新しい IntegrityChecker を作成
func NewIntegrityChecker(objHandler *ObjectiveHandler, delHandler *DeliverableHandler) *IntegrityChecker {
	return &IntegrityChecker{
		objectiveHandler:   objHandler,
		deliverableHandler: delHandler,
	}
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
