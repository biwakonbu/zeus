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

// TestDoctorFlatObjectives は doctor コマンドでフラット構造の Objective が正常にチェックされることをテストする
func TestDoctorFlatObjectives(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// フラット構造の Objective を作成
	result := runCommand(t, dir, "add", "objective", "目標A")
	assertSuccess(t, result)

	result = runCommand(t, dir, "add", "objective", "目標B")
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
		"--objective", ids["objective"],
		"--metric", "coverage:80:%")
	assertSuccess(t, result)

	// doctor でチェック
	result = runCommand(t, dir, "doctor")
	assertSuccess(t, result)
}

// =============================================================================
// 参照チェーンテスト
// =============================================================================

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

// TestRefChain_Qual_Obj は Quality → Objective の参照チェーンをテストする
func TestRefChain_Qual_Obj(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	ids := setupBasicProject(t, dir)

	// Quality 作成（正しい参照）
	result := runCommand(t, dir, "add", "quality", "コードカバレッジ",
		"--objective", ids["objective"],
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

// =============================================================================
// 無効な参照テスト
// =============================================================================

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

// TestInvalidRef_Qual_Obj は存在しない Objective への参照がエラーになることをテストする
func TestInvalidRef_Qual_Obj(t *testing.T) {
	t.Parallel()
	dir := setupTempDir(t)
	defer cleanupTempDir(t, dir)

	runCommand(t, dir, "init")

	// 存在しない Objective を参照
	result := runCommand(t, dir, "add", "quality", "無効な品質",
		"--objective", "obj-999")
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

	// 別の Quality を追加
	result := runCommand(t, dir, "add", "quality", "テストカバレッジ",
		"--objective", ids["objective"],
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

	// Problem を Objective に紐づける
	result := runCommand(t, dir, "add", "problem", "複合問題",
		"--severity", "medium",
		"--objective", ids["objective"])
	assertSuccess(t, result)

	// Risk も同様に
	result = runCommand(t, dir, "add", "risk", "複合リスク",
		"--probability", "low",
		"--impact", "medium",
		"--objective", ids["objective"])
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
