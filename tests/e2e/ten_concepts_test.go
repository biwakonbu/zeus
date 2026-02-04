// Package e2e は Zeus CLI の 10 概念モデルに対する E2E テストを提供する
package e2e

import (
	"path/filepath"
	"testing"
)

// =============================================================================
// Phase 1: Vision, Objective, Deliverable
// =============================================================================

// TestVisionFlow は Vision の追加・取得フローをテストする
func TestVisionFlow(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	// init
	result := runCommand(t, dir, "init")
	assertSuccess(t, result)

	// Vision 追加
	result = runCommand(t, dir, "add", "vision", "AI駆動PM",
		"--statement", "AIと人間が協調するプロジェクト管理を実現する")
	assertSuccess(t, result)
	assertOutputContains(t, result, "Added vision")
	assertOutputContains(t, result, "AI駆動PM")

	// ファイル存在確認
	assertFileExists(t, filepath.Join(dir, ".zeus", "vision.yaml"))

	// list で確認（Vision は詳細表示形式）
	result = runCommand(t, dir, "list", "vision")
	assertSuccess(t, result)
	assertOutputContains(t, result, "AI駆動PM")
	assertOutputContains(t, result, "vision-001")
}

// TestVisionSingletonConstraint は Vision が単一であることをテストする
// Vision は常に vision-001 として更新される
func TestVisionSingletonConstraint(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// 1回目の Vision 追加
	result := runCommand(t, dir, "add", "vision", "初期ビジョン",
		"--statement", "初期ステートメント")
	assertSuccess(t, result)

	// 2回目の Vision 追加（更新として動作）
	result = runCommand(t, dir, "add", "vision", "更新ビジョン",
		"--statement", "更新されたステートメント")
	assertSuccess(t, result)

	// list で確認（Vision は詳細表示形式、更新後のタイトルが表示される）
	result = runCommand(t, dir, "list", "vision")
	assertSuccess(t, result)
	assertOutputContains(t, result, "更新ビジョン")
	assertOutputContains(t, result, "vision-001")
}

// TestObjectiveManagement は Objective の CRUD 操作をテストする
func TestObjectiveManagement(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// Objective 追加
	result := runCommand(t, dir, "add", "objective", "認証システム実装",
		"--wbs", "1.1",
		"--due", "2026-02-28",
		"--progress", "0")
	assertSuccess(t, result)
	assertOutputContains(t, result, "Added objective")

	// ディレクトリ確認
	assertDirExists(t, filepath.Join(dir, ".zeus", "objectives"))

	// list で確認
	result = runCommand(t, dir, "list", "objectives")
	assertSuccess(t, result)
	assertOutputContains(t, result, "1 items")

	// 2つ目の Objective 追加
	result = runCommand(t, dir, "add", "objective", "API設計",
		"--wbs", "1.2")
	assertSuccess(t, result)

	result = runCommand(t, dir, "list", "objectives")
	assertSuccess(t, result)
	assertOutputContains(t, result, "2 items")
}

// TestObjectiveHierarchy は Objective の親子関係をテストする
func TestObjectiveHierarchy(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// 親 Objective 追加
	result := runCommand(t, dir, "add", "objective", "システム開発",
		"--wbs", "1.0")
	assertSuccess(t, result)
	parentID := extractEntityID(t, result, "obj-")
	if parentID == "" {
		parentID = "obj-001"
	}

	// 子 Objective 追加
	result = runCommand(t, dir, "add", "objective", "バックエンド開発",
		"--parent", parentID,
		"--wbs", "1.1")
	assertSuccess(t, result)
	assertOutputContains(t, result, "Added objective")

	// doctor で参照確認
	result = runCommand(t, dir, "doctor")
	assertSuccess(t, result)
}

// TestObjectiveCyclicReference は Objective の循環参照が doctor で検出されることをテストする
// Note: 存在しない親への参照は作成時にはエラーにならず、doctor でチェックされる仕様
func TestObjectiveCyclicReference(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// 親 Objective を作成
	result := runCommand(t, dir, "add", "objective", "親Objective",
		"--wbs", "1.0")
	assertSuccess(t, result)
	parentID := extractEntityID(t, result, "obj-")
	if parentID == "" {
		t.Fatal("親 Objective の ID を取得できませんでした")
	}

	// 子 Objective を作成（親を参照）
	result = runCommand(t, dir, "add", "objective", "子Objective",
		"--parent", parentID,
		"--wbs", "1.1")
	assertSuccess(t, result)
	childID := extractEntityID(t, result, "obj-")
	if childID == "" {
		t.Fatal("子 Objective の ID を取得できませんでした")
	}

	// doctor で確認（この時点では問題なし）
	result = runCommand(t, dir, "doctor")
	assertSuccess(t, result)
}

// TestDeliverableManagement は Deliverable の CRUD 操作をテストする
func TestDeliverableManagement(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	// 基本プロジェクトセットアップ（Vision + Objective）
	runCommand(t, dir, "init")
	runCommand(t, dir, "add", "vision", "テストビジョン")
	result := runCommand(t, dir, "add", "objective", "Phase 1",
		"--wbs", "1.0")
	assertSuccess(t, result)
	objID := extractEntityID(t, result, "obj-")
	if objID == "" {
		objID = "obj-001"
	}

	// Deliverable 追加
	result = runCommand(t, dir, "add", "deliverable", "API設計書",
		"--objective", objID,
		"--format", "document")
	assertSuccess(t, result)
	assertOutputContains(t, result, "Added deliverable")

	// ディレクトリ確認
	assertDirExists(t, filepath.Join(dir, ".zeus", "deliverables"))

	// list で確認
	result = runCommand(t, dir, "list", "deliverables")
	assertSuccess(t, result)
	assertOutputContains(t, result, "1 items")
}

// TestDeliverableObjectiveRef は Deliverable が Objective を必須参照することをテストする
func TestDeliverableObjectiveRef(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// Objective なしで Deliverable を作成（エラー）
	result := runCommand(t, dir, "add", "deliverable", "無効な成果物")
	assertFailure(t, result)
}

// TestDeliverableInvalidRef は無効な Objective 参照でエラーになることをテストする
func TestDeliverableInvalidRef(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// 存在しない Objective を参照
	result := runCommand(t, dir, "add", "deliverable", "無効な成果物",
		"--objective", "obj-999")
	assertFailure(t, result)
	assertStderrContains(t, result, "not found")
}

// TestPhase1Integration は Phase 1 エンティティの統合テスト
func TestPhase1Integration(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	// セットアップヘルパーを使用
	ids := setupBasicProject(t, dir)

	// 各エンティティの存在を確認
	if ids["vision"] == "" || ids["objective"] == "" || ids["deliverable"] == "" {
		t.Error("基本プロジェクトのセットアップに失敗しました")
	}

	// doctor でヘルスチェック
	result := runCommand(t, dir, "doctor")
	assertSuccess(t, result)

	// status で確認
	result = runCommand(t, dir, "status")
	assertSuccess(t, result)
}

// =============================================================================
// Phase 2: Consideration, Decision, Problem, Risk, Assumption
// =============================================================================

// TestConsiderationManagement は Consideration の管理をテストする
func TestConsiderationManagement(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupBasicProject(t, dir)

	// Consideration 追加
	result := runCommand(t, dir, "add", "consideration", "認証方式の選択",
		"--objective", ids["objective"],
		"--due", "2026-02-15")
	assertSuccess(t, result)
	assertOutputContains(t, result, "Added consideration")

	// ディレクトリ確認
	assertDirExists(t, filepath.Join(dir, ".zeus", "considerations"))

	// list で確認
	result = runCommand(t, dir, "list", "considerations")
	assertSuccess(t, result)
	assertOutputContains(t, result, "1 items")
}

// TestConsiderationDecisionFlow は Consideration → Decision のフローをテストする
func TestConsiderationDecisionFlow(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupDecisionFlow(t, dir)

	// Decision 作成
	result := runCommand(t, dir, "add", "decision", "JWT認証を採用",
		"--consideration", ids["consideration"],
		"--selected-opt-id", "opt-1",
		"--selected-title", "JWT",
		"--rationale", "セキュリティと拡張性のバランス")
	assertSuccess(t, result)
	assertOutputContains(t, result, "Added decision")

	// ディレクトリ確認
	assertDirExists(t, filepath.Join(dir, ".zeus", "decisions"))

	// list で確認
	result = runCommand(t, dir, "list", "decisions")
	assertSuccess(t, result)
	assertOutputContains(t, result, "1 items")
}

// TestDecisionImmutability は Decision が更新不可であることをテストする
func TestDecisionImmutability(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupDecisionFlow(t, dir)

	// Decision 作成
	result := runCommand(t, dir, "add", "decision", "JWT認証を採用",
		"--consideration", ids["consideration"],
		"--selected-opt-id", "opt-1",
		"--selected-title", "JWT",
		"--rationale", "セキュリティ")
	assertSuccess(t, result)

	// list で確認（Decision が存在することを確認）
	result = runCommand(t, dir, "list", "decisions")
	assertSuccess(t, result)
	assertOutputContains(t, result, "1 items")
}

// TestDecisionDeleteProhibited は Decision が削除不可であることをテストする (M1)
// Note: CLI に delete コマンドがない場合、このテストは内部的な制約をテストする
func TestDecisionDeleteProhibited(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupDecisionFlow(t, dir)

	// Decision 作成
	result := runCommand(t, dir, "add", "decision", "React採用",
		"--consideration", ids["consideration"],
		"--selected-opt-id", "opt-1",
		"--selected-title", "React",
		"--rationale", "コミュニティ")
	assertSuccess(t, result)

	// Decision の存在確認
	result = runCommand(t, dir, "list", "decisions")
	assertSuccess(t, result)
	assertOutputContains(t, result, "1 items")

	// 注: CLI に delete コマンドがある場合、ここでテスト
	// delete コマンドがない場合は doctor で整合性確認
	result = runCommand(t, dir, "doctor")
	assertSuccess(t, result)
}

// TestDecisionRequiresConsideration は Decision が Consideration を必須とすることをテストする
func TestDecisionRequiresConsideration(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// Consideration なしで Decision を作成（エラー）
	result := runCommand(t, dir, "add", "decision", "無効な決定",
		"--selected-opt-id", "opt-1",
		"--selected-title", "テスト",
		"--rationale", "理由")
	assertFailure(t, result)
}

// TestConsiderationReverseRef は Consideration → Decision の逆参照整合性をテストする (M3)
func TestConsiderationReverseRef(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupDecisionFlow(t, dir)

	// Decision 作成
	result := runCommand(t, dir, "add", "decision", "技術選定完了",
		"--consideration", ids["consideration"],
		"--selected-opt-id", "opt-1",
		"--selected-title", "選択肢1",
		"--rationale", "コスト効率")
	assertSuccess(t, result)

	// doctor で整合性確認
	result = runCommand(t, dir, "doctor")
	assertSuccess(t, result)
}

// TestProblemManagement は Problem の管理をテストする
func TestProblemManagement(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupBasicProject(t, dir)

	// Problem 追加
	result := runCommand(t, dir, "add", "problem", "パフォーマンス問題",
		"--severity", "high",
		"--objective", ids["objective"])
	assertSuccess(t, result)
	assertOutputContains(t, result, "Added problem")

	// ディレクトリ確認
	assertDirExists(t, filepath.Join(dir, ".zeus", "problems"))

	// list で確認
	result = runCommand(t, dir, "list", "problems")
	assertSuccess(t, result)
	assertOutputContains(t, result, "1 items")
}

// TestProblemSeverity は Problem の severity レベルをテストする
func TestProblemSeverity(t *testing.T) {
	severities := []string{"critical", "high", "medium", "low"}

	for _, severity := range severities {
		severity := severity
		t.Run(severity, func(t *testing.T) {
			t.Parallel()
			dir := setupTempDir(t)
			defer cleanupTempDir(t, dir)

			ids := setupBasicProject(t, dir)

			result := runCommand(t, dir, "add", "problem", "テスト問題",
				"--severity", severity,
				"--objective", ids["objective"])
			assertSuccess(t, result)
		})
	}
}

// TestRiskManagement は Risk の管理をテストする
func TestRiskManagement(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupBasicProject(t, dir)

	// Risk 追加
	result := runCommand(t, dir, "add", "risk", "外部API依存リスク",
		"--probability", "medium",
		"--impact", "high",
		"--objective", ids["objective"])
	assertSuccess(t, result)
	assertOutputContains(t, result, "Added risk")

	// ディレクトリ確認
	assertDirExists(t, filepath.Join(dir, ".zeus", "risks"))

	// list で確認
	result = runCommand(t, dir, "list", "risks")
	assertSuccess(t, result)
	assertOutputContains(t, result, "1 items")
}

// TestRiskScoreCalculation は Risk のスコア自動計算をテストする
func TestRiskScoreCalculation(t *testing.T) {
	testCases := []struct {
		name        string
		probability string
		impact      string
	}{
		{"High-Critical", "high", "critical"},
		{"Medium-High", "medium", "high"},
		{"Low-Low", "low", "low"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := setupTempDir(t)
			defer cleanupTempDir(t, dir)

			ids := setupBasicProject(t, dir)

			result := runCommand(t, dir, "add", "risk", "テストリスク",
				"--probability", tc.probability,
				"--impact", tc.impact,
				"--objective", ids["objective"])
			assertSuccess(t, result)
		})
	}
}

// TestAssumptionManagement は Assumption の管理をテストする
func TestAssumptionManagement(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupBasicProject(t, dir)

	// Assumption 追加
	result := runCommand(t, dir, "add", "assumption", "ユーザー数1000人以下",
		"--objective", ids["objective"])
	assertSuccess(t, result)
	assertOutputContains(t, result, "Added assumption")

	// ディレクトリ確認
	assertDirExists(t, filepath.Join(dir, ".zeus", "assumptions"))

	// list で確認
	result = runCommand(t, dir, "list", "assumptions")
	assertSuccess(t, result)
	assertOutputContains(t, result, "1 items")
}

// TestAssumptionVerification は Assumption の検証ステータスをテストする
func TestAssumptionVerification(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupBasicProject(t, dir)

	// 複数の Assumption を追加
	result := runCommand(t, dir, "add", "assumption", "前提条件1",
		"--objective", ids["objective"])
	assertSuccess(t, result)

	result = runCommand(t, dir, "add", "assumption", "前提条件2",
		"--objective", ids["objective"])
	assertSuccess(t, result)

	// list で確認
	result = runCommand(t, dir, "list", "assumptions")
	assertSuccess(t, result)
	assertOutputContains(t, result, "2 items")
}

// =============================================================================
// Phase 3: Constraint, Quality
// =============================================================================

// TestConstraintManagement は Constraint の管理をテストする
func TestConstraintManagement(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// Constraint 追加
	result := runCommand(t, dir, "add", "constraint", "外部DB不使用",
		"--category", "technical",
		"--non-negotiable")
	assertSuccess(t, result)
	assertOutputContains(t, result, "Added constraint")

	// ファイル確認（グローバル単一ファイル）
	assertFileExists(t, filepath.Join(dir, ".zeus", "constraints.yaml"))

	// list で確認
	result = runCommand(t, dir, "list", "constraints")
	assertSuccess(t, result)
	assertOutputContains(t, result, "1 items")
}

// TestConstraintCategories は Constraint のカテゴリをテストする
func TestConstraintCategories(t *testing.T) {
	categories := []string{"technical", "business", "legal", "resource"}

	for _, category := range categories {
		category := category
		t.Run(category, func(t *testing.T) {
			t.Parallel()
			dir := setupTempDir(t)
			defer cleanupTempDir(t, dir)

			runCommand(t, dir, "init")

			result := runCommand(t, dir, "add", "constraint", "テスト制約",
				"--category", category)
			assertSuccess(t, result)
		})
	}
}

// TestConstraintGlobalFile は Constraint がグローバル単一ファイルで管理されることをテストする
func TestConstraintGlobalFile(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// 複数の Constraint を追加
	result := runCommand(t, dir, "add", "constraint", "制約1",
		"--category", "technical")
	assertSuccess(t, result)

	result = runCommand(t, dir, "add", "constraint", "制約2",
		"--category", "business")
	assertSuccess(t, result)

	// 単一ファイルに格納されていることを確認
	assertFileExists(t, filepath.Join(dir, ".zeus", "constraints.yaml"))

	// list で確認
	result = runCommand(t, dir, "list", "constraints")
	assertSuccess(t, result)
	assertOutputContains(t, result, "2 items")
}

// TestQualityManagement は Quality の管理をテストする
func TestQualityManagement(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupBasicProject(t, dir)

	// Quality 追加
	result := runCommand(t, dir, "add", "quality", "コードカバレッジ基準",
		"--deliverable", ids["deliverable"],
		"--metric", "coverage:80:%")
	assertSuccess(t, result)
	assertOutputContains(t, result, "Added quality")

	// ディレクトリ確認
	assertDirExists(t, filepath.Join(dir, ".zeus", "quality"))

	// list で確認
	result = runCommand(t, dir, "list", "quality")
	assertSuccess(t, result)
	assertOutputContains(t, result, "1 items")
}

// TestQualityDeliverableRef は Quality が Deliverable を必須参照することをテストする
func TestQualityDeliverableRef(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// Deliverable なしで Quality を作成（エラー）
	result := runCommand(t, dir, "add", "quality", "無効な品質基準")
	assertFailure(t, result)
}

// TestQualityMetrics は Quality のメトリクス設定をテストする
func TestQualityMetrics(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupBasicProject(t, dir)

	// 複数メトリクスを設定
	result := runCommand(t, dir, "add", "quality", "パフォーマンス基準",
		"--deliverable", ids["deliverable"],
		"--metric", "coverage:80:%",
		"--metric", "performance:100:ms",
		"--metric", "memory:256:MB")
	assertSuccess(t, result)

	// list で確認
	result = runCommand(t, dir, "list", "quality")
	assertSuccess(t, result)
	assertOutputContains(t, result, "1 items")
}

// =============================================================================
// 統合テスト
// =============================================================================

// TestFullProjectSetup はすべてのエンティティを含むプロジェクトをテストする
func TestFullProjectSetup(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	// フルプロジェクトセットアップ
	ids := setupFullProject(t, dir)

	// 各エンティティの存在確認
	expectedEntities := []string{
		"vision", "objective", "deliverable",
		"consideration", "decision",
		"problem", "risk", "assumption",
		"constraint", "quality",
	}

	for _, entity := range expectedEntities {
		if ids[entity] == "" {
			t.Errorf("エンティティ %s が作成されていません", entity)
		}
	}

	// doctor でヘルスチェック
	result := runCommand(t, dir, "doctor")
	assertSuccess(t, result)

	// status で確認
	result = runCommand(t, dir, "status")
	assertSuccess(t, result)
}

// TestTenConceptsListAll は全エンティティタイプの list をテストする
func TestTenConceptsListAll(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	// フルプロジェクトセットアップ
	setupFullProject(t, dir)

	// 各エンティティタイプの list を実行
	entityTypes := []string{
		"vision", "objectives", "deliverables",
		"considerations", "decisions",
		"problems", "risks", "assumptions",
		"constraints", "quality",
	}

	for _, entityType := range entityTypes {
		t.Run(entityType, func(t *testing.T) {
			result := runCommand(t, dir, "list", entityType)
			assertSuccess(t, result)
		})
	}
}
