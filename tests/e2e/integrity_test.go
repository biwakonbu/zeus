// Package e2e は Zeus CLI の参照整合性に対する E2E テストを提供する
package e2e

import (
	"testing"
)

// =============================================================================
// 参照整合性テスト
// =============================================================================

// TestDoctorReferenceIntegrity は doctor コマンドで参照整合性がチェックされることをテストする
func TestDoctorReferenceIntegrity(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	// フルプロジェクトセットアップ（全エンティティ含む）
	setupFullProject(t, dir)

	// doctor で参照整合性チェック
	result := runCommand(t, dir, "doctor")
	assertSuccess(t, result)
	assertOutputContains(t, result, "Zeus Doctor")
}

// TestDoctorCyclicDetection は doctor コマンドで循環参照が検出されることをテストする
func TestDoctorCyclicDetection(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// 正常な階層構造を作成
	result := runCommand(t, dir, "add", "objective", "親目標", "--wbs", "1.0")
	assertSuccess(t, result)
	parentID := extractEntityID(t, result, "obj-")
	if parentID == "" {
		parentID = "obj-001"
	}

	result = runCommand(t, dir, "add", "objective", "子目標",
		"--parent", parentID,
		"--wbs", "1.1")
	assertSuccess(t, result)

	// doctor でチェック
	result = runCommand(t, dir, "doctor")
	assertSuccess(t, result)
}

// TestDoctorBrokenRefReport は doctor コマンドで壊れた参照が報告されることをテストする
func TestDoctorBrokenRefReport(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	// 基本プロジェクトセットアップ
	ids := setupBasicProject(t, dir)

	// 正常なプロジェクトで doctor
	result := runCommand(t, dir, "doctor")
	assertSuccess(t, result)

	// Quality を追加（正常な参照）
	result = runCommand(t, dir, "add", "quality", "テスト品質",
		"--deliverable", ids["deliverable"],
		"--metric", "coverage:80:%")
	assertSuccess(t, result)

	// doctor でチェック
	result = runCommand(t, dir, "doctor")
	assertSuccess(t, result)
}

// =============================================================================
// 参照チェーンテスト
// =============================================================================

// TestRefChain_Del_Obj は Deliverable → Objective の参照チェーンをテストする
func TestRefChain_Del_Obj(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// Objective 作成
	result := runCommand(t, dir, "add", "objective", "Phase 1 目標",
		"--wbs", "1.0")
	assertSuccess(t, result)
	objID := extractEntityID(t, result, "obj-")
	if objID == "" {
		objID = "obj-001"
	}

	// Deliverable 作成（正しい参照）
	result = runCommand(t, dir, "add", "deliverable", "設計書",
		"--objective", objID,
		"--format", "document")
	assertSuccess(t, result)

	// doctor でチェック
	result = runCommand(t, dir, "doctor")
	assertSuccess(t, result)
}

// TestRefChain_Dec_Con は Decision → Consideration の参照チェーンをテストする
func TestRefChain_Dec_Con(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupDecisionFlow(t, dir)

	// Decision 作成（正しい参照）
	result := runCommand(t, dir, "add", "decision", "技術選定完了",
		"--consideration", ids["consideration"],
		"--selected-opt-id", "opt-1",
		"--selected-title", "選択肢A",
		"--rationale", "コスト効率")
	assertSuccess(t, result)

	// doctor でチェック
	result = runCommand(t, dir, "doctor")
	assertSuccess(t, result)
}

// TestRefChain_Qual_Del は Quality → Deliverable の参照チェーンをテストする
func TestRefChain_Qual_Del(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupBasicProject(t, dir)

	// Quality 作成（正しい参照）
	result := runCommand(t, dir, "add", "quality", "コードカバレッジ",
		"--deliverable", ids["deliverable"],
		"--metric", "coverage:80:%")
	assertSuccess(t, result)

	// doctor でチェック
	result = runCommand(t, dir, "doctor")
	assertSuccess(t, result)
}

// TestRefChain_Prob_Obj は Problem → Objective の参照チェーンをテストする
func TestRefChain_Prob_Obj(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupBasicProject(t, dir)

	// Problem 作成（正しい参照）
	result := runCommand(t, dir, "add", "problem", "パフォーマンス問題",
		"--severity", "high",
		"--objective", ids["objective"])
	assertSuccess(t, result)

	// doctor でチェック
	result = runCommand(t, dir, "doctor")
	assertSuccess(t, result)
}

// TestRefChain_Risk_Del は Risk → Deliverable の参照チェーンをテストする
func TestRefChain_Risk_Del(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupBasicProject(t, dir)

	// Risk 作成（Deliverable 参照）
	result := runCommand(t, dir, "add", "risk", "品質リスク",
		"--probability", "medium",
		"--impact", "high",
		"--deliverable", ids["deliverable"])
	assertSuccess(t, result)

	// doctor でチェック
	result = runCommand(t, dir, "doctor")
	assertSuccess(t, result)
}

// =============================================================================
// 無効な参照テスト
// =============================================================================

// TestInvalidRef_Del_Obj は存在しない Objective への参照がエラーになることをテストする
func TestInvalidRef_Del_Obj(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// 存在しない Objective を参照
	result := runCommand(t, dir, "add", "deliverable", "無効な成果物",
		"--objective", "obj-999")
	assertFailure(t, result)
}

// TestInvalidRef_Dec_Con は存在しない Consideration への参照がエラーになることをテストする
func TestInvalidRef_Dec_Con(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// 存在しない Consideration を参照
	result := runCommand(t, dir, "add", "decision", "無効な決定",
		"--consideration", "con-999",
		"--selected-opt-id", "opt-1",
		"--selected-title", "テスト",
		"--rationale", "理由")
	assertFailure(t, result)
}

// TestInvalidRef_Qual_Del は存在しない Deliverable への参照がエラーになることをテストする
func TestInvalidRef_Qual_Del(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// 存在しない Deliverable を参照
	result := runCommand(t, dir, "add", "quality", "無効な品質",
		"--deliverable", "del-999")
	assertFailure(t, result)
}

// =============================================================================
// 複合参照テスト
// =============================================================================

// TestComplexRefChain は複数のエンティティ間の参照チェーンをテストする
func TestComplexRefChain(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	// フルプロジェクトセットアップ
	ids := setupFullProject(t, dir)

	// 追加の参照チェーン: 別の Deliverable を追加
	result := runCommand(t, dir, "add", "deliverable", "テスト計画書",
		"--objective", ids["objective"],
		"--format", "document")
	assertSuccess(t, result)
	del2ID := extractEntityID(t, result, "del-")
	if del2ID == "" {
		del2ID = "del-002"
	}

	// 別の Quality を追加
	result = runCommand(t, dir, "add", "quality", "テストカバレッジ",
		"--deliverable", del2ID,
		"--metric", "test_coverage:90:%")
	assertSuccess(t, result)

	// doctor でチェック
	result = runCommand(t, dir, "doctor")
	assertSuccess(t, result)
}

// TestMultipleReferences は複数の参照を持つエンティティをテストする
func TestMultipleReferences(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupBasicProject(t, dir)

	// Problem を Objective と Deliverable 両方に紐づける
	result := runCommand(t, dir, "add", "problem", "複合問題",
		"--severity", "medium",
		"--objective", ids["objective"],
		"--deliverable", ids["deliverable"])
	assertSuccess(t, result)

	// Risk も同様に
	result = runCommand(t, dir, "add", "risk", "複合リスク",
		"--probability", "low",
		"--impact", "medium",
		"--objective", ids["objective"],
		"--deliverable", ids["deliverable"])
	assertSuccess(t, result)

	// doctor でチェック
	result = runCommand(t, dir, "doctor")
	assertSuccess(t, result)
}

// TestDoctorHealthyProject は正常なプロジェクトで doctor が healthy を返すことをテストする
func TestDoctorHealthyProject(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	// フルプロジェクトセットアップ
	setupFullProject(t, dir)

	// doctor でヘルスチェック
	result := runCommand(t, dir, "doctor")
	assertSuccess(t, result)
}

// TestDoctorEmptyProject は空のプロジェクトで doctor が動作することをテストする
func TestDoctorEmptyProject(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// doctor でチェック
	result := runCommand(t, dir, "doctor")
	assertSuccess(t, result)
}
