package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
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
	if !strings.Contains(errMsg, "cycle") {
		t.Errorf("expected error message to contain 'cycle', got %q", errMsg)
	}
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

	// Objective を作成（UseCase の必須参照先）
	obj := &ObjectiveEntity{
		ID:       "obj-00000001",
		Title:    "テスト目標",
		Status:   ObjectiveStatusInProgress,
		Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.EnsureDir(ctx, "objectives"); err != nil {
		t.Fatalf("EnsureDir objectives failed: %v", err)
	}
	if err := fs.WriteYaml(ctx, "objectives/obj-00000001.yaml", obj); err != nil {
		t.Fatalf("Write objective failed: %v", err)
	}

	// 存在しないサブシステムへの参照を持つ UseCase を直接作成
	brokenUC := &UseCaseEntity{
		ID:          "uc-12345678",
		Title:       "ログイン",
		ObjectiveID: "obj-00000001", // 必須参照
		SubsystemID: "sub-99999999", // 存在しない Subsystem
		Status:      UseCaseStatusDraft,
		Metadata:    Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "usecases/uc-12345678.yaml", brokenUC); err != nil {
		t.Fatalf("Write broken usecase failed: %v", err)
	}

	// ObjectiveHandler を設定
	objHandler := NewObjectiveHandler(fs, nil)
	checker.objectiveHandler = objHandler

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

	// Objective を作成（UseCase の必須参照先）
	obj := &ObjectiveEntity{
		ID:       "obj-00000001",
		Title:    "テスト目標",
		Status:   ObjectiveStatusInProgress,
		Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.EnsureDir(ctx, "objectives"); err != nil {
		t.Fatalf("EnsureDir objectives failed: %v", err)
	}
	if err := fs.WriteYaml(ctx, "objectives/obj-00000001.yaml", obj); err != nil {
		t.Fatalf("Write objective failed: %v", err)
	}

	// ObjectiveHandler を設定
	objHandler := NewObjectiveHandler(fs, nil)
	checker.objectiveHandler = objHandler

	// 複数の壊れた参照を持つ UseCase を作成
	for i := 1; i <= 3; i++ {
		uc := &UseCaseEntity{
			ID:          "uc-1234567" + string(rune('0'+i)),
			Title:       "UseCase",
			ObjectiveID: "obj-00000001", // 必須参照
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

// ===== M3: Decision/Consideration 逆参照整合性テスト =====

// テスト用セットアップ（Decision/Consideration 逆参照チェック用）
func setupDecisionConsiderationIntegrityTest(t *testing.T) (*IntegrityChecker, *ConsiderationHandler, *DecisionHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-decision-consideration-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath+"/considerations", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create considerations dir: %v", err)
	}
	if err := os.MkdirAll(zeusPath+"/decisions", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create decisions dir: %v", err)
	}

	fs := yaml.NewFileManager(zeusPath)
	conHandler := NewConsiderationHandler(fs, nil, nil, nil)
	decHandler := NewDecisionHandler(fs, conHandler, nil)

	// IntegrityChecker に設定
	checker := NewIntegrityChecker(nil, nil)
	checker.SetConsiderationHandler(conHandler)
	checker.SetDecisionHandler(decHandler)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return checker, conHandler, decHandler, zeusPath, cleanup
}

// TestIntegrityChecker_DecisionConsiderationBackRef は Decision/Consideration の双方向参照整合性をテスト
func TestIntegrityChecker_DecisionConsiderationBackRef(t *testing.T) {
	checker, conHandler, decHandler, _, cleanup := setupDecisionConsiderationIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Consideration を作成
	conResult, err := conHandler.Add(ctx, "検討事項 1")
	if err != nil {
		t.Fatalf("Add consideration failed: %v", err)
	}

	// 2. Decision を作成（Consideration を参照）
	decResult, err := decHandler.Add(ctx, "決定事項 1",
		WithDecisionConsideration(conResult.ID),
		WithDecisionSelected(SelectedOption{OptionID: "opt-001", Title: "選択肢 1"}),
		WithDecisionRationale("これが最適な選択です"),
	)
	if err != nil {
		t.Fatalf("Add decision failed: %v", err)
	}

	// 3. Consideration が Decision を逆参照していることを確認
	con, err := conHandler.Get(ctx, conResult.ID)
	if err != nil {
		t.Fatalf("Get consideration failed: %v", err)
	}
	conEntity := con.(*ConsiderationEntity)
	if conEntity.DecisionID != decResult.ID {
		t.Errorf("expected consideration.DecisionID = %q, got %q", decResult.ID, conEntity.DecisionID)
	}

	// 4. 整合性チェック実行（正常なデータ）
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	if !result.Valid {
		t.Error("expected Valid to be true for consistent bidirectional references")
	}

	if len(result.ReferenceErrors) != 0 {
		t.Errorf("expected 0 reference errors, got %d", len(result.ReferenceErrors))
	}
}

// TestIntegrityChecker_OrphanedDecision は孤立 Decision（参照先 Consideration が存在しない）を検出するテスト
func TestIntegrityChecker_OrphanedDecision(t *testing.T) {
	checker, _, _, zeusPath, cleanup := setupDecisionConsiderationIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 存在しない Consideration を参照する Decision を直接作成（孤立 Decision）
	orphanedDec := &DecisionEntity{
		ID:              "dec-001",
		Title:           "孤立した決定事項",
		ConsiderationID: "con-999", // 存在しない Consideration
		DecidedAt:       Now(),
	}
	if err := fs.WriteYaml(ctx, "decisions/dec-001.yaml", orphanedDec); err != nil {
		t.Fatalf("Write orphaned decision failed: %v", err)
	}

	// 整合性チェック実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	// 孤立 Decision はエラーとして検出される
	if result.Valid {
		t.Error("expected Valid to be false for orphaned decision")
	}

	if len(result.ReferenceErrors) != 1 {
		t.Errorf("expected 1 reference error, got %d", len(result.ReferenceErrors))
	}

	// エラー内容確認
	if len(result.ReferenceErrors) > 0 {
		refErr := result.ReferenceErrors[0]
		if refErr.SourceType != "decision" {
			t.Errorf("expected source type 'decision', got %q", refErr.SourceType)
		}
		if refErr.SourceID != "dec-001" {
			t.Errorf("expected source ID 'dec-001', got %q", refErr.SourceID)
		}
		if refErr.TargetType != "consideration" {
			t.Errorf("expected target type 'consideration', got %q", refErr.TargetType)
		}
		if refErr.TargetID != "con-999" {
			t.Errorf("expected target ID 'con-999', got %q", refErr.TargetID)
		}
	}
}

// TestIntegrityChecker_DecisionMissingConsiderationID は ConsiderationID が空の Decision を検出するテスト
func TestIntegrityChecker_DecisionMissingConsiderationID(t *testing.T) {
	checker, _, _, zeusPath, cleanup := setupDecisionConsiderationIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// ConsiderationID が空の Decision を直接作成（必須フィールド欠損）
	invalidDec := &DecisionEntity{
		ID:              "dec-001",
		Title:           "不正な決定事項",
		ConsiderationID: "", // 必須なのに空
		DecidedAt:       Now(),
	}
	if err := fs.WriteYaml(ctx, "decisions/dec-001.yaml", invalidDec); err != nil {
		t.Fatalf("Write invalid decision failed: %v", err)
	}

	// 整合性チェック実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	// ConsiderationID 欠損はエラー
	if result.Valid {
		t.Error("expected Valid to be false for decision missing consideration_id")
	}

	if len(result.ReferenceErrors) != 1 {
		t.Errorf("expected 1 reference error, got %d", len(result.ReferenceErrors))
	}

	// エラー内容確認
	if len(result.ReferenceErrors) > 0 {
		refErr := result.ReferenceErrors[0]
		if refErr.Message != "consideration_id is required but missing" {
			t.Errorf("expected message about missing consideration_id, got %q", refErr.Message)
		}
	}
}

// TestIntegrityChecker_ConsiderationDecisionBackRefInconsistent は Consideration の DecisionID が存在しない場合をテスト
func TestIntegrityChecker_ConsiderationDecisionBackRefInconsistent(t *testing.T) {
	checker, _, _, zeusPath, cleanup := setupDecisionConsiderationIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// Consideration を直接作成（存在しない Decision を参照）
	inconsistentCon := &ConsiderationEntity{
		ID:         "con-001",
		Title:      "不整合な検討事項",
		DecisionID: "dec-999", // 存在しない Decision
		Status:     ConsiderationStatusDecided,
		Metadata:   Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "considerations/con-001.yaml", inconsistentCon); err != nil {
		t.Fatalf("Write inconsistent consideration failed: %v", err)
	}

	// 整合性チェック実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	// Consideration → Decision の不整合参照はエラー
	if result.Valid {
		t.Error("expected Valid to be false for inconsistent consideration→decision reference")
	}

	if len(result.ReferenceErrors) != 1 {
		t.Errorf("expected 1 reference error, got %d", len(result.ReferenceErrors))
	}

	// エラー内容確認
	if len(result.ReferenceErrors) > 0 {
		refErr := result.ReferenceErrors[0]
		if refErr.SourceType != "consideration" {
			t.Errorf("expected source type 'consideration', got %q", refErr.SourceType)
		}
		if refErr.TargetType != "decision" {
			t.Errorf("expected target type 'decision', got %q", refErr.TargetType)
		}
		if refErr.TargetID != "dec-999" {
			t.Errorf("expected target ID 'dec-999', got %q", refErr.TargetID)
		}
	}
}

// TestIntegrityChecker_MultipleDecisionReferenceErrors は複数の Decision 参照エラーをテスト
func TestIntegrityChecker_MultipleDecisionReferenceErrors(t *testing.T) {
	checker, _, _, zeusPath, cleanup := setupDecisionConsiderationIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 複数の孤立 Decision を作成
	for i := 1; i <= 3; i++ {
		dec := &DecisionEntity{
			ID:              fmt.Sprintf("dec-00%d", i),
			Title:           "孤立した決定事項",
			ConsiderationID: "con-999", // 存在しない
			DecidedAt:       Now(),
		}
		if err := fs.WriteYaml(ctx, fmt.Sprintf("decisions/dec-00%d.yaml", i), dec); err != nil {
			t.Fatalf("Write decision failed: %v", err)
		}
	}

	// 整合性チェック実行
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

// TestIntegrityChecker_DecisionConsiderationNilHandlers は nil ハンドラーでの動作をテスト
func TestIntegrityChecker_DecisionConsiderationNilHandlers(t *testing.T) {
	// DecisionHandler と ConsiderationHandler は nil のまま
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

// ===== UseCase → Objective 必須参照テスト =====

// テスト用セットアップ（UseCase → Objective 必須参照チェック用）
func setupUseCaseObjectiveIntegrityTest(t *testing.T) (*IntegrityChecker, *UseCaseHandler, *ObjectiveHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-usecase-objective-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath+"/usecases", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create usecases dir: %v", err)
	}
	if err := os.MkdirAll(zeusPath+"/objectives", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create objectives dir: %v", err)
	}

	fs := yaml.NewFileManager(zeusPath)
	objHandler := NewObjectiveHandler(fs, nil)
	usecaseHandler := NewUseCaseHandler(fs, objHandler, nil, nil)

	checker := NewIntegrityChecker(objHandler, nil)
	checker.SetUseCaseHandler(usecaseHandler)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return checker, usecaseHandler, objHandler, zeusPath, cleanup
}

// TestIntegrityChecker_UseCaseObjectiveRequired は UseCase → Objective 必須参照をテスト
func TestIntegrityChecker_UseCaseObjectiveRequired(t *testing.T) {
	checker, _, objHandler, zeusPath, cleanup := setupUseCaseObjectiveIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// Objective を作成
	objResult, err := objHandler.Add(ctx, "テスト目標")
	if err != nil {
		t.Fatalf("Add objective failed: %v", err)
	}

	// 正しい参照を持つ UseCase を作成
	uc := &UseCaseEntity{
		ID:          "uc-12345678",
		Title:       "正常なユースケース",
		ObjectiveID: objResult.ID,
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
		t.Errorf("expected Valid to be true, got errors: %v", result.ReferenceErrors)
	}
}

// TestIntegrityChecker_UseCaseObjectiveMissing は UseCase の ObjectiveID が空の場合をテスト
func TestIntegrityChecker_UseCaseObjectiveMissing(t *testing.T) {
	checker, _, _, zeusPath, cleanup := setupUseCaseObjectiveIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// ObjectiveID が空の UseCase を作成
	uc := &UseCaseEntity{
		ID:          "uc-12345678",
		Title:       "ObjectiveID なしのユースケース",
		ObjectiveID: "", // 必須なのに空
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

	if result.Valid {
		t.Error("expected Valid to be false for missing objective_id")
	}

	if len(result.ReferenceErrors) != 1 {
		t.Errorf("expected 1 reference error, got %d", len(result.ReferenceErrors))
	}

	if len(result.ReferenceErrors) > 0 {
		refErr := result.ReferenceErrors[0]
		if refErr.Message != "objective_id is required but missing" {
			t.Errorf("expected message about missing objective_id, got %q", refErr.Message)
		}
	}
}

// TestIntegrityChecker_UseCaseObjectiveBroken は UseCase が存在しない Objective を参照している場合をテスト
func TestIntegrityChecker_UseCaseObjectiveBroken(t *testing.T) {
	checker, _, _, zeusPath, cleanup := setupUseCaseObjectiveIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 存在しない Objective を参照する UseCase を作成
	uc := &UseCaseEntity{
		ID:          "uc-12345678",
		Title:       "壊れた参照のユースケース",
		ObjectiveID: "obj-99999999", // 存在しない
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

	if result.Valid {
		t.Error("expected Valid to be false for broken objective reference")
	}

	if len(result.ReferenceErrors) != 1 {
		t.Errorf("expected 1 reference error, got %d", len(result.ReferenceErrors))
	}

	if len(result.ReferenceErrors) > 0 {
		refErr := result.ReferenceErrors[0]
		if refErr.SourceType != "usecase" {
			t.Errorf("expected source type 'usecase', got %q", refErr.SourceType)
		}
		if refErr.TargetType != "objective" {
			t.Errorf("expected target type 'objective', got %q", refErr.TargetType)
		}
		if refErr.TargetID != "obj-99999999" {
			t.Errorf("expected target ID 'obj-99999999', got %q", refErr.TargetID)
		}
	}
}

// ===== Activity 参照警告テスト =====

// テスト用セットアップ（Activity 参照チェック用）
func setupActivityIntegrityTest(t *testing.T) (*IntegrityChecker, *ActivityHandler, *DeliverableHandler, *UseCaseHandler, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-activity-integrity-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	for _, dir := range []string{"activities", "deliverables", "usecases", "objectives"} {
		if err := os.MkdirAll(zeusPath+"/"+dir, 0755); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("failed to create %s dir: %v", dir, err)
		}
	}

	fs := yaml.NewFileManager(zeusPath)
	objHandler := NewObjectiveHandler(fs, nil)
	delHandler := NewDeliverableHandler(fs, objHandler, nil)
	ucHandler := NewUseCaseHandler(fs, objHandler, nil, nil)
	actHandler := NewActivityHandler(fs, ucHandler, delHandler, nil)

	checker := NewIntegrityChecker(objHandler, delHandler)
	checker.SetUseCaseHandler(ucHandler)
	checker.SetActivityHandler(actHandler)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return checker, actHandler, delHandler, ucHandler, zeusPath, cleanup
}

// TestIntegrityChecker_ActivityDependencyWarning は存在しない Activity を依存先として参照している場合をテスト
func TestIntegrityChecker_ActivityDependencyWarning(t *testing.T) {
	checker, _, _, _, zeusPath, cleanup := setupActivityIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 存在しない Activity を依存先として参照する Activity を作成
	act := &ActivityEntity{
		ID:           "act-12345678",
		Title:        "依存先が壊れたアクティビティ",
		Dependencies: []string{"act-99999999"}, // 存在しない
		Status:       ActivityStatusDraft,
		Metadata:     Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "activities/act-12345678.yaml", act); err != nil {
		t.Fatalf("Write activity failed: %v", err)
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

	if len(result.Warnings) > 0 {
		warning := result.Warnings[0]
		if warning.SourceType != "activity" {
			t.Errorf("expected source type 'activity', got %q", warning.SourceType)
		}
		if warning.TargetType != "activity" {
			t.Errorf("expected target type 'activity', got %q", warning.TargetType)
		}
		if warning.TargetID != "act-99999999" {
			t.Errorf("expected target ID 'act-99999999', got %q", warning.TargetID)
		}
	}
}

// TestIntegrityChecker_ActivityParentWarning は存在しない親 Activity を参照している場合をテスト
func TestIntegrityChecker_ActivityParentWarning(t *testing.T) {
	checker, _, _, _, zeusPath, cleanup := setupActivityIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 存在しない親 Activity を参照する Activity を作成
	act := &ActivityEntity{
		ID:       "act-12345678",
		Title:    "親が壊れたアクティビティ",
		ParentID: "act-99999999", // 存在しない
		Status:   ActivityStatusDraft,
		Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "activities/act-12345678.yaml", act); err != nil {
		t.Fatalf("Write activity failed: %v", err)
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

	if len(result.Warnings) > 0 {
		warning := result.Warnings[0]
		if warning.Message != "referenced parent activity not found" {
			t.Errorf("expected message about parent not found, got %q", warning.Message)
		}
	}
}

// TestIntegrityChecker_ActivityDeliverableWarning は存在しない Deliverable を参照している場合をテスト
func TestIntegrityChecker_ActivityDeliverableWarning(t *testing.T) {
	checker, _, _, _, zeusPath, cleanup := setupActivityIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 存在しない Deliverable を参照する Activity を作成
	act := &ActivityEntity{
		ID:                  "act-12345678",
		Title:               "成果物参照が壊れたアクティビティ",
		RelatedDeliverables: []string{"del-99999999"}, // 存在しない
		Status:              ActivityStatusDraft,
		Metadata:            Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "activities/act-12345678.yaml", act); err != nil {
		t.Fatalf("Write activity failed: %v", err)
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

	if len(result.Warnings) > 0 {
		warning := result.Warnings[0]
		if warning.TargetType != "deliverable" {
			t.Errorf("expected target type 'deliverable', got %q", warning.TargetType)
		}
		if warning.TargetID != "del-99999999" {
			t.Errorf("expected target ID 'del-99999999', got %q", warning.TargetID)
		}
	}
}

// TestIntegrityChecker_ActivityUseCaseWarning は存在しない UseCase を参照している場合をテスト
func TestIntegrityChecker_ActivityUseCaseWarning(t *testing.T) {
	checker, _, _, _, zeusPath, cleanup := setupActivityIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 存在しない UseCase を参照する Activity を作成
	act := &ActivityEntity{
		ID:        "act-12345678",
		Title:     "ユースケース参照が壊れたアクティビティ",
		UseCaseID: "uc-99999999", // 存在しない
		Status:    ActivityStatusDraft,
		Metadata:  Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "activities/act-12345678.yaml", act); err != nil {
		t.Fatalf("Write activity failed: %v", err)
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

	if len(result.Warnings) > 0 {
		warning := result.Warnings[0]
		if warning.TargetType != "usecase" {
			t.Errorf("expected target type 'usecase', got %q", warning.TargetType)
		}
	}
}

// TestIntegrityChecker_ActivityNodeDeliverableWarning は Node 内の存在しない Deliverable を参照している場合をテスト
func TestIntegrityChecker_ActivityNodeDeliverableWarning(t *testing.T) {
	checker, _, _, _, zeusPath, cleanup := setupActivityIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// Node 内で存在しない Deliverable を参照する Activity を作成
	act := &ActivityEntity{
		ID:    "act-12345678",
		Title: "ノード内成果物参照が壊れたアクティビティ",
		Nodes: []ActivityNode{
			{
				ID:             "node-001",
				Name:           "テストノード",
				DeliverableIDs: []string{"del-99999999"}, // 存在しない
			},
		},
		Status:   ActivityStatusDraft,
		Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "activities/act-12345678.yaml", act); err != nil {
		t.Fatalf("Write activity failed: %v", err)
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

	if len(result.Warnings) > 0 {
		warning := result.Warnings[0]
		if warning.TargetType != "deliverable" {
			t.Errorf("expected target type 'deliverable', got %q", warning.TargetType)
		}
	}
}

// ===== Activity 循環参照テスト =====

// TestIntegrityChecker_ActivityParentCycle は Activity 親子関係の循環参照をテスト
func TestIntegrityChecker_ActivityParentCycle(t *testing.T) {
	checker, _, _, _, zeusPath, cleanup := setupActivityIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 循環参照を持つ Activity を作成: act-001 → act-002 → act-003 → act-001
	acts := []*ActivityEntity{
		{
			ID:       "act-00000001",
			Title:    "Activity 1",
			ParentID: "act-00000003", // act-003 を親とする
			Status:   ActivityStatusDraft,
			Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
		},
		{
			ID:       "act-00000002",
			Title:    "Activity 2",
			ParentID: "act-00000001", // act-001 を親とする
			Status:   ActivityStatusDraft,
			Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
		},
		{
			ID:       "act-00000003",
			Title:    "Activity 3",
			ParentID: "act-00000002", // act-002 を親とする
			Status:   ActivityStatusDraft,
			Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
		},
	}

	for _, act := range acts {
		if err := fs.WriteYaml(ctx, fmt.Sprintf("activities/%s.yaml", act.ID), act); err != nil {
			t.Fatalf("Write activity failed: %v", err)
		}
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

	if len(result.CycleErrors) > 0 {
		cycleErr := result.CycleErrors[0]
		if cycleErr.EntityType != "activity" {
			t.Errorf("expected entity type 'activity', got %q", cycleErr.EntityType)
		}
		if cycleErr.Message != "circular parent reference detected" {
			t.Errorf("expected message 'circular parent reference detected', got %q", cycleErr.Message)
		}
	}
}

// TestIntegrityChecker_ActivityDependencyCycle は Activity 依存関係の循環参照をテスト
func TestIntegrityChecker_ActivityDependencyCycle(t *testing.T) {
	checker, actHandler, _, _, zeusPath, cleanup := setupActivityIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 循環参照を持つ依存関係を作成: act-001 → act-002 → act-003 → act-001
	acts := []*ActivityEntity{
		{
			ID:           "act-00000001",
			Title:        "Activity 1",
			Dependencies: []string{"act-00000003"}, // act-003 に依存
			Status:       ActivityStatusDraft,
			Metadata:     Metadata{CreatedAt: Now(), UpdatedAt: Now()},
		},
		{
			ID:           "act-00000002",
			Title:        "Activity 2",
			Dependencies: []string{"act-00000001"}, // act-001 に依存
			Status:       ActivityStatusDraft,
			Metadata:     Metadata{CreatedAt: Now(), UpdatedAt: Now()},
		},
		{
			ID:           "act-00000003",
			Title:        "Activity 3",
			Dependencies: []string{"act-00000002"}, // act-002 に依存
			Status:       ActivityStatusDraft,
			Metadata:     Metadata{CreatedAt: Now(), UpdatedAt: Now()},
		},
	}

	for _, act := range acts {
		if err := fs.WriteYaml(ctx, fmt.Sprintf("activities/%s.yaml", act.ID), act); err != nil {
			t.Fatalf("Write activity failed: %v", err)
		}
	}

	// ActivityHandler のキャッシュをクリア（直接ファイルを書いたため）
	_ = actHandler // actHandler は使わないが、セットアップで必要

	// チェック実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	if result.Valid {
		t.Error("expected Valid to be false with dependency cycle")
	}

	// 循環参照エラーが検出されることを確認
	foundDependencyCycle := false
	for _, cycleErr := range result.CycleErrors {
		if cycleErr.EntityType == "activity" && cycleErr.Message == "circular dependency detected" {
			foundDependencyCycle = true
			break
		}
	}

	if !foundDependencyCycle {
		t.Error("expected to find dependency cycle error")
	}
}

// TestIntegrityChecker_ActivitySelfParentReference は自己参照（親が自分自身）をテスト
func TestIntegrityChecker_ActivitySelfParentReference(t *testing.T) {
	checker, _, _, _, zeusPath, cleanup := setupActivityIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 自己参照を持つ Activity を作成
	act := &ActivityEntity{
		ID:       "act-12345678",
		Title:    "自己参照アクティビティ",
		ParentID: "act-12345678", // 自分自身を親とする
		Status:   ActivityStatusDraft,
		Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "activities/act-12345678.yaml", act); err != nil {
		t.Fatalf("Write activity failed: %v", err)
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

// TestIntegrityChecker_ActivityMultipleWarnings は複数の Activity 参照警告をテスト
func TestIntegrityChecker_ActivityMultipleWarnings(t *testing.T) {
	checker, _, _, _, zeusPath, cleanup := setupActivityIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 複数の壊れた参照を持つ Activity を作成
	act := &ActivityEntity{
		ID:                  "act-12345678",
		Title:               "複数の壊れた参照",
		ParentID:            "act-99999999",           // 存在しない親
		Dependencies:        []string{"act-88888888"}, // 存在しない依存先
		RelatedDeliverables: []string{"del-77777777"}, // 存在しない成果物
		UseCaseID:           "uc-66666666",            // 存在しないユースケース
		Status:              ActivityStatusDraft,
		Metadata:            Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "activities/act-12345678.yaml", act); err != nil {
		t.Fatalf("Write activity failed: %v", err)
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

	// 4つの警告が期待される: 親、依存先、成果物、ユースケース
	if len(result.Warnings) != 4 {
		t.Errorf("expected 4 warnings, got %d", len(result.Warnings))
		for i, w := range result.Warnings {
			t.Logf("Warning %d: %s → %s: %s", i, w.SourceID, w.TargetID, w.Message)
		}
	}
}

// TestIntegrityChecker_ActivityValidReferences は正常な Activity 参照をテスト
func TestIntegrityChecker_ActivityValidReferences(t *testing.T) {
	checker, actHandler, delHandler, ucHandler, zeusPath, cleanup := setupActivityIntegrityTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// Objective を作成（UseCase の必須参照先）
	obj := &ObjectiveEntity{
		ID:       "obj-00000001",
		Title:    "テスト目標",
		Status:   ObjectiveStatusInProgress,
		Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "objectives/obj-00000001.yaml", obj); err != nil {
		t.Fatalf("Write objective failed: %v", err)
	}

	// Deliverable を作成
	delResult, err := delHandler.Add(ctx, "テスト成果物", WithDeliverableObjective("obj-00000001"))
	if err != nil {
		t.Fatalf("Add deliverable failed: %v", err)
	}

	// UseCase を作成
	ucResult, err := ucHandler.Add(ctx, "テストユースケース", WithUseCaseObjective("obj-00000001"))
	if err != nil {
		t.Fatalf("Add usecase failed: %v", err)
	}

	// 親 Activity を作成
	parentResult, err := actHandler.Add(ctx, "親アクティビティ")
	if err != nil {
		t.Fatalf("Add parent activity failed: %v", err)
	}

	// 依存先 Activity を作成
	depResult, err := actHandler.Add(ctx, "依存先アクティビティ")
	if err != nil {
		t.Fatalf("Add dependency activity failed: %v", err)
	}

	// 全ての正しい参照を持つ Activity を作成
	act := &ActivityEntity{
		ID:                  "act-12345678",
		Title:               "全ての参照が正常なアクティビティ",
		ParentID:            parentResult.ID,
		Dependencies:        []string{depResult.ID},
		RelatedDeliverables: []string{delResult.ID},
		UseCaseID:           ucResult.ID,
		Status:              ActivityStatusDraft,
		Metadata:            Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "activities/act-12345678.yaml", act); err != nil {
		t.Fatalf("Write activity failed: %v", err)
	}

	// チェック実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected Valid to be true, got errors: %v", result.ReferenceErrors)
	}

	if len(result.Warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(result.Warnings))
		for i, w := range result.Warnings {
			t.Logf("Warning %d: %s → %s: %s", i, w.SourceID, w.TargetID, w.Message)
		}
	}
}
