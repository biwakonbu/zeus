package core

import (
	"context"
	"fmt"
	"path/filepath"
)

// LintError は Lint エラー
type LintError struct {
	EntityType string
	EntityID   string
	Field      string
	Message    string
	Expected   string
	Actual     string
}

func (e *LintError) Error() string {
	if e.Expected != "" && e.Actual != "" {
		return fmt.Sprintf("[%s] %s.%s: %s (expected: %s, actual: %s)",
			e.EntityType, e.EntityID, e.Field, e.Message, e.Expected, e.Actual)
	}
	return fmt.Sprintf("[%s] %s.%s: %s", e.EntityType, e.EntityID, e.Field, e.Message)
}

// LintWarning は Lint 警告
type LintWarning struct {
	EntityType string
	EntityID   string
	Field      string
	Message    string
	Suggested  string
	Actual     string
}

func (w *LintWarning) Warning() string {
	if w.Suggested != "" {
		return fmt.Sprintf("[%s] %s.%s: %s (suggested: %s, actual: %s)",
			w.EntityType, w.EntityID, w.Field, w.Message, w.Suggested, w.Actual)
	}
	return fmt.Sprintf("[%s] %s.%s: %s", w.EntityType, w.EntityID, w.Field, w.Message)
}

// LintResult は Lint チェックの結果
type LintResult struct {
	Valid    bool
	Errors   []*LintError
	Warnings []*LintWarning
}

// LintChecker はデータの仕様準拠をチェックする
type LintChecker struct {
	fileStore FileStore
}

// NewLintChecker は新しい LintChecker を作成
func NewLintChecker(fs FileStore) *LintChecker {
	return &LintChecker{
		fileStore: fs,
	}
}

// CheckAll は全ての Lint チェックを実行
func (l *LintChecker) CheckAll(ctx context.Context) (*LintResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &LintResult{
		Valid:    true,
		Errors:   make([]*LintError, 0),
		Warnings: make([]*LintWarning, 0),
	}

	// ID フォーマットチェック
	idErrors, idWarnings := l.CheckIDFormat(ctx)
	result.Errors = append(result.Errors, idErrors...)
	result.Warnings = append(result.Warnings, idWarnings...)

	// status/progress 整合性チェック
	warnings := l.CheckStatusProgressConsistency(ctx)
	result.Warnings = append(result.Warnings, warnings...)

	// エラーがあれば valid = false
	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result, nil
}

// CheckIDFormat は全エンティティの ID 形式をチェック
func (l *LintChecker) CheckIDFormat(ctx context.Context) ([]*LintError, []*LintWarning) {
	var errors []*LintError
	var warnings []*LintWarning

	// ディレクトリベースのエンティティ（個別ファイル）
	// Note: tasks は TaskStore 形式のため singleFileEntities で処理
	directoryEntities := []struct {
		entityType  string
		directory   string
		expectedFmt string
	}{
		{"objective", "objectives", "obj-NNN"},
		{"deliverable", "deliverables", "del-NNN"},
		{"activity", "activities", "act-NNN"},
		{"consideration", "considerations", "con-NNN"},
		{"decision", "decisions", "dec-NNN"},
		{"problem", "problems", "prob-NNN"},
		{"risk", "risks", "risk-NNN"},
		{"assumption", "assumptions", "assum-NNN"},
		{"quality", "quality", "qual-NNN"},
		{"usecase", "usecases", "uc-XXXXXXXX"},
	}

	for _, entity := range directoryEntities {
		errs, warns := l.checkDirectoryEntityIDs(ctx, entity.entityType, entity.directory, entity.expectedFmt)
		errors = append(errors, errs...)
		warnings = append(warnings, warns...)
	}

	// 単一ファイルエンティティ（actors.yaml, subsystems.yaml, tasks/active.yaml, constraints.yaml）
	singleFileEntities := []struct {
		entityType  string
		filePath    string
		expectedFmt string
	}{
		{"actor", "actors.yaml", "actor-XXXXXXXX"},
		{"subsystem", "subsystems.yaml", "sub-XXXXXXXX"},
		{"task", "tasks/active.yaml", "task-XXXXXXXX"},
		{"constraint", "constraints.yaml", "const-NNN"},
	}

	for _, entity := range singleFileEntities {
		errs, warns := l.checkSingleFileEntityIDs(ctx, entity.entityType, entity.filePath, entity.expectedFmt)
		errors = append(errors, errs...)
		warnings = append(warnings, warns...)
	}

	// Vision の単一エンティティチェック（配列でない単一エンティティ）
	if l.fileStore.Exists(ctx, "vision.yaml") {
		errs, warns := l.checkSingleVisionEntityID(ctx)
		errors = append(errors, errs...)
		warnings = append(warnings, warns...)
	}

	return errors, warnings
}

// checkDirectoryEntityIDs はディレクトリ内のエンティティ ID をチェック
func (l *LintChecker) checkDirectoryEntityIDs(ctx context.Context, entityType, directory, expectedFmt string) ([]*LintError, []*LintWarning) {
	var errors []*LintError
	var warnings []*LintWarning

	if !l.fileStore.Exists(ctx, directory) {
		return errors, warnings
	}

	files, err := l.fileStore.ListDir(ctx, directory)
	if err != nil {
		// ListDir エラーは警告として記録
		warnings = append(warnings, &LintWarning{
			EntityType: entityType,
			EntityID:   directory,
			Field:      "directory",
			Message:    fmt.Sprintf("failed to list directory: %v", err),
		})
		return errors, warnings
	}

	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}

		filePath := filepath.Join(directory, file)
		id, err := l.extractEntityID(ctx, entityType, filePath)
		if err != nil {
			continue // 読み込み失敗はスキップ
		}

		if err := ValidateID(entityType, id); err != nil {
			errors = append(errors, &LintError{
				EntityType: entityType,
				EntityID:   id,
				Field:      "id",
				Message:    "ID format mismatch",
				Expected:   expectedFmt,
				Actual:     id,
			})
		}
	}

	return errors, warnings
}

// checkSingleFileEntityIDs は単一ファイル内のエンティティ ID をチェック
func (l *LintChecker) checkSingleFileEntityIDs(ctx context.Context, entityType, filePath, expectedFmt string) ([]*LintError, []*LintWarning) {
	var errors []*LintError
	var warnings []*LintWarning

	if !l.fileStore.Exists(ctx, filePath) {
		return errors, warnings
	}

	ids, err := l.extractEntityIDsFromFile(ctx, entityType, filePath)
	if err != nil {
		warnings = append(warnings, &LintWarning{
			EntityType: entityType,
			EntityID:   filePath,
			Field:      "file",
			Message:    fmt.Sprintf("failed to read file: %v", err),
		})
		return errors, warnings
	}

	for _, id := range ids {
		if err := ValidateID(entityType, id); err != nil {
			errors = append(errors, &LintError{
				EntityType: entityType,
				EntityID:   id,
				Field:      "id",
				Message:    "ID format mismatch",
				Expected:   expectedFmt,
				Actual:     id,
			})
		}
	}

	return errors, warnings
}

// extractEntityID はファイルからエンティティ ID を抽出
func (l *LintChecker) extractEntityID(ctx context.Context, entityType, filePath string) (string, error) {
	switch entityType {
	case "objective":
		var entity ObjectiveEntity
		if err := l.fileStore.ReadYaml(ctx, filePath, &entity); err != nil {
			return "", err
		}
		return entity.ID, nil
	case "deliverable":
		var entity DeliverableEntity
		if err := l.fileStore.ReadYaml(ctx, filePath, &entity); err != nil {
			return "", err
		}
		return entity.ID, nil
	case "activity":
		var entity ActivityEntity
		if err := l.fileStore.ReadYaml(ctx, filePath, &entity); err != nil {
			return "", err
		}
		return entity.ID, nil
	case "consideration":
		var entity ConsiderationEntity
		if err := l.fileStore.ReadYaml(ctx, filePath, &entity); err != nil {
			return "", err
		}
		return entity.ID, nil
	case "decision":
		var entity DecisionEntity
		if err := l.fileStore.ReadYaml(ctx, filePath, &entity); err != nil {
			return "", err
		}
		return entity.ID, nil
	case "problem":
		var entity ProblemEntity
		if err := l.fileStore.ReadYaml(ctx, filePath, &entity); err != nil {
			return "", err
		}
		return entity.ID, nil
	case "risk":
		var entity RiskEntity
		if err := l.fileStore.ReadYaml(ctx, filePath, &entity); err != nil {
			return "", err
		}
		return entity.ID, nil
	case "assumption":
		var entity AssumptionEntity
		if err := l.fileStore.ReadYaml(ctx, filePath, &entity); err != nil {
			return "", err
		}
		return entity.ID, nil
	case "quality":
		var entity QualityEntity
		if err := l.fileStore.ReadYaml(ctx, filePath, &entity); err != nil {
			return "", err
		}
		return entity.ID, nil
	case "task":
		var entity Task
		if err := l.fileStore.ReadYaml(ctx, filePath, &entity); err != nil {
			return "", err
		}
		return entity.ID, nil
	case "usecase":
		var entity UseCaseEntity
		if err := l.fileStore.ReadYaml(ctx, filePath, &entity); err != nil {
			return "", err
		}
		return entity.ID, nil
	default:
		return "", fmt.Errorf("unknown entity type: %s", entityType)
	}
}

// extractEntityIDsFromFile は単一ファイルから複数エンティティの ID を抽出
func (l *LintChecker) extractEntityIDsFromFile(ctx context.Context, entityType, filePath string) ([]string, error) {
	switch entityType {
	case "actor":
		var store ActorsFile
		if err := l.fileStore.ReadYaml(ctx, filePath, &store); err != nil {
			return nil, err
		}
		ids := make([]string, len(store.Actors))
		for i, actor := range store.Actors {
			ids[i] = actor.ID
		}
		return ids, nil
	case "subsystem":
		var store SubsystemsFile
		if err := l.fileStore.ReadYaml(ctx, filePath, &store); err != nil {
			return nil, err
		}
		ids := make([]string, len(store.Subsystems))
		for i, sub := range store.Subsystems {
			ids[i] = sub.ID
		}
		return ids, nil
	case "task":
		var store TaskStore
		if err := l.fileStore.ReadYaml(ctx, filePath, &store); err != nil {
			return nil, err
		}
		ids := make([]string, len(store.Tasks))
		for i, task := range store.Tasks {
			ids[i] = task.ID
		}
		return ids, nil
	case "constraint":
		var store ConstraintsFile
		if err := l.fileStore.ReadYaml(ctx, filePath, &store); err != nil {
			return nil, err
		}
		ids := make([]string, len(store.Constraints))
		for i, c := range store.Constraints {
			ids[i] = c.ID
		}
		return ids, nil
	default:
		return nil, fmt.Errorf("unknown single-file entity type: %s", entityType)
	}
}

// CheckStatusProgressConsistency は status と progress の整合性をチェック
func (l *LintChecker) CheckStatusProgressConsistency(ctx context.Context) []*LintWarning {
	var warnings []*LintWarning

	// Objective: progress=100 なら status=completed
	warnings = append(warnings, l.checkObjectiveStatusProgress(ctx)...)

	// Deliverable: progress=100 なら status=completed
	warnings = append(warnings, l.checkDeliverableStatusProgress(ctx)...)

	// Activity: progress=100 なら status=completed
	warnings = append(warnings, l.checkActivityStatusProgress(ctx)...)

	return warnings
}

// checkObjectiveStatusProgress は Objective の status/progress 整合性をチェック
func (l *LintChecker) checkObjectiveStatusProgress(ctx context.Context) []*LintWarning {
	var warnings []*LintWarning

	if !l.fileStore.Exists(ctx, "objectives") {
		return warnings
	}

	files, err := l.fileStore.ListDir(ctx, "objectives")
	if err != nil {
		return warnings
	}

	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var obj ObjectiveEntity
		if err := l.fileStore.ReadYaml(ctx, filepath.Join("objectives", file), &obj); err != nil {
			continue
		}
		if obj.Progress == 100 && obj.Status != ObjectiveStatusCompleted {
			warnings = append(warnings, &LintWarning{
				EntityType: "objective",
				EntityID:   obj.ID,
				Field:      "status",
				Message:    "progress=100 but status is not completed",
				Suggested:  string(ObjectiveStatusCompleted),
				Actual:     string(obj.Status),
			})
		}
	}

	return warnings
}

// checkDeliverableStatusProgress は Deliverable の status/progress 整合性をチェック
func (l *LintChecker) checkDeliverableStatusProgress(ctx context.Context) []*LintWarning {
	var warnings []*LintWarning

	if !l.fileStore.Exists(ctx, "deliverables") {
		return warnings
	}

	files, err := l.fileStore.ListDir(ctx, "deliverables")
	if err != nil {
		return warnings
	}

	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var del DeliverableEntity
		if err := l.fileStore.ReadYaml(ctx, filepath.Join("deliverables", file), &del); err != nil {
			continue
		}
		if del.Progress == 100 && del.Status != DeliverableStatusCompleted {
			warnings = append(warnings, &LintWarning{
				EntityType: "deliverable",
				EntityID:   del.ID,
				Field:      "status",
				Message:    "progress=100 but status is not completed",
				Suggested:  string(DeliverableStatusCompleted),
				Actual:     string(del.Status),
			})
		}
	}

	return warnings
}

// checkActivityStatusProgress は Activity の status/progress 整合性をチェック
func (l *LintChecker) checkActivityStatusProgress(ctx context.Context) []*LintWarning {
	var warnings []*LintWarning

	if !l.fileStore.Exists(ctx, "activities") {
		return warnings
	}

	files, err := l.fileStore.ListDir(ctx, "activities")
	if err != nil {
		return warnings
	}

	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var act ActivityEntity
		if err := l.fileStore.ReadYaml(ctx, filepath.Join("activities", file), &act); err != nil {
			continue
		}
		if act.Progress == 100 && act.Status != ActivityStatusCompleted {
			warnings = append(warnings, &LintWarning{
				EntityType: "activity",
				EntityID:   act.ID,
				Field:      "status",
				Message:    "progress=100 but status is not completed",
				Suggested:  string(ActivityStatusCompleted),
				Actual:     string(act.Status),
			})
		}
	}

	return warnings
}

// FixStatusProgressConsistency は status/progress の不整合を自動修正
func (l *LintChecker) FixStatusProgressConsistency(ctx context.Context) (int, error) {
	fixedCount := 0

	// Objective: progress=100 なら status=completed に修正
	count, err := l.fixObjectiveStatusProgress(ctx)
	if err != nil {
		return fixedCount, err
	}
	fixedCount += count

	// Deliverable: progress=100 なら status=completed に修正
	count, err = l.fixDeliverableStatusProgress(ctx)
	if err != nil {
		return fixedCount, err
	}
	fixedCount += count

	// Activity: progress=100 なら status=completed に修正
	count, err = l.fixActivityStatusProgress(ctx)
	if err != nil {
		return fixedCount, err
	}
	fixedCount += count

	return fixedCount, nil
}

// fixObjectiveStatusProgress は Objective の status/progress 不整合を修正
func (l *LintChecker) fixObjectiveStatusProgress(ctx context.Context) (int, error) {
	fixedCount := 0

	if !l.fileStore.Exists(ctx, "objectives") {
		return fixedCount, nil
	}

	files, err := l.fileStore.ListDir(ctx, "objectives")
	if err != nil {
		return fixedCount, nil
	}

	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var obj ObjectiveEntity
		filePath := filepath.Join("objectives", file)
		if err := l.fileStore.ReadYaml(ctx, filePath, &obj); err != nil {
			continue
		}
		if obj.Progress == 100 && obj.Status != ObjectiveStatusCompleted {
			obj.Status = ObjectiveStatusCompleted
			obj.Metadata.UpdatedAt = Now()
			if err := l.fileStore.WriteYaml(ctx, filePath, &obj); err != nil {
				return fixedCount, err
			}
			fixedCount++
		}
	}

	return fixedCount, nil
}

// fixDeliverableStatusProgress は Deliverable の status/progress 不整合を修正
func (l *LintChecker) fixDeliverableStatusProgress(ctx context.Context) (int, error) {
	fixedCount := 0

	if !l.fileStore.Exists(ctx, "deliverables") {
		return fixedCount, nil
	}

	files, err := l.fileStore.ListDir(ctx, "deliverables")
	if err != nil {
		return fixedCount, nil
	}

	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var del DeliverableEntity
		filePath := filepath.Join("deliverables", file)
		if err := l.fileStore.ReadYaml(ctx, filePath, &del); err != nil {
			continue
		}
		if del.Progress == 100 && del.Status != DeliverableStatusCompleted {
			del.Status = DeliverableStatusCompleted
			del.Metadata.UpdatedAt = Now()
			if err := l.fileStore.WriteYaml(ctx, filePath, &del); err != nil {
				return fixedCount, err
			}
			fixedCount++
		}
	}

	return fixedCount, nil
}

// fixActivityStatusProgress は Activity の status/progress 不整合を修正
func (l *LintChecker) fixActivityStatusProgress(ctx context.Context) (int, error) {
	fixedCount := 0

	if !l.fileStore.Exists(ctx, "activities") {
		return fixedCount, nil
	}

	files, err := l.fileStore.ListDir(ctx, "activities")
	if err != nil {
		return fixedCount, nil
	}

	for _, file := range files {
		if !hasYamlSuffix(file) {
			continue
		}
		var act ActivityEntity
		filePath := filepath.Join("activities", file)
		if err := l.fileStore.ReadYaml(ctx, filePath, &act); err != nil {
			continue
		}
		if act.Progress == 100 && act.Status != ActivityStatusCompleted {
			act.Status = ActivityStatusCompleted
			act.Metadata.UpdatedAt = Now()
			if err := l.fileStore.WriteYaml(ctx, filePath, &act); err != nil {
				return fixedCount, err
			}
			fixedCount++
		}
	}

	return fixedCount, nil
}

// checkSingleVisionEntityID は Vision の ID フォーマットをチェック
// Vision は単一エンティティ（配列でない）として vision.yaml に保存される
func (l *LintChecker) checkSingleVisionEntityID(ctx context.Context) ([]*LintError, []*LintWarning) {
	var errors []*LintError
	var warnings []*LintWarning

	var vision Vision
	if err := l.fileStore.ReadYaml(ctx, "vision.yaml", &vision); err != nil {
		warnings = append(warnings, &LintWarning{
			EntityType: "vision",
			EntityID:   "vision.yaml",
			Field:      "file",
			Message:    fmt.Sprintf("failed to read file: %v", err),
		})
		return errors, warnings
	}

	if err := ValidateID("vision", vision.ID); err != nil {
		errors = append(errors, &LintError{
			EntityType: "vision",
			EntityID:   vision.ID,
			Field:      "id",
			Message:    "ID format mismatch",
			Expected:   "vision-NNN",
			Actual:     vision.ID,
		})
	}

	return errors, warnings
}
