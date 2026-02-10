package core

import (
	"context"
	"os"
	"testing"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// テスト用のセットアップ（LintChecker）
func setupLintCheckerTest(t *testing.T) (*LintChecker, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeus-lint-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// .zeus ディレクトリを作成
	zeusPath := tmpDir + "/.zeus"
	if err := os.MkdirAll(zeusPath+"/objectives", 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create objectives dir: %v", err)
	}
	fs := yaml.NewFileManager(zeusPath)
	checker := NewLintChecker(fs)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return checker, zeusPath, cleanup
}

// ===== Vision ID フォーマットテスト =====

func TestLintChecker_VisionValidID(t *testing.T) {
	checker, zeusPath, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 正常な Vision ID を持つデータを作成
	vision := &Vision{
		ID:        "vision-001",
		Statement: "プロジェクトのビジョン",
		Metadata:  Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "vision.yaml", vision); err != nil {
		t.Fatalf("Write vision failed: %v", err)
	}

	// Lint チェック実行
	errors, warnings := checker.CheckIDFormat(ctx)

	if len(errors) != 0 {
		t.Errorf("expected 0 errors for valid vision ID, got %d: %v", len(errors), errors)
	}
	_ = warnings // 警告は無視
}

func TestLintChecker_VisionInvalidID(t *testing.T) {
	checker, zeusPath, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 不正な Vision ID を持つデータを作成
	vision := &Vision{
		ID:        "invalid-vision-id", // 不正な ID
		Statement: "プロジェクトのビジョン",
		Metadata:  Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "vision.yaml", vision); err != nil {
		t.Fatalf("Write vision failed: %v", err)
	}

	// Lint チェック実行
	errors, _ := checker.CheckIDFormat(ctx)

	if len(errors) != 1 {
		t.Errorf("expected 1 error for invalid vision ID, got %d", len(errors))
	}

	// エラー内容確認
	if len(errors) > 0 {
		lintErr := errors[0]
		if lintErr.EntityType != "vision" {
			t.Errorf("expected entity type 'vision', got %q", lintErr.EntityType)
		}
		if lintErr.EntityID != "invalid-vision-id" {
			t.Errorf("expected entity ID 'invalid-vision-id', got %q", lintErr.EntityID)
		}
		if lintErr.Field != "id" {
			t.Errorf("expected field 'id', got %q", lintErr.Field)
		}
	}
}

func TestLintChecker_VisionEmptyID(t *testing.T) {
	checker, zeusPath, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 空の Vision ID を持つデータを作成
	vision := &Vision{
		ID:        "", // 空の ID
		Statement: "プロジェクトのビジョン",
		Metadata:  Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "vision.yaml", vision); err != nil {
		t.Fatalf("Write vision failed: %v", err)
	}

	// Lint チェック実行
	errors, _ := checker.CheckIDFormat(ctx)

	if len(errors) != 1 {
		t.Errorf("expected 1 error for empty vision ID, got %d", len(errors))
	}
}

func TestLintChecker_VisionUUID(t *testing.T) {
	checker, zeusPath, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// UUID 形式の Vision ID を持つデータを作成
	vision := &Vision{
		ID:        "vision-12345678", // UUID 形式
		Statement: "プロジェクトのビジョン",
		Metadata:  Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "vision.yaml", vision); err != nil {
		t.Fatalf("Write vision failed: %v", err)
	}

	// Lint チェック実行
	errors, _ := checker.CheckIDFormat(ctx)

	if len(errors) != 0 {
		t.Errorf("expected 0 errors for UUID vision ID, got %d: %v", len(errors), errors)
	}
}

// ===== Constraint ID フォーマットテスト =====

func TestLintChecker_ConstraintValidID(t *testing.T) {
	checker, zeusPath, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 正常な Constraint ID を持つデータを作成
	constraintsFile := &ConstraintsFile{
		Constraints: []ConstraintEntity{
			{
				ID:       "const-001",
				Title:    "制約条件 1",
				Category: ConstraintCategoryTechnical,
			},
		},
	}
	if err := fs.WriteYaml(ctx, "constraints.yaml", constraintsFile); err != nil {
		t.Fatalf("Write constraints failed: %v", err)
	}

	// Lint チェック実行
	errors, warnings := checker.CheckIDFormat(ctx)

	if len(errors) != 0 {
		t.Errorf("expected 0 errors for valid constraint ID, got %d: %v", len(errors), errors)
	}
	_ = warnings // 警告は無視
}

func TestLintChecker_ConstraintInvalidID(t *testing.T) {
	checker, zeusPath, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 不正な Constraint ID を持つデータを作成
	constraintsFile := &ConstraintsFile{
		Constraints: []ConstraintEntity{
			{
				ID:       "invalid-constraint-id", // 不正な ID
				Title:    "制約条件 1",
				Category: ConstraintCategoryTechnical,
			},
		},
	}
	if err := fs.WriteYaml(ctx, "constraints.yaml", constraintsFile); err != nil {
		t.Fatalf("Write constraints failed: %v", err)
	}

	// Lint チェック実行
	errors, _ := checker.CheckIDFormat(ctx)

	if len(errors) != 1 {
		t.Errorf("expected 1 error for invalid constraint ID, got %d", len(errors))
	}

	// エラー内容確認
	if len(errors) > 0 {
		lintErr := errors[0]
		if lintErr.EntityType != "constraint" {
			t.Errorf("expected entity type 'constraint', got %q", lintErr.EntityType)
		}
		if lintErr.EntityID != "invalid-constraint-id" {
			t.Errorf("expected entity ID 'invalid-constraint-id', got %q", lintErr.EntityID)
		}
		if lintErr.Field != "id" {
			t.Errorf("expected field 'id', got %q", lintErr.Field)
		}
	}
}

func TestLintChecker_ConstraintMultipleInvalidIDs(t *testing.T) {
	checker, zeusPath, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 複数の不正な Constraint ID を持つデータを作成
	constraintsFile := &ConstraintsFile{
		Constraints: []ConstraintEntity{
			{
				ID:       "const-001", // 正常
				Title:    "制約条件 1",
				Category: ConstraintCategoryTechnical,
			},
			{
				ID:       "bad-id", // 不正
				Title:    "制約条件 2",
				Category: ConstraintCategoryTechnical,
			},
			{
				ID:       "const-002", // 正常
				Title:    "制約条件 3",
				Category: ConstraintCategoryTechnical,
			},
			{
				ID:       "wrong", // 不正
				Title:    "制約条件 4",
				Category: ConstraintCategoryTechnical,
			},
		},
	}
	if err := fs.WriteYaml(ctx, "constraints.yaml", constraintsFile); err != nil {
		t.Fatalf("Write constraints failed: %v", err)
	}

	// Lint チェック実行
	errors, _ := checker.CheckIDFormat(ctx)

	if len(errors) != 2 {
		t.Errorf("expected 2 errors for invalid constraint IDs, got %d", len(errors))
	}
}

func TestLintChecker_ConstraintEmptyFile(t *testing.T) {
	checker, zeusPath, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 空の Constraints ファイルを作成
	constraintsFile := &ConstraintsFile{
		Constraints: []ConstraintEntity{},
	}
	if err := fs.WriteYaml(ctx, "constraints.yaml", constraintsFile); err != nil {
		t.Fatalf("Write constraints failed: %v", err)
	}

	// Lint チェック実行
	errors, warnings := checker.CheckIDFormat(ctx)

	if len(errors) != 0 {
		t.Errorf("expected 0 errors for empty constraints file, got %d", len(errors))
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings for empty constraints file, got %d", len(warnings))
	}
}

func TestLintChecker_ConstraintUUID(t *testing.T) {
	checker, zeusPath, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// UUID 形式の Constraint ID を持つデータを作成
	constraintsFile := &ConstraintsFile{
		Constraints: []ConstraintEntity{
			{
				ID:       "const-12345678", // UUID 形式
				Title:    "制約条件 1",
				Category: ConstraintCategoryTechnical,
			},
		},
	}
	if err := fs.WriteYaml(ctx, "constraints.yaml", constraintsFile); err != nil {
		t.Fatalf("Write constraints failed: %v", err)
	}

	// Lint チェック実行
	errors, _ := checker.CheckIDFormat(ctx)

	if len(errors) != 0 {
		t.Errorf("expected 0 errors for UUID constraint ID, got %d: %v", len(errors), errors)
	}
}

// ===== CheckAll テスト =====

func TestLintChecker_CheckAllClean(t *testing.T) {
	checker, zeusPath, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 正常なデータを作成
	vision := &Vision{
		ID:        "vision-001",
		Statement: "プロジェクトのビジョン",
		Metadata:  Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "vision.yaml", vision); err != nil {
		t.Fatalf("Write vision failed: %v", err)
	}

	constraintsFile := &ConstraintsFile{
		Constraints: []ConstraintEntity{
			{
				ID:       "const-001",
				Title:    "制約条件 1",
				Category: ConstraintCategoryTechnical,
			},
		},
	}
	if err := fs.WriteYaml(ctx, "constraints.yaml", constraintsFile); err != nil {
		t.Fatalf("Write constraints failed: %v", err)
	}

	// CheckAll 実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	if !result.Valid {
		t.Error("expected Valid to be true")
	}

	if len(result.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestLintChecker_CheckAllWithErrors(t *testing.T) {
	checker, zeusPath, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 不正なデータを作成
	vision := &Vision{
		ID:        "bad-vision", // 不正
		Statement: "プロジェクトのビジョン",
		Metadata:  Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "vision.yaml", vision); err != nil {
		t.Fatalf("Write vision failed: %v", err)
	}

	constraintsFile := &ConstraintsFile{
		Constraints: []ConstraintEntity{
			{
				ID:       "bad-const", // 不正
				Title:    "制約条件 1",
				Category: ConstraintCategoryTechnical,
			},
		},
	}
	if err := fs.WriteYaml(ctx, "constraints.yaml", constraintsFile); err != nil {
		t.Fatalf("Write constraints failed: %v", err)
	}

	// CheckAll 実行
	result, err := checker.CheckAll(ctx)
	if err != nil {
		t.Fatalf("CheckAll failed: %v", err)
	}

	if result.Valid {
		t.Error("expected Valid to be false")
	}

	if len(result.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(result.Errors))
	}
}

// ===== コンテキストキャンセルテスト =====

func TestLintChecker_ContextCancellation(t *testing.T) {
	checker, _, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	// キャンセル済みのコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// CheckAll
	_, err := checker.CheckAll(ctx)
	if err == nil {
		t.Error("CheckAll should fail with cancelled context")
	}
}

// ===== エラーメッセージテスト =====

func TestLintError_Error(t *testing.T) {
	lintErr := &LintError{
		EntityType: "vision",
		EntityID:   "vision-001",
		Field:      "id",
		Message:    "ID format mismatch",
		Expected:   "vision-NNN",
		Actual:     "bad-id",
	}

	expected := "[vision] vision-001.id: ID format mismatch (expected: vision-NNN, actual: bad-id)"
	if lintErr.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, lintErr.Error())
	}
}

func TestLintWarning_Warning(t *testing.T) {
	lintWarn := &LintWarning{
		EntityType: "objective",
		EntityID:   "obj-001",
		Field:      "status",
		Message:    "status should be completed",
		Suggested:  "completed",
		Actual:     "in_progress",
	}

	expected := "[objective] obj-001.status: status should be completed (suggested: completed, actual: in_progress)"
	if lintWarn.Warning() != expected {
		t.Errorf("expected warning message %q, got %q", expected, lintWarn.Warning())
	}
}

// ===== ファイル不存在時のテスト =====

func TestLintChecker_VisionFileNotExists(t *testing.T) {
	checker, _, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	ctx := context.Background()

	// vision.yaml が存在しない状態で Lint チェック実行
	errors, warnings := checker.CheckIDFormat(ctx)

	// ファイルが存在しない場合はエラーも警告もなし
	if len(errors) != 0 {
		t.Errorf("expected 0 errors when vision.yaml doesn't exist, got %d", len(errors))
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings when vision.yaml doesn't exist, got %d", len(warnings))
	}
}

func TestLintChecker_ConstraintFileNotExists(t *testing.T) {
	checker, _, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	ctx := context.Background()

	// constraints.yaml が存在しない状態で Lint チェック実行
	errors, warnings := checker.CheckIDFormat(ctx)

	// ファイルが存在しない場合はエラーも警告もなし
	if len(errors) != 0 {
		t.Errorf("expected 0 errors when constraints.yaml doesn't exist, got %d", len(errors))
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings when constraints.yaml doesn't exist, got %d", len(warnings))
	}
}

// ===== YAML 未知フィールド検出テスト =====

func TestLintChecker_UnknownFields_Clean(t *testing.T) {
	checker, zeusPath, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	ctx := context.Background()
	fs := yaml.NewFileManager(zeusPath)

	// 正常な Objective を作成（未知フィールドなし）
	obj := &ObjectiveEntity{
		ID:       "obj-001",
		Title:    "テスト目標",
		Status:   "active",
		Metadata: Metadata{CreatedAt: Now(), UpdatedAt: Now()},
	}
	if err := fs.WriteYaml(ctx, "objectives/obj-001.yaml", obj); err != nil {
		t.Fatalf("Write objective failed: %v", err)
	}

	// 未知フィールドチェック実行
	_, warnings := checker.CheckUnknownFields(ctx)

	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings for clean YAML, got %d: %v", len(warnings), warnings)
		for _, w := range warnings {
			t.Logf("  warning: %s", w.Warning())
		}
	}
}

func TestLintChecker_UnknownFields_WithUnknown(t *testing.T) {
	checker, zeusPath, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	ctx := context.Background()

	// 未知フィールドを含む YAML を直接書き込み
	yamlContent := []byte(`id: obj-001
title: "テスト目標"
status: active
unknown_field: "これは未知のフィールド"
metadata:
  created_at: "2026-01-01T00:00:00Z"
  updated_at: "2026-01-01T00:00:00Z"
`)
	objPath := zeusPath + "/objectives/obj-001.yaml"
	if err := os.WriteFile(objPath, yamlContent, 0644); err != nil {
		t.Fatalf("Write YAML failed: %v", err)
	}

	// 未知フィールドチェック実行
	_, warnings := checker.CheckUnknownFields(ctx)

	if len(warnings) != 1 {
		t.Errorf("expected 1 warning for unknown field, got %d", len(warnings))
		for _, w := range warnings {
			t.Logf("  warning: %s", w.Warning())
		}
		return
	}

	w := warnings[0]
	if w.EntityType != "objective" {
		t.Errorf("expected entity type 'objective', got %q", w.EntityType)
	}
	if w.Field != "unknown_fields" {
		t.Errorf("expected field 'unknown_fields', got %q", w.Field)
	}
}

func TestLintChecker_UnknownFields_Activity(t *testing.T) {
	checker, zeusPath, cleanup := setupLintCheckerTest(t)
	defer cleanup()

	ctx := context.Background()

	// activities ディレクトリを作成
	if err := os.MkdirAll(zeusPath+"/activities", 0755); err != nil {
		t.Fatalf("failed to create activities dir: %v", err)
	}

	// 未知フィールド（旧 deliverables）を含む Activity YAML を直接書き込み
	yamlContent := []byte(`id: act-001
title: "テストアクティビティ"
status: active
deliverables:
  - name: "削除済みフィールド"
metadata:
  created_at: "2026-01-01T00:00:00Z"
  updated_at: "2026-01-01T00:00:00Z"
`)
	actPath := zeusPath + "/activities/act-001.yaml"
	if err := os.WriteFile(actPath, yamlContent, 0644); err != nil {
		t.Fatalf("Write YAML failed: %v", err)
	}

	// 未知フィールドチェック実行
	_, warnings := checker.CheckUnknownFields(ctx)

	if len(warnings) != 1 {
		t.Errorf("expected 1 warning for unknown field in activity, got %d", len(warnings))
		for _, w := range warnings {
			t.Logf("  warning: %s", w.Warning())
		}
		return
	}

	w := warnings[0]
	if w.EntityType != "activity" {
		t.Errorf("expected entity type 'activity', got %q", w.EntityType)
	}
}

// Note: Activity status/progress 整合性テストは progress 機能削除に伴い削除
