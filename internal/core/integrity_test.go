package core

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// テスト用のセットアップ（IntegrityChecker）
func setupIntegrityCheckerTest(t *testing.T) (*IntegrityChecker, *ObjectiveHandler, *DeliverableHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-integrity-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath+"/objectives", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create objectives dir: %v", err)
	}
	if err := os.MkdirAll(zeusPath+"/deliverables", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create deliverables dir: %v", err)
	}

	fs := yaml.NewFileManager(zeusPath)
	objHandler := NewObjectiveHandler(fs, nil)
	delHandler := NewDeliverableHandler(fs, objHandler, nil)
	checker := NewIntegrityChecker(objHandler, delHandler)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return checker, objHandler, delHandler, zeusPath, cleanup
}

func TestIntegrityCheckerCheckAllClean(t *testing.T) {
	checker, objHandler, delHandler, _, cleanup := setupIntegrityCheckerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 正しい参照を持つデータを作成
	objResult, err := objHandler.Add(ctx, "親 Objective")
	if err != nil {
		t.Fatalf("Add objective failed: %v", err)
	}

	_, err = delHandler.Add(ctx, "Deliverable",
		WithDeliverableObjective(objResult.ID),
	)
	if err != nil {
		t.Fatalf("Add deliverable failed: %v", err)
	}

	// チェック実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	if !result.Valid {
		t.Error("expected Valid to be true")
	}

	if len(result.ReferenceErrors) != 0 {
		t.Errorf("expected 0 reference errors, got %d", len(result.ReferenceErrors))
	}

	if len(result.CycleErrors) != 0 {
		t.Errorf("expected 0 cycle errors, got %d", len(result.CycleErrors))
	}
}

func TestIntegrityCheckerCheckAllEmpty(t *testing.T) {
	checker, _, _, _, cleanup := setupIntegrityCheckerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 空の状態でチェック
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	if !result.Valid {
		t.Error("expected Valid to be true for empty data")
	}
}

func TestIntegrityCheckerDeliverableToObjectiveReference(t *testing.T) {
	checker, objHandler, delHandler, zeusPath, cleanup := setupIntegrityCheckerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Objective を作成
	objResult, err := objHandler.Add(ctx, "Objective")
	if err != nil {
		t.Fatalf("Add objective failed: %v", err)
	}

	// Deliverable を作成（正しい参照）
	_, err = delHandler.Add(ctx, "Valid Deliverable",
		WithDeliverableObjective(objResult.ID),
	)
	if err != nil {
		t.Fatalf("Add valid deliverable failed: %v", err)
	}

	// 壊れた参照を持つ Deliverable を直接作成
	brokenDel := &DeliverableEntity{
		ID:          "del-999",
		Title:       "Broken Deliverable",
		ObjectiveID: "obj-999", // 存在しない Objective
		Status:      DeliverableStatusPlanned,
		Format:      DeliverableFormatOther,
		Metadata:    Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	fs := yaml.NewFileManager(zeusPath)
	if err := fs.WriteYaml(ctx, "deliverables/del-999.yaml", brokenDel); err != nil {
		t.Fatalf("Write broken deliverable failed: %v", err)
	}

	// チェック実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	if result.Valid {
		t.Error("expected Valid to be false with broken reference")
	}

	if len(result.ReferenceErrors) != 1 {
		t.Errorf("expected 1 reference error, got %d", len(result.ReferenceErrors))
	}

	// エラー内容確認
	if len(result.ReferenceErrors) > 0 {
		refErr := result.ReferenceErrors[0]
		if refErr.SourceType != "deliverable" {
			t.Errorf("expected source type 'deliverable', got %q", refErr.SourceType)
		}
		if refErr.SourceID != "del-999" {
			t.Errorf("expected source ID 'del-999', got %q", refErr.SourceID)
		}
		if refErr.TargetType != "objective" {
			t.Errorf("expected target type 'objective', got %q", refErr.TargetType)
		}
		if refErr.TargetID != "obj-999" {
			t.Errorf("expected target ID 'obj-999', got %q", refErr.TargetID)
		}
	}
}

func TestIntegrityCheckerObjectiveParentReference(t *testing.T) {
	checker, objHandler, _, zeusPath, cleanup := setupIntegrityCheckerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 親 Objective を作成
	parentResult, err := objHandler.Add(ctx, "親 Objective")
	if err != nil {
		t.Fatalf("Add parent failed: %v", err)
	}

	// 子 Objective を作成（正しい参照）
	_, err = objHandler.Add(ctx, "子 Objective",
		WithObjectiveParent(parentResult.ID),
	)
	if err != nil {
		t.Fatalf("Add child failed: %v", err)
	}

	// 壊れた親参照を持つ Objective を直接作成
	brokenObj := &ObjectiveEntity{
		ID:       "obj-999",
		Title:    "Broken Objective",
		ParentID: "obj-888", // 存在しない親
		Status:   ObjectiveStatusNotStarted,
		Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	fs := yaml.NewFileManager(zeusPath)
	if err := fs.WriteYaml(ctx, "objectives/obj-999.yaml", brokenObj); err != nil {
		t.Fatalf("Write broken objective failed: %v", err)
	}

	// チェック実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	if result.Valid {
		t.Error("expected Valid to be false with broken parent reference")
	}

	if len(result.ReferenceErrors) != 1 {
		t.Errorf("expected 1 reference error, got %d", len(result.ReferenceErrors))
	}

	// エラー内容確認
	if len(result.ReferenceErrors) > 0 {
		refErr := result.ReferenceErrors[0]
		if refErr.SourceType != "objective" {
			t.Errorf("expected source type 'objective', got %q", refErr.SourceType)
		}
		if refErr.SourceID != "obj-999" {
			t.Errorf("expected source ID 'obj-999', got %q", refErr.SourceID)
		}
		if refErr.TargetType != "objective" {
			t.Errorf("expected target type 'objective', got %q", refErr.TargetType)
		}
		if refErr.TargetID != "obj-888" {
			t.Errorf("expected target ID 'obj-888', got %q", refErr.TargetID)
		}
	}
}

func TestIntegrityCheckerCycleDetection(t *testing.T) {
	checker, _, _, zeusPath, cleanup := setupIntegrityCheckerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 循環参照を持つ Objective を直接作成
	// obj-001 → obj-002 → obj-003 → obj-001 の循環
	fs := yaml.NewFileManager(zeusPath)

	obj1 := &ObjectiveEntity{
		ID:       "obj-001",
		Title:    "Objective 1",
		ParentID: "obj-003", // obj-003 を親とする
		Status:   ObjectiveStatusNotStarted,
		Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	obj2 := &ObjectiveEntity{
		ID:       "obj-002",
		Title:    "Objective 2",
		ParentID: "obj-001", // obj-001 を親とする
		Status:   ObjectiveStatusNotStarted,
		Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	obj3 := &ObjectiveEntity{
		ID:       "obj-003",
		Title:    "Objective 3",
		ParentID: "obj-002", // obj-002 を親とする
		Status:   ObjectiveStatusNotStarted,
		Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}

	if err := fs.WriteYaml(ctx, "objectives/obj-001.yaml", obj1); err != nil {
		t.Fatalf("Write obj-001 failed: %v", err)
	}
	if err := fs.WriteYaml(ctx, "objectives/obj-002.yaml", obj2); err != nil {
		t.Fatalf("Write obj-002 failed: %v", err)
	}
	if err := fs.WriteYaml(ctx, "objectives/obj-003.yaml", obj3); err != nil {
		t.Fatalf("Write obj-003 failed: %v", err)
	}

	// チェック実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	if result.Valid {
		t.Error("expected Valid to be false with cycle")
	}

	if len(result.CycleErrors) == 0 {
		t.Error("expected at least 1 cycle error")
	}

	// エラー内容確認
	if len(result.CycleErrors) > 0 {
		cycleErr := result.CycleErrors[0]
		if cycleErr.EntityType != "objective" {
			t.Errorf("expected entity type 'objective', got %q", cycleErr.EntityType)
		}
		// 循環パスに全てのノードが含まれていることを確認
		if len(cycleErr.Cycle) < 3 {
			t.Errorf("expected cycle length >= 3, got %d", len(cycleErr.Cycle))
		}
	}
}

func TestIntegrityCheckerSelfReference(t *testing.T) {
	checker, _, _, zeusPath, cleanup := setupIntegrityCheckerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 自己参照を持つ Objective を直接作成
	fs := yaml.NewFileManager(zeusPath)

	selfRefObj := &ObjectiveEntity{
		ID:       "obj-001",
		Title:    "Self Reference Objective",
		ParentID: "obj-001", // 自分自身を親とする
		Status:   ObjectiveStatusNotStarted,
		Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}

	if err := fs.WriteYaml(ctx, "objectives/obj-001.yaml", selfRefObj); err != nil {
		t.Fatalf("Write self-ref objective failed: %v", err)
	}

	// チェック実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	if result.Valid {
		t.Error("expected Valid to be false with self-reference")
	}

	if len(result.CycleErrors) == 0 {
		t.Error("expected at least 1 cycle error for self-reference")
	}
}

func TestIntegrityCheckerContextCancellation(t *testing.T) {
	checker, _, _, _, cleanup := setupIntegrityCheckerTest(t)
	defer cleanup()

	// キャンセル済みのコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// CheckAll（エラーがラップされる可能性があるため errors.Is を使用）
	_, err := checker.CheckAll(ctx)
	if err == nil {
		t.Error("CheckAll should fail with cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// CheckReferences
	_, err = checker.CheckReferences(ctx)
	if err == nil {
		t.Error("CheckReferences should fail with cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// CheckCycles
	_, err = checker.CheckCycles(ctx)
	if err == nil {
		t.Error("CheckCycles should fail with cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestIntegrityCheckerNilHandlers(t *testing.T) {
	// nil ハンドラーでのチェック
	checker := NewIntegrityChecker(nil, nil)

	ctx := context.Background()

	// CheckAll は nil ハンドラーでもエラーなく動作すべき
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll with nil handlers failed: %v", err)
	}

	if !result.Valid {
		t.Error("expected Valid to be true with nil handlers")
	}
}

func TestIntegrityCheckerMultipleErrors(t *testing.T) {
	checker, _, _, zeusPath, cleanup := setupIntegrityCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 複数の壊れた参照を作成
	for i := 1; i <= 3; i++ {
		del := &DeliverableEntity{
			ID:          "del-00" + string(rune('0'+i)),
			Title:       "Broken Deliverable",
			ObjectiveID: "obj-999", // 存在しない
			Status:      DeliverableStatusPlanned,
			Format:      DeliverableFormatOther,
			Metadata:    Metadata{CreatedAt: Now(), UpdatedAt: Now()},
		}
		if err := fs.WriteYaml(ctx, "deliverables/del-00"+string(rune('0'+i))+".yaml", del); err != nil {
			t.Fatalf("Write del failed: %v", err)
		}
	}

	// チェック実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	if result.Valid {
		t.Error("expected Valid to be false")
	}

	if len(result.ReferenceErrors) != 3 {
		t.Errorf("expected 3 reference errors, got %d", len(result.ReferenceErrors))
	}
}

func TestIntegrityCheckerNoParentReference(t *testing.T) {
	checker, objHandler, delHandler, _, cleanup := setupIntegrityCheckerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 親なしの Objective を作成（parent_id が空）
	objResult, err := objHandler.Add(ctx, "Root Objective")
	if err != nil {
		t.Fatalf("Add root objective failed: %v", err)
	}

	// Objective を参照する Deliverable を作成（objective_id は必須）
	_, err = delHandler.Add(ctx, "Valid Deliverable",
		WithDeliverableObjective(objResult.ID),
	)
	if err != nil {
		t.Fatalf("Add deliverable failed: %v", err)
	}

	// チェック実行（親参照なしの Objective は OK）
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	if !result.Valid {
		t.Error("expected Valid to be true (Objective without parent is allowed)")
	}
}

func TestReferenceErrorMessage(t *testing.T) {
	refErr := &ReferenceError{
		SourceType: "deliverable",
		SourceID:   "del-001",
		TargetType: "objective",
		TargetID:   "obj-999",
		Message:    "referenced objective not found",
	}

	expected := "deliverable del-001 → objective obj-999: referenced objective not found"
	if refErr.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, refErr.Error())
	}
}

func TestCycleErrorMessage(t *testing.T) {
	cycleErr := &CycleError{
		EntityType: "objective",
		Cycle:      []string{"obj-001", "obj-002", "obj-003", "obj-001"},
		Message:    "circular parent reference detected",
	}

	errMsg := cycleErr.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}

	// エラーメッセージに cycle が含まれることを確認
	if !contains(errMsg, "cycle") {
		t.Errorf("expected error message to contain 'cycle', got %q", errMsg)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ===== TASK-016: UseCase → Subsystem 参照チェックのテスト =====

// テスト用セットアップ（UseCase/Subsystem 参照チェック用）
func setupUseCaseSubsystemIntegrityTest(t *testing.T) (*IntegrityChecker, *UseCaseHandler, *SubsystemHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-usecase-subsystem-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath+"/usecases", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create usecases dir: %v", err)
	}

	fs := yaml.NewFileManager(zeusPath)
	subsystemHandler := NewSubsystemHandler(fs)
	usecaseHandler := NewUseCaseHandler(fs, nil, nil, nil)

	// IntegrityChecker に設定
	checker := NewIntegrityChecker(nil, nil)
	checker.SetUseCaseHandler(usecaseHandler)
	checker.SetSubsystemHandler(subsystemHandler)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return checker, usecaseHandler, subsystemHandler, zeusPath, cleanup
}

// TestIntegrityCheckerUseCaseSubsystemWarningClean は正常なサブシステム参照をチェック
func TestIntegrityCheckerUseCaseSubsystemWarningClean(t *testing.T) {
	checker, _, subsystemHandler, zeusPath, cleanup := setupUseCaseSubsystemIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// Subsystem を作成
	subResult, err := subsystemHandler.Add(ctx, "認証サブシステム")
	if err != nil {
		t.Fatalf("Add subsystem failed: %v", err)
	}

	// 正しい参照を持つ UseCase を直接作成（ObjectiveID は必須だが、整合性チェックではスキップされる）
	uc := &UseCaseEntity{
		ID:          "uc-12345678",
		Title:       "ログイン",
		ObjectiveID: "obj-00000001", // ダミー（Objective ハンドラーは nil なのでチェックされない）
		SubsystemID: subResult.ID,
		Status:      UseCaseStatusDraft,
		Metadata:    Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "usecases/uc-12345678.yaml", uc); err != nil {
		t.Fatalf("Write usecase failed: %v", err)
	}

	// チェック実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	if !result.Valid {
		t.Error("expected Valid to be true")
	}

	if len(result.Warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(result.Warnings))
	}
}

// TestIntegrityCheckerUseCaseSubsystemWarningNoSubsystem はサブシステム未設定をチェック
func TestIntegrityCheckerUseCaseSubsystemWarningNoSubsystem(t *testing.T) {
	checker, _, _, zeusPath, cleanup := setupUseCaseSubsystemIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// サブシステムなしの UseCase を直接作成
	uc := &UseCaseEntity{
		ID:          "uc-12345678",
		Title:       "ログイン",
		ObjectiveID: "obj-00000001", // ダミー
		SubsystemID: "",             // サブシステム未設定
		Status:      UseCaseStatusDraft,
		Metadata:    Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "usecases/uc-12345678.yaml", uc); err != nil {
		t.Fatalf("Write usecase failed: %v", err)
	}

	// チェック実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	if !result.Valid {
		t.Error("expected Valid to be true (no subsystem is OK)")
	}

	if len(result.Warnings) != 0 {
		t.Errorf("expected 0 warnings for empty subsystem_id, got %d", len(result.Warnings))
	}
}

// TestIntegrityCheckerUseCaseSubsystemWarningBrokenReference は存在しないサブシステム参照をチェック
func TestIntegrityCheckerUseCaseSubsystemWarningBrokenReference(t *testing.T) {
	checker, _, _, zeusPath, cleanup := setupUseCaseSubsystemIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 存在しないサブシステムへの参照を持つ UseCase を直接作成
	brokenUC := &UseCaseEntity{
		ID:          "uc-12345678",
		Title:       "ログイン",
		SubsystemID: "sub-99999999", // 存在しない Subsystem
		Status:      UseCaseStatusDraft,
		Metadata:    Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "usecases/uc-12345678.yaml", brokenUC); err != nil {
		t.Fatalf("Write broken usecase failed: %v", err)
	}

	// チェック実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	// 警告は Valid に影響しない
	if !result.Valid {
		t.Error("expected Valid to be true (warnings don't affect Valid)")
	}

	if len(result.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(result.Warnings))
	}

	// 警告内容確認
	if len(result.Warnings) > 0 {
		warning := result.Warnings[0]
		if warning.SourceType != "usecase" {
			t.Errorf("expected source type 'usecase', got %q", warning.SourceType)
		}
		if warning.SourceID != "uc-12345678" {
			t.Errorf("expected source ID 'uc-12345678', got %q", warning.SourceID)
		}
		if warning.TargetType != "subsystem" {
			t.Errorf("expected target type 'subsystem', got %q", warning.TargetType)
		}
		if warning.TargetID != "sub-99999999" {
			t.Errorf("expected target ID 'sub-99999999', got %q", warning.TargetID)
		}
	}
}

// TestIntegrityCheckerUseCaseSubsystemWarningMultiple は複数の警告をチェック
func TestIntegrityCheckerUseCaseSubsystemWarningMultiple(t *testing.T) {
	checker, _, _, zeusPath, cleanup := setupUseCaseSubsystemIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 複数の壊れた参照を持つ UseCase を作成
	for i := 1; i <= 3; i++ {
		uc := &UseCaseEntity{
			ID:          "uc-1234567" + string(rune('0'+i)),
			Title:       "UseCase",
			SubsystemID: "sub-99999999", // 存在しない
			Status:      UseCaseStatusDraft,
			Metadata:    Metadata{CreatedAt: Now(), UpdatedAt: Now()},
		}
		if err := fs.WriteYaml(ctx, "usecases/uc-1234567"+string(rune('0'+i))+".yaml", uc); err != nil {
			t.Fatalf("Write uc failed: %v", err)
		}
	}

	// チェック実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	// 警告は Valid に影響しない
	if !result.Valid {
		t.Error("expected Valid to be true")
	}

	if len(result.Warnings) != 3 {
		t.Errorf("expected 3 warnings, got %d", len(result.Warnings))
	}
}

// TestIntegrityCheckerUseCaseSubsystemWarningNilHandler は nil ハンドラーでのチェック
func TestIntegrityCheckerUseCaseSubsystemWarningNilHandler(t *testing.T) {
	checker := NewIntegrityChecker(nil, nil)
	// UseCaseHandler と SubsystemHandler は nil のまま

	ctx := context.Background()

	// CheckWarnings は nil ハンドラーでもエラーなく動作すべき
	warnings, err := checker.CheckWarnings(ctx)
	if err != nil {
		t.Fatalf("CheckWarnings with nil handlers failed: %v", err)
	}

	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings with nil handlers, got %d", len(warnings))
	}
}

// TestReferenceWarningMessage は警告メッセージをチェック
func TestReferenceWarningMessage(t *testing.T) {
	warning := &ReferenceWarning{
		SourceType: "usecase",
		SourceID:   "uc-12345678",
		TargetType: "subsystem",
		TargetID:   "sub-99999999",
		Message:    "referenced subsystem not found",
	}

	expected := "usecase uc-12345678 → subsystem sub-99999999: referenced subsystem not found"
	if warning.Warning() != expected {
		t.Errorf("expected warning message %q, got %q", expected, warning.Warning())
	}
}
