package core

import (
	"context"
	"errors"
	"fmt"
)

// 整合性チェック用エラーメッセージ定数
const (
	// 参照エラーメッセージ
	ErrMsgReferencedObjectiveNotFound     = "referenced objective not found"
	ErrMsgReferencedConsiderationNotFound = "referenced consideration not found"
	ErrMsgReferencedDecisionNotFound      = "referenced decision not found"
	ErrMsgReferencedSubsystemNotFound     = "referenced subsystem not found"
	ErrMsgReferencedActorNotFound         = "referenced actor not found"
	ErrMsgReferencedUseCaseNotFound       = "referenced usecase not found"
	// 必須フィールド欠損メッセージ
	ErrMsgObjectiveIDRequired     = "objective_id is required but missing"
	ErrMsgConsiderationIDRequired = "consideration_id is required but missing"

	// 無効なID形式メッセージ
	ErrMsgInvalidSubsystemIDFormat = "invalid subsystem ID format"
	ErrMsgInvalidActorIDFormat     = "invalid actor ID format"
	ErrMsgInvalidUseCaseIDFormat   = "invalid usecase ID format"
)

// ParentResolver は親 ID を解決する関数型（汎用循環参照検出用）
type ParentResolver func(id string) string

// IntegrityChecker は参照整合性をチェックする
type IntegrityChecker struct {
	objectiveHandler     *ObjectiveHandler
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
func NewIntegrityChecker(objHandler *ObjectiveHandler) *IntegrityChecker {
	return &IntegrityChecker{
		objectiveHandler: objHandler,
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
	SourceType string // エンティティ種別（"objective", "quality", "usecase" 等）
	SourceID   string
	TargetType string // 参照先エンティティ種別（"objective", "consideration" 等）
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
// - Decision → Consideration 参照（必須）
// - Quality → Objective 参照（必須）
// - UseCase → Objective 参照（必須）
// - Consideration → Objective 参照（任意）
// - Problem → Objective 参照（任意）
// - Risk → Objective 参照（任意）
// - Assumption → Objective 参照（任意）
// - Consideration ← Decision 逆参照（削除時チェック用）
func (c *IntegrityChecker) CheckReferences(ctx context.Context) ([]*ReferenceError, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var errors []*ReferenceError

	// Decision → Consideration 参照チェック（必須）
	decErrors, err := c.checkDecisionReferences(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, decErrors...)

	// Quality → Objective 参照チェック（必須）
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

	// Consideration → Objective 参照チェック
	conErrors, err := c.checkConsiderationReferences(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, conErrors...)

	// Problem → Objective 参照チェック
	probErrors, err := c.checkProblemReferences(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, probErrors...)

	// Risk → Objective 参照チェック
	riskErrors, err := c.checkRiskReferences(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, riskErrors...)

	// Assumption → Objective 参照チェック
	assumErrors, err := c.checkAssumptionReferences(ctx)
	if err != nil {
		return nil, err
	}
	errors = append(errors, assumErrors...)

	return errors, nil
}

// CheckCycles は循環参照をチェック
func (c *IntegrityChecker) CheckCycles(ctx context.Context) ([]*CycleError, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Objective の親子関係は廃止済みのため、現在チェック対象なし
	return []*CycleError{}, nil
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

// checkQualityReferences は Quality から Objective への参照をチェック（必須）
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
		// ObjectiveID は必須
		if qual.ObjectiveID == "" {
			errors = append(errors, &ReferenceError{
				SourceType: "quality",
				SourceID:   qual.ID,
				TargetType: "objective",
				TargetID:   "",
				Message:    ErrMsgObjectiveIDRequired,
			})
			continue
		}

		// Objective の存在確認
		if c.objectiveHandler != nil {
			_, err := c.objectiveHandler.Get(ctx, qual.ObjectiveID)
			if err == ErrEntityNotFound {
				errors = append(errors, &ReferenceError{
					SourceType: "quality",
					SourceID:   qual.ID,
					TargetType: "objective",
					TargetID:   qual.ObjectiveID,
					Message:    ErrMsgReferencedObjectiveNotFound,
				})
			} else if err != nil {
				return nil, err
			}
		}
	}

	return errors, nil
}

// checkConsiderationReferences は Consideration から Objective への参照をチェック
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

// checkProblemReferences は Problem から Objective への参照をチェック
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
	}

	return errors, nil
}

// checkRiskReferences は Risk から Objective への参照をチェック
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
	}

	return errors, nil
}

// checkAssumptionReferences は Assumption から Objective への参照をチェック
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

	var warnings []*ReferenceWarning

	// Activity → UseCase 参照チェック
	ucWarnings := c.checkActivityUseCaseReferencesWithData(ctx, activities)
	warnings = append(warnings, ucWarnings...)

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
